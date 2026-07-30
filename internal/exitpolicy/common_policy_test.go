package exitpolicy_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles(t *testing.T) {
	t.Parallel()

	want := map[string]struct {
		targets, floors, partials []string
		finalFull                 bool
		recommended               bool
	}{
		exitpolicy.CommonLadderBalanced: {
			targets: []string{"1.5", "2.5", "4.0", "6.0"}, floors: []string{"0", "1.0", "2.0", "3.5"},
			partials: []string{"0", "0.25", "0.25", "1.0"}, finalFull: true,
		},
		exitpolicy.CommonLadderRunner: {
			targets: []string{"2.5", "4.5", "7.0", "999.0"}, floors: []string{"0", "2.0", "3.5", "5.0"},
			partials: []string{"0", "0.15", "0", "0"},
		},
		exitpolicy.CommonLadderHybrid50: {
			targets: []string{"1.8", "3.0", "4.8", "6.5"}, floors: []string{"0", "1.2", "2.5", "3.8"},
			partials: []string{"0", "0.25", "1/3", "0"}, recommended: true,
		},
	}

	got := exitpolicy.RegisteredCommonPolicies()
	if len(got) != len(want) {
		t.Fatalf("registered policies = %d, want exactly %d", len(got), len(want))
	}
	for _, profile := range got {
		expected, ok := want[profile.ID]
		if !ok {
			t.Fatalf("unexpected policy %q", profile.ID)
		}
		if profile.Recommended != expected.recommended || profile.Ladder.FinalTakeFull != expected.finalFull {
			t.Errorf("%s flags = recommended:%v final:%v", profile.ID, profile.Recommended, profile.Ladder.FinalTakeFull)
		}
		if len(profile.Ladder.Rungs) != 4 {
			t.Fatalf("%s rungs = %d, want 4", profile.ID, len(profile.Ladder.Rungs))
		}
		for i, rung := range profile.Ladder.Rungs {
			if rung.TargetPct != expected.targets[i] || rung.StopPct != expected.floors[i] ||
				rung.PartialRatio != expected.partials[i] {
				t.Errorf("%s rung %d = %+v, want target=%s floor=%s partial=%s",
					profile.ID, i, rung, expected.targets[i], expected.floors[i], expected.partials[i])
			}
		}
	}
}

func TestCommonPolicyRegistryReturnsCopies(t *testing.T) {
	t.Parallel()

	first, ok := exitpolicy.CommonPolicyByID(exitpolicy.CommonLadderHybrid50)
	if !ok {
		t.Fatal("HYBRID_50 is not registered")
	}
	first.Ladder.Rungs[0].TargetPct = "999"
	second, _ := exitpolicy.CommonPolicyByID(exitpolicy.CommonLadderHybrid50)
	if second.Ladder.Rungs[0].TargetPct != "1.8" {
		t.Fatalf("registry was mutated through returned slice: %+v", second.Ladder.Rungs[0])
	}
}

func TestExternalRunnerHasNoAutomaticPartials(t *testing.T) {
	t.Parallel()

	policy, err := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderRunner, true)
	if err != nil {
		t.Fatalf("CommonLadderForPosition: %v", err)
	}
	for i, rung := range policy.Rungs {
		if rung.PartialRatio != "0" {
			t.Errorf("adopted RUNNER rung %d partial = %s, want 0", i, rung.PartialRatio)
		}
	}
}

func TestHybrid50T4RaisesTheRunnerFloorWithoutTakingTheRemainder(t *testing.T) {
	t.Parallel()

	policy, err := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderHybrid50, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := exitpolicy.EvaluateLadder(exitpolicy.LadderInput{
		EntryPrice: "10000", ObservedPrice: "11000", HighWater: "10650", Baseline: "10380",
		Policy: policy,
		State: exitpolicy.LadderState{
			PolicyID: policy.PolicyID, ActivatedRung: 3, TakenRatioTotal: "0.5",
			PendingRung: exitpolicy.NoRung,
		},
	})
	if err != nil {
		t.Fatalf("EvaluateLadder: %v", err)
	}
	if got.Baseline != "10380" {
		// 11000 × 0.935 = 10285, so the previous fixed 10380 remains stronger.
		t.Fatalf("baseline = %s, want monotone max 10380", got.Baseline)
	}
	if !got.Proposal.Zero() {
		t.Fatalf("T4 runner made a fixed exit proposal: %+v", got.Proposal)
	}

	got, err = exitpolicy.EvaluateLadder(exitpolicy.LadderInput{
		EntryPrice: "10000", ObservedPrice: "11500", HighWater: "11000", Baseline: "10380",
		Policy: policy,
		State: exitpolicy.LadderState{
			PolicyID: policy.PolicyID, ActivatedRung: 3, TakenRatioTotal: "0.5",
			PendingRung: exitpolicy.NoRung,
		},
	})
	if err != nil {
		t.Fatalf("EvaluateLadder at new high: %v", err)
	}
	if got.Baseline != "10752.5" {
		t.Fatalf("runner baseline = %s, want 11500 × 0.935 = 10752.5", got.Baseline)
	}
	if !got.Proposal.Zero() {
		t.Fatalf("runner high made an exit proposal: %+v", got.Proposal)
	}
}

func TestHybrid50RunnerBreachOutranksAnyPromotion(t *testing.T) {
	t.Parallel()

	policy, _ := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderHybrid50, false)
	got, err := exitpolicy.EvaluateLadder(exitpolicy.LadderInput{
		EntryPrice: "10000", ObservedPrice: "10400", HighWater: "11500", Baseline: "10380",
		Policy: policy,
		State: exitpolicy.LadderState{
			PolicyID: policy.PolicyID, ActivatedRung: 3, TakenRatioTotal: "0.5",
			PendingAction: exitpolicy.ActionLadderPartial, PendingRung: 2,
		},
	})
	if err != nil {
		t.Fatalf("EvaluateLadder: %v", err)
	}
	if got.Action != exitpolicy.ActionLadderStop || got.Proposal.Ratio != "1" || !got.CancelPendingFirst {
		t.Fatalf("breach = action:%s proposal:%+v cancel-first:%v", got.Action, got.Proposal, got.CancelPendingFirst)
	}
}
