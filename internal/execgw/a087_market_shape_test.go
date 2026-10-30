package execgw

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"testing"
)

func TestA087NonfractionalMarketShapeRejected(t *testing.T) {
	for _, market := range []struct{ market, currency string }{{"kr", "KRW"}, {"us", "USD"}} {
		for _, side := range []string{"buy", "sell"} {
			intent := orderintent.PlaceIntent{Symbol: "005930", Market: market.market, Side: side, OrderType: "market", Quantity: 1, CurrencyMode: market.currency}
			if got := checkOrderShape(intent); got == nil || got.Reason != ReasonUnsupportedOrderType {
				t.Errorf("%s %s market rejection = %v", market.market, side, got)
			}
		}
	}
}
