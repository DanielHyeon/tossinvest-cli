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

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// migration_v8_test.go is the v7→v8 step's contract (change add-net-rr-measurement
// task 2.2): the pre-migration backup, the transition with every row preserved,
// the new observation table arriving empty, and an older build refusing a v8
// journal rather than reading it.
//
// It is the same shape as migration_v7_test.go, which is the same shape as v6's.
// The tables change every release; what must not change is that an operator
// upgrading a live account's journal keeps all of it and can always get back to
// the file they started with.
//
// # There is no down-migration, and this change does not invent one
//
// schema.go says it and means it: a down step would have to drop a column, which
// rule 3 forbids. The rollback is "run the previous binary" — pinned by
// TestOlderBuildRefusesTheV8Journal — and the recovery from a step that fails
// partway is the automatic pre-migration backup — pinned by
// TestV8MigrationBacksUpBeforeApplying and
// TestFailedV8MigrationLeavesTheJournalRestorable. Those two paths *are* the
// rollback story, so they are tested rather than described.

// v7AllTables is every table a v7 journal holds. v8 must not lose a row from any
// of them.
var v7AllTables = append(append([]string(nil), v6AllTables...), v7Tables...)

// v8Tables is what this change adds. It must arrive empty: a migration that
// invented an observation would be inventing a measurement nobody took.
var v8Tables = []string{"entry_decision_observations"}

// seedV7Rows writes one row into every table v7 added, on top of the v6 rows.
func seedV7Rows(t *testing.T, j *Journal) {
	t.Helper()
	seedV6Rows(t, j)
	ctx := context.Background()
	stmts := []string{
		`INSERT INTO position_adoptions
		   (id, symbol, market, quantity, cost_basis, cost_basis_src, observed_price,
		    synthetic_stop, observed_at, preimage_digest)
		 VALUES ('adopt-1','005930','kr','10','70000','BROKER_AVG','70500','68385',
		         '2026-03-30T00:30:00Z','digest-1')`,
	}
	for _, stmt := range stmts {
		if _, err := j.db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seeding a v7 row: %v\n%s", err, stmt)
		}
	}
}

