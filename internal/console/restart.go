package console

// restart.go is the console restarting itself, and the soak, from a browser
// (openspec change verify-execution-capability, task 1.8 ①②).
//
// # Why a button for something Ctrl-C already does
//
// The conditional-persistence measurement can only be made by a process that did
// not register the conditional (see the package doc and Console.spent). Until now
// the way past that boundary was written on the page in words: quit the console,
// start it again, press 이어하기. The operator drives this whole change from a
// browser — that is what task 1.6 settled — so an instruction to go back to a
// terminal is a step the tool asks a person to perform on its behalf, and it is the
// only one left.
//
// So the console re-executes itself. What that buys is precisely one thing: a new
// process instance. The evidence record's process.instance_id is minted per
// startup, and internal/verifylive judges the persistence boundary on it and never
// on the PID — which a re-exec preserves. That judgement is pinned from both sides
// in internal/verifylive (TestTheProcessBoundaryIsTheInstanceIDAndNotThePID and
// TestTheProcessBoundaryIsNeverJudgedByPID), because if it ever moved to the PID
// this button would silently stop being a process boundary while still looking like
// one.
//
// # What a restart is, and what it is not
//
// It is a tool restart. It flips no gate, changes no configuration, places and
// cancels nothing, and cannot be reached without the session cookie and the CSRF
// token. What it resets is this process's one-verification cap; what it cannot
// reset is anything on the record, because the record is on disk and the resumed
// run reads it. The batch approval is asked for again from scratch, with a new
// nonce, exactly as it would be after a Ctrl-C.
//
// # Seams, because a test must not fork
//
// Relaunch and RestartSoak are functions the caller supplies. Nothing in this
// package executes a program, and nothing in this package's tests does either: they
// hand the console a function that records the call and assert on the recording.
// cmd/tossctl holds the one implementation of each.

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Errors the restart paths produce before anything happens.
var (
	// errNoRelaunch means the binary has no way to re-execute itself. The button
	// is not drawn, and this is the answer if the route is reached anyway.
	errNoRelaunch = errors.New(
		"이 빌드에는 재시작 배선이 없다. 콘솔을 종료(Ctrl-C)하고 직접 다시 시작하라")
	// errRunInFlight refuses a restart while a verification is walking the
	// catalogue. Restarting then would abandon a run that may have a live order
	// resting, which is the one thing this button must never do.
	errRunInFlight = errors.New(
		"검증이 진행 중이다. 먼저 끝내거나 거부한 뒤에 재시작하라 — 진행 중인 실행을 " +
			"프로세스째 버리면 계좌에 살아 있는 주문이 남을 수 있다")
	// errNoSoakRestart means the soak restart is not wired.
	errNoSoakRestart = errors.New(
		"이 빌드에는 soak 재기동 배선이 없다")
)

// Handoff carries one already-authenticated browser across a restart.
//
// It is an interface so this package writes no file: internal/handoff owns the
// 0600 file, the single-use rule and the window, and cmd/tossctl supplies it. A
// console built without one restarts perfectly well — the operator reads the new
// session URL off the terminal, which is what happened before this existed and
// remains the root of trust either way.
type Handoff interface {
	// Mint writes a fresh single-use token and returns it.
	Mint(now time.Time) (string, error)
	// Consume spends the waiting token. It returns an error for every outcome
	// except "this is the token, and it is still inside its window".
	Consume(token string, now time.Time) error
}

// Relaunch replaces this process with a fresh instance of the same binary.
//
// port is the loopback port the new instance must bind. It is passed rather than
// inferred because the browser is already sitting on that address: a restart that
// came back on a different port would be a restart the operator has to chase
// through a terminal, which is the thing this exists to remove.
//
// An implementation does not return on success.
type Relaunch func(port int) error

// RestartSoak stops the running read-only survey and starts it again, detached.
//
// It returns a line for the operator describing what it did. It never touches a
// credential and it never places anything: the survey it restarts is
// structurally incapable of mutating an account (internal/soak's package doc).
type RestartSoak func() (string, error)

// --- the console restarting itself ------------------------------------------------

