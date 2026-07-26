package engine_test

// runtime_test.go covers the engine runtime's two supervision layers
// (openspec change add-engine-runtime, task 1.3).
//
// The loops here are functions, not the real ones: what is under test is the
// supervisor's classification of a return and its threshold on a live loop's
// cycle health, and driving that through a real reconciliation would be testing
// the reconciliation. The real loops' own suites are next door.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const runtimeAccount = "acct-runtime"

// --- fakes --------------------------------------------------------------------

// recordingAlerts is a mutex-guarded alert sink. The runtime raises alerts from
// its supervisor goroutine and from Run, so an unguarded slice would be a data
// race the -race build finds before any assertion does.
type recordingAlerts struct {
	mu     sync.Mutex
	events []obs.Event
}

func (a *recordingAlerts) Notify(_ context.Context, e obs.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, e)
	return nil
}

func (a *recordingAlerts) count(t obs.EventType) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	n := 0
	for _, e := range a.events {
		if e.Type == t {
			n++
		}
	}
	return n
}

func (a *recordingAlerts) types() []obs.EventType {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]obs.EventType, 0, len(a.events))
	for _, e := range a.events {
		out = append(out, e.Type)
	}
	return out
}

// recordingEscalator stands in for the journal's operating-mode transition.
type recordingEscalator struct {
	mu       sync.Mutex
	triggers []string
	err      error
}

func (e *recordingEscalator) EscalateOperatingMode(_ context.Context, _, trigger string,
	_ journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.triggers = append(e.triggers, trigger)
	if e.err != nil {
		return journal.OperatingModeRecord{}, false, e.err
	}
	return journal.OperatingModeRecord{Mode: journal.ModeEntryBlocked, Cause: trigger}, true, nil
}

func (e *recordingEscalator) seen() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string(nil), e.triggers...)
}

// countingHealth is a settable LoopHealth.
type countingHealth struct {
	mu sync.Mutex
	n  int
}

func (h *countingHealth) ConsecutiveFailures() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.n
}

func (h *countingHealth) set(n int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.n = n
}

// blockingLoop runs until its context ends and returns the context's error, which
// is what all three real loops do.
func blockingLoop(started chan<- struct{}) func(context.Context) error {
	var once sync.Once
	return func(ctx context.Context) error {
		once.Do(func() {
			if started != nil {
				close(started)
			}
		})
		<-ctx.Done()
		return ctx.Err()
	}
}

// --- layer ①: the defensive termination contract ---------------------------------

// TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical is the SHALL NOT:
// 컨텍스트 취소에 의한 반환은 정상 종료이며 critical을 발송하지 않는다.
func TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical(t *testing.T) {
	alerts := &recordingAlerts{}
	var stopped sync.WaitGroup
	stopped.Add(2)

	loop := func(ctx context.Context) error {
		defer stopped.Done()
		<-ctx.Done()
		return ctx.Err()
	}
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: loop},
			{Name: "exit", Run: loop},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run returned %v; a cancellation is a normal stop", err)
	}
	stopped.Wait() // every loop returned before Run did
	if n := alerts.count(obs.EventEngineLoopFailed); n != 0 {
		t.Errorf("a graceful stop raised %d loop-failed alert(s): %v", n, alerts.types())
	}
	if len(alerts.types()) != 0 {
		t.Errorf("a graceful stop raised %v; SIGTERM is not an incident", alerts.types())
	}
}

