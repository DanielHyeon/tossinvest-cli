package verifylive

// transient_test.go covers what the 2026-07-28 US re-measurement found: the
// broker's `already-processing` refusal is not specific to cancels, and a step
// that fails after placing an order was walking away from it.
//
// The two are tested together because they combine: the amend was refused, the
// step returned, and the order it had just placed stayed resting and filled the
// exposure cap that the next step needed.

import (
	"strings"
	"testing"
)

func TestAnAmendRefusedAsTransientIsRetried(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	broker.modifyAlreadyProcessing = 1
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("run: %v", err)
	}

	if v := h.verdict(StepOrderAmend); v != VerdictPass {
		e, _ := LastEntry(h.entries(), StepOrderAmend)
		t.Fatalf("order-amend is %q after one transient refusal: %s", v, e.Reason)
	}
	got, ok := h.observation(StepOrderAmend, "order.amend.retries")
	if !ok {
		t.Fatal("an amend that needed a second attempt is recorded as if it worked first time")
	}
	if got != "1" {
		t.Errorf("order.amend.retries = %q, want 1", got)
	}
}

// TestAnAmendRefusalStillClearsTheOrderTheStepPlaced is the defect that blocked
// sell-boundary on the real account: the step returned on the amend error and
// left its own order resting.
func TestAnAmendRefusalStillClearsTheOrderTheStepPlaced(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	broker.modifyAlreadyProcessing = 99
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("run: %v", err)
	}

	if v := h.verdict(StepOrderAmend); v != VerdictFail {
		t.Fatalf("this test needs the amend to fail, got %q", v)
	}
	for _, a := range Outstanding(h.entries()) {
		if a.Kind == KindOrder {
			t.Errorf("the failed step left the order it placed resting: %+v", a)
		}
	}
	// And the step after it was not blocked by that leftover.
	if v := h.verdict(StepSellBoundary); v == VerdictFail {
		e, _ := LastEntry(h.entries(), StepSellBoundary)
		if strings.Contains(e.Reason, ErrExposureCap.Error()) {
			t.Errorf("sell-boundary was blocked by the previous step's leftover: %s", e.Reason)
		}
	}
}

// TestAPlacementRefusedAsTransientIsNotRetried holds the line the amend retry
// must not cross: repeating a placement can create a second live order.
func TestAPlacementRefusedAsTransientIsNotRetried(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	broker.placeAlreadyProcessing = 1
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("run: %v", err)
	}

	// Which step meets the scripted refusal depends on the catalogue order, so
	// the assertion is about the refusal rather than about a step id: whichever
	// step met it failed, and nothing recorded a placement retry.
	var met bool
	for _, e := range h.entries() {
		if strings.Contains(e.Reason, "already-processing") {
			met = true
			if e.Verdict != VerdictFail {
				t.Errorf("%s met a refused placement and recorded %q", e.StepID, e.Verdict)
			}
		}
		for _, o := range e.Observations {
			if strings.Contains(o.Key, "place.retries") {
				t.Errorf("%s recorded a placement retry (%s=%s); placements are never retried",
					e.StepID, o.Key, o.Value)
			}
		}
	}
	if !met {
		t.Fatal("no step met the scripted placement refusal, so this test proves nothing")
	}
}

// TestTheSweepGoesThroughTheGate: the cleanup a failed step does is a live
// request like any other, and a plan that does not carry the step's cancel line
// must not have one sent on its behalf.
func TestTheSweepGoesThroughTheGate(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	broker.modifyAlreadyProcessing = 99
	h := newHarness(t, broker, alwaysConfirm())

	runner := h.runner(t, Options{HoldingSymbol: "005930"})
	if !runner.Plan(t.Context()).Authorises(StepOrderAmend, MutateCancelOrder, "005930", "", 0) {
		t.Fatal("the plan does not carry order-amend's cancel line, so the sweep would have nothing to use")
	}
}
