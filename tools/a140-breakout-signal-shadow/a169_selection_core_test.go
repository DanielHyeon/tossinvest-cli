package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func a169Fixture(kind string) domain.Ranking {
	return domain.Ranking{Type: kind, MarketCountry: "KR", Duration: "realtime", FetchedAt: time.Date(2026, 10, 23, 0, 0, 30, 0, time.UTC)}
}

func TestA169SelectionFiltersAndReceipt(t *testing.T) {
	r := a169Fixture("MARKET_TRADING_AMOUNT")
	r.RankedAt = "raw-clock"
	r.Items = []domain.RankingItem{
		{Rank: 1, Symbol: "0101N0", Currency: "USD", LastPrice: 0},
		{Rank: 2, Symbol: "000002", Currency: "USD", LastPrice: 0},
		{Rank: 3, Symbol: "000003", Currency: "KRW", LastPrice: math.NaN()},
		{Rank: 4, Symbol: "000004", Currency: "KRW", LastPrice: 999},
		{Rank: 5, Symbol: "000005", Currency: "KRW", LastPrice: 10001},
		{Rank: 6, Symbol: "000006", Currency: "KRW", LastPrice: 1000},
		{Rank: 7, Symbol: "000007", Currency: "KRW", LastPrice: 10000},
	}
	got, err := a169Select("U0", r, 100, true, 10000, 20)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"symbol_format", "currency", "price_invalid", "price_below_floor", "price_above_cap", "", ""}
	if len(got.Rows) != len(want) || len(got.Symbols) != 2 || got.ActualCount != 2 || !got.Short || got.Receipt.RankedAt != "raw-clock" {
		t.Fatalf("result %+v", got)
	}
	for i, w := range want {
		if got.Rows[i].Reason != w || got.Rows[i].Selected != (i >= 5) {
			t.Fatalf("row %d %+v", i, got.Rows[i])
		}
	}
	if got.Symbols[0] != "000006" || got.Symbols[1] != "000007" {
		t.Fatalf("symbols %v", got.Symbols)
	}
	r.Items[5].Symbol = "mutated"
	if got.Receipt.Items[5].Symbol != "000006" {
		t.Fatal("input alias")
	}
	got.Receipt.Items[6].Symbol = "changed"
	if r.Items[6].Symbol != "000007" {
		t.Fatal("result alias")
	}
}

func TestA169SelectionMetadataAndSize(t *testing.T) {
	for _, c := range []struct {
		cohort, kind string
		n            int
		want         int
	}{{"U0", "MARKET_TRADING_AMOUNT", 20, 20}, {"U1", "MARKET_TRADING_VOLUME", 30, 30}} {
		r := a169Fixture(c.kind)
		for i := 0; i < 100; i++ {
			r.Items = append(r.Items, domain.RankingItem{Rank: i + 1, Symbol: fmt.Sprintf("%06d", i+1), Currency: "KRW", LastPrice: 1000})
		}
		got, err := a169Select(c.cohort, r, 100, true, 10000, c.n)
		if err != nil || len(got.Symbols) != c.want || got.ActualCount != c.want || len(got.Rows) != 100 || got.Short {
			t.Fatalf("%s result=%+v err=%v", c.cohort, got, err)
		}
		if got.Rows[c.n].Selected || got.Rows[c.n].Reason != "" {
			t.Fatal("valid overflow marked excluded")
		}
	}
	empty, err := a169Select("U0", a169Fixture("MARKET_TRADING_AMOUNT"), 100, true, 10000, 30)
	if err != nil || !empty.Short || empty.ActualCount != 0 || len(empty.Symbols) != 0 {
		t.Fatalf("empty %+v %v", empty, err)
	}
	if empty.Receipt.Items != nil {
		t.Fatal("nil receipt slice changed")
	}
	nonnil := a169Fixture("MARKET_TRADING_AMOUNT")
	nonnil.Items = []domain.RankingItem{}
	empty, err = a169Select("U0", nonnil, 100, true, 10000, 20)
	if err != nil || empty.Receipt.Items == nil {
		t.Fatal("empty receipt slice changed")
	}
}

