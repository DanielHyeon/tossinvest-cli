package risk

// chain.go is task 2.1: the fixed evaluation order and the steps that are not
// the stop contract (which lives in contract.go).
//
// The steps that depend on account aggregates and journal state — cash, the
// allowlist, re-entry, the totals and the duplicate check — are implemented here
// over the values Input carries. Task 2.3 wires those values to their real
// sources (riskcalc's aggregates, the journal's same-day history, the
// configured allowlist); this package never reads them itself, and that is what
// keeps the chain a pure function.

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// Decision is one chain verdict.
//
// Detail is not decoration: a refusal an operator cannot act on produces a
// restart loop instead of a fix, so every refusal names the numbers it
// compared. The code is what is recorded; the detail is what is read.
type Decision struct {
	Allowed bool
	Reason  ReasonCode
	Detail  string
}

func allow() Decision { return Decision{Allowed: true, Reason: ReasonAllowed} }

func refuse(reason ReasonCode, detail string) Decision {
	return Decision{Reason: reason, Detail: detail}
}

// Input is everything one evaluation reads.
type Input struct {
	// Now is the instant the evaluation happens at. It is an input, not a clock
	// read, so the same snapshot always produces the same verdict.
	Now time.Time
	// Intent is the proposed mutation.
	Intent Intent
	// Account is the state the intent is measured against.
	Account AccountState
	// Policy is the configured limit set.
	Policy Policy
	// Costs is the shared cost model. The zero value is not usable: it is a
	// model whose every rate is zero, which is not a measurement anybody made,
	// so callers pass costs.DefaultModel() or a configured one.
	Costs costs.Model
}

// step is one named rung of the chain.
type step struct {
	name string
	eval func(Input) Decision
}

// entryChain is the order risk-management fixes. Renaming or reordering it
// changes which reason an operator sees for an intent that breaks several rules
// at once, so both the names and the order are pinned by a test and transcribed
// into docs/guardian-chain.md.
var entryChain = []step{
	{"kill_switch", checkKillSwitch},
	{"operating_mode", checkOperatingMode},
	{"entry_gate_latch", checkEntryGateLatch},
	{"symbol_allowlist", checkSymbolAllowlist},
	{"stop_contract", checkStopContract},
	{"order_size", checkOrderSize},
	{"min_reward_risk", checkMinRewardRisk},
	{"cash", checkCash},
	{"same_day_reentry", checkSameDayReentry},
	{"open_exposure", checkOpenExposure},
	{"daily_loss", checkDailyLoss},
	{"duplicate_order", checkDuplicateOrder},
}

// EntryChainSteps returns the entry chain's step names, in evaluation order.
func EntryChainSteps() []string {
	out := make([]string, 0, len(entryChain))
	for _, s := range entryChain {
		out = append(out, s.name)
	}
	return out
}

// Evaluate runs the chain and returns the first refusal, or ALLOWED.
//
// A refusal stops the chain (SHALL — 첫 실패에서 정지): the later checks are not
// evaluated, so their inputs need not even be usable. That is deliberate — it
// means an account whose aggregate valuation is stale still gets a precise
// "kill switch is on" rather than an input error that hides the real answer.
//
// An ALLOWED verdict is not permission to trade. 총계 한도의 최종 권위는 예약
// 트랜잭션이며: the issuer must still record the decision and take the
// reservation atomically (task 4.1), and the gateway re-checks the entry gate at
// submission time.
func Evaluate(in Input) Decision {
	if d := preflight(in); !d.Allowed {
		return d
	}
	if in.Intent.Side == SideSell {
		return evaluateReduction(in)
	}
	for _, s := range entryChain {
		if d := s.eval(in); !d.Allowed {
			return d
		}
	}
	return allow()
}

