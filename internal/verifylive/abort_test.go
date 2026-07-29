package verifylive

// abort_test.go is the escape hatch, and most of what it asserts is what the
// escape hatch is *not*.
//
// The hold rule (verify-holds-what-it-awaits) says objects are released by a
// verdict and never by a clock, and that rule was chosen over a lease on purpose:
// a long wait means the market has not come to the price yet, and a lease that
// expired mid-measurement would cancel the subject of the measurement. This is the
// other half of that decision — the person saying "I am ending this" — and it has
// to stay the *only* other half. A time-based path reintroduced here would undo
// the choice quietly.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// seedRecord writes entries into a harness's record as if an earlier run had.
func seedRecord(t *testing.T, h *harness, entries ...Entry) {
	t.Helper()
	rec, err := OpenRecorder(h.record)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()
	for _, e := range entries {
		if e.Kind == "" {
			e.Kind = KindStep
		}
		if err := rec.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

// heldChain is what an interrupted trigger measurement leaves on the record: a
// conditional that fired, a child order nobody has decided about, and no terminal
// verdict for the step that would release it.
func heldChain(at time.Time) []Entry {
	return []Entry{{
		Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictFail,
		Artifacts: []Artifact{
			{
				Kind: KindOrder, ID: "child-7", Symbol: "005930", CreatedAt: at,
				Deliberate: true, HeldUntil: StepConditionalTrigger, ChainID: "chain-7",
			},
			{
				Kind: KindConditional, ID: "co-7", Symbol: "005930", CreatedAt: at,
				Deliberate: true, HeldUntil: StepConditionalCancel, ChainID: "chain-7",
			},
		},
	}}
}

func TestAbortEndsAHeldChain(t *testing.T) {
	h := triggerHarness(t, nil)
	at := time.Date(2026, 7, 30, 4, 0, 0, 0, time.UTC)
	seedRecord(t, h, heldChain(at)...)
	h.broker.orders["child-7"] = mustOrderJSON("child-7", "005930", "SELL", "PENDING", 1, 70000, "")
	h.broker.conds["co-7"] = domain.ConditionalOrder{
		ID: "co-7", Type: "SINGLE", Status: "WATCHING", Symbol: "005930", Quantity: 1, OrderType: "MARKET",
		First: domain.ConditionalOrderCondition{Type: "STOP", Status: "WATCHING", TriggerPrice: 56000},
	}

	r := h.runner(t, Options{})
	result, err := r.Abort(context.Background(), "장이 닫혔고 오늘은 여기까지")
	if err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if !result.Approved {
		t.Fatal("the abort was not approved although the operator approves everything here")
	}
	if len(result.Targets) != 2 {
		t.Fatalf("targets = %+v, want both held objects — the whole point is reaching what the hold rule "+
			"protects", result.Targets)
	}
	if len(result.Remaining) != 0 {
		t.Errorf("still live afterwards: %+v", result.Remaining)
	}

	entries := h.entries()
	if out := Outstanding(entries); len(out) != 0 {
		t.Errorf("the record still says %+v is live", out)
	}
	// The chain is closed with a reason. Cancelling the objects and saying nothing
	// would leave the evidence describing a measurement that simply stops, which is
	// the shape of M37.
	if !observationEquals(t, entries, StepAbort, "chain.closed.chain-7", "operator-abort") {
		t.Error("the chain the two objects belonged to was not recorded as closed")
	}
	if d := observationDetail(t, entries, StepAbort, "chain.closed.chain-7"); !strings.Contains(d, "장이 닫혔고") {
		t.Errorf("the closure records no reason: %q", d)
	}
	e, ok := LastEntry(entries, StepAbort)
	if !ok {
		t.Fatal("the abort wrote no line to the record")
	}
	if e.Kind != KindCleanup {
		t.Errorf("kind = %q, want %q — an abort measures nothing and every reader that walks the "+
			"procedure has to skip it", e.Kind, KindCleanup)
	}
}

// TestAbortReachesWhatTheCleanupPrologueMayNot is the difference between the two
// paths said in one assertion. The prologue is deliberately blind to a held
// object; the abort exists because something has to see it.
func TestAbortReachesWhatTheCleanupPrologueMayNot(t *testing.T) {
	at := time.Date(2026, 7, 30, 4, 0, 0, 0, time.UTC)
	entries := heldChain(at)
	if pending := PendingCleanup(entries); len(pending) != 0 {
		t.Fatalf("PendingCleanup = %+v, want nothing: the objects are held and no releasing verdict was "+
			"recorded after the line that named the gate", pending)
	}
	if targets := AbortTargets(entries); len(targets) != 2 {
		t.Fatalf("AbortTargets = %+v, want both", targets)
	}
}

// TestAbortIsNeverDecidedByElapsedTime.
//
// The rejected alternative, restated as a test. Two identical records that differ
// only in how long ago they were written have to produce the same target list —
// and a record with nothing outstanding has to produce an empty one however old
// it is, because "it has been a while" is not a reason to cancel anything.
func TestAbortIsNeverDecidedByElapsedTime(t *testing.T) {
	fresh := AbortTargets(heldChain(time.Now().Add(-time.Minute)))
	stale := AbortTargets(heldChain(time.Now().AddDate(0, -1, 0)))
	if len(fresh) != len(stale) {
		t.Fatalf("a month-old record produced %d targets and a minute-old one %d; elapsed time is not an "+
			"input to this decision", len(stale), len(fresh))
	}
	for i := range fresh {
		if fresh[i].ID != stale[i].ID {
			t.Errorf("target %d differs by age alone: %s vs %s", i, fresh[i].ID, stale[i].ID)
		}
	}

	// And an already-finished chain stays finished. A cancelled object is not
	// resurrected by an abort, however long ago it was cancelled.
	done := heldChain(time.Now().AddDate(-1, 0, 0))
	done = append(done, Entry{Kind: KindStep, StepID: StepAbort, Artifacts: []Artifact{
		{Kind: KindOrder, ID: "child-7", Symbol: "005930", Cancelled: true},
		{Kind: KindConditional, ID: "co-7", Symbol: "005930", Cancelled: true},
	}})
	if targets := AbortTargets(done); len(targets) != 0 {
		t.Errorf("AbortTargets = %+v on a record where everything is already cancelled", targets)
	}
}

// TestAbortSendsNothingWhenItIsNotApproved. It is a live cancel like any other, so
// it goes through the batch rail — and the batch rail's answer is respected.
func TestAbortSendsNothingWhenItIsNotApproved(t *testing.T) {
	h := triggerHarness(t, nil)
	h.op = refuseBatch(ErrRefused)
	at := time.Date(2026, 7, 30, 4, 0, 0, 0, time.UTC)
	seedRecord(t, h, heldChain(at)...)

	r := h.runner(t, Options{})
	result, err := r.Abort(context.Background(), "")
	if err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if result.Approved {
		t.Fatal("a refused abort reported itself approved")
	}
	if n := h.broker.countRequests("DELETE /conditional-orders/"); n != 0 {
		t.Errorf("%d conditional cancel(s) were sent after a refusal", n)
	}
	if n := h.broker.countRequests("POST /orders/child-7/cancel"); n != 0 {
		t.Error("an order cancel was sent after a refusal")
	}
	if len(result.Remaining) != 2 {
		t.Errorf("remaining = %+v, want both objects still reported live", result.Remaining)
	}
}

func TestAbortWithNothingLiveTouchesNothing(t *testing.T) {
	h := triggerHarness(t, nil)
	r := h.runner(t, Options{})
	result, err := r.Abort(context.Background(), "")
	if err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if result.Approved || len(result.Targets) != 0 {
		t.Errorf("result = %+v, want nothing to do", result)
	}
	for _, req := range h.broker.seen() {
		if strings.HasPrefix(req, "DELETE") || strings.Contains(req, "/cancel") {
			t.Errorf("a cancel was sent against an empty record: %s", req)
		}
	}
}

// TestNothingBesidesAbortCancelsAHeldObject is the structural half.
//
// sweepStep cancels what a step left resting, and it runs on every path a step
// does not abort on — including the one where the trigger observation fails
// having seen the fire. The child order it holds on that path is the evidence, and
// a sweep that reached it would delete the measurement at the moment it mattered.
func TestNothingBesidesAbortCancelsAHeldObject(t *testing.T) {
	h := triggerHarness(t, nil)
	r := h.runner(t, Options{})
	sr := &stepRun{step: mustStep(t, StepConditionalTrigger)}
	sr.created(KindOrder, "child-9", "005930", time.Now(), "")
	sr.markHeld(KindOrder, "child-9", StepConditionalTrigger, "chain-9", "held")

	r.sweepStep(context.Background(), sr)
	if n := h.broker.countRequests("POST /orders/child-9/cancel"); n != 0 {
		t.Errorf("sweepStep cancelled a held order %d time(s). Letting the child fill IS the measurement, "+
			"and the path that needs it most is the one where the step already failed", n)
	}

	// The ordinary case is unchanged: an order left resting by accident is swept.
	loose := &stepRun{step: mustStep(t, StepOrderCancel)}
	loose.created(KindOrder, "ord-9", "005930", time.Now(), "")
	h.broker.orders["ord-9"] = mustOrderJSON("ord-9", "005930", "BUY", "PENDING", 1, 56000, "")
	r.plan = &Plan{Mutations: []PlannedMutation{
		{Step: StepOrderCancel, Kind: MutateCancelOrder, Symbol: "005930"},
	}}
	r.sweepStep(context.Background(), loose)
	if n := h.broker.countRequests("POST /orders/ord-9/cancel"); n != 1 {
		t.Errorf("an ordinary leftover was cancelled %d time(s), want 1 — the sweep still has to work", n)
	}
}
