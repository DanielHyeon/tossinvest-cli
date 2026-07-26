package console

// console_test.go covers the console's whole reason to exist: that the three
// gates are all required, that failing any of them sends nothing, and that
// passing all three runs exactly the batch the operator was shown.
//
// The broker is a counter (fake_broker_test.go). Every refusal assertion is
// "zero mutating broker calls", because that is the claim task 1.6 makes and the
// only one worth testing at this layer — the plan authorisation, the exposure
// caps and the cancellation rails are internal/verifylive's, tested there, and
// this package deliberately does not reimplement any of them.

import (
	"context"
	"html"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- harness --------------------------------------------------------------------

type harness struct {
	*Console
	broker *fakeBroker
	record string
	srv    *httptest.Server
	client *http.Client
}

// realStarter wires a real verifylive.Runner onto the fake broker.
//
// A stub confirmer would test the console against itself. The whole question is
// whether the console's web approval drives the *actual* runner, so the actual
// runner is what it drives.
func realStarter(broker *fakeBroker, recordPath string) StartVerify {
	return func(
		ctx context.Context,
		confirm verifylive.BatchConfirmer,
		out io.Writer,
	) (verifylive.Summary, []verifylive.Entry, error) {
		var empty verifylive.Summary
		prior, err := verifylive.LoadEntries(recordPath)
		if err != nil {
			return empty, nil, err
		}
		rec, err := verifylive.OpenRecorder(recordPath)
		if err != nil {
			return empty, nil, err
		}
		defer rec.Close()

		runner, err := verifylive.New(verifylive.Options{
			Broker:          broker,
			Recorder:        rec,
			Confirm:         func(verifylive.Mutation) error { return verifylive.ErrNotATerminal },
			ConfirmBatch:    confirm,
			Out:             out,
			AccountRef:      "123-45-678901",
			Symbol:          "005930",
			Offset:          verifylive.DefaultOffset,
			MaxSellQuantity: verifylive.DefaultMaxSellQuantity,
			Prior:           prior,
		})
		if err != nil {
			return empty, nil, err
		}
		summary, runErr := runner.Run(ctx)
		return summary, runner.Entries(), runErr
	}
}

func newHarness(t *testing.T, tweak ...func(*Options)) *harness {
	t.Helper()
	dir := t.TempDir()
	broker := newFakeBroker()
	record := filepath.Join(dir, verifylive.FileName)

	o := Options{
		StartVerify:  realStarter(broker, record),
		VerifyRecord: record,
		SoakRecord:   filepath.Join(dir, soak.FileName),
		Attestation:  filepath.Join(dir, attest.FileName),
		MinSoakDays:  3,
		Out:          io.Discard,
	}
	for _, f := range tweak {
		f(&o)
	}

	c, err := New(o)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	srv := httptest.NewServer(c.Handler())
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}

	h := &harness{Console: c, broker: broker, record: record, srv: srv, client: &http.Client{Jar: jar}}
	t.Cleanup(func() {
		h.stopRun()
		srv.Close()
	})
	return h
}

// stopRun releases a run left blocked on an approval nobody typed.
func (h *harness) stopRun() {
	run := h.currentRun()
	if run == nil {
		return
	}
	run.cancel()
	select {
	case <-run.done:
	case <-time.After(10 * time.Second):
	}
}

func (h *harness) authenticate(t *testing.T) {
	t.Helper()
	resp := h.get(t, "/?session="+h.SessionToken())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("the session URL returned %d, want 200", resp.StatusCode)
	}
}