// preflight refuses an input the chain cannot evaluate at all.
//
// It is separate from the chain because its failures are not policy refusals:
// they mean the caller handed over something unusable, and reporting them as
// (say) STOP_NOT_BELOW_ENTRY would send an operator to the wrong end of the
// pipeline.
func preflight(in Input) Decision {
	if in.Now.IsZero() {
		return refuse(ReasonInputUnavailable,
			"the evaluation instant is an input and none was supplied")
	}
	if strings.TrimSpace(in.Intent.AccountRef) == "" {
		return refuse(ReasonInputUnavailable, "the intent names no account")
	}
	if strings.TrimSpace(in.Intent.Symbol) == "" {
		return refuse(ReasonInputUnavailable, "the intent names no symbol")
	}
	if in.Intent.Side != SideBuy && in.Intent.Side != SideSell {
		return refuse(ReasonInputUnavailable, fmt.Sprintf(
			"intent side %q is neither %s nor %s", string(in.Intent.Side), SideBuy, SideSell))
	}
	if _, err := currencyOf(in.Intent.Market); err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if in.Intent.Side != SideBuy {
		return allow()
	}
	// Entry-only preflight. The entry price anchors every later comparison, and
	// the policy is what those comparisons are against.
	price, err := parseDecimal("entry limit price", in.Intent.LimitPrice)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if price.Sign() <= 0 {
		return refuse(ReasonInputUnavailable, fmt.Sprintf(
			"entry limit price %q is not positive", in.Intent.LimitPrice))
	}
	if err := in.Policy.Validate(); err != nil {
		return refuse(ReasonInputUnavailable, "the configured policy is not usable: "+err.Error())
	}
	// An absent cost model prices every trade as free, which understates the
	// cash an entry needs and drops the break-even price onto the entry price.
	// This is asked on the entry path only: a liquidation must never be refused
	// for anything cost-shaped (§0.3), including the model's absence.
	if !in.Costs.Configured() {
		return refuse(ReasonInputUnavailable,
			"no cost model was supplied; an absent model is not a free trade")
	}
	return allow()
}

// evaluateReduction judges a risk-reducing intent.
//
// Almost nothing applies to it, and each omission is a rule rather than an
// oversight:
//
//   - The **kill switch** is BLOCK-ONLY (SHALL NOT — 어떤 소비자도 강제청산을
//     유발하지 않는다). It stops new exposure; a switch that also stopped exits
//     would turn a safety control into a trap.
//   - The **operating mode** table permits RISK_REDUCING in all three modes,
//     including HALT_ALL, and 수동 flatten-all은 모든 모드에서 통과한다(§0.3).
//   - The **entry latch** is, by name, about entries.
//   - The **allowlist** is an entry control. A symbol removed from it must still
//     be closeable, or the removal would strand the position it was meant to
//     stop growing.
//   - **Cost** never gates a liquidation (SHALL NOT — §0.3, StockOS
//     SELL_COST_BUFFER 미이식).
//   - The **aggregates** measure exposure, which a reduction lowers.
//
// What remains is the one thing a reduction can get wrong: selling more than the
// account holds, which is a short by another name.
func evaluateReduction(in Input) Decision {
	quantity, err := parseWholeNumber("intent quantity", in.Intent.Quantity)
	if err != nil {
		return refuse(ReasonInvalidOrderSize, err.Error())
	}
	if quantity.Sign() <= 0 {
		return refuse(ReasonInvalidOrderSize, fmt.Sprintf(
			"quantity %q is not a positive whole number", in.Intent.Quantity))
	}
	held, err := parseDecimal("held quantity", in.Account.HeldQuantity)
	if err != nil {
		return refuse(ReasonInputUnavailable,
			"the holding for this symbol is unknown, so a reduction cannot be told from a short: "+err.Error())
	}
	if new(big.Rat).SetInt(quantity).Cmp(held) > 0 {
		return refuse(ReasonSellExceedsHoldings, fmt.Sprintf(
			"selling %s of a %s holding would open a short; reductions are reduce-only",
			in.Intent.Quantity, in.Account.HeldQuantity))
	}
	return allow()
}

// --- 1. kill switch -------------------------------------------------------------

func checkKillSwitch(in Input) Decision {
	if in.Account.KillSwitchActive {
		return refuse(ReasonKillSwitchActive,
			"the kill switch is active; it blocks new entries only and never forces a liquidation")
	}
	return allow()
}

// --- 2. operating mode ------------------------------------------------------------

func checkOperatingMode(in Input) Decision {
	switch in.Account.Mode {
	case ModeNormal:
		return allow()
	case ModeEntryBlocked, ModeHaltAll:
		return refuse(ReasonOperatingModeBlocked, fmt.Sprintf(
			"operating mode %s refuses exposure-raising intents", string(in.Account.Mode)))
	default:
		// A mode nobody recognises is not NORMAL. Failing closed here is what
		// stops a typo in a persisted enum from reading as permission.
		return refuse(ReasonOperatingModeBlocked, fmt.Sprintf(
			"operating mode %q is not one this build recognises; an unknown mode is refused, never read as NORMAL",
			string(in.Account.Mode)))
	}
}

