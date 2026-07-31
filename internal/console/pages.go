package console

// pages.go is the console's whole user interface: seven handlers and one
// stylesheet, server-rendered with html/template.
//
// No JavaScript, no CDN, no build step, no external asset of any kind. The page
// an operator approves live orders on has to be readable from its source, and a
// script tag is one more thing that would have to be trusted for that reading to
// mean anything. Progress is a two-second meta refresh — except while the nonce
// form is on screen, where a refresh would wipe what somebody is typing.
//
// The batch summary is rendered by calling verifylive.Batch.Prompt: the exact
// bytes the terminal prints, in a <pre>. Re-rendering the plan in HTML would be a
// second summary to keep in step with the first, and the entire batch model rests
// on the summary being complete.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// pageFuncs are the only things the templates can compute that Go does not hand
// them ready-made. inRedo asks the set itself rather than re-deriving it: the
// table that marks a row "재측정 대상" has to agree with the set the button
// actually starts.
//
// It used to call verifylive.RedoableVerdict on the row's verdict, which was the
// same answer while the set was a function of the verdict alone. It is not any
// more — a passed conditional-register whose order is gone is in the set (see
// verifylive/redo.go) — so a table asking the verdict would have hidden the one
// row that unblocks the chain while the button re-ran it.
var pageFuncs = template.FuncMap{
	"add1": func(value int) int { return value + 1 },
	"inRedo": func(set []verifylive.StepID, id verifylive.StepID) bool {
		for _, s := range set {
			if s == id {
				return true
			}
		}
		return false
	},
	// stepLabel and verdict are the render-layer mapping task 1.8 ③ asks for: the
	// record keeps its English `title` and its verdict values, and the screen puts
	// Korean next to them. verifylive owns both maps and a drift test there fails
	// if a catalogue step ever loses its label.
	"stepLabel": verifylive.StepLabel,
	"verdict":   verifylive.VerdictLabel,
}

// pages is the whole template set. The dashboard screens live in their own const
// (templates_portfolio.go) and are parsed into it here, so they share "head",
// "foot" and the one stylesheet rather than growing a second look. The overview
// joins the same chain and reuses "journalstate" from the portfolio set, so the
// four journal failures are worded once for every screen that can hit them.
var pages = template.Must(template.Must(
	template.Must(
		template.Must(
			template.Must(
				template.Must(
					template.Must(
						template.Must(template.New("console").Funcs(pageFuncs).Parse(pageTemplates)).
							Parse(portfolioTemplates)).
						Parse(settingsTemplates)).
					Parse(overviewTemplates)).
				Parse(ordersTemplates)).
			Parse(signalsTemplates)).
		Parse(optimizationTemplates)).
	Parse(openAPIOnboardingTemplates))

// --- dashboard --------------------------------------------------------------

type dashboardPage struct {
	Nav     string
	CSRF    string
	Refresh bool
	Notice  string
	Snap    snapshot
	Run     *runView
	// CanRestart and CanRestartSoak report that the caller wired the seam. A
	// button for something this build cannot do would be a button that answers
	// with an apology.
	CanRestart     bool
	CanRestartSoak bool
}

// RefreshSeconds is the reload period the head template writes while Refresh is
// true: two seconds, because this screen is following a running verification.
// The positions screen carries its own, slower period (portfolio_pages.go).
func (dashboardPage) RefreshSeconds() int { return 2 }

func (c *Console) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		c.refuse(w, http.StatusNotFound, "그런 경로는 없다",
			"이 콘솔의 화면은 개요·검증 콘솔·포지션·주문·발굴 신호·거래 이력·편입 설정·검증·리포트 아홉뿐이다.")
		return
	}
	page := dashboardPage{
		Nav:            "verify-console",
		CSRF:           c.csrf,
		Notice:         r.URL.Query().Get("notice"),
		Snap:           c.snapshot(),
		CanRestart:     c.opts.Relaunch != nil,
		CanRestartSoak: c.opts.RestartSoak != nil,
	}
	if run := c.currentRun(); run != nil {
		v := run.snapshot()
		page.Run = &v
		page.Refresh = !v.Done && !v.Awaiting
	}
	c.render(w, "dashboard", page)
}

// --- verify -----------------------------------------------------------------

type verifyPage struct {
	Nav      string
	CSRF     string
	Market   string
	Snap     snapshot
	Steps    string
	Run      *runView
	Prompt   string
	Spent    bool
	Notice   string
	Refresh  bool
	Resuming bool
	// CanRestart reports that the restart button is wired, so the page can offer
	// the process boundary rather than describe it as a chore.
	CanRestart bool
	// ShowStart reports that the screen offers a way to begin a verification.
	//
	// It is false only while a run is actually working or waiting for its
	// approval, where the action the operator needs is on the run itself. A run
	// that has *finished* leaves the controls up: on 2026-07-27 three approval
	// windows lapsed and each time the screen came back with no buttons on it at
	// all, so the only way to try again was restarting the console — three live
	// market windows spent on nothing (measurements.md M11, M18). The controls
	// coming back is not a loosening: Spent and "이어할 단계가 없다" still decide
	// whether the button does anything.
	ShowStart bool
}

