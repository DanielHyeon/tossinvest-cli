package candidatesrc

// previous_reading_test.go is tasks 1.2 and 3.1: the two facts a ranking source
// knows about its own reading and used to throw away.
//
// Both are facts nothing downstream can recover. The store never sees the request,
// so it cannot know a hundred rows were asked for; and nothing anywhere holds "what
// this list looked like last time", because the store is keyed by symbol and the
// question is about one source's own consecutive readings. Discarding them here is
// what made candidate.Row.NewlyListed structurally false through five layers.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func rankingOf(symbols ...string) domain.Ranking {
	out := domain.Ranking{}
	for i, s := range symbols {
		out.Items = append(out.Items, domain.RankingItem{Rank: i + 1, Symbol: s})
	}
	return out
}

func popularityOf(symbols ...string) domain.StockRanking {
	out := domain.StockRanking{}
	for i, s := range symbols {
		out.Stocks = append(out.Stocks, domain.RankedStock{Rank: i + 1, Symbol: s})
	}
	return out
}

// newlyListedBySymbol indexes a reading's new-entrant answers.
func newlyListedBySymbol(rows []candidate.Row) map[string]candidate.NewlyListed {
	out := map[string]candidate.NewlyListed{}
	for _, r := range rows {
		out[r.Symbol] = r.NewlyListed
	}
	return out
}

// TestASourcesFirstReadingHasNoAnswerAboutNewEntrants is task 1.2's RED.
//
// "Absent from the previous reading" is vacuously true of every row of a first
// reading. A source that reported `yes` there would declare the entire panel newly
// listed on the first scan after every process start — and seen_late would then be
// measuring when our process came up rather than anything about the market.
func TestASourcesFirstReadingHasNoAnswerAboutNewEntrants(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930", "000660", "035720")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 100)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}

	first, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(first.Rows) != 3 {
		t.Fatalf("%d rows, want 3", len(first.Rows))
	}
	for _, r := range first.Rows {
		if r.NewlyListed.Known() {
			t.Errorf("%s: the source's first reading answered %s about a previous reading it "+
				"does not have; unknown is the only honest answer and it is not `yes`",
				r.Symbol, r.NewlyListed)
		}
	}
}

// TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed.
func TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930", "000660")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 100)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// 000660 leaves, 035720 arrives, 005930 stays.
	f.out = rankingOf("005930", "035720")
	second, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("second Read: %v", err)
	}
	got := newlyListedBySymbol(second.Rows)
	if !got["005930"].No() {
		t.Errorf("005930 was in the previous reading and answered %s", got["005930"])
	}
	if !got["035720"].Yes() {
		t.Errorf("035720 was not in the previous reading and answered %s", got["035720"])
	}

	// And a third reading measures against the second rather than against the first:
	// 000660 is back, and it left in between, so it is a new entrant now.
	f.out = rankingOf("005930", "000660")
	third, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("third Read: %v", err)
	}
	got = newlyListedBySymbol(third.Rows)
	if !got["000660"].Yes() {
		t.Errorf("000660 left the list and came back and answered %s; the comparison is "+
			"against the previous reading, not against every reading ever", got["000660"])
	}
}

// TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList is task 1.2's
// per-market clause.
//
// One adapter instance is handed whichever market the panel asks it for. A single
// symbol set would let the Korean reading decide whether a US symbol had just
// joined the US list — an answer about a different list, arriving as a measurement
// with nothing to distinguish it from a real one.
func TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930", "000660")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 100)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("KR Read: %v", err)
	}

	f.out = rankingOf("AAPL", "MSFT")
	us, err := src.Read(ctx, candidate.MarketUS)
	if err != nil {
		t.Fatalf("US Read: %v", err)
	}
	for _, r := range us.Rows {
		if r.NewlyListed.Known() {
			t.Errorf("%s: the first US reading answered %s, using the KR reading as its "+
				"previous one. A different list is not a previous reading of this one",
				r.Symbol, r.NewlyListed)
		}
	}

	// And the KR series is undisturbed by the US read in between.
	f.out = rankingOf("005930", "035720")
	kr, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("second KR Read: %v", err)
	}
	got := newlyListedBySymbol(kr.Rows)
	if !got["005930"].No() || !got["035720"].Yes() {
		t.Errorf("the KR answers after a US read in between = %v; the two markets' previous "+
			"readings have to be independent", got)
	}
}

