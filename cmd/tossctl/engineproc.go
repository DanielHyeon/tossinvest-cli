package main

// engineproc.go starts and stops the engine from outside it (openspec change
// add-engine-runtime, task 2.1), and holds the constants the boot autostart
// script shares with this binary (task 2.2).
//
// # Why it looks like soakproc.go
//
// Because it is the same mechanism aimed at a different process, and soakproc.go
// already argued the design: pgrep for the process, signal it, wait for it,
// spawn it detached with its output appended to a log beside its data. Two
// mechanisms that both claim to manage one process is one mechanism more than
// there should be, so the autostart script and this file use the same pgrep
// pattern — asserted by a drift test, exactly as the soak's is.
//
// # What is different, and it matters
//
// The soak is structurally incapable of mutating an account. The engine is not:
// once the gate is approved and the interlock is satisfied, its exit loop places
// real orders. So two things here are stricter than the soak's equivalents:
//
//	the start is not blind    it waits briefly and reports the engine's own
//	                          refusal, because "I pressed start and nothing
//	                          happened" is the answer an operator gets otherwise
//	the marker is consulted   a fresh advisory marker means an engine is already
//	                          running and nothing is spawned. The *exclusion* is
//	                          still the engine's own flock — this check exists so
//	                          the console can say "already running" instead of
//	                          spawning a process that immediately refuses
//
// Nothing here holds a broker, a credential or a token. It signals a process and
// spawns `tossctl engine run`, which does its own gate check, its own interlock
// and its own locking.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

// engineLogName is where a detached engine's output goes, beside its journal.
const engineLogName = "engine.log"

// engineProcessPattern is how a running engine is recognised.
//
// It is the autostart script's pattern, unchanged. A process one of them can see
// and the other cannot is a bug that only appears at three in the morning —
// engineproc_test.go asserts the two agree, the way soakproc_test.go does for the
// survey.
const engineProcessPattern = "tossctl engine run"

// engineRestartCap is how many times the autostart script restarts a crashing
// engine before it gives up.
//
// Five. A crash loop is not a transient condition for a process whose whole job
// is to hold a durable, single-writer journal open: the two ways an engine dies
// repeatedly are a corrupt journal and a refused start, and restarting either one
// forever is a way to fill a disk with logs while nobody is told. When the cap is
// reached the script stops and leaves the reason in the log, which is the state
// the operator's next dashboard visit shows as "not running".
const engineRestartCap = 5

// engineStopTimeout bounds the wait for a signalled engine to finish its cycle
// and close the journal.
//
// A cycle is a reconciliation pass at worst — two stabilised snapshots with the
// interval between them, plus the adoption judgement — so this is generous. The
// case it is sized for is a loop mid-request against a broker that is being slow.
const engineStopTimeout = 60 * time.Second

// engineStartProbe is how long a start waits to see whether the engine refused.
//
// Every refusal `engine run` can produce happens before the loops start: the
// flock, the gate, the interlock, the run lock. They are local checks plus one
// account read, so a second and a half is comfortably longer than all of them and
// short enough that an operator watching a page does not think it hung.
// It is a variable rather than a constant only so this package's tests do not
// spend three seconds each proving a spawn happened; nothing else writes it.
var engineStartProbe = 3 * time.Second

// engineLogTail is how much of the log a refusal reports back. The enumerated
// interlock clauses are a handful of lines; this is room for them plus whatever
// cobra printed around them.
const engineLogTail = 4096

// The seams. They are package variables so this package's own tests can drive the
// whole thing without a process table, a signal or a fork — there is no flag and
// no environment variable that reaches them.
var (
	engineFindProcesses = pgrepEngine
	engineSignalProcess = terminatePID
	engineProcessAlive  = pidAlive
	engineSpawnDetached = spawnDetachedEngine
	engineSleep         = time.Sleep
)

// engineLogPath is where the detached engine appends, derived from the journal
// directory so an isolated --config-dir profile keeps its own.
func engineLogPath(dir string) string { return filepath.Join(dir, engineLogName) }

// startEngine is the console's StartEngine seam.
//
// It returns one line describing what happened, which the dashboard prints. On a
// refusal it returns the tail of the engine's own log alongside the error,
// because the reason a start was refused is the engine's to give: the console
// does not evaluate the gate and must not appear to.
func startEngine(root *rootOptions) (string, error) {
	dir, err := engineJournalDir(root)
	if err != nil {
		return "", err
	}
	binary, err := binstamp.SelfPath()
	if err != nil {
		return "", err
	}

	// Advisory, and only to give a better answer than "the engine refused because
	// it could not take the lock". The exclusion is the engine's own flock.
	if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
		return "", fmt.Errorf("엔진이 이미 실행 중이다 (pid %d, 마지막 갱신 %s)",
			status.Marker.PID, status.RefreshedAt.UTC().Format(time.RFC3339))
	}
	if pids, perr := engineFindProcesses(); perr == nil && len(pids) > 0 {
		return "", fmt.Errorf("엔진 프로세스가 이미 있다 (%s)", joinPIDs(pids))
	}

	logPath := engineLogPath(dir)
	wait, err := engineSpawnDetached(binary, logPath, engineArgs(root))
	if err != nil {
		return "", err
	}

	// Watch briefly for an immediate exit. Every refusal happens before the loops
	// start, so an engine that is still alive after the probe is an engine that
	// got past all of them.
	select {
	case exitErr := <-wait:
		tail := readLogTail(logPath, engineLogTail)
		if exitErr == nil {
			return tail, errors.New("엔진이 곧바로 종료했다 (오류 없이)")
		}
		return tail, fmt.Errorf("엔진이 곧바로 종료했다: %w", exitErr)
	case <-time.After(engineStartProbe):
	}
	return fmt.Sprintf("엔진을 시작했다 — 로그 %s", logPath), nil
}

