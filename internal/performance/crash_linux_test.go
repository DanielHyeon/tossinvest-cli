//go:build linux

package performance

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const performanceCrashPath = "TOSSOS_TEST_PERFORMANCE_CRASH_PATH"

func TestPerformanceMigrationAndAppendSurviveSIGKILL(t *testing.T) {
	if path := os.Getenv(performanceCrashPath); path != "" {
		store, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.AppendObservations(context.Background(), []Observation{{
			ID: "crash-observation", PositionID: "position-crash",
			At: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Price: "100",
			Source: "existing-position", SourceVersion: "v1",
		}}); err != nil {
			t.Fatal(err)
		}
		process, _ := os.FindProcess(os.Getpid())
		_ = process.Kill()
		return
	}
	path := filepath.Join(t.TempDir(), "performance.db")
	cmd := exec.Command(os.Args[0], "-test.run=^TestPerformanceMigrationAndAppendSurviveSIGKILL$")
	cmd.Env = append(os.Environ(), performanceCrashPath+"="+path)
	if err := cmd.Run(); err == nil {
		t.Fatal("crash child exited cleanly")
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM price_observations WHERE id='crash-observation'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("crash row count=%d err=%v", count, err)
	}
	if version, err := store.SchemaVersion(context.Background()); err != nil || version != 1 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
}