// RefreshSeconds mirrors dashboardPage's: two seconds while a run is working.
func (verifyPage) RefreshSeconds() int { return 2 }

func (c *Console) handleVerify(w http.ResponseWriter, r *http.Request) {
	c.renderVerify(w, marketOf(r), r.URL.Query().Get("notice"))
}

// marketOf reads the screen's market from the request.
//
// One place, for both the query string and the form, and it normalises through
// verifylive: an unrecognised value is KR rather than an error, because the worst
// this decides is which record is displayed and the safe default is the one every
// existing record lives in.
func marketOf(r *http.Request) string {
	if v := strings.TrimSpace(r.URL.Query().Get("market")); v != "" {
		return verifylive.NormalizeMarket(v)
	}
	return verifylive.NormalizeMarket(r.PostFormValue("market"))
}

func (c *Console) renderVerify(w http.ResponseWriter, market, notice string) {
	var steps bytes.Buffer
	verifylive.WriteSteps(&steps, false, false)

	snap := c.snapshotFor(market)
	page := verifyPage{
		Nav:        "verify",
		CSRF:       c.csrf,
		Market:     snap.Market,
		Snap:       snap,
		Steps:      steps.String(),
		Notice:     notice,
		Resuming:   snap.Verify.Resume,
		CanRestart: c.opts.Relaunch != nil,
	}

	c.mu.Lock()
	page.Spent = c.spent
	c.mu.Unlock()

	page.ShowStart = true
	if run := c.currentRun(); run != nil {
		v := run.snapshot()
		page.Run = &v
		if v.Batch != nil {
			// Summary, not Prompt: the tail Prompt carries tells the reader to type
			// a string, and typing is not how this screen approves anything.
			page.Prompt = v.Batch.Summary()
		}
		// Refresh while the runner is working, never while the form is on screen.
		page.Refresh = !v.Done && !v.Awaiting
		page.ShowStart = v.Done
	}
	c.render(w, "verify", page)
}

// startModeRedo is the form value that asks for a re-measurement.
//
// It is a mode on the existing start route rather than a route of its own,
// because "start a verification" is one act and one CSRF-gated door; a second
// door would be a second place the gates have to be got right.
const startModeRedo = "redo"

// handleStart begins a verification. It sends nothing: the runner's first act is
// to ask for the approval this console cannot supply on anybody's behalf.
//
// The redo mode changes exactly one thing — which steps the runner is willing to
// walk again — and the set comes from the record, not from the form. The plan is
// rebuilt from scratch and a new nonce has to be typed either way, so there is no
// path here that re-measures anything without a person approving that list.
func (c *Console) handleStart(w http.ResponseWriter, r *http.Request) {
	market := marketOf(r)
	var redo []verifylive.StepID
	if r.PostFormValue("mode") == startModeRedo {
		set, err := c.redoSet(market)
		if err != nil {
			c.redirectVerify(w, r, "기록을 읽을 수 없다: "+err.Error()+" 아무것도 전송되지 않았다.")
			return
		}
		if len(set) == 0 {
			c.redirectVerify(w, r, "재측정할 단계가 없다 — fail·skipped 판정이 남아 있지 않다. "+
				"아무것도 전송되지 않았다.")
			return
		}
		redo = set
	} else if snap := c.readVerify(market); snap.Present && len(snap.Pending) == 0 &&
		len(snap.Cleanup) == 0 {
		// The no-op that cost a market window on 2026-07-27: every step already has
		// a terminal verdict, so an ordinary resume walks the catalogue, settles
		// nothing and finishes with no steps recorded. Refusing it here is what
		// makes the disabled button on the page more than decoration — a stale tab
		// posts the same form.
		//
		// A leftover is the exception, and it is why the Cleanup test is here: a
		// record whose steps are all terminal but which still holds an order this
		// tool could not cancel has exactly one thing left to send, and refusing
		// that start is what made the leftover unremovable in the first place.
		notice := "이어할 단계가 없다 — 모든 단계에 판정이 있다. 아무것도 전송되지 않았다."
		if len(snap.Redo) > 0 {
			notice += " 다시 측정하려면 [재측정]을 사용하라."
		}
		c.redirectVerify(w, r, notice)
		return
	}
	if _, err := c.startRun(market, redo); err != nil {
		c.redirectVerify(w, r, err.Error())
		return
	}
	c.redirectVerify(w, r, "")
}

