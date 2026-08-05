package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// schemaV10 is additive and intentionally leaves every new column nullable
// without a default. A NULL in a historical row is evidence of absence, not a
// request to reinterpret history using today's registry.
const schemaV10 = `
ALTER TABLE exit_states ADD COLUMN snapshot_status TEXT;
ALTER TABLE exit_states ADD COLUMN policy_version TEXT;
ALTER TABLE exit_states ADD COLUMN policy_digest TEXT;
ALTER TABLE exit_states ADD COLUMN snapshot_id TEXT;
ALTER TABLE exit_states ADD COLUMN decision_id TEXT;
ALTER TABLE exit_states ADD COLUMN observation_id TEXT;
ALTER TABLE exit_states ADD COLUMN position_generation INTEGER;
ALTER TABLE exit_states ADD COLUMN next_target TEXT;
ALTER TABLE exit_states ADD COLUMN next_protection TEXT;
ALTER TABLE exit_states ADD COLUMN last_observation_source TEXT;
ALTER TABLE exit_states ADD COLUMN last_observed_at TEXT;
ALTER TABLE exit_states ADD COLUMN snapshot_action TEXT;
ALTER TABLE exit_states ADD COLUMN snapshot_ratio TEXT;
ALTER TABLE exit_states ADD COLUMN projected_quantity TEXT;
ALTER TABLE exit_states ADD COLUMN state_only INTEGER;
ALTER TABLE exit_states ADD COLUMN suppressed_reason TEXT;
ALTER TABLE exit_states ADD COLUMN effective_snapshot_json TEXT;

ALTER TABLE exit_events ADD COLUMN position_generation INTEGER;
ALTER TABLE exit_events ADD COLUMN policy_id TEXT;
ALTER TABLE exit_events ADD COLUMN policy_version TEXT;
ALTER TABLE exit_events ADD COLUMN policy_digest TEXT;
ALTER TABLE exit_events ADD COLUMN snapshot_id TEXT;
ALTER TABLE exit_events ADD COLUMN decision_id TEXT;
ALTER TABLE exit_events ADD COLUMN observation_id TEXT;
ALTER TABLE exit_events ADD COLUMN next_target TEXT;
ALTER TABLE exit_events ADD COLUMN next_protection TEXT;
ALTER TABLE exit_events ADD COLUMN observation_source TEXT;
ALTER TABLE exit_events ADD COLUMN observed_at TEXT;
ALTER TABLE exit_events ADD COLUMN projected_quantity TEXT;
ALTER TABLE exit_events ADD COLUMN proposal_ratio TEXT;
ALTER TABLE exit_events ADD COLUMN state_only INTEGER;
ALTER TABLE exit_events ADD COLUMN suppressed_reason TEXT;
ALTER TABLE exit_events ADD COLUMN saved_snapshot_json TEXT;
ALTER TABLE exit_events ADD COLUMN recomputed_snapshot_json TEXT;
ALTER TABLE exit_events ADD COLUMN effective_snapshot_json TEXT;
ALTER TABLE exit_events ADD COLUMN effective_source TEXT;
ALTER TABLE exit_events ADD COLUMN arm_suppressed_reason TEXT;

CREATE UNIQUE INDEX idx_exit_events_decision ON exit_events(decision_id)
	WHERE decision_id IS NOT NULL;

CREATE TABLE exit_snapshot_quarantines (
	position_id          TEXT NOT NULL REFERENCES positions(id),
	position_generation  INTEGER NOT NULL,
	quarantine_version   INTEGER NOT NULL,
	reason               TEXT NOT NULL,
	evidence             TEXT NOT NULL,
	quarantined_at       TEXT NOT NULL,
	released_at          TEXT,
	release_kind         TEXT,
	release_evidence     TEXT,
	PRIMARY KEY(position_id, position_generation, quarantine_version)
) STRICT;

CREATE UNIQUE INDEX idx_exit_snapshot_quarantine_active
	ON exit_snapshot_quarantines(position_id, position_generation)
	WHERE released_at IS NULL;
`

const (
	SnapshotStatusSeed      = "SEED"
	SnapshotStatusEvaluated = "EVALUATED"

	EffectiveSourceRecomputed = "recomputed"
	EffectiveSourceSaved      = "saved_monotone"

	QuarantineReleaseHumanRepair            = "HUMAN_REPAIR"
	QuarantineReleaseAuthoritativeReconcile = "AUTHORITATIVE_RECONCILE"
	// QuarantineReleaseSelectorRevised — the recovery selector that produced the
	// quarantine is not the one running now, the re-judgement chose a verified
	// candidate, and the row closed on that evidence.
	//
	// A third kind rather than reuse: HUMAN_REPAIR names a person in the audit
	// trail and would be a lie here, and AUTHORITATIVE_RECONCILE names the
	// account, which decided nothing about this. "Which evidence closed it" is the
	// whole reason the column exists.
	QuarantineReleaseSelectorRevised = "SELECTOR_REVISED"

	ArmSuppressedWorkingOrder = "working_order_not_cleared"
)

