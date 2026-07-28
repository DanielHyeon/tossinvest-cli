package candidatesrc

// clock_wiring_test.go pins the seam the memory's age is measured against, and the
// two halves of usableAt that refuse a comparison there is no instant for.
//
// # Why the seam needs a test of its own
//
// The injected clock is the only reason the freshness bound is testable at all: with
// clock.System() every reading in a test is microseconds after the last one, so the
// bound is never crossed and a build that dropped it would look identical. Panel is
// the one production caller, it hands the clock to four constructors, and passing nil
// in any of them compiles, keeps every other test green, and silently returns that
// source to the system clock — which is the state F1 found and fixed everywhere
// except here. issues.md I15 recorded the gap: the validity tests all call the two
// constructors directly.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// aFullRanking is a reading of exactly n rows, which is what Panel's official
// sources ask for. Anything shorter is refused as truncated before the clock is
// consulted, and a test built on a short reading would pass against a source holding
// no clock at all.
func aFullRanking(n int) domain.Ranking {
	out := domain.Ranking{}
	for i := 0; i < n; i++ {
		out.Items = append(out.Items, domain.RankingItem{
			Rank: i + 1, Symbol: fmt.Sprintf("%06d", i+1),
		})
	}
	return out
}

// aFullPopularity is the same for the WTS list.
func aFullPopularity(n int) domain.StockRanking {
	out := domain.StockRanking{}
	for i := 0; i < n; i++ {
		out.Stocks = append(out.Stocks, domain.RankedStock{
			Rank: i + 1, Symbol: fmt.Sprintf("%06d", i+1),
		})
	}
	return out
}

// TestThePanelHandsItsClockToEverySourceItBuilds.
//
// Driven through the bound rather than through the field: a source that kept the
// injected clock stops answering when the fake advances past previousReadingTTL, and
// a source that fell back to the system clock keeps answering, because in real time
// these three readings are microseconds apart.
func TestThePanelHandsItsClockToEverySourceItBuilds(t *testing.T) {
	clk := clock.NewFake(t0())
	rankings := &fakeRankings{out: aFullRanking(100)}
	popular := &fakePopular{out: aFullPopularity(30)}
	sources := Panel(candidate.MarketKR, rankings, nil, popular, clk)
	if len(sources) != 4 {
		t.Fatalf("the KR panel has %d sources, want 4 (three rankings and the popularity list)",
			len(sources))
	}
	ctx := context.Background()

	for _, src := range sources {
		src := src
		t.Run(string(src.ID()), func(t *testing.T) {
			// Reading one: nothing to compare against, and it becomes the memory.
			if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
				t.Fatalf("first Read: %v", err)
			}
			// Reading two, one tick later: the memory answers.
			clk.Advance(15 * time.Second)
			fresh, err := src.Read(ctx, candidate.MarketKR)
			if err != nil {
				t.Fatalf("second Read: %v", err)
			}
			if len(fresh.Rows) == 0 {
				t.Fatalf("the second reading carried no rows; a loop over an empty slice asserts " +
					"nothing and passes")
			}
			if got := knownAnswers(fresh.Rows); got != len(fresh.Rows) {
				t.Fatalf("%d of %d rows answered one tick after a whole reading; this test cannot "+
					"say anything about the clock until the ordinary case works",
					got, len(fresh.Rows))
			}
			// Reading three, past the bound: it does not.
			clk.Advance(previousReadingTTL)
			stale, err := src.Read(ctx, candidate.MarketKR)
			if err != nil {
				t.Fatalf("third Read: %v", err)
			}
			if got := knownAnswers(stale.Rows); got != 0 {
				t.Errorf("%d of %d rows still answered %s after the previous reading, measured on "+
					"the injected clock. This source is dating its memory from somewhere else — "+
					"Panel builds four of them and a nil in any one constructor returns that one "+
					"to clock.System(), where these three readings are microseconds apart",
					got, len(stale.Rows), previousReadingTTL)
			}
		})
	}
}

// TestAMemoryWithNoInstantOnEitherSideIsNotAnAnswer is usableAt's two zero-instant
// halves.
//
// Neither is reachable through clock.System(), and that is the argument for pinning
// them rather than against it: an unreachable guard is one nothing fails when it is
// removed, and this one is three lines away from the subtraction it protects.
//
// Of the four cases below only one distinguishes the guard, and that is recorded
// rather than glossed. A dated memory asked about at the zero instant produces a
// hugely negative age and is refused by the sign test; an undated memory asked about
// at a real instant produces a fifty-six-year-old one and is refused by the TTL. The
// case the explicit test catches is both at once — age zero, which reads as a memory
// recorded this instant, and answers `no` for every symbol it holds.
func TestAMemoryWithNoInstantOnEitherSideIsNotAnAnswer(t *testing.T) {
	set := map[string]bool{"005930": true}
	for _, tc := range []struct {
		name   string
		memory previousReading
		asking time.Time
	}{
		{"a memory that was never dated", previousReading{symbols: set}, t0()},
		{"a reading with no instant of its own",
			previousReading{symbols: set, at: t0()}, time.Time{}},
		{"both", previousReading{symbols: set}, time.Time{}},
		{"no memory at all", previousReading{at: t0()}, t0()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.memory.usableAt(tc.asking) {
				t.Errorf("usableAt returned true for %s; the age is a subtraction and a zero "+
					"instant on either side of it is not a statement about when anything was read",
					tc.name)
			}
		})
	}
	// And the ordinary case still answers, so the guards above are a refusal rather
	// than a function that never returns true.
	if !(previousReading{symbols: set, at: t0()}).usableAt(t0().Add(time.Second)) {
		t.Error("a one-second-old memory was refused; every case in this test would pass " +
			"against a usableAt that answered false unconditionally")
	}
}
