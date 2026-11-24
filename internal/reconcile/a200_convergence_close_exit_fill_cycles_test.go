package reconcile_test

// a200_convergence_close_exit_fill_cycles_test.go 는 a200 1단계(1.4 W4·W5·W6, 1.5 범위 밖 둘)의 재현 시험임 — 편집 전 코드의
// 동작을 기록했고, 이 파일의 시험은 a200 뒤에도 그대로 맞는 계약임. 실제 원장(임시 DB) · 운영 배선과 같은 훅 리터럴 · 실제 Converger · 실제 Tracker(운영 조립
// gateway.go:271 과 같은 필드 — Terminal 포함)로 ReconcileDriver.RunOnce 의 순서(수렴 → Refresh → Observe, reconcileloop.go:451-464)를
// 손으로 밟고, 엔진 손절 매도의 체결 커밋을 그 순서의 어느 틈에 넣느냐로 창(W2 · W3 · W4 · W5 · W6)을 만듦.
//
// 이 하네스의 hold 는 **무관리** 포지션(exit 상태 · 발의 없음)을 만듦 — 그래서 W4 · W5 는 a200 뒤에도 귀속 실패로 남고(인정 조건 (다)),
// 관리 포지션의 인정은 a200_recognition_cycles_test.go 가 잼. 이름의 「CurrentBehaviour」 는 편집 전부터 같은 동작(W2 · W3 · W6 ·
// 부분 조정)을 뜻함. 픽스처는 합성 값만 씀.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	a200Account = "acct-7"
	a200Symbol  = "SYN200"
	a200Day     = "2026-03-30"
)

type a200World struct {
	t       *testing.T
	ctx     context.Context
	clk     *clock.Fake
	journal *journal.Journal
	tracker *reconcile.Tracker
	conv    *reconcile.Converger
}

func newA200World(t *testing.T) *a200World {
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
	// 운영 배선(app/engine bindApplyHooks)과 같은 리터럴 — OutcomeRefused 싱크만 없음.
	if err := j.SetApplyHooks(journal.ApplyHooks{
		Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill,
		Exit: journal.ApplyExitFill, Costs: costs.DefaultModel(),
	}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	tracker := &reconcile.Tracker{
		Clock: clk, Gate: noStaleGate(clk), Journal: j, AccountRef: a200Account, Terminal: j,
	}
	if err := tracker.Restore(context.Background()); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	w := &a200World{t: t, ctx: context.Background(), clk: clk, journal: j, tracker: tracker,
		conv: &reconcile.Converger{Journal: j, Credit: tracker, AccountRef: a200Account}}
	return w
}

func (w *a200World) now() string { return w.clk.Now().UTC().Format(time.RFC3339) }

// hold 는 엔진 매수 체결로 연 보유 quantity 주와, 그 종목의 엔진 매도 주문 하나(접수 확정)를 만듦.
func (w *a200World) hold(quantity string) {
	w.t.Helper()
	confirmScopedOrder(w.t, w.journal, "intent-entry", "attempt-entry", a200Account, "kr", a200Day, a200Symbol, "BUY", "o-entry")
	w.clk.Advance(time.Second)
	w.fill("o-entry", "BUY", quantity, quantity, true)
	w.clk.Advance(time.Second)
	confirmScopedOrder(w.t, w.journal, "intent-stop", "attempt-stop", a200Account, "kr", a200Day, a200Symbol, "SELL", "o-stop")
	w.clk.Advance(time.Second)
}

func (w *a200World) fill(orderID, side, ordered, filled string, terminal bool) {
	w.t.Helper()
	state := "OPEN_PARTIALLY_FILLED"
	if terminal {
		state = "FILLED"
	}
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: a200Symbol, Market: "kr", TradingDay: a200Day, Side: side,
		AccountRef: a200Account, State: state, Terminal: terminal,
		Quantity: ordered, FilledQuantity: filled, AveragePrice: "100", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("RecordFill(%s %s/%s): %v", orderID, filled, ordered, err)
	}
}

