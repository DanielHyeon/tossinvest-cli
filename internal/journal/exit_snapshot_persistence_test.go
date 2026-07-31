package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func ratchetSnapshotForState(t *testing.T, state ExitState, observation, price, high,
	baseline string) (exitpolicy.ExitLineSnapshot, exitpolicy.RecoveryPolicyDefinition) {
	t.Helper()
	evaluation := exitpolicy.RatchetSnapshotInput{
		Context: exitpolicy.SnapshotContext{
			PositionID: state.PositionID, PositionGeneration: state.PositionGeneration,
			ObservationID: observation, RemainingQuantity: "10",
		},
		Input: exitpolicy.RatchetInput{
			Entry: "70000", InitialStop: "68000", ObservedPrice: price,
			HighWater: high, Baseline: baseline, RealBreakeven: "70010",
			TakenRatioTotal: state.TakenRatioTotal, Level: exitpolicy.Level(state.RatchetLevel),
		},
	}
	snapshot, err := exitpolicy.EvaluateRatchetSnapshot(evaluation)
	if err != nil {
		t.Fatalf("EvaluateRatchetSnapshot: %v", err)
	}
	snapshot = snapshot.ChangedFromState(high, baseline, exitpolicy.Level(state.RatchetLevel), exitpolicy.NoRung)
	return snapshot, exitpolicy.NewRatchetRecoveryPolicy(evaluation)
}

func judgementForSnapshot(snapshot exitpolicy.ExitLineSnapshot,
	recovery exitpolicy.RecoveryPolicyDefinition) ExitJudgement {
	return ExitJudgement{
		PositionID: snapshot.PositionID, Snapshot: snapshot,
		RecoveryPolicy:    recovery,
		ObservationSource: "test_quote", ObservedAt: time.Date(2026, 7, 31, 14, 0, 0, 0, time.UTC),
		Provenance: ExitDecisionProvenance{
			ObservationID: snapshot.ObservationID, SnapshotID: snapshot.SnapshotID,
			DecisionID: snapshot.DecisionID, Policy: snapshot.Policy,
		},
		ObservedPrice: snapshot.ObservedPrice, HighWater: snapshot.HighWater,
		Baseline: snapshot.CurrentProtection, RatchetLevel: string(snapshot.RatchetLevel),
		ActiveRung: snapshot.ActiveRung,
	}
}

func TestExitSnapshotPersistsWholeRecoveryCandidates(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")

	first, firstRecovery := ratchetSnapshotForState(t, seed, "obs-first", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(first, firstRecovery)); err != nil {
		t.Fatalf("first judgement: %v", err)
	}
	stored := exitStateOf(t, j, seed.PositionID)
	if stored.SnapshotStatus != SnapshotStatusEvaluated || stored.Snapshot.Snapshot == nil {
		t.Fatalf("stored snapshot = %+v", stored.Snapshot)
	}
	if stored.Snapshot.Snapshot.Line != first || stored.Snapshot.Snapshot.OutputDigest == "" {
		t.Fatalf("effective snapshot changed during persistence: got %+v want %+v", stored.Snapshot.Snapshot.Line, first)
	}

	// This evaluator raced from the seed and is weaker on the watermark axis.
	// Recovery must keep the complete saved tuple; no scalar may be mixed in.
	stale, staleRecovery := ratchetSnapshotForState(t, seed, "obs-stale", "70200", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(stale, staleRecovery)); err != nil {
		t.Fatalf("stale judgement: %v", err)
	}
	after := exitStateOf(t, j, seed.PositionID)
	if after.Snapshot.Snapshot == nil || after.Snapshot.Snapshot.Line != first {
		t.Fatalf("saved tuple was not retained whole: %+v", after.Snapshot.Snapshot)
	}
	events, err := j.ExitEvents(context.Background(), seed.PositionID)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1].Evaluation
	if last.EffectiveSource != EffectiveSourceSaved || last.Saved.Snapshot == nil ||
		last.Recomputed.Snapshot == nil || last.Effective.Snapshot == nil {
		t.Fatalf("recovery evidence = %+v", last)
	}
	path := j.Path()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := openTestJournalAt(t, path)
	reopened := exitStateOf(t, restarted, seed.PositionID)
	if reopened.Snapshot.Snapshot == nil || reopened.Snapshot.Snapshot.Line != first {
		t.Fatalf("reopened effective snapshot = %+v", reopened.Snapshot)
	}
}

