package reconcile_test

// a083_diff_stamps_test.go holds the helpers that give a test diff the as-of a
// real comparison carries. They live in their own file so that adding them does
// not make the untouched helpers next door look like modified logic.

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// asOfAt renders a comparison's as-of the way Comparer.Compare does.
//
// A credit and the observation that may spend it are ordered by this stamp
// (a083), so a test about release has to move the clock between them exactly as
// the driver's period does. A test that leaves both on the same instant is
// asserting the fail-closed branch, not the release.
func asOfAt(clk clock.Clock) string {
	return clk.Now().UTC().Format(time.RFC3339)
}

func mismatchDiffAt(clk clock.Clock, symbol, local, broker string) reconcile.Diff {
	diff := mismatchDiff(symbol, local, broker)
	diff.AsOf = asOfAt(clk)
	return diff
}

func cleanDiffAt(clk clock.Clock) reconcile.Diff {
	return reconcile.Diff{AsOf: asOfAt(clk), AccountRef: "acct-7", Matched: 1}
}