// handleRestart re-executes the console.
//
// The response is written before the process is replaced: the page the browser is
// left holding is the one that sends it back in a few seconds, carrying the
// handoff token, and if the relaunch fails the operator is looking at a live
// console with a notice on it rather than at a dead socket.
func (c *Console) handleRestart(w http.ResponseWriter, r *http.Request) {
	if c.opts.Relaunch == nil {
		c.redirectDashboard(w, r, errNoRelaunch.Error())
		return
	}
	if run := c.currentRun(); run != nil && !run.finished() {
		c.redirectDashboard(w, r, errRunInFlight.Error())
		return
	}

	port, ok := c.listeningPort()
	if !ok {
		c.redirectDashboard(w, r, "콘솔이 아직 포트를 열지 않았다. 잠시 후 다시 시도하라.")
		return
	}

	// The handoff is minted here and nowhere else: this request has already
	// cleared the session cookie and the CSRF token, which is the whole warrant
	// for the new process trusting the browser that presents it.
	token, note := c.mintHandoff()

	page := restartPage{
		Target: restartTarget(token),
		Note:   note,
		Delay:  restartRefreshDelay,
	}
	c.render(w, "restart", page)
	c.requestRelaunch(port)
}

// restartRefreshDelay is how long the browser waits before coming back.
//
// Three seconds. A re-exec and a socket rebind are milliseconds; the rest is for a
// machine that is busy, and for the operator to read what is on the page. Coming
// back too early costs a browser error page on a console that is in fact fine,
// which is the worst possible way for a working restart to look.
const restartRefreshDelay = 3

// restartTarget is where the browser goes next: the same origin, because the new
// process binds the same port.
func restartTarget(token string) string {
	if token == "" {
		return pathVerifyConsole
	}
	return pathVerifyConsole + "?handoff=" + urlQueryEscape(token)
}

// mintHandoff produces the token, or explains why the browser will have to be
// pointed at the terminal instead.
func (c *Console) mintHandoff() (token, note string) {
	if c.opts.Handoff == nil {
		return "", "이 빌드는 브라우저 핸드오프를 쓰지 않는다. 새 세션 링크는 터미널에 출력된다."
	}
	token, err := c.opts.Handoff.Mint(c.now())
	if err != nil {
		return "", "핸드오프 토큰을 쓸 수 없었다(" + err.Error() + "). 터미널에 출력되는 새 세션 링크를 사용하라."
	}
	return token, ""
}

// requestRelaunch hands the port to Serve, which is the only place that may take
// the socket down.
//
// The handler cannot do it itself: it is still holding the connection it just
// answered on, and the new process cannot bind a port this one has not released.
func (c *Console) requestRelaunch(port int) {
	select {
	case c.relaunch <- port:
	default:
		// A second press while the first is in flight. One restart is enough.
	}
}

// listeningPort reads the port off the address the listener actually got, which
// is the only place it is known when --port was 0.
func (c *Console) listeningPort() (int, bool) {
	addr := c.Addr()
	if addr == "" {
		return 0, false
	}
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0, false
	}
	port, err := strconv.Atoi(addr[idx+1:])
	if err != nil || port <= 0 {
		return 0, false
	}
	return port, true
}

// --- the browser coming back --------------------------------------------------

// acceptHandoff is the second half of session0's first-visit exchange.
//
// It answers one question — is this the browser the previous process was talking
// to, arriving inside the window — and it can only ever answer it once, because
// Consume spends the token whatever the answer is.
//
// It grants a session and nothing else. Every rail the session gate protects is
// still in front of everything: the CSRF token is this process's, the approval
// nonce is minted fresh and typed by a person, and Console.spent starts false
// because this is a different process, which is the entire reason the restart
// happened.
func (c *Console) acceptHandoff(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("handoff")
	if token == "" {
		return false
	}
	if c.opts.Handoff == nil {
		return false
	}
	if err := c.opts.Handoff.Consume(token, c.now()); err != nil {
		c.refuse(w, http.StatusForbidden, "재시작 핸드오프를 받아들일 수 없다",
			"이 링크는 1회용이고 유효 시간이 짧다 — 이미 사용했거나, 만료되었거나, 일치하지 않는다. "+
				"터미널에 출력된 새 세션 링크(http://127.0.0.1:PORT/?session=...)를 사용하라. "+
				"아무것도 전송되지 않았다.")
		return true
	}
	c.grantSession(w, r)
	return true
}

