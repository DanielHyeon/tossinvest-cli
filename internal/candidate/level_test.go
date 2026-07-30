package candidate

import (
	"testing"
	"time"
)

// levelObs is one reading of KR:005930 that carries a price and nothing else —
// which is what every ranking and every prices call actually returns.
func levelObs(at time.Time, source SourceID, price string) Observation {
	return obs("005930", at, source, Reported{Price: price})
}

// candleObs is a candle reading: the one source that carries the day's extremes.
func candleObs(at time.Time, price, high, low string) Observation {
	return obs("005930", at, SourceOfficialCandles,
		Reported{Price: price, DayHigh: high, DayLow: low})
}

// storedBaseline is what Store.Baseline returns for a candidate whose first
// priced observation has been recorded (D17). Every expansion measured in this
// file supplies one, because a candidate without one is not measurable — which is
// itself pinned, in TestAnExpansionWithNoStoredBaselineIsUnmeasured.
func storedBaseline(price string, at time.Time, source SourceID) Baseline {
	return Baseline{Price: price, At: at, Source: source}
}

// --- task 3.6: expansion against the first price we ever saw --------------------

// TestTheZeroValueOfEveryLevelMetricIsUnmeasured.
//
// This is D10 aimed at the type itself. A struct whose zero value reads as
// "measured, and zero" is a struct that reports a candidate nobody looked at as
// one that was looked at and found harmless — a map lookup that misses, a slice
// that came back empty, a field somebody forgot to assign, and the veto is off
// for that candidate with nothing having failed.
func TestTheZeroValueOfEveryLevelMetricIsUnmeasured(t *testing.T) {
	var e Expansion
	if e.Measured {
		t.Error("the zero Expansion reports itself as measured; an unassigned metric would " +
			"read as a candidate that has not run")
	}
	var r RangePosition
	if r.Measured {
		t.Error("the zero RangePosition reports itself as measured; an unassigned metric would " +
			"read as a candidate that was checked against the day's high")
	}
	if r.PositionMeasured {
		t.Error("the zero RangePosition reports a measured position inside the day's range")
	}
}

// TestAnUnmeasuredAnswerNeverRendersAsNothing is the §3 review's smallest finding
// and the one with the widest surface.
//
// Measured=false with Why=="" is an unmeasured answer that names no missing
// input, and it is the one state TestEveryUnmeasuredAnswerNamesItsMissingInput
// cannot reach — it comes from a struct nobody passed to a measurement at all.
// §5 renders the reason beside the veto, and "unmeasured ()" on a screen whose
// entire job is to stop an unmeasured veto from looking like a passed one teaches
// an operator to skip the column.
func TestAnUnmeasuredAnswerNeverRendersAsNothing(t *testing.T) {
	var e Expansion
	if got := e.Reason(); got != LevelNotEvaluated {
		t.Errorf("the zero Expansion's reason = %q, want %q", got, LevelNotEvaluated)
	}
	var r RangePosition
	if got := r.Reason(); got != LevelNotEvaluated {
		t.Errorf("the zero RangePosition's reason = %q, want %q", got, LevelNotEvaluated)
	}
	// And a measured answer names nothing, because there is nothing to name.
	measured := MeasureRangePosition([]Observation{candleObs(t0, "71000", "72000", "68000")})
	if !measured.Measured {
		t.Fatalf("unmeasured (%q)", measured.Why)
	}
	if got := measured.Reason(); got != "" {
		t.Errorf("a measured range position gives the reason %q, want none", got)
	}
}

