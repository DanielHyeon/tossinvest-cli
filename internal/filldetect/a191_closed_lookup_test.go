package filldetect_test

// a191 수용 조건·발동 조건·간격·시한·알림의 시험 — 시험 더블(a191Pager·fakeLedger) 위에서 Detector 를 돌림.
// 모든 값은 합성임. 모든 시험이 a191Pager 가 받은 질의를 단언함(기존 fakePager 는 status 를 안 봐 우발 CLOSED 질의를 못 잡음).

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a191Alerts 는 알림 통로를 기록함.
type a191Alerts struct {
	mu     sync.Mutex
	events []obs.Event
}

func (a *a191Alerts) Notify(_ context.Context, e obs.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, e)
	return nil
}

func (a *a191Alerts) all() []obs.Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]obs.Event(nil), a.events...)
}

// a191Tracked 는 합성 추적 주문 하나임(KR, 매도, 거래일 a191TradingDay).
func a191Tracked(id string) filldetect.TrackedOrder {
	return filldetect.TrackedOrder{
		OrderID: id, IntentID: "intent-" + id, AccountRef: "acct-test",
		Symbol: a191Symbol, Market: "kr", TradingDay: a191TradingDay, Side: "SELL",
	}
}

// a191FakeOrder 는 fakeTracked 의 기본 계좌(acct-test)와 맞는 합성 주문임.
func a191FakeOrder(id, status, filled string) rawOrder { return a191Order(id, status, filled) }

func a191Fake(t *testing.T, pager *a191Pager, reader *fakeOrderReader, tracked ...filldetect.TrackedOrder) (*filldetect.Detector, *clock.Fake, *fakeLedger, *a191Alerts) {
	t.Helper()
	clk := clock.NewFake(pollStart)
	ledger := newLedger(clk)
	alerts := &a191Alerts{}
	return &filldetect.Detector{
		Orders: pager, Order: reader, Positions: &fakePositions{},
		Tracked: fakeTracked{orders: tracked}, Ledger: ledger, Clock: clk, Alerts: alerts,
	}, clk, ledger, alerts
}

func a191Poll(t *testing.T, d *filldetect.Detector) filldetect.Cycle {
	t.Helper()
	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	return cycle
}

func lastApplied(t *testing.T, l *fakeLedger, id string) filldetect.Snapshot {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.applied) - 1; i >= 0; i-- {
		if l.applied[i].OrderID == id {
			return l.applied[i]
		}
	}
	t.Fatalf("no snapshot applied for %s", id)
	return filldetect.Snapshot{}
}

func reasonsOf(c filldetect.Cycle) map[string]filldetect.UnresolvedReason {
	out := map[string]filldetect.UnresolvedReason{}
	for _, u := range c.Unresolved {
		out[u.OrderID] = u.Reason
	}
	return out
}

// TestA191SingleReadPendingAndClosedListSilentKeepsTodaysSnapshot (3.3) — 못 찾음 → 오늘과 같은 단건 스냅숏, 사유 남음, 실패·래치 없음.
func TestA191SingleReadPendingAndClosedListSilentKeepsTodaysSnapshot(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-1", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-1"))
	cycle := a191Poll(t, d)
	if got := lastApplied(t, ledger, "o-1"); got.Derived.State != "OPEN_UNFILLED" || got.RawStatus != "PENDING" {
		t.Fatalf("applied = %+v, want today's OPEN_UNFILLED single read", got.Derived)
	}
	if len(cycle.FailedClosed) != 0 || cycle.ClosedReads != 1 {
		t.Fatalf("cycle = %+v, want one CLOSED lookup and no fail-closed", cycle)
	}
	if r := reasonsOf(cycle)["o-1"]; r != filldetect.UnresolvedClosedListNotFound {
		t.Fatalf("reason = %q, want %q", r, filldetect.UnresolvedClosedListNotFound)
	}
}

