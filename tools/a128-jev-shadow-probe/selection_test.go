package main

// selection_test.go 는 1.5 표본 선택 규칙(상위 100 중 주문 상한 이하 무작위 K)을 고정함.
// 실 API·실 config 는 쓰지 않음 — httptest mock 과 임시 디렉터리의 config 파일만.

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// judgedSymbols 는 원장 judgment 행의 종목 집합(정렬).
func judgedSymbols(contents ledgerContents) []string {
	set := map[string]bool{}
	for _, row := range contents.Judgments {
		set[row.Symbol] = true
	}
	return sortedKeys(set)
}

// onlySelection 은 모든 judgment 행이 같은 selection 을 공유함을 확인하고 그 값을 돌려줌.
func onlySelection(t *testing.T, contents ledgerContents) selectionRecord {
	t.Helper()
	if len(contents.Judgments) == 0 {
		t.Fatal("no judgment rows")
	}
	first := contents.Judgments[0].Selection
	if first == nil {
		t.Fatal("a judgment row lacks its selection record")
	}
	for _, row := range contents.Judgments {
		if row.Selection == nil || *row.Selection != *first {
			t.Fatalf("rows of one cycle disagree on the selection: %+v vs %+v", row.Selection, first)
		}
	}
	return *first
}

func selectionGaps(contents ledgerContents) []gapRow {
	var gaps []gapRow
	for _, row := range contents.Gaps {
		if row.Stage == "selection" {
			gaps = append(gaps, row)
		}
	}
	return gaps
}

// TestSelectionCapBoundaryAndUnknownPrices 는 KR(상한 통화와 같음): 가격 == 상한은 통과, 상한+1 은
// 탈락, 가격 0·통화 불일치·통화 빈칸은 「가격 모름」으로 빠지고 수가 기록됨. 환율은 부르지 않음.
func TestSelectionCapBoundaryAndUnknownPrices(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.setRanking(
		rankingRow{"AT_CAP", "500000", "KRW"},
		rankingRow{"OVER_CAP", "500001", "KRW"},
		rankingRow{"CHEAP", "70000", "KRW"},
		rankingRow{"NO_PRICE", "0", "KRW"},
		rankingRow{"WRONG_CCY", "100", "USD"},
		rankingRow{"NO_CCY", "100", ""},
	)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if got, want := judgedSymbols(contents), []string{"AT_CAP", "CHEAP"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("judged %v, want %v (== cap passes, cap+1 and unknown prices are excluded)", got, want)
	}
	got := onlySelection(t, contents)
	want := selectionRecord{
		Rule: selectionRule, Seed: testSeed, CapValue: 500000, CapCurrency: "KRW", PriceCurrency: "KRW",
		PriceSource: selectionPriceSource, RankingRequested: 100, RankingReceived: 6,
		ExcludedOverCap: 1, ExcludedUnknownPrice: 3, Candidates: 2,
		Pool: 30, Selected: 2, Judged: 2, FinalistsRequested: 10, Finalists: 2,
	}
	if got != want {
		t.Fatalf("selection = %+v\nwant        %+v", got, want)
	}
	if n := rig.official.callCount("/api/v1/exchange-rate"); n != 0 {
		t.Fatalf("a same-currency market read the exchange rate %d times", n)
	}
	query, _ := url.ParseQuery(rig.official.query("/api/v1/rankings"))
	if query.Get("count") != "100" {
		t.Fatalf("rankings count = %q, want 100", query.Get("count"))
	}
}

