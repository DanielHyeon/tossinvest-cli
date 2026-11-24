package reconcile_test

// a201_fill_credit_test.go 는 a201 체결 credit 의 함수 단위 계약을 고정함(design V2 · V4b · V5b · §3-5 · §3-12 · §3-14).
//
// 실제 RunOnce 경로는 internal/app/engine/a201_fill_credit_cycles_test.go 가 봄. 여기는 부여 대상 계산(V2 의 닫힌 조건 ·
// 판독 실패), 블록 없는 종목, 엄격히 나중 비교, V4b 체결 credit 판, 해제 원인 등식(V5b)을 실제 임시 장부로 봄.

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a201Fill 은 symbol 의 엔진 매수 주문 하나를 확정하고 체결함 — 그 종목의 체결 워터마크가 움직임.
func a201Fill(t *testing.T, w *a192World, symbol, orderID string) {
	t.Helper()
	confirmScopedOrder(t, w.journal, "intent-"+orderID, "attempt-"+orderID, a192Account, "kr", a192Day, symbol, "BUY", orderID)
	if _, err := w.journal.RecordFill(w.ctx, journal.FillObservation{
		OrderID: orderID, Symbol: symbol, Market: "kr", TradingDay: a192Day, Side: "BUY",
		AccountRef: a192Account, State: "FILLED", Terminal: true,
		Quantity: "1", FilledQuantity: "1", AveragePrice: "100", ObservedAt: w.now(),
	}); err != nil {
		t.Fatalf("RecordFill(%s): %v", orderID, err)
	}
}

func a201Quantity(symbol string) reconcile.QuantityMismatch {
	return reconcile.QuantityMismatch{Symbol: symbol, Local: "1", Broker: "0"}
}

// a201ReleaseCauses 는 그 종목 행들의 해제 원인을 진입 순서대로 돌려줌(미해제는 "active").
func a201ReleaseCauses(t *testing.T, w *a192World, symbol string) []string {
	t.Helper()
	history, err := w.journal.ReconcileStateHistory(w.ctx, a192Account)
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	var out []string
	for _, s := range history {
		if s.Symbol != symbol {
			continue
		}
		if s.Active() {
			out = append(out, "active")
			continue
		}
		out = append(out, s.ReleaseCause)
	}
	return out
}

// a201Blocked 는 그 종목에 자동 해제 대상 QM 블록을 세움(비교 as-of 를 돌려줌).
func a201Blocked(t *testing.T, w *a192World, symbol string) string {
	t.Helper()
	diff := reconcile.Diff{AsOf: w.now(), AccountRef: a192Account, Quantities: []reconcile.QuantityMismatch{a201Quantity(symbol)}}
	w.observe(diff)
	return diff.AsOf
}

func TestA201TheFillReleaseCauseIsFillApplied(t *testing.T) {
	if journal.ReconcileReleaseFillApplied != "FILL_APPLIED" || !journal.ValidReconcileReleaseCause(journal.ReconcileReleaseFillApplied) {
		t.Fatalf("ReconcileReleaseFillApplied = %q, want the valid cause FILL_APPLIED", journal.ReconcileReleaseFillApplied)
	}
}

// TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols: V2 (a)(b)(c) — 그 비교의 수량 불일치 종목 중 로컬 읽기
// 전 표식보다 워터마크가 크고 수렴이 반영하지 않은 종목만. 빈 종목 · missing-order 만 · 외부 포지션만인 종목은 아님.
func TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols(t *testing.T) {
	w := newA192World(t, nil)
	a201Fill(t, w, "SYNSTILL", "o-still") // 표식 전 체결 — 워터마크가 표식 이하
	mark := reconcile.MarkBeforeLocalRead(w.ctx, w.journal)
	w.tick()
	for _, s := range []string{a192Symbol, "SYNCONV", "SYNMISS", "SYNEXT"} {
		a201Fill(t, w, s, "o-after-"+s)
	}
	diff := reconcile.Diff{
		AsOf: w.now(), AccountRef: a192Account,
		Quantities: []reconcile.QuantityMismatch{
			a201Quantity(" " + "syn001" + " "), a201Quantity("SYNCONV"), a201Quantity("SYNSTILL"), a201Quantity(""),
		},
		MissingOrders: []reconcile.LocalOrder{{OrderID: "o-miss", Symbol: "SYNMISS", AccountRef: a192Account, Market: "kr"}},
		ExternalPos:   []reconcile.ExternalPosition{{Holding: reconcile.Holding{Symbol: "SYNEXT", Quantity: "1", Market: "kr"}}},
	}
	converged := reconcile.ConvergeReport{Converged: []reconcile.ConvergedPosition{{Market: "kr", Symbol: "synconv"}}}
	got := reconcile.FillOutranSymbols(w.ctx, w.journal, mark, diff, converged)
	if want := []string{a192Symbol}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FillOutranSymbols = %q, want %q", got, want)
	}
}

