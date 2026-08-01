//go:build linux

package journal

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const (
	crashModeMigrationV15BeforeVersion = "migration_v15_before_version"
	crashModeMigrationV15AfterVersion  = "migration_v15_after_version"
	crashModeMigrationV15AfterCommit   = "migration_v15_after_commit"
)

func TestMigrationV15TransactionBoundariesSurviveSIGKILL(t *testing.T) {
	mode := os.Getenv(crashEnvMode)
	if mode == crashModeMigrationV15BeforeVersion || mode == crashModeMigrationV15AfterVersion {
		stage := "before_version"
		if mode == crashModeMigrationV15AfterVersion {
			stage = "after_version"
		}
		openV15CrashChild(func(gotStage string, version int) {
			if version == 15 && gotStage == stage {
				kill()
			}
		})
		return
	}
	if mode == crashModeMigrationV15AfterCommit {
		j := openV15CrashChild(nil)
		if version, err := j.SchemaVersion(context.Background()); err != nil || version != 15 {
			t.Fatalf("child version=%d err=%v", version, err)
		}
		kill()
		return
	}

	for _, test := range []struct {
		name          string
		mode          string
		wantVersion   int
		wantCostField bool
	}{
		{name: "pre-commit before version", mode: crashModeMigrationV15BeforeVersion, wantVersion: 14},
		{name: "after user version before commit", mode: crashModeMigrationV15AfterVersion, wantVersion: 14},
		{name: "after commit", mode: crashModeMigrationV15AfterCommit, wantVersion: 15, wantCostField: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.db")
			old := openJournalAtSchema(t, path, 14)
			seedLegacyV14Outcome(t, old, "legacy-v14")
			if err := old.Close(); err != nil {
				t.Fatal(err)
			}

			runCrashChild(t, "TestMigrationV15TransactionBoundariesSurviveSIGKILL", test.mode, path)
			assertCrashJournalArtifacts(t, path)
			raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer raw.Close()
			var version, columnCount, outcomeCount int
			if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != test.wantVersion {
				t.Fatalf("user_version=%d err=%v want=%d", version, err, test.wantVersion)
			}
			if err := raw.QueryRow(`SELECT count(*) FROM pragma_table_info('trade_outcomes') WHERE name='cost_total'`).Scan(&columnCount); err != nil {
				t.Fatal(err)
			}
			if err := raw.QueryRow(`SELECT count(*) FROM trade_outcomes WHERE position_id='legacy-v14'`).Scan(&outcomeCount); err != nil || outcomeCount != 1 {
				t.Fatalf("legacy outcome count=%d err=%v", outcomeCount, err)
			}
			if (columnCount == 1) != test.wantCostField {
				t.Fatalf("cost_total column count=%d wantPresent=%v", columnCount, test.wantCostField)
			}
		})
	}
}

func openV15CrashChild(hook func(string, int)) *Journal {
	path := os.Getenv(crashEnvPath)
	if path == "" {
		os.Exit(2)
	}
	j, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(15), target: 15},
		migrationHook:     hook,
	})
	if err != nil {
		os.Exit(2)
	}
	return j
}
