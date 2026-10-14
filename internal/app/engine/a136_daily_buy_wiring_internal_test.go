package engine

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a136 D5: Guardian 정책의 하루 매수 상한은 config 한 곳의 해석을 따름 — KRW 무값은 기본값, USD 무값은 정책 검증 거부.
func TestA136ProductionPolicyResolvesTheDailyBuyCeiling(t *testing.T) {
	krw := config.AutomationGate{Enabled: true, MaxOrderQuantity: 10, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 100_000, MaxDailyLossRatio: 0.01, LimitCurrency: "KRW"}
	if got := riskPolicyFromAutomationGate(krw).MaxDailyBuyNotional; got != (riskcalc.Money{Amount: "10000000", Currency: "KRW"}) {
		t.Fatalf("absent KRW ceiling = %+v, want the 10,000,000 KRW default", got)
	}
	// 기본값이 두 패키지(config 상수 · risk.DefaultPolicy)에 있으므로 둘이 같은지 대조함(freeze 리뷰 P3).
	if got, want := riskPolicyFromAutomationGate(krw).MaxDailyBuyNotional, risk.DefaultPolicy().MaxDailyBuyNotional; got != want {
		t.Fatalf("config default %+v and risk.DefaultPolicy %+v disagree", got, want)
	}

	explicit := krw
	explicit.MaxDailyBuyNotional = 5_000_000
	if got := riskPolicyFromAutomationGate(explicit).MaxDailyBuyNotional.Amount; got != "5000000" {
		t.Fatalf("explicit ceiling = %s, want 5000000", got)
	}

	usd := krw
	usd.LimitCurrency = "USD"
	usd.MaxOrderNotional, usd.MaxTotalExposure, usd.MaxDailyLossAmount = 300, 1_000, 50
	if err := riskPolicyFromAutomationGate(usd).Validate(); err == nil {
		t.Fatal("a USD gate without max_daily_buy_notional must not produce a usable Guardian policy (D5)")
	}
	if _, err := newProductionRiskGuardian(usd, nil, nil, "acct", nil); err == nil {
		t.Fatal("the production Guardian must refuse to exist for a USD gate without a daily buy ceiling")
	}
}
