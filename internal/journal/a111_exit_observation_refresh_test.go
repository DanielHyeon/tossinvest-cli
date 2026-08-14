package journal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

var a111ObservedAt = time.Date(2026, 8, 14, 4, 0, 0, 0, time.UTC)

func a111EvaluatedPosition(t *testing.T, source string, observedAt time.Time) (*Journal, ExitState) {
	t.Helper()
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	line, recovery := ratchetSnapshotForState(t, seed, "a111-observation-1", "70500", "70500", "68000")
	judgement := judgementForSnapshot(line, recovery)
	judgement.LifecycleGeneration = seed.LifecycleGeneration
	judgement.ObservationSource = source
	judgement.ObservedAt = observedAt
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	return j, exitStateOf(t, j, seed.PositionID)
}

func a111RefreshForState(t *testing.T, state ExitState, observationID, price, source string,
	observedAt time.Time,
) ExitObservationRefresh {
	t.Helper()
	line, recovery := ratchetSnapshotForState(t, state, observationID, price, state.HighWater, state.Baseline)
	if line.Changed || line.Orderable {
		t.Fatalf("refresh fixture changed the operational line: %+v", line)
	}
	return ExitObservationRefresh{
		PositionID: state.PositionID, LifecycleGeneration: state.LifecycleGeneration,
		Snapshot: line, RecoveryPolicy: recovery, ObservationSource: source, ObservedAt: observedAt,
		Provenance: ExitDecisionProvenance{
			ObservationID: line.ObservationID, SnapshotID: line.SnapshotID,
			DecisionID: line.DecisionID, Policy: line.Policy,
		},
	}
}

func a111RefreshForSnapshot(state ExitState, line exitpolicy.ExitLineSnapshot,
	recovery exitpolicy.RecoveryPolicyDefinition, source string, observedAt time.Time,
) ExitObservationRefresh {
	return ExitObservationRefresh{
		PositionID: state.PositionID, LifecycleGeneration: state.LifecycleGeneration,
		Snapshot: line, RecoveryPolicy: recovery, ObservationSource: source, ObservedAt: observedAt,
		Provenance: ExitDecisionProvenance{
			ObservationID: line.ObservationID, SnapshotID: line.SnapshotID,
			DecisionID: line.DecisionID, Policy: line.Policy,
		},
	}
}

func a111RatchetCandidate(t *testing.T, state ExitState, observationID, quantity string,
	mutate func(*exitpolicy.RatchetInput),
) (exitpolicy.ExitLineSnapshot, exitpolicy.RecoveryPolicyDefinition) {
	t.Helper()
	evaluation := exitpolicy.RatchetSnapshotInput{
		Context: exitpolicy.SnapshotContext{
			PositionID: state.PositionID, PositionGeneration: state.PositionGeneration,
			ObservationID: observationID, RemainingQuantity: quantity,
		},
		Input: exitpolicy.RatchetInput{
			Entry: "70000", InitialStop: "68000", ObservedPrice: "70400",
			HighWater: state.HighWater, Baseline: state.Baseline, RealBreakeven: "70010",
			TakenRatioTotal: state.TakenRatioTotal, Level: exitpolicy.Level(state.RatchetLevel),
		},
	}
	if mutate != nil {
		mutate(&evaluation.Input)
	}
	line, err := exitpolicy.EvaluateRatchetSnapshot(evaluation)
	if err != nil {
		t.Fatalf("EvaluateRatchetSnapshot(%s): %v", observationID, err)
	}
	line = line.ChangedFromState(evaluation.Input.HighWater, evaluation.Input.Baseline,
		evaluation.Input.Level, exitpolicy.NoRung)
	recovery := exitpolicy.NewRatchetRecoveryPolicy(evaluation)
	if err := exitpolicy.ValidateRecoveryDerivation(line, recovery); err != nil {
		t.Fatalf("candidate %s is not an internally valid evaluator tuple: %v", observationID, err)
	}
	return line, recovery
}

type a111PersistedExitTuple struct {
	Baseline, HighWater, RatchetLevel string
	ActiveRung                        int
	UpdatedAt, Status                 string
	PolicyID, PolicyVersion           string
	PolicyDigest                      string
	SnapshotID, DecisionID            string
	ObservationID                     string
	PositionGeneration                int64
	NextTarget, NextProtection        string
	ObservationSource, ObservedAt     string
	Action, Ratio, ProjectedQuantity  string
	StateOnly                         int
	Suppressed, EffectiveJSON         string
	PendingAction, PendingLevel       string
	PendingIntentID                   string
}

