package exitpolicy_test

// helpers_test.go is the two things the property tests need and the package
// under test deliberately does not export: a whole-number renderer and a decimal
// comparison.
//
// The comparison goes through internal/riskcalc rather than through the package
// being tested. A monotone property checked with the same arithmetic that
// produced the values proves only that the code agrees with itself; checked with
// the ledger's own comparison, it proves the ordering the ledger will see.

import (
	"strconv"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func itoa(n int) string { return strconv.Itoa(n) }

func cmpDecimal(t *testing.T, a, b string) int {
	t.Helper()
	cmp, err := riskcalc.CompareDecimal(a, b)
	if err != nil {
		t.Fatalf("comparing %q and %q: %v", a, b, err)
	}
	return cmp
}
