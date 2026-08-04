package engine_test

// a079 task 7: what a release actually buys, and what it does not.
//
// These two tests are the whole safety argument stated as behaviour. Releasing
// returns a position to judgement — which is the point, because a quarantined
// position has no stop evaluation at all. And releasing a position whose cause
// still holds gets it quarantined again on the very next observation, which is
// why the release cannot be used to suppress anything.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestAReleasedPositionIsJudgedAgainOnTheNextObservation(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	ctx := context.Background()
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	// The state the three live positions were in on 2026-08-03: quarantined, and
	// therefore not judged at all.
	if _, err := h.journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	// The observable difference is not the cycle counter — quarantined rows stay
	// in the working set so the loop can raise their refusal alert — it is that
	// nothing about the position is evaluated or advanced.
	h.quote("005930", 70800)
	h.observe()
	if _, ok := h.alerts.first(obs.EventExitJudgementRefused); !ok {
		t.Fatal("a quarantined position was judged without a refusal alert")
	}
	if held := h.state(p.ID); held.HighWater != "70500" {
		t.Fatalf("high water = %s, want it frozen at 70500 while quarantined", held.HighWater)
	}

	if err := h.journal.ReleaseExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq, 1,
		journal.QuarantineReleaseHumanRepair, "LOCAL_OPERATOR released quarantine v1"); err != nil {
		t.Fatalf("ReleaseExitSnapshotQuarantine: %v", err)
	}

	h.quote("005930", 70900)
	cycle := h.observe()

	if cycle.Judged == 0 {
		t.Fatalf("a released position was still not judged: %+v", cycle)
	}
	if resumed := h.state(p.ID); resumed.HighWater != "70900" {
		t.Errorf("high water = %s, want the loop tracking it again at 70900", resumed.HighWater)
	}
	// The whole point of preferring release over re-adoption: the baseline the
	// position was opened with is still the baseline it is judged against.
	if kept := h.state(p.ID); kept.EntryPrice != "70000" || kept.InitialStop != "68000" {
		t.Errorf("release rewrote the baseline: entry=%s stop=%s, want 70000/68000",
			kept.EntryPrice, kept.InitialStop)
	}
}

// The other half of the claim — that a release cannot suppress a cause that
// still holds — is pinned at the ledger boundary where the quarantine is
// actually written: journal.TestAReleasedQuarantineComesBackWhenTheCauseHolds.
// It belongs there because reproducing a live re-quarantine means driving the
// exact judgement that writes one, which is a record() call, not a quote.
