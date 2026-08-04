package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// reconcile_states_test.go pins the durable RECONCILE state (task 4.1).
//
// The Manager's condition on task 0.1 is the first test here: the column has no
// CHECK constraint, so the storage layer has to refuse an unknown cause at
// write time. Everything else follows from "one active state per scope".

func TestUnknownReconcileCauseIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	for _, cause := range []string{"", "SOMETHING_ELSE", "quantity_mismatch", " "} {
		_, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
			AccountRef: "acct-1", Symbol: "005930", Cause: cause, Evidence: "observed",
		})
		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("cause %q must be refused at write time (the column carries no CHECK), got %v", cause, err)
		}
	}

	// Every enumerated cause is accepted, so the refusal above is the list and
	// not an accident.
	for i, cause := range []string{
		ReconcileCauseSnapshotUnavailable, ReconcileCauseSnapshotStale,
		ReconcileCauseQuantityMismatch, ReconcileCauseIdentifierConflict,
		ReconcileCauseAttributionFailed,
	} {
		if !ValidReconcileCause(cause) {
			t.Fatalf("%s must be a valid cause", cause)
		}
		if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
			AccountRef: "acct-1", Symbol: "SYM" + string(rune('A'+i)),
			Cause: cause, Evidence: "observed",
		}); err != nil {
			t.Errorf("cause %s must be accepted: %v", cause, err)
		}
	}
}

func TestReconcileStateRequiresEvidence(t *testing.T) {
	j, _ := openReservationJournal(t)
	_, _, err := j.EnterReconcile(context.Background(), EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930", Cause: ReconcileCauseQuantityMismatch,
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a state an operator cannot act on must be refused, got %v", err)
	}
}

func TestReEnteringAnActiveScopeKeepsTheFirstObservation(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	first, entered, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseQuantityMismatch, Evidence: "engine 10, account 4",
	})
	if err != nil || !entered {
		t.Fatalf("EnterReconcile: entered=%v err=%v", entered, err)
	}

	fake.Advance(time.Minute)
	again, entered, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseSnapshotStale, Evidence: "a later, different observation",
	})
	if err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	if entered {
		t.Fatal("re-entering an active scope must report that it was already active")
	}
	if again.ID != first.ID || again.Cause != first.Cause || !again.EnteredAt.Equal(first.EnteredAt) {
		t.Fatalf("re-entry replaced the original state: %+v vs %+v", again, first)
	}

	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("one active state per scope, got %d", len(active))
	}
}

func TestAccountWideAndSymbolScopesAreSeparate(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Cause: ReconcileCauseSnapshotUnavailable, Evidence: "holdings unreadable",
	}); err != nil {
		t.Fatalf("account-wide EnterReconcile: %v", err)
	}
	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseQuantityMismatch, Evidence: "engine 10, account 4",
	}); err != nil {
		t.Fatalf("symbol EnterReconcile: %v", err)
	}

	// A second account-wide state is what the extra partial UNIQUE index from
	// task 0.1 exists to refuse; EnterReconcile answers before reaching it.
	_, entered, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Cause: ReconcileCauseSnapshotStale, Evidence: "another one",
	})
	if err != nil {
		t.Fatalf("second account-wide EnterReconcile: %v", err)
	}
	if entered {
		t.Fatal("a second active account-wide state must not be entered")
	}

	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("want the account-wide and the symbol state, got %d: %+v", len(active), active)
	}
}

func TestReleaseRequiresACauseAndEvidence(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseQuantityMismatch, Evidence: "engine 10, account 4",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}

	for _, req := range []ReleaseReconcileRequest{
		{AccountRef: "acct-1", Symbol: "005930", Cause: "", Evidence: "matched"},
		{AccountRef: "acct-1", Symbol: "005930", Cause: "BECAUSE", Evidence: "matched"},
		{AccountRef: "acct-1", Symbol: "005930", Cause: ReconcileReleaseRecheckMatched},
	} {
		if _, _, err := j.ReleaseReconcile(ctx, req); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("release %+v must be refused, got %v", req, err)
		}
	}

	state, released, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileReleaseRecheckMatched, Evidence: "a later read agreed: both say 4",
	})
	if err != nil || !released {
		t.Fatalf("ReleaseReconcile: released=%v err=%v", released, err)
	}
	if state.ReleaseCause != ReconcileReleaseRecheckMatched {
		t.Fatalf("release cause = %q", state.ReleaseCause)
	}

	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("the released state is still active: %+v", active)
	}

	// The history keeps both halves of the story.
	history, err := j.ReconcileStateHistory(ctx, "acct-1")
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	if len(history) != 1 || !strings.Contains(history[0].Evidence, "both say 4") {
		t.Fatalf("the release evidence must survive in the history, got %+v", history)
	}
}

