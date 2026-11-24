package engine_test

// a197 — 손절을 막는 외부 매수 미체결 취소(기본 꺼진 토글)를 **실제 주문 경로**(엔진 조립의 게이트웨이 · 원장 · 관측기)로 잼.
// 하네스는 a198 의 진짜 게이트웨이 하네스(newA198Broker · buildE2EStack)를 재사용하고, 미체결 목록 판독만 가짜 판독기로 둠.
// 픽스처는 합성 값만(종목 코드 · 번호 · 가격 모두 합성).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// a197Pager 는 미체결 목록 판독기의 가짜임 — 쪽 · 오류 · 막힘을 시험이 정함.
type a197Pager struct {
	mu      sync.Mutex
	pages   []execgw.OrderPage
	err     error
	block   bool
	calls   int
	queries []execgw.OrderQuery
	onCall  func()
}

func (p *a197Pager) OrdersPageRaw(ctx context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	p.calls++
	p.queries = append(p.queries, q)
	block, err, hook := p.block, p.err, p.onCall
	idx := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "c%d", &idx)
	}
	var page execgw.OrderPage
	if idx < len(p.pages) {
		page = p.pages[idx]
	}
	p.mu.Unlock()
	if hook != nil {
		hook()
	}
	if block {
		<-ctx.Done()
		return execgw.OrderPage{}, ctx.Err()
	}
	if err != nil {
		return execgw.OrderPage{}, err
	}
	return page, nil
}

func (p *a197Pager) set(rows ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	page := execgw.OrderPage{}
	for _, r := range rows {
		page.Orders = append(page.Orders, json.RawMessage(r))
	}
	p.pages = []execgw.OrderPage{page}
}

func (p *a197Pager) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// a197Row 는 공식 OPEN 행 모양의 합성 행임.
func a197Row(id, symbol, side, qty, filled string) string {
	return fmt.Sprintf(`{"orderId":%q,"symbol":%q,"side":%q,"quantity":%q,"price":"69000","currency":"KRW","status":"PENDING","execution":{"filledQuantity":%q}}`,
		id, symbol, side, qty, filled)
}

// a197Recorder 는 critical 입구 · 알림 입구의 기록기임. onRecord 는 기록 순간의 상태를 잡으려고 둠.
type a197Recorder struct {
	mu       sync.Mutex
	events   []obs.Event
	onRecord func(obs.Event)
}

func (r *a197Recorder) Notify(_ context.Context, e obs.Event) error {
	r.mu.Lock()
	r.events = append(r.events, e)
	hook := r.onRecord
	r.mu.Unlock()
	if hook != nil {
		hook(e)
	}
	return nil
}

func (r *a197Recorder) RecordCritical(ctx context.Context, e obs.Event, _ time.Duration) error {
	return r.Notify(ctx, e)
}

func (r *a197Recorder) withKeyPrefix(prefix string) []obs.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []obs.Event
	for _, e := range r.events {
		if strings.HasPrefix(e.Key, prefix) {
			out = append(out, e)
		}
	}
	return out
}

func (r *a197Recorder) ofType(t obs.EventType) []obs.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []obs.Event
	for _, e := range r.events {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

type a197Fixture struct {
	*a198Fixture
	pager    *a197Pager
	alerts2  *a197Recorder
	critical *a197Recorder
}

type a197Options struct {
	toggle       bool
	noExitLog    bool
	notifyOff    bool
	nilOpenOrder bool
}

// newA197Fixture 는 진짜 게이트웨이 엔진 위에 토글 · 가짜 미체결 판독기 · 기록기를 단 관측기를 세움.
func newA197Fixture(t *testing.T, o a197Options) *a197Fixture {
	t.Helper()
	base := newA198Fixture(t)
	f := &a197Fixture{a198Fixture: base, pager: &a197Pager{}, alerts2: &a197Recorder{}, critical: &a197Recorder{}}
	f.rebuild(o)
	return f
}

// rebuild 는 같은 원장 · 브로커 위에 관측기를 새로 만듦 — 재시작의 모양.
func (f *a197Fixture) rebuild(o a197Options) {
	f.t.Helper()
	s := f.s
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: s.engine.Journal, Clock: s.clk, AccountRef: s.engine.AccountRef,
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a197-test",
	})
	if err != nil {
		f.t.Fatalf("NewRiskGuardian: %v", err)
	}
	var exitLog *obs.Logger
	if !o.noExitLog {
		exitLog = obs.NewLogger(obs.LogOptions{Writer: f.sink, JSON: true, Clock: s.clk})
	}
	var pager execgw.OrderPager = f.pager
	if o.nilOpenOrder {
		pager = nil
	}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: s.engine.Journal, Prices: e2eClockDomainPrices{next: s.engine.Official, clk: s.clk}, Retrier: s.engine.Retrier,
		Issuer: guardian, Submit: s.engine.Gateway, Alerts: f.alerts2, Critical: f.critical,
		ExitOrderLog: exitLog, NotificationsEnabled: !o.notifyOff,
		ExternalBuyCancel: o.toggle, OpenOrders: pager,
		Costs: costs.DefaultModel(), AccountRef: s.engine.AccountRef, Clock: s.clk,
	})
	if err != nil {
		f.t.Fatalf("NewExitObserver: %v", err)
	}
	s.observer = observer
}