var (
	ErrExitSnapshotCorrupt      = errors.New("journal: exit snapshot is corrupt")
	ErrExitSnapshotQuarantined  = errors.New("journal: exit snapshot generation is quarantined")
	ErrExitSnapshotReleaseStale = errors.New("journal: exit snapshot quarantine release is stale")
)

func (j *Journal) runExitWriteHook(stage string) error {
	if j.exitWriteHook == nil {
		return nil
	}
	return j.exitWriteHook(stage)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// StoredExitSnapshot is the exact value selected at evaluation time plus the
// observation clock metadata that is deliberately outside the pure evaluator.
type StoredExitSnapshot struct {
	Line              exitpolicy.ExitLineSnapshot         `json:"line"`
	RecoveryPolicy    exitpolicy.RecoveryPolicyDefinition `json:"recovery_policy"`
	ObservationSource string                              `json:"observation_source"`
	ObservedAt        string                              `json:"observed_at"`
	// OutputDigest seals every serialized output field, including the derived
	// next target/protection lines which are not part of a041's decision id.
	// This is corruption evidence, not order authority.
	OutputDigest string `json:"output_digest"`
}

func (s StoredExitSnapshot) validate() error {
	if _, _, err := exitpolicy.SelectRecoverySnapshot(nil, s.Line); err != nil {
		return err
	}
	if err := exitpolicy.ValidateRecoveryDerivation(s.Line, s.RecoveryPolicy); err != nil {
		return err
	}
	if strings.TrimSpace(s.ObservationSource) == "" {
		return errors.New("observation source is empty")
	}
	if _, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(s.ObservedAt)); err != nil {
		return fmt.Errorf("observed_at: %w", err)
	}
	if s.OutputDigest != "" {
		expected, err := snapshotOutputDigest(s.Line, s.RecoveryPolicy)
		if err != nil {
			return err
		}
		if s.OutputDigest != expected {
			return errors.New("output digest does not match the snapshot")
		}
	}
	return nil
}

type ExitSnapshotView struct {
	Snapshot      *StoredExitSnapshot
	UnknownReason string
	Stale         bool
	StaleReason   string
}

func (v ExitSnapshotView) WithFreshness(asOf time.Time, maxAge time.Duration) ExitSnapshotView {
	if v.Snapshot == nil || maxAge <= 0 || asOf.IsZero() {
		return v
	}
	observed, err := time.Parse(time.RFC3339Nano, v.Snapshot.ObservedAt)
	if err != nil {
		v.Stale, v.StaleReason = true, "invalid_observed_at"
		return v
	}
	if asOf.Sub(observed) > maxAge {
		v.Stale, v.StaleReason = true, "observation_older_than_limit"
	} else if observed.After(asOf) {
		v.Stale, v.StaleReason = true, "observation_in_future"
	}
	return v
}

type ExitEvaluation struct {
	Saved           ExitSnapshotView
	Recomputed      ExitSnapshotView
	Effective       ExitSnapshotView
	EffectiveSource string
}

func encodeStoredSnapshot(snapshot StoredExitSnapshot) (string, error) {
	digest, err := snapshotOutputDigest(snapshot.Line, snapshot.RecoveryPolicy)
	if err != nil {
		return "", err
	}
	snapshot.OutputDigest = digest
	if err := snapshot.validate(); err != nil {
		return "", err
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeStoredSnapshot(raw string) (StoredExitSnapshot, error) {
	var out StoredExitSnapshot
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return StoredExitSnapshot{}, err
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return StoredExitSnapshot{}, errors.New("trailing JSON value")
		}
		return StoredExitSnapshot{}, err
	}
	if err := out.validate(); err != nil {
		return StoredExitSnapshot{}, err
	}
	if strings.TrimSpace(out.OutputDigest) == "" {
		return StoredExitSnapshot{}, errors.New("output digest is absent")
	}
	return out, nil
}

func snapshotOutputDigest(line exitpolicy.ExitLineSnapshot, policy exitpolicy.RecoveryPolicyDefinition) (string, error) {
	raw, err := json.Marshal(struct {
		Line   exitpolicy.ExitLineSnapshot         `json:"line"`
		Policy exitpolicy.RecoveryPolicyDefinition `json:"policy"`
	}{Line: line, Policy: policy})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("sha256:%x", sum[:]), nil
}

