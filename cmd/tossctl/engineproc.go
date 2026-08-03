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
// The shared half stops at the pattern. pgrep answers "which processes look like an
// engine"; only this file goes on to ask "which of them is mine", by resolving each
// candidate's --config-dir to a journal directory and comparing it with this
// console's own (change a059). A shell cannot do that resolution without
// reimplementing it, and a reimplementation that drifts is the bug this whole
// arrangement exists to prevent.
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
//	the signal is aimed       at this profile's engine and no other. A host shares
//	                          its PID namespace with its containers, so a widened
//	                          pattern can see an engine holding the exit loops for
//	                          somebody else's journal, and SIGTERM is not a signal
//	                          to send on a guess (change a059)
//	the marker is consulted   but it does not decide. A fresh advisory marker
//	                          supplies the PID and refresh time that make a refusal
//	                          readable; what makes it a refusal is an observed
//	                          process. The *exclusion* is the engine's own flock.
//	                          A marker outlives its process — a container recreate,
//	                          a SIGKILL and a reboot all leave the file behind —
//	                          so letting it refuse on its own is how an engine
//	                          stays down while the file says it is up (change a056)
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
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

// engineLogName is where a detached engine's output goes, beside its journal.
const engineLogName = "engine.log"

// engineProcessPattern is how an engine *candidate* is recognised: the extended
// regular expression both this binary and the autostart script hand to `pgrep -f`.
//
// It is a regular expression rather than the literal "tossctl engine run" it used
// to be, because that literal never matched an engine this console started
// (change a059). engineArgs puts the console's own --config-dir and --session-file
// in front of the subcommand, so the command line is
//
//	/usr/local/bin/tossctl --config-dir … --session-file … engine run
//
// and the words are not adjacent. Measured in production on 2026-08-03: pgrep found
// nothing while the engine ran as pid 16, so the stop button answered "실행 중인
// 엔진을 찾지 못했다" about a process it had started itself.
//
// The leading space before `engine run` is load-bearing: it makes `engine` a whole
// argv token, so a path like /srv/myengine runtime is not an engine. There is no
// `$` anchor, because `tossctl engine run --config-dir X` is a command a person can
// type and an anchor would silently miss it — which is the shape of the bug above.
//
// It selects candidates. It does not establish ownership: this pattern matches every
// profile's engine, and enginePIDsForJournal decides which of them is ours (a059
// design D2). A process one of the two halves can see and the other cannot is a bug
// that only appears at three in the morning — engineproc_test.go asserts the script
// and this constant agree, the way soakproc_test.go does for the survey.
const engineProcessPattern = `tossctl( .*)? engine run`

// engineProcessMatcher is engineProcessPattern compiled once, so ownership can be
// judged from a command line without another trip through pgrep.
var engineProcessMatcher = regexp.MustCompile(engineProcessPattern)

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
	// engineFindProcesses lists the pids of engines running against journalDir.
	engineFindProcesses = pgrepEngine
	// engineListProcesses is the raw enumeration underneath it: one `pid command`
	// line per candidate. It is a seam of its own so a test can exercise the
	// pattern, the parsing and the ownership judgement together — everything except
	// the process table (change a059).
	engineListProcesses = pgrepEngineLines
	engineSignalProcess = terminatePID
	engineProcessAlive  = pidAlive
	engineSpawnDetached = spawnDetachedEngine
	engineSleep         = time.Sleep
)

// engineLogPath is where the detached engine appends, derived from the journal
// directory so an isolated --config-dir profile keeps its own.
func engineLogPath(dir string) string { return filepath.Join(dir, engineLogName) }

// markerRefusesStart decides whether a fresh advisory marker may refuse a start
// (change a056).
//
// It may only when something else agrees with it: either a process was actually
// observed, or the enumeration failed and absence therefore cannot be claimed.
// A fresh marker on its own proves nothing — a container recreate, a SIGKILL and
// a host reboot all delete the process and leave the file, and engine-safety
// already says this signal is advisory and that 배타는 flock이 담당한다.
//
// It is a named function rather than an inline condition so that the reason
// survives the next edit: the shape this replaces was two sequential `return`s
// where the advisory one came first and the process check below it was
// unreachable.
func markerRefusesStart(markerFresh, processObserved, enumerationFailed bool) bool {
	return markerFresh && (processObserved || enumerationFailed)
}

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

	// One enumeration, read by both guards below. Counting twice invites two
	// different answers inside a single decision. It is scoped to this profile's
	// journal, so another profile's engine is not a reason to refuse this one
	// (change a059) — they hold different flocks and are different instances.
	pids, findErr := engineFindProcesses(dir)
	observed := findErr == nil && len(pids) > 0

	// Advisory, and only to give a better answer than "the engine refused because
	// it could not take the lock": the marker carries the PID and the refresh time,
	// which a flock failure does not. The exclusion is the engine's own flock, so
	// this may name a running instance but may not by itself refuse one
	// (change a056 — see markerRefusesStart).
	if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); markerRefusesStart(
		status.Running, observed, findErr != nil) {
		return "", fmt.Errorf("엔진이 이미 실행 중이다 (pid %d, 마지막 갱신 %s)",
			status.Marker.PID, status.RefreshedAt.UTC().Format(time.RFC3339))
	}
	if observed {
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
	// The journal directory comes first because it is what decides *whose* engine
	// each candidate is, and there is no signalling anything without that (change
	// a059). It used to be resolved at the bottom, for the marker sentence only,
	// with its error discarded.
	dir, err := engineJournalDir(root)
	if err != nil {
		return "", err
	}
	pids, err := engineFindProcesses(dir)
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
	if status := enginelock.Read(enginelock.MarkerPath(dir), time.Now()); status.Running {
		return fmt.Sprintf("%s를 종료시켰지만 활성 마커가 아직 신선하다 (%s) — 최대 %s 뒤 사라진다",
			joinPIDs(stopped), status.Marker.StartedAt.UTC().Format(time.RFC3339),
			enginelock.StaleAfter), nil
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
	if root != nil && strings.TrimSpace(root.sessionFile) != "" {
		args = append([]string{"--session-file", root.sessionFile}, args...)
	}
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		args = append([]string{"--config-dir", root.configDir}, args...)
	}
	return args
}

