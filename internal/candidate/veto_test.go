package candidate

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// rankedObs is a ranking row: the only kind of reading that carries a list
// position, which is what seen_late is measured from.
func rankedObs(at time.Time, source SourceID, rank, total int) Observation {
	return obs("005930", at, source, Reported{Rank: rank, RankTotal: total, Price: "70000"})
}

// storedFirstRank is what Store.FirstRank returns for a candidate whose first
// ranked observation has been recorded (task 4.9). Every measured sighting in this
// file supplies one, because a candidate without one is not measurable — the same
// contract storedBaseline carries for the expansion, and pinned the same way in
// TestASightingWithNoStoredRankIsNotMeasuredFromARow.
//
// Corrected 2026-07-28 for the two qualifiers the position now carries. Both are
// spelled as the ordinary measured case — the source held a previous reading and
// this symbol was in it, and the reading arrived whole — because that is the state
// in which a first sighting means what these tests say it means. Leaving them at
// their zero values would make every sighting in this file unmeasured, which would
// have made the tests pass by measuring nothing.
//
// The other two states are exercised where they are the subject:
// TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate and
// TestATruncatedReadingsPositionIsNotAPercentile build their own FirstRank.
func storedFirstRank(rank, total int, at time.Time, source SourceID) FirstRank {
	return FirstRank{
		Rank: rank, Total: total, At: at, Source: source,
		NewlyListed: NewlyListedNo(), Requested: total,
	}
}

// aCandidate is the summary the vetoes judge their inputs' lateness against.
func aCandidate(firstSeen time.Time) Candidate {
	return Candidate{
		Key:         Key{Market: MarketKR, Symbol: "005930"},
		State:       StateActive,
		FirstSeenAt: firstSeen,
		LastSeenAt:  firstSeen,
	}
}

// vetoThresholds is a fully configured set. Only near_high's value comes from the
// contract (design.md 결정된 계약값); the other two are the caller's, because the
// contract fixes no number for them — see VetoThresholds.
func vetoThresholds() VetoThresholds {
	return VetoThresholds{
		SeenLatePercentilePct: "80",
		ExtendedGainPct:       "50",
		NearHighDistancePct:   DefaultNearHighThresholdPct,
	}
}

// --- task 4.4: three states, and the third one is never the second -------------

// TestTheZeroValueOfEveryVetoIsUnmeasured is D10 aimed at the type itself.
//
// The §3 review found Acceleration{}.Computed() == true — a struct nobody
// assigned reporting itself as measured — inside the section built to enforce
// D10. The natural spellings that produce it are all here: a map lookup that
// misses, a slot in make([]Chase, n) that no read filled, a field left unset for
// a candidate outside the hot queue. Each would otherwise read as "we checked
// this one and it is not chasing".
func TestTheZeroValueOfEveryVetoIsUnmeasured(t *testing.T) {
	var v VetoState
	if v.Measured() {
		t.Error("the zero VetoState reports itself as measured")
	}
	if v.Clear() {
		t.Error("the zero VetoState reports itself as clear; a veto nobody ran would read as a " +
			"candidate somebody checked and found safe")
	}
	if v.Dangerous() {
		t.Error("the zero VetoState reports itself as dangerous")
	}
	if got := v.Reason(); got != VetoNotEvaluated {
		t.Errorf("the zero VetoState's reason = %q, want %q", got, VetoNotEvaluated)
	}
	var s Sighting
	if s.Measured {
		t.Error("the zero Sighting reports itself as measured")
	}
	if got := s.Reason(); got != VetoNotEvaluated {
		t.Errorf("the zero Sighting's reason = %q, want %q", got, VetoNotEvaluated)
	}
}

// TestTheZeroChaseDoesNotPassTheVeto is the same rule one level up.
//
// Passed has to mean "all three measured and all three clear". Spelled as
// "nothing fired" it is true of a verdict nobody computed, which is the majority
// of the list under the candle budget.
func TestTheZeroChaseDoesNotPassTheVeto(t *testing.T) {
	var c Chase
	if c.Passed() {
		t.Error("the zero Chase passes the veto; a candidate nobody assessed would be counted " +
			"as one that cleared")
	}
	if !c.HasUnmeasured() {
		t.Error("the zero Chase reports nothing unmeasured")
	}
	if got := len(c.NotMeasured()); got != len(OrderedVetoCodes()) {
		t.Errorf("the zero Chase reports %d unmeasured codes, want all %d", got, len(OrderedVetoCodes()))
	}
	if c.Vetoed() {
		t.Error("the zero Chase reports a fired veto; unmeasured is neither of the other two")
	}
}

// TestAVetoMissingFromAMapIsNotAPass is the first of the three spellings the §3
// review named, written as the code that produces it.
func TestAVetoMissingFromAMapIsNotAPass(t *testing.T) {
	assessed := map[Key]Chase{
		{Market: MarketKR, Symbol: "005930"}: {
			SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto(),
		},
	}
	if !assessed[Key{Market: MarketKR, Symbol: "005930"}].Passed() {
		t.Fatal("the candidate that was assessed does not pass; the test cannot show the miss")
	}
	if assessed[Key{Market: MarketKR, Symbol: "000660"}].Passed() {
		t.Error("a candidate that is not in the map passes the veto; every candidate outside " +
			"the hot queue is exactly this lookup")
	}
}

// TestASliceSlotNobodyFilledIsNotAPass is the second spelling.
func TestASliceSlotNobodyFilledIsNotAPass(t *testing.T) {
	assessed := make([]Chase, 3)
	assessed[0] = Chase{SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto()}
	for i, c := range assessed {
		if i == 0 {
			continue
		}
		if c.Passed() {
			t.Errorf("slot %d passes the veto without anything having been measured into it", i)
		}
	}
}

// TestPassingTheVetoNeedsAllThreeMeasuredAndClear walks all twenty-seven
// combinations.
//
// A default branch in a switch is one of the ways an unmeasured becomes a false,
// and it is invisible in review because it is one line that reads as tidiness.
// Enumerating the states closes it: there is no combination left for a default to
// answer for.
func TestPassingTheVetoNeedsAllThreeMeasuredAndClear(t *testing.T) {
	states := map[string]VetoState{
		"raised":     RaisedVeto(),
		"clear":      ClearedVeto(),
		"unmeasured": UnmeasuredVeto(noDayHighReason()),
	}
	names := []string{"raised", "clear", "unmeasured"}
	for _, a := range names {
		for _, b := range names {
			for _, c := range names {
				chase := Chase{SeenLate: states[a], Extended: states[b], NearHigh: states[c]}
				wantPass := a == "clear" && b == "clear" && c == "clear"
				if chase.Passed() != wantPass {
					t.Errorf("Chase{%s, %s, %s}.Passed() = %v, want %v",
						a, b, c, chase.Passed(), wantPass)
				}
				wantVetoed := a == "raised" || b == "raised" || c == "raised"
				if chase.Vetoed() != wantVetoed {
					t.Errorf("Chase{%s, %s, %s}.Vetoed() = %v, want %v",
						a, b, c, chase.Vetoed(), wantVetoed)
				}
				wantUnmeasured := a == "unmeasured" || b == "unmeasured" || c == "unmeasured"
				if chase.HasUnmeasured() != wantUnmeasured {
					t.Errorf("Chase{%s, %s, %s}.HasUnmeasured() = %v, want %v",
						a, b, c, chase.HasUnmeasured(), wantUnmeasured)
				}
				// And the three predicates cannot disagree with each other.
				if chase.Passed() && (chase.Vetoed() || chase.HasUnmeasured()) {
					t.Errorf("Chase{%s, %s, %s} passes and is also vetoed or unmeasured",
						a, b, c)
				}
			}
		}
	}
}

// noDayHighReason is the reason a candidate with no candle carries, spelled
// once so the tests and the measurement cannot drift apart.
func noDayHighReason() VetoUnmeasured { return VetoUnmeasured(LevelNoDayHigh) }

// TestACandidateWithNoCandleDoesNotPassTheNearHighVeto is the most important test
// in this change.
//
// The intraday high comes only from candles, one call per symbol at five a
// second, so under a list of hundreds most candidates never get one (D13 decision
// 3). Storing false for them switches the veto off for the majority of the list
// while the screen says "not chasing" — D10's second route around D3, and the one
// nothing fails on.
func TestACandidateWithNoCandleDoesNotPassTheNearHighVeto(t *testing.T) {
	// Everything a ranking gives us, and nothing a candle would have.
	readings := []Observation{
		rankedObs(t0, SourceOfficialTradingValue, 140, 150),
		rankedObs(t0.Add(time.Minute), SourceOfficialTradingValue, 138, 150),
	}
	position := MeasureRangePosition(readings)
	if position.Measured {
		t.Fatalf("the readings produced a measured range position (%s%%); the test needs one "+
			"that never got a candle", position.DistancePct)
	}
	if position.Why != LevelNoDayHigh {
		t.Fatalf("range position reason = %q, want %q", position.Why, LevelNoDayHigh)
	}

	got := AssessNearHigh(position, t0.Add(time.Minute), vetoThresholds())
	if got.Clear() {
		t.Error("near_high reads as clear for a candidate that never got a candle; the record " +
			"would say 'no chase risk' where the truth is 'never checked'")
	}
	if got.Measured() {
		t.Error("near_high reports itself as measured without an intraday high")
	}
	if want := noDayHighReason(); got.Reason() != want {
		t.Errorf("reason = %q, want %q — the measurement's own name has to survive to the verdict",
			got.Reason(), want)
	}

	// And it is not a pass, however clear the other two are.
	chase := Chase{SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: got}
	if chase.Passed() {
		t.Error("the candidate passes the chase veto with near_high never measured")
	}
	if codes := chase.NotMeasured(); len(codes) != 1 || codes[0] != VetoNearHigh {
		t.Errorf("unmeasured codes = %v, want [%s]", codes, VetoNearHigh)
	}
}