func validateJudgementSnapshot(positionID string, judgement ExitJudgement, stored StoredExitSnapshot) error {
	if err := stored.validate(); err != nil {
		return err
	}
	line := stored.Line
	if line.PositionID != positionID || line.ObservationID != judgement.Provenance.ObservationID ||
		line.SnapshotID != judgement.Provenance.SnapshotID || line.DecisionID != judgement.Provenance.DecisionID ||
		line.Policy != judgement.Provenance.Policy || line.ObservedPrice != judgement.ObservedPrice ||
		line.HighWater != judgement.HighWater || line.CurrentProtection != judgement.Baseline ||
		string(line.RatchetLevel) != strings.TrimSpace(judgement.RatchetLevel) || line.ActiveRung != judgement.ActiveRung {
		return errors.New("judgement fields do not match immutable snapshot")
	}
	expected := line.ExecutableProposal()
	if expected.Zero() {
		if judgement.Proposal != nil || judgement.ArmSuppressedReason != "" {
			return errors.New("non-orderable snapshot cannot arm or suppress a proposal")
		}
		return nil
	}
	if judgement.Proposal == nil {
		if judgement.ArmSuppressedReason != ArmSuppressedWorkingOrder {
			return errors.New("orderable snapshot without a proposal needs a typed arm-suppression reason")
		}
		return nil
	}
	if judgement.ArmSuppressedReason != "" {
		return errors.New("armed proposal cannot also carry an arm-suppression reason")
	}
	if judgement.Proposal.Action != string(expected.Action) || judgement.Proposal.Level != expected.Level {
		return errors.New("armed proposal does not match immutable snapshot action and level")
	}
	return nil
}

func evaluationForEvent(saved, recomputed, effective *StoredExitSnapshot, source string) *ExitEvaluation {
	evaluation := &ExitEvaluation{EffectiveSource: source}
	if saved != nil {
		copy := *saved
		evaluation.Saved.Snapshot = &copy
	} else {
		evaluation.Saved.UnknownReason = "no_saved_evaluation"
	}
	if recomputed != nil {
		copy := *recomputed
		evaluation.Recomputed.Snapshot = &copy
	}
	if effective != nil {
		copy := *effective
		evaluation.Effective.Snapshot = &copy
	}
	return evaluation
}

func scanExitEvent(row rowScanner, event *ExitEvent) error {
	var evidence exitEventEvidence
	if err := row.Scan(&event.ID, &event.PositionID, &event.ObservedPrice, &event.HighWater,
		&event.BaselineAfter, &event.LevelAfter, &event.Action, &event.ProposedIntentID,
		&event.CreatedAt, &evidence.Generation, &evidence.PolicyID, &evidence.PolicyVersion,
		&evidence.PolicyDigest, &evidence.SnapshotID, &evidence.DecisionID, &evidence.ObservationID,
		&evidence.NextTarget, &evidence.NextProtection, &evidence.ObservationSource,
		&evidence.ObservedAt, &evidence.ProjectedQuantity, &evidence.ProposalRatio,
		&evidence.StateOnly, &evidence.SuppressedReason, &evidence.SavedJSON,
		&evidence.RecomputedJSON, &evidence.EffectiveJSON, &evidence.EffectiveSource,
		&evidence.ArmSuppressedReason, &event.LifecycleGeneration); err != nil {
		return err
	}
	if event.LifecycleGeneration < 1 {
		return fmt.Errorf("%w: invalid event lifecycle generation %d", ErrExitSnapshotCorrupt,
			event.LifecycleGeneration)
	}
	return hydrateExitEventEvidence(event, evidence)
}

type exitEventEvidence struct {
	Generation          sql.NullInt64
	PolicyID            sql.NullString
	PolicyVersion       sql.NullString
	PolicyDigest        sql.NullString
	SnapshotID          sql.NullString
	DecisionID          sql.NullString
	ObservationID       sql.NullString
	NextTarget          sql.NullString
	NextProtection      sql.NullString
	ObservationSource   sql.NullString
	ObservedAt          sql.NullString
	ProjectedQuantity   sql.NullString
	ProposalRatio       sql.NullString
	StateOnly           sql.NullInt64
	SuppressedReason    sql.NullString
	SavedJSON           sql.NullString
	RecomputedJSON      sql.NullString
	EffectiveJSON       sql.NullString
	EffectiveSource     sql.NullString
	ArmSuppressedReason sql.NullString
}

func (e exitEventEvidence) any() bool {
	return e.Generation.Valid || e.PolicyID.Valid || e.PolicyVersion.Valid || e.PolicyDigest.Valid ||
		e.SnapshotID.Valid || e.DecisionID.Valid || e.ObservationID.Valid || e.NextTarget.Valid ||
		e.NextProtection.Valid || e.ObservationSource.Valid || e.ObservedAt.Valid ||
		e.ProjectedQuantity.Valid || e.ProposalRatio.Valid || e.StateOnly.Valid ||
		e.SuppressedReason.Valid || e.SavedJSON.Valid || e.RecomputedJSON.Valid ||
		e.EffectiveJSON.Valid || e.EffectiveSource.Valid || e.ArmSuppressedReason.Valid
}

func (e exitEventEvidence) full() bool {
	return e.Generation.Valid && e.PolicyID.Valid && e.PolicyVersion.Valid && e.PolicyDigest.Valid &&
		e.SnapshotID.Valid && e.DecisionID.Valid && e.ObservationID.Valid &&
		e.ObservationSource.Valid && e.ObservedAt.Valid && e.ProjectedQuantity.Valid &&
		e.StateOnly.Valid && e.RecomputedJSON.Valid && e.EffectiveJSON.Valid && e.EffectiveSource.Valid
}

