package main

// engineproc_test.go covers the engine's process control and the drift between
// this binary and the autostart script (openspec change add-engine-runtime,
// tasks 2.1 and 2.2).
//
// Nothing here forks, signals or greps: the four seams in engineproc.go are
// package variables for exactly that reason, which is soakproc_test.go's
// arrangement and its rationale.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

// engineFakes replaces the process seams for one test.
type engineFakes struct {
	found     []int
	findErr   error
	signalled []int
	signalErr error
	alive     map[int]bool
	spawned   [][]string
	spawnErr  error
	// exit, when non-nil, is delivered on the spawned engine's wait channel — an
	// engine that refused and exited before the probe expired.
	exit    error
	exitNow bool
}

func (f *engineFakes) install(t *testing.T) {
	t.Helper()
	prevFind, prevSignal, prevAlive, prevSpawn, prevSleep :=
		engineFindProcesses, engineSignalProcess, engineProcessAlive, engineSpawnDetached, engineSleep

	engineFindProcesses = func() ([]int, error) { return f.found, f.findErr }
	engineSignalProcess = func(pid int) error {
		if f.signalErr != nil {
			return f.signalErr
		}
		f.signalled = append(f.signalled, pid)
		if f.alive != nil {
			f.alive[pid] = false
		}
		return nil
	}
	engineProcessAlive = func(pid int) bool { return f.alive != nil && f.alive[pid] }
	engineSpawnDetached = func(_, logPath string, args []string) (<-chan error, error) {
		f.spawned = append(f.spawned, args)
		if f.spawnErr != nil {
			return nil, f.spawnErr
		}
		wait := make(chan error, 1)
		if f.exitNow {
			// Write what a refusing engine would have written before exiting, so
			// the tail the start reports is a real read of a real file.
			_ = os.WriteFile(logPath, []byte(
				"기동 인터록 미충족: 루프를 하나도 시작하지 않았다\n"+
					"  - 9. 브로커측 보호 실행(ProtectionReady) — 이 빌드에는 없다 [2c 소관]\n"), 0o600)
			wait <- f.exit
		}
		return wait, nil
	}
	engineSleep = func(time.Duration) {}

	prevProbe := engineStartProbe
	engineStartProbe = 10 * time.Millisecond

	t.Cleanup(func() {
		engineFindProcesses, engineSignalProcess, engineProcessAlive, engineSpawnDetached, engineSleep =
			prevFind, prevSignal, prevAlive, prevSpawn, prevSleep
		engineStartProbe = prevProbe
	})
}

// --- starting -----------------------------------------------------------------

// TestStartingSpawnsTheEngineWithThisProfilesConfigDir. An isolated profile must
// stay isolated across the button, or the console of a test run could start the
// real engine.
func TestStartingSpawnsTheEngineWithThisProfilesConfigDir(t *testing.T) {
	dir := t.TempDir()
	f := &engineFakes{}
	f.install(t)

	note, err := startEngine(&rootOptions{configDir: dir})
	if err != nil {
		t.Fatalf("startEngine: %v", err)
	}
	if len(f.spawned) != 1 {
		t.Fatalf("spawned %d engines, want one", len(f.spawned))
	}
	args := strings.Join(f.spawned[0], " ")
	if !strings.Contains(args, "--config-dir "+dir) {
		t.Errorf("the spawned engine did not inherit the profile: %s", args)
	}
	if !strings.Contains(args, "engine run") {
		t.Errorf("the spawned command is not the engine: %s", args)
	}
	if !strings.Contains(note, engineLogPath(dir)) {
		t.Errorf("the answer does not say where the log is: %q", note)
	}
}

// TestARefusedStartReportsTheEnginesOwnLog is what makes the console unable to
// paper over an unmet interlock: the refusal the engine printed is what comes
// back.
func TestARefusedStartReportsTheEnginesOwnLog(t *testing.T) {
	dir := t.TempDir()
	f := &engineFakes{exitNow: true, exit: errors.New("exit status 1")}
	f.install(t)

	note, err := startEngine(&rootOptions{configDir: dir})
	if err == nil {
		t.Fatal("an engine that exited immediately was reported as started")
	}
	if !strings.Contains(note, "ProtectionReady") {
		t.Errorf("the engine's own refusal did not come back:\n%s", note)
	}
	if !strings.Contains(err.Error(), "곧바로 종료했다") {
		t.Errorf("the error does not say what happened: %v", err)
	}
}

