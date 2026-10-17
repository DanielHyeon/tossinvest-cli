package risk

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// a146 사람 결정 1(2026-10-18): 「Guardian 상한 0」에는 주문 금액(notional) 상한으로 1주도 안 되는 진입도 든다 — 예산은 넉넉하고 한 주 가격(× FX)이
// 주문 금액 상한을 넘으면 Guardian 짝 사이징은 위험 예산 상한 0 과 같은 sentinel 로 거절함(엔진은 그 sentinel 로 스킵을 가름). 경계 한 단위도 잼.
func TestA146ANotionalCapBelowOneUnitIsTheSameZeroSentinel(t *testing.T) {
	now := time.Date(2026, 10, 18, 2, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name, quote, rate, haircut, below, exact string
		market                                   costs.Market
	}{
		// KR identity: 한 주 10 KRW.
		{name: "KR identity", market: costs.MarketKR, quote: "KRW", rate: "1", haircut: "1", below: "9", exact: "10"},
		// US official: 한 주 10 USD × 2 × 1.1 = 22 KRW.
		{name: "US official FX", market: costs.MarketUS, quote: "USD", rate: "2", haircut: "1.1", below: "21", exact: "22"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			policy := accountBaseTestPolicy()
			policy.RiskBudget.Amount = "1000000" // 위험 상한은 넉넉함 — 묶는 것은 notional 뿐
			policy.MaxOrderNotional.Amount = test.below
			if _, err := AccountBaseStrategyEntryQuantity(policy, test.market, "10", "9", fx); !errors.Is(err, ErrStrategyQuantityZero) {
				t.Fatalf("notional %s below one unit: err=%v — want ErrStrategyQuantityZero", test.below, err)
			}
			policy.MaxOrderNotional.Amount = test.exact
			if got, err := AccountBaseStrategyEntryQuantity(policy, test.market, "10", "9", fx); err != nil || got != "1" {
				t.Fatalf("notional %s exactly one unit: got=%s err=%v — want 1", test.exact, got, err)
			}
		})
	}
}
