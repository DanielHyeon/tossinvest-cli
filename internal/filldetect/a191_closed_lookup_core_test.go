package filldetect_test

// a191 핵심 RED — 미체결 목록을 떠났는데 단건 조회가 「살아 있음(PENDING)」을 돌려주는 엔진 주문의 종결을 종결(CLOSED) 목록으로
// 확인함. 이 파일은 base 의 공개 표면만 써서, base 에서는 컴파일되고 **행동으로** 실패하게 둠(종결이 안 쓰이고 추적이 남음).
// 실제 저널 위에서 실제 Detector 를 돌림 — 브로커는 시험 더블(실주문 없음). 값은 전부 합성임(종목·주문번호·수량·가격).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// 합성 고정값 — 실제 종목·주문과 겹치지 않는 이름만 씀.
const (
	a191Symbol     = "SYN001"
	a191Account    = "acct-1"
	a191TradingDay = "2026-03-30"
	a191OrderedAt  = "2026-03-30T10:00:00+09:00"
)

// a191Pager 는 상태 그룹(OPEN/CLOSED)을 가려 답하고 받은 질의를 전부 기록하는 가짜 목록 조회임.
// 기존 fakePager 는 q.Status 를 보지 않으므로(커서로만 고름) CLOSED 질의를 단언할 수 없어 따로 둠.
type a191Pager struct {
	mu          sync.Mutex
	open        execgw.OrderPage
	closed      map[string]execgw.OrderPage // 커서 → 페이지
	closedErr   error
	blockClosed bool // CLOSED 질의가 ctx 가 끝날 때까지 멈춤(시한 시험용)
	queries     []execgw.OrderQuery
}

func newA191Pager() *a191Pager {
	return &a191Pager{closed: map[string]execgw.OrderPage{}}
}

func (p *a191Pager) OrdersPageRaw(ctx context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	p.queries = append(p.queries, q)
	open, closedPage, closedErr, block := p.open, p.closed[cursor], p.closedErr, p.blockClosed
	p.mu.Unlock()
	switch q.Status {
	case "OPEN":
		return open, nil
	case "CLOSED":
		if block {
			<-ctx.Done()
			return execgw.OrderPage{}, ctx.Err()
		}
		if closedErr != nil {
			return execgw.OrderPage{}, closedErr
		}
		return closedPage, nil
	default:
		return execgw.OrderPage{}, fmt.Errorf("a191Pager: unexpected status group %q", q.Status)
	}
}

// closedQueries 는 받은 CLOSED 질의만 돌려줌.
func (p *a191Pager) closedQueries() []execgw.OrderQuery {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []execgw.OrderQuery
	for _, q := range p.queries {
		if q.Status == "CLOSED" {
			out = append(out, q)
		}
	}
	return out
}

func (p *a191Pager) setClosed(cursor string, pg execgw.OrderPage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed[cursor] = pg
}

// a191Order 는 KR 합성 매도 주문 하나의 raw JSON 을 만듦.
func a191Order(id, status, filled string) rawOrder {
	return rawOrder{
		id: id, symbol: a191Symbol, side: "SELL", currency: "KRW", orderedAt: a191OrderedAt,
		status: status, quantity: "10", filled: filled,
	}
}

// a191ClosedPage 는 CLOSED 질의의 단일 페이지임.
func a191ClosedPage(rows ...json.RawMessage) execgw.OrderPage {
	return execgw.OrderPage{Orders: rows}
}

