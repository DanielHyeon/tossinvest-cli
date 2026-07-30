package engine_test

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy(t *testing.T) {
	h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
		options.CommonPolicy = exitpolicy.CommonLadderHybrid50
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70000)
	h.observe()

	state := h.state(p.ID)
	if state.PolicyKind != journal.ExitPolicyLadder || state.PolicyID != exitpolicy.CommonLadderHybrid50 {
		t.Fatalf("opened state = kind:%s id:%s", state.PolicyKind, state.PolicyID)
	}
}

func TestStoredPolicyWinsOverLaterObserverCommonPolicy(t *testing.T) {
	h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
		options.CommonPolicy = exitpolicy.CommonLadderHybrid50
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		PolicyID: exitpolicy.CommonLadderBalanced, EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 71760) // BALANCED T2, below HYBRID_50 T2.
	h.observe()
	state := h.state(p.ID)
	if state.PolicyID != exitpolicy.CommonLadderBalanced || state.ActiveRung != 1 {
		t.Fatalf("state was rebound or judged with the wrong policy: %+v", state)
	}
}
