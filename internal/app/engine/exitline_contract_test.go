package engine_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestOneShareLadderPartialAdvancesProtectionWithoutAnyOrder(t *testing.T) {
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		opts.CommonPolicy = exitpolicy.CommonLadderBalanced
	})
	p := h.entry("005930", "1", "10000", "9800", "10000")
	h.quote("005930", 10260)

	cycle := h.observe()
	if cycle.Err != nil {
		t.Fatalf("ObserveOnce: %v", cycle.Err)
	}
	if cycle.Proposed != 0 || len(h.submit.places) != 0 {
		t.Fatalf("zero-share partial reached order path: cycle=%+v places=%d", cycle, len(h.submit.places))
	}
	state := h.state(p.ID)
	if state.ActiveRung != 1 || state.Baseline != "10100" || state.Pending() {
		t.Fatalf("state = %+v, want rung 1/protection 10100/no pending", state)
	}
}

func TestOneShareLadderFinalAndProtectionBreachEachSellExactlyOne(t *testing.T) {
	t.Run("final", func(t *testing.T) {
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
			opts.CommonPolicy = exitpolicy.CommonLadderBalanced
		})
		h.entry("005930", "1", "10000", "9800", "10000")
		h.quote("005930", 10700)
		cycle := h.observe()
		if cycle.Err != nil || cycle.Proposed != 1 || len(h.submit.places) != 1 {
			t.Fatalf("cycle=%+v places=%d", cycle, len(h.submit.places))
		}
		if got := h.submit.places[0].Intent.Quantity; got != 1 {
			t.Fatalf("final quantity = %v, want 1", got)
		}
	})

	t.Run("breach after zero partial", func(t *testing.T) {
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
			opts.CommonPolicy = exitpolicy.CommonLadderBalanced
		})
		h.entry("005930", "1", "10000", "9800", "10000")
		h.quote("005930", 10260)
		if cycle := h.observe(); cycle.Err != nil || cycle.Proposed != 0 {
			t.Fatalf("promotion cycle = %+v", cycle)
		}
		h.quote("005930", 10000)
		cycle := h.observe()
		if cycle.Err != nil || cycle.Proposed != 1 || len(h.submit.places) != 1 {
			t.Fatalf("breach cycle=%+v places=%d", cycle, len(h.submit.places))
		}
		if got := h.submit.places[0].Intent.Quantity; got != 1 {
			t.Fatalf("breach quantity = %v, want 1", got)
		}
	})
}

func TestOneShareRatchetPartialNeverArmsAZeroQuantityProposal(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "1", "10000", "9800", "10000")
	h.quote("005930", 10200) // +1R: the 40% partial projects to zero whole shares.

	cycle := h.observe()
	if cycle.Err != nil || cycle.Proposed != 0 || len(h.submit.places) != 0 {
		t.Fatalf("cycle=%+v places=%d", cycle, len(h.submit.places))
	}
	if state := h.state(p.ID); state.Pending() {
		t.Fatalf("zero quantity armed a journal proposal: %+v", state)
	}
}