// TestA191ClosedListFillDisagreementNeverRecordsAFill (3.4, m6·m7·m16) — 체결 수량·주문 수량 불일치는 받지 않음.
// 단건 0 → CLOSED >0(증가 방향) 이라 m16 이 감소 가드에 가려지지 않음.
func TestA191ClosedListFillDisagreementNeverRecordsAFill(t *testing.T) {
	for _, tc := range []struct {
		name string
		row  rawOrder
	}{
		{"filled differs", a191FakeOrder("o-2", "FILLED", "10")},
		{"ordered quantity differs", func() rawOrder { o := a191FakeOrder("o-2", "CANCELED", "0"); o.quantity = "9"; return o }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pager := newA191Pager()
			pager.setClosed("", a191ClosedPage(tc.row.json()))
			reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-2", "PENDING", "0"))}
			d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-2"))
			cycle := a191Poll(t, d)
			got := lastApplied(t, ledger, "o-2")
			if got.Derived.Terminal || got.FilledQuantity != 0 || got.Quantity != 10 || cycle.Fills != 0 {
				t.Fatalf("applied = %+v (fills %d), want the single read untouched", got, cycle.Fills)
			}
			if r := reasonsOf(cycle)["o-2"]; r != filldetect.UnresolvedClosedListQuantityDisagrees {
				t.Fatalf("reason = %q", r)
			}
		})
	}
}

