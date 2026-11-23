package reconcile_test

// a200_recognition_cycles_test.go 는 a200 의 새 계약을 실제 Converger · Tracker 로 잼(design §3-4): **관리** 포지션(편입 + 손절 발의)의
// 엔진 매도 체결이 수렴 조정과 경합하면 W4 · W5 는 귀속 실패 차단 없이 지나가고 다음 일치 비교에서 그 종목 진입이 열리며, W6 은
// 오늘과 같다. 1단계 하네스(a200World — a200_convergence_close_exit_fill_cycles_test.go)를 쓰고, 관리 포지션을 만드는 것만 새로
// 둠(그 하네스의 hold 는 무관리 포지션이라 「계속 거절: 무관리」 사례로 남음). 픽스처는 합성 값만 씀.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// holdManaged 는 엔진 매수 체결 31 을 편입하고 exit 상태를 연 뒤, 손절 발의를 무장하고 주문 수량 31 의 매도 주문(접수 확정)을
// 그 발의에 붙임 — ExitObserver.submit 의 arm → attach → submit 순서.
func (w *a200World) holdManaged() string {
	w.t.Helper()
	ctx := w.ctx
	confirmScopedOrder(w.t, w.journal, "intent-entry", "attempt-entry", a200Account, "kr", a200Day, a200Symbol, "BUY", "o-entry")
	w.clk.Advance(time.Second)
	w.fill("o-entry", "BUY", "31", "31", true)
	p, err := w.journal.CurrentPosition(ctx, a200Account, "kr", a200Symbol)
	if err != nil {
		w.t.Fatalf("CurrentPosition: %v", err)
	}
	if _, err := w.journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: p.ID, Symbol: a200Symbol, Market: "kr", Quantity: "31", CostBasis: "100",
		ObservedPrice: "100", SyntheticStop: "95", ObservedAt: w.now(),
	}); err != nil {
		w.t.Fatalf("AdoptPosition: %v", err)
	}
	if _, err := w.journal.OpenAdoptedExitState(ctx, p.ID); err != nil {
		w.t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	w.clk.Advance(time.Second)
	state, err := w.journal.ExitState(ctx, p.ID)
	if err != nil {
		w.t.Fatal(err)
	}
	if err := w.journal.RecordExitJudgement(ctx, journal.ExitJudgement{
		PositionID: p.ID, ObservedPrice: "94", HighWater: state.HighWater, Baseline: state.Baseline,
		RatchetLevel: state.RatchetLevel, ActiveRung: exitpolicy.NoRung,
		Proposal: &journal.ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: state.RatchetLevel},
	}); err != nil {
		w.t.Fatalf("손절 발의 무장: %v", err)
	}
	attempt, err := w.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-stop", AccountRef: a200Account, Market: "kr", TradingDay: a200Day, Symbol: a200Symbol,
			Side: "SELL", OrderType: "LIMIT", Quantity: "31", Price: "94", Currency: "KRW", Source: "engine",
			Fingerprint: "fp-intent-stop",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-stop",
	})
	if err != nil {
		w.t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		w.t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "o-stop"); err != nil {
		w.t.Fatal(err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		w.t.Fatal(err)
	}
	if err := w.journal.AttachExitIntent(ctx, p.ID, "intent-stop"); err != nil {
		w.t.Fatalf("AttachExitIntent: %v", err)
	}
	w.clk.Advance(time.Second)
	return p.ID
}

// assertNoAttributionFailure 는 그 종목에 귀속 실패 행이 한 번도 생기지 않았음을 봄.
func (w *a200World) assertNoAttributionFailure(label string) {
	w.t.Helper()
	for _, r := range w.rows() {
		if strings.HasPrefix(r, journal.ReconcileCauseAttributionFailed) {
			w.t.Fatalf("%s: rows = %q, 인정된 체결은 귀속 실패 차단을 만들지 않아야 함", label, w.rows())
		}
	}
}

// recognitionMarkers 는 그 포지션의 인정 표지(ADJUSTMENT_CLOSE_FILL_RECOGNISED) 행 수를 셈.
func (w *a200World) recognitionMarkers(positionID string) int {
	w.t.Helper()
	events, err := w.journal.ExitEvents(w.ctx, positionID)
	if err != nil {
		w.t.Fatalf("ExitEvents: %v", err)
	}
	n := 0
	for _, e := range events {
		if e.Action == journal.ExitEventAdjustmentCloseFillRecognised {
			n++
		}
	}
	return n
}

// assertOpenAfterOneCleanCycle 은 다음 일치 비교 한 번에서 수량 불일치가 풀리고 그 종목 진입이 열리는 것을 봄.
func (w *a200World) assertOpenAfterOneCleanCycle(label string) {
	w.t.Helper()
	out, err := w.cleanCycle()
	w.logCycle(label+" 다음 일치", out, err)
	if err != nil {
		w.t.Fatalf("%s: 다음 일치 Observe: %v", label, err)
	}
	if got := w.activeCauses(); len(got) != 0 {
		w.t.Fatalf("%s: 다음 일치 뒤 활성 = %q, 수량 불일치는 조정 반영으로 풀려야 함", label, got)
	}
	if rejected := w.tracker.EntryAllowed("kr", a200Symbol); rejected != nil {
		w.t.Fatalf("%s: 다음 일치 뒤에도 그 종목 진입이 막힘: %v", label, rejected)
	}
}

