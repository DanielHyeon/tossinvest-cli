package riskcalc

// confirmed_floor.go is the confirmed-floor quantity: how much an automated
// path may sell while the account is in RECONCILE (design D6, task 4.2;
// order-execution "RECONCILE 상태").
//
//	확정 하한 = max(0, min(신선한 브로커 보유수량, 신선한 매도가능수량) − 로컬 미체결 SELL 수량)
//	신선 = 스냅샷 나이 ≤ staleness 한계 · 스냅샷 부재 시 0
//
// # Why it is a formula and not a judgement
//
// RECONCILE means the engine and the account disagree. The tempting move is to
// pick whichever number looks right and carry on; the spec forbids it (권위
// 값의 불일치는 산식으로 보정하지 않고 RECONCILE 상태로 전이해야 한다) and this
// function is what replaces the judgement. Every term is a broker-reported
// quantity or a subtraction, and there is no branch where a local belief
// increases the answer.
//
// # The three conservative rules, each with the failure it prevents
//
//  1. **A local quantity may never raise the floor** (SHALL NOT). The local
//     ledger does not know about orders a human placed in the app, so its
//     position can be larger than the account's. It appears here only as a
//     subtrahend, and a negative one is refused — subtracting a negative would
//     be raising the bound through the back door.
//  2. **The sellable quantity only lowers.** Its exact meaning is
//     `[미측정 — 2b 2.8]`: settlement, pledges and same-day buys all plausibly
//     move it, and nothing has been measured. Using it inside min() means an
//     unexpectedly small value refuses quantity we might have been allowed to
//     sell, and an unexpectedly large one changes nothing.
//  3. **Absent or stale is zero, not "probably fine"** (SHALL). A snapshot the
//     caller could not read, or read too long ago, produces a floor of 0 — the
//     automated path sells nothing and the exit waits for the operator or for a
//     fresh read. Zero is the answer here, not an error: a caller must not have
//     a separate "I could not compute it" path that quietly does something else.
//
// # Placement
//
// Here rather than in execgw because it is a calculation contract, not gateway
// wiring: pure, stdlib-only, every input explicit, no clock and no network — so
// the freshness judgement cannot be softened by a hidden time.Now() at some
// call site. It also reuses this package's exact decimal layer, which is what
// keeps a fractional-share floor from drifting.
//
// # Manual flatten is not a caller (§0.3)
//
// The flatten saga reads the sellable quantity itself, per symbol, immediately
// before each order, and its P1 behaviour is preserved unchanged. Nothing in
// internal/flatten imports this package, and a test in that package asserts it
// (confirmed_floor_exempt_test.go). Gating a human's emergency liquidation on a
// snapshot that has just been declared unreadable would trap the account in the
// position it is trying to leave.
//
// # Broker-behaviour claims
//
// One, tagged: the semantics of the broker's sellable quantity are
// `[미측정 — 2b 2.8]`. Everything else here is arithmetic.

import (
	"fmt"
	"time"
)

// QuantitySnapshot is one broker-reported quantity and when it was current.
//
// A nil *QuantitySnapshot is "not read"; a zero AsOf is "read, but we cannot
// say when", which is treated the same way — unknown, never fresh.
type QuantitySnapshot struct {
	// Quantity is a non-negative decimal string.
	Quantity string
	AsOf     time.Time
}

// What determined a floor. Stable strings: an operator reads them to know
// whether to wait for a fresh read or to look at the position.
const (
	// FloorBoundHoldings — the broker's holding was the smaller term.
	FloorBoundHoldings = "broker holdings"
	// FloorBoundSellable — the sellable quantity was the smaller term.
	FloorBoundSellable = "broker sellable quantity"
	// FloorBoundLocalSells — resting local sells consumed the remainder.
	FloorBoundLocalSells = "local open sells"
	// FloorBoundNoSnapshot — a required snapshot was not read at all.
	FloorBoundNoSnapshot = "absent snapshot"
	// FloorBoundStaleSnapshot — a snapshot was read, but too long ago.
	FloorBoundStaleSnapshot = "stale snapshot"
)

// ConfirmedFloorInputs is everything the floor is computed from.
type ConfirmedFloorInputs struct {
	// Now is the instant freshness is judged against. Required.
	Now time.Time
	// Bounds may narrow the transcribed staleness default, never widen it.
	Bounds Staleness
	// Symbol is used only in messages.
	Symbol string

	// Holdings is the broker's holding quantity. Nil means it was not read.
	Holdings *QuantitySnapshot
	// Sellable is the broker's sellable quantity. Nil means it was not read.
	Sellable *QuantitySnapshot

	// LocalOpenSellQuantity is the quantity of resting local SELL orders, as a
	// decimal string. Required, and "0" when there are none: an empty value is
	// unknown, and a floor computed as if unknown were zero would authorise
	// selling quantity a live order has already committed.
	LocalOpenSellQuantity string
}

// ConfirmedFloor is the computed bound and why it is what it is.
type ConfirmedFloor struct {
	// Quantity is the decimal string an automated exit may sell, never negative.
	Quantity string
	// Bound names the term that determined it (a FloorBound* constant).
	Bound string
	// Detail explains it for the operator and the audit record.
	Detail string
}

