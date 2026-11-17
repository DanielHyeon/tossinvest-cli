package reconcile_test

// a192_order_terminal_release_test.go 는 missing-order 종목 차단의 종결 증명 해제를 실 journal 로 끝까지 돌림
// (a192 design D2~D6, spec 「missing-order 종목 차단은 …」). 픽스처는 합성 값만 씀.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	a192Account = "acct-7"
	a192Symbol  = "SYN001"
	a192Day     = "2026-03-30"
)

var errA192Injected = errors.New("a192: injected storage failure")

// a192Store 는 실 journal 앞에서 EnterReconcile 결과를 세고, 지정한 실패를 끼워 넣음.
type a192Store struct {
	reconcile.ReconcileStore
	entered         []bool
	failNext        string // "before"(쓰기 없이 실패) | "after"(쓰고 나서 실패)
	beforeInsert    func()
	failAccountWide bool
	failActive      bool
}

func (s *a192Store) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if s.failAccountWide && req.Symbol == "" {
		return journal.ReconcileState{}, false, errA192Injected
	}
	if s.beforeInsert != nil {
		hook := s.beforeInsert
		s.beforeInsert = nil
		hook()
	}
	switch s.failNext {
	case "before":
		s.failNext = ""
		return journal.ReconcileState{}, false, errA192Injected
	case "after":
		s.failNext = ""
		state, entered, err := s.ReconcileStore.EnterReconcile(ctx, req)
		s.entered = append(s.entered, entered)
		if err != nil {
			return state, entered, err
		}
		return journal.ReconcileState{}, false, errA192Injected
	}
	state, entered, err := s.ReconcileStore.EnterReconcile(ctx, req)
	s.entered = append(s.entered, entered)
	return state, entered, err
}

func (s *a192Store) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	if s.failActive {
		return nil, errA192Injected
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

// a192Interleaving 은 해제 트랜잭션 직전에 다른 쓰기를 끼워 넣는 Terminal 임.
type a192Interleaving struct {
	*journal.Journal
	beforeRelease func()
	fail          error
	calls         int
}

func (w *a192Interleaving) ReleaseReconcileOnOrderTerminalProof(ctx context.Context, req journal.OrderTerminalProofRequest,
	evidence string, decide func(journal.OrderTerminalProof) bool) (journal.ReconcileState, bool, error) {
	w.calls++
	if w.beforeRelease != nil {
		w.beforeRelease()
	}
	if w.fail != nil {
		return journal.ReconcileState{}, false, w.fail
	}
	return w.Journal.ReleaseReconcileOnOrderTerminalProof(ctx, req, evidence, decide)
}

type a192World struct {
	t       *testing.T
	ctx     context.Context
	clk     *clock.Fake
	journal *journal.Journal
	gate    *execgw.EntryGate
	tracker *reconcile.Tracker
}

func newA192World(t *testing.T, mutate func(*reconcile.Tracker)) *a192World {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC))
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	gate := noStaleGate(clk)
	tracker := &reconcile.Tracker{
		Clock: clk, Gate: gate, Journal: j, Terminal: j,
		MinInterval: 30 * time.Second, MaxFailures: 3, AccountRef: a192Account,
	}
	if mutate != nil {
		mutate(tracker)
	}
	return &a192World{t: t, ctx: context.Background(), clk: clk, journal: j, gate: gate, tracker: tracker}
}

func (w *a192World) now() string { return w.clk.Now().UTC().Format(time.RFC3339) }

func (w *a192World) tick() { w.clk.Advance(time.Minute) }

func (w *a192World) order(orderID, side string) {
	w.t.Helper()
	confirmScopedOrder(w.t, w.journal, "intent-"+orderID, "attempt-"+orderID, a192Account, "kr", a192Day,
		a192Symbol, side, orderID)
}

func a192Missing(orderID string) reconcile.LocalOrder {
	return reconcile.LocalOrder{OrderID: orderID, OriginalOrderID: orderID, AccountRef: a192Account,
		Symbol: a192Symbol, Market: "kr", TradingDay: a192Day, Side: "SELL"}
}

func (w *a192World) missing(orders ...reconcile.LocalOrder) reconcile.Diff {
	return reconcile.Diff{AsOf: w.now(), AccountRef: a192Account, MissingOrders: orders}
}

func (w *a192World) clean() reconcile.Diff {
	return reconcile.Diff{AsOf: w.now(), AccountRef: a192Account}
}

