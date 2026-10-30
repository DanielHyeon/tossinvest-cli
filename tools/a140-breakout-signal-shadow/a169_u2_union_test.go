package main

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func a169U2UnionFixture() (domain.Ranking, domain.Ranking, time.Time, time.Time, time.Time, map[string][]domain.Candle) {
	day, open, close, bars := a169U2Fixture()
	a := domain.Ranking{Type: "MARKET_TRADING_AMOUNT", MarketCountry: "KR", Duration: "realtime", FetchedAt: open.Add(time.Minute)}
	v := domain.Ranking{Type: "MARKET_TRADING_VOLUME", MarketCountry: "KR", Duration: "realtime", FetchedAt: open.Add(2 * time.Minute)}
	a.Items = []domain.RankingItem{{Rank: 1, Symbol: "000001", Currency: "KRW", LastPrice: 1500, TradingVolume: 7}, {Rank: 2, Symbol: "000002", Currency: "KRW", LastPrice: 1500, TradingVolume: 9}}
	v.Items = []domain.RankingItem{{Rank: 1, Symbol: "000002", Currency: "USD", LastPrice: 1500, TradingVolume: 5}, {Rank: 2, Symbol: "000001", Currency: "KRW", LastPrice: 1500, TradingVolume: 1}}
	return a, v, day, open, close, map[string][]domain.Candle{"000001": bars, "000002": bars}
}
func TestA169U2UnionAnyPassAndVolume(t *testing.T) {
	a, v, day, open, close, d := a169U2UnionFixture()
	r, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Rows) != 2 || len(r.Symbols) != 2 || r.Symbols[0] != "000002" || r.Rows[1].VSource != "volume" || r.Rows[1].VolumeReason != "currency" || r.Rows[1].Score.V != 5 || !r.SelectedAt.Equal(v.FetchedAt) || !r.Short {
		t.Fatalf("unexpected %+v", r)
	}
	a.Items[0].Symbol = "changed"
	if r.Amount.Receipt.Items[0].Symbol != "000001" {
		t.Fatal("input alias")
	}
	v.Items[0].TradingVolume = math.NaN()
	delete(d, "000002")
	r, e = a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil {
		t.Fatal(e)
	}
	for _, row := range r.Rows {
		if row.Symbol == "000002" && row.Exclusion != "numerator_invalid" {
			t.Fatalf("priority %+v", row)
		}
	}
}

