package main

// soakautostart_test.go is written before the code it tests (openspec change
// a101, tasks 2 and 3).
//
// It exists in this shape because runConsole is measured at 0.0% coverage: the
// console's assembly function is not reachable from any test, so a judgement
// placed inside it would be a judgement nobody can exercise. Everything below
// therefore tests functions that take their dependencies as arguments — the same
// reason engineautostart.go gives for its own shape — and runConsole gets only a
// call and a print.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestConfiguredSoakAutostartOffDoesNotStart(t *testing.T) {
	starts := 0
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return false, nil },
		func() (string, error) { starts++; return "started", nil },
	)
	if starts != 0 || note != "" {
		t.Fatalf("starts=%d note=%q, want no action and no output", starts, note)
	}
}

// TestConfiguredSoakAutostartWithoutLoadWiringDoesNothing. A build with no
// config seam has recorded no approval, and inventing one would start a survey
// nobody asked for.
func TestConfiguredSoakAutostartWithoutLoadWiringDoesNothing(t *testing.T) {
	starts := 0
	note := runConfiguredSoakAutostart(nil, func() (string, error) { starts++; return "", nil })
	if starts != 0 || note != "" {
		t.Fatalf("starts=%d note=%q, want no action", starts, note)
	}
}

func TestConfiguredSoakAutostartOnStartsExactlyOnce(t *testing.T) {
	starts := 0
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) { starts++; return "soak을 다시 시작했다.", nil },
	)
	if starts != 1 {
		t.Fatalf("starts=%d, want 1", starts)
	}
	if !strings.Contains(note, "다시 시작했다") {
		t.Fatalf("note=%q, want the seam's own words", note)
	}
}

// TestConfiguredSoakAutostartReadFailureIsFailClosed. An unreadable config is
// not an approval.
func TestConfiguredSoakAutostartReadFailureIsFailClosed(t *testing.T) {
	starts := 0
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return false, errors.New("invalid config") },
		func() (string, error) { starts++; return "", nil },
	)
	if starts != 0 {
		t.Fatalf("starts=%d, want 0", starts)
	}
	if !strings.Contains(note, "invalid config") {
		t.Fatalf("note=%q, want the read failure named", note)
	}
}

// TestConfiguredSoakAutostartSurvivesAStartFailure is the invariant runConsole's
// Function Logic Map names: in this function every optional thing that cannot be
// built costs one line of output and never the console. A survey that will not
// start must come back as a string, not as a reason to have no operator screen.
func TestConfiguredSoakAutostartSurvivesAStartFailure(t *testing.T) {
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) { return "", errors.New("pgrep unavailable") },
	)
	if !strings.Contains(note, "pgrep unavailable") {
		t.Fatalf("note=%q, want the failure reported rather than swallowed", note)
	}
}

// TestConfiguredSoakAutostartKeepsAFailedStartsOwnWords. restartSoak explains
// itself — "pid N이(가) 30s 안에 종료되지 않았다. 새 soak을 시작하지 않았다" — and that
// sentence is the whole diagnosis. Reporting only the error would throw it away.
func TestConfiguredSoakAutostartKeepsAFailedStartsOwnWords(t *testing.T) {
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) {
			return "pid 42이(가) 30s 안에 종료되지 않았다", errors.New("stop timeout")
		},
	)
	for _, want := range []string{"stop timeout", "종료되지 않았다"} {
		if !strings.Contains(note, want) {
			t.Errorf("note=%q, missing %q", note, want)
		}
	}
}

// TestConfiguredSoakAutostartSaysSomethingWhenTheSeamIsSilent. A survey that
// started and said nothing must still produce a line: the operator is being told
// what a process they cannot see did at boot.
func TestConfiguredSoakAutostartSaysSomethingWhenTheSeamIsSilent(t *testing.T) {
	note := runConfiguredSoakAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) { return "   ", nil },
	)
	if !strings.Contains(note, "서베이를 시작했다") {
		t.Errorf("note=%q, want a default statement", note)
	}
}

func TestConfiguredSoakAutostartWithoutStartWiringIsVisible(t *testing.T) {
	note := runConfiguredSoakAutostart(func() (bool, error) { return true, nil }, nil)
	if strings.TrimSpace(note) == "" {
		t.Fatal("an approved autostart with no wiring said nothing")
	}
}