// handleApprove is the approval, and it is a click.
//
// What has to be true for a request to reach the line that approves: it carried a
// session (or a handoff) this process issued, it carried this process's CSRF
// token, a plan is parked and waiting, and the window that plan was minted with is
// still open. That is the whole gate — no typed string (operator-console: 검증 배치
// 승인의 형식, 사용자 결정 2026-07-27), and nothing that a scheduler or an agent
// could satisfy on a person's behalf.
//
// The window is still verifylive's own answer (Batch.Expired), for the reason the
// comparison was borrowed before: a second rule for "too late" would drift from
// the one the terminal enforces.
func (c *Console) handleApprove(w http.ResponseWriter, r *http.Request) {
	run := c.currentRun()
	if run == nil {
		c.redirectVerify(w, r, "진행 중인 검증이 없다. 아무것도 전송되지 않았다.")
		return
	}
	view := run.snapshot()
	if !view.Awaiting || view.Batch == nil {
		c.redirectVerify(w, r, "승인을 기다리는 배치가 없다. 아무것도 전송되지 않았다.")
		return
	}

	var answer error
	if view.Batch.Expired(c.now()) {
		answer = verifylive.ErrConfirmationExpired
	}
	if !run.deliver(answer) {
		c.redirectVerify(w, r, "승인 창이 이미 닫혔다. 아무것도 전송되지 않았다.")
		return
	}
	if answer != nil {
		c.redirectVerify(w, r, refusalNotice(answer))
		return
	}
	c.redirectVerify(w, r, "")
}

// handleAbort is the operator saying no without typing anything.
func (c *Console) handleAbort(w http.ResponseWriter, r *http.Request) {
	run := c.currentRun()
	if run == nil {
		c.redirectVerify(w, r, "진행 중인 검증이 없다.")
		return
	}
	if run.deliver(verifylive.ErrRefused) {
		c.redirectVerify(w, r, "거부했다. 아무것도 전송되지 않았다.")
		return
	}
	run.cancel()
	run.note("중단을 요청했다.")
	c.redirectVerify(w, r, "중단을 요청했다.")
}

func refusalNotice(err error) string {
	switch {
	case err == nil:
		return ""
	case strings.Contains(err.Error(), verifylive.ErrConfirmationExpired.Error()):
		return "승인 창이 만료되었다. 아무것도 전송되지 않았다. 콘솔을 다시 시작해 새 승인을 받아라."
	default:
		return "승인되지 않았다. 아무것도 전송되지 않았고 이 실행은 중단되었다."
	}
}

func (c *Console) redirectVerify(w http.ResponseWriter, r *http.Request, notice string) {
	target := "/verify?market=" + urlQueryEscape(marketOf(r))
	if strings.TrimSpace(notice) != "" {
		target += "&notice=" + urlQueryEscape(notice)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// --- report -----------------------------------------------------------------

type reportPage struct {
	Nav     string
	Refresh bool
	Snap    snapshot
	Text    string
	Error   string
}

func (c *Console) handleReport(w http.ResponseWriter, r *http.Request) {
	market := marketOf(r)
	page := reportPage{Nav: "report", Snap: c.snapshotFor(market)}
	rep, err := c.report(market)
	if err != nil {
		page.Error = err.Error()
	} else {
		var buf bytes.Buffer
		rep.WriteText(&buf)
		page.Text = buf.String()
	}
	c.render(w, "report", page)
}

func (c *Console) handleReportJSON(w http.ResponseWriter, r *http.Request) {
	rep, err := c.report(marketOf(r))
	if err != nil {
		c.refuse(w, http.StatusInternalServerError, "리포트를 읽을 수 없다", err.Error())
		return
	}
	body, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		c.refuse(w, http.StatusInternalServerError, "리포트를 인코딩할 수 없다", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="capability-verify-report.json"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

// --- rendering ----------------------------------------------------------------

func (c *Console) render(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := pages.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "console: rendering "+name+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// refuse is every "no" this console says, in one shape: a status, a reason, and
// the sentence that matters — that nothing was sent.
func (c *Console) refuse(w http.ResponseWriter, status int, title, detail string) {
	var buf bytes.Buffer
	_ = pages.ExecuteTemplate(&buf, "refuse", struct{ Title, Detail string }{title, detail})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

// urlQueryEscape keeps the notice readable in the address bar without pulling in
// a dependency on net/url for one call.
func urlQueryEscape(s string) string {
	var b strings.Builder
	for _, r := range []byte(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.', r == '~':
			b.WriteByte(r)
		case r == ' ':
			b.WriteByte('+')
		default:
			fmt.Fprintf(&b, "%%%02X", r)
		}
	}
	return b.String()
}