// openAndBlock 는 005930 10주(손절 68000)를 세우고, 다음 손절 매도 n 번을 반대 미체결로 거절하게 하고 손절선 아래를 관측함.
func (f *a197Fixture) openAndBlock(n int) {
	f.t.Helper()
	f.p = f.s.entry("005930", "10", "70000", "68000")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.refuseSells(n)
	f.breach()
}

func (f *a197Fixture) attemptCritical() []obs.Event {
	return f.critical.withKeyPrefix(engine.A197AttemptKeyPrefix)
}
func (f *a197Fixture) unclearedCritical() []obs.Event {
	return f.critical.withKeyPrefix(engine.A197UnclearedKeyPrefix)
}

// §4-1 — 외부 매수 1 → 주기 끝에 「시도」 critical → 취소 1(번호 그대로) → CONFIRMED → 다음 주기 손절 접수.
func TestA197AnExternalBuyIsCancelledAtTheCycleEndAndTheStopFollows(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	cancelsAtAlert := -1
	f.critical.onRecord = func(e obs.Event) {
		if strings.HasPrefix(e.Key, engine.A197AttemptKeyPrefix) && cancelsAtAlert < 0 {
			cancelsAtAlert = len(f.broker.cancels())
		}
	}
	f.openAndBlock(1)
	if got := f.broker.cancels(); len(got) != 1 || got[0] != "EXT-1" {
		t.Fatalf("broker cancels = %v, want exactly EXT-1", got)
	}
	if cancelsAtAlert != 0 {
		t.Fatalf("the attempt critical was recorded after %d cancels, want before the cancel (0)", cancelsAtAlert)
	}
	if got := f.cancelStates(); len(got) != 1 || !strings.HasPrefix(got[0], string(journal.StateConfirmed)) {
		t.Fatalf("cancel attempts = %v, want one CONFIRMED", got)
	}
	if q := f.pager.queries; len(q) != 1 || q[0].Status != "OPEN" || q[0].Symbol != "005930" {
		t.Fatalf("open-order queries = %+v, want one OPEN read filtered to 005930", q)
	}
	f.next()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v after the external buy was cancelled, want the stop", f.sells())
	}
}

// §4-2 — 토글 꺼짐: 판독 0 · 취소 0 · park 0, 알림은 D9 문장.
func TestA197TheToggleOffReadsAndCancelsNothing(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: false})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.openAndBlock(2)
	f.next()
	if n := f.pager.callCount(); n != 0 {
		t.Fatalf("open-order reads = %d with the toggle off, want 0", n)
	}
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v with the toggle off, want none", got)
	}
	if got := f.cancelStates(); len(got) != 0 {
		t.Fatalf("cancel attempts = %v with the toggle off, want none", got)
	}
	if len(f.attemptCritical()) != 0 || len(f.unclearedCritical()) != 0 {
		t.Fatalf("a197 criticals with the toggle off: %v", f.critical.events)
	}
	refused := f.alerts2.ofType(obs.EventExitProposalRefused)
	if len(refused) == 0 {
		t.Fatal("no proposal-refused alert for the opposite-pending refusal")
	}
	body := refused[0].Body
	if !strings.Contains(body, "매수 미체결") || strings.Contains(body, "브로커에 닿기 전에") {
		t.Fatalf("refusal body = %q, want the opposite-pending cause and not 'before reaching the broker'", body)
	}
}

