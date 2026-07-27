package riskcalc

// aggregate.go computes the three quantities the aggregate limits are measured
// against. Each is a pure function of its inputs; none of them reads a clock, a
// file or the network.

import (
	"fmt"
	"strings"
	"time"
)

// CheckAutomatedEntry refuses an automated entry that is not a LIMIT order.
//
// The rule is structural, not stylistic: 시장가 진입의 노출 평가가는 정의되지
// 않는다. A market order's fill price is unknown until it exists, so there is no
// number to reserve against, and reserving against a guess is how an account
// ends up over its own limit. Anything that is not exactly LIMIT is refused —
// fail-closed applies to unrecognised order types too, not only to MARKET.
func CheckAutomatedEntry(intent EntryIntent) error {
	if intent.OrderType == OrderTypeLimit {
		return nil
	}
	return &InputError{
		Input: "entry order type",
		Detail: fmt.Sprintf(
			"%q: automated entries are LIMIT only — a non-limit entry has no defined exposure valuation",
			string(intent.OrderType)),
		kind: ErrMarketEntry,
	}
}

// EntryExposure values one proposed entry: 지정가 × 수량 + 과대 추정 비용.
//
// The market-entry check runs here as well as in CheckAutomatedEntry, so no
// caller can reach a number by valuing an intent it never checked.
func EntryExposure(intent EntryIntent) (Money, error) {
	if err := CheckAutomatedEntry(intent); err != nil {
		return Money{}, err
	}
	currency, err := requireCurrency("entry intent", intent.Currency)
	if err != nil {
		return Money{}, err
	}
	price, err := parseNonNegative("entry limit price", intent.LimitPrice)
	if err != nil {
		return Money{}, err
	}
	quantity, err := parseNonNegative("entry quantity", intent.Quantity)
	if err != nil {
		return Money{}, err
	}
	costs, err := parseNonNegative("entry cost over-estimate", intent.CostOverestimate)
	if err != nil {
		return Money{}, err
	}
	return Money{Amount: formatAmount(price*quantity + costs), Currency: currency}, nil
}

// GrossLongInputs is everything the gross long exposure is computed from.
//
// There is no local position ledger in this struct on purpose: external manual
// trades are visible only through Snapshot (SHALL), and a calculation that could
// fall back to local state would under-count exposure exactly when a human had
// been trading in the app.
type GrossLongInputs struct {
	// Now is the instant the freshness of everything else is judged against.
	Now time.Time
	// TargetCurrency is what the result is expressed in. Required.
	TargetCurrency string
	// Bounds may narrow the transcribed staleness defaults, never widen them.
	Bounds Staleness

	Snapshot     AccountSnapshot
	Reservations []EntryReservation
	Rates        []FXRate
}

// GrossLongExposure computes 보유 평가액 + 미체결 매수 평가액 + 유효 진입 예약.
//
// Long-only gross: nothing is netted against anything. A resting sell does not
// reduce the number and neither does a short leg — the limit is on how much the
// account has at risk, and a hedge that has not settled is not a reduction.
//
// Reservations are added to the snapshot rather than reconciled against it.
// Between issuing a decision and the broker showing its order, the exposure
// exists in the reservation only; double-counting an order that appears in both
// is the conservative error, and the reservation's release (design D5) is what
// resolves it — not an offset guessed here.
func GrossLongExposure(in GrossLongInputs) (Money, error) {
	bounds := in.Bounds.resolve()
	target, err := requireCurrency("target currency", in.TargetCurrency)
	if err != nil {
		return Money{}, err
	}
	if err := checkFresh("account snapshot", in.Snapshot.AsOf, in.Now, bounds.AccountSnapshot); err != nil {
		return Money{}, err
	}

	// The three terms of the spec's sum, in the order it lists them. Each is a
	// separate function so the composition stays legible and a missing term
	// would be visible here rather than buried in one long loop.
	total := 0.0
	for _, term := range []func(GrossLongInputs, string, Staleness) (float64, error){
		grossHoldings, grossUnfilledBuys, grossReservations,
	} {
		value, err := term(in, target, bounds)
		if err != nil {
			return Money{}, err
		}
		total += value
	}
	return Money{Amount: formatAmount(total), Currency: target}, nil
}

