package engine

// runtime.go is the engine's lifecycle supervisor (openspec change
// add-engine-runtime, task 1.3; engine-safety "엔진 런타임 수명주기").
//
// # What was missing
//
// Every loop this engine has was built, tested and left without a caller.
// `ReconcileDriver.Run`, `ExitObserver.Run` and `filldetect.Detector.Run` are
// three goroutines nothing started, which meant the questions a supervisor
// answers had never been asked: what happens when one of them returns, and what
// happens when one of them stays alive and stops working. This file answers both,
// and it deliberately answers them separately.
//
// # The two layers, and why one was not enough
//
//	① defensive termination   a loop that returns for any reason other than the
//	                          runtime being cancelled stops the whole runtime,
//	                          raises a critical alert, and makes the process exit
//	                          non-zero.
//	② sustained degradation   a loop that is alive but whose cycles keep failing
//	                          crosses a consecutive-failure threshold, which
//	                          raises a critical alert and tightens the operating
//	                          mode to ENTRY_BLOCKED. The loop keeps retrying.
//
// Layer ① alone would be a supervisor that never fires: the landed loops return
// only on context cancellation ("실패한 사이클은 다음 주기에 재시도" —
// reconcileloop.go, exitloop.go, detect.go all say so in their own words). Layer
// ② alone would leave a loop that panicked its way out of existence unnoticed
// until somebody read a log. Together they are the "부분 생존 금지" rule: the
// engine is all of its loops or none of them, and an engine that is technically
// all of them while two are blind stops taking new exposure.
//
// # Why a cancelled loop is not an incident
//
// A graceful stop is the operator pressing Ctrl-C or the console's [엔진 정지]
// button, and a critical alert for that would be an alert for the one engine
// event that is always deliberate. engine-safety says it outright: 컨텍스트
// 취소에 의한 반환은 정상 종료이며 critical을 발송하지 않는다(SHALL NOT). So the
// classification is "did this loop return because we cancelled it", not "did this
// loop return an error".
//
// # What this file does not own
//
// It constructs no loop. The loops are handed in as [SupervisedLoop] values,
// which is what lets `tossctl engine run` assemble the fill detector — whose
// package transitively imports the WTS client through internal/push, and is
// therefore unspellable inside internal/app/engine (deps_test.go) — while the
// supervision contract stays here with the loops it can name. It also holds no
// lock and no marker: single-instance exclusion is the journal directory's flock,
// taken by the command before any of this exists.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// DefaultDegradationThreshold is how many consecutive failed cycles make a live
// loop an incident (engine-safety: 연속 5주기 실패).
//
// Five, and the number is the same for both supervised loops on purpose. At the
// reconciliation period (60 s) it is five minutes of not knowing what the account
// holds; at the fill-detection period (3 s) it is fifteen seconds of not seeing
// fills. The two durations are wildly different and that is correct — the
// threshold is counted in *cycles* because what it measures is "this is not a
// blip", and a blip is one cycle whatever the cycle is worth.
const DefaultDegradationThreshold = 5

// DefaultHealthInterval is how often the supervisor reads the loops' health.
//
// One second: fast enough that the fill detector's fifteen-second threshold is
// not materially delayed, slow enough that the poll costs nothing. It reads
// counters in memory and makes no request of any kind.
const DefaultHealthInterval = time.Second

// ErrLoopFailed is the defensive termination contract firing: a supervised loop
// returned for a reason other than the runtime being cancelled.
//
// It is a process-ending error rather than a restart, because none of these loops
// has a documented reason to return at all. A supervisor that restarted an
// unexplained return would be guessing that the state it returned from is
// recoverable, and the state it returned from is a live account.
var ErrLoopFailed = errors.New("engine: a supervised loop stopped for a reason that is not a cancellation")

// ErrRuntimeUnavailable means the runtime could not be assembled.
var ErrRuntimeUnavailable = errors.New("engine: the runtime could not be assembled")

// LoopHealth is a loop that can report how many of its cycles have failed in a
// row.
//
// Both supervised loops already answer this without any new bookkeeping —
// ReconcileDriver.Health() and filldetect.Detector.Health().Outage.Consecutive —
// so the supervisor asks one question of one interface rather than knowing two
// loops' internals.
type LoopHealth interface {
	// ConsecutiveFailures is the length of the current run of failed cycles.
	// Zero means the last cycle succeeded.
	ConsecutiveFailures() int
}

