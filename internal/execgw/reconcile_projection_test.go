package execgw_test

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// reconcile_projection_test.go covers the entry gate's half of "journal이
// 권위, in-memory 차단 기계는 투영" (task 4.1).
//
// The gate is asked the questions a restarted process asks: is this symbol
// blocked, is the account blocked, and — the one that matters for the operator
// runbook — *why*, because a RECONCILE state and an unresolved IN_DOUBT are
// different problems with different exits.

func projectionGate(t *testing.T) *execgw.EntryGate {
	t.Helper()
	// An empty (non-nil) threshold map: this file is about latches, and a
	// staleness block would answer every question before they were asked.
	return execgw.NewEntryGate(
		clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
		map[execgw.RequiredQuery]time.Duration{})
}

func activeState(account, symbol, cause, evidence string) journal.ReconcileState {
	return journal.ReconcileState{
		ID: "rec-" + symbol + cause, AccountRef: account, Symbol: symbol,
		Cause: cause, Evidence: evidence,
		EnteredAt: time.Date(2026, 3, 30, 0, 29, 0, 0, time.UTC),
	}
}

// TestARebuiltProjectionBlocksTheScopesTheJournalCarries is the restart
// behaviour: a gate that has never seen a disagreement refuses the symbol
// because the journal says so.
func TestARebuiltProjectionBlocksTheScopesTheJournalCarries(t *testing.T) {
	gate := projectionGate(t)

	if rejected := gate.CheckEntryFor("kr", "005930"); rejected != nil {
		t.Fatalf("precondition: a fresh gate blocks nothing, got %v", rejected)
	}

	gate.RebuildReconcileProjection([]journal.ReconcileState{
		activeState("acct-1", "005930", journal.ReconcileCauseQuantityMismatch,
			"the engine believes 10, the account says 4"),
	})

	rejected := gate.CheckEntryFor("kr", "005930")
	if rejected == nil {
		t.Fatal("a restart must not lose the block; the journal still carries the state")
	}
	if rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %s, want %s", rejected.Reason, execgw.ReasonReconcileMismatch)
	}
	// The scope is honoured: one symbol's disagreement does not stop the rest
	// of the account (§ the reconciliation state table).
	if other := gate.CheckEntryFor("kr", "000660"); other != nil {
		t.Fatalf("a symbol-scoped state must not block another symbol, got %v", other)
	}
}

// TestASymbolStateBlocksEveryMarket: reconcile_states carries no market column
// (D9), so a symbol-scoped state blocks that symbol wherever it trades. That is
// the conservative reading of a scope the schema does not record.
func TestASymbolStateBlocksEveryMarket(t *testing.T) {
	gate := projectionGate(t)
	gate.RebuildReconcileProjection([]journal.ReconcileState{
		activeState("acct-1", "AAPL", journal.ReconcileCauseQuantityMismatch, "disagreement"),
	})
	for _, market := range []string{"kr", "us", ""} {
		if gate.CheckEntryFor(market, "AAPL") == nil {
			t.Errorf("AAPL must be blocked in market %q", market)
		}
	}
}

func TestAnAccountWideStateBlocksEverything(t *testing.T) {
	gate := projectionGate(t)
	gate.RebuildReconcileProjection([]journal.ReconcileState{
		activeState("acct-1", "", journal.ReconcileCauseSnapshotUnavailable, "holdings unreadable"),
	})
	if gate.CheckEntry() == nil {
		t.Fatal("an account-wide RECONCILE state must block the account")
	}
	if gate.CheckEntryFor("kr", "005930") == nil {
		t.Fatal("an account-wide state covers every symbol")
	}
}

// TestTheProjectionReplacesRatherThanMerges: a latch this process raised but
// that the journal does not carry is a latch nothing can release, because the
// row an operator would close does not exist.
func TestTheProjectionReplacesRatherThanMerges(t *testing.T) {
	gate := projectionGate(t)
	gate.BlockSymbol("kr", "005930", execgw.ReasonReconcileMismatch, "a block from a previous life")
	gate.Block(execgw.ReasonReconcilePermanent, "and an account-wide one")

	gate.RebuildReconcileProjection(nil)

	if rejected := gate.CheckEntryFor("kr", "005930"); rejected != nil {
		t.Fatalf("the journal carries no state, so the gate must carry none: %v", rejected)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("the account-wide reconcile latch must be replaced too: %v", rejected)
	}
}