func hydrateExitEventEvidence(event *ExitEvent, evidence exitEventEvidence) error {
	if !evidence.any() {
		event.Evaluation.Saved.UnknownReason = "no_saved_evaluation"
		event.Evaluation.Recomputed.UnknownReason = "legacy_event"
		event.Evaluation.Effective.UnknownReason = "legacy_event"
		return nil
	}
	if !evidence.full() {
		return fmt.Errorf("%w: event evaluation tuple is partial", ErrExitSnapshotCorrupt)
	}
	event.Evaluation.EffectiveSource = evidence.EffectiveSource.String
	event.ArmSuppressedReason = evidence.ArmSuppressedReason.String
	if err := decodeExitEventSnapshot(evidence.SavedJSON, &event.Evaluation.Saved, "no_saved_evaluation"); err != nil {
		return err
	}
	if err := decodeExitEventSnapshot(evidence.RecomputedJSON, &event.Evaluation.Recomputed, "legacy_event"); err != nil {
		return err
	}
	if err := decodeExitEventSnapshot(evidence.EffectiveJSON, &event.Evaluation.Effective, "legacy_event"); err != nil {
		return err
	}
	if err := validateExitEventFlattenedEvidence(*event, evidence); err != nil {
		return err
	}
	return validateExitEventArmSuppression(*event)
}

func validateExitEventFlattenedEvidence(event ExitEvent, evidence exitEventEvidence) error {
	recomputed := event.Evaluation.Recomputed.Snapshot
	if recomputed == nil {
		return fmt.Errorf("%w: recomputed event snapshot is absent", ErrExitSnapshotCorrupt)
	}
	if evidence.StateOnly.Int64 != 0 && evidence.StateOnly.Int64 != 1 {
		return fmt.Errorf("%w: event state-only flag is not canonical", ErrExitSnapshotCorrupt)
	}
	line := recomputed.Line
	if evidence.Generation.Int64 != line.PositionGeneration || evidence.PolicyID.String != line.Policy.ID ||
		evidence.PolicyVersion.String != line.Policy.Version || evidence.PolicyDigest.String != line.Policy.Digest ||
		evidence.SnapshotID.String != line.SnapshotID || evidence.DecisionID.String != line.DecisionID ||
		evidence.ObservationID.String != line.ObservationID ||
		!nullableEventStringMatches(evidence.NextTarget, line.NextTarget) ||
		!nullableEventStringMatches(evidence.NextProtection, line.NextProtection) ||
		evidence.ObservationSource.String != recomputed.ObservationSource ||
		evidence.ObservedAt.String != recomputed.ObservedAt ||
		evidence.ProjectedQuantity.String != line.ProjectedQuantity ||
		!nullableEventStringMatches(evidence.ProposalRatio, line.Ratio) ||
		(line.StateOnly != (evidence.StateOnly.Int64 != 0)) ||
		!nullableEventStringMatches(evidence.SuppressedReason, line.Suppressed) {
		return fmt.Errorf("%w: flattened event columns do not match recomputed snapshot",
			ErrExitSnapshotCorrupt)
	}
	return nil
}

func nullableEventStringMatches(value sql.NullString, expected string) bool {
	present := strings.TrimSpace(expected) != ""
	return value.Valid == present && (!present || value.String == expected)
}