func (w *a192World) observe(diff reconcile.Diff) reconcile.Outcome {
	w.t.Helper()
	out, err := w.tracker.Observe(w.ctx, diff)
	if err != nil {
		w.t.Fatalf("Observe: %v", err)
	}
	return out
}

// cancel 은 감지기가 종결 목록으로 확인한 취소를 장부에 적는 것과 같은 기록임(체결 0).
func (w *a192World) cancel(orderID, side string) {
	w.t.Helper()
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: a192Symbol, Market: "kr", TradingDay: a192Day, Side: side,
		AccountRef: a192Account, State: "CANCELLED", Terminal: true,
		Quantity: "1", FilledQuantity: "0", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("RecordFill(cancel %s): %v", orderID, err)
	}
}

// fill 은 그 종목의 체결 하나(체결 이벤트가 생김)임.
func (w *a192World) fill(orderID, side string) {
	w.t.Helper()
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: a192Symbol, Market: "kr", TradingDay: a192Day, Side: side,
		AccountRef: a192Account, State: "FILLED", Terminal: true,
		Quantity: "1", FilledQuantity: "1", AveragePrice: "100", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("RecordFill(fill %s): %v", orderID, err)
	}
}

func (w *a192World) symbolRow() (journal.ReconcileState, bool) {
	w.t.Helper()
	history, err := w.journal.ReconcileStateHistory(w.ctx, a192Account)
	if err != nil {
		w.t.Fatalf("ReconcileStateHistory: %v", err)
	}
	for _, state := range history {
		if state.Symbol == a192Symbol {
			return state, true
		}
	}
	return journal.ReconcileState{}, false
}

func (w *a192World) assertStillBlocked(label string) {
	w.t.Helper()
	state, ok := w.symbolRow()
	if !ok || !state.Active() {
		w.t.Fatalf("%s: symbol row = %+v (found=%v), want it still active", label, state, ok)
	}
	if w.tracker.EntryAllowed("kr", a192Symbol) == nil {
		w.t.Fatalf("%s: entries on %s are open, want the block kept", label, a192Symbol)
	}
}

// twoCleanComparisons 는 종결 기록 뒤 무장·해제 두 비교를 돌리고 두 결과를 돌려줌.
func (w *a192World) twoCleanComparisons() (reconcile.Outcome, reconcile.Outcome) {
	w.t.Helper()
	w.tick()
	first := w.observe(w.clean())
	w.tick()
	second := w.observe(w.clean())
	return first, second
}

// TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof 는 주 흐름: 이 프로세스가 삽입한 missing-order
// 행은 종결 기록 뒤 첫 일치 비교에서 무장만 하고, 그보다 나중의 일치 비교에서 ORDER_TERMINAL_CONFIRMED 로 풀림.
func TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	if out := w.observe(w.missing(a192Missing("o-a192-1"))); len(out.Added) != 1 {
		t.Fatalf("added = %+v, want the missing-order symbol block", out.Added)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")

	first, second := w.twoCleanComparisons()
	if len(first.Cleared) != 0 || len(first.AwaitingAdjustment) != 1 {
		t.Fatalf("first agreeing comparison = %+v, want the block kept (armed only)", first)
	}
	if len(second.Cleared) != 1 || second.Cleared[0].Symbol != a192Symbol {
		t.Fatalf("second agreeing comparison = %+v, want the block released", second)
	}
	if n, ok := second.Cleared[0].ReleasedByOrderTerminal(); !ok || n != 1 {
		t.Fatalf("released block marker = (%d, %v), want one bound order", n, ok)
	}
	state, ok := w.symbolRow()
	if !ok || state.Active() || state.ReleaseCause != journal.ReconcileReleaseOrderTerminalConfirmed {
		t.Fatalf("symbol row = %+v, want it released with %s", state, journal.ReconcileReleaseOrderTerminalConfirmed)
	}
	if rejected := w.tracker.EntryAllowed("kr", a192Symbol); rejected != nil {
		t.Fatalf("entries on %s still refused after the release: %v", a192Symbol, rejected)
	}
	if rejected := w.gate.CheckEntryFor("kr", a192Symbol); rejected != nil {
		t.Fatalf("the entry gate still latches %s: %v", a192Symbol, rejected)
	}
}

