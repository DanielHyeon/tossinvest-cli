package main

import (
	"fmt"
	"math"
	"testing"
)

// trades 는 세션 sessions 개에 하루 perDay 건씩, 건당 r = value(세션 번호) 인 라벨 제안.
func tradesOf(sessions, perDay int, value func(int) float64) []a203Trade {
	var out []a203Trade
	for s := 0; s < sessions; s++ {
		for i := 0; i < perDay; i++ {
			out = append(out, a203Trade{Session: fmt.Sprintf("KRX:2026-07-%02d", s+1), Labeled: true, Return: value(s)})
		}
	}
	return out
}

func TestA203InsufficientConditions(t *testing.T) {
	t.Parallel()
	pos := func(int) float64 { return 0.01 }
	cases := []struct {
		name        string
		trades      []a203Trade
		gaps, total int
		want        string
	}{
		{"29 labeled", tradesOf(29, 1, pos), 0, 100, a203Insufficient},
		{"30 labeled over 20 sessions", append(tradesOf(20, 1, pos), tradesOf(10, 1, pos)...), 0, 100, a203Pass},
		{"19 proposal sessions", tradesOf(19, 2, pos), 0, 100, a203Insufficient},
		{"20 proposal sessions", tradesOf(20, 2, pos), 0, 100, a203Pass},
		{"gaps exactly 20%", tradesOf(20, 2, pos), 20, 100, a203Pass},
		{"gaps above 20%", tradesOf(20, 2, pos), 21, 100, a203Insufficient},
		{"an unlabeled proposal", append(tradesOf(20, 2, pos), a203Trade{Session: "KRX:2026-07-01"}), 0, 100, a203Insufficient},
	}
	for _, c := range cases {
		if got := a203Judge(c.trades, c.gaps, c.total); got.Verdict != c.want {
			t.Errorf("%s: %s %v, want %s", c.name, got.Verdict, got.Reasons, c.want)
		}
	}
}

func TestA203PassNeedsAllThreeConditions(t *testing.T) {
	t.Parallel()
	// (1) 평균 ≤ 0 → FAIL.
	if j := a203Judge(tradesOf(30, 1, func(int) float64 { return -0.001 }), 0, 100); j.Verdict != a203Fail {
		t.Fatalf("negative mean: %+v", j)
	}
	// (3) 가장 좋은 하루 하나가 평균을 만든 경우: 평균 > 0 이지만 그날을 빼면 ≤ 0 → FAIL.
	oneBig := tradesOf(30, 1, func(s int) float64 {
		if s == 7 {
			return 0.31
		}
		return -0.01
	})
	j := a203Judge(oneBig, 0, 100)
	if j.Mean <= 0 || j.ExcludingBest > 0 || j.Verdict != a203Fail {
		t.Fatalf("best-day carried: %+v", j)
	}
	// (2) 평균 > 0 · 최고일 제외 > 0 이지만 부트스트랩 하한 ≤ 0 → FAIL.
	noisy := tradesOf(30, 1, func(s int) float64 { return []float64{0.05, -0.045}[s%2] })
	j = a203Judge(noisy, 0, 100)
	if j.Mean <= 0 || j.ExcludingBest <= 0 || j.Lower > 0 || j.Verdict != a203Fail {
		t.Fatalf("bootstrap lower: %+v", j)
	}
	if j := a203Judge(tradesOf(30, 1, func(s int) float64 { return 0.01 + 0.001*float64(s%3) }), 0, 100); j.Verdict != a203Pass {
		t.Fatalf("all positive: %+v", j)
	}
}

// a182 report.go(feat/a182-kr-crash-day-rule-replay e0890b870) 의 bootstrapLower 를 같은 입력 · seed 203 으로 돌린 값
// (그 커밋의 percentile · mean · bootstrapLower 원문을 그대로 실행 — 2026-11-27, 원문 발췌 sha256 1ff64eea4e44…).
func TestA203BootstrapMatchesA182(t *testing.T) {
	t.Parallel()
	days := []float64{0.01, -0.02, 0.005, 0.03, -0.001, 0.012, -0.007, 0.004, 0.0, -0.015, 0.021}
	if got := a203BootstrapLower(days, 203); math.Float64bits(got) != 0xbf735ce79c34e029 {
		t.Fatalf("bootstrap lower %v (%x), want a182's -0.004727272727272727", got, math.Float64bits(got))
	}
	if got := a203BootstrapLower([]float64{0.5}, 203); got != 0.5 {
		t.Fatalf("one day %v", got)
	}
	if got := a203BootstrapLower(nil, 203); got != 0 {
		t.Fatalf("empty %v", got)
	}
}