// TestAnExpansionWithNoStoredBaselineIsUnmeasured is the §3 review's P0, stated
// as the rule that closes it.
//
// The baseline used to be the earliest priced row of whatever slice the caller
// handed over, and MeasureExpansion cannot tell a whole history from a window.
// D17 gives it a column instead, and a caller without one now gets a named
// unmeasured answer rather than a number derived from rows that happen to have
// survived retention.
//
// Unmeasured is safe here in a way a number is not: D10 stops an unmeasured veto
// from being counted as a pass, and nothing stops a plausible one.
func TestAnExpansionWithNoStoredBaselineIsUnmeasured(t *testing.T) {
	got := MeasureExpansion(Baseline{}, []Observation{
		levelObs(t0, SourceOfficialPrices, "20000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "30000"),
	})
	if got.Measured {
		t.Fatalf("expansion measured (%s%%) from a series with no stored baseline; the "+
			"earliest surviving row silently became the first price", got.GainPct)
	}
	if got.Why != LevelNoBaseline {
		t.Errorf("reason = %q, want %q", got.Why, LevelNoBaseline)
	}
	// It still reports what it read. An unmeasured answer that carries none of its
	// inputs cannot be acted on.
	if got.FirstPrice != "20000" || got.LastPrice != "30000" {
		t.Errorf("the readings did not survive the unmeasured answer: first=%q last=%q",
			got.FirstPrice, got.LastPrice)
	}
	if got.GainPct != "" || got.Ratio != "" {
		t.Errorf("an unmeasured expansion carries arithmetic: ratio=%q gain=%q",
			got.Ratio, got.GainPct)
	}
}

// TestAWindowedSeriesCannotRebaseTheStoredBaseline is the same P0 from the other
// side: the caller has a baseline, and the slice is a window that starts after it.
//
// This is what `since` produces — SourceObservations' own doc comment recommends
// it ("the last two minutes of one source rather than two trading days of rows to
// use three of them") — and what PruneObservations produces after 48 hours. The
// stored baseline wins both times.
func TestAWindowedSeriesCannotRebaseTheStoredBaseline(t *testing.T) {
	baseline := storedBaseline("10000", t0, SourceOfficialPrices)
	// The window: everything before t0+49h has been pruned or was never asked for.
	got := MeasureExpansion(baseline, []Observation{
		levelObs(t0.Add(49*time.Hour), SourceOfficialPrices, "20000"),
		levelObs(t0.Add(50*time.Hour), SourceOfficialPrices, "30000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.FirstPrice != "10000" || !got.FirstAt.Equal(t0) {
		t.Errorf("baseline = %q at %v, want the stored 10000 at %v — the window re-based it",
			got.FirstPrice, got.FirstAt, t0)
	}
	if got.GainPct != "200" {
		t.Errorf("gain = %s%%, want 200%% — a candidate that tripled reported %s%%",
			got.GainPct, got.GainPct)
	}
	if !got.BaselineStored {
		t.Error("the answer does not report that the baseline came from the candidate summary")
	}
}

// TestABaselineOnlyEverMovesBackwards.
//
// A stored baseline is authority for "the first price we saw", and a reading
// older than it is evidence that it is not. Taking the older one can only make
// the measured expansion larger, which is the direction `extended` is a veto in;
// taking a newer one is the failure the column exists to stop. So the rule is
// one-directional and stated rather than implied.
func TestABaselineOnlyEverMovesBackwards(t *testing.T) {
	baseline := storedBaseline("20000", t0.Add(time.Hour), SourceOfficialPrices)
	got := MeasureExpansion(baseline, []Observation{
		levelObs(t0, SourceWTSPopular, "10000"),
		levelObs(t0.Add(2*time.Hour), SourceOfficialPrices, "30000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.FirstPrice != "10000" || !got.FirstAt.Equal(t0) {
		t.Errorf("baseline = %q at %v, want the older evidence 10000 at %v",
			got.FirstPrice, got.FirstAt, t0)
	}
	if got.BaselineStored {
		t.Error("the answer claims the stored baseline was used; it was overridden by an " +
			"older reading and the record has to say so")
	}
	if got.GainPct != "200" {
		t.Errorf("gain = %s%%, want 200%%", got.GainPct)
	}
}

// TestExpansionIsUndefinedWhenNoSourceEverReportedAPrice.
//
// The failure this catches is the one D10 describes for near_high, one axis
// across: with no first price there is no baseline, and a metric that answers
// "0% — it has not run" for that case hands the `extended` veto (task 4.2) a
// clean bill of health for a candidate whose price nobody has ever seen.
//
// Undefined is not zero, and neither of them is "not extended".
func TestExpansionIsUndefinedWhenNoSourceEverReportedAPrice(t *testing.T) {
	// A trading-value ranking that carried no price at all: entirely ordinary.
	got := MeasureExpansion(Baseline{}, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 4, RankTotal: 100, TradingValue: "1000"}),
		obs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue,
			Reported{Rank: 2, RankTotal: 100, TradingValue: "9000"}),
	})
	if got.Measured {
		t.Fatalf("expansion of a candidate with no price reads as measured (%s%%); "+
			"there is no baseline to expand from", got.GainPct)
	}
	if got.Why != LevelNoPrice {
		t.Errorf("reason = %q, want %q — the record has to say which input was missing",
			got.Why, LevelNoPrice)
	}
}

// TestNoObservationsAtAllIsItsOwnAnswer.
//
// "Nobody has reported a price for this symbol" and "nobody has reported
// anything for this symbol" call for different responses: the first is a source
// that answers without the field, the second is a symbol we have not polled. The
// two reasons are separate for the reason D3 separates seen_late from extended —
// one is fixed by changing the source panel and the other is not.
func TestNoObservationsAtAllIsItsOwnAnswer(t *testing.T) {
	got := MeasureExpansion(Baseline{}, nil)
	if got.Measured {
		t.Fatal("expansion measured from an empty series")
	}
	if got.Why != LevelNoObservations {
		t.Errorf("reason = %q, want %q", got.Why, LevelNoObservations)
	}
}

// TestAnObservationWithoutAPriceIsSkippedAndNotReadAsZero.
//
// The two rows that carry no price are the ranking rows; the third is a prices
// call. Reading the absent ones as 0 would make the baseline zero, and every
// ratio taken against it infinite — the manufactured maximum signal the
// observation validator already refuses for ranks.
func TestAnObservationWithoutAPriceIsSkippedAndNotReadAsZero(t *testing.T) {
	baseline := storedBaseline("50000", t0.Add(30*time.Second), SourceOfficialPrices)
	got := MeasureExpansion(baseline, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 9, RankTotal: 100}),
		obs("005930", t0.Add(15*time.Second), SourceOfficialTradingVolume, Reported{Rank: 7, RankTotal: 100}),
		levelObs(t0.Add(30*time.Second), SourceOfficialPrices, "50000"),
		levelObs(t0.Add(45*time.Second), SourceOfficialPrices, "60000"),
	})
	if !got.Measured {
		t.Fatalf("expansion is unmeasured (%q) although two readings carried a price", got.Why)
	}
	if got.FirstPrice != "50000" {
		t.Errorf("first price = %q, want 50000 — the price-less rows became the baseline",
			got.FirstPrice)
	}
	if got.GainPct != "20" {
		t.Errorf("gain = %s%%, want 20%%", got.GainPct)
	}
	if got.Ratio != "1.2" {
		t.Errorf("ratio = %s, want 1.2", got.Ratio)
	}
	if !got.FirstAt.Equal(t0.Add(30 * time.Second)) {
		t.Errorf("first price instant = %v, want the first reading that carried one", got.FirstAt)
	}
}

