package journal

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV22AddsOnlyAuthoritativeRiskBucketJournalState(t *testing.T) {
	j := openTestJournal(t)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 22 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	for _, table := range []string{
		"risk_bucket_policies", "risk_bucket_snapshots", "risk_bucket_final_decisions",
		"risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_orders",
		"risk_bucket_fills", "risk_bucket_fill_allocations", "risk_bucket_events",
		"risk_bucket_state_snapshots", "risk_bucket_scope_latches",
	} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v22 table %s count=%d err=%v", table, count, err)
		}
	}
}

func TestMigrationV22PreservesV21AsUnknownWithoutBackfill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 21)
	seedExistingRiskReservation(t, old, "legacy-existing", "acct-legacy")
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_events"} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("legacy migration invented %s rows=%d err=%v", table, count, err)
		}
	}
	if _, err := current.ReadRiskBucketState(context.Background(), riskBucketOwnerKey("acct-legacy", "legacy")); !errors.Is(err, ErrRiskBucketStateUnknown) {
		t.Fatalf("legacy state error=%v", err)
	}
}

func TestMigrationV22FailureRollsBackEveryTableAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 21)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(21), migration{Version: 22, SQL: schemaV22 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 22}})
	if err == nil {
		t.Fatal("broken v22 migration succeeded")
	}
	raw, openErr := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 21 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	var count int
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name LIKE 'risk_bucket_%'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back v22 artifacts=%d err=%v", count, err)
	}
}
