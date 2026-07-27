package risk

// contract.go is task 2.2: the stop contract, the risk-based quantity and the
// minimum reward:risk — the three rules that judge an intent on its own terms,
// before any account state is consulted.
//
// # Why the arithmetic here is exact and riskcalc's is not
//
// riskcalc values an account in float64 and documents a bounded relative error,
// which is right for a valuation compared against a limit with a fail-closed
// tie rule. Two of the three rules here are different in kind:
//
//   - The quantity is a **count**. floor(100000/300) must be 333, and a float
//     quotient that lands at 332.9999999999999 loses a whole unit of exposure —
//     not a rounding error but a different order. StockOS hit the same edge from
//     the other side (a090 review M1: 이중 내림이면 199로 1주 과소).
//   - The reward:risk has an **inclusive boundary at exactly 2.0**. "미달" means
//     below, so a ratio of exactly 2 passes; deciding that by a float comparison
//     would make the rule's boundary an artefact of the price's binary
//     expansion.
//
// So both are computed with math/big rationals over the decimal strings. The
// break-even comparison stays in the cost model's float64 domain, because a
// break-even price is a valuation and its input rates are approximations to
// begin with.
//
// # What that means for the net ratio recorded beside the gate
//
// netrr.go computes (target − B) / (B − stop) with the same exact rationals, but B
// arrives as a decimal string rendered from float64 arithmetic — so the observed
// net ratio inherits the cost model's bounded relative error in its last
// significant digits. As an observation that is a limitation to record, not a
// defect: no verdict turns on it.
//
// A change promoting the net ratio to a gate must resolve this **first**, and the
// reason is the one this comment opens with. Two concrete counterexamples are
// pinned in netrr_test.go: with every rate zero the net ratio equals the gross one
// exactly, and at a high-precision entry price the rendered B can land below the
// entry, making the net ratio *larger* than the gross one. A gate whose boundary
// moves with the binary expansion of a price is the failure these rationals exist
// to prevent. The conservative fix is to round B **up** before comparing — a larger
// break-even lowers the net ratio, so the error is monotonically safe — or to carry
// the cost arithmetic in rationals end to end.

import (
	"fmt"
	"math/big"
	"strings"
)

// RewardRisk returns (target − entry) / (entry − stop) exactly.
//
// It refuses rather than returning zero when the ratio does not exist — a
// zero-wide or inverted stop, or an unreadable price. risk-management is
// explicit about this (SHALL — 0 대체 금지): zero is a *measured* refusal, and
// an audit record that cannot tell "the trade offered nothing" from "nobody
// could compute what it offered" has lost the distinction that matters.
func RewardRisk(entryPrice, stopPrice, targetPrice string) (*big.Rat, error) {
	entry, err := parseDecimal("entry price", entryPrice)
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
	risk := new(big.Rat).Sub(entry, stop)
	if risk.Sign() <= 0 {
		return nil, fmt.Errorf("entry %s is not above stop %s, so the risk per unit does not exist",
			entryPrice, stopPrice)
	}
	reward := new(big.Rat).Sub(target, entry)
	return new(big.Rat).Quo(reward, risk), nil
}

// RiskBasedQuantity returns floor(위험예산 / (entry − stop)) as a whole number.
//
// The grade multiplier StockOS applies here is fixed at 1.0 until P3
// (risk-management: 등급배수는 P3까지 1.0 고정 — 보수 하한), so this is the
// unmodified quotient. The floor is applied **once**, to the exact rational: a
// second rounding of an already-rounded intermediate is what a090's review
// pinned as a one-unit undersize.
//
// A stop that is not below the entry yields zero rather than an error, because
// zero is the fail-closed answer risk-management specifies (stop 폭 0 이하는
// 수량 0). The chain never reaches this with such a stop — the stop contract
// refuses first — but a later caller might, and it must not get a negative
// quantity or a division by zero.
func RiskBasedQuantity(riskBudget, entryPrice, stopPrice string) (string, error) {
	budget, err := parseDecimal("risk budget", riskBudget)
	if err != nil {
		return "", err
	}
	if budget.Sign() < 0 {
		return "", fmt.Errorf("risk budget %q is negative", riskBudget)
	}
	entry, err := parseDecimal("entry price", entryPrice)
	if err != nil {
		return "", err
	}
	stop, err := parseDecimal("stop price", stopPrice)
	if err != nil {
		return "", err
	}
	width := new(big.Rat).Sub(entry, stop)
	if width.Sign() <= 0 {
		return "0", nil
	}
	quotient := new(big.Rat).Quo(budget, width)
	// Floor of a non-negative rational is integer division of its numerator by
	// its denominator; big.Int.Quo truncates, which is the same thing here
	// because the quotient cannot be negative.
	floored := new(big.Int).Quo(quotient.Num(), quotient.Denom())
	return floored.String(), nil
}

// --- chain steps ------------------------------------------------------------

