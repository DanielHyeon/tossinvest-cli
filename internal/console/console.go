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
	"crypto/tls"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"strconv"
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

	// Remote opts into the authenticated VPN console. Its zero value preserves
	// the original loopback-only terminal-possession mode.
	Remote RemoteAccess

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
	// CheckOpenAPI and SaveOpenAPI are narrow, secret-free orchestration seams.
	// The console never resolves credential paths or constructs an API client.
	CheckOpenAPI CheckOpenAPICredentials
	SaveOpenAPI  SaveOpenAPICredentials
	// Binary fingerprints the installed executable. The console takes one reading
	// at construction and compares it per render, so a console left running across
	// a reinstall says so instead of quietly being the old build. Nil uses
	// binstamp.Self.
	Binary func() (binstamp.Stamp, error)

	// SystemUpdater is bound to the running executable and its fixed
	// `.candidate`/`.rollback` siblings. The HTTP request supplies only the
	// candidate hash the operator reviewed.
	SystemUpdater SystemUpdater
	// ReleaseDownloader discovers and fully verifies the constructor-bound
	// official release. ReleaseCandidateStager publishes only its returned bytes
	// to the updater's fixed sibling path. They are separate capabilities so a
	// downloader cannot install and an installer cannot select a URL.
	ReleaseDownloader      ReleaseDownloader
	ReleaseCandidateStager ReleaseCandidateStager
	// AcquireUpdateEngineLock takes the real journal-directory engine flock.
	// Nil refuses installation; an advisory marker is not an exclusion.
	AcquireUpdateEngineLock AcquireUpdateEngineLock
	// CheckUpdateVerifyActivity is the strict external verification check run at
	// the updater's commit boundary. Nil or any error refuses installation.
	CheckUpdateVerifyActivity CheckUpdateVerifyActivity

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
	// ExitPolicies is the optimization page's load/save-only config seam. It
	// carries no broker, gate, trading-toggle, or journal authority.
	ExitPolicies ExitPolicySettings
	// Optimization is the narrow, durable settings lifecycle command service.
	// Its interface has no journal, broker, lane, gate, kill-switch, or LIVE
	// mutation method; nil leaves every lifecycle control read-only.
	Optimization OptimizationCommander
	// MarketSchedule is a read-only projection of scheduler desired/effective
	// state and its calendar provenance. Its single method returns plain display
	// data; it cannot edit configuration, start the engine, or reach a broker.
	MarketSchedule MarketScheduleReader
	// Performance is the derived performance.db read model. Its interface has one
	// query method and exposes no collector, pruning, journal, broker, config, or
	// operating-control capability.
	Performance PerformanceReader
	// StrategyRuntime is the a047 read-only dormant lane projection. It carries
	// display booleans/enums only and cannot edit a lane, start the engine, mint
	// an activation, or reach an account.
	StrategyRuntime StrategyRuntimeReader

	// Orders is the read-only view of the account's order record the orders
	// screen reads (orders.go). It declares one method, behind which the caller
	// makes the three broker calls one refresh costs — the pending order group,
	// the finished one and the conditional endpoint — so this package can neither
	// spend the budget three times over nor report one endpoint's silence as
	// another's zero. Nil leaves the
	// orders screen unmeasured with the reason seam 미배선 and every other screen
	// working.
	Orders OrdersReader

	// Signals is the discovery store's read for the /signals screen (change
	// add-candidate-discovery; signals.go). One method, behind which the caller
	// opens internal/candidate's store and runs its assessment — this package
	// never holds the handle that can promote, cool or prune, and the screen
	// causes no scan and therefore spends none of the account's rate budget. Nil
	// leaves the screen unmeasured with the reason seam 미배선; an empty list
	// would read as "the market is quiet".
	Signals SignalsReader

	// GateLimits reads the engine's automation-gate ceilings for the overview's
	// safety panel (change console-operator-overview; overview.go). It is a seam
	// of its own rather than a third method on Settings: that one writes config
	// and this one only reads it, and a screen that wants to show a limit must
	// not thereby hold the ability to edit the adoption block. Nil renders the
	// limits as seam 미배선 and leaves every other panel working.
	GateLimits GateLimitsReader

	// Limits is the settings screen's editor for those same ceilings (change
	// console-sets-guardian-limits; settings_limits.go). It is a third seam
	// rather than a write method on GateLimits, for the reason that one is a
	// third seam rather than a method on Settings: the overview shows a limit and
	// must not thereby be able to change one.
	//
	// Its Save takes config.GuardianLimits, which carries the five ceilings and
	// the currency and has no field for `enabled`. That is how "editing a ceiling
	// cannot turn the gate on" survives future edits to the handlers — there is
	// nowhere in the message to put the switch. It stays true after
	// console-owns-the-operating-toggles gave the console a way to write the
	// switch: the way is a different seam. Nil renders the limit section
	// read-only with the reason seam 미배선.
	Limits LimitSettings

	// --- the operating toggles (change console-owns-the-operating-toggles) ---

	// TradingPolicy is the editor for the four `trading` toggles the engine's
	// exit path uses. Its Save takes config.TradingPolicy — four fields, no
	// amend, no conditional, no fractional — so the three the engine never
	// reaches cannot be moved from here.
	TradingPolicy TradingPolicySettings

	// Gate writes engine.automation_gate.enabled, and only that key.
	//
	// A fifth seam rather than a method on Limits, and the separation is the
	// safety property: a limit save emits no bytes for the switch and a switch
	// save emits none for the ceilings, so neither can move the other through a
	// read taken outside the file lock. Nil renders the gate section read-only.
	Gate GateSwitch

	// EngineBoot reads and writes only engine.autostart. It is deliberately
	// separate from Gate: process lifecycle approval does not grant order
	// capability, and a gate save must not silently approve reboot starts.
	EngineBoot EngineBootSettings

	// PositionPolicies is an engine-owned local command capability. It exposes
	// no journal handle, SQL, broker, config, or operating toggle.
	PositionPolicies PositionPolicyCommander

	// Protections is an engine-owned capability. Nil is the shipped default and
	// renders OFF/UNWIRED read-only state. The console never receives a broker,
	// journal handle, activation toggle, symbol field, or numeric trigger input.
	Protections ProtectionCommander

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
	// EngineBootNote is a display-only result from the startup autostart
	// decision made by cmd/tossctl before the HTTP listener opens.
	EngineBootNote string
}