func (h *harness) get(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := h.client.Get(h.srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func (h *harness) post(t *testing.T, path string, form url.Values) *http.Response {
	t.Helper()
	resp, err := h.client.PostForm(h.srv.URL+path, form)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func body(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading the body: %v", err)
	}
	return string(b)
}

// waitForBatch blocks until the runner has parked a plan for approval.
func (h *harness) waitForBatch(t *testing.T) runView {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if run := h.currentRun(); run != nil {
			if v := run.snapshot(); v.Awaiting {
				return v
			} else if v.Done {
				t.Fatalf("the run finished before asking for an approval: %+v", v.Summary)
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("no batch was offered for approval within the deadline")
	return runView{}
}

func (h *harness) waitForFinish(t *testing.T) runView {
	t.Helper()
	run := h.currentRun()
	if run == nil {
		t.Fatal("there is no run to wait for")
	}
	select {
	case <-run.done:
	case <-time.After(20 * time.Second):
		t.Fatal("the run did not finish within the deadline")
	}
	return run.snapshot()
}

// startAndWait begins a verification and returns the parked plan.
func (h *harness) startAndWait(t *testing.T) runView {
	t.Helper()
	h.authenticate(t)
	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}})
	return h.waitForBatch(t)
}

// --- the loopback rule -------------------------------------------------------------

// TestListenBindsTheLoopbackInterface. There is no host option and no flag; this
// pins the one address the console can ever be reached at.
func TestListenBindsTheLoopbackInterface(t *testing.T) {
	ln, err := Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	if err := loopbackOnly(ln); err != nil {
		t.Fatalf("the console's own listener is not loopback: %v", err)
	}
	if host, _, _ := strings.Cut(ln.Addr().String(), ":"); host != "127.0.0.1" {
		t.Errorf("Listen bound %s, want 127.0.0.1", ln.Addr())
	}
}

// TestServeRefusesANonLoopbackListener.
//
// The refusal is checked against the listener rather than a configured string,
// so it holds whoever opened the socket — including a future edit that added the
// flag this console does not have.
func TestServeRefusesANonLoopbackListener(t *testing.T) {
	h := newHarness(t)

	ln, err := newAnyInterfaceListener()
	if err != nil {
		t.Skipf("no non-loopback listener available in this environment: %v", err)
	}
	defer ln.Close()

	err = h.Serve(context.Background(), ln)
	if err == nil {
		t.Fatal("Serve accepted a non-loopback listener")
	}
	if !strings.Contains(err.Error(), "non-loopback") {
		t.Errorf("err = %v, want the non-loopback refusal", err)
	}
	// And the socket is shut, not left listening for anybody who can route to it.
	if _, err := ln.Accept(); err == nil {
		t.Error("the refused listener is still accepting connections")
	}
}

// --- the session gate ----------------------------------------------------------------

// TestEveryRouteRefusesARequestWithoutTheSessionToken.
//
// Possession of the terminal is the authentication, so a request that cannot show
// the token gets nothing — not the dashboard, not the report, and above all not
// the approval form's target.
func TestEveryRouteRefusesARequestWithoutTheSessionToken(t *testing.T) {
	h := newHarness(t)

	for _, route := range []struct {
		method, path string
	}{
		{http.MethodGet, "/"},
		{http.MethodGet, "/verify"},
		{http.MethodGet, "/report"},
		{http.MethodGet, "/report.json"},
		{http.MethodPost, "/verify/start"},
		{http.MethodPost, "/verify/approve"},
		{http.MethodPost, "/verify/abort"},
	} {
		req := httptest.NewRequest(route.method, route.path,
			strings.NewReader("csrf="+h.csrf))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		h.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s = %d, want 403 without a session", route.method, route.path, rec.Code)
		}
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) were made without a session", n)
	}
	if h.currentRun() != nil {
		t.Error("a verification was started without a session")
	}
}

// TestTheURLTokenIsExchangedForACookieAndLeavesTheAddressBar.
func TestTheURLTokenIsExchangedForACookieAndLeavesTheAddressBar(t *testing.T) {
	h := newHarness(t)

	resp := h.get(t, "/?session="+h.SessionToken())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if strings.Contains(resp.Request.URL.RawQuery, "session=") {
		t.Errorf("the token is still in the address bar: %s", resp.Request.URL)
	}

	u, _ := url.Parse(h.srv.URL)
	var found bool
	for _, ck := range h.client.Jar.Cookies(u) {
		if ck.Name == sessionCookie && ck.Value == h.SessionToken() {
			found = true
		}
	}
	if !found {
		t.Error("no session cookie was set from the URL token")
	}

	// A wrong token is refused even though it is spelled like the real thing.
	fresh := newHarness(t)
	resp = fresh.get(t, "/?session="+strings.Repeat("A", len(fresh.SessionToken())))
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("a wrong token returned %d, want 403", resp.StatusCode)
	}
}