func TestA169U2UnionFullRowsAndShort(t *testing.T) {
	a, v, day, open, close, d := a169U2UnionFixture()
	a.Items = nil
	v.Items = nil
	d = map[string][]domain.Candle{}
	_, _, _, bars := a169U2Fixture()
	for i := 1; i <= 35; i++ {
		symbol := fmt.Sprintf("%06d", i)
		a.Items = append(a.Items, domain.RankingItem{Rank: i, Symbol: symbol, Currency: "KRW", LastPrice: 1500, TradingVolume: 1})
		d[symbol] = bars
	}
	for i := 36; i <= 70; i++ {
		symbol := fmt.Sprintf("%06d", i)
		v.Items = append(v.Items, domain.RankingItem{Rank: i - 35, Symbol: symbol, Currency: "KRW", LastPrice: 1500, TradingVolume: 1})
		d[symbol] = bars
	}
	a.Items[34].TradingVolume = 100
	v.Items[34].TradingVolume = 200
	for _, n := range []int{20, 30} {
		r, e := a169U2Union(a, v, 100, 100, true, true, 10000, n, day, open, close, d)
		if e != nil || len(r.Rows) != 70 || len(r.Symbols) != n || r.Short || r.Symbols[0] != "000070" || r.Symbols[1] != "000035" {
			t.Fatalf("n=%d %+v %v", n, r.Symbols, e)
		}
	}
	a.Items = nil
	v.Items = nil
	r, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil || !r.Short || len(r.Rows) != 0 || len(r.Symbols) != 0 {
		t.Fatalf("empty %+v %v", r, e)
	}
	v.FetchedAt = open
	if _, e = a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d); e == nil || e.Error() != "envelope_invalid" {
		t.Fatalf("empty bad time %v", e)
	}
}
func TestA169U2UnionMissingTimeTieAndCopy(t *testing.T) {
	a, v, day, open, close, d := a169U2UnionFixture()
	a.RankedAt = "later-looking"
	v.RankedAt = "earlier-looking"
	a.FetchedAt = open.Add(3 * time.Minute)
	v.FetchedAt = open.Add(2 * time.Minute)
	v.Items[0].Currency = "KRW"
	v.Items[0].TradingVolume = 1
	v.Items[1].TradingVolume = 1
	r, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil || !r.SelectedAt.Equal(a.FetchedAt) || r.Symbols[0] != "000002" {
		t.Fatalf("time/tie %+v %v", r, e)
	}
	r.Amount.Receipt.Items[0].Symbol = "result-mutated"
	if a.Items[0].Symbol != "000001" {
		t.Fatal("reverse alias")
	}
	delete(d, "000001")
	d["000002"] = []domain.Candle{}
	r, e = a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil || r.Rows[0].Exclusion != "baseline_unavailable" || r.Rows[1].Exclusion != "history_lt_20" || len(r.Symbols) != 0 {
		t.Fatalf("missing vs empty %+v %v", r.Rows, e)
	}
	v.FetchedAt = day.AddDate(0, 0, -1).Add(10 * time.Hour)
	if _, e = a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d); e == nil || e.Error() != "envelope_invalid" {
		t.Fatalf("other day %v", e)
	}
}

func TestA169U2UnionTieAndDeterminism(t *testing.T) {
	a, v, day, open, close, d := a169U2UnionFixture()
	a.Items = []domain.RankingItem{{Rank: 1, Symbol: "000001", Currency: "KRW", LastPrice: 1500, TradingVolume: 1}, {Rank: 2, Symbol: "000002", Currency: "KRW", LastPrice: 1500, TradingVolume: 1}}
	v.Items = []domain.RankingItem{{Rank: 3, Symbol: "000003", Currency: "KRW", LastPrice: 1500, TradingVolume: 1}}
	d["000003"] = d["000001"]
	r, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil || !reflect.DeepEqual(r.Symbols, []string{"000003", "000001", "000002"}) {
		t.Fatalf("ties %+v %v", r.Symbols, e)
	}
	reordered := map[string][]domain.Candle{}
	reordered["000003"] = d["000003"]
	reordered["000002"] = d["000002"]
	reordered["000001"] = d["000001"]
	other, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, reordered)
	if e != nil || !reflect.DeepEqual(r, other) {
		t.Fatalf("nondeterministic %+v %v", other, e)
	}
	bad := append([]domain.Candle(nil), d["000003"]...)
	bad[0].Volume = math.NaN()
	reordered["000003"] = bad
	r, e = a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, reordered)
	if e != nil || r.Rows[2].Exclusion != "baseline_invalid" {
		t.Fatalf("score failure %+v %v", r.Rows, e)
	}
}

func TestA169U2UnionZeroAndInvalidReceipt(t *testing.T) {
	a, v, day, open, close, d := a169U2UnionFixture()
	a.Items = a.Items[:1]
	v.Items = nil
	a.Items[0].TradingVolume = 0
	r, e := a169U2Union(a, v, 100, 100, true, true, 10000, 20, day, open, close, d)
	if e != nil || len(r.Symbols) != 1 || !r.Rows[0].Score.ZeroV || r.Rows[0].Score.RatioPPM != 0 || r.Rows[0].Exclusion != "" {
		t.Fatalf("zero %+v %v", r, e)
	}
	if _, e = a169U2Union(a, v, 99, 100, true, true, 10000, 20, day, open, close, d); e == nil {
		t.Fatal("invalid count accepted")
	}
}