// TestTheProjectionLeavesOtherProducersAlone: fill detection and the flatten
// saga raise their own blocks, and rebuilding the reconcile view is not an
// opinion about theirs.
func TestTheProjectionLeavesOtherProducersAlone(t *testing.T) {
	gate := projectionGate(t)
	gate.BlockSymbol("kr", "005930", execgw.ReasonBrokerStateUnknown, "fill detection could not derive a state")
	gate.Block(execgw.ReasonFlattenInProgress, "a flatten saga ran")

	gate.RebuildReconcileProjection(nil)

	if gate.CheckEntry() == nil {
		t.Fatal("the flatten latch must survive a reconcile rebuild")
	}
	gate.Clear(execgw.ReasonFlattenInProgress)
	rejected := gate.CheckEntryFor("kr", "005930")
	if rejected == nil || rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("fill detection's symbol block must survive a reconcile rebuild, got %v", rejected)
	}
}

// TestOperatorOnlyCausesProjectToTheOperatorOnlyReason keeps the two release
// semantics apart. An identifier in conflicting contexts, or a broker record
// nothing local explains, is not disproved by looking again — openapi documents
// CANCEL_REJECTED/REPLACE_REJECTED as separate order records whose shape is
// [형태 미측정 — 2b 2.1] — so those must not land on the reason a clean
// reconciliation clears.
func TestOperatorOnlyCausesProjectToTheOperatorOnlyReason(t *testing.T) {
	autoClearable := []string{
		journal.ReconcileCauseSnapshotUnavailable,
		journal.ReconcileCauseSnapshotStale,
		journal.ReconcileCauseQuantityMismatch,
	}
	operatorOnly := []string{
		journal.ReconcileCauseIdentifierConflict,
		journal.ReconcileCauseAttributionFailed,
		"A_CAUSE_FROM_A_NEWER_BUILD",
	}
	for _, cause := range autoClearable {
		if got := execgw.ReconcileReasonFor(cause); got != execgw.ReasonReconcileMismatch {
			t.Errorf("cause %s maps to %s, want %s", cause, got, execgw.ReasonReconcileMismatch)
		}
	}
	for _, cause := range operatorOnly {
		if got := execgw.ReconcileReasonFor(cause); got != execgw.ReasonReconcilePermanent {
			t.Errorf("cause %s maps to %s, want the operator-only %s",
				cause, got, execgw.ReasonReconcilePermanent)
		}
	}
}

// TestReconcileAndUnresolvedAreNotConflated: the gateway refuses an entry on an
// UNRESOLVED_IN_DOUBT symbol from the journal's attempt table (checkSymbolFree),
// and on a RECONCILE symbol from this projection. They are different reasons
// with different exits — an operator resolving an attempt does not clear a
// reconciliation, and vice versa — so nothing in the projection may produce the
// other's code.
func TestReconcileAndUnresolvedAreNotConflated(t *testing.T) {
	gate := projectionGate(t)
	gate.RebuildReconcileProjection([]journal.ReconcileState{
		activeState("acct-1", "005930", journal.ReconcileCauseQuantityMismatch, "disagreement"),
		activeState("acct-1", "000660", journal.ReconcileCauseIdentifierConflict, "order 42 on two symbols"),
		activeState("acct-1", "", journal.ReconcileCauseSnapshotStale, "holdings 40s old"),
	})

	for _, block := range gate.SymbolBlocks() {
		if block.Reason == execgw.ReasonUnresolvedInDoubt {
			t.Errorf("the RECONCILE projection produced %s on %s; that code belongs to the "+
				"attempt table's unresolved block", block.Reason, block.Symbol)
		}
	}
	for reason := range gate.Blocks() {
		if reason == execgw.ReasonUnresolvedInDoubt {
			t.Errorf("the RECONCILE projection produced the account latch %s", reason)
		}
	}

	// And every cause it did produce is a reconcile-family code, so an operator
	// reading the reason knows which runbook to open.
	blocks := gate.SymbolBlocks()
	if len(blocks) != 2 {
		t.Fatalf("want the two symbol states, got %d: %+v", len(blocks), blocks)
	}
	for _, block := range blocks {
		switch block.Reason {
		case execgw.ReasonReconcileMismatch, execgw.ReasonReconcilePermanent:
		default:
			t.Errorf("unexpected reason %s on %s", block.Reason, block.Symbol)
		}
	}
}

// TestAReleasedStateIsNotProjected: the caller may hand over a history slice;
// only the active rows block.
func TestAReleasedStateIsNotProjected(t *testing.T) {
	gate := projectionGate(t)
	released := activeState("acct-1", "005930", journal.ReconcileCauseQuantityMismatch, "was a disagreement")
	released.ReleasedAt = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)
	released.ReleaseCause = journal.ReconcileReleaseRecheckMatched

	gate.RebuildReconcileProjection([]journal.ReconcileState{released})
	if rejected := gate.CheckEntryFor("kr", "005930"); rejected != nil {
		t.Fatalf("a released state must not block: %v", rejected)
	}
}
