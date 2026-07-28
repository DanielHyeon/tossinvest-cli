package candidate

// migrated_firstrank_test.go pins what happens to a first sighting position that
// was stored before the two facts beside it existed — the rows the schema-4 rung
// creates by backfilling, and every row an older build wrote.
//
// # The claim that was made and the claim that is true
//
// store_test.go said such a store "gets its answer back one scan after the upgrade,
// when a source reports a reading it can qualify". It does not, and three separate
// things stop it: NoteFirstRank returns early whenever first_rank is set, recordFirsts
// does not offer a reading for a candidate that already has one, and — the one that
// would still hold if the other two were changed — the facts describe the reading
// the position came from, that reading is gone, and a later reading's answer is an
// answer about a later reading. Writing one in would be the substitution design D3
// exists to refuse, arriving from behind.
//
// What is true is that the state is bounded by the candidate's life rather than by
// the store's. When the life ends and the symbol crosses again, Promote clears the
// position and the next one is taken from a reading a running source can qualify.
// A migrated store recovers at the rate its candidates turn over.
//
// Both halves are asserted here because the first one is the kind of claim that gets
// quietly repaired into the wrong thing: "fill it from the current reading" is one
// line, looks like a fix, and would make every migrated candidate measurable against
// a fact about a different moment.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// unqualifiedFirstRank is the shape an older build stored: a position, its instant
// and its source, and nothing about the reading it came from.
func unqualifiedFirstRank(rank, total int, at time.Time, source SourceID) FirstRank {
	return FirstRank{Rank: rank, Total: total, At: at, Source: source}
}

// TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan.
func TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	// A candidate with a position and no qualifiers — a v3 store one rung up, or a
	// row written by the build this change is fixing.
	src := &scriptedSource{
		id:       SourceOfficialTradingValue,
		readings: [][]Row{{heldRow("005930", 12, 100)}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE candidates SET first_rank_newly_listed = NULL, first_rank_requested = NULL
		  WHERE market = 'KR' AND symbol = '005930'`); err != nil {
		t.Fatalf("blanking the qualifiers: %v", err)
	}

	// Five fully qualified scans, each one offering a reading this source can answer
	// about. None of them may reach back into the stored position.
	for i := 0; i < 5; i++ {
		clk.Advance(20 * time.Second)
		opts.Schedule = NewSchedule(nil)
		if _, err := Cycle(ctx, s, opts); err != nil {
			t.Fatalf("Cycle %d: %v", i+2, err)
		}
	}

	stored, found, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil || !found {
		t.Fatalf("FirstRank: %v found=%v", err, found)
	}
	if stored.Rank != 12 || stored.Total != 100 {
		t.Errorf("the stored position moved to %d of %d; first_rank is written once and five "+
			"later readings must not be able to change where this candidate entered",
			stored.Rank, stored.Total)
	}
	if stored.NewlyListed.Known() || stored.Requested != 0 {
		t.Errorf("the qualifiers were filled in as %s / %d from a later reading. They describe "+
			"the reading this position came from, and that reading is gone — writing a later "+
			"one's answer beside it is the substitution D3 refuses",
			stored.NewlyListed, stored.Requested)
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: clk.Now(), Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	state, sighting := seenLateOf(t, verdicts, "005930")
	if sighting.Measured {
		t.Errorf("the migrated position is measured at %s%% after five qualified scans",
			sighting.PercentilePct)
	}
	if state.Reason() != VetoNewEntrantUnknown {
		t.Errorf("reason = %q, want %q", state.Reason(), VetoNewEntrantUnknown)
	}
}

// TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds is the other half, and it
// is what bounds the state above.
//
// Expiry clears the position (D1), the next crossing records one from a reading the
// source can qualify, and the candidate is measurable from then on. This is the
// honest version of "the store gets its answer back": at the rate candidates turn
// over, not at the rate scans run.
func TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	src := &scriptedSource{
		id:       SourceOfficialTradingValue,
		readings: [][]Row{{heldRow("005930", 60, 100)}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE candidates SET first_rank_newly_listed = NULL, first_rank_requested = NULL
		  WHERE market = 'KR' AND symbol = '005930'`); err != nil {
		t.Fatalf("blanking the qualifiers: %v", err)
	}

	clk.Advance(DefaultCoolingTTL + DefaultStalenessTTL + time.Minute)
	reborn := clk.Now()
	src.readings = [][]Row{{enteredRow("005930", 4, 100)}}
	opts.Schedule = NewSchedule(nil)
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("second Cycle: %v", err)
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: reborn, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	_, sighting := seenLateOf(t, verdicts, "005930")
	if !sighting.Measured {
		t.Fatalf("the new life is unmeasured (%s); its position came from a reading taken "+
			"after the upgrade and the source answered about it", sighting.Reason())
	}
	if sighting.Rank != 4 {
		t.Errorf("the new life's position = %d, want 4 — the dead life's 60th does not "+
			"survive the expiry", sighting.Rank)
	}
}

// TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext states the store-level
// rule the two tests above depend on, and states it against the exact offer somebody
// repairing this would make: the same candidate, a later reading, fully qualified.
func TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, MarketKR, "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	stored, err := s.NoteFirstRank(ctx, MarketKR, "005930",
		unqualifiedFirstRank(12, 100, t0, SourceOfficialTradingValue))
	if err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}
	if stored.NewlyListed.Known() || stored.Requested != 0 {
		t.Fatalf("the unqualified write came back as %s / %d", stored.NewlyListed, stored.Requested)
	}

	// A later, fully qualified reading of the same candidate.
	again, err := s.NoteFirstRank(ctx, MarketKR, "005930",
		storedFirstRank(12, 100, t0.Add(time.Minute), SourceOfficialTradingValue))
	if err != nil {
		t.Fatalf("second NoteFirstRank: %v", err)
	}
	if again.NewlyListed.Known() || again.Requested != 0 {
		t.Errorf("a later reading's answers were written beside a position they are not "+
			"about: %s / %d. The instant differs, so this is a different reading, and its "+
			"answer to \"was this symbol in my previous reading\" is about that one",
			again.NewlyListed, again.Requested)
	}
	if !again.At.Equal(t0) || again.Rank != 12 {
		t.Errorf("the stored position moved to %d at %v", again.Rank, again.At)
	}
}
