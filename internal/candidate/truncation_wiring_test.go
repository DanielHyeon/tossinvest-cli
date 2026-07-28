package candidate

// truncation_wiring_test.go is section 3 driven end to end, and it exists because
// section 3 had no test that used the wiring at all.
//
// # What was unpinned
//
// truncation_test.go asserts the refusal against a FirstRank built by hand, and
// previous_reading_test.go asserts that the adapters report the requested count.
// Between those two ends sit two assignments, and nothing crossed them:
//
//	scan.go   Reported{… RankRequested: r.RankRequested …}   reading → observation
//	scan.go   FirstRank{… Requested: o.Reported.RankRequested …}   observation → store
//
// Setting either of them to 0 left the whole suite green. The identical mutation on
// the sibling fact one line over — NewlyListed — fails three tests, so section 1 was
// pinned and section 3 was not. Every other scan-level fixture in this package sets
// RankRequested equal to the total, so no wired test ever produced a truncated
// reading and the two assignments were carrying a fact nobody checked had arrived.
//
// # Why it is a Cycle and not a unit test
//
// A unit test of either assignment would pin that assignment. What has to hold is
// that a source reporting a short reading produces a candidate whose seen_late names
// truncation as the reason — through Collect, RecordObservations, recordFirsts,
// NoteFirstRank, Assess and MeasureFirstSighting — and each of those is a place the
// two numbers can be dropped.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// truncatedRow is a ranked row from a reading that arrived short: the source asked
// for a hundred and three came back. Its new-entrant fact is measured, so that the
// refusal under test is the truncation one and not the reading-unqualified one.
func truncatedRow(symbol string, rank, arrived, requested int) Row {
	r := heldRow(symbol, rank, arrived)
	r.RankRequested = requested
	return r
}

// TestATruncatedReadingReachesTheVerdictAsTruncated is F3, and it is the shape of
// TestASessionStartDoesNotStampThePanelAsSeenLate one section over.
func TestATruncatedReadingReachesTheVerdictAsTruncated(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	// Three rows of the hundred this source asked for. The top row's percentile
	// against what arrived is 66.7 — the middle of a list, measured and
	// unremarkable — which is why a refusal is needed rather than a caveat.
	src := &scriptedSource{
		id: SourceOfficialTradingValue,
		readings: [][]Row{{
			truncatedRow("005930", 1, 3, 100),
			truncatedRow("000660", 2, 3, 100),
			truncatedRow("035720", 3, 3, 100),
		}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}

	// The scan's own account of the reading, which is what answers "does the
	// endpoint return the hundred rows it is asked for" in one scan rather than
	// several candidate lives later.
	facts, ok := res.Scan.Readings[SourceOfficialTradingValue]
	if !ok {
		t.Fatalf("the scan recorded no reading for %s; the per-source record is what makes a "+
			"short endpoint visible before the veto census does", SourceOfficialTradingValue)
	}
	if facts.Requested != 100 || facts.Arrived != 3 || !facts.Truncation().Yes() {
		t.Errorf("the recorded reading = %d requested, %d arrived, %s; want 100, 3, truncated",
			facts.Requested, facts.Arrived, facts.Truncation())
	}
	// The position is stored, unlike an unqualified one. Both facts were taken here,
	// so the refusal it produces names a live fault an operator can chase; declining
	// to store it would trade that diagnosis for NO_FIRST_RANK.
	if res.Scan.FirstRanks != 3 {
		t.Errorf("%d first ranks stored, want 3 — a truncated reading is a measured reading "+
			"and its position is the honest record of where this candidate was first seen",
			res.Scan.FirstRanks)
	}
	if res.Scan.FirstRanksHeld != 0 {
		t.Errorf("%d positions held; a truncated reading carries both qualifiers and is not "+
			"the unqualified case", res.Scan.FirstRanksHeld)
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: t0, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if len(verdicts) != 3 {
		t.Fatalf("%d candidates, want 3", len(verdicts))
	}
	for _, v := range verdicts {
		state, sighting := v.Chase.SeenLate, v.Sighting
		if state.Measured() {
			t.Errorf("%s: seen_late is %v from a reading that asked for %d rows and received "+
				"%d. The percentile normalises by what arrived, so the top of this reading "+
				"comes out mid-list of a list that never existed",
				v.Summary.Symbol, state, facts.Requested, facts.Arrived)
		}
		if state.Reason() != VetoReadingTruncated {
			t.Errorf("%s: reason = %q, want %q. The two numbers travel reading → observation "+
				"→ candidates.first_rank_requested, and every one of those steps can drop "+
				"one of them without anything else failing",
				v.Summary.Symbol, state.Reason(), VetoReadingTruncated)
		}
		if !sighting.Truncation.Yes() {
			t.Errorf("%s: the stored truncation fact = %s, want truncated",
				v.Summary.Symbol, sighting.Truncation)
		}
		if sighting.PercentilePct != "" {
			t.Errorf("%s: percentile = %q on a refused sighting; absent is not a number",
				v.Summary.Symbol, sighting.PercentilePct)
		}
	}

	// And the per-source reduction says whose readings these were, which is the
	// difference between "seen_late is unmeasured" and "this endpoint is short".
	sightings := res.Sightings
	if len(sightings) != 1 {
		t.Fatalf("%d sources in the sighting reduction, want 1", len(sightings))
	}
	if sightings[0].Source != SourceOfficialTradingValue {
		t.Errorf("the reduction attributes the refusals to %q", sightings[0].Source)
	}
	if got := sightings[0].NotMeasured[VetoReadingTruncated]; got != 3 {
		t.Errorf("%s refusals attributed to this source = %d, want 3",
			VetoReadingTruncated, got)
	}
	if sightings[0].Measured != 0 {
		t.Errorf("%d of this source's sightings were measured", sightings[0].Measured)
	}
}

// TestAWholeReadingOfTheSameLengthIsMeasured is the control beside it.
//
// Three of three requested is a complete reading of a three-row list, and its
// positions are perfectly good percentiles. Without this, the test above would pass
// just as well against a build that refused every short list — which is the floor D4
// declines to invent.
func TestAWholeReadingOfTheSameLengthIsMeasured(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	src := &scriptedSource{
		id: SourceOfficialTradingValue,
		readings: [][]Row{{
			truncatedRow("005930", 1, 3, 3),
			truncatedRow("000660", 2, 3, 3),
			truncatedRow("035720", 3, 3, 3),
		}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()
	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if facts := res.Scan.Readings[SourceOfficialTradingValue]; !facts.Truncation().No() {
		t.Fatalf("three of three requested was recorded as %s", facts.Truncation())
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: t0, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	_, sighting := seenLateOf(t, verdicts, "005930")
	if !sighting.Measured {
		t.Fatalf("a complete three-row reading was refused as %s; the length of the list is "+
			"not what decides this", sighting.Reason())
	}
	if sighting.PercentilePct != "66.666666666666" {
		t.Errorf("percentile = %q, want 66.666666666666 (1st of 3). This is the number the "+
			"truncated reading above produced too, and it is why that one is refused rather "+
			"than annotated: nothing about it looks wrong", sighting.PercentilePct)
	}
}