// SupervisedLoop is one loop under the runtime's supervision.
type SupervisedLoop struct {
	// Name identifies the loop in alerts and logs. Required.
	Name string
	// Run is the loop. Required. It is expected to return only when ctx ends.
	Run func(ctx context.Context) error
	// Health reports the loop's consecutive cycle failures. Optional: a loop with
	// none is covered by the defensive termination contract only, which is the
	// right arrangement for the exit observer — its degradation ladder is its own
	// landed 60-second outage contract and a second threshold on top would be two
	// definitions of "exit observation is down".
	Health LoopHealth
	// Trigger is the journal ModeTrigger the degradation threshold escalates
	// with. Required when Health is set, and validated against the journal's
	// closed enumeration at assembly time — a supervisor that discovered its
	// trigger was unknown at the moment it needed to escalate would alert and
	// change nothing.
	Trigger string
}

// RuntimeOptions is the supervisor's wiring.
type RuntimeOptions struct {
	// Loops are the loops to run. At least one is required.
	Loops []SupervisedLoop
	// Auxiliary is work the runtime starts and drains but does not supervise.
	// Optional. See AuxiliaryExecutor (auxiliary.go) for why alert delivery
	// cannot be a supervised loop: a delivery fault must not stop the loop that
	// places a stop-loss.
	//
	// The runtime still waits for everything here before it returns, because the
	// caller closes the journal the moment it does.
	Auxiliary []AuxiliaryExecutor
	// AccountRef scopes the escalation. Required when any loop carries Health:
	// an operating-mode transition is per account, and an unscoped one is not a
	// transition anybody can act on.
	AccountRef string
	// Alerts receives the critical alerts both layers raise. Optional only so the
	// loop can be unit tested; *obs.Notifier is what production passes.
	Alerts ExitAlerter
	// Escalate persists the ENTRY_BLOCKED tightening. Optional; nil leaves the
	// alert as the only consequence. *journal.Journal satisfies it.
	Escalate execgw.ModeEscalation
	// Announcer is told about a transition that changed something. Optional.
	Announcer journal.ModeAnnouncer
	// Log receives the structured lines. Optional.
	Log *obs.Logger
	// Clock drives the health poll. Defaults to the system clock.
	Clock clock.Clock
	// Threshold overrides DefaultDegradationThreshold.
	Threshold int
	// HealthInterval overrides DefaultHealthInterval.
	HealthInterval time.Duration
	// Recover is the restart recovery sequence, run to completion before any loop
	// starts. Optional; nil skips it, which is what a unit test wants and what no
	// production wiring does — `tossctl engine run` passes reconcile.Recovery's,
	// and a failure there stops the start rather than beginning to trade over an
	// unresolved attempt.
	Recover func(ctx context.Context) error
}

// Runtime supervises the engine's loops.
type Runtime struct {
	opts      RuntimeOptions
	clk       clock.Clock
	threshold int
	interval  time.Duration

	// escalated latches the degradation alert per loop, so one sustained outage
	// is one alert and one transition rather than one per health poll. It clears
	// when the loop's cycles recover, which is what makes a second outage
	// audible.
	mu        sync.Mutex
	escalated map[string]bool
}

