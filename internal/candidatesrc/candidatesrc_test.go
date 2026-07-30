package candidatesrc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type fakeRankings struct {
	out    domain.Ranking
	err    error
	gotTyp string
	gotMkt string
	gotDur string
	gotN   int
}

func (f *fakeRankings) Rankings(_ context.Context, typ, market, duration string,
	_ bool, count int) (domain.Ranking, error) {
	f.gotTyp, f.gotMkt, f.gotDur, f.gotN = typ, market, duration, count
	return f.out, f.err
}

type fakeBudget struct{ b official.RateBudget }

func (f fakeBudget) RateBudget(string) official.RateBudget { return f.b }

type fakePopular struct {
	out     domain.StockRanking
	err     error
	gotSize int
}

func (f *fakePopular) GetStockRanking(_ context.Context, size int) (domain.StockRanking, error) {
	f.gotSize = size
	return f.out, f.err
}

// TestTheOfficialRankingAsksForTheRealtimeList.
//
// A 1d ranking is a summary of a move that has already finished, and using one
// would make this change an expensive way to reproduce a gainers list. The
// duration is therefore not a caller's option.
func TestTheOfficialRankingAsksForTheRealtimeList(t *testing.T) {
	f := &fakeRankings{out: domain.Ranking{Items: []domain.RankingItem{
		{Rank: 1, Symbol: "005930", TradingAmount: 1e9, TradingVolume: 5000, LastPrice: 70000},
	}}}
	src, mkErr := OfficialRanking(f, nil, RankingTradingAmount, 100, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}

	if _, err := src.Read(context.Background(), candidate.MarketKR); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if f.gotDur != "realtime" {
		t.Errorf("duration = %q, want realtime", f.gotDur)
	}
	if f.gotTyp != RankingTradingAmount || f.gotMkt != candidate.MarketKR {
		t.Errorf("asked for %s/%s", f.gotTyp, f.gotMkt)
	}
}

// TestTheRequestedCountIsCappedAtTheDocumentedMaximum.
//
// The API caps count at 100 and truncates silently. A caller that asked for 150
// and received 100 would compute rank percentiles against a list length that
// never existed — a normalisation error produced by the request rather than by
// the arithmetic.
func TestTheRequestedCountIsCappedAtTheDocumentedMaximum(t *testing.T) {
	f := &fakeRankings{}
	src, mkErr := OfficialRanking(f, nil, RankingTopGainers, 500, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}
	if _, err := src.Read(context.Background(), candidate.MarketKR); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if f.gotN != 100 {
		t.Errorf("count = %d, want 100", f.gotN)
	}
}

// TestTheRankTotalIsTheListWeActuallyReceived is D8's normalisation input. The
// percentile is rank over list length, so the length has to be what came back
// rather than what was requested — a short response is common and a rank of 5
// means something very different in a list of 8 than in a list of 100.
func TestTheRankTotalIsTheListWeActuallyReceived(t *testing.T) {
	f := &fakeRankings{out: domain.Ranking{Items: []domain.RankingItem{
		{Rank: 1, Symbol: "005930"}, {Rank: 2, Symbol: "000660"}, {Rank: 3, Symbol: "035720"},
	}}}
	src, mkErr := OfficialRanking(f, nil, RankingTradingAmount, 100, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}
	got, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	for _, r := range got.Rows {
		if r.RankTotal != 3 {
			t.Errorf("%s: RankTotal = %d, want 3 (the list that arrived, not the 100 requested)",
				r.Symbol, r.RankTotal)
		}
	}
}

// TestARateLimitedRankingIsReportedAsOne.
//
// The scan counts rate-limited losses separately, and for this source that count
// is the measurement: the official RANKING group has no published limit, and
// hybrid gives the endpoint no WTS fallback, so a 429 removes the source rather
// than degrading it. Filing it as an ordinary failure would discard the one
// signal available about a limit nobody documented.
func TestARateLimitedRankingIsReportedAsOne(t *testing.T) {
	f := &fakeRankings{err: fmt.Errorf("get /rankings: %w", official.ErrRateLimited)}
	src, mkErr := OfficialRanking(f, nil, RankingTradingAmount, 100, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}
	_, err := src.Read(context.Background(), candidate.MarketKR)
	if !errors.Is(err, candidate.ErrRateLimited) {
		t.Fatalf("a 429 came back as %v, want candidate.ErrRateLimited", err)
	}
}

