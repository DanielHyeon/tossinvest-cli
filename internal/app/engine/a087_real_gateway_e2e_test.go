package engine_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

func TestA087RealGatewayLowerLimitAndDormantFallback(t *testing.T) {
	for _, tc := range []struct {
		name    string
		inject  bool
		outside bool
		want    string
	}{
		{"lower limit", true, false, "50400"},
		{"dormant fallback", false, false, "67900"},
		{"outside session", true, true, "67900"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newE2EStack(t)
			guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: s.engine.Journal, Clock: s.clk, AccountRef: s.engine.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a087-e2e"})
			if err != nil {
				t.Fatal(err)
			}
			opts := engine.ExitObserverOptions{Journal: s.engine.Journal, Prices: e2eClockDomainPrices{next: s.engine.Official, clk: s.clk}, Retrier: s.engine.Retrier, Issuer: guardian, Submit: s.engine.Gateway, Alerts: s.engine.Notifier, Costs: costs.DefaultModel(), AccountRef: s.engine.AccountRef, Clock: s.clk}
			if tc.inject {
				engine.SetA087LowerLimitForTest(&opts, "kr", "005930", "50400", s.clk.Now())
			}
			s.observer, err = engine.NewExitObserver(opts)
			if err != nil {
				t.Fatal(err)
			}
			s.entry("005930", "10", "70000", "68000")
			s.broker.quote("005930", "70100")
			if cycle := s.observe(); cycle.Opened != 1 {
				t.Fatalf("opened = %d", cycle.Opened)
			}
			if tc.outside {
				s.clk.Advance(6*time.Hour + 30*time.Minute)
			}
			s.broker.quote("005930", "67900")
			if cycle := s.observe(); cycle.Proposed != 1 {
				t.Fatalf("proposed = %d, cycle = %+v", cycle.Proposed, cycle)
			}
			orders := s.broker.placed()
			if len(orders) != 1 {
				t.Fatalf("orders = %v", orders)
			}
			if id, ok := orders[0]["clientOrderId"].(string); !ok || id == "" {
				t.Fatalf("missing clientOrderId in %v", orders[0])
			}
			delete(orders[0], "clientOrderId")
			want := map[string]any{"symbol": "005930", "side": "SELL", "orderType": "LIMIT", "quantity": "10", "price": tc.want, "timeInForce": "DAY", "confirmHighValueOrder": false}
			if !reflect.DeepEqual(orders[0], want) {
				t.Fatalf("wire order = %v, want %v", orders[0], want)
			}
		})
	}
}
