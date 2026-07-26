package riskcalc

// decimal.go is exact decimal-string arithmetic: addition, subtraction and
// comparison performed on the digits themselves, with no float64 anywhere on
// the path.
//
// # Why this sits beside the float64 valuation in aggregate.go
//
// The two answer different requirements, and the difference is deliberate
// rather than an inconsistency. aggregate.go *values* an account — prices times
// quantities, converted through FX — and documents a bounded relative error
// with a fail-closed tie rule. The reservation ledger is not a valuation:
// order-execution "원자적 위험 예약" says 예약 산술은 decimal 문자열 연산이며
// float 누적을 사용하지 않는다(SHALL NOT).
//
// The reason is concrete. Ten reservations of "0.1" have to sum to exactly "1",
// because that sum is compared against a limit; a float64 accumulation of the
// same ten values is 0.9999999999999999, which is *under* the limit and
// therefore admits an eleventh. A ledger whose total drifts below the truth as
// it grows is a ledger that leaks limit headroom in exactly the direction that
// costs money.
//
// # Why here and not in internal/journal
//
// The journal already imports this package for the staleness bounds, so the
// dependency exists in that direction and cannot be reversed. And "how two
// decimals are added" is a calculation contract, not a storage detail: the same
// answer has to hold for a reservation row, for a confirmed-floor quantity and
// for anything a later change measures against a limit.
//
// # Representation
//
// A value is (sign, integer digits, fractional digits) with no leading zeros in
// the integer part and no trailing zeros in the fractional part, so every value
// has exactly one spelling and "1.10" and "1.1" compare equal. Zero is always
// positive: "-0" is not a number this package produces.

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotDecimal means a string is not a decimal number this package can read.
// It is separate from ErrUnknownInput because the callers differ: a malformed
// decimal in a reservation is a programming error, not a stale broker read.
var ErrNotDecimal = errors.New("riskcalc: not a decimal string")

// AddDecimal returns a + b exactly.
func AddDecimal(a, b string) (string, error) {
	x, y, err := parseOperands(a, b)
	if err != nil {
		return "", err
	}
	return addDecimals(x, y).String(), nil
}

// SubDecimal returns a − b exactly. The result may be negative; callers that
// need a floor at zero use MaxDecimal("0", …) so the clamp is visible at the
// call site rather than hidden in the arithmetic.
func SubDecimal(a, b string) (string, error) {
	x, y, err := parseOperands(a, b)
	if err != nil {
		return "", err
	}
	return addDecimals(x, y.negated()).String(), nil
}

// CompareDecimal returns -1, 0 or +1 as a is less than, equal to or greater
// than b.
func CompareDecimal(a, b string) (int, error) {
	x, y, err := parseOperands(a, b)
	if err != nil {
		return 0, err
	}
	return compareDecimals(x, y), nil
}

// MinDecimal returns the smaller of the two, in its canonical spelling.
func MinDecimal(a, b string) (string, error) {
	cmp, err := CompareDecimal(a, b)
	if err != nil {
		return "", err
	}
	if cmp <= 0 {
		return CanonicalDecimal(a)
	}
	return CanonicalDecimal(b)
}

// MaxDecimal returns the larger of the two, in its canonical spelling.
func MaxDecimal(a, b string) (string, error) {
	cmp, err := CompareDecimal(a, b)
	if err != nil {
		return "", err
	}
	if cmp >= 0 {
		return CanonicalDecimal(a)
	}
	return CanonicalDecimal(b)
}

