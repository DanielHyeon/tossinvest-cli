package journal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func ratchetSnapshotForState(t *testing.T, state ExitState, observation, price, high, baseline string) exitpolicy.ExitLineSnapshot {
	t.Helper()
	snapshot, err := exitpolicy.EvaluateRatchetSnapshot(exitpolicy.RatchetSnapshotInput{
		Context: exitpolicy.SnapshotContext{
			PositionID: state.PositionID, PositionGeneration: state.PositionGeneration,
			ObservationID: observation, RemainingQuantity: "10",
		},
		Input: exitpolicy.RatchetInput{
			Entry: "70000", InitialStop: "68000", ObservedPrice: price,
			HighWater: high, Baseline: baseline, RealBreakeven: "70010",
			TakenRatioTotal: state.TakenRatioTotal, Level: exitpolicy.Level(state.RatchetLevel),
		},
	})
	if err != nil {
		t.Fatalf("EvaluateRatchetSnapshot: %v", err)
	}
	return snapshot
}

func judgementForSnapshot(snapshot exitpolicy.ExitLineSnapshot) ExitJudgement {
	return ExitJudgement{
		PositionID: snapshot.PositionID, Snapshot: snapshot,
		RecoveryPolicy:    exitpolicy.NewRatchetRecoveryPolicy(exitpolicy.DefaultRatchetConfig(), "70010", "0"),
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

	first := ratchetSnapshotForState(t, seed, "obs-first", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(first)); err != nil {
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
	stale := ratchetSnapshotForState(t, seed, "obs-stale", "70200", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(stale)); err != nil {
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

func TestExitSnapshotWriteStagesRollbackAsOneTransaction(t *testing.T) {
	for _, stage := range []string{"after_state", "after_arm", "after_event"} {
		t.Run(stage, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			before, err := j.ExitEvents(context.Background(), seed.PositionID)
			if err != nil {
				t.Fatal(err)
			}
			snapshot := ratchetSnapshotForState(t, seed, "obs-crash-"+stage, "67900", "70000", "68000")
			judgement := judgementForSnapshot(snapshot)
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
	snapshot := ratchetSnapshotForState(t, seed, "obs-forged-next", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot)); err != nil {
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
	snapshot := ratchetSnapshotForState(t, seed, "obs-duplicate", "67900", "70000", "68000")
	judgement := judgementForSnapshot(snapshot)
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
	snapshot := ratchetSnapshotForState(t, first, "obs-corrupt", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot)); err != nil {
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
	snapshot := ratchetSnapshotForState(t, seed, "obs-read-model", "70500", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snapshot)); err != nil {
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
