package official

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for the additive raw-order reads (harden-execution-base task 2.9,
// issues.md Manager decision (c)).
//
// The point of these methods is what the adapted ones lose: Orders() keeps only
// the first page and adaptOrder() drops canceledAt, and the engine's IN_DOUBT
// resolution needs both. So these tests check the pagination envelope survives and
// the payload arrives byte-for-byte.

const rawOrderPage1 = `{"result":{"orders":[
  {"orderId":"O-1","symbol":"005930","side":"BUY","orderType":"LIMIT","status":"OPEN",
   "quantity":"10","price":"70000","currency":"KRW","orderedAt":"2026-03-29T09:30:00+09:00",
   "canceledAt":null,"execution":{"filledQuantity":"0"}}
],"nextCursor":"cursor-2","hasNext":true}}`

const rawOrderPage2 = `{"result":{"orders":[
  {"orderId":"O-2","symbol":"005930","side":"BUY","orderType":"LIMIT","status":"CLOSED",
   "quantity":"5","price":"69000","currency":"KRW","orderedAt":"2026-03-29T09:31:00+09:00",
   "canceledAt":"2026-03-29T09:35:00+09:00","execution":{"filledQuantity":"2"}}
],"nextCursor":null,"hasNext":false}}`

func rawTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithAccountSeq(3),
	)
}

// TestOrdersPageRawKeepsTheCursorAndTheRawPayload is the whole reason the method
// exists: Orders() discards nextCursor/hasNext and adaptOrder discards canceledAt.
func TestOrdersPageRawKeepsTheCursorAndTheRawPayload(t *testing.T) {
	var gotHeader, gotCursor, gotStatus string
	c := rawTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orders" {
			http.NotFound(w, r)
			return
		}
		gotHeader = r.Header.Get("X-Tossinvest-Account")
		gotCursor = r.URL.Query().Get("cursor")
		gotStatus = r.URL.Query().Get("status")
		if gotCursor == "cursor-2" {
			_, _ = w.Write([]byte(rawOrderPage2))
			return
		}
		_, _ = w.Write([]byte(rawOrderPage1))
	})

	page, err := c.OrdersPageRaw(context.Background(), OrdersFilter{Status: "OPEN", Symbol: "005930"}, "")
	if err != nil {
		t.Fatalf("OrdersPageRaw: %v", err)
	}
	if gotHeader != "3" {
		t.Errorf("X-Tossinvest-Account: got %q, want 3 — the account-scoped path must be reused", gotHeader)
	}
	if gotStatus != "OPEN" {
		t.Errorf("status param: got %q, want OPEN", gotStatus)
	}
	if len(page.Orders) != 1 {
		t.Fatalf("orders: got %d, want 1", len(page.Orders))
	}
	if !page.HasNext || page.NextCursor != "cursor-2" {
		t.Errorf("pagination envelope lost: hasNext=%v cursor=%q", page.HasNext, page.NextCursor)
	}
	if !strings.Contains(string(page.Orders[0]), `"orderId":"O-1"`) {
		t.Errorf("raw payload: %s", page.Orders[0])
	}

	// Second page, driven by the cursor argument.
	page2, err := c.OrdersPageRaw(context.Background(), OrdersFilter{Status: "OPEN"}, "cursor-2")
	if err != nil {
		t.Fatalf("OrdersPageRaw(page 2): %v", err)
	}
	if gotCursor != "cursor-2" {
		t.Errorf("cursor param: got %q, want cursor-2", gotCursor)
	}
	if page2.HasNext || page2.NextCursor != "" {
		t.Errorf("last page must report no successor: hasNext=%v cursor=%q", page2.HasNext, page2.NextCursor)
	}
	// canceledAt is the field the adapted path drops; it must be here.
	if !strings.Contains(string(page2.Orders[0]), `"canceledAt":"2026-03-29T09:35:00+09:00"`) {
		t.Errorf("canceledAt did not survive: %s", page2.Orders[0])
	}
}

