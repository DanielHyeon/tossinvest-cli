package exitpolicy

// decimal.go is the arithmetic this package's rules are decided by: exact
// rationals over decimal strings, and one rendering rule per kind of output.
//
// The parser is internal/risk's, character for character in what it accepts: an
// optional sign, digits, at most one point. Exponents and rational literals are
// refused even though big.Rat reads them, because no writer in TossOS produces
// them and accepting them would let one number have two spellings that compare
// unequal as text.
//
// # Rendering, and why the direction of the rounding is a safety property
//
// A price computed as `entry + risk × r` terminates in decimal whenever its
// inputs do, so it is printed exactly and no rounding happens at all. The two
// values that need not terminate are rendered with a chosen direction:
//
//   - A protective price rounds **up** (ceilAt). For a long position the
//     protected baseline is a floor under the price, so rounding it up can only
//     make the protection trigger sooner. Rounding to nearest would, half the
//     time, place the baseline a hair below where the policy said — a weakening,
//     which §0.9 does not permit even at 1e-12.
//   - A taken ratio rounds **up** as well. A larger recorded take suppresses more
//     future take-profit proposals and understates the remaining quantity, both
//     of which are the quiet direction.
//   - An R multiple rounds to nearest, because it is an audit number: it is
//     rendered for the operator and for exit_events, and every decision that
//     depends on it was already made on the exact rational.

import (
	"fmt"
	"math/big"
	"strings"
)

// scale is how many fractional digits a non-terminating value is rendered with.
//
// Twelve, the same as internal/position's average price, and for the same
// reason: prices carry five or six significant digits in KRW and three or four
// in USD, so twelve leaves more than ten digits of headroom below the smallest
// tick either market quotes.
const scale = 12

// ratioScale is the scale of a taken fraction. A fraction has no units, so the
// headroom argument above does not apply; twelve digits of a ratio is a
// millionth of a share in a million-share position.
const ratioScale = 12

// rMultipleScale is what an R multiple is *shown* at. Six digits is more than an
// operator reads and far more than the numbers being compared need, and no
// decision is made on the rendered form.
const rMultipleScale = 6

var (
	zero    = new(big.Rat)
	one     = big.NewRat(1, 1)
	hundred = big.NewRat(100, 1)
)

// parseRat reads a decimal string exactly.
func parseRat(field, value string) (*big.Rat, error) {
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

// parseRatio accepts the decimal spelling used by persisted ratios and the
// exact rational spelling used by immutable policy tables (currently 1/3).
// Persisted execution totals remain decimals; only a policy ratio may use a
// slash, so no journal representation changes.
func parseRatio(field, value string) (*big.Rat, error) {
	raw := strings.TrimSpace(value)
	if !strings.Contains(raw, "/") {
		return parseRat(field, raw)
	}
	numerator, denominator, ok := strings.Cut(raw, "/")
	if !ok || numerator == "" || denominator == "" ||
		!onlyDigits(numerator) || !onlyDigits(denominator) {
		return nil, fmt.Errorf("%s %q is not a decimal or positive rational", field, value)
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok || r.Denom().Sign() <= 0 {
		return nil, fmt.Errorf("%s %q is not a decimal or positive rational", field, value)
	}
	return r, nil
}

// parseRatOr reads a decimal string, treating "" as the given default. It is for
// the columns whose empty spelling has a defined meaning — a taken ratio that has
// never moved is "0", not unknown.
func parseRatOr(field, value, fallback string) (*big.Rat, error) {
	if strings.TrimSpace(value) == "" {
		return parseRat(field, fallback)
	}
	return parseRat(field, value)
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

// formatPrice renders a price: exactly when it terminates, rounded up when it
// does not.
func formatPrice(r *big.Rat) string {
	if exact, ok := exactDecimal(r); ok {
		return exact
	}
	return ceilAt(r, scale)
}

// formatRatio renders a fraction of a position, rounded up when it does not
// terminate, and never above 1.
func formatRatio(r *big.Rat) string {
	if r.Cmp(one) >= 0 {
		return "1"
	}
	if exact, ok := exactDecimal(r); ok {
		return exact
	}
	out := ceilAt(r, ratioScale)
	// ceilAt can only reach 1 from below by rounding, and a ratio that rounds to
	// the whole position would claim the position is fully taken when it is not.
	if out == "1" {
		return trimDecimal(new(big.Rat).Sub(one, big.NewRat(1, pow10(ratioScale))).FloatString(ratioScale))
	}
	return out
}

// formatRMultiple renders an R multiple for the audit trail.
func formatRMultiple(r *big.Rat) string {
	if exact, ok := exactDecimal(r); ok {
		return exact
	}
	return trimDecimal(r.FloatString(rMultipleScale))
}

// exactDecimal returns the value's exact decimal spelling, or false when it has
// none. A rational terminates in base ten exactly when its reduced denominator
// has no prime factor other than 2 and 5.
func exactDecimal(r *big.Rat) (string, bool) {
	den := new(big.Int).Set(r.Denom())
	for _, p := range []int64{2, 5} {
		prime := big.NewInt(p)
		q, m := new(big.Int), new(big.Int)
		for {
			q.QuoRem(den, prime, m)
			if m.Sign() != 0 {
				break
			}
			den.Set(q)
		}
	}
	if den.Cmp(big.NewInt(1)) != 0 {
		return "", false
	}
	// The digits needed are bounded by the powers of 2 and 5 that were removed;
	// FloatString at a scale that large is exact, and the trim removes the
	// padding.
	digits := len(r.Denom().String()) * 4
	if digits < 1 {
		digits = 1
	}
	return trimDecimal(r.FloatString(digits)), true
}

// ceilAt renders the smallest decimal with `digits` fractional places that is not
// below r.
func ceilAt(r *big.Rat, digits int) string {
	shift := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	num := new(big.Int).Mul(r.Num(), shift)
	q, m := new(big.Int).QuoRem(num, r.Denom(), new(big.Int))
	// QuoRem truncates towards zero, so a negative value with a remainder has
	// already been rounded up and a positive one has not.
	if m.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	return trimDecimal(new(big.Rat).SetFrac(q, shift).FloatString(digits))
}

func pow10(n int) int64 {
	out := int64(1)
	for i := 0; i < n; i++ {
		out *= 10
	}
	return out
}

// trimDecimal removes the zeros FloatString pads with, so a value has the single
// spelling riskcalc.CanonicalDecimal produces.
func trimDecimal(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "" || s == "-" || s == "-0" {
		return "0"
	}
	return s
}
