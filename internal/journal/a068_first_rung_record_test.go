package journal

// a068 at the record boundary: the transition that quarantined a live position.
//
// The exitpolicy tests pin the comparison. This one pins what the ledger does
// with it, because the quarantine is written here — inside record(), in the same
// transaction, and the position is refused by the observation loop from the next
// cycle on.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// ladderSnapshotForState evaluates the common hybrid ladder the live account
// runs, from the state the ledger currently holds.
func ladderSnapshotForState(t *testing.T, state ExitState, observation, price, high string,
	rung int) (exitpolicy.ExitLineSnapshot, exitpolicy.RecoveryPolicyDefinition) {
	t.Helper()
	policy, err := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderHybrid50, false)
	if err != nil {
		t.Fatalf("CommonLadderForPosition: %v", err)
	}
	identity, err := policy.Identity()
	if err != nil {
		t.Fatalf("policy identity: %v", err)
	}
	evaluation := exitpolicy.LadderSnapshotInput{
		Context: exitpolicy.SnapshotContext{
			PositionID: state.PositionID, PositionGeneration: state.PositionGeneration,
			ObservationID: observation, RemainingQuantity: "10",
		},
		Input: exitpolicy.LadderInput{
			EntryPrice: "70000", InitialStop: "68000", ObservedPrice: price,
			HighWater: high, Baseline: state.Baseline,
			State: exitpolicy.LadderState{
				PolicyID: identity.ID, PolicyVersion: identity.Version, PolicyDigest: identity.Digest,
				ActivatedRung: rung, TakenRatioTotal: state.TakenRatioTotal,
				PendingRung: exitpolicy.NoRung,
			},
			Policy: policy,
		},
	}
	snapshot, err := exitpolicy.EvaluateLadderSnapshot(evaluation)
	if err != nil {
		t.Fatalf("EvaluateLadderSnapshot: %v", err)
	}
	snapshot = snapshot.ChangedFromState(high, state.Baseline, exitpolicy.LevelNone, rung)
	return snapshot, exitpolicy.NewLadderRecoveryPolicy(evaluation)
}

// TestTheFirstRungActivationIsRecordedRatherThanQuarantined is a068 task 3.1.
//
// The shape is the 466100 shape: a canonical snapshot already exists at no rung,
// and the next observation crosses the first take-profit line. Before a068 this
// wrote an ambiguous_recovery quarantine and the position stopped being judged.
func TestTheFirstRungActivationIsRecordedRatherThanQuarantined(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, seed := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	// One judgement below the first target: a canonical snapshot at no rung.
	below, belowRecovery := ladderSnapshotForState(t, seed, "obs-below", "70500", "70500", exitpolicy.NoRung)
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(below, belowRecovery)); err != nil {
		t.Fatalf("the below-target judgement was refused: %v", err)
	}
	if below.ActiveRung != exitpolicy.NoRung {
		t.Fatalf("fixture activated a rung too early: %d", below.ActiveRung)
	}

	// The crossing. 70000 * 1.018 = 71260, so 71500 activates rung 0.
	state, err := j.ExitState(ctx, position.ID)
	if err != nil {
		t.Fatal(err)
	}
	crossing, crossingRecovery := ladderSnapshotForState(t, state, "obs-crossing", "71500", "71500", exitpolicy.NoRung)
	if crossing.ActiveRung != 0 {
		t.Fatalf("fixture did not activate the first rung: rung=%d target=%s", crossing.ActiveRung, below.NextTarget)
	}
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(crossing, crossingRecovery)); err != nil {
		t.Fatalf("the first rung activation was refused: %v", err)
	}

	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if active {
		t.Fatal("activating the first rung quarantined the position")
	}
	after, err := j.ExitState(ctx, position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.ActiveRung != 0 {
		t.Fatalf("stored rung = %d, want the activated rung 0", after.ActiveRung)
	}
	if after.Baseline != crossing.CurrentProtection {
		t.Fatalf("stored protection = %s, want the raised %s", after.Baseline, crossing.CurrentProtection)
	}
}

// TestAGenuinelyAmbiguousJudgementIsStillQuarantined is task 3.2: the ledger's
// fail-closed behaviour is untouched where it was right.
func TestAGenuinelyAmbiguousJudgementIsStillQuarantined(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	_, seed := openedPosition(t, j, "10")

	high, highRecovery := ladderSnapshotForState(t, seed, "obs-high", "71500", "71500", exitpolicy.NoRung)
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(high, highRecovery)); err != nil {
		t.Fatalf("the first judgement was refused: %v", err)
	}

	// A candidate that is further along the rung axis but behind on protection:
	// the tuple no policy stage can produce, and the one this guard is for.
	crossed := high
	crossed.ActiveRung = high.ActiveRung + 1
	crossed.CurrentProtection = "68500"
	crossed.ObservationID = "obs_crossed_axes"
	err := j.RecordExitJudgement(ctx, judgementForSnapshot(crossed, highRecovery))
	if err == nil {
		t.Fatal("a crossed-axes candidate was accepted")
	}
	if !errors.Is(err, ErrExitSnapshotQuarantined) && !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("crossed-axes error = %v, want a quarantine or a refusal", err)
	}
}
