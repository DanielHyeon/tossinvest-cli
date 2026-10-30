package main

import (
	"math"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func a169U2Fixture() (time.Time, time.Time, time.Time, []domain.Candle) {
	k := time.FixedZone("KST", 9*60*60)
	day := time.Date(2026, 10, 31, 0, 0, 0, 0, k)
	open := day.Add(9 * time.Hour)
	close := day.Add(15*time.Hour + 30*time.Minute)
	bars := make([]domain.Candle, 20)
	for i := range bars {
		bars[i] = domain.Candle{Time: day.AddDate(0, 0, -i-1), Volume: 10}
	}
	return day, open, close, bars
}
func TestA169U2ScoreBoundaries(t *testing.T) {
	day, open, close, bars := a169U2Fixture()
	orig := append([]domain.Candle(nil), bars...)
	got, err := a169U2Score(5, bars, day, open, close, open.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if got.V != 5 || got.S20 != 200 || got.RatioPPM != 500000 || got.IncreasePPM != 1625000 || got.PriorDays != 20 || got.ZeroVolumeDays != 0 {
		t.Fatalf("unexpected result: %+v", got)
	}
	for i := range bars {
		if bars[i] != orig[i] {
			t.Fatal("input mutated")
		}
	}
	cases := []struct {
		name                       string
		v                          float64
		bars                       []domain.Candle
		day, open, close, selected time.Time
		want                       string
	}{
		{"envelope", math.NaN(), bars, day.Add(time.Second), open, close, open.Add(time.Minute), "envelope_invalid"},
		{"numerator", math.NaN(), bars, day, open, close, open.Add(time.Minute), "numerator_invalid"},
		{"history", 1, bars[:19], day, open, close, open.Add(time.Minute), "history_lt_20"},
		{"at_open", 1, bars, day, open, close, open, "envelope_invalid"},
		{"fraction", 1.5, bars, day, open, close, open.Add(time.Minute), "numerator_invalid"},
		{"excess", 9007199254740992, bars, day, open, close, open.Add(time.Minute), "numerator_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, e := a169U2Score(tc.v, tc.bars, tc.day, tc.open, tc.close, tc.selected)
			if e == nil || e.Error() != tc.want {
				t.Fatalf("got %v want %s", e, tc.want)
			}
		})
	}
}

func TestA169U2ScoreHistoryAndFloors(t *testing.T) {
	day, open, close, bars := a169U2Fixture()
	// Unordered, zero Time, future and unused old invalid volume.
	bars = append(bars, domain.Candle{Time: day.AddDate(0, 0, -21), Volume: math.NaN()}, domain.Candle{Time: day, Volume: 999}, domain.Candle{Time: day.Add(time.Hour), Volume: 999}, domain.Candle{})
	bars[0], bars[19] = bars[19], bars[0]
	got, e := a169U2Score(0, bars, day, open, close, close)
	if e != nil || got.S20 != 200 || got.PriorDays != 21 || !got.ZeroV || got.IncreasePPM != 0 || !got.PriorBarAt.Equal(day.AddDate(0, 0, -1)) || !got.NextBarAt.Equal(day) {
		t.Fatalf("got %+v %v", got, e)
	}
	_, _, _, base := a169U2Fixture()
	dup := append(append([]domain.Candle(nil), base...), domain.Candle{Time: base[0].Time.Add(time.Hour), Volume: 0})
	if _, e = a169U2Score(1, dup, day, open, close, close); e == nil || e.Error() != "baseline_invalid" {
		t.Fatalf("duplicate: %v", e)
	}
	for i := range base {
		base[i].Volume = 0
	}
	if _, e = a169U2Score(1, base, day, open, close, close); e == nil || e.Error() != "baseline_zero" {
		t.Fatalf("zero: %v", e)
	}
	base[0].Volume = 60
	got, e = a169U2Score(1, base, day, open, open.Add(3*time.Second), open.Add(2*time.Second))
	if e != nil || got.RatioPPM != 333333 || got.IncreasePPM != 499999 || got.ZeroVolumeDays != 19 {
		t.Fatalf("two floors: %+v %v", got, e)
	}
}

func TestA169U2ScoreValidationAndOverflow(t *testing.T) {
	day, open, close, base := a169U2Fixture()
	for _, v := range []float64{math.Inf(1), math.Inf(-1), -1, 0.1, 9007199254740992} {
		if _, e := a169U2Score(v, base, day, open, close, close); e == nil || e.Error() != "numerator_invalid" {
			t.Fatalf("numerator %v: %v", v, e)
		}
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 0.1, 9007199254740992} {
		b := append([]domain.Candle(nil), base...)
		b[0].Volume = v
		if _, e := a169U2Score(1, b, day, open, close, close); e == nil || e.Error() != "baseline_invalid" {
			t.Fatalf("volume %v: %v", v, e)
		}
	}
	b := append([]domain.Candle(nil), base...)
	for i := range b {
		b[i].Volume = 0
	}
	b[0].Volume = 1
	if _, e := a169U2Score(9007199254740991, b, day, open, close, close); e == nil || e.Error() != "score_overflow" {
		t.Fatalf("overflow %v", e)
	}
	if _, e := a169U2Score(9007199254740991, base, day, open, close, close); e == nil || e.Error() != "score_overflow" {
		t.Fatalf("ratio overflow %v", e)
	}
	for _, tc := range []struct {
		name                  string
		day, open, close, sel time.Time
		n                     int
	}{
		{"zero_day", time.Time{}, open, close, close, 20}, {"other_open", day, day.AddDate(0, 0, -1), close, close, 20}, {"other_close", day, open, day.AddDate(0, 0, 1), close, 20}, {"reverse", day, close, open, close, 20}, {"after", day, open, close, close.Add(time.Nanosecond), 20}, {"too_many", day, open, close, close, 26},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bars := base
			if tc.n == 26 {
				bars = make([]domain.Candle, 26)
			}
			if _, e := a169U2Score(0, bars, tc.day, tc.open, tc.close, tc.sel); e == nil || e.Error() != "envelope_invalid" {
				t.Fatalf("got %v", e)
			}
		})
	}
	dup := append(append([]domain.Candle(nil), base[:19]...), base[0])
	shortDup := append(append([]domain.Candle(nil), dup[:18]...), dup[19])
	if _, e := a169U2Score(1, shortDup, day, open, close, close); e == nil || e.Error() != "baseline_invalid" {
		t.Fatalf("duplicate before short history: %v", e)
	}
	if _, e := a169U2Score(math.NaN(), dup, day, open, close, close); e == nil || e.Error() != "numerator_invalid" {
		t.Fatalf("priority %v", e)
	}
	if _, e := a169U2Score(1, dup[:19], day, open, close, close); e == nil || e.Error() != "history_lt_20" {
		t.Fatalf("history %v", e)
	}
	if _, e := a169U2Score(1, dup, day, open, close, close); e == nil || e.Error() != "baseline_invalid" {
		t.Fatalf("duplicate priority %v", e)
	}
}

