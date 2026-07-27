package journal

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
)

// TestTransitionTable is the MutationAttempt lifecycle from the order-execution
// spec, as a table. Every pair not listed as legal must be refused: the journal is
// the record a restart reasons from, so a state it could not have reached must not
// be writable.
func TestTransitionTable(t *testing.T) {
	legal := map[AttemptState][]AttemptState{
		StateRecorded: {StateDispatchStarted, StateNotDispatched},
		// A dispatch can end four ways: acknowledged, unknown, provably never
		// sent, or definitively rejected by the broker.
		StateDispatchStarted: {StateAcked, StateInDoubt, StateNotDispatched, StateFailedConfirmed},
		StateAcked:           {StateConfirmed, StateInDoubt},
		StateInDoubt:         {StateConfirmed, StateFailedConfirmed, StateUnresolvedInDoubt},
		// Operator resolution only (no automatic path reaches these).
		StateUnresolvedInDoubt: {StateConfirmed, StateFailedConfirmed},
		StateConfirmed:         nil,
		StateNotDispatched:     nil,
		StateFailedConfirmed:   nil,
	}

	all := []AttemptState{
		StateRecorded, StateDispatchStarted, StateAcked, StateInDoubt,
		StateConfirmed, StateNotDispatched, StateFailedConfirmed, StateUnresolvedInDoubt,
	}

	for from, allowed := range legal {
		allowedSet := map[AttemptState]bool{}
		for _, to := range allowed {
			allowedSet[to] = true
			if err := ValidateTransition(from, to); err != nil {
				t.Errorf("ValidateTransition(%s, %s) = %v, want nil", from, to, err)
			}
		}
		for _, to := range all {
			if allowedSet[to] {
				continue
			}
			if err := ValidateTransition(from, to); !errors.Is(err, ErrIllegalTransition) {
				t.Errorf("ValidateTransition(%s, %s) = %v, want ErrIllegalTransition", from, to, err)
			}
		}

		got := AllowedNext(from)
		sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
		want := append([]AttemptState(nil), allowed...)
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		if len(got) != len(want) {
			t.Errorf("AllowedNext(%s) = %v, want %v", from, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("AllowedNext(%s) = %v, want %v", from, got, want)
				break
			}
		}
	}
}

// TestTransitionTableRejectsSelfAndUnknown covers the edges around the table.
func TestTransitionTableRejectsSelfAndUnknown(t *testing.T) {
	if err := ValidateTransition(StateRecorded, StateRecorded); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("a state must not transition to itself: %v", err)
	}
	if err := ValidateTransition(AttemptState("FROB"), StateAcked); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("unknown from-state must be refused: %v", err)
	}
	if err := ValidateTransition(StateRecorded, AttemptState("FROB")); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("unknown to-state must be refused: %v", err)
	}
	// The initial write is the only way into RECORDED.
	if err := ValidateTransition(AttemptState(""), StateRecorded); err != nil {
		t.Errorf("initial transition into RECORDED must be legal: %v", err)
	}
	if err := ValidateTransition(AttemptState(""), StateDispatchStarted); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("an attempt cannot start at DISPATCH_STARTED: %v", err)
	}
}

// TestIllegalTransitionIsRejectedByTheJournal proves the table is enforced on the
// stored record, not just as a pure function.
func TestIllegalTransitionIsRejectedByTheJournal(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	// RECORDED → CONFIRMED: nothing was ever sent, so this must be unwritable.
	if err := a.Settle(ctx, StateConfirmed, "wishful", ""); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("Settle(CONFIRMED) from RECORDED: want ErrIllegalTransition, got %v", err)
	}
	// RECORDED → NOT_DISPATCHED is the legal safe close.
	if err := a.Settle(ctx, StateNotDispatched, "restart", ""); err != nil {
		t.Fatalf("Settle(NOT_DISPATCHED) from RECORDED: %v", err)
	}
	// A terminal attempt is final.
	if err := a.Settle(ctx, StateConfirmed, "no", ""); err == nil {
		t.Fatal("a terminal attempt must not be re-settled")
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateNotDispatched {
		t.Fatalf("state = %s, want NOT_DISPATCHED", rec.State)
	}
	if got := attemptHistory(t, j, "attempt-1"); len(got) != 2 {
		t.Fatalf("history = %v, want 2 entries (refused transitions must not be recorded)", got)
	}
}

