package journal

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// migration_v5_test.go covers task 0.2: the automatic pre-migration backup, the
// restore path after a migration that dies partway, and the two schema
// transitions that matter to a live account — v4 forward to v5 with every row
// preserved, and an older build refusing a v5 journal instead of misreading it.

// migrationTestInstant is the fake clock every migration test opens with, so the
// backup file names in the assertions are exact rather than approximate.
var migrationTestInstant = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

// migrationsThrough is the forward-step list an older build would have shipped.
func migrationsThrough(version int) []migration {
	var steps []migration
	for _, m := range migrations {
		if m.Version <= version {
			steps = append(steps, m)
		}
	}
	return steps
}

// openJournalAtSchema opens a journal and stops the migration at an older
// version, which is how these tests get a genuine v4 database on disk instead of
// a hand-written approximation of one.
func openJournalAtSchema(t *testing.T, path string, version int) *Journal {
	t.Helper()
	j, err := Open(context.Background(), Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(version), target: version},
	})
	if err != nil {
		t.Fatalf("Open at schema v%d: %v", version, err)
	}
	return j
}

// v4Tables are every table a v4 journal holds. The migration must not lose a row
// from any of them: they are the audit trail of a live account's mutations.
var v4Tables = []string{
	"alert_outbox", "attempt_transitions", "fill_events", "fill_snapshots",
	"flatten_sagas", "flatten_steps", "intents", "lineage_edges", "mutation_attempts",
}