func TestSavedMonotoneRecoveryCannotArmRecomputedOrder(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")

	saved, savedRecovery := ratchetSnapshotForState(t, seed, "obs-saved-safer", "70500", "70000", "68000")
	if saved.Orderable {
		t.Fatalf("saved fixture unexpectedly orderable: %+v", saved)
	}
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(saved, savedRecovery)); err != nil {
		t.Fatal(err)
	}

	stale, staleRecovery := ratchetSnapshotForState(t, seed, "obs-stale-order", "67900", "70000", "68000")
	if !stale.Orderable {
		t.Fatalf("stale fixture must propose an order: %+v", stale)
	}
	judgement := judgementForSnapshot(stale, staleRecovery)
	judgement.Proposal = &ExitProposal{Action: string(stale.Action), Level: stale.Level,
		IntentID: "exit-stale", Provenance: judgement.Provenance}
	result, err := j.RecordExitJudgementResult(context.Background(), judgement)
	if err != nil {
		t.Fatal(err)
	}
	if result.EffectiveSource != EffectiveSourceSaved ||
		result.ArmOutcome != ExitArmSuppressedSavedMonotone || result.ArmedProposal != nil {
		t.Fatalf("durable result = %+v, want saved/no arm", result)
	}
	after := exitStateOf(t, j, seed.PositionID)
	if after.Pending() || after.Snapshot.Snapshot == nil || after.Snapshot.Snapshot.Line != saved {
		t.Fatalf("saved recovery armed or replaced state: %+v", after)
	}
	events, err := j.ExitEvents(context.Background(), seed.PositionID)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.ProposedIntentID != "" || last.ArmSuppressedReason != "" ||
		last.Evaluation.EffectiveSource != EffectiveSourceSaved ||
		last.Evaluation.Recomputed.Snapshot == nil || !last.Evaluation.Recomputed.Snapshot.Line.Orderable {
		t.Fatalf("saved recovery audit evidence = %+v", last)
	}
}

func TestExitSnapshotWriteStagesRollbackAsOneTransaction(t *testing.T) {
	for _, stage := range []string{"after_state", "after_arm", "after_event"} {
		t.Run(stage, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			before, err := j.ExitEvents(context.Background(), seed.PositionID)
			if err != nil {
				t.Fatal(err)
			}
			snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-crash-"+stage, "67900", "70000", "68000")
			judgement := judgementForSnapshot(snapshot, recovery)
			judgement.Proposal = &ExitProposal{
				Action: string(snapshot.Action), Level: snapshot.Level, IntentID: "exit-fault",
				Provenance: judgement.Provenance,
			}
			j.exitWriteHook = func(got string) error {
				if got == stage {
					return errors.New("injected write failure")
				}
				return nil
			}
			if err := j.RecordExitJudgement(context.Background(), judgement); err == nil {
				t.Fatal("faulted write unexpectedly committed")
			}
			j.exitWriteHook = nil
			path := j.Path()
			if err := j.Close(); err != nil {
				t.Fatal(err)
			}
			restarted := openTestJournalAt(t, path)
			after := exitStateOf(t, restarted, seed.PositionID)
			if after.SnapshotStatus != SnapshotStatusSeed || after.Pending() || after.Baseline != seed.Baseline {
				t.Fatalf("partial state survived rollback: %+v", after)
			}
			events, err := restarted.ExitEvents(context.Background(), seed.PositionID)
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != len(before) {
				t.Fatalf("events = %d, want rollback to %d", len(events), len(before))
			}
		})
	}
}