// TestStartingIsRefusedWhileTheMarkerIsFresh. Advisory, and only so the answer is
// "already running" rather than a spawned process that immediately loses the race
// for the flock.
func TestStartingIsRefusedWhileTheMarkerIsFresh(t *testing.T) {
	dir := t.TempDir()
	release, err := enginelock.Hold(context.Background(), enginelock.MarkerPath(dir), time.Now())
	if err != nil {
		t.Fatalf("enginelock.Hold: %v", err)
	}
	t.Cleanup(release)

	f := &engineFakes{}
	f.install(t)

	if _, err := startEngine(&rootOptions{configDir: dir}); err == nil {
		t.Fatal("a second engine was started while the marker was fresh")
	}
	if len(f.spawned) != 0 {
		t.Errorf("a process was spawned anyway: %v", f.spawned)
	}
}

// TestAStaleMarkerDoesNotBlockAStart. A crashed engine costs one refused start at
// most; the exclusion is the flock and it is gone with the process.
func TestAStaleMarkerDoesNotBlockAStart(t *testing.T) {
	dir := t.TempDir()
	release, err := enginelock.Hold(context.Background(), enginelock.MarkerPath(dir),
		time.Now().Add(-2*enginelock.StaleAfter))
	if err != nil {
		t.Fatalf("enginelock.Hold: %v", err)
	}
	t.Cleanup(release)

	f := &engineFakes{}
	f.install(t)

	if _, err := startEngine(&rootOptions{configDir: dir}); err != nil {
		t.Fatalf("a stale marker blocked a start: %v", err)
	}
	if len(f.spawned) != 1 {
		t.Errorf("spawned %d engines, want one", len(f.spawned))
	}
}

// TestStartingIsRefusedWhenAProcessIsAlreadyThere, even with no marker — a
// machine where the marker could not be written still must not get two engines
// through this path.
func TestStartingIsRefusedWhenAProcessIsAlreadyThere(t *testing.T) {
	dir := t.TempDir()
	f := &engineFakes{found: []int{4242}}
	f.install(t)

	if _, err := startEngine(&rootOptions{configDir: dir}); err == nil {
		t.Fatal("a second engine was started while pgrep found one")
	}
	if len(f.spawned) != 0 {
		t.Errorf("a process was spawned anyway: %v", f.spawned)
	}
}

// --- stopping -----------------------------------------------------------------

// TestStoppingSignalsAndWaits is the signal discipline: ask, then wait for the
// journal to be closed. Nothing is killed.
func TestStoppingSignalsAndWaits(t *testing.T) {
	dir := t.TempDir()
	f := &engineFakes{found: []int{4242}, alive: map[int]bool{4242: true}}
	f.install(t)

	note, err := stopEngine(&rootOptions{configDir: dir})
	if err != nil {
		t.Fatalf("stopEngine: %v", err)
	}
	if len(f.signalled) != 1 || f.signalled[0] != 4242 {
		t.Fatalf("signalled %v, want [4242]", f.signalled)
	}
	if !strings.Contains(note, "4242") {
		t.Errorf("the answer does not name what was stopped: %q", note)
	}
}

// TestStoppingNeverSignalsThisProcess. pgrep should not match the console, but a
// guess is not a licence.
func TestStoppingNeverSignalsThisProcess(t *testing.T) {
	f := &engineFakes{found: []int{os.Getpid()}}
	f.install(t)

	if _, err := stopEngine(&rootOptions{configDir: t.TempDir()}); err != nil {
		t.Fatalf("stopEngine: %v", err)
	}
	if len(f.signalled) != 0 {
		t.Fatalf("the stop signalled %v, which includes this process", f.signalled)
	}
}

