package journal

// a084: a quarantine records which recovery selector judged it, and a build whose
// selector has changed re-judges it exactly once.
//
// The rows this exists for are real: three ambiguous_recovery quarantines written
// by the comparator a078 replaced, still active, still stopping their positions
// from being judged at all — stop included — because the quarantine cuts in ahead
// of the comparison and the fixed code is never reached (a078 issues.md I1).

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// TestANewQuarantineRecordsTheSelectorThatJudgedIt: without the stamp there is no
// way to tell this build's judgement from its predecessor's.
func TestANewQuarantineRecordsTheSelectorThatJudgedIt(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if q.SelectorRevision != exitpolicy.RecoverySelectorRevision {
		t.Fatalf("selector revision = %d, want the current %d",
			q.SelectorRevision, exitpolicy.RecoverySelectorRevision)
	}

	active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
	if err != nil || !ok {
		t.Fatalf("ActiveExitSnapshotQuarantine: %v ok=%v", err, ok)
	}
	if active.SelectorRevision != exitpolicy.RecoverySelectorRevision {
		t.Errorf("read back revision = %d, want %d",
			active.SelectorRevision, exitpolicy.RecoverySelectorRevision)
	}
}

// TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown: NULL is not "the current
// selector decided this". A backfill would say exactly that, and it would be
// false for the only rows that matter.
func TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	// A row as the v29 database left it.
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_snapshot_quarantines SET selector_revision=NULL WHERE position_id=?`,
		position.ID); err != nil {
		t.Fatalf("clearing the stamp: %v", err)
	}

	active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
	if err != nil || !ok {
		t.Fatalf("ActiveExitSnapshotQuarantine: %v ok=%v", err, ok)
	}
	if active.SelectorRevision != 0 {
		t.Fatalf("unstamped revision = %d, want 0 meaning unknown", active.SelectorRevision)
	}
	if !active.NeedsReJudgement() {
		t.Error("an unstamped quarantine must be re-judged: which selector wrote it is unknown, " +
			"and the rows this change exists for are exactly these")
	}
}

// TestAQuarantineFromTheCurrentSelectorIsNotReJudged is the loop guard. The retry
// is once per revision, not once per observation.
func TestAQuarantineFromTheCurrentSelectorIsNotReJudged(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if q.NeedsReJudgement() {
		t.Fatal("a quarantine this build's own selector wrote must not be re-judged; " +
			"the same comparator on the same frozen inputs cannot produce a new answer")
	}
}

// TestMigrationV30LeavesExistingQuarantinesUnstamped is the migration half of the
// NULL contract: a database that predates the column comes up with its rows still
// saying "unknown", not "the current selector decided this".
func TestMigrationV30LeavesExistingQuarantinesUnstamped(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")

	old := openJournalAtSchema(t, path, 29)
	beforeRows := journalV25RowFingerprints(t, old, nil)
	beforeColumns := journalTableColumns(t, old, mapKeys(beforeRows))
	if _, err := old.db.ExecContext(ctx, `INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES('pos-legacy','acct-7','kr','005930',1,'OPEN','10','70000','2026-08-04T00:00:00Z')`); err != nil {
		t.Fatalf("seeding a v29 position: %v", err)
	}
	if _, err := old.db.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines
		(position_id,position_generation,quarantine_version,reason,evidence,quarantined_at)
		VALUES('pos-legacy',1,1,'ambiguous_recovery','exitpolicy: recovery candidate identity mismatch',
		'2026-08-04T13:31:05Z')`); err != nil {
		t.Fatalf("seeding a v29 quarantine: %v", err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	current, err := Open(ctx, Options{Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatalf("migrating to v%d: %v", SchemaVersion, err)
	}
	defer current.Close()
	if version, err := current.SchemaVersion(ctx); err != nil || version != SchemaVersion {
		t.Fatalf("version=%d err=%v", version, err)
	}
	assertColumnsOnlyAppended(t, beforeColumns, journalTableColumns(t, current, mapKeys(beforeRows)))

	var revision sql.NullInt64
	if err := current.db.QueryRowContext(ctx,
		`SELECT selector_revision FROM exit_snapshot_quarantines WHERE position_id='pos-legacy'`).
		Scan(&revision); err != nil {
		t.Fatalf("reading the migrated row: %v", err)
	}
	if revision.Valid {
		t.Fatalf("the migration backfilled selector_revision=%d; a row written before the column "+
			"existed was not judged by the selector running now, and saying so would keep the very "+
			"rows this change exists to re-judge out of reach forever", revision.Int64)
	}
}
