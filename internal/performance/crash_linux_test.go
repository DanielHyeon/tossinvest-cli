//go:build linux

package performance

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const performanceCrashPath = "TOSSOS_TEST_PERFORMANCE_CRASH_PATH"
const performanceCrashMode = "TOSSOS_TEST_PERFORMANCE_CRASH_MODE"
const performanceCrashPhase = "TOSSOS_TEST_PERFORMANCE_CRASH_PHASE"

func TestPerformanceMigrationAndAppendSIGKILLPhasesAreAllOrNone(t *testing.T) {
	if path, mode, phase := os.Getenv(performanceCrashPath), os.Getenv(performanceCrashMode), os.Getenv(performanceCrashPhase); path != "" && mode != "" && phase != "" {
		killAt := phase
		transactionPhaseHook = func(got string) {
			if got == killAt {
				process, _ := os.FindProcess(os.Getpid())
				_ = process.Kill()
			}
		}
		if mode == "migration" {
			_, _ = Open(path)
			return
		}
		store, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		transactionPhaseHook = func(got string) {
			if got == killAt {
				process, _ := os.FindProcess(os.Getpid())
				_ = process.Kill()
			}
		}
		at := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		trade := measuredTrade(at)
		_, _ = store.Collect(context.Background(), trade, []Observation{{
			ID: "phase-observation", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute),
			Price: "105", Source: "existing-position", SourceVersion: "v1",
		}}, at.Add(time.Hour))
		return
	}

	for _, test := range []struct {
		mode   string
		phases []string
	}{
		{mode: "migration", phases: []string{"migration_after_schema", "migration_after_version"}},
		{mode: "append", phases: []string{"collect_after_trade", "collect_after_observations", "collect_after_snapshot"}},
	} {
		for _, phase := range test.phases {
			t.Run(test.mode+"/"+phase, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "performance.db")
				if test.mode == "append" {
					store, err := Open(path)
					if err != nil {
						t.Fatal(err)
					}
					if err := store.Close(); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestPerformanceMigrationAndAppendSIGKILLPhasesAreAllOrNone$")
				cmd.Env = append(os.Environ(), performanceCrashPath+"="+path, performanceCrashMode+"="+test.mode, performanceCrashPhase+"="+phase)
				if err := cmd.Run(); err == nil {
					t.Fatal("crash child exited cleanly")
				}

				raw, err := sql.Open("sqlite", "file:"+path)
				if err != nil {
					t.Fatal(err)
				}
				if test.mode == "migration" {
					var version, tables int
					if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
						t.Fatal(err)
					}
					if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('performance_trades','price_observations','measurement_snapshots','metric_observations')`).Scan(&tables); err != nil {
						t.Fatal(err)
					}
					if version != 0 || tables != 0 {
						t.Fatalf("partial migration version=%d tables=%d", version, tables)
					}
				} else {
					for _, table := range []string{"performance_trades", "price_observations", "measurement_snapshots", "metric_observations"} {
						var count int
						if err := raw.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
							t.Fatalf("partial append table=%s count=%d err=%v", table, count, err)
						}
					}
				}
				if err := raw.Close(); err != nil {
					t.Fatal(err)
				}
				store, err := Open(path)
				if err != nil {
					t.Fatalf("reopen after crash: %v", err)
				}
				_ = store.Close()
			})
		}
	}
}

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