// TestTheConsoleServesNothingButItsThreePages. A route nobody wrote is a route
// nobody reviewed.
func TestTheConsoleServesNothingButItsThreePages(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	for _, path := range []string{"/config", "/gate", "/order", "/credentials", "/verify/run"} {
		resp := h.get(t, path)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, resp.StatusCode)
		}
	}
}

// TestTheApprovalRoutesRefuseAGET. A confirmation that could be triggered by a
// link is not a confirmation.
func TestTheApprovalRoutesRefuseAGET(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	for _, path := range []string{"/verify/start", "/verify/approve", "/verify/abort"} {
		resp := h.get(t, path)
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want 405", path, resp.StatusCode)
		}
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) were made by a GET", n)
	}
}

// --- the approval triple ----------------------------------------------------------

// TestAWrongCSRFTokenSendsNothing. The nonce is right; the form is not one this
// console drew.
func TestAWrongCSRFTokenSendsNothing(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	resp := h.post(t, "/verify/approve", url.Values{
		"csrf":  {"not-the-csrf-token"},
		"nonce": {view.Batch.Nonce},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after a CSRF failure", n)
	}
	if !h.currentRun().snapshot().Awaiting {
		t.Error("the CSRF failure disturbed the pending approval; it should have been ignored entirely")
	}
}

// TestAMissingSessionOnTheApprovalSendsNothing — the same POST, from a client
// that never had the token.
func TestAMissingSessionOnTheApprovalSendsNothing(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	anonymous := &http.Client{}
	resp, err := anonymous.PostForm(h.srv.URL+"/verify/approve", url.Values{
		"csrf":  {h.csrf},
		"nonce": {view.Batch.Nonce},
	})
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from a session-less approval", n)
	}
}

// TestAWrongNonceSendsNothingAndStopsTheRun.
//
// This is the TTY contract: "anything else aborts the run". Nothing was sent, so
// the cost of a typo is starting the form again — and a mismatch does not become
// five minutes of guesses.
func TestAWrongNonceSendsNothingAndStopsTheRun(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	resp := h.post(t, "/verify/approve", url.Values{
		"csrf":  {h.csrf},
		"nonce": {view.Batch.Nonce + "X"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want the verify page back", resp.StatusCode)
	}

	final := h.waitForFinish(t)
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after a wrong nonce", n)
	}
	if !final.Summary.Halted {
		t.Error("the run continued after the approval was refused")
	}
	if len(h.broker.placements()) != 0 {
		t.Errorf("orders were placed after a wrong nonce: %+v", h.broker.placements())
	}
	// The record keeps the refusal but records no step, so the next attempt is
	// not mistaken for a restart over finished work.
	entries := loadRecord(t, h.record)
	if n := verifylive.StepCount(entries); n != 0 {
		t.Errorf("%d step(s) were recorded for a refused approval", n)
	}
	if len(entries) == 0 {
		t.Error("the refusal was not recorded at all; a decision belongs in the audit trail")
	}
}

// TestAnExpiredNonceSendsNothing. The window is the same one the terminal has,
// checked by the same verifylive.Batch.Verify.
func TestAnExpiredNonceSendsNothing(t *testing.T) {
	// The console's clock runs two hours ahead of the runner's, so the batch is
	// minted with a live expiry and read after it has passed.
	h := newHarness(t, func(o *Options) {
		o.Now = func() time.Time { return time.Now().UTC().Add(2 * time.Hour) }
	})
	view := h.startAndWait(t)

	h.post(t, "/verify/approve", url.Values{
		"csrf":  {h.csrf},
		"nonce": {view.Batch.Nonce},
	})

	final := h.waitForFinish(t)
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after an expired approval", n)
	}
	if !final.Summary.Halted {
		t.Error("the run continued past an expired approval")
	}
}

