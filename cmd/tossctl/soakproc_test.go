package main

// soakproc_test.go covers the two pieces of task 1.8 that live in the command
// rather than in a package: the argument list a console restart re-executes with,
// and the stop-then-start the soak button performs.
//
// Nothing here signals, spawns or execs. soakproc.go's four operations are package
// variables for exactly this reason, and every test swaps them for recorders.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- the re-executed argument list -------------------------------------------------

// TestArgvWithPortKeepsTheCommandLineAndPinsThePort.
//
// The port is the one thing the operator did not choose when they started a console
// without --port, and it is the one thing that has to survive: the browser is
// already sitting on it. Everything else — the subcommand, --config-dir, whatever
// they typed — is theirs and is kept.
func TestArgvWithPortKeepsTheCommandLineAndPinsThePort(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "no port was given",
			in:   []string{"/usr/local/bin/tossctl", "console"},
			want: []string{"/usr/local/bin/tossctl", "console", "--port", "45678"},
		},
		{
			name: "a port was given as two arguments",
			in:   []string{"tossctl", "console", "--port", "9000"},
			want: []string{"tossctl", "console", "--port", "45678"},
		},
		{
			name: "a port was given as one argument",
			in:   []string{"tossctl", "console", "--port=9000"},
			want: []string{"tossctl", "console", "--port", "45678"},
		},
		{
			name: "other flags are kept, in order",
			in:   []string{"tossctl", "--config-dir", "/tmp/p", "console", "--port", "1", "--verbose"},
			want: []string{"tossctl", "--config-dir", "/tmp/p", "console", "--verbose", "--port", "45678"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := argvWithPort(tc.in, 45678)
			if strings.Join(got, " ") != strings.Join(tc.want, " ") {
				t.Errorf("argvWithPort =\n  %v\nwant\n  %v", got, tc.want)
			}
		})
	}
}

// TestArgvWithPortNeverLosesTheProgramName. argv[0] is what an exec is given as the
// process name; dropping it would produce a process that cannot say what it is.
func TestArgvWithPortNeverLosesTheProgramName(t *testing.T) {
	got := argvWithPort([]string{"--port", "1"}, 2)
	if len(got) == 0 || got[0] != "--port" {
		t.Errorf("argvWithPort mangled argv[0]: %v", got)
	}
}

// --- the soak restart ------------------------------------------------------------

// procFakes replaces soakproc.go's four operations for one test.
type procFakes struct {
	found     []int
	findErr   error
	signalled []int
	signalErr error
	// aliveFor counts how many liveness polls each pid survives.
	aliveFor map[int]int
	spawned  [][2]string
	spawnErr error
	slept    time.Duration
}

func (f *procFakes) install(t *testing.T) {
	t.Helper()
	if f.aliveFor == nil {
		f.aliveFor = map[int]int{}
	}
	oldFind, oldSignal, oldAlive, oldSpawn, oldSleep :=
		soakFindProcesses, soakSignalProcess, soakProcessAlive, soakSpawnDetached, soakSleep
	t.Cleanup(func() {
		soakFindProcesses, soakSignalProcess, soakProcessAlive, soakSpawnDetached, soakSleep =
			oldFind, oldSignal, oldAlive, oldSpawn, oldSleep
	})

	soakFindProcesses = func() ([]int, error) { return f.found, f.findErr }
	soakSignalProcess = func(pid int) error {
		if f.signalErr != nil {
			return f.signalErr
		}
		f.signalled = append(f.signalled, pid)
		return nil
	}
	soakProcessAlive = func(pid int) bool {
		if f.aliveFor[pid] <= 0 {
			return false
		}
		f.aliveFor[pid]--
		return true
	}
	soakSpawnDetached = func(bin, log string) error {
		if f.spawnErr != nil {
			return f.spawnErr
		}
		f.spawned = append(f.spawned, [2]string{bin, log})
		return nil
	}
	soakSleep = func(d time.Duration) { f.slept += d }
}

