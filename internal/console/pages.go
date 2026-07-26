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

var pages = template.Must(template.New("console").Parse(pageTemplates))

// --- dashboard --------------------------------------------------------------

type dashboardPage struct {
	Nav     string
	Refresh bool
	Snap    snapshot
	Run     *runView
}

func (c *Console) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		c.refuse(w, http.StatusNotFound, "그런 경로는 없다",
			"이 콘솔의 화면은 대시보드·검증·리포트 셋뿐이다.")
		return
	}
	page := dashboardPage{Nav: "dashboard", Snap: c.snapshot()}
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
	Snap     snapshot
	Steps    string
	Run      *runView
	Prompt   string
	Spent    bool
	Notice   string
	Refresh  bool
	Resuming bool
}

func (c *Console) handleVerify(w http.ResponseWriter, r *http.Request) {
	c.renderVerify(w, r.URL.Query().Get("notice"))
}

func (c *Console) renderVerify(w http.ResponseWriter, notice string) {
	var steps bytes.Buffer
	verifylive.WriteSteps(&steps, false)

	snap := c.snapshot()
	page := verifyPage{
		Nav:      "verify",
		CSRF:     c.csrf,
		Snap:     snap,
		Steps:    steps.String(),
		Notice:   notice,
		Resuming: snap.Verify.Resume,
	}

	c.mu.Lock()
	page.Spent = c.spent
	c.mu.Unlock()

	if run := c.currentRun(); run != nil {
		v := run.snapshot()
		page.Run = &v
		if v.Batch != nil {
			page.Prompt = v.Batch.Prompt()
		}
		// Refresh while the runner is working, never while the form is on screen.
		page.Refresh = !v.Done && !v.Awaiting
	}
	c.render(w, "verify", page)
}

// handleStart begins a verification. It sends nothing: the runner's first act is
// to ask for the approval this console cannot supply on anybody's behalf.
func (c *Console) handleStart(w http.ResponseWriter, r *http.Request) {
	if _, err := c.startRun(); err != nil {
		c.redirectVerify(w, r, err.Error())
		return
	}
	c.redirectVerify(w, r, "")
}

// handleApprove is the typed confirmation.
//
// verifylive.Batch.Verify is the judge — the same call the terminal makes, with
// the same TTL and the same exact-match rule — and its answer goes to the runner
// whatever it is. There is no branch here that approves anything.
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

	answer := view.Batch.Verify(r.PostFormValue("nonce"), c.now())
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
		return "확인 문자열이 만료되었다. 아무것도 전송되지 않았다. 콘솔을 다시 시작해 새 승인을 받아라."
	default:
		return "입력한 문자열이 일치하지 않는다. 아무것도 전송되지 않았고 이 실행은 중단되었다."
	}
}

func (c *Console) redirectVerify(w http.ResponseWriter, r *http.Request, notice string) {
	target := "/verify"
	if strings.TrimSpace(notice) != "" {
		target += "?notice=" + urlQueryEscape(notice)
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

func (c *Console) handleReport(w http.ResponseWriter, _ *http.Request) {
	page := reportPage{Nav: "report", Snap: c.snapshot()}
	rep, err := c.report()
	if err != nil {
		page.Error = err.Error()
	} else {
		var buf bytes.Buffer
		rep.WriteText(&buf)
		page.Text = buf.String()
	}
	c.render(w, "report", page)
}

func (c *Console) handleReportJSON(w http.ResponseWriter, _ *http.Request) {
	rep, err := c.report()
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
	w.Header().Set("Referrer-Policy", "no-referrer")
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