// TestOrdersPageRawCursorArgumentOverridesTheFilter keeps the two ways of passing
// a cursor from disagreeing silently.
func TestOrdersPageRawCursorArgumentOverridesTheFilter(t *testing.T) {
	var gotCursor string
	c := rawTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCursor = r.URL.Query().Get("cursor")
		_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`))
	})

	if _, err := c.OrdersPageRaw(context.Background(),
		OrdersFilter{Cursor: "from-filter"}, "from-argument"); err != nil {
		t.Fatalf("OrdersPageRaw: %v", err)
	}
	if gotCursor != "from-argument" {
		t.Errorf("cursor: got %q, want the explicit argument to win", gotCursor)
	}
}

// TestOrdersPageRawPassesEveryFilter mirrors what Orders() sends, so the two paths
// cannot drift into filtering differently.
func TestOrdersPageRawPassesEveryFilter(t *testing.T) {
	var got map[string]string
	c := rawTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = map[string]string{}
		for _, k := range []string{"status", "symbol", "from", "to", "cursor", "limit"} {
			got[k] = r.URL.Query().Get(k)
		}
		_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`))
	})

	if _, err := c.OrdersPageRaw(context.Background(), OrdersFilter{
		Status: "CLOSED", Symbol: "005930", From: "2026-03-01", To: "2026-03-31", Limit: 50,
	}, "cur"); err != nil {
		t.Fatalf("OrdersPageRaw: %v", err)
	}
	want := map[string]string{
		"status": "CLOSED", "symbol": "005930", "from": "2026-03-01",
		"to": "2026-03-31", "cursor": "cur", "limit": "50",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("param %q: got %q, want %q", k, got[k], v)
		}
	}
}

// TestOrderRawByIDReturnsTheUnwrappedOrder: the single-order read the IN_DOUBT
// resolution uses for cancel/amend, where canceledAt decides the outcome.
func TestOrderRawByIDReturnsTheUnwrappedOrder(t *testing.T) {
	var gotHeader, gotPath string
	c := rawTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Tossinvest-Account")
		// EscapedPath, not Path: the server decodes %2F back into a slash, so the
		// decoded form cannot show whether the id was escaped at all.
		gotPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"result":{"orderId":"O 1/x","symbol":"005930","status":"CLOSED",` +
			`"quantity":"2","canceledAt":"2026-03-29T09:35:00+09:00","execution":{"filledQuantity":"1"}}}`))
	})

	raw, err := c.OrderRawByID(context.Background(), "O 1/x")
	if err != nil {
		t.Fatalf("OrderRawByID: %v", err)
	}
	if gotHeader != "3" {
		t.Errorf("X-Tossinvest-Account: got %q, want 3", gotHeader)
	}
	if gotPath != "/api/v1/orders/O%201%2Fx" {
		t.Errorf("path escaping: got %q", gotPath)
	}
	if !strings.Contains(string(raw), `"canceledAt":"2026-03-29T09:35:00+09:00"`) {
		t.Errorf("raw order: %s", raw)
	}
	if strings.Contains(string(raw), `"result"`) {
		t.Errorf("the envelope must be unwrapped: %s", raw)
	}
}

// TestRawReadsClassifyErrorsLikeEveryOtherRead keeps the new methods on the shared
// error contract — the retry matrix classifies by those sentinels.
func TestRawReadsClassifyErrorsLikeEveryOtherRead(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		check  func(error) bool
	}{
		// Only the auth row needed widening. a082 wraps that sentinel to carry its
		// status code, which == rejects and every production consumer accepts.
		// The other two rows keep == on purpose: identity there also asserts that
		// nothing has started attaching a response body to them, and errors.Is
		// would admit exactly that.
		{"rate limited", http.StatusTooManyRequests, `{}`, func(err error) bool { return err == ErrRateLimited }},
		{"auth", http.StatusUnauthorized, `{"message":"bad key"}`, func(err error) bool { return errors.Is(err, ErrAuth) }},
		{"server", http.StatusInternalServerError, `{}`, func(err error) bool { return err == ErrServer }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := rawTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := c.OrdersPageRaw(context.Background(), OrdersFilter{}, "")
			if err == nil || !tc.check(err) {
				t.Errorf("got %v", err)
			}
		})
	}
}
