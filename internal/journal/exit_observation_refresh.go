package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// ErrExitObservationConflict means a refresh no longer describes the current
// evaluated state. Callers must re-read on the next observation; it never
// authorises a write, proposal, or order.
var ErrExitObservationConflict = errors.New("journal: exit observation refresh conflicts with current state")

// ErrExitObservationStale means the candidate is older than durable evidence.
var ErrExitObservationStale = errors.New("journal: exit observation refresh is stale")

// ExitObservationRefresh replaces only the observation-bound representation of
// an already evaluated, semantically unchanged effective line. It is purposely
// separate from ExitJudgement: a heartbeat is not another exit decision.
type ExitObservationRefresh struct {
	PositionID          string
	LifecycleGeneration int64
	Snapshot            exitpolicy.ExitLineSnapshot
	RecoveryPolicy      exitpolicy.RecoveryPolicyDefinition
	ObservationSource   string
	ObservedAt          time.Time
	Provenance          ExitDecisionProvenance
}

// RefreshExitObservation atomically writes a newer complete effective snapshot
// without appending an exit event or touching proposal state.
func (j *Journal) RefreshExitObservation(ctx context.Context, request ExitObservationRefresh) error {
	id := strings.TrimSpace(request.PositionID)
	if id == "" {
		return fmt.Errorf("%w: an observation refresh needs a position", ErrInvalidRequest)
	}
	if err := request.Provenance.validate(); err != nil || request.Provenance.zero() {
		if err != nil {
			return err
		}
		return fmt.Errorf("%w: an observation refresh needs complete provenance", ErrInvalidRequest)
	}
	if !sameExitDecisionProvenance(request.Provenance, ExitDecisionProvenance{
		ObservationID: request.Snapshot.ObservationID, SnapshotID: request.Snapshot.SnapshotID,
		DecisionID: request.Snapshot.DecisionID, Policy: request.Snapshot.Policy,
	}) {
		return fmt.Errorf("%w: refresh provenance must match its snapshot", ErrInvalidRequest)
	}
	if request.ObservedAt.IsZero() {
		return fmt.Errorf("%w: refresh observed_at is required", ErrInvalidRequest)
	}
	source := strings.TrimSpace(request.ObservationSource)
	if _, _, ok := observationSourceOrder(source); !ok {
		return fmt.Errorf("%w: invalid observation source %q", ErrExitObservationConflict, source)
	}
	if request.Snapshot.Orderable || !request.Snapshot.ExecutableProposal().Zero() {
		return fmt.Errorf("%w: orderable snapshots require a full judgement", ErrExitObservationConflict)
	}
	candidate := StoredExitSnapshot{
		Line: request.Snapshot, RecoveryPolicy: request.RecoveryPolicy,
		ObservationSource: source, ObservedAt: request.ObservedAt.UTC().Format(time.RFC3339Nano),
	}
	judgement := ExitJudgement{PositionID: id, LifecycleGeneration: request.LifecycleGeneration,
		Snapshot: request.Snapshot, RecoveryPolicy: request.RecoveryPolicy,
		ObservationSource: source, ObservedAt: request.ObservedAt, Provenance: request.Provenance,
		ObservedPrice: request.Snapshot.ObservedPrice, HighWater: request.Snapshot.HighWater,
		Baseline: request.Snapshot.CurrentProtection, RatchetLevel: string(request.Snapshot.RatchetLevel),
		ActiveRung: request.Snapshot.ActiveRung}
	if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE serializes temporal CAS.
	if err != nil {
		return fmt.Errorf("journal: refreshing exit observation of %s: %w", id, err)
	}
	defer tx.Rollback()
	current, err := scanExitProgress(ctx, tx, id)
	if err != nil {
		return err
	}
	if current.Completed {
		return fmt.Errorf("%w: %s", ErrExitStateCompleted, id)
	}
	if current.Effective == nil {
		return fmt.Errorf("%w: %s is not evaluated", ErrExitObservationConflict, id)
	}
	status, proposalPending, err := exitObservationRefreshGuardTx(ctx, tx, id)
	if err != nil {
		return err
	}
	if status != SnapshotStatusEvaluated || proposalPending {
		return fmt.Errorf("%w: refresh requires evaluated state without a pending proposal", ErrExitObservationConflict)
	}
	expectedLifecycle := request.LifecycleGeneration
	if expectedLifecycle == 0 {
		expectedLifecycle = current.LifecycleGeneration
	}
	if expectedLifecycle != current.LifecycleGeneration {
		return fmt.Errorf("%w: lifecycle generation changed", ErrExitObservationConflict)
	}
	var lifecycleGeneration int64
	var lifecycleStatus positionpolicy.Status
	err = tx.QueryRowContext(ctx, `SELECT adoption_generation,status FROM position_policy_lifecycles
		WHERE position_id=? ORDER BY adoption_generation DESC LIMIT 1`, id).Scan(&lifecycleGeneration, &lifecycleStatus)
	if errors.Is(err, sql.ErrNoRows) {
		lifecycleGeneration, lifecycleStatus = 1, positionpolicy.StatusManaged
	} else if err != nil {
		return err
	}
	if lifecycleStatus != positionpolicy.StatusManaged || lifecycleGeneration != expectedLifecycle {
		return fmt.Errorf("%w: lifecycle is no longer managed", ErrExitObservationConflict)
	}
	decision, err := compareObservationEvidence(*current.Effective, candidate)
	if err != nil {
		return err
	}
	if decision == observationNoop {
		return nil
	}
	if current.PositionGeneration != request.Snapshot.PositionGeneration ||
		!sameExitOperationalLine(current.Effective.Line, request.Snapshot) {
		return fmt.Errorf("%w: candidate changes effective exit semantics", ErrExitObservationConflict)
	}
	raw, err := encodeStoredSnapshot(candidate)
	if err != nil {
		return err
	}
	line := candidate.Line
	_, err = tx.ExecContext(ctx, `UPDATE exit_states SET
		baseline_price=?,high_water=?,ratchet_level=?,active_rung=?,updated_at=?,
		snapshot_status=?,policy_id=?,policy_version=?,policy_digest=?,snapshot_id=?,decision_id=?,observation_id=?,
		position_generation=?,next_target=?,next_protection=?,last_observation_source=?,last_observed_at=?,
		snapshot_action=?,snapshot_ratio=?,projected_quantity=?,state_only=?,suppressed_reason=?,effective_snapshot_json=?
		WHERE position_id=?`,
		line.CurrentProtection, line.HighWater, string(line.RatchetLevel), nullableRung(line.ActiveRung), j.nowString(),
		SnapshotStatusEvaluated, line.Policy.ID, line.Policy.Version, line.Policy.Digest, line.SnapshotID,
		line.DecisionID, line.ObservationID, line.PositionGeneration, nullableString(line.NextTarget),
		nullableString(line.NextProtection), candidate.ObservationSource, candidate.ObservedAt,
		nullableString(string(line.Action)), nullableString(line.Ratio), line.ProjectedQuantity, boolInt(line.StateOnly),
		nullableString(line.Suppressed), raw, id)
	if err != nil {
		return fmt.Errorf("journal: refreshing exit observation of %s: %w", id, err)
	}
	if err := j.runExitWriteHook("after_refresh_state"); err != nil {
		return err
	}
	return tx.Commit()
}