// mismatch 는 수량 불일치(로컬 local · 계좌 broker)와, 계좌 열린 주문 목록에서 사라진 엔진 매도(F4)를 담은 비교임.
func (w *a200World) mismatch(local, broker string) reconcile.Diff {
	return reconcile.Diff{
		AsOf: w.now(), AccountRef: a200Account,
		Quantities: []reconcile.QuantityMismatch{{Symbol: a200Symbol, Local: local, Broker: broker}},
		MissingOrders: []reconcile.LocalOrder{{OrderID: "o-stop", OriginalOrderID: "o-stop", AccountRef: a200Account,
			Symbol: a200Symbol, Market: "kr", TradingDay: a200Day, Side: "SELL"}},
	}
}

func (w *a200World) clean() reconcile.Diff {
	return reconcile.Diff{AsOf: w.now(), AccountRef: a200Account, Matched: 1}
}

// converge · refresh · observe 는 RunOnce 의 세 단계임. 수렴 오류는 RunOnce 처럼 기록만 하고 다음 단계로 감.
func (w *a200World) converge(diff reconcile.Diff) (reconcile.ConvergeReport, error) {
	return w.conv.ConvergeQuantities(w.ctx, diff)
}

func (w *a200World) refresh() error { return w.tracker.Refresh(w.ctx) }

func (w *a200World) observe(diff reconcile.Diff) (reconcile.Outcome, error) {
	return w.tracker.Observe(w.ctx, diff)
}

// cleanCycle 은 다음 주기 하나(시계 1분 전진 · 일치 비교)를 RunOnce 순서로 돌림.
func (w *a200World) cleanCycle() (reconcile.Outcome, error) {
	w.clk.Advance(time.Minute)
	diff := w.clean()
	if _, err := w.converge(diff); err != nil {
		w.t.Fatalf("converge(clean): %v", err)
	}
	if err := w.refresh(); err != nil {
		return reconcile.Outcome{}, err
	}
	return w.observe(diff)
}

// rows 는 그 종목의 reconcile_states 행 전부를 「cause/release_cause/활성」 으로 돌려줌.
func (w *a200World) rows() []string {
	w.t.Helper()
	history, err := w.journal.ReconcileStateHistory(w.ctx, a200Account)
	if err != nil {
		w.t.Fatalf("ReconcileStateHistory: %v", err)
	}
	var out []string
	for _, s := range history {
		if s.Symbol != a200Symbol {
			continue
		}
		state := "active"
		if !s.Active() {
			state = "released:" + s.ReleaseCause
		}
		evidence := s.Evidence
		if i := strings.Index(evidence, "("); i > 0 && strings.HasPrefix(evidence, "position projection") {
			evidence = evidence[:i+len("(SELL_ON_CLOSED)")+1]
		} else if len(evidence) > 60 {
			evidence = evidence[:60]
		}
		out = append(out, s.Cause+"|"+state+"|"+evidence)
	}
	return out
}

func (w *a200World) blocks() []string {
	var out []string
	for _, b := range w.tracker.Blocks() {
		out = append(out, b.Cause+"/"+string(b.Release))
	}
	return out
}

func (w *a200World) logCycle(label string, out reconcile.Outcome, err error) {
	w.t.Helper()
	// EntryAllowed 는 *RejectedError 를 돌려줌 — nil 포인터를 %v 로 찍으면 「execgw: rejected」 로 보이므로 막힘 여부만 찍음.
	w.t.Logf("[%s] err=%v added=%d cleared=%d awaiting=%d blocked=%v rows=%q blocks=%q entry_refused=%v", label, err,
		len(out.Added), len(out.Cleared), len(out.AwaitingAdjustment), out.Blocked, w.rows(), w.blocks(),
		w.tracker.EntryAllowed("kr", a200Symbol) != nil)
}

// runCleanCycles 는 일치 비교 n 주기를 돌려 각 결과를 기록하고 마지막 결과를 돌려줌.
func (w *a200World) runCleanCycles(n int) {
	w.t.Helper()
	for i := 1; i <= n; i++ {
		out, err := w.cleanCycle()
		w.logCycle("일치 주기 "+string(rune('0'+i)), out, err)
	}
}

