package candidate

// tally_alarm_test.go is section 4: the tally's own arithmetic, and the fact that
// it needs no thresholds to work.
//
// The tallies here are built by hand rather than produced by an assessment, and
// that is the only way to write these tests: TallyVetoes cannot currently produce a
// contradiction, because the buckets it fills are disjoint by construction. That is
// the property under test. The alarm exists for the edit that breaks it — counting
// an unmeasured veto as a pass is a one-line change in TallyVetoes' switch, it
// compiles, and every existing test stays green because none of them assert on a
// number that would move.

import (
	"strings"
	"testing"
)

// talliedAs is a VetoTally with the given per-code counts, keyed exactly like the
// real one so a missing key cannot be what makes an assertion pass.
func talliedAs(total, passed int, raised, notMeasured map[VetoCode]int,
	reasons map[VetoUnmeasured]int) VetoTally {

	out := VetoTally{
		Total: total, Passed: passed,
		Raised:      map[VetoCode]int{},
		NotMeasured: map[VetoCode]int{},
		Reasons:     map[VetoUnmeasured]int{},
	}
	for _, code := range VetoCodes {
		out.Raised[code] = raised[code]
		out.NotMeasured[code] = notMeasured[code]
	}
	for why, n := range reasons {
		out.Reasons[why] = n
	}
	return out
}

// TestAConsistentTallyRaisesNothing is the negative control, and it uses the shape
// production actually produces today: no thresholds, so every candidate is
// unmeasured on two codes and nothing passed.
func TestAConsistentTallyRaisesNothing(t *testing.T) {
	tally := talliedAs(40, 0,
		map[VetoCode]int{VetoNearHigh: 2},
		map[VetoCode]int{VetoSeenLate: 40, VetoExtended: 40, VetoNearHigh: 38},
		map[VetoUnmeasured]int{VetoThresholdAbsent: 80, VetoUnmeasured(LevelNoDayHigh): 38})
	if got := tally.Anomalies(); len(got) != 0 {
		t.Errorf("a consistent tally reported %d anomalies: %v", len(got), got)
	}
}

// TestAnUnmeasuredVetoCountedAsAPassIsCaughtByTheArithmetic is task 4.1's RED.
//
// Forty candidates, all of them unmeasured on near_high because nobody spent a
// candle on them, and somebody has counted them as passes. The near_high sum is
// 40 + 0 + 40 = 80 against a total of 40.
func TestAnUnmeasuredVetoCountedAsAPassIsCaughtByTheArithmetic(t *testing.T) {
	tally := talliedAs(40, 40,
		nil,
		map[VetoCode]int{VetoNearHigh: 40},
		map[VetoUnmeasured]int{VetoUnmeasured(LevelNoDayHigh): 40})

	got := tally.Anomalies()
	if len(got) == 0 {
		t.Fatal("forty candidates were counted as passing while all forty were unmeasured " +
			"on near_high, and the tally reported nothing wrong")
	}
	var found *TallyAnomaly
	for i := range got {
		if got[i].Kind == TallyOvercounted && got[i].Code == VetoNearHigh {
			found = &got[i]
		}
	}
	if found == nil {
		t.Fatalf("no overcount was reported for near_high; got %v", got)
	}
	if !strings.Contains(found.Arithmetic(), "40") || !strings.Contains(found.Arithmetic(), "80") {
		t.Errorf("the finding does not carry the numbers that produced it: %q",
			found.Arithmetic())
	}
}

// TestTheAlarmCatchesEveryUnmeasuredReasonRatherThanTheThresholdOne is task 4.4,
// and it is the reason this check was chosen over the threshold-shaped one.
//
// THRESHOLD_ABSENT is one of many. The edit D10 exists to stop counts *any*
// unmeasured veto as a pass, and under the candle budget the reason will usually be
// NO_DAY_HIGH — a screen watching only for an absent threshold would never see it.
func TestTheAlarmCatchesEveryUnmeasuredReasonRatherThanTheThresholdOne(t *testing.T) {
	for _, why := range []VetoUnmeasured{
		VetoUnmeasured(LevelNoDayHigh), VetoInputTooOld, VetoNotRanked,
		VetoBaselineTooLate, VetoNewEntrantUnknown, VetoReadingTruncated,
		VetoNoFirstRank, VetoInstantUnknown,
	} {
		t.Run(string(why), func(t *testing.T) {
			// Ten candidates, all unmeasured on seen_late for this reason, all counted
			// as passes. No THRESHOLD_ABSENT anywhere in the tally.
			tally := talliedAs(10, 10, nil,
				map[VetoCode]int{VetoSeenLate: 10},
				map[VetoUnmeasured]int{why: 10})
			got := tally.Anomalies()
			if len(got) == 0 {
				t.Fatalf("ten candidates unmeasured under %s were counted as passes and "+
					"nothing was reported. An alarm that only watches THRESHOLD_ABSENT is "+
					"blind to every other way the same bug arrives", why)
			}
			for _, a := range got {
				if a.Kind == TallyPassedWithoutThreshold {
					t.Errorf("the threshold contradiction fired for a tally with no "+
						"THRESHOLD_ABSENT in it: %v", a)
				}
			}
		})
	}
}

