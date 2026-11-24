package engine_test

// a201_stale_quantity_comparison_repro_test.go 는 a201 1단계(1.4)의 재현 하네스임. 편집 전 동작을 못 박던
// 「CurrentBehaviour」 시험 6 개는 v3 계약으로 뒤집혀 a201_fill_credit_cycles_test.go 로 옮겨졌고(2.2), 이 파일에는
// 그 시험들이 같이 쓰는 하네스(a201World)만 남음.
//
// a200 의 재현은 수렴 → Refresh → Observe 를 손으로 밟았음. 이 파일은 실제 ReconcileDriver.RunOnce 를 거침(수집 2회 ·
// 안정화 · LocalStateFromJournal · Compare · Ingest · 수렴 · Refresh · Observe · 편입 판정). 엔진 매도의 체결 커밋을 주기
// 안의 정해진 지점에 끼우는 방법은 production 편집 없이 **이미 주입 가능한 인터페이스 두 개**를 감싸는 것임:
//
//   - Converger.Journal(reconcile.ConvergeStore) — Positions 호출 직전(W2: 로컬 읽기 뒤 · 수렴의 Positions 읽기 전),
//     Positions 반환 직후(W3-a: Positions 읽기 뒤 · 워터마크 읽기 전), FillWatermark 반환 직후(조정 tx 직전).
//   - Tracker.Journal(reconcile.ReconcileStore) — Refresh 의 ActiveReconcileStates 호출 시점(수렴 뒤 · Observe 전).
//
// 픽스처는 합성 값만 씀.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const a201Day = "2026-03-30"

// a201 끼움 지점 이름.
const (
	a201BeforePositions = "before-positions" // W2
	a201AfterPositions  = "after-positions"  // W3-a
	a201AfterWatermark  = "after-watermark"  // 조정 tx 직전
)

// a201ConvergeStore 는 수렴이 쓰는 원장을 감싸 정해진 지점에서 한 번 체결을 커밋함. 나머지 메서드는 실제 원장 그대로.
type a201ConvergeStore struct {
	*journal.Journal
	at    string
	fire  func()
	fired bool
}

func (s *a201ConvergeStore) once(point string) {
	if s.at == point && !s.fired && s.fire != nil {
		s.fired = true
		s.fire()
	}
}

func (s *a201ConvergeStore) Positions(ctx context.Context, account string) ([]journal.Position, error) {
	s.once(a201BeforePositions)
	p, err := s.Journal.Positions(ctx, account)
	s.once(a201AfterPositions)
	return p, err
}

func (s *a201ConvergeStore) FillWatermark(ctx context.Context, symbol string) (int64, error) {
	w, err := s.Journal.FillWatermark(ctx, symbol)
	s.once(a201AfterWatermark)
	return w, err
}

// a201ReconcileStore 는 추적기 원장을 감싸 Refresh(수렴 뒤 · Observe 전)의 첫 ActiveReconcileStates 에서 한 번 체결을 커밋함.
type a201ReconcileStore struct {
	reconcile.ReconcileStore
	fire  func()
	armed bool
}

