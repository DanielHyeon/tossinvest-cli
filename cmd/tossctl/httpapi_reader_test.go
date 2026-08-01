package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

type httpAPIHoldingsFixture struct {
	calls int
	rows  []domain.Position
}

func (f *httpAPIHoldingsFixture) Holdings(context.Context, string) ([]domain.Position, error) {
	f.calls++
	return append([]domain.Position(nil), f.rows...), nil
}

type httpAPIOrdersFixture struct {
	calls   int
	reading console.OrdersReading
}

func (f *httpAPIOrdersFixture) Orders(context.Context) (console.OrdersReading, error) {
	f.calls++
	return f.reading, nil
}

func TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	holdings := &httpAPIHoldingsFixture{rows: []domain.Position{{
		Symbol: "005930", Name: "삼성전자", MarketCode: "KR", Quantity: 3, AveragePrice: 70000, CurrentPrice: 71000,
	}}}
	orders := &httpAPIOrdersFixture{reading: console.OrdersReading{
		AccountRef: "raw-account-1234", Open: []console.OrderRecord{{
			ID: "order-1", Symbol: "005930", Market: "KR", Side: "BUY", Status: "OPEN", Quantity: "1",
		}},
	}}
	reader := &httpAPIReader{
		now: func() time.Time { return now }, holdings: holdings, orders: orders,
		accountRef: func() (string, error) { return "raw-account-1234", nil },
	}

	for range 2 {
		positions, err := reader.Positions(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(positions.Items) != 1 || positions.Items[0].Symbol != "005930" ||
			positions.Items[0].ManagementStatus != "unknown" || strings.Contains(positions.Items[0].AccountLabel, "raw-account") {
			t.Fatalf("position projection=%+v", positions.Items)
		}
		projectedOrders, err := reader.Orders(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(projectedOrders.Items) != 1 || projectedOrders.Items[0].ID != "order-1" ||
			strings.Contains(projectedOrders.Items[0].AccountLabel, "raw-account") {
			t.Fatalf("order projection=%+v", projectedOrders.Items)
		}
	}
	if holdings.calls != 1 || orders.calls != 1 {
		t.Fatalf("broker calls holdings/orders=%d/%d want=1/1 within cache interval", holdings.calls, orders.calls)
	}
}
