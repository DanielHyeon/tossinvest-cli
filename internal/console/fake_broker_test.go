package console

// fake_broker_test.go is the account internal/console's tests run against.
//
// It exists so the approval tests can assert the thing that actually matters:
// that a request with a missing session, a wrong CSRF token or a wrong nonce
// results in **zero mutating broker calls**. Counting HTTP verbs would prove the
// same thing one layer down, but internal/verifylive already does that against
// its own httptest server; here the claim is about the console's gates, and the
// cheapest honest way to watch them is a broker that counts.
//
// The reads are as plausible as they need to be for verifylive to build a plan:
// one account, no holdings (so the sell-side and conditional steps skip with a
// reason), one quote, one price band.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	trading "github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// TestMain installs internal/testenv's guard for the whole package.
//
// The console's production code makes no outbound request at all, but its tests
// build a verifylive.Runner, and that package exists to place live orders. A test
// here that somehow reached a Toss hostname panics instead of trading — WORKFLOW's
// 불변 규칙: "문구가 아니라 테스트 인프라가 막는다".
func TestMain(m *testing.M) {
	previous := http.DefaultTransport
	http.DefaultTransport = &testenv.Guard{Base: previous, OnBlock: func(err *testenv.ErrRealHost) {
		panic("console tests: " + err.Error())
	}}
	code := m.Run()
	http.DefaultTransport = previous
	os.Exit(code)
}

// fakeBroker is a minimal Open API that counts what it was asked to change.
type fakeBroker struct {
	mu sync.Mutex
	// mutations counts every call that could change the account. It is the number
	// the refusal tests assert is zero.
	mutations int
	// placed records the (symbol, side, quantity) of each placement, so the
	// approved-flow test can check the runner sent what the plan authorised.
	placed []placedOrder
	seq    int
}

type placedOrder struct {
	Symbol   string
	Side     string
	Quantity float64
}

func newFakeBroker() *fakeBroker { return &fakeBroker{} }

func (f *fakeBroker) mutationCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mutations
}

func (f *fakeBroker) placements() []placedOrder {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]placedOrder(nil), f.placed...)
}

// --- reads ---------------------------------------------------------------------

func (f *fakeBroker) Accounts(context.Context) ([]domain.Account, error) {
	return []domain.Account{{ID: "1", DisplayName: "123-45-678901"}}, nil
}

// usProbeSymbol is the US holding this fake account carries, standing in for the
// real one (MWG, 115 shares at $1.27). The US probes need a symbol the account
// already owns, so the fake owns one.
const usProbeSymbol = "MWG"

func (f *fakeBroker) Holdings(context.Context, string) ([]domain.Position, error) {
	return []domain.Position{{Symbol: usProbeSymbol, Quantity: 115}}, nil
}

func (f *fakeBroker) SellableQuantity(_ context.Context, symbol string) (domain.SellableQuantity, error) {
	if symbol == usProbeSymbol {
		return domain.SellableQuantity{Symbol: symbol, Quantity: 115}, nil
	}
	return domain.SellableQuantity{}, errors.New("nothing held")
}

func (f *fakeBroker) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	out := make([]domain.Quote, 0, len(symbols))
	for _, s := range symbols {
		if s == usProbeSymbol {
			out = append(out, domain.Quote{Symbol: s, Last: 1.27, Currency: "USD"})
			continue
		}
		out = append(out, domain.Quote{Symbol: s, Last: 70000, Currency: "KRW"})
	}
	return out, nil
}

func (f *fakeBroker) PriceLimits(_ context.Context, symbol string) (domain.PriceLimits, error) {
	if symbol == usProbeSymbol {
		// US has no daily band; the endpoint returns null for both bounds.
		return domain.PriceLimits{Symbol: symbol}, nil
	}
	return domain.PriceLimits{Symbol: "005930", UpperLimit: 91000, LowerLimit: 49000}, nil
}

func (f *fakeBroker) OrdersPageRaw(context.Context, official.OrdersFilter, string) (official.RawOrderPage, error) {
	return official.RawOrderPage{}, nil
}

func (f *fakeBroker) OrderRawByID(_ context.Context, id string) (json.RawMessage, error) {
	return json.RawMessage(`{"orderId":"` + id + `","status":"CANCELED"}`), nil
}

func (f *fakeBroker) ConditionalOrders(context.Context, string, string, string, int) (domain.ConditionalOrderList, error) {
	return domain.ConditionalOrderList{}, nil
}

func (f *fakeBroker) ConditionalOrder(context.Context, string) (domain.ConditionalOrder, error) {
	return domain.ConditionalOrder{}, errors.New("no such conditional order")
}

// --- the mutation surface ---------------------------------------------------------

func (f *fakeBroker) PlaceOrder(_ context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	f.seq++
	f.placed = append(f.placed, placedOrder{
		Symbol: intent.Symbol, Side: intent.Side, Quantity: intent.Quantity,
	})
	return trading.MutationResult{OrderID: orderIDFor(f.seq)}, nil
}

func (f *fakeBroker) CancelOrder(_ context.Context, orderID string) (trading.MutationResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	return trading.MutationResult{CurrentOrderID: orderID}, nil
}

func (f *fakeBroker) ModifyOrder(_ context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	return trading.MutationResult{CurrentOrderID: intent.OrderID}, nil
}

func (f *fakeBroker) CreateConditionalOrder(context.Context, official.ConditionalCreateBody) (domain.ConditionalOrderRef, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	return domain.ConditionalOrderRef{ID: "co-1"}, nil
}

func (f *fakeBroker) ModifyConditionalOrderRef(context.Context, string, official.ConditionalModifyBody) (domain.ConditionalOrderRef, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	return domain.ConditionalOrderRef{ID: "co-2"}, nil
}

func (f *fakeBroker) CancelConditionalOrder(context.Context, string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutations++
	return nil
}

func orderIDFor(n int) string { return "ord-" + string(rune('a'+n%26)) }