// TestA201AnUnreadableJournalGrantsNoFillCredit: V2 판독 실패 — 표식을 못 읽었거나(닫힌 장부 · 장부 없음) 종목 워터마크를
// 못 읽으면 아무 종목도 주지 않음.
func TestA201AnUnreadableJournalGrantsNoFillCredit(t *testing.T) {
	w := newA192World(t, nil)
	mark := reconcile.MarkBeforeLocalRead(w.ctx, w.journal)
	w.tick()
	a201Fill(t, w, a192Symbol, "o-after")
	diff := reconcile.Diff{AsOf: w.now(), AccountRef: a192Account, Quantities: []reconcile.QuantityMismatch{a201Quantity(a192Symbol)}}
	if got := reconcile.FillOutranSymbols(w.ctx, w.journal, mark, diff, reconcile.ConvergeReport{}); len(got) != 1 {
		t.Fatalf("control FillOutranSymbols = %q, want the moved symbol", got)
	}

	if got := reconcile.FillOutranSymbols(w.ctx, w.journal, reconcile.MarkBeforeLocalRead(w.ctx, nil), diff, reconcile.ConvergeReport{}); len(got) != 0 {
		t.Fatalf("with no journal for the mark FillOutranSymbols = %q, want none", got)
	}
	closed := openJournal(t)
	if err := closed.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got := reconcile.FillOutranSymbols(w.ctx, w.journal, reconcile.MarkBeforeLocalRead(w.ctx, closed), diff, reconcile.ConvergeReport{}); len(got) != 0 {
		t.Fatalf("with an unreadable mark FillOutranSymbols = %q, want none", got)
	}
	if got := reconcile.FillOutranSymbols(w.ctx, closed, mark, diff, reconcile.ConvergeReport{}); len(got) != 0 {
		t.Fatalf("with an unreadable symbol watermark FillOutranSymbols = %q, want none", got)
	}
}

// TestA201AFillCreditNeedsAStandingBlock: V2 (d) — 부여 때 그 종목에 자동 해제 대상 블록이 없으면 credit 을 세우지 않음.
// 그 뒤 Refresh 로 들어온 같은 종목 행은 그 낡은 부여로 풀리지 않음.
func TestA201AFillCreditNeedsAStandingBlock(t *testing.T) {
	w := newA192World(t, nil)
	w.tracker.FillsOutranAdjustment(w.now(), a192Symbol)
	if _, _, err := w.journal.EnterReconcile(w.ctx, journal.EnterReconcileRequest{
		AccountRef: a192Account, Symbol: a192Symbol, Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "another producer",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	if err := w.tracker.Refresh(w.ctx); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	w.tick()
	out := w.observe(w.clean())
	if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("cleared=%+v awaiting=%+v — a grant with no standing block must not release a later row", out.Cleared, out.AwaitingAdjustment)
	}
}

// TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison: V3 · V4 — 같은 as-of 의 비교는 그 credit 을 못 쓰고,
// 엄격히 나중의 깨끗한 비교가 FILL_APPLIED 로 풂.
func TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison(t *testing.T) {
	w := newA192World(t, nil)
	at := a201Blocked(t, w, a192Symbol)
	w.tracker.FillsOutranAdjustment(at, a192Symbol)
	same := w.clean()
	same.AsOf = at
	if out := w.observe(same); len(out.Cleared) != 0 {
		t.Fatalf("the same comparison released %+v", out.Cleared)
	}
	w.tick()
	out := w.observe(w.clean())
	if len(out.Cleared) != 1 {
		t.Fatalf("cleared=%+v, want the block released", out.Cleared)
	}
	if got := a201ReleaseCauses(t, w, a192Symbol); !reflect.DeepEqual(got, []string{journal.ReconcileReleaseFillApplied}) {
		t.Fatalf("release causes = %q, want FILL_APPLIED", got)
	}
}

// TestA201AnAuthorityErrorOnACleanComparisonRefutesAFillCredit: V4b 체결 credit 판(조정 credit 판은
// a201_authority_refute_test.go) — 같은 맵 · 같은 반박 호출이라 체결 credit 도 같은 자리에서 소멸함.
func TestA201AnAuthorityErrorOnACleanComparisonRefutesAFillCredit(t *testing.T) {
	clk, j, store, tracker, threshold := a201PendingPermanentWorld(t)
	tracker.FillsOutranAdjustment(threshold, "AAPL")

	a110Advance(clk)
	store.failReads = true
	if out, err := tracker.Observe(context.Background(), a201ExternalOnly(clk)); err == nil {
		t.Fatalf("external-only Observe succeeded despite the authority outage: %+v", out)
	}
	store.failReads = false

	a110Advance(clk)
	out := a110Observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("cleared=%+v awaiting=%+v — the fill credit the external position disputed must not release", out.Cleared, out.AwaitingAdjustment)
	}
	if cause, active := a201ReleaseCause(t, j, "AAPL"); !active {
		t.Fatalf("AAPL row released with %q, want it active", cause)
	}
}