// TestTheLatestPriceIsTheLatestAndNotTheEarliestRow.
//
// A scan merges several sources into one write, and a caller that concatenates
// two reads hands this an unsorted slice. An end picked by position in the slice
// rather than by instant is an end that changes with the order the caller
// happened to build its list in — and it moves the `extended` veto with it,
// silently, in whichever direction the ordering fell.
func TestTheLatestPriceIsTheLatestAndNotTheEarliestRow(t *testing.T) {
	baseline := storedBaseline("40000", t0, SourceWTSPopular)
	got := MeasureExpansion(baseline, []Observation{
		levelObs(t0.Add(60*time.Second), SourceOfficialPrices, "80000"),
		levelObs(t0, SourceWTSPopular, "40000"),
		levelObs(t0.Add(30*time.Second), SourceOfficialPrices, "60000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.FirstPrice != "40000" || !got.FirstAt.Equal(t0) {
		t.Errorf("baseline = %q at %v, want 40000 at %v", got.FirstPrice, got.FirstAt, t0)
	}
	if got.LastPrice != "80000" || !got.LastAt.Equal(t0.Add(60*time.Second)) {
		t.Errorf("latest = %q at %v, want 80000 at %v — the slice order decided the latest price",
			got.LastPrice, got.LastAt, t0.Add(60*time.Second))
	}
	if got.GainPct != "100" {
		t.Errorf("gain = %s%%, want 100%%", got.GainPct)
	}
}

// TestReadingsThatShareAnInstantResolveDeliberately.
//
// scan.go stamps one instant on a whole pass, so every source's row for a symbol
// carries the same one: ties are the ordinary case here, not an edge case. The
// rule has to be chosen rather than fallen into — and the version this replaces
// fell into the wrong one. `first` used a strict Before and `last` a strict
// After, so both ends kept the earliest matching index, and the field named
// LastPrice resolved to the oldest-inserted row of the newest instant.
//
// The rule: the baseline keeps the first reading of the instant in the caller's
// order, the latest price keeps the last. For the store's (observed_at, id) that
// is the oldest-inserted and the newest-inserted row, which is what the names say.
func TestReadingsThatShareAnInstantResolveDeliberately(t *testing.T) {
	baseline := storedBaseline("10000", t0, SourceOfficialPrices)
	got := MeasureExpansion(baseline, []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0, SourceWTSPopular, "11000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.FirstPrice != "10000" || got.FirstSource != SourceOfficialPrices {
		t.Errorf("first = %q from %q, want the first row of the instant (10000)",
			got.FirstPrice, got.FirstSource)
	}
	if got.LastPrice != "11000" || got.LastSource != SourceWTSPopular {
		t.Errorf("last = %q from %q, want the last row of the instant (11000) — both ends "+
			"kept the same index and LastPrice resolved to the oldest-inserted row",
			got.LastPrice, got.LastSource)
	}

	// The same rule on the range position's choice of candle.
	pos := MeasureRangePosition([]Observation{
		candleObs(t0, "70000", "71000", "69000"),
		candleObs(t0, "70000", "72000", "69000"),
	})
	if !pos.Measured {
		t.Fatalf("unmeasured (%q)", pos.Why)
	}
	if pos.High != "72000" {
		t.Errorf("high = %q, want the last candle of the instant (72000)", pos.High)
	}
}

// TestTheFirstPriceMayComeFromAnySourceThatCarriedOne.
//
// D9 keeps the rate and acceleration series per source, because TradingValue is
// a different cumulative in every source — different unit, different aggregation
// window — so differencing one against another measures the sources rather than
// the market.
//
// A price is not a cumulative. It is the last trade in the market this candidate
// belongs to, and every source is reporting the same number. Expansion is a
// property of the candidate (D1: one symbol raised by two sources is one
// candidate), so its baseline is the earliest price anyone reported, and the
// record names the source that reported it so a disagreement between two sources
// is answerable afterwards rather than invisible.
func TestTheFirstPriceMayComeFromAnySourceThatCarriedOne(t *testing.T) {
	baseline := storedBaseline("10000", t0, SourceWTSPopular)
	got := MeasureExpansion(baseline, []Observation{
		levelObs(t0, SourceWTSPopular, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "11000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.FirstSource != SourceWTSPopular {
		t.Errorf("first source = %q, want %q — the record cannot say who set the baseline",
			got.FirstSource, SourceWTSPopular)
	}
	if got.LastSource != SourceOfficialPrices {
		t.Errorf("last source = %q, want %q", got.LastSource, SourceOfficialPrices)
	}
	if got.GainPct != "10" {
		t.Errorf("gain = %s%%, want 10%%", got.GainPct)
	}
}

// TestASinglePricedObservationHasNotExpandedYet.
//
// One price is a measurement: as far as this record goes, the candidate has not
// moved since we saw it. That is deliberately not the same claim as "it did not
// run before we saw it" — that question is seen_late's (task 4.1), and D3 keeps
// the two apart because they are fixed by different things. Answering unmeasured
// here would fold them back together.
func TestASinglePricedObservationHasNotExpandedYet(t *testing.T) {
	baseline := storedBaseline("70000", t0, SourceOfficialPrices)
	got := MeasureExpansion(baseline, []Observation{levelObs(t0, SourceOfficialPrices, "70000")})
	if !got.Measured {
		t.Fatalf("a single priced reading is unmeasured (%q); it is a baseline and a "+
			"current price at once", got.Why)
	}
	if got.GainPct != "0" || got.Ratio != "1" {
		t.Errorf("gain = %s%% ratio = %s, want 0%% and 1", got.GainPct, got.Ratio)
	}
}

// TestABaselineThatIsNotANumberIsNotAMeasurement.
//
// strconv.ParseFloat's error is the one every careless conversion drops, and the
// value it hands back with it is 0. A thousands separator, a currency mark, an
// API that started sending "N/A" — each becomes a baseline of zero, and the ratio
// taken against it is infinite. Refusing loudly costs a metric; accepting quietly
// costs the veto.
//
// The denominator is the one place this is fatal. A malformed row elsewhere in
// the series is skipped and counted — see
// TestOneUnreadablePriceDoesNotDiscardTheWholeSeries.
func TestABaselineThatIsNotANumberIsNotAMeasurement(t *testing.T) {
	for _, text := range []string{"1,000,000", "N/A", "70000원", "NaN", "+Inf", "1e5"} {
		got := MeasureExpansion(storedBaseline(text, t0, SourceOfficialPrices),
			[]Observation{levelObs(t0.Add(time.Minute), SourceOfficialPrices, "80000")})
		if got.Measured {
			t.Errorf("a baseline of %q was measured (%s%%); ParseFloat's discarded error is a "+
				"baseline of zero", text, got.GainPct)
		}
		if got.Why != LevelUnreadableDecimal {
			t.Errorf("a baseline of %q gave reason %q, want %q", text, got.Why, LevelUnreadableDecimal)
		}
	}

	// And a series in which every reading is malformed has no current price at
	// all. That is unreadable rather than absent: the source carried something.
	all := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices),
		[]Observation{levelObs(t0.Add(time.Minute), SourceOfficialPrices, "N/A")})
	if all.Measured {
		t.Errorf("a series of nothing but malformed prices was measured: %s%%", all.GainPct)
	}
	if all.Why != LevelUnreadableDecimal {
		t.Errorf("reason = %q, want %q", all.Why, LevelUnreadableDecimal)
	}
}

// TestOneUnreadablePriceDoesNotDiscardTheWholeSeries is the §3 review's finding
// about how much one bad row used to cost.
//
// A single unreadable price anywhere in the slice returned
// Why=UNREADABLE_DECIMAL with FirstPrice, LastPrice, FirstAt and FirstSource all
// empty. Nine sources feed this series and raw rows are retained for 48 hours, so
// one malformed row from one of them turned `extended` off for that candidate for
// two days — and the answer carried none of the inputs it had successfully read.
//
// MeasureRangePosition already did the opposite: it records the strings it read
// before it checks whether they are usable. The two are consistent now, in that
// direction.
func TestOneUnreadablePriceDoesNotDiscardTheWholeSeries(t *testing.T) {
	got := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceWTSPopular, "N/A"),
		levelObs(t0.Add(2*time.Minute), SourceOfficialPrices, "12000"),
	})
	if !got.Measured {
		t.Fatalf("one unreadable row out of three turned the measurement off (%q)", got.Why)
	}
	if got.FirstPrice != "10000" || got.LastPrice != "12000" {
		t.Errorf("first=%q last=%q, want 10000 and 12000", got.FirstPrice, got.LastPrice)
	}
	if got.GainPct != "20" {
		t.Errorf("gain = %s%%, want 20%%", got.GainPct)
	}
	// Skipped, not ignored. A rising count is a source contract moving, and it is
	// the only trace of it once the measurement carries on regardless.
	if got.Unreadable != 1 {
		t.Errorf("unreadable readings = %d, want 1 — the malformed row left no trace",
			got.Unreadable)
	}
}

