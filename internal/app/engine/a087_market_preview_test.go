package engine_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA087EngineAndDefaultServiceRejectNonfractionalMarket(t *testing.T) {
	s := newE2EStack(t)
	defaultService := trading.NewService(config.Trading{}, nil)
	for _, market := range []struct{ market, currency string }{{"kr", "KRW"}, {"us", "USD"}} {
		for _, side := range []string{"buy", "sell"} {
			intent := orderintent.PlaceIntent{Symbol: "005930", Market: market.market, Side: side, OrderType: "market", Quantity: 1, CurrencyMode: market.currency}
			if got := defaultService.PreviewPlace(intent); got.LiveReady {
				t.Errorf("default %s %s market accepted", market.market, side)
			}
			if got := engine.A087EnginePreviewForTest(s.engine, intent); got.LiveReady {
				t.Errorf("engine %s %s market accepted", market.market, side)
			}
		}
	}
}