// TestAnUnmeasuredVetoNeverRendersAsNothing.
//
// §5 prints the reason beside the verdict, and "unmeasured ()" on the screen
// whose entire job is to stop an unmeasured veto from looking like a passed one
// teaches an operator to skip the column. level.go pins the same rule for its own
// metrics.
func TestAnUnmeasuredVetoNeverRendersAsNothing(t *testing.T) {
	for _, v := range []VetoState{{}, UnmeasuredVeto("")} {
		if got := v.Reason(); got == "" {
			t.Error("an unmeasured veto names no reason")
		}
		if got := v.String(); got != "unmeasured ("+string(VetoNotEvaluated)+")" {
			t.Errorf("String() = %q, want it to name the reason", got)
		}
	}
	if got := ClearedVeto().Reason(); got != "" {
		t.Errorf("a measured veto gives the reason %q, want none", got)
	}
	if got := RaisedVeto().String(); got != "raised" {
		t.Errorf("String() = %q, want %q", got, "raised")
	}
}

// TestAnAbsentThresholdIsNotAPassedVeto.
//
// The contract fixes only near_high's number, so the other two arrive from a
// caller and a caller can forget. Comparing against an empty threshold with the
// obvious parse-or-zero produces `> 0`, which is true of every candidate, or
// `< 0`, which is true of none — and the second is a silent pass for the whole
// list.
func TestAnAbsentThresholdIsNotAPassedVeto(t *testing.T) {
	in := VetoInputs{
		Candidate: aCandidate(t0),
		Sighting: MeasureFirstSighting(aCandidate(t0),
			storedFirstRank(12, 150, t0, SourceOfficialTradingValue),
			[]Observation{rankedObs(t0, SourceOfficialTradingValue, 12, 150)}),
		Expansion: MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
			levelObs(t0, SourceOfficialPrices, "10000"),
			levelObs(t0.Add(time.Minute), SourceOfficialPrices, "11000"),
		}),
		Range: MeasureRangePosition([]Observation{candleObs(t0.Add(time.Minute), "71000", "72000", "68000")}),
		At:    t0.Add(time.Minute),
		// Thresholds left at their zero value.
	}
	got := AssessChase(in)
	if got.Passed() {
		t.Fatal("a candidate assessed against no thresholds at all passes the chase veto")
	}
	for _, code := range OrderedVetoCodes() {
		state := got.State(code)
		if state.Measured() {
			t.Errorf("%s is measured against an absent threshold", code)
		}
		if state.Reason() != VetoThresholdAbsent {
			t.Errorf("%s reason = %q, want %q", code, state.Reason(), VetoThresholdAbsent)
		}
	}
}

// TestAnUnreadableThresholdIsNotAPassedVeto keeps the two apart for parseFigure's
// reason: an absent threshold is a configuration nobody wrote, an unreadable one
// is a configuration somebody broke, and the remedies differ.
func TestAnUnreadableThresholdIsNotAPassedVeto(t *testing.T) {
	th := VetoThresholds{
		SeenLatePercentilePct: "eighty",
		ExtendedGainPct:       "1e2",
		NearHighDistancePct:   "2,0",
	}
	position := MeasureRangePosition([]Observation{candleObs(t0, "71000", "72000", "68000")})
	got := AssessNearHigh(position, t0, th)
	if got.Measured() {
		t.Error("near_high is measured against a threshold that is not a number")
	}
	if got.Reason() != VetoThresholdUnreadable {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoThresholdUnreadable)
	}
}

// TestAMeasuredInputWithUnusableFiguresIsStillNotAPass closes the ignored second
// return.
//
// level.go answers the two comparisons as (value, measured) pairs, and
// `near, _ := r.NearHigh(th)` compiles and turns every unmeasured candidate into
// a clear one. Expansion and RangePosition have exported fields, so a caller —
// §5's fixtures, a future lane — can hand over a struct that says Measured and
// carries figures that are not numbers. That reaches the second return without
// going through the measurement, which is the only way to test that anybody is
// reading it.
func TestAMeasuredInputWithUnusableFiguresIsStillNotAPass(t *testing.T) {
	want := VetoUnmeasured(LevelUnreadableDecimal)

	extended := AssessExtended(
		Expansion{Measured: true, FirstPrice: "N/A", LastPrice: "1,200", FirstAt: t0, LastAt: t0},
		aCandidate(t0), t0, vetoThresholds())
	if extended.Clear() {
		t.Error("extended reads as clear from prices that are not numbers; the measured return " +
			"of GainExceeds was dropped")
	}
	if extended.Reason() != want {
		t.Errorf("extended = %v, want unmeasured (%s)", extended, want)
	}
	near := AssessNearHigh(
		RangePosition{Measured: true, High: "N/A", Price: "1,200", At: t0},
		t0, vetoThresholds())
	if near.Clear() {
		t.Error("near_high reads as clear from a high that is not a number; the measured return " +
			"of NearHigh was dropped")
	}
	if near.Reason() != want {
		t.Errorf("near_high = %v, want unmeasured (%s)", near, want)
	}
}

// TestTheStateOfAnUnknownCodeIsNotAPass closes the last default branch.
func TestTheStateOfAnUnknownCodeIsNotAPass(t *testing.T) {
	c := Chase{SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto()}
	got := c.State(VetoCode("chase_risk"))
	if got.Clear() || got.Measured() {
		t.Error("an unrecognised veto code answers as measured and clear")
	}
	if got.Reason() == "" {
		t.Error("an unrecognised veto code answers with no reason")
	}
}

// --- task 4.3: the sign of near_high -------------------------------------------

// TestTheNearHighVetoFiresBelowTheThresholdAndNotAtIt.
//
// D3's table said "상한 초과" for years of drafts, which reads as `>` and inverts
// the veto: it fires on candidates far from the high and passes the ones sitting
// on it. The stored quantity is the *distance* to the day's high and smaller is
// more dangerous, so the comparison is `<` and the boundary is pinned to the
// digit — 1.99% raises it, 2.00% does not.
//
// The last row is why the arithmetic is on exact rationals. (2.50 − 2.45) ÷ 2.50
// × 100 is exactly 2, and in float64 it is 1.99999999999999267253 — one of the
// 9,598 US cent pairs the §3 review swept that fire a veto the contract says must
// not fire. level.go already pins this for RangePosition.NearHigh; this pins that
// the veto reaches for that comparison rather than deriving its own.
func TestTheNearHighVetoFiresBelowTheThresholdAndNotAtIt(t *testing.T) {
	for _, tc := range []struct {
		high, price string
		want        bool
		why         string
	}{
		{"100", "98.01", true, "1.99% below the high is inside the threshold"},
		{"100", "98", false, "2.00% is the boundary and the comparison is <"},
		{"100", "97.99", false, "2.01% is outside"},
		{"2.50", "2.45", false, "exactly 2.00% on US cents; float64 makes it 1.99999999999999"},
		{"100", "100", true, "a price sitting on the day's high is the most dangerous position"},
	} {
		position := MeasureRangePosition([]Observation{candleObs(t0, tc.price, tc.high, "1")})
		if !position.Measured {
			t.Fatalf("high=%s price=%s: unmeasured (%q)", tc.high, tc.price, position.Why)
		}
		got := AssessNearHigh(position, t0, vetoThresholds())
		if !got.Measured() {
			t.Fatalf("high=%s price=%s: veto unmeasured (%q)", tc.high, tc.price, got.Reason())
		}
		if got.Dangerous() != tc.want {
			t.Errorf("high=%s price=%s (distance %s%%): near_high = %v, want %v — %s",
				tc.high, tc.price, position.DistancePct, got.Dangerous(), tc.want, tc.why)
		}
	}
}

// --- task 4.7: an input age ceiling --------------------------------------------

