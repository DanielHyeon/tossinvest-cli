package engine

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// a138: Guardian 정책의 최대 보유 포지션 수는 config 한 곳의 해석을 따름 — 없음 · 0 은 6, 정수가 아니거나 1 미만이면 정책 검증 거부.
func TestA138ProductionPolicyResolvesTheOpenPositionMaximum(t *testing.T) {
	gate := config.AutomationGate{Enabled: true, MaxOrderQuantity: 10, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 100_000, MaxDailyLossRatio: 0.01, LimitCurrency: "KRW"}
	if got := riskPolicyFromAutomationGate(gate).MaxOpenPositions; got != 6 {
		t.Fatalf("absent maximum = %d, want the default 6", got)
	}
	// 기본값이 두 패키지(config 상수 · risk.DefaultPolicy)에 있으므로 둘이 같은지 대조함.
	if got, want := riskPolicyFromAutomationGate(gate).MaxOpenPositions, risk.DefaultPolicy().MaxOpenPositions; got != want {
		t.Fatalf("config default %d and risk.DefaultPolicy %d disagree", got, want)
	}
	usd := gate
	usd.LimitCurrency, usd.MaxOrderNotional, usd.MaxTotalExposure, usd.MaxDailyLossAmount, usd.MaxDailyBuyNotional = "USD", 300, 1_000, 50, 2_000
	if got := riskPolicyFromAutomationGate(usd).MaxOpenPositions; got != 6 {
		t.Fatalf("the default is currency-independent; USD gate maximum = %d, want 6", got)
	}
	explicit := gate
	explicit.MaxOpenPositions = 4
	if got := riskPolicyFromAutomationGate(explicit).MaxOpenPositions; got != 4 {
		t.Fatalf("explicit maximum = %d, want 4", got)
	}
	for _, bad := range []float64{-1, 2.5, math.NaN(), math.Inf(1), 1e300} {
		invalid := gate
		invalid.MaxOpenPositions = bad
		if err := riskPolicyFromAutomationGate(invalid).Validate(); err == nil {
			t.Errorf("max_open_positions %v must not produce a usable Guardian policy", bad)
		}
		// 생산 factory 경로(freeze 리뷰 P2-3): 한도 부재로 거부하고, 문구가 운영자가 적은 값을 말함(0 으로 바꾼 값이 아님).
		_, err := newProductionRiskGuardian(invalid, nil, nil, "acct", nil)
		if !errors.Is(err, ErrLimitsRequired) || !strings.Contains(err.Error(), "max_open_positions is "+strconv.FormatFloat(bad, 'g', -1, 64)) {
			t.Errorf("the production Guardian must refuse max_open_positions %v naming the written value, got %v", bad, err)
		}
	}
}
