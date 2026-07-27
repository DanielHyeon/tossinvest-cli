package exitpolicy

// decimal_internal_test.go pins the rendering rules, which are safety rules
// rather than formatting ones: the direction a non-terminating value rounds in
// decides whether a rounding error can weaken a protective level.
//
// It is an internal test because the functions are not part of the package's
// contract — what callers see is that a baseline is a decimal string — and
// exporting them so a black-box test could reach them would make the rounding
// direction something a caller could pick.

import (
	"math/big"
	"testing"
)

func TestATerminatingValueIsPrintedExactly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		num, den int64
		want     string
	}{
		{9900, 1, "9900"},
		{1, 8, "0.125"},
		{-1, 2, "-0.5"},
		{10030, 1, "10030"},
		{0, 1, "0"},
		{1, 1024, "0.0009765625"},
		{3, 400, "0.0075"},
	}
	for _, tc := range cases {
		got := formatPrice(big.NewRat(tc.num, tc.den))
		if got != tc.want {
			t.Errorf("formatPrice(%d/%d) = %s, want %s", tc.num, tc.den, got, tc.want)
		}
	}
}

// TestANonTerminatingPriceRoundsUp is the §0.9 direction. A protective baseline
// is a floor under a long position, so rounding it up can only make it fire
// sooner; rounding to nearest would place it a hair *below* the policy's answer
// half the time, which is a weakening however small.
func TestANonTerminatingPriceRoundsUp(t *testing.T) {
	t.Parallel()

	// 1/3 = 0.333…: the twelve-digit rendering must be above the true value.
	got := formatPrice(big.NewRat(1, 3))
	if got != "0.333333333334" {
		t.Fatalf("formatPrice(1/3) = %s, want it rounded up", got)
	}
	rendered, ok := new(big.Rat).SetString(got)
	if !ok {
		t.Fatalf("the rendering %q is not a decimal", got)
	}
	if rendered.Cmp(big.NewRat(1, 3)) <= 0 {
		t.Error("the rendered price is not above the exact one; the protection was weakened by rounding")
	}

	// 2/3 = 0.666… rounds to nearest *up* anyway; the case that separates the two
	// rules is one whose nearest neighbour is below it.
	if up := formatPrice(big.NewRat(1, 30000000000000)); up != "0.000000000001" {
		t.Errorf("formatPrice(tiny) = %s, want the smallest representable step up", up)
	}
}

// TestANegativePriceStillRoundsUp — towards zero for a negative value, which is
// still "not below the exact value". No price in TossOS is negative, but the R
// multipliers are and the same helper renders them during composition.
func TestANegativePriceStillRoundsUp(t *testing.T) {
	t.Parallel()

	got := formatPrice(new(big.Rat).Neg(big.NewRat(1, 3)))
	rendered, _ := new(big.Rat).SetString(got)
	if rendered.Cmp(new(big.Rat).Neg(big.NewRat(1, 3))) < 0 {
		t.Errorf("formatPrice(−1/3) = %s, which is below the exact value", got)
	}
}

// TestANonTerminatingRatioRoundsUp is the same argument for a taken fraction: a
// larger recorded take suppresses more future take-profit proposals and
// understates what is left, both of which are the quiet direction.
func TestANonTerminatingRatioRoundsUp(t *testing.T) {
	t.Parallel()

	got := formatRatio(big.NewRat(1, 3))
	rendered, _ := new(big.Rat).SetString(got)
	if rendered.Cmp(big.NewRat(1, 3)) <= 0 {
		t.Errorf("formatRatio(1/3) = %s, want it above the exact value", got)
	}
}

// TestARatioNeverRoundsUpToTheWholePosition guards the one case where rounding
// up is the wrong direction: claiming the position is fully taken when a sliver
// remains would let `taken_ratio_total` reach 1 while quantity is still held.
func TestARatioNeverRoundsUpToTheWholePosition(t *testing.T) {
	t.Parallel()

	almost := new(big.Rat).Sub(one, big.NewRat(1, 100000000000000000))
	got := formatRatio(almost)
	if got == "1" {
		t.Fatal("a fraction below 1 rendered as the whole position")
	}
	rendered, _ := new(big.Rat).SetString(got)
	if rendered.Cmp(one) >= 0 {
		t.Errorf("formatRatio = %s, which is not below 1", got)
	}
}

func TestAWholeRatioIsExactlyOne(t *testing.T) {
	t.Parallel()

	if got := formatRatio(one); got != "1" {
		t.Errorf("formatRatio(1) = %s, want 1", got)
	}
}

// TestAnRMultipleRoundsToNearest — it is an audit number. Every decision that
// depended on it was made on the exact rational, so rounding it in a direction
// would misreport rather than protect.
func TestAnRMultipleRoundsToNearest(t *testing.T) {
	t.Parallel()

	if got := formatRMultiple(big.NewRat(1, 3)); got != "0.333333" {
		t.Errorf("formatRMultiple(1/3) = %s, want 0.333333", got)
	}
	if got := formatRMultiple(big.NewRat(2, 3)); got != "0.666667" {
		t.Errorf("formatRMultiple(2/3) = %s, want 0.666667", got)
	}
	if got := formatRMultiple(big.NewRat(2, 1)); got != "2" {
		t.Errorf("formatRMultiple(2) = %s, want the exact spelling", got)
	}
}

func TestExponentsAreRefused(t *testing.T) {
	t.Parallel()

	for _, v := range []string{"1e4", "1E4", "5/2", "1.2.3", "", "  ", "abc", "1."} {
		if _, err := parseRat("value", v); err == nil {
			t.Errorf("parseRat(%q) was accepted; a value with two spellings breaks text comparison", v)
		}
	}
}

func TestTheEmptySpellingOfZeroIsOnlyAcceptedWhereItIsDefined(t *testing.T) {
	t.Parallel()

	if _, err := parseRat("value", ""); err == nil {
		t.Error("parseRat accepted the empty string; an empty decimal is unknown, not zero")
	}
	got, err := parseRatOr("taken ratio", "", "0")
	if err != nil {
		t.Fatalf("parseRatOr: %v", err)
	}
	if got.Sign() != 0 {
		t.Errorf("parseRatOr(\"\") = %v, want zero", got)
	}
}
