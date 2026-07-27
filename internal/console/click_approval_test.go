package console

// click_approval_test.go covers console-click-approval: the console's batch
// approval is a click, and the start screen never offers an action that would do
// nothing.
//
// What the tests hold onto is the part that did not change. A click is still the
// third gate — session, CSRF, and a POST from the screen that is showing the plan
// — and every refusal still ends with zero mutating broker calls. What is gone is
// the typed string, and these tests fail if it comes back.

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- the approval is a click -------------------------------------------------------

// TestApprovalIsOneClickWithNothingTyped.
//
// The form carries the CSRF token and nothing else. That has to be enough, and it
// has to drive the real runner: the assertion is that orders were actually placed
// against the fake broker, not that a handler returned 200.
func TestApprovalIsOneClickWithNothingTyped(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	if len(view.Batch.Plan.Mutations) == 0 {
		t.Fatal("the plan offered for approval is empty; there would be nothing to prove")
	}

	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}})
	final := h.waitForFinish(t)

	if len(h.broker.placements()) == 0 {
		t.Fatal("the clicked approval sent nothing; the click is not an approval")
	}
	if final.Summary.Halted {
		t.Error("the run halted after being approved")
	}
	// The approval is on the record as granted, with the plan it covered.
	var approvals int
	for _, e := range loadRecord(t, h.record) {
		if e.Kind != verifylive.KindApproval {
			continue
		}
		approvals++
		if e.Verdict != verifylive.VerdictPass {
			t.Errorf("the approval entry says %q, want pass", e.Verdict)
		}
	}
	if approvals != 1 {
		t.Errorf("%d approval entries were recorded, want exactly 1", approvals)
	}
}

// TestTheApprovalScreenAsksForNoTypedString.
//
// The screen must not describe an approval method that is not the one in force.
// It still has to show the complete plan — that is what the batch model rests on —
// so this checks both directions at once.
func TestTheApprovalScreenAsksForNoTypedString(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	page := body(t, h.get(t, "/verify"))
	if strings.Contains(page, view.Batch.Nonce) {
		t.Errorf("the approval screen still displays a confirmation string:\n%s", truncateForLog(page))
	}
	for _, banned := range []string{"확인 문자열", "입력하라", `name="nonce"`} {
		if strings.Contains(page, banned) {
			t.Errorf("the approval screen still asks for typing (%q):\n%s", banned, truncateForLog(page))
		}
	}
	// The list itself is still there, rendered from the same source the terminal
	// prints, so the operator approves a complete list rather than a summary of it.
	if !strings.Contains(flatten(page), flatten(view.Batch.Summary())) {
		t.Errorf("the approval screen does not show the plan summary:\n%s", truncateForLog(page))
	}
}

// TestAnExpiredApprovalWindowStillSendsNothing.
//
// Removing the typing does not remove the window: an approval clicked after the
// plan's prices went stale is refused, and nothing is sent.
func TestAnExpiredApprovalWindowStillSendsNothing(t *testing.T) {
	h := newHarness(t, func(o *Options) {
		o.Now = func() time.Time { return time.Now().UTC().Add(2 * time.Hour) }
	})
	h.startAndWait(t)

	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}})
	final := h.waitForFinish(t)

	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after an expired approval window", n)
	}
	if !final.Summary.Halted {
		t.Error("the run continued past an expired approval window")
	}
}

// TestTheApprovalRecordNamesTheClickChannel.
//
// The record is what proves later that a person was shown a list and answered. It
// said "one typed expiring string" for every run; for a console run that sentence
// is now false, and a false sentence in the evidence is worse than no sentence.
func TestTheApprovalRecordNamesTheClickChannel(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)
	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}})
	h.waitForFinish(t)

	var detail string
	for _, e := range loadRecord(t, h.record) {
		if e.Kind != verifylive.KindApproval {
			continue
		}
		for _, o := range e.Observations {
			if o.Key == "approval.model" {
				detail = o.Detail
			}
		}
	}
	if detail == "" {
		t.Fatal("the approval entry carries no approval.model observation")
	}
	if strings.Contains(detail, "typed") {
		t.Errorf("the record calls a clicked approval typed: %q", detail)
	}
	if !strings.Contains(detail, "click") {
		t.Errorf("the record does not name the click channel: %q", detail)
	}
}

// --- the start screen offers no no-op ------------------------------------------------

// TestResumeIsNotOfferedWhenNothingIsPending.
//
// 2026-07-27, live: [이어하기] was pressed twice against a record whose every step
// already had a terminal verdict. Both runs finished with "0 step(s) recorded" and
// the market window was spent. The button must not be offered for that.
func TestResumeIsNotOfferedWhenNothingIsPending(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	resume, ok := formSection(page, `value="resume"`)
	if !ok {
		t.Fatalf("the resume form is gone entirely; it should be present and disabled:\n%s",
			truncateForLog(page))
	}
	if !strings.Contains(resume, "disabled") {
		t.Errorf("the resume button is live with nothing to resume:\n%s", resume)
	}
	if !strings.Contains(page, "이어할 단계가 없다") {
		t.Errorf("the screen does not say why resuming is disabled:\n%s", truncateForLog(page))
	}
	// And the action that would actually measure something is still offered.
	if !strings.Contains(page, `name="mode" value="redo"`) {
		t.Errorf("the re-measurement is not offered as the alternative:\n%s", truncateForLog(page))
	}
}

// TestResumeStaysOfferedWhileAStepIsPending — the restart handoff still works.
//
// conditional-persist halts with a non-terminal verdict on purpose: the next
// process resumes it. That is the one case where resuming is the right button, so
// the guard must not swallow it.
func TestResumeStaysOfferedWhileAStepIsPending(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepConditionalPersist: verifylive.VerdictAwaitingRestart,
	})
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	resume, ok := formSection(page, `value="resume"`)
	if !ok {
		t.Fatalf("the resume form is missing while a step is pending:\n%s", truncateForLog(page))
	}
	if strings.Contains(resume, "disabled") {
		t.Errorf("resuming is disabled while a step is still waiting for a process:\n%s", resume)
	}
}

// TestResumeWithNothingPendingStartsNoRun — the guard behind the button.
//
// A form submitted from a stale tab must not produce the empty run either. The
// record is the judge, read at the moment of the POST.
func TestResumeWithNothingPendingStartsNoRun(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	h.authenticate(t)

	page := body(t, h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"resume"}}))
	if !strings.Contains(page, "이어할 단계가 없다") {
		t.Errorf("the refusal is not on the page:\n%s", truncateForLog(page))
	}
	if h.currentRun() != nil {
		t.Error("a run was started with nothing to resume")
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from a no-op resume", n)
	}
}

// formSection returns the <form> element containing marker.
func formSection(page, marker string) (string, bool) {
	for _, part := range strings.Split(page, "<form") {
		if strings.Contains(part, marker) {
			section, _, _ := strings.Cut(part, "</form>")
			return section, true
		}
	}
	return "", false
}