// TestAFirstPriceOfZeroCannotBeADenominator.
//
// D9 states the rule for acceleration — a denominator that cannot be used does
// not produce a value, and the value it would produce is not a pass. This is the
// same rule on the same shape of arithmetic: 0 gives +Inf and a negative gives a
// sign inversion, and both clear every threshold they are compared against.
func TestAFirstPriceOfZeroCannotBeADenominator(t *testing.T) {
	for _, text := range []string{"0", "0.0", "-100"} {
		got := MeasureExpansion(storedBaseline(text, t0, SourceOfficialPrices),
			[]Observation{levelObs(t0.Add(time.Minute), SourceOfficialPrices, "80000")})
		if got.Measured {
			t.Errorf("a baseline of %q was measured as %s%%", text, got.GainPct)
		}
		if got.Why != LevelNonPositivePrice {
			t.Errorf("a baseline of %q gave reason %q, want %q", text, got.Why, LevelNonPositivePrice)
		}
		if got.Ratio != "" || got.GainPct != "" {
			t.Errorf("a baseline of %q produced ratio=%q gain=%q; an unmeasured metric carries "+
				"no arithmetic at all", text, got.Ratio, got.GainPct)
		}
	}
}

// TestGainExceedsIsExactAndThreeState pins the comparison `extended` (task 4.2)
// is written against.
//
// The threshold is compared on exact rationals recomputed from the two reported
// strings, not on the rendered GainPct — which is truncated at metricScale digits
// whenever the value does not terminate in base ten. A rendering must not be what
// decides a veto.
func TestGainExceedsIsExactAndThreeState(t *testing.T) {
	measured := MeasureExpansion(storedBaseline("100", t0, SourceOfficialPrices),
		[]Observation{levelObs(t0.Add(time.Minute), SourceOfficialPrices, "130")})
	if !measured.Measured {
		t.Fatalf("unmeasured (%q)", measured.Why)
	}
	for _, tc := range []struct {
		threshold string
		want      bool
	}{
		{"29.99", true},
		{"30", false},
		{"30.01", false},
	} {
		got, ok := measured.GainExceeds(tc.threshold)
		if !ok {
			t.Fatalf("GainExceeds(%s) reports unmeasured on a measured expansion", tc.threshold)
		}
		if got != tc.want {
			t.Errorf("a 30%% gain against a threshold of %s = %v, want %v", tc.threshold, got, tc.want)
		}
	}

	// Unmeasured is not "not extended". A caller that reads the first return
	// without the second has turned the veto off for every candidate nobody has
	// priced, which is D10's whole subject.
	var unmeasured Expansion
	if exceeds, ok := unmeasured.GainExceeds("30"); ok || exceeds {
		t.Errorf("an unmeasured expansion answered the threshold: exceeds=%v measured=%v",
			exceeds, ok)
	}
}