// --- 3. entry gate latch -----------------------------------------------------------

func checkEntryGateLatch(in Input) Decision {
	if !in.Account.EntryBlockedLatch {
		return allow()
	}
	reason := strings.TrimSpace(in.Account.EntryBlockedReason)
	if reason == "" {
		reason = "no reason recorded"
	}
	return refuse(ReasonEntryGateBlocked, fmt.Sprintf(
		"the entry gate is latched (%s); it clears through the gate's own rules, not through this chain", reason))
}

// --- 4. allowlist ---------------------------------------------------------------------

func checkSymbolAllowlist(in Input) Decision {
	symbol := strings.ToUpper(strings.TrimSpace(in.Intent.Symbol))
	for _, allowed := range in.Account.AllowedSymbols {
		if strings.ToUpper(strings.TrimSpace(allowed)) == symbol {
			return allow()
		}
	}
	return refuse(ReasonSymbolNotAllowed, fmt.Sprintf(
		"%s is not on the entry allowlist (%d symbols); an empty or missing allowlist admits nothing",
		in.Intent.Symbol, len(in.Account.AllowedSymbols)))
}

// --- 6. order size -----------------------------------------------------------------------

// checkOrderSize is the chain's fifth rung: the quantity itself, then the
// per-trade risk budget, then the configured per-order ceilings.
//
// The risk budget is checked before the configured ceilings because it is
// derived from the stop the previous rung just validated — it is the sizing rule
// rather than an envelope around it — and because an intent that is too large
// for its own stop is wrong in a way that raising a configured limit would not
// fix.
//
// The chain recomputes the budget-derived cap rather than trusting the issuer's
// arithmetic. Guardian is the checkpoint; a quantity the caller asserts is
// exactly the thing a checkpoint exists to re-derive.
func checkOrderSize(in Input) Decision {
	quantity, err := parseWholeNumber("intent quantity", in.Intent.Quantity)
	if err != nil {
		return refuse(ReasonInvalidOrderSize, err.Error())
	}
	if quantity.Sign() <= 0 {
		return refuse(ReasonInvalidOrderSize, fmt.Sprintf(
			"quantity %q is not a positive whole number", in.Intent.Quantity))
	}

	currency, err := currencyOf(in.Intent.Market)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if limit := in.Policy.LimitCurrency(); limit != currency {
		// a090's mixed-currency guard: a KRW budget divided by a USD stop width
		// is a quantity in no currency at all. Normalisation belongs to
		// riskcalc, with its FX staleness bound.
		return refuse(ReasonInputUnavailable, fmt.Sprintf(
			"the intent is priced in %s and the configured limits are in %s; "+
				"a cross-currency size is refused rather than computed at an implicit rate",
			currency, limit))
	}

	sized, err := RiskBasedQuantity(in.Policy.RiskBudget.Amount, in.Intent.LimitPrice, in.Intent.StopPrice)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	capacity, err := parseWholeNumber("risk-based quantity", sized)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if capacity.Sign() <= 0 {
		return refuse(ReasonInvalidOrderSize, fmt.Sprintf(
			"a risk budget of %s %s cannot cover one unit of risk between entry %s and stop %s",
			in.Policy.RiskBudget.Amount, currency, in.Intent.LimitPrice, in.Intent.StopPrice))
	}
	if quantity.Cmp(capacity) > 0 {
		return refuse(ReasonInvalidOrderSize, fmt.Sprintf(
			"quantity %s is above the %s units a risk budget of %s %s allows over a stop %s wide",
			in.Intent.Quantity, sized, in.Policy.RiskBudget.Amount, currency,
			widthOf(in.Intent.LimitPrice, in.Intent.StopPrice)))
	}

	maxQuantity, err := parseWholeNumber("maximum order quantity", in.Policy.MaxOrderQuantity)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if quantity.Cmp(maxQuantity) > 0 {
		return refuse(ReasonMaxOrderExceeded, fmt.Sprintf(
			"quantity %s is above the per-order ceiling of %s units",
			in.Intent.Quantity, in.Policy.MaxOrderQuantity))
	}

	notional, d := entryNotional(in, "0")
	if !d.Allowed {
		return d
	}
	over, d := exceedsInclusiveLimit(notional, in.Policy.MaxOrderNotional)
	if !d.Allowed {
		return d
	}
	if over {
		return refuse(ReasonMaxOrderExceeded, fmt.Sprintf(
			"notional %s %s is above the per-order ceiling of %s %s",
			notional.Amount, notional.Currency,
			in.Policy.MaxOrderNotional.Amount, in.Policy.MaxOrderNotional.Currency))
	}
	return allow()
}