func TestExitSnapshotOutputDigestRejectsForgedDerivedLine(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-forged-next", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot, recovery)); err != nil {
		t.Fatal(err)
	}
	state := exitStateOf(t, j, seed.PositionID)
	forged := *state.Snapshot.Snapshot
	forged.Line.NextTarget = "99999"
	if _, err := encodeStoredSnapshot(forged); !errors.Is(err, exitpolicy.ErrRecoveryIdentity) {
		t.Fatalf("forged next line with a newly computed digest = %v, want semantic derivation refusal", err)
	}
	raw, err := json.Marshal(forged) // deliberately retain the old output digest
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE exit_states SET next_target=?, effective_snapshot_json=? WHERE position_id=?`,
		forged.Line.NextTarget, string(raw), seed.PositionID); err != nil {
		t.Fatal(err)
	}
	results, err := j.OpenExitStateResults(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !errors.Is(results[0].Corruption, ErrExitSnapshotCorrupt) ||
		results[0].State.Snapshot.UnknownReason != "invalid_effective_snapshot" {
		t.Fatalf("forged derived line was not typed as corruption: %+v", results)
	}
}

func TestKnownLegacyIdentityIsResolvedInMemoryWithoutBackfill(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	if _, err := j.db.Exec(`UPDATE exit_states SET snapshot_status=NULL, policy_version=NULL,
		policy_digest=NULL, snapshot_id=NULL, decision_id=NULL, observation_id=NULL,
		position_generation=NULL, next_target=NULL, next_protection=NULL,
		last_observation_source=NULL, last_observed_at=NULL, snapshot_action=NULL,
		snapshot_ratio=NULL, projected_quantity=NULL, state_only=NULL,
		suppressed_reason=NULL, effective_snapshot_json=NULL WHERE position_id=?`, seed.PositionID); err != nil {
		t.Fatal(err)
	}
	results, err := j.OpenExitStateResults(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	want, err := exitpolicy.LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Corruption != nil || results[0].State.PolicyIdentity != want ||
		results[0].State.Snapshot.UnknownReason != "legacy_snapshot_absent" {
		t.Fatalf("legacy read = %+v, want exact in-memory identity and unknown snapshot", results)
	}
	var version, digest any
	if err := j.db.QueryRow(`SELECT policy_version,policy_digest FROM exit_states WHERE position_id=?`, seed.PositionID).
		Scan(&version, &digest); err != nil {
		t.Fatal(err)
	}
	if version != nil || digest != nil {
		t.Fatalf("legacy identity was backfilled on disk: version=%v digest=%v", version, digest)
	}
}

func TestExitSnapshotDuplicateDecisionIsNotRearmed(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-duplicate", "67900", "70000", "68000")
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.Proposal = &ExitProposal{Action: string(snapshot.Action), Level: snapshot.Level,
		IntentID: "exit-duplicate", Provenance: judgement.Provenance}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(context.Background(), judgement); !errors.Is(err, ErrProposalPending) {
		t.Fatalf("duplicate error = %v, want conservative no-submit signal", err)
	}
	events, err := j.ExitEvents(context.Background(), seed.PositionID)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range events {
		if event.Evaluation.Recomputed.Snapshot != nil && event.Evaluation.Recomputed.Snapshot.Line.DecisionID == snapshot.DecisionID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("decision events = %d, want exactly one", count)
	}
}

func TestExitSnapshotCorruptionIsPerPositionAndNeverRecomputed(t *testing.T) {
	j := exitFixture(t)
	_, first := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, first, "obs-corrupt", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot, recovery)); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE exit_states SET effective_snapshot_json=effective_snapshot_json || '{}'
		WHERE position_id=?`, first.PositionID); err != nil {
		t.Fatal(err)
	}

	// Add a second healthy row: one semantic defect must not turn the account
	// scan into a global SQL/driver failure.
	o := place(t, j, order{intentID: "i-second", attemptID: "a-second", orderID: "o-second", decisionID: "d-second", symbol: "000660"})
	if _, err := j.RecordFill(context.Background(), terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	if _, err := j.OpenExitState(context.Background(), ExitStateSeed{PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000"}); err != nil {
		t.Fatal(err)
	}
	results, err := j.OpenExitStateResults(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want both positions", len(results))
	}
	corrupt := 0
	for _, result := range results {
		if result.Corruption != nil {
			corrupt++
			if !errors.Is(result.Corruption, ErrExitSnapshotCorrupt) ||
				!strings.Contains(result.State.Snapshot.UnknownReason, "invalid") {
				t.Fatalf("typed corruption = %+v", result)
			}
		}
	}
	if corrupt != 1 {
		t.Fatalf("corrupt rows = %d, want one", corrupt)
	}
}

func TestReadOnlyExposesTypedSavedRecomputedEffectiveAndFreshness(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-read-model", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot, recovery)); err != nil {
		t.Fatal(err)
	}
	ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: j.Path()})
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	positions, err := ro.LivePositionExits(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 1 || positions[0].Exit.Snapshot.Snapshot == nil ||
		positions[0].Exit.Snapshot.Snapshot.Line != snapshot {
		t.Fatalf("read-only effective snapshot = %+v", positions)
	}
	events, err := ro.AccountExitEvents(context.Background(), "acct-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	latest := events[len(events)-1].Event.Evaluation
	if latest.Recomputed.Snapshot == nil || latest.Effective.Snapshot == nil ||
		latest.Saved.UnknownReason != "no_saved_evaluation" {
		t.Fatalf("typed event views = %+v", latest)
	}
	observedAt, err := time.Parse(time.RFC3339Nano, latest.Effective.Snapshot.ObservedAt)
	if err != nil {
		t.Fatal(err)
	}
	if view := latest.Effective.WithFreshness(observedAt.Add(-time.Second), time.Minute); !view.Stale || view.StaleReason != "observation_in_future" {
		t.Fatalf("future freshness = %+v", view)
	}
	if view := latest.Effective.WithFreshness(observedAt.Add(2*time.Minute), time.Minute); !view.Stale || view.StaleReason != "observation_older_than_limit" {
		t.Fatalf("old freshness = %+v", view)
	}
}

func TestSeedRejectsEveryPartialV10OutputColumn(t *testing.T) {
	columns := []struct {
		name  string
		value any
	}{
		{"snapshot_id", "snapshot"},
		{"decision_id", "decision"},
		{"observation_id", "observation"},
		{"next_target", "71000"},
		{"next_protection", "68000"},
		{"last_observation_source", "test"},
		{"last_observed_at", "2026-07-31T14:00:00Z"},
		{"snapshot_action", "NONE"},
		{"snapshot_ratio", "0.25"},
		{"projected_quantity", "1"},
		{"state_only", 0},
		{"suppressed_reason", "blocked_by_pending_proposal"},
		{"effective_snapshot_json", "{}"},
	}
	for _, column := range columns {
		t.Run(column.name, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			query := fmt.Sprintf("UPDATE exit_states SET %s=? WHERE position_id=?", column.name)
			if _, err := j.db.Exec(query, column.value, seed.PositionID); err != nil {
				t.Fatal(err)
			}
			results, err := j.OpenExitStateResults(context.Background(), "acct-1")
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || !errors.Is(results[0].Corruption, ErrExitSnapshotCorrupt) ||
				results[0].State.Snapshot.UnknownReason != "partial_seed_tuple" {
				t.Fatalf("partial %s result = %+v", column.name, results)
			}
		})
	}
}