// TestA192WithoutATerminalReaderTheTrackerBehavesAsToday: Terminal 이 nil 이면 같은 흐름에서 해제가 없고
// 오늘처럼 「조정 대기」로만 남음.
func TestA192WithoutATerminalReaderTheTrackerBehavesAsToday(t *testing.T) {
	w := newA192World(t, func(tr *reconcile.Tracker) { tr.Terminal = nil })
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	for i, out := range []reconcile.Outcome{first, second} {
		if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 || !out.Blocked {
			t.Fatalf("comparison %d = %+v, want today's awaiting-adjustment outcome", i+1, out)
		}
	}
	w.assertStillBlocked("no terminal reader")
}

// TestA192ARestartedTrackerNeverReleasesThisWay: 결속은 프로세스 기억 — 재기동한 추적기는 남은 행을 이 경로로
// 풀지 않음(운영자 해제로 돌아감).
func TestA192ARestartedTrackerNeverReleasesThisWay(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))

	// 차단이 걸린 채 재기동 — 그 뒤에도 주문은 한동안 missing 으로 보이다가 종결됨(실제 순서).
	restarted := &reconcile.Tracker{
		Clock: w.clk, Gate: noStaleGate(w.clk), Journal: w.journal, Terminal: w.journal,
		MinInterval: 30 * time.Second, MaxFailures: 3, AccountRef: a192Account,
	}
	if err := restarted.Restore(w.ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	w.tracker, w.gate = restarted, restarted.Gate
	w.tick()
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a restarted tracker released %+v / %+v; the binding must not survive a restart", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("restarted")
}

