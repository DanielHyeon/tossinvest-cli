package main

import (
	"fmt"
	"reflect"
	"testing"
)

// ucWindow 는 d−1 … d−21 날짜(가짜 — 정렬 의미만).
func ucWindow() []string {
	out := make([]string, 21)
	for i := range out {
		out[i] = fmt.Sprintf("2026-06-%02d", 30-i)
	}
	return out
}

// ucDaily 는 모든 창 날짜에 종가 close · 거래량 base 인 일봉, d−1 거래량만 prev.
func ucDaily(window []string, close, base, prev float64) map[string]a203Day {
	days := map[string]a203Day{}
	for i, date := range window {
		v := base
		if i == 0 {
			v = prev
		}
		days[date] = a203Day{Close: close, Volume: v}
	}
	return days
}

func shares(picks []a203Pick) map[string]string {
	out := map[string]string{}
	for _, p := range picks {
		out[p.Symbol] = p.Share
	}
	return out
}

func TestA203UCUsesOnlyPriorSessions(t *testing.T) {
	t.Parallel()
	w := ucWindow()
	daily := map[string]map[string]a203Day{}
	var pop []string
	for i := 0; i < 40; i++ {
		code := fmt.Sprintf("TST%03d", i)
		pop = append(pop, code)
		daily[code] = ucDaily(w, 10_000, 100, float64(100+i))
	}
	picks, n := a203SelectUC(w, pop, daily, 500_000)
	// 당일(d) 봉을 아무렇게나 바꿔도 같음 — 함수는 창 밖 날짜를 읽지 않음.
	for _, code := range pop {
		daily[code]["2026-07-01"] = a203Day{Close: 999_999, Volume: 1e12}
	}
	again, _ := a203SelectUC(w, pop, daily, 500_000)
	if !reflect.DeepEqual(picks, again) || n != 40 || len(picks) != 30 {
		t.Fatalf("picks changed with day d or wrong size: %d %d", n, len(picks))
	}
}

func TestA203UCSharesTiesAndBounds(t *testing.T) {
	t.Parallel()
	w := ucWindow()
	daily := map[string]map[string]a203Day{}
	var pop []string
	add := func(code string, days map[string]a203Day) {
		pop = append(pop, code)
		daily[code] = days
	}
	// A 후보: T(d−1) 큰 16 개(가격 10,000 · d−1 거래량 1000+i) — 하나는 A 밖으로 밀려 B 경쟁에 남음.
	for i := 0; i < 16; i++ {
		add(fmt.Sprintf("AAA%03d", i), ucDaily(w, 10_000, 1000, float64(10_000+i)))
	}
	// B 후보: T 작지만 R 큼. BBB000 과 BBB001 은 R 동률 → 코드 오름차순.
	for i := 0; i < 20; i++ {
		ratio := float64(50 - i)
		if i == 1 {
			ratio = 50
		}
		add(fmt.Sprintf("BBB%03d", i), ucDaily(w, 10_000, 10, 10*ratio))
	}
	// 경계: 종가 999 원(자격 없음) · 1,000 원(자격) · 명목 상한 초과(자격 없음) · 상한과 같음(자격).
	add("LOW999", ucDaily(w, 999, 1e9, 1e9))
	add("LOW1K0", ucDaily(w, 1000, 1, 1))
	add("CAPOVR", ucDaily(w, 500_001, 1e9, 1e9))
	add("CAPEQL", ucDaily(w, 500_000, 1, 1))
	picks, n := a203SelectUC(w, pop, daily, 500_000)
	got := shares(picks)
	if n != 38 {
		t.Fatalf("eligible %d, want 38 (999 원 · 상한 초과 제외)", n)
	}
	if _, ok := got["LOW999"]; ok {
		t.Fatal("a sub-1,000 close was picked")
	}
	if _, ok := got["CAPOVR"]; ok {
		t.Fatal("a close above the notional cap was picked")
	}
	countA := 0
	for code, share := range got {
		if share == a203ShareA {
			countA++
			if code[:3] != "AAA" {
				t.Errorf("A holds %s", code)
			}
		}
	}
	if countA != 15 || got["AAA000"] == a203ShareA {
		t.Fatalf("A = %d (AAA000 %q), want the top 15 by T", countA, got["AAA000"])
	}
	// A 에 든 종목은 B 에 안 듦 — B 는 R 순, 동률은 코드 오름차순(BBB000 < BBB001).
	if got["BBB000"] != a203ShareB || got["BBB001"] != a203ShareB || picks[15].Symbol != "BBB000" || picks[16].Symbol != "BBB001" {
		t.Fatalf("B order: %v", picks[15:18])
	}
	if len(picks) != 30 {
		t.Fatalf("picks %d", len(picks))
	}
}

