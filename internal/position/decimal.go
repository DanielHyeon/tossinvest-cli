package position

// decimal.go is the cost-basis arithmetic: the two operations
// internal/riskcalc's exact decimal package does not have, and the rule for the
// one of them that cannot always be exact.
//
// riskcalc owns addition, subtraction and comparison, and the projection's
// quantities use it unchanged — a quantity is a ledger number and must not
// drift. A cost basis additionally needs `quantity × price` and
// `cost ÷ quantity`. The product of two terminating decimals terminates, so
// multiplication is exact here too. The quotient of two terminating decimals
// need not (1 ÷ 3), so division is the only rounded value in the projection and
// it is confined to avg_price.

import (
	"fmt"
	"math/big"
	"strings"
)

// Unknown is the average price of a position whose cost basis the broker has
// not told us. It is "" and not "0": the same convention
// fill_snapshots.average_price and position_adjustments' price pair already use
// (internal/journal), and for the same reason — 0 is a price, and pricing an
// unknown acquisition at zero understates every break-even computed from it.
const Unknown = ""

// avgPriceScale is how many fractional digits an average price carries.
//
// Twelve, because the numbers it has to separate are prices (five or six
// significant digits in KRW, three or four in USD) divided by quantities, and
// twelve leaves more than ten digits of headroom below the smallest tick either
// market quotes. Rounding is to nearest with halves away from zero
// (big.Rat.FloatString), so the error of one update is at most 5e-13 of one
// unit and does not favour a direction.
//
// The rounding is not accumulated inside one update — cost is recomputed from
// the stored average, so a long-lived position's basis drifts by at most
// 5e-13 × (number of buys) × quantity. At a thousand scale-ins of a million
// shares that is under a millionth of one unit of currency, which is smaller
// than the smallest amount either market can settle.
const avgPriceScale = 12

// reaverage returns the instance's average price after a BUY observation.
//
// One formula covers all three ways the hook is reached (a fill, an
// EXECUTION_CORRECTION at unchanged quantity, and a terminal observation that
// moved nothing), because all three are the same question: how much did this
// order's contribution to the position change?
//
//	contribution = cumulative filled × average price
//	Δcost        = contribution now − contribution before
//	new average  = (held × average + Δcost) ÷ (held + Δquantity)
//
// A correction has Δquantity = 0 and a non-zero Δcost, so it moves the average
// and not the quantity — which is exactly what a restatement of the price of an
// already-observed execution means.
func reaverage(base Instance, nextQuantity string, ev Event) (string, error) {
	deltaCost, known, err := contributionDelta(ev)
	if err != nil {
		return "", err
	}
	if !known {
		return Unknown, nil
	}
	heldCost, known, err := costOf(base.Quantity, base.AvgPrice)
	if err != nil {
		return "", err
	}
	if !known {
		// Once a position's basis is unknown it stays unknown: the missing piece
		// is not recoverable by averaging in the pieces that came after it.
		return Unknown, nil
	}

	quantity, err := parseRat("quantity", nextQuantity)
	if err != nil {
		return "", err
	}
	if quantity.Sign() == 0 {
		// Nothing is held, so there is no unit cost to state. The previous
		// average is kept rather than blanked: it is what the instance was
		// acquired at, and trade_outcomes (task 8.1) freezes against it.
		return base.AvgPrice, nil
	}
	cost := new(big.Rat).Add(heldCost, deltaCost)
	return DivideDecimal(cost, quantity), nil
}

// contributionDelta is how much this observation changed the order's
// contribution to the position's cost.
//
// The bool is false when the broker did not report a price for a quantity it
// did report: the contribution is then genuinely unknown, and saying 0 would be
// a claim nobody made.
func contributionDelta(ev Event) (*big.Rat, bool, error) {
	now, known, err := costOf(orZero(ev.OrderFilled), ev.OrderAvgPrice)
	if err != nil || !known {
		return nil, false, err
	}
	before, known, err := costOf(orZero(ev.PrevOrderFilled), ev.PrevOrderAvgPrice)
	if err != nil || !known {
		return nil, false, err
	}
	return new(big.Rat).Sub(now, before), true, nil
}

// costOf is quantity × price, with "no quantity" beating "no price": an order
// that has filled nothing contributes nothing whatever its average price says.
func costOf(quantity, price string) (*big.Rat, bool, error) {
	q, err := parseRat("quantity", quantity)
	if err != nil {
		return nil, false, err
	}
	if q.Sign() == 0 {
		return new(big.Rat), true, nil
	}
	if orEmpty(price) == Unknown {
		return nil, false, nil
	}
	p, err := parseRat("price", price)
	if err != nil {
		return nil, false, err
	}
	return new(big.Rat).Mul(q, p), true, nil
}

// DivideDecimal returns a ÷ b as a decimal string, rounded to avgPriceScale
// fractional digits and stripped of the zeros that rounding introduced.
//
// It is exported so the one rounded number in the projection has one
// implementation and one test, rather than a copy in every caller that needs an
// average.
func DivideDecimal(a, b *big.Rat) string {
	if b.Sign() == 0 {
		// Callers check for a zero denominator before dividing; returning the
		// unknown price rather than panicking keeps the invariant that this
		// package never stops a live fill loop.
		return Unknown
	}
	return trimDecimal(new(big.Rat).Quo(a, b).FloatString(avgPriceScale))
}

// trimDecimal removes the trailing zeros FloatString pads with, so a value has
// the same single spelling riskcalc.CanonicalDecimal produces.
func trimDecimal(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	if s == "-0" {
		return "0"
	}
	return s
}

// parseRat reads a decimal string exactly. Exponents are refused along with
// everything else that is not digits and at most one point: this package's
// inputs are ledger text, and a value nobody can read exactly must not become an
// approximation of itself (the same rule as riskcalc.parseDecimal).
func parseRat(field, value string) (*big.Rat, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, fmt.Errorf("%w: %s is empty; an empty decimal is unknown, not zero",
			ErrInvalidEvent, field)
	}
	body := strings.TrimPrefix(strings.TrimPrefix(raw, "+"), "-")
	ints, frac, _ := strings.Cut(body, ".")
	if strings.ContainsAny(body, "eE") || strings.Contains(frac, ".") ||
		(ints == "" && frac == "") || !onlyDigits(ints) || !onlyDigits(frac) {
		return nil, fmt.Errorf("%w: %s %q is not a decimal", ErrInvalidEvent, field, value)
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, fmt.Errorf("%w: %s %q is not a decimal", ErrInvalidEvent, field, value)
	}
	return r, nil
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
