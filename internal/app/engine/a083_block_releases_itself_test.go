package engine_test

// a083_block_releases_itself_test.go is change a083's RED정본: the driver cycle
// that proves a converged quantity mismatch releases its own block.
//
// The unit tests in internal/reconcile could not catch this. They call
// AdjustmentApplied outside a cycle and then observe a clean diff, which is the
// driver's order reversed:
//
//	테스트     Observe(mismatch) → Converge(mismatch) → Observe(clean)
//	드라이버   Converge(mismatch) → Observe(mismatch) → [다음 주기] Observe(clean)
//	                                ~~~~~~~~~~~~~~~~ 테스트에 없던 단계
//
// That missing step is the whole defect: the observation that immediately
// follows the credit is the *same* pre-adjustment comparison, and it used to
// discard the credit before the next cycle's agreeing re-read could spend it.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// TestTheCycleAfterAConvergenceReleasesTheBlock walks the two cycles the live
// ledger has been stuck between since 2026-08-03.
func TestTheCycleAfterAConvergenceReleasesTheBlock(t *testing.T) {
	ctx := context.Background()
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	if first := h.cycle(); first.Err != nil {
		t.Fatalf("the fold/adopt cycle failed: %v", first.Err)
	}
	if got := h.position("005930").Quantity; got != "10" {
		t.Fatalf("projection = %q after the fold, want 10", got)
	}

	// The owner sells six shares in the app. The account is the authority.
	h.holdings.items[0].Quantity = "4"

	// --- cycle 1: the disagreement is seen, converged, and observed ----------
	converging := h.cycle()
	if converging.Err != nil {
		t.Fatalf("the converging cycle failed: %v", converging.Err)
	}
	if converging.Converged != 1 {
		t.Fatalf("converged = %d, want the mismatch folded onto the account's number",
			converging.Converged)
	}
	if got := h.position("005930").Quantity; got != "4" {
		t.Fatalf("projection = %q after the convergence, want the account's 4", got)
	}
	if converging.Blocked != 1 || converging.Released != 0 {
		t.Fatalf("converging cycle = {blocked:%d released:%d}, want the block raised and nothing released: "+
			"the comparison this cycle observed is the one the adjustment was computed from",
			converging.Blocked, converging.Released)
	}

	// --- cycle 2: the re-read after the adjustment agrees --------------------
	releasing := h.cycle()
	if releasing.Err != nil {
		t.Fatalf("the releasing cycle failed: %v", releasing.Err)
	}
	if releasing.Released != 1 {
		t.Fatalf("released = %d, want the block released by the re-read after the adjustment "+
			"(blocked=%d)", releasing.Released, releasing.Blocked)
	}
	if releasing.Blocked != 0 {
		t.Errorf("blocked = %d after the release, want none", releasing.Blocked)
	}

	states, err := h.journal.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("active reconcile states = %+v, want the durable row closed", states)
	}
	if rejected := h.tracker.EntryAllowed("kr", "005930"); rejected != nil {
		t.Errorf("the released symbol must be tradable again, got %v", rejected)
	}
}

// TestTheConvergingCycleKeepsItsCreditForTheNextRead is the same walk seen from
// the credit's side: the observation that shares a comparison with the
// adjustment must leave the credit alone rather than spend it for nothing.
func TestTheConvergingCycleKeepsItsCreditForTheNextRead(t *testing.T) {
	ctx := context.Background()
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	if first := h.cycle(); first.Err != nil {
		t.Fatalf("the fold/adopt cycle failed: %v", first.Err)
	}
	h.holdings.items[0].Quantity = "4"

	if cycle := h.cycle(); cycle.Converged != 1 {
		t.Fatalf("converged = %d (%v)", cycle.Converged, cycle.Err)
	}

	// The block is durable and the gate is latched at this point: the engine has
	// written the adjustment but has not yet re-read the account.
	states, err := h.journal.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(states) != 1 || states[0].Symbol != "005930" {
		t.Fatalf("active states = %+v, want exactly the converging symbol blocked", states)
	}
	if states[0].Cause != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("cause = %q, want %s", states[0].Cause, journal.ReconcileCauseQuantityMismatch)
	}
	if rejected := h.tracker.EntryAllowed("kr", "005930"); rejected == nil {
		t.Fatal("the block must still stand in the cycle that wrote the adjustment")
	} else if rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %q, want %s", rejected.Reason, execgw.ReasonReconcileMismatch)
	}

	// And the very next cycle spends the credit it kept.
	if cycle := h.cycle(); cycle.Released != 1 {
		t.Fatalf("released = %d on the re-read after the adjustment, want 1 (%v)",
			cycle.Released, cycle.Err)
	}
}