// TestTheApprovedFlowRunsExactlyTheApprovedBatch.
//
// The positive case, and the one that has to be exact: the record's approval line
// carries the digest of the plan, and that digest has to be the digest of the plan
// the page rendered. Anything else would mean the operator approved one list and
// the runner walked another.
func TestTheApprovedFlowRunsExactlyTheApprovedBatch(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)

	if len(view.Batch.Plan.Mutations) == 0 {
		t.Fatal("the plan offered for approval is empty; there would be nothing to prove")
	}
	page := body(t, h.get(t, "/verify"))
	if !strings.Contains(page, view.Batch.Nonce) {
		t.Error("the page does not display the nonce it is asking for")
	}
	// The page must carry the terminal's summary verbatim — the batch model's
	// safety claim is that the list is complete, and a second rendering would be a
	// second list to keep in step. Compared with whitespace collapsed, because the
	// terminal wraps at eighty-odd columns and the browser does not.
	shown := flatten(page)
	if !strings.Contains(shown, flatten(view.Batch.Prompt())) {
		t.Errorf("the page does not show the same batch summary the CLI prints:\n%s", truncateForLog(page))
	}
	for _, m := range view.Batch.Plan.Mutations {
		if !strings.Contains(shown, flatten(m.Headline())) {
			t.Errorf("the batch summary on the page omits %q", m.Headline())
		}
		if !strings.Contains(shown, flatten(m.Ends)) {
			t.Errorf("the batch summary on the page does not say how %s ends", m.Kind)
		}
	}

	h.post(t, "/verify/approve", url.Values{
		"csrf":  {h.csrf},
		"nonce": {view.Batch.Nonce},
	})
	final := h.waitForFinish(t)

	if n := h.broker.mutationCount(); n == 0 {
		t.Fatal("the approved run sent nothing")
	}
	for _, p := range h.broker.placements() {
		if p.Symbol != "005930" || !strings.EqualFold(p.Side, "buy") || p.Quantity != 1 {
			t.Errorf("an order outside the approved shape was placed: %+v", p)
		}
	}
	if final.Err != "" && strings.Contains(final.Err, verifylive.ErrOutsidePlan.Error()) {
		t.Errorf("the run sent something outside the plan: %s", final.Err)
	}

	approval := approvalEntry(t, loadRecord(t, h.record))
	if approval.Verdict != verifylive.VerdictPass {
		t.Errorf("the recorded approval verdict is %q, want pass", approval.Verdict)
	}
	if got, want := observation(approval, "approval.plan_digest"), view.Batch.Plan.Digest(); got != want {
		t.Errorf("the run recorded plan digest %q; the page showed %q", got, want)
	}
}

// TestAbortRefusesWithoutTypingAnything.
func TestAbortRefusesWithoutTypingAnything(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)

	h.post(t, "/verify/abort", url.Values{"csrf": {h.csrf}})
	final := h.waitForFinish(t)

	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) after an abort", n)
	}
	if !final.Summary.Halted {
		t.Error("the run continued after being aborted")
	}
}

// --- the process boundary ---------------------------------------------------------

// TestASecondVerificationInTheSameProcessIsRefused.
//
// conditional-persist measures whether the broker still holds the protection after
// the registering process has exited. A console that let you press the button
// again would be a console that could not measure it.
func TestASecondVerificationInTheSameProcessIsRefused(t *testing.T) {
	h := newHarness(t)
	view := h.startAndWait(t)
	h.post(t, "/verify/approve", url.Values{"csrf": {h.csrf}, "nonce": {view.Batch.Nonce}})
	h.waitForFinish(t)

	page := body(t, h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}}))
	if !strings.Contains(page, "콘솔을 종료") {
		t.Errorf("a second run in the same process was not refused with the restart instruction:\n%s",
			truncateForLog(page))
	}
	before := h.broker.mutationCount()
	if run := h.currentRun(); !run.finished() {
		t.Error("a second run was started in the same process")
	}
	if h.broker.mutationCount() != before {
		t.Error("the refused second run still sent something")
	}
}

