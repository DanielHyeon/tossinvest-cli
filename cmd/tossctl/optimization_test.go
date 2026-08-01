package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

func TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.db")
	commander, err := newConsoleOptimizationCommander(context.Background(), journalPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer commander.Close()
	if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
		t.Fatalf("optimization constructor touched trading journal: %v", err)
	}
	controlPath := filepath.Join(dir, strategyopt.DatabaseFileName)
	info, err := os.Stat(controlPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("control store mode = %o, want 600", info.Mode().Perm())
	}
	view, err := commander.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if view.Evidence.Status != strategyopt.EvidenceUnavailable || view.Snapshot.EffectiveEntry {
		t.Fatalf("unintegrated provider/authority = %+v / %+v", view.Evidence, view.Snapshot)
	}
	field, ok := view.Registry.Field("exit.common-policy")
	if !ok || len(field.Descriptor.Options) != len(exitpolicy.RegisteredCommonPolicies()) {
		t.Fatalf("a041 descriptor adapter = %+v, %v", field, ok)
	}
}

type fixedOptimizationEvidence struct{ evidence strategyopt.Evidence }

func (f fixedOptimizationEvidence) ReadEvidence(context.Context) (strategyopt.Evidence, error) {
	return f.evidence, nil
}

func TestConsoleOptimizationCommanderUsesPerformanceEvidenceProvider(t *testing.T) {
	dir := t.TempDir()
	want := strategyopt.Evidence{Status: strategyopt.EvidenceComplete, Digest: strings.Repeat("a", 64),
		ObservedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}
	commander, err := newConsoleOptimizationCommander(context.Background(), filepath.Join(dir, "journal.db"),
		fixedOptimizationEvidence{evidence: want})
	if err != nil {
		t.Fatal(err)
	}
	defer commander.Close()
	view, err := commander.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if view.Evidence.Status != want.Status || view.Evidence.Digest != want.Digest ||
		!view.Evidence.ObservedAt.Equal(want.ObservedAt) {
		t.Fatalf("evidence=%+v want=%+v", view.Evidence, want)
	}
}

func TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams(t *testing.T) {
	dir := t.TempDir()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	fixture, err := performance.Open(filepath.Join(dir, consolePerformanceDatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.Close(); err != nil {
		t.Fatal(err)
	}
	capabilities, err := openConsolePerformanceCapabilities(dir, func() time.Time { return asOf })
	if err != nil {
		t.Fatal(err)
	}
	if capabilities.Performance == nil || capabilities.Evidence == nil {
		t.Fatalf("capabilities=%+v", capabilities)
	}
	if _, err := os.Stat(filepath.Join(dir, consolePerformanceDatabaseFileName)); err != nil {
		t.Fatalf("profile performance DB not opened: %v", err)
	}
	evidence, err := capabilities.Evidence.ReadEvidence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Status != strategyopt.EvidenceStale || !evidence.ObservedAt.IsZero() || len(evidence.Digest) != 64 {
		t.Fatalf("empty performance DB evidence=%+v", evidence)
	}
	if err := capabilities.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := capabilities.Performance.Dashboard(context.Background(), performance.DefaultQuery(asOf)); err == nil {
		t.Fatal("performance read still succeeds after capability lifecycle close")
	}
}

func TestRunConsoleWiresAndClosesPerformanceWithoutJournalCollection(t *testing.T) {
	source, err := os.ReadFile("console.go")
	if err != nil {
		t.Fatal(err)
	}
	code := string(source)
	for _, want := range []string{
		"openConsolePerformanceCapabilities(filepath.Dir(journalPath), time.Now)",
		"defer performanceCapabilities.Close()",
		"newConsoleOptimizationCommander(ctx, journalPath, performanceCapabilities.Evidence)",
		"Performance:      performanceCapabilities.Performance",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("runConsole production wiring missing %q", want)
		}
	}
	for _, forbidden := range []string{"journal.Open(", "performancejournal.New(", ".Collect(", ".Prune("} {
		if strings.Contains(code, forbidden) {
			t.Errorf("runConsole acquired forbidden journal/collector authority %q", forbidden)
		}
	}
}

func TestConsolePerformanceConstructorUsesReadOnlyOpenOnly(t *testing.T) {
	source, err := os.ReadFile("optimization.go")
	if err != nil {
		t.Fatal(err)
	}
	code := string(source)
	if !strings.Contains(code, "performance.OpenReadOnly(") {
		t.Fatal("console performance constructor does not use the read-only opener")
	}
	for _, forbidden := range []string{"performance.Open(", ".Collect(", ".Prune(", ".AppendTrade(", ".AppendObservations("} {
		if strings.Contains(code, forbidden) {
			t.Errorf("console performance constructor contains writer authority %q", forbidden)
		}
	}
}

func TestConsolePerformanceCapabilitiesFailWithoutPartialReadAuthority(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, consolePerformanceDatabaseFileName)
	capabilities, err := openConsolePerformanceCapabilities(dir, time.Now)
	if !errors.Is(err, performance.ErrDatabaseMissing) || capabilities.Performance != nil || capabilities.Evidence != nil {
		t.Fatalf("missing capabilities=%+v err=%v, want typed missing and no partial capability", capabilities, err)
	}
	if _, statErr := os.Stat(missingPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("console read open created missing performance DB: %v", statErr)
	}
	blocked := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	capabilities, err = openConsolePerformanceCapabilities(blocked, time.Now)
	if err == nil || capabilities.Performance != nil || capabilities.Evidence != nil {
		t.Fatalf("capabilities=%+v err=%v, want no partial capability", capabilities, err)
	}
}
