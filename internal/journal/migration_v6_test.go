package journal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// migration_v6_test.go covers add-core-domain task 0.1's second half: the
// contract around the v5→v6 step itself — the automatic pre-migration backup,
// the restore path after a step that dies partway, the transition with every
// row preserved, and an older build refusing a v6 journal rather than reading
// it as though the columns it does not know about were absent.
//
// It is the same shape as migration_v5_test.go on purpose. The tables change
// every release; what must not change is that an operator upgrading a live
// account's journal keeps all of it and can always get back to the file they
// started with.

// v5Tables is every table a v5 journal holds. The v6 migration must not lose a
// row from any of them: together they are the audit trail of a live account's
// decisions, mutations and fills.
var v5Tables = append(append([]string(nil), v4Tables...),
	"decisions", "risk_reservations", "spent_nonces", "reconcile_states",
	"execution_corrections")

// v6Tables is what this change adds. They must arrive empty: a migration that
// invented a position would be inventing a holding.
var v6Tables = []string{
	"positions", "position_adjustments", "operating_modes",
	"exit_states", "exit_events", "trade_outcomes",
}

// seedV5Rows writes one row into every table v5 added, on top of the v4 rows.
func seedV5Rows(t *testing.T, j *Journal) {
	t.Helper()
	seedV4Rows(t, j)
	ctx := context.Background()
	stmts := []string{
		`INSERT INTO decisions
		   (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
		    risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
		 VALUES ('decision-1','acct-1',0,'EXPOSURE_RAISING','RISK_INTENT','{}','hash-1',
		         'cid-1','{"max_notional":"1000000"}','nonce-1','2026-03-30T00:30:00Z',
		         '2026-03-30T00:31:00Z')`,
		`INSERT INTO risk_reservations
		   (id, decision_id, account_ref, kind, amount, currency, snapshot_as_of, state)
		 VALUES ('res-1','decision-1','acct-1','OPEN_EXPOSURE','700000','KRW',
		         '2026-03-30T00:29:55Z','HELD')`,
		`INSERT INTO spent_nonces (nonce, decision_id, consumed_at)
		 VALUES ('nonce-1','decision-1','2026-03-30T00:30:01Z')`,
		`INSERT INTO reconcile_states (id, account_ref, symbol, cause, evidence, entered_at)
		 VALUES ('rec-1','acct-1','AAPL','QUANTITY_MISMATCH','local 3, broker 4',
		         '2026-03-30T00:30:00Z')`,
		`INSERT INTO execution_corrections
		   (id, account_ref, order_id, prev_avg_price, new_avg_price, prev_filled_amount,
		    new_filled_amount, cumulative_qty, observed_at)
		 VALUES ('corr-1','acct-1','order-1','200.5','200.6','601.5','601.8','3',
		         '2026-03-30T00:30:02Z')`,
	}
	for _, stmt := range stmts {
		if _, err := j.db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seeding a v5 row: %v\n%s", err, stmt)
		}
	}
}

// TestMigrationV5ToV6PreservesEveryRow is the transition contract for this
// change: an operator upgrading a journal that already holds decisions,
// reservations and fills keeps every one of them, and the core-domain tables
// arrive empty.
func TestMigrationV5ToV6PreservesEveryRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 5)
	seedV5Rows(t, old)
	before := countRows(t, old.db, v5Tables)
	var preimage string
	if err := old.db.QueryRowContext(ctx,
		"SELECT risk_preimage FROM decisions WHERE id = 'decision-1'").Scan(&preimage); err != nil {
		t.Fatal(err)
	}
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
	if SchemaVersion != 6 {
		t.Fatalf("this file is the v5→v6 contract; SchemaVersion is %d", SchemaVersion)
	}

	after := countRows(t, j.db, v5Tables)
	for _, table := range v5Tables {
		if before[table] != after[table] {
			t.Errorf("%s rows: %d before the migration, %d after", table, before[table], after[table])
		}
		if after[table] == 0 {
			t.Errorf("%s was empty before the migration; the fixture proves nothing", table)
		}
	}

	// Historical rows are not rewritten. The preimage is the one that matters
	// most: the gateway re-derives a hash from this text, so a migration that
	// touched it would invalidate every decision on disk.
	var kept string
	if err := j.db.QueryRowContext(ctx,
		"SELECT risk_preimage FROM decisions WHERE id = 'decision-1'").Scan(&kept); err != nil {
		t.Fatal(err)
	}
	if kept != preimage {
		t.Errorf("risk_preimage = %q, want the pre-migration value %q", kept, preimage)
	}

	for _, table := range v6Tables {
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

// TestMigrationV4ToV6IsOneRunOfTwoSteps covers the operator who skipped a
// release: v4 straight to head, both steps applied, one backup taken (of the v4
// file, before the first step) and nothing lost.
func TestMigrationV4ToV6IsOneRunOfTwoSteps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 4)
	seedV4Rows(t, old)
	before := countRows(t, old.db, v4Tables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)
	version, err := j.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != SchemaVersion {
		t.Fatalf("schema version = %d, want %d", version, SchemaVersion)
	}
	if got := countRows(t, j.db, v4Tables); !sameCounts(got, before) {
		t.Errorf("rows after the two-step migration = %v, want %v", got, before)
	}

	// One backup, named for the whole run rather than for each step: it is the
	// state to go back to, and there is only one of those.
	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want exactly one pre-migration backup for the run, got %v", backups)
	}
	if want := "journal.db.v4-pre-v6.20260330T003000Z.bak"; filepath.Base(backups[0]) != want {
		t.Errorf("backup name = %q, want %q", filepath.Base(backups[0]), want)
	}
	assertBackupIsV4(t, backups[0], before)
}

