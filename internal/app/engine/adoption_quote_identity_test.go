package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

type quoteIdentityPrices struct {
	quotes []domain.Quote
	asked  []string
}

func (f *quoteIdentityPrices) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	f.asked = append([]string(nil), symbols...)
	return append([]domain.Quote(nil), f.quotes...), nil
}

func TestObserveCandidatesRequiresMarketCurrencyIdentity(t *testing.T) {
	tests := []struct {
		name     string
		market   string
		currency string
		want     string
	}{
		{name: "US USD", market: "us", currency: "USD", want: "200"},
		{name: "US KRW", market: "us", currency: "KRW"},
		{name: "US empty", market: "us", currency: ""},
		{name: "KR KRW", market: "kr", currency: "KRW", want: "200"},
		{name: "KR USD", market: "kr", currency: "USD"},
		{name: "unknown market", market: "jp", currency: "JPY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := &quoteIdentityPrices{quotes: []domain.Quote{{Symbol: "AAPL", Last: 200, Currency: tt.currency}}}
			driver := &ReconcileDriver{opts: ReconcileDriverOptions{Prices: prices},
				clk: clock.NewFake(time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC))}
			got, _, err := driver.observeCandidates(context.Background(), []candidate{{
				position: journal.Position{Market: tt.market, Symbol: "AAPL"},
			}})
			if err != nil {
				t.Fatal(err)
			}
			if observed := got[adoptionQuoteKey(tt.market, "AAPL")]; observed != tt.want {
				t.Fatalf("observed=%q map=%v want=%q", observed, got, tt.want)
			}
		})
	}
}

func TestObserveCandidatesRefusesCrossMarketDuplicateSymbol(t *testing.T) {
	prices := &quoteIdentityPrices{quotes: []domain.Quote{{Symbol: "DUPL", Last: 200, Currency: "USD"}}}
	driver := &ReconcileDriver{opts: ReconcileDriverOptions{Prices: prices}, clk: clock.System()}
	got, _, err := driver.observeCandidates(context.Background(), []candidate{
		{position: journal.Position{Market: "kr", Symbol: "DUPL"}},
		{position: journal.Position{Market: "us", Symbol: "DUPL"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || len(prices.asked) != 1 || prices.asked[0] != "DUPL" {
		t.Fatalf("observations=%v asked=%v", got, prices.asked)
	}
}
