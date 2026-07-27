// Package console is the local operator console for the one-off live-account
// verification (openspec change verify-execution-capability, task 1.6).
//
// # What it is, and what it is not
//
// It is a stopgap operator surface: one person, one loopback socket, one
// verification. It exists because the measurement the change needs
// (internal/verifylive) is easier to read and answer on a screen than in a
// terminal, and for no other reason. It is **not** the Phase 4 web daemon — it
// holds no scheduler, no engine, no multi-user notion, no remote access, and it
// must not grow one. When Phase 4 arrives this package is deleted, not extended.
//
// # The approval is the whole design
//
// tasks.md 1.6 permits the typed confirmation to move from the TTY to a browser
// form only under an equivalence condition, and every line of this package is
// that condition:
//
//	session      possession of the terminal that started the console is the
//	             authentication. A crypto-random token is printed in the URL and
//	             exchanged for a cookie on first visit; it is valid for as long as
//	             this process lives, and the URL exchange happens once because the
//	             cookie replaces it, not because the token is spent. Every route
//	             requires it, and a request without it is refused before it can
//	             reach anything. (The single-use credential is the restart handoff
//	             token — restart.go — which is a different thing.)
//	CSRF         every state-changing POST carries a second token, minted with the
//	             session and never printed anywhere a page can be tricked into
//	             replaying.
//	nonce        the approval itself is verifylive's own Batch nonce, displayed on
//	             the page and typed back by the person reading it, checked by
//	             verifylive.Batch.Verify — the same comparison, the same TTL, the
//	             same refusal.
//
// All three, or nothing is sent. The runner's rails are untouched: this package
// supplies a verifylive.BatchConfirmer and nothing else, so the plan authorisation,
// the exposure caps, the cancel paths and ErrOutsidePlan behave exactly as they do
// at a terminal. There is no route that places an order, toggles a gate, edits
// configuration or reveals a credential; everything except the approval form is
// read-only.
//
// --confirm-each is deliberately not offered here. The per-mutation gate is a
// terminal affordance — the operator is already watching a stream — and a web
// surface that asked fifteen separate questions would be a surface people click
// through. The console is batch-only, and cmd/tossctl's console wiring passes a
// per-mutation confirmer that refuses, so the finer gate fails closed rather than
// silently approving if it is ever reached.
//
// # The process boundary is preserved
//
// verifylive's conditional-persistence step cannot pass inside the process that
// registered the conditional. The console therefore performs at most one
// step-walking verification per process: once a run has written a step, starting
// another is refused and the page says to stop the console and start it again. A
// restarted console reads the record, notices the unfinished verification and
// labels itself a resume.
package console

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// SessionCookie is the cookie the URL token is exchanged for on first visit.
const sessionCookie = "tossos_console_session"

// Errors the console can produce before it serves anything.
var (
	// ErrNonLoopback means the listener is not bound to a loopback address.
	//
	// It is checked against the listener rather than against a configured string
	// because a string is what a future flag would set: the console must refuse an
	// externally reachable socket whoever opened it. There is no interface option
	// and there is no flag; this is the second lock on the same door.
	ErrNonLoopback = errors.New(
		"console: refusing to serve on a non-loopback address — this console authenticates by possession " +
			"of the terminal that started it and is only ever reachable from this machine")
	// ErrNoVerifyWiring means the console was built without a way to run the
	// verification. It is a programming error, refused at construction.
	ErrNoVerifyWiring = errors.New("console: a verification starter is required")
)

// StartVerify runs one verification to completion.
//
// It is the console's only route to a live account, and it is supplied by
// cmd/tossctl's console wiring — never by a flag, an environment variable or a
// default. confirm is the console's web batch confirmer, which is what makes the
// approval a person's; out receives the runner's operator progress verbatim.
//
// redo is the re-measurement set (task 1.7): the steps whose terminal verdict the
// operator asked to attempt again. It is empty for an ordinary run. It changes
// which steps the runner walks and nothing else — the plan is still rebuilt from
// scratch and the batch approval is still asked for with a new nonce, so a redo
// cannot send anything a fresh run could not.
//
// The console never sees credentials, a broker or a record path through it: the
// three things it would need to send a request of its own.
type StartVerify func(
	ctx context.Context,
	confirm verifylive.BatchConfirmer,
	out io.Writer,
	market string,
	redo []verifylive.StepID,
) (verifylive.Summary, []verifylive.Entry, error)

