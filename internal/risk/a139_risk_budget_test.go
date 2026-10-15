package risk

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a139: 건당 위험 예산 = floor(일일 손실 한도 / 3) — 통화 최소 단위(KRW 1 · USD 0.01)로 내림, 0 으로 내려가거나 최소 단위를 모르면 오류.
func TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit(t *testing.T) {
	for _, c := range []struct {
		in   riskcalc.Money
		want string
	}{
		{riskcalc.Money{Amount: "100000", Currency: "KRW"}, "33333"},
		{riskcalc.Money{Amount: "200000", Currency: "krw "}, "66666"},
		{riskcalc.Money{Amount: "99999", Currency: "KRW"}, "33333"},
		{riskcalc.Money{Amount: "3", Currency: "KRW"}, "1"},
		{riskcalc.Money{Amount: "100000.9", Currency: "KRW"}, "33333"},
		{riskcalc.Money{Amount: "50", Currency: "USD"}, "16.66"},
		{riskcalc.Money{Amount: "50.75", Currency: "USD"}, "16.91"},
		{riskcalc.Money{Amount: "60", Currency: "USD"}, "20"},
		{riskcalc.Money{Amount: "0.03", Currency: "USD"}, "0.01"},
		{riskcalc.Money{Amount: "0.3", Currency: "USD"}, "0.1"},
	} {
		got, err := PerTradeRiskBudget(c.in)
		if err != nil || got.Amount != c.want || got.Currency != "KRW" && got.Currency != "USD" {
			t.Errorf("PerTradeRiskBudget(%+v) = %+v, %v; want %s", c.in, got, err, c.want)
		}
	}
	for _, bad := range []riskcalc.Money{
		{Amount: "2", Currency: "KRW"},      // 1/3 이 0 으로 내려감
		{Amount: "0.02", Currency: "USD"},   // 〃
		{Amount: "0", Currency: "KRW"},      // 양수가 아님
		{Amount: "-30", Currency: "KRW"},    // 〃
		{Amount: "NaN", Currency: "KRW"},    // 해석 불가
		{Amount: "", Currency: "KRW"},       // 미설정
		{Amount: "1e5", Currency: "KRW"},    // 지수 표기는 decimal 이 아님
		{Amount: "100000", Currency: "JPY"}, // 최소 단위를 모르는 통화
		{Amount: "100000", Currency: ""},
	} {
		if got, err := PerTradeRiskBudget(bad); err == nil {
			t.Errorf("PerTradeRiskBudget(%+v) = %+v; want an error", bad, got)
		}
	}
	if PerTradeRiskBudgetDivisor != 3 {
		t.Fatalf("the divisor is the user's decision of 2026-10-15: 3, got %d", PerTradeRiskBudgetDivisor)
	}
}

// 기본 정책도 같은 규칙 — 주석이 적는 유도와 값이 갈라지지 않게 함수의 답과 대조함.
func TestA139TheDefaultPolicyFollowsTheSameRule(t *testing.T) {
	p := DefaultPolicy()
	want, err := PerTradeRiskBudget(p.MaxDailyLoss)
	if err != nil {
		t.Fatal(err)
	}
	if p.RiskBudget != want || p.RiskBudget != (riskcalc.Money{Amount: "33333", Currency: "KRW"}) {
		t.Fatalf("default risk budget = %+v, want %+v (= a third of %+v)", p.RiskBudget, want, p.MaxDailyLoss)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("the default policy must validate: %v", err)
	}
}

// 사이징이 실제로 1/3 로 줄어드는지 — 수량 · notional 상한이 묶지 않는 가격(진입 10,000 · 손절 9,000, 폭 1,000)에서
// 위험 기반 상한이 floor(33,333 / 1,000) = 33(이전 100).
func TestA139TheRiskCapIsAThirdOfTheDailyLoss(t *testing.T) {
	got, err := StrategyEntryQuantity(DefaultPolicy(), "10000", "9000")
	if err != nil {
		t.Fatal(err)
	}
	if got != "33" {
		t.Fatalf("Guardian sizing at a 1,000 stop = %s, want 33 (a third of the daily loss over the stop)", got)
	}
}