// sameExitOperationalLine is D1's semantic classifier. Provenance and the
// observed price intentionally vary on a valid heartbeat; every field that can
// alter execution, recovery, or operator line semantics must remain identical.
func sameExitOperationalLine(a, b exitpolicy.ExitLineSnapshot) bool {
	return a.Policy == b.Policy && a.PositionID == b.PositionID &&
		a.PositionGeneration == b.PositionGeneration && a.EntryPrice == b.EntryPrice &&
		a.InitialStop == b.InitialStop && a.CurrentProtection == b.CurrentProtection &&
		a.HighWater == b.HighWater && a.RatchetLevel == b.RatchetLevel && a.ActiveRung == b.ActiveRung &&
		a.NextTarget == b.NextTarget && a.NextProtection == b.NextProtection && a.Action == b.Action &&
		a.Level == b.Level && a.Ratio == b.Ratio && a.ProjectedQuantity == b.ProjectedQuantity &&
		a.Orderable == b.Orderable && a.StateOnly == b.StateOnly && a.Suppressed == b.Suppressed &&
		a.CancelPendingFirst == b.CancelPendingFirst
}

type observationDecision uint8

const (
	observationReplace observationDecision = iota
	observationNoop
)

func observationSourceOrder(source string) (rank int, sequence uint64, ok bool) {
	if source == "quote_fetched_at" {
		return 2, 0, true
	}
	if !strings.HasPrefix(source, "cycle:") {
		return 0, 0, false
	}
	n, err := strconv.ParseUint(strings.TrimPrefix(source, "cycle:"), 10, 64)
	if err != nil || n == 0 {
		return 0, 0, false
	}
	return 1, n, true
}

