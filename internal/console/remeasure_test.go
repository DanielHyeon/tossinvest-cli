package console

// remeasure_test.go covers task 1.7's two console-facing halves: the re-measure
// start mode and the market-hours advisory.
//
// The claims worth testing here are not "the button works". They are:
//
//	a passed step is never re-measured  — the redo set comes from
//	                                      verifylive.RedoSet against the record,
//	                                      and the record is on disk
//	the form cannot name a step         — a step id in the POST body is ignored;
//	                                      only the file decides
//	nothing is sent without a nonce     — a re-measurement is an ordinary run with
//	                                      an ordinary batch approval in front of it
//	the advisory never blocks           — it is a paragraph, not a gate

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// seedVerdicts writes a terminal verdict for every step, overriding the named
// ones. It is what the record looked like after the first live run: a full sweep
// of verdicts, almost none of them measurements.
func seedVerdicts(t *testing.T, path string, base verifylive.Verdict, override map[verifylive.StepID]verifylive.Verdict) {
	t.Helper()
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	now := time.Now().UTC()
	for _, step := range verifylive.Steps() {
		verdict := base
		if v, ok := override[step.ID]; ok {
			verdict = v
		}
		if err := rec.Append(verifylive.Entry{
			StepID: step.ID, Title: step.Title, Verdict: verdict, Mutating: step.Mutates,
			Reason:     "seeded by the test",
			AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now,
		}); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

// --- the start screen -------------------------------------------------------------

// TestTheStartScreenOffersARemeasureAndNamesWhatItWouldTouch.
//
// "대상 단계 수를 시작 화면에 표기" (tasks.md 1.7). The count and the ids, before
// anybody presses anything.
func TestTheStartScreenOffersARemeasureAndNamesWhatItWouldTouch(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel:        verifylive.VerdictFail,
		verifylive.StepOrderAmend:         verifylive.VerdictFail,
		verifylive.StepSellBoundary:       verifylive.VerdictSkipped,
		verifylive.StepConditionalTrigger: verifylive.VerdictDeferred,
	})
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	if !strings.Contains(page, "재측정") {
		t.Fatalf("the start screen offers no re-measurement:\n%s", truncateForLog(page))
	}
	if !strings.Contains(page, "재측정 3단계") {
		t.Errorf("the start screen does not say how many steps a re-measurement would touch:\n%s",
			truncateForLog(page))
	}
	for _, want := range []verifylive.StepID{
		verifylive.StepOrderCancel, verifylive.StepOrderAmend, verifylive.StepSellBoundary,
	} {
		if !strings.Contains(page, string(want)) {
			t.Errorf("the re-measurement list omits %s", want)
		}
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from merely rendering the start screen", n)
	}
}

// TestTheStartScreenOffersNoRemeasureWhenThereIsNothingToRedo.
func TestTheStartScreenOffersNoRemeasureWhenThereIsNothingToRedo(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	if strings.Contains(page, `name="mode" value="redo"`) {
		t.Errorf("the start screen offers a re-measurement with nothing to re-measure:\n%s",
			truncateForLog(page))
	}
}

// TestARemeasureAskedForWithNothingToRedoSendsNothing — the guard behind the
// button, for a form submitted after the record changed.
func TestARemeasureAskedForWithNothingToRedoSendsNothing(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	h.authenticate(t)

	page := body(t, h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"redo"}}))
	if !strings.Contains(page, "재측정할 단계가 없다") {
		t.Errorf("the refusal is not on the page:\n%s", truncateForLog(page))
	}
	if h.currentRun() != nil {
		t.Error("a run was started with an empty redo set")
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from an empty re-measurement", n)
	}
}

// --- the run ------------------------------------------------------------------------