// TestTwoSymbolsInOneSeriesIsAWiringDefectAndNotAMeasurement.
//
// These take one candidate's observations, as Store.Observations returns them.
// A caller that hands over two symbols' rows would otherwise get a confident
// number computed from one symbol's baseline and another's current price. It is
// recorded as unmeasured because unmeasured is the one state that is never read
// as safe (D10).
func TestTwoSymbolsInOneSeriesIsAWiringDefectAndNotAMeasurement(t *testing.T) {
	series := []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		obs("000660", t0.Add(time.Minute), SourceOfficialPrices, Reported{Price: "200000"}),
	}
	got := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), series)
	if got.Measured {
		t.Errorf("expansion measured across two symbols: %s%%", got.GainPct)
	} else if got.Why != LevelMixedCandidates {
		t.Errorf("reason = %q, want %q", got.Why, LevelMixedCandidates)
	}
	if r := MeasureRangePosition(series); r.Measured {
		t.Errorf("range position measured across two symbols: %s%%", r.DistancePct)
	} else if r.Why != LevelMixedCandidates {
		t.Errorf("reason = %q, want %q", r.Why, LevelMixedCandidates)
	}
}

// --- task 3.7: where the price sits in the day's range -------------------------

// TestRangePositionIsUnmeasuredWhenNoCandleEverArrived is the most important test
// in this file.
//
// The intraday high is not in the ranking response and not in the prices response
// (`lastPrice` only). It comes from candles, one call per symbol at five calls a
// second, and the hot queue spends that budget on a fraction of the candidates
// (D13 decision 3). So a candidate with no day high is not an edge case — it is
// the ordinary state of most of the list, all day.
//
// If that state produces a number, near_high answers "not dangerous" for most of
// the watchlist and the screen says "no chase risk" where the truth is "never
// checked". Nothing fails while this happens, which is why it is a test and not
// a review comment.
func TestRangePositionIsUnmeasuredWhenNoCandleEverArrived(t *testing.T) {
	got := MeasureRangePosition([]Observation{
		obs("005930", t0, SourceOfficialTradingValue,
			Reported{Rank: 3, RankTotal: 100, TradingValue: "1000", Price: "70000"}),
		levelObs(t0.Add(30*time.Second), SourceOfficialPrices, "72000"),
	})
	if got.Measured {
		t.Fatalf("range position measured (%s%%) from readings that carried no day high; "+
			"near_high would report \"not dangerous\" for a candidate nobody looked at",
			got.DistancePct)
	}
	if got.Why != LevelNoDayHigh {
		t.Errorf("reason = %q, want %q — 미측정 has to name its missing input", got.Why, LevelNoDayHigh)
	}
	// And the arithmetic must not have happened at all. There is no NaN to worry
	// about now that the value is a decimal string, and the property that replaces
	// it is the same one: empty, not "0". A zero distance is a price sitting
	// exactly on the day's high — the most dangerous position there is.
	if got.DistancePct != "" {
		t.Errorf("DistancePct = %q; an unmeasured position carries no arithmetic", got.DistancePct)
	}
	// The veto cannot be answered from it either.
	if near, ok := got.NearHigh("2.0"); ok || near {
		t.Errorf("an unmeasured position answered near_high: near=%v measured=%v", near, ok)
	}
}