func (w *a200World) activeCauses() []string {
	var out []string
	for _, r := range w.rows() {
		parts := strings.Split(r, "|")
		if parts[1] == "active" {
			out = append(out, parts[0])
		}
	}
	return out
}

// TestA200AnUnmanagedW4FillStaysAnAttributionFailure: 수렴 조정 커밋 뒤 · Refresh 전 체결 커밋. 이 하네스의 hold 는 무관리 포지션
// (exit 상태 · 발의 없음)이라 인정 조건 (다)를 못 지나 a200 뒤에도 귀속 실패로 남음 — 그것이 맞는 계약(design §2 계속 거절 3).
func TestA200AnUnmanagedW4FillStaysAnAttributionFailure(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	diff := w.mismatch("31", "0")
	report, err := w.converge(diff)
	t.Logf("[수렴] converged=%d refused=%v credited=%v err=%v", len(report.Converged), report.Refused, report.Credited, err)
	w.fill("o-stop", "SELL", "31", "31", true) // W4
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("W4 주기", out, oerr)
	w.runCleanCycles(3)
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseAttributionFailed {
		t.Fatalf("일치 3 주기 뒤 활성 행 = %q, 현재 동작은 ATTRIBUTION_FAILED 하나가 남음", got)
	}
	if w.tracker.EntryAllowed("kr", a200Symbol) == nil {
		t.Fatal("현재 동작은 그 종목 진입이 막혀 있음")
	}
}

// TestA200AnUnmanagedW5FillStaysAnAttributionFailure: Refresh 뒤 · Observe 의 persist 전 체결 커밋 — 무관리 포지션이라 a200 뒤에도
// 오늘과 같음(Observe 오류 F13 · 귀속 실패 고착).
func TestA200AnUnmanagedW5FillStaysAnAttributionFailure(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	w.fill("o-stop", "SELL", "31", "31", true) // W5
	out, oerr := w.observe(diff)
	w.logCycle("W5 주기", out, oerr)
	if oerr == nil || !strings.Contains(oerr.Error(), "already owned by cause "+journal.ReconcileCauseAttributionFailed) {
		t.Fatalf("Observe 오류 = %v, 현재 동작은 persist 가 귀속 실패 행을 권위로 받고 오류(F13)", oerr)
	}
	w.runCleanCycles(3)
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseAttributionFailed {
		t.Fatalf("일치 3 주기 뒤 활성 행 = %q, 현재 동작은 ATTRIBUTION_FAILED 하나가 남음", got)
	}
}

// TestA200CurrentBehaviourW6FillAfterTheQuantityBlockPersisted: 수량 불일치 행이 선 뒤 체결 커밋 — 거절은 행을 안 씀(F10).
func TestA200CurrentBehaviourW6FillAfterTheQuantityBlockPersisted(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("W6 주기(체결 전)", out, oerr)
	w.fill("o-stop", "SELL", "31", "31", true) // W6
	t.Logf("[W6 체결 뒤] rows=%q", w.rows())
	w.runCleanCycles(2)
	if got := w.activeCauses(); len(got) != 0 {
		t.Fatalf("일치 2 주기 뒤 활성 행 = %q, 현재 동작은 수량 불일치가 크레딧으로 풀리고 남는 차단 없음", got)
	}
	for _, r := range w.rows() {
		if strings.HasPrefix(r, journal.ReconcileCauseAttributionFailed) {
			t.Fatalf("rows = %q, W6 에서는 귀속 실패 행이 생기지 않음", w.rows())
		}
	}
}

