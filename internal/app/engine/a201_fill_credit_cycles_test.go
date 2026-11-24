package engine_test

// a201_fill_credit_cycles_test.go 는 a201 v3 계약(design §3 1~9 · 13)을 실제 ReconcileDriver.RunOnce 로 고정함.
//
// 하네스는 1단계 재현 파일(a201_stale_quantity_comparison_repro_test.go)의 a201World 를 그대로 씀 — 체결을 주기 안
// 정해진 지점(로컬 읽기 뒤 · 수렴 Positions 앞뒤 · 조정 tx 직전 · Refresh)에 끼움. 1단계의 「CurrentBehaviour」 시험은
// 계약이 뒤집혀 이 파일의 시험으로 옮겨짐(W2 · W3-a · tx 직전 · W3-b · 앞 크레딧 위 W2 → 다음 깨끗한 주기 FILL_APPLIED,
// 판 중단 뒤 종목 B 는 남는 것 그대로). 계좌 열린 주문 목록은 바꿀 때마다 비우고 다시 채움(1단계 「중복 행」 결함).
// 픽스처는 합성 값만 씀.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a201FillApplied 는 새 해제 원인의 장부 문자열임 — 편집 전 코드에서도 컴파일되도록 상수 대신 글자로 고정함.
const a201FillApplied = "FILL_APPLIED"

// releasedAs 는 그 종목의 마지막 QM 행이 cause 로 풀렸고 미해제 행이 없으며 진입이 열렸음을 단언함.
func (w *a201World) releasedAs(symbol, cause string) {
	w.t.Helper()
	rows := w.rows(symbol)
	want := journal.ReconcileCauseQuantityMismatch + "|released:" + cause
	if len(rows) == 0 || rows[len(rows)-1] != want || len(w.activeCauses(symbol)) != 0 {
		w.t.Fatalf("rows for %s = %q, want the last one %q and none active", symbol, rows, want)
	}
	if rejected := w.tracker.EntryAllowed("kr", symbol); rejected != nil {
		w.t.Fatalf("entry for %s still refused after the release: %v", symbol, rejected)
	}
}

// releasedByFill 은 다음 일치 주기 한 번에 그 종목 차단이 FILL_APPLIED 로 풀림을 단언함.
func (w *a201World) releasedByFill(symbol string) {
	w.t.Helper()
	c := w.next()
	w.logCycle("다음 일치 주기", symbol, c)
	if c.Err != nil || c.Released != 1 {
		w.t.Fatalf("next clean cycle = %+v, want one release", c)
	}
	w.releasedAs(symbol, a201FillApplied)
}

// rebuyAdopted 는 해제 뒤 사람이 다시 산 5 주가 편입됨(자동 손절 대상이 됨)을 단언함.
func (w *a201World) rebuyAdopted(symbol string) {
	w.t.Helper()
	w.holds(symbol, "5", "100", 100)
	c := w.next()
	w.logCycle("재매수 주기(계좌 5 · 로컬 0)", symbol, c)
	if c.Err != nil || c.Adopted != 1 || !w.position(symbol).Adopted() {
		w.t.Fatalf("rebuy cycle = %+v, want the re-bought holding adopted once the block is released", c)
	}
}

// openAtBroker 는 주문 하나가 계좌 열린 주문 목록에 있음을 더함(종목 · 주문번호 지정).
func (w *a201World) openAtBroker(orderID, symbol, quantity string) {
	w.orders.orders = append(w.orders.orders, json.RawMessage(`{"orderId":"`+orderID+`","market":"kr","orderDate":"`+
		a201Day+`","symbol":"`+symbol+`","side":"SELL","status":"OPEN","quantity":"`+quantity+`","price":"100"}`))
}

// cancel 은 감지기가 종결 목록으로 확인한 취소(체결 0)를 장부에 적음 — 체결 이벤트는 생기지 않음.
func (w *a201World) cancel(symbol, orderID string) {
	w.t.Helper()
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: symbol, Market: "kr", TradingDay: a201Day, Side: "SELL",
		AccountRef: reconcileAccount, State: "CANCELLED", Terminal: true,
		Quantity: "1", FilledQuantity: "0", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("RecordFill(cancel %s): %v", orderID, err)
	}
}