func validateExitEventArmSuppression(event ExitEvent) error {
	saved := event.Evaluation.Saved.Snapshot
	recomputed := event.Evaluation.Recomputed.Snapshot
	effective := event.Evaluation.Effective.Snapshot
	source := strings.TrimSpace(event.Evaluation.EffectiveSource)
	reason := strings.TrimSpace(event.ArmSuppressedReason)
	if saved == nil && recomputed == nil && effective == nil && source == "" {
		if reason != "" {
			return fmt.Errorf("%w: legacy event cannot assert arm-suppression evidence",
				ErrExitSnapshotCorrupt)
		}
		return nil
	}
	if recomputed == nil || effective == nil {
		return fmt.Errorf("%w: event evaluation tuple is incomplete", ErrExitSnapshotCorrupt)
	}
	if recomputed.Line.PositionID != event.PositionID || effective.Line.PositionID != event.PositionID ||
		(saved != nil && saved.Line.PositionID != event.PositionID) {
		return fmt.Errorf("%w: event evaluation belongs to another position", ErrExitSnapshotCorrupt)
	}
	var expectedSource exitpolicy.RecoverySource
	switch source {
	case EffectiveSourceRecomputed:
		if !reflect.DeepEqual(*effective, *recomputed) {
			return fmt.Errorf("%w: recomputed source does not match effective snapshot",
				ErrExitSnapshotCorrupt)
		}
		expectedSource = exitpolicy.RecoveryRecomputed
	case EffectiveSourceSaved:
		if saved == nil || !reflect.DeepEqual(*effective, *saved) {
			return fmt.Errorf("%w: saved source does not match effective snapshot",
				ErrExitSnapshotCorrupt)
		}
		expectedSource = exitpolicy.RecoverySavedMonotone
	default:
		return fmt.Errorf("%w: unknown effective source %q", ErrExitSnapshotCorrupt, source)
	}
	var savedLine *exitpolicy.ExitLineSnapshot
	if saved != nil {
		line := saved.Line
		savedLine = &line
	}
	selected, selectedSource, err := exitpolicy.SelectRecoverySnapshot(savedLine, recomputed.Line)
	if err != nil || selectedSource != expectedSource || selected != effective.Line {
		return fmt.Errorf("%w: effective event snapshot does not match recovery selection",
			ErrExitSnapshotCorrupt)
	}
	if reason != "" && reason != ArmSuppressedWorkingOrder {
		return fmt.Errorf("%w: unknown arm-suppression reason %q", ErrExitSnapshotCorrupt, reason)
	}
	armed := event.Action != "" || event.ProposedIntentID != ""
	if (event.Action == "") != (event.ProposedIntentID == "") {
		return fmt.Errorf("%w: event arm action and intent evidence are partial", ErrExitSnapshotCorrupt)
	}
	if source == EffectiveSourceSaved {
		if armed || reason != "" {
			return fmt.Errorf("%w: saved-monotone event cannot carry arm evidence", ErrExitSnapshotCorrupt)
		}
		return nil
	}
	if recomputed.Line.Orderable && armed {
		if reason != "" || event.Action != string(recomputed.Line.Action) {
			return fmt.Errorf("%w: armed event contradicts immutable recomputed proposal",
				ErrExitSnapshotCorrupt)
		}
		return nil
	}
	if recomputed.Line.Orderable && !armed && reason != ArmSuppressedWorkingOrder {
		return fmt.Errorf("%w: orderable event without an arm needs typed suppression evidence",
			ErrExitSnapshotCorrupt)
	}
	if !recomputed.Line.Orderable && (armed || reason != "") {
		return fmt.Errorf("%w: arm-suppression reason lacks complete orderable evaluation evidence",
			ErrExitSnapshotCorrupt)
	}
	return nil
}

func decodeExitEventSnapshot(raw sql.NullString, view *ExitSnapshotView, absent string) error {
	if !raw.Valid {
		view.UnknownReason = absent
		return nil
	}
	snapshot, err := decodeStoredSnapshot(raw.String)
	if err != nil {
		view.UnknownReason = "invalid_stored_snapshot"
		return fmt.Errorf("%w: %v", ErrExitSnapshotCorrupt, err)
	}
	view.Snapshot = &snapshot
	return nil
}

func quarantineExitSnapshotTx(ctx context.Context, tx *sql.Tx, id string, generation int64,
	reason, evidence, now string) (ExitSnapshotQuarantine, error) {
	if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation); err != nil {
		return ExitSnapshotQuarantine{}, err
	} else if ok {
		if !active.NeedsReJudgement() {
			return active, nil
		}
		// A different selector wrote the active row, this build re-judged the
		// generation, and it still cannot choose one verified candidate. Closing the
		// old row and opening one stamped with the current revision is what keeps
		// the retry to once per revision — leaving the old row would make every
		// later observation re-judge the same frozen inputs forever (a084 D6).
		if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSelectorRevised,
			"the recovery selector changed since this quarantine was written; the generation was "+
				"re-judged and the cause still holds, so it is recorded again under the current selector",
			now); err != nil {
			return ExitSnapshotQuarantine{}, err
		}
	}
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+1 FROM exit_snapshot_quarantines
		WHERE position_id=? AND position_generation=?`, id, generation).Scan(&version); err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	q := ExitSnapshotQuarantine{PositionID: id, PositionGeneration: generation, Version: version,
		Reason: reason, Evidence: evidence, QuarantinedAt: now,
		SelectorRevision: exitpolicy.RecoverySelectorRevision}
	_, err := tx.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines
		(position_id,position_generation,quarantine_version,reason,evidence,quarantined_at,selector_revision)
		VALUES(?,?,?,?,?,?,?)`, id, generation, version, reason, evidence, now, q.SelectorRevision)
	return q, err
}

type ExitStateResult struct {
	State      ExitState
	Corruption error
}