// TestTheReportedRateBudgetTravelsWithTheReading is D13 decision 2 end to end:
// the header the official client now keeps has to reach the scan result, or it
// was recorded for nobody.
func TestTheReportedRateBudgetTravelsWithTheReading(t *testing.T) {
	reset := time.Date(2026, 7, 28, 9, 1, 0, 0, time.UTC)
	f := &fakeRankings{out: domain.Ranking{Items: []domain.RankingItem{{Rank: 1, Symbol: "005930"}}}}
	budget := fakeBudget{b: official.RateBudget{
		Limit: 10, Remaining: 4, Reset: reset, Reported: true,
	}}

	src, mkErr := OfficialRanking(f, budget, RankingTradingAmount, 100, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}
	got, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !got.Budget.Reported || got.Budget.Remaining != 4 || got.Budget.Limit != 10 {
		t.Fatalf("budget = %+v, want 4 of 10 reported", got.Budget)
	}
	if !got.Budget.Reset.Equal(reset) {
		t.Errorf("reset = %v, want %v", got.Budget.Reset, reset)
	}
}

// TestAnUnreportedBudgetStaysUnreported is the absent-versus-zero rule crossing
// the package boundary. A client with no accessor and a response with no headers
// must both arrive as "nothing was said", never as "nothing left".
func TestAnUnreportedBudgetStaysUnreported(t *testing.T) {
	f := &fakeRankings{out: domain.Ranking{Items: []domain.RankingItem{{Rank: 1, Symbol: "005930"}}}}

	noAccessorSrc, noAccessorMkErr := OfficialRanking(f, nil, RankingTradingAmount, 100, nil)
	if noAccessorMkErr != nil {
		t.Fatalf("OfficialRanking: %v", noAccessorMkErr)
	}
	noAccessor, err := noAccessorSrc.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if noAccessor.Budget.Reported || noAccessor.Budget.Tight(5) {
		t.Errorf("a client with no budget accessor reported %+v", noAccessor.Budget)
	}

	silentSrc, silentMkErr := OfficialRanking(f, fakeBudget{}, RankingTradingAmount, 100, nil)
	if silentMkErr != nil {
		t.Fatalf("OfficialRanking: %v", silentMkErr)
	}
	silent, err := silentSrc.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if silent.Budget.Reported || silent.Budget.Tight(5) {
		t.Errorf("a response with no rate headers reported %+v", silent.Budget)
	}
}

// TestThePopularityRankingReportsNoTradingFigures pins a fact about the source
// rather than about our code, because section 3 depends on it.
//
// domain.RankedStock is rank, code, symbol, name and market — there is no trading
// value in it at all. So this source can raise a candidate and can never
// contribute to its rate or acceleration, and the fields must come back empty
// rather than as "0". A zero here would be a fabricated data point that a
// per-source rate series would then difference against a real one.
func TestThePopularityRankingReportsNoTradingFigures(t *testing.T) {
	f := &fakePopular{out: domain.StockRanking{Stocks: []domain.RankedStock{
		{Rank: 1, Symbol: "005930", Name: "삼성전자"},
	}}}
	got, err := WTSPopular(f, 30, nil).Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(got.Rows))
	}
	r := got.Rows[0]
	if r.TradingValue != "" || r.TradingVolume != "" || r.Price != "" {
		t.Errorf("the popularity ranking reported figures it does not carry: %+v", r)
	}
	if r.Rank != 1 || r.RankTotal != 1 {
		t.Errorf("rank = %d/%d, want 1/1", r.Rank, r.RankTotal)
	}
}

// TestThePopularityRankingFallsBackToTheProductCode: some WTS rows carry the
// product code and no symbol. Dropping those rows would quietly shrink the list
// — and the list length is the percentile denominator.
func TestThePopularityRankingFallsBackToTheProductCode(t *testing.T) {
	f := &fakePopular{out: domain.StockRanking{Stocks: []domain.RankedStock{
		{Rank: 1, ProductCode: "A005930"},
	}}}
	got, err := WTSPopular(f, 30, nil).Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got.Rows) != 1 || got.Rows[0].Symbol != "A005930" {
		t.Fatalf("rows = %+v, want the product code used as the symbol", got.Rows)
	}
}

// TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking.
//
// This is not tidiness. A scan treats "this source answered and did not list the
// symbol" as evidence the symbol left the list, and starts its cooling clock —
// which ends in expiry, which discards first_seen_at. A source that structurally
// cannot see a market would answer with rows for a different one, so every US
// candidate would be cooled on every scan by a source that was never looking at
// them.
func TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking(t *testing.T) {
	official := &fakeRankings{}
	wts := &fakePopular{}

	kr := Panel(candidate.MarketKR, official, nil, wts, nil)
	us := Panel(candidate.MarketUS, official, nil, wts, nil)

	if !hasSource(kr, candidate.SourceWTSPopular) {
		t.Error("the KR panel is missing the WTS popularity ranking")
	}
	if hasSource(us, candidate.SourceWTSPopular) {
		t.Error("the US panel includes the Korean popularity ranking; it would cool " +
			"every US candidate on every scan")
	}
	if len(us) == 0 {
		t.Fatal("the US panel is empty — the official rankings serve both markets")
	}
}

// TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly.
//
// Spec Requirement 5 says the official sources must suffice alone; the converse
// — WTS alone — is the configuration that stops working the day a session
// expires. A panel built without an official client is allowed to exist, but the
// caller has to be able to see that it did, which the source list makes plain.
func TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly(t *testing.T) {
	only := Panel(candidate.MarketKR, nil, nil, &fakePopular{}, nil)
	if len(only) != 1 || only[0].ID() != candidate.SourceWTSPopular {
		t.Fatalf("panel = %v, want just the WTS source", only)
	}
	if len(Panel(candidate.MarketKR, nil, nil, nil, nil)) != 0 {
		t.Error("a panel with no clients is not empty")
	}
}

func hasSource(panel []candidate.Source, id candidate.SourceID) bool {
	for _, s := range panel {
		if s.ID() == id {
			return true
		}
	}
	return false
}

// TestEveryPanelSourceHasItsOwnID is the section-2 review's P0 at the place that
// produced it.
//
// The three official rankings shipped under one source id. That is not cosmetic:
// the scan may only cool a candidate when every source that raised it answered,
// and the check is keyed by id — so the two rankings that answered vouched for
// the one that was rate limited, and a candidate only ever raised by the missing
// list got cooled by a scan that had never looked at it. From there the cooling
// clock expires it and first_seen_at is gone.
func TestEveryPanelSourceHasItsOwnID(t *testing.T) {
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		panel := Panel(market, &fakeRankings{}, nil, &fakePopular{}, nil)
		seen := map[candidate.SourceID]bool{}
		for _, src := range panel {
			id := src.ID()
			if seen[id] {
				t.Errorf("%s panel has two sources under the id %q", market, id)
			}
			seen[id] = true
		}
		if len(panel) == 0 {
			t.Errorf("%s panel is empty", market)
		}
	}
}

// TestAnUnknownRankingTypeIsRefused. A fallback id would put the new source back
// into a collision with a real one, which is the defect this guard exists for.
func TestAnUnknownRankingTypeIsRefused(t *testing.T) {
	if _, err := OfficialRanking(&fakeRankings{}, nil, "MARKET_SOMETHING_NEW", 100, nil); err == nil {
		t.Fatal("an unknown ranking type was accepted and given some source id")
	}
}

// TestAnInfiniteTradingValueIsAbsentRatherThanInfinite.
//
// internal/official's parseDecimal returns whatever ParseFloat produced when err
// is nil, and ParseFloat accepts "NaN", "Inf" and "Infinity" without complaint.
// Formatted naively those round-trip straight back, and an infinite trading value
// clears every threshold it is compared against — a manufactured maximum signal,
// which is the failure the observation validator already refuses for ranks.
func TestAnInfiniteTradingValueIsAbsentRatherThanInfinite(t *testing.T) {
	f := &fakeRankings{out: domain.Ranking{Items: []domain.RankingItem{
		{Rank: 1, Symbol: "005930",
			TradingAmount: math.Inf(1), TradingVolume: math.NaN(), LastPrice: 70000},
	}}}
	src, mkErr := OfficialRanking(f, nil, RankingTradingAmount, 100, nil)
	if mkErr != nil {
		t.Fatalf("OfficialRanking: %v", mkErr)
	}
	got, err := src.Read(context.Background(), candidate.MarketKR)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	r := got.Rows[0]
	if r.TradingValue != "" {
		t.Errorf("an infinite trading value came through as %q", r.TradingValue)
	}
	if r.TradingVolume != "" {
		t.Errorf("a NaN trading volume came through as %q", r.TradingVolume)
	}
	if r.Price != "70000" {
		t.Errorf("a finite price was damaged: %q", r.Price)
	}
}

// TestThePopularityRankingRefusesAMarketItCannotSee puts the guard on the source
// rather than only on Panel. A caller building a panel by hand, or reusing a KR
// panel for a US scan, would otherwise file Korean rows under US — and a scan
// treats a responding source as evidence about the market it was asked for.
func TestThePopularityRankingRefusesAMarketItCannotSee(t *testing.T) {
	f := &fakePopular{out: domain.StockRanking{Stocks: []domain.RankedStock{
		{Rank: 1, Symbol: "005930"},
	}}}
	if _, err := WTSPopular(f, 30, nil).Read(context.Background(), candidate.MarketUS); err == nil {
		t.Fatal("the Korean popularity ranking answered a US scan")
	}
	if f.gotSize != 0 {
		t.Error("it called the client before refusing")
	}
}

// TestTheRankingsPathIsTheOneTheClientRequests. The budget map is keyed by path,
// so a drifted copy makes the lookup miss — and a miss reads as "no headers were
// sent", which is indistinguishable from a server that never sends them. The
// measurement would go dark without failing.
func TestTheRankingsPathIsTheOneTheClientRequests(t *testing.T) {
	if rankingsPath != official.PathRankings {
		t.Fatalf("rankingsPath = %q, official.PathRankings = %q", rankingsPath, official.PathRankings)
	}
}
