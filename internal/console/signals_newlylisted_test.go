package console

// signals_newlylisted_test.go is task 1.4: the new-entrant marker, which until this
// change could not appear on the screen at all.
//
// candidate.Row.NewlyListed was declared on five types, copied through the scan,
// persisted, read back and rendered here — and no production source ever assigned
// it, so `if s.NewlyListed` was structurally false and this cell was always blank.
// Nothing failed. A screen that says nothing about a fact and a screen whose fact is
// always negative look identical, which is why the negative and the unknown are
// rendered too: a blank is what the defect looked like.

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// sightingVerdict is one candidate whose first sighting was measured, carrying the
// given new-entrant answer.
func sightingVerdict(symbol string, sighting candidate.Sighting) candidate.Verdict {
	v := measuredAndClear(symbol)
	v.Sighting = sighting
	return v
}

// aMeasuredSighting is what MeasureFirstSighting produces for a qualified reading,
// built here directly because this test is about the rendering rather than about
// the measurement.
func aMeasuredSighting(rank, total int, newly candidate.NewlyListed) candidate.Sighting {
	return candidate.Sighting{
		Measured: true, Rank: rank, RankTotal: total, PercentilePct: "88",
		At: signalsNow.Add(-20 * time.Minute), Source: candidate.SourceOfficialTradingValue,
		FirstSeenAt: signalsNow.Add(-20 * time.Minute),
		NewlyListed: newly,
	}
}

// TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly.
func TestTheNewEntrantMarkerRendersAllThreeStatesDistinctly(t *testing.T) {
	rows := []candidate.Verdict{
		sightingVerdict("005930", aMeasuredSighting(12, 100, candidate.NewlyListedYes())),
		sightingVerdict("000660", aMeasuredSighting(12, 100, candidate.NewlyListedNo())),
		sightingVerdict("035720", aMeasuredSighting(12, 100, candidate.NewlyListedUnknown())),
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: oneMarket(rows...)}))

	texts := map[string]string{}
	for _, symbol := range []string{"005930", "000660", "035720"} {
		row := signalsRowFor(t, page, symbol)
		if !strings.Contains(row, "12 / 100") {
			t.Errorf("%s: the row does not carry the position it was measured from:\n%s",
				symbol, row)
		}
		texts[symbol] = row
	}

	if !strings.Contains(texts["005930"], "신규 진입") {
		t.Errorf("a symbol the source reported as absent from its previous reading carries no "+
			"new-entrant marker. This is the marker that could not be drawn at all before "+
			"this change:\n%s", texts["005930"])
	}
	if strings.Contains(texts["000660"], "신규 진입 미상") {
		t.Errorf("a measured `no` renders as the unknown state:\n%s", texts["000660"])
	}
	if !strings.Contains(texts["000660"], "직전 읽기에 있었음") {
		t.Errorf("a symbol the source reported as already listed says nothing about it; a "+
			"blank cell is exactly what the broken version looked like:\n%s", texts["000660"])
	}
	if !strings.Contains(texts["035720"], "신규 진입 미상") {
		t.Errorf("an unknown new-entrant fact renders blank rather than naming itself:\n%s",
			texts["035720"])
	}
	// And the three are actually different strings on the page, which a page-wide
	// Contains would not have shown.
	if texts["005930"] == texts["000660"] || texts["000660"] == texts["035720"] {
		t.Error("two of the three states render identically")
	}
}

// TestARefusedSightingNamesWhyRatherThanShowingTheMarker.
//
// When the fact is unknown the sighting is unmeasured, so the cell carries the
// reason rather than a position. That is the ordinary path for the first scan of a
// session, and the operator's next action — wait for the second reading — is only
// legible if the reason is the one that says the source had no previous reading.
func TestARefusedSightingNamesWhyRatherThanShowingTheMarker(t *testing.T) {
	refused := candidate.Sighting{
		Rank: 4, RankTotal: 100, Why: candidate.VetoNewEntrantUnknown,
		At: signalsNow, Source: candidate.SourceOfficialTradingValue,
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(sightingVerdict("005930", refused)),
	}))
	row := signalsRowFor(t, page, "005930")
	if !strings.Contains(row, string(candidate.VetoNewEntrantUnknown)) {
		t.Errorf("the row does not name %s; an unmeasured cell with no reason is the display "+
			"this screen exists to remove:\n%s", candidate.VetoNewEntrantUnknown, row)
	}
	if strings.Contains(row, "신규 진입 미상") {
		t.Errorf("the refused row also draws the marker; the cell is one thing or the "+
			"other:\n%s", row)
	}
}

// TestATruncatedReadingIsNamedOnTheScreenToo.
func TestATruncatedReadingIsNamedOnTheScreenToo(t *testing.T) {
	refused := candidate.Sighting{
		Rank: 1, RankTotal: 3, Why: candidate.VetoReadingTruncated,
		At: signalsNow, Source: candidate.SourceOfficialTradingValue,
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(sightingVerdict("005930", refused)),
	}))
	row := signalsRowFor(t, page, "005930")
	if !strings.Contains(row, string(candidate.VetoReadingTruncated)) {
		t.Errorf("the row does not name %s, so a reading that came back short reads like any "+
			"other missing measurement:\n%s", candidate.VetoReadingTruncated, row)
	}
}