// TestALoopReturningOnItsOwnStopsEverythingAndIsCritical is the scenario "루프의
// 비정상 반환": the rest are stopped, a critical alert goes out, and the process
// is told to fail.
func TestALoopReturningOnItsOwnStopsEverythingAndIsCritical(t *testing.T) {
	alerts := &recordingAlerts{}
	boom := errors.New("the reconciliation driver hit a state nobody wrote")
	survivorStopped := make(chan struct{})

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: func(context.Context) error { return boom }},
			{Name: "exit", Run: func(ctx context.Context) error {
				<-ctx.Done()
				close(survivorStopped)
				return ctx.Err()
			}},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	runErr := rt.Run(context.Background())
	if !errors.Is(runErr, engine.ErrLoopFailed) {
		t.Fatalf("Run returned %v, want ErrLoopFailed", runErr)
	}
	if !strings.Contains(runErr.Error(), "reconcile") {
		t.Errorf("the failure %q does not name the loop that stopped", runErr)
	}
	select {
	case <-survivorStopped:
	default:
		t.Error("the surviving loop was not stopped; 부분 생존은 금지된다")
	}
	if n := alerts.count(obs.EventEngineLoopFailed); n != 1 {
		t.Errorf("loop-failed alerts = %d, want exactly one: %v", n, alerts.types())
	}
	if obs.SeverityOf(obs.EventEngineLoopFailed) != obs.SeverityCritical {
		t.Error("engine.loop_failed is not graded critical")
	}
}

// TestALoopThatSimplyReturnsIsAlsoAFailure. A nil return from a loop documented
// as "runs until the context ends" is the more surprising of the two failures,
// and a supervisor that only reacted to errors would treat it as success.
func TestALoopThatSimplyReturnsIsAlsoAFailure(t *testing.T) {
	alerts := &recordingAlerts{}
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops: []engine.SupervisedLoop{
			{Name: "filldetect", Run: func(context.Context) error { return nil }},
			{Name: "exit", Run: blockingLoop(nil)},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	if err := rt.Run(context.Background()); !errors.Is(err, engine.ErrLoopFailed) {
		t.Fatalf("Run returned %v, want ErrLoopFailed", err)
	}
	if n := alerts.count(obs.EventEngineLoopFailed); n != 1 {
		t.Errorf("loop-failed alerts = %d, want one", n)
	}
}

// TestALoopFailingDuringAShutdownIsStillReported: the two halves of the graceful
// test are both required. A loop that returns a genuine failure at the same
// moment the operator cancels is a failure, not a clean stop.
func TestALoopFailingDuringAShutdownIsStillReported(t *testing.T) {
	alerts := &recordingAlerts{}
	release := make(chan struct{})
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: func(context.Context) error {
				<-release
				return errors.New("the journal went away")
			}},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	close(release)

	if err := <-done; !errors.Is(err, engine.ErrLoopFailed) {
		t.Fatalf("Run returned %v; a real failure during a shutdown is still a failure", err)
	}
	if n := alerts.count(obs.EventEngineLoopFailed); n != 1 {
		t.Errorf("loop-failed alerts = %d, want one", n)
	}
}

// --- layer ②: the sustained degradation threshold --------------------------------

