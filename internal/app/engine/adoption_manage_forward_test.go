package engine_test

// adoption_manage_forward_test.go is task 2.3's second half and design A2's
// honest statement.
//
// The first half — "the adoption transaction proposes nothing" — is in
// reconcileloop_test.go, where the adoption actually happens. This file is about
// what happens *next*, and it exists because the absolute claim is not available:
// the adoption observation and the first exit observation are different values
// taken at different instants from different reads, so a proposal on the first
// tick after adopting is possible and is **normal exit behaviour**.
//
// What has to be true is narrower and much more useful:
//
//   - the ratchet judges the first observation against the *adoption* t0 and
//     nothing else. The original cost basis is not in the comparison, so a
//     position 50 % under water against what it cost and a position 50 % up
//     behave identically once adopted at the same price;
//   - a first observation below the synthetic stop liquidates, exactly as it
//     would for a position the engine opened. That is the protection working,
//     not the adoption misfiring.

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// adopt folds a holding into the projection and adopts it at the given
// observation, returning the position.
//
// The cost basis is deliberately a parameter: every test below runs the same
// scenario at wildly different cost bases to show the number never enters the
// judgement.
func (h *exitHarness) adopt(symbol, quantity, costBasis, observed, stop string) journal.Position {
	h.t.Helper()
	ctx := context.Background()

	watermark, err := h.journal.FillWatermark(ctx, symbol)
	if err != nil {
		h.t.Fatal(err)
	}
	fold, err := h.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: exitAccount, Market: "kr", Symbol: symbol, Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: quantity, NewAvgPrice: costBasis,
		BrokerAsOf: h.clk.Now().Format("2006-01-02T15:04:05Z07:00"),
		Evidence:   "the account holds it and no local instance explains it",
	})
	if err != nil {
		h.t.Fatalf("folding %s in: %v", symbol, err)
	}
	if _, err := h.journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: fold.Position.ID, Symbol: symbol, Market: "kr", Quantity: quantity,
		CostBasis: costBasis, ObservedPrice: observed, SyntheticStop: stop,
		ObservedAt: journal.RFC3339(h.clk.Now()),
	}); err != nil {
		h.t.Fatalf("adopting %s: %v", symbol, err)
	}
	return fold.Position
}

