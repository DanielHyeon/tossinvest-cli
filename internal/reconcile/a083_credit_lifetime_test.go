package reconcile_test

// a083_credit_lifetime_test.go pins how long an adjustment credit lives.
//
// The rule the reconciliation spec states is "비영구 차단의 자동 해제는 조정
// 이벤트가 반영된 뒤의 재조회 일치에만 근거한다". Everything here is about the
// two words the old implementation could not check: *그 뒤의*. A credit carries
// the as-of of the comparison it was computed from, and only an observation of a
// strictly later comparison is the re-read that rule means.
//
// Why this file exists at all: the driver converges a disagreeing comparison and
// then observes that same comparison in the same cycle. Without the stamp, that
// observation spent the credit before any re-read existed, and
// ADJUSTMENT_APPLIED became unreachable — zero of them in the production ledger
// against eight blocks that never lifted.

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// TestTheSameComparisonNeitherSpendsNorDiscardsACredit is the defect itself,
// in one function: the driver's order, at the tracker's level.
func TestTheSameComparisonNeitherSpendsNorDiscardsACredit(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	// Cycle 1, exactly as ReconcileDriver.RunOnce runs it: converge the
	// comparison, then observe that same comparison.
	comparison := mismatchDiffAt(clk, "AAPL", "10", "4")
	tracker.AdjustmentApplied(comparison.AsOf, "AAPL")
	out := observe(t, tracker, comparison)

	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want nothing: this comparison is the one the adjustment "+
			"was computed from, not a re-read of it", out.Cleared)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("the block must stand through the cycle that wrote the adjustment")
	}

	// Cycle 2: the re-read after the adjustment. The credit survived cycle 1 and
	// is spent here.
	clk.Advance(30 * time.Second)
	out = observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" {
		t.Fatalf("cleared = %+v, want the credit kept by cycle 1 to release here "+
			"(awaiting = %+v)", out.Cleared, out.AwaitingAdjustment)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Fatalf("the released symbol must be tradable again, got %v", rejected)
	}
}

// TestAnUndatedComparisonCannotSpendACredit: an observation that cannot say when
// it was true cannot be shown to come after the adjustment. Fail closed.
func TestAnUndatedComparisonCannotSpendACredit(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)

	observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")
	clk.Advance(30 * time.Second)

	out := observe(t, tracker, reconcile.Diff{AccountRef: "acct-7", Matched: 1})
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want an undated comparison to release nothing", out.Cleared)
	}
	if len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("awaiting = %+v, want the block held", out.AwaitingAdjustment)
	}
	if tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("the block must still stand")
	}
}

// TestACreditFromALaterComparisonIsNotSpent: a credit stamped after the
// observation cannot be the adjustment this observation is a re-read of.
func TestACreditFromALaterComparisonIsNotSpent(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)

	observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))
	observing := cleanDiffAt(clk)

	// The credit belongs to a comparison collected after the one being observed.
	clk.Advance(30 * time.Second)
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")

	out := observe(t, tracker, observing)
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want a credit from a later comparison to release nothing", out.Cleared)
	}
	if tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("the block must still stand")
	}
}

// TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit is the hole the
// adversarial review of a083 found in a083's own first design.
//
// Expiry has to be per symbol. AAPL's adjustment is answered by what the re-read
// says about AAPL; MSFT disagreeing says nothing about it. Discarding AAPL's
// credit over MSFT would strand AAPL forever — nothing writes an adjustment for
// a symbol the comparison already agrees about, so it could never earn another.
func TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	// Both symbols disagree and both are converged.
	both := reconcile.Diff{
		AsOf:       asOfAt(clk),
		AccountRef: "acct-7",
		Quantities: []reconcile.QuantityMismatch{
			{Symbol: "AAPL", Local: "10", Broker: "4"},
			{Symbol: "MSFT", Local: "5", Broker: "2"},
		},
	}
	tracker.AdjustmentApplied(both.AsOf, "AAPL", "MSFT")
	observe(t, tracker, both)
	if len(tracker.Blocks()) != 2 {
		t.Fatalf("blocks = %+v, want both symbols blocked", tracker.Blocks())
	}

	// The re-read agrees about AAPL and still disagrees about MSFT. The whole
	// comparison blocks entry, so nothing is released — but AAPL's credit must
	// survive, because this comparison answered it.
	clk.Advance(30 * time.Second)
	observe(t, tracker, reconcile.Diff{
		AsOf:       asOfAt(clk),
		AccountRef: "acct-7",
		Quantities: []reconcile.QuantityMismatch{{Symbol: "MSFT", Local: "5", Broker: "3"}},
	})

	// MSFT is re-converged; AAPL is not, and cannot be — it agrees.
	tracker.AdjustmentApplied(asOfAt(clk), "MSFT")

	clk.Advance(30 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 2 {
		t.Fatalf("cleared = %+v (awaiting = %+v), want both released: AAPL on the credit it "+
			"kept through MSFT's disagreement, MSFT on its own re-convergence",
			out.Cleared, out.AwaitingAdjustment)
	}
	for _, symbol := range []string{"AAPL", "MSFT"} {
		if rejected := gate.CheckEntryFor("us", symbol); rejected != nil {
			t.Errorf("%s must be tradable again, got %v", symbol, rejected)
		}
	}
}

// TestARefutedCreditIsDiscardedPerSymbol keeps the other half of the rule: a
// re-read that still disagrees about the credited symbol spends the credit for
// nothing, and a later coincidence must not resurrect it.
func TestARefutedCreditIsDiscardedPerSymbol(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)

	observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")

	clk.Advance(30 * time.Second)
	observe(t, tracker, mismatchDiffAt(clk, "AAPL", "9", "4"))

	clk.Advance(30 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want the refuted credit not to release a later coincidence",
			out.Cleared)
	}
	if tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("the block must still stand")
	}
}