// TestA192ASameObjectRestoreDropsBindings: 같은 추적기에 Restore 가 다시 불려도 결속·무장은 남지 않음.
func TestA192ASameObjectRestoreDropsBindings(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	w.tick()
	if first := w.observe(w.clean()); len(first.Cleared) != 0 {
		t.Fatalf("first comparison released %+v, want arming only", first.Cleared)
	}
	if err := w.tracker.Restore(w.ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("released after Restore: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("same-object restore")
}

// TestA192ARowAnotherProducerInsertedFirstIsNeverBound 는 Manager 필수 조건 1: 투영 거절이 행을 먼저 만들고,
// 추적기가 Refresh 전에 같은 종목의 missing-order 로 기억에 새 블록을 만들어도 장부는 「삽입 안 함」이라 결속이 없음.
func TestA192ARowAnotherProducerInsertedFirstIsNeverBound(t *testing.T) {
	store := &a192Store{}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		store.ReconcileStore = tr.Journal
		tr.Journal = store
	})
	w.order("o-a192-1", "SELL")
	// 다른 생산자(투영 거절)가 같은 cause 로 먼저 만든 행 — 추적기는 아직 Refresh 하지 않음.
	if _, entered, err := w.journal.EnterReconcile(w.ctx, journal.EnterReconcileRequest{
		AccountRef: a192Account, Symbol: a192Symbol, Cause: journal.ReconcileCauseQuantityMismatch,
		Evidence: "position projection refused transition (synthetic)",
	}); err != nil || !entered {
		t.Fatalf("seeding the other producer's row: entered=%v err=%v", entered, err)
	}
	if out := w.observe(w.missing(a192Missing("o-a192-1"))); len(out.Added) != 1 {
		t.Fatalf("added = %+v, want the tracker to believe it raised the block", out.Added)
	}
	// 대조: 추적기의 쓰기는 실제로 「기존 행」을 만났음.
	if len(store.entered) != 1 || store.entered[0] {
		t.Fatalf("EnterReconcile results = %v, want exactly one write that inserted nothing", store.entered)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("another producer's row was released: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("other producer first")
}

// TestA192AnOverlappingMissingOrderOnAnExistingRowIsNeverBound: 추적기가 이미 아는 행(Refresh 로 들어옴)에
// 겹친 missing-order 는 쓰기 자체가 없어 결속이 없음.
func TestA192AnOverlappingMissingOrderOnAnExistingRowIsNeverBound(t *testing.T) {
	store := &a192Store{}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		store.ReconcileStore = tr.Journal
		tr.Journal = store
	})
	w.order("o-a192-1", "SELL")
	if _, _, err := w.journal.EnterReconcile(w.ctx, journal.EnterReconcileRequest{
		AccountRef: a192Account, Symbol: a192Symbol, Cause: journal.ReconcileCauseQuantityMismatch,
		Evidence: "an earlier process raised this",
	}); err != nil {
		t.Fatal(err)
	}
	if err := w.tracker.Refresh(w.ctx); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	w.observe(w.missing(a192Missing("o-a192-1")))
	if len(store.entered) != 0 {
		t.Fatalf("EnterReconcile results = %v, want no write for a block the tracker already holds", store.entered)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("an overlapping row was released: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("overlap")
}

// TestA192AnUnprovenInsertIsNeverBound: 삽입이 커밋된 뒤 오류가 돌아오면 삽입 증명이 없음 — 재시도는 기존 행을 만남.
func TestA192AnUnprovenInsertIsNeverBound(t *testing.T) {
	store := &a192Store{failNext: "after"}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		store.ReconcileStore = tr.Journal
		tr.Journal = store
	})
	w.order("o-a192-1", "SELL")
	if _, err := w.tracker.Observe(w.ctx, w.missing(a192Missing("o-a192-1"))); !errors.Is(err, errA192Injected) {
		t.Fatalf("first Observe err = %v, want the injected failure", err)
	}
	w.tick()
	w.observe(w.missing(a192Missing("o-a192-1")))
	if len(store.entered) != 2 || !store.entered[0] || store.entered[1] {
		t.Fatalf("EnterReconcile results = %v, want [inserted-but-unacknowledged, existing]", store.entered)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("an unproven insert was released: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("unproven insert")
}

// TestA192ARetriedInsertIsNeverBound(M14): 첫 쓰기가 실패해 pending 으로 남았다가 나중 관측에서 삽입된 블록은
// 결속하지 않음 — 그 사이의 관측(둘째 주문·수량 흔적)이 결속에 없기 때문.
func TestA192ARetriedInsertIsNeverBound(t *testing.T) {
	store := &a192Store{failNext: "before"}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		store.ReconcileStore = tr.Journal
		tr.Journal = store
	})
	w.order("o-a192-1", "SELL")
	if _, err := w.tracker.Observe(w.ctx, w.missing(a192Missing("o-a192-1"))); !errors.Is(err, errA192Injected) {
		t.Fatalf("first Observe err = %v, want the injected failure", err)
	}
	w.tick()
	w.observe(w.missing(a192Missing("o-a192-1")))
	// 대조: 재시도는 실제로 행을 삽입함.
	if len(store.entered) != 1 || !store.entered[0] {
		t.Fatalf("EnterReconcile results = %v, want the retry to insert the row", store.entered)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a retried insert was released: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("retried insert")
}

// TestA192AQuantityDisputeDisqualifiesTheBinding: 행이 살아 있는 동안 같은 종목 수량 불일치가 한 번이라도
// 보이면 이 경로로는 풀리지 않음(수량 규칙이 지배 — I2).
func TestA192AQuantityDisputeDisqualifiesTheBinding(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	diff := w.missing(a192Missing("o-a192-1"))
	diff.Quantities = []reconcile.QuantityMismatch{{Symbol: a192Symbol, Local: "1", Broker: "0"}}
	w.observe(diff)
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a binding that saw a quantity dispute was released: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("quantity dispute")
}

// TestA192EveryBoundOrderMustBeTerminal: 같은 종목 주문 둘이 결속되면 둘 다 종결돼야 풀림.
func TestA192EveryBoundOrderMustBeTerminal(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.order("o-a192-2", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1"), a192Missing("o-a192-2")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("released with one bound order still open: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.tick()
	w.cancel("o-a192-2", "SELL")
	first, second = w.twoCleanComparisons()
	if len(first.Cleared) != 0 || len(second.Cleared) != 1 {
		t.Fatalf("after both cancels = %+v / %+v, want arm then release", first.Cleared, second.Cleared)
	}
}

// TestA192AForeignAccountObservationNeverAuthorisesRelease: 다른 계좌의 missing-order 로 걸린 블록은 차단은
// 하되 해제 근거가 되지 않음.
func TestA192AForeignAccountObservationNeverAuthorisesRelease(t *testing.T) {
	w := newA192World(t, nil)
	confirmScopedOrder(t, w.journal, "intent-foreign", "attempt-foreign", "acct-other", "kr", a192Day,
		a192Symbol, "SELL", "o-a192-foreign")
	foreign := a192Missing("o-a192-foreign")
	foreign.AccountRef = "acct-other"
	if out := w.observe(w.missing(foreign)); len(out.Added) != 1 {
		t.Fatalf("added = %+v, want the fail-closed block", out.Added)
	}
	w.tick()
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: "o-a192-foreign", Symbol: a192Symbol, Market: "kr", TradingDay: a192Day, Side: "SELL",
		AccountRef: "acct-other", State: "CANCELLED", Terminal: true, Quantity: "1", FilledQuantity: "0",
		ObservedAt: w.now(),
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a foreign-account observation released the block: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("foreign account")
}

// TestA192AnAmendmentAfterBindingKeepsTheBlock: 결속 뒤 그 번호를 겨눈 정정이 생기면 유지(lineage 는 판정 때 다시 봄).
func TestA192AnAmendmentAfterBindingKeepsTheBlock(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	if _, err := w.journal.Prepare(w.ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-amend", AccountRef: a192Account, Market: "kr", TradingDay: a192Day,
			Symbol: a192Symbol, Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "99",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-intent-amend",
		},
		Kind: journal.KindAmend, AttemptID: "attempt-amend", TargetOrderID: "o-a192-1",
	}); err != nil {
		t.Fatalf("Prepare(AMEND): %v", err)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("an amended order released the block: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("amended")
}

// TestA192AnOpenOrderWithTheSameIDKeepsTheBlock: 장부는 종결인데 브로커 open 목록에 같은 번호가 보이면 유지.
// 공백이 다른 번호는 다른 주문이라 막지 않음(대조).
func TestA192AnOpenOrderWithTheSameIDKeepsTheBlock(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	withOpen := func(id string) reconcile.Diff {
		d := w.clean()
		d.ExternalOrd = []reconcile.ExternalOrder{{BrokerOrder: reconcile.BrokerOrder{OrderID: id, Symbol: a192Symbol},
			Provenance: reconcile.ProvenanceExternal}}
		return d
	}
	for i := 0; i < 2; i++ {
		w.tick()
		if out := w.observe(withOpen("o-a192-1")); len(out.Cleared) != 0 {
			t.Fatalf("released while the broker still shows the order open: %+v", out.Cleared)
		}
	}
	w.tick()
	first := w.observe(withOpen(" o-a192-1"))
	w.tick()
	second := w.observe(withOpen(" o-a192-1"))
	if len(first.Cleared) != 0 || len(second.Cleared) != 1 {
		t.Fatalf("a differently spelled open id = %+v / %+v, want arm then release", first.Cleared, second.Cleared)
	}
}

// TestA192AFillAfterTheBlockKeepsIt: 행을 삽입하기 전 워터마크 뒤에 그 종목 체결이 커밋되면 유지 —
// 삽입 뒤의 체결과, 워터마크 읽기와 삽입 사이의 체결 둘 다.
func TestA192AFillAfterTheBlockKeepsIt(t *testing.T) {
	t.Run("after the insert", func(t *testing.T) {
		w := newA192World(t, nil)
		w.order("o-a192-1", "SELL")
		w.order("o-a192-buy", "BUY")
		w.observe(w.missing(a192Missing("o-a192-1")))
		w.tick()
		w.fill("o-a192-buy", "BUY")
		w.cancel("o-a192-1", "SELL")
		first, second := w.twoCleanComparisons()
		if len(first.Cleared)+len(second.Cleared) != 0 {
			t.Fatalf("released after a later fill: %+v / %+v", first.Cleared, second.Cleared)
		}
		w.assertStillBlocked("fill after insert")
	})
	t.Run("between the watermark and the insert", func(t *testing.T) {
		store := &a192Store{}
		w := newA192World(t, func(tr *reconcile.Tracker) {
			store.ReconcileStore = tr.Journal
			tr.Journal = store
		})
		w.order("o-a192-1", "SELL")
		w.order("o-a192-buy", "BUY")
		filled := false
		store.beforeInsert = func() { w.fill("o-a192-buy", "BUY"); filled = true }
		w.observe(w.missing(a192Missing("o-a192-1")))
		if !filled {
			t.Fatal("the interleaved fill did not run")
		}
		w.tick()
		w.cancel("o-a192-1", "SELL")
		first, second := w.twoCleanComparisons()
		if len(first.Cleared)+len(second.Cleared) != 0 {
			t.Fatalf("released after a fill between watermark and insert: %+v / %+v", first.Cleared, second.Cleared)
		}
		w.assertStillBlocked("fill before insert")
	})
}

// TestA192ABindingChangeRearms: 무장 뒤 결속에 주문이 늘면 다시 무장해야 하고 한 주기 일찍 풀리지 않음.
func TestA192ABindingChangeRearms(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.order("o-a192-2", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 0 {
		t.Fatalf("armed comparison released %+v", out.Cleared)
	}
	w.tick()
	w.observe(w.missing(a192Missing("o-a192-2")))
	w.tick()
	w.cancel("o-a192-2", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared) != 0 {
		t.Fatalf("released on the first comparison after the binding grew: %+v", first.Cleared)
	}
	if len(second.Cleared) != 1 {
		t.Fatalf("second comparison after re-arming = %+v, want the release", second.Cleared)
	}
}

// TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock: 1차 판정 뒤·해제 tx 앞에 그 종목 체결이 커밋되면
// tx 안의 재판정이 해제를 거절함.
func TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock(t *testing.T) {
	interleave := &a192Interleaving{}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		interleave.Journal = tr.Terminal.(*journal.Journal)
		tr.Terminal = interleave
	})
	w.order("o-a192-1", "SELL")
	w.order("o-a192-buy", "BUY")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	interleave.beforeRelease = func() { w.fill("o-a192-buy", "BUY"); interleave.beforeRelease = nil }
	first, second := w.twoCleanComparisons()
	if interleave.calls != 1 {
		t.Fatalf("release attempts = %d, want the release path reached exactly once", interleave.calls)
	}
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a stale verdict released the block: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 0 {
		t.Fatalf("released after the interleaved fill: %+v", out.Cleared)
	}
	w.assertStillBlocked("interleaved fill")
}

// TestA192AnAuthorityErrorStillRecordsTheTaint: 외부 포지션만 있는 비교(깨끗한 갈래)가 영구 보류 권위 읽기
// 실패로 일찍 돌아가도 실격 흔적은 남음.
func TestA192AnAuthorityErrorStillRecordsTheTaint(t *testing.T) {
	store := &a192Store{}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		store.ReconcileStore = tr.Journal
		tr.Journal = store
		tr.MaxFailures = 2
	})
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	// 같은 분쟁 둘째 관측이 영구 승격을 만들고, 그 쓰기를 실패시켜 pending 영구를 남김.
	store.failAccountWide = true
	w.tick()
	if _, err := w.tracker.Observe(w.ctx, w.missing(a192Missing("o-a192-1"))); !errors.Is(err, errA192Injected) {
		t.Fatalf("threshold Observe err = %v, want the injected permanent write failure", err)
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	// 권위 읽기가 실패하는 동안, 그 종목의 외부 포지션만 있는 비교.
	store.failActive = true
	w.tick()
	external := w.clean()
	external.ExternalPos = []reconcile.ExternalPosition{{Holding: reconcile.Holding{Symbol: a192Symbol, Quantity: "1"},
		Provenance: reconcile.ProvenanceExternal}}
	if _, err := w.tracker.Observe(w.ctx, external); !errors.Is(err, errA192Injected) {
		t.Fatalf("external-only Observe err = %v, want the authority read failure", err)
	}
	store.failActive, store.failAccountWide = false, false
	first, second := w.twoCleanComparisons()
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("the taint was lost behind the authority error: %+v / %+v", first.Cleared, second.Cleared)
	}
	w.assertStillBlocked("authority error")
}

