// Package riskcalc is the calculation contract behind the aggregate limits
// (extend-execution-contract task 6.1).
//
// It is the artifact the order-execution requirement "총계 한도의 계산 계약"
// asks for: 총 개방 노출·일일 손실·현금은 계산 계약이 정의되어야 하며(SHALL),
// 정의되지 않은 양에 예약을 걸어서는 안 된다(SHALL NOT). A limit over a quantity
// nobody has defined is not a control — it is a number that happens to be
// compared against something.
//
// # The structural decisions this package transcribes
//
// The numbers may be replaced by a later change, but only in the conservative
// direction (WORKFLOW §0.9); the shapes below are settled here.
//
//  1. **Automated entries are LIMIT only** (SHALL). A market order has no price
//     until it fills, so its exposure valuation does not exist. Entry exposure is
//     `지정가 × 수량 + 과대 추정 비용` — the cost term is an over-estimate on
//     purpose, because under-estimating it under-reserves.
//  2. **Exposure is long-only gross** (SHALL): 보유 평가액 + 미체결 매수 평가액 +
//     유효 진입 예약. Nothing nets anything else off. A short leg cannot reduce
//     the number, and neither can an unfilled sell.
//  3. **Daily loss is realised, after costs** (SHALL). Unrealised is out of scope
//     until a later change adds it — and adding it can only raise the measured
//     loss, which is the conservative direction.
//  4. **Currency is normalised at a fresh rate, and a stale rate fails closed**
//     (SHALL). There is no implicit 1:1.
//  5. **External manual trades reach this package only through the broker
//     snapshot** (SHALL). There is deliberately no local-ledger input here: the
//     local ledger does not know about trades a human placed in the app, so a
//     calculation fed from it would silently under-count exposure.
//  6. **Staleness bounds are named, conservative defaults** (SHALL — neither
//     "always stale" nor "never stale"). See the constants below.
//
// Any input that is stale or unknown refuses the entry (SHALL). That is why
// every function here returns an error rather than a best effort: a caller
// cannot accidentally proceed on a number that was never established.
//
// # Everything is an explicit input
//
// There is no clock, no network client and no configuration lookup in this
// package. The caller passes `Now`, the snapshot's as-of, the rates and their
// as-ofs. That is what makes these functions total and testable, and it is also
// the property that stops a hidden `time.Now()` from turning a stale snapshot
// into a fresh one at a later call site.
//
// # Decimals
//
// Amounts arrive and leave as decimal strings, the convention the journal uses
// for every broker decimal (a binary float cannot hold them exactly, and these
// values end up in audit records). Arithmetic is done in float64 after parsing,
// exactly as internal/journal's fill ledger does, and results are rendered with
// `strconv.FormatFloat(v, 'f', -1, 64)` — the shortest string that round-trips.
//
// The precision claim is therefore bounded, not exact: KRW integer amounts are
// exact up to 2^53 (≈9×10^15, far beyond any account this trades), and
// fractional/foreign amounts accumulate at most a relative error of about 1e-16
// per operation. Because that error can fall either way, no comparison against a
// limit is made on equality: WithinLimit treats a value within a scale-relative
// tolerance of the limit as *exceeding* it, using the same 1e-9 relative
// tolerance the fill ledger uses for quantities. Ties fail closed.
package riskcalc

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Staleness bounds. These are the D6a conservative defaults, transcribed.
//
// Provenance: the fill detector polls every 3s (internal/filldetect.DefaultConfig,
// itself the retry matrix's single-order cadence) and the fill-detection SLO
// targets 10s from broker-visible to local commit. Ten seconds is therefore three
// polls plus margin — long enough that a healthy account never trips it, short
// enough that a snapshot older than it genuinely describes a different account
// state. An FX rate moves far more slowly than a position does, so 60s buys
// tolerance for a slower quote feed without letting a real currency move go
// unnoticed inside one decision.
//
// They are not tuning knobs. WORKFLOW §0.9 allows sizing-adjacent changes in the
// conservative direction only, so a later change may *shorten* these; lengthening
// either one requires a human's approval and an audit record (§0.5, §0.7). The
// code enforces the first half of that: a caller-supplied bound narrows the
// window and can never widen it (see Staleness.resolve).
const (
	// AccountSnapshotStaleness bounds holdings, unfilled orders and realised
	// P&L — everything read from the broker's account snapshot.
	AccountSnapshotStaleness = 10 * time.Second
	// FXRateStaleness bounds the rate used to normalise currencies.
	FXRateStaleness = 60 * time.Second
)