// TestExpectCauseReleasesOnlyYourOwnState is what keeps a producer from
// closing somebody else's block: a clean quantity comparison says nothing about
// an identifier conflict.
func TestExpectCauseReleasesOnlyYourOwnState(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseIdentifierConflict, Evidence: "order 42 seen on two symbols",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}

	_, released, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileReleaseRecheckMatched, Evidence: "quantities agreed",
		ExpectCause: ReconcileCauseQuantityMismatch,
	})
	if err != nil {
		t.Fatalf("ReleaseReconcile: %v", err)
	}
	if released {
		t.Fatal("a quantity re-check must not release an identifier conflict")
	}
	active, _ := j.ActiveReconcileStates(ctx)
	if len(active) != 1 {
		t.Fatalf("the identifier conflict must still be active, got %+v", active)
	}
}

// TestUnknownReleaseCauseIsRefused is the other half of the Manager's condition
// on task 0.1: `release_cause` carries no CHECK either, so the storage layer is
// what refuses one nobody enumerated. Task 6.3 extends the set with
// ADJUSTMENT_APPLIED, and this test is the enumeration.
func TestUnknownReleaseCauseIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	for _, cause := range []string{"", "ADJUSTED", "adjustment_applied", " ", "CLEAN"} {
		if ValidReconcileReleaseCause(cause) {
			t.Errorf("release cause %q must not be one this build writes", cause)
		}
		if _, _, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
			AccountRef: "acct-1", Symbol: "005930", Cause: cause, Evidence: "matched",
		}); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("release cause %q must be refused at write time, got %v", cause, err)
		}
	}

	for _, cause := range []string{
		ReconcileReleaseRecheckMatched, ReconcileReleaseAdjustmentApplied,
		ReconcileReleaseOperator,
	} {
		if !ValidReconcileReleaseCause(cause) {
			t.Errorf("%s must be a valid release cause", cause)
		}
	}
}

// TestAnAdjustmentAppliedReleaseIsRecorded walks the release cause task 6.3
// added all the way to disk: the row closes, and the history says an adjustment
// is what closed it rather than a coincidence.
func TestAnAdjustmentAppliedReleaseIsRecorded(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileCauseQuantityMismatch, Evidence: "engine 10, account 4",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}

	state, released, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause:       ReconcileReleaseAdjustmentApplied,
		Evidence:    "adjustment converged the projection to 4 and the re-read agreed",
		ExpectCause: ReconcileCauseQuantityMismatch,
	})
	if err != nil || !released {
		t.Fatalf("ReleaseReconcile: released=%v err=%v", released, err)
	}
	if state.ReleaseCause != ReconcileReleaseAdjustmentApplied {
		t.Fatalf("release cause = %q, want %s", state.ReleaseCause, ReconcileReleaseAdjustmentApplied)
	}

	history, err := j.ReconcileStateHistory(ctx, "acct-1")
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	if len(history) != 1 || history[0].ReleaseCause != ReconcileReleaseAdjustmentApplied {
		t.Fatalf("history = %+v, want the adjustment recorded as the cause", history)
	}
	if !strings.Contains(history[0].Evidence, "converged the projection") {
		t.Fatalf("the release evidence must survive: %+v", history[0])
	}
}

func TestReleasingNothingIsNotAnError(t *testing.T) {
	j, _ := openReservationJournal(t)
	_, released, err := j.ReleaseReconcile(context.Background(), ReleaseReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930",
		Cause: ReconcileReleaseRecheckMatched, Evidence: "agreed",
	})
	if err != nil {
		t.Fatalf("releasing an inactive scope must not be an error: %v", err)
	}
	if released {
		t.Fatal("nothing was active, so nothing was released")
	}
}

func TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	for _, req := range []EnterReconcileRequest{
		{AccountRef: "acct-1", Cause: ReconcileCauseQuantityMismatch, Evidence: "permanent account guard"},
		{AccountRef: "acct-1", Symbol: "005930", Cause: ReconcileCauseIdentifierConflict, Evidence: "symbol identity conflict"},
	} {
		if _, _, err := j.EnterReconcile(ctx, req); err != nil {
			t.Fatalf("EnterReconcile(%q): %v", req.Symbol, err)
		}
	}

	_, err := j.ReleaseReconciles(ctx, []ReleaseReconcileRequest{
		{AccountRef: "acct-1", Cause: ReconcileReleaseOperator, Evidence: "operator verified", ExpectCause: ReconcileCauseQuantityMismatch},
		// Deliberately wrong: the second scope is owned by IDENTIFIER_CONFLICT.
		{AccountRef: "acct-1", Symbol: "005930", Cause: ReconcileReleaseOperator, Evidence: "operator verified", ExpectCause: ReconcileCauseQuantityMismatch},
	})
	if err == nil || !strings.Contains(err.Error(), "owned by") {
		t.Fatalf("ReleaseReconciles error = %v, want exact-cause refusal", err)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("atomic refusal released part of the guard set: %+v", active)
	}
	for _, state := range active {
		if !state.ReleasedAt.IsZero() {
			t.Fatalf("state %s was partially released: %+v", state.ID, state)
		}
	}
}
