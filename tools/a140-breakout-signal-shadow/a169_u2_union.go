package main

import (
	"errors"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

type a169U2UnionRow struct {
	Symbol                                         string
	Amount, Volume                                 domain.RankingItem
	HasAmount, HasVolume                           bool
	AmountReason, VolumeReason, VSource, Exclusion string
	Score                                          a169U2Result
	Selected                                       bool
}
type a169U2UnionResult struct {
	Amount, Volume a169Selection
	Rows           []a169U2UnionRow
	Symbols        []string
	SelectedAt     time.Time
	Short          bool
}

// a169U2Union은 정렬 영수증의 통과 행 합집합에 순수 점수를 적용한다.
func a169U2Union(amount, volume domain.Ranking, amountCount, volumeCount int, amountCaution, volumeCaution bool, cap float64, n int, day, open, close time.Time, daily map[string][]domain.Candle) (a169U2UnionResult, error) {
	for _, at := range []time.Time{amount.FetchedAt, volume.FetchedAt} {
		_, err := a169U2Score(0, nil, day, open, close, at)
		if err == nil || err.Error() != "history_lt_20" {
			return a169U2UnionResult{}, errors.New("envelope_invalid")
		}
	}
	a, err := a169Select("U0", amount, amountCount, amountCaution, cap, n)
	if err != nil {
		return a169U2UnionResult{}, err
	}
	v, err := a169Select("U1", volume, volumeCount, volumeCaution, cap, n)
	if err != nil {
		return a169U2UnionResult{}, err
	}
	result := a169U2UnionResult{Amount: a, Volume: v, SelectedAt: amount.FetchedAt}
	if volume.FetchedAt.After(result.SelectedAt) {
		result.SelectedAt = volume.FetchedAt
	}
	bySymbol := map[string]*a169U2UnionRow{}
	eligible := map[string]bool{}
	for _, row := range a.Rows {
		r := bySymbol[row.Item.Symbol]
		if r == nil {
			r = &a169U2UnionRow{Symbol: row.Item.Symbol}
			bySymbol[row.Item.Symbol] = r
		}
		r.Amount = row.Item
		r.HasAmount = true
		r.AmountReason = row.Reason
		if row.Reason == "" {
			eligible[row.Item.Symbol] = true
		}
	}
	for _, row := range v.Rows {
		r := bySymbol[row.Item.Symbol]
		if r == nil {
			r = &a169U2UnionRow{Symbol: row.Item.Symbol}
			bySymbol[row.Item.Symbol] = r
		}
		r.Volume = row.Item
		r.HasVolume = true
		r.VolumeReason = row.Reason
		if row.Reason == "" {
			eligible[row.Item.Symbol] = true
		}
	}
	for symbol := range eligible {
		r := bySymbol[symbol]
		raw := r.Amount.TradingVolume
		r.VSource = "amount"
		if r.HasVolume {
			raw = r.Volume.TradingVolume
			r.VSource = "volume"
		}
		if !a169U2Volume(raw) {
			r.Exclusion = "numerator_invalid"
		} else if bars, ok := daily[symbol]; !ok {
			r.Exclusion = "baseline_unavailable"
		} else {
			r.Score, err = a169U2Score(raw, bars, day, open, close, result.SelectedAt)
			if err != nil {
				r.Exclusion = err.Error()
			}
		}
		result.Rows = append(result.Rows, *r)
	}
	sort.Slice(result.Rows, func(i, j int) bool { return result.Rows[i].Symbol < result.Rows[j].Symbol })
	order := make([]int, 0, len(result.Rows))
	for i := range result.Rows {
		if result.Rows[i].Exclusion == "" {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		x, y := result.Rows[order[i]], result.Rows[order[j]]
		if x.Score.RatioPPM != y.Score.RatioPPM {
			return x.Score.RatioPPM > y.Score.RatioPPM
		}
		if x.HasVolume != y.HasVolume {
			return x.HasVolume
		}
		if x.HasVolume && x.Volume.Rank != y.Volume.Rank {
			return x.Volume.Rank < y.Volume.Rank
		}
		if x.HasAmount != y.HasAmount {
			return x.HasAmount
		}
		if x.HasAmount && x.Amount.Rank != y.Amount.Rank {
			return x.Amount.Rank < y.Amount.Rank
		}
		return x.Symbol < y.Symbol
	})
	for _, i := range order {
		if len(result.Symbols) == n {
			break
		}
		result.Rows[i].Selected = true
		result.Symbols = append(result.Symbols, result.Rows[i].Symbol)
	}
	result.Short = len(result.Symbols) < n
	return result, nil
}