// --- 1단계 재현 6 의 v3 판(design §3-1) ------------------------------------------------------------

// TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment: 로컬 읽기 뒤 · 수렴 Positions 읽기 전 체결(W2).
// 그 주기에는 차단이 지금과 같이 서고, 다음 깨끗한 주기에 FILL_APPLIED 로 풀리며 재매수가 편입됨.
func TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }

	c := w.next()
	w.logCycle("W2 주기", symbol, c)
	if !w.converge.fired {
		t.Fatal("the fill hook never ran; the cycle did not reach the convergence")
	}
	if c.Err != nil || c.Converged != 0 || c.Released != 0 {
		t.Fatalf("W2 cycle = %+v, want no error, no convergence and no release", c)
	}
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
	w.rebuyAdopted(symbol)
}

// TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment: Positions 읽기 뒤 · 워터마크 읽기 전 체결(W3-a).
func TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201AfterPositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }

	c := w.next()
	w.logCycle("W3-a 주기", symbol, c)
	if c.Err == nil || !strings.Contains(c.Err.Error(), "expected quantity 31 but found 0") {
		t.Fatalf("W3-a cycle err = %v, want the quantity revalidation's stale", c.Err)
	}
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
}

// TestA201AFillJustBeforeTheAdjustmentTxIsCreditedLikeAnAdjustment: 워터마크 읽기 뒤 · 조정 tx 직전 체결.
func TestA201AFillJustBeforeTheAdjustmentTxIsCreditedLikeAnAdjustment(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201AfterWatermark
	w.converge.fire = func() { w.stopFills(symbol, "31") }

	c := w.next()
	w.logCycle("tx 직전 체결 주기", symbol, c)
	if c.Err == nil || !strings.Contains(c.Err.Error(), "expected quantity 31 but found 0") {
		t.Fatalf("cycle err = %v, want the quantity revalidation (a)", c.Err)
	}
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
}

// TestA201W3bWatermarkOnlyStaleThenTheFillIsCreditedLikeAnAdjustment: 워터마크 재검증 (b)만 걸리는 합성 모양.
// 다른 계좌의 같은 종목 체결이 조정 tx 직전에 워터마크만 움직이고, 엔진 매도 체결은 그 뒤 Refresh 에 커밋됨.
func TestA201W3bWatermarkOnlyStaleThenTheFillIsCreditedLikeAnAdjustment(t *testing.T) {
	const symbol = "SYN201"
	const other = "acct-other"
	w := newA201World(t)
	w.start(symbol)
	w.confirm(other, symbol, "BUY", "o-other")
	w.converge.at = a201AfterWatermark
	w.converge.fire = func() { w.fill(other, symbol, "o-other", "BUY", "1") }
	w.store.armed = true
	w.store.fire = func() { w.stopFills(symbol, "31") }

	c := w.next()
	w.logCycle("W3-b 주기", symbol, c)
	if c.Err == nil || !strings.Contains(c.Err.Error(), "fill watermark") {
		t.Fatalf("W3-b cycle err = %v, want the watermark revalidation's stale", c.Err)
	}
	if w.store.armed {
		t.Fatal("the Refresh hook never ran")
	}
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
}

// TestA201AFillCreditReplacesTheEarlierCreditTheStaleComparisonRefuted: 이미 걸린 차단 위의 경합(앞 크레딧 위 W2).
// 사람의 앱 매도 11 주로 31 → 20 조정(크레딧) · QM 이 선 다음 주기에 엔진 매도 20 주 체결이 로컬 읽기 뒤에 커밋됨.
// 낡은 비교가 앞 크레딧을 반박하는 것은 지금과 같고, 그 주기의 체결 credit 이 다음 깨끗한 주기에 풂.
func TestA201AFillCreditReplacesTheEarlierCreditTheStaleComparisonRefuted(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.holds(symbol, "20", "100", 100)
	w.orders.orders = nil
	w.stopOpenAtBroker(symbol, "31")
	c := w.next()
	w.logCycle("외부 매도 주기(로컬 31 · 계좌 20)", symbol, c)
	if c.Err != nil || c.Converged != 1 {
		t.Fatalf("external sale cycle = %+v, want one convergence", c)
	}
	w.accountSold()
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "20") }
	c = w.next()
	w.logCycle("W2 주기(로컬 20 · 계좌 0)", symbol, c)
	if c.Err != nil || c.Converged != 0 || c.Released != 0 {
		t.Fatalf("W2 cycle = %+v, want no convergence and no release", c)
	}
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
}