// checkStopContract is the chain's fifth rung: No Stop = No Trade, a protective
// stop, a target above the entry, and a target above the cost-inclusive
// break-even.
//
// It runs before any sizing so that "손절가가 없거나 보호적이지 않은 진입 의도는
// 수량 계산 이전 단계에서 거부되어야 한다(SHALL)" is true structurally, not by
// accident of what the numbers happen to be.
func checkStopContract(in Input) Decision {
	if strings.TrimSpace(in.Intent.StopPrice) == "" {
		return refuse(ReasonStopMissing,
			"the intent carries no stop price; an entry without one cannot be sized or protected")
	}
	entry, err := parseDecimal("entry price", in.Intent.LimitPrice)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	stop, err := parseDecimal("stop price", in.Intent.StopPrice)
	if err != nil {
		return refuse(ReasonInvalidTargetStop, err.Error())
	}
	if stop.Sign() <= 0 {
		return refuse(ReasonInvalidTargetStop,
			fmt.Sprintf("stop price %q is not a positive price", in.Intent.StopPrice))
	}
	target, err := parseDecimal("target price", in.Intent.TargetPrice)
	if err != nil {
		return refuse(ReasonInvalidTargetStop, err.Error())
	}
	if target.Sign() <= 0 {
		return refuse(ReasonInvalidTargetStop,
			fmt.Sprintf("target price %q is not a positive price", in.Intent.TargetPrice))
	}
	if stop.Cmp(entry) >= 0 {
		return refuse(ReasonStopNotBelowEntry, fmt.Sprintf(
			"stop %s is not below entry %s; TossOS is long-only, so a protective stop is under the entry",
			in.Intent.StopPrice, in.Intent.LimitPrice))
	}
	if target.Cmp(entry) <= 0 {
		return refuse(ReasonTargetNotAboveEntry, fmt.Sprintf(
			"target %s is not above entry %s", in.Intent.TargetPrice, in.Intent.LimitPrice))
	}

	// 실질 본전: the break-even sell price at the strict fee-and-tax floor.
	//
	// Quantity is passed as 1 rather than the intent's own: with a zero profit
	// floor the quantity cancels out of the formula exactly, and using the
	// intent's would make an unsizeable intent (quantity 0) report a cost
	// failure instead of the size failure the next rung is about to name.
	breakEven, err := in.Costs.BreakEvenSellPrice(in.Intent.LimitPrice, "1", in.Intent.Market)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	floor, err := parseDecimal("break-even sell price", breakEven)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if target.Cmp(floor) < 0 {
		return refuse(ReasonTargetBelowBreakEven, fmt.Sprintf(
			"target %s is below the break-even sell price %s (commission and tax on both legs); "+
				"this trade loses money even when it reaches its target",
			in.Intent.TargetPrice, breakEven))
	}
	return allow()
}

// checkMinRewardRisk is the chain's seventh rung, and it judges the **gross**
// ratio: (target − entry) / (entry − stop), against Policy.MinRewardRisk.
//
// Gross is the whole of what this gate has ever measured, and naming it matters
// now that a net ratio exists beside it. The net one — (target − B) / (B − stop),
// where B is the cost-inclusive break-even — is computed by risk.NetRewardRisk and
// recorded as an *observation* (change add-net-rr-measurement). It is not an input
// here and this rung's verdicts are unchanged by its existence.
//
// A refusal to compute maps to MIN_RR_NOT_MET, not to a pass and not to a zero:
// whatever the reason, the intent has not demonstrated the ratio it is required
// to demonstrate.
func checkMinRewardRisk(in Input) Decision {
	minimum, err := parseDecimal("minimum reward:risk", in.Policy.MinRewardRisk)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	ratio, err := RewardRisk(in.Intent.LimitPrice, in.Intent.StopPrice, in.Intent.TargetPrice)
	if err != nil {
		return refuse(ReasonMinRRNotMet, fmt.Sprintf(
			"the reward:risk cannot be computed (%s); an uncomputable ratio is refused, never read as zero", err))
	}
	if ratio.Cmp(minimum) < 0 {
		return refuse(ReasonMinRRNotMet, fmt.Sprintf(
			"reward:risk %s is below the minimum %s", trimRatio(ratio), in.Policy.MinRewardRisk))
	}
	return allow()
}

// trimRatio renders a ratio for an operator: four decimals, without the trailing
// zeros that make 1.5000 read as more precision than anyone measured.
func trimRatio(r *big.Rat) string {
	s := r.FloatString(4)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}

// --- decimal parsing ----------------------------------------------------------
//
// The accepted shape is the journal's: an optional sign, digits, at most one
// point. Exponents and rational literals ("1e3", "5/2") are refused even though
// big.Rat would read them, because no writer in TossOS produces them and
// accepting them would mean a decimal column could hold two spellings of the
// same number that compare unequal as text.

func parseDecimal(field, value string) (*big.Rat, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, fmt.Errorf("%s is empty; an empty decimal is unknown, not zero", field)
	}
	if !isPlainDecimal(raw) {
		return nil, fmt.Errorf("%s %q is not a decimal", field, value)
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, fmt.Errorf("%s %q is not a decimal", field, value)
	}
	return r, nil
}

func isPlainDecimal(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	ints, frac, hasPoint := strings.Cut(s, ".")
	if hasPoint && frac == "" {
		return false
	}
	if ints == "" && frac == "" {
		return false
	}
	if strings.Contains(frac, ".") {
		return false
	}
	return onlyDigits(ints) && onlyDigits(frac)
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// parseWholeNumber reads a count. A fractional quantity is refused rather than
// rounded: TossOS has no fractional-share order path, so a fractional intent is
// a defect upstream, and rounding it would hide the defect behind an order the
// caller did not ask for.
func parseWholeNumber(field, value string) (*big.Int, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, fmt.Errorf("%s is empty; an empty count is unknown, not zero", field)
	}
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("%s %q is not a whole number", field, value)
	}
	return n, nil
}
