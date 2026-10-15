package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a139: 생산 정책의 건당 위험 예산은 게이트 일일 손실의 1/3(최소 단위 내림) — 유도는 risk.PerTradeRiskBudget 한 곳.
func TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss(t *testing.T) {
	krw := config.AutomationGate{Enabled: true, MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 100_000, MaxDailyLossRatio: 0.01, LimitCurrency: "KRW"}
	policy := riskPolicyFromAutomationGate(krw)
	if policy.RiskBudget != (riskcalc.Money{Amount: "33333", Currency: "KRW"}) {
		t.Fatalf("KRW risk budget = %+v, want 33333 KRW", policy.RiskBudget)
	}
	if policy.MaxDailyLoss.Amount != "100000" {
		t.Fatalf("the daily loss itself must not change, got %+v", policy.MaxDailyLoss)
	}
	// 게이트에서 사이징까지: 폭 1,000 손절이면 위험 기반 상한 33(이전 100).
	if got, err := risk.StrategyEntryQuantity(policy, "10000", "9000"); err != nil || got != "33" {
		t.Fatalf("production sizing = %s, %v; want 33", got, err)
	}

	usd := config.AutomationGate{Enabled: true, MaxOrderQuantity: 100, MaxOrderNotional: 300,
		MaxTotalExposure: 1_000, MaxDailyLossAmount: 50.75, MaxDailyLossRatio: 0.01, LimitCurrency: "USD", MaxDailyBuyNotional: 2_000}
	if got := riskPolicyFromAutomationGate(usd).RiskBudget; got != (riskcalc.Money{Amount: "16.91", Currency: "USD"}) {
		t.Fatalf("USD risk budget = %+v, want 16.91 USD", got)
	}

	tiny := usd
	tiny.MaxDailyLossAmount = 0.02
	if err := riskPolicyFromAutomationGate(tiny).Validate(); err == nil {
		t.Fatal("a daily loss whose third rounds to zero must not produce a usable Guardian policy")
	}
	if _, err := newProductionRiskGuardian(tiny, nil, nil, "acct", nil); !errors.Is(err, ErrLimitsRequired) {
		t.Fatalf("the production Guardian must refuse to exist when the risk budget rounds to zero, got %v", err)
	}
}

// freeze 리뷰 P3 · 변이 「유도 실패 시 dailyLoss 로 폴백」: 표에 없는 통화는 생산 배선 수준에서도 거부되고, 문구가 유도 실패를 말함.
func TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName(t *testing.T) {
	jpy := config.AutomationGate{Enabled: true, MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 100_000, MaxDailyLossRatio: 0.01, LimitCurrency: "JPY", MaxDailyBuyNotional: 1_000_000}
	if got := riskPolicyFromAutomationGate(jpy).RiskBudget; got.Amount != "" {
		t.Fatalf("an underivable budget must not fall back to any amount, got %+v", got)
	}
	_, err := newProductionRiskGuardian(jpy, nil, nil, "acct", nil)
	// 사전 검사의 문구(「cannot be derived」와 원인)를 단언 — Validate 의 빈 금액 문구로 대신 거부되는 것과 갈라짐(변이 N12).
	if !errors.Is(err, ErrLimitsRequired) || !strings.Contains(err.Error(), "per-trade risk budget cannot be derived") ||
		!strings.Contains(err.Error(), "JPY") {
		t.Fatalf("want ErrLimitsRequired naming why the per-trade risk budget cannot be derived, got %v", err)
	}
	// 기동 audit 값은 생산 유도의 답 — 같은 함수(66666 = floor(200000 / 3)).
	if got := perTradeRiskBudgetAudit(config.AutomationGate{MaxDailyLossAmount: 200_000, LimitCurrency: "KRW"}); got.value != "66666" {
		t.Fatalf("audit value = %q, want the production derivation 66666", got.value)
	}
	if got := perTradeRiskBudgetAudit(config.AutomationGate{}); got.value != "" || got.detail == "" {
		t.Fatalf("an underivable budget is audited as empty with a reason, got %+v", got)
	}
}

// freeze 리뷰 P2-3: 사이징 규칙이 바뀌었으므로 결정 preimage 의 정책 버전도 바뀜.
func TestA139TheProductionPolicyVersionNamesTheNewRule(t *testing.T) {
	if productionGuardianPolicyVersion != "engine.automation_gate/risk-policy-v2" {
		t.Fatalf("policy version = %q, want risk-policy-v2", productionGuardianPolicyVersion)
	}
}