// TestTheNearHighBoundaryHoldsExactly is tasks.md 4.3's boundary, pinned on the
// arithmetic that decides it.
//
//	1.99% true / 2.00% false / 2.01% false, and the comparison is `<`.
//
// The last row is why this is not a float64. (2.50 − 2.45) ÷ 2.50 × 100 is exactly
// 2, and in float64 it is 1.99999999999999267253, which fires a veto the contract
// says must not fire. It is not a curiosity: of the 20,000 two-decimal price pairs
// below $10,000 whose exact distance is 2.00%, 9,598 come out below the threshold
// in float64. Whole-unit KRW prices are all clean, so nothing showed on the market
// this was first tested against.
func TestTheNearHighBoundaryHoldsExactly(t *testing.T) {
	for _, tc := range []struct {
		high, price  string
		wantDistance string
		wantNear     bool
		why          string
	}{
		{"100", "98.01", "1.99", true, "1.99% is inside the threshold"},
		{"100", "98", "2", false, "2.00% is the boundary and the comparison is <"},
		{"100", "97.99", "2.01", false, "2.01% is outside"},
		{"2.50", "2.45", "2", false, "exactly 2.00% on US cents; float64 says 1.9999999999999927"},
	} {
		got := MeasureRangePosition([]Observation{candleObs(t0, tc.price, tc.high, "0.01")})
		if !got.Measured {
			t.Fatalf("high=%s price=%s unmeasured (%q)", tc.high, tc.price, got.Why)
		}
		if got.DistancePct != tc.wantDistance {
			t.Errorf("high=%s price=%s: distance = %s%%, want %s%%",
				tc.high, tc.price, got.DistancePct, tc.wantDistance)
		}
		near, ok := got.NearHigh("2.0")
		if !ok {
			t.Fatalf("high=%s price=%s: NearHigh reports unmeasured", tc.high, tc.price)
		}
		if near != tc.wantNear {
			t.Errorf("high=%s price=%s: near_high = %v, want %v (%s)",
				tc.high, tc.price, near, tc.wantNear, tc.why)
		}
	}
}

// TestAZeroDayHighIsNotADayHigh.
//
// The package doc names this one: "a zero day high would make every
// high-proximity comparison pass, which is the quiet form of turning a veto off".
// The store keeps the column nullable so absence stays absent, and this is the
// same rule one layer up — a source that starts sending 0 for a field it used to
// omit must not be believed.
func TestAZeroDayHighIsNotADayHigh(t *testing.T) {
	for _, high := range []string{"0", "0.00", "-1"} {
		got := MeasureRangePosition([]Observation{candleObs(t0, "70000", high, "0")})
		if got.Measured {
			t.Errorf("a day high of %q was measured: %s%%", high, got.DistancePct)
		}
		if got.Why != LevelNonPositiveDayHigh {
			t.Errorf("a day high of %q gave reason %q, want %q", high, got.Why, LevelNonPositiveDayHigh)
		}
	}
}

