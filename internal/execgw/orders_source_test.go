package execgw_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// End-to-end pagination tests (harden-execution-base task 2.9).
//
// The fake-pager tests in indoubt_test.go pin the loop's logic; these pin that the
// loop, the official client and a real HTTP server agree — that the cursor the
// broker sends is the cursor the next request carries, and that the two defences
// fire against a server that misbehaves rather than against a mock that was told to.

// ordersServer serves a scripted /api/v1/orders and counts requests.
type ordersServer struct {
	t *testing.T

	mu      sync.Mutex
	cursors []string
	calls   int
	handler func(cursor string, call int) string
}

func newOrdersServer(t *testing.T, handler func(cursor string, call int) string) (*official.Client, *ordersServer) {
	t.Helper()
	s := &ordersServer{t: t, handler: handler}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		if r.URL.Path != "/api/v1/orders" {
			http.NotFound(w, r)
			return
		}
		s.mu.Lock()
		cursor := r.URL.Query().Get("cursor")
		s.cursors = append(s.cursors, cursor)
		s.calls++
		call := s.calls
		s.mu.Unlock()
		_, _ = w.Write([]byte(s.handler(cursor, call)))
	}))
	t.Cleanup(srv.Close)

	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))
	return client, s
}

func (s *ordersServer) seen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.cursors...)
}

func pageBody(orderID, canceledAt, next string) string {
	canceled := "null"
	if canceledAt != "" {
		canceled = fmt.Sprintf("%q", canceledAt)
	}
	tail := `"nextCursor":null,"hasNext":false`
	if next != "" {
		tail = fmt.Sprintf(`"nextCursor":%q,"hasNext":true`, next)
	}
	return fmt.Sprintf(`{"result":{"orders":[{"orderId":%q,"symbol":"005930","side":"BUY",`+
		`"status":"CLOSED","quantity":"2","price":"70000","currency":"KRW",`+
		`"orderedAt":"2026-03-30T10:30:00+09:00","canceledAt":%s,`+
		`"execution":{"filledQuantity":"0"}}],%s}}`, orderID, canceled, tail)
}

// TestScanOrdersWalksARealPaginatedEndpoint: three pages, each cursor followed
// exactly once, every order returned.
func TestScanOrdersWalksARealPaginatedEndpoint(t *testing.T) {
	client, srv := newOrdersServer(t, func(cursor string, _ int) string {
		switch cursor {
		case "":
			return pageBody("O-1", "", "c2")
		case "c2":
			return pageBody("O-2", "", "c3")
		case "c3":
			return pageBody("O-3", "2026-03-30T10:35:00+09:00", "")
		default:
			return `{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`
		}
	})

	orders, err := execgw.ScanOrders(context.Background(), execgw.OfficialOrders{Client: client},
		execgw.OrderQuery{Status: "CLOSED", Symbol: "005930"}, 10)
	if err != nil {
		t.Fatalf("ScanOrders: %v", err)
	}
	if len(orders) != 3 {
		t.Fatalf("orders: got %d, want 3", len(orders))
	}
	if want := []string{"", "c2", "c3"}; !equalStrings(srv.seen(), want) {
		t.Errorf("cursors sent: got %v, want %v", srv.seen(), want)
	}

	// The whole reason for the raw path: canceledAt survives to the derivation.
	derived := brokerstate.DeriveOfficialOrder(orders[2], brokerstate.Lineage{})
	if derived.State != brokerstate.StateCancelled {
		t.Errorf("derived state: got %s (%s) — canceledAt did not survive the raw page",
			derived.State, derived.Reason)
	}
}