// TestA192APermanentBlockStaysOperatorOnly: 같은 missing-order 분쟁 3회로 생긴 계좌 영구 행은 종결 증명으로
// 풀리지 않음 — 종목 행만 풀릴 수 있음(P3).
func TestA192APermanentBlockStaysOperatorOnly(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	for i := 0; i < 3; i++ {
		if i > 0 {
			w.tick()
		}
		w.observe(w.missing(a192Missing("o-a192-1")))
	}
	if !w.tracker.Permanent() {
		t.Fatal("three observations of the same missing order must promote the account block")
	}
	w.tick()
	w.cancel("o-a192-1", "SELL")
	_, second := w.twoCleanComparisons()
	if len(second.Cleared) != 1 || second.Cleared[0].Symbol != a192Symbol {
		t.Fatalf("second comparison = %+v, want only the symbol row released", second.Cleared)
	}
	if !w.tracker.Permanent() || w.tracker.EntryAllowed("kr", "SYNX") == nil {
		t.Fatal("the account-wide permanent block must stay until an operator clears it")
	}
}

// TestA192TheAdjustmentRulesHoldWithTheReaderOn 은 판정을 켠 추적기로 기존 안전 흐름을 다시 돌림(a083):
// 조정 없는 수량 일치는 풀지 않고, 조정 크레딧 해제는 여전히 ADJUSTMENT_APPLIED 임.
func TestA192TheAdjustmentRulesHoldWithTheReaderOn(t *testing.T) {
	t.Run("coincidental agreement", func(t *testing.T) {
		w := newA192World(t, nil)
		diff := w.clean()
		diff.Quantities = []reconcile.QuantityMismatch{{Symbol: a192Symbol, Local: "10", Broker: "7"}}
		w.observe(diff)
		first, second := w.twoCleanComparisons()
		if len(first.Cleared)+len(second.Cleared) != 0 {
			t.Fatalf("an agreement nobody wrote anything for released: %+v / %+v", first.Cleared, second.Cleared)
		}
	})
	t.Run("credited release", func(t *testing.T) {
		w := newA192World(t, nil)
		diff := w.clean()
		diff.Quantities = []reconcile.QuantityMismatch{{Symbol: a192Symbol, Local: "10", Broker: "7"}}
		w.observe(diff)
		w.tracker.AdjustmentApplied(diff.AsOf, a192Symbol)
		w.tick()
		out := w.observe(w.clean())
		if len(out.Cleared) != 1 {
			t.Fatalf("credited comparison = %+v, want the adjustment release", out)
		}
		if _, byTerminal := out.Cleared[0].ReleasedByOrderTerminal(); byTerminal {
			t.Fatal("a credited release must not be marked as an order-terminal release")
		}
		state, ok := w.symbolRow()
		if !ok || state.ReleaseCause != journal.ReconcileReleaseAdjustmentApplied {
			t.Fatalf("symbol row = %+v, want ADJUSTMENT_APPLIED", state)
		}
	})
}

