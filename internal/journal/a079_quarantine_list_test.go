package journal

// a079 task 3: the list an operator is shown before lifting a quarantine.
//
// The generation predicate is the part worth testing hardest. A released and
// re-adopted position keeps its old instance's quarantine row forever, and
// offering a release for that row would be a button that changes nothing the
// engine can observe.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestActiveQuarantinesCarryTheProtectionTheStoredSnapshotKeeps(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, seed := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	snapshot, recovery := ladderSnapshotForState(t, seed, "obs-1", "70500", "70500", exitpolicy.NoRung)
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(snapshot, recovery)); err != nil {
		t.Fatalf("the judgement was refused: %v", err)
	}
	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	rows, err := j.ActiveExitSnapshotQuarantines(ctx)
	if err != nil {
		t.Fatalf("ActiveExitSnapshotQuarantines: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("active quarantines = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.PositionID != position.ID || row.Generation != position.InstanceSeq {
		t.Fatalf("row identity = %s/gen %d, want %s/gen %d",
			row.PositionID, row.Generation, position.ID, position.InstanceSeq)
	}
	if row.Version != 1 || row.Reason != "ambiguous_recovery" {
		t.Fatalf("row = v%d %q, want v1 ambiguous_recovery", row.Version, row.Reason)
	}
	if row.Symbol != position.Symbol || row.Market != position.Market {
		t.Fatalf("row names %s/%s, want %s/%s", row.Market, row.Symbol, position.Market, position.Symbol)
	}
	if row.Evidence == "" || row.QuarantinedAt == "" {
		t.Fatalf("row must carry the stored evidence and timestamp: %+v", row)
	}
	// The operator's whole decision rests on what is still protected.
	if row.Protection != snapshot.CurrentProtection {
		t.Fatalf("protection = %q, want the stored %q", row.Protection, snapshot.CurrentProtection)
	}
	if row.ProtectionUnknown != "" {
		t.Fatalf("a readable snapshot must not carry an unknown reason: %q", row.ProtectionUnknown)
	}
}

func TestAReleasedQuarantineLeavesTheActiveList(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "evidence"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if err := j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, 1,
		QuarantineReleaseHumanRepair, "operator repaired"); err != nil {
		t.Fatalf("ReleaseExitSnapshotQuarantine: %v", err)
	}

	rows, err := j.ActiveExitSnapshotQuarantines(ctx)
	if err != nil {
		t.Fatalf("ActiveExitSnapshotQuarantines: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("a released quarantine is still listed: %+v", rows)
	}
}

func TestAQuarantineFromADeadGenerationIsNotOffered(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	// A quarantine belonging to a generation the position no longer is. The
	// public writer refuses this on purpose, so the row is planted directly —
	// which is exactly how such rows exist on disk: they were written when that
	// generation was current, and the position was re-adopted afterwards.
	if _, err := j.db.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines
		(position_id,position_generation,quarantine_version,reason,evidence,quarantined_at)
		VALUES(?,?,?,?,?,?)`, position.ID, position.InstanceSeq+1, 1,
		"ambiguous_recovery", "from a generation that no longer exists",
		"2026-08-01T00:00:00Z"); err != nil {
		t.Fatalf("planting the stale row: %v", err)
	}

	rows, err := j.ActiveExitSnapshotQuarantines(ctx)
	if err != nil {
		t.Fatalf("ActiveExitSnapshotQuarantines: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("a dead generation's quarantine was offered for release: %+v", rows)
	}
}

func TestAQuarantineWithNoReadableSnapshotSaysSoRatherThanShowingBlank(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	// stored_snapshot_corrupt is precisely the reason whose snapshot may not
	// decode, so the list has to survive it — an operator cannot act on a row
	// that refuses to appear.
	if _, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"stored_snapshot_corrupt", "effective snapshot did not decode"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		`UPDATE exit_states SET effective_snapshot_json=? WHERE position_id=?`,
		"{not json", position.ID); err != nil {
		t.Fatalf("corrupting the stored snapshot: %v", err)
	}

	rows, err := j.ActiveExitSnapshotQuarantines(ctx)
	if err != nil {
		t.Fatalf("a corrupt snapshot must not fail the whole list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("active quarantines = %d, want the corrupt one listed", len(rows))
	}
	if rows[0].Protection != "" {
		t.Fatalf("an unreadable snapshot must not report a protection line: %q", rows[0].Protection)
	}
	if rows[0].ProtectionUnknown == "" {
		t.Fatal("an unreadable snapshot must say why the protection line is missing")
	}
}

// TestAReleaseDoesNotConsumeTheLedgersAbilityToQuarantineAgain is a079 task 7.2
// at the ledger boundary.
//
// What it proves: releasing closes exactly one row, and the same generation can
// be quarantined again immediately, under a fresh version, unreleased.
//
// What it deliberately does not do: fabricate a crossed-axes judgement and push
// it through RecordExitJudgement. That path is unreachable by construction —
// ValidateRecoveryDerivation re-runs the exact evaluator input and requires the
// stored line to match it field for field, so a hand-edited candidate is refused
// as invalid long before recovery selection compares anything. The judgement
// half of the claim is pinned where it is reachable:
// exitpolicy.TestCrossedAxesAcrossRungsAreStillRefused (the comparison still
// returns ErrRecoveryAmbiguous) and
// journal.TestAGenuinelyAmbiguousJudgementIsStillQuarantined (record() still
// quarantines on it). Neither is touched by a079, which is the point: a release
// changes no judgement rule, so a cause that held before a release still holds
// after one.
func TestAReleaseDoesNotConsumeTheLedgersAbilityToQuarantineAgain(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	first, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if err := j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, first.Version,
		QuarantineReleaseHumanRepair, "LOCAL_OPERATOR released quarantine v1"); err != nil {
		t.Fatalf("ReleaseExitSnapshotQuarantine: %v", err)
	}
	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if active {
		t.Fatal("the release did not clear the active quarantine")
	}

	again, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "the same cause, one observation later")
	if err != nil {
		t.Fatalf("re-quarantining a released generation: %v", err)
	}

	if again.Version <= first.Version {
		t.Fatalf("re-quarantine version = %d, want a new version above %d", again.Version, first.Version)
	}
	active, ok, err := j.ActiveExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("a release left the generation permanently unquarantinable")
	}
	if active.Version != again.Version || active.ReleasedAt != "" {
		t.Fatalf("active quarantine = %+v, want the new unreleased v%d", active, again.Version)
	}
	// The first release stays exactly one release: replaying it must not close
	// the new row.
	if err := j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, first.Version,
		QuarantineReleaseHumanRepair, "replayed"); !errors.Is(err, ErrExitSnapshotReleaseStale) {
		t.Fatalf("replaying the old release = %v, want ErrExitSnapshotReleaseStale", err)
	}
}

// TestReleasingAQuarantineWritesNothingButTheQuarantineRow is the a079 spec's
// "기준선 보존" scenario, checked the only way that actually settles it: dump the
// two tables a release must never touch, release, dump again, compare.
//
// The claim matters because the alternative an operator has today — 자동관리 해제
// followed by 새 generation 재편입 — rewrites exactly these rows, replacing the
// entry, the initial stop and the high-water with values derived from today's
// price.
func TestReleasingAQuarantineWritesNothingButTheQuarantineRow(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, seed := openedPosition(t, j, "10")
	position := currentPosition(t, j, o)

	snapshot, recovery := ladderSnapshotForState(t, seed, "obs-1", "70500", "70500", exitpolicy.NoRung)
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(snapshot, recovery)); err != nil {
		t.Fatalf("the judgement was refused: %v", err)
	}
	q, err := j.QuarantineExitSnapshot(ctx, position.ID, position.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch")
	if err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	before := dumpTables(t, j, "exit_states", "positions", "position_policy_lifecycles")

	if err := j.ReleaseExitSnapshotQuarantine(ctx, position.ID, position.InstanceSeq, q.Version,
		QuarantineReleaseHumanRepair, "LOCAL_OPERATOR released quarantine v1"); err != nil {
		t.Fatalf("ReleaseExitSnapshotQuarantine: %v", err)
	}

	after := dumpTables(t, j, "exit_states", "positions", "position_policy_lifecycles")
	for table, rows := range before {
		if rows != after[table] {
			t.Errorf("the release rewrote %s\nbefore: %s\nafter:  %s", table, rows, after[table])
		}
	}
	state := exitStateOf(t, j, position.ID)
	if state.EntryPrice != seed.EntryPrice || state.InitialStop != seed.InitialStop {
		t.Errorf("baseline moved: entry=%s stop=%s, want %s/%s",
			state.EntryPrice, state.InitialStop, seed.EntryPrice, seed.InitialStop)
	}
	if state.HighWater != snapshot.HighWater {
		t.Errorf("high water = %s, want the stored %s", state.HighWater, snapshot.HighWater)
	}
}

// dumpTables renders whole tables as one comparable string per table.
func dumpTables(t *testing.T, j *Journal, tables ...string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, table := range tables {
		rows, err := j.db.QueryContext(context.Background(), "SELECT * FROM "+table) //nolint:gosec // fixed test literals
		if err != nil {
			t.Fatalf("dumping %s: %v", table, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			t.Fatal(err)
		}
		var b strings.Builder
		for rows.Next() {
			cells := make([]any, len(columns))
			for i := range cells {
				cells[i] = new(sql.NullString)
			}
			if err := rows.Scan(cells...); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			for i, cell := range cells {
				value := cell.(*sql.NullString)
				fmt.Fprintf(&b, "%s=%q;", columns[i], value.String)
			}
			b.WriteString("\n")
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			t.Fatal(err)
		}
		out[table] = b.String()
	}
	return out
}