func a111ReadExitTuple(t *testing.T, j *Journal, positionID string) a111PersistedExitTuple {
	t.Helper()
	var got a111PersistedExitTuple
	err := j.db.QueryRow(`SELECT baseline_price,high_water,ratchet_level,coalesce(active_rung,-1),
		updated_at,coalesce(snapshot_status,''),coalesce(policy_id,''),coalesce(policy_version,''),
		coalesce(policy_digest,''),coalesce(snapshot_id,''),coalesce(decision_id,''),
		coalesce(observation_id,''),coalesce(position_generation,-1),coalesce(next_target,''),
		coalesce(next_protection,''),coalesce(last_observation_source,''),coalesce(last_observed_at,''),
		coalesce(snapshot_action,''),coalesce(snapshot_ratio,''),coalesce(projected_quantity,''),
		coalesce(state_only,0),coalesce(suppressed_reason,''),coalesce(effective_snapshot_json,''),
		coalesce(pending_action,''),coalesce(pending_level,''),coalesce(pending_intent_id,'')
		FROM exit_states WHERE position_id=?`, positionID).Scan(
		&got.Baseline, &got.HighWater, &got.RatchetLevel, &got.ActiveRung,
		&got.UpdatedAt, &got.Status, &got.PolicyID, &got.PolicyVersion, &got.PolicyDigest,
		&got.SnapshotID, &got.DecisionID, &got.ObservationID, &got.PositionGeneration,
		&got.NextTarget, &got.NextProtection, &got.ObservationSource, &got.ObservedAt,
		&got.Action, &got.Ratio, &got.ProjectedQuantity, &got.StateOnly,
		&got.Suppressed, &got.EffectiveJSON, &got.PendingAction, &got.PendingLevel,
		&got.PendingIntentID)
	if err != nil {
		t.Fatalf("read persisted exit tuple: %v", err)
	}
	return got
}

func a111AssertTupleMatchesRefresh(t *testing.T, got a111PersistedExitTuple,
	want ExitObservationRefresh,
) {
	t.Helper()
	line := want.Snapshot
	stateOnly := 0
	if line.StateOnly {
		stateOnly = 1
	}
	checks := map[string][2]any{
		"baseline":            {got.Baseline, line.CurrentProtection},
		"high_water":          {got.HighWater, line.HighWater},
		"ratchet_level":       {got.RatchetLevel, string(line.RatchetLevel)},
		"active_rung":         {got.ActiveRung, line.ActiveRung},
		"snapshot_status":     {got.Status, SnapshotStatusEvaluated},
		"policy_id":           {got.PolicyID, line.Policy.ID},
		"policy_version":      {got.PolicyVersion, line.Policy.Version},
		"policy_digest":       {got.PolicyDigest, line.Policy.Digest},
		"snapshot_id":         {got.SnapshotID, line.SnapshotID},
		"decision_id":         {got.DecisionID, line.DecisionID},
		"observation_id":      {got.ObservationID, line.ObservationID},
		"position_generation": {got.PositionGeneration, line.PositionGeneration},
		"next_target":         {got.NextTarget, line.NextTarget},
		"next_protection":     {got.NextProtection, line.NextProtection},
		"observation_source":  {got.ObservationSource, want.ObservationSource},
		"observed_at":         {got.ObservedAt, want.ObservedAt.UTC().Format(time.RFC3339Nano)},
		"action":              {got.Action, string(line.Action)},
		"ratio":               {got.Ratio, line.Ratio},
		"projected_quantity":  {got.ProjectedQuantity, line.ProjectedQuantity},
		"state_only":          {got.StateOnly, stateOnly},
		"suppressed":          {got.Suppressed, line.Suppressed},
	}
	for field, values := range checks {
		if values[0] != values[1] {
			t.Errorf("%s = %v, want %v", field, values[0], values[1])
		}
	}
	if got.EffectiveJSON == "" {
		t.Error("effective snapshot JSON was not replaced with the refreshed tuple")
	}
}

func a111EventCount(t *testing.T, j *Journal, positionID string) int {
	t.Helper()
	events, err := j.ExitEvents(context.Background(), positionID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	return len(events)
}

func a111AssertRefreshNoWrite(t *testing.T, j *Journal, positionID string,
	request ExitObservationRefresh, wantErr error,
) {
	t.Helper()
	before := a111ReadExitTuple(t, j, positionID)
	events := a111EventCount(t, j, positionID)
	j.clk.(*clock.Fake).Advance(time.Second)
	refreshWrites := 0
	j.exitWriteHook = func(stage string) error {
		if stage == "after_refresh_state" {
			refreshWrites++
		}
		return nil
	}
	defer func() { j.exitWriteHook = nil }()
	err := j.RefreshExitObservation(context.Background(), request)
	if wantErr == nil && err != nil {
		t.Fatalf("no-op refresh error = %v", err)
	}
	if wantErr != nil && !errors.Is(err, wantErr) {
		t.Fatalf("refresh error = %v, want %v", err, wantErr)
	}
	if refreshWrites != 0 {
		t.Fatalf("no-write refresh performed %d refresh state writes", refreshWrites)
	}
	if after := a111ReadExitTuple(t, j, positionID); after != before {
		t.Fatalf("no-write refresh changed complete tuple or updated_at:\nbefore=%+v\nafter=%+v", before, after)
	}
	if got := a111EventCount(t, j, positionID); got != events {
		t.Fatalf("no-write refresh changed events: %d -> %d", events, got)
	}
}

func TestA111RefreshAtomicallyReplacesTheCompleteTupleWithoutAnEvent(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	beforeEvents := a111EventCount(t, j, state.PositionID)
	request := a111RefreshForState(t, state, "a111-observation-2", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))

	if err := j.RefreshExitObservation(context.Background(), request); err != nil {
		t.Fatalf("RefreshExitObservation: %v", err)
	}

	got := a111ReadExitTuple(t, j, state.PositionID)
	a111AssertTupleMatchesRefresh(t, got, request)
	reloaded := exitStateOf(t, j, state.PositionID)
	if reloaded.Snapshot.Snapshot == nil || reloaded.Snapshot.Snapshot.Line != request.Snapshot ||
		reloaded.Snapshot.Snapshot.ObservationSource != request.ObservationSource ||
		reloaded.Snapshot.Snapshot.ObservedAt != request.ObservedAt.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("decoded refreshed tuple = %+v", reloaded.Snapshot)
	}
	if after := a111EventCount(t, j, state.PositionID); after != beforeEvents {
		t.Fatalf("refresh appended an exit event: before=%d after=%d", beforeEvents, after)
	}
	if reloaded.Pending() {
		t.Fatalf("refresh armed a proposal: %+v", reloaded)
	}
}

