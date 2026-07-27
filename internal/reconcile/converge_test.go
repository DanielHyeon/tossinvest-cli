package reconcile_test

// converge_test.go is the other half of task 6.3's release rule: a block lifts
// on "the re-read after an adjustment agreed", and something has to issue the
// adjustment. Without this the automatic release is unreachable on the path it
// was designed for and every ordinary quantity disagreement becomes an operator
// ticket (issues.md, task 6.3 "운영상 대가").

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

var convergeNow = time.Date(2026, 3, 30, 2, 30, 0, 0, time.UTC)

type recordingCrediter struct{ credited [][]string }

func (c *recordingCrediter) AdjustmentApplied(symbols ...string) {
	c.credited = append(c.credited, append([]string(nil), symbols...))
}

// heldPosition puts a filled buy in the projection.
func heldPosition(t *testing.T, j *journal.Journal, symbol, quantity string) {
	t.Helper()
	confirmedOrder(t, j, "intent-"+symbol, "attempt-"+symbol, "o-"+symbol, symbol, "BUY")
	if _, err := j.RecordFill(context.Background(), journal.FillObservation{
		OrderID: "o-" + symbol, Symbol: symbol, Market: "us", State: "FILLED", Terminal: true,
		Quantity: quantity, FilledQuantity: quantity, AveragePrice: "200",
		ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

// stampedMismatch is mismatchDiff with the broker as-of an adjustment needs:
// nothing may be written from a comparison that cannot say when it was true.
func stampedMismatch(symbol, local, broker string) reconcile.Diff {
	diff := mismatchDiff(symbol, local, broker)
	diff.AsOf = "2026-03-30T02:30:00Z"
	return diff
}

// TestAQuantityMismatchConvergesToTheAccount is the normal path: the account is
// the authority and the projection is written to agree with it.
func TestAQuantityMismatchConvergesToTheAccount(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "10", "7"))
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	if len(report.Converged) != 1 {
		t.Fatalf("converged = %+v, want the one mismatch", report.Converged)
	}
	got := report.Converged[0]
	if got.Quantity != "7" || !got.Applied {
		t.Errorf("converged = %+v, want the account's 7, applied", got)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if local.Positions["AAPL"] != "7" {
		t.Errorf("the projection still says %q; the account said 7", local.Positions["AAPL"])
	}
	if len(credit.credited) != 1 || credit.credited[0][0] != "AAPL" {
		t.Errorf("credited = %v, want the adjusted symbol", credit.credited)
	}
}

// TestTheAdjustmentIsUNKNOWNAndNotEXTERNAL: the engine has an instance, so the
// difference is one the journal cannot explain rather than somebody else's
// shares. The two names mean different things to whoever reads the row.
func TestTheAdjustmentIsUNKNOWNAndNotEXTERNAL(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	c := &reconcile.Converger{Journal: j, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "10", "7"))
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	adjustments, err := j.PositionAdjustments(ctx, report.Converged[0].PositionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(adjustments) != 1 {
		t.Fatalf("adjustments = %+v, want one", adjustments)
	}
	if adjustments[0].Kind != "UNKNOWN" {
		t.Errorf("kind = %s, want UNKNOWN", adjustments[0].Kind)
	}
	if adjustments[0].ExpectedPrevQuantity != "10" {
		t.Errorf("expected-prev = %s; the compare half of compare-and-append is the evidence",
			adjustments[0].ExpectedPrevQuantity)
	}
	if adjustments[0].Evidence == "" {
		t.Error("an adjustment carries the evidence the operator has to act on")
	}
}

// TestConvergenceMakesTheBlockReleasable is the point of the whole file: with
// the credit in hand the next agreeing re-read closes the block with
// ADJUSTMENT_APPLIED, which is unreachable without an adjustment.
func TestConvergenceMakesTheBlockReleasable(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	tracker := trackerOn(clock.NewFake(convergeNow), noStaleGate(clock.NewFake(convergeNow)), j)
	if _, err := tracker.Observe(ctx, stampedMismatch("AAPL", "10", "7")); err != nil {
		t.Fatalf("Observe(mismatch): %v", err)
	}
	if len(tracker.Blocks()) != 1 {
		t.Fatalf("blocks = %+v, want the symbol blocked", tracker.Blocks())
	}

	c := &reconcile.Converger{Journal: j, Credit: tracker, AccountRef: "acct-7"}
	if _, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "10", "7")); err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}

	clean := reconcile.Diff{AsOf: "2026-03-30T02:31:00Z", AccountRef: "acct-7", Matched: 1}
	out, err := tracker.Observe(ctx, clean)
	if err != nil {
		t.Fatalf("Observe(clean): %v", err)
	}
	if len(out.Cleared) != 1 {
		t.Fatalf("cleared = %+v, want the adjusted symbol released", out.Cleared)
	}
	if len(tracker.Blocks()) != 0 {
		t.Errorf("blocks = %+v, want none", tracker.Blocks())
	}
}

