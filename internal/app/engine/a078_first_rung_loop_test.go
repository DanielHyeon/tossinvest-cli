package engine_test

// a078 end to end: crossing the first take-profit line must not end a position's
// judgement.
//
// The unit and ledger tests pin the comparison and the write. This one pins the
// consequence the operator actually lives with — that the loop keeps judging the
// position on the cycle after the crossing, so its stop stays under evaluation.
// On 2026-08-03 it did not: one crossing quarantined the position and the engine
// skipped it from then on.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement(t *testing.T) {
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

	// Below the first target: a canonical snapshot is persisted at no rung. This
	// is the state every ladder holding sits in before it earns anything.
	h.quote("005930", 70500)
	h.observe()
	if before := h.state(p.ID); before.ActiveRung != exitpolicy.NoRung {
		t.Fatalf("fixture activated a rung before the crossing: %d", before.ActiveRung)
	}

	// The crossing.
	h.quote("005930", 71500)
	h.observe()

	if _, active, err := h.journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if active {
		t.Fatal("crossing the first take-profit line quarantined the position")
	}
	if alert, ok := h.alerts.first(obs.EventExitJudgementRefused); ok {
		t.Fatalf("the loop refused to judge a healthy position: %s", alert.Body)
	}
	after := h.state(p.ID)
	if after.ActiveRung != 0 {
		t.Fatalf("stored rung = %d, want the first rung activated", after.ActiveRung)
	}

	// The cycle after the crossing is the one that matters: a quarantined
	// position is skipped from here on, and its stop stops being evaluated.
	h.quote("005930", 71800)
	cycle := h.observe()
	if cycle.Judged == 0 {
		t.Fatalf("the position was not judged after activating its first rung: %+v", cycle)
	}
	if got := h.alerts.count(obs.EventExitJudgementRefused); got != 0 {
		t.Errorf("judgement refusals after the crossing = %d, want none", got)
	}
	if raised := h.state(p.ID); raised.HighWater != "71800" {
		t.Errorf("high water = %s, want the loop still tracking it at 71800", raised.HighWater)
	}
}
