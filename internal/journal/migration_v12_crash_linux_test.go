//go:build linux

package journal

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

const crashModeMigrationV12AfterCommit = "migration_v12_after_commit"

func TestMigrationV12CommitAndUserVersionSurviveSIGKILL(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeMigrationV12AfterCommit {
		j := openCrashChildJournal()
		if version, err := j.SchemaVersion(context.Background()); err != nil || version != 13 {
			t.Fatalf("child schema version=%d err=%v", version, err)
		}
		kill()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 11)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	runCrashChild(t, "TestMigrationV12CommitAndUserVersionSurviveSIGKILL",
		crashModeMigrationV12AfterCommit, path)
	assertCrashJournalArtifacts(t, path)

	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	var version int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 13 {
		t.Fatalf("raw user_version after crash=%d err=%v", version, err)
	}
	for _, name := range []string{"idx_exit_events_proposed_intent", "position_policy_lifecycles", "position_policy_events", "protection_sagas"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("raw artifact %s count=%d err=%v", name, count, err)
		}
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	reopened := openTestJournalAt(t, path)
	if after := countRows(t, reopened.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed across migration crash: before=%v after=%v", before, after)
	}
}