// TestA200CurrentBehaviourW2FillBeforeTheConvergenceRead (1.5 범위 밖 (i)): 로컬 읽기(비교) 뒤 · 수렴의 Positions 읽기 전 체결.
func TestA200CurrentBehaviourW2FillBeforeTheConvergenceRead(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	// 비교는 체결 전 로컬(31)로 계산됨.
	diff := w.mismatch("31", "0")
	w.fill("o-stop", "SELL", "31", "31", true) // W2
	report, err := w.converge(diff)
	t.Logf("[수렴] converged=%d refused=%v credited=%v err=%v", len(report.Converged), report.Refused, report.Credited, err)
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("W2 주기", out, oerr)
	w.runCleanCycles(5)
	t.Logf("[W2 결과] 일치 5 주기 뒤 활성 = %q", w.activeCauses())
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("일치 5 주기 뒤 활성 = %q, 현재 동작은 크레딧 없는 수량 불일치 고착", got)
	}
}

// a200StaleStore 는 수렴이 워터마크를 읽은 직후(조정 트랜잭션 전)에 체결 하나를 끼워 넣는 ConvergeStore 임 — W3.
type a200StaleStore struct {
	reconcile.ConvergeStore
	between func()
}

func (s *a200StaleStore) FillWatermark(ctx context.Context, symbol string) (int64, error) {
	mark, err := s.ConvergeStore.FillWatermark(ctx, symbol)
	if s.between != nil {
		hook := s.between
		s.between = nil
		hook()
	}
	return mark, err
}

// TestA200CurrentBehaviourW3FillInsideTheConvergenceWindow (1.5 범위 밖 (i)): 수렴의 읽기 뒤 · 조정 커밋 전 체결.
func TestA200CurrentBehaviourW3FillInsideTheConvergenceWindow(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	diff := w.mismatch("31", "0")
	w.conv.Journal = &a200StaleStore{ConvergeStore: w.journal, between: func() {
		w.fill("o-stop", "SELL", "31", "31", true) // W3
	}}
	report, err := w.converge(diff)
	t.Logf("[수렴] converged=%d refused=%v credited=%v err=%v", len(report.Converged), report.Refused, report.Credited, err)
	w.conv.Journal = w.journal
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("W3 주기", out, oerr)
	w.runCleanCycles(5)
	t.Logf("[W3 결과] 일치 5 주기 뒤 활성 = %q", w.activeCauses())
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("일치 5 주기 뒤 활성 = %q, 현재 동작은 크레딧 없는 수량 불일치 고착", got)
	}
}

// TestA200CurrentBehaviourAPartialIntermediateConvergenceThenOverfill (1.5 범위 밖 (ii)): 계좌가 부분 체결 중간값(11)을 보여
// 수렴이 31→11 로 제자리 조정한 뒤, 그 매도의 누적 20 체결(delta 20 > 11)이 옴. 그다음 주기에 계좌는 11(잔량 주문이 아직 일함).
func TestA200CurrentBehaviourAPartialIntermediateConvergenceThenOverfill(t *testing.T) {
	w := newA200World(t)
	w.hold("31")
	diff := w.mismatch("31", "11")
	report, err := w.converge(diff)
	t.Logf("[수렴 31→11] converged=%+v err=%v", report.Converged, err)
	w.fill("o-stop", "SELL", "31", "20", false) // 조정 뒤 · Refresh 전(W4 와 같은 틈)
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("부분 조정 주기", out, oerr)
	local, lerr := reconcile.LocalStateFromJournal(w.ctx, w.journal, a200Account)
	t.Logf("[투영] positions=%v err=%v", local.Positions, lerr)
	w.runCleanCycles(2)
	// 잔량 11 체결(누적 31).
	w.fill("o-stop", "SELL", "31", "31", true)
	local, lerr = reconcile.LocalStateFromJournal(w.ctx, w.journal, a200Account)
	t.Logf("[잔량 체결 뒤 투영] positions=%v err=%v rows=%q", local.Positions, lerr, w.rows())
	t.Logf("[부분 조정 결과] 활성 = %q", w.activeCauses())
	if got := w.activeCauses(); len(got) != 0 {
		t.Fatalf("활성 = %q, 현재 동작은 OVERSELL 수량 불일치가 수렴 크레딧으로 풀림", got)
	}
}
