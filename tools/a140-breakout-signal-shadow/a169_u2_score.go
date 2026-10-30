package main

import (
	"errors"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

type a169U2Result struct {
	V, S20, RatioPPM, IncreasePPM int64
	PriorDays, ZeroVolumeDays     int
	ZeroV                         bool
	PriorBarAt, NextBarAt         time.Time
}

func a169U2Volume(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 9007199254740991 && math.Trunc(v) == v
}

// a169U2Score는 1개 후보의 순수 점수만 계산한다. 실패 시 결과는 항상 비어 있다.
func a169U2Score(v float64, bars []domain.Candle, day, open, close, selected time.Time) (a169U2Result, error) {
	fail := func(reason string) (a169U2Result, error) { return a169U2Result{}, errors.New(reason) }
	kst := time.FixedZone("KST", 9*60*60)
	d := day.In(kst)
	midnight := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, kst)
	if len(bars) > 25 || day.IsZero() || !day.Equal(midnight) || open.IsZero() || close.IsZero() || selected.IsZero() || open.Before(midnight) || !open.Before(midnight.AddDate(0, 0, 1)) || close.Before(midnight) || !close.Before(midnight.AddDate(0, 0, 1)) || !close.After(open) || !selected.After(open) || selected.After(close) {
		return fail("envelope_invalid")
	}
	if !a169U2Volume(v) {
		return fail("numerator_invalid")
	}
	prior := make([]domain.Candle, 0, len(bars))
	seen := make(map[string]bool)
	var next time.Time
	for _, bar := range bars {
		if bar.Time.IsZero() {
			continue
		}
		if !bar.Time.Before(midnight) {
			if next.IsZero() || bar.Time.Before(next) {
				next = bar.Time
			}
			continue
		}
		date := bar.Time.In(kst).Format("2006-01-02")
		if seen[date] {
			return fail("baseline_invalid")
		}
		seen[date] = true
		prior = append(prior, bar)
	}
	if len(prior) < 20 {
		return fail("history_lt_20")
	}
	sort.Slice(prior, func(i, j int) bool { return prior[i].Time.After(prior[j].Time) })
	sum := new(big.Int)
	zeros := 0
	for _, bar := range prior[:20] {
		if !a169U2Volume(bar.Volume) {
			return fail("baseline_invalid")
		}
		if bar.Volume == 0 {
			zeros++
		}
		sum.Add(sum, big.NewInt(int64(bar.Volume)))
	}
	if sum.Sign() == 0 {
		return fail("baseline_zero")
	}
	ratio := new(big.Int).Mul(big.NewInt(int64(v)), big.NewInt(20_000_000))
	ratio.Quo(ratio, sum)
	increase := new(big.Int).Mul(ratio, big.NewInt(int64(close.Sub(open))))
	increase.Quo(increase, big.NewInt(int64(selected.Sub(open))))
	if !sum.IsInt64() || !ratio.IsInt64() || !increase.IsInt64() {
		return fail("score_overflow")
	}
	return a169U2Result{V: int64(v), S20: sum.Int64(), RatioPPM: ratio.Int64(), IncreasePPM: increase.Int64(), PriorDays: len(prior), ZeroVolumeDays: zeros, ZeroV: v == 0, PriorBarAt: prior[0].Time, NextBarAt: next}, nil
}
