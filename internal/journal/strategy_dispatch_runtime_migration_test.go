package journal

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV25AddsOnlyDurableStrategyDispatchStateAndPreservesV24(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 24)
	if _, err := old.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,scope_market)
		VALUES('v24-survivor','acct-v24','AAPL','QUANTITY_MISMATCH','survives','2026-03-30T00:00:00Z','US')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(25), target: 25}})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	if version, err := current.SchemaVersion(context.Background()); err != nil || version != 25 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	for _, table := range []string{
		"strategy_dispatch_owner_epochs", "strategy_dispatch_owner_current", "strategy_dispatch_market_authorities",
		"strategy_dispatch_leases", "strategy_dispatch_outcomes",
	} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v25 table %s count=%d err=%v", table, count, err)
		}
	}
	var survivor int
	if err := current.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE id='v24-survivor' AND scope_market='US'`).Scan(&survivor); err != nil || survivor != 1 {
		t.Fatalf("v24 row survivor=%d err=%v", survivor, err)
	}
}

func TestMigrationV25FailureRollsBackEveryDispatchTableAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 24)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(24), migration{Version: 25, SQL: schemaV25 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 25}})
	if err == nil {
		t.Fatal("broken v25 migration succeeded")
	}
	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 24 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	var tables int
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name LIKE 'strategy_dispatch_%'`).Scan(&tables); err != nil || tables != 0 {
		t.Fatalf("rolled-back dispatch tables=%d err=%v", tables, err)
	}
}

func TestReleasedV24BuildRefusesV25Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	current := openJournalAtSchema(t, path, 25)
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(24), target: 24}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("v24 build open error=%v", err)
	}
}
