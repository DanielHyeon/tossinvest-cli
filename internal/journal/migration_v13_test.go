package journal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV12ToV13IsAdditiveAndPreservesExistingRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 12)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != SchemaVersion {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	if after := countRows(t, j.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed: before=%v after=%v", before, after)
	}
	for _, name := range []string{"position_policy_lifecycles", "protection_sagas", "protection_mutation_attempts", "idx_protection_sagas_live_claim"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v12/v13 artifact %s count=%d err=%v", name, count, err)
		}
	}
}

func TestFailedV13MigrationRollsBackTablesAndUserVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 12)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(12), migration{Version: 13, SQL: schemaV13 + `
		PRAGMA user_version = 13;
		INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 13},
	})
	if err == nil {
		t.Fatal("broken v13 migration unexpectedly opened")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 || !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("backup/error=%v/%v", backups, err)
	}

	survivor := openJournalAtSchema(t, path, 12)
	if version, versionErr := survivor.SchemaVersion(context.Background()); versionErr != nil || version != 12 {
		t.Fatalf("survivor version=%d err=%v", version, versionErr)
	}
	if got := countRows(t, survivor.db, v8AllTables); !sameCounts(got, before) {
		t.Fatalf("survivor rows=%v want=%v", got, before)
	}
	var count int
	if err := survivor.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='protection_sagas'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial v13 table survived: count=%d err=%v", count, err)
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}
	assertBackupAtVersion(t, backups[0], 12, before, "protection_sagas")
}