// TestADayHighThatIsNotANumberIsNotADayHigh.
func TestADayHighThatIsNotANumberIsNotADayHigh(t *testing.T) {
	got := MeasureRangePosition([]Observation{candleObs(t0, "70000", "seventy thousand", "60000")})
	if got.Measured {
		t.Fatalf("an unreadable day high was measured: %s%%", got.DistancePct)
	}
	if got.Why != LevelUnreadableDecimal {
		t.Errorf("reason = %q, want %q", got.Why, LevelUnreadableDecimal)
	}
}

// TestTheDistanceIsMeasuredFromTheHighTheCandleReported.
//
// 71000 against a high of 72000 is 1.3888…% below it — 25/18, which does not
// terminate in base ten, so it is rendered truncated at metricScale digits the
// same way every other non-terminating value in this package is. The veto's own
// comparison does not use the rendering (see NearHigh).
func TestTheDistanceIsMeasuredFromTheHighTheCandleReported(t *testing.T) {
	got := MeasureRangePosition([]Observation{candleObs(t0, "71000", "72000", "68000")})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.DistancePct != "1.388888888888" {
		t.Errorf("distance = %s%%, want 1.388888888888%%", got.DistancePct)
	}
	if got.High != "72000" || got.Price != "71000" || got.Low != "68000" {
		t.Errorf("the reported strings did not survive: high=%q price=%q low=%q",
			got.High, got.Price, got.Low)
	}
	if !got.At.Equal(t0) || got.Source != SourceOfficialCandles {
		t.Errorf("the measurement is dated %v from %q, want %v from %q — task 4.7 needs the "+
			"instant to decide whether this is a measurement or a memory",
			got.At, got.Source, t0, SourceOfficialCandles)
	}
}

// TestTheFreshestCandleIsTheOneMeasured.
//
// A candle is one call per symbol at five a second, so a candidate that gets one
// at all gets it rarely and the old readings stay in the table. Measuring from
// the oldest surviving candle would report the morning's range at the close.
func TestTheFreshestCandleIsTheOneMeasured(t *testing.T) {
	got := MeasureRangePosition([]Observation{
		candleObs(t0.Add(2*time.Minute), "80000", "80500", "70000"),
		candleObs(t0, "70000", "71000", "69000"),
		levelObs(t0.Add(3*time.Minute), SourceOfficialPrices, "81000"),
	})
	if !got.Measured {
		t.Fatalf("unmeasured (%q)", got.Why)
	}
	if got.High != "80500" || !got.At.Equal(t0.Add(2*time.Minute)) {
		t.Errorf("measured from the candle at %v with high %q, want the freshest one",
			got.At, got.High)
	}
}

// TestTheHighAndThePriceComeFromTheSameReading.
//
// The tempting alternative is to take the freshest price from anywhere and the
// freshest high from a candle. That mixes two instants into one percentage, and
// it fails in the unsafe direction: when the price has fallen since the candle,
// the distance grows and near_high answers "not dangerous" on the strength of a
// high that is minutes old.
//
// One reading, one instant, one source — which is also what makes task 4.7's age
// limit a single comparison rather than a pair of them.
func TestTheHighAndThePriceComeFromTheSameReading(t *testing.T) {
	got := MeasureRangePosition([]Observation{
		// A candle with a high and no price of its own.
		obs("005930", t0, SourceOfficialCandles, Reported{DayHigh: "80000", DayLow: "70000"}),
		// A fresher price from somewhere else. Borrowing it would produce 1.25%.
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "79000"),
	})
	if got.Measured {
		t.Fatalf("a high from one reading was combined with a price from another: %s%%",
			got.DistancePct)
	}
	if got.Why != LevelNoPrice {
		t.Errorf("reason = %q, want %q", got.Why, LevelNoPrice)
	}
}

// TestAPriceAtOrAboveTheRecordedHighIsNotFurtherFromIt.
//
// The last trade in a candle can equal the candle's own high, and a later candle
// can arrive with a price that has already passed the previous high. The distance
// is then zero or negative, and a negative distance is the most dangerous
// position there is rather than a missing measurement — so it is reported as the
// number it is, and the near_high comparison (`<`) reads it correctly.
func TestAPriceAtOrAboveTheRecordedHighIsNotFurtherFromIt(t *testing.T) {
	at := MeasureRangePosition([]Observation{candleObs(t0, "72000", "72000", "70000")})
	if !at.Measured {
		t.Fatalf("a price exactly at the high is unmeasured (%q)", at.Why)
	}
	if at.DistancePct != "0" {
		t.Errorf("distance at the high = %s%%, want 0", at.DistancePct)
	}
	if near, ok := at.NearHigh("2.0"); !ok || !near {
		t.Errorf("a price on the day's high is not near it: near=%v measured=%v", near, ok)
	}

	above := MeasureRangePosition([]Observation{candleObs(t0, "73000", "72000", "70000")})
	if !above.Measured {
		t.Fatalf("a price above the recorded high is unmeasured (%q)", above.Why)
	}
	if above.DistancePct != "-1.388888888888" {
		t.Errorf("distance above the high = %s%%, want a negative number — a price past the "+
			"high must not read as further from it", above.DistancePct)
	}
	if near, ok := above.NearHigh("2.0"); !ok || !near {
		t.Errorf("a price past the day's high is not near it: near=%v measured=%v", near, ok)
	}
}

