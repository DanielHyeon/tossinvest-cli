package console

// band_alarm_test.go is the collapse alarm at the browser surface.
//
// The requirement does not name a surface: "측정된 기록이 하나 이상인데 그 전부가 같은
// 교차 집합을 냈으면, 보고 표면은 그것을 정상 수치로 표시하지 않고 경보로 표시해야
// 한다." The scan output learned that first and this page did not, so the same
// candidate.BandTally rendered as an alarm in a terminal and as an ordinary row in a
// browser — and the browser is the one an operator leaves open.
//
// Same rule, same numbers, different sentence. The judgement is
// candidate.BandTally.Collapsed and it is asserted on both surfaces so that the two
// cannot drift into disagreeing about whether a scale resolved anything.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// marketWithBands is one market block whose shadow tallies are supplied directly.
//
// They are built here rather than run through TallyBands because the state under
// test is a distribution, and producing a collapsed one from real candidates would
// make the test depend on the scale's values — the very numbers this change moved.
//
// It carries no verdicts, and that is not laziness: TestOnlyTheListedFilesCanNameThe
// ChaseVerdict keeps a list of the files allowed to name one, and a file that does
// not need a chase must not be added to it. The band block renders from the tally
// alone.
func marketWithBands(bands map[candidate.VetoCode]candidate.BandTally) SignalsReading {
	return SignalsReading{
		At: signalsNow,
		Markets: []SignalsMarket{{
			Market: candidate.MarketKR,
			Bands:  bands,
			Panel:  SignalsPanel{Known: true, At: signalsNow, Attempted: 5, Responded: 5},
		}},
	}
}

// bandTally is a tally of the extended code with the crossings named.
func bandTally(measured int, crossed map[string]int) candidate.BandTally {
	full := map[string]int{}
	for _, band := range candidate.BandsFor(candidate.VetoExtended) {
		full[band] = crossed[band]
	}
	return candidate.BandTally{
		Code: candidate.VetoExtended, Total: measured + 25, Measured: measured,
		Crossed:     full,
		NotMeasured: map[candidate.VetoUnmeasured]int{candidate.VetoNoObservations: 25},
	}
}

// TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing.
//
// 131 measured records that all crossed exactly band 0 is not a hypothetical: it is
// what this repository's own KR store produces, and before this test the page drew
// it as a row of honest counts with nothing to distinguish it from a working scale.
func TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing(t *testing.T) {
	tally := bandTally(131, map[string]int{"0": 131})
	if !tally.Collapsed() {
		t.Fatal("the fixture is not collapsed; this test would pass for the wrong reason")
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: marketWithBands(map[candidate.VetoCode]candidate.BandTally{
			candidate.VetoExtended: tally,
		}),
	}))

	if !strings.Contains(page, "경보") {
		t.Errorf("131 measured records that all produced the same crossings render with no "+
			"alarm; a scale that resolved nothing reads as a working one:\n%s",
			truncateForLog(page))
	}
	if !strings.Contains(page, "131") {
		t.Error("the measured count is gone from the page; the alarm goes beside the numbers " +
			"rather than instead of them — they are what it is about")
	}
	// The counts stay too, for the same reason.
	if !strings.Contains(page, "extended") {
		t.Error("the alarm does not name the code whose scale resolved nothing")
	}
}

// TestAShadowScaleThatSeparatedSomethingRaisesNoAlarm is the negative control.
//
// An alarm that is on every day is one nobody reads, and this page's other alarm
// carries the same control for the same reason. The US half of this repository's
// store is exactly this state: band 0 split the candidates and the rest did not.
func TestAShadowScaleThatSeparatedSomethingRaisesNoAlarm(t *testing.T) {
	tally := bandTally(169, map[string]int{"0": 165})
	if tally.Collapsed() {
		t.Fatal("the fixture is collapsed; the control is not controlling anything")
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: marketWithBands(map[candidate.VetoCode]candidate.BandTally{
			candidate.VetoExtended: tally,
		}),
	}))
	if strings.Contains(page, "경보") {
		t.Errorf("a scale that separated 165 from 4 renders an alarm:\n%s", truncateForLog(page))
	}
	if !strings.Contains(page, "169") {
		t.Error("the measured count is missing from the ordinary render")
	}
}
