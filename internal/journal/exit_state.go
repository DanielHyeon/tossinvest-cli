package journal

// exit_state.go is the storage half of the exit policy (change add-core-domain
// task 7.3; exit-policy, design D5/D7).
//
// # The split, and the one that is enforced rather than agreed
//
// internal/exitpolicy owns the rule — the trigger table, the rung set, the
// composition, the monotone guarantees. This file owns the row those values live
// in: opening it at t0, advancing it when a judgement moves something, and
// appending the judgement history the provenance join reads.
//
// It owns *most* of that row. Four columns — the cumulative taken fraction and
// the three that make up a pending proposal — are written only in apply_hook.go,
// and the rule is enforced by a test that fails if any other production file in
// this package so much as spells their names
// (TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook). That is why the arming
// step of a judgement is a call into that file rather than a statement here, and
// why this file's own reads never select them: the judgement transaction reaches
// them through armExitProposalTx and through nothing else.
//
// # What a judgement writes
//
// One transaction per evaluation (RecordExitJudgement): the advanced state, the
// proposal if there is one, and the history row. They cannot be three commits —
// a crash between arming and recording would leave a proposal nobody can explain,
// and a crash between recording and arming would re-propose the same level at the
// next observation. "발의 기록 후 제출 전 크래시" is recoverable precisely because
// the record and the arming are one write and the submission is the next one.
//
// # Two invariants are re-checked here
//
// The baseline and the watermark never descend. internal/exitpolicy guarantees
// it and the schema comments assert it, and this file refuses a write that would
// break it anyway — because the guarantee is about a decision function and the
// column is about every writer that will ever exist, including a future one that
// reconstructs a state from somewhere else.
//
// # Broker-behaviour claims
//
// None. Every value here is either an internal decision or a price the caller
// observed; this file adds no interpretation of either.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrPositionNotExitEligible means the position carries neither an entry
// decision nor an adoption record, so there is no stop to build a baseline out
// of and no risk to measure R against.
//
// exit-policy makes this a refusal plus an operator alert rather than a
// best-effort guess (entry 결정도 편입 기록도 없는 포지션은 exit 정책의 대상이
// 아니며 발견 시 알림을 발송한다 SHALL). The alert is the observation loop's; the
// refusal is here, because this is where a caller would otherwise create the
// state that makes the position managed.
//
// The eligibility test is internal/position.ExitEligible and is spelled nowhere
// else: two sources justify a baseline, and a second copy of "which columns
// count" is how they drift apart.
var ErrPositionNotExitEligible = errors.New(
	"journal: the position carries neither an entry decision nor an adoption record, " +
		"so it is not an exit-policy target")

// ErrExitStateExists means the position already has one. A second exit state is
// a second baseline for one position, which is the thing "포지션당 정책 하나"
// forbids.
var ErrExitStateExists = errors.New("journal: the position already has an exit state")

// ErrExitStateCompleted means the position's exit policy is finished. A finished
// policy judges nothing further; re-opening one would resurrect a protection
// decision the position has already been closed out of.
var ErrExitStateCompleted = errors.New("journal: the exit state is completed")

// ErrExitLifecycleStale refuses a judgement produced by a released or older
// operator-managed generation. It is a quarantine outcome: nothing is retried
// or rebound onto the current generation.
var ErrExitLifecycleStale = errors.New("journal: exit lifecycle generation is stale or released")

// ExitStateSeed is what opening an exit state needs. Everything else in the row
// is derived from it by internal/exitpolicy, so the two systems cannot disagree
// about what t0 means.
type ExitStateSeed struct {
	PositionID string
	// PolicyKind is RATCHET or LADDER. Empty means RATCHET, which is the default
	// design D5 specifies; LADDER is chosen by configuration and never by
	// accident.
	PolicyKind string
	// PolicyID snapshots the exact immutable ladder. Empty LADDER means the
	// pre-v9 default_v1 policy; RATCHET has no policy ID.
	PolicyID string
	// PolicyIdentity is the immutable runtime meaning of PolicyKind/PolicyID.
	// a042 owns the schema columns that will persist it; until then an omitted
	// value is accepted only for the exact pinned pre-a042 compatibility table.
	PolicyIdentity exitpolicy.PolicyIdentity
	// EntryPrice is the position's cost basis, and InitialStop the entry
	// decision's stop price. Their difference is the R denominator, frozen here.
	EntryPrice  string
	InitialStop string
}

