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

// migration_v7_test.go is the v6→v7 step's contract (change
// adopt-external-positions task 1.1): the pre-migration backup, the transition
// with every row preserved, the additive column arriving NULL on rows that
// existed before it, and an older build refusing a v7 journal rather than
// reading a position as unadopted because it cannot see the column that says
// otherwise.
//
// It is deliberately the same shape as migration_v6_test.go. The tables change
// every release; what must not change is that an operator upgrading a live
// account's journal keeps all of it and can always get back to the file they
// started with.

// v6Tables is every table a v6 journal holds. v7 must not lose a row from any of
// them.
var v6AllTables = append(append([]string(nil), v5Tables...), v6Tables...)

// v7Tables is what this change adds. It must arrive empty: a migration that
// invented an adoption would be inventing a stop for somebody's shares.
var v7Tables = []string{"position_adoptions"}

// seedV6Rows writes one row into every table v6 added, on top of the v5 rows.
func seedV6Rows(t *testing.T, j *Journal) {
	t.Helper()
	seedV5Rows(t, j)
	ctx := context.Background()
	stmts := []string{
		`INSERT INTO positions
		   (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		    quantity, avg_price, opened_at)
		 VALUES ('pos-1','acct-1','kr','005930',1,'decision-1','OPEN','10','70000',
		         '2026-03-30T00:30:00Z')`,
		`INSERT INTO position_adjustments
		   (id, position_id, kind, expected_prev_quantity, prev_quantity, new_quantity,
		    prev_avg_price, new_avg_price, broker_as_of, evidence, created_at)
		 VALUES ('adj-1','pos-1','EXTERNAL','10','10','11','70000','70000',
		         '2026-03-30T00:30:01Z','the account holds one more','2026-03-30T00:30:02Z')`,
		`INSERT INTO operating_modes (id, account_ref, mode, cause, actor, created_at)
		 VALUES ('mode-1','acct-1','NORMAL','startup','OPERATOR','2026-03-30T00:30:00Z')`,
		`INSERT INTO exit_states
		   (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		    baseline_price, high_water, ratchet_level, updated_at)
		 VALUES ('pos-1','RATCHET','70000','68000','2000','68000','70000','NONE',
		         '2026-03-30T00:30:00Z')`,
		`INSERT INTO exit_events (position_id, observed_price, high_water, baseline_after,
		    level_after, action, created_at)
		 VALUES ('pos-1','70500','70500','68000','NONE','','2026-03-30T00:30:03Z')`,
		`INSERT INTO trade_outcomes
		   (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		    initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		 VALUES ('pos-1','1000','0.5','2000','10',60,'BREAKEVEN',NULL,
		         '2026-03-30T00:31:00Z')`,
	}
	for _, stmt := range stmts {
		if _, err := j.db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seeding a v6 row: %v\n%s", err, stmt)
		}
	}
}

// TestMigrationV6ToV7PreservesEveryRow is the transition contract: every row of
// a journal that already holds positions, exit states and frozen outcomes
// survives, the adoption table arrives empty, and the column added to a table
// with rows in it reads NULL rather than a value the migration invented.
func TestMigrationV6ToV7PreservesEveryRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 6)
	seedV6Rows(t, old)
	before := countRows(t, old.db, v6AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	version, err := j.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != SchemaVersion {
		t.Fatalf("schema version after upgrade = %d, want %d", version, SchemaVersion)
	}
	if SchemaVersion != 7 {
		t.Fatalf("this file is the v6→v7 contract; SchemaVersion is %d", SchemaVersion)
	}

	after := countRows(t, j.db, v6AllTables)
	for _, table := range v6AllTables {
		if before[table] != after[table] {
			t.Errorf("%s rows: %d before the migration, %d after", table, before[table], after[table])
		}
		if after[table] == 0 {
			t.Errorf("%s was empty before the migration; the fixture proves nothing", table)
		}
	}

	// The additive column on a table that already had rows. NULL is the whole
	// point: a pre-v7 position was never adopted, and a DEFAULT would have made
	// every historical row claim an adoption id nothing points at.
	var adoption any
	if err := j.db.QueryRowContext(ctx,
		"SELECT adoption_id FROM positions WHERE id = 'pos-1'").Scan(&adoption); err != nil {
		t.Fatalf("the v6 position must read back through the new column: %v", err)
	}
	if adoption != nil {
		t.Errorf("adoption_id on a migrated v6 row = %v, want NULL", adoption)
	}

	// Historical rows are not rewritten: the entry decision the position was
	// opened under is untouched, which is what makes "adoption_id is the second
	// justification, not a replacement" true on disk.
	var decision string
	if err := j.db.QueryRowContext(ctx,
		"SELECT entry_decision_id FROM positions WHERE id = 'pos-1'").Scan(&decision); err != nil {
		t.Fatal(err)
	}
	if decision != "decision-1" {
		t.Errorf("entry_decision_id = %q, want the pre-migration value", decision)
	}

	for _, table := range v7Tables {
		var n int
		if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Errorf("new table %s: %v", table, err)
			continue
		}
		if n != 0 {
			t.Errorf("%s has %d rows after the migration, want 0", table, n)
		}
	}
	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after the migration: %v", err)
	}
}

