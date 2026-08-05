package reconcile_test

// a083_converge_credit_test.go: the converger is the only production crediter,
// so it is the only place that can say which comparison a credit stands on.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)


// TestTheCreditCarriesTheComparisonItWasComputedFrom: the tracker cannot tell a
// re-read from the read it is answering unless the crediter says which
// comparison it stood on, and the converger is the only production caller.
func TestTheCreditCarriesTheComparisonItWasComputedFrom(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	diff := stampedMismatch("AAPL", "10", "7")
	if _, err := c.ConvergeQuantities(ctx, diff); err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}

	if len(credit.comparisons) != 1 {
		t.Fatalf("comparisons = %+v, want exactly one credit", credit.comparisons)
	}
	if credit.comparisons[0] != diff.AsOf {
		t.Errorf("credited comparison = %q, want the diff's as-of %q; a credit stamped with "+
			"anything else cannot be ordered against the re-read that spends it",
			credit.comparisons[0], diff.AsOf)
	}
}

// TestAReappliedAdjustmentIsStillCredited: what the release rule requires is
// that something was written for this symbol before the re-read, not that this
// process was the one that wrote it.
func TestAReappliedAdjustmentIsStillCredited(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	diff := stampedMismatch("AAPL", "10", "7")
	if _, err := c.ConvergeQuantities(ctx, diff); err != nil {
		t.Fatalf("first ConvergeQuantities: %v", err)
	}
	report, err := c.ConvergeQuantities(ctx, diff)
	if err != nil {
		t.Fatalf("second ConvergeQuantities: %v", err)
	}
	if len(report.Converged) != 1 || report.Converged[0].Applied {
		t.Fatalf("converged = %+v, want the same adjustment recognised as already on disk",
			report.Converged)
	}
	if len(credit.comparisons) != 2 || credit.comparisons[1] != diff.AsOf {
		t.Errorf("comparisons = %+v, want the re-applied adjustment credited with the same as-of",
			credit.comparisons)
	}
}