// TestARestartedConsoleLabelsItselfAResume. The record is how the new process
// knows what the old one did, and the page has to say so before anybody clicks.
func TestARestartedConsoleLabelsItselfAResume(t *testing.T) {
	h := newHarness(t)
	seedAwaitingRestart(t, h.record)

	// A second console over the same record: this is what starting the binary
	// again looks like.
	restarted := newHarness(t, func(o *Options) {
		o.VerifyRecord = h.record
		o.StartVerify = realStarter(h.broker, h.record)
	})
	restarted.authenticate(t)

	page := body(t, restarted.get(t, "/verify"))
	for _, want := range []string{"이어하기", string(verifylive.StepConditionalPersist)} {
		if !strings.Contains(page, want) {
			t.Errorf("the resumed console's page does not mention %q:\n%s", want, truncateForLog(page))
		}
	}

	dash := body(t, restarted.get(t, "/"))
	if !strings.Contains(dash, "새 프로세스") {
		t.Errorf("the dashboard does not report the awaiting-restart state:\n%s", truncateForLog(dash))
	}
	if !strings.Contains(dash, "co-1") {
		t.Errorf("the dashboard does not name the conditional order left live:\n%s", truncateForLog(dash))
	}
}

// --- the read-only pages --------------------------------------------------------

func TestTheDashboardReportsAnUnstartedMachineWithoutFailing(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	page := body(t, h.get(t, "/"))
	for _, want := range []string{"soak", "attestation", "tossctl soak run"} {
		if !strings.Contains(page, want) {
			t.Errorf("the dashboard does not mention %q on a machine with no records", want)
		}
	}
	if !strings.Contains(page, "게이트를 켜지 않는다") {
		t.Error("the dashboard does not say the console is not the place gates are turned on")
	}
}

func TestTheDashboardMasksTheAccountNumber(t *testing.T) {
	h := newHarness(t)
	seedAwaitingRestart(t, h.record)
	h.authenticate(t)

	for _, path := range []string{"/", "/verify", "/report"} {
		if page := body(t, h.get(t, path)); strings.Contains(page, "123-45-678901") {
			t.Errorf("%s printed the unmasked account number", path)
		}
	}
}

func TestTheReportPageAndItsJSONDownload(t *testing.T) {
	h := newHarness(t)
	seedAwaitingRestart(t, h.record)
	h.authenticate(t)

	page := body(t, h.get(t, "/report"))
	if !strings.Contains(page, "DISABLED") {
		t.Errorf("the report page does not carry the replay verdict:\n%s", truncateForLog(page))
	}

	resp := h.get(t, "/report.json")
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Errorf("Content-Type = %q, want JSON", got)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, "attachment") {
		t.Errorf("Content-Disposition = %q, want an attachment", got)
	}
	if raw := body(t, resp); !strings.Contains(raw, `"replay_enabled": false`) {
		t.Errorf("the JSON download does not carry replay_enabled=false:\n%s", truncateForLog(raw))
	}
}

// TestTheVerifyPageShowsTheStepListBeforeAnythingStarts — the same content as
// `tossctl verify run --list`, so an operator reads the procedure before pressing
// anything.
func TestTheVerifyPageShowsTheStepListBeforeAnythingStarts(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	page := body(t, h.get(t, "/verify"))
	for _, s := range verifylive.Steps() {
		if !strings.Contains(page, string(s.ID)) {
			t.Errorf("the step list omits %s", s.ID)
		}
	}
	if !strings.Contains(page, "--confirm-each") {
		t.Error("the page does not say the per-mutation gate is not offered here")
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from merely opening the page", n)
	}
}

// --- lifecycle ---------------------------------------------------------------------