// Zero reports whether the floor authorises nothing.
func (f ConfirmedFloor) Zero() bool { return f.Quantity == "0" }

// ConfirmedFloorQuantity computes the confirmed floor.
//
// It returns an error only for inputs that are malformed — a quantity that is
// not a decimal, a negative one, a missing evaluation instant. An absent or
// stale snapshot is not an error: it is a floor of zero, which is the answer
// the spec gives it.
func ConfirmedFloorQuantity(in ConfirmedFloorInputs) (ConfirmedFloor, error) {
	if in.Now.IsZero() {
		return ConfirmedFloor{}, unknown("evaluation instant",
			"riskcalc takes the instant as an input and was given none")
	}
	bound := in.Bounds.resolve().AccountSnapshot
	symbol := describe(in.Symbol)

	// The local quantity is validated first, so a caller that passes a negative
	// one is refused even when the snapshots would have produced a zero floor.
	// A negative subtrahend is the one shape that could raise the bound, and it
	// must not depend on which branch the snapshots take.
	localSells, err := parseNonNegativeString("local open sell quantity for "+symbol, in.LocalOpenSellQuantity)
	if err != nil {
		return ConfirmedFloor{}, err
	}

	holdings, ok, err := freshQuantity("holdings of "+symbol, in.Holdings, in.Now, bound)
	if err != nil {
		return ConfirmedFloor{}, err
	}
	if !ok {
		return zeroFloor(in.Holdings, "holdings", symbol), nil
	}
	sellable, ok, err := freshQuantity("sellable quantity of "+symbol, in.Sellable, in.Now, bound)
	if err != nil {
		return ConfirmedFloor{}, err
	}
	if !ok {
		return zeroFloor(in.Sellable, "sellable quantity", symbol), nil
	}

	// min(): the sellable quantity is used in the lowering direction only. Its
	// exact semantics are [미측정 — 2b 2.8], so a value larger than the holding
	// changes nothing and a smaller one is respected.
	base, err := MinDecimal(holdings, sellable)
	if err != nil {
		return ConfirmedFloor{}, unknown("confirmed floor for "+symbol, err.Error())
	}
	limiting := FloorBoundHoldings
	if base == sellable && sellable != holdings {
		limiting = FloorBoundSellable
	}

	remainder, err := SubDecimal(base, localSells)
	if err != nil {
		return ConfirmedFloor{}, unknown("confirmed floor for "+symbol, err.Error())
	}
	// max(0, …): resting local sells can exceed what the broker shows — that is
	// exactly the disagreement RECONCILE describes — and a negative floor would
	// read as a buy.
	floor, err := MaxDecimal("0", remainder)
	if err != nil {
		return ConfirmedFloor{}, unknown("confirmed floor for "+symbol, err.Error())
	}
	if floor != base {
		limiting = FloorBoundLocalSells
	}

	return ConfirmedFloor{
		Quantity: floor,
		Bound:    limiting,
		Detail: fmt.Sprintf(
			"min(holdings %s, sellable %s) − local open sells %s, floored at 0",
			holdings, sellable, localSells),
	}, nil
}

// freshQuantity reads one snapshot and reports whether it may be used.
//
// The bool is false for "absent or stale", which the caller turns into a zero
// floor. The error is reserved for a value that is not a usable quantity at
// all: a broker quantity that is negative or unparseable is a payload nobody
// should compute from, however fresh it is.
func freshQuantity(label string, snap *QuantitySnapshot, now time.Time, bound time.Duration) (string, bool, error) {
	if snap == nil {
		return "", false, nil
	}
	quantity, err := parseNonNegativeString(label, snap.Quantity)
	if err != nil {
		return "", false, err
	}
	if err := checkFresh(label, snap.AsOf, now, bound); err != nil {
		return "", false, nil
	}
	return quantity, true, nil
}

// parseNonNegativeString validates a decimal exactly and returns its canonical
// spelling, so the arithmetic below never sees "07.50".
func parseNonNegativeString(label, value string) (string, error) {
	canonical, err := CanonicalDecimal(value)
	if err != nil {
		return "", unknown(label, err.Error())
	}
	negative, err := IsNegativeDecimal(canonical)
	if err != nil {
		return "", unknown(label, err.Error())
	}
	if negative {
		return "", unknown(label, fmt.Sprintf("%q is negative", canonical))
	}
	return canonical, nil
}

// zeroFloor explains a floor of zero: whether the snapshot was never read or
// was read too long ago, because the operator's next action differs.
func zeroFloor(snap *QuantitySnapshot, what, symbol string) ConfirmedFloor {
	if snap == nil {
		return ConfirmedFloor{
			Quantity: "0",
			Bound:    FloorBoundNoSnapshot,
			Detail: fmt.Sprintf(
				"the %s of %s was not read; the automated path sells nothing without one", what, symbol),
		}
	}
	return ConfirmedFloor{
		Quantity: "0",
		Bound:    FloorBoundStaleSnapshot,
		Detail: fmt.Sprintf(
			"the %s of %s is older than its staleness bound; a stale quantity is not a fresh one",
			what, symbol),
	}
}