// NewRuntime validates the wiring.
//
// Everything it refuses is a wiring mistake that would otherwise be discovered by
// a supervisor at the moment it was needed: a loop with no Run, a health source
// with no trigger, a trigger the journal's closed enumeration does not know.
func NewRuntime(opts RuntimeOptions) (*Runtime, error) {
	if len(opts.Loops) == 0 {
		return nil, fmt.Errorf("%w: no loops were supplied; a runtime with nothing to supervise is a "+
			"process that would report success while trading nothing", ErrRuntimeUnavailable)
	}
	seen := map[string]bool{}
	for _, loop := range opts.Loops {
		name := strings.TrimSpace(loop.Name)
		switch {
		case name == "":
			return nil, fmt.Errorf("%w: a supervised loop has no name; the alert it raises has to say "+
				"which loop stopped", ErrRuntimeUnavailable)
		case seen[name]:
			return nil, fmt.Errorf("%w: two supervised loops are both called %q", ErrRuntimeUnavailable, name)
		case loop.Run == nil:
			return nil, fmt.Errorf("%w: the %s loop has no Run function", ErrRuntimeUnavailable, name)
		}
		seen[name] = true
		if loop.Health == nil {
			continue
		}
		if !journal.AutomaticTrigger(loop.Trigger) {
			return nil, fmt.Errorf("%w: the %s loop escalates with %q, which is not one of the "+
				"enumerated automatic triggers (%s)", ErrRuntimeUnavailable, name, loop.Trigger,
				strings.Join(journal.AutomaticTriggers(), ", "))
		}
		if strings.TrimSpace(opts.AccountRef) == "" {
			return nil, fmt.Errorf("%w: the %s loop can tighten the operating mode and no account was "+
				"named; a transition is scoped to one account", ErrRuntimeUnavailable, name)
		}
	}

	// Auxiliary executors are validated against the same seen set, not one of
	// their own: the name is what a stop record points at, and two things the
	// runtime starts sharing one name make that record ambiguous at exactly the
	// moment somebody is reading it to find out what died.
	for _, aux := range opts.Auxiliary {
		name := strings.TrimSpace(aux.Name)
		switch {
		case name == "":
			return nil, fmt.Errorf("%w: an auxiliary executor has no name; the record of its stop has "+
				"to say what stopped", ErrRuntimeUnavailable)
		case seen[name]:
			return nil, fmt.Errorf("%w: %q is already the name of something this runtime starts",
				ErrRuntimeUnavailable, name)
		case aux.Run == nil:
			return nil, fmt.Errorf("%w: the %s executor has no Run function", ErrRuntimeUnavailable, name)
		}
		seen[name] = true
	}

	r := &Runtime{
		opts:      opts,
		clk:       opts.Clock,
		threshold: opts.Threshold,
		interval:  opts.HealthInterval,
		escalated: map[string]bool{},
	}
	if r.clk == nil {
		r.clk = clock.System()
	}
	if r.threshold <= 0 {
		r.threshold = DefaultDegradationThreshold
	}
	if r.interval <= 0 {
		r.interval = DefaultHealthInterval
	}
	return r, nil
}

// LoopNames lists the supervised loops, in the order they were supplied. It is
// what the command prints on start so an operator can see the loop set.
func (r *Runtime) LoopNames() []string {
	out := make([]string, 0, len(r.opts.Loops))
	for _, loop := range r.opts.Loops {
		out = append(out, loop.Name)
	}
	return out
}

// stop is one loop's exit.
type stop struct {
	name string
	err  error
}