// OpenExitState creates the protection state of one position.
//
// It is the moment D5's first correction takes effect: the row is written with
// the entry decision's stop already in `baseline_price`, so there is no instant
// at which the position is open and the ledger holds no stop for it.
func (j *Journal) OpenExitState(ctx context.Context, seed ExitStateSeed) (ExitState, error) {
	id := strings.TrimSpace(seed.PositionID)
	if id == "" {
		return ExitState{}, fmt.Errorf("%w: an exit state needs a position", ErrInvalidRequest)
	}
	kind := strings.TrimSpace(seed.PolicyKind)
	if kind == "" {
		kind = ExitPolicyRatchet
	}
	if kind != ExitPolicyRatchet && kind != ExitPolicyLadder {
		return ExitState{}, fmt.Errorf("%w: %q is neither %s nor %s; a position has exactly one policy",
			ErrInvalidRequest, seed.PolicyKind, ExitPolicyRatchet, ExitPolicyLadder)
	}
	policyID := strings.TrimSpace(seed.PolicyID)
	switch kind {
	case ExitPolicyRatchet:
		if policyID != "" {
			return ExitState{}, fmt.Errorf("%w: RATCHET carries no ladder policy id", ErrInvalidRequest)
		}
	case ExitPolicyLadder:
		if policyID == "" {
			policyID = "default_v1"
		} else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" {
			return ExitState{}, fmt.Errorf("%w: unknown ladder policy %q", ErrInvalidRequest, policyID)
		}
	}
	// The arithmetic is the policy package's, not this file's. The t0 seed is the
	// same for both kinds — the entry stop is the baseline whatever comes after
	// it — so one call covers both.
	open, err := exitpolicy.OpenRatchetState(seed.EntryPrice, seed.InitialStop)
	if err != nil {
		return ExitState{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return ExitState{}, fmt.Errorf("journal: opening the exit state of %s: %w", id, err)
	}
	defer tx.Rollback()

	var decisionID, adoptionID sql.NullString
	var positionGeneration int64
	err = tx.QueryRowContext(ctx,
		`SELECT entry_decision_id, adoption_id, instance_seq FROM positions WHERE id = ?`, id).
		Scan(&decisionID, &adoptionID, &positionGeneration)
	if errors.Is(err, sql.ErrNoRows) {
		return ExitState{}, fmt.Errorf("%w: %s", ErrPositionNotFound, id)
	}
	if err != nil {
		return ExitState{}, fmt.Errorf("journal: reading position %s: %w", id, err)
	}
	if !position.ExitEligible(decisionID.String, adoptionID.String) {
		return ExitState{}, fmt.Errorf("%w: %s", ErrPositionNotExitEligible, id)
	}
	policyIdentity, err := seedPolicyIdentity(kind, policyID,
		strings.TrimSpace(adoptionID.String) != "", seed.PolicyIdentity)
	if err != nil {
		return ExitState{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO exit_states
		  (position_id, policy_kind, policy_id, entry_price, initial_stop, initial_risk,
		   baseline_price, high_water, ratchet_level, active_rung, completed, updated_at,
		   snapshot_status, policy_version, policy_digest, position_generation)
		VALUES (?,?,?,?,?,?,?,?,?,NULL,0,?,?,?,?,?)`,
		id, kind, policyIdentity.ID, seed.EntryPrice, seed.InitialStop, open.InitialRisk,
		open.Baseline, open.HighWater, string(open.Level), now, SnapshotStatusSeed,
		policyIdentity.Version, policyIdentity.Digest, positionGeneration); err != nil {
		if isUniqueViolation(err) {
			return ExitState{}, fmt.Errorf("%w: %s", ErrExitStateExists, id)
		}
		return ExitState{}, fmt.Errorf("journal: opening the exit state of %s: %w", id, err)
	}
	if err := appendExitEventTx(ctx, tx, exitEventRow{
		PositionID: id, HighWater: open.HighWater, BaselineAfter: open.Baseline,
		LevelAfter: string(open.Level), Action: ExitEventOpened, CreatedAt: now,
	}); err != nil {
		return ExitState{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExitState{}, fmt.Errorf("journal: opening the exit state of %s: %w", id, err)
	}
	state, err := j.ExitState(ctx, id)
	state.PolicyIdentity = policyIdentity
	return state, err
}

func seedPolicyIdentity(kind, policyID string, adopted bool,
	claimed exitpolicy.PolicyIdentity) (exitpolicy.PolicyIdentity, error) {
	var (
		expected exitpolicy.PolicyIdentity
		err      error
	)
	if kind == ExitPolicyRatchet {
		expected, err = exitpolicy.LegacyRatchetPolicyIdentity()
	} else {
		expected, err = exitpolicy.LegacyLadderPolicyIdentity(policyID, adopted)
	}
	if err != nil {
		return exitpolicy.PolicyIdentity{}, err
	}
	if strings.TrimSpace(claimed.ID) == "" && strings.TrimSpace(claimed.Version) == "" &&
		strings.TrimSpace(claimed.Digest) == "" {
		return expected, nil
	}
	if err := claimed.Validate(); err != nil {
		return exitpolicy.PolicyIdentity{}, err
	}
	if strings.TrimSpace(claimed.ID) != expected.ID ||
		strings.TrimSpace(claimed.Version) != expected.Version ||
		strings.TrimSpace(claimed.Digest) != expected.Digest {
		return exitpolicy.PolicyIdentity{}, fmt.Errorf(
			"%w: exit state seed claims %s@%s/%s, pinned meaning is %s@%s/%s",
			exitpolicy.ErrPolicyIdentityConflict, claimed.ID, claimed.Version, claimed.Digest,
			expected.ID, expected.Version, expected.Digest)
	}
	return expected, nil
}

// ExitJudgement is one evaluation's outcome, as the row and the history should
// record it.
//
// The fields are the decision-time ones. The two the original calls
// execution-time — the cumulative taken fraction and the completion flag — are
// not here: a fill moves those, inside its own transaction, through the apply
// hook, and an evaluation that could move them would be a five-second loop with
// write access to what has been sold.
type ExitJudgement struct {
	PositionID string
	// LifecycleGeneration is copied from the exact ExitState observed by the
	// engine. Zero is accepted only for legacy in-process callers and resolves to
	// the current generation; production always supplies a positive value.
	LifecycleGeneration int64
	// Snapshot is the exact a041 value consumed by the execution path. A zero
	// value is accepted only by the pre-a042 compatibility path and never
	// overwrites a previously evaluated v10 snapshot.
	Snapshot          exitpolicy.ExitLineSnapshot
	RecoveryPolicy    exitpolicy.RecoveryPolicyDefinition
	ObservationSource string
	ObservedAt        time.Time
	// Provenance binds this write to the one immutable evaluator snapshot.
	// It is a typed a041 seam; a042 persists the corresponding columns.
	Provenance ExitDecisionProvenance
	// ObservedPrice is the price judged, empty when the judgement had none.
	ObservedPrice string
	// HighWater and Baseline are the state after this judgement. Neither may be
	// below what the row holds.
	HighWater string
	Baseline  string
	// RatchetLevel is the level after, under RATCHET. Under LADDER it stays
	// NONE — the schema's CHECK admits no other value and the rung is the ladder's
	// step.
	RatchetLevel string
	// ActiveRung is the rung after, under LADDER. exitpolicy.NoRung writes NULL,
	// which is what "no rung has been reached" is stored as.
	ActiveRung int
	// Proposal is what the judgement asks to be submitted, or nil. Arming it is
	// part of this transaction; submitting it is the caller's next step, and the
	// gap between the two is the crash window the arming exists to close.
	Proposal *ExitProposal
	// ArmSuppressedReason records why an otherwise executable immutable
	// snapshot was persisted without arming its proposal. It is deliberately
	// separate from Snapshot.Suppressed: the evaluator made a proposal, while
	// the execution precondition withheld only its arming.
	ArmSuppressedReason string
}

// ExitProposal is a proposal being armed.
type ExitProposal struct {
	// Provenance must exactly match the judgement that arms this proposal.
	Provenance ExitDecisionProvenance
	// Action is one of exitpolicy's orderable actions.
	Action string
	// Level is the proposal's identity: the ratchet level, or the rung index in
	// decimal text. One column, because a position has one policy (D7).
	Level string
	// IntentID may be empty. The proposal is armed *before* the intent exists —
	// that ordering is what makes a crash between arming and submitting
	// recoverable, and it is why the column carries no foreign key (issues.md,
	// task 0.1).
	IntentID string
}

type ExitArmOutcome string

const (
	ExitArmNotRequested            ExitArmOutcome = "not_requested"
	ExitArmArmed                   ExitArmOutcome = "armed"
	ExitArmSuppressedSavedMonotone ExitArmOutcome = "suppressed_saved_monotone"
	ExitArmSuppressedWorkingOrder  ExitArmOutcome = "suppressed_working_order"
)

// ExitJudgementResult is the only execution authority returned after the
// transaction commits. Callers must not infer submission from their mutable
// request value because recovery may select the previously saved candidate.
type ExitJudgementResult struct {
	EffectiveSource string
	ArmOutcome      ExitArmOutcome
	ArmedProposal   *ExitProposal
}

// ExitDecisionProvenance is the schema-independent handoff from the exit-line
// evaluator to the journal and mutation gateway. It authorises nothing; the
// Guardian decision remains the only order authority.
type ExitDecisionProvenance struct {
	ObservationID string
	SnapshotID    string
	DecisionID    string
	Policy        exitpolicy.PolicyIdentity
}

func (p ExitDecisionProvenance) zero() bool {
	return strings.TrimSpace(p.ObservationID) == "" && strings.TrimSpace(p.SnapshotID) == "" &&
		strings.TrimSpace(p.DecisionID) == "" && strings.TrimSpace(p.Policy.ID) == "" &&
		strings.TrimSpace(p.Policy.Version) == "" && strings.TrimSpace(p.Policy.Digest) == ""
}

func (p ExitDecisionProvenance) validate() error {
	if p.zero() {
		return nil
	}
	if strings.TrimSpace(p.ObservationID) == "" || strings.TrimSpace(p.SnapshotID) == "" ||
		strings.TrimSpace(p.DecisionID) == "" {
		return fmt.Errorf("%w: exit decision provenance requires observation, snapshot, and decision ids",
			ErrInvalidRequest)
	}
	if err := p.Policy.Validate(); err != nil {
		return err
	}
	return nil
}

func sameExitDecisionProvenance(a, b ExitDecisionProvenance) bool {
	return strings.TrimSpace(a.ObservationID) == strings.TrimSpace(b.ObservationID) &&
		strings.TrimSpace(a.SnapshotID) == strings.TrimSpace(b.SnapshotID) &&
		strings.TrimSpace(a.DecisionID) == strings.TrimSpace(b.DecisionID) &&
		strings.TrimSpace(a.Policy.ID) == strings.TrimSpace(b.Policy.ID) &&
		strings.TrimSpace(a.Policy.Version) == strings.TrimSpace(b.Policy.Version) &&
		strings.TrimSpace(a.Policy.Digest) == strings.TrimSpace(b.Policy.Digest)
}

// RecordExitJudgement advances the state, arms the proposal if there is one, and
// appends the history row — in one transaction.
func (j *Journal) RecordExitJudgement(ctx context.Context, judgement ExitJudgement) error {
	_, err := j.RecordExitJudgementResult(ctx, judgement)
	return err
}

// RecordExitJudgementResult returns only the proposal that was durably armed
// by the committed transaction. The request proposal is not execution
// authority because saved-monotone recovery may supersede it.
func (j *Journal) RecordExitJudgementResult(ctx context.Context,
	judgement ExitJudgement) (ExitJudgementResult, error) {
	result := ExitJudgementResult{ArmOutcome: ExitArmNotRequested}
	err := j.recordExitJudgementTx(ctx, judgement, &result)
	return result, err
}

func (j *Journal) recordExitJudgementTx(ctx context.Context, judgement ExitJudgement,
	result *ExitJudgementResult) error {
	id := strings.TrimSpace(judgement.PositionID)
	if id == "" {
		return fmt.Errorf("%w: a judgement needs a position", ErrInvalidRequest)
	}
	if err := judgement.Provenance.validate(); err != nil {
		return err
	}
	if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedReason) != "" {
		return fmt.Errorf("%w: legacy judgement cannot assert an arm-suppression reason", ErrInvalidRequest)
	}
	if judgement.Proposal != nil {
		if err := validateProposal(*judgement.Proposal); err != nil {
			return err
		}
		if err := judgement.Proposal.Provenance.validate(); err != nil {
			return err
		}
		if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() ||
			(!judgement.Provenance.zero() &&
				!sameExitDecisionProvenance(judgement.Provenance, judgement.Proposal.Provenance)) {
			return fmt.Errorf("%w: proposal provenance must match its exit judgement", ErrInvalidRequest)
		}
	}
	var recomputed *StoredExitSnapshot
	if !judgement.Provenance.zero() {
		candidate := StoredExitSnapshot{Line: judgement.Snapshot, RecoveryPolicy: judgement.RecoveryPolicy,
			ObservationSource: strings.TrimSpace(judgement.ObservationSource),
			ObservedAt:        judgement.ObservedAt.UTC().Format(time.RFC3339Nano)}
		if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
		}
		recomputed = &candidate
	}

	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return fmt.Errorf("journal: recording the exit judgement of %s: %w", id, err)
	}
	defer tx.Rollback()

	current, err := scanExitProgress(ctx, tx, id)
	if err != nil {
		return err
	}
	if current.Completed {
		return fmt.Errorf("%w: %s", ErrExitStateCompleted, id)
	}
	expectedLifecycle := judgement.LifecycleGeneration
	if expectedLifecycle == 0 {
		expectedLifecycle = current.LifecycleGeneration
	}
	if expectedLifecycle != current.LifecycleGeneration {
		return fmt.Errorf("%w: judgement generation %d, current exit generation %d",
			ErrExitLifecycleStale, expectedLifecycle, current.LifecycleGeneration)
	}
	var lifecycleGeneration int64
	var lifecycleStatus positionpolicy.Status
	err = tx.QueryRowContext(ctx, `SELECT adoption_generation,status
		FROM position_policy_lifecycles WHERE position_id=?
		ORDER BY adoption_generation DESC LIMIT 1`, id).Scan(&lifecycleGeneration, &lifecycleStatus)
	if errors.Is(err, sql.ErrNoRows) {
		lifecycleGeneration, lifecycleStatus = 1, positionpolicy.StatusManaged
	} else if err != nil {
		return fmt.Errorf("journal: reading exit lifecycle of %s: %w", id, err)
	}
	if lifecycleStatus != positionpolicy.StatusManaged || lifecycleGeneration != expectedLifecycle {
		return fmt.Errorf("%w: lifecycle generation %d is %s; judgement holds generation %d",
			ErrExitLifecycleStale, lifecycleGeneration, lifecycleStatus, expectedLifecycle)
	}
	if recomputed != nil && recomputed.Line.PositionGeneration != current.PositionGeneration {
		return fmt.Errorf("%w: snapshot generation %d does not match position generation %d",
			ErrInvalidRequest, recomputed.Line.PositionGeneration, current.PositionGeneration)
	}
	if recomputed != nil {
		var duplicate int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM exit_events WHERE decision_id=?`,
			recomputed.Line.DecisionID).Scan(&duplicate)
		if err == nil {
			// The deterministic decision was already committed by another
			// observer. Returning ErrProposalPending is intentional: the engine
			// treats it as a conservative no-submit outcome, whereas nil would
			// make both observers submit the same already-armed proposal.
			return ErrProposalPending
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("journal: checking duplicate exit decision: %w", err)
		}
	}
	// Legacy callers have no complete candidates to compare, so retain the old
	// scalar monotonicity guards. Snapshot callers are decided below as one
	// coherent tuple; checking their individual scalars first would reject the
	// very stale-but-valid saved-snapshot recovery a042 requires.
	if recomputed == nil {
		if err := notBelow("high water", id, judgement.HighWater, current.HighWater); err != nil {
			return err
		}
		if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err != nil {
			return err
		}
	}
	level := strings.TrimSpace(judgement.RatchetLevel)
	if level == "" {
		level = current.RatchetLevel
	}

	effectiveSource := ""
	var saved, effective *StoredExitSnapshot
	if current.Effective != nil {
		copy := *current.Effective
		saved = &copy
	}
	if recomputed != nil {
		var savedLine *exitpolicy.ExitLineSnapshot
		if saved != nil {
			line := saved.Line
			savedLine = &line
		}
		selected, source, selectErr := exitpolicy.SelectRecoverySnapshot(savedLine, recomputed.Line)
		if selectErr != nil {
			if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGeneration,
				QuarantineReasonAmbiguousRecovery, selectErr.Error(), now); qerr != nil {
				return qerr
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			return fmt.Errorf("%w: %v", ErrExitSnapshotQuarantined, selectErr)
		}
		// The selection succeeded. If a superseded selector had this generation
		// quarantined, that row is what stopped it being judged at all, and it is
		// closed here — in the same transaction as the judgement it earned, so the
		// two cannot come apart (a084).
		if err := releaseReJudgedQuarantineTx(ctx, tx, id,
			recomputed.Line.PositionGeneration, now); err != nil {
			return err
		}
		if source == exitpolicy.RecoverySavedMonotone {
			effective = saved
			effectiveSource = EffectiveSourceSaved
		} else {
			copy := *recomputed
			copy.Line = selected
			effective = &copy
			effectiveSource = EffectiveSourceRecomputed
		}
		judgement.HighWater = effective.Line.HighWater
		judgement.Baseline = effective.Line.CurrentProtection
		judgement.RatchetLevel = string(effective.Line.RatchetLevel)
		judgement.ActiveRung = effective.Line.ActiveRung
		if effectiveSource == EffectiveSourceSaved {
			judgement.Proposal = nil
			judgement.ArmSuppressedReason = ""
		}
		level = strings.TrimSpace(judgement.RatchetLevel)
	}

	updateSQL := `
		UPDATE exit_states
		   SET baseline_price = ?, high_water = ?, ratchet_level = ?, active_rung = ?, updated_at = ?
		 WHERE position_id = ?`
	args := []any{judgement.Baseline, judgement.HighWater, level,
		nullableRung(judgement.ActiveRung), now, id}
	if effective != nil {
		raw, err := encodeStoredSnapshot(*effective)
		if err != nil {
			return err
		}
		line := effective.Line
		updateSQL = `UPDATE exit_states SET baseline_price=?,high_water=?,ratchet_level=?,active_rung=?,updated_at=?,
			snapshot_status=?,policy_id=?,policy_version=?,policy_digest=?,snapshot_id=?,decision_id=?,observation_id=?,
			position_generation=?,next_target=?,next_protection=?,last_observation_source=?,last_observed_at=?,
			snapshot_action=?,snapshot_ratio=?,projected_quantity=?,state_only=?,suppressed_reason=?,effective_snapshot_json=?
			WHERE position_id=?`
		args = []any{line.CurrentProtection, line.HighWater, string(line.RatchetLevel), nullableRung(line.ActiveRung), now,
			SnapshotStatusEvaluated, line.Policy.ID, line.Policy.Version, line.Policy.Digest, line.SnapshotID,
			line.DecisionID, line.ObservationID, line.PositionGeneration, nullableString(line.NextTarget),
			nullableString(line.NextProtection), effective.ObservationSource, effective.ObservedAt,
			nullableString(string(line.Action)), nullableString(line.Ratio), line.ProjectedQuantity,
			boolInt(line.StateOnly), nullableString(line.Suppressed), raw, id}
	}
	if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil {
		return fmt.Errorf("journal: recording the exit judgement of %s: %w", id, err)
	}
	if err := j.runExitWriteHook("after_state"); err != nil {
		return err
	}

	action := ""
	intentID := ""
	if judgement.Proposal != nil {
		action, intentID = judgement.Proposal.Action, judgement.Proposal.IntentID
		if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil {
			return err
		}
		if err := j.runExitWriteHook("after_arm"); err != nil {
			return err
		}
	}
	event := exitEventRow{
		PositionID: id, ObservedPrice: judgement.ObservedPrice,
		HighWater: judgement.HighWater, BaselineAfter: judgement.Baseline,
		LevelAfter: levelAfter(level, judgement.ActiveRung), Action: action,
		ProposedIntentID: intentID, CreatedAt: now,
		ArmSuppressedReason: judgement.ArmSuppressedReason,
	}
	if recomputed != nil && effective != nil {
		event.Evaluation = evaluationForEvent(saved, recomputed, effective, effectiveSource)
	}
	if err := appendExitEventTx(ctx, tx, event); err != nil {
		return err
	}
	if err := j.runExitWriteHook("after_event"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: recording the exit judgement of %s: %w", id, err)
	}
	result.EffectiveSource = effectiveSource
	switch {
	case effectiveSource == EffectiveSourceSaved:
		result.ArmOutcome = ExitArmSuppressedSavedMonotone
	case judgement.Proposal != nil:
		proposal := *judgement.Proposal
		result.ArmOutcome = ExitArmArmed
		result.ArmedProposal = &proposal
	case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder,
		judgement.ArmSuppressedReason == ArmSuppressedReJudge:
		result.ArmOutcome = ExitArmSuppressedWorkingOrder
	}
	return nil
}