// TestA201TheTrailingSymbolWhoseFillLandsAfterObserveStaysStuck: design §4 의 남는 것 — 판 중단 뒤 종목 B.
// 앞 종목 A 의 W3-a stale 이 수렴 판을 멈추고, B 의 체결은 그 주기 뒤(Observe 뒤 · 다음 로컬 읽기 전)에 커밋됨.
// B 는 로컬 읽기 뒤 체결 조건이 거짓이라 credit 이 없고 고착 그대로임(A 의 해제는 W3-a 시험이 고정).
func TestA201TheTrailingSymbolWhoseFillLandsAfterObserveStaysStuck(t *testing.T) {
	const a, b = "SYN201A", "SYN201B"
	w := newA201World(t)
	w.hold(a, "31")
	w.hold(b, "17")
	w.holds(a, "31", "100", 100)
	w.holds(b, "17", "100", 100)
	w.stopOpenAtBroker(a, "31")
	w.stopOpenAtBroker(b, "17")
	if c := w.next(); c.Err != nil || len(w.activeCauses(a))+len(w.activeCauses(b)) != 0 {
		t.Fatalf("start cycle = %+v", c)
	}
	w.accountSold()
	w.converge.at = a201AfterPositions
	w.converge.fire = func() { w.stopFills(a, "31") }

	c := w.next()
	w.logCycle("A stale 주기", b, c)
	if c.Err == nil || !strings.Contains(c.Err.Error(), a) || c.Converged != 0 {
		t.Fatalf("cycle = %+v, want A's stale to stop the pass before B converges", c)
	}
	w.clk.Advance(10 * time.Second)
	w.stopFills(b, "17") // Observe 뒤 · 다음 로컬 읽기 전
	w.settleClean(b, 5)
	w.assertStuck(b)
}

// --- 실제 차이 · 상쇄 · 반박 · 재기동 · credit 없음(design §3-2~6) ---------------------------------------

// a201R11Setup 은 codex r1 R1-1 순서의 주기 k 까지 만듦: 앞 주기 수렴 31→20(크레딧 · QM) 뒤, 스냅숏 15(사람 5 주
// 매도) · 로컬 읽기 20 · 그 뒤 엔진 매도 5 주 체결(로컬 15). 그 주기는 수렴이 stale 이고 해제 없음.
func a201R11Setup(t *testing.T, symbol string) *a201World {
	t.Helper()
	w := newA201World(t)
	w.start(symbol)
	w.holds(symbol, "20", "100", 100)
	w.orders.orders = nil
	w.stopOpenAtBroker(symbol, "31")
	c := w.next()
	w.logCycle("앞 주기(수렴 31→20 · QM · 크레딧)", symbol, c)
	if c.Converged != 1 {
		t.Fatalf("setup cycle = %+v", c)
	}
	w.confirm(reconcileAccount, symbol, "SELL", "o-sell5")
	w.clk.Advance(time.Second)
	w.holdings.items = nil
	w.orders.orders = nil
	w.holds(symbol, "15", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	w.openAtBroker("o-sell5", symbol, "5")
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.fill(reconcileAccount, symbol, "o-sell5", "SELL", "5") }
	c = w.next()
	w.logCycle("R1-1 주기 k(스냅숏 15 · 로컬 읽기 20 · 체결 뒤 로컬 15)", symbol, c)
	if !w.converge.fired {
		t.Fatal("fill hook never ran")
	}
	if c.Released != 0 {
		t.Fatalf("cycle k released %d", c.Released)
	}
	w.assertStuck(symbol)
	return w
}

