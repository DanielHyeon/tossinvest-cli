package risk

// netrr.go is the net reward:risk observation (change add-net-rr-measurement,
// design D3).
//
// # It is a measurement, not a rung
//
// Nothing in entryChain calls this. The minimum-reward-risk gate still judges the
// gross ratio against 2.0 and this change alters no verdict — that is the whole
// contract of it. What this function produces is a number written beside the
// verdict so that a later change can choose a net threshold from observed data
// instead of from the three gross-basis precedents that have been cited for it.
//
// # Why it is separate from RewardRisk
//
// RewardRisk is pure arithmetic over three prices. This needs a market and a cost
// model, because 실질본전 depends on both. Widening RewardRisk's signature to take
// them would put a cost dependency on the gate's own function for the benefit of
// something the gate does not use.
//
// # The break-even is the stop contract's, not a second opinion
//
// B comes from `Costs.BreakEvenSellPrice(entry, "1", market)` — the identical call
// checkStopContract makes when it compares the target (contract.go). Two
// break-even definitions would let the rung that refuses a below-break-even target
// and the observation that measures the same trade disagree about where break-even
// is. Quantity 1 for the same reason it does there: with a zero profit floor the
// quantity cancels out exactly.
//
// # Precision, stated rather than fixed
//
// B is float64 arithmetic rendered to a decimal string, and this package parses
// that string into an exact rational. So the ratio below is exact arithmetic over
// an approximate input: it inherits the cost model's bounded relative error in its
// last significant digits. For an observation that is a recorded limitation and
// not a defect — no verdict turns on it. A change promoting this to a gate must
// deal with it first, because a gate boundary decided by the binary expansion of a
// price is exactly what contract.go's rationals exist to avoid.

import (
	"fmt"
	"math/big"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// NetRewardRisk returns (target − B) / (B − stop) exactly, where B is the
// fee-and-tax break-even sell price for the entry.
//
// Like RewardRisk it refuses rather than returning zero when the ratio does not
// exist. The reasons differ, though, and the difference matters to whoever reads
// the observation:
//
//	B could not be computed   the cost model has no rates for the market, or the
//	                          entry price is unusable. Nothing is known.
//	B is not above the stop   the costs alone eat through the stop, so there is no
//	                          risk-per-unit left to divide by. The trade is worse
//	                          than "poor ratio" and a number here would flatter it.
//
// A caller recording an observation writes the empty string for both cases: an
// absent measurement, never a zero one (risk-management: 0 대체 금지).
func NetRewardRisk(
	model costs.Model, market costs.Market, entryPrice, stopPrice, targetPrice string,
) (*big.Rat, error) {
	breakEven, err := NetBreakEven(model, market, entryPrice)
	if err != nil {
		return nil, err
	}
	stop, err := parseDecimal("stop price", stopPrice)
	if err != nil {
		return nil, err
	}
	target, err := parseDecimal("target price", targetPrice)
	if err != nil {
		return nil, err
	}
	risk := new(big.Rat).Sub(breakEven, stop)
	if risk.Sign() <= 0 {
		return nil, fmt.Errorf(
			"the break-even sell price %s is not above the stop %s, so the cost-inclusive "+
				"risk per unit does not exist", trimRatio(breakEven), stopPrice)
	}
	reward := new(big.Rat).Sub(target, breakEven)
	return new(big.Rat).Quo(reward, risk), nil
}

// NetBreakEven returns 실질본전 for one entry, as an exact rational over the cost
// model's rendered decimal.
//
// Exported because the observation records the value as well as the ratio, and
// because task 3.2's single-definition check needs to compare it against the number
// checkStopContract used. Having one function both consumers call is what makes
// that comparison a tautology rather than a coincidence.
func NetBreakEven(model costs.Model, market costs.Market, entryPrice string) (*big.Rat, error) {
	if !model.Configured() {
		return nil, fmt.Errorf("no cost model was supplied; an absent model is not a free trade")
	}
	rendered, err := model.BreakEvenSellPrice(entryPrice, "1", market)
	if err != nil {
		return nil, err
	}
	return parseDecimal("break-even sell price", rendered)
}

// RatioText renders a ratio for storage: exact to a fixed scale, with the
// trailing zeros trimmed.
//
// Ten decimals rather than the four trimRatio uses for operators. These strings
// are the analysis input a threshold gets chosen from, and the difference between
// 1.9999999999 and 2.0 is precisely the kind of boundary case that change will
// need to see rather than a rounding of it.
func RatioText(r *big.Rat) string {
	if r == nil {
		return ""
	}
	s := r.FloatString(10)
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}

// EntryRatios is both ratios and the break-even, as the observation stores them.
//
// Empty strings mean "not computed". The struct deliberately has no error field:
// a caller that could not measure still records what it does know, because "this
// entry was judged and its net ratio is unknown" is a fact the analysis needs and
// a dropped row is not.
type EntryRatios struct {
	BreakEvenPrice  string
	GrossRewardRisk string
	NetRewardRisk   string
}

// MeasureEntry computes what an observation records for one entry.
//
// It never fails. Each field is filled if it can be and left empty if it cannot,
// which is what lets the recording site be a straight line with no branch that
// could accidentally turn a measurement failure into a verdict (design D5: 본전
// 산출 실패가 판정을 바꾸지 않고 관측 항목의 결측으로 기록된다, and no new reason
// code is introduced for it).
func MeasureEntry(
	model costs.Model, market costs.Market, entryPrice, stopPrice, targetPrice string,
) EntryRatios {
	var out EntryRatios
	if breakEven, err := NetBreakEven(model, market, entryPrice); err == nil {
		out.BreakEvenPrice = RatioText(breakEven)
	}
	if gross, err := RewardRisk(entryPrice, stopPrice, targetPrice); err == nil {
		out.GrossRewardRisk = RatioText(gross)
	}
	if net, err := NetRewardRisk(model, market, entryPrice, stopPrice, targetPrice); err == nil {
		out.NetRewardRisk = RatioText(net)
	}
	return out
}