// TestACandleOlderThanTheCeilingIsAMemoryAndNotAMeasurement.
//
// A candle costs one call per symbol at five a second, so a high that was fetched
// once gets reused for as long as the row is retained — two trading days. The
// price in that same reading is what the distance is measured from, so a reading
// that has aged says nothing about where the price is now, and the error points
// at "not dangerous": the price has risen since and the recorded distance has
// not. Without the ceiling the failure task 4.4 blocks reappears one level down
// and looks identical on a screen.
func TestACandleOlderThanTheCeilingIsAMemoryAndNotAMeasurement(t *testing.T) {
	// 4% below the day's high: comfortably clear while it is fresh.
	position := MeasureRangePosition([]Observation{candleObs(t0, "96", "100", "80")})
	if !position.Measured {
		t.Fatalf("unmeasured (%q)", position.Why)
	}

	fresh := AssessNearHigh(position, t0.Add(MaxVetoInputAge-time.Second), vetoThresholds())
	if !fresh.Clear() {
		t.Errorf("a candle one second inside the ceiling is %v, want clear", fresh)
	}
	stale := AssessNearHigh(position, t0.Add(MaxVetoInputAge), vetoThresholds())
	if stale.Clear() {
		t.Error("a candle at the ceiling still reads as 'not near the high'; a not-dangerous " +
			"built from an old reading is a memory, not a measurement")
	}
	if stale.Reason() != VetoInputTooOld {
		t.Errorf("reason = %q, want %q", stale.Reason(), VetoInputTooOld)
	}
	// The three minutes tasks.md 4.7 names is on the wrong side of the ceiling.
	if got := AssessNearHigh(position, t0.Add(3*time.Minute), vetoThresholds()); got.Measured() {
		t.Error("a three-minute-old candle is still counted as a measurement")
	}
}

// TestAnAbsentInputAgeLimitIsTheDefaultAndNotNoLimit.
//
// A zero duration meaning "no ceiling" is the same class of defect as a zero
// DayHigh meaning "at the high": the field nobody filled in turns the guard off.
//
// # Both halves, because one of them alone is satisfied by the opposite defect
//
// The §4 review found this test green with inputAge's fallback deleted entirely.
// A ceiling of 0 or −1h makes `age >= ceiling` true of every non-negative age, so
// "an hour-old candle is unmeasured" is equally satisfied by "the ceiling rejects
// everything" — which is the §3 ZERO_ELAPSED_SECONDS pattern verbatim: an
// assertion that holds under both the fix and its inverse. So the reading inside
// the default is asserted too, and it is the half that dies under the mutant.
func TestAnAbsentInputAgeLimitIsTheDefaultAndNotNoLimit(t *testing.T) {
	position := MeasureRangePosition([]Observation{candleObs(t0, "96", "100", "80")})
	for _, ceiling := range []time.Duration{0, -time.Hour} {
		th := vetoThresholds()
		th.MaxInputAge = ceiling
		if got := AssessNearHigh(position, t0.Add(time.Hour), th); got.Measured() {
			t.Errorf("an hour-old candle is measured under a ceiling of %v: %v", ceiling, got)
		} else if got.Reason() != VetoInputTooOld {
			t.Errorf("under a ceiling of %v: reason = %q, want %q", ceiling, got.Reason(),
				VetoInputTooOld)
		}
		// And the default is a two-minute ceiling rather than a zero one: a reading
		// one second inside MaxVetoInputAge is still a measurement.
		fresh := AssessNearHigh(position, t0.Add(MaxVetoInputAge-time.Second), th)
		if !fresh.Clear() {
			t.Errorf("a candle one second inside the default ceiling is %v under a configured "+
				"ceiling of %v, want clear — a bound nobody set is MaxVetoInputAge, not zero, "+
				"and a zero one rejects every reading there is", fresh, ceiling)
		}
	}
}

// TestAReadingLaterThanTheInstantAskedAboutIsNotAMeasurement.
//
// A negative age passes every ceiling there is. metrics.go refuses the same shape
// under READINGS_ALL_LATER rather than measuring from a reading that has not
// happened yet at the instant being evaluated.
func TestAReadingLaterThanTheInstantAskedAboutIsNotAMeasurement(t *testing.T) {
	position := MeasureRangePosition([]Observation{candleObs(t0.Add(time.Hour), "96", "100", "80")})
	got := AssessNearHigh(position, t0, vetoThresholds())
	if got.Measured() {
		t.Errorf("a reading an hour after the instant asked about is measured: %v", got)
	}
	if got.Reason() != VetoInputAfterInstant {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoInputAfterInstant)
	}
}

// TestAnAssessmentWithNoInstantCannotJudgeAnInputsAge.
func TestAnAssessmentWithNoInstantCannotJudgeAnInputsAge(t *testing.T) {
	position := MeasureRangePosition([]Observation{candleObs(t0, "96", "100", "80")})
	got := AssessNearHigh(position, time.Time{}, vetoThresholds())
	if got.Measured() {
		t.Errorf("a veto measured against no instant at all: %v", got)
	}
	if got.Reason() != VetoInstantUnknown {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoInstantUnknown)
	}
}

// --- task 4.1: seen_late --------------------------------------------------------

// TestASymbolAlreadyHighInItsListWhenWeFirstSawItIsSeenLate.
//
// D8: 12위로 진입한 종목과 목록 하단으로 진입한 종목은 다른 사건이고, 전자가 많다면 우리
// 스캔이 늦다는 뜻이다. The percentile is normalised by the list's own length because
// the two lists this system reads are different lengths — the official ranking
// serves up to 100 rows and the WTS popularity list 30.
//
// Corrected 2026-07-28: this comment said the KR panel returns 150 rows and quoted
// D8's 148th-place example. No panel has ever returned 150 rows. The fixture below
// keeps its synthetic 150-row list because what is under test is the arithmetic of
// normalising by whatever length arrived, and a list length in a fixture is not a
// claim about the sources — the comment was.
func TestASymbolAlreadyHighInItsListWhenWeFirstSawItIsSeenLate(t *testing.T) {
	c := aCandidate(t0)
	late := MeasureFirstSighting(c, storedFirstRank(12, 150, t0, SourceOfficialTradingValue), []Observation{
		rankedObs(t0, SourceOfficialTradingValue, 12, 150),
		rankedObs(t0.Add(time.Minute), SourceOfficialTradingValue, 8, 150),
	})
	if !late.Measured {
		t.Fatalf("unmeasured (%q)", late.Why)
	}
	if late.PercentilePct != "92" {
		t.Errorf("percentile = %s, want 92 (12th of 150)", late.PercentilePct)
	}
	if got := AssessSeenLate(late, vetoThresholds()); !got.Dangerous() {
		t.Errorf("a symbol first seen 12th of 150 is %v, want raised", got)
	}

	early := MeasureFirstSighting(c, storedFirstRank(148, 150, t0, SourceOfficialTradingValue), []Observation{
		rankedObs(t0, SourceOfficialTradingValue, 148, 150),
		rankedObs(t0.Add(time.Minute), SourceOfficialTradingValue, 30, 150),
	})
	if !early.Measured {
		t.Fatalf("unmeasured (%q)", early.Why)
	}
	if got := AssessSeenLate(early, vetoThresholds()); !got.Clear() {
		t.Errorf("a symbol first seen 148th of 150 is %v, want clear — this is the one we caught "+
			"early, and the whole claim of the change is that it exists", got)
	}
}

// TestAFirstSightingWeNoLongerHoldIsNotAFirstSighting is D17's shape applied to
// seen_late.
//
// The rank at first sighting lives only in the observations table, which is
// pruned after two trading days (D11) and which callers window (`since`). When
// the earliest row we still hold postdates first_seen_at, the position it carries
// is somebody else's fact — and it can be anywhere in the list, so the error has
// no safe direction and the answer is unmeasured.
func TestAFirstSightingWeNoLongerHoldIsNotAFirstSighting(t *testing.T) {
	c := aCandidate(t0)
	// Everything before t0+30m has been pruned; the earliest surviving row shows
	// the symbol near the top, but that is not where it entered. Nothing was
	// stored either, which is the state a store from before schema v3 is in.
	got := MeasureFirstSighting(c, FirstRank{}, []Observation{
		rankedObs(t0.Add(30*time.Minute), SourceOfficialTradingValue, 9, 150),
		rankedObs(t0.Add(31*time.Minute), SourceOfficialTradingValue, 7, 150),
	})
	if got.Measured {
		t.Fatalf("a sighting measured (%s%%) from a reading half an hour after first_seen_at; "+
			"the surviving row silently became the first sighting", got.PercentilePct)
	}
	if got.Why != VetoNoFirstSighting {
		t.Errorf("reason = %q, want %q", got.Why, VetoNoFirstSighting)
	}
	if state := AssessSeenLate(got, vetoThresholds()); state.Clear() {
		t.Error("seen_late reads as clear for a candidate whose first sighting we no longer hold")
	}
	// It still reports what it read.
	if got.Rank != 9 || got.RankTotal != 150 {
		t.Errorf("the reading did not survive the unmeasured answer: %d/%d", got.Rank, got.RankTotal)
	}
}