// TestA192ReleaseWriteFailureKeepsTheGate: 종결 증명 해제의 장부 기록이 실패하면 Observe 가 오류를 돌려주고
// 블록·게이트는 그대로 남음(spec 「해제 기록이 실패한다」 — 편입 판정으로 가지 않는 것은 RunOnce 의 기존 오류 갈래).
func TestA192ReleaseWriteFailureKeepsTheGate(t *testing.T) {
	interleave := &a192Interleaving{fail: errA192Injected}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		interleave.Journal = tr.Terminal.(*journal.Journal)
		tr.Terminal = interleave
	})
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	w.tick()
	w.observe(w.clean())
	w.tick()
	out, err := w.tracker.Observe(w.ctx, w.clean())
	if !errors.Is(err, errA192Injected) {
		t.Fatalf("Observe err = %v, want the injected release failure", err)
	}
	if interleave.calls != 1 || len(out.Cleared) != 0 {
		t.Fatalf("calls = %d, cleared = %+v; want one failed attempt and nothing released", interleave.calls, out.Cleared)
	}
	w.assertStillBlocked("release write failure")
	if rejected := w.gate.CheckEntryFor("kr", a192Symbol); rejected == nil {
		t.Fatal("the entry gate opened although the release did not commit")
	}
}

// TestA192AFillBeforeTheBlockDoesNotKeepIt: 행을 삽입하기 **전**에 커밋된 그 종목 체결은 해제를 막지 않음 —
// 워터마크는 삽입 전 값이라 그 체결을 「뒤」로 세지 않음(정상 입력).
func TestA192AFillBeforeTheBlockDoesNotKeepIt(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.order("o-a192-buy", "BUY")
	w.fill("o-a192-buy", "BUY")
	w.tick()
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	first, second := w.twoCleanComparisons()
	if len(first.Cleared) != 0 || len(second.Cleared) != 1 {
		t.Fatalf("with a fill committed before the block = %+v / %+v, want arm then release", first.Cleared, second.Cleared)
	}
}