// TestA201ARealDifferenceReleasesOnlyThroughTheAdjustment: R1-1 순서(design §3-2) — 다음 주기 실제 계좌 10 · 로컬 15
// 는 비교가 진입을 막아 해제 없음, 그 주기 수렴 15→10 의 조정 credit 이 다음 일치에서 ADJUSTMENT_APPLIED 로 풂.
func TestA201ARealDifferenceReleasesOnlyThroughTheAdjustment(t *testing.T) {
	const symbol = "SYN201"
	w := a201R11Setup(t, symbol)
	w.holdings.items = nil
	w.orders.orders = nil
	w.holds(symbol, "10", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	c := w.next()
	w.logCycle("주기 k+1(스냅숏 10 · 로컬 15)", symbol, c)
	if c.Released != 0 || c.Converged != 1 {
		t.Fatalf("cycle k+1 = %+v, want the real difference converged and nothing released", c)
	}
	w.assertStuck(symbol)
	c = w.next()
	w.logCycle("주기 k+2(스냅숏 10 · 로컬 10)", symbol, c)
	if c.Err != nil || c.Released != 1 {
		t.Fatalf("cycle k+2 = %+v, want one release", c)
	}
	w.releasedAs(symbol, journal.ReconcileReleaseAdjustmentApplied)
}

// TestA201AnOffsettingExternalTradeReleasesAsFillApplied: R1-1 변형(design §3-3 · V8-1 공시 동작) — 실제 외부 분쟁이
// 섞인 주기 k 뒤에 사람이 5 주를 되사 새 비교가 일치하면 주기 k 의 체결 credit 으로 풀림.
func TestA201AnOffsettingExternalTradeReleasesAsFillApplied(t *testing.T) {
	const symbol = "SYN201"
	w := a201R11Setup(t, symbol)
	w.holdings.items = nil
	w.orders.orders = nil
	w.holds(symbol, "15", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	c := w.next()
	w.logCycle("주기 k+1(사람 5 주 재매수: 스냅숏 15 · 로컬 15)", symbol, c)
	if c.Err != nil || c.Released != 1 {
		t.Fatalf("cycle k+1 = %+v, want one release", c)
	}
	w.releasedAs(symbol, a201FillApplied)
}

// TestA201ALaterDisputeRefutesTheFillCredit: design §3-4 — 체결 credit 뒤의 비교가 그 종목을 분쟁(외부 포지션)으로
// 보면 credit 이 소멸하고, 그 뒤 일치 주기에는 풀리지 않음. 일치 주기는 체결 · 조정 없이 계좌 = 로컬(접힌 5 주)로 만들어
// 조정 credit 이 덮어쓰는 경로 없이 반박 자체를 잼(구현 리뷰 r1 R2).
func TestA201ALaterDisputeRefutesTheFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("W2 주기", symbol, c)
	w.assertStuck(symbol)

	w.holds(symbol, "5", "100", 100) // 사람 재매수 → 외부 포지션(분쟁)
	c = w.next()
	w.logCycle("외부 포지션 주기", symbol, c)
	if c.Err != nil || c.Released != 0 || c.Awaiting != 1 {
		t.Fatalf("dispute cycle = %+v, want the block awaiting and nothing released", c)
	}
	// 계좌 5 주 그대로 — 외부 포지션 주기에 접힌 5 주와 같아 체결 · 조정 없이 일치함.
	for i := 1; i <= 2; i++ {
		c = w.next()
		w.logCycle("계좌 = 로컬 주기", symbol, c)
		if c.Err != nil || c.Converged != 0 || c.Released != 0 {
			t.Fatalf("agreeing cycle %d = %+v, want no convergence and no release — the fill credit was refuted", i, c)
		}
		w.assertStuck(symbol)
	}
}

// TestA201ARestartDropsTheFillCredit: design §3-6 — 체결 credit 은 조정 credit 처럼 프로세스 기억이라 Restore 뒤에는
// 없음. 그 뒤 일치 주기에도 차단이 남음.
func TestA201ARestartDropsTheFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("W2 주기", symbol, c)
	w.assertStuck(symbol)
	if err := w.tracker.Restore(w.ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	w.settleClean(symbol, 3)
	w.assertStuck(symbol)
	w.rebuyRefused(symbol)
	w.assertStuck(symbol)
}

// a201EnterRefusingStore 는 차단 행 삽입을 실패시켜 그 주기 Observe 를 오류로 끝냄.
type a201EnterRefusingStore struct {
	reconcile.ReconcileStore
}

func (s *a201EnterRefusingStore) EnterReconcile(context.Context,
	journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	return journal.ReconcileState{}, false, errors.New("a201 injected enter failure")
}

// TestA201AnObserveErrorCycleGrantsNoFillCredit: design §3-5 · V3 — Observe 가 오류인 주기에는 체결 credit 을 주지 않음.
// 그 뒤 일치 주기들에 FILL_APPLIED 해제가 없음.
func TestA201AnObserveErrorCycleGrantsNoFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.tracker.Journal = &a201EnterRefusingStore{ReconcileStore: w.store}
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("W2 주기(Observe 쓰기 실패)", symbol, c)
	if c.Err == nil || !strings.Contains(c.Err.Error(), "a201 injected enter failure") {
		t.Fatalf("W2 cycle err = %v, want the Observe write failure", c.Err)
	}
	w.tracker.Journal = w.store
	w.settleClean(symbol, 3)
	for _, row := range w.rows(symbol) {
		if strings.HasSuffix(row, "released:"+a201FillApplied) {
			t.Fatalf("rows for %s = %q — an Observe error cycle must not grant a fill credit", symbol, w.rows(symbol))
		}
	}
}

// TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore: RunOnce B2 — 로컬 읽기 앞에 들어간 표식 읽기는 실패해도
// 주기를 따로 실패시키지 않고(「모름」 표식), 주기는 지금처럼 로컬 읽기 오류로 끝남.
func TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	if err := w.journal.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	c := w.next()
	if c.Err == nil || !strings.Contains(c.Err.Error(), "reading the local state for reconciliation") {
		t.Fatalf("cycle err = %v, want the local read's failure", c.Err)
	}
	if c.Released != 0 || c.Blocked != 0 {
		t.Fatalf("cycle = %+v, want it to stop before any observation", c)
	}
}

// TestA201AConvergedSymbolGetsNoFillCredit: design §3-8 · (q6) — 수렴이 조정을 커밋한 뒤 Refresh 에 체결(W4)이면 그
// 종목은 「미반영」이 아니라 체결 credit 이 없고, a200 의 ATTRIBUTION_FAILED 처분이 그대로임.
func TestA201AConvergedSymbolGetsNoFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.store.armed = true
	w.store.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("W4 주기", symbol, c)
	if c.Converged != 1 {
		t.Fatalf("W4 cycle = %+v, want the convergence committed before the fill", c)
	}
	w.settleClean(symbol, 2)
	got := w.activeCauses(symbol)
	if len(got) != 1 || got[0] != journal.ReconcileCauseAttributionFailed {
		t.Fatalf("active rows for %s = %q, want ATTRIBUTION_FAILED kept", symbol, got)
	}
	for _, row := range w.rows(symbol) {
		if strings.HasSuffix(row, "released:"+a201FillApplied) {
			t.Fatalf("rows for %s = %q — a converged symbol must not get a fill credit", symbol, w.rows(symbol))
		}
	}
}

// --- 그 주기 Observe 무변(design §3-7) ---------------------------------------------------------------

