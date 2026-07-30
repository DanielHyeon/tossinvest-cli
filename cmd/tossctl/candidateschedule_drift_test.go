package main

// candidateschedule_drift_test.go holds the polling schedule to the panel.
//
// # The mistake this is written about
//
// Retiring a source is two edits in two packages: the source leaves
// candidatesrc.Panel and its interval leaves candidateIntervals. Doing one of them
// is the easiest error in the whole change and it fails nothing. An interval with
// no source behind it makes no call, raises no error and returns no row; all it
// does is tell the next reader that the source is alive and read at that cadence,
// and that reader then tunes a number for a source nobody reads.
//
// # The obligation runs one way
//
// A source on the panel with no interval configured is NOT a violation. It gets
// unconfiguredFloor — the slowest configured cadence in the contract, applied as
// the cautious default for exactly this case, a panel gaining a source before
// somebody writes its interval down (internal/candidate/source.go). Asserting the
// two sets are equal would make that designed behaviour fail a test, so the
// assertion is containment: every scheduled id has a source that can produce it.
//
// # Why the comparison is against candidatesrc.Panel and not buildCandidatePanel
//
// buildCandidatePanel loads Open API credentials and constructs a client, so it
// cannot run in a test at all. Panel is a pure function of its arguments and this
// command already imports it.

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// scheduleProbeRankings is a RankingReader that is never read. Panel needs a
// non-nil reader to build the official sources and this test only asks what ids
// they carry.
type scheduleProbeRankings struct{}

func (scheduleProbeRankings) Rankings(context.Context, string, string, string, bool, int) (
	domain.Ranking, error) {
	return domain.Ranking{}, nil
}

// scheduleProbeBudget is the same for the optional rate-budget accessor.
type scheduleProbeBudget struct{}

func (scheduleProbeBudget) RateBudget(string) official.RateBudget {
	return official.RateBudget{}
}

// scheduleProbePopular is the same for the WTS popularity list.
type scheduleProbePopular struct{}

func (scheduleProbePopular) GetStockRanking(context.Context, int) (domain.StockRanking, error) {
	return domain.StockRanking{}, nil
}

// everySourceThePanelCanBuild is the union over the markets, which is the maximal
// set an interval may legitimately name.
//
// KR carries it: the WTS popularity list is a Korean-market product and Panel
// leaves it out of every other market, so the US panel is a subset. Both are built
// anyway, because "US is a subset" is a property of Panel rather than of this test,
// and a change to it should not quietly shrink what this guard compares against.
func everySourceThePanelCanBuild(t *testing.T) map[candidate.SourceID]bool {
	t.Helper()
	out := map[candidate.SourceID]bool{}
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		panel := candidatesrc.Panel(market, scheduleProbeRankings{}, scheduleProbeBudget{},
			scheduleProbePopular{}, nil)
		if len(panel) == 0 {
			t.Fatalf("the %s panel is empty with every reader supplied; a containment "+
				"check against an empty set fails everything or nothing and means neither",
				market)
		}
		for _, src := range panel {
			out[src.ID()] = true
		}
	}
	return out
}

// scheduledWithNoSource is the rule, as a function so that both directions of it
// can be run over a fixture.
//
// A rule spelled only inside a t.Errorf loop is one nothing can check: a guard that
// stopped catching its own case reports success, which is the failure mode this
// whole change is about.
func scheduledWithNoSource(intervals map[candidate.SourceID]candidate.Interval,
	buildable map[candidate.SourceID]bool) []string {

	var out []string
	for id := range intervals {
		if !buildable[id] {
			out = append(out, string(id))
		}
	}
	sort.Strings(out)
	return out
}

// TestNoIntervalNamesASourceNoPanelBuilds.
func TestNoIntervalNamesASourceNoPanelBuilds(t *testing.T) {
	buildable := everySourceThePanelCanBuild(t)
	intervals := candidateIntervals()
	if len(intervals) == 0 {
		t.Fatal("candidateIntervals is empty; every source would fall back to the " +
			"unconfigured floor and this containment check would hold vacuously")
	}
	if orphans := scheduledWithNoSource(intervals, buildable); len(orphans) > 0 {
		t.Errorf("candidateIntervals sets a cadence for %v and no market's panel builds "+
			"those sources.\n\n"+
			"This costs nothing at runtime and that is the problem: the entry makes no "+
			"call and reports no error, it only tells whoever reads this map next that "+
			"the source is alive and polled at that interval. Retiring a source is two "+
			"edits — the panel and this map — and doing one of them fails nothing "+
			"until here.\n\n"+
			"Sources the panel can build: %v", orphans, sortedBuildableIDs(buildable))
	}
}

// TestTheScheduleGuardIsContainmentAndNotEquality runs the rule over both
// directions.
//
// The one it must catch is an interval with no source. The one it must NOT catch is
// a source with no interval: that source gets unconfiguredFloor, which
// internal/candidate/source.go put there for a panel that gains a source before
// anybody writes its cadence down. An equality check would make that designed
// behaviour a violation, and it is the check somebody reaches for while tightening
// this file.
func TestTheScheduleGuardIsContainmentAndNotEquality(t *testing.T) {
	const retired = candidate.SourceOfficialGainers
	live := candidate.SourceOfficialTradingValue
	every := candidate.Interval{Every: 15 * time.Second, Floor: 5 * time.Second}

	scheduledOnly := scheduledWithNoSource(
		map[candidate.SourceID]candidate.Interval{live: every, retired: every},
		map[candidate.SourceID]bool{live: true})
	if len(scheduledOnly) != 1 || scheduledOnly[0] != string(retired) {
		t.Errorf("an interval for %q with no source behind it was read as %v, want "+
			"exactly that id. This is the case the guard exists for and a guard that "+
			"stopped catching it passes", retired, scheduledOnly)
	}

	panelOnly := scheduledWithNoSource(
		map[candidate.SourceID]candidate.Interval{live: every},
		map[candidate.SourceID]bool{live: true, candidate.SourceWTSPopular: true})
	if len(panelOnly) != 0 {
		t.Errorf("a panel source with no configured interval was reported as a "+
			"violation (%v). It is not one — it reads at the unconfigured floor, which "+
			"is the cautious default written for exactly this case — and requiring the "+
			"two sets to be equal turns a designed behaviour into a test failure",
			panelOnly)
	}
}

// sortedBuildableIDs renders a set stably so two runs produce the same message.
//
// Its own function rather than candidate.go's sortedSourceIDs: that one orders a
// map of ReadingFacts, which is what a scan report holds, and this one is over a
// set.
func sortedBuildableIDs(set map[candidate.SourceID]bool) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, string(id))
	}
	sort.Strings(out)
	return out
}
