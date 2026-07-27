package main

// engine_test.go is the boot sequence of `tossctl engine run` (openspec change
// add-engine-runtime, tasks 1.1, 1.2 and 1.3).
//
// # What is faked and why
//
// Two seams: engineAssemble (the engine profile) and engineRuntimeFactory (the
// loop set). Both are package variables, as soakproc.go's process seams are, and
// for the same reason — a test of the *sequence* must not need a broker, a
// journal on an allowlisted filesystem or a fork. What the loops do once they
// start is internal/app/engine's suites' business; what this file asserts is the
// order the refusals happen in, which is the part engine-safety fixes.
//
// The engine.Context values here are hand-built and carry only the automation
// status, which is the one field the sequence branches on. A test that needed a
// real one would need a verified gate, and interlock clause 6 makes that
// unreachable outside internal/app/engine's own test-only seam — which is the
// safety property working, not a gap in the coverage.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// --- seams ---------------------------------------------------------------------

// stubAssembly replaces the engine assembly for one test and reports whether it
// was reached.
func stubAssembly(t *testing.T, ectx *engine.Context, err error) func() bool {
	t.Helper()
	previous := engineAssemble
	called := false
	engineAssemble = func(context.Context, *rootOptions, clock.Clock, *obs.Logger) (*engine.Context, error) {
		called = true
		return ectx, err
	}
	t.Cleanup(func() { engineAssemble = previous })
	return func() bool { return called }
}

// stubRuntime replaces the loop assembly. The default refuses, so a test that
// reaches it fails with a recognisable error instead of starting three loops
// against a hand-built Context.
func stubRuntime(t *testing.T, build func(*engine.Context) (*engine.Runtime, error)) {
	t.Helper()
	previous := engineRuntimeFactory
	engineRuntimeFactory = func(ectx *engine.Context, _ clock.Clock, _ *obs.Logger) (*engine.Runtime, error) {
		return build(ectx)
	}
	t.Cleanup(func() { engineRuntimeFactory = previous })
}

var errStubRuntimeReached = errors.New("the loop assembly was reached")

// gateOffEngine and verifiedEngine are the two automation states the sequence
// branches on.
func gateOffEngine() *engine.Context {
	return &engine.Context{Automation: engine.AutomationStatus{Enabled: false}}
}

func verifiedEngine() *engine.Context {
	return &engine.Context{Automation: engine.AutomationStatus{
		Enabled: true, Verified: true, AccountRef: "123-45",
	}}
}

// --- the exclusion is the first thing that happens -------------------------------

// TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled is review round 3's
// TOCTOU fix, asserted from the only angle that proves it: with the lock already
// held, the assembly must never run — so the journal open and its migration are
// inside the exclusion rather than racing it.
func TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled(t *testing.T) {
	dir := testenv.Isolate(t)
	held, err := enginelock.Acquire(dir)
	if err != nil {
		t.Skipf("this platform has no file lock: %v", err)
	}
	t.Cleanup(held.Release)

	assembled := stubAssembly(t, gateOffEngine(), nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	_, _, err = runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, enginelock.ErrAlreadyRunning) {
		t.Fatalf("a second instance returned %v, want ErrAlreadyRunning", err)
	}
	if assembled() {
		t.Error("the engine was assembled before the lock was taken; the journal open is outside the exclusion")
	}
}

// TestTheLockIsReleasedWhenTheCommandReturns, so the console's [엔진 시작] after
// a refused start is not itself refused.
func TestTheLockIsReleasedWhenTheCommandReturns(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, gateOffEngine(), nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errEngineGateOff) {
		t.Fatalf("first run: %v", err)
	}
	lock, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("the lock was not released: %v", err)
	}
	lock.Release()
}

// --- the gate ---------------------------------------------------------------------

// TestAGateOffEngineRefusesWithoutEnumeratingClauses is the scenario "게이트 OFF
// 기동". The two refusals have different grounds and must not be conflated: with
// the gate off there is no interlock, so a list of unmet clauses would be
// inventing one.
func TestAGateOffEngineRefusesWithoutEnumeratingClauses(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, gateOffEngine(), nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errEngineGateOff) {
		t.Fatalf("err = %v, want errEngineGateOff", err)
	}
	if !strings.Contains(err.Error(), "기동할 루프 집합이 없다") {
		t.Errorf("the refusal does not say why: %v", err)
	}
	if strings.Contains(errOut, "인터록") {
		t.Errorf("the gate-off refusal enumerated interlock clauses:\n%s", errOut)
	}
	// And nothing was left claiming an engine is alive.
	if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
		t.Error("a refused start left an active marker behind")
	}
}