// §4-3 — 외부 매수 2 + 외부 매도 1: 주기당 1 건, 다음 거절 주기에 둘째, 매도 무접촉.
func TestA197OneExternalBuyPerCycleAndNeverASell(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-S", "005930", "SELL", "2", "0"), a197Row("EXT-1", "005930", "BUY", "3", "0"), a197Row("EXT-2", "005930", "BUY", "1", "0"))
	f.openAndBlock(2)
	if got := f.broker.cancels(); len(got) != 1 || got[0] != "EXT-1" {
		t.Fatalf("first cycle cancels = %v, want only EXT-1", got)
	}
	f.next()
	if got := f.broker.cancels(); len(got) != 2 || got[1] != "EXT-2" {
		t.Fatalf("second cycle cancels = %v, want EXT-1 then EXT-2", got)
	}
	for _, c := range f.broker.cancels() {
		if c == "EXT-S" {
			t.Fatal("an external sell was cancelled")
		}
	}
}

// a197EngineOrderNumber 는 엔진 발주 attempt 를 남김 — 종결 상태별로(귀속은 상태 무관 · D2).
func (f *a197Fixture) a197EngineOrderNumber(orderID string, end journal.AttemptState) {
	f.t.Helper()
	ctx := context.Background()
	decisionID := "d-a197-" + orderID
	if _, err := f.s.engine.Journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: f.s.engine.AccountRef, SafetyClass: journal.SafetyClassRiskReducing, Kind: journal.KindPlace,
		Preimage: journal.ReductionIntent{AccountRef: f.s.engine.AccountRef, Market: "kr", Symbol: "005930", Side: "SELL",
			MaxQuantity: "1", Reason: "a197 synthetic engine order"},
		Nonce: "nonce-" + decisionID, IssuedAt: f.s.clk.Now(), ExpiresAt: f.s.clk.Now().Add(time.Hour),
	}); err != nil {
		f.t.Fatalf("RecordDecision: %v", err)
	}
	a, err := f.s.engine.Journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: "i-a197-" + orderID, Market: "kr", TradingDay: "2026-03-29", AccountRef: f.s.engine.AccountRef,
			Symbol: "005930", Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "1", Price: "60000",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-a197-" + orderID},
		Kind: journal.KindPlace, AttemptID: "a-a197-" + orderID, AccountRef: f.s.engine.AccountRef, DecisionID: decisionID,
		SafetyClass: journal.SafetyClassRiskReducing, ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		f.t.Fatalf("Prepare: %v", err)
	}
	must := func(err error) {
		if err != nil {
			f.t.Fatalf("attempt step: %v", err)
		}
	}
	must(a.MarkDispatchStarted(ctx))
	must(a.MarkAcked(ctx, orderID))
	must(a.MarkInDoubt(ctx, journal.ReasonDispatchAmbiguous, "a197 synthetic"))
	r, err := f.s.engine.Journal.Resume(ctx, a.ID())
	must(err)
	switch end {
	case journal.StateFailedConfirmed:
		must(r.ResolveFailed(ctx, "", "a197 synthetic"))
	case journal.StateUnresolvedInDoubt:
		must(r.ResolveUnresolved(ctx, "", "a197 synthetic"))
	}
}

// §4-4 — 엔진 귀속 번호(상태 무관 — 여기서는 종결 FAILED)는 대상이 아님.
func TestA197AnEngineAttributedBuyIsNotATarget(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.a197EngineOrderNumber("ENG-1", journal.StateFailedConfirmed)
	f.pager.set(a197Row("ENG-1", "005930", "BUY", "1", "0"))
	f.openAndBlock(1)
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v, want none — the number is the engine's", got)
	}
}

