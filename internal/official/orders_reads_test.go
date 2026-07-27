package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

// TestAdaptOrderUnit verifies the pure adapter for a single Order.
func TestAdaptOrderUnit(t *testing.T) {
	raw := apiOrder{
		OrderID:   "abc123",
		Symbol:    "005930",
		Side:      "BUY",
		OrderType: "LIMIT",
		Status:    "OPEN",
		Quantity:  "10",
		Price:     "70000",
		Currency:  "KRW",
		OrderedAt: "2026-03-29T09:30:00+09:00",
		Execution: apiOrderExecution{
			FilledQuantity:     "0",
			AverageFilledPrice: "",
		},
	}

	got := adaptOrder(raw)

	if got.ID != "abc123" {
		t.Fatalf("ID: want abc123, got %q", got.ID)
	}
	if got.Symbol != "005930" {
		t.Fatalf("Symbol: want 005930, got %q", got.Symbol)
	}
	if got.Side != "BUY" {
		t.Fatalf("Side: want BUY, got %q", got.Side)
	}
	if got.Status != "OPEN" {
		t.Fatalf("Status: want OPEN, got %q", got.Status)
	}
	if got.Quantity != 10 {
		t.Fatalf("Quantity: want 10, got %v", got.Quantity)
	}
	if got.Price != 70000 {
		t.Fatalf("Price: want 70000, got %v", got.Price)
	}
	if got.FilledQuantity != 0 {
		t.Fatalf("FilledQuantity: want 0, got %v", got.FilledQuantity)
	}
	if got.AverageExecutionPrice != 0 {
		t.Fatalf("AverageExecutionPrice: want 0 (null), got %v", got.AverageExecutionPrice)
	}
	if got.OrderDate != "2026-03-29" {
		t.Fatalf("OrderDate: want 2026-03-29, got %q", got.OrderDate)
	}
	if got.SubmittedAt == nil {
		t.Fatal("SubmittedAt: expected non-nil for valid orderedAt")
	}
	// 09:30 KST = 00:30 UTC on same calendar day.
	wantUTC := time.Date(2026, 3, 29, 0, 30, 0, 0, time.UTC)
	if !got.SubmittedAt.UTC().Equal(wantUTC) {
		t.Fatalf("SubmittedAt UTC: want %v, got %v", wantUTC, got.SubmittedAt.UTC())
	}
	// Fields not in /orders response.
	if got.Name != "" {
		t.Fatalf("Name: expected empty, got %q", got.Name)
	}
	if got.Market != "" {
		t.Fatalf("Market: expected empty, got %q", got.Market)
	}
}

// TestAdaptOrderFilledExecution verifies a fully-filled order adapter.
func TestAdaptOrderFilledExecution(t *testing.T) {
	raw := apiOrder{
		OrderID:   "xyz789",
		Symbol:    "AAPL",
		Side:      "SELL",
		Status:    "CLOSED",
		Quantity:  "5",
		Price:     "200",
		OrderedAt: "2026-03-28T22:00:00+00:00",
		Execution: apiOrderExecution{
			FilledQuantity:     "5",
			AverageFilledPrice: "201.5",
		},
	}

	got := adaptOrder(raw)

	if got.FilledQuantity != 5 {
		t.Fatalf("FilledQuantity: want 5, got %v", got.FilledQuantity)
	}
	if got.AverageExecutionPrice != 201.5 {
		t.Fatalf("AverageExecutionPrice: want 201.5, got %v", got.AverageExecutionPrice)
	}
}

// TestAdaptOrderNoOrderedAt verifies nil SubmittedAt when orderedAt is empty.
func TestAdaptOrderNoOrderedAt(t *testing.T) {
	raw := apiOrder{OrderID: "x", Symbol: "X", OrderedAt: ""}
	got := adaptOrder(raw)
	if got.SubmittedAt != nil {
		t.Fatalf("SubmittedAt: expected nil for empty orderedAt, got %v", got.SubmittedAt)
	}
	if got.OrderDate != "" {
		t.Fatalf("OrderDate: expected empty for empty orderedAt, got %q", got.OrderDate)
	}
}

// TestAdaptOrdersBatch verifies the slice adapter.
func TestAdaptOrdersBatch(t *testing.T) {
	raw := []apiOrder{
		{OrderID: "a", Symbol: "A", Quantity: "1"},
		{OrderID: "b", Symbol: "B", Quantity: "2"},
	}
	got := adaptOrders(raw)
	if len(got) != 2 {
		t.Fatalf("len: want 2, got %d", len(got))
	}
	if got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("IDs: want a,b got %q,%q", got[0].ID, got[1].ID)
	}
}