// CanonicalDecimal returns the single spelling this package uses for a value:
// no plus sign, no leading zeros, no trailing fractional zeros, no "-0".
//
// Two strings that canonicalise to the same text are the same number, which is
// what lets a stored decimal be compared byte-for-byte when that is cheaper
// than parsing it again.
func CanonicalDecimal(s string) (string, error) {
	d, err := parseDecimal("decimal", s)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// IsNegativeDecimal reports whether a value is below zero.
func IsNegativeDecimal(s string) (bool, error) {
	d, err := parseDecimal("decimal", s)
	if err != nil {
		return false, err
	}
	return d.neg, nil
}

// --- internals --------------------------------------------------------------

// decimal is a parsed value. ints is "0" or digits with no leading zero; frac
// is digits with no trailing zero (possibly empty).
type decimal struct {
	neg  bool
	ints string
	frac string
}

func parseOperands(a, b string) (decimal, decimal, error) {
	x, err := parseDecimal("left operand", a)
	if err != nil {
		return decimal{}, decimal{}, err
	}
	y, err := parseDecimal("right operand", b)
	if err != nil {
		return decimal{}, decimal{}, err
	}
	return x, y, nil
}

// parseDecimal reads a decimal string without going through a float. Anything
// that is not a sign, digits and at most one point is refused: a value nobody
// can read exactly must not be turned into an approximation of itself.
func parseDecimal(field, s string) (decimal, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return decimal{}, fmt.Errorf("%w: %s is empty; an empty decimal is unknown, not zero",
			ErrNotDecimal, field)
	}
	neg := false
	switch raw[0] {
	case '+':
		raw = raw[1:]
	case '-':
		neg = true
		raw = raw[1:]
	}
	ints, frac, hasPoint := strings.Cut(raw, ".")
	if strings.Contains(frac, ".") {
		return decimal{}, fmt.Errorf("%w: %s %q has more than one point", ErrNotDecimal, field, s)
	}
	if ints == "" && frac == "" {
		return decimal{}, fmt.Errorf("%w: %s %q has no digits", ErrNotDecimal, field, s)
	}
	if !onlyDigits(ints) || !onlyDigits(frac) {
		return decimal{}, fmt.Errorf("%w: %s %q is not a decimal", ErrNotDecimal, field, s)
	}
	_ = hasPoint
	return decimal{neg: neg, ints: ints, frac: frac}.canonical(), nil
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// canonical strips the redundant zeros and normalises negative zero away.
func (d decimal) canonical() decimal {
	ints := strings.TrimLeft(d.ints, "0")
	if ints == "" {
		ints = "0"
	}
	frac := strings.TrimRight(d.frac, "0")
	neg := d.neg
	if ints == "0" && frac == "" {
		neg = false
	}
	return decimal{neg: neg, ints: ints, frac: frac}
}

func (d decimal) String() string {
	out := d.ints
	if d.frac != "" {
		out += "." + d.frac
	}
	if d.neg {
		out = "-" + out
	}
	return out
}

func (d decimal) isZero() bool { return d.ints == "0" && d.frac == "" }

func (d decimal) negated() decimal {
	if d.isZero() {
		return d
	}
	return decimal{neg: !d.neg, ints: d.ints, frac: d.frac}
}

// addDecimals implements signed addition on magnitudes, which is the only place
// the sign rules live.
func addDecimals(x, y decimal) decimal {
	if x.neg == y.neg {
		sum := addMagnitudes(x, y)
		sum.neg = x.neg
		return sum.canonical()
	}
	switch compareMagnitudes(x, y) {
	case 0:
		return decimal{ints: "0"}
	case 1:
		diff := subMagnitudes(x, y)
		diff.neg = x.neg
		return diff.canonical()
	default:
		diff := subMagnitudes(y, x)
		diff.neg = y.neg
		return diff.canonical()
	}
}

func compareDecimals(x, y decimal) int {
	switch {
	case x.neg && !y.neg:
		return -1
	case !x.neg && y.neg:
		return 1
	}
	cmp := compareMagnitudes(x, y)
	if x.neg {
		return -cmp
	}
	return cmp
}

// align pads both fractional parts to the same length and returns the two
// values as plain digit strings plus the fractional width.
func align(x, y decimal) (string, string, int) {
	width := len(x.frac)
	if len(y.frac) > width {
		width = len(y.frac)
	}
	return x.ints + x.frac + strings.Repeat("0", width-len(x.frac)),
		y.ints + y.frac + strings.Repeat("0", width-len(y.frac)),
		width
}

func compareMagnitudes(x, y decimal) int {
	a, b, _ := align(x, y)
	a = strings.TrimLeft(a, "0")
	b = strings.TrimLeft(b, "0")
	switch {
	case len(a) != len(b):
		if len(a) > len(b) {
			return 1
		}
		return -1
	case a > b:
		return 1
	case a < b:
		return -1
	default:
		return 0
	}
}

func addMagnitudes(x, y decimal) decimal {
	a, b, width := align(x, y)
	return split(addDigits(a, b), width)
}

// subMagnitudes computes |x| − |y| and requires |x| >= |y|; the caller decides
// which way round that is.
func subMagnitudes(x, y decimal) decimal {
	a, b, width := align(x, y)
	return split(subDigits(a, b), width)
}

// split turns a digit string back into integer and fractional parts, given how
// many of the trailing digits are fractional.
func split(digits string, width int) decimal {
	for len(digits) <= width {
		digits = "0" + digits
	}
	return decimal{ints: digits[:len(digits)-width], frac: digits[len(digits)-width:]}
}

func addDigits(a, b string) string {
	if len(a) < len(b) {
		a, b = b, a
	}
	out := make([]byte, len(a)+1)
	carry := 0
	for i := 0; i < len(a); i++ {
		da := int(a[len(a)-1-i] - '0')
		db := 0
		if i < len(b) {
			db = int(b[len(b)-1-i] - '0')
		}
		sum := da + db + carry
		out[len(out)-1-i] = byte('0' + sum%10)
		carry = sum / 10
	}
	out[0] = byte('0' + carry)
	return string(out)
}

// subDigits computes a − b for digit strings with a >= b.
func subDigits(a, b string) string {
	out := make([]byte, len(a))
	borrow := 0
	for i := 0; i < len(a); i++ {
		da := int(a[len(a)-1-i] - '0')
		db := 0
		if i < len(b) {
			db = int(b[len(b)-1-i] - '0')
		}
		diff := da - db - borrow
		if diff < 0 {
			diff += 10
			borrow = 1
		} else {
			borrow = 0
		}
		out[len(out)-1-i] = byte('0' + diff)
	}
	return string(out)
}