// TestThePositionInsideTheRangeNeedsBothEnds.
//
// Position is the literal range position: 0 at the day's low, 1 at its high. It
// needs a low, and a low can be absent on its own — so it carries its own
// measured flag rather than borrowing the distance's. Collapsing the two would
// mean a missing low either suppressed the distance (losing the veto's input) or
// produced a position computed against a low of zero, which puts every price at
// the top of its range.
func TestThePositionInsideTheRangeNeedsBothEnds(t *testing.T) {
	both := MeasureRangePosition([]Observation{candleObs(t0, "75000", "80000", "70000")})
	if !both.PositionMeasured {
		t.Fatalf("a reading with both ends has no position")
	}
	if both.Position != "0.5" {
		t.Errorf("position = %s, want 0.5", both.Position)
	}

	noLow := MeasureRangePosition([]Observation{
		obs("005930", t0, SourceOfficialCandles, Reported{Price: "75000", DayHigh: "80000"}),
	})
	if !noLow.Measured {
		t.Fatalf("a missing low suppressed the distance from the high (%q); the veto's input "+
			"does not need a low", noLow.Why)
	}
	if noLow.PositionMeasured {
		t.Errorf("a position of %s was computed with no low", noLow.Position)
	}
	if noLow.Position != "" {
		t.Errorf("an unmeasured position carries the value %q", noLow.Position)
	}
}

// TestADegenerateRangeHasNoPositionButStillHasADistance.
//
// high == low happens: a halted symbol, a limit-up print, the first candle of the
// session. The position is 0/0 there and the distance is still a fact.
func TestADegenerateRangeHasNoPositionButStillHasADistance(t *testing.T) {
	got := MeasureRangePosition([]Observation{candleObs(t0, "70000", "70000", "70000")})
	if !got.Measured {
		t.Fatalf("unmeasured (%q); the distance from the high is 0 and that is a measurement", got.Why)
	}
	if got.PositionMeasured {
		t.Errorf("a position of %s was computed from a zero-width range", got.Position)
	}
	if got.Position != "" {
		t.Errorf("Position = %q; 0/0 was evaluated", got.Position)
	}
}

// TestNoObservationsHasItsOwnReasonForTheRangeToo.
func TestNoObservationsHasItsOwnReasonForTheRangeToo(t *testing.T) {
	got := MeasureRangePosition(nil)
	if got.Measured {
		t.Fatal("range position measured from an empty series")
	}
	if got.Why != LevelNoObservations {
		t.Errorf("reason = %q, want %q", got.Why, LevelNoObservations)
	}
}

// TestEveryUnmeasuredAnswerNamesItsMissingInput.
//
// A reason code that is sometimes empty is a reason code an operator learns to
// ignore, and the screen requirement (spec: 미측정을 안전으로 보이게 해서는 안 된다)
// depends on being able to say *why* something was not measured.
//
// Reason() rather than Why, because the zero value is one of the cases and it is
// the one a struct nobody measured produces.
func TestEveryUnmeasuredAnswerNamesItsMissingInput(t *testing.T) {
	baselines := []Baseline{
		{},
		storedBaseline("0", t0, SourceOfficialPrices),
		storedBaseline("abc", t0, SourceOfficialPrices),
		storedBaseline("10000", t0, SourceOfficialPrices),
	}
	series := [][]Observation{
		nil,
		{obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 1, RankTotal: 10})},
		{levelObs(t0, SourceOfficialPrices, "0")},
		{levelObs(t0, SourceOfficialPrices, "abc")},
		{
			levelObs(t0, SourceOfficialPrices, "10000"),
			obs("000660", t0, SourceOfficialPrices, Reported{Price: "20000"}),
		},
	}
	for i, in := range series {
		for j, b := range baselines {
			if e := MeasureExpansion(b, in); !e.Measured && e.Reason() == "" {
				t.Errorf("series %d baseline %d: expansion is unmeasured with no reason", i, j)
			}
		}
		if r := MeasureRangePosition(in); !r.Measured && r.Reason() == "" {
			t.Errorf("series %d: range position is unmeasured with no reason", i)
		}
	}
	// The zero values, which no series can produce.
	if (Expansion{}).Reason() == "" {
		t.Error("the zero Expansion is unmeasured with no reason")
	}
	if (RangePosition{}).Reason() == "" {
		t.Error("the zero RangePosition is unmeasured with no reason")
	}
}