// Options configures a console.
type Options struct {
	// Port is the loopback port. Zero lets the OS pick a free one, which is the
	// default: the URL is printed either way and nothing else has to find it.
	Port int

	// StartVerify is how a verification is run. Required.
	StartVerify StartVerify

	// SoakRecord, VerifyRecord and Attestation are the local files the dashboard
	// reads. All three are optional: a missing file is a state the dashboard
	// reports, not an error.
	SoakRecord   string
	VerifyRecord string
	// VerifyRecordUS is the US market's evidence record. A capability verdict
	// belongs to an account and a market, so the two never share a file.
	VerifyRecordUS string
	Attestation    string

	// MinSoakDays is the consecutive-day bar the soak is judged against.
	MinSoakDays int
	// RequiredEndpoints are the endpoints the engine's automation gate needs
	// attested. Supplied by the caller so this package does not import the engine.
	RequiredEndpoints []string

	// Now is the clock. Injectable so the tests can drive an expiry.
	Now func() time.Time
	// Out receives the console's own operator lines: the URL, and what is left
	// live on the account when it shuts down.
	Out io.Writer

	// Relaunch re-executes this binary so a NEW process instance starts, which is
	// what the conditional-persistence measurement needs and what resets this
	// process's one-verification cap (task 1.8 ①). Nil hides the button and the
	// page says to restart by hand. See restart.go.
	Relaunch Relaunch
	// Handoff carries one already-authenticated browser across that restart. Nil
	// means the operator reads the new session URL off the terminal, which is what
	// happened before this existed.
	Handoff Handoff
	// RestartSoak stops the read-only survey and starts it again, detached (task
	// 1.8 ②). Nil hides the button.
	RestartSoak RestartSoak
	// Binary fingerprints the installed executable. The console takes one reading
	// at construction and compares it per render, so a console left running across
	// a reinstall says so instead of quietly being the old build. Nil uses
	// binstamp.Self.
	Binary func() (binstamp.Stamp, error)

	// --- the dashboard (change add-operator-dashboard) ---
	//
	// All three are optional and all three fail to a state the page describes
	// rather than to an error: an operator opens this console precisely when they
	// do not know what is running.

	// Holdings is the read-only broker the positions screen reads (holdings.go).
	// It declares one method and cmd/tossctl supplies it; nil leaves the broker
	// half of the screen unwired and the journal half working.
	Holdings HoldingsReader

	// JournalPath is the engine's journal database, opened read-only per request
	// with journal.OpenReadOnly. Empty means the journal half is unwired — it is
	// never resolved from the environment here, so a test cannot reach the
	// developer's real data directory by forgetting to set it.
	JournalPath string

	// RunLockPath is internal/runlock's advisory marker. A fresh one means
	// another process is spending this account's rate budget on a live
	// verification, and the broker refresh yields to it. Empty disables that
	// half of the check; the in-process half always applies.
	RunLockPath string

	// Settings is the console's one config write surface: the engine.adoption
	// block, read raw and written surgically (change console-adoption-controls;
	// settings.go). Nil leaves the settings screen read-only-with-an-explanation
	// and hides the per-symbol designation buttons.
	Settings AdoptionSettings

	// GateLimits reads the engine's automation-gate ceilings for the overview's
	// safety panel (change console-operator-overview; overview.go). It is a seam
	// of its own rather than a third method on Settings: that one writes config
	// and this one only reads it, and a screen that wants to show a limit must
	// not thereby hold the ability to edit the adoption block. Nil renders the
	// limits as seam 미배선 and leaves every other panel working.
	GateLimits GateLimitsReader

	// --- the engine (change add-engine-runtime, task 2.1) ---
	//
	// The console shows whether the engine is running and can start and stop the
	// process. It cannot make the engine *able* to trade: that is decided by the
	// §0.7-approved gate configuration and the startup interlock, both of which
	// live in the engine process. A refused start comes back as the engine's own
	// reason and is displayed verbatim, which is what makes "the console cannot
	// bypass the interlock" observable rather than merely true.

	// EngineMarker is internal/enginelock's advisory marker. It is the only thing
	// this package reads to decide whether an engine is running — the exclusion
	// itself is a flock the engine holds, and a dashboard cannot ask about a lock
	// without fighting the engine for it. Empty leaves the status section unwired.
	EngineMarker string
	// StartEngine spawns the engine process and reports what happened. Nil hides
	// the button.
	StartEngine StartEngine
	// StopEngine signals the running engine and waits for it. Nil hides the
	// button.
	StopEngine StopEngine
}

