package candidate

// retiredsupporter_cooling_test.go pins the safety argument of
// `retire-gainers-source`, both halves of it.
//
// # The argument, and the condition the draft left off
//
// Taking a source off the panel is safe because Candidate.Sources accumulates over
// a candidate's whole life while `heard` is built from this scan's panel and the
// schedule's not-due list. A supporter that is in neither is a source that is gone,
// coverageAnswered skips it, and the candidate keeps cooling on the evidence of the
// supporters that are still there. That path is not new and it is not a workaround:
// it was written for an operator dropping WTS, which the design calls routine.
//
// It holds only while the candidate has another supporter. coverageAnswered counts
// the supporters it could have heard from and ends on `present > 0`, so a candidate
// whose ONLY supporter is the retired source produces present == 0, returns false,
// and coolAbsent skips it — for the rest of its life. It can then leave only through
// the staleness fallback, which store.go reserves for the scan that died.
//
// So the two facts are not independent arguments. The second is safe because the
// first is true: this source never raised anything, measured at 0 candidates and 0
// observations in the operator's store on 2026-07-28, because the API refused every
// request it ever made. The two tests below are those two halves, and the second is
// written as the current behaviour rather than as the desired one.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// TestALiveSupporterStillCoolsACandidateARetiredSourceAlsoRaised is the first half.
func TestALiveSupporterStillCoolsACandidateARetiredSourceAlsoRaised(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	retired := &fakeSource{
		id:   SourceOfficialGainers,
		rows: []Row{pricedRow("005930", 3, 100, "10000")},
	}
	live := &fakeSource{
		id:   SourceOfficialTradingValue,
		rows: []Row{pricedRow("005930", 7, 100, "10000")},
	}
	both := cycleOpts(MarketKR, retired, live)
	if _, err := Cycle(ctx, s, both); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	c, found, err := s.Candidate(ctx, MarketKR, "005930", clk.Now())
	if err != nil || !found {
		t.Fatalf("Candidate: %v (found %v)", err, found)
	}
	if len(c.Sources) != 2 {
		t.Fatalf("005930 has supporters %v, want both — the arrangement under test is a "+
			"candidate carrying a retired supporter alongside a live one", c.Sources)
	}

	// The source comes off the panel. The live ranking answers with a list that has
	// rows in it and no longer carries the symbol: a zero-row reading responds
	// without vouching, and would make this pass for a different reason.
	live.rows = []Row{pricedRow("000660", 5, 100, "5000")}
	clk.Advance(time.Minute)
	res, err := Cycle(ctx, s, cycleOpts(MarketKR, live))
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	if res.Scan.Cooled != 1 {
		t.Errorf("cooled = %d, want 1. A supporter that is no longer in the panel and was "+
			"not passed over by the schedule is a source that is gone — coverageAnswered "+
			"skips it — and if that ever stops being true, every candidate this source "+
			"ever touched becomes un-coolable and can leave only through the staleness "+
			"fallback that store.go reserves for the scan that died", res.Scan.Cooled)
	}
}

// TestACandidateWhoseOnlySupporterIsRetiredIsNeverCooled is the honest half.
//
// This is the current behaviour, not the wanted one, and it is the reason
// `retire-gainers-source` measured the store instead of arguing. present > 0 in
// coverageAnswered means a candidate with no supporter this scan could have heard
// from is left alone — correctly, because nothing in the scan saw it — and for a
// candidate whose only supporter is off the panel forever, "this scan" is every
// scan there will ever be.
//
// The measurement that makes it safe: no candidate in the operator's store has
// SourceOfficialGainers among its sources, and no observation was ever written
// under that id, because the API refused every request the source made. If a
// deployment is ever found with such a candidate in it, this test is the
// description of what happens to it.
//
// If this test ever goes RED it is good news — cooling has learned to reach a
// candidate whose supporters are all gone — and the right response is to delete
// this test and the issues.md entry that goes with it rather than to restore the
// behaviour it pins.
func TestACandidateWhoseOnlySupporterIsRetiredIsNeverCooled(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	retired := &fakeSource{
		id:   SourceOfficialGainers,
		rows: []Row{pricedRow("005930", 3, 100, "10000")},
	}
	if _, err := Cycle(ctx, s, cycleOpts(MarketKR, retired)); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	c, found, err := s.Candidate(ctx, MarketKR, "005930", clk.Now())
	if err != nil || !found {
		t.Fatalf("Candidate: %v (found %v)", err, found)
	}
	if len(c.Sources) != 1 || c.Sources[0] != SourceOfficialGainers {
		t.Fatalf("005930 has supporters %v, want the retired source alone", c.Sources)
	}

	live := &fakeSource{
		id:   SourceOfficialTradingValue,
		rows: []Row{pricedRow("000660", 5, 100, "5000")},
	}
	clk.Advance(time.Minute)
	res, err := Cycle(ctx, s, cycleOpts(MarketKR, live))
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	if res.Scan.Cooled != 0 {
		t.Errorf("cooled = %d, want 0 — coverageAnswered requires at least one supporter "+
			"this scan could have heard from, and this candidate has none", res.Scan.Cooled)
	}
	if _, stillThere, cerr := s.Candidate(ctx, MarketKR, "005930", clk.Now()); cerr != nil ||
		!stillThere {
		t.Fatalf("005930: %v (found %v) — this test is about a candidate that stays active, "+
			"so it has to still be there to say anything", cerr, stillThere)
	}
}