// §4-5 — 식별 불완전 행이 하나라도 있으면 그 주기 취소 0.
func TestA197AnUnreadableRowStopsEveryCancel(t *testing.T) {
	for name, bad := range map[string]string{
		"no order id":    `{"orderId":"","symbol":"005930","side":"BUY","quantity":"1","price":"1","currency":"KRW","execution":{"filledQuantity":"0"}}`,
		"no side":        `{"orderId":"X-9","symbol":"005930","side":"","quantity":"1","price":"1","currency":"KRW","execution":{"filledQuantity":"0"}}`,
		"other symbol":   a197Row("X-9", "000660", "BUY", "1", "0"),
		"bad quantity":   a197Row("X-9", "005930", "BUY", "x", "0"),
		"not an object":  `[]`,
		"no fill amount": `{"orderId":"X-9","symbol":"005930","side":"BUY","quantity":"1","price":"1","currency":"KRW"}`,
	} {
		t.Run(name, func(t *testing.T) {
			f := newA197Fixture(t, a197Options{toggle: true})
			f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"), bad)
			f.openAndBlock(1)
			if got := f.broker.cancels(); len(got) != 0 {
				t.Fatalf("broker cancels = %v, want none with an unreadable row", got)
			}
		})
	}
}

// §4-6 — 읽지 못함(오류 · 다음 쪽 있음 ∧ 커서 없음 · 쪽 상한 · 시한 초과) → 취소 0.
func TestA197AnIncompleteReadCancelsNothing(t *testing.T) {
	cases := map[string]func(p *a197Pager){
		"read error": func(p *a197Pager) { p.err = errors.New("synthetic read failure") },
		"next page without a cursor": func(p *a197Pager) {
			p.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
			p.pages[0].HasNext = true
		},
		"pages never end": func(p *a197Pager) {
			for i := 0; i < 50; i++ {
				p.pages = append(p.pages, execgw.OrderPage{Orders: []json.RawMessage{json.RawMessage(a197Row(fmt.Sprintf("EXT-%d", i), "005930", "BUY", "1", "0"))},
					HasNext: true, NextCursor: fmt.Sprintf("c%d", i+1)})
			}
		},
		"read times out": func(p *a197Pager) { p.block = true },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			f := newA197Fixture(t, a197Options{toggle: true})
			setup(f.pager)
			start := time.Now()
			f.openAndBlock(1)
			if got := f.broker.cancels(); len(got) != 0 {
				t.Fatalf("broker cancels = %v, want none after an incomplete read", got)
			}
			if elapsed := time.Since(start); elapsed > 10*time.Second {
				t.Fatalf("the cycle took %s, want the read bounded", elapsed)
			}
		})
	}
}

// §4-7 — 매수 0 → 취소 0.
func TestA197NoBuyMeansNoCancel(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-S", "005930", "SELL", "2", "0"))
	f.openAndBlock(1)
	if f.pager.callCount() != 1 {
		t.Fatalf("open-order reads = %d, want 1", f.pager.callCount())
	}
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v, want none", got)
	}
}

// §4-8 — 익절 발의가 같은 사유로 거절 → 판독 0.
func TestA197ATakeProfitRefusalOpensNothing(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.p = f.s.entry("005930", "10", "70000", "68000")
	f.broker.refuseSells(1)
	for _, px := range []string{"70100", "71000", "72000"} {
		f.broker.quote("005930", px)
		f.s.observe()
		f.s.clk.Advance(6 * time.Second)
	}
	if f.pager.callCount() != 0 || len(f.broker.cancels()) != 0 {
		t.Fatalf("reads=%d cancels=%v after a take-profit refusal, want none", f.pager.callCount(), f.broker.cancels())
	}
}

// §4-10 — 한 번만(D7 · M-1): 확정 거절 → 두 번째 critical, 다음 거절 주기 · 재시작 · 날짜 경계 뒤에도 그 번호 재취소 0.
func TestA197ANumberIsCancelledOnceForItsLifetime(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.broker.planCancel("EXT-1", a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-canceled")})
	f.openAndBlock(5)
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("first cycle cancels = %v, want 1", got)
	}
	if got := f.cancelStates(); len(got) != 1 || !strings.HasPrefix(got[0], string(journal.StateFailedConfirmed)) {
		t.Fatalf("cancel attempts = %v, want one FAILED_CONFIRMED", got)
	}
	if n := len(f.unclearedCritical()); n != 1 {
		t.Fatalf("uncleared criticals = %d, want 1", n)
	}
	f.next()
	f.rebuild(a197Options{toggle: true})
	f.next()
	// KR 날짜 경계(KST 자정)를 넘김 — 같은 번호는 여전히 「이미 냈다」.
	f.s.clk.Advance(24 * time.Hour)
	f.next()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("cancels after more refusals, a restart and a date boundary = %v, want still 1", got)
	}
	if n := len(f.attemptCritical()); n != 1 {
		t.Fatalf("attempt criticals = %d, want 1 per (position, order)", n)
	}
	if n := f.pager.callCount(); n < 2 {
		t.Fatalf("open-order reads = %d, want one per refused cycle", n)
	}
}