// TestARemeasureRunsOnlyTheFailedStepAndStillAsksForANewNonce.
//
// The whole point of task 1.7: a terminal fail becomes attemptable again, through
// the same door as everything else. The plan is rebuilt, the nonce is new, and the
// steps that passed are left exactly where they are.
func TestARemeasureRunsOnlyTheFailedStepAndStillAsksForANewNonce(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"redo"}})
	view := h.waitForBatch(t)

	if len(view.Redo) != 1 || view.Redo[0] != verifylive.StepOrderCancel {
		t.Fatalf("the run's redo set is %v, want only %s", view.Redo, verifylive.StepOrderCancel)
	}
	if view.Batch == nil || strings.TrimSpace(view.Batch.Nonce) == "" {
		t.Fatal("the re-measurement did not ask for a batch approval")
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Fatalf("%d mutating broker call(s) before the nonce was typed", n)
	}
	// Every planned line belongs to the step being re-measured.
	for _, m := range view.Batch.Plan.Mutations {
		if m.Step != verifylive.StepOrderCancel {
			t.Errorf("the plan carries a line for %s, which is not in the redo set", m.Step)
		}
	}

	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}, "nonce": {view.Batch.Nonce}})
	final := h.waitForFinish(t)

	if got := len(h.broker.placements()); got != 1 {
		t.Errorf("%d order(s) were placed, want 1 — only order-cancel should have run: %+v",
			got, h.broker.placements())
	}
	settled, ran := map[verifylive.StepID]bool{}, map[verifylive.StepID]bool{}
	for _, o := range final.Summary.Outcomes {
		if o.AlreadySettled {
			settled[o.Step] = true
			continue
		}
		ran[o.Step] = true
	}
	if !ran[verifylive.StepOrderCancel] {
		t.Error("the step the operator asked to re-measure did not run")
	}
	for _, step := range verifylive.Steps() {
		if step.ID == verifylive.StepOrderCancel {
			continue
		}
		if !settled[step.ID] {
			t.Errorf("%s was re-run; a re-measurement must touch only the fail/skipped steps", step.ID)
		}
	}
}

// TestARemeasureSendsNothingWhenTheNonceIsWrong. The redo path has no shortcut
// around the approval — it is the same confirmer, the same Batch.Verify.
func TestARemeasureSendsNothingWhenTheNonceIsWrong(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"redo"}})
	view := h.waitForBatch(t)

	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}, "nonce": {view.Batch.Nonce + "X"}})
	final := h.waitForFinish(t)

	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after a wrong nonce on a re-measurement", n)
	}
	if !final.Summary.Halted {
		t.Error("the re-measurement continued after the approval was refused")
	}
	// And the failed step is still failed, so the button is still there.
	if got := verifylive.RedoSet(loadRecord(t, h.record)); len(got) != 1 {
		t.Errorf("the redo set after a refusal is %v, want the one failed step still on it", got)
	}
}

// TestARemeasureCannotBeStartedWithoutTheCSRFToken.
func TestARemeasureCannotBeStartedWithoutTheCSRFToken(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {"wrong"}, "mode": {"redo"}})
	if h.currentRun() != nil {
		t.Error("a re-measurement was started without the CSRF token")
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from a CSRF-less re-measurement", n)
	}
}

// TestTheRedoSetComesFromTheRecordAndNotFromTheForm.
//
// This is the injection case, and it is the reason handleStart reads the file
// instead of trusting the request: a step named in the POST body must not become a
// step that gets a second live order.
func TestTheRedoSetComesFromTheRecordAndNotFromTheForm(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{
		"csrf": {h.csrf},
		"mode": {"redo"},
		// Everything a caller might hope the handler reads.
		"redo":  {string(verifylive.StepIdempotency), string(verifylive.StepOrderAmend)},
		"steps": {string(verifylive.StepIdempotency)},
		"step":  {string(verifylive.StepOrderAmend)},
	})
	view := h.waitForBatch(t)

	if len(view.Redo) != 1 || view.Redo[0] != verifylive.StepOrderCancel {
		t.Fatalf("the run's redo set is %v; the form must not be able to name a step", view.Redo)
	}
	for _, m := range view.Batch.Plan.Mutations {
		if m.Step != verifylive.StepOrderCancel {
			t.Errorf("the form put %s on the plan", m.Step)
		}
	}
}

