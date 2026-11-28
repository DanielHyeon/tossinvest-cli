package engine

// a198 보강 2(D10) — 세션 해소 실행자의 스윕 안에서 엔진이 낸 IN_DOUBT 취소를 닫는 자리의 성질(패키지 안 시험, a193 픽스처):
// 브로커 요청 0 · 실행자 알림 0 · 발주 해소보다 먼저 · 엔진이 낸 취소만(flatten 의 것은 무변) · ACKED 무변.

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// engineCancel 은 엔진 귀속 발주(소유 PLACE, 접수 확정 번호 orderID)와 그 주문을 겨눈 취소 attempt(state 까지)를 남기고,
// closed 가 참이면 그 주문의 무체결 취소 종결 기록을 체결 감지 입구로 씀. 취소 intent 의 source 를 정함.
func (f *a193Fixture) engineCancel(orderID, source string, to journal.AttemptState, closed bool) string {
	f.t.Helper()
	ctx := context.Background()
	f.ids++
	n := f.ids
	owner, err := f.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: fmt.Sprintf("i-a198-owner-%d", n), Market: "kr", TradingDay: "2026-03-30", AccountRef: a193Account,
			Symbol: a193Symbol, Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "3", Price: "1000",
			Currency: "KRW", Source: "engine", Fingerprint: fmt.Sprintf("fp-a198-owner-%d", n),
		},
		Kind: journal.KindPlace, AttemptID: fmt.Sprintf("a-a198-owner-%d", n), AccountRef: a193Account,
	})
	if err != nil {
		f.t.Fatalf("Prepare(owner): %v", err)
	}
	for _, step := range []func() error{
		func() error { return owner.MarkDispatchStarted(ctx) },
		func() error { return owner.MarkAcked(ctx, orderID) },
		func() error { return owner.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "") },
	} {
		if err := step(); err != nil {
			f.t.Fatalf("driving the owner: %v", err)
		}
	}
	id := fmt.Sprintf("a-a198-cancel-%d", n)
	c, err := f.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: fmt.Sprintf("i-a198-cancel-%d", n), Market: "kr", TradingDay: "2026-03-30", AccountRef: a193Account,
			Symbol: a193Symbol, Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "3", Price: "1000",
			Currency: "KRW", Source: source, Fingerprint: fmt.Sprintf("fp-a198-cancel-%d", n),
		},
		Kind: journal.KindCancel, AttemptID: id, AccountRef: a193Account, TargetOrderID: orderID,
	})
	if err != nil {
		f.t.Fatalf("Prepare(cancel): %v", err)
	}
	if to != journal.StateRecorded {
		if err := c.MarkDispatchStarted(ctx); err != nil {
			f.t.Fatal(err)
		}
	}
	switch to {
	case journal.StateRecorded, journal.StateDispatchStarted:
	case journal.StateInDoubt:
		if err := c.MarkInDoubt(ctx, journal.ReasonDispatchAmbiguous, "5xx"); err != nil {
			f.t.Fatal(err)
		}
	case journal.StateAcked:
		if err := c.MarkAcked(ctx, "C-"+orderID); err != nil {
			f.t.Fatal(err)
		}
	default:
		f.t.Fatalf("unsupported cancel state %s", to)
	}
	if closed {
		if _, err := f.journal.RecordFill(ctx, journal.FillObservation{
			OrderID: orderID, AccountRef: a193Account, Symbol: a193Symbol, Market: "kr", TradingDay: "2026-03-30",
			Side: "BUY", State: "CANCELLED", Terminal: true, Quantity: "3", FilledQuantity: "0",
			ObservedAt: f.clk.Now().Format(time.RFC3339),
		}); err != nil {
			f.t.Fatalf("RecordFill: %v", err)
		}
	}
	return id
}

// gstack review r2 G6 — D10 의 「닫음」 줄은 종목을 실음(a193 해소 줄과 같은 필드), 대상 주문 번호는 싣지 않음.
func TestA198TheCancelClosedLineNamesTheSymbol(t *testing.T) {
	f := newA193Fixture(t)
	id := f.engineCancel("O-a198-8", engineSource, journal.StateInDoubt, true)
	var lines bytes.Buffer
	s := f.settler(nil, nil)
	s.deps.Log = obs.NewLogger(obs.LogOptions{Writer: &lines, JSON: true, Clock: f.clk})
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	var closed string
	for _, l := range strings.Split(lines.String(), "\n") {
		if strings.Contains(l, id) && strings.Contains(l, "closed this engine cancel") {
			closed = l
		}
	}
	if closed == "" || !strings.Contains(closed, `"`+obs.FieldSymbol+`":"`+a193Symbol+`"`) {
		t.Fatalf("closed line %q does not name symbol %s (log %s)", closed, a193Symbol, lines.String())
	}
	if strings.Contains(closed, "O-a198-8") {
		t.Fatalf("closed line carries the target order number: %s", closed)
	}
}

// D4 재결정(gstack review r2 G1) — 전용 로거가 없으면 취소 실패 줄 대기열은 늘 비어 있음.
func TestA198WithoutTheDedicatedLoggerTheCancelQueueStaysEmpty(t *testing.T) {
	o := &ExitObserver{}
	o.noteClearCancelFailure(managed{}, journal.LiveOrder{OrderID: "O-x"}, journal.StateFailedConfirmed, "x", true)
	if n := len(o.clearCancelPending); n != 0 {
		t.Fatalf("%d queued line(s) without a dedicated logger", n)
	}
	o.flushClearCancelFailures()
	if len(o.clearCancelLogged) != 0 {
		t.Fatalf("marks set without a dedicated logger: %v", o.clearCancelLogged)
	}
}