// Run starts the loops and supervises them until ctx ends or one of them stops.
//
// The return is the whole contract:
//
//	nil                 every loop was cancelled and drained, journal-safe
//	ErrLoopFailed       one returned for another reason; the rest were stopped
//	the recovery's err   the restart sequence did not complete, so nothing started
//
// Nothing is restarted and nothing is left running: when this returns, every
// goroutine it started has returned too. That is what lets the caller close the
// journal immediately afterwards without racing a loop that is still writing.
func (r *Runtime) Run(ctx context.Context) error {
	if r.opts.Recover != nil {
		// Before any loop. The recovery sequence latches the entry gate in its own
		// constructor and clears it only on completion (internal/reconcile/
		// recovery.go), so a runtime that started its loops first would be
		// reconciling and observing against a state nobody had resolved.
		if err := r.opts.Recover(ctx); err != nil {
			return err
		}
	}

	loopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stops := make(chan stop, len(r.opts.Loops))
	var wg sync.WaitGroup
	for _, loop := range r.opts.Loops {
		wg.Add(1)
		go func(loop SupervisedLoop) {
			defer wg.Done()
			stops <- stop{name: loop.Name, err: loop.Run(loopCtx)}
		}(loop)
	}

	// The auxiliary executors start with the loops and are drained with them, and
	// that is all they share. Nothing here sends to stops, so one that returns
	// cannot become the first stop and cannot bring the engine down. The channel
	// arithmetic above is untouched: stops is buffered to len(Loops) and the
	// senders are still exactly the loops.
	//
	// They go into the same wg on purpose. wg.Wait() below is already this
	// function's journal-close contract — "every goroutine it started has
	// returned too" — and a second WaitGroup would be a second place to forget
	// it rather than a second guarantee.
	for _, aux := range r.opts.Auxiliary {
		wg.Add(1)
		go func(aux AuxiliaryExecutor) {
			defer wg.Done()
			// loopCtx, not ctx: runAuxiliary judges the stop against what it is
			// handed, and the two differ when a supervised loop fails.
			r.runAuxiliary(loopCtx, aux)
		}(aux)
	}

	supervised := make(chan struct{})
	var supervisorWG sync.WaitGroup
	supervisorWG.Add(1)
	go func() {
		defer supervisorWG.Done()
		r.superviseHealth(loopCtx, supervised)
	}()

	// The first stop decides everything. Whether it is a cancellation or not, the
	// rest come down with it: engine-safety's 부분 생존 금지 is not a preference
	// about tidiness, it is that an engine reconciling without observing exits, or
	// observing without detecting fills, is worse than an engine that stopped.
	first := <-stops
	cancel()
	close(supervised)
	wg.Wait()
	supervisorWG.Wait()
	drained := drain(stops)

	if r.gracefulStop(ctx, first.err) {
		r.log(obs.EventEngineHeartbeat, false,
			"loop", first.name,
			obs.FieldDetail, "the runtime was cancelled; every loop was drained and the journal can be closed")
		return nil
	}

	// A loop returned on its own. That is the defensive contract, and it is
	// critical whatever the error says — including a nil error, which means a loop
	// documented as "runs until the context ends" decided otherwise and said
	// nothing about why.
	failure := fmt.Errorf("%w: %s returned %s", ErrLoopFailed, first.name, describe(first.err))
	r.alert(ctx, obs.Event{
		Type:  obs.EventEngineLoopFailed,
		Key:   string(obs.EventEngineLoopFailed) + "|" + first.name,
		Title: "the " + first.name + " loop stopped and the engine was shut down",
		Body: "this loop is documented as returning only when its context ends, so a return is a state " +
			"nobody wrote a recovery for. Every other loop was cancelled and drained and the process is " +
			"exiting non-zero: a partially alive engine keeps trading with one of its senses missing",
		Fields: map[string]any{
			obs.FieldAccount: r.opts.AccountRef,
			"loop":           first.name,
			obs.FieldReason:  describe(first.err),
			"stopped_with":   strings.Join(namesOf(drained), ", "),
		},
	})
	r.log(obs.EventEngineLoopFailed, true, "loop", first.name, obs.FieldReason, describe(first.err))
	return failure
}

// gracefulStop reports whether a loop's return was the runtime being cancelled.
//
// Both halves are required. The caller's context must actually be done — a loop
// that manufactured a context.Canceled out of an inner context is not a graceful
// stop — and the error must be that cancellation, so a loop that returned a real
// failure at the same moment somebody pressed Ctrl-C is still reported.
func (r *Runtime) gracefulStop(ctx context.Context, err error) bool {
	cause := ctx.Err()
	return cause != nil && err != nil && errors.Is(err, cause)
}

// superviseHealth polls the loops' cycle health until the runtime stops.
func (r *Runtime) superviseHealth(ctx context.Context, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		default:
		}
		r.CheckHealth(ctx)
		if err := r.clk.Sleep(ctx, r.interval); err != nil {
			return
		}
	}
}

// CheckHealth applies the degradation threshold once.
//
// It is exported so a test can drive the decision without a clock, and so the
// decision is one function rather than something spread through a ticker: the
// question "has this loop failed five cycles running, and have we already said
// so" is the whole of layer ②.
func (r *Runtime) CheckHealth(ctx context.Context) {
	for _, loop := range r.opts.Loops {
		if loop.Health == nil {
			continue
		}
		consecutive := loop.Health.ConsecutiveFailures()
		if consecutive < r.threshold {
			if consecutive == 0 {
				// Recovered. Unlatching here rather than never is what makes a
				// second outage audible; the operating mode stays where it is,
				// because relaxing it is a person's decision (§0.7).
				r.clearLatch(loop.Name)
			}
			continue
		}
		if !r.takeLatch(loop.Name) {
			continue
		}
		r.escalate(ctx, loop, consecutive)
	}
}