func TestA111RefreshRejectsInvalidRequestsBeforeStartingAWrite(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ExitObservationRefresh)
		want   error
	}{
		{"blank_position", func(r *ExitObservationRefresh) { r.PositionID = " \t" }, ErrInvalidRequest},
		{"zero_provenance", func(r *ExitObservationRefresh) { r.Provenance = ExitDecisionProvenance{} }, ErrInvalidRequest},
		{"malformed_provenance", func(r *ExitObservationRefresh) { r.Provenance.DecisionID = "" }, ErrInvalidRequest},
		{"mismatched_provenance", func(r *ExitObservationRefresh) { r.Provenance.SnapshotID = "els_mismatch" }, ErrInvalidRequest},
		{"zero_observed_at", func(r *ExitObservationRefresh) { r.ObservedAt = time.Time{} }, ErrInvalidRequest},
		{"malformed_source", func(r *ExitObservationRefresh) { r.ObservationSource = "cycle:not-a-number" }, ErrExitObservationConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
			request := a111RefreshForState(t, state, "a111-invalid-"+tc.name, "70400",
				"quote_fetched_at", a111ObservedAt.Add(time.Second))
			tc.mutate(&request)
			a111AssertRefreshNoWrite(t, j, state.PositionID, request, tc.want)
		})
	}
}

func TestA111RefreshIsIdempotentAndDurableAcrossRestart(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	request := a111RefreshForState(t, state, "a111-observation-replay", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))
	if err := j.RefreshExitObservation(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	committed := a111ReadExitTuple(t, j, state.PositionID)
	events := a111EventCount(t, j, state.PositionID)
	j.clk.(*clock.Fake).Advance(time.Second)
	refreshWrites := 0
	j.exitWriteHook = func(stage string) error {
		if stage == "after_refresh_state" {
			refreshWrites++
		}
		return nil
	}
	if err := j.RefreshExitObservation(context.Background(), request); err != nil {
		t.Fatalf("exact replay: %v", err)
	}
	j.exitWriteHook = nil
	if refreshWrites != 0 {
		t.Fatalf("exact replay performed %d refresh state writes after the journal clock advanced", refreshWrites)
	}
	if replayed := a111ReadExitTuple(t, j, state.PositionID); replayed != committed {
		t.Fatalf("exact replay churned state or updated_at:\nfirst=%+v\nagain=%+v", committed, replayed)
	}
	if got := a111EventCount(t, j, state.PositionID); got != events {
		t.Fatalf("exact replay appended event: %d -> %d", events, got)
	}

	path := j.Path()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := openTestJournalAt(t, path)
	got := exitStateOf(t, restarted, state.PositionID)
	if got.Snapshot.Snapshot == nil || got.Snapshot.Snapshot.Line != request.Snapshot ||
		got.Snapshot.Snapshot.ObservationSource != request.ObservationSource {
		t.Fatalf("restart lost refreshed tuple: %+v", got.Snapshot)
	}
}

