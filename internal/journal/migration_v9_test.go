package journal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

var v8AllTables = append(append([]string(nil), v7AllTables...), v8Tables...)

func seedV8Rows(t *testing.T, j *Journal) {
	t.Helper()
	seedV7Rows(t, j)
	_, err := j.db.Exec(`
		INSERT INTO entry_decision_observations
		  (id, account_ref, market, symbol, entry_price, cost_scope,
		   cost_model_fingerprint, outcome, observed_at)
		VALUES ('obs-v8','acct-1','kr','005930','70000','FEE_TAX_ONLY',
		        'cost-v1','REFUSED_CHAIN','2026-03-30T00:30:00Z')`)
	if err != nil {
		t.Fatalf("seeding v8 observation: %v", err)
	}
}

func TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 8)
	seedV8Rows(t, old)
	before := countRows(t, old.db, v8AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	version, err := j.SchemaVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if version != 9 {
		t.Fatalf("schema version = %d, want 9", version)
	}
	after := countRows(t, j.db, v8AllTables)
	if !sameCounts(after, before) {
		t.Fatalf("rows changed across v9 migration: before=%v after=%v", before, after)
	}
	for table, column := range map[string]string{
		"exit_states": "policy_id", "position_adoptions": "exit_policy_id",
	} {
		var notNull int
		var defaultValue any
		err := j.db.QueryRow(`SELECT "notnull", dflt_value FROM pragma_table_info(?) WHERE name=?`,
			table, column).Scan(&notNull, &defaultValue)
		if err != nil {
			t.Fatalf("%s.%s: %v", table, column, err)
		}
		if notNull != 0 || defaultValue != nil {
			t.Fatalf("%s.%s is not nullable/no-default", table, column)
		}
	}
}

func TestV8BuildRefusesV9AndV9BacksUpBeforeApplying(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	old := openJournalAtSchema(t, path, 8)
	seedV8Rows(t, old)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 ||
		filepath.Base(backups[0]) != "journal.db.v8-pre-v9.20260330T003000Z.bak" {
		t.Fatalf("v9 backup = %v", backups)
	}
	if info, err := os.Stat(backups[0]); err != nil || info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("v9 backup permissions/info: %v %+v", err, info)
	}

	_, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(8), target: 8},
	})
	if !errors.Is(err, ErrSchemaTooNew) || !strings.Contains(err.Error(), "9") ||
		!strings.Contains(err.Error(), "8") {
		t.Fatalf("v8 refusal = %v", err)
	}
}
