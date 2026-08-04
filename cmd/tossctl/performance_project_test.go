package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestProjectLaneAttributionEmptyJournalDoesNotInventAccountOrDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, journal.DBFileName)
	now := time.Date(2026, 8, 4, 4, 0, 0, 0, time.UTC)
	j, err := journal.Open(context.Background(), journal.Options{Path: path, Clock: clock.NewFake(now),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	result, err := projectLaneAttributionOnce(context.Background(), path, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Accounts != 0 || result.EvidenceRows != 0 || !result.CalculatedAt.Equal(now) {
		t.Fatalf("result=%+v", result)
	}
	if _, err := os.Stat(filepath.Join(dir, consolePerformanceDatabaseFileName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty journal created derived database: %v", err)
	}
}

func TestPerformanceProjectorRejectsIntervalBeforeOpeningDatabases(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runPerformanceAttributionProjector(context.Background(), &out, &rootOptions{configDir: dir}, time.Second, time.Now)
	if err == nil || !strings.Contains(err.Error(), "projection interval") {
		t.Fatalf("error=%v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("invalid interval touched journal: %v", statErr)
	}
}

func TestPerformanceProjectionCommandIsSeparateNonTradingLane(t *testing.T) {
	cmd := newPerformanceProjectAttributionCmd(&rootOptions{})
	if cmd.Annotations["mutating"] != "false" || cmd.Annotations["source"] != "official" {
		t.Fatalf("annotations=%v", cmd.Annotations)
	}
	body, err := os.ReadFile("performance_project.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"internal/official", "internal/execgw", "internal/trading", "order place", "AutomationGate"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("projector acquired forbidden capability %q", forbidden)
		}
	}
}
