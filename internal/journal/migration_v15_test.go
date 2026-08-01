package journal

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func seedLegacyV14Outcome(t *testing.T, j *Journal, positionID string) {
	t.Helper()
	insertPosition(t, j, positionID, nil)
	if _, err := j.db.Exec(`INSERT INTO trade_outcomes
		(position_id,realized_pnl_after_costs,realized_r,initial_risk,initial_quantity,
		 held_seconds,exit_ratchet_level,exit_rung,closed_at)
		VALUES (?, '12.5','0.25','5','10',60,NULL,NULL,'2026-03-30T01:00:00Z')`, positionID); err != nil {
		t.Fatalf("seed legacy v14 outcome: %v", err)
	}
}

func TestMigrationV14ToV15IsAdditiveNullableAndIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 14)
	seedLegacyV14Outcome(t, old, "legacy-v14")
	before := schemaObjects(t, old.db)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 15)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 15 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	after := schemaObjects(t, j.db)
	for name, ddl := range before {
		if name == "trade_outcomes" {
			// ALTER TABLE appends the nullable column to this table's stored DDL;
			// every other released object remains byte-identical.
			continue
		}
		if got, exists := after[name]; !exists || got != ddl {
			t.Fatalf("v15 changed released schema object %s: before=%q after=%q", name, ddl, got)
		}
	}
	var cost sql.NullString
	if err := j.db.QueryRow(`SELECT cost_total FROM trade_outcomes WHERE position_id='legacy-v14'`).Scan(&cost); err != nil {
		t.Fatal(err)
	}
	if cost.Valid {
		t.Fatalf("legacy cost_total=%q, want NULL (no invented zero/backfill)", cost.String)
	}
	var notNull int
	var defaultValue sql.NullString
	if err := j.db.QueryRow(`SELECT "notnull", dflt_value FROM pragma_table_info('trade_outcomes') WHERE name='cost_total'`).Scan(&notNull, &defaultValue); err != nil {
		t.Fatal(err)
	}
	if notNull != 0 || defaultValue.Valid {
		t.Fatalf("cost_total notnull=%d default=%+v, want nullable with no default", notNull, defaultValue)
	}
	var triggerCount int
	if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name='trade_outcomes_no_update'`).Scan(&triggerCount); err != nil || triggerCount != 1 {
		t.Fatalf("immutability trigger count=%d err=%v", triggerCount, err)
	}

	// Reopen at the same target: the additive step is not replayed and the row
	// remains byte-for-byte historical, including its NULL measurement.
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openJournalAtSchema(t, path, 15)
	defer reopened.Close()
	if err := reopened.db.QueryRow(`SELECT cost_total FROM trade_outcomes WHERE position_id='legacy-v14'`).Scan(&cost); err != nil || cost.Valid {
		t.Fatalf("idempotent reopen cost=%+v err=%v", cost, err)
	}
}

func TestV15TradeOutcomeAuthorityBytesRejectUpdateButRetentionDeleteWorks(t *testing.T) {
	j := openTestJournal(t)
	seedLegacyV14Outcome(t, j, "frozen-v15")
	if _, err := j.db.Exec(`UPDATE trade_outcomes SET cost_total='0' WHERE position_id='frozen-v15'`); err == nil || !strings.Contains(err.Error(), "trade outcome is immutable") {
		t.Fatalf("authority UPDATE err=%v, want immutability refusal", err)
	}
	if _, err := j.db.Exec(`UPDATE trade_outcomes SET realized_r='999' WHERE position_id='frozen-v15'`); err == nil {
		t.Fatal("released frozen outcome column was mutable")
	}
	result, err := j.db.Exec(`DELETE FROM trade_outcomes WHERE position_id='frozen-v15'`)
	if err != nil {
		t.Fatalf("retention DELETE was blocked: %v", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		t.Fatalf("retention DELETE affected %d rows", rows)
	}
}

func TestFailedV15MigrationRollsBackColumnTriggerAndVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 14)
	seedLegacyV14Outcome(t, old, "legacy-v14")
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(14), migration{Version: 15, SQL: schemaV15 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 15},
	})
	if err == nil {
		t.Fatal("broken v15 migration opened")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 || !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("backup/error=%v/%v", backups, err)
	}

	survivor := openJournalAtSchema(t, path, 14)
	defer survivor.Close()
	if version, versionErr := survivor.SchemaVersion(context.Background()); versionErr != nil || version != 14 {
		t.Fatalf("survivor version=%d err=%v", version, versionErr)
	}
	var columnCount, triggerCount int
	if err := survivor.db.QueryRow(`SELECT count(*) FROM pragma_table_info('trade_outcomes') WHERE name='cost_total'`).Scan(&columnCount); err != nil {
		t.Fatal(err)
	}
	if err := survivor.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='trade_outcomes_no_update'`).Scan(&triggerCount); err != nil {
		t.Fatal(err)
	}
	if columnCount != 0 || triggerCount != 0 {
		t.Fatalf("failed migration leaked column=%d trigger=%d", columnCount, triggerCount)
	}
}