// seedV4Rows writes one row into every v4 table.
func seedV4Rows(t *testing.T, j *Journal) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`INSERT INTO intents
		   (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
		    time_in_force, quantity, price, currency, source, fingerprint, notes)
		 VALUES ('intent-1','2026-03-30T00:30:00Z','us','2026-03-30','acct-1','AAPL','BUY',
		         'LIMIT','DAY','10','200.5','USD','engine','fp-1','v4 row')`,
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		 VALUES ('attempt-1','intent-1','PLACE','RECORDED',1,'fp-1','2026-03-30T00:30:00Z')`,
		`INSERT INTO attempt_transitions (attempt_id, from_state, to_state, at)
		 VALUES ('attempt-1','','RECORDED','2026-03-30T00:30:00Z')`,
		`INSERT INTO lineage_edges
		   (parent_order_id, child_order_id, relation, parent_filled_quantity,
		    requested_quantity, intent_id, attempt_id, created_at)
		 VALUES ('order-1','order-2','replaces','3','7','intent-1','attempt-1',
		         '2026-03-30T00:30:00Z')`,
		`INSERT INTO fill_snapshots (order_id, symbol, filled_quantity, average_price, committed_at)
		 VALUES ('order-1','AAPL','3','200.5','2026-03-30T00:30:00Z')`,
		`INSERT INTO fill_events
		   (order_id, symbol, delta_quantity, cumulative_quantity, average_price, committed_at)
		 VALUES ('order-1','AAPL','3','3','200.5','2026-03-30T00:30:00Z')`,
		`INSERT INTO alert_outbox (event_key, event_type, severity, state, created_at)
		 VALUES ('key-1','IN_DOUBT','critical','PENDING','2026-03-30T00:30:00Z')`,
		`INSERT INTO flatten_sagas (id, account_ref, phase, started_at, updated_at)
		 VALUES ('saga-1','acct-1','CANCELLING','2026-03-30T00:30:00Z','2026-03-30T00:30:00Z')`,
		`INSERT INTO flatten_steps
		   (saga_id, kind, symbol, target_order_id, state, created_at, updated_at)
		 VALUES ('saga-1','CANCEL','AAPL','order-1','PENDING','2026-03-30T00:30:00Z',
		         '2026-03-30T00:30:00Z')`,
	}
	for _, stmt := range stmts {
		if _, err := j.db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seeding a v4 row: %v\n%s", err, stmt)
		}
	}
}

// countRows reports the row count of every v4 table.
func countRows(t *testing.T, db *sql.DB, tables []string) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, table := range tables {
		var n int
		// Table names come from this file, never from input.
		if err := db.QueryRowContext(context.Background(),
			"SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Fatalf("counting %s: %v", table, err)
		}
		out[table] = n
	}
	return out
}

// TestMigrationV4ToV5PreservesEveryRow is the transition contract: an operator
// upgrading a journal that holds a live account's history keeps all of it, and
// the new tables arrive empty.
func TestMigrationV4ToV5PreservesEveryRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 4)
	seedV4Rows(t, old)
	before := countRows(t, old.db, v4Tables)
	var intentNotes string
	if err := old.db.QueryRowContext(ctx,
		"SELECT notes FROM intents WHERE id = 'intent-1'").Scan(&intentNotes); err != nil {
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

	after := countRows(t, j.db, v4Tables)
	for _, table := range v4Tables {
		if before[table] != after[table] {
			t.Errorf("%s rows: %d before the migration, %d after", table, before[table], after[table])
		}
		if after[table] == 0 {
			t.Errorf("%s was empty before the migration; the fixture proves nothing", table)
		}
	}

	// Historical rows are not rewritten, only extended.
	var notes, filledAmount string
	if err := j.db.QueryRowContext(ctx,
		"SELECT notes FROM intents WHERE id = 'intent-1'").Scan(&notes); err != nil {
		t.Fatal(err)
	}
	if notes != intentNotes {
		t.Errorf("intent notes = %q, want the pre-migration value %q", notes, intentNotes)
	}
	if err := j.db.QueryRowContext(ctx,
		"SELECT filled_amount FROM fill_snapshots WHERE order_id = 'order-1'").Scan(&filledAmount); err != nil {
		t.Fatalf("the v4 snapshot must read back through the new column: %v", err)
	}
	if filledAmount != "" {
		t.Errorf("filled_amount on a migrated v4 row = %q, want the empty string", filledAmount)
	}

	// The v5 tables exist and start empty: the migration invents no decisions.
	for _, table := range []string{
		"decisions", "risk_reservations", "spent_nonces", "reconcile_states",
		"execution_corrections",
	} {
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

// TestOlderBuildRefusesTheV5Journal is the rollback story stated as a test: there
// is no down-migration, so the previous build must refuse the file rather than
// read a v5 journal as if the columns it does not know about were absent.
func TestOlderBuildRefusesTheV5Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	seedV4Rows(t, j)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	// The same code with the previous build's understanding of the schema.
	_, err := Open(context.Background(), Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(4), target: 4},
	})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("an older build opening a v5 journal: want ErrSchemaTooNew, got %v", err)
	}
	if !strings.Contains(err.Error(), "5") || !strings.Contains(err.Error(), "4") {
		t.Errorf("the refusal must name both versions: %v", err)
	}

	// Refusing must not have touched the file: it is still a v5 journal with its
	// rows, and no backup was taken for a migration that never ran.
	reopened := openTestJournalAt(t, path)
	if got := countRows(t, reopened.db, []string{"intents"})["intents"]; got != 1 {
		t.Fatalf("intents after the refusal = %d, want 1", got)
	}
	if backups := backupsIn(t, filepath.Dir(path)); len(backups) != 0 {
		t.Errorf("a refused open must not create a backup, found %v", backups)
	}
}

// TestMigrationBacksUpBeforeApplying pins task 0.2's first half: the copy is
// taken before the first step runs, next to the journal, named after the
// versions and the instant, and readable as the old schema.
func TestMigrationBacksUpBeforeApplying(t *testing.T) {
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

	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want exactly one pre-migration backup, got %v", backups)
	}
	backup := backups[0]
	if want := "journal.db.v4-pre-v5.20260330T003000Z.bak"; filepath.Base(backup) != want {
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

	// The live journal records where the copy went, so an operator recovering
	// after a failure does not have to guess.
	var recorded string
	if err := j.db.QueryRowContext(ctx,
		"SELECT value FROM schema_meta WHERE key = 'pre_migration_backup_path'").Scan(&recorded); err != nil {
		t.Fatalf("schema_meta must record the backup path: %v", err)
	}
	if recorded != backup {
		t.Errorf("recorded backup path = %q, want %q", recorded, backup)
	}

	// The copy is the pre-migration journal: version 4, all rows, self-contained
	// (no -wal beside it) and intact.
	assertBackupIsV4(t, backup, before)
}

// TestFreshJournalIsNotBackedUp keeps the data directory from filling with
// copies of nothing: a database created by this run has no history to lose.
func TestFreshJournalIsNotBackedUp(t *testing.T) {
	dir := t.TempDir()
	openTestJournalAt(t, filepath.Join(dir, "journal.db"))
	if backups := backupsIn(t, dir); len(backups) != 0 {
		t.Fatalf("a first-run migration must not leave a backup, found %v", backups)
	}
}

// TestFailedMigrationLeavesTheJournalRestorable is the failure contract. A step
// that dies partway must leave (a) the live journal exactly as it was, still
// openable by the build that wrote it, and (b) a backup that restores to the
// same thing — because the additive rules give us no down-migration, only the
// copy.
func TestFailedMigrationLeavesTheJournalRestorable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 4)
	seedV4Rows(t, old)
	before := countRows(t, old.db, v4Tables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	// A v5 step that creates one table and then fails: the half that succeeded
	// must not survive the transaction.
	broken := append(migrationsThrough(4), migration{
		Version: 5,
		SQL: `CREATE TABLE decisions (id TEXT PRIMARY KEY) STRICT;
		      INSERT INTO table_that_does_not_exist (x) VALUES (1);`,
	})
	_, err := Open(ctx, Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 5},
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

	// (a) The live journal is untouched: still v4, still all rows, no half-built
	// table, and not corrupt.
	survivor := openJournalAtSchema(t, path, 4)
	defer survivor.Close()
	if got := countRows(t, survivor.db, v4Tables); !sameCounts(got, before) {
		t.Errorf("rows after the failed migration = %v, want %v", got, before)
	}
	var halfBuilt int
	if err := survivor.db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='decisions'").Scan(&halfBuilt); err != nil {
		t.Fatal(err)
	}
	if halfBuilt != 0 {
		t.Error("the failed step's first statement must have rolled back with the transaction")
	}
	version, err := survivor.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 4 {
		t.Errorf("schema version after the failed migration = %d, want 4", version)
	}
	if err := survivor.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after the failed migration: %v", err)
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	// (b) The documented restore works: put the backup in place of the live
	// files and the journal comes back with its history, then migrates forward.
	assertBackupIsV4(t, backups[0], before)
	restoreBackup(t, backups[0], path)

	restored := openTestJournalAt(t, path)
	if got := countRows(t, restored.db, v4Tables); !sameCounts(got, before) {
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

// TestBackupNeverOverwritesAnExistingBackup: the name carries a timestamp, but
// two migrations within the same second (or a retried upgrade under a fake
// clock) must not destroy the earlier copy — the earlier one may be the only
// good state left.
func TestBackupNeverOverwritesAnExistingBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")

	old := openJournalAtSchema(t, path, 4)
	seedV4Rows(t, old)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	occupied := filepath.Join(dir, "journal.db.v4-pre-v5.20260330T003000Z.bak")
	if err := os.WriteFile(occupied, []byte("an earlier backup"), 0o600); err != nil {
		t.Fatal(err)
	}

	openTestJournalAt(t, path)

	kept, err := os.ReadFile(occupied)
	if err != nil {
		t.Fatal(err)
	}
	if string(kept) != "an earlier backup" {
		t.Fatal("the existing backup was overwritten")
	}
	backups := backupsIn(t, dir)
	if len(backups) != 2 {
		t.Fatalf("want the existing backup plus a new one, got %v", backups)
	}
}

// backupsIn lists the pre-migration backups in a directory.
func backupsIn(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.bak"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(matches)
	return matches
}

// assertBackupIsV4 opens the backup read-only, without the journal package's own
// migration, and checks it is the pre-migration database.
func assertBackupIsV4(t *testing.T, backup string, want map[string]int) {
	t.Helper()
	ctx := context.Background()

	for _, sidecar := range []string{backup + "-wal", backup + "-shm"} {
		if _, err := os.Stat(sidecar); err == nil {
			t.Errorf("the backup must be self-contained, found %s", sidecar)
		}
	}

	db, err := sql.Open("sqlite", "file:"+backup+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 4 {
		t.Errorf("backup schema version = %d, want 4 (the pre-migration state)", version)
	}
	var integrity string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(strings.TrimSpace(integrity), "ok") {
		t.Errorf("backup integrity_check = %q", integrity)
	}
	if got := countRows(t, db, v4Tables); !sameCounts(got, want) {
		t.Errorf("backup rows = %v, want %v", got, want)
	}
	var hasDecisions int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='decisions'").Scan(&hasDecisions); err != nil {
		t.Fatal(err)
	}
	if hasDecisions != 0 {
		t.Error("the backup was taken after the migration started")
	}
}

// restoreBackup performs the documented restore: move the live files aside, put
// the backup in their place.
func restoreBackup(t *testing.T, backup, path string) {
	t.Helper()
	for _, live := range []string{path, path + "-wal", path + "-shm"} {
		if _, err := os.Stat(live); err == nil {
			if err := os.Rename(live, live+".aside"); err != nil {
				t.Fatal(err)
			}
		}
	}
	data, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func sameCounts(got, want map[string]int) bool {
	if len(got) != len(want) {
		return false
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}
