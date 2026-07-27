package journal

import (
	"context"
	"strings"
	"testing"
)

// trade_analytics_test.go is trade-analytics' "합성 R의 구분 집계" (change
// adopt-external-positions task 2.6).
//
// The scenario the spec names is the one under test: an account with both kinds
// of round trip reports them apart, with both sample counts.

// TestOutcomesAreSplitByWhatJustifiedThePosition is the spec's scenario: three
// engine entries and two adoptions come back as two populations, and the split
// is the `adoption_id IS NOT NULL` join rather than anything inferred.
func TestOutcomesAreSplitByWhatJustifiedThePosition(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	entered := roundTrip(t, j, "10", "70000", "72000")
	adopted := adoptedHolding(t, j, "10", "55000", "70000", "66500")
	exitLoopSell(t, j, adopted, "i-exit", "a-exit", "o-exit", "10", "72000")

	got, err := j.ClassifiedTradeOutcomes(ctx, "acct-1")
	if err != nil {
		t.Fatalf("ClassifiedTradeOutcomes: %v", err)
	}
	if len(got.Entered) != 1 || got.Entered[0].PositionID != entered {
		t.Errorf("entered = %+v, want the engine's own round trip", got.Entered)
	}
	if len(got.Adopted) != 1 || got.Adopted[0].PositionID != adopted {
		t.Errorf("adopted = %+v, want the adopted round trip", got.Adopted)
	}
	// Both are real rows with real numbers; the split is about which denominator
	// they came from, not about one of them being empty.
	if got.Adopted[0].RealizedR == "" || got.Entered[0].RealizedR == "" {
		t.Errorf("one population has no realised R: %+v", got)
	}
}

// TestMixedAggregatesCarryBothSampleCounts is the SHALL that makes the blended
// number honest. The counts live on the same struct as Combined, so a caller
// cannot render the mixed figure without holding them.
func TestMixedAggregatesCarryBothSampleCounts(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	roundTrip(t, j, "10", "70000", "72000")
	adopted := adoptedHolding(t, j, "10", "55000", "70000", "66500")
	exitLoopSell(t, j, adopted, "i-exit", "a-exit", "o-exit", "10", "72000")

	classified, err := j.ClassifiedTradeOutcomes(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	agg := AggregateBySource(classified)

	if agg.Entered.Trades != 1 || agg.Adopted.Trades != 1 {
		t.Fatalf("populations = %d measured, %d synthetic; want one of each",
			agg.Entered.Trades, agg.Adopted.Trades)
	}
	if agg.Combined.Trades != 2 {
		t.Errorf("combined trades = %d, want 2", agg.Combined.Trades)
	}
	if !agg.Mixed() {
		t.Error("Mixed() is false for an account holding both kinds of round trip")
	}
	note := agg.Note()
	if !strings.Contains(note, "1") || !strings.Contains(note, "합성") {
		t.Errorf("the caveat must name both sample counts and say which is synthetic: %q", note)
	}
	// The two ΣR are different numbers over different denominators; blending them
	// is what the note exists to qualify.
	if agg.Entered.SumRealizedR == agg.Adopted.SumRealizedR {
		t.Errorf("both populations reported ΣR %s; the fixture proves nothing",
			agg.Entered.SumRealizedR)
	}
}

// TestASinglePopulationNeedsNoCaveat keeps the note from becoming noise: an
// account with only engine entries is not a mixed aggregate.
func TestASinglePopulationNeedsNoCaveat(t *testing.T) {
	j := outcomeFixture(t)
	roundTrip(t, j, "10", "70000", "72000")

	classified, err := j.ClassifiedTradeOutcomes(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	agg := AggregateBySource(classified)
	if agg.Mixed() || agg.Note() != "" {
		t.Errorf("a single-population account produced the mixed caveat %q", agg.Note())
	}
	if agg.Combined.Trades != agg.Entered.Trades {
		t.Errorf("combined = %d, entered = %d; with no adoptions they are the same set",
			agg.Combined.Trades, agg.Entered.Trades)
	}
}

// TestClassificationSurvivesTheAggregateBeingDerived pins that nothing is
// stored: the classification is a read-time join, and the frozen rows carry no
// column saying which population they belong to.
func TestClassificationSurvivesTheAggregateBeingDerived(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	adopted := adoptedHolding(t, j, "10", "55000", "70000", "66500")
	exitLoopSell(t, j, adopted, "i-exit", "a-exit", "o-exit", "10", "72000")

	columns := tableColumns(t, j, "trade_outcomes")
	for _, c := range columns {
		if strings.Contains(c, "adoption") || strings.Contains(c, "source") ||
			strings.Contains(c, "provenance") {
			t.Errorf("trade_outcomes gained a %q column; the split is a join and the schema is "+
				"unchanged (trade-analytics SHALL)", c)
		}
	}
	got, err := j.ClassifiedTradeOutcomes(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Adopted) != 1 || len(got.Entered) != 0 {
		t.Errorf("classified = %+v, want the one adopted round trip", got)
	}
}