// TestAnEmptyReadingIsStillAReadingOfThisList.
//
// The scan refuses to count an empty reading as coverage, because "answered with no
// rows" is not evidence that a symbol left. That is a different question from this
// one: as far as "what did this source see last time" goes, it saw nothing, and the
// rows of the next reading really were absent from it. Skipping the swap would make
// the next reading compare against a list two readings old.
func TestAnEmptyReadingIsStillAReadingOfThisList(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 100)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}
	f.out = domain.Ranking{}
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("empty Read: %v", err)
	}
	f.out = rankingOf("005930")
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("third Read: %v", err)
	}
	got := newlyListedBySymbol(back.Rows)
	if !got["005930"].Yes() {
		t.Errorf("005930 answered %s after the list it was in went empty and it came back; "+
			"the previous reading was the empty one", got["005930"])
	}
}

// TestTheRequestedRowCountTravelsBesideTheOneThatArrived is task 3.1.
//
// The source knows both numbers and used to report one. RankTotal alone cannot tell
// a three-row list from a hundred-row request that returned three, and the
// percentile normalises by RankTotal — so the top row of the second comes out at the
// 66.7th percentile of a list that never existed.
func TestTheRequestedRowCountTravelsBesideTheOneThatArrived(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930", "000660", "035720")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 100)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	got, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	for _, r := range got.Rows {
		if r.RankRequested != 100 {
			t.Errorf("%s: RankRequested = %d, want 100 — the number this source asked for, "+
				"which is the only number that can say the reading came back short",
				r.Symbol, r.RankRequested)
		}
		if r.RankTotal != 3 {
			t.Errorf("%s: RankTotal = %d, want 3 (what arrived)", r.Symbol, r.RankTotal)
		}
	}
}

// TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked.
//
// OfficialRanking caps count at the API's documented maximum, and what is reported
// has to be the capped number: a caller that asked for 500 and had it clipped to 100
// would otherwise mark every reading truncated for the rest of the process's life,
// on a comparison against a request that was never made.
func TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked(t *testing.T) {
	f := &fakeRankings{out: rankingOf("005930")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 500)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	got, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if f.gotN != 100 {
		t.Fatalf("the endpoint was asked for %d rows, want the capped 100", f.gotN)
	}
	if got.Rows[0].RankRequested != f.gotN {
		t.Errorf("RankRequested = %d and the endpoint was asked for %d; the two have to be "+
			"the same number or the truncation comparison is against a request nobody made",
			got.Rows[0].RankRequested, f.gotN)
	}
}

// TestTheWTSPopularityListReportsTheSameTwoFacts.
//
// It is the other ranking source and it has the same two numbers: a size it asks
// for and a previous reading it holds. Both adapters have to report them or the
// store learns the fact from one source and not the other, which is worse than
// learning it from neither — the panel would then be half measurable for reasons
// nothing on the screen explains.
func TestTheWTSPopularityListReportsTheSameTwoFacts(t *testing.T) {
	f := &fakePopular{out: popularityOf("005930", "000660")}
	src := WTSPopular(f, 30)
	ctx := context.Background()

	first, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("first Read: %v", err)
	}
	for _, r := range first.Rows {
		if r.NewlyListed.Known() {
			t.Errorf("%s: the first WTS reading answered %s", r.Symbol, r.NewlyListed)
		}
		if r.RankRequested != 30 {
			t.Errorf("%s: RankRequested = %d, want 30", r.Symbol, r.RankRequested)
		}
	}

	f.out = popularityOf("005930", "035720")
	second, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("second Read: %v", err)
	}
	got := newlyListedBySymbol(second.Rows)
	if !got["005930"].No() || !got["035720"].Yes() {
		t.Errorf("the WTS answers on the second reading = %v, want 005930 no and 035720 yes", got)
	}
}

// TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime.
//
// This source falls back to ProductCode when Symbol is empty, so the previous
// reading's set has to be keyed by the same string the rows are. A set built from
// Symbol alone would hold "" for those rows and every fallback row would answer
// `yes` on every single reading — a permanent new-entrant marker produced entirely
// by the indexing.
func TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime(t *testing.T) {
	byProductCode := domain.StockRanking{Stocks: []domain.RankedStock{
		{Rank: 1, ProductCode: "A005930"}, {Rank: 2, ProductCode: "A000660"},
	}}
	f := &fakePopular{out: byProductCode}
	src := WTSPopular(f, 30)
	ctx := context.Background()

	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}
	second, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("second Read: %v", err)
	}
	for _, r := range second.Rows {
		if !r.NewlyListed.No() {
			t.Errorf("%s answered %s on an unchanged list; the previous reading is indexed by "+
				"a different string from the one the rows are reported under",
				r.Symbol, r.NewlyListed)
		}
	}
}
