package console

// retry_after_run_test.go covers the dead end that cost three live-market windows
// on 2026-07-27 (measurements.md M11, M18).
//
// Every one of them ended the same way: an approval window lapsed, the run
// finished having sent nothing, and the verification screen then had no button on
// it at all — the start controls were rendered only when no run existed, so the
// only way back was restarting the console. These tests hold the screen to the
// rule that a finished run is not a dead end, without loosening either of the
// guards that legitimately disable the button.

import (
	"net/url"
	"strings"
	"testing"
)

// startForm reports whether the page offers a way to begin a verification.
func startForm(page string) bool {
	return strings.Contains(page, `action="/verify/start"`)
}

// disabledStart reports whether that offer is disabled.
func disabledStart(page string) bool {
	i := strings.Index(page, `action="/verify/start"`)
	if i < 0 {
		return false
	}
	rest := page[i:]
	if end := strings.Index(rest, "</form>"); end >= 0 {
		rest = rest[:end]
	}
	return strings.Contains(rest, "disabled")
}

// TestAFinishedRunLeavesTheStartControlsOnTheScreen.
//
// The run here sends nothing — the operator refuses the batch — so the process is
// not spent and starting again is a legitimate thing to want. Before this change
// the screen answered with no buttons whatsoever.
func TestAFinishedRunLeavesTheStartControlsOnTheScreen(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)
	h.post(t, "/verify/abort", url.Values{"csrf": {h.csrf}})
	h.waitForFinish(t)

	page := body(t, h.get(t, "/verify"))
	if !startForm(page) {
		t.Fatalf("after a run that sent nothing the screen offers no way to try again — the operator has to "+
			"restart the console:\n%s", truncateForLog(page))
	}
	if disabledStart(page) {
		t.Errorf("the start control is disabled although nothing was sent and the process is not spent:\n%s",
			truncateForLog(page))
	}
}

// TestALiveRunHidesTheStartControls. The other half of the rule: while a run is
// waiting for its approval, the action the screen offers is the approval.
func TestALiveRunHidesTheStartControls(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)
	defer h.stopRun()

	page := body(t, h.get(t, "/verify"))
	if startForm(page) {
		t.Errorf("the screen offers to start a second verification while one is waiting for approval:\n%s",
			truncateForLog(page))
	}
}

// TestASpentProcessStillDisablesTheStartControls. Restoring the controls must not
// restore the ability to walk the catalogue twice in one process: conditional
// persistence is measured across a process boundary and cannot be measured
// without one.
func TestASpentProcessStillDisablesTheStartControls(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)
	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}})
	h.waitForFinish(t)

	h.mu.Lock()
	spent := h.spent
	h.mu.Unlock()
	if !spent {
		t.Fatal("this test needs a run that walked the catalogue, so the process is spent")
	}

	page := body(t, h.get(t, "/verify"))
	if !startForm(page) {
		t.Fatalf("a spent process shows no start control at all, so it cannot explain why:\n%s",
			truncateForLog(page))
	}
	if !disabledStart(page) {
		t.Errorf("a spent process offers an enabled start control:\n%s", truncateForLog(page))
	}
}