func compareObservationEvidence(current, candidate StoredExitSnapshot) (observationDecision, error) {
	curAt, err := time.Parse(time.RFC3339Nano, current.ObservedAt)
	if err != nil {
		return 0, fmt.Errorf("%w: current observed_at is invalid", ErrExitObservationConflict)
	}
	candAt, err := time.Parse(time.RFC3339Nano, candidate.ObservedAt)
	if err != nil {
		return 0, fmt.Errorf("%w: candidate observed_at is invalid", ErrExitObservationConflict)
	}
	if candAt.Before(curAt) {
		return 0, ErrExitObservationStale
	}
	if candAt.After(curAt) {
		return observationReplace, nil
	}
	curRank, curSequence, curOK := observationSourceOrder(current.ObservationSource)
	candRank, candSequence, candOK := observationSourceOrder(candidate.ObservationSource)
	if !curOK || !candOK {
		return 0, fmt.Errorf("%w: equal-time source is malformed", ErrExitObservationConflict)
	}
	if candRank > curRank {
		return observationReplace, nil
	}
	if candRank < curRank {
		return 0, ErrExitObservationStale
	}
	if candRank == 1 { // cycle:N total order at a frozen timestamp.
		switch {
		case candSequence > curSequence:
			return observationReplace, nil
		case candSequence < curSequence:
			return 0, ErrExitObservationStale
		}
	}
	if candidate.Line.ObservationID == current.Line.ObservationID &&
		candidate.Line.SnapshotID == current.Line.SnapshotID &&
		candidate.Line.DecisionID == current.Line.DecisionID &&
		candidate.ObservationSource == current.ObservationSource {
		return observationNoop, nil
	}
	return 0, fmt.Errorf("%w: equal-time evidence has a different identity", ErrExitObservationConflict)
}

// MaxExitObservationCycle returns the largest valid fallback sequence in the
// current managed working set. It lets a restarted observer continue the
// durable cycle:N order instead of reusing a frozen-clock identity.
func (j *Journal) MaxExitObservationCycle(ctx context.Context, accountRef string) (uint64, error) {
	rows, err := j.db.QueryContext(ctx, `SELECT coalesce(e.effective_snapshot_json,'')
		FROM exit_states e JOIN positions p ON p.id=e.position_id
		WHERE p.account_ref=? AND e.completed=0`+currentManagedExitLifecycle, strings.TrimSpace(accountRef))
	if err != nil {
		return 0, fmt.Errorf("journal: reading maximum exit observation cycle: %w", err)
	}
	defer rows.Close()
	var maximum uint64
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return 0, err
		}
		stored, err := decodeStoredSnapshot(raw)
		if err != nil {
			continue // corrupt or incomplete state must not manufacture an ordering claim.
		}
		if rank, sequence, ok := observationSourceOrder(strings.TrimSpace(stored.ObservationSource)); ok && rank == 1 && sequence > maximum {
			maximum = sequence
		}
	}
	return maximum, rows.Err()
}

// SameExitOperationalLine exposes the observer-side classifier while retaining
// the journal's single definition of the fields a refresh may never change.
func SameExitOperationalLine(a, b exitpolicy.ExitLineSnapshot) bool {
	return sameExitOperationalLine(a, b)
}
