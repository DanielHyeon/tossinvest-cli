//go:build linux

package journal

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

const crashModeMigrationV14AfterCommit = "migration_v14_after_commit"

func TestMigrationV14CommitAndUserVersionSurviveSIGKILL(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeMigrationV14AfterCommit {
		j := openCrashChildJournalAtVersion(14)
		if version, err := j.SchemaVersion(context.Background()); err != nil || version != 14 {
			t.Fatalf("child version=%d err=%v", version, err)
		}
		kill()
		return
	}
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 13)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	runCrashChild(t, "TestMigrationV14CommitAndUserVersionSurviveSIGKILL", crashModeMigrationV14AfterCommit, path)
	assertCrashJournalArtifacts(t, path)
	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 14 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	for _, name := range []string{"strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage", "strategy_attempt_refusals", "idx_strategy_execution_reverse"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("artifact %s count=%d err=%v", name, count, err)
		}
	}
	reopened := openJournalAtSchema(t, path, 14)
	defer reopened.Close()
	if after := countRows(t, reopened.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed before=%v after=%v", before, after)
	}
}

func openCrashChildJournalAtVersion(version int) *Journal {
	path := os.Getenv(crashEnvPath)
	if path == "" {
		os.Exit(2)
	}
	j, err := Open(context.Background(), Options{
		Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(version), target: version},
	})
	if err != nil {
		os.Exit(2)
	}
	return j
}
