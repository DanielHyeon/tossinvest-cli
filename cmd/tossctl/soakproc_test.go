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
	"regexp"
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
	// spawnArgs is the argv each spawn was given, and findRecords is the record
	// each lookup was asked about — the two things a060 is about.
	spawnArgs   [][]string
	findRecords []string
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

	soakFindProcesses = func(record string) ([]int, error) {
		f.findRecords = append(f.findRecords, record)
		return f.found, f.findErr
	}
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
	soakSpawnDetached = func(bin, log string, args []string) error {
		if f.spawnErr != nil {
			return f.spawnErr
		}
		f.spawned = append(f.spawned, [2]string{bin, log})
		f.spawnArgs = append(f.spawnArgs, args)
		return nil
	}
	soakSleep = func(d time.Duration) { f.slept += d }
}

// TestRestartingTheSoakInterruptsItThenStartsItAgain.
func TestRestartingTheSoakInterruptsItThenStartsItAgain(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{found: []int{4242}, aliveFor: map[int]int{4242: 3}}
	f.install(t)

	note, err := restartSoak(nil, record)
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

	note, err := restartSoak(nil, record)
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

	_, err := restartSoak(nil, record)
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

	if _, err := restartSoak(nil, record); err != nil {
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

	if _, err := restartSoak(nil, record); err == nil {
		t.Fatal("a failed search reported success")
	}
	if len(f.spawned) != 0 {
		t.Errorf("a survey was started without knowing whether one was already running: %v", f.spawned)
	}
}

// TestTheLogSitsBesideTheRecord preserves the legacy soak-autostart.sh layout,
// which is where the operator is already looking.
func TestTheLogSitsBesideTheRecord(t *testing.T) {
	got := soakLogPath("/home/x/.local/share/tossos/capability-soak.jsonl")
	want := "/home/x/.local/share/tossos/soak.log"
	if got != want {
		t.Errorf("soakLogPath = %s, want %s", got, want)
	}
}

// TestTheSoakSpawnCarriesThisProfile (change a060).
//
// The console computes the record path, the log path and the credential location
// from its own --config-dir, draws all three on the screen, and then used to spawn
// `tossctl soak run` with no flags at all — a child on the default profile.
//
// Measured in production on 2026-08-03: /var/lib/tossos/config/soak.log is nothing
// but "soak: no Open API credentials" repeated, while the credentials sit in
// /var/lib/tossos/config/openapi-credentials.json. The button had never once
// worked in the container.
func TestTheSoakSpawnCarriesThisProfile(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{}
	f.install(t)

	if _, err := restartSoak(&rootOptions{
		configDir:   "/var/lib/tossos/config",
		sessionFile: "/run/tossos/session.json",
	}, record); err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if len(f.spawnArgs) != 1 {
		t.Fatalf("spawned %d surveys, want one", len(f.spawnArgs))
	}
	args := strings.Join(f.spawnArgs[0], " ")
	for _, want := range []string{
		"--config-dir /var/lib/tossos/config",
		"--session-file /run/tossos/session.json",
		"soak run",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("the spawned survey did not inherit the profile: %q is missing from %q",
				want, args)
		}
	}
}

// TestTheSoakPatternMatchesWhatTheConsoleSpawns (change a060) replaces
// TestTheProcessPatternMatchesTheAutostartScript, which compared the constant with
// a literal this test file wrote down itself:
//
//	if soakProcessPattern != "tossctl soak run" { … }
//
// tools/soak-autostart.sh was never in this repository, and its installed copy was
// retired on 2026-08-03 (a060 I3). So that assertion had no durable second half to
// check against and could only ever report that somebody changed the value, never
// that the value was right. The engine's pattern passed three tests of that family
// while being wrong (a059).
//
// This binds the pattern to the thing it has to match: the command line soakArgs
// builds. Break the spawn or break the pattern and one of them fails.
func TestTheSoakPatternMatchesWhatTheConsoleSpawns(t *testing.T) {
	pattern, err := regexp.Compile(soakProcessPattern)
	if err != nil {
		t.Fatalf("soakProcessPattern is not a valid expression: %v", err)
	}
	for _, tc := range []struct {
		name string
		root *rootOptions
	}{
		{"no flags — the default-profile invocation", nil},
		{"a container console", &rootOptions{
			configDir:   "/var/lib/tossos/config",
			sessionFile: "/run/tossos/session.json",
		}},
		{"config dir only", &rootOptions{configDir: "/var/lib/tossos/config"}},
	} {
		command := "/usr/local/bin/tossctl " + strings.Join(soakArgs(tc.root), " ")
		if !pattern.MatchString(command) {
			t.Errorf("%s: the console cannot find the survey it spawns\n  pattern: %s\n  command: %s",
				tc.name, soakProcessPattern, command)
		}
	}
}