// TestSelectionConvertsUSPricesWithTheOfficialRate 는 US: 가격 × 환율(USD→KRW) ≤ 상한. 환산 없이 비교하면
// 400.01 USD 도 500000 이하로 통과해 버림.
func TestSelectionConvertsUSPricesWithTheOfficialRate(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.fxRate = "1250"
	rig.official.setRanking(
		rankingRow{"AT_CAP", "400", "USD"},      // 400 × 1250 = 500000 → 통과
		rankingRow{"OVER_CAP", "400.01", "USD"}, // 500012.5 → 탈락
		rankingRow{"CHEAP", "100", "USD"},
	)
	if err := rig.probe.cycle(context.Background(), "US", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if got, want := judgedSymbols(contents), []string{"AT_CAP", "CHEAP"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("judged %v, want %v", got, want)
	}
	got := onlySelection(t, contents)
	if got.FXRate != 1250 || got.FXPair != "USD/KRW" || got.PriceCurrency != "USD" || got.CapCurrency != "KRW" ||
		got.ExcludedOverCap != 1 || got.Candidates != 2 {
		t.Fatalf("US selection = %+v", got)
	}
	query, _ := url.ParseQuery(rig.official.query("/api/v1/exchange-rate"))
	if query.Get("baseCurrency") != "USD" || query.Get("quoteCurrency") != "KRW" {
		t.Fatalf("exchange rate asked for %v, want USD → KRW", query)
	}
}

// TestSelectionSeedReproduces 는 같은 후보·같은 시드 → 같은 선택, 그리고 알고리즘 자체를 골든으로 고정함
// (원장에 적힌 옛 시드가 코드 변경 뒤에도 같은 선택을 재현하게).
func TestSelectionSeedReproduces(t *testing.T) {
	t.Parallel()
	var candidates []domain.RankingItem
	for index := 0; index < 30; index++ {
		candidates = append(candidates, domain.RankingItem{Rank: index + 1, Symbol: fmt.Sprintf("C%02d", index)})
	}
	symbolsOf := func(items []domain.RankingItem) []string {
		out := make([]string, 0, len(items))
		for _, item := range items {
			out = append(out, item.Symbol)
		}
		return out
	}
	a, b := symbolsOf(pickSample(candidates, 10, 7)), symbolsOf(pickSample(candidates, 10, 7))
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed, different selection: %v vs %v", a, b)
	}
	if !sort.StringsAreSorted(a) {
		t.Fatalf("the sample must be returned in ranking order: %v", a)
	}
	if strings.Join(a, ",") != goldenSeed7 {
		t.Fatalf("seed 7 selection = %s, want the pinned %s (changing the algorithm breaks old seeds)", strings.Join(a, ","), goldenSeed7)
	}
	differs := false
	for seed := uint64(8); seed < 12; seed++ {
		if !reflect.DeepEqual(symbolsOf(pickSample(candidates, 10, seed)), a) {
			differs = true
		}
	}
	if !differs {
		t.Fatal("four other seeds all reproduced seed 7's selection — the seed is not used")
	}

	// 사이클 단위: 고정 시드의 두 rig 는 같은 종목을 판단함.
	var rows []rankingRow
	for index := 0; index < 30; index++ {
		rows = append(rows, rankingRow{fmt.Sprintf("C%02d", index), "1000", "KRW"})
	}
	var runs [][]string
	for range 2 {
		rig := newTestRig(t)
		rig.probe.seed = func() uint64 { return 7 }
		rig.probe.cfg.Pool = 10 // 풀 10 = 골든의 K — 풀 추출이 같은 pickSample 절차임을 사이클 단위로 봄
		rig.official.setRanking(rows...)
		if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
			t.Fatal(err)
		}
		runs = append(runs, judgedSymbols(rig.contents(t)))
		if got := onlySelection(t, rig.contents(t)); got.Seed != 7 {
			t.Fatalf("recorded seed = %d, want 7", got.Seed)
		}
	}
	if !reflect.DeepEqual(runs[0], runs[1]) || strings.Join(runs[0], ",") != goldenSeed7 {
		t.Fatalf("cycle selections = %v, want both %s", runs, goldenSeed7)
	}
	if selectionRuleFor(10, 30) != selectionRule || selectionRule != "jev_screen10_of_random30_of_top100_le_order_cap" ||
		selectionRuleFor(5, 20) != "jev_screen5_of_random20_of_top100_le_order_cap" {
		t.Fatalf("rule ids = %q / %q", selectionRuleFor(10, 30), selectionRuleFor(5, 20))
	}
	if newSelectionSeed() > seedMask {
		t.Fatal("seeds must fit in 53 bits so JSON readers keep them exact")
	}
}

// goldenSeed7 은 30 후보(C00..C29)·K=10·시드 7 의 선택 — pickSample 의 재현 절차를 못 박음.
// 값은 pickSample 이 아니라 문서화된 절차(NewPCG(7, selectionPCGStream).Perm(30)[:10] 정렬)를
// 따로 짠 프로그램으로 계산해 넣음(2026-10-14, go1.26).
const goldenSeed7 = "C02,C03,C09,C16,C19,C20,C21,C24,C26,C29"

