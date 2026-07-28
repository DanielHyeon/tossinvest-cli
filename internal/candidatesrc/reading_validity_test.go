package candidatesrc

// reading_validity_test.go is the two conditions that decide whether a source's
// remembered reading may answer anything: it has to have arrived whole, and it has
// to be recent enough to be the reading one tick ago.
//
// # Why existence was not enough
//
// Both adapters return before the swap when the read fails, so the memory survives
// an outage untouched. That is the case where it is least entitled to speak and
// where it used to speak loudest: cooling plus staleness is forty minutes, so any
// gap longer than that expires the whole watchlist, and the recovering scan
// re-promotes the entire panel while every symbol that never left the list is
// answered `no` from a set recorded before the outage began. `no` is a measured
// answer, so each of those positions became a first sighting seen_late was allowed
// to measure — which is the path design D3 exists to cut, reached from the other
// side.
//
// # Why a short reading was not enough either
//
// The swap used to be unconditional. Three rows arriving out of a hundred requested
// then became the yardstick, and the next complete reading found ninety-seven
// symbols "absent from the previous reading" and reported all of them as new
// entrants — stored on the candidate rows and rendered as 신규 진입 for the rest of
// those lives.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// t0 is an ordinary trading instant. A function rather than a package variable so
// no test can move another's.
func t0() time.Time { return time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC) }

// knownAnswers counts the rows of a reading whose new-entrant fact was measured.
func knownAnswers(rows []candidate.Row) int {
	n := 0
	for _, r := range rows {
		if r.NewlyListed.Known() {
			n++
		}
	}
	return n
}

// TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer is F1's probe.
//
// A source that spends longer than previousReadingTTL failing comes back holding a
// set from before the failure. Every symbol still on the list would answer `no`
// against it, and `no` is the answer that qualifies a first sighting — arriving at
// exactly the moment the whole watchlist has expired and is being re-promoted.
func TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer(t *testing.T) {
	clk := clock.NewFake(t0())
	f := &fakeRankings{out: rankingOf("005930", "000660")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 2, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// The outage. Every one of these returns before the swap, so the memory is
	// exactly as it was — which is the whole point.
	f.err = official.ErrRateLimited
	for i := 0; i < 3; i++ {
		clk.Advance(previousReadingTTL / 2)
		if _, err := src.Read(ctx, candidate.MarketKR); !errors.Is(err, candidate.ErrRateLimited) {
			t.Fatalf("read %d during the outage returned %v, want a rate-limited loss", i, err)
		}
	}

	f.err = nil
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the recovering Read: %v", err)
	}
	if len(back.Rows) != 2 {
		t.Fatalf("the recovering reading carried %d rows, want 2", len(back.Rows))
	}
	if got := knownAnswers(back.Rows); got != 0 {
		t.Errorf("%d of %d rows carried a measured new-entrant fact after an outage of %s. "+
			"The set they were compared against was recorded before it began, so `no` there "+
			"means \"this symbol was in the list an hour ago\" — and by now cooling plus "+
			"staleness has expired the watchlist, so that answer is about to qualify a "+
			"re-promotion as a first sighting",
			got, len(back.Rows), 3*(previousReadingTTL/2))
	}

	// And the recovering reading is itself remembered, so the next one can answer.
	clk.Advance(time.Second)
	next, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the Read after recovery: %v", err)
	}
	if got := knownAnswers(next.Rows); got != 2 {
		t.Errorf("%d of %d rows carried an answer one tick after recovery; the refusal is "+
			"about the age of the memory, not a source that has stopped answering", got,
			len(next.Rows))
	}
}

// TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore pins the bound to the value it
// was derived from rather than to a number spelled here.
//
// previousReadingTTL is candidate.DefaultStalenessTTL: twice the last rung of
// candidate.BackoffLadder, which is the gap the design guarantees a source may go
// without being read, and also the span after which a candidate nobody re-promoted
// begins to cool. Below it the memory answers; at it and above it, it does not.
func TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore(t *testing.T) {
	if previousReadingTTL != candidate.DefaultStalenessTTL {
		t.Fatalf("previousReadingTTL = %s and candidate.DefaultStalenessTTL = %s; the bound "+
			"is that value rather than a number chosen here, and the argument written at "+
			"previousReadingTTL is about that value's own derivation",
			previousReadingTTL, candidate.DefaultStalenessTTL)
	}
	if want := 2 * candidate.BackoffLadder[len(candidate.BackoffLadder)-1]; previousReadingTTL != want {
		t.Errorf("previousReadingTTL = %s and twice the longest backoff rung is %s. The "+
			"derivation is the ladder's, and a ladder that climbed somewhere else has to "+
			"move this bound with it", previousReadingTTL, want)
	}

	for _, tc := range []struct {
		name string
		gap  time.Duration
		want int
	}{
		{"one tick", 15 * time.Second, 2},
		{"a full backoff ladder", candidate.BackoffLadder[len(candidate.BackoffLadder)-1], 2},
		{"one instant short of the bound", previousReadingTTL - time.Nanosecond, 2},
		{"exactly the bound, which is closed at the near end", previousReadingTTL, 0},
		{"an outage", 4 * previousReadingTTL, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(t0())
			f := &fakeRankings{out: rankingOf("005930", "000660")}
			src, err := OfficialRanking(f, nil, RankingTradingAmount, 2, clk)
			if err != nil {
				t.Fatalf("OfficialRanking: %v", err)
			}
			ctx := context.Background()
			if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
				t.Fatalf("first Read: %v", err)
			}
			clk.Advance(tc.gap)
			second, err := src.Read(ctx, candidate.MarketKR)
			if err != nil {
				t.Fatalf("second Read: %v", err)
			}
			if got := knownAnswers(second.Rows); got != tc.want {
				t.Errorf("%d of %d rows answered after a gap of %s, want %d",
					got, len(second.Rows), tc.gap, tc.want)
			}
		})
	}
}