// TestA201TheFillCycleStillBindsAnotherOrderOfAnA192Symbol: R2-1 첫 반례 입력 — 체결 경합 주기에도 Observe 가 그대로
// 돌아 a192 결속 종목 C 의 새 missing-order(o-c2)를 결속에 더함. 그래서 o-c1 이 취소돼도 살아 있는 o-c2 가 C 를 붙잡고,
// o-c2 가 취소된 뒤에야 ORDER_TERMINAL_CONFIRMED 로 풀림.
func TestA201TheFillCycleStillBindsAnotherOrderOfAnA192Symbol(t *testing.T) {
	const a, cs = "SYN201", "SYN201C"
	w := newA201World(t)
	w.hold(a, "31")
	w.holds(a, "31", "100", 100)
	w.stopOpenAtBroker(a, "31")
	w.confirm(reconcileAccount, cs, "SELL", "o-c1") // 계좌 열린 목록에 없음 → missing-order 블록 · 결속
	c := w.next()
	w.logCycle("결속 주기", cs, c)
	if got := w.activeCauses(cs); c.Err != nil || len(got) != 1 {
		t.Fatalf("binding cycle = %+v rows=%q, want C's missing-order block", c, w.rows(cs))
	}

	w.accountSold()
	w.confirm(reconcileAccount, cs, "SELL", "o-c2")
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(a, "31") }
	c = w.next()
	w.logCycle("W2 주기(C 의 o-c1 · o-c2 missing)", cs, c)
	w.assertStuck(a)

	w.cancel(cs, "o-c1")
	w.openAtBroker("o-c2", cs, "1") // o-c2 는 계좌에 살아 있음 — missing 이 아님
	for i := 1; i <= 3; i++ {
		c = w.next()
		w.logCycle("o-c2 열림 주기", cs, c)
		if got := w.activeCauses(cs); len(got) != 1 {
			t.Fatalf("cycle %d: rows for %s = %q — the live bound o-c2 must keep the block", i, cs, w.rows(cs))
		}
	}
	w.orders.orders = nil
	w.cancel(cs, "o-c2")
	for i := 1; i <= 3 && len(w.activeCauses(cs)) != 0; i++ {
		c = w.next()
		w.logCycle("o-c2 취소 뒤 주기", cs, c)
	}
	w.releasedAs(cs, journal.ReconcileReleaseOrderTerminalConfirmed)
}

// TestA201TheFillCycleStillCountsAndBreaksTheStreak: R2-1 둘째 반례 입력 — 체결 경합 주기의 Observe 가 분쟁을 세고
// (연속 1), 다음 깨끗한 비교가 연속을 끊음(0). 주기를 건너뛰면 이 단절이 사라져 비연속 승격이 생겼음(v2 기각 사유).
func TestA201TheFillCycleStillCountsAndBreaksTheStreak(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("W2 주기", symbol, c)
	if got := w.tracker.Failures(); got != 1 || w.tracker.Permanent() {
		t.Fatalf("after the W2 cycle failures=%d permanent=%v, want 1/false", got, w.tracker.Permanent())
	}
	c = w.next()
	w.logCycle("다음 주기", symbol, c)
	if got := w.tracker.Failures(); got != 0 {
		t.Fatalf("after the clean cycle failures=%d, want the streak broken", got)
	}
}

// TestA201AdoptionPriceFailuresStillCountAcrossTheFillCycle: R2-2 입력 — 체결 경합 주기도 편입 가격 조회 실패를 실패
// 주기로 세어 연속 실패 수가 끊기지 않음(운영 모드 강화가 가려지지 않음).
func TestA201AdoptionPriceFailuresStillCountAcrossTheFillCycle(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.start(symbol)
	w.holds("SYN201E", "3", "100", 100) // 엔진이 모르는 보유 → 편입 후보
	w.prices.err = errors.New("a201 injected price read failure")
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c1 := w.next()
	w.logCycle("W2 주기(가격 실패)", symbol, c1)
	c2 := w.next()
	w.logCycle("다음 주기(가격 실패)", symbol, c2)
	if c1.Err == nil || c2.Err == nil {
		t.Fatalf("cycles err = %v / %v, want both to carry the price failure", c1.Err, c2.Err)
	}
	if h := w.driver.Health(); h.Consecutive != 2 {
		t.Fatalf("consecutive failed cycles = %d, want 2", h.Consecutive)
	}
}

