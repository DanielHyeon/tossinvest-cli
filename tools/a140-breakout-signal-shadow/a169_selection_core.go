package main

import (
	"errors"
	"math"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

type a169SelectionRow struct {
	Item     domain.RankingItem
	Reason   string
	Selected bool
}

type a169Selection struct {
	Receipt        domain.Ranking
	RequestedCount int
	ExcludeCaution bool
	Rows           []a169SelectionRow
	Symbols        []string
	ActualCount    int
	Short          bool
}

type a169SecurityType struct {
	Raw   string
	Group string
}

// a169Select는 검증된 영수증 주장만 받아 순위를 바꾸지 않고 모든 후보를 판정함.
func a169Select(cohort string, ranking domain.Ranking, count int, caution bool, cap float64, n int) (a169Selection, error) {
	expected := ""
	switch cohort {
	case "U0":
		expected = "MARKET_TRADING_AMOUNT"
	case "U1":
		expected = "MARKET_TRADING_VOLUME"
	default:
		return a169Selection{}, errors.New("invalid cohort")
	}
	if ranking.Type != expected || ranking.MarketCountry != "KR" || ranking.Duration != "realtime" || ranking.FetchedAt.IsZero() || count != 100 || !caution || len(ranking.Items) > 100 || math.IsNaN(cap) || math.IsInf(cap, 0) || cap < 1000 || (n != 20 && n != 30) {
		return a169Selection{}, errors.New("invalid selection receipt or policy")
	}
	seen := make(map[string]bool, len(ranking.Items))
	lastRank := 0
	for _, item := range ranking.Items {
		if item.Rank <= lastRank || seen[item.Symbol] {
			return a169Selection{}, errors.New("invalid rank or duplicate symbol")
		}
		lastRank = item.Rank
		seen[item.Symbol] = true
	}
	result := a169Selection{Receipt: ranking, RequestedCount: count, ExcludeCaution: caution, Rows: make([]a169SelectionRow, 0, len(ranking.Items)), Symbols: make([]string, 0, n)}
	if ranking.Items != nil {
		result.Receipt.Items = append(make([]domain.RankingItem, 0, len(ranking.Items)), ranking.Items...)
	}
	for _, item := range ranking.Items {
		row := a169SelectionRow{Item: item}
		switch {
		case !a169NumericSymbol(item.Symbol):
			row.Reason = "symbol_format"
		case item.Currency != "KRW":
			row.Reason = "currency"
		case math.IsNaN(item.LastPrice) || math.IsInf(item.LastPrice, 0) || item.LastPrice <= 0:
			row.Reason = "price_invalid"
		case item.LastPrice < 1000:
			row.Reason = "price_below_floor"
		case item.LastPrice > cap:
			row.Reason = "price_above_cap"
		case len(result.Symbols) < n:
			row.Selected = true
			result.Symbols = append(result.Symbols, item.Symbol)
		}
		result.Rows = append(result.Rows, row)
	}
	result.ActualCount = len(result.Symbols)
	result.Short = result.ActualCount < n
	return result, nil
}

func a169NumericSymbol(symbol string) bool {
	if len(symbol) != 6 {
		return false
	}
	for i := 0; i < len(symbol); i++ {
		if symbol[i] < '0' || symbol[i] > '9' {
			return false
		}
	}
	return true
}

func a169ClassifyType(raw string) a169SecurityType {
	group := "미분류"
	switch raw {
	case "STOCK", "FOREIGN_STOCK":
		group = "STOCK"
	case "ETF", "FOREIGN_ETF":
		group = "ETF"
	case "ETN":
		group = "ETN"
	case "DEPOSITARY_RECEIPT", "INFRASTRUCTURE_FUND", "REIT", "STOCK_WARRANTS":
		group = "기타"
	}
	return a169SecurityType{Raw: raw, Group: group}
}
