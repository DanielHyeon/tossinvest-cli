package measure_test

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/measure"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// counterfactual_test.go is change add-net-rr-measurement tasks 5.1–5.5 and 5.7.
//
// The property that matters most here is not any particular number. It is that the
// output cannot be read as more than it is: a boundary map from a chosen grid,
// under unmeasured rates, on a left-truncated population, with fabricated US
// limits. Each of those four is asserted, because each one is a way this artifact
// could mislead the change that reads it.

var harnessNow = time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC)

func krGrid(t *testing.T) measure.MarketReport {
	t.Helper()
	report, err := measure.Grid(costs.DefaultModel(), measure.DefaultKRGrid(), harnessNow)
	if err != nil {
		t.Fatalf("Grid(kr): %v", err)
	}
	return report
}

// TestOnlyTheTwoGeometryRungsRefuse is task 5.2, and it is the test the whole
// harness rests on. Ten of the twelve rungs read account state; if any of them
// could refuse, the map's counts would be measuring this fixture instead of the
// geometry — which trade-analytics explicitly forbids reporting as a property of
// the geometry.
func TestOnlyTheTwoGeometryRungsRefuse(t *testing.T) {
	for _, spec := range []measure.GridSpec{measure.DefaultKRGrid(), measure.DefaultUSGrid()} {
		t.Run(string(spec.Market), func(t *testing.T) {
			report, err := measure.Grid(costs.DefaultModel(), spec, harnessNow)
			if err != nil {
				t.Fatalf("Grid: %v", err)
			}
			if len(report.RefusedElsewhere) != 0 {
				t.Fatalf("rungs outside the geometry refused %v. The fixture failed to hold "+
					"the non-geometry steps unconstrained, so every count in this map is a "+
					"statement about the fixture", report.RefusedElsewhere)
			}
			if len(report.Points) == 0 {
				t.Fatal("the grid produced no points")
			}
			// Both constrained rungs must actually fire somewhere, or the sweep is
			// not covering the boundary it claims to map.
			if report.RefusedBelowBreakEven == 0 {
				t.Error("no point was refused for a target under break-even; the sweep does " +
					"not reach the stop-contract boundary")
			}
			if report.RefusedMinRewardRisk == 0 {
				t.Error("no point was refused for a thin gross ratio; the sweep does not " +
					"reach the minimum-reward-risk boundary")
			}
			if report.Allowed == 0 {
				t.Error("no point was allowed; a map with no allowed population has no " +
					"threshold boundary to draw")
			}
		})
	}
}

// TestTheTwoRefusalReasonsAreReportedApart is the same requirement's second half
// (SHALL NOT — 하나의 거부율로 합산해서는 안 된다). "The target cannot clear costs"
// and "the ratio is thin" call for different fixes, and one number hides which.
func TestTheTwoRefusalReasonsAreReportedApart(t *testing.T) {
	report := krGrid(t)

	refused := len(report.Points) - report.Allowed
	if got := report.RefusedBelowBreakEven + report.RefusedMinRewardRisk; got != refused {
		t.Fatalf("the two counted rungs account for %d of %d refusals; the remainder is "+
			"unattributed", got, refused)
	}
	if report.RefusedBelowBreakEven == report.RefusedMinRewardRisk {
		t.Log("the two counts happen to be equal; that is a coincidence of this grid, " +
			"not a merge")
	}

	rendered := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{report},
	}.Render()
	if !strings.Contains(rendered, "stop_contract") || !strings.Contains(rendered, "min_reward_risk") {
		t.Errorf("the rendered output must name both rungs separately:\n%s", rendered)
	}
}

// TestNoAllowedPointHasANonPositiveNetRatio is task 5.4's left truncation, proved
// rather than only annotated. The stop-contract rung refuses a target below the
// cost-inclusive break-even, so the allowed population cannot contain a net ratio
// at or below zero — and a later change reading this map must not mistake that
// for a fact about entries.
func TestNoAllowedPointHasANonPositiveNetRatio(t *testing.T) {
	for _, spec := range []measure.GridSpec{measure.DefaultKRGrid(), measure.DefaultUSGrid()} {
		t.Run(string(spec.Market), func(t *testing.T) {
			report, err := measure.Grid(costs.DefaultModel(), spec, harnessNow)
			if err != nil {
				t.Fatal(err)
			}
			sawAllowed := false
			for _, p := range report.Points {
				if !p.Allowed {
					continue
				}
				sawAllowed = true
				net, ok := new(big.Rat).SetString(p.Net)
				if !ok {
					continue
				}
				if net.Sign() <= 0 {
					t.Errorf("allowed point %s/%s has net ratio %s. If this ever appears, the "+
						"stop-contract rung stopped bounding the allowed population and the "+
						"truncation note is wrong", p.StopWidth, p.RewardMultiple, p.Net)
				}
			}
			if !sawAllowed {
				t.Fatal("no allowed point, so the truncation claim is untested here")
			}
		})
	}
}

