package riskcalc_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// decimal_test.go pins the exact arithmetic the reservation ledger is required
// to use (order-execution "원자적 위험 예약": 예약 산술은 decimal 문자열
// 연산이며 float 누적을 사용하지 않는다 — SHALL NOT).
//
// The tests that matter are the ones a float64 implementation fails. Every
// other case here is scaffolding that keeps those honest.

// TestTenTenthsSumToExactlyOne is the case the requirement exists for.
//
// Ten reservations of "0.1" is not a contrived input: it is a fractional-share
// account taking ten positions. Accumulated in float64 the sum is
// 0.9999999999999999, which is under a limit of "1" and therefore admits an
// eleventh reservation the limit was written to refuse.
func TestTenTenthsSumToExactlyOne(t *testing.T) {
	t.Parallel()

	total := "0"
	for i := 0; i < 10; i++ {
		next, err := riskcalc.AddDecimal(total, "0.1")
		if err != nil {
			t.Fatalf("AddDecimal(%q, 0.1): %v", total, err)
		}
		total = next
	}
	if total != "1" {
		t.Fatalf("ten tenths must sum to exactly 1, got %q", total)
	}

	// The float64 control: the same accumulation, to show the assertion above
	// is not vacuous.
	var float float64
	for i := 0; i < 10; i++ {
		float += 0.1
	}
	if float == 1 {
		t.Fatal("float64 accumulation happens to be exact on this platform; " +
			"the test above no longer distinguishes the two implementations")
	}

	cmp, err := riskcalc.CompareDecimal(total, "1")
	if err != nil {
		t.Fatalf("CompareDecimal: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("the exact total must compare equal to the limit, got %d", cmp)
	}
	if strconv.FormatFloat(float, 'f', -1, 64) == "1" {
		t.Fatal("expected the float control to print as something other than 1")
	}
}

// TestBeyondFloat64Precision covers amounts a KRW account reaches: 2^53 is
// about 9×10^15 won, and above it float64 cannot represent a fractional part at
// all, so two different amounts become the same number.
func TestBeyondFloat64Precision(t *testing.T) {
	t.Parallel()

	held := "10000000000000000.4"
	limit := "10000000000000000.5"

	cmp, err := riskcalc.CompareDecimal(held, limit)
	if err != nil {
		t.Fatalf("CompareDecimal: %v", err)
	}
	if cmp != -1 {
		t.Fatalf("%s must be below %s, got %d", held, limit, cmp)
	}

	a, _ := strconv.ParseFloat(held, 64)
	b, _ := strconv.ParseFloat(limit, 64)
	if a != b {
		t.Skip("float64 distinguishes these values on this platform; the case is no longer discriminating")
	}
}

func TestAddDecimal(t *testing.T) {
	t.Parallel()

	cases := []struct{ a, b, want string }{
		{"0", "0", "0"},
		{"1", "2", "3"},
		{"0.1", "0.2", "0.3"},
		{"1.05", "2.95", "4"},
		{"9.99", "0.01", "10"},
		{"-1.5", "1.5", "0"},
		{"-1.5", "0.5", "-1"},
		{"1.5", "-2.5", "-1"},
		{"-1.5", "-2.5", "-4"},
		{"0.0001", "0.0002", "0.0003"},
		{"999999999999999999999", "1", "1000000000000000000000"},
		{"1.10", "0.90", "2"},
	}
	for _, c := range cases {
		got, err := riskcalc.AddDecimal(c.a, c.b)
		if err != nil {
			t.Fatalf("AddDecimal(%q,%q): %v", c.a, c.b, err)
		}
		if got != c.want {
			t.Errorf("AddDecimal(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestSubDecimal(t *testing.T) {
	t.Parallel()

	cases := []struct{ a, b, want string }{
		{"3", "1", "2"},
		{"0.3", "0.1", "0.2"},
		{"1", "1", "0"},
		{"1", "2", "-1"},
		{"-1", "-2", "1"},
		{"10", "0.0001", "9.9999"},
		{"1000000000000000000000", "1", "999999999999999999999"},
	}
	for _, c := range cases {
		got, err := riskcalc.SubDecimal(c.a, c.b)
		if err != nil {
			t.Fatalf("SubDecimal(%q,%q): %v", c.a, c.b, err)
		}
		if got != c.want {
			t.Errorf("SubDecimal(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestCompareDecimalIgnoresSpelling(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b string
		want int
	}{
		{"1.10", "1.1", 0},
		{"01", "1", 0},
		{"+1", "1", 0},
		{"0", "-0", 0},
		{"0.0", "0", 0},
		{"2", "10", -1},
		{"10", "2", 1},
		{"-10", "-2", -1},
		{"-1", "1", -1},
	}
	for _, c := range cases {
		got, err := riskcalc.CompareDecimal(c.a, c.b)
		if err != nil {
			t.Fatalf("CompareDecimal(%q,%q): %v", c.a, c.b, err)
		}
		if got != c.want {
			t.Errorf("CompareDecimal(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestMinMaxDecimal(t *testing.T) {
	t.Parallel()

	min, err := riskcalc.MinDecimal("7.50", "7.5000")
	if err != nil {
		t.Fatalf("MinDecimal: %v", err)
	}
	if min != "7.5" {
		t.Errorf("MinDecimal = %q, want the canonical 7.5", min)
	}
	max, err := riskcalc.MaxDecimal("0", "-3")
	if err != nil {
		t.Fatalf("MaxDecimal: %v", err)
	}
	if max != "0" {
		t.Errorf("MaxDecimal(0,-3) = %q, want 0", max)
	}
}

func TestNegativeZeroIsNeverProduced(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"-0", "-0.0", "-0.000"} {
		got, err := riskcalc.CanonicalDecimal(in)
		if err != nil {
			t.Fatalf("CanonicalDecimal(%q): %v", in, err)
		}
		if got != "0" {
			t.Errorf("CanonicalDecimal(%q) = %q, want 0", in, got)
		}
	}
	diff, err := riskcalc.SubDecimal("1.5", "1.5")
	if err != nil {
		t.Fatalf("SubDecimal: %v", err)
	}
	if diff != "0" {
		t.Errorf("1.5 − 1.5 = %q, want 0", diff)
	}
}

func TestMalformedDecimalsAreRefused(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "   ", "abc", "1.2.3", "1e5", "1,000", "-", ".", "+"} {
		if _, err := riskcalc.AddDecimal(in, "1"); !errors.Is(err, riskcalc.ErrNotDecimal) {
			t.Errorf("AddDecimal(%q, 1) must report ErrNotDecimal, got %v", in, err)
		}
	}
	// A bare point with digits on one side is a decimal; only a lone point is not.
	for _, in := range []string{"1.", ".5"} {
		if _, err := riskcalc.CanonicalDecimal(in); err != nil {
			t.Errorf("CanonicalDecimal(%q) must be accepted: %v", in, err)
		}
	}
}

func TestIsNegativeDecimal(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]bool{"-0.1": true, "0": false, "-0": false, "5": false} {
		got, err := riskcalc.IsNegativeDecimal(in)
		if err != nil {
			t.Fatalf("IsNegativeDecimal(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("IsNegativeDecimal(%q) = %v, want %v", in, got, want)
		}
	}
}
