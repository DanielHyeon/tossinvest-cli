package risk

import (
	"math/big"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// netrr_test.go is change add-net-rr-measurement tasks 3.1, 3.2, 3.5, 3.6 and 3.7.
//
// The values here are TossOS's own. Where a StockOS post-mortem number is nearby
// it is named as a neighbour, never as a target — the two formulas differ and
// pretending otherwise would turn a coincidence in the second decimal into an
// evidence claim.

// stockosRates is the rate set StockOS's 058 post-mortem used: 0.015% commission
// on each leg and a 0.20% sell-side transaction tax, 0.23% round trip.
//
// It is *not* TossOS's DefaultModel, whose seven rates are all `[미검증]`
// over-estimates. It appears here because task 3.5 pins a value against it and
// because it is the only rate set with a published worked example to sit beside.
func stockosRates(t *testing.T) costs.Model {
	t.Helper()
	model, err := costs.NewModel(map[string]string{
		costs.KeyKRBuyCommissionRate:     "0.00015",
		costs.KeyKRSellCommissionRate:    "0.00015",
		costs.KeyKRSellTaxRate:           "0.0020",
		costs.KeyUSBuyCommissionRate:     "0.00015",
		costs.KeyUSSellCommissionRate:    "0.00015",
		costs.KeyUSSellRegulatoryFeeRate: "0",
		costs.KeyUSFXConversionFeeRate:   "0",
	})
	if err != nil {
		t.Fatalf("building the StockOS rate set: %v", err)
	}
	return model
}

// zeroRates is a model an operator configured to charge nothing. Configured, so
// it is a measurement somebody made — unlike the zero value, which is the absence
// of one and is refused everywhere.
func zeroRates(t *testing.T) costs.Model {
	t.Helper()
	model, err := costs.NewModel(map[string]string{
		costs.KeyKRBuyCommissionRate:     "0",
		costs.KeyKRSellCommissionRate:    "0",
		costs.KeyKRSellTaxRate:           "0",
		costs.KeyUSBuyCommissionRate:     "0",
		costs.KeyUSSellCommissionRate:    "0",
		costs.KeyUSSellRegulatoryFeeRate: "0",
		costs.KeyUSFXConversionFeeRate:   "0",
	})
	if err != nil {
		t.Fatalf("building the zero rate set: %v", err)
	}
	return model
}

func mustNetRR(t *testing.T, m costs.Model, market costs.Market, entry, stop, target string) *big.Rat {
	t.Helper()
	got, err := NetRewardRisk(m, market, entry, stop, target)
	if err != nil {
		t.Fatalf("NetRewardRisk(%s, %s, %s): %v", entry, stop, target, err)
	}
	return got
}

// TestNetRewardRiskIsBelowGrossUnderRealRates is the ordinary case the whole
// change is about: a trade that clears the 2.0 gross gate offers materially less
// once the round trip is paid for.
func TestNetRewardRiskIsBelowGrossUnderRealRates(t *testing.T) {
	model := costs.DefaultModel()

	gross, err := RewardRisk("10000", "9800", "10400")
	if err != nil {
		t.Fatal(err)
	}
	net := mustNetRR(t, model, costs.MarketKR, "10000", "9800", "10400")

	if gross.Cmp(big.NewRat(2, 1)) != 0 {
		t.Fatalf("the fixture must sit exactly on the gate: gross = %s", trimRatio(gross))
	}
	if net.Cmp(gross) >= 0 {
		t.Errorf("net %s is not below gross %s under non-zero rates",
			trimRatio(net), trimRatio(gross))
	}
	if net.Cmp(big.NewRat(2, 1)) >= 0 {
		t.Errorf("net = %s; an intent exactly on the gross gate must fall under it on a net basis",
			trimRatio(net))
	}
}

// TestNetRewardRiskUnderTheStockOSRateSet is task 3.5.
//
// The geometry is 058's: entry 100, stop 99.30, target 101.05 — gross exactly
// 1.50. TossOS computes 0.8807183…
//
// # This is not a reproduction of 058's 0.88, and must not be cited as one
//
// 058 subtracted the round-trip cost from the reward and added it to the risk:
// 0.82 / 0.93 = 0.8817204…. TossOS divides by (1 − sell-side rate) instead,
// because the sale is itself taxed and adding the sell-side cost rather than
// grossing up by its complement understates the break-even (costs.go's own
// reasoning). The two agree to two decimals and diverge at the third:
//
//	058     0.8817204…
//	TossOS  0.8807183…
//
// Same rates, same prices, different formulas. Reporting the match as evidence
// would be reporting an arithmetic coincidence as a validation.
func TestNetRewardRiskUnderTheStockOSRateSet(t *testing.T) {
	model := stockosRates(t)

	gross, err := RewardRisk("100", "99.30", "101.05")
	if err != nil {
		t.Fatal(err)
	}
	if gross.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("gross = %s, want exactly 1.5", trimRatio(gross))
	}

	breakEven, err := NetBreakEven(model, costs.MarketKR, "100")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := breakEven.FloatString(6), "100.230496"; got != want {
		t.Errorf("break-even = %s, want %s", got, want)
	}

	net := mustNetRR(t, model, costs.MarketKR, "100", "99.30", "101.05")
	if got, want := net.FloatString(7), "0.8807183"; got != want {
		t.Errorf("net = %s, want %s", got, want)
	}

	// The neighbour, spelled out so nobody has to rediscover why they differ.
	const stockos058 = "0.8817204"
	if net.FloatString(7) == stockos058 {
		t.Error("the two formulas are not the same and must not be made to agree; " +
			"see this test's comment")
	}
}

