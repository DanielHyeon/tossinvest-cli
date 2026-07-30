package verifylive

// pricing_trigger_test.go is the other half of pricing_test.go, and it is testing
// the opposite property.
//
// Everything in pricing_test.go proves an order cannot fill. This file proves one
// deliberately can: the trigger observation needs a stop that the market will
// reach, because the only way to measure what happens after a conditional order
// fires is to make one fire. The cases are therefore the ones that would silently
// produce a *useless* measurement — a trigger that fires under both readings of
// the broker's rule at once, a trigger below the bid that fires under neither, and
// a grid with no room between the two.

import (
	"math"
	"strings"
	"testing"
)

// TestNearStopTriggerSitsBetweenTheBidAndTheLastTrade is the placement rule.
//
// A stop sell fires when the price falls to the trigger. Between the best bid and
// the last trade is the one interval where the two candidate readings of "the
// price" disagree *right now*:
//
//	broker reads the bid          bid <= trigger already, so it fires at once
//	broker reads the last trade   it fires when a trade prints at or below trigger
//
// That disagreement is the measurement. Anywhere else on the grid both readings
// answer the same thing and the run learns nothing about which one the broker uses.
func TestNearStopTriggerSitsBetweenTheBidAndTheLastTrade(t *testing.T) {
	got, err := NearStopTrigger(299.94, 299.88, MarketUS)
	if err != nil {
		t.Fatalf("NearStopTrigger: %v", err)
	}
	if !(got.Price > 299.88 && got.Price <= 299.94) {
		t.Fatalf("trigger %v is not inside (bid 299.88, last 299.94]", got.Price)
	}
	if got.Price != 299.89 {
		t.Errorf("trigger = %v, want 299.89 — the first tick above the bid, which puts the longest "+
			"observable delay between a bid-basis fire and a last-basis one", got.Price)
	}
	if math.Abs(math.Mod(math.Round(got.Price*10000), 100)) > 1e-9 {
		t.Errorf("trigger %v is not on the US cent grid", got.Price)
	}
	if got.Clamped {
		t.Error("a trigger inside the spread was reported as clamped by a daily band")
	}
	if !strings.Contains(got.Basis, "299.88") || !strings.Contains(got.Basis, "299.94") {
		t.Errorf("basis %q does not name both ends of the interval it was chosen from", got.Basis)
	}
}

// TestNearStopTriggerRefusesRatherThanGuess. There is no safe fallback: a trigger
// this function invents outside the interval measures the wrong thing, and a
// trigger it invents *below* the bid is a live stop that never fires and has to be
// cleaned up. Every one of these has to come back as an error so the step skips.
func TestNearStopTriggerRefusesRatherThanGuess(t *testing.T) {
	cases := []struct {
		name       string
		last, bid  float64
		market     string
		wantReason string
	}{
		{"no last trade", 0, 299.88, MarketUS, "last"},
		{"negative last", -1, 299.88, MarketUS, "last"},
		{"no bid", 299.94, 0, MarketUS, "bid"},
		{"negative bid", 299.94, -1, MarketUS, "bid"},
		{"bid at the last trade", 299.94, 299.94, MarketUS, "no room"},
		{"crossed book", 299.88, 299.94, MarketUS, "no room"},
		// A US trade can print inside the quoted spread (measurements.md M49), so a
		// last trade less than one tick above the bid is a real shape, not a
		// contrivance — and there is no grid point to put a trigger on.
		{"no tick fits", 299.885, 299.88, MarketUS, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NearStopTrigger(c.last, c.bid, c.market)
			if err == nil {
				t.Fatalf("NearStopTrigger(%v, %v) = %v with no error; it must refuse rather than "+
					"invent a trigger price", c.last, c.bid, got.Price)
			}
			if got.Price != 0 {
				t.Errorf("a refused trigger still carried a price %v", got.Price)
			}
		})
	}
}

