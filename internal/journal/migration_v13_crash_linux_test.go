//go:build linux

package journal

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

const crashModeMigrationV13AfterCommit = "migration_v13_after_commit"

func TestMigrationV13CommitAndUserVersionSurviveSIGKILL(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeMigrationV13AfterCommit {
		j := openCrashChildJournal()
		if version, err := j.SchemaVersion(context.Background()); err != nil || version != SchemaVersion {
			t.Fatalf("child schema version=%d err=%v", version, err)
		}
		kill()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 12)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	runCrashChild(t, "TestMigrationV13CommitAndUserVersionSurviveSIGKILL", crashModeMigrationV13AfterCommit, path)
	assertCrashJournalArtifacts(t, path)

	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != SchemaVersion {
		t.Fatalf("raw user_version=%d err=%v", version, err)
	}
	for _, name := range []string{"protection_sagas", "protection_mutation_attempts", "idx_protection_sagas_live_claim"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("raw artifact %s count=%d err=%v", name, count, err)
		}
	}

	reopened := openTestJournalAt(t, path)
	if after := countRows(t, reopened.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed across migration crash: before=%v after=%v", before, after)
	}
}