// grantSession sets the cookie and redirects with the credential stripped out of
// the address bar. It is shared by the session-token path and the handoff path so
// the two cannot drift into setting different cookies.
func (c *Console) grantSession(w http.ResponseWriter, r *http.Request) {
	if c.remote != nil {
		peer, ok := c.remote.peer(r)
		if !ok {
			c.refuse(w, http.StatusForbidden, "원격 접속 주소를 확인할 수 없다",
				"재시작 핸드오프 세션을 만들지 않았다.")
			return
		}
		if err := c.remote.record(RemoteAccessEvent{
			Action: RemoteAccessLogin, Peer: peer.String(), Detail: "restart handoff",
		}); err != nil {
			c.refuse(w, http.StatusServiceUnavailable, "접근 감사를 기록할 수 없다",
				"감사 기록 없이 원격 세션을 만들지 않았다.")
			return
		}
		c.remote.issueSession(w, r, peer)
		target := *r.URL
		q := target.Query()
		q.Del("handoff")
		target.RawQuery = q.Encode()
		http.Redirect(w, r, target.RequestURI(), http.StatusSeeOther)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    c.session,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	target := *r.URL
	q := target.Query()
	q.Del("session")
	q.Del("handoff")
	target.RawQuery = q.Encode()
	http.Redirect(w, r, target.RequestURI(), http.StatusSeeOther)
}

// --- the soak ------------------------------------------------------------------

// handleSoakRestart stops and restarts the read-only survey.
//
// The console does not do this itself and cannot: RestartSoak is cmd/tossctl's, it
// signals a process and spawns one, and nothing in this package knows how. What
// the console contributes is the gate in front of it.
func (c *Console) handleSoakRestart(w http.ResponseWriter, r *http.Request) {
	c.openAPIMu.Lock()
	defer c.openAPIMu.Unlock()
	if c.opts.CheckOpenAPI != nil {
		result := c.opts.CheckOpenAPI(r.Context())
		switch result.State {
		case OpenAPICredentialsReady:
			// Continue to the existing process seam below.
		case OpenAPICredentialsMissing:
			http.Redirect(w, r, "/openapi/login?reason=missing", http.StatusSeeOther)
			return
		case OpenAPICredentialsRejected:
			http.Redirect(w, r, "/openapi/login?reason=rejected", http.StatusSeeOther)
			return
		case OpenAPICredentialsSavedNotStarted:
			http.Redirect(w, r, "/openapi/login?reason=pending", http.StatusSeeOther)
			return
		default:
			c.redirectDashboard(w, r, credentialMessage(result,
				"Open API 자격증명을 확인하지 못했다. soak은 시작하지 않았다."))
			return
		}
	}
	c.restartSoakLocked(w, r)
}

func (c *Console) restartSoakLocked(w http.ResponseWriter, r *http.Request) {
	if c.opts.RestartSoak == nil {
		c.redirectDashboard(w, r, errNoSoakRestart.Error())
		return
	}
	note, err := c.opts.RestartSoak()
	if err != nil {
		c.redirectDashboard(w, r, "soak 재기동 실패: "+err.Error())
		return
	}
	if strings.TrimSpace(note) == "" {
		note = "soak을 다시 시작했다."
	}
	c.redirectDashboard(w, r, note)
}

// redirectDashboard sends a restart or soak-restart result back to the screen
// whose button produced it. That screen is the verification console: reading
// the outcome of a button somewhere the button is not is how an operator ends
// up pressing it twice.
func (c *Console) redirectDashboard(w http.ResponseWriter, r *http.Request, notice string) {
	target := pathVerifyConsole
	if strings.TrimSpace(notice) != "" {
		target += "?notice=" + urlQueryEscape(notice)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// --- the page ------------------------------------------------------------------

// restartPage is the interstitial the browser holds while the process is replaced.
type restartPage struct {
	Nav    string
	Target string
	Note   string
	Delay  int
}

// Refresh is what the head template reads. The restart page refreshes to a
// different URL than itself, so it renders its own meta tag rather than the
// shared two-second one.
func (restartPage) Refresh() bool { return false }

// RefreshContent is the meta refresh: wait, then come back to the new process.
func (p restartPage) RefreshContent() string {
	return fmt.Sprintf("%d; url=%s", p.Delay, p.Target)
}
