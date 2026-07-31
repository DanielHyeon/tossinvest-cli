package journal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV10ToV11AddsIntentIndexAndPreservesRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 10)
	insertPosition(t, old, "v11-position", nil)
	insertIntent(t, old, "v11-intent")
	if _, err := old.db.Exec(`INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('v11-position','LEGACY','v11-intent','2026-03-30T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 11 {
		t.Fatalf("schema version = %d, err=%v", version, err)
	}
	var indexSQL string
	if err := j.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='index' AND name=?`,
		"idx_exit_events_proposed_intent").Scan(&indexSQL); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(indexSQL, "proposed_intent_id") {
		t.Fatalf("unexpected index definition: %s", indexSQL)
	}
	var count int
	if err := j.db.QueryRow(`SELECT count(*) FROM exit_events WHERE proposed_intent_id='v11-intent'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("preserved exit event count=%d err=%v", count, err)
	}
}

func TestFailedV11MigrationRollsBackIndexAndVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 10)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(10), migration{Version: 11, SQL: `
		CREATE INDEX idx_exit_events_proposed_intent ON exit_events(proposed_intent_id);
		INSERT INTO table_that_does_not_exist(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 11},
	})
	if err == nil {
		t.Fatal("broken v11 migration unexpectedly opened")
	}

	survivor := openJournalAtSchema(t, path, 10)
	if version, versionErr := survivor.SchemaVersion(context.Background()); versionErr != nil || version != 10 {
		t.Fatalf("survivor version=%d err=%v", version, versionErr)
	}
	var count int
	if err := survivor.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name=?`,
		"idx_exit_events_proposed_intent").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partial v11 index survived rollback")
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("backups=%v", backups)
	}
	restoreBackup(t, backups[0], path)
	restored := openTestJournalAt(t, path)
	if version, versionErr := restored.SchemaVersion(context.Background()); versionErr != nil || version != 11 {
		t.Fatalf("restored version=%d err=%v", version, versionErr)
	}
}
