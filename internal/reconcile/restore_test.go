package reconcile_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// restore_test.go is the restart test task 4.1 asks for: 재시작 차단 유지.
//
// A tracker and a gate are built, a disagreement is observed, and then both are
// thrown away — the way a process restart throws away everything in memory. New
// ones are built against the *same journal file*, restored, and asked the same
// question. If the answer changes, the engine would open a position into an
// account it still disagrees with.

func trackerOn(clk clock.Clock, gate *execgw.EntryGate, j *journal.Journal) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock:       clk,
		Gate:        gate,
		Journal:     j,
		MinInterval: 30 * time.Second,
		MaxFailures: 3,
		AccountRef:  "acct-7",
	}
}

// noStaleGate blocks on latches only; a staleness block would answer every
// question before the reconcile state was consulted.
func noStaleGate(clk clock.Clock) *execgw.EntryGate {
	return execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
}

func TestARestartKeepsTheReconcileBlock(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, first)

	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("precondition: the disagreement must block AAPL")
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("precondition: the gate must carry the block")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// --- restart: everything in memory is gone -------------------------------

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)

	if restartedGate.CheckEntryFor("us", "AAPL") != nil {
		t.Fatal("precondition: a fresh gate starts empty, which is what Restore has to fix")
	}
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	rejected := restartedGate.CheckEntryFor("us", "AAPL")
	if rejected == nil {
		t.Fatal("the block was lost across the restart; the journal still carries the disagreement")
	}
	if rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %s, want %s", rejected.Reason, execgw.ReasonReconcileMismatch)
	}
	if restarted.EntryAllowed("us", "AAPL") == nil {
		t.Fatal("the tracker's own view must be restored too")
	}
	// Unrelated symbols are still tradable: the restore honours the scope.
	if restartedGate.CheckEntryFor("us", "MSFT") != nil {
		t.Fatal("a symbol-scoped block must not survive as an account-wide one")
	}
}

func TestARestartKeepsAPermanentMismatchPermanent(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	tracker := trackerOn(clk, noStaleGate(clk), first)
	for i := 0; i < 3; i++ {
		if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
			t.Fatalf("Observe %d: %v", i, err)
		}
		clk.Advance(31 * time.Second)
	}
	if !tracker.Permanent() {
		t.Fatal("precondition: three consecutive failures make the mismatch permanent")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if !restarted.Permanent() {
		t.Fatal("a permanent mismatch must survive a restart as permanent")
	}
	if restartedGate.CheckEntry() == nil {
		t.Fatal("a permanent mismatch is account-wide; the account must be blocked")
	}

	// The dangerous case: one clean pass after the restart must not release a
	// block only an operator can clear.
	if _, err := restarted.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if !restarted.Permanent() {
		t.Fatal("a clean reconciliation must not clear a permanent mismatch")
	}
	if restartedGate.CheckEntry() == nil {
		t.Fatal("the account block must survive a clean pass after a restart")
	}
}

func TestAnOperatorResolutionSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"

	first := openJournalAt(t, path)
	tracker := trackerOn(clk, noStaleGate(clk), first)
	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if err := tracker.Resolve(ctx, "daniel", "compared the app to the journal by hand"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second := openJournalAt(t, path)
	restartedGate := noStaleGate(clk)
	restarted := trackerOn(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if rejected := restartedGate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Fatalf("a resolved block must not come back on restart: %v", rejected)
	}
	if len(restarted.Blocks()) != 0 {
		t.Fatalf("the tracker restored blocks an operator had cleared: %+v", restarted.Blocks())
	}
}

// TestRestoreProjectsStatesThisTrackerDidNotEnter: the gate is the projection of
// the *whole* table, not only of this tracker's rows. An identifier conflict
// recorded by the resolution path still has to block after a restart.
func TestRestoreProjectsStatesThisTrackerDidNotEnter(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)

	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "AAPL",
		Cause:    journal.ReconcileCauseIdentifierConflict,
		Evidence: "order 42 turned up under two symbols",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}

	gate := noStaleGate(clk)
	tracker := trackerOn(clk, gate, j)
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	rejected := gate.CheckEntryFor("us", "AAPL")
	if rejected == nil {
		t.Fatal("a state another producer recorded must still block after a restart")
	}
	if rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("an identifier conflict is operator-only; reason = %s", rejected.Reason)
	}
	// It is not this tracker's row, so its own block set stays empty and a
	// clean pass of *its* comparison must not release it.
	if len(tracker.Blocks()) != 0 {
		t.Fatalf("the tracker adopted somebody else's state: %+v", tracker.Blocks())
	}
	if _, err := tracker.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("a clean quantity comparison released an identifier conflict it says nothing about")
	}
}

// TestWriteThroughRecordsTheStateWithItsEvidence pins that the journal — not
// the block map — is where the disagreement lives.
func TestWriteThroughRecordsTheStateWithItsEvidence(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	tracker := trackerOn(clk, noStaleGate(clk), j)

	if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("want one persisted state, got %+v", active)
	}
	state := active[0]
	if state.Symbol != "AAPL" || state.Cause != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("persisted state = %+v", state)
	}
	if state.Evidence == "" {
		t.Fatal("the persisted state must carry the evidence an operator acts on")
	}

	// A clean pass closes it, and the journal says why.
	clk.Advance(31 * time.Second)
	if _, err := tracker.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if active, err = j.ActiveReconcileStates(ctx); err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("a clean reconciliation must close the state, got %+v", active)
	}
	history, err := j.ReconcileStateHistory(ctx, "acct-7")
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	if len(history) != 1 || history[0].ReleaseCause != journal.ReconcileReleaseRecheckMatched {
		t.Fatalf("want one released state with a recorded cause, got %+v", history)
	}
}