// TestOlderBuildRefusesTheV6Journal is the rollback story: there is no
// down-migration, so the previous build must refuse the file rather than read a
// v6 journal as though positions and exit states did not exist.
func TestOlderBuildRefusesTheV6Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	seedV5Rows(t, j)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	// The same code with the previous release's understanding of the schema.
	_, err := Open(context.Background(), Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(5), target: 5},
	})
	if err == nil {
		t.Fatal("a v5 build opening a v6 journal must refuse")
	}
	if !strings.Contains(err.Error(), "6") || !strings.Contains(err.Error(), "5") {
		t.Errorf("the refusal must name both versions: %v", err)
	}

	// The refusal must not have touched the file: it is still a v6 journal with
	// its rows, and no backup was taken for a migration that never ran.
	reopened := openTestJournalAt(t, path)
	if got := countRows(t, reopened.db, []string{"decisions"})["decisions"]; got != 1 {
		t.Fatalf("decisions after the refusal = %d, want 1", got)
	}
	if backups := backupsIn(t, filepath.Dir(path)); len(backups) != 0 {
		t.Errorf("a refused open must not create a backup, found %v", backups)
	}
}

// TestV6MigrationBacksUpBeforeApplying pins the copy: taken before the step
// runs, next to the journal, named after the versions and the injected clock's
// instant, unreadable by anyone but the owner, and recorded in the live
// database so an operator who finds a dead process does not have to guess.
func TestV6MigrationBacksUpBeforeApplying(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 5)
	seedV5Rows(t, old)
	before := countRows(t, old.db, v5Tables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)

	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want exactly one pre-migration backup, got %v", backups)
	}
	backup := backups[0]
	if want := "journal.db.v5-pre-v6.20260330T003000Z.bak"; filepath.Base(backup) != want {
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
	if recordedVersion != "5" {
		t.Errorf("recorded backup version = %q, want \"5\" (the build to restore with)", recordedVersion)
	}

	assertBackupIsV5(t, backup, before)
}

// TestFailedV6MigrationLeavesTheJournalRestorable is the failure contract. A
// step that dies partway must leave (a) the live journal exactly as it was,
// still openable by the build that wrote it, and (b) a backup that restores to
// the same thing — the additive rules give us no down-migration, only the copy.
func TestFailedV6MigrationLeavesTheJournalRestorable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 5)
	seedV5Rows(t, old)
	before := countRows(t, old.db, v5Tables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	// A v6 step that creates one table and then fails: the half that succeeded
	// must not survive the transaction.
	broken := append(migrationsThrough(5), migration{
		Version: 6,
		SQL: `CREATE TABLE positions (id TEXT PRIMARY KEY) STRICT;
		      INSERT INTO table_that_does_not_exist (x) VALUES (1);`,
	})
	_, err := Open(ctx, Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 6},
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
	if !strings.Contains(err.Error(), "restore") {
		t.Errorf("the failure must point at the restore procedure:\n%v", err)
	}

	// (a) The live journal is untouched: still v5, still all rows, no half-built
	// table, and not corrupt.
	survivor := openJournalAtSchema(t, path, 5)
	if got := countRows(t, survivor.db, v5Tables); !sameCounts(got, before) {
		t.Errorf("rows after the failed migration = %v, want %v", got, before)
	}
	var halfBuilt int
	if err := survivor.db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='positions'").Scan(&halfBuilt); err != nil {
		t.Fatal(err)
	}
	if halfBuilt != 0 {
		t.Error("the failed step's first statement must have rolled back with the transaction")
	}
	version, err := survivor.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 5 {
		t.Errorf("schema version after the failed migration = %d, want 5", version)
	}
	if err := survivor.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after the failed migration: %v", err)
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	// (b) The documented restore works: put the backup in place of the live
	// files and the journal comes back with its history, then migrates forward.
	assertBackupIsV5(t, backups[0], before)
	restoreBackup(t, backups[0], path)

	restored := openTestJournalAt(t, path)
	if got := countRows(t, restored.db, v5Tables); !sameCounts(got, before) {
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

// assertBackupIsV5 opens the backup read-only, without the journal package's own
// migration, and checks it is the pre-migration database.
func assertBackupIsV5(t *testing.T, backup string, want map[string]int) {
	t.Helper()
	assertBackupAtVersion(t, backup, 5, want, "positions")
}