// TestA192ACreditWinsOverTheTerminalProof: 같은 블록에 쓸 수 있는 조정 크레딧이 있으면 수량 규칙이 먼저 —
// 해제 원인은 ADJUSTMENT_APPLIED 임(I2).
func TestA192ACreditWinsOverTheTerminalProof(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 0 {
		t.Fatalf("arming comparison released %+v", out.Cleared)
	}
	w.tracker.AdjustmentApplied(w.now(), a192Symbol)
	w.tick()
	out := w.observe(w.clean())
	if len(out.Cleared) != 1 {
		t.Fatalf("credited comparison = %+v, want a release", out.Cleared)
	}
	if _, byTerminal := out.Cleared[0].ReleasedByOrderTerminal(); byTerminal {
		t.Fatal("the credit must win over the terminal proof")
	}
	state, ok := w.symbolRow()
	if !ok || state.ReleaseCause != journal.ReconcileReleaseAdjustmentApplied {
		t.Fatalf("symbol row = %+v, want ADJUSTMENT_APPLIED", state)
	}
	// 조정으로 풀린 블록의 결속·유지 사유는 남지 않음(코드 리뷰 2 #1).
	if reasons := w.tracker.OrderTerminalHoldReasons(); len(reasons) != 0 {
		t.Fatalf("hold reasons after the credited release = %v, want none", reasons)
	}
}

