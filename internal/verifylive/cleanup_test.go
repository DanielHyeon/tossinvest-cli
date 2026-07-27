package verifylive

// cleanup_test.go pins the property the 2026-07-27 US session proved the tool did
// not have: it could create a live order it was then unable to remove.
//
// The sequence that produced it is reproduced here rather than hand-written into a
// record file, because the bug was in the *interaction* between three rules that
// are each correct on their own — a cancel that failed, a terminal verdict that a
// resume skips, and an exposure cap counted off the record — and a hand-seeded
// record would let one of them drift without a test noticing.

import (
	"context"
	"strings"
	"testing"
)

// leftover runs one invocation whose every cancel is refused, which is what leaves
// an order this tool created resting on the account.
func leftover(t *testing.T, h *harness, broker *fakeBroker) Artifact {
	t.Helper()
	broker.cancelAlreadyProcessing = 99
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("first invocation ended with %v (expected: it could not cancel)", err)
	}
	broker.cancelAlreadyProcessing = 0

	out := Outstanding(h.entries())
	var orders []Artifact
	for _, a := range out {
		if a.Kind == "order" {
			orders = append(orders, a)
		}
	}
	if len(orders) != 1 {
		t.Fatalf("this test needs exactly one leftover order to work with, got %d: %+v", len(orders), out)
	}
	return orders[0]
}

func TestALeftoverOrderIsCancelledOnTheNextRun(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())
	left := leftover(t, h, broker)

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("second invocation: %v", err)
	}

	for _, a := range Outstanding(h.entries()) {
		if a.Kind == "order" && a.ID == left.ID {
			t.Fatalf("the order this tool could not cancel is still outstanding after a second run: %+v", a)
		}
	}
}

func TestTheLeftoverCancelIsOnTheApprovedList(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)
	left := leftover(t, h, broker)

	runner := h.runner(t, Options{HoldingSymbol: "005930"})
	plan := runner.Plan(context.Background())

	var found bool
	for _, m := range plan.Mutations {
		if m.Step == StepCleanup && m.Kind == MutateCancelOrder {
			found = true
			if m.MaxQuantity != 0 {
				t.Errorf("a cancel carries no quantity, got MaxQuantity=%v", m.MaxQuantity)
			}
		}
	}
	if !found {
		t.Fatalf("the plan does not list the cancel of leftover order %s, so approving it would not authorise "+
			"the one request the account needs:\n%+v", left.ID, plan.Mutations)
	}
	if !plan.Authorises(StepCleanup, MutateCancelOrder, left.Symbol, "", 0) {
		t.Errorf("the plan lists the cleanup but does not authorise it for %s", left.Symbol)
	}
}

// TestNoLeftoversMeansNoCleanupLines keeps the plan honest on the ordinary path:
// a clean record must not show an operator a cancel that has nothing to cancel.
func TestNoLeftoversMeansNoCleanupLines(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())

	runner := h.runner(t, Options{HoldingSymbol: "005930"})
	for _, m := range runner.Plan(context.Background()).Mutations {
		if m.Step == StepCleanup {
			t.Fatalf("a run with nothing outstanding planned a cleanup line: %+v", m)
		}
	}
}

// TestTheRedoThatWasBlockedByTheLeftoverCanRun is the symptom the operator
// actually met: order-cancel failed, its order stayed resting, and the exposure
// cap then refused the very re-measurement that would have fixed it.
func TestTheRedoThatWasBlockedByTheLeftoverCanRun(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())
	leftover(t, h, broker)

	if v := h.verdict(StepOrderCancel); v != VerdictFail {
		t.Fatalf("this test needs order-cancel to have failed, got %q", v)
	}

	if _, err := h.run(Options{
		HoldingSymbol: "005930",
		Redo:          []StepID{StepOrderCancel},
	}); err != nil {
		t.Logf("redo invocation: %v", err)
	}

	if v := h.verdict(StepOrderCancel); v != VerdictPass {
		e, _ := LastEntry(h.entries(), StepOrderCancel)
		t.Fatalf("the re-measurement is still %q: %s", v, e.Reason)
	}
	if out := Outstanding(h.entries()); len(out) > 0 {
		t.Errorf("the account is left holding %+v", out)
	}
}

