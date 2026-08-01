package journal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV13ToV14IsAdditiveAndPreservesLegacyRiskIntentBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 13)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	legacy := RiskIntent{AccountRef: "acct", Market: "KR", Symbol: "005930", Side: "BUY", Quantity: "1", EntryPrice: "100", StopPrice: "99", TargetPrice: "103", PolicyVersion: "legacy-v1"}
	canonicalBefore, err := legacy.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	// Historical transition test: stop at v14 so a later additive migration
	// cannot silently turn this into a different transition.
	j := openJournalAtSchema(t, path, 14)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 14 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	if after := countRows(t, j.db, v8AllTables); !sameCounts(before, after) {
		t.Fatalf("rows changed before=%v after=%v", before, after)
	}
	canonicalAfter, err := legacy.Canonical()
	if err != nil || canonicalAfter != canonicalBefore {
		t.Fatalf("RiskIntent canonical changed: before=%q after=%q err=%v", canonicalBefore, canonicalAfter, err)
	}
	for _, name := range []string{"strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage", "strategy_attempt_refusals", "idx_strategy_execution_reverse", "strategy_decision_lineage_no_update", "strategy_attempt_lineage_update_guard"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("artifact %s count=%d err=%v", name, count, err)
		}
	}
}
func TestFailedV14MigrationRollsBackAndBacksUpV13(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 13)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	old.Close()
	broken := append(migrationsThrough(13), migration{Version: 14, SQL: schemaV14 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 14}})
	if err == nil {
		t.Fatal("broken v14 opened")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 || !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("backup/error=%v/%v", backups, err)
	}
	survivor := openJournalAtSchema(t, path, 13)
	defer survivor.Close()
	if got := countRows(t, survivor.db, v8AllTables); !sameCounts(got, before) {
		t.Fatalf("survivor rows=%v want=%v", got, before)
	}
	var count int
	if err := survivor.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='strategy_decision_lineage'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial table count=%d err=%v", count, err)
	}
}