func scanExitStateResult(row rowScanner) (ExitStateResult, error) {
	var (
		s                                                                        ExitState
		policyID, status, version, digest, snapshotID, decisionID, observationID sql.NullString
		nextTarget, nextProtection, source, observedAt, action, ratio            sql.NullString
		projected, suppressed, effectiveJSON                                     sql.NullString
		generation, rung                                                         sql.NullInt64
		stateOnly                                                                sql.NullInt64
		done                                                                     int
	)
	err := row.Scan(&s.PositionID, &s.PolicyKind, &policyID, &s.EntryPrice, &s.InitialStop, &s.InitialRisk,
		&s.Baseline, &s.HighWater, &s.RatchetLevel, &rung, &s.TakenRatioTotal,
		&s.PendingAction, &s.PendingLevel, &s.PendingIntentID, &done, &s.UpdatedAt,
		&status, &version, &digest, &snapshotID, &decisionID, &observationID, &generation,
		&nextTarget, &nextProtection, &source, &observedAt, &action, &ratio, &projected,
		&stateOnly, &suppressed, &effectiveJSON, &s.LifecycleGeneration)
	if errors.Is(err, sql.ErrNoRows) {
		return ExitStateResult{}, ErrExitStateNotFound
	}
	if err != nil {
		return ExitStateResult{}, fmt.Errorf("journal: reading an exit state: %w", err)
	}
	s.ActiveRung = exitpolicy.NoRung
	if rung.Valid {
		s.ActiveRung = int(rung.Int64)
	}
	s.Completed = done != 0
	if s.LifecycleGeneration < 1 {
		return ExitStateResult{}, fmt.Errorf("%w: invalid exit lifecycle generation %d", ErrExitSnapshotCorrupt,
			s.LifecycleGeneration)
	}
	s.PositionGeneration = -1
	if policyID.Valid {
		s.PolicyID = policyID.String
	}
	if generation.Valid {
		s.PositionGeneration = generation.Int64
	}
	if status.Valid {
		s.SnapshotStatus = status.String
	}

	result := ExitStateResult{State: s}
	// policy_id predates v10. Its presence alone is legacy evidence that can be
	// resolved against a041's pinned compatibility table at read time; it must
	// not be mistaken for a partially written v10 tuple.
	v10Evidence := []bool{status.Valid, version.Valid, digest.Valid, snapshotID.Valid,
		decisionID.Valid, observationID.Valid, generation.Valid, nextTarget.Valid,
		nextProtection.Valid, source.Valid, observedAt.Valid, action.Valid, ratio.Valid,
		projected.Valid, stateOnly.Valid, suppressed.Valid, effectiveJSON.Valid}
	anyV10 := false
	for _, present := range v10Evidence {
		anyV10 = anyV10 || present
	}
	if !anyV10 {
		result.State.Snapshot.UnknownReason = "legacy_snapshot_absent"
		// RATCHET and all non-RUNNER legacy ladder IDs have one exact a041
		// meaning independent of adoption context. Resolve that value in memory
		// only; the NULL v10 columns remain untouched on disk.
		if s.PolicyKind != ExitPolicyLadder || strings.TrimSpace(s.PolicyID) != exitpolicy.CommonLadderRunner {
			identity, identityErr := legacyPolicyIdentity(s, false)
			if identityErr == nil {
				result.State.PolicyIdentity = identity
			} else {
				result.State.Snapshot.UnknownReason = "legacy_policy_identity_unknown"
			}
		} else {
			result.State.Snapshot.UnknownReason = "legacy_adoption_context_required"
		}
		return result, nil
	}
	if !status.Valid {
		reason := "partial_snapshot_tuple"
		result.State.Snapshot.UnknownReason = reason
		result.Corruption = fmt.Errorf("%w: %s", ErrExitSnapshotCorrupt, reason)
		return result, nil
	}
	if !policyID.Valid || !version.Valid || !digest.Valid || !generation.Valid {
		result.State.Snapshot.UnknownReason = "partial_policy_tuple"
		result.Corruption = fmt.Errorf("%w: partial policy tuple", ErrExitSnapshotCorrupt)
		return result, nil
	}
	identity := exitpolicy.PolicyIdentity{ID: policyID.String, Version: version.String, Digest: digest.String}
	if err := identity.Validate(); err != nil {
		result.State.Snapshot.UnknownReason = "invalid_policy_identity"
		result.Corruption = fmt.Errorf("%w: %v", ErrExitSnapshotCorrupt, err)
		return result, nil
	}
	result.State.PolicyIdentity = identity
	if status.String == SnapshotStatusSeed {
		if snapshotID.Valid || decisionID.Valid || observationID.Valid || nextTarget.Valid ||
			nextProtection.Valid || source.Valid || observedAt.Valid || action.Valid || ratio.Valid ||
			projected.Valid || stateOnly.Valid || suppressed.Valid || effectiveJSON.Valid {
			result.State.Snapshot.UnknownReason = "partial_seed_tuple"
			result.Corruption = fmt.Errorf("%w: seed carries partial evaluation evidence", ErrExitSnapshotCorrupt)
		} else {
			result.State.Snapshot.UnknownReason = "not_evaluated_yet"
		}
		return result, nil
	}
	full := status.String == SnapshotStatusEvaluated && snapshotID.Valid && decisionID.Valid &&
		observationID.Valid && generation.Valid && source.Valid && observedAt.Valid &&
		projected.Valid && stateOnly.Valid && effectiveJSON.Valid
	if !full {
		result.State.Snapshot.UnknownReason = "partial_evaluated_tuple"
		result.Corruption = fmt.Errorf("%w: evaluated snapshot tuple is incomplete", ErrExitSnapshotCorrupt)
		return result, nil
	}
	stored, err := decodeStoredSnapshot(effectiveJSON.String)
	if err != nil {
		result.State.Snapshot.UnknownReason = "invalid_effective_snapshot"
		result.Corruption = fmt.Errorf("%w: %v", ErrExitSnapshotCorrupt, err)
		return result, nil
	}
	line := stored.Line
	if line.Policy != identity || line.PositionID != s.PositionID || line.PositionGeneration != generation.Int64 ||
		line.SnapshotID != snapshotID.String || line.DecisionID != decisionID.String ||
		line.ObservationID != observationID.String || line.CurrentProtection != s.Baseline ||
		line.HighWater != s.HighWater || line.NextTarget != nextTarget.String ||
		line.NextProtection != nextProtection.String || string(line.Action) != action.String ||
		line.Ratio != ratio.String || line.ProjectedQuantity != projected.String || line.StateOnly != (stateOnly.Valid && stateOnly.Int64 != 0) ||
		line.Suppressed != suppressed.String || stored.ObservationSource != source.String || stored.ObservedAt != observedAt.String {
		result.State.Snapshot.UnknownReason = "flattened_snapshot_mismatch"
		result.Corruption = fmt.Errorf("%w: flattened columns do not match effective snapshot", ErrExitSnapshotCorrupt)
		return result, nil
	}
	result.State.Snapshot.Snapshot = &stored
	return result, nil
}

