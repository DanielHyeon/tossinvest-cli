package journal

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV9ToV10IsAdditiveNullableAndPreservesRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 9)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != SchemaVersion {
		t.Fatalf("schema version = %d, err=%v", version, err)
	}
	if after := countRows(t, j.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed: before=%v after=%v", before, after)
	}
	columns := map[string][]string{
		"exit_states": {"snapshot_status", "policy_version", "policy_digest", "snapshot_id", "decision_id",
			"observation_id", "position_generation", "next_target", "next_protection", "last_observation_source",
			"last_observed_at", "snapshot_action", "snapshot_ratio", "projected_quantity", "state_only",
			"suppressed_reason", "effective_snapshot_json"},
		"exit_events": {"position_generation", "policy_id", "policy_version", "policy_digest", "snapshot_id",
			"decision_id", "observation_id", "next_target", "next_protection", "observation_source", "observed_at",
			"projected_quantity", "proposal_ratio", "state_only", "suppressed_reason", "saved_snapshot_json",
			"recomputed_snapshot_json", "effective_snapshot_json", "effective_source", "arm_suppressed_reason"},
	}
	for table, names := range columns {
		for _, name := range names {
			var notNull int
			var defaultValue any
			if err := j.db.QueryRow(`SELECT "notnull",dflt_value FROM pragma_table_info(?) WHERE name=?`, table, name).
				Scan(&notNull, &defaultValue); err != nil {
				t.Fatalf("%s.%s missing: %v", table, name, err)
			}
			if notNull != 0 || defaultValue != nil {
				t.Fatalf("%s.%s must be nullable/no-default", table, name)
			}
		}
	}
}

func TestFailedV10MigrationRollsBackDDLMetadataAndUserVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 9)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(9), migration{Version: 10, SQL: `
		ALTER TABLE exit_states ADD COLUMN snapshot_status TEXT;
		CREATE TABLE exit_snapshot_quarantines(position_id TEXT PRIMARY KEY) STRICT;
		INSERT INTO table_that_does_not_exist(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 10},
	})
	if err == nil {
		t.Fatal("broken v10 migration unexpectedly opened")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 || !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("backup/error = %v / %v", backups, err)
	}
	if !strings.Contains(err.Error(), "Reconcile open orders") {
		t.Fatalf("v10 migration failure omitted broker reconciliation instruction: %v", err)
	}

	survivor := openJournalAtSchema(t, path, 9)
	if version, versionErr := survivor.SchemaVersion(context.Background()); versionErr != nil || version != 9 {
		t.Fatalf("survivor version = %d, err=%v", version, versionErr)
	}
	if got := countRows(t, survivor.db, v8AllTables); !sameCounts(got, before) {
		t.Fatalf("survivor rows = %v, want %v", got, before)
	}
	for _, check := range []struct{ table, column string }{
		{"exit_states", "snapshot_status"}, {"exit_snapshot_quarantines", "position_id"},
	} {
		var count int
		if err := survivor.db.QueryRow(`SELECT count(*) FROM pragma_table_info(?) WHERE name=?`,
			check.table, check.column).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("partial v10 artifact survived: %s.%s", check.table, check.column)
		}
	}
	assertBackupAtVersion(t, backups[0], 9, before, "exit_snapshot_quarantines")
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	// The documented backup-only restore is self-contained: the live DB/WAL/SHM
	// trio is moved aside, the single VACUUM INTO copy is restored, and a fresh
	// head build can migrate it forward without losing the v9 rows.
	restoreBackup(t, backups[0], path)
	restored := openTestJournalAt(t, path)
	if version, versionErr := restored.SchemaVersion(context.Background()); versionErr != nil || version != SchemaVersion {
		t.Fatalf("restored version = %d, err=%v", version, versionErr)
	}
	if got := countRows(t, restored.db, v8AllTables); !sameCounts(got, before) {
		t.Fatalf("restored rows = %v, want %v", got, before)
	}
}

func TestExitSnapshotQuarantineIsGenerationScopedAndReleaseIsCAS(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq, "partial_tuple", "policy_version is NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq+1, "bad", "future"); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("future generation quarantine = %v", err)
	}
	if err := j.ReleaseExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq, q.Version+1,
		QuarantineReleaseHumanRepair, "fixed"); !errors.Is(err, ErrExitSnapshotReleaseStale) {
		t.Fatalf("stale release = %v", err)
	}
	if err := j.ReleaseExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq, q.Version,
		QuarantineReleaseAuthoritativeReconcile, "broker reconciled"); err != nil {
		t.Fatal(err)
	}
	q2, err := j.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq, "again", "new evidence")
	if err != nil {
		t.Fatal(err)
	}
	if q2.Version != q.Version+1 {
		t.Fatalf("quarantine version = %d, want %d", q2.Version, q.Version+1)
	}
}