func TestA111RefreshRejectsARealReleasedLifecycleAndPreservesItsTuple(t *testing.T) {
	j := exitFixture(t)
	positionID := adoptedHolding(t, j, "10", "55000", "70000", "68000")
	seed := exitStateOf(t, j, positionID)
	line, recovery := ratchetSnapshotForState(t, seed, "a111-released-evaluated", "70500", "70500", seed.Baseline)
	judgement := judgementForSnapshot(line, recovery)
	judgement.LifecycleGeneration = seed.LifecycleGeneration
	judgement.ObservationSource = "quote_fetched_at"
	judgement.ObservedAt = a111ObservedAt
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	evaluated := exitStateOf(t, j, positionID)
	request := a111RefreshForState(t, evaluated, "a111-after-release", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))

	released, err := j.ApplyPositionPolicy(context.Background(), positionpolicy.Request{
		PositionID: positionID, ExpectedGeneration: 1, ExpectedVersion: 0,
		Action: positionpolicy.ActionRelease, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonRelease, At: a111ObservedAt.Add(500 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("ApplyPositionPolicy(RELEASE): %v", err)
	}
	if released.Status != positionpolicy.StatusReleased {
		t.Fatalf("actual lifecycle status = %s, want RELEASED", released.Status)
	}
	before := a111ReadExitTuple(t, j, positionID)
	events := a111EventCount(t, j, positionID)
	if err := j.RefreshExitObservation(context.Background(), request); !errors.Is(err, ErrExitObservationConflict) {
		t.Fatalf("released refresh error = %v, want typed conflict", err)
	}
	if after := a111ReadExitTuple(t, j, positionID); after != before {
		t.Fatalf("released refresh changed the complete tuple:\nbefore=%+v\nafter=%+v", before, after)
	}
	if got := a111EventCount(t, j, positionID); got != events {
		t.Fatalf("released refresh changed exit events: %d -> %d", events, got)
	}
	if current, err := j.PositionPolicy(context.Background(), positionID); err != nil ||
		current.Status != positionpolicy.StatusReleased {
		t.Fatalf("released lifecycle reopened: state=%+v err=%v", current, err)
	}
}

func TestA111RefreshRejectsARealCurrentPendingProposalAndPreservesItsEvidence(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	// The compatibility judgement is a real journal lifecycle transition: it
	// arms pending_* without fabricating or replacing the complete v10 effective
	// snapshot. That leaves a semantically flat, otherwise valid refresh to prove
	// the pending precondition itself is checked.
	judgement := ExitJudgement{
		PositionID: state.PositionID, LifecycleGeneration: state.LifecycleGeneration,
		ObservedPrice: state.Snapshot.Snapshot.Line.ObservedPrice,
		HighWater:     state.HighWater, Baseline: state.Baseline,
		RatchetLevel: state.RatchetLevel, ActiveRung: state.ActiveRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionBaselineBreach), Level: RatchetNone,
			IntentID: "a111-real-pending-intent",
		},
	}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("RecordExitJudgement(pending): %v", err)
	}
	pending := exitStateOf(t, j, state.PositionID)
	if !pending.Pending() {
		t.Fatalf("fixture did not create a real current proposal: %+v", pending)
	}
	request := a111RefreshForState(t, pending, "a111-after-pending", "70400",
		"quote_fetched_at", a111ObservedAt.Add(2*time.Second))
	before := a111ReadExitTuple(t, j, state.PositionID)
	if before.PendingAction == "" || before.PendingLevel == "" || before.PendingIntentID == "" {
		t.Fatalf("pending fixture lacks durable proposal evidence: %+v", before)
	}
	events := a111EventCount(t, j, state.PositionID)
	if err := j.RefreshExitObservation(context.Background(), request); !errors.Is(err, ErrExitObservationConflict) {
		t.Fatalf("pending refresh error = %v, want typed conflict", err)
	}
	if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
		t.Fatalf("pending refresh changed the complete tuple:\nbefore=%+v\nafter=%+v", before, after)
	}
	if got := a111EventCount(t, j, state.PositionID); got != events {
		t.Fatalf("pending refresh changed exit events: %d -> %d", events, got)
	}
}

func TestA111RefreshExplicitlyRejectsAnIdenticalExecutableSnapshotWithoutWriting(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	currentLine, currentRecovery := ratchetSnapshotForState(t, seed,
		"a111-current-executable", "67900", seed.HighWater, seed.Baseline)
	if !currentLine.Orderable || currentLine.ExecutableProposal().Zero() {
		t.Fatalf("fixture lacks executable snapshot evidence: %+v", currentLine)
	}
	judgement := judgementForSnapshot(currentLine, currentRecovery)
	judgement.LifecycleGeneration = seed.LifecycleGeneration
	judgement.ObservationSource = "quote_fetched_at"
	judgement.ObservedAt = a111ObservedAt
	judgement.ArmSuppressedReason = ArmSuppressedWorkingOrder
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("persist current executable snapshot: %v", err)
	}
	current := exitStateOf(t, j, seed.PositionID)
	if current.SnapshotStatus != SnapshotStatusEvaluated || current.Snapshot.Snapshot == nil ||
		!current.Snapshot.Snapshot.Line.Orderable || current.Snapshot.Snapshot.Line.ExecutableProposal().Zero() ||
		current.Pending() {
		t.Fatalf("persisted current state lacks suppressed executable evidence: %+v", current)
	}

	persisted := *current.Snapshot.Snapshot
	if persisted.Line != currentLine || persisted.ObservationSource != "quote_fetched_at" ||
		persisted.ObservedAt != a111ObservedAt.Format(time.RFC3339Nano) {
		t.Fatalf("persisted executable tuple drifted from the exact judgement: %+v", persisted)
	}
	if err := exitpolicy.ValidateRecoveryDerivation(persisted.Line, persisted.RecoveryPolicy); err != nil {
		t.Fatalf("persisted executable recovery is invalid: %v", err)
	}
	request := a111RefreshForSnapshot(current, persisted.Line, persisted.RecoveryPolicy,
		persisted.ObservationSource, a111ObservedAt)
	before := a111ReadExitTuple(t, j, current.PositionID)
	if before.PendingAction != "" || before.PendingLevel != "" || before.PendingIntentID != "" {
		t.Fatalf("suppressed executable fixture unexpectedly armed pending evidence: %+v", before)
	}
	a111AssertRefreshNoWrite(t, j, current.PositionID, request, ErrExitObservationConflict)
	after := a111ReadExitTuple(t, j, current.PositionID)
	if after.PendingAction != before.PendingAction || after.PendingLevel != before.PendingLevel ||
		after.PendingIntentID != before.PendingIntentID {
		t.Fatalf("rejected executable refresh changed pending evidence: before=%+v after=%+v", before, after)
	}
}