// TestAdaptOrdersEmpty verifies that nil input returns a non-nil empty slice.
func TestAdaptOrdersEmpty(t *testing.T) {
	got := adaptOrders(nil)
	if got == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

// TestOrdersIntegration tests Orders() against an httptest server.
func TestOrdersIntegration(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/orders":
			gotHeader = r.Header.Get("X-Tossinvest-Account")
			q := r.URL.Query()
			if q.Get("status") != "OPEN" {
				t.Errorf("status: want OPEN, got %q", q.Get("status"))
			}
			_, _ = w.Write([]byte(`{"result":{"orders":[{"orderId":"abc123","symbol":"005930","side":"BUY","orderType":"LIMIT","timeInForce":"DAY","status":"OPEN","quantity":"10","price":"70000","currency":"KRW","orderedAt":"2026-03-29T09:30:00+09:00","canceledAt":null,"orderAmount":null,"execution":{"filledQuantity":"0","averageFilledPrice":null,"filledAmount":null,"commission":null,"tax":null,"filledAt":null,"settlementDate":null}}],"nextCursor":null,"hasNext":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithAccountSeq(3),
	)

	got, err := c.Orders(context.Background(), OrdersFilter{Status: "OPEN"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 order, got %d", len(got))
	}
	if got[0].ID != "abc123" {
		t.Fatalf("ID: want abc123, got %q", got[0].ID)
	}
	if got[0].Quantity != 10 {
		t.Fatalf("Quantity: want 10, got %v", got[0].Quantity)
	}
	if gotHeader != "3" {
		t.Fatalf("X-Tossinvest-Account: want 3, got %q", gotHeader)
	}
}

// TestOrdersFilterEmpty verifies that Orders() with zero filter omits all params.
func TestOrdersFilterEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/orders":
			q := r.URL.Query()
			for _, key := range []string{"status", "symbol", "from", "to", "cursor", "limit"} {
				if q.Get(key) != "" {
					t.Errorf("param %q: expected empty/absent, got %q", key, q.Get(key))
				}
			}
			_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithAccountSeq(1), // skip lazy fetch; test focuses on query param absence
	)

	got, err := c.Orders(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 orders, got %d", len(got))
	}
}

// ---------------------------------------------------------------------------
// OrderByID integration tests
// ---------------------------------------------------------------------------

// TestOrderByIDIntegration tests OrderByID() against an httptest server.
func TestOrderByIDIntegration(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/orders/abc123":
			gotHeader = r.Header.Get("X-Tossinvest-Account")
			_, _ = w.Write([]byte(`{"result":{"orderId":"abc123","symbol":"005930","side":"BUY","orderType":"LIMIT","timeInForce":"DAY","status":"OPEN","quantity":"10","price":"70000","currency":"KRW","orderedAt":"2026-03-29T09:30:00+09:00","canceledAt":null,"orderAmount":null,"execution":{"filledQuantity":"0","averageFilledPrice":null,"filledAmount":null,"commission":null,"tax":null,"filledAt":null,"settlementDate":null}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithAccountSeq(5),
	)

	got, err := c.OrderByID(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "abc123" {
		t.Fatalf("ID: want abc123, got %q", got.ID)
	}
	if got.Symbol != "005930" {
		t.Fatalf("Symbol: want 005930, got %q", got.Symbol)
	}
	if got.Price != 70000 {
		t.Fatalf("Price: want 70000, got %v", got.Price)
	}
	if gotHeader != "5" {
		t.Fatalf("X-Tossinvest-Account: want 5, got %q", gotHeader)
	}
}

// ---------------------------------------------------------------------------
// The raw-preserving read (change console-orders-screen, task 1.1-1.4)
// ---------------------------------------------------------------------------

// ordersRawServer serves one page of two orders: one the broker priced and
// filled, one it sent nulls for, and one priced at a genuine zero.
func ordersRawServer(t *testing.T, hasNext bool, nextCursor string) *httptest.Server {
	t.Helper()
	next := "null"
	if nextCursor != "" {
		next = `"` + nextCursor + `"`
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/orders":
			_, _ = w.Write([]byte(`{"result":{"orders":[` +
				// A pending market order: price is null and the whole execution
				// object is null, which is what the API sends while an order is
				// alive.
				`{"orderId":"pending-1","symbol":"005930","side":"BUY","orderType":"MARKET",` +
				`"timeInForce":"DAY","status":"PENDING","quantity":"10","price":null,"currency":"KRW",` +
				`"orderedAt":"2026-03-29T09:30:00+09:00","canceledAt":null,"orderAmount":null,` +
				`"execution":null},` +
				// A filled order whose average price really is zero-something.
				`{"orderId":"filled-1","symbol":"AAPL","side":"SELL","orderType":"LIMIT",` +
				`"timeInForce":"DAY","status":"FILLED","quantity":"5","price":"0","currency":"USD",` +
				`"orderedAt":"2026-03-28T22:00:00Z","canceledAt":null,"orderAmount":null,` +
				`"execution":{"filledQuantity":"5","averageFilledPrice":"0","filledAmount":"0",` +
				`"commission":null,"tax":null,"filledAt":null,"settlementDate":null}}` +
				`],"nextCursor":` + next + `,"hasNext":` + boolText(hasNext) + `}}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func boolText(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func ordersRawClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return New(
		Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithAccountSeq(3),
	)
}

// TestTheAdaptedOrderReadCannotTellAnAbsentPriceFromAZeroOne pins the reason the
// raw read below exists, and pins Orders' behaviour while it is at it.
//
// It is not a complaint about Orders: float64 is the right shape for a CLI that
// prints a portfolio, and every existing caller depends on exactly this. It is
// the measurement that says a screen which has to render "the broker sent no
// value" as something other than 0 cannot be built on it.
func TestTheAdaptedOrderReadCannotTellAnAbsentPriceFromAZeroOne(t *testing.T) {
	srv := ordersRawServer(t, false, "")
	defer srv.Close()

	got, err := ordersRawClient(t, srv).Orders(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 orders, got %d", len(got))
	}
	if got[0].Price != 0 || got[1].Price != 0 {
		t.Fatalf("Orders no longer collapses a null price to 0 (%v, %v); if that changed on purpose "+
			"the raw read's reason for existing changed with it", got[0].Price, got[1].Price)
	}
	if got[0].AverageExecutionPrice != 0 || got[1].AverageExecutionPrice != 0 {
		t.Fatalf("Orders no longer collapses a null execution to 0 (%v, %v)",
			got[0].AverageExecutionPrice, got[1].AverageExecutionPrice)
	}
}

// TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne.
//
// The screen's whole claim is "what the broker did not send is not zero". The
// API sends price as null for a market order and the entire execution object as
// null while an order is pending, so every live order arrives with an average
// fill price of nothing — and rendering that as 0 says the order filled.
func TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne(t *testing.T) {
	srv := ordersRawServer(t, false, "")
	defer srv.Close()

	page, err := ordersRawClient(t, srv).OrdersRaw(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Orders) != 2 {
		t.Fatalf("want 2 orders, got %d", len(page.Orders))
	}

	pending, filled := page.Orders[0], page.Orders[1]
	for _, tc := range []struct{ field, got, want string }{
		{"pending price", pending.Price, ""},
		{"pending filledQuantity", pending.FilledQuantity, ""},
		{"pending averageFilledPrice", pending.AverageFilledPrice, ""},
		{"filled price", filled.Price, "0"},
		{"filled filledQuantity", filled.FilledQuantity, "5"},
		{"filled averageFilledPrice", filled.AverageFilledPrice, "0"},
		{"pending quantity", pending.Quantity, "10"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q; an absent value that arrives as a number cannot be told "+
				"from a measured zero once it is on the screen", tc.field, tc.got, tc.want)
		}
	}
	if pending.Status != "PENDING" {
		t.Errorf("status = %q, want PENDING unmodified; internal/brokerstate is what derives a "+
			"state from it and it needs the broker's own word", pending.Status)
	}
	if pending.ID != "pending-1" || filled.ID != "filled-1" {
		t.Errorf("order ids = %q, %q; the order number is a column on the screen",
			pending.ID, filled.ID)
	}
}

// TestTheRawOrderReadDerivesTheMarketFromTheCurrency.
//
// The /orders response carries no market and no name. currency is decoded by the
// adapted read and then dropped, so a market column is either derived here or it
// does not exist — and a screen with no market cannot answer "which of these is
// in the session that is open right now".
func TestTheRawOrderReadDerivesTheMarketFromTheCurrency(t *testing.T) {
	srv := ordersRawServer(t, false, "")
	defer srv.Close()

	page, err := ordersRawClient(t, srv).OrdersRaw(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Orders[0].Currency != "KRW" || page.Orders[0].Market != "KR" {
		t.Errorf("KRW order: currency %q market %q, want KRW/KR",
			page.Orders[0].Currency, page.Orders[0].Market)
	}
	if page.Orders[1].Currency != "USD" || page.Orders[1].Market != "US" {
		t.Errorf("USD order: currency %q market %q, want USD/US",
			page.Orders[1].Currency, page.Orders[1].Market)
	}
	if got := marketFromCurrency("JPY"); got != "" {
		t.Errorf("marketFromCurrency(JPY) = %q, want empty; a currency this build does not know "+
			"is an unknown market, not a guessed one", got)
	}
}

// TestTheRawOrderReadReportsThatThePageWasTruncated.
//
// Orders decodes nextCursor and hasNext and drops both, so a caller counting a
// first page has no way to know there is a second one. A count taken from a
// truncated page is a confidently short number, which on this screen means
// leftovers that are filling the exposure cap disappear.
func TestTheRawOrderReadReportsThatThePageWasTruncated(t *testing.T) {
	srv := ordersRawServer(t, true, "cursor-2")
	defer srv.Close()

	page, err := ordersRawClient(t, srv).OrdersRaw(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasNext {
		t.Error("HasNext is false on a page the broker said has a successor; the count would be " +
			"rendered as a settled number")
	}
	if page.NextCursor != "cursor-2" {
		t.Errorf("NextCursor = %q, want cursor-2", page.NextCursor)
	}
}