// exitProgress is the part of the row this file may read: everything except the
// four the apply point owns.
type exitProgress struct {
	PolicyKind          string
	PositionGeneration  int64
	LifecycleGeneration int64
	Baseline            string
	HighWater           string
	RatchetLevel        string
	ActiveRung          int
	Completed           bool
	Effective           *StoredExitSnapshot
}

func scanExitProgress(ctx context.Context, tx *sql.Tx, positionID string) (exitProgress, error) {
	var (
		out       exitProgress
		rung      sql.NullInt64
		done      int
		effective sql.NullString
	)
	err := tx.QueryRowContext(ctx, `
		SELECT e.policy_kind, p.instance_seq, coalesce(e.lifecycle_generation,1),
		       e.baseline_price, e.high_water, e.ratchet_level,
		       e.active_rung, e.completed, e.effective_snapshot_json
		  FROM exit_states e JOIN positions p ON p.id=e.position_id WHERE e.position_id = ?`, positionID).
		Scan(&out.PolicyKind, &out.PositionGeneration, &out.LifecycleGeneration, &out.Baseline, &out.HighWater,
			&out.RatchetLevel, &rung, &done, &effective)
	if errors.Is(err, sql.ErrNoRows) {
		return exitProgress{}, fmt.Errorf("%w: position %s", ErrExitStateNotFound, positionID)
	}
	if err != nil {
		return exitProgress{}, fmt.Errorf("journal: reading the exit state of %s: %w", positionID, err)
	}
	out.ActiveRung = exitpolicy.NoRung
	if rung.Valid {
		out.ActiveRung = int(rung.Int64)
	}
	out.Completed = done != 0
	if effective.Valid {
		stored, err := decodeStoredSnapshot(effective.String)
		if err != nil {
			return exitProgress{}, fmt.Errorf("%w: %v", ErrExitSnapshotCorrupt, err)
		}
		out.Effective = &stored
	}
	return out, nil
}