// TestAPreviousLifesReadingsAreNotThisCandidatesFirstSighting.
//
// D1: a candidate that expires and crosses again is a *new* candidate with a new
// first_seen_at, and raw rows outlive it by two trading days (D11). So the
// earliest ranked row in the table can predate first_seen_at, and it belongs to a
// life that ended.
//
// The direction is the unsafe one and that is why it is pinned. The dead life
// entered at the bottom of the list and this one entered near the top, so reading
// the old row answers "we caught it early" about the entry that is actually late,
// and seen_late clears. Promote resets the first rank on expiry, so what the new
// life measures from is its own promotion's position; the old row is in the slice
// and is not consulted.
func TestAPreviousLifesReadingsAreNotThisCandidatesFirstSighting(t *testing.T) {
	// The candidate was found at 148th, cooled, expired, and crossed again two
	// hours later — this time already 5th.
	reborn := aCandidate(t0.Add(2 * time.Hour))
	got := MeasureFirstSighting(reborn,
		storedFirstRank(5, 150, t0.Add(2*time.Hour), SourceOfficialTradingValue),
		[]Observation{
			rankedObs(t0, SourceOfficialTradingValue, 148, 150),
			rankedObs(t0.Add(2*time.Hour), SourceOfficialTradingValue, 5, 150),
			rankedObs(t0.Add(2*time.Hour+time.Minute), SourceOfficialTradingValue, 4, 150),
		})
	if !got.Measured {
		t.Fatalf("unmeasured (%q); the row at first_seen_at is right there", got.Why)
	}
	if got.Rank != 5 {
		t.Fatalf("first sighting = %d of %d at %v; the previous life's row became this life's "+
			"first sighting", got.Rank, got.RankTotal, got.At)
	}
	if state := AssessSeenLate(got, vetoThresholds()); !state.Dangerous() {
		t.Errorf("seen_late = %v, want raised — a candidate that came back at 5th of 150 reads "+
			"as one we caught at 148th", state)
	}
}

// TestAReadWhoseErrorWasDroppedIsNotAPass.
//
// `rows, _ := store.Observations(...)` compiles and hands over a nil slice, and a
// scan loop is the code most likely to write it because it must keep going. Every
// measurement then has nothing to measure, and the verdict has to say so rather
// than report three quiet vetoes.
func TestAReadWhoseErrorWasDroppedIsNotAPass(t *testing.T) {
	c := aCandidate(t0)
	var dropped []Observation // what a failed read leaves behind
	got := AssessChase(VetoInputs{
		Candidate:  c,
		Sighting:   MeasureFirstSighting(c, FirstRank{}, dropped),
		Expansion:  MeasureExpansion(Baseline{}, dropped),
		Range:      MeasureRangePosition(dropped),
		At:         t0,
		Thresholds: vetoThresholds(),
	})
	if got.Passed() {
		t.Error("a candidate whose observations never arrived passes the chase veto")
	}
	if len(got.NotMeasured()) != len(OrderedVetoCodes()) {
		t.Errorf("unmeasured codes = %v, want all three", got.NotMeasured())
	}
	for _, code := range OrderedVetoCodes() {
		if got.State(code).Reason() == "" {
			t.Errorf("%s is unmeasured and names no reason", code)
		}
	}
}

// TestTheFirstSightingBoundaryIsTheStalenessTTL.
//
// Ten minutes is the same derivation DefaultStalenessTTL and MaxRankPriorAge use:
// twice the longest planned backoff, so an ordinary 429 retreat does not turn the
// veto's input into somebody else's. Closed at the near end, as stateAt's is.
//
// Task 4.9 moved what the boundary applies to. It used to choose which row became
// the first sighting; now it bounds the *stored* position's instant, and the rows
// are not consulted — the observations here are empty on purpose, which is the
// state of a candidate whose raw rows have been pruned and the state in which the
// column earns its existence.
func TestTheFirstSightingBoundaryIsTheStalenessTTL(t *testing.T) {
	c := aCandidate(t0)
	inside := MeasureFirstSighting(c,
		storedFirstRank(12, 150, t0.Add(DefaultStalenessTTL-time.Second), SourceOfficialTradingValue),
		nil)
	if !inside.Measured {
		t.Errorf("a stored first rank one second inside the TTL is unmeasured (%q)", inside.Why)
	}
	at := MeasureFirstSighting(c,
		storedFirstRank(12, 150, t0.Add(DefaultStalenessTTL), SourceOfficialTradingValue), nil)
	if at.Measured {
		t.Error("a stored first rank exactly at the TTL is still counted as the first sighting")
	}
	if at.Why != VetoFirstRankNotFirst {
		t.Errorf("reason = %q, want %q", at.Why, VetoFirstRankNotFirst)
	}
	// Both ends, because a stored rank from before first_seen_at is a previous
	// life's and has the same unsafe direction as one from after it.
	early := MeasureFirstSighting(c,
		storedFirstRank(148, 150, t0.Add(-DefaultStalenessTTL), SourceOfficialTradingValue), nil)
	if early.Measured {
		t.Error("a stored first rank a full TTL before first_seen_at is counted as this life's")
	}
	// And a stored position with no instant cannot be judged at all.
	undated := MeasureFirstSighting(c, FirstRank{Rank: 12, Total: 150}, nil)
	if undated.Measured || undated.Why != VetoFirstRankUndated {
		t.Errorf("an undated stored first rank = measured %v (%q), want unmeasured (%q)",
			undated.Measured, undated.Why, VetoFirstRankUndated)
	}
}

// TestASightingWithNoStoredRankIsNotMeasuredFromARow is task 4.9's contract, and
// it is D17's rule one veto across: the column is the measurement and a surviving
// row is never substituted for it.
//
// The four nothings are kept apart because the remedies are. NO_FIRST_RANK is the
// one this task adds: rows we hold include one that could have been the first
// sighting, and no column beside it — a store from before schema v3, or a caller
// that never wrote one. Answering from that row is exactly what this change stops.
func TestASightingWithNoStoredRankIsNotMeasuredFromARow(t *testing.T) {
	c := aCandidate(t0)
	got := MeasureFirstSighting(c, FirstRank{}, []Observation{
		rankedObs(t0, SourceOfficialTradingValue, 12, 150),
	})
	if got.Measured {
		t.Fatalf("measured (%s%%) from a row with nothing stored beside it; the row silently "+
			"became the first sighting", got.PercentilePct)
	}
	if got.Why != VetoNoFirstRank {
		t.Errorf("reason = %q, want %q", got.Why, VetoNoFirstRank)
	}
	// It still reports what it read, so a screen can show how close that row is to
	// first_seen_at.
	if got.Rank != 12 || got.RankTotal != 150 {
		t.Errorf("the reading did not survive the unmeasured answer: %d/%d", got.Rank, got.RankTotal)
	}
	if state := AssessSeenLate(got, vetoThresholds()); state.Clear() || state.Measured() {
		t.Errorf("seen_late = %v for a candidate with no stored first rank, want unmeasured", state)
	}
	// A rank of zero is not a position, so it is not a stored first rank either.
	for _, bad := range []FirstRank{
		{Rank: 0, Total: 150, At: t0}, {Rank: 12, Total: 0, At: t0}, {Rank: -1, Total: 150, At: t0},
	} {
		if MeasureFirstSighting(c, bad, nil).Measured {
			t.Errorf("%+v is treated as a stored first rank", bad)
		}
	}
}

// TestASightingWithNoRankIsUnmeasured. The prices and candles sources are not
// rankings and never carry a position, so a candidate raised only by them has no
// "where did it enter" at all — which is not the same fact as entering at the
// bottom.
func TestASightingWithNoRankIsUnmeasured(t *testing.T) {
	got := MeasureFirstSighting(aCandidate(t0), FirstRank{}, []Observation{
		levelObs(t0, SourceOfficialPrices, "70000"),
		candleObs(t0, "70000", "72000", "68000"),
	})
	if got.Measured {
		t.Fatalf("measured (%s%%) with no ranked reading at all", got.PercentilePct)
	}
	if got.Why != VetoNotRanked {
		t.Errorf("reason = %q, want %q", got.Why, VetoNotRanked)
	}
	empty := MeasureFirstSighting(aCandidate(t0), FirstRank{}, nil)
	if empty.Why != VetoNoObservations {
		t.Errorf("reason for an empty series = %q, want %q", empty.Why, VetoNoObservations)
	}
	mixed := MeasureFirstSighting(aCandidate(t0),
		storedFirstRank(12, 150, t0, SourceOfficialTradingValue), []Observation{
			rankedObs(t0, SourceOfficialTradingValue, 12, 150),
			obs("000660", t0, SourceOfficialTradingValue, Reported{Rank: 3, RankTotal: 150}),
		})
	if mixed.Measured || mixed.Why != VetoMixedCandidates {
		t.Errorf("two symbols in one slice = measured %v (%q), want unmeasured (%q) even with a "+
			"stored rank; the caller has lost track of which candidate it is assessing",
			mixed.Measured, mixed.Why, VetoMixedCandidates)
	}
}

// TestSeenLateWithoutACandidateCannotKnowWhenWeFirstSaw. Without first_seen_at
// there is nothing to check the stored position's instant against, so a stored
// rank does not rescue it.
func TestSeenLateWithoutACandidateCannotKnowWhenWeFirstSaw(t *testing.T) {
	got := MeasureFirstSighting(Candidate{},
		storedFirstRank(12, 150, t0, SourceOfficialTradingValue), []Observation{
			rankedObs(t0, SourceOfficialTradingValue, 12, 150),
		})
	if got.Measured {
		t.Fatalf("measured (%s%%) with no candidate summary to date the reading against",
			got.PercentilePct)
	}
	if got.Why != VetoNoCandidate {
		t.Errorf("reason = %q, want %q", got.Why, VetoNoCandidate)
	}
}

