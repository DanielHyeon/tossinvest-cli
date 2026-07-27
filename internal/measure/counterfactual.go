package measure

// counterfactual.go is the measurement this change exists to produce (tasks
// 5.1–5.4, 5.7; design D4).
//
// # The question it answers, and the one it does not
//
// A later change has to pick a net reward:risk threshold. Today the only inputs
// available for that are three cited precedents, all of them gross-basis, and zero
// live entry verdicts — `evaluateChain` has no production caller yet, so the
// observation table this change adds starts empty and stays that way until entries
// are switched on.
//
// So the harness manufactures the population instead: a grid of entry geometries
// pushed through the *real* chain, showing where each candidate threshold's
// boundary falls. That is a **boundary map**, and the distinction from a
// distribution is not pedantry — a grid's density is whatever its author chose, so
// counting how many grid points a threshold refuses measures the author's choice
// of grid, not the market. Anything derived from the shape of these counts is
// circular. What is not circular is where the boundary *is*: that is a property of
// the chain and the cost model.
//
// # Three things the output must say about itself
//
//	left truncation      The stop-contract rung refuses any target under break-even,
//	                     so no allowed point can have a net ratio ≤ 0. The allowed
//	                     population is truncated by construction and a reader who
//	                     misses that will read the truncation as a fact about
//	                     entries.
//	fabricated US policy `checkOrderSize` runs before `min_reward_risk` and refuses
//	                     cross-currency, and DefaultPolicy is KRW throughout. A US
//	                     grid point only evaluates at all if somebody invents a USD
//	                     limit set — so the US rows sit on numbers with no provenance
//	                     and have to say so.
//	unmeasured rates     Every rate in DefaultModel is `[미검증]`. Every number here
//	                     carries the model's fingerprint so that a re-run after the
//	                     2b measurement cannot be averaged together with this one.
//
// # No order path, no ledger, no account
//
// risk.Evaluate is a pure function over values and internal/risk imports no
// storage, so this file needs none either: it constructs Inputs and reads verdicts.
// counterfactual_isolation_test.go pins that this file imports nothing that could
// place an order or write a row.

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ThresholdCandidates are the net-basis values the map is drawn for.
//
// 1.3 is StockOS's 058 post-mortem prescription, 1.5 and 2.0 are the KRX and US
// values in its shipped `early_entry_geometry.py`. They are candidates, not
// recommendations: this change picks no threshold, and the reason it picks none is
// that until now there was nothing to pick from.
var ThresholdCandidates = []string{"1.3", "1.5", "2.0"}

// GridSpec describes one market's sweep.
type GridSpec struct {
	Market costs.Market
	// EntryPrices in the market's own currency.
	EntryPrices []string
	// StopWidths are fractions of the entry price: "0.007" is StockOS's 0.70%.
	StopWidths []string
	// RewardMultiples set the target: target = entry + m × (entry − stop). Sweeping
	// the multiple rather than the target price is what makes the gross ratio the
	// swept axis, since gross RR *is* the multiple.
	RewardMultiples []string
}

// DefaultKRGrid is the KR sweep. The stop widths bracket StockOS's 0.70% on both
// sides, because that number is the one the post-mortem blamed and the map's job
// is to show what a threshold would have done to it.
func DefaultKRGrid() GridSpec {
	return GridSpec{
		Market:          costs.MarketKR,
		EntryPrices:     []string{"10000", "70000"},
		StopWidths:      []string{"0.003", "0.007", "0.01", "0.02", "0.05"},
		RewardMultiples: []string{"1.0", "1.3", "1.5", "1.75", "2.0", "2.5", "3.0"},
	}
}

// DefaultUSGrid is the US sweep. See the fabricated-policy note on Report.
func DefaultUSGrid() GridSpec {
	return GridSpec{
		Market:          costs.MarketUS,
		EntryPrices:     []string{"100", "250"},
		StopWidths:      []string{"0.003", "0.007", "0.01", "0.02", "0.05"},
		RewardMultiples: []string{"1.0", "1.3", "1.5", "1.75", "2.0", "2.5", "3.0"},
	}
}

// GridPoint is one geometry and what the chain said about it.
type GridPoint struct {
	Market                     costs.Market
	Entry, Stop, Target        string
	StopWidth, RewardMultiple  string
	Allowed                    bool
	StoppedStep, ReasonCode    string
	BreakEven, Gross, Net      string
}

// ThresholdOutcome is one candidate's boundary on one market's grid.
type ThresholdOutcome struct {
	Threshold string
	// BelowThreshold counts chain-allowed points whose net ratio is under the
	// candidate. These are the entries the candidate would newly refuse.
	BelowThreshold int
	// AtOrAbove counts the allowed points it would keep.
	AtOrAbove int
	// Unmeasured counts allowed points with no net ratio at all.
	Unmeasured int
	// MinKept is the smallest net ratio the candidate keeps, and MaxRefused the
	// largest it refuses. The pair is the boundary; the counts are grid density and
	// mean nothing on their own.
	MinKept, MaxRefused string
}

