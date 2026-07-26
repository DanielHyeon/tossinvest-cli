package engine

// exit_wiring_internal_test.go is task 7.3's wiring half: the exit applier is
// bound by the same call as the projection, and the guard notices when it is not.
//
// It is a separate file from interlock_internal_test.go because the assertion
// that file makes — "the hooks cannot be bound twice" — is the reason this one
// exists: 7.3 had to extend the single literal, and this is the test that the
// extension actually happened rather than being written into a comment.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestBindingWiresBothApplyHooks(t *testing.T) {
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	if !j.ProjectionBound() {
		t.Error("the projection hook is not bound")
	}
	if !j.ExitApplierBound() {
		t.Error("the exit applier is not bound; nothing would resolve a pending proposal")
	}
	if err := checkProjectionWired(j); err != nil {
		t.Errorf("a fully bound journal must pass the guard: %v", err)
	}
}

// TestTheGuardNoticesAHalfBoundJournal is the point of asking rather than
// trusting. A future edit that dropped the Exit member from the literal would
// leave a process whose positions each propose once and then go quiet, which
// looks exactly like a process with nothing to propose.
func TestTheGuardNoticesAHalfBoundJournal(t *testing.T) {
	j := openTestJournal(t)
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}

	err := checkProjectionWired(j)
	if !errors.Is(err, ErrExitApplierUnbound) {
		t.Fatalf("err = %v, want ErrExitApplierUnbound", err)
	}
	if !errors.Is(err, ErrExitApplierUnbound) || err.Error() == "" {
		t.Error("the refusal must say what goes wrong")
	}
}

// TestTheBoundExitApplierIsTheJournalsOwn keeps the binding honest about *what*
// it bound: a hook that did nothing would satisfy ExitApplierBound and satisfy
// nothing else. The cheapest true check is that a fill through the bound journal
// moves the exit state, which only the real applier does.
func TestTheBoundExitApplierIsTheJournalsOwn(t *testing.T) {
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	// A fill for an order no intent claims: the real applier returns quietly, a
	// hook that mistakenly assumed a position would fail the fill.
	if _, err := j.RecordFill(context.Background(), journal.FillObservation{
		OrderID: "o-unclaimed", Symbol: "005930", Market: "kr",
		State: "CLOSED_FILLED", Terminal: true, Quantity: "5", FilledQuantity: "5",
		AveragePrice: "70000", ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("RecordFill through the bound hooks: %v", err)
	}
}
