package candidate

import (
	"context"
	"testing"
	"time"
)

// wiring_test.go is issues.md 4: the two stored firsts that section 3 and section
// 4 built and left without a caller.
//
// Both columns exist, both are once-per-life, and until a scan writes them the two
// vetoes that read them degrade to unmeasured on exactly the longest-running
// candidates — the ones the question is about. The defect is D17's, twice: the
// baseline and the first sighting position both lived only in the prunable
// observations table, and both were given columns precisely so retention could not
// move them. A column nobody writes is the same as no column.

// pricedRow is a ranking row that carries a price, which is what a scan needs in
// order to have anything to record as a baseline.
func pricedRow(symbol string, rank, total int, price string) Row {
	return Row{
		Symbol: symbol, Rank: rank, RankTotal: total,
		TradingValue: "1000000", TradingVolume: "500", Price: price,
	}
}

// TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw.
//
// The write is the durable half of D17 and of its other half (D20). Without it
// `extended` answers NO_BASELINE and `seen_late` answers NO_FIRST_RANK for every
// candidate, forever — and neither of those is a failure anybody would notice,
// because unmeasured is the ordinary state of both vetoes under the candle budget.
func TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}},
		},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	baseline, found, err := s.Baseline(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if !found {
		t.Fatal("the scan promoted no candidate at all")
	}
	if !baseline.Recorded() {
		t.Error("the scan saw a price and recorded no baseline; `extended` reads NO_BASELINE " +
			"for the life of this candidate and D17's column has no writer")
	}
	if baseline.Price != "10000" {
		t.Errorf("baseline price = %q, want %q", baseline.Price, "10000")
	}
	if !baseline.At.Equal(t0) {
		t.Errorf("baseline instant = %s, want %s — a baseline whose age is unknown cannot be read",
			baseline.At, t0)
	}
	if baseline.Source != SourceOfficialTradingValue {
		t.Errorf("baseline source = %q, want %q", baseline.Source, SourceOfficialTradingValue)
	}

	first, found, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if !found {
		t.Fatal("the scan promoted no candidate at all")
	}
	if !first.Recorded() {
		t.Error("the scan saw a ranked reading and recorded no first rank; `seen_late` reads " +
			"NO_FIRST_RANK for the life of this candidate and D20's repair has no writer")
	}
	if first.Rank != 12 || first.Total != 100 {
		t.Errorf("first rank = %d of %d, want 12 of 100", first.Rank, first.Total)
	}
	if !first.At.Equal(t0) {
		t.Errorf("first rank instant = %s, want %s", first.At, t0)
	}
}

// TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank.
//
// Once is the whole contract of both columns. A scan writes on every tick, so the
// store's idempotence is what keeps the baseline from following the price up — and
// a baseline that follows the price reports a candidate which has doubled as
// having gained nothing, which is the direction D10 forbids.
func TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{src},
	}); err != nil {
		t.Fatalf("first Collect: %v", err)
	}

	src.rows = []Row{pricedRow("005930", 3, 100, "20000")}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0.Add(60 * time.Second), Sources: []Source{src},
	}); err != nil {
		t.Fatalf("second Collect: %v", err)
	}

	baseline, _, err := s.Baseline(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if baseline.Price != "10000" {
		t.Errorf("baseline price = %q after a scan at 20000, want %q — the baseline followed "+
			"the price and the doubling disappeared", baseline.Price, "10000")
	}

	first, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if first.Rank != 12 {
		t.Errorf("first rank = %d after a second scan at 3rd, want 12 — the first sighting "+
			"followed the symbol up the list, which is what seen_late is supposed to notice",
			first.Rank)
	}

	// And the point of storing them: the expansion is measured from the first
	// price rather than from whatever rows are still held.
	rows, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	e := MeasureExpansion(baseline, rows)
	if !e.Measured || e.GainPct != "100" {
		t.Errorf("expansion = %+v, want a measured 100%% gain", e)
	}
}

// TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone.
//
// The prices and candles endpoints are not rankings and carry no position. D8
// refuses to convert "no position" into "entered at the bottom", and the store
// refuses a rank of zero — so a candidate raised only by such a source has no
// first rank, seen_late says so by name, and nothing is manufactured.
func TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialPrices, rows: []Row{{Symbol: "005930", Price: "10000"}}},
		},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	first, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if first.Recorded() {
		t.Errorf("first rank = %d of %d for a source that reports no position; "+
			"a manufactured entry position answers seen_late with a number nobody measured",
			first.Rank, first.Total)
	}
	// The baseline is still recorded: a price is a price whoever reported it.
	baseline, _, err := s.Baseline(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if !baseline.Recorded() {
		t.Error("a priced reading from a source with no rank recorded no baseline either")
	}
}

// TestANewLifeRecordsItsOwnFirstPriceAndFirstRank.
//
// D1: a candidate that expires and crosses again is a new candidate. Promote
// clears both columns on expiry, and the scan has to fill them again from the new
// life's reading — otherwise the new life runs with no baseline and no entry
// position at all, which is worse than the dead life's values it was protecting
// against.
func TestANewLifeRecordsItsOwnFirstPriceAndFirstRank(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 148, 150, "10000")}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{src},
	}); err != nil {
		t.Fatalf("first Collect: %v", err)
	}

	// Long enough that the first life has cooled implicitly and expired.
	later := t0.Add(DefaultStalenessTTL + DefaultCoolingTTL + time.Minute)
	src.rows = []Row{pricedRow("005930", 4, 150, "30000")}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: later, Sources: []Source{src},
	}); err != nil {
		t.Fatalf("second Collect: %v", err)
	}

	c, found, err := s.Candidate(ctx, MarketKR, "005930", later)
	if err != nil || !found {
		t.Fatalf("Candidate: %v (found %v)", err, found)
	}
	if !c.FirstSeenAt.Equal(later) {
		t.Fatalf("first_seen_at = %s, want %s — the second crossing was not a new life",
			c.FirstSeenAt, later)
	}

	baseline, _, err := s.Baseline(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if baseline.Price != "30000" {
		t.Errorf("baseline price = %q, want %q — the new life is measured from the dead one",
			baseline.Price, "30000")
	}
	first, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if first.Rank != 4 || !first.At.Equal(later) {
		t.Errorf("first rank = %d at %s, want 4 at %s — the new life inherited the dead "+
			"life's entry position, which is exactly D20's failure", first.Rank, first.At, later)
	}
}