// §4-11 · §4-19 — 알림 key 는 (포지션, 번호), 계좌 · 주문 번호 미기재, 첫 알림에 인계 문장.
func TestA197TheAttemptAlertCarriesNoAccountOrNumberAndHandsOver(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.openAndBlock(1)
	got := f.attemptCritical()
	if len(got) != 1 {
		t.Fatalf("attempt criticals = %d, want 1", len(got))
	}
	e := got[0]
	if !strings.Contains(e.Key, f.p.ID) {
		t.Fatalf("key %q does not name the position", e.Key)
	}
	text := e.Title + e.Body + fmt.Sprint(e.Fields)
	for _, forbidden := range []string{"EXT-1", f.s.engine.AccountRef, "123-45"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("the alert carries %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(e.Body, "해동") || !strings.Contains(e.Body, "진입") {
		t.Fatalf("the first alert body does not hand over the park procedure: %q", e.Body)
	}
	if e.Type != obs.EventExitLiquidationDelayed {
		t.Fatalf("type = %s, want the existing critical kind %s", e.Type, obs.EventExitLiquidationDelayed)
	}
}

// §4-12 — 5xx → 같은 단계 안에서 park(계정 래치 없음) → 다음 주기 손절 제출 · 같은 종목 진입만 막힘.
func TestA197AnAmbiguousCancelIsParkedWithoutAnAccountLatch(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.openAndBlock(1)
	f.broker.refuseCancels(0)
	if got := f.cancelStates(); len(got) != 1 || !strings.HasPrefix(got[0], string(journal.StateUnresolvedInDoubt)) {
		t.Fatalf("cancel attempts = %v, want one parked UNRESOLVED_IN_DOUBT", got)
	}
	if blocks := f.s.engine.Entry.Blocks(); len(blocks) != 0 {
		if _, ok := blocks[execgw.ReasonUnresolvedInDoubt]; ok {
			t.Fatalf("entry latches = %v, want no account-wide unresolved latch", blocks)
		}
	}
	u := f.unclearedCritical()
	if len(u) != 1 || !strings.Contains(u[0].Body, "attempt-resolve") {
		t.Fatalf("uncleared criticals = %+v, want one carrying the operator procedure", u)
	}
	f.next()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v after the park, want the stop submitted", f.sells())
	}
}

// §4-12 갈래 — 취소가 실행되지 않은 5xx(매수 남음): 다음 손절도 반대 미체결로 거절, 그 번호 재취소 0.
func TestA197AParkedCancelIsNotRetriedWhileTheBuyRemains(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.openAndBlock(3)
	f.broker.refuseCancels(0)
	f.next()
	f.next()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("cancels = %v, want 1 — a parked number is not cancelled again", got)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("sells = %v, want none while the buy still refuses the stop", f.sells())
	}
}

// §4-13 — 주기 끝 실행: 앞 포지션이 열 조건이어도 뒤 포지션 손절은 단계 전에 나감, 판독이 막혀도.
func TestA197TheStageRunsAfterEveryJudgementOfTheCycle(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.s.entryAs("a", "000660", "10", "70000", "68000")
	f.p = f.s.entryAs("b", "005930", "10", "70000", "68000")
	f.broker.quote("000660", "70100")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.pager.set(a197Row("EXT-1", "000660", "BUY", "3", "0"))
	sellsAtRead := -1
	f.pager.onCall = func() {
		if sellsAtRead < 0 {
			sellsAtRead = len(f.sells())
		}
	}
	f.broker.refuseSells(1) // 000660 이 먼저 판정되어 거절됨
	f.broker.quote("000660", "67900")
	f.broker.quote("005930", "67900")
	f.s.observe()
	if sellsAtRead != 1 {
		t.Fatalf("sells already placed when the stage read the open orders = %d, want 1 (the later position's stop first)", sellsAtRead)
	}
}