// Sentinel errors. Every failure here is fail-closed: the caller refuses the
// entry. They are distinguished so an operator can tell "we could not read it"
// from "we read it too long ago" from "this shape is not allowed at all".
var (
	// ErrStaleInput: the input exists but is older than its bound.
	ErrStaleInput = errors.New("riskcalc: input older than its staleness bound")
	// ErrUnknownInput: the input is absent, unreadable, or not a usable value.
	ErrUnknownInput = errors.New("riskcalc: required input is unknown")
	// ErrMarketEntry: an automated entry that is not a LIMIT order. Its exposure
	// valuation is undefined, so there is nothing to measure against a limit.
	ErrMarketEntry = errors.New("riskcalc: automated entries are LIMIT only")
)

// InputError names the input that failed and why.
//
// Fail-closed is only operable if an operator can see which input stopped
// trading; an opaque refusal produces a restart loop instead of a fix.
type InputError struct {
	// Input names the failing input in the operator's vocabulary, e.g.
	// "account snapshot", "FX rate USD→KRW".
	Input string
	// Detail explains what was wrong with it.
	Detail string
	// Age and Bound are set for a staleness failure.
	Age   time.Duration
	Bound time.Duration

	kind error
}

func (e *InputError) Error() string {
	if e.Bound > 0 {
		return fmt.Sprintf("riskcalc: %s is %s old, past the %s bound (%s)",
			e.Input, e.Age, e.Bound, e.Detail)
	}
	return fmt.Sprintf("riskcalc: %s: %s", e.Input, e.Detail)
}

func (e *InputError) Unwrap() error { return e.kind }

func unknown(input, detail string) error {
	return &InputError{Input: input, Detail: detail, kind: ErrUnknownInput}
}

func stale(input string, age, bound time.Duration) error {
	return &InputError{
		Input: input, Detail: "a limit decision cannot rest on it",
		Age: age, Bound: bound, kind: ErrStaleInput,
	}
}

// Money is an amount in one currency, as a decimal string.
//
// The currency is not optional anywhere in this package: an amount whose
// currency nobody recorded cannot be summed with another, and assuming the
// account's currency is precisely the guess that makes a foreign position look
// 1350× smaller than it is.
type Money struct {
	Amount   string
	Currency string
}

// OrderType is the entry's order type. Only LIMIT is valid for an automated
// entry; the type exists so "which of these did the caller mean" is answered by
// the compiler rather than by a string comparison at each call site.
type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

// EntryIntent is one proposed automated entry, as the sizing decision describes
// it. It is not an order and does not become one: this package only values it.
type EntryIntent struct {
	Symbol    string
	OrderType OrderType
	// LimitPrice is the price the exposure is valued at. Required, because a
	// market entry is refused before this is read.
	LimitPrice string
	Quantity   string
	// CostOverestimate is commission, tax and slippage headroom, deliberately
	// over-stated. It is required: an empty one is an unknown input, not a zero,
	// because "this trade is free" is a claim about the broker nobody measured.
	CostOverestimate string
	Currency         string
}

// Holding is one position in the broker's account snapshot.
type Holding struct {
	Symbol   string
	Quantity string
	// Price is the valuation price the snapshot carries, in Currency.
	Price    string
	Currency string
}

// UnfilledBuy is one resting buy order in the broker's account snapshot. Its
// remaining quantity is exposure the account has committed to but not yet taken.
type UnfilledBuy struct {
	Symbol            string
	LimitPrice        string
	RemainingQuantity string
	Currency          string
}

// EntryReservation is a live risk reservation held by a decision that has been
// issued but whose order is not yet visible in the snapshot
// (journal.risk_reservations, kind OPEN_EXPOSURE). Counting it is what stops two
// decisions in the same window from each fitting under the limit alone.
type EntryReservation struct {
	DecisionID string
	Amount     string
	Currency   string
}