// TestNetRewardRiskInTheUSMarket is task 3.5's other market. The US rate set is
// heavier — the FX conversion fee applies to both legs — so the same geometry
// loses more.
func TestNetRewardRiskInTheUSMarket(t *testing.T) {
	model := costs.DefaultModel()

	kr := mustNetRR(t, model, costs.MarketKR, "100", "99.30", "101.05")
	us, err := NetRewardRisk(model, costs.MarketUS, "100", "99.30", "101.05")

	// Under the default rates the US round trip is wide enough that the
	// break-even lands above the target *and* above the stop is still true, so
	// the ratio exists and is negative — a trade that cannot make money.
	if err != nil {
		t.Fatalf("the US ratio must be computable for this geometry: %v", err)
	}
	if us.Sign() >= 0 {
		t.Errorf("US net = %s; the default US rates put break-even above this target, "+
			"so the ratio is negative", trimRatio(us))
	}
	if us.Cmp(kr) >= 0 {
		t.Errorf("US net %s is not below KR net %s under the heavier US rate set",
			trimRatio(us), trimRatio(kr))
	}
}

// TestTheBreakEvenIsTheStopContractRungsOwn is task 3.2: one definition of 실질본전.
//
// If the observation and the rung disagreed about where break-even is, the rung
// would refuse a target the observation calls profitable, and an operator reading
// both would have no way to tell which was lying.
func TestTheBreakEvenIsTheStopContractRungsOwn(t *testing.T) {
	for _, market := range []costs.Market{costs.MarketKR, costs.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			model := costs.DefaultModel()
			const entry = "10000"

			// What the rung compares against, taken from the rung's own call.
			rendered, err := model.BreakEvenSellPrice(entry, "1", market)
			if err != nil {
				t.Fatal(err)
			}
			rung, err := parseDecimal("break-even sell price", rendered)
			if err != nil {
				t.Fatal(err)
			}

			observed, err := NetBreakEven(model, market, entry)
			if err != nil {
				t.Fatal(err)
			}
			if rung.Cmp(observed) != 0 {
				t.Fatalf("the rung uses %s and the observation uses %s; there is one break-even",
					rung.FloatString(10), observed.FloatString(10))
			}
		})
	}

	// And the rung really is refusing at that number: a target one tick under it
	// is TARGET_BELOW_BREAK_EVEN, a target on it is not.
	model := costs.DefaultModel()
	// The rendered string itself, which is the value the rung parses. Re-rounding
	// the rational to a fixed scale could land a tick under it and would make this
	// a test of FloatString rather than of the shared definition.
	onIt, err := model.BreakEvenSellPrice("10000", "1", costs.MarketKR)
	if err != nil {
		t.Fatal(err)
	}
	breakEven, err := NetBreakEven(model, costs.MarketKR, "10000")
	if err != nil {
		t.Fatal(err)
	}

	in := entryInput()
	in.Intent.LimitPrice = "10000"
	in.Intent.StopPrice = "9000"
	in.Intent.TargetPrice = onIt
	if got := checkStopContract(in); !got.Allowed {
		t.Errorf("a target exactly at the shared break-even must clear the rung: %s %s",
			got.Reason, got.Detail)
	}

	below := new(big.Rat).Sub(breakEven, big.NewRat(1, 100))
	in.Intent.TargetPrice = below.FloatString(8)
	if got := checkStopContract(in); got.Allowed || got.Reason != ReasonTargetBelowBreakEven {
		t.Errorf("a target under the shared break-even must be TARGET_BELOW_BREAK_EVEN, got %s",
			got.Reason)
	}
}