func TestLegacyDetectionRequiresEveryV10ColumnToBeNull(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	if _, err := j.db.Exec(`UPDATE exit_states SET snapshot_status=NULL, policy_version=NULL,
		policy_digest=NULL, snapshot_id=NULL, decision_id=NULL, observation_id=NULL,
		position_generation=NULL, next_target=NULL, next_protection=NULL,
		last_observation_source=NULL, last_observed_at=NULL, snapshot_action=NULL,
		snapshot_ratio=NULL, projected_quantity='1', state_only=NULL,
		suppressed_reason=NULL, effective_snapshot_json=NULL WHERE position_id=?`, seed.PositionID); err != nil {
		t.Fatal(err)
	}
	results, err := j.OpenExitStateResults(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !errors.Is(results[0].Corruption, ErrExitSnapshotCorrupt) ||
		results[0].State.Snapshot.UnknownReason != "partial_snapshot_tuple" {
		t.Fatalf("single-column v10 evidence was treated as legacy: %+v", results)
	}
}

func TestEvaluatedTupleRequiresEveryNonOptionalFlattenedColumn(t *testing.T) {
	for _, test := range []struct{ column, reason string }{
		{"position_generation", "partial_policy_tuple"},
		{"projected_quantity", "partial_evaluated_tuple"},
		{"state_only", "partial_evaluated_tuple"},
	} {
		t.Run(test.column, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-null-"+test.column, "70500", "70000", "68000")
			if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot, recovery)); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(fmt.Sprintf("UPDATE exit_states SET %s=NULL WHERE position_id=?", test.column),
				seed.PositionID); err != nil {
				t.Fatal(err)
			}
			results, err := j.OpenExitStateResults(context.Background(), "acct-1")
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || !errors.Is(results[0].Corruption, ErrExitSnapshotCorrupt) ||
				results[0].State.Snapshot.UnknownReason != test.reason {
				t.Fatalf("NULL %s result = %+v", test.column, results)
			}
		})
	}
}