// stopEngine is the console's StopEngine seam: the signal discipline, not a kill.
func stopEngine(root *rootOptions) (string, error) {
	pids, err := engineFindProcesses()
	if err != nil {
		return "", err
	}
	stopped := make([]int, 0, len(pids))
	for _, pid := range pids {
		if pid == os.Getpid() {
			continue // this console; pgrep should not match it, but a guess is not a licence
		}
		if err := engineSignalProcess(pid); err != nil {
			return "", fmt.Errorf("pid %d에 종료 시그널을 보낼 수 없다: %w", pid, err)
		}
		stopped = append(stopped, pid)
	}
	if len(stopped) == 0 {
		return "실행 중인 엔진을 찾지 못했다.", nil
	}
	for _, pid := range stopped {
		if err := waitForEngineExit(pid, engineStopTimeout); err != nil {
			return "", err
		}
	}
	if dir, derr := engineJournalDir(root); derr == nil {
		if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
			return fmt.Sprintf("%s를 종료시켰지만 활성 마커가 아직 신선하다 (%s) — 최대 %s 뒤 사라진다",
				joinPIDs(stopped), status.Marker.StartedAt.UTC().Format(time.RFC3339),
				enginelock.StaleAfter), nil
		}
	}
	return fmt.Sprintf("%s에 종료 시그널을 보내 루프 완주·journal 정합 close까지 기다렸다",
		joinPIDs(stopped)), nil
}

// waitForEngineExit gives a signalled engine time to close its journal.
//
// The bound is a parameter rather than the constant read inline so this
// function's own behaviour — report, never kill — is testable in milliseconds
// instead of in the minute the production value is.
func waitForEngineExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !engineProcessAlive(pid) {
			return nil
		}
		engineSleep(100 * time.Millisecond)
	}
	return fmt.Errorf("pid %d이(가) %s 안에 종료되지 않았다. 한 번 더 시그널을 보내면 즉시 종료하지만 "+
		"journal은 정합하게 닫히지 않는다 — 재기동 시 복구 절차가 미결 시도를 정리한다",
		pid, timeout)
}

// engineArgs is the command line a spawned engine gets: the subcommand plus the
// --config-dir the console itself was started with, so an isolated profile stays
// isolated across the button.
func engineArgs(root *rootOptions) []string {
	args := []string{"engine", "run"}
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		args = append([]string{"--config-dir", root.configDir}, args...)
	}
	return args
}

// --- the real implementations -------------------------------------------------

// pgrepEngine finds running engines the way the autostart script does.
func pgrepEngine() ([]int, error) {
	out, err := exec.Command("pgrep", "-f", engineProcessPattern).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil // pgrep's "nothing matched"
		}
		return nil, fmt.Errorf("실행 중인 엔진을 찾을 수 없다 (pgrep): %w", err)
	}
	var pids []int
	for _, line := range strings.Fields(string(out)) {
		pid, convErr := strconv.Atoi(line)
		if convErr != nil || pid <= 0 {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// spawnDetachedEngine starts `tossctl engine run` so it outlives this process,
// and returns a channel carrying its exit.
//
// The channel is what lets the start report a refusal. Nothing blocks on it: an
// engine that is still running when the probe expires simply leaves a goroutine
// waiting, which ends when the engine does.
func spawnDetachedEngine(binary, logPath string, args []string) (<-chan error, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return nil, fmt.Errorf("로그 디렉터리를 만들 수 없다 (%s): %w", filepath.Dir(logPath), err)
	}
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("로그 파일을 열 수 없다 (%s): %w", logPath, err)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("엔진을 시작할 수 없다 (%s %s): %w", binary, strings.Join(args, " "), err)
	}
	wait := make(chan error, 1)
	go func() {
		defer logFile.Close()
		wait <- cmd.Wait()
	}()
	return wait, nil
}

// readLogTail returns the last n bytes of a log, for reporting a refusal.
func readLogTail(path string, n int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > n {
		data = data[len(data)-n:]
	}
	return strings.TrimSpace(string(data))
}

// terminatePID asks an engine to stop.
//
// SIGTERM rather than the soak's SIGINT: the engine is started detached, with no
// controlling terminal, so an interrupt is not the signal a supervisor or a
// button would send. `engine run` installs handlers for both and treats the first
// delivery of either as graceful (engine.go engineStopSignals), so the two are
// interchangeable for correctness and SIGTERM is the honest one for a process
// nobody is sitting in front of.
func terminatePID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}
