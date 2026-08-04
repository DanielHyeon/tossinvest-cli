//go:build tossos_testseams

package risk

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

func TestTestOnlySealAccountBaseFXPairedScope(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	policy := accountBaseTestPolicy()
	for _, test := range []struct {
		name, rate, haircut, quote string
		market                     costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, quote: "KRW", rate: "1", haircut: "1"},
		{name: "US official", market: costs.MarketUS, quote: "USD", rate: "1400", haircut: "1.01"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx, err := TestOnlySealAccountBaseFX(now, test.market, policy, test.rate, test.haircut)
			if err != nil {
				t.Fatal(err)
			}
			if fx.QuoteCurrency() != test.quote || fx.AccountCurrency() != "KRW" || fx.Digest() == "" || !fx.EvaluatedAt().Equal(now) {
				t.Fatalf("test seam scope = quote=%q base=%q digest=%q at=%s", fx.QuoteCurrency(), fx.AccountCurrency(), fx.Digest(), fx.EvaluatedAt())
			}
		})
	}

	if _, err := TestOnlySealAccountBaseFX(now, costs.MarketUS, policy, "1400", "0.99"); !errors.Is(err, ErrAccountBaseFXUnavailable) {
		t.Fatalf("unsafe haircut error = %v, want ErrAccountBaseFXUnavailable", err)
	}
}