// TestA191ClosedListReadFailureDoesNotFailTheCycle (3.5, m10) — CLOSED 실패는 주기를 안 깨고 같은 주기 재조회 없음.
func TestA191ClosedListReadFailureDoesNotFailTheCycle(t *testing.T) {
	pager := newA191Pager()
	pager.closedErr = errA191Throttled
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-3", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-3"))
	cycle := a191Poll(t, d)
	if len(pager.closedQueries()) != 1 {
		t.Fatalf("CLOSED queries = %d, want exactly 1 (no same-cycle retry)", len(pager.closedQueries()))
	}
	if got := lastApplied(t, ledger, "o-3"); got.Derived.State != "OPEN_UNFILLED" {
		t.Fatalf("applied = %+v", got.Derived)
	}
	if r := reasonsOf(cycle)["o-3"]; r != filldetect.UnresolvedClosedListReadFailed {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191ClosedLookupIsThrottledPerOrder (3.6, m11) — 첫 조회 즉시, 30s 안 0회, 30s 뒤 1회.
func TestA191ClosedLookupIsThrottledPerOrder(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-4", "PENDING", "0"))}
	d, clk, _, _ := a191Fake(t, pager, reader, a191Tracked("o-4"))
	a191Poll(t, d)
	clk.Advance(3 * time.Second)
	cycle := a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 1 {
		t.Fatalf("CLOSED queries after 3s = %d, want 1", n)
	}
	if r := reasonsOf(cycle)["o-4"]; r != filldetect.UnresolvedClosedLookupDeferred {
		t.Fatalf("reason inside the interval = %q", r)
	}
	clk.Advance(27 * time.Second)
	a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 2 {
		t.Fatalf("CLOSED queries after 30s = %d, want 2", n)
	}
}

// TestA191ClosedRowFromAnotherScopeIsNotAccepted (3.7, m8) — 같은 id 라도 거래일이 다르면 안 받음.
func TestA191ClosedRowFromAnotherScopeIsNotAccepted(t *testing.T) {
	pager := newA191Pager()
	other := a191FakeOrder("o-5", "CANCELED", "0")
	other.orderedAt = "2026-03-31T10:00:00+09:00"
	pager.setClosed("", a191ClosedPage(other.json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-5", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-5"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-5").Derived.Terminal {
		t.Fatal("a row from another trading day closed the order")
	}
	if r := reasonsOf(cycle)["o-5"]; r != filldetect.UnresolvedClosedListScopeMismatch {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191DuplicateClosedRowsForOneIDAreAmbiguous (3.7, m9) — 같은 id·다른 범위 두 행 → 모호.
func TestA191DuplicateClosedRowsForOneIDAreAmbiguous(t *testing.T) {
	pager := newA191Pager()
	other := a191FakeOrder("o-6", "CANCELED", "0")
	other.orderedAt = "2026-03-31T10:00:00+09:00"
	pager.setClosed("", a191ClosedPage(a191FakeOrder("o-6", "CANCELED", "0").json(), other.json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-6", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-6"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-6").Derived.Terminal {
		t.Fatal("an ambiguous pair closed the order")
	}
	if r := reasonsOf(cycle)["o-6"]; r != filldetect.UnresolvedClosedListAmbiguous {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191SameScopeRowsWithDifferentPayloadsAreAmbiguous (3.7b, m20) — 같은 범위·다른 status 두 행은 어느 순서로 와도 모호;
// raw 바이트가 같은 중복은 하나로 셈.
func TestA191SameScopeRowsWithDifferentPayloadsAreAmbiguous(t *testing.T) {
	canceled := a191FakeOrder("o-7", "CANCELED", "0").json()
	partial := a191FakeOrder("o-7", "PARTIAL_FILLED", "0").json()
	for _, tc := range []struct {
		name string
		rows []json.RawMessage
		want filldetect.UnresolvedReason
	}{
		{"canceled first", []json.RawMessage{canceled, partial}, filldetect.UnresolvedClosedListAmbiguous},
		{"partial first", []json.RawMessage{partial, canceled}, filldetect.UnresolvedClosedListAmbiguous},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pager := newA191Pager()
			pager.setClosed("", a191ClosedPage(tc.rows...))
			reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-7", "PENDING", "0"))}
			d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-7"))
			cycle := a191Poll(t, d)
			if lastApplied(t, ledger, "o-7").Derived.Terminal {
				t.Fatal("conflicting rows closed the order")
			}
			if r := reasonsOf(cycle)["o-7"]; r != tc.want {
				t.Fatalf("reason = %q, want %q", r, tc.want)
			}
		})
	}
	t.Run("byte-identical duplicate counts once", func(t *testing.T) {
		pager := newA191Pager()
		pager.setClosed("", a191ClosedPage(canceled, canceled))
		reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-7", "PENDING", "0"))}
		d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-7"))
		a191Poll(t, d)
		if got := lastApplied(t, ledger, "o-7"); got.Derived.State != "CANCELLED" {
			t.Fatalf("applied = %+v, want CANCELLED", got.Derived)
		}
	})
}

// TestA191ClosedRowIDMustMatchByteForByte (3.7, m15) — 공백·대소문자만 다른 id 는 같은 주문이 아님.
func TestA191ClosedRowIDMustMatchByteForByte(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191FakeOrder("O-8", "CANCELED", "0").json(), a191FakeOrder(" o-8", "CANCELED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-8", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-8"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-8").Derived.Terminal {
		t.Fatal("a case- or space-variant id closed the order")
	}
	if r := reasonsOf(cycle)["o-8"]; r != filldetect.UnresolvedClosedListNotFound {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191ReplacedWithoutLineageIsNotAccepted (3.8, m4) — REPLACED + 후속 없음은 UNKNOWN 이라 안 받음.
func TestA191ReplacedWithoutLineageIsNotAccepted(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191FakeOrder("o-9", "REPLACED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-9", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-9"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-9").Derived.Terminal {
		t.Fatal("REPLACED without a successor closed the order")
	}
	if r := reasonsOf(cycle)["o-9"]; r != filldetect.UnresolvedClosedListNotCleanTerminal {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191PartialFilledInTheClosedGroupIsNotTerminal (3.8, m5) — CLOSED 그룹의 PARTIAL_FILLED 는 비종결이라 안 받음.
func TestA191PartialFilledInTheClosedGroupIsNotTerminal(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191FakeOrder("o-10", "PARTIAL_FILLED", "4").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-10", "PARTIAL_FILLED", "4"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-10"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-10").Derived.Terminal {
		t.Fatal("PARTIAL_FILLED from the CLOSED group closed the order")
	}
	if r := reasonsOf(cycle)["o-10"]; r != filldetect.UnresolvedClosedListNotCleanTerminal {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191UnknownSingleReadIsNeverLookedUp (3.8, m21 — M2 = 제외) — UNKNOWN 단건은 CLOSED 를 안 묻고 기존 fail-closed 그대로.
func TestA191UnknownSingleReadIsNeverLookedUp(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191FakeOrder("o-11", "CANCELED", "4").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-11", "PENDING", "4"))} // pending_with_fill → UNKNOWN
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-11"))
	cycle := a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 0 {
		t.Fatalf("CLOSED queries for an UNKNOWN single read = %d, want 0", n)
	}
	if got := lastApplied(t, ledger, "o-11"); !got.Derived.FailClosed {
		t.Fatalf("applied = %+v, want the single read's UNKNOWN untouched", got.Derived)
	}
	if len(cycle.Unresolved) != 0 {
		t.Fatalf("unresolved = %+v, want none (not a candidate)", cycle.Unresolved)
	}
}

// TestA191OrderStillInTheOpenListIsNeverLookedUp (3.9, m3).
func TestA191OrderStillInTheOpenListIsNeverLookedUp(t *testing.T) {
	pager := newA191Pager()
	pager.open = page("", a191FakeOrder("o-12", "PENDING", "0"))
	d, _, _, _ := a191Fake(t, pager, &fakeOrderReader{}, a191Tracked("o-12"))
	a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 0 {
		t.Fatalf("CLOSED queries for an order in the open list = %d, want 0", n)
	}
}

// TestA191TerminalSingleReadIsNeverLookedUp (3.9, m2).
func TestA191TerminalSingleReadIsNeverLookedUp(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-13", "FILLED", "10"))}
	d, _, _, _ := a191Fake(t, pager, reader, a191Tracked("o-13"))
	a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 0 {
		t.Fatalf("CLOSED queries for a terminal single read = %d, want 0", n)
	}
}

// TestA191ExistingTrackedReadGuardsRunBeforeTheLookup (3.9, m17) — 범위 불일치 단건은 기존 가드가 먼저 주기를 멈춤, CLOSED 0.
func TestA191ExistingTrackedReadGuardsRunBeforeTheLookup(t *testing.T) {
	pager := newA191Pager()
	wrong := a191FakeOrder("o-14", "PENDING", "0")
	wrong.side = "BUY"
	reader := &fakeOrderReader{orders: rawOrders(wrong)}
	d, _, _, _ := a191Fake(t, pager, reader, a191Tracked("o-14"))
	if _, err := d.PollOnce(context.Background()); err == nil || !strings.Contains(err.Error(), "does not match requested scope") {
		t.Fatalf("poll error = %v, want the existing scope guard", err)
	}
	if n := len(pager.closedQueries()); n != 0 {
		t.Fatalf("CLOSED queries before the scope guard = %d, want 0", n)
	}
}

// TestA191ClosedQueryNamesTheSymbolAndTheTradingDayWindow (3.10, m12·m13).
func TestA191ClosedQueryNamesTheSymbolAndTheTradingDayWindow(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-15", "PENDING", "0"))}
	d, _, _, _ := a191Fake(t, pager, reader, a191Tracked("o-15"))
	a191Poll(t, d)
	qs := pager.closedQueries()
	if len(qs) != 1 {
		t.Fatalf("CLOSED queries = %d, want 1", len(qs))
	}
	q := qs[0]
	if q.Symbol != a191Symbol || q.From != "2026-03-29" || q.To != "2026-03-31" || q.Limit != 100 {
		t.Fatalf("query = %+v, want symbol %s, 2026-03-29..2026-03-31, limit 100", q, a191Symbol)
	}
}

// TestA191ClosedWalkStopsAtThePageCeiling (3.10) — 4 페이지째가 필요하면 결론 불가(truncated).
func TestA191ClosedWalkStopsAtThePageCeiling(t *testing.T) {
	pager := newA191Pager()
	other := func(id string) json.RawMessage { return a191FakeOrder(id, "CANCELED", "0").json() }
	pager.setClosed("", execgw191Page("c1", other("x-1")))
	pager.setClosed("c1", execgw191Page("c2", other("x-2")))
	pager.setClosed("c2", execgw191Page("c3", other("x-3")))
	pager.setClosed("c3", execgw191Page("", a191FakeOrder("o-16", "CANCELED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-16", "PENDING", "0"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-16"))
	cycle := a191Poll(t, d)
	if lastApplied(t, ledger, "o-16").Derived.Terminal {
		t.Fatal("a row beyond the page ceiling was used")
	}
	if r := reasonsOf(cycle)["o-16"]; r != filldetect.UnresolvedClosedListTruncated {
		t.Fatalf("reason = %q", r)
	}
	if n := len(pager.closedQueries()); n != 3 {
		t.Fatalf("CLOSED pages read = %d, want 3", n)
	}
}

// TestA191ClosedRowWithoutExecutionIsNotAccepted (3.12a, m22 — M6) — execution 이 없거나 filledQuantity 가 null 이면 안 받음.
func TestA191ClosedRowWithoutExecutionIsNotAccepted(t *testing.T) {
	base := `"orderId":"o-17","symbol":"` + a191Symbol + `","side":"SELL","status":"CANCELED","quantity":"10","currency":"KRW","orderedAt":"` + a191OrderedAt + `"`
	for name, row := range map[string]string{
		"no execution":          `{` + base + `}`,
		"null execution":        `{` + base + `,"execution":null}`,
		"null filled quantity":  `{` + base + `,"execution":{"filledQuantity":null}}`,
		"empty filled quantity": `{` + base + `,"execution":{"filledQuantity":""}}`,
		"non-numeric filled":    `{` + base + `,"execution":{"filledQuantity":"garbage"}}`,
		"number-typed filled":   `{` + base + `,"execution":{"filledQuantity":0}}`,
	} {
		t.Run(name, func(t *testing.T) {
			pager := newA191Pager()
			pager.setClosed("", a191ClosedPage(json.RawMessage(row)))
			reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-17", "PENDING", "0"))}
			d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-17"))
			cycle := a191Poll(t, d)
			if lastApplied(t, ledger, "o-17").Derived.Terminal {
				t.Fatal("a row without a filled quantity closed the order")
			}
			if r := reasonsOf(cycle)["o-17"]; r != filldetect.UnresolvedClosedListExecutionMissing {
				t.Fatalf("reason = %q", r)
			}
		})
	}
}

// TestA191OnlyOneClosedLookupPerCycle (3.12a, m23·m26 — M8 ②) — 후보 셋이 세 주기에 걸쳐 하나씩, 미뤄진 것의 lastLookup 은 그대로.
func TestA191OnlyOneClosedLookupPerCycle(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(
		a191FakeOrder("o-18a", "PENDING", "0"), a191FakeOrder("o-18b", "PENDING", "0"), a191FakeOrder("o-18c", "PENDING", "0"))}
	d, clk, _, _ := a191Fake(t, pager, reader, a191Tracked("o-18a"), a191Tracked("o-18b"), a191Tracked("o-18c"))
	for i := 1; i <= 3; i++ {
		cycle := a191Poll(t, d)
		if cycle.ClosedReads != 1 {
			t.Fatalf("cycle %d CLOSED lookups = %d, want 1", i, cycle.ClosedReads)
		}
		deferred := 0
		for _, u := range cycle.Unresolved {
			if u.Reason == filldetect.UnresolvedClosedLookupDeferred {
				deferred++
			}
		}
		if deferred != 2 {
			t.Fatalf("cycle %d deferred = %d, want 2", i, deferred)
		}
		clk.Advance(3 * time.Second)
	}
	if n := len(pager.closedQueries()); n != 3 {
		t.Fatalf("CLOSED queries after 3 cycles = %d, want 3 (one per candidate)", n)
	}
}

// TestA191SlowClosedLookupIsCutAtTheDeadline (3.12a·4.6, m24 — M8 ①) — 멈춘 CLOSED 는 2s 근처에서 끊기고 다른 주문은 같은 주기에 적용됨.
func TestA191SlowClosedLookupIsCutAtTheDeadline(t *testing.T) {
	pager := newA191Pager()
	pager.blockClosed = true
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-19", "PENDING", "0"), a191FakeOrder("o-19b", "FILLED", "10"))}
	d, _, ledger, _ := a191Fake(t, pager, reader, a191Tracked("o-19"), a191Tracked("o-19b"))
	start := time.Now()
	cycle := a191Poll(t, d)
	elapsed := time.Since(start)
	if elapsed > 3*time.Second {
		t.Fatalf("cycle took %v with a stalled CLOSED lookup, want <= 2s + scheduling slack", elapsed)
	}
	if got := lastApplied(t, ledger, "o-19b"); got.Derived.State != "FILLED" {
		t.Fatalf("the other order's snapshot = %+v, want applied in the same cycle", got.Derived)
	}
	if r := reasonsOf(cycle)["o-19"]; r != filldetect.UnresolvedClosedListReadFailed {
		t.Fatalf("reason = %q", r)
	}
}

// TestA191UnresolvedIsReportedOnlyAfterItPersists (3.11, m14) — 30s 전엔 알림 0, 30s 뒤 1, 그 뒤 반복 없음.
func TestA191UnresolvedIsReportedOnlyAfterItPersists(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-20", "PENDING", "0"))}
	d, clk, _, alerts := a191Fake(t, pager, reader, a191Tracked("o-20"))
	a191Poll(t, d)
	clk.Advance(27 * time.Second)
	a191Poll(t, d)
	if n := len(alerts.all()); n != 0 {
		t.Fatalf("alerts before 30s = %d, want 0", n)
	}
	clk.Advance(3 * time.Second)
	a191Poll(t, d)
	clk.Advance(30 * time.Second)
	a191Poll(t, d)
	if n := len(alerts.all()); n != 1 {
		t.Fatalf("alerts after the order stayed unresolved = %d, want exactly 1", n)
	}
}

// TestA191JustPlacedOrderMissingFromTheOpenListIsNotEscalated (3.11, m18) — 다음 주기 OPEN 에 나타나면 상태가 지워져 다시 생겨도 즉시 조회.
func TestA191JustPlacedOrderMissingFromTheOpenListIsNotEscalated(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-21", "PENDING", "0"))}
	d, clk, _, alerts := a191Fake(t, pager, reader, a191Tracked("o-21"))
	a191Poll(t, d) // OPEN 에 없음 → 조회 1
	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.open = page("", a191FakeOrder("o-21", "PENDING", "0"))
	pager.mu.Unlock()
	a191Poll(t, d) // OPEN 에 나타남 → 상태 삭제
	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.open = execgw191Page("")
	pager.mu.Unlock()
	a191Poll(t, d) // 다시 사라짐 → 간격 없이 즉시 조회(상태가 지워졌으므로)
	if n := len(pager.closedQueries()); n != 2 {
		t.Fatalf("CLOSED queries = %d, want 2 (state pruned when it reappeared)", n)
	}
	if n := len(alerts.all()); n != 0 {
		t.Fatalf("alerts after the first episode resolved = %d, want 0", n)
	}
	// 새 사건은 다시 사라진 순간부터 셈 — 30s 전엔 0, 30s 뒤 1.
	clk.Advance(27 * time.Second)
	a191Poll(t, d)
	if n := len(alerts.all()); n != 0 {
		t.Fatalf("alerts 27s into the second episode = %d, want 0", n)
	}
	clk.Advance(3 * time.Second)
	a191Poll(t, d)
	if n := len(alerts.all()); n != 1 {
		t.Fatalf("alerts 30s into the second episode = %d, want 1", n)
	}
}

// TestA191OrderThatStaysInTheOpenListIsNeverReported (코드 리뷰 1 #4) — OPEN 에 계속 있는 주문은 폴링이 이어져도 알림·조회 0.
func TestA191OrderThatStaysInTheOpenListIsNeverReported(t *testing.T) {
	pager := newA191Pager()
	pager.open = page("", a191FakeOrder("o-24", "PENDING", "0"))
	d, clk, _, alerts := a191Fake(t, pager, &fakeOrderReader{}, a191Tracked("o-24"))
	for i := 0; i < 15; i++ {
		a191Poll(t, d)
		clk.Advance(3 * time.Second)
	}
	if n := len(alerts.all()) + len(pager.closedQueries()); n != 0 {
		t.Fatalf("alerts+CLOSED queries for an order that stayed open = %d, want 0", n)
	}
}

// TestA191AcceptedLookupKeepsTheIntervalWhenTheCycleFailsLater (코드 리뷰 1 #2) — 받아들인 뒤 같은 주기의 잔고 sweep 이 실패하면
// 스냅숏은 적용되지 않음; 다음 주기 30s 안에는 다시 묻지 않음.
func TestA191AcceptedLookupKeepsTheIntervalWhenTheCycleFailsLater(t *testing.T) {
	pager := newA191Pager()
	pager.setClosed("", a191ClosedPage(a191FakeOrder("o-25", "CANCELED", "0").json()))
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-25", "PENDING", "0"))}
	d, clk, _, _ := a191Fake(t, pager, reader, a191Tracked("o-25"))
	positions := &fakePositions{err: errA191Throttled}
	d.Positions = positions
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("want the sweep failure to fail the cycle")
	}
	clk.Advance(3 * time.Second)
	positions.mu.Lock()
	positions.err = nil
	positions.mu.Unlock()
	cycle := a191Poll(t, d)
	if n := len(pager.closedQueries()); n != 1 {
		t.Fatalf("CLOSED queries = %d, want 1 (the interval survives an accepted-but-unapplied lookup)", n)
	}
	if r := reasonsOf(cycle)["o-25"]; r != filldetect.UnresolvedClosedLookupDeferred {
		t.Fatalf("reason = %q, want deferred", r)
	}
}

// TestA191UnresolvedAlertCarriesOnlyTheInternalIdentifier (3.12a, m25 — M1) — 실을 것은 intent_id·market·reason·unresolved_seconds 뿐.
func TestA191UnresolvedAlertCarriesOnlyTheInternalIdentifier(t *testing.T) {
	pager := newA191Pager()
	reader := &fakeOrderReader{orders: rawOrders(a191FakeOrder("o-22", "PENDING", "0"))}
	d, clk, _, alerts := a191Fake(t, pager, reader, a191Tracked("o-22"))
	a191Poll(t, d)
	clk.Advance(30 * time.Second)
	a191Poll(t, d)
	events := alerts.all()
	if len(events) != 1 {
		t.Fatalf("alerts = %d, want 1", len(events))
	}
	e := events[0]
	if e.Type != filldetect.EventFillOrderUnresolved || obs.SeverityOf(e.Type) != obs.SeverityNormal {
		t.Fatalf("event type = %q (%s)", e.Type, obs.SeverityOf(e.Type))
	}
	want := map[string]bool{obs.FieldIntentID: true, obs.FieldMarket: true, obs.FieldReason: true, "unresolved_seconds": true}
	if len(e.Fields) != len(want) {
		t.Fatalf("fields = %v, want exactly %v", e.Fields, want)
	}
	for k := range e.Fields {
		if !want[k] {
			t.Fatalf("unexpected alert field %q in %v", k, e.Fields)
		}
	}
	if e.Fields[obs.FieldIntentID] != "intent-o-22" || e.Fields[obs.FieldMarket] != "kr" {
		t.Fatalf("fields = %v", e.Fields)
	}
	var parts []string
	parts = append(parts, e.Title, e.Body, e.Key)
	for _, v := range e.Fields {
		parts = append(parts, toString(v))
	}
	all := strings.Join(parts, " ")
	for _, banned := range []string{a191Symbol, "acct", "10"} {
		if strings.Contains(all, banned) {
			t.Fatalf("alert leaks %q: %q", banned, all)
		}
	}
	if strings.Contains(strings.ReplaceAll(all, "intent-o-22", ""), "o-22") {
		t.Fatalf("alert leaks the broker order id: %q", all)
	}
}

// TestJournalTrackedCarriesTheIntentID (3.12a) — 경보 식별자(내부 intent id)가 저널 추적 행에서 옮겨짐.
func TestJournalTrackedCarriesTheIntentID(t *testing.T) {
	d, j, _ := a191JournalDetector(t, newA191Pager(), &fakeOrderReader{})
	_ = d
	recordConfirmedA191Sell(t, j, "o-23")
	tracked, err := filldetect.JournalTracked{Journal: j, AccountRef: a191Account}.TrackedOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].IntentID != "intent-o-23" {
		t.Fatalf("tracked = %+v, want one order carrying intent-o-23", tracked)
	}
}

func toString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// execgw191Page 는 커서가 있는 CLOSED 페이지임.
func execgw191Page(next string, rows ...json.RawMessage) execgw.OrderPage {
	return execgw.OrderPage{Orders: rows, NextCursor: next, HasNext: next != ""}
}

// TestA191IncompleteOpenOrderNotTrackedIsStillApplied — collect B10(else@426) 진입 + B12(if@426) 거짓 갈래: 정규 범위가 불완전한
// OPEN 행이라도 같은 id 가 추적되지 않으면 주기는 멈추지 않고 그 행이 그대로 적용됨(기존 시험이 밟지 않던 갈래, Manager 검증 2026-11-17).
func TestA191IncompleteOpenOrderNotTrackedIsStillApplied(t *testing.T) {
	pager := newA191Pager()
	incomplete := a191FakeOrder("o-26", "PENDING", "0")
	incomplete.orderedAt = "not-a-time" // 거래일을 못 구해 정규 범위가 불완전해짐
	pager.open = page("", incomplete)
	d, _, ledger, _ := a191Fake(t, pager, &fakeOrderReader{})
	cycle := a191Poll(t, d)
	if cycle.OpenOrders != 1 {
		t.Fatalf("open orders = %d, want 1", cycle.OpenOrders)
	}
	if got := lastApplied(t, ledger, "o-26"); got.TradingDay != "" || got.Derived.State != "OPEN_UNFILLED" {
		t.Fatalf("applied = %+v, want the incomplete-scope row applied as read", got)
	}
	if n := len(pager.closedQueries()); n != 0 {
		t.Fatalf("CLOSED queries = %d, want 0", n)
	}
}
