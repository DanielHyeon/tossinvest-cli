package engine_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/hybrid"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// spyTransport is the WTS spy. It is installed as the only transport available
// to both clients under test and routes by host: requests to the official
// httptest server are forwarded, everything else is recorded and refused
// without ever being dialled.
//
// A host-level spy rather than a fake WTS server, because the web-session client
// does not derive every URL from its configured base URLs (ensureTradingMetadata
// fetches https://www.tossinvest.com/account literally). Intercepting at the
// transport is therefore the only way to guarantee a test can neither reach Toss
// nor miss an attempt to.
type spyTransport struct {
	officialHost string
	inner        http.RoundTripper

	mu      sync.Mutex
	foreign []string
}

func (s *spyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == s.officialHost {
		return s.inner.RoundTrip(req)
	}
	s.mu.Lock()
	s.foreign = append(s.foreign, req.Method+" "+req.URL.Host+req.URL.Path)
	s.mu.Unlock()
	return nil, fmt.Errorf("spyTransport: blocked non-official request to %s", req.URL.Host)
}

func (s *spyTransport) calls() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.foreign...)
}

func (s *spyTransport) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.foreign = nil
}

// officialStub answers the endpoints an order mutation matrix touches and counts
// what it was asked for.
func officialStub(t *testing.T) (*httptest.Server, func() map[string]int) {
	t.Helper()
	var mu sync.Mutex
	seen := map[string]int{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen[r.Method+" "+r.URL.Path]++
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case path == "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case path == "/api/v1/orders" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"result":{"orderId":"O-1"}}`))
		case strings.HasPrefix(path, "/api/v1/orders/") && strings.HasSuffix(path, "/cancel"):
			_, _ = w.Write([]byte(`{"result":{"orderId":"O-1-cancel"}}`))
		case strings.HasPrefix(path, "/api/v1/orders/") && strings.HasSuffix(path, "/modify"):
			_, _ = w.Write([]byte(`{"result":{"orderId":"O-2"}}`))
		case path == "/api/v1/conditional-orders" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"result":{"conditionalOrderId":"CO-1","clientOrderId":"cid-1"}}`))
		case strings.HasPrefix(path, "/api/v1/conditional-orders/") && strings.HasSuffix(path, "/modify"):
			_, _ = w.Write([]byte(`{"result":{}}`))
		case strings.HasPrefix(path, "/api/v1/conditional-orders/") && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"result":{}}`))
		// The derived cancel/amend pre-check (task 4.1) reads the single order.
		case strings.HasPrefix(path, "/api/v1/orders/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"result":{"orderId":"O-1","status":"OPEN","quantity":"1","execution":{"filledQuantity":"0"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return srv, func() map[string]int {
		mu.Lock()
		defer mu.Unlock()
		out := map[string]int{}
		for k, v := range seen {
			out[k] = v
		}
		return out
	}
}