// recordConfirmedA191Sell 은 엔진이 낸 합성 매도 intent 를 저널에 확정 상태로 둠.
func recordConfirmedA191Sell(t *testing.T, j *journal.Journal, orderID string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + orderID, Market: "kr", TradingDay: a191TradingDay, AccountRef: a191Account,
			Symbol: a191Symbol, Side: "SELL", OrderType: "LIMIT", Quantity: "10",
			Price: "1", Currency: "KRW", Source: "engine", Fingerprint: "fp-" + orderID,
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

// a191JournalDetector 는 실제 저널·실제 Detector 위에 a191Pager 와 단건 조회 더블을 꽂음.
func a191JournalDetector(t *testing.T, pager *a191Pager, reader *fakeOrderReader) (*filldetect.Detector, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(pollStart)
	j := openLedgerJournal(t, clk, filepath.Join(t.TempDir(), "journal.db"))
	return &filldetect.Detector{
		Orders:    pager,
		Order:     reader,
		Positions: &fakePositions{},
		Tracked:   filldetect.JournalTracked{Journal: j, AccountRef: a191Account},
		Ledger:    filldetect.JournalLedger{Journal: j, AccountRef: a191Account},
		Clock:     clk,
	}, j, clk
}

// TestA191OrderThatLeftTheOpenListAndIsCanceledInTheClosedListIsRecordedTerminal (tasks 3.1) — 결함 그대로의 입력:
// OPEN 에 없음 + 단건 PENDING·취소 시각 없음·체결 0 + CLOSED 에 같은 주문 CANCELED·체결 0 → 종결 CANCELLED, 추적 종료.
func TestA191OrderThatLeftTheOpenListAndIsCanceledInTheClosedListIsRecordedTerminal(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191Order("o-a191-1", "CANCELED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191Order("o-a191-1", "PENDING", "0"))}
	d, j, _ := a191JournalDetector(t, pager, reader)
	ctx := context.Background()
	recordConfirmedA191Sell(t, j, "o-a191-1")

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(cycle.FailedClosed) != 0 {
		t.Fatalf("FailedClosed = %v, want none", cycle.FailedClosed)
	}
	snap, err := j.LookupFill(ctx, "o-a191-1")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Terminal || snap.State != "CANCELLED" || snap.FilledQuantity != "0" {
		t.Fatalf("snapshot = %+v, want terminal CANCELLED with 0 filled", snap)
	}
	tracked, err := filldetect.JournalTracked{Journal: j, AccountRef: a191Account}.TrackedOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 0 {
		t.Fatalf("tracked after the terminal snapshot = %+v, want none", tracked)
	}
}

// TestA191TerminalSnapshotLeavesTheReconcileLocalState (tasks 3.2) — 같은 흐름 뒤 대사의 로컬 열린 주문에서도 빠짐
// (missing-order 의 원천이 사라짐).
func TestA191TerminalSnapshotLeavesTheReconcileLocalState(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191Order("o-a191-2", "CANCELED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191Order("o-a191-2", "PENDING", "0"))}
	d, j, _ := a191JournalDetector(t, pager, reader)
	ctx := context.Background()
	recordConfirmedA191Sell(t, j, "o-a191-2")

	before, err := reconcile.LocalStateFromJournal(ctx, j, a191Account)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.OpenOrders) != 1 {
		t.Fatalf("local open orders before the poll = %d, want 1 (the confirmed order)", len(before.OpenOrders))
	}
	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}
	after, err := reconcile.LocalStateFromJournal(ctx, j, a191Account)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.OpenOrders) != 0 || len(after.ScopedOpenOrders) != 0 {
		t.Fatalf("local open orders after the terminal snapshot = %+v, want none", after.OpenOrders)
	}
}

// TestA191AcceptedClosedRowKeepsTheSingleReadExecution (tasks 3.4b, 변이 m19) — 받아들일 때 바뀌는 것은 상태뿐임.
// 원장 직전 0 → 단건 PARTIAL_FILLED 4(평균가 1) + CLOSED CANCELED 4(평균가 2) → 체결 4 가 **단건 평균가**로 기록되고 종결.
func TestA191AcceptedClosedRowKeepsTheSingleReadExecution(t *testing.T) {
	pager := newA191Pager()
	closedRow := a191Order("o-a191-3", "CANCELED", "4")
	closedRow.avgPrice = "2"
	closedRow.canceled = "2026-03-30T10:05:00+09:00"
	pager.setClosed("", a191ClosedPage(closedRow.json()))
	single := a191Order("o-a191-3", "PARTIAL_FILLED", "4")
	single.avgPrice = "1"
	reader := &fakeOrderReader{orders: rawOrders(single)}
	d, j, _ := a191JournalDetector(t, pager, reader)
	ctx := context.Background()
	recordConfirmedA191Sell(t, j, "o-a191-3")

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if cycle.Fills != 1 {
		t.Fatalf("fills = %d, want 1 (the single read's 4 against an empty ledger)", cycle.Fills)
	}
	snap, err := j.LookupFill(ctx, "o-a191-3")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Terminal || snap.State != "CANCELLED_PARTIALLY_FILLED" || snap.FilledQuantity != "4" || snap.AveragePrice != "1" {
		t.Fatalf("snapshot = %+v, want terminal CANCELLED_PARTIALLY_FILLED, 4 filled at the single read's average 1", snap)
	}
}

// errA191Throttled 는 CLOSED 질의 실패를 흉내 냄.
var errA191Throttled = errors.New("a191: closed list throttled")

// a191Since 는 시험 시계 기준 경과를 한 줄로 씀(가독성용).
func a191Since(clk *clock.Fake, from time.Time) time.Duration { return clk.Now().Sub(from) }