// TestTheSightingPercentileIsTheSameOneTheRankMoveUses.
//
// Two files computing "position as a share of its list" with two spellings is how
// a screen ends up showing a rank move of 20 points beside a first sighting at a
// percentile that cannot be compared with it.
func TestTheSightingPercentileIsTheSameOneTheRankMoveUses(t *testing.T) {
	for _, tc := range []struct{ rank, total int }{{1, 150}, {12, 150}, {75, 100}, {148, 150}} {
		mine := formatDecimal(percentileOf(tc.rank, tc.total))
		theirs := formatDecimal(percentile(obs("005930", t0, SourceOfficialTradingValue,
			Reported{Rank: tc.rank, RankTotal: tc.total})))
		if mine != theirs {
			t.Errorf("%d of %d: the sighting says %s and the rank move says %s",
				tc.rank, tc.total, mine, theirs)
		}
	}
}

// --- task 4.2 / 4.2b: extended --------------------------------------------------

// TestACandidateThatHasRunPastTheGainCeilingIsExtended.
func TestACandidateThatHasRunPastTheGainCeilingIsExtended(t *testing.T) {
	c := aCandidate(t0)
	readings := []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "16000"),
	}
	e := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), readings)
	if !e.Measured {
		t.Fatalf("unmeasured (%q)", e.Why)
	}
	if e.GainPct != "60" {
		t.Fatalf("gain = %s%%, want 60%%", e.GainPct)
	}
	if got := AssessExtended(e, c, t0.Add(time.Minute), vetoThresholds()); !got.Dangerous() {
		t.Errorf("a candidate 60%% above its first price is %v against a 50%% ceiling, want raised", got)
	}

	quiet := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "11000"),
	})
	if got := AssessExtended(quiet, c, t0.Add(time.Minute), vetoThresholds()); !got.Clear() {
		t.Errorf("a candidate 10%% above its first price is %v, want clear", got)
	}
}

// TestAnUnmeasuredExpansionIsNotAnUnextendedCandidate.
//
// The §3 review's P0 was that MeasureExpansion came back Measured with a re-based
// number after a prune. It now answers NO_BASELINE instead, and this is the half
// of that fix that lives in the veto: an unmeasured expansion must not read as
// "has not run".
func TestAnUnmeasuredExpansionIsNotAnUnextendedCandidate(t *testing.T) {
	e := MeasureExpansion(Baseline{}, []Observation{
		levelObs(t0, SourceOfficialPrices, "20000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "30000"),
	})
	got := AssessExtended(e, aCandidate(t0), t0.Add(time.Minute), vetoThresholds())
	if got.Clear() {
		t.Error("extended reads as clear for a candidate with no stored baseline")
	}
	if want := VetoUnmeasured(LevelNoBaseline); got.Reason() != want {
		t.Errorf("reason = %q, want %q — the measurement's own name has to survive", got.Reason(), want)
	}
	// The zero Expansion is the map-miss spelling of the same thing.
	zero := AssessExtended(Expansion{}, aCandidate(t0), t0.Add(time.Minute), vetoThresholds())
	if zero.Clear() || zero.Reason() == "" {
		t.Errorf("the zero Expansion produces %v", zero)
	}
}

// TestABaselineThatPostdatesFirstSightingIsNotABaseline is tasks.md 4.2b.
//
// D17's migration backfill fills first_price from the earliest surviving raw row,
// and if the rows before it were already pruned the stored value is later than
// the real first price. A late baseline *understates* the expansion, which pushes
// `extended` towards false — the direction D10 forbids, arriving through a
// migration rather than through a missing input.
func TestABaselineThatPostdatesFirstSightingIsNotABaseline(t *testing.T) {
	c := aCandidate(t0)
	late := t0.Add(11 * time.Minute)
	e := MeasureExpansion(storedBaseline("20000", late, SourceOfficialPrices), []Observation{
		levelObs(late, SourceOfficialPrices, "20000"),
		levelObs(late.Add(time.Minute), SourceOfficialPrices, "21000"),
	})
	if !e.Measured {
		t.Fatalf("unmeasured (%q)", e.Why)
	}
	if e.GainPct != "5" {
		t.Fatalf("gain = %s%%, want 5%% — the test needs a baseline that flatters the candidate",
			e.GainPct)
	}
	got := AssessExtended(e, c, late.Add(time.Minute), vetoThresholds())
	if got.Clear() {
		t.Error("extended reads as clear from a baseline recorded eleven minutes after we first " +
			"saw the candidate; a late baseline understates every expansion measured from it")
	}
	if got.Reason() != VetoBaselineTooLate {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoBaselineTooLate)
	}
}

// TestTheBaselineLatenessBoundaryIsTheStalenessTTL. Ten minutes is the window a
// scan interval and the planned backoff ladder explain; past it, the gap is
// something else.
func TestTheBaselineLatenessBoundaryIsTheStalenessTTL(t *testing.T) {
	c := aCandidate(t0)
	for _, tc := range []struct {
		lag  time.Duration
		want bool
	}{
		{DefaultStalenessTTL - time.Second, true},
		{DefaultStalenessTTL, false},
	} {
		at := t0.Add(tc.lag)
		e := MeasureExpansion(storedBaseline("20000", at, SourceOfficialPrices), []Observation{
			levelObs(at, SourceOfficialPrices, "20000"),
			levelObs(at.Add(30*time.Second), SourceOfficialPrices, "21000"),
		})
		got := AssessExtended(e, c, at.Add(30*time.Second), vetoThresholds())
		if got.Measured() != tc.want {
			t.Errorf("a baseline %v after first_seen_at is measured=%v, want %v",
				tc.lag, got.Measured(), tc.want)
		}
	}
}

// TestExtendedNeedsACandidateToJudgeItsBaselineAgainst. Without first_seen_at
// there is nothing to call the baseline late against, so the lateness is unknown
// rather than absent.
func TestExtendedNeedsACandidateToJudgeItsBaselineAgainst(t *testing.T) {
	e := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "11000"),
	})
	got := AssessExtended(e, Candidate{}, t0.Add(time.Minute), vetoThresholds())
	if got.Clear() {
		t.Error("extended reads as clear with no candidate summary to date the baseline against")
	}
	if got.Reason() != VetoNoCandidate {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoNoCandidate)
	}
}

// TestALatestPriceOlderThanTheCeilingIsAMemoryToo.
//
// The same argument as task 4.7's candle, on the other end of the expansion: the
// gain is last ÷ first, and a stale `last` understates it while the price keeps
// moving. Understating is the direction that turns `extended` off, so the ceiling
// applies here as well.
func TestALatestPriceOlderThanTheCeilingIsAMemoryToo(t *testing.T) {
	c := aCandidate(t0)
	e := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "11000"),
	})
	fresh := AssessExtended(e, c, t0.Add(time.Minute), vetoThresholds())
	if !fresh.Clear() {
		t.Fatalf("a fresh reading is %v, want clear", fresh)
	}
	stale := AssessExtended(e, c, t0.Add(time.Minute).Add(MaxVetoInputAge), vetoThresholds())
	if stale.Measured() {
		t.Errorf("an expansion whose latest price is %v old is still measured: %v",
			MaxVetoInputAge, stale)
	}
	if stale.Reason() != VetoInputTooOld {
		t.Errorf("reason = %q, want %q", stale.Reason(), VetoInputTooOld)
	}
}

// --- task 4.5: the veto is not a term in a sum ----------------------------------

// TestTheTopAccelerationDoesNotOffsetAPriceOnItsHigh is the spec scenario
// "강한 가속도가 고점 근접을 상쇄하지 않는다".
//
// D3: 점수에 섞으면 다른 축이 상쇄한다. "거래대금 가속도가 대단하니 고점 근처인 건 감수"라는
// 합산은 정확히 추격매수의 논리다.
func TestTheTopAccelerationDoesNotOffsetAPriceOnItsHigh(t *testing.T) {
	// Money arriving five times faster than in the prior window: every shadow
	// threshold crossed, which is the strongest signal this change can record.
	series, err := NewSourceSeries([]Observation{
		valueObs("005930", t0, SourceOfficialTradingValue, "0"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "600"),
	})
	if err != nil {
		t.Fatalf("NewSourceSeries: %v", err)
	}
	accel := Accelerate(series, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	if !accel.Computed() {
		t.Fatalf("acceleration not computed (%s)", accel.Why())
	}
	for _, th := range ShadowThresholds {
		if !accel.Crossed(th) {
			t.Fatalf("acceleration %s did not cross %s; the test needs the strongest signal "+
				"available to try to buy off the veto", accel.Ratio, th)
		}
	}

	// And the price is sitting 0.5% below the day's high.
	position := MeasureRangePosition([]Observation{candleObs(t0.Add(60*time.Second), "99.5", "100", "80")})
	chase := AssessChase(VetoInputs{
		Candidate: aCandidate(t0),
		Sighting: MeasureFirstSighting(aCandidate(t0),
			storedFirstRank(148, 150, t0, SourceOfficialTradingValue),
			[]Observation{rankedObs(t0, SourceOfficialTradingValue, 148, 150)}),
		Expansion:  MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{levelObs(t0.Add(60*time.Second), SourceOfficialPrices, "10100")}),
		Range:      position,
		At:         t0.Add(60 * time.Second),
		Thresholds: vetoThresholds(),
	})
	if !chase.NearHigh.Dangerous() {
		t.Fatalf("near_high = %v, want raised — the test needs the veto to be live", chase.NearHigh)
	}
	if chase.Passed() {
		t.Error("a candidate with the top acceleration passes the chase veto while sitting on " +
			"its day high; the strong axis bought off the veto, which is the logic of chasing")
	}
	if codes := chase.Raised(); len(codes) != 1 || codes[0] != VetoNearHigh {
		t.Errorf("raised codes = %v, want [%s]", codes, VetoNearHigh)
	}
}