// TestNetEqualsGrossWhenEveryRateIsZero is task 3.6's first limit, and it is the
// counterexample that killed the abandoned change's monotonicity claim: 순 RR <
// 총 RR does not hold universally, because a configured zero-rate model puts
// break-even exactly on the entry.
//
// In an observation this is not a defect. It is a recorded fact about a model
// somebody configured, and the equality is arithmetically correct. A change
// promoting the net ratio to a gate has to confront it, which is why the test says
// so rather than asserting an inequality that is false.
func TestNetEqualsGrossWhenEveryRateIsZero(t *testing.T) {
	model := zeroRates(t)
	if !model.Configured() {
		t.Fatal("the fixture must be a configured model, not the zero value")
	}

	for _, market := range []costs.Market{costs.MarketKR, costs.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			gross, err := RewardRisk("10000", "9800", "10400")
			if err != nil {
				t.Fatal(err)
			}
			net := mustNetRR(t, model, market, "10000", "9800", "10400")
			if net.Cmp(gross) != 0 {
				t.Errorf("with every rate zero the two ratios must be equal: net %s, gross %s",
					trimRatio(net), trimRatio(gross))
			}
		})
	}
}

// TestNetCanExceedGrossFromFloatingPointBreakEven is task 3.6's second limit and
// the sharper one.
//
// The cost model computes in float64 and renders with strconv. At a high-precision
// entry price the rendered break-even can land *below* the entry even with
// non-negative rates, which makes the net ratio larger than the gross one. Again:
// in an observation this is a fact to record, not a bug to hide — the number is
// what the model produced. As a gate it would be a boundary decided by the binary
// expansion of a price, which is the exact failure contract.go's rationals exist
// to prevent, so a promoting change must fix precision first.
//
// The test asserts the *possibility* rather than a specific magic price: pinning
// one input would make this a test of strconv's rounding rather than of the
// limitation.
func TestNetCanExceedGrossFromFloatingPointBreakEven(t *testing.T) {
	model := zeroRates(t)

	// Entry prices with more significant digits than a float64 can hold exactly.
	candidates := []string{
		"0.10000000000000000555",
		"1234567890123456789.7",
		"0.1000000000000000055511151231257827021181583404541015625",
		"70000.000000000000001",
		"3.14159265358979311599796346854418516159057617187500001",
	}
	found := ""
	for _, entry := range candidates {
		breakEven, err := NetBreakEven(model, costs.MarketKR, entry)
		if err != nil {
			continue
		}
		exact, err := parseDecimal("entry", entry)
		if err != nil {
			continue
		}
		if breakEven.Cmp(exact) < 0 {
			found = entry
			break
		}
	}
	if found == "" {
		t.Skip("no candidate entry price rendered a break-even below itself on this platform; " +
			"the limitation is real but its witnesses depend on strconv's rounding")
	}

	t.Logf("break-even rendered below the entry at %s — the net ratio will exceed the gross one", found)

	breakEven, err := NetBreakEven(model, costs.MarketKR, found)
	if err != nil {
		t.Fatal(err)
	}
	exact, err := parseDecimal("entry", found)
	if err != nil {
		t.Fatal(err)
	}
	if breakEven.Cmp(exact) >= 0 {
		t.Fatalf("the witness stopped witnessing: break-even %s is not below entry %s",
			breakEven.FloatString(30), exact.FloatString(30))
	}
	// The consequence, stated as the arithmetic it is: with B < entry the reward
	// numerator grows and the risk denominator shrinks, so net > gross.
	if !observationOnly() {
		t.Fatal("if this ever becomes a gate input, this test must be replaced by a " +
			"precision contract rather than deleted")
	}
}

// observationOnly documents, in code, that nothing in entryChain consumes the net
// ratio. It is asserted rather than commented because the comment above depends on
// it being true.
func observationOnly() bool {
	for _, s := range entryChain {
		// The rungs are compared by name; the net ratio has no rung and adding one
		// would mean adding a name here.
		if s.name == "min_net_reward_risk" {
			return false
		}
	}
	return true
}