// TestA192AForeignComparisonDoesNotDropTheBinding: 계좌가 다른 비교는 해제 근거가 아니지만 결속을 버리지도 않음 —
// 그 뒤 제 계좌의 비교들은 정상대로 무장·해제함(판정에 넘기는 계좌는 추적기 계좌임).
func TestA192AForeignComparisonDoesNotDropTheBinding(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	w.tick()
	foreign := w.clean()
	foreign.AccountRef = "acct-other"
	if out := w.observe(foreign); len(out.Cleared) != 0 {
		t.Fatalf("a foreign-account comparison released %+v", out.Cleared)
	}
	first, second := w.twoCleanComparisons()
	if len(first.Cleared) != 0 || len(second.Cleared) != 1 {
		t.Fatalf("own-account comparisons after a foreign one = %+v / %+v, want arm then release", first.Cleared, second.Cleared)
	}
}

// TestA192AFailClosedFlipBetweenVerdictAndReleaseKeepsTheBlock: 1차 판정 뒤·해제 tx 앞에 결속 주문이 체결 이벤트 없이
// fail_closed 로 바뀌어도(브로커 상태 모호) tx 안의 재판정이 해제를 거절하고, 그 사유를 남김(코드 리뷰 1 #2·#3).
func TestA192AFailClosedFlipBetweenVerdictAndReleaseKeepsTheBlock(t *testing.T) {
	interleave := &a192Interleaving{}
	w := newA192World(t, func(tr *reconcile.Tracker) {
		interleave.Journal = tr.Terminal.(*journal.Journal)
		tr.Terminal = interleave
	})
	w.order("o-a192-1", "SELL")
	w.observe(w.missing(a192Missing("o-a192-1")))
	w.tick()
	w.cancel("o-a192-1", "SELL")
	before, err := w.journal.MaxFillEventID(w.ctx)
	if err != nil {
		t.Fatal(err)
	}
	interleave.beforeRelease = func() {
		interleave.beforeRelease = nil
		_, _ = w.journal.RecordFill(w.ctx, journal.FillObservation{
			OrderID: "o-a192-1", Symbol: a192Symbol, Market: "kr", TradingDay: a192Day, Side: "SELL",
			AccountRef: a192Account, State: "UNKNOWN_BROKER_STATE", FailClosed: true, Reason: "synthetic",
			Quantity: "1", FilledQuantity: "0", ObservedAt: w.now(),
		})
	}
	first, second := w.twoCleanComparisons()
	if interleave.calls != 1 {
		t.Fatalf("release attempts = %d, want the release path reached once", interleave.calls)
	}
	// 대조: 끼워 넣은 쓰기는 체결 이벤트 없이 fail_closed 만 세웠음.
	record, err := w.journal.LookupFillScoped(w.ctx, journal.FillSnapshotScope{OrderID: "o-a192-1", AccountRef: a192Account,
		Market: "kr", TradingDay: a192Day, Symbol: a192Symbol, Side: "SELL"})
	if err != nil || !record.Terminal || !record.FailClosed {
		t.Fatalf("snapshot after the flip = %+v err=%v, want terminal and fail_closed", record, err)
	}
	if after, err := w.journal.MaxFillEventID(w.ctx); err != nil || after != before {
		t.Fatalf("fill events %d → %d (err %v), want none added", before, after, err)
	}
	if len(first.Cleared)+len(second.Cleared) != 0 {
		t.Fatalf("a stale verdict released the block: %+v / %+v", first.Cleared, second.Cleared)
	}
	if reason := w.tracker.OrderTerminalHoldReasons()[a192Symbol]; reason != "tx_not_clean_terminal" {
		t.Fatalf("hold reason = %q, want the in-transaction reason recorded", reason)
	}
	w.assertStillBlocked("fail-closed flip")
}