// TestA200W4AManagedStopFillLeavesNoBlock: 수렴 조정 커밋 뒤 · Refresh 전 체결.
func TestA200W4AManagedStopFillLeavesNoBlock(t *testing.T) {
	w := newA200World(t)
	w.holdManaged()
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	w.fill("o-stop", "SELL", "31", "31", true) // W4
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	out, oerr := w.observe(diff)
	w.logCycle("W4 주기", out, oerr)
	if oerr != nil {
		t.Fatalf("W4 Observe 오류: %v", oerr)
	}
	w.assertNoAttributionFailure("W4")
	w.assertOpenAfterOneCleanCycle("W4")
}

// TestA200W5AManagedStopFillLeavesNoBlock: Refresh 뒤 · persist 전 체결 — 오늘의 Observe 오류(F13)가 없어야 함.
func TestA200W5AManagedStopFillLeavesNoBlock(t *testing.T) {
	w := newA200World(t)
	w.holdManaged()
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
	if oerr != nil {
		t.Fatalf("W5 Observe 오류: %v — 인정되면 persist 가 부딪칠 귀속 실패 행이 없어야 함", oerr)
	}
	w.assertNoAttributionFailure("W5")
	w.assertOpenAfterOneCleanCycle("W5")
}

// TestA200W6AManagedStopFillIsUnchanged: 수량 불일치 행이 선 뒤 체결 — 오늘과 같은 결과(귀속 실패 없음 · 다음 일치에서 해제).
// 판정은 활성 차단 확인(enterReconcileScopeInTx 의 F10)보다 앞이라, 차단이 선 채로도 네 조건이 맞으면 표지 1 이 남고 그 차단은
// 건드리지 않음(gstack r1 state-machines) — 표지 수와 「활성은 수량 불일치 하나뿐」 을 함께 못 박음.
func TestA200W6AManagedStopFillIsUnchanged(t *testing.T) {
	w := newA200World(t)
	id := w.holdManaged()
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if _, err := w.observe(diff); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	w.fill("o-stop", "SELL", "31", "31", true) // W6
	w.assertNoAttributionFailure("W6")
	if n := w.recognitionMarkers(id); n != 1 {
		t.Fatalf("W6 표지 %d 개, 활성 차단이 있어도 네 조건이 맞으면 표지 1", n)
	}
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("W6 체결 직후 활성 = %q, 이미 선 수량 불일치 하나뿐이어야 함(인정은 차단을 건드리지 않음)", got)
	}
	w.assertOpenAfterOneCleanCycle("W6")
}

// TestA200AnUnrecognisedW6FillWritesNoNewRow 는 gstack r1 state-machines 의 미시험 칸(Manager 2026-11-24): 수량 불일치 차단이
// 이미 선 창(W6)에 (라)가 깨진 체결(같은 intent 에 PLACE attempt 둘)이 오면 인정하지 않고, 차단 경로의 F10 이 새 행을 쓰지 않음 —
// base 와 같은 결과(표지 0 · 귀속 실패 0 · 활성은 수량 불일치 하나뿐). (나) 실패도 같은 F10 경로.
func TestA200AnUnrecognisedW6FillWritesNoNewRow(t *testing.T) {
	w := newA200World(t)
	id := w.holdManaged()
	if _, err := w.journal.Prepare(w.ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-stop", AccountRef: a200Account, Market: "kr", TradingDay: a200Day, Symbol: a200Symbol,
			Side: "SELL", OrderType: "LIMIT", Quantity: "31", Price: "94", Currency: "KRW", Source: "engine",
			Fingerprint: "fp-intent-stop",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-stop-2",
	}); err != nil {
		t.Fatalf("둘째 PLACE attempt: %v", err)
	}
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if _, err := w.observe(diff); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	w.fill("o-stop", "SELL", "31", "31", true) // W6, (라) 실패
	w.assertNoAttributionFailure("W6 (라) 실패")
	if n := w.recognitionMarkers(id); n != 0 {
		t.Fatalf("표지 %d 개, (라)가 깨진 체결은 인정하지 않음", n)
	}
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("활성 = %q, 이미 선 수량 불일치 하나뿐이어야 함(F10 — 새 행 없음)", got)
	}
}

// TestA200ARestartAfterARecognisedFillLeavesTheQuantityBlockAwaiting 는 공시 시험(design D6, codex r1 P2-2): 인정 뒤 다음 일치 전에
// 재기동하면 조정 반영 근거(프로세스 기억)를 잃어 수량 불일치가 오늘처럼 조정 대기로 남음 — 귀속 실패는 여전히 없음.
func TestA200ARestartAfterARecognisedFillLeavesTheQuantityBlockAwaiting(t *testing.T) {
	w := newA200World(t)
	w.holdManaged()
	diff := w.mismatch("31", "0")
	if _, err := w.converge(diff); err != nil {
		t.Fatalf("converge: %v", err)
	}
	w.fill("o-stop", "SELL", "31", "31", true)
	if err := w.refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if _, err := w.observe(diff); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	restarted := &reconcile.Tracker{
		Clock: w.clk, Gate: noStaleGate(w.clk), Journal: w.journal, AccountRef: a200Account, Terminal: w.journal,
	}
	if err := restarted.Restore(context.Background()); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	w.tracker = restarted
	w.conv.Credit = restarted
	out, err := w.cleanCycle()
	w.logCycle("재기동 뒤 일치", out, err)
	if err != nil {
		t.Fatalf("재기동 뒤 일치 Observe: %v", err)
	}
	w.assertNoAttributionFailure("재기동")
	if got := w.activeCauses(); len(got) != 1 || got[0] != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("재기동 뒤 활성 = %q, 오늘처럼 수량 불일치가 조정 대기로 남음(공시된 한계)", got)
	}
}