// §4-23 — 닫기는 브로커를 부르지 않고 실행자 알림도 내지 않음.
func TestA198TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert(t *testing.T) {
	f := newA193Fixture(t)
	id := f.engineCancel("O-a198-1", engineSource, journal.StateInDoubt, true)
	f.clk.Advance(time.Hour) // 10분 알림 한계를 넘김 — CANCEL 에는 그래도 알림 0 이어야 함
	s := f.settler(nil, nil)
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := f.state(id); got.State != journal.StateConfirmed || got.ReasonCode != journal.ReasonResolvedFound {
		t.Fatalf("cancel %s is %s|%s, want CONFIRMED|%s", id, got.State, got.ReasonCode, journal.ReasonResolvedFound)
	}
	if n := f.pager.count(); n != 0 || f.account.calls.Load() != 0 {
		t.Fatalf("broker reads during the close: lists %d, account %d", n, f.account.calls.Load())
	}
	if n := f.crit.calls(); n != 0 {
		t.Fatalf("%d settler alert(s) for an engine cancel — the clearing alerts own it", n)
	}
}

// §4-23 (Manager Q5) — 같은 스윕의 발주 해소가 막혀 있는 동안 취소는 이미 닫혀 있음(닫기가 먼저).
func TestA198TheCancelIsClosedBeforeThePlaceResolution(t *testing.T) {
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	id := f.engineCancel("O-a198-2", engineSource, journal.StateInDoubt, true)
	res := newA193BlockingResolver()
	s := f.settler(nil, res)
	done := make(chan error, 1)
	go func() { done <- s.sweep(context.Background()) }()
	select {
	case <-res.entered:
	case <-time.After(10 * time.Second):
		t.Fatal("the place resolution never started")
	}
	if got := f.state(id).State; got != journal.StateConfirmed {
		t.Fatalf("cancel %s is %s while the place resolution blocks, want CONFIRMED — closing comes first", id, got)
	}
	close(res.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// §4-32 (Manager Q7) — flatten 이 낸 취소는 대상이 종결이어도 닫지 않음. §4-30 — ACKED 취소도 무변.
func TestA198TheSettlerLeavesFlattenAndAckedCancelsAlone(t *testing.T) {
	f := newA193Fixture(t)
	flat := f.engineCancel("O-a198-3", "flatten-all", journal.StateInDoubt, true)
	acked := f.engineCancel("O-a198-4", engineSource, journal.StateAcked, true)
	// gstack review r2 state-machines: 진행 중 전송(RECORDED·DISPATCH_STARTED)인 엔진 취소도 대상이 종결이어도 무변.
	recorded := f.engineCancel("O-a198-5", engineSource, journal.StateRecorded, true)
	started := f.engineCancel("O-a198-6", engineSource, journal.StateDispatchStarted, true)
	s := f.settler(nil, nil)
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := f.state(flat).State; got != journal.StateInDoubt {
		t.Fatalf("flatten cancel %s moved to %s — the flatten saga settles its own cancels", flat, got)
	}
	if got := f.state(acked).State; got != journal.StateAcked {
		t.Fatalf("ACKED cancel %s moved to %s — the session never finishes an ACKED write", acked, got)
	}
	for id, want := range map[string]journal.AttemptState{recorded: journal.StateRecorded, started: journal.StateDispatchStarted} {
		if got := f.state(id).State; got != want {
			t.Fatalf("in-flight cancel %s moved %s → %s — the session never ends a transmission in progress", id, want, got)
		}
	}
}

// gstack review r2 testing — 원장 닫기가 실패해도 스윕은 발주 해소로 계속 가고(오류 반환 없음), 같은 실패는 한 번만 경고하며,
// 원장이 회복되면 다음 스윕에서 닫힘.
func TestA198AFailedCancelCloseDoesNotStopThePlaceResolution(t *testing.T) {
	f := newA193Fixture(t)
	place := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	id := f.engineCancel("O-a198-7", engineSource, journal.StateInDoubt, true)
	db, err := sql.Open("sqlite", "file:"+f.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TRIGGER a198_refuse_cancel_close BEFORE UPDATE OF state ON mutation_attempts
		WHEN NEW.kind = 'CANCEL' BEGIN SELECT RAISE(ABORT, 'a198 fixture: the ledger refuses the cancel close'); END`); err != nil {
		t.Fatal(err)
	}
	var lines bytes.Buffer
	res := newA193BlockingResolver()
	close(res.release)
	s := f.settler(nil, res)
	s.deps.Log = obs.NewLogger(obs.LogOptions{Writer: &lines, JSON: true, Clock: f.clk})
	for i := 0; i < 2; i++ {
		if err := s.sweep(context.Background()); err != nil {
			t.Fatalf("sweep %d returned %v — a failed cancel close must not stop the sweep", i, err)
		}
	}
	if got := f.state(id).State; got != journal.StateInDoubt {
		t.Fatalf("cancel %s is %s while the ledger refuses the close", id, got)
	}
	if res.calls[place] == 0 {
		t.Fatalf("the IN_DOUBT place was never resolved behind the failed cancel close (calls %v)", res.calls)
	}
	if n := strings.Count(lines.String(), id); n != 1 {
		t.Fatalf("%d warning line(s) name cancel %s over two sweeps, want exactly one:\n%s", n, id, lines.String())
	}
	if _, err := db.Exec(`DROP TRIGGER a198_refuse_cancel_close`); err != nil {
		t.Fatal(err)
	}
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := f.state(id).State; got != journal.StateConfirmed {
		t.Fatalf("cancel %s is %s after the ledger recovered, want CONFIRMED", id, got)
	}
}