// TestShutdownWaitsForARunInProgress.
//
// Ctrl-C must not abandon a step in the middle of cancelling an order. The
// cancellation is handed to the runner and the shutdown waits for it to settle.
func TestShutdownWaitsForARunInProgress(t *testing.T) {
	var out strings.Builder
	h := newHarness(t, func(o *Options) { o.Out = &out })
	h.startAndWait(t)

	ln, err := Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- h.Serve(ctx, ln) }()

	// Wait until it is actually serving before asking it to stop.
	deadline := time.Now().Add(5 * time.Second)
	for h.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()

	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("Serve: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("Serve did not return")
	}

	if !h.currentRun().finished() {
		t.Error("the console shut down while a run was still walking steps")
	}
	if !strings.Contains(out.String(), "Waiting for its steps to settle") {
		t.Errorf("the shutdown did not say it was waiting for the run:\n%s", out.String())
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) from a run that was never approved", n)
	}
}

func TestNewRefusesAConsoleWithNoWayToRunAVerification(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("New accepted a console with no verification starter")
	}
}

func TestTheURLCarriesTheSessionTokenOnce(t *testing.T) {
	h := newHarness(t)
	if h.URL() != "" {
		t.Errorf("URL() = %q before Serve, want empty", h.URL())
	}

	ln, err := Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- h.Serve(ctx, ln) }()

	deadline := time.Now().Add(5 * time.Second)
	for h.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	url := h.URL()
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Errorf("URL = %q, want a loopback address", url)
	}
	if !strings.Contains(url, "?session="+h.SessionToken()) {
		t.Errorf("URL = %q, want it to carry the one-time token", url)
	}
	cancel()
	<-done
}

// --- fixtures ------------------------------------------------------------------------

func loadRecord(t *testing.T, path string) []verifylive.Entry {
	t.Helper()
	entries, err := verifylive.LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	return entries
}

func approvalEntry(t *testing.T, entries []verifylive.Entry) verifylive.Entry {
	t.Helper()
	for _, e := range entries {
		if e.Kind == verifylive.KindApproval {
			return e
		}
	}
	t.Fatal("the record carries no approval line")
	return verifylive.Entry{}
}

func observation(e verifylive.Entry, key string) string {
	for _, o := range e.Observations {
		if o.Key == key {
			return o.Value
		}
	}
	return ""
}

// seedAwaitingRestart writes the record a console restart is supposed to pick up:
// a verification that stopped at the persistence check with a conditional order
// still registered.
func seedAwaitingRestart(t *testing.T, path string) {
	t.Helper()
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	now := time.Now().UTC()
	for _, e := range []verifylive.Entry{
		{
			StepID: verifylive.StepReadFixtures, Title: "fixtures", Verdict: verifylive.VerdictPass,
			AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now,
			Observations: []verifylive.Observation{{Key: "order.status.observed", Value: "FILLED"}},
		},
		{
			StepID: verifylive.StepConditionalRegister, Title: "register", Verdict: verifylive.VerdictPass,
			AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now, Mutating: true,
			Artifacts: []verifylive.Artifact{{
				Kind: "conditional-order", ID: "co-1", Symbol: "005930", CreatedAt: now, Deliberate: true,
			}},
		},
		{
			StepID: verifylive.StepConditionalPersist, Title: "persist",
			Verdict: verifylive.VerdictAwaitingRestart, Reason: "start a new process",
			AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now,
		},
	} {
		if err := rec.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

// flatten un-escapes HTML and collapses runs of whitespace, so a comparison
// against terminal output is about the words rather than about the wrapping.
func flatten(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}

func truncateForLog(s string) string {
	if len(s) <= 1500 {
		return s
	}
	return s[:1500] + "…"
}

// newAnyInterfaceListener opens a socket that is NOT loopback.
//
// 0.0.0.0 rather than a routable address: it is reachable from off the machine
// wherever one exists, it needs no interface enumeration, and it is exactly the
// mistake a future "just let me reach it from my phone" edit would make.
func newAnyInterfaceListener() (net.Listener, error) {
	return net.Listen("tcp", "0.0.0.0:0")
}