// TestNearStopTriggerAtTheLastTradeIsAllowedAndOnlyJust.
//
// One tick above the bid can land exactly on the last trade — the KR grid is
// coarse enough that a 5원 tick swallows the whole spread. That trigger fires
// immediately under both readings, so it settles nothing about the basis, but it
// still fires: the primary measurement (a conditional order produces a child
// order, and how long that takes to become visible) is intact. Refusing here
// would throw away the measurement to protect a secondary one.
func TestNearStopTriggerAtTheLastTradeIsAllowedAndOnlyJust(t *testing.T) {
	got, err := NearStopTrigger(3295, 3290, MarketKR)
	if err != nil {
		t.Fatalf("NearStopTrigger: %v", err)
	}
	if got.Price != 3295 {
		t.Fatalf("trigger = %v, want 3295 — one 5원 tick above the 3290 bid", got.Price)
	}
	if got.Distinguishes {
		t.Error("a trigger sitting on the last trade cannot tell a bid basis from a last-trade basis, " +
			"and the record must not claim it can")
	}

	spread, err := NearStopTrigger(3300, 3290, MarketKR)
	if err != nil {
		t.Fatalf("NearStopTrigger: %v", err)
	}
	if spread.Price != 3295 || !spread.Distinguishes {
		t.Errorf("trigger = %+v, want 3295 strictly below the last trade and distinguishing", spread)
	}
}

// TestNearStopTriggerCrossesTheUSDollarTickBoundary. Below a dollar US ticks at a
// hundredth of a cent and at or above it at a cent, and a bid on one side of that
// line with a last trade on the other is where an off-grid price would be born.
func TestNearStopTriggerCrossesTheUSDollarTickBoundary(t *testing.T) {
	got, err := NearStopTrigger(1.02, 0.9999, MarketUS)
	if err != nil {
		t.Fatalf("NearStopTrigger: %v", err)
	}
	if got.Price <= 0.9999 || got.Price > 1.02 {
		t.Fatalf("trigger %v left the interval (0.9999, 1.02]", got.Price)
	}
	if got.Price != 1.00 {
		t.Errorf("trigger = %v, want 1.00 — the first price above the bid that is on a valid grid", got.Price)
	}

	// Wholly below a dollar, the fine grid applies.
	sub, err := NearStopTrigger(0.3965, 0.38, MarketUS)
	if err != nil {
		t.Fatalf("NearStopTrigger: %v", err)
	}
	if sub.Price != 0.3801 {
		t.Errorf("trigger = %v, want 0.3801 — one hundredth of a cent above the bid", sub.Price)
	}
}

// TestNearStopTriggerIsTheOnlyFunctionThatAimsAtTheMarket.
//
// The Far* family is the safety arithmetic every other step in this package rests
// on, and its whole claim is that it takes no argument that could point a price at
// the market. Adding a direction flag to one of them would put twelve steps'
// worth of "this cannot fill" behind a boolean — so the separation is asserted in
// the source rather than trusted to review.
func TestNearStopTriggerIsTheOnlyFunctionThatAimsAtTheMarket(t *testing.T) {
	for _, name := range []string{"FarBuyLimit", "FarSellLimit", "FarStopTrigger"} {
		params := functionParams(t, "pricing.go", name)
		want := []string{"last", "lowerLimit", "offset", "market"}
		if name == "FarSellLimit" {
			want = []string{"last", "upperLimit", "offset", "market"}
		}
		if strings.Join(params, ",") != strings.Join(want, ",") {
			t.Errorf("%s takes (%s); it must keep taking exactly (%s). A near/far or direction argument "+
				"here would put every step's safety arithmetic behind one branch",
				name, strings.Join(params, ", "), strings.Join(want, ", "))
		}
	}

	// And nothing outside the trigger step may call the near one.
	allowed := map[string]bool{"pricing.go": true, "steps_trigger.go": true}
	for file, src := range packageFiles(t, false) {
		if allowed[file] {
			continue
		}
		if callsFunction(t, file, src, "NearStopTrigger") {
			t.Errorf("%s calls NearStopTrigger. The one place this tool prices an order it intends to "+
				"fill is the opted-in trigger step", file)
		}
	}
}
