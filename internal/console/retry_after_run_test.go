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
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// seedLeftoverOrder appends the line an order-cancel failure leaves behind: a
// mutating step that created an order and never recorded cancelling it.
func seedLeftoverOrder(t *testing.T, path string) {
	t.Helper()
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	now := time.Now().UTC()
	if err := rec.Append(verifylive.Entry{
		StepID: verifylive.StepOrderCancel, Title: "place and cancel",
		Verdict: verifylive.VerdictFail, Reason: "409 already-processing",
		AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now, Mutating: true,
		Artifacts: []verifylive.Artifact{{
			Kind: "order", ID: "order-left", Symbol: "005930", CreatedAt: now,
		}},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
}

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

// TestALeftoverKeepsResumeAvailable.
//
// "Every step has a terminal verdict" is what the no-op guard refuses, and it is
// the right rule — except when the record also holds an order this tool could not
// cancel. Then the run has exactly one live request left in it, and refusing the
// start is what made that order unremovable.
func TestALeftoverKeepsResumeAvailable(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	seedLeftoverOrder(t, h.record)
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	if !startForm(page) {
		t.Fatalf("no start control on a record that still holds a leftover:\n%s", truncateForLog(page))
	}
	if disabledStart(page) {
		t.Errorf("the resume is disabled although a leftover order is still on the account:\n%s",
			truncateForLog(page))
	}

	// And the guard behind the button agrees with the button: a stale tab posting
	// the same form must not be told there is nothing to do.
	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"resume"}})
	defer h.stopRun()
	if run := h.currentRun(); run == nil {
		page = body(t, h.get(t, "/verify"))
		if strings.Contains(page, "이어할 단계가 없다") {
			t.Fatalf("the server refused the resume as a no-op while a leftover was outstanding:\n%s",
				truncateForLog(page))
		}
		t.Fatal("the resume did not start a run")
	}
}