// TestTheVetoCannotSeeAScoreToBeOffsetBy fixes 4.5 structurally rather than at
// today's call site.
//
// The behavioural test above says the current wiring does not offset. This says
// there is nothing to offset with: the verdict carries no number anybody could
// add to, and the inputs the verdict is computed from cannot hold an
// acceleration, a rank move or a tally. Adding one is what a future lane would do
// first, and the diff would look reasonable.
// The §4 review found the walk one level deep while its doc claimed the whole
// closure: Sighting.Rank, Sighting.RankTotal and Expansion.Unreadable are already
// ints reachable from VetoInputs, and an Acceleration parked inside any of the
// three measurement structs would have passed. So it recurses, and the claim it
// makes is now two:
//
//	no score type anywhere in the closure   Acceleration, RankMove and the two
//	                                        tallies cannot be reached from a veto
//	                                        input by any path
//	no number that is not named here        every numeric field is in the
//	                                        allowlist below with the reason it is
//	                                        not a score, so a new one fails
//
// The allowlist is the honest version of the original claim. The measurements
// legitimately carry the source's own integers — D6 requires the reading to travel
// beside anything derived from it — and the verdict itself still carries none:
// Chase and VetoState have no numeric field and no entry here.
func TestTheVetoCannotSeeAScoreToBeOffsetBy(t *testing.T) {
	scores := map[reflect.Type]bool{
		reflect.TypeOf(Acceleration{}):      true,
		reflect.TypeOf(RankMove{}):          true,
		reflect.TypeOf(CrossingTally{}):     true,
		reflect.TypeOf(ThresholdCrossing{}): true,
		// The shadow bands (task 4.10) are here for D18's reason rather than D3's.
		// They are not a score; they are the record of a quantity nobody has
		// approved a threshold for, and a veto that could see one would be deciding
		// against a number chosen without a source — which is what D6 forbids and
		// what the whole band arrangement exists to avoid. Keeping them out of the
		// closure is the input half of that; band.go's own static test closes the
		// other direction.
		reflect.TypeOf(ShadowBand{}):   true,
		reflect.TypeOf(BandCrossing{}): true,
		reflect.TypeOf(BandTally{}):    true,
	}
	allowed := map[string]string{
		"VetoInputs.Candidate.SourcesAttempted": "D4 completeness, not a score",
		"VetoInputs.Candidate.SourcesResponded": "D4 completeness, not a score",
		"VetoInputs.Sighting.Rank":              "the reading's own list position (D6)",
		"VetoInputs.Sighting.RankTotal":         "the list's own length, which normalises it (D8)",
		"VetoInputs.Expansion.Unreadable":       "a count of malformed rows beside the answer",
		"VetoInputs.Thresholds.MaxInputAge":     "a duration, and the ceiling is not compared with anything measured",
	}
	numeric := func(k reflect.Kind) bool {
		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		}
		return false
	}
	// Recursion stops at the package boundary, so time.Time's unexported wall clock
	// is not walked. Everything a veto input is built from is declared here.
	pkg := reflect.TypeOf(VetoInputs{}).PkgPath()
	seen := map[reflect.Type]bool{}
	used := map[string]bool{}
	var walk func(typ reflect.Type, path string)
	walk = func(typ reflect.Type, path string) {
		if typ.Kind() != reflect.Struct || seen[typ] {
			return
		}
		seen[typ] = true
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			name := path + "." + f.Name
			// An embedded struct keeps the outer path: Candidate.Key's fields are
			// Candidate's fields as far as anybody reading them is concerned.
			if f.Anonymous {
				name = path
			}
			elem := f.Type
			for elem.Kind() == reflect.Pointer || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Array || elem.Kind() == reflect.Map {
				elem = elem.Elem()
			}
			if scores[f.Type] || scores[elem] {
				t.Errorf("%s is a %s; the veto can see a score and a later reader will sum it",
					name, f.Type)
			}
			if numeric(f.Type.Kind()) {
				if _, ok := allowed[name]; !ok {
					t.Errorf("%s is a %s; a veto that carries an unaccounted number is a term "+
						"in somebody's sum — add it to the allowlist with the reason it is not "+
						"a score, or keep it out of the closure", name, f.Type)
				}
				used[name] = true
			}
			if elem.Kind() == reflect.Struct && elem.PkgPath() == pkg {
				walk(elem, name)
			}
		}
	}
	for _, typ := range []reflect.Type{
		reflect.TypeOf(VetoInputs{}), reflect.TypeOf(Chase{}), reflect.TypeOf(VetoState{}),
	} {
		walk(typ, typ.Name())
	}
	// A stale allowlist is the way this test rots into a rubber stamp: an entry for
	// a field that no longer exists is a permission nobody is using and nobody will
	// re-examine.
	for name := range allowed {
		if !used[name] {
			t.Errorf("the allowlist names %s and the walk never reached it", name)
		}
	}
	// And the state itself has no conversion to one.
	if reflect.TypeOf(VetoState{}).Kind() != reflect.Struct {
		t.Error("VetoState is not a struct; a defined numeric or string type can be added or " +
			"concatenated into something else")
	}
}

// --- task 4.6: a vetoed candidate is still recorded -----------------------------

// TestAVetoedCandidateIsStillStoredAndReported.
//
// D3: veto가 걸린 후보도 저장하고 화면에 보여준다. 지우면 "우리가 늦게 본 종목이 얼마나
// 되는가"를 나중에 셀 수 없고, 그 수가 이 시스템이 실제로 조기인지 아닌지의 유일한 증거다.
func TestAVetoedCandidateIsStillStoredAndReported(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	readings := []Observation{
		obs("005930", t0, SourceOfficialTradingValue,
			Reported{Rank: 5, RankTotal: 150, Price: "99500"}),
		obs("005930", t0, SourceOfficialCandles,
			Reported{Price: "99500", DayHigh: "100000", DayLow: "80000"}),
	}
	if err := s.RecordObservations(ctx, readings); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := s.Promote(ctx, MarketKR, "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.NoteFirstPrice(ctx, MarketKR, "005930", "99500", t0, SourceOfficialCandles); err != nil {
		t.Fatalf("NoteFirstPrice: %v", err)
	}
	if _, err := s.NoteFirstRank(ctx, MarketKR, "005930",
		storedFirstRank(5, 150, t0, SourceOfficialTradingValue)); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}

	c, found, err := s.Candidate(ctx, MarketKR, "005930", t0)
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	baseline, _, err := s.Baseline(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	first, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	rows, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}

	chase := AssessChase(VetoInputs{
		Candidate:  c,
		Sighting:   MeasureFirstSighting(c, first, rows),
		Expansion:  MeasureExpansion(baseline, rows),
		Range:      MeasureRangePosition(rows),
		At:         t0,
		Thresholds: vetoThresholds(),
	})
	if !chase.Vetoed() {
		t.Fatalf("the candidate is not vetoed: seen_late=%v extended=%v near_high=%v",
			chase.SeenLate, chase.Extended, chase.NearHigh)
	}

	// Assessing it changed nothing in the store.
	after, found, err := s.Candidate(ctx, MarketKR, "005930", t0)
	if err != nil || !found {
		t.Fatalf("the vetoed candidate is gone from the store: %v found=%v", err, found)
	}
	if !after.FirstSeenAt.Equal(c.FirstSeenAt) {
		t.Errorf("first_seen_at moved from %v to %v", c.FirstSeenAt, after.FirstSeenAt)
	}
	stillThere, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(stillThere) != len(rows) {
		t.Errorf("the vetoed candidate kept %d of %d observations; the share of candidates we "+
			"saw late is the only evidence this system is early", len(stillThere), len(rows))
	}

	tally := TallyVetoes([]Chase{chase})
	if tally.Vetoed != 1 {
		t.Errorf("the tally counts %d vetoed candidates, want 1", tally.Vetoed)
	}
	if tally.Total != 1 {
		t.Errorf("the tally counts %d candidates in total, want 1", tally.Total)
	}
}