// TestAnUnmetInterlockIsEnumerated is the scenario "게이트 ON + 인터록 미충족
// 기동": no loop starts, and the operator is shown the clauses.
func TestAnUnmetInterlockIsEnumerated(t *testing.T) {
	dir := testenv.Isolate(t)
	refusal := fmt.Errorf("%w: %w", engine.ErrAutomationGateRefused, engine.ErrProtectionNotWired)
	stubAssembly(t, nil, refusal)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errEngineInterlockUnmet) {
		t.Fatalf("err = %v, want errEngineInterlockUnmet", err)
	}
	if !strings.Contains(errOut, "ProtectionReady") {
		t.Errorf("the refusal does not enumerate the unmet clause:\n%s", errOut)
	}
	if !strings.Contains(errOut, "  - ") {
		t.Errorf("the clauses are not enumerated as a list:\n%s", errOut)
	}
}

// TestANonInterlockFailureIsReportedAsItself. A journal that will not open is not
// a gate problem, and dressing it as one sends the operator to the wrong file.
func TestANonInterlockFailureIsReportedAsItself(t *testing.T) {
	dir := testenv.Isolate(t)
	boom := errors.New("engine: opening the journal: filesystem is not in the allowlist")
	stubAssembly(t, nil, boom)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the underlying failure", err)
	}
}

// --- the other process on this account ---------------------------------------------

// TestAFreshVerifyRunLockRefusesTheStart is the scenario "검증 실행 중 기동 시도".
func TestAFreshVerifyRunLockRefusesTheStart(t *testing.T) {
	dir := testenv.Isolate(t)
	lock, err := runlock.Acquire(filepath.Join(dir, runlock.FileName), time.Now())
	if err != nil {
		t.Fatalf("runlock.Acquire: %v", err)
	}
	t.Cleanup(lock.Release)

	stubAssembly(t, verifiedEngine(), nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errVerifyInProgress) {
		t.Fatalf("err = %v, want errVerifyInProgress", err)
	}
	if !strings.Contains(errOut, "검증이 끝난 뒤 다시 시작하라") {
		t.Errorf("the refusal does not say what to do next:\n%s", errOut)
	}
}

// TestAStaleVerifyRunLockDoesNotRefuse. A crashed verification costs one refused
// start at most, not a wedged engine — the same direction internal/runlock takes
// for the soak.
func TestAStaleVerifyRunLockDoesNotRefuse(t *testing.T) {
	dir := testenv.Isolate(t)
	lock, err := runlock.Acquire(filepath.Join(dir, runlock.FileName),
		time.Now().Add(-2*runlock.StaleAfter))
	if err != nil {
		t.Fatalf("runlock.Acquire: %v", err)
	}
	t.Cleanup(lock.Release)

	stubAssembly(t, verifiedEngine(), nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errStubRuntimeReached) {
		t.Fatalf("err = %v; a stale verification marker must not refuse the start", err)
	}
}

// --- the advisory marker --------------------------------------------------------

// TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter. It is what the console's
// status line and the autostart script's precheck read, and it must be honest in
// both directions.
func TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, verifiedEngine(), nil)

	var duringRun enginelock.Status
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) {
		duringRun = enginelock.Read(enginelock.MarkerPath(dir), time.Now())
		return nil, errStubRuntimeReached
	})

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errStubRuntimeReached) {
		t.Fatalf("err = %v", err)
	}
	if !duringRun.Running {
		t.Error("no active marker was held while the loops were being assembled")
	}
	if duringRun.Marker.PID != os.Getpid() {
		t.Errorf("the marker names pid %d, want this process", duringRun.Marker.PID)
	}
	if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
		t.Error("the marker survived the command returning")
	}
}

// TestTheMarkerSitsBesideTheJournalAndFollowsConfigDir. An isolated profile must
// not report itself as the real engine on the operator's dashboard.
func TestTheMarkerSitsBesideTheJournalAndFollowsConfigDir(t *testing.T) {
	dir := t.TempDir()
	got, err := engineJournalDir(&rootOptions{configDir: dir})
	if err != nil {
		t.Fatalf("engineJournalDir: %v", err)
	}
	if got != dir {
		t.Errorf("journal dir = %s, want the --config-dir override %s", got, dir)
	}
}

// --- the stop discipline ------------------------------------------------------------