// TestRestartingTheSoakInterruptsItThenStartsItAgain.
func TestRestartingTheSoakInterruptsItThenStartsItAgain(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{found: []int{4242}, aliveFor: map[int]int{4242: 3}}
	f.install(t)

	note, err := restartSoak(record)
	if err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if len(f.signalled) != 1 || f.signalled[0] != 4242 {
		t.Fatalf("signalled %v, want [4242]", f.signalled)
	}
	if len(f.spawned) != 1 {
		t.Fatalf("spawned %d survey(s), want 1", len(f.spawned))
	}
	if got := f.spawned[0][1]; got != soakLogPath(record) {
		t.Errorf("the new survey logs to %s, want %s — the restart must append to the log that is "+
			"already being read", got, soakLogPath(record))
	}
	if !strings.Contains(note, "4242") {
		t.Errorf("the operator is not told what was stopped: %q", note)
	}
	if f.slept == 0 {
		t.Error("the restart did not wait for the interrupted survey to close its record")
	}
}

// TestRestartingWithNothingRunningJustStartsOne. The autostart script's own case:
// the machine rebooted and nobody noticed.
func TestRestartingWithNothingRunningJustStartsOne(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{}
	f.install(t)

	note, err := restartSoak(record)
	if err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if len(f.signalled) != 0 {
		t.Errorf("signalled %v with nothing running", f.signalled)
	}
	if len(f.spawned) != 1 {
		t.Fatalf("spawned %d survey(s), want 1", len(f.spawned))
	}
	if !strings.Contains(note, "찾지 못했다") {
		t.Errorf("the note does not say nothing was running: %q", note)
	}
}

// TestASurveyThatWillNotStopBlocksTheRestart.
//
// Two surveys appending to one record and competing for one rate limit is a worse
// state than a survey that needs a person, so nothing is spawned.
func TestASurveyThatWillNotStopBlocksTheRestart(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{found: []int{7}, aliveFor: map[int]int{7: 1 << 30}}
	f.install(t)

	_, err := restartSoak(record)
	if err == nil {
		t.Fatal("a survey that ignored SIGINT did not block the restart")
	}
	if !strings.Contains(err.Error(), "종료되지 않았다") {
		t.Errorf("the error does not say what happened: %v", err)
	}
	if len(f.spawned) != 0 {
		t.Fatalf("a second survey was started anyway: %v", f.spawned)
	}
}

// TestTheRestartNeverSignalsThisProcess. pgrep matches on a command line, and a
// pattern that ever matched the console would have it kill itself.
func TestTheRestartNeverSignalsThisProcess(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{found: []int{os.Getpid()}}
	f.install(t)

	if _, err := restartSoak(record); err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if len(f.signalled) != 0 {
		t.Fatalf("the restart signalled %v, which includes this process", f.signalled)
	}
}

// TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted.
func TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{findErr: errors.New("pgrep: not found")}
	f.install(t)

	if _, err := restartSoak(record); err == nil {
		t.Fatal("a failed search reported success")
	}
	if len(f.spawned) != 0 {
		t.Errorf("a survey was started without knowing whether one was already running: %v", f.spawned)
	}
}

// TestTheLogSitsBesideTheRecord, which is where soak-autostart.sh puts it and where
// the operator is already looking.
func TestTheLogSitsBesideTheRecord(t *testing.T) {
	got := soakLogPath("/home/x/.local/share/tossos/capability-soak.jsonl")
	want := "/home/x/.local/share/tossos/soak.log"
	if got != want {
		t.Errorf("soakLogPath = %s, want %s", got, want)
	}
}

// TestTheProcessPatternMatchesTheAutostartScript.
//
// soak-autostart.sh and this file are two halves of one mechanism. A survey one of
// them can see and the other cannot is a bug that only shows up at three in the
// morning.
func TestTheProcessPatternMatchesTheAutostartScript(t *testing.T) {
	if soakProcessPattern != "tossctl soak run" {
		t.Errorf("the pattern is %q; soak-autostart.sh greps for \"tossctl soak run\"", soakProcessPattern)
	}
}