// TestTheUSMapDeclaresItsFabricatedPolicy is task 5.3. checkOrderSize runs before
// min_reward_risk and refuses cross-currency, and DefaultPolicy is KRW throughout
// — so a US grid point only evaluates because somebody invented USD limits, and
// the artifact has to carry that.
func TestTheUSMapDeclaresItsFabricatedPolicy(t *testing.T) {
	us, err := measure.Grid(costs.DefaultModel(), measure.DefaultUSGrid(), harnessNow)
	if err != nil {
		t.Fatal(err)
	}
	if us.PolicyProvenance == "" {
		t.Fatal("the US map must declare that its limits were fabricated")
	}
	if !strings.Contains(us.PolicyProvenance, "no provenance") {
		t.Errorf("the note must say the numbers have no provenance: %q", us.PolicyProvenance)
	}

	kr := krGrid(t)
	if kr.PolicyProvenance != "" {
		t.Errorf("KR runs on the real DefaultPolicy currency and needs no such note: %q",
			kr.PolicyProvenance)
	}

	rendered := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{kr, us},
	}.Render()
	if !strings.Contains(rendered, measure.FabricatedUSPolicyNote) {
		t.Errorf("the fabricated-policy note must reach the artifact:\n%s", rendered)
	}
}

// TestTheThresholdOutcomeIsABoundaryNotARate is task 5.1's naming rule with teeth.
// The counts depend on the grid's density; the boundary values do not. So the two
// candidates that differ must differ in where the boundary lands, and the artifact
// must say the counts are not a distribution.
func TestTheThresholdOutcomeIsABoundaryNotARate(t *testing.T) {
	report := krGrid(t)
	if len(report.Thresholds) != len(measure.ThresholdCandidates) {
		t.Fatalf("thresholds = %d, want one per candidate", len(report.Thresholds))
	}

	// Monotonicity: a higher candidate cannot refuse fewer points than a lower one.
	// If it did, the boundary computation is wrong rather than the grid unusual.
	previous := -1
	for _, o := range report.Thresholds {
		if previous >= 0 && o.BelowThreshold < previous {
			t.Errorf("candidate %s refuses %d, fewer than the candidate below it (%d)",
				o.Threshold, o.BelowThreshold, previous)
		}
		previous = o.BelowThreshold

		// The boundary is the pair, and it must be ordered: everything refused is
		// under everything kept.
		if o.MaxRefused == "" || o.MinKept == "" {
			continue
		}
		refused, _ := new(big.Rat).SetString(o.MaxRefused)
		kept, _ := new(big.Rat).SetString(o.MinKept)
		if refused.Cmp(kept) >= 0 {
			t.Errorf("candidate %s: largest refused %s is not below smallest kept %s",
				o.Threshold, o.MaxRefused, o.MinKept)
		}
	}

	rendered := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{report},
	}.Render()
	if !strings.Contains(rendered, "boundary map, not a distribution") {
		t.Errorf("the artifact must refuse the word distribution:\n%s", rendered)
	}
}

// TestTheArtifactDeclaresWhatItCannotGive is task 5.7. Four downstream
// requirements were once claimed for this change; it supplies one of them and the
// artifact says so at the top, where a reader who takes only the tables has still
// been told.
func TestTheArtifactDeclaresWhatItCannotGive(t *testing.T) {
	model := costs.DefaultModel()
	kr := krGrid(t)
	us, err := measure.Grid(model, measure.DefaultUSGrid(), harnessNow)
	if err != nil {
		t.Fatal(err)
	}
	rendered := measure.Report{
		CostModelFingerprint: model.Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{kr, us},
	}.Render()

	// The declaration comes before the data.
	declaration := strings.Index(rendered, "What this output can and cannot give")
	firstMarket := strings.Index(rendered, "## Market")
	if declaration < 0 || firstMarket < 0 || declaration > firstMarket {
		t.Fatalf("the declaration must precede the tables:\n%s", rendered)
	}

	for _, required := range []string{
		"boundary map",                   // ① supplied
		"only from the real-trade popul", // ② conditional
		"미검증",                            // ③ not supplied: rates unmeasured
		"needs closed positions",         // ④ not supplied
		model.Fingerprint(),              // the model behind every number
		"FEE_TAX_ONLY",                   // the scope
		"수수료·세금 차감 후 RR",                 // the metric's honest name
	} {
		if !strings.Contains(rendered, required) {
			t.Errorf("the artifact is missing %q:\n%s", required, rendered)
		}
	}
	if !strings.Contains(rendered, measure.LeftTruncationNote) {
		t.Error("the left-truncation note must be in the artifact")
	}
	if !strings.Contains(rendered, measure.StopWidthCircularityNote) {
		t.Error("the `k` circularity note must be in the artifact")
	}
}