// TestEveryAssessedCandidateLandsInExactlyOneBucket.
//
// The §3 review's second P0 was TallyCrossings dropping a candidate out of both
// halves, which turned the screen's default reading from "mostly unchecked" into
// "all measured and quiet". CrossingTally answered it with a Total that makes the
// disjointness arithmetic rather than asserted; the same guard belongs here,
// because under the candle budget the unmeasured bucket is the majority.
func TestEveryAssessedCandidateLandsInExactlyOneBucket(t *testing.T) {
	in := []Chase{
		{SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto()},
		{SeenLate: RaisedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto()},
		{SeenLate: ClearedVeto(), Extended: ClearedVeto(), NearHigh: UnmeasuredVeto(noDayHighReason())},
		// The one that used to disappear: nobody assessed it at all.
		{},
		// Raised and unmeasured at once: it counts as vetoed, once.
		{SeenLate: RaisedVeto(), Extended: UnmeasuredVeto(VetoInputTooOld), NearHigh: ClearedVeto()},
	}
	got := TallyVetoes(in)
	if got.Total != len(in) {
		t.Errorf("Total = %d, want %d", got.Total, len(in))
	}
	if got.Total != got.Passed+got.Vetoed+got.Unmeasured {
		t.Errorf("Total %d != Passed %d + Vetoed %d + Unmeasured %d — a candidate fell out of "+
			"every bucket", got.Total, got.Passed, got.Vetoed, got.Unmeasured)
	}
	if got.Passed != 1 {
		t.Errorf("Passed = %d, want 1", got.Passed)
	}
	if got.Vetoed != 2 {
		t.Errorf("Vetoed = %d, want 2", got.Vetoed)
	}
	if got.Unmeasured != 2 {
		t.Errorf("Unmeasured = %d, want 2 — the candidate nobody assessed and the one with no "+
			"candle are both 'we did not check', not 'we checked and it is fine'", got.Unmeasured)
	}
	// Per code, so an operator can see which input is missing.
	if got.NotMeasured[VetoNearHigh] != 2 {
		t.Errorf("NotMeasured[%s] = %d, want 2", VetoNearHigh, got.NotMeasured[VetoNearHigh])
	}
	if got.Raised[VetoSeenLate] != 2 {
		t.Errorf("Raised[%s] = %d, want 2", VetoSeenLate, got.Raised[VetoSeenLate])
	}
	// Every code has a key even at zero, so a missing column cannot read as a
	// column nobody needed.
	for _, code := range OrderedVetoCodes() {
		if _, ok := got.Raised[code]; !ok {
			t.Errorf("Raised has no entry for %s", code)
		}
		if _, ok := got.NotMeasured[code]; !ok {
			t.Errorf("NotMeasured has no entry for %s", code)
		}
	}
	if got.Reasons[noDayHighReason()] != 1 {
		t.Errorf("Reasons[%s] = %d, want 1", noDayHighReason(), got.Reasons[noDayHighReason()])
	}
}

// TestSeenLateAndALateEntryAreNotTheSameCode is the spec scenario "늦게 본 것과
// 늦은 자리는 다르게 기록된다".
//
// D3 keeps them apart because the remedies differ: many seen_late means the scan
// interval or the source panel is late, many extended means the threshold is.
func TestSeenLateAndALateEntryAreNotTheSameCode(t *testing.T) {
	c := aCandidate(t0)
	th := vetoThresholds()

	// Seen late: already 12th of 150 the first time we looked, and it has barely
	// moved since.
	sawLate := AssessChase(VetoInputs{
		Candidate: c,
		Sighting: MeasureFirstSighting(c, storedFirstRank(12, 150, t0, SourceOfficialTradingValue),
			[]Observation{rankedObs(t0, SourceOfficialTradingValue, 12, 150)}),
		Expansion: MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
			levelObs(t0, SourceOfficialPrices, "10000"),
			levelObs(t0.Add(time.Minute), SourceOfficialPrices, "10100"),
		}),
		Range:      MeasureRangePosition([]Observation{candleObs(t0.Add(time.Minute), "10100", "12000", "9000")}),
		At:         t0.Add(time.Minute),
		Thresholds: th,
	})
	if got := sawLate.Raised(); len(got) != 1 || got[0] != VetoSeenLate {
		t.Errorf("raised = %v, want [%s]", got, VetoSeenLate)
	}

	// Late position: we caught it at the bottom of the list and it has doubled.
	ranAway := AssessChase(VetoInputs{
		Candidate: c,
		Sighting: MeasureFirstSighting(c, storedFirstRank(148, 150, t0, SourceOfficialTradingValue),
			[]Observation{rankedObs(t0, SourceOfficialTradingValue, 148, 150)}),
		Expansion: MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
			levelObs(t0, SourceOfficialPrices, "10000"),
			levelObs(t0.Add(time.Minute), SourceOfficialPrices, "20000"),
		}),
		Range:      MeasureRangePosition([]Observation{candleObs(t0.Add(time.Minute), "20000", "24000", "9000")}),
		At:         t0.Add(time.Minute),
		Thresholds: th,
	})
	if got := ranAway.Raised(); len(got) != 1 || got[0] != VetoExtended {
		t.Errorf("raised = %v, want [%s]", got, VetoExtended)
	}
}

// TestANonPositiveThresholdIsNotAThreshold is the §4 review's P0.
//
// thresholdReason refused only the empty and the unparseable, and "0", "-0",
// "0.0" and "-1" all parse. The realistic input is §5 rendering the contract knob
// with strconv.FormatFloat, where an absent YAML key produces "0" rather than "".
// near_high is the only veto with an approved threshold (D18), so a zero there is
// the whole live veto surface reading clear for every candidate below its high.
// It is task 1.1's zero DayHigh on the other operand of the same comparison.
func TestANonPositiveThresholdIsNotAThreshold(t *testing.T) {
	position := MeasureRangePosition([]Observation{candleObs(t0, "96", "100", "80")})
	if !position.Measured {
		t.Fatalf("unmeasured (%q)", position.Why)
	}
	for _, raw := range []string{"0", "-0", "0.0", "-0.0", "-1", "-2.5"} {
		th := vetoThresholds()
		th.NearHighDistancePct = raw
		got := AssessNearHigh(position, t0, th)
		if got.Measured() {
			t.Errorf("near_high against a threshold of %q is %v; a candidate 4%% below its high "+
				"reads as measured and clear, and so does every other candidate in the list", raw, got)
		}
		if got.Reason() != VetoThresholdNotPositive {
			t.Errorf("near_high against %q: reason = %q, want %q", raw, got.Reason(),
				VetoThresholdNotPositive)
		}
	}

	// And the other two, where the comparison is `>` rather than `<`. A
	// non-positive threshold there fires *more*, so the refusal turns a raised into
	// an unmeasured and never a clear into one — which is why it is safe in all
	// three codes rather than only in the one that motivated it.
	c := aCandidate(t0)
	sighting := MeasureFirstSighting(c, storedFirstRank(148, 150, t0, SourceOfficialTradingValue),
		[]Observation{rankedObs(t0, SourceOfficialTradingValue, 148, 150)})
	expansion := MeasureExpansion(storedBaseline("10000", t0, SourceOfficialPrices), []Observation{
		levelObs(t0, SourceOfficialPrices, "10000"),
		levelObs(t0.Add(time.Minute), SourceOfficialPrices, "10100"),
	})
	for _, raw := range []string{"0", "-0", "0.0", "-1"} {
		th := vetoThresholds()
		th.SeenLatePercentilePct = raw
		if got := AssessSeenLate(sighting, th); got.Measured() {
			t.Errorf("seen_late against a threshold of %q is %v, want unmeasured", raw, got)
		} else if got.Reason() != VetoThresholdNotPositive {
			t.Errorf("seen_late against %q: reason = %q, want %q", raw, got.Reason(),
				VetoThresholdNotPositive)
		}
		th = vetoThresholds()
		th.ExtendedGainPct = raw
		got := AssessExtended(expansion, c, t0.Add(time.Minute), th)
		if got.Measured() {
			t.Errorf("extended against a threshold of %q is %v, want unmeasured", raw, got)
		} else if got.Reason() != VetoThresholdNotPositive {
			t.Errorf("extended against %q: reason = %q, want %q", raw, got.Reason(),
				VetoThresholdNotPositive)
		}
	}
}

// TestABaselineFromAPreviousLifeIsNotABaseline is the §4 review's P1-1.
//
// AssessExtended refused only a baseline that postdates first_seen_at (4.2b).
// MeasureExpansion moves the baseline *backwards* onto any older priced row in the
// slice — deliberately, because a baseline may move backwards on new evidence —
// and raw rows outlive a candidate life by 48h (D11) while Promote clears
// first_price on expiry. So the dead life's price becomes the live life's
// baseline, and the direction is the unsafe one: a candidate that has doubled
// since it was promoted reads as one that has fallen a third and has not run.
//
// It is TestAPreviousLifesReadingsAreNotThisCandidatesFirstSighting's shape on the
// other side of the same window, one veto across.
func TestABaselineFromAPreviousLifeIsNotABaseline(t *testing.T) {
	// The previous life traded at 30,000 and expired. This one was promoted two
	// hours later at 10,000 and has since doubled.
	reborn := aCandidate(t0.Add(2 * time.Hour))
	e := MeasureExpansion(storedBaseline("10000", t0.Add(2*time.Hour), SourceOfficialPrices),
		[]Observation{
			levelObs(t0, SourceOfficialPrices, "30000"), // the dead life's row, still retained
			levelObs(t0.Add(2*time.Hour), SourceOfficialPrices, "10000"),
			levelObs(t0.Add(2*time.Hour+time.Minute), SourceOfficialPrices, "20000"),
		})
	if !e.Measured {
		t.Fatalf("unmeasured (%q)", e.Why)
	}
	if e.BaselineStored {
		t.Fatalf("the series did not re-base the baseline; the test cannot show the defect")
	}
	got := AssessExtended(e, reborn, t0.Add(2*time.Hour+time.Minute), vetoThresholds())
	if got.Clear() {
		t.Errorf("extended = %v from a baseline two hours before this candidate was first seen "+
			"(first %s at %v, last %s, gain %s%%); a candidate that has doubled reads as one "+
			"that has not run", got, e.FirstPrice, e.FirstAt, e.LastPrice, e.GainPct)
	}
	if got.Reason() != VetoBaselineTooEarly {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoBaselineTooEarly)
	}
}