// TestAnEngineThatWillNotGoIsReportedRatherThanKilled. The wait is bounded and
// then it gives up — it never escalates to a kill, because a killed engine is a
// journal that was not closed and the operator was not told.
func TestAnEngineThatWillNotGoIsReportedRatherThanKilled(t *testing.T) {
	f := &engineFakes{alive: map[int]bool{4242: true}}
	f.install(t)

	err := waitForEngineExit(4242, 10*time.Millisecond)
	if err == nil {
		t.Fatal("an engine that never exited was reported as stopped")
	}
	if !strings.Contains(err.Error(), "재기동 시 복구 절차") {
		t.Errorf("the error does not say what happens next: %v", err)
	}
	if f.alive[4242] {
		// Still alive, and nothing killed it — which is the assertion.
		return
	}
	t.Error("the wait ended by killing the process")
}

// TestNoEngineToStopIsNotAFailure.
func TestNoEngineToStopIsNotAFailure(t *testing.T) {
	f := &engineFakes{}
	f.install(t)
	note, err := stopEngine(&rootOptions{configDir: t.TempDir()})
	if err != nil {
		t.Fatalf("stopEngine: %v", err)
	}
	if !strings.Contains(note, "찾지 못했다") {
		t.Errorf("the answer does not say nothing was running: %q", note)
	}
}

// --- the script and the constants are two halves of one mechanism ----------------

// autostartScript reads the prepared script out of the repository.
func autostartScript(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "tools", "engine-autostart.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

// TestTheEngineProcessPatternMatchesTheAutostartScript. A process one of them can see
// and the other cannot is a bug that only shows up at three in the morning —
// soakproc_test.go's assertion, aimed at the engine.
func TestTheEngineProcessPatternMatchesTheAutostartScript(t *testing.T) {
	if engineProcessPattern != "tossctl engine run" {
		t.Errorf("the pattern is %q; the script greps for \"tossctl engine run\"", engineProcessPattern)
	}
	script := autostartScript(t)
	if !strings.Contains(script, `ENGINE_PATTERN="`+engineProcessPattern+`"`) {
		t.Errorf("tools/engine-autostart.sh does not grep for %q", engineProcessPattern)
	}
}

// TestTheRestartCapMatchesTheAutostartScript. The cap is the thing that stops a
// refusing engine being restarted forever, and a script with a different one
// would make the Go constant a comment.
func TestTheRestartCapMatchesTheAutostartScript(t *testing.T) {
	script := autostartScript(t)
	if !strings.Contains(script, "RESTART_CAP="+strconv.Itoa(engineRestartCap)) {
		t.Errorf("tools/engine-autostart.sh does not carry RESTART_CAP=%d", engineRestartCap)
	}
	if engineRestartCap <= 0 {
		t.Error("a non-positive restart cap would mean an engine is never restarted at all")
	}
}

// TestTheStalenessWindowMatchesTheAutostartScript. The script's precheck reads
// the same advisory marker the console does, so it has to believe it for the same
// length of time.
func TestTheStalenessWindowMatchesTheAutostartScript(t *testing.T) {
	script := autostartScript(t)
	minutes := int(enginelock.StaleAfter / time.Minute)
	if !strings.Contains(script, "MARKER_STALE_MINUTES="+strconv.Itoa(minutes)) {
		t.Errorf("tools/engine-autostart.sh does not carry MARKER_STALE_MINUTES=%d", minutes)
	}
	if !strings.Contains(script, enginelock.MarkerFileName) {
		t.Errorf("the script does not read %s; its precheck is looking at a different file",
			enginelock.MarkerFileName)
	}
}

// TestTheScriptIsPreparedAndNotInstalled is the §0.7 boundary, written down where
// somebody about to install it will read it.
//
// "부팅마다 주문 능력 프로세스 자동 기동" is an operational configuration change,
// not a tool convenience, so the read-only soak's precedent does not transfer
// (review round 1, P2). The script says so at the top; this test keeps it saying
// so.
func TestTheScriptIsPreparedAndNotInstalled(t *testing.T) {
	script := autostartScript(t)
	for _, want := range []string{"준비만", "§0.7"} {
		if !strings.Contains(script, want) {
			t.Errorf("tools/engine-autostart.sh does not carry %q; the installation boundary has to be "+
				"legible to whoever opens the file", want)
		}
	}
	// And nothing in this repository installs it.
	for _, name := range []string{"engineproc.go", "console.go", "engine.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if strings.Contains(string(data), "engine-autostart.sh") &&
			!strings.Contains(string(data), "// ") {
			t.Errorf("%s references the autostart script outside a comment", name)
		}
	}
}