func TestA111RefreshRejectsSeedAndEveryStrongerOrIncompatibleCandidateWithoutWriting(t *testing.T) {
	t.Run("seed", func(t *testing.T) {
		j := exitFixture(t)
		_, seed := openedPosition(t, j, "10")
		request := a111RefreshForState(t, seed, "a111-seed-refresh", "70000",
			"quote_fetched_at", a111ObservedAt)
		before := a111ReadExitTuple(t, j, seed.PositionID)
		if err := j.RefreshExitObservation(context.Background(), request); !errors.Is(err, ErrExitObservationConflict) {
			t.Fatalf("seed refresh error = %v, want conflict", err)
		}
		if after := a111ReadExitTuple(t, j, seed.PositionID); after != before {
			t.Fatalf("seed refresh wrote state: before=%+v after=%+v", before, after)
		}
	})

	for _, tc := range []struct {
		name   string
		mutate func(*ExitObservationRefresh)
	}{
		{"lifecycle_generation", func(r *ExitObservationRefresh) { r.LifecycleGeneration++ }},
		{"position_generation", func(r *ExitObservationRefresh) { r.Snapshot.PositionGeneration++ }},
		{"policy_identity", func(r *ExitObservationRefresh) { r.Snapshot.Policy.ID = "forged-policy" }},
		{"semantic_quantity", func(r *ExitObservationRefresh) { r.Snapshot.ProjectedQuantity = "9" }},
		{"semantic_protection", func(r *ExitObservationRefresh) { r.Snapshot.CurrentProtection = "69000" }},
		{"arm_suppression", func(r *ExitObservationRefresh) { r.Snapshot.Suppressed = "working_order" }},
		{"orderable", func(r *ExitObservationRefresh) {
			r.Snapshot.Orderable = true
			r.Snapshot.ProjectedQuantity = "10"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
			request := a111RefreshForState(t, state, "a111-reject-"+tc.name, "70400",
				"quote_fetched_at", a111ObservedAt.Add(time.Second))
			tc.mutate(&request)
			before := a111ReadExitTuple(t, j, state.PositionID)
			events := a111EventCount(t, j, state.PositionID)
			if err := j.RefreshExitObservation(context.Background(), request); err == nil {
				t.Fatal("incompatible refresh unexpectedly succeeded")
			}
			if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
				t.Fatalf("rejected refresh wrote state: before=%+v after=%+v", before, after)
			}
			if got := a111EventCount(t, j, state.PositionID); got != events {
				t.Fatalf("rejected refresh appended event: %d -> %d", events, got)
			}
		})
	}
}

func TestA111OperationalEqualityChecksEveryD1FieldUsingEvaluatorGeneratedDonors(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	base := state.Snapshot.Snapshot.Line

	strong, strongRecovery := a111RatchetCandidate(t, state, "a111-donor-strong", "10",
		func(in *exitpolicy.RatchetInput) {
			in.ObservedPrice, in.HighWater = "72000", "72000"
		})
	entry, entryRecovery := a111RatchetCandidate(t, state, "a111-donor-entry", "10",
		func(in *exitpolicy.RatchetInput) { in.Entry = "69000" })
	initialStop, initialStopRecovery := a111RatchetCandidate(t, state, "a111-donor-stop", "10",
		func(in *exitpolicy.RatchetInput) { in.InitialStop = "67500" })
	oneShare, oneShareRecovery := a111RatchetCandidate(t, state, "a111-donor-state-only", "1",
		func(in *exitpolicy.RatchetInput) {
			in.ObservedPrice, in.HighWater = "72000", "72000"
		})
	suppressed, suppressedRecovery := a111RatchetCandidate(t, state, "a111-donor-suppressed", "10",
		func(in *exitpolicy.RatchetInput) {
			in.ObservedPrice, in.HighWater = "72000", "72000"
			in.PendingAction = exitpolicy.ActionRatchetPartial
		})
	cancelFirst, cancelFirstRecovery := a111RatchetCandidate(t, state, "a111-donor-cancel-first", "10",
		func(in *exitpolicy.RatchetInput) {
			in.ObservedPrice, in.HighWater, in.Baseline = "69900", "72000", "70010"
			in.Level = exitpolicy.LevelBreakeven
			in.PendingAction = exitpolicy.ActionRatchetPartial
		})
	ladder, ladderRecovery := ladderSnapshotForState(t, state, "a111-donor-ladder", "71500", "71500", exitpolicy.NoRung)
	if err := exitpolicy.ValidateRecoveryDerivation(ladder, ladderRecovery); err != nil {
		t.Fatalf("ladder donor is not an internally valid evaluator tuple: %v", err)
	}
	generationState := state
	generationState.PositionGeneration++
	generation, generationRecovery := a111RatchetCandidate(t, generationState, "a111-donor-generation", "10", nil)

	type donor struct {
		line     exitpolicy.ExitLineSnapshot
		recovery exitpolicy.RecoveryPolicyDefinition
	}
	donors := map[string]donor{
		"strong":       {strong, strongRecovery},
		"entry":        {entry, entryRecovery},
		"initial_stop": {initialStop, initialStopRecovery},
		"one_share":    {oneShare, oneShareRecovery},
		"suppressed":   {suppressed, suppressedRecovery},
		"cancel_first": {cancelFirst, cancelFirstRecovery},
		"ladder":       {ladder, ladderRecovery},
		"generation":   {generation, generationRecovery},
	}

	// Every value below comes from an independently evaluated and recovery-validated
	// tuple. Copying exactly one value into the base isolates the equality predicate;
	// the second half of the test submits each complete donor to the journal so
	// temporal/provenance validation alone cannot make the test pass.
	for _, tc := range []struct {
		name  string
		donor string
		copy  func(*exitpolicy.ExitLineSnapshot, exitpolicy.ExitLineSnapshot)
	}{
		{"policy", "ladder", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Policy = b.Policy }},
		{"position_generation", "generation", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) {
			a.PositionGeneration = b.PositionGeneration
		}},
		{"entry_price", "entry", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.EntryPrice = b.EntryPrice }},
		{"initial_stop", "initial_stop", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.InitialStop = b.InitialStop }},
		{"current_protection", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) {
			a.CurrentProtection = b.CurrentProtection
		}},
		{"high_water", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.HighWater = b.HighWater }},
		{"ratchet_level", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.RatchetLevel = b.RatchetLevel }},
		{"active_rung", "ladder", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.ActiveRung = b.ActiveRung }},
		{"next_target", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.NextTarget = b.NextTarget }},
		{"next_protection", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) {
			a.NextProtection = b.NextProtection
		}},
		{"action", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Action = b.Action }},
		{"action_level", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Level = b.Level }},
		{"ratio", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Ratio = b.Ratio }},
		{"projected_quantity", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) {
			a.ProjectedQuantity = b.ProjectedQuantity
		}},
		{"orderable", "strong", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Orderable = b.Orderable }},
		{"state_only", "one_share", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.StateOnly = b.StateOnly }},
		{"suppressed", "suppressed", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) { a.Suppressed = b.Suppressed }},
		{"cancel_pending_first", "cancel_first", func(a *exitpolicy.ExitLineSnapshot, b exitpolicy.ExitLineSnapshot) {
			a.CancelPendingFirst = b.CancelPendingFirst
		}},
	} {
		t.Run("isolated_"+tc.name, func(t *testing.T) {
			candidate := base
			tc.copy(&candidate, donors[tc.donor].line)
			if candidate == base {
				t.Fatalf("evaluator donor did not vary %s", tc.name)
			}
			if sameExitOperationalLine(base, candidate) {
				t.Fatalf("operational equality ignored %s", tc.name)
			}
		})
	}

	for name, candidate := range donors {
		t.Run("journal_rejects_valid_"+name, func(t *testing.T) {
			before := a111ReadExitTuple(t, j, state.PositionID)
			events := a111EventCount(t, j, state.PositionID)
			request := a111RefreshForSnapshot(state, candidate.line, candidate.recovery,
				"quote_fetched_at", a111ObservedAt.Add(time.Second))
			if err := j.RefreshExitObservation(context.Background(), request); err == nil {
				t.Fatal("strictly newer, internally valid semantic change used the refresh path")
			}
			if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
				t.Fatalf("valid semantic conflict wrote state: before=%+v after=%+v", before, after)
			}
			if got := a111EventCount(t, j, state.PositionID); got != events {
				t.Fatalf("valid semantic conflict changed events: %d -> %d", events, got)
			}
		})
	}
}