// TestScanOrdersRefusesARepeatingCursor: a broker that hands back the same cursor
// would otherwise be an infinite loop inside a safety procedure.
func TestScanOrdersRefusesARepeatingCursor(t *testing.T) {
	client, srv := newOrdersServer(t, func(string, int) string {
		return pageBody("O-loop", "", "stuck")
	})

	_, err := execgw.ScanOrders(context.Background(), execgw.OfficialOrders{Client: client},
		execgw.OrderQuery{Status: "OPEN"}, 50)
	if !errors.Is(err, execgw.ErrCursorLoop) {
		t.Fatalf("want ErrCursorLoop, got %v", err)
	}
	if calls := len(srv.seen()); calls > 3 {
		t.Errorf("the loop was detected only after %d requests; it must stop on the first repeat", calls)
	}
}

// TestScanOrdersStopsAtThePageCeiling: an endlessly paginating list fails closed
// rather than returning a partial answer that would read as "absent".
func TestScanOrdersStopsAtThePageCeiling(t *testing.T) {
	client, srv := newOrdersServer(t, func(_ string, call int) string {
		return pageBody(fmt.Sprintf("O-%d", call), "", fmt.Sprintf("c%d", call+1))
	})

	orders, err := execgw.ScanOrders(context.Background(), execgw.OfficialOrders{Client: client},
		execgw.OrderQuery{Status: "OPEN"}, 4)
	if !errors.Is(err, execgw.ErrTooManyPages) {
		t.Fatalf("want ErrTooManyPages, got %v", err)
	}
	if orders != nil {
		t.Error("a truncated walk must return no orders at all — a partial list reads as absence")
	}
	if calls := len(srv.seen()); calls != 4 {
		t.Errorf("requests: got %d, want the 4-page ceiling", calls)
	}
}

// TestScanOrdersPropagatesBrokerErrors keeps the official sentinels intact through
// the adapter, because the retry matrix classifies on them.
func TestScanOrdersPropagatesBrokerErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))

	_, err := execgw.ScanOrders(context.Background(), execgw.OfficialOrders{Client: client},
		execgw.OrderQuery{Status: "OPEN"}, 10)
	if !errors.Is(err, official.ErrRateLimited) {
		t.Errorf("want official.ErrRateLimited to survive the adapter, got %v", err)
	}
	if execgw.ClassifyQueryError(err) != execgw.ClassRateLimited {
		t.Errorf("the retry matrix must still classify it: got %q", execgw.ClassifyQueryError(err))
	}
}

// TestOfficialAccountReadsHoldingsAndBuyingPower covers the absence cross-check's
// data source, including the "you own none of it" case.
func TestOfficialAccountReadsHoldingsAndBuyingPower(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/buying-power":
			if got := r.URL.Query().Get("currency"); got != "KRW" {
				t.Errorf("currency: got %q", got)
			}
			_, _ = w.Write([]byte(`{"result":{"cashBuyingPower":"1234567.89","currency":"KRW"}}`))
		case r.URL.Path == "/api/v1/holdings" && r.URL.Query().Get("symbol") == "005930":
			_, _ = w.Write([]byte(`{"result":{"items":[{"symbol":"005930","quantity":"7"}]}}`))
		case r.URL.Path == "/api/v1/holdings":
			_, _ = w.Write([]byte(`{"result":{"items":[]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))
	account := execgw.OfficialAccount{Client: client}
	ctx := context.Background()

	bp, err := account.BuyingPower(ctx, "KRW")
	if err != nil {
		t.Fatalf("BuyingPower: %v", err)
	}
	if bp != 1234567.89 {
		t.Errorf("buying power: got %v", bp)
	}

	held, err := account.HoldingQuantity(ctx, "005930")
	if err != nil {
		t.Fatalf("HoldingQuantity: %v", err)
	}
	if held != 7 {
		t.Errorf("holding: got %v, want 7", held)
	}

	none, err := account.HoldingQuantity(ctx, "000660")
	if err != nil {
		t.Fatalf("HoldingQuantity(unheld): %v — an unheld symbol is zero, not an error", err)
	}
	if none != 0 {
		t.Errorf("unheld symbol: got %v, want 0", none)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}