// widthOf renders the stop width for an operator message, or "unknown" when the
// prices cannot be read (the caller has already refused in that case).
func widthOf(entryPrice, stopPrice string) string {
	entry, err := parseDecimal("entry", entryPrice)
	if err != nil {
		return "unknown"
	}
	stop, err := parseDecimal("stop", stopPrice)
	if err != nil {
		return "unknown"
	}
	return trimRatio(new(big.Rat).Sub(entry, stop))
}

// --- 8. cash ---------------------------------------------------------------------------

// checkCash compares the intent's cash requirement — notional **plus the
// over-estimated cost** — against available cash. The boundary is inclusive:
// cash that exactly covers both is enough.
func checkCash(in Input) Decision {
	required, d := entryNotionalWithCosts(in)
	if !d.Allowed {
		return d
	}
	available, err := moneyIn("available cash", in.Account.CashAvailable, required.Currency)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	cmp, err := riskcalc.CompareDecimal(required.Amount, available)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if cmp > 0 {
		return refuse(ReasonInsufficientCash, fmt.Sprintf(
			"the entry needs %s %s including estimated cost but %s %s is available",
			required.Amount, required.Currency, available, required.Currency))
	}
	return allow()
}

// --- 9. same-day re-entry -----------------------------------------------------------------

func checkSameDayReentry(in Input) Decision {
	if in.Account.PendingBuy {
		return refuse(ReasonPendingBuyOrderBlocked, fmt.Sprintf(
			"%s already has an unfilled buy; a second entry would size against exposure the first has not taken yet",
			in.Intent.Symbol))
	}
	count := in.Account.SameDayEntryCount
	if count < 0 {
		return refuse(ReasonInputUnavailable, fmt.Sprintf(
			"same-day entry count for %s is %d", in.Intent.Symbol, count))
	}
	if count == 0 {
		return allow()
	}
	if count >= in.Policy.Reentry.MaxEntriesPerSymbol {
		return refuse(ReasonSameDayReentryBlocked, fmt.Sprintf(
			"%s already had %d of at most %d entries today",
			in.Intent.Symbol, count, in.Policy.Reentry.MaxEntriesPerSymbol))
	}
	if in.Policy.Reentry.Cooldown <= 0 {
		return allow()
	}
	if in.Account.LastEntryAt.IsZero() {
		// The count says an entry happened; without its instant there is no way
		// to show the cooldown has elapsed, and "we lost the timestamp" must not
		// be the way to skip a wait.
		return refuse(ReasonSameDayReentryCooldownActive, fmt.Sprintf(
			"%s had %d entries today but no recorded time for the last one, so the %s cooldown cannot be shown to have elapsed",
			in.Intent.Symbol, count, in.Policy.Reentry.Cooldown))
	}
	elapsed := in.Now.Sub(in.Account.LastEntryAt)
	if elapsed < 0 {
		return refuse(ReasonInputUnavailable, fmt.Sprintf(
			"the last entry on %s is dated %s in the future; the clocks disagree", in.Intent.Symbol, -elapsed))
	}
	if elapsed < in.Policy.Reentry.Cooldown {
		return refuse(ReasonSameDayReentryCooldownActive, fmt.Sprintf(
			"%s was entered %s ago, inside the %s cooldown",
			in.Intent.Symbol, elapsed.Round(time.Second), in.Policy.Reentry.Cooldown))
	}
	return allow()
}

// --- 10. open exposure ----------------------------------------------------------------------

// checkOpenExposure adds this intent to the account's existing exposure and
// compares the total against the ceiling.
//
// The boundary is ≥ — reaching the limit blocks. riskcalc.WithinLimit is the
// comparison, not a local one, because it already encodes that rule together
// with the tie tolerance the reservation ledger uses; two implementations of
// "is this under the limit" would eventually disagree at the boundary, which is
// the only place the answer matters.
func checkOpenExposure(in Input) Decision {
	entry, d := entryNotionalWithCosts(in)
	if !d.Allowed {
		return d
	}
	existing, err := magnitudeIn("open exposure", in.Account.OpenExposure, entry.Currency)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	projected, err := riskcalc.AddDecimal(existing, entry.Amount)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	within, err := riskcalc.WithinLimit(
		riskcalc.Money{Amount: projected, Currency: entry.Currency}, in.Policy.MaxOpenExposure)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if !within {
		return refuse(ReasonOpenExposureExceeded, fmt.Sprintf(
			"open exposure would reach %s %s against a ceiling of %s %s; reaching the ceiling blocks",
			projected, entry.Currency, in.Policy.MaxOpenExposure.Amount, in.Policy.MaxOpenExposure.Currency))
	}
	return allow()
}