// AccountSnapshot is the broker's view of the account at one instant.
//
// This is the *only* source of positions and resting orders in this package, by
// design: manual trades a human placed outside the engine appear here and
// nowhere else (SHALL — 외부 수동 거래는 브로커 스냅샷을 통해서만 반영된다).
type AccountSnapshot struct {
	// AsOf is when the broker's data was current. A zero value is unknown, never
	// "just now".
	AsOf         time.Time
	Holdings     []Holding
	UnfilledBuys []UnfilledBuy
}

// RealizedPnL is one realised result for a trading day, after costs.
type RealizedPnL struct {
	// TradingDay is market-local, YYYY-MM-DD. It must match the day being
	// measured: a daily limit computed over the wrong day is not the limit
	// anyone configured.
	TradingDay string
	// Amount is signed: negative is a loss.
	Amount   string
	Currency string
	// CostsDeducted asserts that commission and tax are already subtracted. A
	// false value is refused rather than used, because "실현 손익(비용 차감 후)"
	// is the definition of the quantity, not a preference about it.
	CostsDeducted bool
	// AsOf carries the account snapshot's bound: this number comes from the
	// broker, like the positions do.
	AsOf time.Time
}

// FXRate converts one currency into another at a known instant.
type FXRate struct {
	From string
	To   string
	// Rate is units of To per one unit of From, as a decimal string.
	Rate string
	AsOf time.Time
}

// Staleness bounds one calculation's inputs.
//
// A zero field means "use the transcribed default". A non-zero field narrows the
// window; it can never widen it (see resolve) — §0.9's one-way ratchet expressed
// in code rather than in a comment nobody reads.
type Staleness struct {
	AccountSnapshot time.Duration
	FXRate          time.Duration
}

// DefaultStaleness returns the transcribed D6a bounds.
func DefaultStaleness() Staleness {
	return Staleness{AccountSnapshot: AccountSnapshotStaleness, FXRate: FXRateStaleness}
}

// resolve applies "narrower wins" to both bounds.
func (s Staleness) resolve() Staleness {
	return Staleness{
		AccountSnapshot: narrower(s.AccountSnapshot, AccountSnapshotStaleness),
		FXRate:          narrower(s.FXRate, FXRateStaleness),
	}
}

func narrower(want, def time.Duration) time.Duration {
	if want <= 0 || want > def {
		return def
	}
	return want
}

// checkFresh judges one as-of against a bound.
//
// A zero as-of is unknown rather than fresh, and a future one is refused: a
// negative age means the two clocks disagree, and a clock nobody can trust is
// not a freshness proof.
func checkFresh(input string, asOf, now time.Time, bound time.Duration) error {
	if now.IsZero() {
		return unknown("evaluation instant", "riskcalc takes the instant as an input and was given none")
	}
	if asOf.IsZero() {
		return unknown(input, "it carries no as-of, so its age cannot be established")
	}
	age := now.Sub(asOf)
	if age < 0 {
		return &InputError{
			Input: input, Detail: "it is dated in the future; the clocks disagree",
			Age: age, Bound: bound, kind: ErrStaleInput,
		}
	}
	if age > bound {
		return stale(input, age, bound)
	}
	return nil
}

// parseAmount reads a decimal string the way the journal does, and refuses
// everything the journal refuses: an empty value is unknown rather than zero,
// and a non-finite one is unusable.
func parseAmount(input, value string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, unknown(input, "no value was supplied; an empty decimal is unknown, not zero")
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, unknown(input, fmt.Sprintf("%q is not a decimal", trimmed))
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, unknown(input, fmt.Sprintf("%q is not finite", trimmed))
	}
	return v, nil
}

// parseNonNegative additionally refuses a negative value, for the quantities and
// prices where a negative would silently shrink the aggregate.
func parseNonNegative(input, value string) (float64, error) {
	v, err := parseAmount(input, value)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, unknown(input, fmt.Sprintf("%q is negative", strings.TrimSpace(value)))
	}
	return v, nil
}

func requireCurrency(input, currency string) (string, error) {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		return "", unknown(input, "it carries no currency, and there is no default one")
	}
	return c, nil
}

// formatAmount renders a computed amount the way the journal renders decimals.
func formatAmount(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// nearlyEqual is the fill ledger's scale-relative tolerance
// (internal/journal.nearlyZero), reused so this package's idea of "the same
// number" matches the one already in the durable path.
func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