func TestA111RefreshCannotOverwriteAConcurrentStrongerJudgement(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	staleRefresh := a111RefreshForState(t, state, "a111-stale-flat", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))

	stronger, recovery := ratchetSnapshotForState(t, state, "a111-stronger", "71000", "71000", state.Baseline)
	judgement := judgementForSnapshot(stronger, recovery)
	judgement.LifecycleGeneration = state.LifecycleGeneration
	judgement.ObservationSource = "quote_fetched_at"
	judgement.ObservedAt = a111ObservedAt.Add(2 * time.Second)
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("stronger judgement: %v", err)
	}
	before := a111ReadExitTuple(t, j, state.PositionID)
	a111AssertRefreshNoWrite(t, j, state.PositionID, staleRefresh, ErrExitObservationStale)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
		t.Fatalf("stronger tuple regressed: before=%+v after=%+v", before, after)
	}
}

func TestA111RefreshFailureRollsBackTheWholeTuple(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	request := a111RefreshForState(t, state, "a111-fault", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))
	before := a111ReadExitTuple(t, j, state.PositionID)
	events := a111EventCount(t, j, state.PositionID)
	j.exitWriteHook = func(stage string) error {
		if stage == "after_refresh_state" {
			return errors.New("a111 injected refresh failure")
		}
		return nil
	}
	if err := j.RefreshExitObservation(context.Background(), request); err == nil {
		t.Fatal("faulted refresh unexpectedly committed")
	}
	j.exitWriteHook = nil
	if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
		t.Fatalf("partial refresh survived rollback: before=%+v after=%+v", before, after)
	}
	if got := a111EventCount(t, j, state.PositionID); got != events {
		t.Fatalf("faulted refresh changed events: %d -> %d", events, got)
	}
}

func TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	newer := a111RefreshForState(t, state, "a111-official-newer", "70400",
		"quote_fetched_at", a111ObservedAt.Add(2*time.Second))
	if err := j.RefreshExitObservation(context.Background(), newer); err != nil {
		t.Fatal(err)
	}
	committed := a111ReadExitTuple(t, j, state.PositionID)

	older := a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-official-older", "70300",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))
	a111AssertRefreshNoWrite(t, j, state.PositionID, older, ErrExitObservationStale)
	ambiguous := a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-official-other", "70300",
		"quote_fetched_at", newer.ObservedAt)
	a111AssertRefreshNoWrite(t, j, state.PositionID, ambiguous, ErrExitObservationConflict)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != committed {
		t.Fatalf("temporal conflict wrote state: committed=%+v after=%+v", committed, after)
	}
	a111AssertRefreshNoWrite(t, j, state.PositionID, newer, nil)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != committed {
		t.Fatalf("exact duplicate churned updated_at or tuple: committed=%+v after=%+v", committed, after)
	}
}