// --- 11. daily loss ---------------------------------------------------------------------------

// checkDailyLoss applies three rules in the order of how badly they fail.
//
// Equity first: a loss ratio over zero or negative equity is not a large number,
// it is an undefined one, and risk-management says the account is blocked
// outright (계좌자본 0 이하이면 즉시 차단). Then the absolute ceiling, then the
// ratio — both with the aggregate's ≥ boundary.
func checkDailyLoss(in Input) Decision {
	currency, err := currencyOf(in.Intent.Market)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	equityAmount, err := moneyIn("account equity", in.Account.AccountEquity, currency)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	equity, err := parseDecimal("account equity", equityAmount)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if equity.Sign() <= 0 {
		return refuse(ReasonDailyLossLimitReached, fmt.Sprintf(
			"account equity is %s %s; an account with no capital has no loss budget to measure against",
			equityAmount, currency))
	}

	lossAmount, err := magnitudeIn("daily realised loss", in.Account.DailyRealizedLoss, currency)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	within, err := riskcalc.WithinLimit(
		riskcalc.Money{Amount: lossAmount, Currency: currency}, in.Policy.MaxDailyLoss)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	if !within {
		return refuse(ReasonDailyLossLimitReached, fmt.Sprintf(
			"the day's realised loss is %s %s against a ceiling of %s %s; reaching the ceiling blocks",
			lossAmount, currency, in.Policy.MaxDailyLoss.Amount, in.Policy.MaxDailyLoss.Currency))
	}

	loss, err := parseDecimal("daily realised loss", lossAmount)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	ratio, err := parseDecimal("maximum daily loss ratio", in.Policy.MaxDailyLossRatio)
	if err != nil {
		return refuse(ReasonInputUnavailable, err.Error())
	}
	// loss / equity >= ratio, without the division: loss >= ratio × equity.
	threshold := new(big.Rat).Mul(ratio, equity)
	if loss.Cmp(threshold) >= 0 {
		return refuse(ReasonDailyLossLimitReached, fmt.Sprintf(
			"the day's realised loss %s %s has reached %s of equity %s %s (ceiling %s)",
			lossAmount, currency, trimRatio(new(big.Rat).Quo(loss, equity)), equityAmount, currency,
			in.Policy.MaxDailyLossRatio))
	}
	return allow()
}

// --- 12. duplicate ------------------------------------------------------------------------------

func checkDuplicateOrder(in Input) Decision {
	if in.Account.DuplicateOrder {
		return refuse(ReasonDuplicateOrder, fmt.Sprintf(
			"an equivalent intent for %s is already in flight", in.Intent.Symbol))
	}
	return allow()
}

// --- shared valuation ------------------------------------------------------------------------------

// entryNotional values the intent through riskcalc, which owns the calculation
// contract for exposure (지정가 × 수량 + 과대 추정 비용) and refuses a non-LIMIT
// entry on the way. costOverestimate is "0" when the bare notional is wanted.
func entryNotional(in Input, costOverestimate string) (riskcalc.Money, Decision) {
	currency, err := currencyOf(in.Intent.Market)
	if err != nil {
		return riskcalc.Money{}, refuse(ReasonInputUnavailable, err.Error())
	}
	value, err := riskcalc.EntryExposure(riskcalc.EntryIntent{
		Symbol:           in.Intent.Symbol,
		OrderType:        riskcalc.OrderTypeLimit,
		LimitPrice:       in.Intent.LimitPrice,
		Quantity:         in.Intent.Quantity,
		CostOverestimate: costOverestimate,
		Currency:         currency,
	})
	if err != nil {
		return riskcalc.Money{}, refuse(ReasonInputUnavailable, err.Error())
	}
	return value, allow()
}