// TestAPassBesideAnAbsentThresholdIsADirectContradiction is task 4.2.
//
// This one does not depend on the sums at all. A veto with no threshold cannot be
// clear, so it cannot contribute to a pass — the two counts cannot both be non-zero
// however the buckets are arranged.
func TestAPassBesideAnAbsentThresholdIsADirectContradiction(t *testing.T) {
	// Arranged so the inequality holds for every code: the overcount check finds
	// nothing here, and the contradiction still has to be reported.
	tally := talliedAs(100, 3, nil,
		map[VetoCode]int{VetoSeenLate: 20, VetoExtended: 20},
		map[VetoUnmeasured]int{VetoThresholdAbsent: 40})

	got := tally.Anomalies()
	if len(got) != 1 || got[0].Kind != TallyPassedWithoutThreshold {
		t.Fatalf("anomalies = %v, want exactly the threshold contradiction", got)
	}
	if got[0].Reason != VetoThresholdAbsent || got[0].ReasonCount != 40 || got[0].Passed != 3 {
		t.Errorf("the finding = %+v; it has to carry both counts that contradict", got[0])
	}
}

// TestTheAlarmNeedsNoThresholdToBeComputed is design D5's whole argument, stated as
// a test.
//
// The draft alarm was derived from the applied thresholds, and neither reporting
// surface has them — CycleResult has no threshold field and neither does
// SignalsMarket. The implementation that compiles reads a package constant instead,
// which makes the alarm a function of the constant rather than of what was applied,
// and its branch unreachable. This check is arithmetic on counts, so it works today
// with no threshold configured anywhere; the previous one would have guarded nothing
// until some later change gave it something to read.
func TestTheAlarmNeedsNoThresholdToBeComputed(t *testing.T) {
	// The tally this system produces right now: two of the three vetoes have no
	// approved threshold, so every candidate is unmeasured on both.
	broken := talliedAs(25, 25, nil,
		map[VetoCode]int{VetoSeenLate: 25, VetoExtended: 25},
		map[VetoUnmeasured]int{VetoThresholdAbsent: 50})
	if len(broken.Anomalies()) == 0 {
		t.Error("with no threshold configured anywhere, a tally that counted every " +
			"unmeasured candidate as a pass reported nothing")
	}

	// And it is the same function, with no thresholds passed to it. There is no
	// argument here to be wrong about.
	healthy := talliedAs(25, 0, nil,
		map[VetoCode]int{VetoSeenLate: 25, VetoExtended: 25},
		map[VetoUnmeasured]int{VetoThresholdAbsent: 50})
	if got := healthy.Anomalies(); len(got) != 0 {
		t.Errorf("the ordinary no-threshold tally reported %v; the structural zero is not a "+
			"fault and must not be alarmed on", got)
	}
}

// TestEveryAnomalyKindStatesItsOwnArithmetic.
//
// A finding that says only "inconsistent" gives a reader nothing to check, which is
// the same failure the reason codes exist to avoid one level down.
func TestEveryAnomalyKindStatesItsOwnArithmetic(t *testing.T) {
	kinds := map[TallyAnomalyKind]string{}
	for _, tally := range []VetoTally{
		talliedAs(5, 5, nil, map[VetoCode]int{VetoNearHigh: 5},
			map[VetoUnmeasured]int{VetoUnmeasured(LevelNoDayHigh): 5}),
		talliedAs(100, 3, nil, map[VetoCode]int{VetoSeenLate: 20},
			map[VetoUnmeasured]int{VetoThresholdAbsent: 20}),
	} {
		for _, a := range tally.Anomalies() {
			text := a.Arithmetic()
			if strings.TrimSpace(text) == "" {
				t.Errorf("%s produced an empty arithmetic line", a.Kind)
			}
			kinds[a.Kind] = text
		}
	}
	if len(kinds) != 2 {
		t.Errorf("the two kinds produced %d distinct findings: %v; both have to be "+
			"reachable or one of them is documentation", len(kinds), kinds)
	}
	if kinds[TallyOvercounted] == kinds[TallyPassedWithoutThreshold] {
		t.Error("the two kinds word themselves identically")
	}
}

// TestTheStructurallyZeroNotesAreUntouched is task 4.5.
//
// The identity alarm goes beside the existing sentences rather than instead of them.
// They say different things: the note explains a zero that is honest while no
// threshold is approved, and the alarm says the counts contradict each other. A
// change that replaced one with the other would leave the zero unexplained, and the
// most natural "fix" for an unexplained zero is the one D18 forbids.
func TestTheStructurallyZeroNotesAreUntouched(t *testing.T) {
	ordinary := talliedAs(30, 0, nil,
		map[VetoCode]int{VetoSeenLate: 30, VetoExtended: 30, VetoNearHigh: 29},
		map[VetoUnmeasured]int{VetoThresholdAbsent: 60, VetoUnmeasured(LevelNoDayHigh): 29})
	if got := ordinary.Anomalies(); len(got) != 0 {
		t.Errorf("the state this system is in every day reported %v as an anomaly; the "+
			"structural zero is not a fault", got)
	}
}
