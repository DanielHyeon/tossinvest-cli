package main

// a203_universe.go 는 종목군 U-C(design §5, 결정 H4 · M6 ~ M9) — 세션 d 의 30 개를 d−1 까지의 일봉만으로 고름.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// a203PAFingerprint 는 P(a) 정렬 목록(코드 오름차순, 줄마다 코드 + "\n")의 커밋된 sha256(design §5 · M9).
// 변수인 이유: 시험이 가짜 종목 코드 모집단으로 바꿔 끼움(실제 541 목록은 저장소에 없음).
var a203PAFingerprint = "08a1ac60788dbc1ecf7ed0193bbd238a61a2d5a3a8f996d817fef82eba48f56f"

const (
	a203PAFile      = "universe/pa-kr-2026-11-02_2026-11-26.txt" // 상태 디렉터리 기준
	a203ShareA      = "A"                                        // d−1 거래대금 상위
	a203ShareB      = "B"                                        // 거래대금 급증 R 상위
	a203ShareFill   = "F"                                        // B 부족분을 T(d−1) 순으로 채움
	a203ShareA140   = "a140"                                     // 대조 1 — a140 원장 종목군 그대로
	a203PerShare    = 15
	a203MinCloseKRW = 1000
	a203RWindow     = 20 // R 분모 = d−2 … d−21
)

// errA203Population 은 M9 — 사전등록 모집단이 아님(종료 코드 2, 전송 0).
var errA203Population = errors.New("the P(a) population file does not match the preregistered fingerprint")

// a203Day 는 일봉 한 개에서 쓰는 두 값.
type a203Day struct{ Close, Volume float64 }

// a203Pick 은 선정 한 종목과 그 몫.
type a203Pick struct {
	Symbol string `json:"symbol"`
	Share  string `json:"share"`
}

// a203ListFingerprint 는 P(a) 와 같은 형식(코드 오름차순 · 줄마다 코드 + "\n")의 sha256 과 그 바이트.
func a203ListFingerprint(codes []string) (string, []byte) {
	sorted := append([]string(nil), codes...)
	sort.Strings(sorted)
	var b strings.Builder
	for _, code := range sorted {
		b.WriteString(code + "\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), []byte(b.String())
}

// a203ReadPA 는 P(a) 파일을 읽고 지문을 검증함 — 다르면 errA203Population(M9).
func a203ReadPA(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errA203Population, err)
	}
	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != a203PAFingerprint {
		return nil, fmt.Errorf("%w: %s is sha256 %s, want %s", errA203Population, path, got, a203PAFingerprint)
	}
	return strings.Fields(string(raw)), nil
}

// a203Union 은 P = P(a) ∪ P(b) 를 코드 오름차순으로.
func a203Union(lists ...[]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range lists {
		for _, code := range list {
			if !seen[code] {
				seen[code] = true
				out = append(out, code)
			}
		}
	}
	sort.Strings(out)
	return out
}

// a203DailyIndex 는 일봉을 KST 날짜 → 값으로 펼침(같은 날짜가 둘이면 뒤엣것 — 표준 해독기 그대로).
func a203DailyIndex(chart domain.Chart) map[string]a203Day {
	out := make(map[string]a203Day, len(chart.Candles))
	for _, c := range chart.Candles {
		out[c.Time.In(a203Seoul).Format("2006-01-02")] = a203Day{Close: c.Close, Volume: c.Volume}
	}
	return out
}

// a203SelectUC 는 세션 d 의 선정(design §5 U-C).
//
// window[0] = d−1, window[k−1] = d−k(공식 달력 세션, M6). 당일 d 와 그 뒤 일봉은 받지도 않음.
// 자격 = d−1 일봉이 있고 종가 1,000 원 이상 · 명목 상한 이하(M6: d−1 없으면 자격 없음).
// T = 종가 × 거래량. R = T(d−1) / mean(T(d−2) … T(d−21)) — 그 20 개 중 하나라도 없거나(M6) 평균이 0 이면(M7) 정의 안 됨.
// A = T 내림차순 min(15, n) · B = A 밖에서 R 내림차순 ≤ 15 · F = 나머지 자리를 T 순으로(M8 — n < 30 이면 n 개 전부).
// 동률은 언제나 종목 코드 오름차순.
func a203SelectUC(window []string, population []string, daily map[string]map[string]a203Day, notionalCapKRW uint64) (picks []a203Pick, eligible int) {
	if len(window) < 1+a203RWindow {
		return nil, 0
	}
	type cand struct {
		code     string
		t, r     float64
		rDefined bool
	}
	var cands []cand
	for _, code := range population {
		days := daily[code]
		prev, ok := days[window[0]]
		if !ok || prev.Close < a203MinCloseKRW || prev.Close > float64(notionalCapKRW) {
			continue
		}
		c := cand{code: code, t: prev.Close * prev.Volume}
		c.r, c.rDefined = a203Ratio(window, days)
		cands = append(cands, c)
	}
	byT := func(list []cand) {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].t != list[j].t {
				return list[i].t > list[j].t
			}
			return list[i].code < list[j].code
		})
	}
	byT(cands)
	taken := map[string]bool{}
	take := func(c cand, share string) {
		taken[c.code] = true
		picks = append(picks, a203Pick{Symbol: c.code, Share: share})
	}
	for _, c := range cands[:min(a203PerShare, len(cands))] {
		take(c, a203ShareA)
	}
	var rs []cand
	for _, c := range cands {
		if !taken[c.code] && c.rDefined {
			rs = append(rs, c)
		}
	}
	sort.SliceStable(rs, func(i, j int) bool {
		if rs[i].r != rs[j].r {
			return rs[i].r > rs[j].r
		}
		return rs[i].code < rs[j].code
	})
	for _, c := range rs[:min(a203PerShare, len(rs))] {
		take(c, a203ShareB)
	}
	for _, c := range cands { // 이미 T 순
		if len(picks) >= 2*a203PerShare {
			break
		}
		if !taken[c.code] {
			take(c, a203ShareFill)
		}
	}
	return picks, len(cands)
}

// a203Ratio 는 R(x, d) = T(d−1) / mean(T(d−2) … T(d−21)) — 분모에 d−1 을 넣지 않음. 창의 20 세션 중 하나라도 없거나(M6)
// 분모가 0 이면(M7) 정의 안 됨. window[0] = d−1(있다고 가정 — 자격 검사가 먼저 걸러 냄).
func a203Ratio(window []string, days map[string]a203Day) (float64, bool) {
	sum := 0.0
	for _, date := range window[1 : 1+a203RWindow] {
		day, ok := days[date]
		if !ok {
			return 0, false
		}
		sum += day.Close * day.Volume
	}
	mean := sum / a203RWindow
	if mean <= 0 {
		return 0, false
	}
	prev := days[window[0]]
	return prev.Close * prev.Volume / mean, true
}