// §4-13 갈래 — 전용 로거가 없는 조립에서도 단계가 돎(endExitOrderCycle B1 앞).
func TestA197TheStageRunsWithoutTheDedicatedLogger(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true, noExitLog: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.openAndBlock(1)
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("cancels without the exit-order logger = %v, want 1", got)
	}
}

// §4-14 — 같은 주기에 열 조건 둘 → 판독 1 · 취소 ≤ 1, 다음 주기에 다른 포지션.
func TestA197OnePositionPerCycle(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.s.entryAs("a", "000660", "10", "70000", "68000")
	f.p = f.s.entryAs("b", "005930", "10", "70000", "68000")
	f.broker.quote("000660", "70100")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.pager.set(a197Row("EXT-A", "000660", "BUY", "3", "0"), a197Row("EXT-B", "005930", "BUY", "3", "0"))
	f.broker.refuseSells(10)
	f.broker.quote("000660", "67900")
	f.broker.quote("005930", "67900")
	f.s.observe()
	if n := f.pager.callCount(); n != 1 {
		t.Fatalf("reads in one cycle = %d, want 1", n)
	}
	first := f.broker.cancels()
	if len(first) > 1 {
		t.Fatalf("cancels in one cycle = %v, want at most 1", first)
	}
	f.s.clk.Advance(6 * time.Second)
	f.s.observe()
	q := f.pager.queries
	if len(q) != 2 || q[0].Symbol == q[1].Symbol {
		t.Fatalf("queries over two cycles = %+v, want the other position second", q)
	}
}

// §4-15 — 같은 종목에 엔진 발주 park → 판독 0(D1).
func TestA197AParkedEngineOrderOnTheSymbolOpensNothing(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.a197EngineOrderNumber("ENG-P", journal.StateUnresolvedInDoubt)
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.openAndBlock(1)
	if n := f.pager.callCount(); n != 0 {
		t.Fatalf("open-order reads = %d with a parked engine order on the symbol, want 0", n)
	}
}

// §4-15a — 실행 직전 재검증: ctx 가 취소된 주기의 끝에서는 아무것도 하지 않음.
func TestA197ACancelledCycleRunsNoStage(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.p = f.s.entry("005930", "10", "70000", "68000")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.refuseSells(1)
	f.broker.quote("005930", "67900")
	ctx, cancel := context.WithCancel(context.Background())
	f.alerts2.onRecord = func(e obs.Event) {
		if e.Type == obs.EventExitProposalRefused {
			cancel()
		}
	}
	f.s.observer.ObserveOnce(ctx)
	if n := f.pager.callCount(); n != 0 || len(f.broker.cancels()) != 0 {
		t.Fatalf("reads=%d cancels=%v after the cycle's context ended, want none", n, f.broker.cancels())
	}
}

// §4-16 — 한 주기의 요청 수 계측(판독 쪽 수 · 취소 수). 사전 단건 조회는 브로커 어댑터가 취소마다 1 회(precheck.go) — 여기서는 판독 · 취소만 셈.
func TestA197MeasureTheRequestsOfOneCycle(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	start := time.Now()
	f.openAndBlock(1)
	t.Logf("A197 MEASURE reads=%d cancels=%d cycle_wall=%s", f.pager.callCount(), len(f.broker.cancels()), time.Since(start).Round(time.Millisecond))
}

// 알림이 꺼진 엔진은 critical 을 쓰지 않음(a091 과 같은 규칙 — 보낼 수 없는 critical 이 진입을 멈추지 않게).
func TestA197NotificationsOffWriteNoCritical(t *testing.T) {
	f := newA197Fixture(t, a197Options{toggle: true, notifyOff: true})
	f.pager.set(a197Row("EXT-1", "005930", "BUY", "3", "0"))
	f.openAndBlock(1)
	if len(f.broker.cancels()) != 1 {
		t.Fatalf("cancels = %v, want 1", f.broker.cancels())
	}
	if n := len(f.attemptCritical()) + len(f.unclearedCritical()); n != 0 {
		t.Fatalf("a197 criticals with notifications off = %d, want 0", n)
	}
}
