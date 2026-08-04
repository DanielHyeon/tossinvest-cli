package execgw_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

type boundaryProvider struct {
	snapshot protectionreadiness.ReadinessSnapshot
	err      error
	calls    int
}

func TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker(t *testing.T) {
	for _, quantity := range []float64{1.5} {
		broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "unexpected"}}
		provider := &boundaryProvider{snapshot: protectionreadiness.DefaultSnapshot()}
		adapter, err := protection.NewReadinessAdapter(provider, "acct-7", "production")
		if err != nil {
			t.Fatal(err)
		}
		gw, j, clk := newGatewayWithReadiness(t, broker, adapter)
		intent := placeIntent()
		intent.Quantity = quantity
		_, _ = gw.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: entryDecision(t, j, clk, intent, testLimits())})
		if provider.calls != 0 {
			t.Fatalf("quantity %v reached readiness provider calls=%d", quantity, provider.calls)
		}
		if places, _, _ := broker.totals(); places != 0 {
			t.Fatalf("quantity %v reached broker calls=%d", quantity, places)
		}
	}
}

func (provider *boundaryProvider) Current(context.Context) (protectionreadiness.ReadinessSnapshot, error) {
	provider.calls++
	return provider.snapshot, provider.err
}

func TestProductionDefaultRefusesKRAndUSBuyBeforeBroker(t *testing.T) {
	for _, market := range []string{"kr", "us"} {
		t.Run(market, func(t *testing.T) {
			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "unexpected"}}
			provider := &boundaryProvider{snapshot: protectionreadiness.DefaultSnapshot()}
			adapter, err := protection.NewReadinessAdapter(provider, "acct-7", "production")
			if err != nil {
				t.Fatal(err)
			}
			gw, j, clk := newGatewayWithReadiness(t, broker, adapter)
			intent := placeIntent()
			intent.Market = market
			if market == "us" {
				intent.Symbol, intent.CurrencyMode = "AAPL", "USD"
			}
			_, _ = gw.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: entryDecision(t, j, clk, intent, testLimits())})
			if places, _, _ := broker.totals(); places != 0 {
				t.Fatalf("%s buy broker calls=%d", market, places)
			}
		})
	}
}

func TestReductionNeverReadsReadinessProvider(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *execgw.Gateway, *journal.Journal, *clock.Fake) (execgw.Outcome, error)
		want [3]int
	}{
		{
			name: "sell place",
			run: func(t *testing.T, gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				sell := orderintent.PlaceIntent{Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit", Quantity: 1, Price: 70000, CurrencyMode: "KRW"}
				return gw.Place(context.Background(), execgw.PlaceRequest{Intent: sell, Decision: exitDecision(t, j, clk, journal.KindPlace, sell.Market, sell.Symbol, sell.Side, sell.Quantity)})
			},
			want: [3]int{1, 0, 0},
		},
		{
			name: "cancel",
			run: func(t *testing.T, gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
				return gw.Cancel(context.Background(), execgw.CancelRequest{
					Intent:   intent,
					Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
					Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
				})
			},
			want: [3]int{0, 1, 0},
		},
		{
			name: "reducing amend",
			run: func(t *testing.T, gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				quantity := 1.0
				return gw.Amend(context.Background(), execgw.AmendRequest{
					Intent:   orderintent.AmendIntent{OrderID: "O-1", Quantity: &quantity},
					Symbol:   "005930",
					Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
					Decision: exitDecision(t, j, clk, journal.KindAmend, "kr", "005930", "BUY", 2),
				})
			},
			want: [3]int{0, 0, 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			broker := &fakeBroker{result: domain.MutationResult{Status: "accepted", OrderID: "reduction-1", CurrentOrderID: "reduction-1"}}
			provider := &boundaryProvider{err: errors.New("must not be read")}
			adapter, err := protection.NewReadinessAdapter(provider, "acct-7", "production")
			if err != nil {
				t.Fatal(err)
			}
			gw, j, clk := newGatewayWithReadiness(t, broker, adapter)
			out, err := test.run(t, gw, j, clk)
			if err != nil || out.State != journal.StateConfirmed {
				t.Fatalf("reduction blocked out=%+v err=%v", out, err)
			}
			if provider.calls != 0 {
				t.Fatalf("reduction waited on provider calls=%d", provider.calls)
			}
			places, cancels, amends := broker.totals()
			if got := [3]int{places, cancels, amends}; got != test.want {
				t.Fatalf("broker calls=%v, want %v", got, test.want)
			}
		})
	}
}

func TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "us-1"}}
	gw, j, clk := newGatewayWithMarketFixture(t, broker, func(market string, check int) (bool, string) {
		if market == "kr" {
			return false, "kr-build-drift"
		}
		return true, "us-generation-7"
	})
	kr := placeIntent()
	_, _ = gw.Place(context.Background(), execgw.PlaceRequest{Intent: kr, Decision: entryDecision(t, j, clk, kr, testLimits())})
	if places, _, _ := broker.totals(); places != 0 {
		t.Fatalf("KR drift reached broker calls=%d", places)
	}

	us := placeIntent()
	us.Market, us.Symbol, us.CurrencyMode = "us", "AAPL", "USD"
	usLimits := testLimits()
	usLimits.Currency = "USD"
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{Intent: us, Decision: entryDecision(t, j, clk, us, usLimits)})
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("valid US blocked out=%+v err=%v", out, err)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Fatalf("US broker calls=%d", places)
	}

	driftBroker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "never"}}
	drift, dj, dc := newGatewayWithMarketFixture(t, driftBroker, func(_ string, check int) (bool, string) {
		if check == 1 {
			return true, "generation-1"
		}
		return true, "generation-2"
	})
	request := placeRequest(t, dj, dc)
	_, _ = drift.Place(context.Background(), request)
	if places, _, _ := driftBroker.totals(); places != 0 {
		t.Fatalf("dispatch snapshot drift reached broker calls=%d", places)
	}
}

func newGatewayWithReadiness(t *testing.T, broker trading.Broker, adapter *protection.ReadinessAdapter) (*execgw.Gateway, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	opts := execgw.Options{Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk, AccountRef: "acct-7", Source: "test", ProtectionReadiness: adapter}
	opts.UseReadinessAdapterForTest()
	gw, err := execgw.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	return gw, j, clk
}

func newGatewayWithMarketFixture(t *testing.T, broker trading.Broker, fixture func(market string, check int) (bool, string)) (*execgw.Gateway, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	opts := execgw.Options{Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk, AccountRef: "acct-7", Source: "test"}
	opts.SetMarketProtectionForTest(fixture)
	gw, err := execgw.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	return gw, j, clk
}