// TestAShortReadingDoesNotBecomeTheYardstickForTheNextWholeOne is F2's probe,
// scaled down: three rows arrive out of the hundred this source asked for, and then
// a whole reading arrives.
//
// Under the unconditional swap the whole reading found every symbol outside those
// three "absent from the previous reading" and reported each of them as a new
// entrant. Those are measured facts: they reach candidates.first_rank_newly_listed
// and the 신규 진입 marker, and nothing downstream can tell them from real ones.
func TestAShortReadingDoesNotBecomeTheYardstickForTheNextWholeOne(t *testing.T) {
	clk := clock.NewFake(t0())
	whole := rankingOf("005930", "000660", "035720", "068270")
	f := &fakeRankings{out: whole}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 4, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// The degraded reading: one row of the four requested.
	clk.Advance(15 * time.Second)
	f.out = rankingOf("005930")
	short, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("short Read: %v", err)
	}
	if len(short.Rows) != 1 || short.Rows[0].RankRequested != 4 || short.Rows[0].RankTotal != 1 {
		t.Fatalf("the short reading = %d rows, %d requested, %d arrived; the fixture is not "+
			"the shape this test is about", len(short.Rows),
			short.Rows[0].RankRequested, short.Rows[0].RankTotal)
	}

	// And the whole one behind it.
	clk.Advance(15 * time.Second)
	f.out = whole
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the whole Read after the short one: %v", err)
	}
	if len(back.Rows) != 4 {
		t.Fatalf("the whole reading carried %d rows, want 4", len(back.Rows))
	}
	for _, r := range back.Rows {
		if r.NewlyListed.Yes() {
			t.Errorf("%s was reported as a new entrant on a list it never left. The reading "+
				"in between arrived with one row of four requested, and a truncated reading "+
				"is not a list — treating it as one turns a degradation into a panel full of "+
				"신규 진입", r.Symbol)
		}
		if !r.NewlyListed.No() {
			t.Errorf("%s answered %s; the last whole reading is still the previous reading "+
				"and this symbol was in it", r.Symbol, r.NewlyListed)
		}
	}
}

// TestAGenuinelyNewSymbolIsStillFoundAfterAShortReading is the other direction, so
// that the fix above is a narrowing rather than a silencing.
func TestAGenuinelyNewSymbolIsStillFoundAfterAShortReading(t *testing.T) {
	clk := clock.NewFake(t0())
	f := &fakeRankings{out: rankingOf("005930", "000660")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 2, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}
	clk.Advance(15 * time.Second)
	f.out = rankingOf("005930")
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("short Read: %v", err)
	}
	clk.Advance(15 * time.Second)
	f.out = rankingOf("005930", "068270")
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("third Read: %v", err)
	}
	got := newlyListedBySymbol(back.Rows)
	if !got["068270"].Yes() {
		t.Errorf("068270 answered %s; it was not in the last whole reading and this one is "+
			"whole, so the comparison is available and its answer is yes", got["068270"])
	}
	if !got["005930"].No() {
		t.Errorf("005930 answered %s; it has been in every reading", got["005930"])
	}
}

// TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh.
//
// The age is a subtraction, and a subtraction with no floor reads a memory recorded
// after the reading asking about it as arbitrarily fresh. A negative age is refused
// for the same reason metrics.go refuses READINGS_ALL_LATER: it is not a statement
// about the market.
func TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh(t *testing.T) {
	clk := clock.NewFake(t0())
	f := &fakeRankings{out: rankingOf("005930", "000660")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 2, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}
	clk.Set(t0().Add(-time.Hour))
	second, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("second Read: %v", err)
	}
	if got := knownAnswers(second.Rows); got != 0 {
		t.Errorf("%d of %d rows answered against a memory recorded an hour after the reading "+
			"asking; a previous reading in the future is not a previous reading",
			got, len(second.Rows))
	}
}

// TestTheWTSMemoryCarriesTheSameTwoConditions. Both adapters or neither: a fact
// learned from one source and not the other leaves the panel half measurable for a
// reason nothing on the screen explains.
func TestTheWTSMemoryCarriesTheSameTwoConditions(t *testing.T) {
	clk := clock.NewFake(t0())
	f := &fakePopular{out: popularityOf("005930", "000660")}
	src := WTSPopular(f, 2, clk)
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// Short: one of the two requested.
	clk.Advance(5 * time.Second)
	f.out = popularityOf("005930")
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("short Read: %v", err)
	}
	clk.Advance(5 * time.Second)
	f.out = popularityOf("005930", "000660")
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("third Read: %v", err)
	}
	got := newlyListedBySymbol(back.Rows)
	if !got["000660"].No() {
		t.Errorf("000660 answered %s after a one-row reading in between; the WTS adapter has "+
			"to hold the same two conditions as the official one", got["000660"])
	}

	// Stale: past the bound the answer goes back to unknown.
	clk.Advance(previousReadingTTL)
	stale, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("fourth Read: %v", err)
	}
	if n := knownAnswers(stale.Rows); n != 0 {
		t.Errorf("%d of %d WTS rows answered from a memory %s old", n, len(stale.Rows),
			previousReadingTTL)
	}
}