func (s *a201ReconcileStore) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	if s.armed && s.fire != nil {
		s.armed = false
		s.fire()
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

type a201World struct {
	*driverHarness
	ctx      context.Context
	converge *a201ConvergeStore
	store    *a201ReconcileStore
}

// newA201World 는 운영 조립(gateway.go bindApplyHooks · Tracker · Converger 리터럴)과 같은 필드로 드라이버를 만듦.
// 다른 점: 진입 게이트의 신선도 요구를 비움(대사 차단만 보이게), 알림 전송기는 기록용 가짜, 수렴 알림기 없음.
func newA201World(t *testing.T) *a201World {
	t.Helper()
	ctx := context.Background()
	clk := clock.NewFake(reconcileLoopNow)
	j, err := journal.Open(ctx, journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{
		Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill,
		Exit: journal.ApplyExitFill, Costs: costs.DefaultModel(),
	}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	store := &a201ReconcileStore{ReconcileStore: j}
	tracker := &reconcile.Tracker{
		Clock: clk, Gate: execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
		Journal: store, AccountRef: reconcileAccount, Terminal: j,
	}
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	conv := &a201ConvergeStore{Journal: j}
	h := &driverHarness{
		t: t, journal: j, clk: clk,
		orders: &fakeOrderPager{}, holdings: &fakeHoldings{}, balance: &fakeBalance{},
		prices: &fakePrices{last: map[string]float64{}}, alerts: &fakeAlerts{},
		tracker: tracker, names: &engine.InstrumentNames{},
	}
	driver, err := engine.NewReconcileDriver(engine.ReconcileDriverOptions{
		Journal: j,
		Collector: &reconcile.Collector{
			Orders: h.orders, Positions: h.holdings, Balance: h.balance,
			Currencies: []string{"KRW"}, AccountRef: reconcileAccount, Clock: clk,
		},
		Tracker:  tracker,
		Ingest:   &reconcile.Ingestor{Journal: j, AccountRef: reconcileAccount, DefaultMarket: "kr"},
		Converge: &reconcile.Converger{Journal: conv, Credit: tracker, AccountRef: reconcileAccount},
		Prices:   h.prices, Alerts: h.alerts, Names: h.names,
		AccountRef: reconcileAccount, Clock: clk, DefaultMarket: "kr",
		Adoption: config.Adoption{Enabled: true, DefaultStopPct: 0.05},
	})
	if err != nil {
		t.Fatalf("NewReconcileDriver: %v", err)
	}
	h.driver = driver
	return &a201World{driverHarness: h, ctx: ctx, converge: conv, store: store}
}

func (w *a201World) now() string { return w.clk.Now().UTC().Format(time.RFC3339) }

func (w *a201World) confirm(account, symbol, side, orderID string) {
	w.t.Helper()
	attempt, err := w.journal.Prepare(w.ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + orderID, AccountRef: account, Market: "kr", TradingDay: a201Day,
			Symbol: symbol, Side: side, OrderType: "LIMIT", Quantity: "1", Price: "100",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-" + orderID,
		},
		Kind: journal.KindPlace, AttemptID: "attempt-" + orderID,
	})
	if err != nil {
		w.t.Fatalf("Prepare(%s): %v", orderID, err)
	}
	if err := attempt.MarkDispatchStarted(w.ctx); err != nil {
		w.t.Fatal(err)
	}
	if err := attempt.MarkAcked(w.ctx, orderID); err != nil {
		w.t.Fatal(err)
	}
	if err := attempt.Settle(w.ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		w.t.Fatal(err)
	}
}

func (w *a201World) fill(account, symbol, orderID, side, quantity string) {
	w.t.Helper()
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: symbol, Market: "kr", TradingDay: a201Day, Side: side,
		AccountRef: account, State: "FILLED", Terminal: true,
		Quantity: quantity, FilledQuantity: quantity, AveragePrice: "100", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("RecordFill(%s %s): %v", orderID, quantity, err)
	}
}

// hold 는 엔진 매수 체결로 연 보유 quantity 주와 그 종목의 엔진 매도 주문(접수 확정 · 아직 체결 없음)을 만듦.
// 계좌도 같은 수량을 보유한 상태에서 시작함.
func (w *a201World) hold(symbol, quantity string) {
	w.t.Helper()
	w.confirm(reconcileAccount, symbol, "BUY", "o-entry-"+symbol)
	w.clk.Advance(time.Second)
	w.fill(reconcileAccount, symbol, "o-entry-"+symbol, "BUY", quantity)
	w.clk.Advance(time.Second)
	w.confirm(reconcileAccount, symbol, "SELL", "o-stop-"+symbol)
	w.clk.Advance(time.Second)
}

// stopFills 는 그 종목의 엔진 매도가 전량 체결됐다는 감지기 커밋을 흉내 냄.
func (w *a201World) stopFills(symbol, quantity string) {
	w.fill(reconcileAccount, symbol, "o-stop-"+symbol, "SELL", quantity)
}

// stopOpenAtBroker 는 그 종목의 엔진 매도가 계좌 열린 주문 목록에 있음을 만듦(체결 전의 정상 상태 — missing-order 없음).
func (w *a201World) stopOpenAtBroker(symbol, quantity string) {
	w.orders.orders = append(w.orders.orders, json.RawMessage(`{"orderId":"o-stop-`+symbol+
		`","market":"kr","orderDate":"`+a201Day+`","symbol":"`+symbol+`","side":"SELL","status":"OPEN","quantity":"`+
		quantity+`","price":"100"}`))
}

// accountSold 는 계좌 쪽이 이미 매도를 반영했음(보유 목록 · 열린 주문 목록에서 사라짐)을 만듦 — 감지기보다 계좌가 앞섬.
func (w *a201World) accountSold() {
	w.holdings.items = nil
	w.orders.orders = nil
}