// TestEngineMutationMatrixNeverReachesWTS runs every order mutation the engine
// can perform — place, cancel, amend, and the three conditional-order verbs —
// and asserts the WTS spy recorded zero calls, while the official stub recorded
// all of them.
//
// The spy is proven live first, by driving the upstream hybrid broker (what the
// CLI uses when no official client exists) through the same transport and
// watching it try to reach Toss. Without that control arm, "zero WTS calls"
// could just mean the spy was never wired up.
func TestEngineMutationMatrixNeverReachesWTS(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	srv, officialCalls := officialStub(t)
	spy := &spyTransport{
		officialHost: strings.TrimPrefix(srv.URL, "http://"),
		inner:        srv.Client().Transport,
	}
	spyClient := &http.Client{Transport: spy}

	// --- control arm: the CLI's WTS path does reach out, so the spy works ----
	wts := tossclient.New(tossclient.Config{
		HTTPClient:  spyClient,
		APIBaseURL:  "http://wts.invalid",
		InfoBaseURL: "http://wts.invalid",
		CertBaseURL: "http://wts.invalid",
		Session:     &session.Session{Provider: "test", Cookies: map[string]string{"session": "x"}},
		TradingPolicy: config.Trading{
			Place: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		},
	})
	wtsBroker := hybrid.New(wts, nil, hybrid.Policy{Prefer: "wts"}, io.Discard).Broker()
	if _, err := wtsBroker.PlacePendingOrder(context.Background(), placeIntentFixture()); err == nil {
		t.Fatal("control arm: the WTS place must fail — the spy refuses every non-official host")
	}
	if len(spy.calls()) == 0 {
		t.Fatal("control arm: spy recorded nothing, so a later zero would prove nothing")
	}
	spy.reset()

	// --- engine arm ---------------------------------------------------------
	eng, err := engine.New(engine.Options{
		ConfigDir: dir,
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(spyClient),
		},
	})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}

	ctx := context.Background()
	svc := eng.TradingServiceForTest()

	place := placeIntentFixture()
	if _, err := svc.Place(ctx, place, executeOpts(svc.PreviewPlace(place).ConfirmToken)); err != nil {
		t.Errorf("place through engine: %v", err)
	}

	cancel := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	if _, err := svc.Cancel(ctx, cancel, executeOpts(svc.PreviewCancel(cancel).ConfirmToken)); err != nil {
		t.Errorf("cancel through engine: %v", err)
	}

	newPrice := 70500.0
	amend := orderintent.AmendIntent{OrderID: "O-1", Price: &newPrice}
	if _, err := svc.Amend(ctx, amend, executeOpts(svc.PreviewAmend(amend).ConfirmToken)); err != nil {
		t.Errorf("amend through engine: %v", err)
	}

	condPlace := orderintent.ConditionalPlaceIntent{
		Symbol: "005930", Type: "SINGLE", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 1,
		First: orderintent.ConditionLeg{OrderSide: "SELL", TriggerPrice: 71000, OrderPrice: 70900},
	}
	if _, err := eng.ConditionalForTest().CreateConditionalOrder(ctx, condPlace); err != nil {
		t.Errorf("conditional place through engine: %v", err)
	}
	condModify := orderintent.ConditionalModifyIntent{
		ID: "CO-1", Type: "SINGLE", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 1,
		First: orderintent.ConditionLeg{OrderSide: "SELL", TriggerPrice: 71500, OrderPrice: 71400},
	}
	if err := eng.ConditionalForTest().ModifyConditionalOrder(ctx, condModify); err != nil {
		t.Errorf("conditional modify through engine: %v", err)
	}
	if err := eng.ConditionalForTest().CancelConditionalOrder(ctx, orderintent.ConditionalCancelIntent{ID: "CO-1"}); err != nil {
		t.Errorf("conditional cancel through engine: %v", err)
	}

	if foreign := spy.calls(); len(foreign) != 0 {
		t.Errorf("engine reached %d non-official endpoint(s): %v", len(foreign), foreign)
	}

	calls := officialCalls()
	for _, want := range []string{
		"POST /api/v1/orders",
		"POST /api/v1/orders/O-1/cancel",
		"POST /api/v1/orders/O-1/modify",
		"POST /api/v1/conditional-orders",
		"POST /api/v1/conditional-orders/CO-1/modify",
		"DELETE /api/v1/conditional-orders/CO-1",
	} {
		if calls[want] == 0 {
			t.Errorf("official endpoint %q was never called; got %v", want, calls)
		}
	}
}

// TestEngineCancelWorksWithoutWTSSession pins engine-safety's requirement that
// cancel/amend do not depend on a web session: upstream's trading.Service asks
// the broker for available actions first, and the hybrid broker always answers
// that from WTS. The engine's broker must answer without one.
//
// Since task 4.1 the answer is derived from the official single-order read
// (precheck.go), so this test asserts both halves of that: the pre-check produces
// a real derived state, and it does so without touching any non-official host.
func TestEngineCancelWorksWithoutWTSSession(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	srv, officialCalls := officialStub(t)
	spy := &spyTransport{officialHost: strings.TrimPrefix(srv.URL, "http://"), inner: srv.Client().Transport}

	eng, err := engine.New(engine.Options{
		ConfigDir: dir,
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(&http.Client{Transport: spy}),
		},
	})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}

	actions, err := eng.BrokerForTest().GetOrderAvailableActions(context.Background(), "O-1")
	if err != nil {
		t.Fatalf("pre-check must not fail without a WTS session: %v", err)
	}
	if got := actions[engine.ActionKeyChecked]; got != true {
		t.Errorf("%s = %v, want true", engine.ActionKeyChecked, got)
	}
	if got := actions[engine.ActionKeyState]; got != "OPEN_UNFILLED" {
		t.Errorf("%s = %v, want OPEN_UNFILLED", engine.ActionKeyState, got)
	}
	if calls := officialCalls(); calls["GET /api/v1/orders/O-1"] != 1 {
		t.Errorf("want one official single-order read, got %v", calls)
	}
	if foreign := spy.calls(); len(foreign) != 0 {
		t.Errorf("pre-check reached %v", foreign)
	}
}

func placeIntentFixture() orderintent.PlaceIntent {
	return orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 70000, CurrencyMode: "KRW",
	}
}

func executeOpts(token string) trading.ExecuteOptions {
	return trading.ExecuteOptions{Execute: true, Confirm: token}
}