// --- the real implementations -------------------------------------------------

// pgrepEngine finds the engines running against journalDir.
//
// Two steps, and they are different questions. pgrep answers "which processes look
// like an engine", which is all a shell can ask and all the autostart script does
// ask. This then answers "which of those is mine", which it can only do because it
// has the command lines and the same directory resolution the console itself uses.
func pgrepEngine(journalDir string) ([]int, error) {
	lines, err := engineListProcesses()
	if err != nil {
		return nil, err
	}
	// A failure to resolve the default is not fatal here: it only means a flag-free
	// engine cannot be proven to be ours, and enginePIDsForJournal drops what it
	// cannot prove.
	fallback, _ := engineJournalDir(nil)
	return enginePIDsForJournal(lines, journalDir, fallback), nil
}

// pgrepEngineLines enumerates engine candidates the way the autostart script does,
// with `-a` so each pid arrives with the command line that has to be judged.
func pgrepEngineLines() ([]string, error) {
	out, err := exec.Command("pgrep", "-a", "-f", engineProcessPattern).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil // pgrep's "nothing matched"
		}
		// Every other failure — pgrep missing, -a unsupported, permission denied —
		// is an enumeration that did not happen. It is not an absence, and a056
		// keeps its refusal on it.
		return nil, fmt.Errorf("실행 중인 엔진을 찾을 수 없다 (pgrep): %w", err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

// enginePIDsForJournal keeps the pids whose command line proves they are running
// against journalDir (change a059).
//
// Ownership is the journal directory, not the flag string: engine-safety defines an
// instance by the flock on that directory, and a console told the default path
// explicitly is the same instance as an autostart that left the flag off. Resolving
// both sides through engineJournalDir is what makes those two agree.
//
// The judgement itself is pidsOwnedBy, which the survey shares (change a060). This
// is the engine's half of it: the engine's matcher, and the engine's answer to
// "which resource does this command line own".
func enginePIDsForJournal(lines []string, journalDir, defaultDir string) []int {
	return pidsOwnedBy(lines, engineProcessMatcher, journalDir, func(configDir string) string {
		if strings.TrimSpace(configDir) == "" {
			return defaultDir
		}
		return configDir
	})
}

// pidsOwnedBy keeps the pids whose command line proves they own `want` (change a060).
//
// Both console-managed processes ask this question and neither may guess at it: the
// answer decides who gets a signal. What differs between them is only the matcher
// and what counts as the owned resource — the engine's journal directory, the
// survey's record path — so resolve maps a command line's --config-dir to that
// resource, using whichever function the console itself uses for its own. An empty
// configDir means the process named no profile, and resolve answers with the
// default one.
//
// Everything unprovable is dropped: an unparsable line, a pid with no command line
// behind it, a resource that cannot be resolved. "I could not tell" has to mean
// "not mine", because the list being built is a list of processes to signal.
func pidsOwnedBy(lines []string, matcher *regexp.Regexp, want string,
	resolve func(configDir string) string) []int {
	if strings.TrimSpace(want) == "" {
		return nil
	}
	target := filepath.Clean(want)

	var pids []int
	for _, line := range lines {
		pid, command, ok := splitProcessLine(line)
		if !ok || !matcher.MatchString(command) {
			continue
		}
		owned := resolve(commandConfigDir(command))
		if strings.TrimSpace(owned) == "" || filepath.Clean(owned) != target {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}

// splitProcessLine reads one `pgrep -a` line: the pid, then the command line.
//
// A line with no command line after the pid fails here rather than yielding a pid
// nobody can vouch for.
func splitProcessLine(line string) (int, string, bool) {
	number, command, found := strings.Cut(strings.TrimSpace(line), " ")
	if !found {
		return 0, "", false
	}
	pid, err := strconv.Atoi(number)
	if err != nil || pid <= 0 {
		return 0, "", false
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return 0, "", false
	}
	return pid, command, true
}

// commandConfigDir pulls --config-dir out of a command line, in both the
// spellings cobra accepts. An empty result means the process did not name one and
// is therefore using the default.
func commandConfigDir(command string) string {
	fields := strings.Fields(command)
	for i, field := range fields {
		switch {
		case field == "--config-dir" && i+1 < len(fields):
			return fields[i+1]
		case strings.HasPrefix(field, "--config-dir="):
			return strings.TrimPrefix(field, "--config-dir=")
		}
	}
	return ""
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