func TestA111CycleSequenceOrdersFrozenClockRefreshAndSurvivesRestart(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "cycle:10", a111ObservedAt)
	cycle11 := a111RefreshForState(t, state, "a111-cycle-11", "70400", "cycle:11", a111ObservedAt)
	if err := j.RefreshExitObservation(context.Background(), cycle11); err != nil {
		t.Fatalf("cycle:11 refresh: %v", err)
	}
	committed := a111ReadExitTuple(t, j, state.PositionID)

	for _, tc := range []struct {
		name    string
		request ExitObservationRefresh
		want    error
	}{
		{"lower", a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-cycle-9", "70300", "cycle:9", a111ObservedAt), ErrExitObservationStale},
		{"same_sequence_other_identity", a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-cycle-11-other", "70300", "cycle:11", a111ObservedAt), ErrExitObservationConflict},
		{"malformed", a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-cycle-bad", "70300", "cycle:not-a-number", a111ObservedAt), ErrExitObservationConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a111AssertRefreshNoWrite(t, j, state.PositionID, tc.request, tc.want)
		})
	}
	a111AssertRefreshNoWrite(t, j, state.PositionID, cycle11, nil)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != committed {
		t.Fatalf("same-cycle replay churned state: before=%+v after=%+v", committed, after)
	}

	path := j.Path()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := openTestJournalAt(t, path)
	cycle12 := a111RefreshForState(t, exitStateOf(t, restarted, state.PositionID),
		"a111-cycle-12", "70300", "cycle:12", a111ObservedAt)
	if err := restarted.RefreshExitObservation(context.Background(), cycle12); err != nil {
		t.Fatalf("restart did not preserve durable cycle ordering: %v", err)
	}
}

func TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	type cycleRow struct {
		id, account, symbol, source, lifecycle string
		completed                              bool
		corrupt                                bool
	}
	insert := func(row cycleRow) {
		t.Helper()
		if _, err := j.db.ExecContext(ctx, `INSERT INTO positions
			(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
			VALUES(?,?,?,?,1,'OPEN','10','70000','2026-08-14T03:00:00Z')`,
			row.id, row.account, "kr", row.symbol); err != nil {
			t.Fatalf("insert position %s: %v", row.id, err)
		}
		if _, err := j.db.ExecContext(ctx, `INSERT INTO exit_states
			(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
			high_water,ratchet_level,completed,updated_at,lifecycle_generation,position_generation)
			VALUES(?,?,?,?,?,?,?,?,?,?,1,1)`, row.id, ExitPolicyRatchet, "70000", "68000", "2000",
			"68000", "70000", RatchetNone, boolInt(row.completed), "2026-08-14T03:00:00Z"); err != nil {
			t.Fatalf("insert exit state %s: %v", row.id, err)
		}
		raw := "{corrupt"
		if !row.corrupt {
			state := exitStateOf(t, j, row.id)
			line, recovery := ratchetSnapshotForState(t, state, "a111-cycle-evidence-"+row.id,
				"70400", "70000", "68000")
			var err error
			raw, err = encodeStoredSnapshot(StoredExitSnapshot{
				Line: line, RecoveryPolicy: recovery, ObservationSource: row.source,
				ObservedAt: "2026-08-14T04:00:00Z",
			})
			if err != nil {
				t.Fatalf("encode cycle evidence %s: %v", row.id, err)
			}
		}
		if _, err := j.db.ExecContext(ctx,
			`UPDATE exit_states SET effective_snapshot_json=?,snapshot_status=? WHERE position_id=?`,
			raw, SnapshotStatusEvaluated, row.id); err != nil {
			t.Fatalf("store cycle evidence %s: %v", row.id, err)
		}
		if row.lifecycle != "" {
			if _, err := j.db.ExecContext(ctx, `INSERT INTO position_policy_lifecycles
				(position_id,adoption_generation,version,status,observed_at,updated_at)
				VALUES(?,1,1,?,'2026-08-14T03:00:00Z','2026-08-14T03:00:00Z')`,
				row.id, row.lifecycle); err != nil {
				t.Fatalf("insert lifecycle %s: %v", row.id, err)
			}
		}
	}

	for _, row := range []cycleRow{
		{id: "cycle-managed-7", account: "acct-cycle", symbol: "CYC07", source: "cycle:7", lifecycle: string(positionpolicy.StatusManaged)},
		{id: "cycle-managed-12", account: "acct-cycle", symbol: "CYC12", source: "cycle:12"},
		{id: "cycle-corrupt", account: "acct-cycle", symbol: "CYCBAD", corrupt: true},
		{id: "cycle-zero", account: "acct-cycle", symbol: "CYCZERO", source: "cycle:0"},
		{id: "cycle-malformed", account: "acct-cycle", symbol: "CYCMAL", source: "cycle:nope"},
		{id: "cycle-official", account: "acct-cycle", symbol: "CYCOFF", source: "quote_fetched_at"},
		{id: "cycle-completed", account: "acct-cycle", symbol: "CYCDONE", source: "cycle:77", completed: true},
		{id: "cycle-released", account: "acct-cycle", symbol: "CYCREL", source: "cycle:88", lifecycle: string(positionpolicy.StatusReleased)},
		{id: "cycle-other-account", account: "acct-other", symbol: "CYCOTH", source: "cycle:99"},
	} {
		insert(row)
	}

	got, err := j.MaxExitObservationCycle(ctx, " acct-cycle ")
	if err != nil {
		t.Fatalf("MaxExitObservationCycle: %v", err)
	}
	if got != 12 {
		t.Fatalf("maximum cycle = %d, want only current active managed cycle:12", got)
	}
}