func TestA169SelectionRejectsInvalidMetadata(t *testing.T) {
	good := a169Fixture("MARKET_TRADING_AMOUNT")
	cases := []struct {
		name    string
		cohort  string
		r       domain.Ranking
		count   int
		caution bool
		cap     float64
		n       int
	}{
		{"cohort", "bad", good, 100, true, 10000, 20}, {"type", "U1", good, 100, true, 10000, 20}, {"count", "U0", good, 99, true, 10000, 20}, {"caution", "U0", good, 100, false, 10000, 20}, {"n", "U0", good, 100, true, 10000, 21}, {"capnan", "U0", good, 100, true, math.NaN(), 20}, {"capinf", "U0", good, 100, true, math.Inf(1), 20}, {"caplow", "U0", good, 100, true, 999, 20},
	}
	bad := good
	bad.MarketCountry = "US"
	cases = append(cases, struct {
		name    string
		cohort  string
		r       domain.Ranking
		count   int
		caution bool
		cap     float64
		n       int
	}{"market", "U0", bad, 100, true, 10000, 20})
	bad = good
	bad.Duration = "daily"
	cases = append(cases, struct {
		name    string
		cohort  string
		r       domain.Ranking
		count   int
		caution bool
		cap     float64
		n       int
	}{"duration", "U0", bad, 100, true, 10000, 20})
	bad = good
	bad.FetchedAt = time.Time{}
	cases = append(cases, struct {
		name    string
		cohort  string
		r       domain.Ranking
		count   int
		caution bool
		cap     float64
		n       int
	}{"time", "U0", bad, 100, true, 10000, 20})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := a169Select(c.cohort, c.r, c.count, c.caution, c.cap, c.n); err == nil {
				t.Fatal("accepted invalid metadata")
			}
		})
	}
}

func TestA169SecurityTypeExact(t *testing.T) {
	cases := map[string]string{"STOCK": "STOCK", "FOREIGN_STOCK": "STOCK", "ETF": "ETF", "FOREIGN_ETF": "ETF", "ETN": "ETN", "DEPOSITARY_RECEIPT": "기타", "INFRASTRUCTURE_FUND": "기타", "REIT": "기타", "STOCK_WARRANTS": "기타", "DR": "미분류", "WARRANTS": "미분류", "": "미분류", " ETF ": "미분류", "etf": "미분류"}
	for raw, want := range cases {
		got := a169ClassifyType(raw)
		if got.Raw != raw || got.Group != want {
			t.Fatalf("raw %q got %+v", raw, got)
		}
	}
}

func TestA169SelectionRejectsRanksDuplicatesAndOver100(t *testing.T) {
	base := a169Fixture("MARKET_TRADING_AMOUNT")
	row := domain.RankingItem{Rank: 1, Symbol: "000001", Currency: "KRW", LastPrice: 1000}
	cases := map[string][]domain.RankingItem{"zero-rank": {{Rank: 0, Symbol: "000001"}}, "rank-tie": {row, {Rank: 1, Symbol: "000002"}}, "rank-backward": {{Rank: 2, Symbol: "000001"}, {Rank: 1, Symbol: "000002"}}, "duplicate-symbol": {row, {Rank: 2, Symbol: "000001"}}}
	for name, items := range cases {
		t.Run(name, func(t *testing.T) {
			r := base
			r.Items = items
			if _, err := a169Select("U0", r, 100, true, 10000, 20); err == nil {
				t.Fatal("invalid sequence accepted")
			}
		})
	}
	for i := 0; i < 101; i++ {
		base.Items = append(base.Items, domain.RankingItem{Rank: i + 1, Symbol: fmt.Sprintf("%06d", i+1)})
	}
	if _, err := a169Select("U0", base, 100, true, 10000, 20); err == nil {
		t.Fatal("over 100 accepted")
	}
}

func TestA169SelectionNonFinitePrices(t *testing.T) {
	r := a169Fixture("MARKET_TRADING_VOLUME")
	r.Items = []domain.RankingItem{{Rank: 1, Symbol: "000001", Currency: "KRW", LastPrice: math.Inf(1)}, {Rank: 2, Symbol: "000002", Currency: "KRW", LastPrice: math.Inf(-1)}, {Rank: 3, Symbol: "000003", Currency: "KRW", LastPrice: 0}}
	got, err := a169Select("U1", r, 100, true, 10000, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range got.Rows {
		if row.Reason != "price_invalid" {
			t.Fatalf("reason %+v", row)
		}
	}
}

func TestA169SecurityTypeOpenAPIEnum(t *testing.T) {
	data, err := os.ReadFile("../../docs/migration/openapi.latest.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	got := doc.Components.Schemas["StockInfo"].Properties["securityType"].Enum
	want := []string{"STOCK", "FOREIGN_STOCK", "DEPOSITARY_RECEIPT", "INFRASTRUCTURE_FUND", "REIT", "ETF", "FOREIGN_ETF", "ETN", "STOCK_WARRANTS"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("security enum = %v", got)
	}
	for _, value := range got {
		if a169ClassifyType(value).Group == "미분류" {
			t.Fatalf("unclassified enum %q", value)
		}
	}
}