func TestLegacyJudgementRejectsEveryArmSuppressionReason(t *testing.T) {
	for _, reason := range []string{ArmSuppressedWorkingOrder, "unknown"} {
		t.Run(reason, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			judgement := ExitJudgement{PositionID: seed.PositionID, HighWater: seed.HighWater,
				Baseline: seed.Baseline, RatchetLevel: seed.RatchetLevel, ActiveRung: seed.ActiveRung,
				ArmSuppressedReason: reason}
			if err := j.RecordExitJudgement(context.Background(), judgement); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("error = %v, want invalid request", err)
			}
			if after := exitStateOf(t, j, seed.PositionID); after.SnapshotStatus != SnapshotStatusSeed {
				t.Fatalf("legacy suppression changed state: %+v", after)
			}
		})
	}
}

func TestOrderableSnapshotMustArmItsExactProposal(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*ExitJudgement)
	}{
		{"missing", func(j *ExitJudgement) { j.Proposal = nil }},
		{"unknown_suppression", func(j *ExitJudgement) {
			j.Proposal = nil
			j.ArmSuppressedReason = "unknown"
		}},
		{"different_action", func(j *ExitJudgement) { j.Proposal.Action = string(exitpolicy.ActionLadderStop) }},
		{"different_level", func(j *ExitJudgement) { j.Proposal.Level = "PROFIT_LOCK" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-proposal-"+test.name, "67900", "70000", "68000")
			judgement := judgementForSnapshot(snapshot, recovery)
			judgement.Proposal = &ExitProposal{
				Action: string(snapshot.Action), Level: snapshot.Level,
				IntentID: "exit-exact", Provenance: judgement.Provenance,
			}
			test.mutate(&judgement)
			if err := j.RecordExitJudgement(context.Background(), judgement); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("error = %v, want invalid request", err)
			}
			after := exitStateOf(t, j, seed.PositionID)
			if after.SnapshotStatus != SnapshotStatusSeed || after.Pending() {
				t.Fatalf("invalid proposal changed durable state: %+v", after)
			}
		})
	}
}

func TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-working-order", "67900", "70000", "68000")
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.ArmSuppressedReason = ArmSuppressedWorkingOrder
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}
	after := exitStateOf(t, j, seed.PositionID)
	if after.SnapshotStatus != SnapshotStatusEvaluated || after.Snapshot.Snapshot == nil ||
		after.Snapshot.Snapshot.Line != snapshot || after.Pending() {
		t.Fatalf("suppressed arm did not retain exact snapshot without pending proposal: %+v", after)
	}
	events, err := j.ExitEvents(context.Background(), seed.PositionID)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.ArmSuppressedReason != ArmSuppressedWorkingOrder || last.ProposedIntentID != "" ||
		last.Evaluation.Recomputed.Snapshot == nil || !last.Evaluation.Recomputed.Snapshot.Line.Orderable {
		t.Fatalf("typed arm suppression evidence = %+v", last)
	}
	ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: j.Path()})
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	accountEvents, err := ro.AccountExitEvents(context.Background(), "acct-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := accountEvents[len(accountEvents)-1].Event.ArmSuppressedReason; got != ArmSuppressedWorkingOrder {
		t.Fatalf("read-model arm suppression = %q", got)
	}
}

func TestExitEventReadRejectsForgedArmSuppressionEvidence(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*testing.T, *Journal, ExitState, exitpolicy.ExitLineSnapshot)
	}{
		{"unknown_reason", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET arm_suppressed_reason='unknown' WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"deleted_reason_null", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET arm_suppressed_reason=NULL WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"deleted_reason_empty", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET arm_suppressed_reason='' WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"forged_source", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET effective_source='saved_monotone' WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"missing_effective", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET effective_snapshot_json=NULL WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"known_reason_nonorderable", func(t *testing.T, j *Journal, seed ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			other, recovery := ratchetSnapshotForState(t, seed, "obs-nonorderable-evidence", "70500", "70000", "68000")
			raw, err := encodeStoredSnapshot(StoredExitSnapshot{Line: other, RecoveryPolicy: recovery,
				ObservationSource: "test_quote", ObservedAt: time.Date(2026, 7, 31, 14, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE exit_events SET recomputed_snapshot_json=?, effective_snapshot_json=? WHERE decision_id=?`, raw, raw, snapshot.DecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{"forged_armed_action", func(t *testing.T, j *Journal, _ ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			_, err := j.db.Exec(`UPDATE exit_events SET arm_suppressed_reason=NULL, action='LADDER_STOP', proposed_intent_id='forged' WHERE decision_id=?`, snapshot.DecisionID)
			if err != nil {
				t.Fatal(err)
			}
		}},
		{"swapped_valid_effective", func(t *testing.T, j *Journal, seed ExitState, snapshot exitpolicy.ExitLineSnapshot) {
			other, recovery := ratchetSnapshotForState(t, seed, "obs-other-event", "70500", "70000", "68000")
			otherJSON, err := encodeStoredSnapshot(StoredExitSnapshot{Line: other, RecoveryPolicy: recovery,
				ObservationSource: "test_quote", ObservedAt: time.Date(2026, 7, 31, 14, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE exit_events SET effective_snapshot_json=? WHERE decision_id=?`, otherJSON, snapshot.DecisionID); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			snapshot, recovery := ratchetSnapshotForState(t, seed, "obs-event-"+test.name, "67900", "70000", "68000")
			judgement := judgementForSnapshot(snapshot, recovery)
			judgement.ArmSuppressedReason = ArmSuppressedWorkingOrder
			if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
				t.Fatal(err)
			}
			test.mutate(t, j, seed, snapshot)
			if _, err := j.ExitEvents(context.Background(), seed.PositionID); !errors.Is(err, ErrExitSnapshotCorrupt) {
				t.Fatalf("event read error = %v, want typed corruption", err)
			}
			ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: j.Path()})
			if err != nil {
				t.Fatal(err)
			}
			defer ro.Close()
			events, err := ro.AccountExitEvents(context.Background(), "acct-1", 10)
			if err != nil {
				t.Fatal(err)
			}
			var got *ExitEvent
			for i := range events {
				if events[i].Event.PositionID == seed.PositionID {
					got = &events[i].Event
				}
			}
			if got == nil || got.Evaluation.Effective.Snapshot != nil ||
				got.Evaluation.Effective.UnknownReason != "invalid_arm_suppression_evidence" {
				t.Fatalf("account event did not fail closed: %+v", got)
			}
		})
	}
}