// TestACoincidentalAgreementStillDoesNotRelease is the rule this file must not
// weaken: the release needs an adjustment, not a lucky re-read.
func TestACoincidentalAgreementStillDoesNotRelease(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	tracker := trackerOn(clock.NewFake(convergeNow), noStaleGate(clock.NewFake(convergeNow)), j)
	if _, err := tracker.Observe(ctx, stampedMismatch("AAPL", "10", "7")); err != nil {
		t.Fatal(err)
	}
	clean := reconcile.Diff{AsOf: "2026-03-30T02:31:00Z", AccountRef: "acct-7", Matched: 1}
	out, err := tracker.Observe(ctx, clean)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Cleared) != 0 || len(tracker.Blocks()) != 1 {
		t.Errorf("an agreement nobody wrote anything for must not release: cleared=%+v blocks=%+v",
			out.Cleared, tracker.Blocks())
	}
}

// TestASymbolWithNoLiveInstanceIsRefusedRatherThanFolded keeps the two writers
// apart: a holding with no local instance is the Ingestor's case.
func TestASymbolWithNoLiveInstanceIsRefusedRatherThanFolded(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)

	c := &reconcile.Converger{Journal: j, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "0", "7"))
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	if len(report.Converged) != 0 {
		t.Fatalf("converged = %+v, want nothing", report.Converged)
	}
	if report.Refused["AAPL"] == "" {
		t.Error("the refusal has to say why, or an operator cannot act on it")
	}
}

// TestAStaleAdjustmentStopsThePass: the rest of the diff came from the same
// view, and that view has been shown to have moved.
func TestAStaleAdjustmentStopsThePass(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	c := &reconcile.Converger{Journal: j, AccountRef: "acct-7"}
	// The expected-prev value is wrong, which is exactly what a view that moved
	// under the comparison looks like from inside the commit.
	_, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "4", "7"))
	if err == nil {
		t.Fatal("an adjustment computed against a quantity the projection does not hold must be refused")
	}
	local, lerr := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if lerr != nil {
		t.Fatal(lerr)
	}
	if local.Positions["AAPL"] != "10" {
		t.Errorf("the projection moved on a discarded adjustment: %q", local.Positions["AAPL"])
	}
}

// TestNoMismatchesIsANoOp: the pass writes nothing and credits nothing, so a
// clean loop cannot manufacture a release.
func TestNoMismatchesIsANoOp(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, reconcile.Diff{AsOf: "x", AccountRef: "acct-7"})
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	if len(report.Converged) != 0 || len(credit.credited) != 0 {
		t.Errorf("a clean diff wrote %+v and credited %v", report.Converged, credit.credited)
	}
}

// TestConvergingAManagedPositionToZeroAlerts is design A7's operator half: when
// the account says a position the engine was protecting is gone, the exit state
// is completed and a person is told.
//
// The alert matters more than it looks. The screen stops listing the position
// and no trade outcome is frozen for it, so without this the only evidence that
// the engine was ever protecting those shares is a completed row nobody reads.
func TestConvergingAManagedPositionToZeroAlerts(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	held, err := j.CurrentPosition(ctx, "acct-7", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: held.ID, Symbol: "AAPL", Market: "us", Quantity: "10",
		CostBasis: "200", ObservedPrice: "210", SyntheticStop: "199.5",
		ObservedAt: "2026-03-30T02:00:00Z",
	}); err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, held.ID); err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}

	alerter := &recordingAlerter{}
	c := &reconcile.Converger{Journal: j, Alert: alerter, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "10", "0"))
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	if report.Closed != 1 {
		t.Fatalf("report.Closed = %d, want 1", report.Closed)
	}
	if len(report.Converged) != 1 || !report.Converged[0].ClosedExitState {
		t.Fatalf("converged = %+v, want one entry reporting the closed exit state", report.Converged)
	}
	if len(alerter.closed) != 1 {
		t.Fatalf("managed-close alerts = %d, want 1", len(alerter.closed))
	}
	got := alerter.closed[0]
	if got.Symbol != "AAPL" || got.PositionID != held.ID || got.PrevQuantity != "10" {
		t.Errorf("alert = %+v, want AAPL/%s/10", got, held.ID)
	}
	if !got.Adopted {
		t.Error("the alert must say the position was one the engine adopted; the operator's next " +
			"question is whether the engine or a person put the protection there")
	}
}

// TestConvergingAnUnmanagedPositionToZeroDoesNotAlert keeps the alert about
// protection ending rather than about a quantity changing.
func TestConvergingAnUnmanagedPositionToZeroDoesNotAlert(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")

	alerter := &recordingAlerter{}
	c := &reconcile.Converger{Journal: j, Alert: alerter, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "10", "0"))
	if err != nil {
		t.Fatalf("ConvergeQuantities: %v", err)
	}
	if report.Closed != 0 || len(alerter.closed) != 0 {
		t.Errorf("closed = %d, alerts = %d; nothing was protecting this position",
			report.Closed, len(alerter.closed))
	}
}