// TestFiveConsecutiveFailuresEscalateOnceAndTheLoopKeepsRunning is the scenario
// "reconcile driver 지속 실패": critical + ENTRY_BLOCKED, and the loop is not
// stopped — engine-safety says 루프는 재시도를 계속한다.
func TestFiveConsecutiveFailuresEscalateOnceAndTheLoopKeepsRunning(t *testing.T) {
	alerts := &recordingAlerts{}
	escalator := &recordingEscalator{}
	health := &countingHealth{}
	stopped := make(chan struct{})

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Escalate:   escalator,
		Loops: []engine.SupervisedLoop{{
			Name:    "reconcile",
			Run:     func(ctx context.Context) error { <-ctx.Done(); close(stopped); return ctx.Err() },
			Health:  health,
			Trigger: journal.ModeTriggerReconcileCycleFailure,
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx := context.Background()

	// Four failures is not an incident: the threshold is what separates a bad
	// minute from an engine that cannot see the account.
	health.set(engine.DefaultDegradationThreshold - 1)
	rt.CheckHealth(ctx)
	if got := escalator.seen(); len(got) != 0 {
		t.Fatalf("four consecutive failures escalated: %v", got)
	}

	health.set(engine.DefaultDegradationThreshold)
	rt.CheckHealth(ctx)
	if got := escalator.seen(); len(got) != 1 || got[0] != journal.ModeTriggerReconcileCycleFailure {
		t.Fatalf("escalations = %v, want one RECONCILE_CYCLE_FAILURE", got)
	}
	if n := alerts.count(obs.EventEngineLoopDegraded); n != 1 {
		t.Fatalf("degraded alerts = %d, want one", n)
	}

	// Still failing, ten cycles later: one condition, one alert, one transition.
	health.set(engine.DefaultDegradationThreshold * 2)
	rt.CheckHealth(ctx)
	rt.CheckHealth(ctx)
	if got := escalator.seen(); len(got) != 1 {
		t.Fatalf("a sustained outage escalated %d times: %v", len(got), got)
	}
	if n := alerts.count(obs.EventEngineLoopDegraded); n != 1 {
		t.Fatalf("degraded alerts = %d; one condition is one alert", n)
	}

	// The loop was never cancelled by the threshold.
	select {
	case <-stopped:
		t.Fatal("the degradation threshold stopped the loop; it must keep retrying")
	default:
	}
}

// TestARecoveredLoopCanRaiseTheAlarmAgain. The latch is per outage, not per
// process: a loop that recovers and then fails again is a second incident.
//
// The operating mode is NOT relaxed by the recovery — that is §0.7's人 decision —
// so a second escalation is a no-op transition and the point of it is the alert.
func TestARecoveredLoopCanRaiseTheAlarmAgain(t *testing.T) {
	alerts := &recordingAlerts{}
	escalator := &recordingEscalator{}
	health := &countingHealth{}

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Escalate:   escalator,
		Loops: []engine.SupervisedLoop{{
			Name:    "filldetect",
			Run:     blockingLoop(nil),
			Health:  health,
			Trigger: journal.ModeTriggerFillDetectionFailure,
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx := context.Background()

	health.set(engine.DefaultDegradationThreshold)
	rt.CheckHealth(ctx)
	health.set(0)
	rt.CheckHealth(ctx)
	health.set(engine.DefaultDegradationThreshold)
	rt.CheckHealth(ctx)

	if got := escalator.seen(); len(got) != 2 {
		t.Fatalf("escalations = %v, want two — a recovered loop that fails again is a new incident", got)
	}
	if n := alerts.count(obs.EventEngineLoopDegraded); n != 2 {
		t.Fatalf("degraded alerts = %d, want two", n)
	}
}

// TestTheDegradationAlertGoesOutEvenWhenTheModeTransitionFails. The alert and the
// transition answer different questions and the alert must not depend on the
// journal being writable.
func TestTheDegradationAlertGoesOutEvenWhenTheModeTransitionFails(t *testing.T) {
	alerts := &recordingAlerts{}
	escalator := &recordingEscalator{err: errors.New("the journal is read-only")}
	health := &countingHealth{}

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Escalate:   escalator,
		Loops: []engine.SupervisedLoop{{
			Name: "reconcile", Run: blockingLoop(nil), Health: health,
			Trigger: journal.ModeTriggerReconcileCycleFailure,
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	health.set(engine.DefaultDegradationThreshold)
	rt.CheckHealth(context.Background())

	if n := alerts.count(obs.EventEngineLoopDegraded); n != 1 {
		t.Errorf("degraded alerts = %d, want one even though the transition failed", n)
	}
}

// TestTheSupervisorPollsHealthWhileTheLoopsRun proves the threshold is applied by
// the running runtime and not only by a test calling CheckHealth by hand.
func TestTheSupervisorPollsHealthWhileTheLoopsRun(t *testing.T) {
	alerts := &recordingAlerts{}
	escalator := &recordingEscalator{}
	health := &countingHealth{}
	health.set(engine.DefaultDegradationThreshold)
	clk := clock.NewFake(time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC))

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef:     runtimeAccount,
		Alerts:         alerts,
		Escalate:       escalator,
		Clock:          clk,
		HealthInterval: time.Second,
		Loops: []engine.SupervisedLoop{{
			Name: "reconcile", Run: blockingLoop(nil), Health: health,
			Trigger: journal.ModeTriggerReconcileCycleFailure,
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	// The supervisor's first pass runs before it sleeps, so waiting for it to be
	// asleep is waiting for the pass to have happened.
	if !clk.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the supervisor never entered its health-poll wait")
	}
	if got := escalator.seen(); len(got) != 1 {
		t.Fatalf("the running supervisor escalated %v, want one", got)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// --- assembly-time refusals -------------------------------------------------------

// TestTheRuntimeRefusesWiringItCannotSupervise. Every one of these would
// otherwise be found by a supervisor at the moment it was needed, which for the
// trigger means "we alerted and changed nothing".
func TestTheRuntimeRefusesWiringItCannotSupervise(t *testing.T) {
	health := &countingHealth{}
	cases := map[string]engine.RuntimeOptions{
		"no loops": {AccountRef: runtimeAccount},
		"a loop with no name": {AccountRef: runtimeAccount,
			Loops: []engine.SupervisedLoop{{Run: blockingLoop(nil)}}},
		"a loop with no Run": {AccountRef: runtimeAccount,
			Loops: []engine.SupervisedLoop{{Name: "reconcile"}}},
		"two loops with one name": {AccountRef: runtimeAccount, Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: blockingLoop(nil)},
			{Name: "reconcile", Run: blockingLoop(nil)},
		}},
		"a health source with no trigger": {AccountRef: runtimeAccount, Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: blockingLoop(nil), Health: health},
		}},
		"a trigger the journal does not enumerate": {AccountRef: runtimeAccount, Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: blockingLoop(nil), Health: health, Trigger: "RECONCILE_IS_SAD"},
		}},
		"an escalating loop with no account": {Loops: []engine.SupervisedLoop{
			{Name: "reconcile", Run: blockingLoop(nil), Health: health,
				Trigger: journal.ModeTriggerReconcileCycleFailure},
		}},
	}
	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := engine.NewRuntime(opts); !errors.Is(err, engine.ErrRuntimeUnavailable) {
				t.Fatalf("NewRuntime returned %v, want ErrRuntimeUnavailable", err)
			}
		})
	}
}