// notBelow refuses a protective value that would descend.
func notBelow(what, positionID, next, current string) error {
	cmp, err := riskcalc.CompareDecimal(next, current)
	if err != nil {
		return fmt.Errorf("%w: %s for position %s: %v", ErrInvalidRequest, what, positionID, err)
	}
	if cmp < 0 {
		return fmt.Errorf("%w: the %s of position %s would move from %s down to %s; it is monotone",
			ErrInvalidRequest, what, positionID, current, next)
	}
	return nil
}

func validateProposal(p ExitProposal) error {
	action := exitpolicy.Action(strings.TrimSpace(p.Action))
	if !action.Orderable() {
		return fmt.Errorf(
			"%w: %q is not an orderable exit action; a state-only promotion has no fill to resolve it",
			ErrInvalidRequest, p.Action)
	}
	if strings.TrimSpace(p.Level) == "" {
		return fmt.Errorf("%w: a proposal needs the level it belongs to, or it cannot be de-duplicated",
			ErrInvalidRequest)
	}
	return nil
}

// nullableRung stores exitpolicy.NoRung as NULL. "No rung has been reached" is
// the absence of a rung and not the rung before the first one, and a stored −1
// would sort below every real index in any future query that orders by it.
func nullableRung(rung int) any {
	if rung < 0 {
		return nil
	}
	return rung
}

