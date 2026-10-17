package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

func TestListPendingOrdersParsesStructuredFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/trading/orders/histories/all/pending" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
		  "result": [
		    {
		      "orderNo": 5,
		      "orderId": "opaque-order-id",
		      "stockCode": "US20220809012",
		      "tradeType": "buy",
		      "status": "체결대기",
		      "quantity": 1,
		      "pendingQuantity": 1,
		      "orderPrice": 500,
		      "orderedAt": "2026-03-11T19:10:26.582935Z"
		    }
		  ]
		}`))
	}))
	defer server.Close()

	client := New(Config{
		HTTPClient:  server.Client(),
		APIBaseURL:  server.URL,
		InfoBaseURL: server.URL,
		CertBaseURL: server.URL,
		Session: &session.Session{
			Cookies: map[string]string{"SESSION": "test-session"},
		},
	})

	orders, err := client.ListPendingOrders(context.Background())
	if err != nil {
		t.Fatalf("ListPendingOrders returned error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 pending order, got %d", len(orders))
	}
	if orders[0].ID != "5" {
		t.Fatalf("expected parsed order id 5, got %q", orders[0].ID)
	}
	if orders[0].Symbol != "US20220809012" {
		t.Fatalf("expected parsed symbol, got %q", orders[0].Symbol)
	}
	if orders[0].Price != 500 {
		t.Fatalf("expected parsed price 500, got %v", orders[0].Price)
	}
	if orders[0].Status != "체결대기" {
		t.Fatalf("expected parsed status, got %q", orders[0].Status)
	}
}

func TestListPendingOrdersRejectsAbsentOrNullResult(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"absent": `{}`,
		"null":   `{"result":null}`,
		"object": `{"result":{}}`,
		"scalar": `{"result":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
			if orders, err := c.ListPendingOrders(context.Background()); err == nil || orders != nil {
				t.Fatalf("orders=%v err=%v, want schema error and nil orders", orders, err)
			}
		})
	}
}

func TestListPendingOrdersAcceptsExplicitEmptyResult(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"result":[]}`)) }))
	defer srv.Close()
	c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
	orders, err := c.ListPendingOrders(context.Background())
	if err != nil || orders == nil || len(orders) != 0 {
		t.Fatalf("orders=%v err=%v, want non-nil empty", orders, err)
	}
}

func TestHasPendingOrderRejectsUnknownPendingCollection(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	defer srv.Close()
	c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
	found, err := c.HasPendingOrder(context.Background(), "1")
	if err == nil || found {
		t.Fatalf("found=%v err=%v, want error and no authoritative false", found, err)
	}
}

func TestFindOrderRejectsUnknownPendingCollectionWithoutCompletedFallback(t *testing.T) {
	t.Parallel()
	var completedCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/trading/orders/histories/all/pending" {
			completedCalls++
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
	if _, err := c.FindOrder(context.Background(), "1", "us"); err == nil {
		t.Fatal("want pending collection error")
	}
	if completedCalls != 0 {
		t.Fatalf("completed history calls=%d, want 0", completedCalls)
	}
}