// --- the approval the restart button leaves behind ---------------------------

// TestSoakRestartRecordsTheApproval. The operator pressing restart is the act
// this change persists: without it the intent dies with the container and the
// next deploy silently stops the attestation clock.
func TestSoakRestartRecordsTheApproval(t *testing.T) {
	saved := []bool{}
	note, err := rememberSoakApproval(
		func(on bool) error { saved = append(saved, on); return nil },
		"pid 42에 SIGINT를 보내 정상 종료시킨 뒤 다시 시작했다", nil,
	)
	if err != nil {
		t.Fatalf("rememberSoakApproval: %v", err)
	}
	if len(saved) != 1 || !saved[0] {
		t.Fatalf("saved=%v, want exactly one true", saved)
	}
	if !strings.Contains(note, "다시 시작했다") {
		t.Fatalf("note=%q, want the restart's own words preserved", note)
	}
}

// TestSoakRestartFailureRecordsNothing. There is no survey to keep running, so
// there is nothing to approve.
func TestSoakRestartFailureRecordsNothing(t *testing.T) {
	saved := 0
	restartErr := errors.New("pid 42이(가) 종료되지 않았다")
	_, err := rememberSoakApproval(
		func(bool) error { saved++; return nil }, "", restartErr,
	)
	if !errors.Is(err, restartErr) {
		t.Fatalf("err=%v, want the restart failure unchanged", err)
	}
	if saved != 0 {
		t.Fatalf("saved=%d, want 0", saved)
	}
}

// TestSoakApprovalFailureDoesNotUndoTheRestart is the one that matters most.
//
// The survey is already running by this point. Reporting the bookkeeping failure
// as a restart failure would tell the operator the button did not work, and the
// obvious response — press it again — kills the survey that just started.
func TestSoakApprovalFailureDoesNotUndoTheRestart(t *testing.T) {
	note, err := rememberSoakApproval(
		func(bool) error { return errors.New("config is read-only") },
		"실행 중인 soak을 찾지 못했다. 새로 시작했다", nil,
	)
	if err != nil {
		t.Fatalf("err=%v, want nil — the restart itself succeeded", err)
	}
	if !strings.Contains(note, "새로 시작했다") {
		t.Fatalf("note=%q, want the restart result kept", note)
	}
	if !strings.Contains(note, "config is read-only") {
		t.Fatalf("note=%q, want the bookkeeping failure still visible", note)
	}
}

// TestSoakRestartWithoutASaveSeamStillRestarts. A build with no config seam can
// still bounce the survey; it just cannot remember that it did.
func TestSoakRestartWithoutASaveSeamStillRestarts(t *testing.T) {
	note, err := rememberSoakApproval(nil, "새로 시작했다", nil)
	if err != nil {
		t.Fatalf("err=%v, want nil", err)
	}
	if !strings.Contains(note, "새로 시작했다") {
		t.Fatalf("note=%q", note)
	}
}

// TestConsoleSoakBootSavePersistsAndAuditsTheApproval mirrors the engine seam's
// own test: the key lands in config and the act lands in the audit log.
func TestConsoleSoakBootSavePersistsAndAuditsTheApproval(t *testing.T) {
	configDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	boot := newConsoleSoakBoot(&rootOptions{configDir: configDir})
	if boot == nil {
		t.Fatal("soak boot seam is nil")
	}
	if err := boot.Save(true); err != nil {
		t.Fatalf("Save: %v", err)
	}

	on, err := config.NewService(filepath.Join(configDir, "config.json")).LoadSoakAutostart()
	if err != nil {
		t.Fatalf("LoadSoakAutostart: %v", err)
	}
	if !on {
		t.Fatal("soak.autostart was not persisted")
	}

	// And the boot decision reads back exactly what the button wrote — the two
	// halves of a101 have to meet on the same key.
	if on, err := boot.Load(); err != nil || !on {
		t.Fatalf("boot.Load() = (%v, %v), want (true, nil)", on, err)
	}

	auditBody, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("reading audit: %v", err)
	}
	for _, want := range []string{
		`"action":"soak.autostart"`,
		`"setting":"soak.autostart"`,
		`"old":"false"`,
		`"new":"true"`,
	} {
		if !strings.Contains(string(auditBody), want) {
			t.Fatalf("audit missing %q: %s", want, auditBody)
		}
	}
}