// --- restart recovery ---------------------------------------------------------------

// TestRecoveryRunsBeforeAnyLoopStarts is task 1.4's ordering half: the restart
// sequence latches the entry gate in its own constructor and clears it on
// completion, so a loop that started first would be trading over an unresolved
// attempt.
func TestRecoveryRunsBeforeAnyLoopStarts(t *testing.T) {
	var order []string
	var mu sync.Mutex
	note := func(s string) { mu.Lock(); order = append(order, s); mu.Unlock() }

	started := make(chan struct{})
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Recover: func(context.Context) error {
			note("recover")
			return nil
		},
		Loops: []engine.SupervisedLoop{{Name: "reconcile", Run: func(ctx context.Context) error {
			note("loop")
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	<-started
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != "recover" || order[1] != "loop" {
		t.Fatalf("order = %v, want recovery before the loops", order)
	}
}

// TestAnIncompleteRecoveryStartsNothing: fail-closed. The recovery's own error is
// returned verbatim so the operator sees what did not settle.
func TestAnIncompleteRecoveryStartsNothing(t *testing.T) {
	incomplete := errors.New("reconcile: restart recovery did not complete")
	var startedLoop bool

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Recover:    func(context.Context) error { return incomplete },
		Loops: []engine.SupervisedLoop{{Name: "reconcile", Run: func(ctx context.Context) error {
			startedLoop = true
			<-ctx.Done()
			return ctx.Err()
		}}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	if err := rt.Run(context.Background()); !errors.Is(err, incomplete) {
		t.Fatalf("Run returned %v, want the recovery's own error", err)
	}
	if startedLoop {
		t.Error("a loop started over an unresolved restart")
	}
}