func legacyPolicyIdentity(state ExitState, adopted bool) (exitpolicy.PolicyIdentity, error) {
	if state.PolicyKind == ExitPolicyLadder {
		return exitpolicy.LegacyLadderPolicyIdentity(state.PolicyID, adopted)
	}
	return exitpolicy.LegacyRatchetPolicyIdentity()
}

func (j *Journal) OpenExitStateResults(ctx context.Context, accountRef string) ([]ExitStateResult, error) {
	rows, err := j.db.QueryContext(ctx, exitStateSelect+` e JOIN positions p ON p.id=e.position_id
		WHERE p.account_ref=? AND e.completed=0`+currentManagedExitLifecycle+`
		ORDER BY p.market,p.symbol,p.instance_seq`, strings.TrimSpace(accountRef))
	if err != nil {
		return nil, fmt.Errorf("journal: listing exit state results: %w", err)
	}
	defer rows.Close()
	var out []ExitStateResult
	for rows.Next() {
		result, err := scanExitStateResult(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing exit state results: %w", err)
	}
	return out, nil
}

type ExitSnapshotQuarantine struct {
	PositionID         string
	PositionGeneration int64
	Version            int64
	Reason             string
	Evidence           string
	QuarantinedAt      string
	ReleasedAt         string
	ReleaseKind        string
	ReleaseEvidence    string
	// SelectorRevision is the recovery-selector revision that produced this
	// judgement. Zero means the row predates the column: which selector decided it
	// is unknown, which is a different fact from "the current one decided it" and
	// is why the migration does not backfill (a084).
	SelectorRevision int64
}

// NeedsReJudgement reports that the selector which produced this quarantine is
// not the one running now, so the same comparison may reach a different answer.
//
// Unknown counts as different. A row written before the stamp existed was, by
// construction, judged by an older selector — the stamp was added because one of
// them was wrong.
func (q ExitSnapshotQuarantine) NeedsReJudgement() bool {
	return q.SelectorRevision != exitpolicy.RecoverySelectorRevision
}

func (j *Journal) QuarantineExitSnapshot(ctx context.Context, positionID string, generation int64,
	reason, evidence string) (ExitSnapshotQuarantine, error) {
	id, why, proof := strings.TrimSpace(positionID), strings.TrimSpace(reason), strings.TrimSpace(evidence)
	if id == "" || generation < 0 || why == "" || proof == "" {
		return ExitSnapshotQuarantine{}, fmt.Errorf("%w: quarantine needs position, generation, reason, and evidence", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	defer tx.Rollback()
	var actual int64
	if err := tx.QueryRowContext(ctx, `SELECT instance_seq FROM positions WHERE id=?`, id).Scan(&actual); err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	if actual != generation {
		return ExitSnapshotQuarantine{}, fmt.Errorf("%w: position %s is generation %d, not %d", ErrInvalidRequest, id, actual, generation)
	}
	if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation); err != nil {
		return ExitSnapshotQuarantine{}, err
	} else if ok {
		if !active.NeedsReJudgement() {
			return active, nil
		}
		// Same rule as the judgement path: a superseded row is closed on the
		// selector-revision evidence and replaced by one this build owns, so the
		// retry cannot repeat at this revision (a084 D6).
		if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSelectorRevised,
			"the recovery selector changed since this quarantine was written; the generation was "+
				"re-judged and a cause still holds, so it is recorded again under the current selector",
			j.nowString()); err != nil {
			return ExitSnapshotQuarantine{}, err
		}
	}
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+1 FROM exit_snapshot_quarantines WHERE position_id=? AND position_generation=?`, id, generation).Scan(&version); err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	q := ExitSnapshotQuarantine{PositionID: id, PositionGeneration: generation, Version: version,
		Reason: why, Evidence: proof, QuarantinedAt: j.nowString(),
		SelectorRevision: exitpolicy.RecoverySelectorRevision}
	if _, err := tx.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines
		(position_id,position_generation,quarantine_version,reason,evidence,quarantined_at,selector_revision)
		VALUES(?,?,?,?,?,?,?)`, id, generation, version, q.Reason, q.Evidence, q.QuarantinedAt,
		q.SelectorRevision); err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExitSnapshotQuarantine{}, err
	}
	return q, nil
}

