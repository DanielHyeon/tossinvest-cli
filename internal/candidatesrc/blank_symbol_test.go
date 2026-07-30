package candidatesrc

// blank_symbol_test.go is the fixture issues.md I15 recorded as missing: a reading
// that carries rows with no symbol.
//
// # Why the gap mattered
//
// I15 said the row set and the remembered set "are filtered on the same condition",
// and they are — both drop a blank symbol. What was not filtered on that condition
// was the decision that a reading arrived *whole*. That comparison was made by the
// caller against len(items), which counts the rows both sets then throw away, so a
// reading could be judged complete on rows the memory never held.
//
// Two shapes came out of it, both of them ending in a measured `yes` on a symbol
// that had not gone anywhere:
//
//	some blank   the memory keeps fewer symbols than the request, and the next whole
//	             reading reports the difference as new entrants.
//	all blank    the reading has no rows at all — the scan files it under Missing —
//	             and the memory becomes an empty non-nil set, which usableAt accepts.
//	             Every row of the next reading is then absent from it.
//
// Both are permanent once they land: candidates.first_rank_newly_listed is
// write-once, so the answer that came out of a blank row answers for the rest of
// that candidate's life.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// rankingWithBlanks is a ranking whose rows arrive in the order given, with "" for
// a row the endpoint sent without a symbol. The blanks are counted in Items — that
// is the point — and dropped from the rows and from the memory.
func rankingWithBlanks(symbols ...string) domain.Ranking {
	out := domain.Ranking{}
	for i, s := range symbols {
		out.Items = append(out.Items, domain.RankingItem{Rank: i + 1, Symbol: s})
	}
	return out
}

// TestAReadingThatLostRowsToBlankSymbolsIsNotAWholeReading is the partial case.
//
// Three rows of the three requested, two of them blank. Under the caller-side
// comparison this was whole, so it replaced a three-symbol memory with a one-symbol
// one, and the next complete reading answered `yes` for the other two.
func TestAReadingThatLostRowsToBlankSymbolsIsNotAWholeReading(t *testing.T) {
	clk := clock.NewFake(t0())
	whole := rankingOf("005930", "000660", "035720")
	f := &fakeRankings{out: whole}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 3, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// The degraded reading: three rows arrive, two of them without a symbol.
	clk.Advance(15 * time.Second)
	f.out = rankingWithBlanks("005930", "", "  ")
	blank, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the blank Read: %v", err)
	}
	if len(blank.Rows) != 1 {
		t.Fatalf("the blank reading carried %d rows, want 1; the fixture is not the shape this "+
			"test is about", len(blank.Rows))
	}

	// And the whole one behind it.
	clk.Advance(15 * time.Second)
	f.out = whole
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the whole Read after the blank one: %v", err)
	}
	if len(back.Rows) != 3 {
		t.Fatalf("the whole reading carried %d rows, want 3", len(back.Rows))
	}
	for _, r := range back.Rows {
		if !r.NewlyListed.No() {
			t.Errorf("%s answered %s on a list it never left. The reading in between arrived "+
				"with three rows of the three requested and two of them blank, so the set it "+
				"would have become holds one symbol — and the two it lost come back as new "+
				"entrants, measured and written to a write-once column", r.Symbol, r.NewlyListed)
		}
	}
}

// TestAReadingOfNothingButBlankSymbolsDoesNotBecomeThePreviousReading is the all-blank
// case, and it is the worse one.
//
// The reading has no rows, so candidate.Collect files it under Missing and it vouches
// for nothing. The memory, though, was replaced by an empty *non-nil* set — and
// usableAt's test is `symbols == nil`, so the empty set answers. Every row of the
// next reading is absent from it, so every row is a new entrant.
//
// The source here has no earlier memory at all, which is the honest starting state: if
// the all-blank reading is allowed to become one, `unknown` turns into `yes` for the
// whole panel on the first reading that carries rows.
func TestAReadingOfNothingButBlankSymbolsDoesNotBecomeThePreviousReading(t *testing.T) {
	clk := clock.NewFake(t0())
	f := &fakeRankings{out: rankingWithBlanks("", " ", "")}
	src, err := OfficialRanking(f, nil, RankingTradingAmount, 3, clk)
	if err != nil {
		t.Fatalf("OfficialRanking: %v", err)
	}
	ctx := context.Background()
	nothing, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the all-blank Read: %v", err)
	}
	if len(nothing.Rows) != 0 {
		t.Fatalf("the all-blank reading carried %d rows, want none", len(nothing.Rows))
	}

	clk.Advance(15 * time.Second)
	f.out = rankingOf("005930", "000660", "035720")
	first, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the Read after it: %v", err)
	}
	if len(first.Rows) != 3 {
		t.Fatalf("the reading after the blank one carried %d rows, want 3", len(first.Rows))
	}
	if got := knownAnswers(first.Rows); got != 0 {
		t.Errorf("%d of %d rows answered against a reading that carried no symbols at all. An "+
			"empty set is evidence that everything here is new; a reading whose rows were all "+
			"blank is evidence of nothing, and the difference is the whole of the three-state "+
			"contract", got, len(first.Rows))
	}
}

// TestAGenuinelyNewSymbolIsStillFoundAfterABlankReading is the control, so that the
// two above are a narrowing rather than a source that has stopped answering.
func TestAGenuinelyNewSymbolIsStillFoundAfterABlankReading(t *testing.T) {
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
	f.out = rankingWithBlanks("005930", "")
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("blank Read: %v", err)
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
			"whole, so the comparison is available", got["068270"])
	}
	if !got["005930"].No() {
		t.Errorf("005930 answered %s; it has been in every reading", got["005930"])
	}
}

// TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps.
//
// A stock with neither a Symbol nor a ProductCode is dropped by wtsSymbol, and this
// adapter counted it towards the requested size for exactly as long as the other one
// did. Both adapters or neither: a fact learned from one source and not the other
// leaves the panel half measurable for a reason nothing on the screen explains.
func TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps(t *testing.T) {
	clk := clock.NewFake(t0())
	whole := popularityOf("005930", "000660")
	f := &fakePopular{out: whole}
	src := WTSPopular(f, 2, clk)
	ctx := context.Background()
	if _, err := src.Read(ctx, candidate.MarketKR); err != nil {
		t.Fatalf("first Read: %v", err)
	}

	// Two rows of the two requested, one of them carrying no identity at all.
	clk.Advance(5 * time.Second)
	f.out = domain.StockRanking{Stocks: []domain.RankedStock{
		{Rank: 1, Symbol: "005930"}, {Rank: 2},
	}}
	blank, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the blank Read: %v", err)
	}
	if len(blank.Rows) != 1 {
		t.Fatalf("the blank WTS reading carried %d rows, want 1", len(blank.Rows))
	}

	clk.Advance(5 * time.Second)
	f.out = whole
	back, err := src.Read(ctx, candidate.MarketKR)
	if err != nil {
		t.Fatalf("the whole Read after it: %v", err)
	}
	if len(back.Rows) != 2 {
		t.Fatalf("the whole WTS reading carried %d rows, want 2", len(back.Rows))
	}
	got := newlyListedBySymbol(back.Rows)
	if !got["000660"].No() {
		t.Errorf("000660 answered %s after a reading whose second row had no identity; the WTS "+
			"adapter drops that row from its memory too, so counting it towards the size makes "+
			"a short reading look whole", got["000660"])
	}
}