// TestUnresolvedInDoubtIsOperatorOnly pins the one gap between the lifecycle table
// and the automatic API: the table permits UNRESOLVED_IN_DOUBT → CONFIRMED /
// FAILED_CONFIRMED, but no automatic method may take it. An engine that could
// resolve its own permanent block would defeat the point of having one.
func TestUnresolvedInDoubtIsOperatorOnly(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.MarkInDoubt(ctx, "transport", ""); err != nil {
		t.Fatal(err)
	}
	if err := a.Settle(ctx, StateUnresolvedInDoubt, "could_not_prove", ""); err != nil {
		t.Fatal(err)
	}

	// Legal in the table…
	if err := ValidateTransition(StateUnresolvedInDoubt, StateConfirmed); err != nil {
		t.Fatalf("table must keep the operator path open: %v", err)
	}
	// …but not reachable through Settle.
	if err := a.Settle(ctx, StateConfirmed, "operator", ""); !errors.Is(err, ErrUnexpectedState) {
		t.Fatalf("Settle from UNRESOLVED_IN_DOUBT: want ErrUnexpectedState, got %v", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateUnresolvedInDoubt {
		t.Fatalf("state = %s, want UNRESOLVED_IN_DOUBT", rec.State)
	}
}

// TestRecoverPendingAppliesRestartRules is the spec's restart behaviour:
// RECORDED-only attempts are closed as NOT_DISPATCHED with no blocking, and
// anything that had started dispatching becomes IN_DOUBT and blocks its symbol
// until the resolution procedure finishes.
func TestRecoverPendingAppliesRestartRules(t *testing.T) {
	path := tempJournalPath(t)
	j := openTestJournalAt(t, path)
	ctx := context.Background()

	// (1) RECORDED only — crashed before dispatch.
	recorded := testRequest()
	recorded.Intent.ID = "intent-recorded"
	recorded.AttemptID = "attempt-recorded"
	recorded.Intent.Symbol = "AAPL"
	if _, err := j.Prepare(ctx, recorded); err != nil {
		t.Fatal(err)
	}

	// (2) DISPATCH_STARTED — crashed mid-flight.
	dispatched := testRequest()
	dispatched.Intent.ID = "intent-dispatched"
	dispatched.AttemptID = "attempt-dispatched"
	dispatched.Intent.Symbol = "MSFT"
	a2, err := j.Prepare(ctx, dispatched)
	if err != nil {
		t.Fatal(err)
	}
	if err := a2.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}

	// (3) ACKED — the broker took it, we had not confirmed yet.
	acked := testRequest()
	acked.Intent.ID = "intent-acked"
	acked.AttemptID = "attempt-acked"
	acked.Intent.Symbol = "TSLA"
	a3, err := j.Prepare(ctx, acked)
	if err != nil {
		t.Fatal(err)
	}
	if err := a3.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a3.MarkAcked(ctx, "order-3"); err != nil {
		t.Fatal(err)
	}

	// (4) already terminal — must be left alone.
	done := testRequest()
	done.Intent.ID = "intent-done"
	done.AttemptID = "attempt-done"
	done.Intent.Symbol = "NVDA"
	a4, err := j.Prepare(ctx, done)
	if err != nil {
		t.Fatal(err)
	}
	if err := a4.Settle(ctx, StateNotDispatched, "guardian_refused", ""); err != nil {
		t.Fatal(err)
	}

	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	// --- restart ---
	restarted := openTestJournalAt(t, path)
	report, err := restarted.RecoverPending(ctx)
	if err != nil {
		t.Fatalf("RecoverPending: %v", err)
	}

	if strings.Join(report.NotDispatched, ",") != "attempt-recorded" {
		t.Errorf("NotDispatched = %v, want [attempt-recorded]", report.NotDispatched)
	}
	if strings.Join(report.InDoubt, ",") != "attempt-dispatched" {
		t.Errorf("InDoubt = %v, want [attempt-dispatched]", report.InDoubt)
	}

	blocked := map[string]AttemptState{}
	for _, b := range report.Blocked {
		blocked[b.Symbol] = b.State
		if b.Market != "us" {
			t.Errorf("blocked entry %+v: market must come from the intent", b)
		}
	}
	if len(blocked) != 2 {
		t.Fatalf("Blocked = %+v, want MSFT (IN_DOUBT) and TSLA (ACKED)", report.Blocked)
	}
	if blocked["MSFT"] != StateInDoubt {
		t.Errorf("MSFT blocked state = %s, want IN_DOUBT", blocked["MSFT"])
	}
	if blocked["TSLA"] != StateAcked {
		t.Errorf("TSLA blocked state = %s, want ACKED", blocked["TSLA"])
	}
	if _, ok := blocked["AAPL"]; ok {
		t.Error("a RECORDED-only attempt must not block its symbol")
	}

	// Stored states after recovery.
	wantStates := map[string]AttemptState{
		"attempt-recorded":   StateNotDispatched,
		"attempt-dispatched": StateInDoubt,
		"attempt-acked":      StateAcked,
		"attempt-done":       StateNotDispatched,
	}
	for id, want := range wantStates {
		rec, err := restarted.LookupAttempt(ctx, id)
		if err != nil {
			t.Fatalf("LookupAttempt(%s): %v", id, err)
		}
		if rec.State != want {
			t.Errorf("%s state = %s, want %s", id, rec.State, want)
		}
	}

	// The safe close and the doubt are both recorded with a reason an operator
	// can read.
	rec, err := restarted.LookupAttempt(ctx, "attempt-recorded")
	if err != nil {
		t.Fatal(err)
	}
	if rec.ReasonCode != ReasonRestartNotDispatched {
		t.Errorf("reason = %q, want %q", rec.ReasonCode, ReasonRestartNotDispatched)
	}
	if rec.SettledAt == "" {
		t.Error("a safely closed attempt needs a settled_at timestamp")
	}
	rec2, err := restarted.LookupAttempt(ctx, "attempt-dispatched")
	if err != nil {
		t.Fatal(err)
	}
	if rec2.ReasonCode != ReasonRestartInDoubt {
		t.Errorf("reason = %q, want %q", rec2.ReasonCode, ReasonRestartInDoubt)
	}
}

// TestRecoverPendingIsIdempotent covers a restart loop: running recovery twice must
// not move an attempt further or duplicate history.
func TestRecoverPendingIsIdempotent(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}

	first, err := j.RecoverPending(ctx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := j.RecoverPending(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.InDoubt) != 1 {
		t.Fatalf("first pass InDoubt = %v", first.InDoubt)
	}
	if len(second.InDoubt) != 0 {
		t.Errorf("second pass must not re-transition: %v", second.InDoubt)
	}
	if len(second.Blocked) != 1 {
		t.Errorf("an IN_DOUBT attempt must keep blocking after later passes: %+v", second.Blocked)
	}
	if got := attemptHistory(t, j, "attempt-1"); len(got) != 3 {
		t.Fatalf("history = %v, want 3 entries", got)
	}
}

// TestPendingAttempts lists what the resolution engine has to work through.
func TestPendingAttempts(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	second := testRequest()
	second.Intent.ID = "intent-2"
	second.AttemptID = "attempt-2"
	a2, err := j.Prepare(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	if err := a2.Settle(ctx, StateNotDispatched, "r", ""); err != nil {
		t.Fatal(err)
	}

	pending, err := j.PendingAttempts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].ID != a.ID() {
		t.Fatalf("PendingAttempts = %+v, want only %s", pending, a.ID())
	}
}

// TestResumeAttempt lets a later pass (IN_DOUBT resolution, operator action) keep
// working on an attempt recorded by an earlier process.
func TestResumeAttempt(t *testing.T) {
	path := tempJournalPath(t)
	j := openTestJournalAt(t, path)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	restarted := openTestJournalAt(t, path)
	resumed, err := restarted.Resume(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if resumed.State() != StateDispatchStarted || resumed.AttemptNo() != 1 ||
		resumed.IntentID() != "intent-1" || resumed.Kind() != KindPlace {
		t.Fatalf("resumed handle = %+v", resumed)
	}
	if err := resumed.MarkInDoubt(ctx, "restart", "found mid-flight"); err != nil {
		t.Fatalf("MarkInDoubt on a resumed handle: %v", err)
	}
	if _, err := restarted.Resume(ctx, "missing"); !errors.Is(err, ErrAttemptNotFound) {
		t.Errorf("Resume(missing): want ErrAttemptNotFound, got %v", err)
	}
}