func activeExitSnapshotQuarantineTx(ctx context.Context, tx *sql.Tx, id string, generation int64) (ExitSnapshotQuarantine, bool, error) {
	var (
		q        ExitSnapshotQuarantine
		revision sql.NullInt64
	)
	err := tx.QueryRowContext(ctx, `SELECT position_id,position_generation,quarantine_version,reason,evidence,quarantined_at,selector_revision
		FROM exit_snapshot_quarantines WHERE position_id=? AND position_generation=? AND released_at IS NULL`, id, generation).
		Scan(&q.PositionID, &q.PositionGeneration, &q.Version, &q.Reason, &q.Evidence, &q.QuarantinedAt, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		return ExitSnapshotQuarantine{}, false, nil
	}
	// NULL stays zero rather than becoming the current revision: the reader must
	// not manufacture the very fact the column exists to record (a084).
	q.SelectorRevision = revision.Int64
	return q, err == nil, err
}

// releaseExitSnapshotQuarantineTx closes one active row inside a caller's
// transaction.
//
// It exists so a re-judgement's release and the judgement it earned commit
// together. Releasing outside the judgement transaction and then failing to
// record the judgement would leave a generation with no quarantine and no
// evaluation, which the next observation reads as healthy — the one outcome this
// change must not be able to produce (a084 D8).
func releaseExitSnapshotQuarantineTx(ctx context.Context, tx *sql.Tx,
	active ExitSnapshotQuarantine, kind, evidence, now string) error {
	result, err := tx.ExecContext(ctx, `UPDATE exit_snapshot_quarantines
		SET released_at=?,release_kind=?,release_evidence=?
		WHERE position_id=? AND position_generation=? AND quarantine_version=? AND released_at IS NULL`,
		now, kind, evidence, active.PositionID, active.PositionGeneration, active.Version)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrExitSnapshotReleaseStale
	}
	return nil
}

// releaseReJudgedQuarantineTx closes an active quarantine whose selector has been
// superseded, when the re-judgement it allowed chose a verified candidate.
//
// It is a no-op when there is no active row, and when the active row was written
// by the selector running now: that generation was never re-judged, so nothing
// about it has been re-decided and releasing it would be a release nobody earned.
func releaseReJudgedQuarantineTx(ctx context.Context, tx *sql.Tx, id string,
	generation int64, now string) error {
	active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation)
	if err != nil || !ok || !active.NeedsReJudgement() {
		return err
	}
	return releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSelectorRevised,
		"the recovery selector changed since this quarantine was written; the re-judgement chose "+
			"one verified candidate and this generation is judged again", now)
}

func (j *Journal) ReleaseExitSnapshotQuarantine(ctx context.Context, positionID string, generation, expectedVersion int64,
	kind, evidence string) error {
	kind = strings.TrimSpace(kind)
	if strings.TrimSpace(positionID) == "" || generation < 0 || expectedVersion <= 0 || strings.TrimSpace(evidence) == "" {
		return fmt.Errorf("%w: release needs position, generation, positive version, and evidence", ErrInvalidRequest)
	}
	if kind != QuarantineReleaseHumanRepair && kind != QuarantineReleaseAuthoritativeReconcile &&
		kind != QuarantineReleaseSelectorRevised {
		return fmt.Errorf("%w: invalid quarantine release kind %q", ErrInvalidRequest, kind)
	}
	result, err := j.db.ExecContext(ctx, `UPDATE exit_snapshot_quarantines
		SET released_at=?,release_kind=?,release_evidence=?
		WHERE position_id=? AND position_generation=? AND quarantine_version=? AND released_at IS NULL`,
		j.nowString(), kind, strings.TrimSpace(evidence), strings.TrimSpace(positionID), generation, expectedVersion)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrExitSnapshotReleaseStale
	}
	return nil
}

func (j *Journal) ActiveExitSnapshotQuarantine(ctx context.Context, positionID string, generation int64) (ExitSnapshotQuarantine, bool, error) {
	tx, err := j.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ExitSnapshotQuarantine{}, false, err
	}
	defer tx.Rollback()
	q, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, strings.TrimSpace(positionID), generation)
	if err != nil {
		return ExitSnapshotQuarantine{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return ExitSnapshotQuarantine{}, false, err
	}
	return q, ok, nil
}
