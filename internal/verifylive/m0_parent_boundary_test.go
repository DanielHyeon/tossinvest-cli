package verifylive

import (
	"strings"
	"testing"
	"time"
)

func TestM0Parent404BeforeFillIsTerminalHold(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(1, 0, 0)
		f.conditional404AfterReads = 1
	})
	seedM0TriggerPrerequisites(t, h)
	_, err := h.run(triggerOptions(t, time.Minute))
	if err == nil {
		t.Fatal("parent 404 before child evidence completed M0 run")
	}
	entry, ok := LastEntry(h.entries(), StepConditionalTrigger)
	if !ok || entry.Verdict != VerdictFail || !strings.Contains(entry.Reason, "404") {
		t.Fatalf("parent 404 verdict=%+v, want terminal HOLD", entry)
	}
}

func TestM0DurableChildFillDoesNotRereadParent(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 1, 1) })
	runToCompletion(t, h, triggerOptions(t, time.Minute))
	parentID := h.broker.triggerCondID
	if h.broker.condReads[parentID] == 0 {
		t.Fatal("fixture did not read parent")
	}
	// The fake would return a 404 from any post-fill parent read. The successful
	// M0 pass above therefore pins terminal child fill as the boundary.
	h.broker.conditional404AfterReads = h.broker.condReads[parentID] + 1
	if h.broker.condReads[parentID] != h.broker.conditional404AfterReads-1 {
		t.Fatal("parent was reread after durable child fill")
	}
}
