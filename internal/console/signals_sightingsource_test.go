package console

// signals_sightingsource_test.go is the per-source first-sighting block: which
// source's readings seen_late is refusing.
//
// The veto census on this page counts reasons. It cannot say whose readings produced
// them, and that is the difference between a number to watch and a source to go and
// look at — in particular, whether the official rankings endpoint is returning the
// hundred rows it is asked for. If it is not, every first sighting taken from it is
// refused, seen_late has no distribution at all, and the follow-up change that
// chooses a threshold has nothing to choose from. That question has never been
// measured, and this block is where one session's screen answers it.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// twoSourcesOneShort is a market whose refusals come from two different sources —
// one endpoint coming back short, one panel that has just started.
func twoSourcesOneShort() SignalsReading {
	short := candidate.Sighting{
		Rank: 1, RankTotal: 3, Why: candidate.VetoReadingTruncated,
		At: signalsNow, Source: candidate.SourceOfficialTradingValue,
	}
	unqualified := candidate.Sighting{
		Rank: 4, RankTotal: 30, Why: candidate.VetoNewEntrantUnknown,
		At: signalsNow, Source: candidate.SourceWTSPopular,
	}
	reading := oneMarket(
		sightingVerdict("005930", short),
		sightingVerdict("000660", short),
		sightingVerdict("035720", unqualified),
	)
	reading.Markets[0].Sightings =
		candidate.TallySightingSources(reading.Markets[0].Verdicts)
	return reading
}

// TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem.
func TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: twoSourcesOneShort()}))

	for _, want := range []string{
		string(candidate.SourceOfficialTradingValue),
		string(candidate.SourceWTSPopular),
		string(candidate.VetoReadingTruncated),
		string(candidate.VetoNewEntrantUnknown),
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the page does not name %q. Without it a screen of refusals says how "+
				"many and why and never whose, and the two sources here call for opposite "+
				"actions: one endpoint is short, one panel has just started", want)
		}
	}

	block, ok := sectionAfter(page, "최초 관측 — 원천별")
	if !ok {
		t.Fatalf("the per-source first-sighting block is not on the page")
	}
	// The counts are the reducer's, and they have to be attributed rather than
	// summed: two refusals under the truncated reason belong to the official
	// ranking, and one under the unqualified reason to the WTS list.
	if !strings.Contains(block, "0 / 2") {
		t.Errorf("the official ranking's two refused sightings are not rendered as 0 of 2:\n%s",
			block)
	}
	if !strings.Contains(block, "0 / 1") {
		t.Errorf("the WTS list's single refused sighting is not rendered as 0 of 1:\n%s", block)
	}
}

// TestThePerSourceBlockIsAbsentRatherThanEmptyWhenNothingHasASighting.
//
// A table of zeroes reads as "every source is fine". A market with no stored first
// sighting anywhere has not measured any source, and the page has other places that
// already say so — the veto census leads with the unmeasured count.
func TestThePerSourceBlockIsAbsentRatherThanEmptyWhenNothingHasASighting(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(measuredAndClear("005930")),
	}))
	if strings.Contains(page, "최초 관측 — 원천별") {
		t.Error("the per-source block is drawn for a market in which no candidate holds a " +
			"first sighting; an empty table of sources reads as a set of sources that are fine")
	}
}

// sectionAfter returns the page text from the first occurrence of `heading` to the
// end of the enclosing table, so an assertion about this block cannot be satisfied by
// a number somewhere else on a long page.
func sectionAfter(page, heading string) (string, bool) {
	i := strings.Index(page, heading)
	if i < 0 {
		return "", false
	}
	rest := page[i:]
	if j := strings.Index(rest, "</table>"); j >= 0 {
		return rest[:j], true
	}
	return rest, true
}