// MarketReport is one market's map.
type MarketReport struct {
	Market costs.Market
	Points []GridPoint
	// The two constrained rungs, reported apart (SHALL — 분리 보고). Merging them
	// would collapse "the target cannot clear costs" and "the ratio is thin" into
	// one number, and they call for different fixes.
	RefusedBelowBreakEven int
	RefusedMinRewardRisk  int
	// RefusedElsewhere must be zero. Any other rung refusing means the fixture
	// failed to hold the non-geometry steps unconstrained, and the map would be
	// measuring those fixtures instead of the geometry.
	RefusedElsewhere map[string]int
	Allowed          int
	Thresholds       []ThresholdOutcome
	// PolicyProvenance is empty for KR and set for any market whose limits had to
	// be invented.
	PolicyProvenance string
}

// Report is the whole run, with the declarations that make it readable.
type Report struct {
	CostModelFingerprint string
	CostScope            string
	Markets              []MarketReport
	// RealTradePopulation is the non-synthetic half, when one was supplied.
	RealTradePopulation *PopulationReport
}

// Grid evaluates one market and returns its map.
//
// The chain is the production one. That is the point: a re-implementation of the
// rungs here would measure the re-implementation.
func Grid(model costs.Model, spec GridSpec, now time.Time) (MarketReport, error) {
	policy, provenance, err := gridPolicy(spec.Market)
	if err != nil {
		return MarketReport{}, err
	}
	report := MarketReport{
		Market:           spec.Market,
		RefusedElsewhere: map[string]int{},
		PolicyProvenance: provenance,
	}

	for _, entry := range spec.EntryPrices {
		for _, width := range spec.StopWidths {
			for _, multiple := range spec.RewardMultiples {
				point, err := evaluatePoint(model, policy, spec.Market, entry, width, multiple, now)
				if err != nil {
					return MarketReport{}, err
				}
				report.Points = append(report.Points, point)
				switch {
				case point.Allowed:
					report.Allowed++
				case point.ReasonCode == string(risk.ReasonTargetBelowBreakEven):
					report.RefusedBelowBreakEven++
				case point.StoppedStep == "min_reward_risk":
					report.RefusedMinRewardRisk++
				default:
					report.RefusedElsewhere[point.StoppedStep+"/"+point.ReasonCode]++
				}
			}
		}
	}
	report.Thresholds = thresholdOutcomes(report.Points)
	return report, nil
}

// evaluatePoint builds one Input and reads the verdict.
func evaluatePoint(
	model costs.Model, policy risk.Policy, market costs.Market,
	entry, width, multiple string, now time.Time,
) (GridPoint, error) {
	entryRat, ok := new(big.Rat).SetString(entry)
	if !ok {
		return GridPoint{}, fmt.Errorf("measure: entry price %q is not a decimal", entry)
	}
	widthRat, ok := new(big.Rat).SetString(width)
	if !ok {
		return GridPoint{}, fmt.Errorf("measure: stop width %q is not a decimal", width)
	}
	multipleRat, ok := new(big.Rat).SetString(multiple)
	if !ok {
		return GridPoint{}, fmt.Errorf("measure: reward multiple %q is not a decimal", multiple)
	}

	risked := new(big.Rat).Mul(entryRat, widthRat)
	stop := new(big.Rat).Sub(entryRat, risked)
	target := new(big.Rat).Add(entryRat, new(big.Rat).Mul(multipleRat, risked))

	point := GridPoint{
		Market:         market,
		Entry:          entry,
		Stop:           priceText(stop),
		Target:         priceText(target),
		StopWidth:      width,
		RewardMultiple: multiple,
	}
	ratios := risk.MeasureEntry(model, market, point.Entry, point.Stop, point.Target)
	point.BreakEven = ratios.BreakEvenPrice
	point.Gross = ratios.GrossRewardRisk
	point.Net = ratios.NetRewardRisk

	verdict := risk.Evaluate(risk.Input{
		Now:     now,
		Intent:  gridIntent(market, point),
		Account: unconstrainedAccount(market),
		Policy:  policy,
		Costs:   model,
	})
	point.Allowed = verdict.Allowed
	point.StoppedStep = verdict.Step
	point.ReasonCode = string(verdict.Reason)
	return point, nil
}