func TestA203UCRDenominatorAndMissingSessions(t *testing.T) {
	t.Parallel()
	w := ucWindow()
	// R 분모에 d−1 미포함: d−1 거래대금만 바꾸면 분모(20 세션 평균 = 10,000 × 10)는 그대로 — R 이 d−1 에 정비례.
	days := ucDaily(w, 10_000, 10, 100)
	if r, ok := a203Ratio(w, days); !ok || r != 10 {
		t.Fatalf("R = %v %v, want exactly 10", r, ok)
	}
	days[w[0]] = a203Day{Close: 10_000, Volume: 200}
	if r, ok := a203Ratio(w, days); !ok || r != 20 {
		t.Fatalf("R after doubling d-1 = %v, want 20 (denominator unchanged)", r)
	}
	// M7 분모 0 → 미정의.
	if _, ok := a203Ratio(w, ucDaily(w, 10_000, 0, 100)); ok {
		t.Fatal("a zero denominator must leave R undefined")
	}
	// M6 창 결측(d−21) → 미정의, 그래도 A 자격 유지.
	gap := ucDaily(w, 10_000, 10, 100)
	delete(gap, w[20])
	if _, ok := a203Ratio(w, gap); ok {
		t.Fatal("a missing d-21 must leave R undefined")
	}
	daily := map[string]map[string]a203Day{"BIGTTT": ucDaily(w, 10_000, 1000, 5000), "SURGEX": gap}
	pop := []string{"BIGTTT", "SURGEX"}
	if picks, n := a203SelectUC(w, pop, daily, 500_000); n != 2 || len(picks) != 2 || shares(picks)["SURGEX"] != a203ShareA {
		t.Fatalf("M6/M8: a window gap keeps A eligibility and n<30 uses all, A first: %v n=%d", picks, n)
	}
	// d−1 결측(M6): 자격 없음.
	delete(daily["SURGEX"], w[0])
	if picks, n := a203SelectUC(w, pop, daily, 500_000); n != 1 || len(picks) != 1 {
		t.Fatalf("missing d-1 must drop eligibility: %v n=%d", picks, n)
	}
}

func TestA203UCBFillAndRUndefined(t *testing.T) {
	t.Parallel()
	w := ucWindow()
	daily := map[string]map[string]a203Day{}
	var pop []string
	for i := 0; i < 40; i++ {
		code := fmt.Sprintf("TST%03d", i)
		pop = append(pop, code)
		daily[code] = ucDaily(w, 10_000, 100, float64(1000-i)) // T 순 = 코드 순
	}
	// A = TST000..014. A 밖에서 R 이 정의된 종목을 셋만 남김: 나머지는 분모 0(M7) 또는 창 결측(M6).
	for i := 15; i < 40; i++ {
		code := fmt.Sprintf("TST%03d", i)
		switch {
		case i < 18:
		case i%2 == 0:
			for _, date := range w[1:] {
				daily[code][date] = a203Day{Close: 10_000, Volume: 0} // 분모 0 → R 미정의
			}
		default:
			delete(daily[code], w[21-1])
		}
	}
	picks, _ := a203SelectUC(w, pop, daily, 500_000)
	got := shares(picks)
	b, fill := 0, 0
	for _, s := range got {
		switch s {
		case a203ShareB:
			b++
		case a203ShareFill:
			fill++
		}
	}
	if b != 3 || fill != 12 || len(picks) != 30 {
		t.Fatalf("B %d fill %d total %d, want 3 · 12 · 30", b, fill, len(picks))
	}
	// 채움은 A · B 밖의 T 순(= 코드 순): TST018 부터.
	if got["TST018"] != a203ShareFill || picks[18].Symbol != "TST018" {
		t.Fatalf("fill order: %v", picks[18:20])
	}
}

func TestA203PopulationFingerprint(t *testing.T) {
	t.Parallel()
	sum, raw := a203ListFingerprint([]string{"TST002", "TST001"})
	if string(raw) != "TST001\nTST002\n" || sum != shaOf(raw) {
		t.Fatalf("list bytes %q sum %s", raw, sum)
	}
	if got := a203Union([]string{"TST002", "TST001"}, []string{"TST003", "TST001"}); !reflect.DeepEqual(got, []string{"TST001", "TST002", "TST003"}) {
		t.Fatalf("union %v", got)
	}
}