func grossHoldings(in GrossLongInputs, target string, bounds Staleness) (float64, error) {
	total := 0.0
	for _, h := range in.Snapshot.Holdings {
		label := "holding " + describe(h.Symbol)
		currency, err := requireCurrency(label, h.Currency)
		if err != nil {
			return 0, err
		}
		quantity, err := parseNonNegative(label+" quantity", h.Quantity)
		if err != nil {
			return 0, err
		}
		price, err := parseNonNegative(label+" price", h.Price)
		if err != nil {
			return 0, err
		}
		value, err := convert(quantity*price, currency, target, in.Rates, in.Now, bounds.FXRate)
		if err != nil {
			return 0, err
		}
		total += value
	}
	return total, nil
}

func grossUnfilledBuys(in GrossLongInputs, target string, bounds Staleness) (float64, error) {
	total := 0.0
	for _, b := range in.Snapshot.UnfilledBuys {
		label := "unfilled buy " + describe(b.Symbol)
		currency, err := requireCurrency(label, b.Currency)
		if err != nil {
			return 0, err
		}
		quantity, err := parseNonNegative(label+" remaining quantity", b.RemainingQuantity)
		if err != nil {
			return 0, err
		}
		price, err := parseNonNegative(label+" limit price", b.LimitPrice)
		if err != nil {
			return 0, err
		}
		value, err := convert(quantity*price, currency, target, in.Rates, in.Now, bounds.FXRate)
		if err != nil {
			return 0, err
		}
		total += value
	}
	return total, nil
}

func grossReservations(in GrossLongInputs, target string, bounds Staleness) (float64, error) {
	total := 0.0
	for _, r := range in.Reservations {
		label := "entry reservation " + describe(r.DecisionID)
		currency, err := requireCurrency(label, r.Currency)
		if err != nil {
			return 0, err
		}
		amount, err := parseNonNegative(label+" amount", r.Amount)
		if err != nil {
			return 0, err
		}
		value, err := convert(amount, currency, target, in.Rates, in.Now, bounds.FXRate)
		if err != nil {
			return 0, err
		}
		total += value
	}
	return total, nil
}

// DailyLossInputs is everything the day's realised loss is computed from.
type DailyLossInputs struct {
	Now time.Time
	// TradingDay is the market-local day being measured, YYYY-MM-DD.
	TradingDay     string
	TargetCurrency string
	Bounds         Staleness

	Realized []RealizedPnL
	Rates    []FXRate
}

// DailyLoss computes the day's realised loss after costs, as a non-negative
// magnitude in the target currency.
//
// A profitable day is a loss of zero, never a negative loss: a signed result
// would read as headroom under the limit and let a good morning fund a worse
// afternoon than the limit allows.
//
// Unrealised P&L is deliberately absent. A later change may add it (SHALL —
// 미실현 포함 여부는 후속이 보수 방향으로만 추가), and including it can only
// raise the measured loss, so its absence is the permissive-but-defined position
// this change fixes rather than an oversight.
func DailyLoss(in DailyLossInputs) (Money, error) {
	bounds := in.Bounds.resolve()
	target, err := requireCurrency("target currency", in.TargetCurrency)
	if err != nil {
		return Money{}, err
	}
	day := strings.TrimSpace(in.TradingDay)
	if day == "" {
		return Money{}, unknown("trading day", "a daily loss needs the day it is measured over")
	}

	net := 0.0
	for _, r := range in.Realized {
		label := "realised P&L for " + day
		if strings.TrimSpace(r.TradingDay) != day {
			return Money{}, unknown(label, fmt.Sprintf(
				"an entry is dated %q; a daily limit measured over another day is not the configured limit",
				strings.TrimSpace(r.TradingDay)))
		}
		if !r.CostsDeducted {
			return Money{}, unknown(label,
				"costs are not deducted; the limit is defined over realised P&L after costs")
		}
		if err := checkFresh(label, r.AsOf, in.Now, bounds.AccountSnapshot); err != nil {
			return Money{}, err
		}
		currency, err := requireCurrency(label, r.Currency)
		if err != nil {
			return Money{}, err
		}
		// Signed: a realised result may be a gain or a loss.
		amount, err := parseAmount(label+" amount", r.Amount)
		if err != nil {
			return Money{}, err
		}
		value, err := convert(amount, currency, target, in.Rates, in.Now, bounds.FXRate)
		if err != nil {
			return Money{}, err
		}
		net += value
	}

	loss := 0.0
	if net < 0 {
		loss = -net
	}
	return Money{Amount: formatAmount(loss), Currency: target}, nil
}