// Console is the server.
type Console struct {
	opts    Options
	now     func() time.Time
	out     io.Writer
	handler http.Handler
	remote  *remoteRuntime

	// session and csrf are minted once, per process. There is no way to set
	// either from outside: a preset session token would be a non-interactive
	// approval path with extra steps.
	session string
	csrf    string
	// policySeal signs opaque, version-bound policy action tokens. It is random
	// per process and cannot be supplied through Options.
	policySeal []byte

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
	// openAPIMu keeps a credential check/save/restart transaction on one
	// credential generation even when two browser requests arrive together.
	openAPIMu sync.Mutex

	// activityMu serializes executable installation with the two routes that can
	// start account/engine work. A successful commit leaves updateCommitted set
	// until this old process exits, so a delayed request cannot start work in the
	// binary that is about to be replaced.
	activityMu      sync.Mutex
	updateCommitted bool
	// releaseMu serializes discovery through candidate publication without
	// blocking engine/verification starts on network or TUF work.
	releaseMu sync.Mutex
	// releaseReceipt is process-local evidence for the exact candidate SHA. It
	// intentionally disappears on restart; a surviving local candidate then
	// returns to provenance-unknown until it is verified again.
	releaseReceiptMu sync.Mutex
	releaseReceipt   *signedReleaseReceipt

	// holdings is the lazy, TTL'd cache in front of Options.Holdings. It is the
	// only thing in this process that can make a broker request of its own, and
	// holdings.go is where its rate-budget contract is written down.
	holdings *holdingsCache
	// ordersCache is the lazy, TTL'd cache in front of Options.Orders. One
	// refresh through it is the three broker calls the orders screen's rate-budget
	// contract allows, and orders.go is where that contract is written down.
	ordersCache *ordersCache

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
	now := o.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	remote, err := newRemoteRuntime(o.Remote, now)
	if err != nil {
		return nil, err
	}
	c := &Console{
		opts:       o,
		now:        now,
		out:        o.Out,
		remote:     remote,
		session:    newToken(32),
		csrf:       newToken(16),
		policySeal: []byte(newToken(32)),
		relaunch:   make(chan int, 1),
	}
	if c.out == nil {
		c.out = io.Discard
	}
	if c.opts.Binary == nil {
		c.opts.Binary = binstamp.Self
	}
	if note := strings.TrimSpace(o.EngineBootNote); note != "" {
		c.engineNote = note
		c.engineNoteAt = c.now()
	}
	// A console that cannot fingerprint itself keeps the zero stamp, and
	// binstamp.Stamp.Same then answers "unchanged": an unanswerable question must
	// never turn into a warning the operator cannot act on.
	c.startedWith, _ = c.opts.Binary()
	c.holdings = newHoldingsCache(o.Holdings, holdingsTTL)
	c.ordersCache = newOrdersCache(o.Orders, ordersTTL)
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
	if c.remote != nil {
		if c.remote.trustedNetwork {
			return strings.TrimSuffix(c.remote.publicURL.String(), "/") + "/"
		}
		return strings.TrimSuffix(c.remote.publicURL.String(), "/") + "/login"
	}
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

// ListenOn opens the explicitly configured remote socket. It is only used after
// RemoteAccess has passed the all-or-nothing validation in New.
func ListenOn(bind string, port int) (net.Listener, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(bind))
	if err != nil {
		return nil, fmt.Errorf("console: remote bind must be an IP literal: %w", err)
	}
	network := "tcp6"
	if addr.Unmap().Is4() {
		network = "tcp4"
	}
	ln, err := net.Listen(network, net.JoinHostPort(bind, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("console: binding %s:%d: %w", bind, port, err)
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
	if err := c.listenerAllowed(ln); err != nil {
		_ = ln.Close()
		return err
	}
	c.mu.Lock()
	c.addr = ln.Addr().String()
	c.mu.Unlock()

	srv := &http.Server{
		Handler:           c.handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if c.remote != nil {
		srv.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{c.remote.certificate},
		}
	}
	c.writeBanner()

	served := make(chan error, 1)
	go func() {
		if c.remote != nil {
			served <- srv.ServeTLS(ln, "", "")
			return
		}
		served <- srv.Serve(ln)
	}()

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
	var ln net.Listener
	if c.remote != nil {
		ln, err = ListenOn(c.remote.bind.String(), o.Port)
	} else {
		ln, err = Listen(o.Port)
	}
	if err != nil {
		return err
	}
	return c.Serve(ctx, ln)
}

func (c *Console) listenerAllowed(ln net.Listener) error {
	if c.remote == nil {
		return loopbackOnly(ln)
	}
	if ln == nil {
		return errors.New("console: there is no remote listener")
	}
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok || tcp.IP == nil {
		return fmt.Errorf("console: remote listener %v is not a TCP address", ln.Addr())
	}
	actual, ok := netip.AddrFromSlice(tcp.IP)
	if !ok {
		return fmt.Errorf("console: remote listener has an invalid address: %s", ln.Addr())
	}
	actual = actual.Unmap()
	if actual != c.remote.bind {
		return fmt.Errorf("console: remote listener %s does not match configured bind %s", actual, c.remote.bind)
	}
	return nil
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
	if c.remote != nil {
		if c.remote.trustedNetwork {
			fmt.Fprintf(c.out, "  trusted-network VPN mode. There is no application login. TLS, allowed CIDRs,\n"+
				"  exact host/origin checks, CSRF, and action audit remain required.\n"+
				"  Ctrl-C stops the console.\n\n")
			return
		}
		fmt.Fprintf(c.out, "  remote VPN token mode. TLS, allowed CIDRs, host/origin checks, a separate login\n"+
			"  token, and an audited short-lived session are all required. The token is never printed.\n"+
			"  Ctrl-C stops the console.\n\n")
		return
	}
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
	mux.HandleFunc("/healthz", c.handleHealth)
	if c.remote != nil && !c.remote.trustedNetwork {
		mux.HandleFunc("/login", c.handleRemoteLogin)
		mux.HandleFunc("/logout", c.session0(c.mutating(c.handleRemoteLogout)))
	}
	mux.HandleFunc("/", c.session0(c.handleDashboard))
	// The two dashboard screens (change add-operator-dashboard). Both are GET
	// readings and neither is behind `mutating`: there is nothing on them to
	// submit, which is what static_test.go's two route tests assert from opposite
	// directions.
	mux.HandleFunc("/positions", c.session0(c.readOnly(c.handlePositions)))
	mux.HandleFunc("/history", c.session0(c.handleHistory))
	mux.HandleFunc("/settings", c.session0(c.handleSettings))
	mux.HandleFunc("/settings/save", c.session0(c.mutating(c.handleSettingsSave)))
	mux.HandleFunc("/settings/include", c.session0(c.mutating(c.handleSettingsInclude)))
	mux.HandleFunc("/settings/exclude", c.session0(c.mutating(c.handleSettingsExclude)))
	// The Guardian-limit editor (change console-sets-guardian-limits). Two acts,
	// both behind the same two gates, and neither can write the automation gate's
	// own switch: the seam they save through takes a type with no field for it.
	mux.HandleFunc("/settings/limits", c.session0(c.mutating(c.handleSettingsLimits)))
	mux.HandleFunc("/settings/limits/preset", c.session0(c.mutating(c.handleSettingsLimitPreset)))
	mux.HandleFunc("/settings/trading", c.session0(c.mutating(c.handleSettingsTrading)))
	// The gate switch says what it is in its path (change
	// console-owns-the-operating-toggles): the limit routes deliberately avoid
	// gate vocabulary because they cannot touch the switch, and this one must
	// carry it for the opposite reason — an audit reader has to be able to tell
	// which line turned the engine loose.
	mux.HandleFunc("/settings/gate", c.session0(c.mutating(c.handleSettingsGate)))
	mux.HandleFunc("/settings/autostart",
		c.session0(c.mutating(c.startExclusive(c.handleSettingsAutostart))))
	mux.HandleFunc("/settings/system-update/install",
		c.session0(c.mutating(c.handleSystemUpdateInstall)))
	mux.HandleFunc("/settings/system-update/download",
		c.session0(c.mutating(c.handleSystemUpdateDownload)))
	mux.HandleFunc("/optimization", c.session0(c.handleOptimization))
	mux.HandleFunc("/performance-history", c.session0(c.handlePerformanceHistory))
	mux.HandleFunc("/strategy-runtime/market-schedule", c.session0(c.handleMarketSchedule))
	mux.HandleFunc("/strategy-runtime", c.session0(c.handleStrategyRuntime))
	mux.HandleFunc("/optimization/exit-policy",
		c.session0(c.mutating(c.handleOptimizationSave, 4096)))
	mux.HandleFunc("/optimization/exit-protection/preview",
		c.session0(c.mutating(c.handleProtectionPreview, 4096)))
	mux.HandleFunc("/optimization/exit-protection/apply",
		c.session0(c.mutating(c.handleProtectionApply, 4096)))
	mux.HandleFunc("/position-management", c.session0(c.readOnly(c.handlePositionManagement)))
	mux.HandleFunc("/position-management/preview",
		c.session0(c.mutating(c.handlePositionPolicyPreview, 4096)))
	mux.HandleFunc("/position-management/apply",
		c.session0(c.mutating(c.handlePositionPolicyApply, 4096)))
	mux.HandleFunc("/verify", c.session0(c.handleVerify))
	mux.HandleFunc("/verify/start", c.session0(c.mutating(c.startExclusive(c.handleStart))))
	mux.HandleFunc("/verify/approve", c.session0(c.mutating(c.handleApprove)))
	mux.HandleFunc("/verify/abort", c.session0(c.mutating(c.handleAbort)))
	mux.HandleFunc("/restart", c.session0(c.mutating(c.handleRestart)))
	mux.HandleFunc("/soak/restart", c.session0(c.mutating(c.handleSoakRestart)))
	mux.HandleFunc("/openapi/login", c.session0(c.credentialHTTPS(c.handleOpenAPILogin)))
	mux.HandleFunc("/openapi/login/save",
		c.session0(c.credentialHTTPS(c.mutating(c.handleOpenAPILoginSave, 8192))))
	// The engine's process control (add-engine-runtime task 2.1). Two acts, both
	// behind the same two gates as everything else that acts, and neither of them
	// touches the account: they start and stop a process whose *ability* to trade
	// was decided by a §0.7 gate approval and is re-checked by its own startup
	// interlock every time it comes up.
	mux.HandleFunc("/engine/start", c.session0(c.mutating(c.startExclusive(c.handleEngineStart))))
	mux.HandleFunc("/engine/stop", c.session0(c.mutating(c.handleEngineStop)))
	mux.HandleFunc("/report", c.session0(c.handleReport))
	mux.HandleFunc("/report.json", c.session0(c.handleReportJSON))
	// The operator overview (change console-operator-overview). It registers
	// itself from overview.go, which is the case the route table's static guards
	// could not see until this change widened them.
	c.registerOverview(mux)
	// The orders screen (change console-orders-screen). Same arrangement: it
	// registers itself from orders.go, and it is the one path in this table the
	// account-verb guard grants an exception to — byte-exact, and only while the
	// `readOnly` wrapper below is on it.
	c.registerOrders(mux)
	// The discovery screen (change add-candidate-discovery, task 5.5). Same
	// arrangement again, and it touches no account at all: what it renders is a
	// read of internal/candidate's own store, whose dependency closure is
	// {internal/clock}.
	c.registerSignals(mux)
	if c.remote != nil {
		return c.remote.security(mux)
	}
	return mux
}

// session0 is the gate every route goes through.
//
// The name is deliberately ugly: it is the first thing in every chain and it
// should be obvious in a diff when one is missing.
func (c *Console) session0(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.remote != nil {
			if c.remote.trustedNetwork {
				next(w, r)
				return
			}
			if c.remote.hasSession(r) {
				next(w, r)
				return
			}
			if c.acceptHandoff(w, r) {
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
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
func (c *Console) mutating(next http.HandlerFunc, bodyLimits ...int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			c.refuse(w, http.StatusMethodNotAllowed, "POST 전용",
				"승인 경로는 폼 제출로만 도달한다. 아무것도 전송되지 않았다.")
			return
		}
		if c.remote != nil && !c.remote.sameOriginForMutation(r) {
			c.refuse(w, http.StatusForbidden, "요청 출처가 일치하지 않는다",
				"원격 쓰기 요청은 설정된 HTTPS 주소에서 시작되어야 한다. 아무것도 전송되지 않았다.")
			return
		}
		mediaType, _, mediaErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if mediaErr != nil || mediaType != "application/x-www-form-urlencoded" {
			c.refuse(w, http.StatusBadRequest, "폼 형식이 허용되지 않는다",
				"상태 변경은 콘솔이 그린 URL-encoded 폼으로만 제출할 수 있다. 아무것도 전송되지 않았다.")
			return
		}
		if len(bodyLimits) > 0 && bodyLimits[0] > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, bodyLimits[0])
		}
		if err := r.ParseForm(); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				c.refuse(w, http.StatusRequestEntityTooLarge, "요청 본문이 너무 크다",
					"Open API 자격증명 폼의 허용 크기를 넘었다. 아무것도 저장하거나 시작하지 않았다.")
				return
			}
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

// readOnly is mutating's mirror: a route that only reads must answer GET and
// HEAD and refuse every other method (change console-orders-screen, task 3.2).
//
// The name states the guarantee at the call site — c.session0(c.readOnly(h)) —
// and it is deliberately not `reading`, which is what design.md D3 first called
// it: this package already spells `reading` for the (value, measured, reason)
// triple every screen renders through (overview.go). Go would have compiled both,
// because a method name lives in its type's namespace, and that is exactly the
// kind of collision that reads fine to whoever wrote it (issues.md I-2, Manager
// ruling 2026-07-28).
//
// # Why a wrapper rather than a comment
//
// The route table's static guards read the registration, and what a registration
// carries is {path, session gate, CSRF gate}. There is no method in it. So a
// guard asked to confirm that an exempted route is "a GET" has only one thing to
// look at — whether the CSRF gate is absent — and that turns the sentence into
// "it is not protected". The exception would then be granted BECAUSE the route is
// unprotected, and in that state a POST to it is served on a session cookie with
// no CSRF token at all.
//
// This makes the method a fact of the chain instead. static_test.go reads it off
// the same registration the other two gates are read off, and the runtime
// refusal is the second lock (orders_static_test.go posts to /orders and expects
// 405).
//
// Go 1.22's method patterns — HandleFunc("GET /orders", …) — would say the same
// thing to the mux and nothing at all to the guards: the route extractor reads
// the literal as the path, so the table would hold "GET /orders" and every
// path-keyed comparison in static_test.go would silently stop matching.
//
// It is not applied to every read route on this console. The rest are covered by
// not naming an account verb in the first place, and a wrapper on all of them
// would make the one place it is load-bearing invisible.
func (c *Console) readOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다",
				"이 경로는 조회만 한다. 주문을 내거나 정정·취소하는 수단은 이 콘솔에 없다. "+
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