// TestTheSoakPatternIgnoresTheOtherSubcommands. restartSoak sends SIGINT to what
// this matches, and the engine is the process that must never be in that list.
func TestTheSoakPatternIgnoresTheOtherSubcommands(t *testing.T) {
	pattern := regexp.MustCompile(soakProcessPattern)
	for _, command := range []string{
		"/usr/local/bin/tossctl --config-dir /var/lib/tossos/config --session-file " +
			"/run/tossos/session.json engine run",
		"/usr/local/bin/tossctl --config-dir /var/lib/tossos/config console --port 37085",
		"/usr/local/bin/tossctl --config-dir /var/lib/tossos/config httpapi --port 37086",
		"/usr/local/bin/tossctl --config-dir /srv/mysoak runtime console",
	} {
		if pattern.MatchString(command) {
			t.Errorf("the soak pattern matches a command that is not a survey:\n  %s", command)
		}
	}
}

// TestOnlyThisRecordsSoakIsFound (change a060). Widening the pattern lets this
// console see every profile's survey, and a host shares its PID namespace with its
// containers. A survey's identity is the record it appends to — soakproc.go's own
// header says two surveys on one record is the thing to avoid — so that is what
// ownership is judged on.
func TestOnlyThisRecordsSoakIsFound(t *testing.T) {
	ours := "/var/lib/tossos/config/capability-soak.jsonl"
	got := pidsOwnedBy([]string{
		"31 /usr/local/bin/tossctl --config-dir /var/lib/tossos/config --session-file " +
			"/run/tossos/session.json soak run",
		"4242 /usr/local/bin/tossctl --config-dir /tmp/other-profile soak run",
		"4243 /usr/local/bin/tossctl --config-dir=/var/lib/tossos/config soak run",
		"4244 notapid",
		"4245 /usr/local/bin/tossctl --config-dir /var/lib/tossos/config engine run",
	}, soakProcessMatcher, ours, soakRecordForConfigDir)

	if want := []int{31, 4243}; !pidsEqual(got, want) {
		t.Errorf("pidsOwnedBy = %v, want %v — this console must find its own survey and "+
			"only its own, and never the engine", got, want)
	}
}

// TestTheRestartDoesNotSignalAnotherRecordsSoak runs the real discovery against two
// surveys. Interrupting somebody else's survey costs them a cycle and their record
// its continuity.
func TestTheRestartDoesNotSignalAnotherRecordsSoak(t *testing.T) {
	dir := t.TempDir()
	record := filepath.Join(dir, "capability-soak.jsonl")
	f := &procFakes{aliveFor: map[int]int{31: 1, 4242: 1}}
	f.install(t)

	prev := soakListProcesses
	soakListProcesses = func() ([]string, error) {
		return []string{
			"31 /usr/local/bin/tossctl --config-dir " + dir + " soak run",
			"4242 /usr/local/bin/tossctl --config-dir /tmp/other-profile soak run",
		}, nil
	}
	prevFind := soakFindProcesses
	soakFindProcesses = pgrepSoak
	t.Cleanup(func() { soakListProcesses, soakFindProcesses = prev, prevFind })

	if _, err := restartSoak(&rootOptions{configDir: dir}, record); err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if !pidsEqual(f.signalled, []int{31}) {
		t.Fatalf("signalled %v, want [31] — 4242 appends to another record", f.signalled)
	}
}

// TestTheRestartAsksAboutThisConsolesRecord pins the wiring the two tests above
// rely on.
func TestTheRestartAsksAboutThisConsolesRecord(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{}
	f.install(t)

	if _, err := restartSoak(&rootOptions{configDir: filepath.Dir(record)}, record); err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
	if len(f.findRecords) != 1 || f.findRecords[0] != record {
		t.Errorf("the lookup was asked about %v, want [%s]", f.findRecords, record)
	}
}