// WithinLimit reports whether a normalised aggregate is below a configured
// limit.
//
// Three fail-closed rules live here rather than at each call site:
//
//   - an unset or non-positive limit is an unknown input, not "unlimited";
//   - the two currencies must already match — normalisation happens before the
//     comparison, never inside it;
//   - a value within the fill ledger's scale-relative tolerance of the limit
//     counts as *exceeding* it. Reaching the limit is not staying under it, and
//     float accumulation can fall either side of an exact tie.
func WithinLimit(value, limit Money) (bool, error) {
	valueCurrency, err := requireCurrency("aggregate value", value.Currency)
	if err != nil {
		return false, err
	}
	limitCurrency, err := requireCurrency("configured limit", limit.Currency)
	if err != nil {
		return false, err
	}
	if valueCurrency != limitCurrency {
		return false, unknown("configured limit", fmt.Sprintf(
			"the value is in %s and the limit in %s; normalise before comparing",
			valueCurrency, limitCurrency))
	}
	limitValue, err := parseAmount("configured limit", limit.Amount)
	if err != nil {
		return false, err
	}
	if limitValue <= 0 {
		return false, unknown("configured limit",
			"it is not a positive amount; an absent limit is not an unlimited one")
	}
	amount, err := parseAmount("aggregate value", value.Amount)
	if err != nil {
		return false, err
	}
	if nearlyEqual(amount, limitValue) {
		return false, nil
	}
	return amount < limitValue, nil
}

// convert normalises one amount into the target currency.
//
// A same-currency amount needs no rate at all: requiring an identity rate would
// make every domestic calculation fail whenever the quote feed was down, which
// is a fail-closed nobody intended. Every other conversion needs a rate that
// exists, is positive and is fresh — a missing rate is an unknown input, never
// an implicit 1:1.
func convert(amount float64, from, to string, rates []FXRate, now time.Time, bound time.Duration) (float64, error) {
	if from == to {
		return amount, nil
	}
	label := fmt.Sprintf("FX rate %s→%s", from, to)
	for _, r := range rates {
		if !strings.EqualFold(strings.TrimSpace(r.From), from) ||
			!strings.EqualFold(strings.TrimSpace(r.To), to) {
			continue
		}
		if err := checkFresh(label, r.AsOf, now, bound); err != nil {
			return 0, err
		}
		rate, err := parseAmount(label, r.Rate)
		if err != nil {
			return 0, err
		}
		if rate <= 0 {
			return 0, unknown(label, fmt.Sprintf("%q is not a positive rate", strings.TrimSpace(r.Rate)))
		}
		return amount * rate, nil
	}
	return 0, unknown(label, "no rate was supplied; a foreign amount is never converted at an implicit 1:1")
}

// describe labels an item in an error message without asserting anything about
// the identifier's shape.
func describe(id string) string {
	if strings.TrimSpace(id) == "" {
		return "(unnamed)"
	}
	return id
}