// TestTheBaselineIdentityWindowIsSymmetric pins both ends of it to the same
// number. A legitimate baseline is written at promotion, within seconds of
// first_seen_at; the only thing the earlier end refuses is a baseline the series
// moved backwards onto.
func TestTheBaselineIdentityWindowIsSymmetric(t *testing.T) {
	for _, tc := range []struct {
		lag  time.Duration
		want bool
	}{
		{-DefaultStalenessTTL, false},
		{-DefaultStalenessTTL + time.Second, true},
		{0, true},
		{DefaultStalenessTTL - time.Second, true},
		{DefaultStalenessTTL, false},
	} {
		firstSeen := t0.Add(time.Hour)
		at := firstSeen.Add(tc.lag)
		e := MeasureExpansion(storedBaseline("20000", at, SourceOfficialPrices), []Observation{
			levelObs(at, SourceOfficialPrices, "20000"),
			levelObs(firstSeen.Add(DefaultStalenessTTL), SourceOfficialPrices, "21000"),
		})
		got := AssessExtended(e, aCandidate(firstSeen),
			firstSeen.Add(DefaultStalenessTTL), vetoThresholds())
		if got.Measured() != tc.want {
			t.Errorf("a baseline %v from first_seen_at is measured=%v (%v), want %v",
				tc.lag, got.Measured(), got, tc.want)
		}
	}
}

// TestRemovingAVetoCodeCannotRemoveItsVeto is the §4 review's P1-3.
//
// Every predicate here — Passed, Vetoed, HasUnmeasured, Raised, NotMeasured and
// TallyVetoes — is defined by iterating OrderedVetoCodes rather than naming the three
// fields. That refuses a *fourth* code correctly: it has no field, so Chase.State
// answers unmeasured and it can never be a pass. The other direction was
// unguarded. As an exported slice the list is a shared backing array anybody can
// shorten or overwrite, and a code that leaves the list stops being consulted —
// Chase.Passed() then returns true with a veto raised in a field nobody reads.
//
// An array is a value type: `range` still works, len still works, and a caller who
// takes it gets a copy rather than the list the verdict is defined by.
func TestRemovingAVetoCodeCannotRemoveItsVeto(t *testing.T) {
	want := []VetoCode{VetoSeenLate, VetoExtended, VetoNearHigh}
	codes := OrderedVetoCodes()
	if len(codes) != len(want) {
		t.Fatalf("OrderedVetoCodes has %d entries, want %d — every predicate in this file is defined "+
			"by this list", len(codes), len(want))
	}
	for i := range want {
		if codes[i] != want[i] {
			t.Errorf("OrderedVetoCodes[%d] = %q, want %q (D3's order, which every tally and every "+
				"screen lists them in)", i, codes[i], want[i])
		}
	}

	// Taking the list and writing to it must not reach the verdict. Restoring it is
	// OrderedVetoCodes returns a value, so overwriting the copy cannot damage the
	// package invariant or any test that runs later.
	codes[0] = VetoNearHigh

	chase := Chase{SeenLate: RaisedVeto(), Extended: ClearedVeto(), NearHigh: ClearedVeto()}
	if chase.Passed() {
		t.Error("a Chase with seen_late RAISED passes the veto after a caller overwrote the " +
			"code list; removing a code removed its veto")
	}
	if !chase.Vetoed() {
		t.Error("a Chase with seen_late RAISED reports nothing vetoed after the list was rewritten")
	}
	if got := TallyVetoes([]Chase{chase}); got.Vetoed != 1 || got.Raised[VetoSeenLate] != 1 {
		t.Errorf("the tally counts vetoed=%d raised[seen_late]=%d, want 1 and 1",
			got.Vetoed, got.Raised[VetoSeenLate])
	}
}

// --- task 4.9: the first rank is a column, not a query --------------------------

// TestARankedRowFromTheGapBetweenTwoLivesIsNotThisLifesFirstSighting is the §4
// review's P1-2, reproduced end to end through the real store.
//
// D20's draft argued the ±TTL window could not reach across two lives because the
// lives are at least the cooling TTL apart. That is true of *promotions* and false
// of *observations*: D8 separates the two on purpose, so ranked rows keep arriving
// for a symbol that is cooling, expired, or was never promoted at all. A row from
// the gap lands inside the new life's window and becomes its "first sighting".
//
// The direction is the unsafe one, as it was in D20's own example: the row from
// the gap has the symbol near the bottom of the list, so a candidate promoted at
// 4th of 150 reads as one we caught at 148th and seen_late clears.
func TestARankedRowFromTheGapBetweenTwoLivesIsNotThisLifesFirstSighting(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	// Life one: promoted 5th of 150, cooled a minute later, expired at +31m.
	if err := s.RecordObservations(ctx, []Observation{
		rankedObs(t0, SourceOfficialTradingValue, 5, 150),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := s.Promote(ctx, MarketKR, "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := s.Cool(ctx, MarketKR, "005930", t0.Add(time.Minute)); err != nil {
		t.Fatalf("Cool: %v", err)
	}
	if _, err := s.NoteFirstRank(ctx, MarketKR, "005930",
		storedFirstRank(5, 150, t0, SourceOfficialTradingValue)); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}

	// The gap. Nobody promotes it and the rows keep arriving anyway (D8) — this one
	// nine minutes before the new life begins, with the symbol near the bottom.
	gap := t0.Add(51 * time.Minute)
	if err := s.RecordObservations(ctx, []Observation{
		rankedObs(gap, SourceOfficialTradingValue, 148, 150),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	// Life two: it crosses again an hour after the first, already 4th of 150.
	reborn := t0.Add(time.Hour)
	if err := s.RecordObservations(ctx, []Observation{
		rankedObs(reborn, SourceOfficialTradingValue, 4, 150),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	c, err := s.Promote(ctx, MarketKR, "005930", reborn)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if !c.FirstSeenAt.Equal(reborn) {
		t.Fatalf("the second life kept the first's first_seen_at (%v); the test needs two lives",
			c.FirstSeenAt)
	}
	if _, err := s.NoteFirstRank(ctx, MarketKR, "005930",
		storedFirstRank(4, 150, reborn, SourceOfficialTradingValue)); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}

	first, found, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil || !found {
		t.Fatalf("FirstRank: %v found=%v", err, found)
	}
	if first.Rank != 4 || first.Total != 150 {
		t.Fatalf("the stored first rank = %d of %d, want 4 of 150 — the expired life's rank "+
			"survived into the new one", first.Rank, first.Total)
	}

	rows, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	got := MeasureFirstSighting(c, first, rows)
	if !got.Measured {
		t.Fatalf("unmeasured (%q); the stored first rank is right there", got.Why)
	}
	if got.Rank != 4 || got.RankTotal != 150 {
		t.Fatalf("first sighting = %d of %d at %v; a row from the gap between two lives became "+
			"this life's first sighting", got.Rank, got.RankTotal, got.At)
	}
	if state := AssessSeenLate(got, vetoThresholds()); !state.Dangerous() {
		t.Errorf("seen_late = %v (percentile %s), want raised — a candidate promoted 4th of 150 "+
			"reads as one we caught at 148th", state, got.PercentilePct)
	}
}

// TestASightingRankedBeyondItsListIsNotAPercentile.
//
// percentileOf is (total − rank) ÷ total, so a rank past the end of its list is a
// negative percentile — a hand-built Sighting{Rank: 200, RankTotal: 150} is −33%,
// which is below every threshold there is and clears seen_late for a symbol that
// cannot be where it says it is. Rank <= 0 and RankTotal <= 0 were already
// refused; this is the third impossible reading of the same pair.
//
// Observation.validate blocks it on the way into the store, so the exposure is the
// same one the other two already have a test for: §5's fixtures, a future lane, a
// struct literal built outside a measurement.
func TestASightingRankedBeyondItsListIsNotAPercentile(t *testing.T) {
	for _, s := range []Sighting{
		{Measured: true, Rank: 200, RankTotal: 150, At: t0, FirstSeenAt: t0},
		{Measured: true, Rank: 151, RankTotal: 150, At: t0, FirstSeenAt: t0},
	} {
		exceeds, measured := s.PercentileExceeds("80")
		if measured {
			t.Errorf("Sighting{Rank: %d, RankTotal: %d} is measured (exceeds=%v); its percentile "+
				"is %s%%", s.Rank, s.RankTotal, exceeds,
				formatDecimal(percentileOf(s.Rank, s.RankTotal)))
		}
		if state := AssessSeenLate(s, vetoThresholds()); state.Clear() {
			t.Errorf("seen_late is clear for a symbol ranked %d of %d", s.Rank, s.RankTotal)
		}
	}
	// The boundary: last in the list is a legitimate reading and stays one.
	if _, measured := (Sighting{Measured: true, Rank: 150, RankTotal: 150}).PercentileExceeds("80"); !measured {
		t.Error("a symbol ranked last of its list is refused; that is the ordinary bottom of the " +
			"list and the entry we most want to be able to count")
	}
}
