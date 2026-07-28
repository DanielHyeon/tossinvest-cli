package candidatesrc

// retiredsource_test.go pins the one thing `retire-gainers-source` changes: the
// gainers ranking is not on any market's panel.
//
// # Why the assertion is an id and not a length
//
// "the KR panel has three sources" breaks for reasons that have nothing to do with
// this decision — a market without a WTS session builds one fewer — and it does not
// break for the reason that matters, because a panel that swapped one source for
// another still has three. So the length is not the property. The property is that
// a particular id is absent, and it is asserted in both markets because the
// membership rule is per market (Panel's doc) and a source could come back on one
// of them.
//
// # What it does not assert
//
// It does not say the constant is gone, because the constant is deliberately not
// gone (design D2). candidate.SourceOfficialGainers, candidatesrc.RankingTopGainers
// and the rankingSourceID entry all stay, so that reversing this decision is one
// slice element rather than a rediscovery of why the three rankings needed separate
// ids. What is retired is the wiring, and this is the test that says so.

import (
	"sort"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// sameSourceSet compares two id sets.
//
// It lives here rather than in clock_wiring_test.go, which is its other caller, for
// the reason the whole change keeps its diff narrow: the edit to that file has to
// stay at the one function that had to change.
func sameSourceSet(got, want map[candidate.SourceID]bool) bool {
	if len(got) != len(want) {
		return false
	}
	for id := range want {
		if !got[id] {
			return false
		}
	}
	return true
}

// sortedIDs renders a set stably, so a failure reads the same on two runs.
func sortedIDs(set map[candidate.SourceID]bool) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, string(id))
	}
	sort.Strings(out)
	return out
}

// TestNoMarketPanelBuildsTheGainersRanking.
//
// Every reader is non-nil, which is the maximal panel: a nil official reader would
// produce no rankings at all and the assertion would hold vacuously.
func TestNoMarketPanelBuildsTheGainersRanking(t *testing.T) {
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		market := market
		t.Run(market, func(t *testing.T) {
			panel := Panel(market, &fakeRankings{}, nil, &fakePopular{}, nil)
			if len(panel) == 0 {
				t.Fatalf("the %s panel is empty; an assertion about what it does not "+
					"contain says nothing about an empty panel", market)
			}
			for _, src := range panel {
				if src.ID() != candidate.SourceOfficialGainers {
					continue
				}
				t.Errorf("the %s panel builds %s. That ranking rejects the duration this "+
					"package sends — the API answers 400 unsupported-ranking-duration, so "+
					"it has never returned a row — and a source that cannot answer reports "+
					"itself as a degradation on every scan, which is how the degradation "+
					"flag stopped meaning anything.\n\n"+
					"If it is coming back, the four conditions are in the proposal: a "+
					"measured threshold for seen_late, shadow bands that resolve the range "+
					"this list lives in, an intraday observation of how often the 1d "+
					"ranking updates, and a design for ranks measured over different "+
					"windows sharing one percentile.", market, src.ID())
			}
		})
	}
}

// TestTheRetiredSourceKeepsItsIdentity is the other half of D2, and it is the half
// somebody deletes while tidying.
//
// The constant, the ranking type and the mapping between them stay. Two reasons and
// they are independent: reversing the decision has to be one edit rather than a
// rediscovery of why each ranking type needed its own source id, and a row already
// filed under this id in a store would otherwise be read back as a string nothing
// recognises.
func TestTheRetiredSourceKeepsItsIdentity(t *testing.T) {
	if RankingTopGainers != "TOP_GAINERS" {
		t.Errorf("RankingTopGainers = %q; it is the API's own string and the snapshot "+
			"guard compares against that value", RankingTopGainers)
	}
	id, ok := rankingSourceID[RankingTopGainers]
	if !ok {
		t.Fatalf("rankingSourceID no longer maps %s. Retiring the wiring is one slice "+
			"element; deleting the mapping is a rediscovery of why three rankings on one "+
			"id let a rate limit on one of them cool candidates only that one ever raised",
			RankingTopGainers)
	}
	if id != candidate.SourceOfficialGainers {
		t.Errorf("rankingSourceID[%s] = %q, want %q — a stored observation carries the "+
			"id it was written with", RankingTopGainers, id, candidate.SourceOfficialGainers)
	}
}