func TestA169U2ScoreExactLimitAndLatestTwenty(t *testing.T) {
	day, open, close, bars := a169U2Fixture()
	for i := range bars {
		bars[i].Volume = 9007199254740991
	}
	got, e := a169U2Score(9007199254740991, bars, day, open, close, close)
	if e != nil || got.RatioPPM != 1000000 || got.IncreasePPM != 1000000 || got.S20 != 180143985094819820 {
		t.Fatalf("max exact: %+v %v", got, e)
	}
	for i := 21; i <= 25; i++ {
		bars = append(bars, domain.Candle{Time: day.AddDate(0, 0, -i), Volume: 0})
	}
	got, e = a169U2Score(0, bars, day, open, close, close)
	if e != nil || got.PriorDays != 25 || got.ZeroVolumeDays != 0 {
		t.Fatalf("latest20: %+v %v", got, e)
	}
	bars[24].Time = bars[23].Time.Add(time.Hour)
	if _, e = a169U2Score(0, bars, day, open, close, close); e == nil || e.Error() != "baseline_invalid" {
		t.Fatalf("old duplicate: %v", e)
	}
	_, _, _, bars = a169U2Fixture()
	if _, e = a169U2Score(10, bars, day, open, close, open.Add(time.Nanosecond)); e == nil || e.Error() != "score_overflow" {
		t.Fatalf("increase overflow: %v", e)
	}
}