// Console is the server.
type Console struct {
	opts    Options
	now     func() time.Time
	out     io.Writer
	handler http.Handler

	// session and csrf are minted once, per process. There is no way to set
	// either from outside: a preset session token would be a non-interactive
	// approval path with extra steps.
	session string
	csrf    string

	// startedWith is the fingerprint of the binary this process was loaded from,
	// taken once. Everything after compares against it: a console that has been
	// running since before the last install is an old build wearing a current
	// page, and the dashboard says so rather than letting the operator discover it
	// through a behaviour that is missing.
	startedWith binstamp.Stamp
	// relaunch carries a restart request from the handler to Serve, which owns the
	// listener and is therefore the only thing that may release the port the new
	// process has to bind.
	relaunch chan int

	// holdings is the lazy, TTL'd cache in front of Options.Holdings. It is the
	// only thing in this process that can make a broker request of its own, and
	// holdings.go is where its rate-budget contract is written down.
	holdings *holdingsCache

	mu   sync.Mutex
	addr string
	run  *runState
	// engineNote is the last thing the engine's start or stop said. It is kept
	// rather than passed through a redirect's query string because the answer that
	// matters most — the enumerated interlock clauses a refused start printed — is
	// several lines long and has to survive the dashboard's own refresh.
	engineNote   string
	engineNoteAt time.Time
	// spent records that a verification has walked at least one step in this
	// process. The conditional-persistence measurement is "it survived the
	// process exiting", so the next attempt needs a new process and this is what
	// makes that non-negotiable.
	spent bool
}

// New builds a console.
func New(o Options) (*Console, error) {
	if o.StartVerify == nil {
		return nil, ErrNoVerifyWiring
	}
	c := &Console{
		opts:     o,
		now:      o.Now,
		out:      o.Out,
		session:  newToken(32),
		csrf:     newToken(16),
		relaunch: make(chan int, 1),
	}
	if c.now == nil {
		c.now = func() time.Time { return time.Now().UTC() }
	}
	if c.out == nil {
		c.out = io.Discard
	}
	if c.opts.Binary == nil {
		c.opts.Binary = binstamp.Self
	}
	// A console that cannot fingerprint itself keeps the zero stamp, and
	// binstamp.Stamp.Same then answers "unchanged": an unanswerable question must
	// never turn into a warning the operator cannot act on.
	c.startedWith, _ = c.opts.Binary()
	c.holdings = newHoldingsCache(o.Holdings, holdingsTTL)
	c.handler = c.routes()
	return c, nil
}

// SessionToken returns the token the URL carries. It is minted once per process
// and stays valid for that process's lifetime; it is not single-use.
func (c *Console) SessionToken() string { return c.session }

// Handler is the console's whole HTTP surface.
func (c *Console) Handler() http.Handler { return c.handler }

// Addr reports where the console is listening, once it is.
func (c *Console) Addr() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addr
}

// URL is what an operator pastes into a browser.
func (c *Console) URL() string {
	addr := c.Addr()
	if addr == "" {
		return ""
	}
	return fmt.Sprintf("http://%s/?session=%s", addr, c.session)
}

// Listen opens the console's socket.
//
// The address is not a parameter and not an option: 127.0.0.1 is spelled here,
// once, and Serve refuses anything else regardless of who opened the listener.
func Listen(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("console: binding 127.0.0.1:%d: %w", port, err)
	}
	return ln, nil
}

