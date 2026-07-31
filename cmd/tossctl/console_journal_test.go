package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func createConsoleTestJournal(t *testing.T, path string) {
	t.Helper()
	writer, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("create journal %s: %v", path, err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close journal %s: %v", path, err)
	}
}

// TestConsoleJournalPathFollowsTheEngineProfile pins the same active-profile
// rule used by the engine. Merely having another readable database must never
// cause fallback to it.
func TestConsoleJournalPathFollowsTheEngineProfile(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)
	defaultPath := filepath.Join(dataDir, journal.DBFileName)
	createConsoleTestJournal(t, defaultPath)

	configDir := t.TempDir()
	profilePath := filepath.Join(configDir, journal.DBFileName)
	createConsoleTestJournal(t, profilePath)

	t.Run("explicit", func(t *testing.T) {
		got, err := consoleJournalPath(&rootOptions{configDir: configDir})
		if err != nil {
			t.Fatalf("consoleJournalPath: %v", err)
		}
		if got != profilePath {
			t.Fatalf("explicit profile path = %q, want %q", got, profilePath)
		}
		reader, err := journal.OpenReadOnly(context.Background(), journal.ReadOnlyOptions{Path: got})
		if err != nil {
			t.Fatalf("open selected profile read-only: %v", err)
		}
		defer reader.Close()
		if reader.Path() != profilePath {
			t.Fatalf("opened journal = %q, want %q", reader.Path(), profilePath)
		}
	})

	t.Run("explicit missing never falls back", func(t *testing.T) {
		missingDir := filepath.Join(t.TempDir(), "missing-profile")
		got, err := consoleJournalPath(&rootOptions{configDir: missingDir})
		if err != nil {
			t.Fatalf("consoleJournalPath: %v", err)
		}
		if want := filepath.Join(missingDir, journal.DBFileName); got != want {
			t.Fatalf("missing explicit profile path = %q, want %q", got, want)
		}
		if got == defaultPath {
			t.Fatal("missing explicit profile silently fell back to the default journal")
		}
		if _, err := journal.OpenReadOnly(context.Background(), journal.ReadOnlyOptions{Path: got}); !errors.Is(err, journal.ErrJournalMissing) {
			t.Fatalf("missing selected journal = %v, want ErrJournalMissing without fallback", err)
		}
	})

	t.Run("default", func(t *testing.T) {
		got, err := consoleJournalPath(&rootOptions{})
		if err != nil {
			t.Fatalf("consoleJournalPath: %v", err)
		}
		if got != defaultPath {
			t.Fatalf("default profile path = %q, want %q", got, defaultPath)
		}
	})

	t.Run("whitespace is an explicit engine profile", func(t *testing.T) {
		got, err := consoleJournalPath(&rootOptions{configDir: " "})
		if err != nil {
			t.Fatalf("consoleJournalPath: %v", err)
		}
		if want := filepath.Join(" ", journal.DBFileName); got != want {
			t.Fatalf("whitespace profile path = %q, want %q", got, want)
		}
	})
}