// TestUnmeasurableInputsProduceAbsenceNotZero is task 3.7. A break-even that
// cannot be computed leaves the observation field empty; it does not become a
// refusal, and it does not become a zero.
func TestUnmeasurableInputsProduceAbsenceNotZero(t *testing.T) {
	t.Run("an unconfigured model measures nothing", func(t *testing.T) {
		var absent costs.Model
		got := MeasureEntry(absent, costs.MarketKR, "10000", "9800", "10400")
		if got.BreakEvenPrice != "" || got.NetRewardRisk != "" {
			t.Errorf("an absent cost model must measure nothing, got %+v", got)
		}
		// The gross ratio needs no model, so it is still measured. Recording what
		// is knowable is the point: a row of all-empties would lose the geometry
		// too.
		if got.GrossRewardRisk != "2" {
			t.Errorf("the gross ratio needs no cost model: %+v", got)
		}
	})

	t.Run("an unknown market measures no net ratio", func(t *testing.T) {
		got := MeasureEntry(costs.DefaultModel(), costs.Market("xx"), "10000", "9800", "10400")
		if got.BreakEvenPrice != "" || got.NetRewardRisk != "" {
			t.Errorf("an unknown market must measure no net ratio, got %+v", got)
		}
		if got.GrossRewardRisk != "2" {
			t.Errorf("the gross ratio is still knowable: %+v", got)
		}
	})

	t.Run("a missing target measures no ratio at all", func(t *testing.T) {
		got := MeasureEntry(costs.DefaultModel(), costs.MarketKR, "10000", "9800", "")
		if got.GrossRewardRisk != "" || got.NetRewardRisk != "" {
			t.Errorf("no target means no ratio, got %+v", got)
		}
		// The break-even needs only the entry, so it is still measured.
		if got.BreakEvenPrice == "" {
			t.Errorf("the break-even depends on the entry alone: %+v", got)
		}
	})

	t.Run("the verdict is unaffected either way", func(t *testing.T) {
		// The same intent the third case measured incompletely is still refused by
		// the rung that always refused it, with the code it always used. No new
		// reason code exists for a measurement failure.
		in := entryInput()
		in.Intent.TargetPrice = ""
		got := Evaluate(in)
		if got.Allowed {
			t.Fatal("a target-less entry is refused, as it always was")
		}
		if got.Reason != ReasonInvalidTargetStop {
			t.Errorf("reason = %s, want the pre-existing code", got.Reason)
		}
		if strings.Contains(string(got.Reason), "NET") {
			t.Error("this change introduces no net-ratio reason code")
		}
	})
}

// TestBreakEvenBelowStopHasNoRatio: when the round trip eats through the stop
// there is no cost-inclusive risk per unit, and a number here would flatter the
// trade rather than describe it.
func TestBreakEvenBelowStopHasNoRatio(t *testing.T) {
	model := costs.DefaultModel()
	// A stop above the break-even: the "risk" is negative.
	_, err := NetRewardRisk(model, costs.MarketKR, "10000", "10100", "11000")
	if err == nil {
		t.Fatal("a stop above the break-even has no ratio, and must not report zero")
	}
	if !strings.Contains(err.Error(), "not above the stop") {
		t.Errorf("the refusal must name what it compared: %v", err)
	}

	got := MeasureEntry(model, costs.MarketKR, "10000", "10100", "11000")
	if got.NetRewardRisk != "" {
		t.Errorf("an uncomputable net ratio is absent, not zero: %q", got.NetRewardRisk)
	}
}

// TestRatioTextKeepsTheDigitsAThresholdWillBeChosenFrom: the storage rendering
// must not round away the boundary cases a later change needs to see.
func TestRatioTextKeepsTheDigitsAThresholdWillBeChosenFrom(t *testing.T) {
	justUnder := new(big.Rat).Sub(big.NewRat(2, 1), big.NewRat(1, 10000000000))
	if got := RatioText(justUnder); got != "1.9999999999" {
		t.Errorf("RatioText(2 − 1e-10) = %q, want the digit that says it is under", got)
	}
	if got := RatioText(big.NewRat(2, 1)); got != "2" {
		t.Errorf("RatioText(2) = %q, want %q", got, "2")
	}
	if got := RatioText(big.NewRat(3, 2)); got != "1.5" {
		t.Errorf("RatioText(1.5) = %q, want the trailing zeros trimmed", got)
	}
	if got := RatioText(nil); got != "" {
		t.Errorf("RatioText(nil) = %q, want the empty string", got)
	}
}
