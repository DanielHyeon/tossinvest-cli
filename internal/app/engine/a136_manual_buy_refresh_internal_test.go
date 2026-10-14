package engine

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// a136 D9=b: refresher 는 OPEN 한 번 + CLOSED 를 페이지 끝까지 읽고, 엔진 주문을 빼고, 결과(실패 포함)를 보드에 게시함.

type a136Pages struct {
	open    []string
	closed  [][]string
	calls   []official.OrdersFilter
	cursors []string
	err     error
}

func (p *a136Pages) OrdersPageRaw(_ context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error) {
	p.calls, p.cursors = append(p.calls, filter), append(p.cursors, cursor)
	if p.err != nil {
		return official.RawOrderPage{}, p.err
	}
	raw := func(rows []string) []json.RawMessage {
		out := []json.RawMessage{}
		for _, r := range rows {
			out = append(out, json.RawMessage(r))
		}
		return out
	}
	if filter.Status == "OPEN" {
		return official.RawOrderPage{Orders: raw(p.open)}, nil
	}
	page := 0
	if cursor != "" {
		page = int(cursor[0] - '0')
	}
	out := official.RawOrderPage{Orders: raw(p.closed[page])}
	if page+1 < len(p.closed) {
		out.HasNext, out.NextCursor = true, string(rune('0'+page+1))
	}
	return out, nil
}

type a136EngineOrders map[string]bool

func (o a136EngineOrders) EngineBrokerOrderIDs(context.Context, string) (map[string]bool, error) {
	return o, nil
}

func TestA136TheRefresherPublishesTodaysManualBuys(t *testing.T) {
	now := time.Date(2026, 10, 14, 2, 0, 0, 0, time.UTC) // KST 11:00 — 라벨 2026-10-14
	pages := &a136Pages{
		open: []string{`{"orderId":"m-open","side":"BUY","status":"OPEN","quantity":"2","price":"1000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
			`{"orderId":"engine-1","side":"BUY","status":"OPEN","quantity":"9","price":"9999","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`},
		closed: [][]string{
			{`{"orderId":"m-c1","side":"BUY","status":"CLOSED","quantity":"1","price":"500","currency":"KRW","orderedAt":"2026-10-14T01:30:00Z","execution":{"filledQuantity":"1","filledAmount":"500"}}`},
			{`{"orderId":"m-c2","side":"BUY","status":"CLOSED","quantity":"1","price":"3","currency":"USD","orderedAt":"2026-10-14T01:40:00Z","execution":{"filledQuantity":"1","filledAmount":"3"}}`},
		},
	}
	board := execgw.NewManualBuyBoard()
	r := &manualBuyRefresh{board: board, reads: pages, orders: a136EngineOrders{"engine-1": true}, account: "acct",
		clock: clock.NewFake(now), interval: time.Hour}
	r.refreshOnce(context.Background())
	got, ok := board.Snapshot()
	if !ok || got.Err != nil || got.Day != "2026-10-14" || !got.AsOf.Equal(now) {
		t.Fatalf("snapshot = %+v", got)
	}
	if got.ByCurrency["KRW"] != "2500" || got.ByCurrency["USD"] != "3" {
		t.Fatalf("totals = %v, want KRW 2500 (engine order excluded) and USD 3", got.ByCurrency)
	}
	if len(pages.calls) != 3 || pages.calls[0].Status != "OPEN" || pages.calls[1].From != "2026-10-13" ||
		pages.calls[1].To != "2026-10-15" || pages.cursors[2] != "1" {
		t.Fatalf("reads = %+v cursors = %v, want OPEN once then CLOSED pages walked to the end", pages.calls, pages.cursors)
	}
}

func TestA136AFailedReadIsPublishedAsUnknown(t *testing.T) {
	board := execgw.NewManualBuyBoard()
	board.Publish(execgw.ManualBuySnapshot{AsOf: time.Unix(0, 0), Day: "2026-10-14", ByCurrency: map[string]string{"KRW": "0"}})
	r := &manualBuyRefresh{board: board, reads: &a136Pages{err: errors.New("broker down")}, orders: a136EngineOrders{},
		clock: clock.NewFake(time.Date(2026, 10, 14, 2, 0, 0, 0, time.UTC)), interval: time.Hour}
	r.refreshOnce(context.Background())
	got, _ := board.Snapshot()
	if got.Err == nil {
		t.Fatal("a failed read must replace the previous snapshot with an unknown one, never keep the old value")
	}
}