// TestA201WithoutAnAuthorityErrorADisputedCreditStillTakesTheAwaitingBranch: V4b 를 오류 갈래 안에만 둔 이유를 고정함 —
// 오류 없는 깨끗한 비교에서 credit 있는 종목이 외부 포지션으로 분쟁이면 지금처럼 B10(Awaiting)으로 가고 a192 후보 판정
// (orderTerminalCandidate)을 부르지 않음. 반박을 앞으로 옮기면 그 블록이 B8 → a192 후보로 갈래가 바뀌어 보류 사유가 남음.
func TestA201WithoutAnAuthorityErrorADisputedCreditStillTakesTheAwaitingBranch(t *testing.T) {
	w := newA192World(t, nil)
	w.order("o-a201", "SELL")
	w.observe(w.missing(a192Missing("o-a201"))) // 이 프로세스가 삽입 · 결속한 missing-order 블록
	at := w.now()
	w.tracker.AdjustmentApplied(at, a192Symbol)
	w.tick()
	diff := w.clean()
	diff.ExternalPos = []reconcile.ExternalPosition{{Holding: reconcile.Holding{Symbol: a192Symbol, Quantity: "1", Market: "kr"}}}
	out := w.observe(diff)
	if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("cleared=%+v awaiting=%+v, want the block awaiting", out.Cleared, out.AwaitingAdjustment)
	}
	if reasons := w.tracker.OrderTerminalHoldReasons(); reasons[a192Symbol] != "" {
		t.Fatalf("hold reasons = %v — the credited, disputed block must not reach the order-terminal candidate", reasons)
	}
}

// TestA201AGrantPrunesALeftoverFillMarkAfterARestart: V5b · design §4 — 재기동(Restore)은 credit 맵만 비우고 곁 맵은 남김.
// 다음 주기의 부여 호출(대상 종목 없음)이 값이 달라진 곁 맵 항목을 정리하므로, 그 뒤 우연히 같은 글자의 as-of 로 선
// 조정 credit 의 해제도 ADJUSTMENT_APPLIED 로 남음(정리가 없으면 등식이 참으로 남아 FILL_APPLIED 로 잘못 기록됨).
func TestA201AGrantPrunesALeftoverFillMarkAfterARestart(t *testing.T) {
	w := newA192World(t, nil)
	at := a201Blocked(t, w, a192Symbol)
	w.tracker.FillsOutranAdjustment(at, a192Symbol)
	if err := w.tracker.Restore(w.ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	w.tick()
	w.tracker.FillsOutranAdjustment(w.now()) // 재기동 뒤 첫 주기의 부여 호출 — 대상 종목 없음
	w.tracker.AdjustmentApplied(at, a192Symbol)
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 1 {
		t.Fatalf("cleared=%+v, want the adjustment credit's release", out.Cleared)
	}
	if got := a201ReleaseCauses(t, w, a192Symbol); !reflect.DeepEqual(got, []string{journal.ReconcileReleaseAdjustmentApplied}) {
		t.Fatalf("release causes = %q, want ADJUSTMENT_APPLIED", got)
	}
}

// TestA201ALaterAdjustmentNamesTheRelease: V5b — 체결 credit 뒤 그 종목에 조정이 커밋되면(더 나중 as-of 로 덮어씀) 해제
// 원인은 ADJUSTMENT_APPLIED.
func TestA201ALaterAdjustmentNamesTheRelease(t *testing.T) {
	w := newA192World(t, nil)
	at := a201Blocked(t, w, a192Symbol)
	w.tracker.FillsOutranAdjustment(at, a192Symbol)
	w.tick()
	w.tracker.AdjustmentApplied(w.now(), a192Symbol) // 다음 주기 수렴(Observe 앞)
	a201Blocked(t, w, a192Symbol)
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 1 {
		t.Fatalf("cleared=%+v, want the block released", out.Cleared)
	}
	if got := a201ReleaseCauses(t, w, a192Symbol); !reflect.DeepEqual(got, []string{journal.ReconcileReleaseAdjustmentApplied}) {
		t.Fatalf("release causes = %q, want ADJUSTMENT_APPLIED", got)
	}
}

// TestA201ALeftoverFillMarkDoesNotRenameALaterAdjustmentRelease: V5b — FILL_APPLIED 해제 뒤 남은 곁 맵 항목이 그 종목의
// 다음 조정 credit 해제의 원인을 바꾸지 않음(값이 달라 등식이 거짓).
func TestA201ALeftoverFillMarkDoesNotRenameALaterAdjustmentRelease(t *testing.T) {
	w := newA192World(t, nil)
	at := a201Blocked(t, w, a192Symbol)
	w.tracker.FillsOutranAdjustment(at, a192Symbol)
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 1 {
		t.Fatalf("cleared=%+v, want the fill credit's release", out.Cleared)
	}
	w.clk.Advance(2 * time.Minute)
	w.tracker.AdjustmentApplied(w.now(), a192Symbol)
	a201Blocked(t, w, a192Symbol)
	w.tick()
	if out := w.observe(w.clean()); len(out.Cleared) != 1 {
		t.Fatalf("cleared=%+v, want the adjustment credit's release", out.Cleared)
	}
	want := []string{journal.ReconcileReleaseFillApplied, journal.ReconcileReleaseAdjustmentApplied}
	if got := a201ReleaseCauses(t, w, a192Symbol); !reflect.DeepEqual(got, want) {
		t.Fatalf("release causes = %q, want %q", got, want)
	}
}
