package filldetect_test

// a087 tasks 3.5 · 3.6 — 시장가 보호 청산(가격 없는 KR 매도)의 체결가 확정과 부분체결 잔량을 체결 감지가 어떻게
// 닫는지 못 박음. 변경 없음을 고정하는 시험임: 체결 감지는 주문 유형도 intents.price 도 읽지 않고(열거:
// openspec/changes/a087-…/analysis/intents-price-null/ — 비시험 Go 에서 intents.price 를 읽는 SQL 은 둘뿐이고
// 체결 경로가 아님), 체결가는 브로커 `execution.averageFilledPrice` 에서만 옴(payload.go parseSnapshot).
// 실제 저널 위에서 실제 Detector 를 돌림 — 브로커는 시험 더블(실주문 없음).

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const a087OrderedAt = "2026-03-30T10:00:00+09:00"

// recordConfirmedMarketSell 는 a087 이후 엔진이 남기는 보호 청산 intent 를 저널에 확정 상태로 둠 — 가격 없음(NULL).
func recordConfirmedMarketSell(t *testing.T, j *journal.Journal, orderID string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + orderID, Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "005930", Side: "SELL", OrderType: "MARKET", Quantity: "10",
			Price: "", Currency: "KRW", Source: "engine", Fingerprint: "fp-" + orderID,
		},
		Kind: journal.KindPlace, AttemptID: "attempt-" + orderID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
}

func a087MarketSellOrder(status, filled, avg, canceledAt string) rawOrder {
	return rawOrder{
		id: "o-mkt", symbol: "005930", side: "SELL", currency: "KRW", orderedAt: a087OrderedAt,
		status: status, quantity: "10", filled: filled, avgPrice: avg, canceled: canceledAt,
		filledAt: pollStart.Add(-time.Second).Format(time.RFC3339),
	}
}

func a087Tracked(t *testing.T, j *journal.Journal) []filldetect.TrackedOrder {
	t.Helper()
	tracked, err := filldetect.JournalTracked{Journal: j, AccountRef: "acct-1"}.TrackedOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return tracked
}

// TestA087MarketSellFillPriceIsTheBrokersAverage (3.5) — 가격 없는 시장가 매도의 체결가는 브로커가 보고한 평균
// 체결가로 원장에 확정되고, intent 의 가격 칸은 비어 있는 채로 남음(관측가가 체결가로 새지 않음).
func TestA087MarketSellFillPriceIsTheBrokersAverage(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", a087MarketSellOrder("PARTIAL_FILLED", "4", "70100", "")))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()
	recordConfirmedMarketSell(t, j, "o-mkt")

	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 1: %v", err)
	}
	// 전량 체결 — 미체결 목록에서 빠지고 id 로 읽힘.
	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("")
	pager.mu.Unlock()
	d.Order = &fakeOrderReader{orders: rawOrders(a087MarketSellOrder("FILLED", "10", "70050", ""))}
	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 2: %v", err)
	}

	snap, err := j.LookupFill(ctx, "o-mkt")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Terminal || snap.FilledQuantity != "10" || snap.AveragePrice != "70050" {
		t.Fatalf("snapshot = %+v, want terminal 10 @ the broker's 70050", snap)
	}
	events, err := j.FillEvents(ctx, "o-mkt")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].AveragePrice != "70100" || events[1].AveragePrice != "70050" {
		t.Fatalf("fill events = %+v, want 4 @ 70100 then 6 more with cumulative average 70050", events)
	}
	intent, err := j.LookupIntent(ctx, "intent-o-mkt")
	if err != nil {
		t.Fatal(err)
	}
	if intent.Price != "" || intent.OrderType != "MARKET" {
		t.Fatalf("intent after the fill = %+v, want the unpriced market intent untouched", intent)
	}
}

// TestA087MarketSellRemainderIsTrackedUntilTheBrokerEndsIt (3.6) — 부분체결 뒤 잔량이 브로커에 남아 있는 동안
// 그 주문은 추적 집합에 남아 매 주기 읽히고, 브로커가 잔량을 취소로 끝내면(CANCELED + canceledAt) 종결 스냅숏
// CANCELLED_PARTIALLY_FILLED 로 닫히며 체결 수량은 4 그대로(잔량이 체결로 둔갑하지 않음).
// 잔량을 **다시 내는** 것은 체결 감지가 아니라 청산 루프의 몫임 — analysis/partial-fill-3.6/README.md.
func TestA087MarketSellRemainderIsTrackedUntilTheBrokerEndsIt(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", a087MarketSellOrder("PARTIAL_FILLED", "4", "70100", "")))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()
	recordConfirmedMarketSell(t, j, "o-mkt")

	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 1: %v", err)
	}
	if got := a087Tracked(t, j); len(got) != 1 || got[0].OrderID != "o-mkt" {
		t.Fatalf("tracked while the remainder is working = %+v, want o-mkt", got)
	}

	// 잔량이 그대로 남은 주기 — 같은 누적은 새 체결이 아님.
	clk.Advance(3 * time.Second)
	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll 2: %v", err)
	}
	if cycle.Fills != 0 {
		t.Fatalf("an unchanged remainder produced %d fills, want 0", cycle.Fills)
	}

	// 브로커가 잔량을 끝냄 — 미체결 목록에서 빠지고 id 로 읽히는 취소(부분체결 보존).
	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("")
	pager.mu.Unlock()
	d.Order = &fakeOrderReader{orders: rawOrders(
		a087MarketSellOrder("CANCELED", "4", "70100", "2026-03-30T10:00:09+09:00"))}
	cycle, err = d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll 3: %v", err)
	}
	if cycle.TrackedReads != 1 || cycle.Fills != 0 || len(cycle.FailedClosed) != 0 {
		t.Fatalf("cycle = %+v, want one tracked read, no new fill, no fail-closed", cycle)
	}
	snap, err := j.LookupFill(ctx, "o-mkt")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Terminal || snap.State != "CANCELLED_PARTIALLY_FILLED" || snap.FilledQuantity != "4" {
		t.Fatalf("snapshot = %+v, want terminal CANCELLED_PARTIALLY_FILLED with 4 filled", snap)
	}
	if got := a087Tracked(t, j); len(got) != 0 {
		t.Fatalf("tracked after the remainder ended = %+v, want none", got)
	}
}