// priceText renders a computed price. Eight decimals, trimmed: enough that a
// 0.3%-of-100 stop width is exact, and the chain parses decimal strings anyway.
func priceText(r *big.Rat) string {
	s := r.FloatString(8)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// thresholdOutcomes draws the boundary for each candidate.
//
// Only chain-allowed points are considered. A candidate threshold is a rung that
// would run *after* the existing ones, so what it can newly refuse is what they
// already let through.
func thresholdOutcomes(points []GridPoint) []ThresholdOutcome {
	out := make([]ThresholdOutcome, 0, len(ThresholdCandidates))
	for _, candidate := range ThresholdCandidates {
		threshold, ok := new(big.Rat).SetString(candidate)
		if !ok {
			continue
		}
		outcome := ThresholdOutcome{Threshold: candidate}
		var minKept, maxRefused *big.Rat
		for _, p := range points {
			if !p.Allowed {
				continue
			}
			net, ok := new(big.Rat).SetString(p.Net)
			if !ok {
				outcome.Unmeasured++
				continue
			}
			if net.Cmp(threshold) < 0 {
				outcome.BelowThreshold++
				if maxRefused == nil || net.Cmp(maxRefused) > 0 {
					maxRefused = net
				}
				continue
			}
			outcome.AtOrAbove++
			if minKept == nil || net.Cmp(minKept) < 0 {
				minKept = net
			}
		}
		outcome.MinKept = risk.RatioText(minKept)
		outcome.MaxRefused = risk.RatioText(maxRefused)
		out = append(out, outcome)
	}
	return out
}

// gridIntent is the geometry as an intent. Quantity 1 keeps the order-size rung
// away from the geometry: the map is about prices, and a size that bumped a
// notional ceiling would show up as a refusal the geometry did not cause.
func gridIntent(market costs.Market, p GridPoint) risk.Intent {
	return risk.Intent{
		AccountRef:  "counterfactual",
		Market:      market,
		Symbol:      "GRID",
		Side:        risk.SideBuy,
		Quantity:    "1",
		LimitPrice:  p.Entry,
		StopPrice:   p.Stop,
		TargetPrice: p.Target,
	}
}

// unconstrainedAccount pins every rung that is not about geometry (SHALL — 기하
// 외의 모든 판정 단계를 비구속으로 고정).
//
// Ten of the twelve rungs read account state. If any of them were left at a value
// that could refuse, the map's refusal counts would be measuring this fixture
// rather than the geometry — which is exactly the reading trade-analytics forbids.
// Each field below is set to the value that cannot refuse, and the grid asserts
// RefusedElsewhere is empty so that a rung added later cannot silently start
// constraining the sweep.
func unconstrainedAccount(market costs.Market) risk.AccountState {
	currency := "KRW"
	headroom := "1000000000000"
	if market == costs.MarketUS {
		currency = "USD"
		headroom = "1000000000"
	}
	return risk.AccountState{
		KillSwitchActive:   false,                    // 1. kill switch: off
		Mode:               risk.ModeNormal,          // 2. operating mode: NORMAL
		EntryBlockedLatch:  false,                    // 3. entry latch: clear
		AllowedSymbols:     []string{"GRID"},         // 4. allowlist: the grid symbol
		CashAvailable:      money(headroom, currency),// 8. cash: unlimited headroom
		SameDayEntryCount:  0,                        // 9. re-entry: none today
		PendingBuy:         false,                    // 9. no unfilled buy
		OpenExposure:       money("0", currency),     // 10. open exposure: nothing open
		DailyRealizedLoss:  money("0", currency),     // 11. daily loss: none
		AccountEquity:      money(headroom, currency),// 11. equity: unlimited
		DuplicateOrder:     false,                    // 12. duplicate: none
	}
}

func money(amount, currency string) riskcalc.Money {
	return riskcalc.Money{Amount: amount, Currency: currency}
}

// FabricatedUSPolicyNote is the provenance warning the US rows carry (task 5.3).
const FabricatedUSPolicyNote = "US limits are fabricated. checkOrderSize runs before " +
	"min_reward_risk and refuses a cross-currency intent as INPUT_UNAVAILABLE, and " +
	"risk.DefaultPolicy() is KRW in every field — so a US grid point cannot be evaluated " +
	"at all without inventing a USD limit set. These numbers have no provenance and must " +
	"not be cited as policy."

// gridPolicy returns limits generous enough that only the geometry can refuse.
func gridPolicy(market costs.Market) (risk.Policy, string, error) {
	policy := risk.DefaultPolicy()
	currency := "KRW"
	headroom := "1000000000000"
	provenance := ""
	if market == costs.MarketUS {
		currency = "USD"
		headroom = "1000000000"
		provenance = FabricatedUSPolicyNote
	}
	policy.MaxOrderQuantity = "1000000"
	policy.MaxOrderNotional = money(headroom, currency)
	policy.MaxOpenExposure = money(headroom, currency)
	policy.MaxDailyLoss = money(headroom, currency)
	policy.RiskBudget = money(headroom, currency)
	if err := policy.Validate(); err != nil {
		return risk.Policy{}, "", fmt.Errorf("measure: the grid policy is unusable: %w", err)
	}
	return policy, provenance, nil
}

// --- rendering ----------------------------------------------------------------

// LeftTruncationNote is the structural limit on the allowed population (task 5.4).
const LeftTruncationNote = "Left-truncated. The stop-contract rung (5) already refuses any " +
	"target below the cost-inclusive break-even, so no chain-allowed point can carry a net " +
	"ratio at or below zero. The absence of such points is a property of the existing chain, " +
	"not of entries."

// BoundaryMapNote is the naming rule (task 5.1).
const BoundaryMapNote = "This is a boundary map, not a distribution. The grid's density is " +
	"chosen by whoever wrote the spec, so the *counts* below measure that choice; only the " +
	"boundary values (the largest ratio a candidate refuses and the smallest it keeps) are " +
	"properties of the chain and the cost model."

// StopWidthCircularityNote is task 5.5's warning, carried even when no real
// population was supplied.
const StopWidthCircularityNote = "The minimum stop-width constant `k` is NOT settled by this " +
	"output. The grid's stop widths are values the spec's author chose, so deriving a floor " +
	"from them is circular. The only non-synthetic source is the real-trade population; " +
	"without it, `k` remains open."

// Render writes the report as Markdown.
//
// The declaration block comes first and is not optional (SHALL — 산출물은 무엇을 줄
// 수 있고 무엇을 줄 수 없는지 함께 선언해야 한다). A reader who takes only the tables
// away should still have been told what the tables cannot support.
func (r Report) Render() string {
	var b strings.Builder
	b.WriteString("# Counterfactual entry-geometry boundary map\n\n")

	b.WriteString("## What this output can and cannot give\n\n")
	b.WriteString("| Downstream requirement | This output | Why |\n|---|---|---|\n")
	b.WriteString("| ① gross/net boundary map | **yes** | the synthetic grid below |\n")
	b.WriteString("| ② stop-width distribution (the basis for `k`) | **only from the real-trade population** | " +
		"the grid's widths are chosen, so deriving `k` from them is circular |\n")
	b.WriteString("| ③ measured cost ratios per market | **no** | all seven rates are `[미검증]`; " +
		"today's values restate the placeholders |\n")
	b.WriteString("| ④ declared target vs realised exit | **no** | needs closed positions; this change " +
		"only records the target |\n\n")

	fmt.Fprintf(&b, "- Cost model fingerprint: `%s`\n", r.CostModelFingerprint)
	fmt.Fprintf(&b, "- Cost scope: `%s` — commission and tax on both legs. Slippage is **not** included, "+
		"so the metric is 수수료·세금 차감 후 RR.\n", r.CostScope)
	fmt.Fprintf(&b, "- %s\n", BoundaryMapNote)
	fmt.Fprintf(&b, "- %s\n", LeftTruncationNote)
	fmt.Fprintf(&b, "- %s\n\n", StopWidthCircularityNote)

	for _, m := range r.Markets {
		fmt.Fprintf(&b, "## Market `%s`\n\n", m.Market)
		if m.PolicyProvenance != "" {
			fmt.Fprintf(&b, "> ⚠️ %s\n\n", m.PolicyProvenance)
		}
		fmt.Fprintf(&b, "Grid points: %d. Chain-allowed: %d.\n\n", len(m.Points), m.Allowed)
		b.WriteString("Refusals, by rung — reported apart because they are different facts:\n\n")
		fmt.Fprintf(&b, "- `stop_contract` (target below break-even): %d\n", m.RefusedBelowBreakEven)
		fmt.Fprintf(&b, "- `min_reward_risk` (gross ratio under 2.0): %d\n", m.RefusedMinRewardRisk)
		if len(m.RefusedElsewhere) > 0 {
			b.WriteString("- ⚠️ **other rungs refused, so this map is measuring the fixture " +
				"rather than the geometry**:\n")
			for _, key := range sortedKeys(m.RefusedElsewhere) {
				fmt.Fprintf(&b, "  - `%s`: %d\n", key, m.RefusedElsewhere[key])
			}
		}
		b.WriteString("\n### Candidate net thresholds\n\n")
		b.WriteString("| Candidate | would refuse | would keep | largest refused | smallest kept |\n")
		b.WriteString("|---|---|---|---|---|\n")
		for _, o := range m.Thresholds {
			fmt.Fprintf(&b, "| %s | %d | %d | %s | %s |\n",
				o.Threshold, o.BelowThreshold, o.AtOrAbove, dash(o.MaxRefused), dash(o.MinKept))
		}
		b.WriteString("\n")
	}

	if r.RealTradePopulation != nil {
		b.WriteString(r.RealTradePopulation.render())
	} else {
		b.WriteString("## Real-trade population\n\nNone supplied. `k` remains open — see the " +
			"declaration above.\n")
	}
	return b.String()
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
