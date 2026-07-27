package risk

import (
	"math/big"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// netrr_inversion_test.go is change add-net-rr-measurement task 6.6 (round 3's P0).
//
// # The trap this file exists to keep visible
//
// "A net threshold of 1.5 is obviously stricter than a gross threshold of 2.0,
// because costs only make the ratio worse" is false, and it is false in the
// direction that matters: at wide stops the net rule is *looser*. A change that
// switched the basis on that intuition would relax the gate while believing it had
// tightened it, which is a §0.9 violation nobody would notice.
//
// # The algebra
//
// Write the entry as 1, the stop width as s (so the stop is 1 − s), and the
// break-even as c (so B = c, with c > 1 whenever any rate is positive). Then
//
//	the gross rule needs   target ≥ 1 + R·s
//	the net rule needs     target ≥ c + r·(c − 1 + s)
//
// Setting them equal and solving for s:
//
//	1 + R·s = c + r·c − r + r·s
//	s(R − r) = (c − 1) + r(c − 1) = (1 + r)(c − 1)
//	s*       = (1 + r)(c − 1) / (R − r)
//
// Below s* the net rule demands more; above it, less. At r = 1.5 and R = 2.0 the
// denominator is 0.5, so s* = 5(c − 1) — five times the round-trip cost ratio.
//
// ⚠️ The third edition of this change's tasks carried `((1+r)c − 1 − R)/(r − R)`
// here. That is wrong (it gives 97.49% for KR) and must not be reintroduced.
//
// # Why this is a test and not a comment
//
// The crossover is a function of the cost model, and every rate in that model is
// an unmeasured over-estimate. When 2b replaces them the crossover moves — downward,
// widening the range in which a net threshold relaxes the gate — and the specs that
// quote 2.51% and 10.62% become wrong. Asserting the numbers here means the rate
// measurement cannot land without this failing and forcing those quotes to be
// updated.

// breakEvenRatio is c: the break-even as a multiple of the entry, from the live
// model rather than from a rate arithmetic re-derived here. Re-deriving it would
// test the re-derivation.
func breakEvenRatio(t *testing.T, model costs.Model, market costs.Market) *big.Rat {
	t.Helper()
	// Entry 1 makes B numerically equal to c. The cost model is scale-free in the
	// rates, so any entry would do; 1 removes a division.
	breakEven, err := NetBreakEven(model, market, "1")
	if err != nil {
		t.Fatalf("break-even for %s: %v", market, err)
	}
	return breakEven
}

// crossoverStopWidth is s* = (1 + r)(c − 1) / (R − r).
func crossoverStopWidth(c, r, R *big.Rat) *big.Rat {
	numerator := new(big.Rat).Mul(
		new(big.Rat).Add(big.NewRat(1, 1), r),
		new(big.Rat).Sub(c, big.NewRat(1, 1)))
	return new(big.Rat).Quo(numerator, new(big.Rat).Sub(R, r))
}

// TestTheCrossoverMatchesTheSpecsQuotedNumbers pins the two values the spec
// carries, computed from internal/costs rather than transcribed.
func TestTheCrossoverMatchesTheSpecsQuotedNumbers(t *testing.T) {
	model := costs.DefaultModel()
	r := big.NewRat(3, 2) // net 1.5
	R := big.NewRat(2, 1) // gross 2.0

	for _, tc := range []struct {
		market costs.Market
		wantC  string // c to seven decimals
		wantS  string // s* as a percentage, four decimals
	}{
		{costs.MarketKR, "1.0050201", "2.5100"},
		{costs.MarketUS, "1.0212336", "10.6168"},
	} {
		t.Run(string(tc.market), func(t *testing.T) {
			c := breakEvenRatio(t, model, tc.market)
			if got := c.FloatString(7); got != tc.wantC {
				t.Errorf("break-even ratio c = %s, want %s. The rates changed; every "+
					"crossover number in risk-management's spec is now stale", got, tc.wantC)
			}
			s := crossoverStopWidth(c, r, R)
			percent := new(big.Rat).Mul(s, big.NewRat(100, 1))
			if got := percent.FloatString(4); got != tc.wantS {
				t.Errorf("crossover stop width s* = %s%%, want %s%%", got, tc.wantS)
			}
		})
	}
}

// TestTheShortFormHolds: at r = 1.5 and R = 2.0 the denominator is 0.5, so
// s* = 5(c − 1). The spec states the short form and it is worth pinning, because
// the short form is what a reader will do arithmetic with.
func TestTheShortFormHolds(t *testing.T) {
	model := costs.DefaultModel()
	r, R := big.NewRat(3, 2), big.NewRat(2, 1)

	for _, market := range []costs.Market{costs.MarketKR, costs.MarketUS} {
		c := breakEvenRatio(t, model, market)
		full := crossoverStopWidth(c, r, R)
		short := new(big.Rat).Mul(big.NewRat(5, 1), new(big.Rat).Sub(c, big.NewRat(1, 1)))
		if full.Cmp(short) != 0 {
			t.Errorf("%s: (1+r)(c−1)/(R−r) = %s but 5(c−1) = %s",
				market, full.FloatString(10), short.FloatString(10))
		}
	}
}

// TestTheDiscardedFormulaIsStillWrong keeps the third edition's error identifiable.
// Somebody re-deriving this will produce candidate expressions; this one looks
// plausible and gives 97.49% for KR, which would read as "the inversion is
// unreachable in practice" — the opposite of the truth.
func TestTheDiscardedFormulaIsStillWrong(t *testing.T) {
	model := costs.DefaultModel()
	c := breakEvenRatio(t, model, costs.MarketKR)
	r, R := big.NewRat(3, 2), big.NewRat(2, 1)

	// ((1+r)c − 1 − R) / (r − R)
	wrong := new(big.Rat).Quo(
		new(big.Rat).Sub(
			new(big.Rat).Sub(new(big.Rat).Mul(new(big.Rat).Add(big.NewRat(1, 1), r), c), big.NewRat(1, 1)),
			R),
		new(big.Rat).Sub(r, R))
	correct := crossoverStopWidth(c, r, R)

	if wrong.Cmp(correct) == 0 {
		t.Fatal("the two formulas agree, so this guard has stopped guarding anything")
	}
	if got := new(big.Rat).Mul(wrong, big.NewRat(100, 1)).FloatString(2); got != "97.49" {
		t.Errorf("the discarded formula gives %s%% for KR; the recorded value was 97.49%%", got)
	}
}

// TestTheInversionIsRealAtWideStops is the claim itself, verified against the
// chain's own arithmetic rather than against the algebra that predicts it. Below
// the crossover the net rule demands a higher target; above it, a lower one.
func TestTheInversionIsRealAtWideStops(t *testing.T) {
	model := costs.DefaultModel()
	const entry = "10000"
	r, R := big.NewRat(3, 2), big.NewRat(2, 1)

	c := breakEvenRatio(t, model, costs.MarketKR)
	crossover := crossoverStopWidth(c, r, R) // ≈ 2.51%

	// A stop wider than the crossover: the spec's worked example is −5%.
	wide := targetsAt(t, model, entry, big.NewRat(5, 100), r, R)
	if wide.net.Cmp(wide.gross) >= 0 {
		t.Errorf("at a 5%% stop the net-1.5 target %s is not below the gross-2.0 target %s; "+
			"the inversion the spec warns about is not reproducing",
			wide.net.FloatString(4), wide.gross.FloatString(4))
	}
	// The spec's numbers, as percentages above the entry: net +8.755%, gross +10%.
	if got := percentAbove(wide.gross, entry).FloatString(3); got != "10.000" {
		t.Errorf("gross-2.0 target at a 5%% stop = +%s%%, want +10.000%%", got)
	}
	if got := percentAbove(wide.net, entry).FloatString(3); got != "8.755" {
		t.Errorf("net-1.5 target at a 5%% stop = +%s%%, want +8.755%%", got)
	}
	// And that looser target is one today's gate refuses: its gross ratio is 1.751.
	grossOfNetTarget, err := RewardRisk(entry, "9500", wide.net.FloatString(8))
	if err != nil {
		t.Fatal(err)
	}
	if got := grossOfNetTarget.FloatString(3); got != "1.751" {
		t.Errorf("the net-1.5 target's gross ratio = %s, want 1.751", got)
	}
	if grossOfNetTarget.Cmp(R) >= 0 {
		t.Error("that target would pass the current gate, which would make it no relaxation")
	}

	// A stop narrower than the crossover: the ordering flips back.
	narrow := targetsAt(t, model, entry, big.NewRat(1, 100), r, R)
	if narrow.net.Cmp(narrow.gross) <= 0 {
		t.Errorf("at a 1%% stop (under the %s%% crossover) the net rule must be the stricter "+
			"one: net %s, gross %s", new(big.Rat).Mul(crossover, big.NewRat(100, 1)).FloatString(2),
			narrow.net.FloatString(4), narrow.gross.FloatString(4))
	}
}

type requiredTargets struct{ gross, net *big.Rat }

// targetsAt returns the minimum target each rule demands for one stop width.
func targetsAt(t *testing.T, model costs.Model, entry string, width, r, R *big.Rat) requiredTargets {
	t.Helper()
	e, ok := new(big.Rat).SetString(entry)
	if !ok {
		t.Fatalf("entry %q", entry)
	}
	risked := new(big.Rat).Mul(e, width)
	stop := new(big.Rat).Sub(e, risked)

	// gross: target = entry + R × (entry − stop)
	gross := new(big.Rat).Add(e, new(big.Rat).Mul(R, risked))

	// net: target = B + r × (B − stop)
	b, err := NetBreakEven(model, costs.MarketKR, entry)
	if err != nil {
		t.Fatal(err)
	}
	net := new(big.Rat).Add(b, new(big.Rat).Mul(r, new(big.Rat).Sub(b, stop)))
	return requiredTargets{gross: gross, net: net}
}

func percentAbove(target *big.Rat, entry string) *big.Rat {
	e, _ := new(big.Rat).SetString(entry)
	return new(big.Rat).Mul(
		new(big.Rat).Quo(new(big.Rat).Sub(target, e), e),
		big.NewRat(100, 1))
}

// TestAnyNetThresholdBelowGrossInvertsSomewhere is the general consequence, and
// the one a promoting change has to answer: for every r < R the crossover is
// finite and positive, so there is always a stop width past which the net rule is
// looser.
//
// Whether that width is *reachable* is a second question, and the test keeps the
// two apart. TossOS has no upper bound on stop width, but a stop is below zero
// past 100%, so a crossover above 100% is not admissible input. The candidates
// actually under discussion — 1.3 and 1.5 — invert at 1.6% and 2.5%, which is
// inside the range StockOS was trading at 0.70%.
//
// This leaves exactly two safe forms: a net threshold at or above the gross one,
// or gross kept with net added as a conjunction. The first is safe because the
// crossover formula has no solution at r ≥ R at all — the test states that as the
// division being undefined rather than as a number.
func TestAnyNetThresholdBelowGrossInvertsSomewhere(t *testing.T) {
	model := costs.DefaultModel()
	R := big.NewRat(2, 1)
	c := breakEvenRatio(t, model, costs.MarketKR)
	hundredPercent := big.NewRat(1, 1)

	reachable := 0
	for _, candidate := range []string{"1.3", "1.5", "1.9", "1.99"} {
		r, ok := new(big.Rat).SetString(candidate)
		if !ok {
			t.Fatalf("candidate %q", candidate)
		}
		s := crossoverStopWidth(c, r, R)
		if s.Sign() <= 0 {
			t.Errorf("candidate %s has a non-positive crossover %s, which would mean the "+
				"net rule is never stricter", candidate, s.FloatString(6))
			continue
		}
		percent := new(big.Rat).Mul(s, big.NewRat(100, 1)).FloatString(4)
		if s.Cmp(hundredPercent) >= 0 {
			t.Logf("net %s inverts only above a %s%% stop, which is not admissible input "+
				"(the stop would be at or below zero)", candidate, percent)
			continue
		}
		reachable++
		t.Logf("net %s inverts above a %s%% stop — admissible, so this candidate relaxes "+
			"the gate on real geometry", candidate, percent)
	}
	if reachable == 0 {
		t.Fatal("no candidate inverted within an admissible stop width; the whole warning " +
			"in risk-management would be moot and this test would be the place to notice")
	}

	// The safe form. At r = R the denominator is zero: there is no crossover, so
	// there is no width at which the net rule becomes looser. Asserting the
	// division is undefined is the honest statement of "never inverts" — computing
	// it would panic, which is what the recover here shows.
	func() {
		defer func() {
			if recover() == nil {
				t.Error("r = R produced a finite crossover; the division by (R − r) must be " +
					"undefined, which is precisely why a net threshold at or above the gross " +
					"one cannot invert")
			}
		}()
		_ = crossoverStopWidth(c, R, R)
	}()
}