func (w *a201World) rows(symbol string) []string {
	w.t.Helper()
	history, err := w.journal.ReconcileStateHistory(w.ctx, reconcileAccount)
	if err != nil {
		w.t.Fatalf("ReconcileStateHistory: %v", err)
	}
	var out []string
	for _, s := range history {
		if s.Symbol != symbol {
			continue
		}
		state := "active"
		if !s.Active() {
			state = "released:" + s.ReleaseCause
		}
		out = append(out, s.Cause+"|"+state)
	}
	return out
}

func (w *a201World) activeCauses(symbol string) []string {
	var out []string
	for _, r := range w.rows(symbol) {
		if parts := strings.Split(r, "|"); parts[1] == "active" {
			out = append(out, parts[0])
		}
	}
	return out
}

func (w *a201World) blocks() []string {
	var out []string
	for _, b := range w.tracker.Blocks() {
		out = append(out, b.Symbol+":"+b.Cause+"/"+string(b.Release))
	}
	return out
}

func (w *a201World) logCycle(label, symbol string, c engine.ReconcileCycle) {
	w.t.Helper()
	w.t.Logf("[%s] err=%v stable=%v folded=%d converged=%d closed=%d blocked=%d released=%d awaiting=%d "+
		"unmanaged=%d adopted=%d deferred=%d rows(%s)=%q blocks=%q entry_refused=%v",
		label, c.Err, c.Stable, c.Folded, c.Converged, c.Closed, c.Blocked, c.Released, c.Awaiting,
		c.Unmanaged, c.Adopted, c.Deferred, symbol, w.rows(symbol), w.blocks(),
		w.tracker.EntryAllowed("kr", symbol) != nil)
}

// next 는 운영 루프처럼 한 주기(60초) 뒤 RunOnce 를 돌림.
func (w *a201World) next() engine.ReconcileCycle {
	w.clk.Advance(time.Minute)
	return w.cycle()
}

// settleClean 은 계좌 · 로컬이 모두 비어 일치하는 주기 n 번을 돌려 기록함.
func (w *a201World) settleClean(symbol string, n int) {
	w.t.Helper()
	for i := 1; i <= n; i++ {
		c := w.next()
		w.logCycle("일치 주기 "+string(rune('0'+i)), symbol, c)
		if c.Err != nil {
			w.t.Fatalf("clean cycle %d: %v", i, c.Err)
		}
	}
}

// start 는 보유 31 주 · 계좌 31 주로 깨끗한 주기 하나를 돌린 뒤, 계좌가 매도를 먼저 반영한 상태를 만듦.
func (w *a201World) start(symbol string) {
	w.t.Helper()
	w.hold(symbol, "31")
	w.holds(symbol, "31", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	c := w.next()
	w.logCycle("시작 주기(로컬 31 · 계좌 31)", symbol, c)
	if c.Err != nil || len(w.activeCauses(symbol)) != 0 {
		w.t.Fatalf("start cycle = %+v rows=%q", c, w.rows(symbol))
	}
	w.accountSold()
}

// assertStuck 은 「그 종목 QM 행 1 이 미해제 · 진입 거절」 을 단언함(현재 동작).
func (w *a201World) assertStuck(symbol string) {
	w.t.Helper()
	if got := w.activeCauses(symbol); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		w.t.Fatalf("active rows for %s = %q, want one QUANTITY_MISMATCH (current behaviour)", symbol, got)
	}
	if w.tracker.EntryAllowed("kr", symbol) == nil {
		w.t.Fatalf("entry for %s allowed; current behaviour keeps it refused", symbol)
	}
}

// rebuyRefused 는 사람이 앱으로 그 종목을 다시 사 계좌가 5 주를 보유했을 때 편입되지 않음(자동 손절 없음)을 기록함.
func (w *a201World) rebuyRefused(symbol string) {
	w.t.Helper()
	w.holds(symbol, "5", "100", 100)
	c := w.next()
	w.logCycle("재매수 주기(계좌 5 · 로컬 0)", symbol, c)
	if c.Err != nil {
		w.t.Fatalf("rebuy cycle: %v", c.Err)
	}
	if c.Adopted != 0 || w.position(symbol).Adopted() {
		w.t.Fatalf("re-bought holding adopted (adopted=%d); current behaviour refuses it under the block", c.Adopted)
	}
}

// 1단계 「CurrentBehaviour」 시험 6 개(W2 · W3-a · tx 직전 · W3-b · 앞 크레딧 위 W2 · 판 중단 뒤 종목 B)는 v3 계약으로
// 뒤집혀 a201_fill_credit_cycles_test.go 로 옮겨짐(a201 tasks 2.2).