// TestSelectionFXFailureLeavesAGapAndNoJudgment 는 환율 실패(오류·짝 불일치·0)가 gap(selection) 1행이
// 되고 랭킹도 판단도 부르지 않음을 봄 — 옛 규칙이나 환율 1 로 조용히 폴백하지 않음.
func TestSelectionFXFailureLeavesAGapAndNoJudgment(t *testing.T) {
	t.Parallel()
	cases := map[string]func(*officialMock){
		"read failed":   func(m *officialMock) { m.failStatus["/api/v1/exchange-rate"] = 503 },
		"pair reversed": func(m *officialMock) { m.fxBase, m.fxQuote = "KRW", "USD" },
		"zero rate":     func(m *officialMock) { m.fxRate = "0" },
	}
	for name, breakFX := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rig := newTestRig(t)
			rig.official.setRanking(rankingRow{"CHEAP", "100", "USD"})
			rig.official.mu.Lock()
			breakFX(rig.official)
			rig.official.mu.Unlock()
			if err := rig.probe.cycle(context.Background(), "US", rig.session(), at(10, 0)); err != nil {
				t.Fatal(err)
			}
			contents := rig.contents(t)
			gaps := selectionGaps(contents)
			if len(contents.Judgments) != 0 || len(gaps) != 1 || !strings.Contains(gaps[0].Reason, "exchange rate") {
				t.Fatalf("an FX failure must leave one selection gap and no judgment: %+v", contents)
			}
			if rig.official.callCount("/api/v1/rankings") != 0 || len(rig.typeSafe.requests()) != 0 {
				t.Fatal("a cycle without an order cap conversion must not rank or judge")
			}
		})
	}
}

// TestSelectionNeedsTheConfigCap 는 config 부재·상한 0·통화 이상·출처 미배선이 각각 gap(selection) 이 되고
// 판단하지 않음을 봄. config 는 실 로더(internal/config)로 임시 파일을 읽음.
func TestSelectionNeedsTheConfigCap(t *testing.T) {
	t.Parallel()
	writeConfig := func(t *testing.T, body string) string {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	cases := map[string]struct {
		source func(t *testing.T) func(context.Context) (orderCap, error)
		reason string
	}{
		"config missing": {func(t *testing.T) func(context.Context) (orderCap, error) {
			path := filepath.Join(t.TempDir(), "config.json")
			return func(ctx context.Context) (orderCap, error) { return loadOrderCap(ctx, path) }
		}, "not found"},
		"cap zero": {func(t *testing.T) func(context.Context) (orderCap, error) {
			path := writeConfig(t, `{"engine":{"automation_gate":{"limit_currency":"KRW"}}}`)
			return func(ctx context.Context) (orderCap, error) { return loadOrderCap(ctx, path) }
		}, "max_order_notional"},
		"currency unknown": {func(t *testing.T) func(context.Context) (orderCap, error) {
			path := writeConfig(t, `{"engine":{"automation_gate":{"max_order_notional":500000,"limit_currency":"EUR"}}}`)
			return func(ctx context.Context) (orderCap, error) { return loadOrderCap(ctx, path) }
		}, "limit_currency"},
		"no source": {func(*testing.T) func(context.Context) (orderCap, error) { return nil }, "no order cap source"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rig := newTestRig(t)
			rig.probe.cfg.CapSource = tc.source(t)
			if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
				t.Fatal(err)
			}
			contents := rig.contents(t)
			gaps := selectionGaps(contents)
			if len(contents.Judgments) != 0 || len(gaps) != 1 || !strings.Contains(gaps[0].Reason, tc.reason) {
				t.Fatalf("want one selection gap naming %q and no judgment: %+v", tc.reason, contents)
			}
			if rig.official.callCount("/api/v1/rankings") != 0 {
				t.Fatal("no ranking read may happen without an order cap")
			}
		})
	}

	// 정상 config: 실 로더가 값과 통화(대문자 정규화)를 읽음.
	path := writeConfig(t, `{"engine":{"automation_gate":{"max_order_notional":500000,"limit_currency":" krw "}}}`)
	got, err := loadOrderCap(context.Background(), path)
	if err != nil || got != (orderCap{Value: 500000, Currency: "KRW"}) {
		t.Fatalf("loadOrderCap = %+v, %v", got, err)
	}
}

// TestSelectionWithNoCandidateLeavesAGap 는 상한 이하 후보가 0 이면 gap(selection) 으로 남기고 판단하지 않음.
func TestSelectionWithNoCandidateLeavesAGap(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.setRanking(rankingRow{"PRICEY", "900000", "KRW"}, rankingRow{"NO_PRICE", "0", "KRW"})
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	gaps := selectionGaps(contents)
	if len(contents.Judgments) != 0 || len(gaps) != 1 ||
		!strings.Contains(gaps[0].Reason, "received 2, over cap 1, unknown price 1") {
		t.Fatalf("want one selection gap with the exclusion counts: %+v", contents)
	}
	if len(rig.typeSafe.requests()) != 0 {
		t.Fatal("no judgment may be requested without a candidate")
	}
}