// TestAnOrdinaryStartIsStillNotARemeasure. The default button keeps the
// always-resume behaviour: a settled step stays settled.
func TestAnOrdinaryStartIsStillNotARemeasure(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "mode": {"resume"}})
	final := h.waitForFinish(t)

	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from an ordinary resume over a fully settled record", n)
	}
	for _, o := range final.Summary.Outcomes {
		if !o.AlreadySettled {
			t.Errorf("%s ran on an ordinary resume; every step already has a terminal verdict", o.Step)
		}
	}
}

// --- the market-hours advisory ---------------------------------------------------

func atKST(t *testing.T, y int, m time.Month, d, hh, mm int) time.Time {
	t.Helper()
	return time.Date(y, m, d, hh, mm, 0, 0, time.FixedZone("KST", 9*60*60))
}

// TestTheStartScreenWarnsOutsideMarketHours — and names the measured code, so an
// operator can check the warning against the evidence rather than believe it.
func TestTheStartScreenWarnsOutsideMarketHours(t *testing.T) {
	// The Sunday evening the first live run happened on.
	sunday := atKST(t, 2026, 7, 26, 21, 31)
	h := newHarness(t, func(o *Options) { o.Now = func() time.Time { return sunday } })
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	for _, want := range []string{"KR 정규장", "order-hours-closed", "422"} {
		if !strings.Contains(page, want) {
			t.Errorf("the start screen does not mention %q outside market hours:\n%s",
				want, truncateForLog(page))
		}
	}
	if !strings.Contains(page, "막지는 않는다") {
		t.Errorf("the advisory does not say it is advisory:\n%s", truncateForLog(page))
	}
}

// TestTheAdvisoryIsQuietInsideMarketHours.
func TestTheAdvisoryIsQuietInsideMarketHours(t *testing.T) {
	monday := atKST(t, 2026, 7, 27, 11, 0)
	h := newHarness(t, func(o *Options) { o.Now = func() time.Time { return monday } })
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	if strings.Contains(page, "order-hours-closed") {
		t.Errorf("the closed-market warning is shown during the session:\n%s", truncateForLog(page))
	}
	if !strings.Contains(page, "KR 정규장 시간이다") {
		t.Errorf("the page does not report the session reading at all:\n%s", truncateForLog(page))
	}
}

// TestTheAdvisoryNeverBlocksTheStart.
//
// The task says advisory and means it: the window in which the broker *does*
// accept an order is unmeasured, so a refusal here would be the tool asserting
// something nobody has observed.
func TestTheAdvisoryNeverBlocksTheStart(t *testing.T) {
	sunday := atKST(t, 2026, 7, 26, 21, 31)
	h := newHarness(t, func(o *Options) { o.Now = func() time.Time { return sunday } })
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}})
	view := h.waitForBatch(t)
	if view.Batch == nil {
		t.Fatal("a run started outside market hours did not reach the approval")
	}
	if len(view.Batch.Plan.Mutations) == 0 {
		t.Error("the plan was emptied because the market is closed; the advisory must not change the plan")
	}
}

// TestTheRemeasureScreenCarriesTheAdvisoryToo. Both start paths, per the task.
func TestTheRemeasureScreenCarriesTheAdvisoryToo(t *testing.T) {
	sunday := atKST(t, 2026, 7, 26, 21, 31)
	h := newHarness(t, func(o *Options) { o.Now = func() time.Time { return sunday } })
	seedVerdicts(t, h.record, verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	if !strings.Contains(page, "order-hours-closed") {
		t.Errorf("the re-measure screen does not warn about the closed market:\n%s", truncateForLog(page))
	}
	if !strings.Contains(page, "재측정 1단계") {
		t.Errorf("the re-measure offer is missing from the same screen:\n%s", truncateForLog(page))
	}
}
