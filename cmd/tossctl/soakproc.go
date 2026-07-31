package main

// soakproc.go stops and restarts the read-only survey from outside it
// (openspec change verify-execution-capability, task 1.8 ②).
//
// # Why this exists and why it looks like a shell script
//
// The operator drives everything from the console now. The soak, though, is a
// separate long-running process started by ~/.local/share/tossos/bin/
// soak-autostart.sh, and until this existed the only way to bounce it was a
// terminal. So this does exactly what that script does — pgrep for the survey,
// signal it, start it again detached with its output appended to the same log —
// because two mechanisms that both claim to manage the same process are one
// mechanism more than there should be.
//
// # It cannot touch an account
//
// Nothing here holds a broker, a credential or a token. It signals a process and
// spawns `tossctl soak run`, whose own package is structurally incapable of
// mutating anything (internal/soak's package doc and its import-graph test). The
// restart is therefore a tool restart in the strictest sense: the worst it can do
// is cost the survey a cycle.
//
// # SIGINT, and then patience
//
// `soak run` installs a signal handler so an interrupt closes the record cleanly
// rather than leaving the last cycle half-written. Killing it would work and would
// throw that away, so this asks and then waits. If it will not go, nothing is
// spawned: two surveys appending to one record and competing for one rate limit is
// worse than a survey that needs a person.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
)

// soakLogName is where a detached survey's output goes. It is the name
// soak-autostart.sh already uses, in the directory the record already lives in.
const soakLogName = "soak.log"

// soakProcessPattern is how a running survey is recognised. It is
// soak-autostart.sh's pattern, unchanged.
const soakProcessPattern = "tossctl soak run"

// soakStopTimeout bounds the wait for an interrupted survey to finish its cycle.
//
// A cycle is a handful of reads and the interrupt is checked between them, so this
// is generous; the case it is sized for is a survey that is mid-request against a
// broker that is being slow.
const soakStopTimeout = 30 * time.Second

// The seams. They are package variables so cmd/tossctl's own tests can drive the
// whole restart without a process table, a signal or a fork — there is no flag and
// no environment variable that reaches them.
var (
	// soakFindProcesses lists the pids of running surveys.
	soakFindProcesses = pgrepSoak
	// soakSignalProcess asks one to stop.
	soakSignalProcess = interruptPID
	// soakProcessAlive reports that a pid is still there.
	soakProcessAlive = pidAlive
	// soakSpawnDetached starts a new survey that outlives this process.
	soakSpawnDetached = spawnDetachedSoak
	// soakSleep is the wait between liveness polls.
	soakSleep = time.Sleep
)

// soakLogPath is where the detached survey appends, derived from the record's
// location so an isolated --config-dir profile keeps its own.
func soakLogPath(recordPath string) string {
	return filepath.Join(filepath.Dir(recordPath), soakLogName)
}

// restartSoak is the console's RestartSoak seam.
//
// It returns one line describing what happened, which the dashboard prints
// verbatim: the operator pressed a button on a process they cannot see, and the
// answer has to say whether anything was actually stopped.
func restartSoak(recordPath string, prepareSpawn ...func() error) (string, error) {
	binary, err := binstamp.SelfPath()
	if err != nil {
		return "", err
	}
	logPath := soakLogPath(recordPath)

	pids, err := soakFindProcesses()
	if err != nil {
		return "", err
	}

	stopped := make([]int, 0, len(pids))
	for _, pid := range pids {
		if pid == os.Getpid() {
			continue // this console; pgrep should not match it, but a guess is not a licence
		}
		if err := soakSignalProcess(pid); err != nil {
			return "", fmt.Errorf("pid %d에 SIGINT를 보낼 수 없다: %w", pid, err)
		}
		stopped = append(stopped, pid)
	}

	for _, pid := range stopped {
		if err := waitForExit(pid); err != nil {
			return "", err
		}
	}

	if len(prepareSpawn) > 0 && prepareSpawn[0] != nil {
		if err := prepareSpawn[0](); err != nil {
			return "", fmt.Errorf("새 soak 시작 직전 token cache를 준비하지 못했다: %w", err)
		}
	}
	if err := soakSpawnDetached(binary, logPath); err != nil {
		return "", err
	}

	switch len(stopped) {
	case 0:
		return fmt.Sprintf("실행 중인 soak을 찾지 못했다. 새로 시작했다 — 로그 %s", logPath), nil
	case 1:
		return fmt.Sprintf("pid %d에 SIGINT를 보내 정상 종료시킨 뒤 다시 시작했다 — 로그 %s",
			stopped[0], logPath), nil
	default:
		return fmt.Sprintf("soak 프로세스 %s를 정상 종료시킨 뒤 하나로 다시 시작했다 — 로그 %s",
			joinPIDs(stopped), logPath), nil
	}
}

// waitForExit gives an interrupted survey time to close its record.
func waitForExit(pid int) error {
	deadline := time.Now().Add(soakStopTimeout)
	for time.Now().Before(deadline) {
		if !soakProcessAlive(pid) {
			return nil
		}
		soakSleep(100 * time.Millisecond)
	}
	return fmt.Errorf("pid %d이(가) %s 안에 종료되지 않았다. 새 soak을 시작하지 않았다 — "+
		"기록 하나에 두 서베이가 붙는 편이 더 나쁘다", pid, soakStopTimeout)
}

func joinPIDs(pids []int) string {
	parts := make([]string, 0, len(pids))
	for _, pid := range pids {
		parts = append(parts, strconv.Itoa(pid))
	}
	return strings.Join(parts, ", ")
}

// soakReExec is the survey's own upgrade path: replace this process with the
// binary now installed at its path, keeping the arguments it was started with.
//
// It runs only at a cycle boundary, and only when the fingerprint moved. The
// record is untouched: it is append-only, every cycle is synced before the boundary
// is reached, and the successor opens the same path.
func soakReExec() error {
	path, err := binstamp.SelfPath()
	if err != nil {
		return err
	}
	return reexecSelf(path, os.Args)
}

// --- the real implementations -------------------------------------------------

// pgrepSoak finds running surveys the way soak-autostart.sh does.
//
// pgrep rather than a /proc walk: the autostart script is the other half of this
// mechanism and it uses pgrep, so a survey that one of them can see and the other
// cannot would be a bug that only appears at three in the morning.
func pgrepSoak() ([]int, error) {
	out, err := exec.Command("pgrep", "-f", soakProcessPattern).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil // pgrep's "nothing matched"
		}
		return nil, fmt.Errorf("실행 중인 soak을 찾을 수 없다 (pgrep): %w", err)
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

func interruptPID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(os.Interrupt)
}

// pidAlive reports whether a pid can still be signalled. Signal 0 is the portable
// "does this exist" probe.
func pidAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(nil) == nil
}

// spawnDetachedSoak starts `tossctl soak run` so it outlives this process.
//
// setsid is what soak-autostart.sh uses and it is what makes the survey survive the
// console being closed: a new session means no controlling terminal, so the
// operator's Ctrl-C reaches the console and not the survey behind it.
func spawnDetachedSoak(binary, logPath string) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return fmt.Errorf("로그 디렉터리를 만들 수 없다 (%s): %w", filepath.Dir(logPath), err)
	}
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("로그 파일을 열 수 없다 (%s): %w", logPath, err)
	}
	defer logFile.Close()

	cmd := exec.Command(binary, "soak", "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("soak을 다시 시작할 수 없다 (%s soak run): %w", binary, err)
	}
	// Nothing waits for it. It is meant to run for days, and the console that
	// started it may itself be restarted in a minute.
	go func() { _ = cmd.Wait() }()
	return nil
}
