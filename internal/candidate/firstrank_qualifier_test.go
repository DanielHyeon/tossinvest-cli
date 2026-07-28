package candidate

// firstrank_qualifier_test.go is which of a tick's readings claims a symbol's first
// sighting position, and which of them is allowed to store one.
//
// # The failure this closes
//
// Collect took the position from the panel-order-first source that carried one, and
// recordFirsts then decided separately whether that reading could qualify it. Those
// two rules disagree whenever one source in a tick can qualify a symbol and an
// earlier one cannot: the earlier reading claimed the position, the later one was
// dropped, and the scan stored nothing while a perfectly good position had been in
// hand.
//
// It is not a rare alignment. A source that persistently comes back short never
// replaces its own memory — that is the F2 rule, and it is correct — so its
// NewlyListed stays unknown for as long as the degradation lasts. Being first in
// Panel's order it then claims every symbol it lists and blocks first_rank for all of
// them indefinitely, which turns one degraded endpoint into a market with no
// measurable seen_late. Whether the official rankings endpoint returns the hundred
// rows it is asked for has never been measured, and this change exists to make that
// answerable rather than to convert it into silence.
//
// # Why the hold is kept anyway
//
// When *nothing* in the panel can qualify the symbol — the session's first tick, and
// every reading a one-shot `tossctl candidate scan` takes — the position is still
// held rather than stored, because first_rank is write-once and an unqualified one
// answers for the rest of that candidate's life. The repair narrows the hold to the
// case it was written for; it does not remove it.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// unqualifiedFrom and qualifiedFrom are one source's reading of one symbol, at the
// same position, differing only in whether the source could say what the position
// meant. Both carry a price so either can also win the baseline.
func unqualifiedFrom(id SourceID, symbol string, rank, total int) *scriptedSource {
	return &scriptedSource{id: id, readings: [][]Row{{unqualifiedRow(symbol, rank, total)}}}
}

func qualifiedFrom(id SourceID, symbol string, rank, total int) *scriptedSource {
	return &scriptedSource{id: id, readings: [][]Row{{heldRow(symbol, rank, total)}}}
}

// TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne.
//
// Two sources, one tick, one symbol. The trading-value list is first in Panel's order
// and holds no previous reading; the trading-volume list holds one and answers about
// it. The position that is stored has to be the second one — the first is not a
// better answer for having been read first.
func TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne(t *testing.T) {
	ctx := context.Background()
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clock.NewFake(t0))

	opts := cycleOpts(MarketKR,
		unqualifiedFrom(SourceOfficialTradingValue, "005930", 40, 100),
		qualifiedFrom(SourceOfficialTradingVolume, "005930", 4, 100))
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if res.Scan.FirstRanks != 1 {
		t.Errorf("%d first ranks stored, want 1. A qualified reading arrived in this tick and "+
			"the scan stored nothing, because the position had already been claimed by the "+
			"source that happened to be first in the panel", res.Scan.FirstRanks)
	}
	if res.Scan.FirstRanksHeld != 0 {
		t.Errorf("%d positions held; the hold is for a tick in which no source could qualify "+
			"the symbol, and one of these two could", res.Scan.FirstRanksHeld)
	}

	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if !stored.Recorded() {
		t.Fatalf("nothing was recorded as the first sighting")
	}
	if stored.Source != SourceOfficialTradingVolume {
		t.Errorf("the stored position came from %q, want %q — the reading that could say what "+
			"the position meant", stored.Source, SourceOfficialTradingVolume)
	}
	if stored.Rank != 4 || stored.Total != 100 {
		t.Errorf("the stored position = %d of %d, want 4 of 100 (the qualified reading's, not "+
			"the panel-earlier one's 40th)", stored.Rank, stored.Total)
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: t0, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	state, sighting := seenLateOf(t, verdicts, "005930")
	if !sighting.Measured {
		t.Fatalf("seen_late is unmeasured (%s) in a tick that contained an answer",
			sighting.Reason())
	}
	if sighting.PercentilePct != "96" {
		t.Errorf("percentile = %q, want 96 (4th of 100)", sighting.PercentilePct)
	}
	if !state.Dangerous() {
		t.Errorf("seen_late = %v for a candidate first seen 4th of a hundred", state)
	}
}

// TestPanelOrderStillDecidesBetweenTwoReadingsOfTheSameStanding is the control.
//
// The preference is between a qualified reading and an unqualified one and nothing
// else. Two qualified readings in one tick are both good answers, and the tie is
// broken the way this file has always broken it — the panel's own order, which is
// where the trading-value list is first for a stated reason.
func TestPanelOrderStillDecidesBetweenTwoReadingsOfTheSameStanding(t *testing.T) {
	ctx := context.Background()
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clock.NewFake(t0))

	opts := cycleOpts(MarketKR,
		qualifiedFrom(SourceOfficialTradingValue, "005930", 40, 100),
		qualifiedFrom(SourceOfficialTradingVolume, "005930", 4, 100))
	opts.Thresholds = seenLateThresholds()
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("Cycle: %v", err)
	}

	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if stored.Source != SourceOfficialTradingValue || stored.Rank != 40 {
		t.Errorf("the stored position = %d of %d from %q, want 40 of 100 from %q. Both readings "+
			"qualify, so nothing displaces the first one and the rule stays \"panel order among "+
			"equals\" rather than \"the highest position wins\" — which would be a scan choosing "+
			"the most alarming of several true answers",
			stored.Rank, stored.Total, stored.Source, SourceOfficialTradingValue)
	}
}

// TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld is the other control, and it
// is the case recordFirsts' refusal was written for.
func TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld(t *testing.T) {
	ctx := context.Background()
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clock.NewFake(t0))

	opts := cycleOpts(MarketKR,
		unqualifiedFrom(SourceOfficialTradingValue, "005930", 40, 100),
		unqualifiedFrom(SourceOfficialTradingVolume, "005930", 4, 100))
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if res.Scan.FirstRanks != 0 || res.Scan.FirstRanksHeld != 1 {
		t.Errorf("%d stored and %d held, want 0 and 1. Preferring a qualified reading must not "+
			"turn into taking the best of a bad set: neither of these could say what its "+
			"position meant, and first_rank is written once",
			res.Scan.FirstRanks, res.Scan.FirstRanksHeld)
	}
	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if stored.Recorded() {
		t.Errorf("a position was stored from a tick in which no reading could qualify it: "+
			"%d of %d", stored.Rank, stored.Total)
	}
}

// TestAReadingThatNeverRecordedItsRequestIsHeldToo is the second half of the
// qualifier, which nothing drove.
//
// The predicate has two clauses and only the new-entrant one was reachable from a
// scan-level fixture: every other one in this package sets RankRequested equal to the
// total. A row whose source measured the new-entrant fact and recorded no request is
// the shape a schema-4-era producer leaves behind, and its truncation is unknown — so
// the position it carries could be the top of a one-row list, whose percentile is 0,
// and 0 clears every seen_late threshold that could ever exist (issues.md I3).
func TestAReadingThatNeverRecordedItsRequestIsHeldToo(t *testing.T) {
	ctx := context.Background()
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clock.NewFake(t0))

	unrecorded := heldRow("005930", 1, 1)
	unrecorded.RankRequested = 0
	src := &scriptedSource{
		id:       SourceOfficialTradingValue,
		readings: [][]Row{{unrecorded}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if res.Scan.FirstRanks != 0 || res.Scan.FirstRanksHeld != 1 {
		t.Errorf("%d stored and %d held for a reading that recorded no request, want 0 and 1",
			res.Scan.FirstRanks, res.Scan.FirstRanksHeld)
	}
	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if stored.Recorded() {
		t.Errorf("the position was stored with no request beside it: %d of %d. The percentile "+
			"of the top of a one-row list is 0, and 0 is below every threshold this veto could "+
			"ever be given", stored.Rank, stored.Total)
	}
}

// TestAQualifiedReadingIsAlsoPreferredWhenItArrivesForOnlySomeSymbols.
//
// The preference is per symbol rather than per reading, because that is how the two
// maps are keyed and because a degraded source rarely fails uniformly: it lists some
// of what it listed last time. A symbol only the unqualified source carries stays
// held, and one both carry is stored — in the same tick.
func TestAQualifiedReadingIsAlsoPreferredWhenItArrivesForOnlySomeSymbols(t *testing.T) {
	ctx := context.Background()
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clock.NewFake(t0))

	first := &scriptedSource{
		id: SourceOfficialTradingValue,
		readings: [][]Row{{
			unqualifiedRow("005930", 40, 100),
			unqualifiedRow("000660", 41, 100),
		}},
	}
	second := &scriptedSource{
		id:       SourceOfficialTradingVolume,
		readings: [][]Row{{heldRow("005930", 4, 100)}},
	}
	opts := cycleOpts(MarketKR, first, second)
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if res.Scan.FirstRanks != 1 || res.Scan.FirstRanksHeld != 1 {
		t.Errorf("%d stored and %d held, want 1 and 1 — one symbol was answered for in this "+
			"tick and the other was not", res.Scan.FirstRanks, res.Scan.FirstRanksHeld)
	}
	answered, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank(005930): %v", err)
	}
	if !answered.Recorded() || answered.Source != SourceOfficialTradingVolume {
		t.Errorf("005930's stored position = %+v; the qualified reading covered this symbol",
			answered)
	}
	unanswered, _, err := s.FirstRank(ctx, MarketKR, "000660")
	if err != nil {
		t.Fatalf("FirstRank(000660): %v", err)
	}
	if unanswered.Recorded() {
		t.Errorf("000660's position was stored from %q, which held no previous reading; no "+
			"source in this tick said anything about this symbol's standing",
			unanswered.Source)
	}
}