// EntryExposureValue is what one entry adds to the account's open exposure:
// 지정가 × 수량 + 과대 추정 비용, valued through riskcalc's calculation contract.
//
// It is exported for one caller: the issuer (task 4.1), which has to hold that
// amount back in the reservation ledger. 총계 한도의 최종 권위는 예약 트랜잭션 —
// so the reservation must consume exactly the number the chain just found room
// for. A second arithmetic at the call site would be a second answer, and two
// answers for one aggregate differ at the boundary, which is the only place the
// aggregate decides anything.
//
// The refusal it can return is the chain's own INPUT_UNAVAILABLE: an input the
// chain could not evaluate is not an amount anybody may reserve against.
func EntryExposureValue(in Input) (riskcalc.Money, Decision) {
	if d := preflight(in); !d.Allowed {
		return riskcalc.Money{}, d
	}
	if in.Intent.Side != SideBuy {
		return riskcalc.Money{}, refuse(ReasonInputUnavailable,
			"only an entry adds open exposure; a reduction lowers it")
	}
	return entryNotionalWithCosts(in)
}

// entryNotionalWithCosts values the intent including the cost model's estimate
// of what placing it costs. Both the cash check and the exposure check use it,
// so an entry is measured the same way against both.
func entryNotionalWithCosts(in Input) (riskcalc.Money, Decision) {
	bare, d := entryNotional(in, "0")
	if !d.Allowed {
		return riskcalc.Money{}, d
	}
	cost, err := in.Costs.EstimateTradeCost(bare.Amount, costs.SideBuy, in.Intent.Market)
	if err != nil {
		return riskcalc.Money{}, refuse(ReasonInputUnavailable, err.Error())
	}
	return entryNotional(in, cost.Total)
}

// exceedsInclusiveLimit reports whether a value is past a per-order ceiling.
//
// Per-order ceilings are inclusive (design D2: 주문 단위 한도는 포함 상한),
// unlike the aggregates, so this is an exact `>` on the decimals rather than
// riskcalc.WithinLimit's tie-fails-closed comparison. The difference is
// deliberate and is the reason the two live in different helpers.
func exceedsInclusiveLimit(value riskcalc.Money, limit riskcalc.Money) (bool, Decision) {
	amount, err := moneyIn("configured per-order ceiling", limit, value.Currency)
	if err != nil {
		return false, refuse(ReasonInputUnavailable, err.Error())
	}
	cmp, err := riskcalc.CompareDecimal(value.Amount, amount)
	if err != nil {
		return false, refuse(ReasonInputUnavailable, err.Error())
	}
	return cmp > 0, allow()
}

// magnitudeIn returns an aggregate valuation, having checked what moneyIn checks
// and one thing more: that it is not negative.
//
// The extra rule closes a fail-open at the package boundary. Both aggregates the
// chain measures against are **magnitudes** by their producer's contract —
// riskcalc.GrossLongExposure sums non-negative terms, and riskcalc.DailyLoss
// returns `max(0, −net)` (aggregate.go: `loss := 0.0; if net < 0 { loss = -net }`).
// A caller that instead passed the signed P&L would hand a day that lost 100,000
// over as "−100,000", and every comparison in this file would read that as a day
// well inside its limit. That is the one input error here that opens the gate
// instead of closing it, so it is refused by name rather than compared.
func magnitudeIn(label string, m riskcalc.Money, want string) (string, error) {
	amount, err := moneyIn(label, m, want)
	if err != nil {
		return "", err
	}
	value, err := parseDecimal(label, amount)
	if err != nil {
		return "", err
	}
	if value.Sign() < 0 {
		return "", fmt.Errorf(
			"%s is %s %s; this aggregate is a magnitude, and a negative one means signed P&L was passed where a loss was expected — "+
				"comparing it would read a losing day as an unused limit",
			label, amount, want)
	}
	return amount, nil
}

// moneyIn returns an amount, having checked that it exists and is expressed in
// the expected currency. Currencies are never converted here: riskcalc owns
// conversion, together with the staleness bound that makes a rate usable.
func moneyIn(label string, m riskcalc.Money, want string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(m.Currency))
	if currency == "" {
		return "", fmt.Errorf("%s carries no currency; an amount nobody named a currency for cannot be compared", label)
	}
	if currency != want {
		return "", fmt.Errorf("%s is in %s but the comparison is in %s; normalise before comparing, never inside a comparison",
			label, currency, want)
	}
	if _, err := parseDecimal(label, m.Amount); err != nil {
		return "", err
	}
	return strings.TrimSpace(m.Amount), nil
}
