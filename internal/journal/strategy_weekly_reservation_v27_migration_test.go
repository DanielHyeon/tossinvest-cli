package journal

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV27AddsPairedWeeklyAuthorityWithoutChangingV26Rows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 26)
	before := journalV25RowFingerprints(t, old, nil)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	if version, err := current.SchemaVersion(context.Background()); err != nil || version != SchemaVersion {
		t.Fatalf("version=%d err=%v", version, err)
	}
	after := journalV25RowFingerprints(t, current, mapKeys(before))
	for table, want := range before {
		if after[table] != want {
			t.Fatalf("v26 table %s changed: before=%s after=%s", table, want, after[table])
		}
	}
	for _, name := range []string{"strategy_weekly_reservation_scopes", "strategy_weekly_market_reservations", "strategy_weekly_reservation_receipts", "strategy_weekly_first_leg_bindings"} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("table %s count=%d err=%v", name, count, err)
		}
	}
}

func TestMigrationV27FailureRollsBackWeeklyObjectsAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 26)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(26), migration{Version: 27, SQL: schemaV27 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 27}})
	if err == nil {
		t.Fatal("broken v27 migration succeeded")
	}
	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version, objects int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 26 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name LIKE 'strategy_weekly_%'`).Scan(&objects); err != nil || objects != 0 {
		t.Fatalf("rolled-back weekly objects=%d err=%v", objects, err)
	}
}

func TestReleasedV26BuildRefusesV27Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	current := openTestJournalAt(t, path)
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(26), target: 26}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("v26 open error=%v", err)
	}
}
