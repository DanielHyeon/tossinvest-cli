package engine_test

// a084 end to end: a quarantine written by a selector this build no longer runs
// gets re-judged once, and the position goes back under evaluation.
//
// This is the consequence a078 named and could not fix. It stopped the *next*
// position from being quarantined for crossing its first rung; the three already
// in the ledger stayed unjudged — stop included — because the quarantine cuts in
// ahead of the comparison, so the fixed comparator is never reached
// (a078 issues.md I1). PLTR has been in that state since 2026-08-04T13:31:05Z.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"

	_ "modernc.org/sqlite"
)

// openRaw reaches the journal file directly. The journal API has no way to write
// an unstamped quarantine and should not grow one: unstamped rows are a fact
// about databases that predate a084, not something the engine may produce.
func openRaw(t *testing.T, h *exitHarness) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", h.dbPath)
	if err != nil {
		t.Fatalf("opening the journal file: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// supersededQuarantine writes a quarantine and then clears its stamp, which is
// exactly the on-disk shape of a row written before the column existed.
func supersededQuarantine(t *testing.T, h *exitHarness, p journal.Position) {
	t.Helper()
	ctx := context.Background()
	if _, err := h.journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}
	if _, err := openRaw(t, h).ExecContext(ctx,
		`UPDATE exit_snapshot_quarantines SET selector_revision=NULL WHERE position_id=?`,
		p.ID); err != nil {
		t.Fatalf("clearing the stamp: %v", err)
	}
}

func TestASupersededQuarantineIsReJudgedAndReleased(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	supersededQuarantine(t, h, p)

	// The cycle that meets it. The old behaviour skipped the position outright, so
	// the assertion is on what only a real judgement moves: the stored high water.
	before := h.state(p.ID).HighWater
	h.quote("005930", 70800)
	h.observe()
	if after := h.state(p.ID).HighWater; after == before {
		t.Fatalf("high water stayed at %s: a quarantine from a selector this build does not "+
			"run must be re-judged once, not skipped forever", before)
	}
	if got := h.alerts.count(obs.EventExitJudgementRefused); got != 0 {
		t.Errorf("judgement refusals during the re-judgement = %d, want none", got)
	}

	q, active, err := h.journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)
	if err != nil {
		t.Fatal(err)
	}
	if active {
		t.Fatalf("the quarantine is still active after a successful re-judgement: %+v", q)
	}

	var rows int
	var kind sql.NullString
	if err := openRaw(t, h).QueryRowContext(ctx,
		`SELECT count(*), max(release_kind) FROM exit_snapshot_quarantines WHERE position_id=?`,
		p.ID).Scan(&rows, &kind); err != nil {
		t.Fatalf("reading the quarantine history: %v", err)
	}
	if rows != 1 {
		t.Fatalf("quarantine rows = %d, want the one row, closed", rows)
	}
	if kind.String != journal.QuarantineReleaseSelectorRevised {
		t.Errorf("release kind = %q, want %s: the ledger has to distinguish a machine "+
			"re-judgement from an operator repair", kind.String, journal.QuarantineReleaseSelectorRevised)
	}

	// And it is judged from here on, which is the whole point.
	h.quote("005930", 71000)
	if next := h.observe(); next.Judged == 0 {
		t.Fatalf("the position fell out of judgement again: %+v", next)
	}
	if got := h.alerts.count(obs.EventExitJudgementRefused); got != 0 {
		t.Errorf("judgement refusals after the release = %d, want none", got)
	}
}

func TestAQuarantineThisSelectorWroteIsStillSkipped(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	// Stamped with the revision running now: the same comparator on the same
	// frozen inputs cannot reach a different answer, so retrying is churn.
	if _, err := h.journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	before := h.state(p.ID).HighWater
	h.quote("005930", 70800)
	h.observe()

	if after := h.state(p.ID).HighWater; after != before {
		t.Fatalf("high water moved %s -> %s: a quarantine this build's own selector wrote "+
			"must still stop the generation being judged", before, after)
	}
	if got := h.alerts.count(obs.EventExitJudgementRefused); got == 0 {
		t.Error("a skipped generation must still say so; the operator's only signal that a " +
			"position is unprotected is this alert")
	}
	if _, active, err := h.journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if !active {
		t.Fatal("the quarantine was released without anything having re-decided it")
	}
}
