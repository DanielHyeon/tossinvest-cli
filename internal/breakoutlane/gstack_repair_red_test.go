package breakoutlane

import "testing"

// This grouped contract locks the exact L2 boundaries identified by the final
// gstack review. It is deliberately package-local so it can also prevent the
// frozen state and v1 threshold vocabulary from silently drifting.
func TestGstackRepairFrozenVocabularyAndV1Thresholds(t *testing.T) {
	if v1PPMScale != 1_000_000 || v1OpeningRangeBars != 15 || v1OpeningRangeMinutes != 15 || v1BreakoutBufferPPM != 100_000 || v1RetestToleranceMinPPM != 100_000 || v1RetestToleranceMaxPPM != 250_000 || v1TimeoutKRClosedBars != 8 || v1TimeoutUSClosedBars != 10 || v1RVOLMinPPM != 1_500_000 || v1UpperWickRangeMaxPPM != 350_000 {
		t.Fatal("frozen v1 constants drifted")
	}
	for _, tc := range []struct {
		phase phase
		want  string
	}{
		{phaseDiscovered, "DISCOVERED"}, {phaseRangeLocked, "RANGE_LOCKED"}, {phaseBreakoutClosed, "BREAKOUT_CLOSED"}, {phaseRetestWait, "RETEST_WAIT"}, {phaseReclaimed, "RECLAIMED"}, {phaseArmed, "ARMED"}, {phaseProposed, "PROPOSED"}, {phaseInvalidated, "INVALIDATED"}, {phaseTimedOut, "TIMED_OUT"}, {phaseConsumed, "CONSUMED"},
	} {
		if string(tc.phase) != tc.want || (Decision{phase: tc.phase}).Phase() != tc.want {
			t.Fatalf("phase=%q want=%q", tc.phase, tc.want)
		}
	}
	for _, tolerance := range []uint64{v1RetestToleranceMinPPM, v1RetestToleranceMaxPPM} {
		i := fixtureConfigInput(t)
		i.RetestTolerancePPM = tolerance
		i.Digest = V1ConfigDigest(i)
		if _, err := NewV1Config(i); err != nil {
			t.Fatalf("retest tolerance %d rejected: %v", tolerance, err)
		}
	}
	for _, invalid := range []uint64{v1RetestToleranceMinPPM - 1, v1RetestToleranceMaxPPM + 1} {
		i := fixtureConfigInput(t)
		i.RetestTolerancePPM = invalid
		i.Digest = V1ConfigDigest(i)
		if _, err := NewV1Config(i); err == nil {
			t.Fatalf("retest tolerance %d accepted", invalid)
		}
	}
	for _, openingMinutes := range []uint64{v1OpeningRangeMinutes - 1, v1OpeningRangeMinutes + 1} {
		i := fixtureConfigInput(t)
		i.OpeningRangeMinutes = openingMinutes
		i.Digest = V1ConfigDigest(i)
		if _, err := NewV1Config(i); err == nil {
			t.Fatalf("opening range %d accepted", openingMinutes)
		}
	}
	if _, err := NewEvidenceSnapshot(fixtureInput(t)); err != nil {
		t.Fatalf("%d opening bars rejected: %v", v1OpeningRangeBars, err)
	}
	tooShort := fixtureInput(t)
	tooShort.Bars = tooShort.Bars[:v1OpeningRangeBars-1]
	if _, err := NewEvidenceSnapshot(tooShort); err == nil {
		t.Fatalf("%d opening bars accepted", v1OpeningRangeBars-1)
	}
	config := fixtureConfig(t, "cfg")
	if !BreakoutCloseQualifies(101, 100, 9, config) || BreakoutCloseQualifies(101, 100, 20, config) || !BreakoutCloseQualifies(102, 100, 20, config) {
		t.Fatal("1 tick versus 0.10 ATR breakout buffer boundary drifted")
	}
	for _, tc := range []struct {
		name string
		mut  func(*EvidenceInput)
		want bool
	}{
		{"rvol exact", func(i *EvidenceInput) {
			b := i.Bars[15].value
			b.RVOLPPM = v1RVOLMinPPM
			i.Bars[15] = ClosedBar{value: b}
		}, true},
		{"rvol below", func(i *EvidenceInput) {
			b := i.Bars[15].value
			b.RVOLPPM = v1RVOLMinPPM - 1
			i.Bars[15] = ClosedBar{value: b}
		}, false},
		{"wick exact", func(i *EvidenceInput) {
			b := i.Bars[15].value
			b.UpperWickRangePPM = v1UpperWickRangeMaxPPM
			i.Bars[15] = ClosedBar{value: b}
		}, true},
		{"wick above", func(i *EvidenceInput) {
			b := i.Bars[15].value
			b.UpperWickRangePPM = v1UpperWickRangeMaxPPM + 1
			i.Bars[15] = ClosedBar{value: b}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := fixtureInput(t)
			tc.mut(&i)
			got := Evaluate(snapshot(t, i), nil).Phase() == "PROPOSED"
			if got != tc.want {
				t.Fatalf("proposed=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestGstackRepairFailedReclaimAndRiskRewardBoundary(t *testing.T) {
	i := fixtureInput(t)
	b := i.Bars[17].value
	b.CloseMinor, b.HighMinor, b.VolumeExpanded = 99, 100, true
	i.Bars[17] = ClosedBar{value: b}
	if d := Evaluate(snapshot(t, i), nil); d.Phase() != "INVALIDATED" || d.ProposalID() != "" {
		t.Fatalf("volume-expanded failed reclaim proposed: %+v", d)
	}
	for _, tc := range []struct {
		name   string
		target uint64
		want   RefusalCode
	}{
		{"minimum RR exact", 123, RefusalNone},
		{"minimum RR one minor below", 122, RefusalNonProtectiveTarget},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := SizingInput{ProposedEntryMinor: 100, StopMinor: 90, TargetMinor: tc.target, RiskBudgetAccountMinor: 1_000, NotionalCapAccountMinor: 10_000, FinalCap: 10, MinRiskRewardPPM: 2_000_000}
			if got := size(in, fixtureQuote(t, 5, 6), fixtureFX(t, 10), 10).Refusal; got != tc.want {
				t.Fatalf("refusal=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestGstackRepairQuoteAndFXExactBoundaries(t *testing.T) {
	base := fixtureConfigInput(t)
	base.MaxQuoteAgeMS, base.MaxSpreadPPM, base.MaxEntryDriftPPM = 10, 200, 200
	base.Digest = V1ConfigDigest(base)
	config, err := NewV1Config(base)
	if err != nil {
		t.Fatal(err)
	}
	quote := func(t *testing.T, bid, ask, source uint64) QuoteSeal {
		t.Helper()
		i := QuoteSealInput{BidMinor: bid, AskMinor: ask, LastMinor: ask, Currency: "USD", SourceObservedAtMS: source, ReceivedAtMS: source}
		i.Digest = QuoteSealDigest(i)
		q, err := NewQuoteSeal(i)
		if err != nil {
			t.Fatal(err)
		}
		return q
	}
	for _, tc := range []struct {
		name, want string
		quote      QuoteSeal
		evaluated  uint64
		entry      uint64
	}{
		{"age exact", "", quote(t, 999_900, 1_000_100, 0), 10, 1_000_100},
		{"age one over", "QUOTE_STALE", quote(t, 999_900, 1_000_100, 0), 11, 1_000_100},
		{"spread exact", "", quote(t, 999_900, 1_000_100, 0), 0, 1_000_100},
		{"spread one over", "SPREAD_TOO_WIDE", quote(t, 999_900, 1_000_101, 0), 0, 1_000_101},
		{"drift exact", "", quote(t, 1_000_000, 1_000_200, 0), 0, 1_000_000},
		{"drift one over", "ENTRY_DRIFT_EXCEEDED", quote(t, 1_000_001, 1_000_201, 0), 0, 1_000_000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateQuote(tc.quote, tc.evaluated, tc.entry, config); string(got) != tc.want {
				t.Fatalf("refusal=%q want=%q", got, tc.want)
			}
		})
	}
	for _, tc := range []struct {
		name                   string
		direction              FXDirection
		rateNum, rateDen       uint64
		scale                  uint32
		wantCapacity, wantCost uint64
	}{
		{"canonical scale 2", FXAccountToInstrument, 2, 3, 2, 6, 7},
		{"inverse scale 6", FXInstrumentToAccount, 3, 2, 6, 6, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: tc.direction, RateNum: tc.rateNum, RateDen: tc.rateDen, Scale: tc.scale, AsOfMS: 1, FreshUntilMS: 10}
			i.Digest = FXSealDigest(i)
			f, err := NewFXSeal(i)
			if err != nil {
				t.Fatal(err)
			}
			capacity, capacityRefusal := convertCapacity(10, f, 10)
			cost, costRefusal := convertCost(10, f, 10)
			if capacityRefusal != RefusalNone || costRefusal != RefusalNone || capacity != tc.wantCapacity || cost != tc.wantCost {
				t.Fatalf("capacity=%d/%q cost=%d/%q", capacity, capacityRefusal, cost, costRefusal)
			}
		})
	}
}

func TestGstackRepairRetestQualifiesToleranceEndpoints(t *testing.T) {
	for _, tc := range []struct {
		name         string
		tolerancePPM uint64
		closeMinor   uint64
		want         bool
	}{
		{"100000 exact distance accepts", 100_000, 99, true},
		{"100000 next minor rejects", 100_000, 98, false},
		{"250000 integer floor accepts", 250_000, 98, true},
		{"250000 next minor rejects", 250_000, 97, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := fixtureConfigInput(t)
			input.RetestTolerancePPM = tc.tolerancePPM
			input.Digest = V1ConfigDigest(input)
			config, err := NewV1Config(input)
			if err != nil {
				t.Fatal(err)
			}
			if got := RetestQualifies(tc.closeMinor, 100, 10, config); got != tc.want {
				t.Fatalf("RetestQualifies(close=%d, resistance=100, ATR=10, tolerance=%d)=%v want=%v", tc.closeMinor, tc.tolerancePPM, got, tc.want)
			}
		})
	}
}

func fixtureConfigInput(t *testing.T) V1ConfigInput {
	t.Helper()
	i := V1ConfigInput{Version: LaneVersionV1, TickMinor: 1, OpeningRangeMinutes: v1OpeningRangeMinutes, BreakoutBufferPPM: v1BreakoutBufferPPM, RetestTolerancePPM: v1RetestToleranceMinPPM, TimeoutKR: v1TimeoutKRClosedBars, TimeoutUS: v1TimeoutUSClosedBars, RVOLMinPPM: v1RVOLMinPPM, UpperWickRangeMaxPPM: v1UpperWickRangeMaxPPM, MaxQuoteAgeMS: 10, MaxSpreadPPM: 100_000, MaxEntryDriftPPM: 100_000}
	i.Digest = V1ConfigDigest(i)
	return i
}