// Serve runs the console until ctx is cancelled.
//
// A run in progress is not abandoned: the cancellation is handed to the runner,
// and the shutdown waits for the steps to settle so the account is not left
// mid-cancel. Whatever the run leaves live is printed afterwards with the
// runner's own naming.
func (c *Console) Serve(ctx context.Context, ln net.Listener) error {
	if err := loopbackOnly(ln); err != nil {
		_ = ln.Close()
		return err
	}
	c.mu.Lock()
	c.addr = ln.Addr().String()
	c.mu.Unlock()

	srv := &http.Server{Handler: c.handler, ReadHeaderTimeout: 10 * time.Second}
	c.writeBanner()

	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	relaunchPort := 0
	select {
	case err := <-served:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case relaunchPort = <-c.relaunch:
		fmt.Fprintf(c.out, "\n재시작 요청 — 이 프로세스를 같은 바이너리로 다시 실행한다 (포트 %d 유지).\n", relaunchPort)
	case <-ctx.Done():
	}

	c.settle()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	<-served

	if relaunchPort != 0 && c.opts.Relaunch != nil {
		// The socket is closed, so the port the new process has to bind is free.
		// A successful implementation replaces this process and never returns.
		return c.opts.Relaunch(relaunchPort)
	}
	return nil
}

// ListenAndServe is the whole lifecycle, for the command.
func ListenAndServe(ctx context.Context, o Options) error {
	c, err := New(o)
	if err != nil {
		return err
	}
	ln, err := Listen(o.Port)
	if err != nil {
		return err
	}
	return c.Serve(ctx, ln)
}

// loopbackOnly is the bind refusal.
func loopbackOnly(ln net.Listener) error {
	if ln == nil {
		return fmt.Errorf("%w: there is no listener", ErrNonLoopback)
	}
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("%w: %s is not a TCP address", ErrNonLoopback, ln.Addr())
	}
	if tcp.IP == nil || !tcp.IP.IsLoopback() {
		return fmt.Errorf("%w: %s", ErrNonLoopback, ln.Addr())
	}
	return nil
}

func (c *Console) writeBanner() {
	fmt.Fprintf(c.out, "tossctl console — %s\n", c.URL())
	fmt.Fprintf(c.out, "  loopback only. The link above carries this process's session token — valid until the\n"+
		"  console stops. Opening it in this machine's browser is what authenticates you, so do\n"+
		"  not paste it anywhere else.\n")
	fmt.Fprintf(c.out, "  read-only everywhere except the verification approval, which is a click on the\n"+
		"  screen that shows the plan. Ctrl-C stops the console.\n\n")
}

// settle waits for a run in progress, then reports what is still live.
func (c *Console) settle() {
	c.mu.Lock()
	run := c.run
	c.mu.Unlock()
	if run == nil || run.finished() {
		c.writeOutstanding(run)
		return
	}

	fmt.Fprintf(c.out, "\nstopping — a verification is in progress. Waiting for its steps to settle; the account\n"+
		"must not be left mid-cancel.\n")
	run.cancel()
	select {
	case <-run.done:
	case <-time.After(settleTimeout):
		fmt.Fprintf(c.out, "the run did not settle within %s. `tossctl verify status` lists anything left live.\n",
			settleTimeout)
		return
	}
	c.writeOutstanding(run)
}

// settleTimeout bounds the wait. It is generous because a step in the middle of
// cancelling an order has to be allowed to finish cancelling it.
const settleTimeout = 2 * time.Minute

// writeOutstanding prints the leftovers with the runner's own naming, so the
// console and `tossctl verify status` describe the same account in the same words.
func (c *Console) writeOutstanding(run *runState) {
	if run == nil {
		return
	}
	summary := run.snapshot().Summary
	if len(summary.Outstanding) == 0 {
		return
	}
	fmt.Fprintf(c.out, "\nstill live on the account:\n")
	for _, a := range summary.Outstanding {
		why := "NOT CANCELLED — deal with this"
		if a.Deliberate {
			why = "left on purpose; the resumed verification cancels it"
		}
		fmt.Fprintf(c.out, "  %s %s (%s) — %s\n", a.Kind, a.ID, a.Symbol, why)
	}
}

// --- routing ------------------------------------------------------------------