// levelAfter is what the history row records as the level: the ratchet level
// under RATCHET, the rung index under LADDER. `exit_events.level_after` is one
// column for the same reason the pending column is (D7).
func levelAfter(ratchetLevel string, rung int) string {
	if rung >= 0 {
		return fmt.Sprintf("%d", rung)
	}
	return ratchetLevel
}

// --- the judgement history ----------------------------------------------------

// Actions recorded in `exit_events` that are not proposals. They are the
// lifecycle of the state itself, and they are in the same log as the proposals
// because the provenance join is "what did this position's protection do", not
// "what did it order".
const (
	ExitEventOpened            = "OPENED"
	ExitEventReadopted         = "READOPTED"
	ExitEventProposalRefused   = "PROPOSAL_REFUSED"
	ExitEventProposalCancelled = "PROPOSAL_CANCELLED"
	ExitEventProposalFilled    = "PROPOSAL_FILLED"
	ExitEventCompleted         = "COMPLETED"
	// ExitEventAdjustmentClosed is a position that reached zero through an
	// adjustment rather than through a fill: somebody sold it outside the engine
	// (change adopt-external-positions, design A7).
	//
	// It is a different action from COMPLETED on purpose. COMPLETED is the exit
	// policy finishing its own work — the engine sold the position and the closing
	// fill resolved the state inside its transaction. This one records that the
	// position went to zero somewhere the engine cannot see, which is why the same
	// event is also the marker for "no trade outcome was frozen and none should
	// be": there is no sell leg to price, and inventing one from the adjustment
	// would report the whole position as a loss.
	ExitEventAdjustmentClosed = "ADJUSTMENT_CLOSED"
)

