package journal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 11)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 12 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	if after := countRows(t, j.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed: before=%v after=%v", before, after)
	}
	for _, name := range []string{"idx_exit_events_proposed_intent", "position_policy_lifecycles", "position_policy_events"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("v11/v12 artifact %s missing after migration", name)
		}
	}
}

func TestFailedV12MigrationRollsBackTablesAndUserVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 11)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(11), migration{Version: 12, SQL: `
		ALTER TABLE exit_states ADD COLUMN lifecycle_generation INTEGER;
		ALTER TABLE exit_events ADD COLUMN lifecycle_generation INTEGER;
		CREATE TABLE position_policy_lifecycles(position_id TEXT PRIMARY KEY) STRICT;
		PRAGMA user_version = 12;
		INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 12},
	})
	if err == nil {
		t.Fatal("broken v12 migration unexpectedly opened")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 || !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("backup/error=%v/%v", backups, err)
	}

	survivor := openJournalAtSchema(t, path, 11)
	if version, versionErr := survivor.SchemaVersion(context.Background()); versionErr != nil || version != 11 {
		t.Fatalf("survivor version=%d err=%v", version, versionErr)
	}
	if got := countRows(t, survivor.db, v8AllTables); !sameCounts(got, before) {
		t.Fatalf("survivor rows=%v want=%v", got, before)
	}
	var count int
	if err := survivor.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='position_policy_lifecycles'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partial v12 table survived failed migration")
	}
	for _, table := range []string{"exit_states", "exit_events"} {
		if err := survivor.db.QueryRow(`SELECT count(*) FROM pragma_table_info(?) WHERE name='lifecycle_generation'`, table).
			Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("partial v12 column survived failed migration: %s.lifecycle_generation", table)
		}
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}
	assertBackupAtVersion(t, backups[0], 11, before, "position_policy_lifecycles")

	restoreBackup(t, backups[0], path)
	restored := openTestJournalAt(t, path)
	if version, versionErr := restored.SchemaVersion(context.Background()); versionErr != nil || version != 12 {
		t.Fatalf("restored version=%d err=%v", version, versionErr)
	}
}