// TestTheConditionalLeftForPersistenceIsNotCleanedUp is the one artifact the
// prologue must leave alone: cancelling it would delete the thing the next step
// exists to observe.
func TestTheConditionalLeftForPersistenceIsNotCleanedUp(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())

	// One invocation gets as far as conditional-persist, which stops the run with
	// the conditional deliberately alive.
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("first invocation: %v", err)
	}
	if h.verdict(StepConditionalPersist) != VerdictAwaitingRestart {
		t.Fatalf("this test needs the run to stop at conditional-persist, got %q", h.verdict(StepConditionalPersist))
	}
	var conditional string
	for _, a := range Outstanding(h.entries()) {
		if a.Kind == "conditional-order" {
			conditional = a.ID
		}
	}
	if conditional == "" {
		t.Fatal("this test needs a conditional order left alive by the persistence design")
	}

	runner := h.runner(t, Options{HoldingSymbol: "005930"})
	for _, m := range runner.Plan(context.Background()).Mutations {
		if m.Step == StepCleanup && m.Kind == MutateCancelConditional {
			t.Fatalf("the prologue planned to cancel the conditional order the next step has to read: %+v", m)
		}
	}
}

// TestARefusedBatchSendsNoCleanup: the prologue is a live request like any other.
func TestARefusedBatchSendsNoCleanup(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())
	left := leftover(t, h, broker)

	refusing := newHarness(t, broker, refuseBatch(ErrRefused))
	refusing.record = h.record
	broker.requests = nil
	if _, err := refusing.run(Options{HoldingSymbol: "005930"}); err == nil {
		t.Log("refused batch returned no error")
	}

	for _, req := range broker.requests {
		if strings.Contains(req, "/cancel") {
			t.Fatalf("a refused approval still sent a cancel: %q (order %s)", req, left.ID)
		}
	}
}

// TestAFailedCleanupIsRecordedAndDoesNotStopTheRun keeps the failure honest: the
// run carries on, the exposure cap still refuses what it always refused, and
// nothing reports the leftover as gone.
func TestAFailedCleanupIsRecordedAndDoesNotStopTheRun(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())
	left := leftover(t, h, broker)

	broker.cancelAlreadyProcessing = 99
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("second invocation: %v", err)
	}

	entries := h.entries()
	var cleanup *Entry
	for i := range entries {
		if entries[i].StepID == StepCleanup {
			cleanup = &entries[i]
		}
	}
	if cleanup == nil {
		t.Fatal("a cleanup that ran and failed left no line on the record")
	}
	if cleanup.Verdict == VerdictPass {
		t.Errorf("a cleanup whose cancel was refused was recorded as %q", cleanup.Verdict)
	}
	var still bool
	for _, a := range Outstanding(entries) {
		if a.ID == left.ID {
			still = true
		}
	}
	if !still {
		t.Error("a failed cleanup stopped reporting the order that is still on the account")
	}
	// The run kept walking: steps after the prologue still produced verdicts.
	if _, ok := LastEntry(entries, StepConditionalCancel); !ok {
		t.Error("the run stopped at the failed cleanup instead of measuring what it still could")
	}
}

// TestCleanupIsNotAMeasuredStep: every reader that walks the procedure has to
// skip it, or a housekeeping line would read as a measured capability.
func TestCleanupIsNotAMeasuredStep(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())
	leftover(t, h, broker)

	before := StepCount(h.entries())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("second invocation: %v", err)
	}
	entries := h.entries()

	var cleanup *Entry
	for i := range entries {
		if entries[i].StepID == StepCleanup {
			cleanup = &entries[i]
		}
	}
	if cleanup == nil {
		t.Fatal("no cleanup line was written")
	}
	if cleanup.Kind != KindCleanup {
		t.Errorf("the cleanup line is kind %q, so step readers will count it as a measurement", cleanup.Kind)
	}

	stepsAfter := StepCount(entries)
	measured := 0
	for _, e := range entries {
		if e.Kind == KindStep && e.StepID != StepCleanup {
			measured++
		}
	}
	if stepsAfter != measured {
		t.Errorf("StepCount counts %d lines but only %d are measurements (before this run: %d)",
			stepsAfter, measured, before)
	}
	for _, id := range RedoSet(entries) {
		if id == StepCleanup {
			t.Error("the redo set offers to re-measure the cleanup, which measures nothing")
		}
	}
}
