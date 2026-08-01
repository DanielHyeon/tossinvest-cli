package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
)

func TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.db")
	commander, err := newConsoleOptimizationCommander(context.Background(), journalPath)
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