// TestMigrationV7ToV8PreservesEveryRow is the transition contract: every row of a
// v7 journal survives, the observation table arrives empty, and the decision
// preimage — the text a hash is taken over — is byte-identical afterwards.
func TestMigrationV7ToV8PreservesEveryRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 7)
	seedV7Rows(t, old)
	before := countRows(t, old.db, v7AllTables)
	var preimage, hash string
	if err := old.db.QueryRowContext(ctx,
		"SELECT risk_preimage, risk_hash FROM decisions WHERE id = 'decision-1'").
		Scan(&preimage, &hash); err != nil {
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
	if SchemaVersion != 8 {
		t.Fatalf("this file is the v7→v8 contract; SchemaVersion is %d", SchemaVersion)
	}

	after := countRows(t, j.db, v7AllTables)
	for _, table := range v7AllTables {
		if before[table] != after[table] {
			t.Errorf("%s rows: %d before the migration, %d after", table, before[table], after[table])
		}
		if after[table] == 0 {
			t.Errorf("%s was empty before the migration; the fixture proves nothing", table)
		}
	}

	// The preimage and its hash are the sealed pair this change is not allowed to
	// touch (design D1). A migration that rewrote either would invalidate every
	// decision on disk, so it is asserted here and not only in the golden test.
	var keptPreimage, keptHash string
	if err := j.db.QueryRowContext(ctx,
		"SELECT risk_preimage, risk_hash FROM decisions WHERE id = 'decision-1'").
		Scan(&keptPreimage, &keptHash); err != nil {
		t.Fatal(err)
	}
	if keptPreimage != preimage {
		t.Errorf("risk_preimage = %q, want the pre-migration value %q", keptPreimage, preimage)
	}
	if keptHash != hash {
		t.Errorf("risk_hash = %q, want the pre-migration value %q", keptHash, hash)
	}

	for _, table := range v8Tables {
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

// TestV8IsPurelyAdditive is task 2.1's "기존 테이블·열·인덱스 무변경" as a
// mechanical check rather than a promise: every schema object a v7 journal
// carries must exist in a v8 journal with byte-identical DDL, and v8 may only
// *add*.
//
// This is stronger than counting rows. A migration that recreated a table with
// the same columns in a different order, or dropped an index, would keep every
// row and still break the readers this change is not supposed to touch.
func TestV8IsPurelyAdditive(t *testing.T) {
	ctx := context.Background()

	v7Path := filepath.Join(t.TempDir(), "journal.db")
	v7 := openJournalAtSchema(t, v7Path, 7)
	oldDDL := schemaObjects(t, v7.db)
	if err := v7.Close(); err != nil {
		t.Fatal(err)
	}

	v8 := openTestJournal(t)
	newDDL := schemaObjects(t, v8.db)

	for name, ddl := range oldDDL {
		got, ok := newDDL[name]
		if !ok {
			t.Errorf("v8 dropped the v7 schema object %q; the additive rules forbid it", name)
			continue
		}
		if got != ddl {
			t.Errorf("v8 changed the DDL of %q:\n v7: %s\n v8: %s", name, ddl, got)
		}
	}

	added := make([]string, 0, len(newDDL))
	for name := range newDDL {
		if _, existed := oldDDL[name]; !existed {
			added = append(added, name)
		}
	}
	sort.Strings(added)
	want := []string{
		"entry_decision_observations",
		"idx_entry_observations_decision",
		"idx_entry_observations_observed",
		"idx_entry_observations_outcome",
	}
	if strings.Join(added, ",") != strings.Join(want, ",") {
		t.Errorf("v8 added %v, want exactly %v", added, want)
	}

	// The observation row must not be reachable from the contract tables by a
	// constraint: decision_id carries no foreign key (D1 / R2-4, the same reason
	// spent_nonces.decision_id does not).
	var fks int
	if err := v8.db.QueryRowContext(ctx,
		"SELECT count(*) FROM pragma_foreign_key_list('entry_decision_observations')").
		Scan(&fks); err != nil {
		t.Fatal(err)
	}
	if fks != 0 {
		t.Errorf("entry_decision_observations declares %d foreign keys, want 0", fks)
	}
}

// schemaObjects maps every user table, index and trigger to its stored DDL.
// Auto-created indexes (PRIMARY KEY, UNIQUE constraints) carry a NULL sql and are
// recorded by name alone, which is enough to notice one disappearing.
func schemaObjects(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT name, COALESCE(sql, '') FROM sqlite_master
		  WHERE type IN ('table','index','trigger','view')
		    AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatalf("reading sqlite_master: %v", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var name, ddl string
		if err := rows.Scan(&name, &ddl); err != nil {
			t.Fatal(err)
		}
		out[name] = ddl
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestOlderBuildRefusesTheV8Journal is the §0.6 rollback contract. There is no
// down-migration, so the previous build must refuse the file rather than read it:
// a v7 binary opening a v8 journal cannot see the observation table, and writing
// entries into a journal whose measurement half it does not maintain would
// produce exactly the silent gap this change exists to close.
func TestOlderBuildRefusesTheV8Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	seedV7Rows(t, j)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := Open(context.Background(), Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(7), target: 7},
	})
	if err == nil {
		t.Fatal("a v7 build opening a v8 journal must refuse")
	}
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Errorf("the refusal must be ErrSchemaTooNew: %v", err)
	}
	if !strings.Contains(err.Error(), "8") || !strings.Contains(err.Error(), "7") {
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

// TestV8MigrationBacksUpBeforeApplying pins the copy: taken before the step runs,
// named after the versions and the injected clock's instant, unreadable by anyone
// but the owner, and recorded in the live database.
func TestV8MigrationBacksUpBeforeApplying(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 7)
	seedV7Rows(t, old)
	before := countRows(t, old.db, v7AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openTestJournalAt(t, path)

	backups := backupsIn(t, dir)
	if len(backups) != 1 {
		t.Fatalf("want exactly one pre-migration backup, got %v", backups)
	}
	backup := backups[0]
	if want := "journal.db.v7-pre-v8.20260330T003000Z.bak"; filepath.Base(backup) != want {
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
	if recordedVersion != "7" {
		t.Errorf("recorded backup version = %q, want \"7\" (the build to restore with)", recordedVersion)
	}

	assertBackupAtVersion(t, backup, 7, before, "entry_decision_observations")
}

// TestFailedV8MigrationLeavesTheJournalRestorable is the failure contract: a step
// that dies partway leaves the live journal exactly as it was and a backup that
// restores to the same thing.
func TestFailedV8MigrationLeavesTheJournalRestorable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 7)
	seedV7Rows(t, old)
	before := countRows(t, old.db, v7AllTables)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	broken := append(migrationsThrough(7), migration{
		Version: 8,
		SQL: `CREATE TABLE entry_decision_observations (id TEXT PRIMARY KEY) STRICT;
		      INSERT INTO table_that_does_not_exist (x) VALUES (1);`,
	})
	_, err := Open(ctx, Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 8},
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

	// (a) The live journal is untouched: still v7, still all rows, no half-built
	// table, not corrupt.
	survivor := openJournalAtSchema(t, path, 7)
	if got := countRows(t, survivor.db, v7AllTables); !sameCounts(got, before) {
		t.Errorf("rows after the failed migration = %v, want %v", got, before)
	}
	var halfBuilt int
	if err := survivor.db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='entry_decision_observations'").
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
	if version != 7 {
		t.Errorf("schema version after the failed migration = %d, want 7", version)
	}
	if err := survivor.Close(); err != nil {
		t.Fatal(err)
	}

	// (b) The documented restore works and then migrates forward.
	assertBackupAtVersion(t, backups[0], 7, before, "entry_decision_observations")
	restoreBackup(t, backups[0], path)

	restored := openTestJournalAt(t, path)
	if got := countRows(t, restored.db, v7AllTables); !sameCounts(got, before) {
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