// escalate raises the critical alert and tightens the operating mode.
//
// The loop is not stopped and not restarted: engine-safety says 루프는 계속
// 재시도한다, which is the landed per-loop decision this threshold sits on top of.
// What changes is the account's willingness to take new exposure.
func (r *Runtime) escalate(ctx context.Context, loop SupervisedLoop, consecutive int) {
	r.alert(ctx, obs.Event{
		Type:  obs.EventEngineLoopDegraded,
		Key:   string(obs.EventEngineLoopDegraded) + "|" + loop.Name,
		Title: fmt.Sprintf("the %s loop has failed %d cycles in a row", loop.Name, consecutive),
		Body: "the loop is still running and still retrying, but it has not completed a cycle for " +
			"long enough that the engine no longer knows what the account holds. New entries are " +
			"blocked until the loop recovers and an operator relaxes the mode",
		Fields: map[string]any{
			obs.FieldAccount:       r.opts.AccountRef,
			"loop":                 loop.Name,
			"consecutive_failures": consecutive,
			"threshold":            r.threshold,
			"trigger":              loop.Trigger,
		},
	})

	if r.opts.Escalate == nil || strings.TrimSpace(r.opts.AccountRef) == "" {
		return
	}
	_, changed, err := r.opts.Escalate.EscalateOperatingMode(ctx, r.opts.AccountRef, loop.Trigger, r.opts.Announcer)
	if err != nil {
		r.log(obs.EventOperatingMode, true,
			"loop", loop.Name, obs.FieldReason, err.Error(),
			obs.FieldDetail, "the sustained-failure tightening did not reach the operating mode, so a "+
				"restart would lift the block")
		return
	}
	r.log(obs.EventOperatingMode, false,
		"loop", loop.Name, "trigger", loop.Trigger, "changed", changed,
		obs.FieldDetail, "sustained cycle failure tightened the operating mode to ENTRY_BLOCKED")
}

func (r *Runtime) takeLatch(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.escalated[name] {
		return false
	}
	r.escalated[name] = true
	return true
}

func (r *Runtime) clearLatch(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.escalated, name)
}

func (r *Runtime) alert(ctx context.Context, e obs.Event) {
	if r.opts.Alerts == nil {
		return
	}
	// The alert is delivered on a context of its own: the runtime raises both of
	// these while it is shutting down, and sending them through the cancelled
	// context would drop the one alert that explains the shutdown.
	alertCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), alertDeliveryBound)
	defer cancel()
	if err := r.opts.Alerts.Notify(alertCtx, e); err != nil {
		r.log(obs.EventAlertUndelivered, true, obs.FieldReason, err.Error(), "alert", string(e.Type))
	}
}

// alertDeliveryBound is how long the runtime waits for an alert it raises while
// stopping. Generous enough for obs.Notifier's three bounded publish attempts,
// finite so a dead transport cannot hold the shutdown open.
const alertDeliveryBound = 30 * time.Second

func (r *Runtime) log(t obs.EventType, warn bool, args ...any) {
	if r.opts.Log == nil {
		return
	}
	args = append([]any{obs.FieldAccount, r.opts.AccountRef}, args...)
	if warn {
		r.opts.Log.Warn(t, args...)
		return
	}
	r.opts.Log.Event(t, args...)
}

// drain empties the stop channel of everything already sent, without waiting.
func drain(stops chan stop) []stop {
	var out []stop
	for {
		select {
		case s := <-stops:
			out = append(out, s)
		default:
			return out
		}
	}
}

func namesOf(stops []stop) []string {
	out := make([]string, 0, len(stops))
	for _, s := range stops {
		out = append(out, s.name)
	}
	sort.Strings(out)
	return out
}

// describe renders a loop's return for an operator. A nil error is the case worth
// spelling out: "it returned" and "it returned an error" are different failures
// and the first one is the more surprising.
func describe(err error) string {
	if err == nil {
		return "nil (it simply returned, with no error to explain it)"
	}
	return err.Error()
}
