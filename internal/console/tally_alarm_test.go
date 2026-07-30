package console

// tally_alarm_test.go is task 4.3 at the browser surface.
//
// Same rule as the scan output's, same numbers, different sentence — and the point
// of asserting both is that the two must not be able to disagree about whether the
// numbers add up. The judgement is internal/candidate's; what is checked here is
// that this page renders it and that it does not fire on the ordinary state.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// marketWithTally is one market block whose veto tally is supplied directly, so a
// contradiction can be shown that TallyVetoes cannot currently produce.
func marketWithTally(t candidate.VetoTally) SignalsReading {
	return SignalsReading{
		At: signalsNow,
		Markets: []SignalsMarket{{
			Market:   candidate.MarketKR,
			Verdicts: []candidate.Verdict{neverChecked("005930")},
			Vetoes:   t,
			Panel:    SignalsPanel{Known: true, At: signalsNow, Attempted: 5, Responded: 5},
		}},
	}
}

// TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself.
func TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself(t *testing.T) {
	tally := candidate.VetoTally{
		Total: 12, Passed: 12,
		Raised:      map[candidate.VetoCode]int{},
		NotMeasured: map[candidate.VetoCode]int{candidate.VetoNearHigh: 12},
		Reasons: map[candidate.VetoUnmeasured]int{
			candidate.VetoUnmeasured(candidate.LevelNoDayHigh): 12,
		},
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: marketWithTally(tally)}))

	if !strings.Contains(page, "경보") {
		t.Errorf("a tally whose buckets sum past its own total renders with no alarm; twelve "+
			"passes and twelve unmeasured on near_high read as a calm screen:\n%s",
			truncateForLog(page))
	}
	if !strings.Contains(page, "near_high") {
		t.Error("the alarm does not name the code whose arithmetic failed")
	}
	// The counts stay on the page. The alarm is about them, so a page that replaced
	// them would have hidden its own evidence.
	if !strings.Contains(page, "<dd>12") {
		t.Error("the passed count is gone from the page; the alarm goes beside the number " +
			"rather than instead of it")
	}
}

// TestTheOrdinarySignalsScreenRaisesNoAlarm is the negative control.
//
// Every candidate is unmeasured on seen_late and extended because neither has an
// approved threshold. That is the state this system is in every day, and an alarm
// on it would be noise from the moment it shipped — which is how a real alarm stops
// being read.
func TestTheOrdinarySignalsScreenRaisesNoAlarm(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(neverChecked("005930"), neverChecked("000660")),
	}))
	if strings.Contains(page, "경보") {
		t.Errorf("the ordinary no-threshold screen renders an alarm:\n%s", truncateForLog(page))
	}
	if !strings.Contains(page, signalsPassedNote) {
		t.Error("and the sentence explaining the structural zero is gone; the alarm was " +
			"added beside it, not instead of it")
	}
}