// TestTheSecondStopSignalExitsAtOnce. The first cancels and waits; the second is
// an operator saying the wait is over, and the journal's own restart recovery is
// designed for the state a hard exit leaves.
func TestTheSecondStopSignalExitsAtOnce(t *testing.T) {
	var delivered chan<- os.Signal
	previousNotify, previousStop := engineSignalNotify, engineSignalStop
	engineSignalNotify = func(c chan<- os.Signal, _ ...os.Signal) { delivered = c }
	engineSignalStop = func(chan<- os.Signal) {}
	t.Cleanup(func() { engineSignalNotify, engineSignalStop = previousNotify, previousStop })

	exited := make(chan int, 1)
	previousExit := engineHardExit
	engineHardExit = func(code int) { exited <- code }
	t.Cleanup(func() { engineHardExit = previousExit })

	ctx, cancel := context.WithCancel(context.Background())
	var out strings.Builder
	stop := watchStopSignals(cancel, &out)
	defer stop()

	if delivered == nil {
		t.Fatal("watchStopSignals subscribed to nothing")
	}
	delivered <- os.Interrupt
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("the first signal did not cancel the runtime")
	}
	select {
	case code := <-exited:
		t.Fatalf("the first signal exited the process with %d; it must wait for the loops", code)
	case <-time.After(50 * time.Millisecond):
	}

	delivered <- os.Interrupt
	select {
	case code := <-exited:
		if code != engineHardExitCode {
			t.Errorf("hard exit code = %d, want %d", code, engineHardExitCode)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the second signal did not exit at once")
	}
}

// TestBothStopSignalsAreGracefulOnFirstDelivery. SIGINT is a terminal and SIGTERM
// is the console button and any supervisor; treating one of them as a kill would
// make [엔진 정지] the thing that leaves the journal half-written.
func TestBothStopSignalsAreGracefulOnFirstDelivery(t *testing.T) {
	want := map[os.Signal]bool{os.Interrupt: false, syscall.SIGTERM: false}
	for _, sig := range engineStopSignals() {
		if _, ok := want[sig]; !ok {
			t.Errorf("unexpected stop signal %v", sig)
			continue
		}
		want[sig] = true
	}
	for sig, seen := range want {
		if !seen {
			t.Errorf("%v is not in the stop set", sig)
		}
	}
}

// --- the loop set --------------------------------------------------------------------

// TestTheLoopSetIsTheSpecifiedThree reads the assembly's source, because what is
// being asserted is what the wiring *names* — a loop that is constructed and not
// supervised, or supervised under the wrong trigger, is exactly the mistake a
// runtime test would only find on the day it mattered.
func TestTheLoopSetIsTheSpecifiedThree(t *testing.T) {
	src := readSource(t, "engine.go")
	for _, want := range []string{
		`Name:    "reconcile"`,
		`Name: "exit"`,
		`Name:    "filldetect"`,
		"journal.ModeTriggerReconcileCycleFailure",
		"journal.ModeTriggerFillDetectionFailure",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("the loop set does not carry %q; engine-safety fixes it at three loops "+
				"(대사·exit 관측·체결 감지)", want)
		}
	}
}

// TestTheExitObserverDefersToFillDetection. The SLO seam existed with nothing to
// answer it because no build constructed a detector; this is the wiring that makes
// exit-policy's 체결 감지 SLO에 양보 reachable.
func TestTheExitObserverDefersToFillDetection(t *testing.T) {
	src := readSource(t, "engine.go")
	if !strings.Contains(src, "SLO:       detectorPressure{detector: detector}") {
		t.Error("the exit observer is built with no SLO source; the deference rule is unreachable again")
	}
	if !strings.Contains(src, "p.detector.Health().EntryBlocked") {
		t.Error("the deference verdict is not the detector's own; a second definition of 'behind' would drift")
	}
}

// TestTheRestartRecoveryRunsBeforeTheLoops (task 1.4, the wiring half). The
// sequence itself is internal/reconcile's and is tested there; what this asserts
// is that the runtime is handed it rather than starting loops over an unresolved
// restart.
func TestTheRestartRecoveryRunsBeforeTheLoops(t *testing.T) {
	src := readSource(t, "engine.go")
	if !strings.Contains(src, "ectx.Recovery(") {
		t.Fatal("the runtime is assembled without the restart recovery")
	}
	if !strings.Contains(src, "Recover: func(ctx context.Context) error {") {
		t.Error("the recovery is not handed to the runtime's Recover hook, so nothing runs it first")
	}
}

// TestTheHintConsumerIsValidatedAtAssembly is the SHALL from engine-safety: an
// unrouted Refresh is refused when the detector is built, not discovered when the
// hint loop returns and the defensive termination contract reads it as an
// incident.
func TestTheHintConsumerIsValidatedAtAssembly(t *testing.T) {
	ectx := verifiedEngine()
	if _, err := engineFillDetector(ectx, clock.System(), &filldetect.Hints{}); err == nil {
		t.Fatal("a hint consumer with no refresh function was accepted at assembly time")
	}
	// And a build with no hints at all is fine: the engine holds no WTS session.
	if _, err := engineFillDetector(ectx, clock.System(), nil); err != nil {
		t.Fatalf("the detector refused to be built without hints: %v", err)
	}
}
