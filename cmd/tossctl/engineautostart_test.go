package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestConfiguredEngineAutostartOffDoesNotStart(t *testing.T) {
	starts := 0
	note := runConfiguredEngineAutostart(
		func() (bool, error) { return false, nil },
		func() (string, error) { starts++; return "started", nil },
	)
	if starts != 0 || note != "" {
		t.Fatalf("starts=%d note=%q, want no action", starts, note)
	}
}

func TestConfiguredEngineAutostartOnStartsExactlyOnce(t *testing.T) {
	starts := 0
	note := runConfiguredEngineAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) { starts++; return "engine started", nil },
	)
	if starts != 1 {
		t.Fatalf("starts=%d, want 1", starts)
	}
	if !strings.Contains(note, "engine started") {
		t.Fatalf("note=%q", note)
	}
}

func TestConfiguredEngineAutostartReadFailureIsFailClosed(t *testing.T) {
	starts := 0
	note := runConfiguredEngineAutostart(
		func() (bool, error) { return false, errors.New("invalid config") },
		func() (string, error) { starts++; return "", nil },
	)
	if starts != 0 {
		t.Fatalf("starts=%d, want 0", starts)
	}
	for _, want := range []string{"자동 시작 안 함", "invalid config"} {
		if !strings.Contains(note, want) {
			t.Fatalf("note=%q, missing %q", note, want)
		}
	}
}

func TestConfiguredEngineAutostartPreservesStartRefusal(t *testing.T) {
	starts := 0
	note := runConfiguredEngineAutostart(
		func() (bool, error) { return true, nil },
		func() (string, error) {
			starts++
			return "interlock: attestation expired", errors.New("engine refused")
		},
	)
	if starts != 1 {
		t.Fatalf("starts=%d, want 1", starts)
	}
	for _, want := range []string{"engine refused", "attestation expired"} {
		if !strings.Contains(note, want) {
			t.Fatalf("note=%q, missing %q", note, want)
		}
	}
}

func TestConfiguredEngineAutostartWithoutStartWiringIsVisible(t *testing.T) {
	note := runConfiguredEngineAutostart(
		func() (bool, error) { return true, nil },
		nil,
	)
	if !strings.Contains(note, "기동 배선이 없다") {
		t.Fatalf("note=%q", note)
	}
}

func TestConsoleEngineBootSavePersistsAndAuditsTheApproval(t *testing.T) {
	configDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	boot := newConsoleEngineBoot(&rootOptions{configDir: configDir})
	if boot == nil {
		t.Fatal("engine boot seam is nil")
	}
	if err := boot.Save(true); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg, err := config.NewService(filepath.Join(configDir, "config.json")).Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.Autostart {
		t.Fatal("engine.autostart was not persisted")
	}
	auditBody, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("reading audit: %v", err)
	}
	for _, want := range []string{
		`"action":"engine.autostart"`,
		`"setting":"engine.autostart"`,
		`"old":"false"`,
		`"new":"true"`,
	} {
		if !strings.Contains(string(auditBody), want) {
			t.Fatalf("audit missing %q: %s", want, auditBody)
		}
	}
}

func TestFinishConsoleLeavesNativeEngineAlone(t *testing.T) {
	serveErr := errors.New("server closed")
	stops := 0
	got := finishConsole(false, serveErr, func() (string, error) {
		stops++
		return "", nil
	}, &bytes.Buffer{})
	if !errors.Is(got, serveErr) || stops != 0 {
		t.Fatalf("got=%v stops=%d, want original error and no engine stop", got, stops)
	}
}

func TestFinishConsoleStopsContainerEngineAndPrintsItsResult(t *testing.T) {
	var output bytes.Buffer
	stops := 0
	if err := finishConsole(true, nil, func() (string, error) {
		stops++
		return "engine journal closed", nil
	}, &output); err != nil {
		t.Fatalf("finishConsole: %v", err)
	}
	if stops != 1 || !strings.Contains(output.String(), "engine journal closed") {
		t.Fatalf("stops=%d output=%q", stops, output.String())
	}
}

func TestFinishConsoleReturnsStopFailureWhenServerWasClean(t *testing.T) {
	stopErr := errors.New("engine stop timeout")
	var output bytes.Buffer
	got := finishConsole(true, nil, func() (string, error) {
		return "", stopErr
	}, &output)
	if !errors.Is(got, stopErr) || !strings.Contains(output.String(), stopErr.Error()) {
		t.Fatalf("got=%v output=%q", got, output.String())
	}
}

func TestFinishConsolePreservesServerFailureOverStopFailure(t *testing.T) {
	serveErr := errors.New("listen failed")
	stopErr := errors.New("engine stop timeout")
	var output bytes.Buffer
	got := finishConsole(true, serveErr, func() (string, error) {
		return "", stopErr
	}, &output)
	if !errors.Is(got, serveErr) {
		t.Fatalf("got=%v, want serve error %v", got, serveErr)
	}
	if !strings.Contains(output.String(), stopErr.Error()) {
		t.Fatalf("output=%q, want stop failure visible", output.String())
	}
}