// TestA201AnotherSymbolsDisputeKeepsTheFillReleaseWaiting: design §3-9 — 다른 종목 Z 의 수량 불일치가 비교를 막는 동안
// A 는 풀리지 않고(credit 은 남음), 그 분쟁이 사라진 첫 깨끗한 비교에서 A 는 FILL_APPLIED · Z 는 ADJUSTMENT_APPLIED.
func TestA201AnotherSymbolsDisputeKeepsTheFillReleaseWaiting(t *testing.T) {
	const a, z = "SYN201", "SYN201Z"
	w := newA201World(t)
	w.hold(a, "31")
	w.hold(z, "10")
	w.holds(a, "31", "100", 100)
	w.holds(z, "10", "100", 100)
	w.stopOpenAtBroker(a, "31")
	w.stopOpenAtBroker(z, "10")
	if c := w.next(); c.Err != nil {
		t.Fatalf("start cycle = %+v", c)
	}
	w.holdings.items = nil
	w.orders.orders = nil
	w.holds(z, "10", "100", 100)
	w.stopOpenAtBroker(z, "10")
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(a, "31") }
	c := w.next()
	w.logCycle("W2 주기", a, c)
	w.assertStuck(a)

	w.holdings.items = nil
	w.holds(z, "9", "100", 100) // 사람의 앱 매도 1 주 → Z 수량 불일치(비교가 진입을 막음)
	c = w.next()
	w.logCycle("Z 분쟁 주기", a, c)
	if c.Released != 0 {
		t.Fatalf("Z dispute cycle = %+v, want nothing released while the comparison blocks", c)
	}
	w.assertStuck(a)
	c = w.next()
	w.logCycle("첫 깨끗한 주기", a, c)
	if c.Err != nil || c.Released != 2 {
		t.Fatalf("first clean cycle = %+v, want A and Z released", c)
	}
	w.releasedAs(a, a201FillApplied)
	w.releasedAs(z, journal.ReconcileReleaseAdjustmentApplied)
}

// --- a192 결속과의 순서(design §3-13 · V5) ---------------------------------------------------------

// TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit: 수량 흔적 없는 missing-order 결속 행 — 로컬 읽기 뒤 그 주문의
// 취소가 커밋돼도 수량 불일치가 없어 체결 credit 이 서지 않고, a192 경로(ORDER_TERMINAL_CONFIRMED)로 풀림.
func TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.confirm(reconcileAccount, symbol, "SELL", "o-a192") // 보유 없음 · 계좌 열린 목록에 없음 → missing-order
	c := w.next()
	w.logCycle("missing-order 주기(결속)", symbol, c)
	if c.Err != nil || c.Blocked != 1 {
		t.Fatalf("first cycle = %+v, want the missing-order block", c)
	}
	w.store.armed = true
	w.store.fire = func() { w.cancel(symbol, "o-a192") } // 로컬 읽기 뒤(Refresh) 취소 커밋
	for i := 1; i <= 4 && len(w.activeCauses(symbol)) != 0; i++ {
		c = w.next()
		w.logCycle("취소 뒤 주기", symbol, c)
		if c.Err != nil {
			t.Fatalf("cycle %d: %v", i, c.Err)
		}
	}
	w.releasedAs(symbol, journal.ReconcileReleaseOrderTerminalConfirmed)
}

// TestA201ABoundRowWithAQuantityMismatchReleasesAsFillApplied: missing-order 로 삽입 · 결속한 행 위에 나중 수량 불일치
// (W2)가 오면 결속은 실격(a192 는 「조정 규칙만」에 맡김)이고, 그 주기의 체결 credit 이 다음 깨끗한 주기에 풂.
func TestA201ABoundRowWithAQuantityMismatchReleasesAsFillApplied(t *testing.T) {
	const symbol = "SYN201"
	w := newA201World(t)
	w.hold(symbol, "31")
	w.holds(symbol, "31", "100", 100) // 엔진 매도가 계좌 열린 목록에 없음 → missing-order(결속)
	c := w.next()
	w.logCycle("missing-order 주기(결속)", symbol, c)
	if c.Err != nil || c.Blocked != 1 {
		t.Fatalf("binding cycle = %+v, want the missing-order block", c)
	}
	w.accountSold()
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c = w.next()
	w.logCycle("W2 주기(같은 행 위 수량 불일치)", symbol, c)
	w.assertStuck(symbol)
	w.releasedByFill(symbol)
}