// TestOlderBuildRefusesTheV7Journal is the §0.6 rollback contract. There is no
// down-migration, so the previous build must refuse the file: a v6 binary that
// read a v7 journal would see every adopted position as one with no entry
// decision, decide it is not an exit-policy target, and leave a managed position
// unprotected while reporting it as unmanaged.
func TestOlderBuildRefusesTheV7Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	seedV6Rows(t, j)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := Open(context.Background(), Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(6), target: 6},
	})
	if err == nil {
		t.Fatal("a v6 build opening a v7 journal must refuse")
	}
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Errorf("the refusal must be ErrSchemaTooNew: %v", err)
	}
	if !strings.Contains(err.Error(), "7") || !strings.Contains(err.Error(), "6") {
		t.Errorf("the refusal must name both versions: %v", err)
	}

	// The refusal must not have touched the file, and it must not have taken a
	// backup for a migration that never ran.
	reopened := openTestJournalAt(t, path)
	if got := countRows(t, reopened.db, []string{"positions"})["positions"]; got != 1 {
		t.Fatalf("positions after the refusal = %d, want 1", got)
	}
	if backups := backupsIn(t, filepath.Dir(path)); len(backups) != 0 {
		t.Errorf("a refused open must not create a backup, found %v", backups)
	}
}

// TestV7MigrationBacksUpBeforeApplying pins the copy: taken before the step
// runs, named after the versions and the injected clock's instant, unreadable by
// anyone but the owner, and recorded in the live database.
func TestV7MigrationBacksUpBeforeApplying(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 6)
	seedV6Rows(t, old)
	before := countRows(t, old.db, v6AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)

	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want exactly one pre-migration backup, got %v", backups)
	}
	backup := backups[0]
	if want := "journal.db.v6-pre-v7.20260330T003000Z.bak"; filepath.Base(backup) != want {
		t.Errorf("backup name = %q, want %q (versions + the injected clock's instant)",
			filepath.Base(backup), want)
	}
	info, err := os.Stat(backup)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("backup permissions = %o, want no group/other access", perm)
	}

	var recorded, recordedVersion string
	if err := j.db.QueryRowContext(ctx,
		"SELECT value FROM schema_meta WHERE key = 'pre_migration_backup_path'").Scan(&recorded); err != nil {
		t.Fatalf("schema_meta must record the backup path: %v", err)
	}
	if recorded != backup {
		t.Errorf("recorded backup path = %q, want %q", recorded, backup)
	}
	if err := j.db.QueryRowContext(ctx,
		"SELECT value FROM schema_meta WHERE key = 'pre_migration_backup_version'").Scan(&recordedVersion); err != nil {
		t.Fatalf("schema_meta must record the backed-up version: %v", err)
	}
	if recordedVersion != "6" {
		t.Errorf("recorded backup version = %q, want \"6\" (the build to restore with)", recordedVersion)
	}

	assertBackupAtVersion(t, backup, 6, before, "position_adoptions")
}

// TestFailedV7MigrationLeavesTheJournalRestorable is the failure contract: a
// step that dies partway leaves the live journal exactly as it was and a backup
// that restores to the same thing.
func TestFailedV7MigrationLeavesTheJournalRestorable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 6)
	seedV6Rows(t, old)
	before := countRows(t, old.db, v6AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(6), migration{
		Version: 7,
		SQL: `CREATE TABLE position_adoptions (id TEXT PRIMARY KEY) STRICT;
		      INSERT INTO table_that_does_not_exist (x) VALUES (1);`,
	})
	_, err := Open(ctx, Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 7},
	})
	if err == nil {
		t.Fatal("a failing migration step must fail the open")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want one backup taken before the failed step, got %v", backups)
	}
	if !strings.Contains(err.Error(), backups[0]) {
		t.Errorf("the failure must tell the operator where the backup is:\n%v", err)
	}

	// (a) The live journal is untouched: still v6, still all rows, no half-built
	// table, not corrupt.
	survivor := openJournalAtSchema(t, path, 6)
	if got := countRows(t, survivor.db, v6AllTables); !sameCounts(got, before) {
		t.Errorf("rows after the failed migration = %v, want %v", got, before)
	}
	var halfBuilt int
	if err := survivor.db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='position_adoptions'").
		Scan(&halfBuilt); err != nil {
		t.Fatal(err)
	}
	if halfBuilt != 0 {
		t.Error("the failed step's first statement must have rolled back with the transaction")
	}
	version, err := survivor.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 6 {
		t.Errorf("schema version after the failed migration = %d, want 6", version)
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	// (b) The documented restore works and then migrates forward.
	assertBackupAtVersion(t, backups[0], 6, before, "position_adoptions")
	restoreBackup(t, backups[0], path)

	restored := openTestJournalAt(t, path)
	if got := countRows(t, restored.db, v6AllTables); !sameCounts(got, before) {
		t.Errorf("rows after restoring the backup = %v, want %v", got, before)
	}
	version, err = restored.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != SchemaVersion {
		t.Errorf("the restored journal must migrate forward: version = %d, want %d",
			version, SchemaVersion)
	}
}