func TestA111EqualTimeOfficialEvidenceOutranksCycleButNotTheReverse(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "cycle:10", a111ObservedAt)
	official := a111RefreshForState(t, state, "a111-official-wins", "70400",
		"quote_fetched_at", a111ObservedAt)
	if err := j.RefreshExitObservation(context.Background(), official); err != nil {
		t.Fatalf("official did not supersede equal-time cycle: %v", err)
	}
	committed := a111ReadExitTuple(t, j, state.PositionID)
	cycle := a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-cycle-loses", "70300",
		"cycle:999", a111ObservedAt)
	a111AssertRefreshNoWrite(t, j, state.PositionID, cycle, ErrExitObservationStale)
	otherOfficial := a111RefreshForState(t, exitStateOf(t, j, state.PositionID), "a111-official-other", "70300",
		"quote_fetched_at", a111ObservedAt)
	a111AssertRefreshNoWrite(t, j, state.PositionID, otherOfficial, ErrExitObservationConflict)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != committed {
		t.Fatalf("equal-time loser wrote state: committed=%+v after=%+v", committed, after)
	}
}

func TestA111SameCycleRaceLeavesExactlyOneCompleteWinner(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "cycle:20", a111ObservedAt)
	beforeEvents := a111EventCount(t, j, state.PositionID)
	left := a111RefreshForState(t, state, "a111-cycle-race-left", "70400", "cycle:21", a111ObservedAt)
	right := a111RefreshForState(t, state, "a111-cycle-race-right", "70300", "cycle:21", a111ObservedAt)

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, request := range []ExitObservationRefresh{left, right} {
		request := request
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- j.RefreshExitObservation(context.Background(), request)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrExitObservationConflict):
			conflicts++
		default:
			t.Fatalf("race result = %v, want success or typed conflict", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("race results success/conflict = %d/%d, want 1/1", success, conflicts)
	}
	got := exitStateOf(t, j, state.PositionID)
	if got.Snapshot.Snapshot == nil {
		t.Fatal("race left no complete effective snapshot")
	}
	winner := got.Snapshot.Snapshot.Line.ObservationID
	if winner != left.Snapshot.ObservationID && winner != right.Snapshot.ObservationID {
		t.Fatalf("race winner = %q, want one submitted identity", winner)
	}
	loser := left
	if winner == left.Snapshot.ObservationID {
		loser = right
	}
	committed := a111ReadExitTuple(t, j, state.PositionID)
	a111AssertRefreshNoWrite(t, j, state.PositionID, loser, ErrExitObservationConflict)
	if after := a111ReadExitTuple(t, j, state.PositionID); after != committed {
		t.Fatalf("deterministic race-loser replay changed winner: before=%+v after=%+v", committed, after)
	}
	if events := a111EventCount(t, j, state.PositionID); events != beforeEvents {
		t.Fatalf("cycle race appended events: %d -> %d", beforeEvents, events)
	}
	t.Logf("durable cycle race winner: %s", fmt.Sprintf("%s/%s", got.Snapshot.Snapshot.ObservationSource, winner))
}

func TestA111RefreshRejectsCompletedStateWithoutReopeningIt(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	request := a111RefreshForState(t, state, "a111-completed", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))
	if _, err := j.db.Exec(`UPDATE exit_states SET completed=1 WHERE position_id=?`, state.PositionID); err != nil {
		t.Fatal(err)
	}
	before := a111ReadExitTuple(t, j, state.PositionID)
	if err := j.RefreshExitObservation(context.Background(), request); !errors.Is(err, ErrExitStateCompleted) {
		t.Fatalf("completed refresh error = %v, want ErrExitStateCompleted", err)
	}
	if after := a111ReadExitTuple(t, j, state.PositionID); after != before {
		t.Fatalf("completed refresh wrote state: before=%+v after=%+v", before, after)
	}
}

func TestA111RefreshPreservesTheExistingPolicyAndGenerationOnAValidCandidate(t *testing.T) {
	j, state := a111EvaluatedPosition(t, "quote_fetched_at", a111ObservedAt)
	request := a111RefreshForState(t, state, "a111-policy-control", "70400",
		"quote_fetched_at", a111ObservedAt.Add(time.Second))
	if request.Snapshot.Policy == (exitpolicy.PolicyIdentity{}) || request.Snapshot.PositionGeneration < 1 {
		t.Fatalf("fixture lacks policy/generation evidence: %+v", request.Snapshot)
	}
	if err := j.RefreshExitObservation(context.Background(), request); err != nil {
		t.Fatalf("valid policy/generation refresh: %v", err)
	}
	got := exitStateOf(t, j, state.PositionID)
	if got.PolicyIdentity != request.Snapshot.Policy ||
		got.PositionGeneration != request.Snapshot.PositionGeneration ||
		got.LifecycleGeneration != request.LifecycleGeneration {
		t.Fatalf("authority tuple drifted: %+v", got)
	}
}