// TestTheFirstObservationAfterAdoptionAppliesTheRatchetNormally is the P1 ≠ P0
// case: the price has moved between the adoption observation and the first exit
// observation, and the ratchet judges the new price against the adopted t0.
func TestTheFirstObservationAfterAdoptionAppliesTheRatchetNormally(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.adopt("005930", "10", "55000", "70000", "66500")

	// P1 below the synthetic stop. Risk per unit is 3500, so 66000 is a breach.
	h.prices.last["005930"] = 66000
	cycle := h.observer.ObserveOnce(context.Background())
	if cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	if cycle.Opened != 1 {
		t.Fatalf("opened = %d, want the adopted position's exit state", cycle.Opened)
	}
	if cycle.Unmanaged != 0 {
		t.Errorf("unmanaged = %d; an adopted position is managed", cycle.Unmanaged)
	}
	if cycle.Proposed != 1 {
		t.Fatalf("proposals = %d, want 1: a first observation below the synthetic stop is a breach, "+
			"and liquidating it is the protection working rather than the adoption misfiring",
			cycle.Proposed)
	}

	events, err := h.journal.ExitEvents(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Action == "" {
		t.Errorf("the breach judgement recorded no action: %+v", last)
	}
	if last.ObservedPrice != "66000" {
		t.Errorf("the judgement was made at %q, want the first exit observation", last.ObservedPrice)
	}
}

// TestTheCostBasisDoesNotChangeTheFirstJudgement is the manage-forward claim
// stated so it cannot pass by accident: two accounts identical except for what
// the shares originally cost produce the same cycle.
//
// The two bases are ±50 % of the adoption price, which is where a cost-anchored
// t0 does its damage: the +50 % holding would be instantly partial-taken as a
// winner, and the −50 % one instantly stopped out as a loser, on the first tick
// after adopting.
func TestTheCostBasisDoesNotChangeTheFirstJudgement(t *testing.T) {
	cases := map[string]string{
		"50% under water against cost": "140000",
		"50% up against cost":          "35000",
		"exactly at cost":              "70000",
		"no cost basis reported":       "",
	}
	for name, basis := range cases {
		t.Run(name, func(t *testing.T) {
			h := newExitHarness(t, nil)
			p := h.adopt("005930", "10", basis, "70000", "66500")

			// A first observation 1 % above the adoption price: +0.2R, which is
			// below every ratchet trigger and every partial.
			h.prices.last["005930"] = 70700
			cycle := h.observer.ObserveOnce(context.Background())
			if cycle.Err != nil {
				t.Fatalf("cycle: %v", cycle.Err)
			}
			if cycle.Proposed != 0 {
				t.Errorf("proposals = %d at +0.2R from the adoption price; the cost basis (%q) must "+
					"not be part of the judgement", cycle.Proposed, basis)
			}

			state, err := h.journal.ExitState(context.Background(), p.ID)
			if err != nil {
				t.Fatal(err)
			}
			if state.EntryPrice != "70000" || state.InitialRisk != "3500" {
				t.Errorf("t0 = %s/%s with a cost basis of %q; the adoption observation is the whole "+
					"of t0", state.EntryPrice, state.InitialRisk, basis)
			}
			if state.Baseline != "66500" {
				t.Errorf("baseline = %q, want the synthetic stop", state.Baseline)
			}
			// The watermark moved to the observation, which is the ratchet doing
			// its ordinary work from the adoption forward.
			if state.HighWater != "70700" {
				t.Errorf("high water = %q, want the observation", state.HighWater)
			}
		})
	}
}

// TestAnAdoptedWinnerRatchetsFromTheAdoptionPrice is design A2's "귀결의 명시"
// made concrete: the protection floor a long-held winner acquires is measured
// from the day it was adopted, not from what it cost.
//
// At +0.8R above the adoption price the baseline is promoted to the
// cost-inclusive break-even *of the adoption price*. A position bought at 35,000
// and adopted at 70,000 is therefore protected at roughly 70,000 — its historical
// gain is preserved and is not part of the R scale.
func TestAnAdoptedWinnerRatchetsFromTheAdoptionPrice(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.adopt("005930", "10", "35000", "70000", "66500")

	// +1.0R = 73,500. Enough to promote past BREAKEVEN.
	h.prices.last["005930"] = 73500
	if cycle := h.observer.ObserveOnce(context.Background()); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}

	state, err := h.journal.ExitState(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Baseline == "66500" {
		t.Fatal("the baseline never moved at +1.0R; the ratchet is not running on the adopted position")
	}
	// The promoted baseline is around the adoption price and nowhere near the
	// 35,000 the shares cost.
	if got := ratOfText(t, state.Baseline); got < 69000 || got > 72000 {
		t.Errorf("baseline = %s; it must be measured from the 70000 adoption price, not from the "+
			"35000 cost basis", state.Baseline)
	}
	if state.RatchetLevel == journal.RatchetNone {
		t.Errorf("ratchet level = %q at +1.0R, want a promotion", state.RatchetLevel)
	}
}

// TestAnAdoptedPositionIsNotReportedUnmanaged closes the loop with the exit
// observer's own enumeration: the single eligibility predicate is what it asks,
// so an adopted position is managed there too.
func TestAnAdoptedPositionIsNotReportedUnmanaged(t *testing.T) {
	h := newExitHarness(t, nil)
	h.adopt("005930", "10", "55000", "70000", "66500")

	h.prices.last["005930"] = 70000
	cycle := h.observer.ObserveOnce(context.Background())
	if cycle.Unmanaged != 0 {
		t.Errorf("unmanaged = %d for an adopted position", cycle.Unmanaged)
	}
	if cycle.Opened != 1 || cycle.Judged != 1 {
		t.Errorf("cycle = %+v, want the adopted position opened and judged", cycle)
	}
}

func ratOfText(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		t.Fatalf("parsing %q: %v", s, err)
	}
	return v
}