type exitEventRow struct {
	PositionID          string
	ObservedPrice       string
	HighWater           string
	BaselineAfter       string
	LevelAfter          string
	Action              string
	ProposedIntentID    string
	CreatedAt           string
	ArmSuppressedReason string
	Evaluation          *ExitEvaluation
}

func appendExitEventTx(ctx context.Context, tx *sql.Tx, row exitEventRow) error {
	var lifecycleGeneration int64
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(lifecycle_generation,1)
		FROM exit_states WHERE position_id=?`, row.PositionID).Scan(&lifecycleGeneration); err != nil {
		return fmt.Errorf("journal: binding exit event to lifecycle of %s: %w", row.PositionID, err)
	}
	var savedJSON, recomputedJSON, effectiveJSON any
	var generation, policyID, policyVersion, policyDigest, snapshotID, decisionID, observationID any
	var nextTarget, nextProtection, source, observedAt, projected, ratio, stateOnly, suppressed, effectiveSource any
	if row.Evaluation != nil && row.Evaluation.Effective.Snapshot != nil && row.Evaluation.Recomputed.Snapshot != nil {
		encode := func(view ExitSnapshotView) (any, error) {
			if view.Snapshot == nil {
				return nil, nil
			}
			return encodeStoredSnapshot(*view.Snapshot)
		}
		var err error
		if savedJSON, err = encode(row.Evaluation.Saved); err != nil {
			return err
		}
		if recomputedJSON, err = encode(row.Evaluation.Recomputed); err != nil {
			return err
		}
		if effectiveJSON, err = encode(row.Evaluation.Effective); err != nil {
			return err
		}
		// The event identity belongs to the newly evaluated candidate. The
		// selected effective snapshot may be the previously saved candidate; using
		// its decision id here would collide on every legitimate stale recovery
		// and erase which observation triggered this event.
		e := row.Evaluation.Recomputed.Snapshot
		line := e.Line
		generation, policyID, policyVersion, policyDigest = line.PositionGeneration, line.Policy.ID, line.Policy.Version, line.Policy.Digest
		snapshotID, decisionID, observationID = line.SnapshotID, line.DecisionID, line.ObservationID
		nextTarget, nextProtection, source, observedAt = nullableString(line.NextTarget), nullableString(line.NextProtection), e.ObservationSource, e.ObservedAt
		projected, ratio, stateOnly, suppressed = line.ProjectedQuantity, nullableString(line.Ratio), boolInt(line.StateOnly), nullableString(line.Suppressed)
		effectiveSource = row.Evaluation.EffectiveSource
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after,
		   action, proposed_intent_id, created_at, position_generation, policy_id,
		   policy_version, policy_digest, snapshot_id, decision_id, observation_id,
		   next_target, next_protection, observation_source, observed_at, projected_quantity,
		   proposal_ratio, state_only, suppressed_reason, saved_snapshot_json,
		   recomputed_snapshot_json, effective_snapshot_json, effective_source,
		   arm_suppressed_reason,lifecycle_generation)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.PositionID, nullableString(row.ObservedPrice), nullableString(row.HighWater),
		nullableString(row.BaselineAfter), nullableString(row.LevelAfter),
		nullableString(row.Action), nullableString(row.ProposedIntentID), row.CreatedAt,
		generation, policyID, policyVersion, policyDigest, snapshotID, decisionID, observationID,
		nextTarget, nextProtection, source, observedAt, projected, ratio, stateOnly, suppressed,
		savedJSON, recomputedJSON, effectiveJSON, effectiveSource,
		nullableString(row.ArmSuppressedReason), lifecycleGeneration)
	if err != nil {
		return fmt.Errorf("journal: recording the exit judgement of %s: %w", row.PositionID, err)
	}
	return nil
}

// ExitEvent is one recorded judgement.
type ExitEvent struct {
	ID                  int64
	PositionID          string
	ObservedPrice       string
	HighWater           string
	BaselineAfter       string
	LevelAfter          string
	Action              string
	ProposedIntentID    string
	CreatedAt           string
	ArmSuppressedReason string
	LifecycleGeneration int64
	Evaluation          ExitEvaluation
}

// ExitEvents returns one position's judgement history, oldest first.
func (j *Journal) ExitEvents(ctx context.Context, positionID string) ([]ExitEvent, error) {
	id := strings.TrimSpace(positionID)
	rows, err := j.db.QueryContext(ctx, `
		SELECT id, position_id, coalesce(observed_price,''), coalesce(high_water,''),
		       coalesce(baseline_after,''), coalesce(level_after,''), coalesce(action,''),
		       coalesce(proposed_intent_id,''), created_at, position_generation, policy_id,
		       policy_version, policy_digest, snapshot_id, decision_id, observation_id,
		       next_target, next_protection, observation_source, observed_at, projected_quantity,
		       proposal_ratio, state_only, suppressed_reason, saved_snapshot_json,
		       recomputed_snapshot_json, effective_snapshot_json, effective_source,
		       arm_suppressed_reason,coalesce(lifecycle_generation,1)
		  FROM exit_events WHERE position_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("journal: reading the exit history of %s: %w", id, err)
	}
	defer rows.Close()

	var out []ExitEvent
	for rows.Next() {
		var e ExitEvent
		if err := scanExitEvent(rows, &e); err != nil {
			return nil, fmt.Errorf("journal: reading the exit history of %s: %w", id, err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading the exit history of %s: %w", id, err)
	}
	return out, nil
}

// --- what the fill transaction needs to know ----------------------------------

// exitFillContext is everything the apply hook needs about a fill that this file
// can read for it: which position it lands on, which intent asked for it, and
// what the projection now holds.
type exitFillContext struct {
	PositionID string
	// IntentID is the intent behind the broker order, empty when no local intent
	// claims it (an external order).
	IntentID string
	// Side is the intent's direction. A take-profit is a SELL; a BUY is a
	// scale-in and moves no taken fraction.
	Side string
	// QuantityAfter is what the projection holds after this fill, and
	// PositionState its state. Both are read after the Project hook has run,
	// which is why the apply contract runs Project first.
	QuantityAfter string
	PositionState string
}

// exitContextForFill resolves a fill onto the position and intent it belongs to.
//
// The bool is false when there is nothing for the exit policy to do: an order no
// local intent claims, or a symbol the projection holds no instance of. Neither
// is an error — an external order is a fact about the account, not a fault.
func exitContextForFill(ctx context.Context, tx *ApplyTx, fill AppliedFill) (exitFillContext, bool, error) {
	origin, found, err := resolveFillOrigin(ctx, tx, fill)
	if err != nil {
		return exitFillContext{}, false, err
	}
	if !found {
		return exitFillContext{}, false, nil
	}
	out := exitFillContext{
		IntentID: origin.IntentID,
		Side:     origin.Side,
	}

	instance, err := tx.Query(ctx, `
		SELECT id, quantity, state FROM positions
		 WHERE account_ref = ? AND market = ? AND symbol = ?
		 ORDER BY instance_seq DESC LIMIT 1`, origin.AccountRef, origin.Market, origin.Symbol)
	if err != nil {
		return exitFillContext{}, false, fmt.Errorf(
			"journal: reading the position of %s/%s: %w", origin.AccountRef, origin.Symbol, err)
	}
	defer instance.Close()
	if !instance.Next() {
		return exitFillContext{}, false, instance.Err()
	}
	if err := instance.Scan(&out.PositionID, &out.QuantityAfter, &out.PositionState); err != nil {
		return exitFillContext{}, false, fmt.Errorf(
			"journal: reading the position of %s/%s: %w", origin.AccountRef, origin.Symbol, err)
	}
	return out, true, nil
}

// markExitCompletedTx closes the exit state of a position that has closed.
//
// `completed` is not one of the guarded four, so it is written here — but it is
// written from inside the fill transaction all the same, because the fact it
// records is "the position reached CLOSED", which only that transaction knows.
func markExitCompletedTx(ctx context.Context, tx *ApplyTx, positionID string) error {
	if _, err := tx.Exec(ctx,
		`UPDATE exit_states SET completed = 1, updated_at = ? WHERE position_id = ?`,
		tx.Now(), positionID); err != nil {
		return fmt.Errorf("journal: completing the exit state of %s: %w", positionID, err)
	}
	return appendExitEventFromHook(ctx, tx, exitEventRow{
		PositionID: positionID, Action: ExitEventCompleted, CreatedAt: tx.Now(),
	})
}

// appendExitEventFromHook is appendExitEventTx against the hook's handle. The two
// exist separately because *ApplyTx deliberately does not expose its *sql.Tx —
// the whole point of the handle is that a hook cannot reach past it.
func appendExitEventFromHook(ctx context.Context, tx *ApplyTx, row exitEventRow) error {
	result, err := tx.Exec(ctx, `
		INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after,
		   action, proposed_intent_id, created_at, lifecycle_generation)
		SELECT ?,?,?,?,?,?,?,?,coalesce(lifecycle_generation,1)
		  FROM exit_states WHERE position_id=?`,
		row.PositionID, nullableString(row.ObservedPrice), nullableString(row.HighWater),
		nullableString(row.BaselineAfter), nullableString(row.LevelAfter),
		nullableString(row.Action), nullableString(row.ProposedIntentID), row.CreatedAt, row.PositionID)
	if err != nil {
		return fmt.Errorf("journal: recording the exit judgement of %s: %w", row.PositionID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: confirming the exit event of %s: %w", row.PositionID, err)
	}
	if affected != 1 {
		return fmt.Errorf("%w: position %s", ErrExitStateNotFound, row.PositionID)
	}
	return nil
}

// rollBackRungTx undoes a rung promotion whose proposal never became an order.
//
// It is the ladder half of "거부·취소 시 해당 레벨은 재발의 가능해진다". The rung
// index is what the ladder de-duplicates on, so a rung that stayed promoted after
// its proposal was refused could never be proposed again. The baseline is
// deliberately *not* rolled back with it: the lock that rung set is protection
// already granted, and §0.9 permits no movement in the other direction. The next
// evaluation therefore re-proposes the rung against a baseline that already
// reflects it, which is the conservative combination.
func rollBackRungTx(ctx context.Context, tx *sql.Tx, positionID string, toRung int, now string) error {
	if _, err := tx.ExecContext(ctx,
		`UPDATE exit_states SET active_rung = ?, updated_at = ? WHERE position_id = ?`,
		nullableRung(toRung), now, positionID); err != nil {
		return fmt.Errorf("journal: rolling the rung of %s back: %w", positionID, err)
	}
	return nil
}

// isUniqueViolation reports a primary-key or unique-index collision, which the
// callers here turn into a named error rather than a driver message.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "UNIQUE")
}