func (c *Console) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", c.session0(c.handleDashboard))
	// The two dashboard screens (change add-operator-dashboard). Both are GET
	// readings and neither is behind `mutating`: there is nothing on them to
	// submit, which is what static_test.go's two route tests assert from opposite
	// directions.
	mux.HandleFunc("/positions", c.session0(c.handlePositions))
	mux.HandleFunc("/history", c.session0(c.handleHistory))
	mux.HandleFunc("/settings", c.session0(c.handleSettings))
	mux.HandleFunc("/settings/save", c.session0(c.mutating(c.handleSettingsSave)))
	mux.HandleFunc("/settings/include", c.session0(c.mutating(c.handleSettingsInclude)))
	mux.HandleFunc("/verify", c.session0(c.handleVerify))
	mux.HandleFunc("/verify/start", c.session0(c.mutating(c.handleStart)))
	mux.HandleFunc("/verify/approve", c.session0(c.mutating(c.handleApprove)))
	mux.HandleFunc("/verify/abort", c.session0(c.mutating(c.handleAbort)))
	mux.HandleFunc("/restart", c.session0(c.mutating(c.handleRestart)))
	mux.HandleFunc("/soak/restart", c.session0(c.mutating(c.handleSoakRestart)))
	// The engine's process control (add-engine-runtime task 2.1). Two acts, both
	// behind the same two gates as everything else that acts, and neither of them
	// touches the account: they start and stop a process whose *ability* to trade
	// was decided by a §0.7 gate approval and is re-checked by its own startup
	// interlock every time it comes up.
	mux.HandleFunc("/engine/start", c.session0(c.mutating(c.handleEngineStart)))
	mux.HandleFunc("/engine/stop", c.session0(c.mutating(c.handleEngineStop)))
	mux.HandleFunc("/report", c.session0(c.handleReport))
	mux.HandleFunc("/report.json", c.session0(c.handleReportJSON))
	// The operator overview (change console-operator-overview). It registers
	// itself from overview.go, which is the case the route table's static guards
	// could not see until this change widened them.
	c.registerOverview(mux)
	return mux
}

// session0 is the gate every route goes through.
//
// The name is deliberately ugly: it is the first thing in every chain and it
// should be obvious in a diff when one is missing.
func (c *Console) session0(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.hasSessionCookie(r) {
			next(w, r)
			return
		}
		// First visit: the token is in the URL. Exchange it for a cookie and
		// redirect, so the token stops being in the address bar.
		if token := r.URL.Query().Get("session"); tokenEqual(token, c.session) {
			c.grantSession(w, r)
			return
		}
		// Or the browser is coming back from a restart this console's predecessor
		// started. The token is single-use and short-lived, it was minted by an
		// already-authenticated session, and it grants a session and nothing more
		// — see restart.go.
		if c.acceptHandoff(w, r) {
			return
		}
		c.refuse(w, http.StatusForbidden, "세션 토큰이 없거나 일치하지 않는다",
			"이 콘솔은 기동할 때 출력한 세션 링크로만 열 수 있다. 그 토큰은 1회용이 아니라 "+
				"이 콘솔 프로세스가 살아 있는 동안 유효하다(1회용인 것은 재시작 핸드오프 토큰 쪽이다). "+
				"터미널에 인쇄된 http://127.0.0.1:PORT/?session=... 주소를 그대로 사용하라. "+
				"아무것도 전송되지 않았다.")
	}
}

// mutating is the second gate: a state-changing request must be a POST and must
// carry the CSRF token.
//
// "State-changing" here means the approval flow and nothing else. There is no
// other POST on this console.
func (c *Console) mutating(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			c.refuse(w, http.StatusMethodNotAllowed, "POST 전용",
				"승인 경로는 폼 제출로만 도달한다. 아무것도 전송되지 않았다.")
			return
		}
		if err := r.ParseForm(); err != nil {
			c.refuse(w, http.StatusBadRequest, "폼을 읽을 수 없다", "아무것도 전송되지 않았다.")
			return
		}
		if !tokenEqual(r.PostFormValue("csrf"), c.csrf) {
			c.refuse(w, http.StatusForbidden, "CSRF 토큰이 없거나 일치하지 않는다",
				"이 폼은 콘솔이 방금 그린 페이지에서만 제출할 수 있다. 페이지를 새로 열고 다시 시도하라. "+
					"아무것도 전송되지 않았다.")
			return
		}
		next(w, r)
	}
}

func (c *Console) hasSessionCookie(r *http.Request) bool {
	ck, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	return tokenEqual(ck.Value, c.session)
}

// tokenEqual compares in constant time. The tokens are long random strings and
// the console is on loopback, so this is belt and braces — but a token comparison
// that leaks its prefix is the kind of thing nobody notices until it matters.
func tokenEqual(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(want)) == 1
}

// newToken returns a URL-safe random token.
func newToken(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic("console: crypto/rand is unavailable: " + err.Error())
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}