// TestEveryNumberCarriesItsCostModel is 5.7's other half: two runs under different
// rates must be distinguishable from their artifacts alone.
func TestEveryNumberCarriesItsCostModel(t *testing.T) {
	cheaper, err := costs.NewModel(map[string]string{
		costs.KeyKRSellTaxRate: "0.0018", costs.KeyKRBuyCommissionRate: "0.00015",
	})
	if err != nil {
		t.Fatal(err)
	}
	measured, err := measure.Grid(cheaper, measure.DefaultKRGrid(), harnessNow)
	if err != nil {
		t.Fatal(err)
	}
	placeholder := krGrid(t)

	a := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{placeholder},
	}.Render()
	b := measure.Report{
		CostModelFingerprint: cheaper.Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{measured},
	}.Render()

	if a == b {
		t.Fatal("two rate sets produced identical artifacts")
	}
	if strings.Contains(a, cheaper.Fingerprint()) {
		t.Error("the placeholder artifact must not carry the measured model's fingerprint")
	}

	// Cheaper costs push break-even down, so a given geometry keeps more of its
	// reward: the same candidate refuses no more than before.
	for i := range placeholder.Thresholds {
		if measured.Thresholds[i].BelowThreshold > placeholder.Thresholds[i].BelowThreshold {
			t.Errorf("candidate %s refuses more under cheaper rates (%d) than under the "+
				"over-estimates (%d)", placeholder.Thresholds[i].Threshold,
				measured.Thresholds[i].BelowThreshold, placeholder.Thresholds[i].BelowThreshold)
		}
	}
}

// TestTheGridRunsTheProductionChain: a re-implementation of the rungs here would
// measure the re-implementation. The map is only worth anything because the same
// function that judges a live entry judged these.
func TestTheGridRunsTheProductionChain(t *testing.T) {
	report := krGrid(t)
	steps := map[string]bool{}
	for _, s := range risk.EntryChainSteps() {
		steps[s] = true
	}
	sawRefusal := false
	for _, p := range report.Points {
		if p.Allowed {
			continue
		}
		sawRefusal = true
		if !steps[p.StoppedStep] {
			t.Errorf("point %s/%s reports step %q, which is not a rung of the production "+
				"chain", p.StopWidth, p.RewardMultiple, p.StoppedStep)
		}
	}
	if !sawRefusal {
		t.Fatal("no refusal, so the step names went unchecked")
	}
}

// --- 5.5: the real-trade population ------------------------------------------

// TestTheDocumentedFallbackDeclaresItsSampleSize is the honesty requirement on the
// fallback. Eight hand-copied trades support no rate, and the artifact has to make
// that impossible to miss.
func TestTheDocumentedFallbackDeclaresItsSampleSize(t *testing.T) {
	entries, source := measure.StockOS058Entries()
	if len(entries) != 8 || source.SampleSize != 8 {
		t.Fatalf("the fallback is %d entries claiming %d", len(entries), source.SampleSize)
	}
	report, err := measure.Population(costs.DefaultModel(), entries, source, harnessNow)
	if err != nil {
		t.Fatalf("Population: %v", err)
	}

	rendered := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{krGrid(t)},
		RealTradePopulation:  &report,
	}.Render()

	for _, required := range []string{
		"Rows: **8**",
		"is not a rate",
		"transcription, not a sample",
		measure.StopWidthCircularityNote,
	} {
		if !strings.Contains(rendered, required) {
			t.Errorf("the population section is missing %q:\n%s", required, rendered)
		}
	}
}

// TestTheRealPopulationSuppliesTheStopWidths is what only it can give. The
// post-mortem's geometry is a 0.70% stop, and that number is the one a `k`
// argument would have to be made against.
func TestTheRealPopulationSuppliesTheStopWidths(t *testing.T) {
	entries, source := measure.StockOS058Entries()
	report, err := measure.Population(costs.DefaultModel(), entries, source, harnessNow)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.StopWidths) != len(entries) {
		t.Fatalf("stop widths = %d, want one per entry", len(report.StopWidths))
	}
	for _, w := range report.StopWidths {
		if w != "0.007" {
			t.Errorf("stop width = %q, want 058's 0.70%%", w)
		}
	}

	// And the entries themselves: gross 1.5 clears nothing, because today's gate is
	// 2.0. That is the honest headline — the geometry this change was motivated by
	// is *already* refused, on the gross basis, before any net threshold exists.
	for _, e := range report.Entries {
		if e.Gross != "1.5" {
			t.Errorf("%s gross = %q, want 1.5", e.Label, e.Gross)
		}
		if e.Allowed {
			t.Errorf("%s was allowed; the current 2.0 gross gate already refuses it", e.Label)
		}
		if e.StoppedStep != "min_reward_risk" {
			t.Errorf("%s stopped at %q, want min_reward_risk", e.Label, e.StoppedStep)
		}
	}
}

// TestAMissingPopulationIsNotAMissingMeasurement is the requirement that the
// external data is optional (SHALL — 외부 데이터 부재가 측정 부재가 되어서는 안 된다).
func TestAMissingPopulationIsNotAMissingMeasurement(t *testing.T) {
	rendered := measure.Report{
		CostModelFingerprint: costs.DefaultModel().Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{krGrid(t)},
	}.Render()

	if !strings.Contains(rendered, "## Market") {
		t.Error("the boundary map must still be produced without any external data")
	}
	if !strings.Contains(rendered, "None supplied") {
		t.Error("the absence must be stated rather than left as a missing section")
	}
	if !strings.Contains(rendered, "`k` remains open") {
		t.Error("the artifact must say what the absence costs")
	}
}
