package engine

// a098 4.2·4.3 — 실행자의 죽음은 관측되고, 게이트를 잠그고, 그 이상은 안 한다.
//
// 넷을 한 파일에서 본다. 넷이 서로를 반증하기 때문이다:
//
//	R3   죽으면 잠긴다
//	R13  **정상 종료는 안 잠근다** — 두 경우. ① 런타임 취소 ② 감독 루프 하나가 실패해
//	     엔진이 내려가는 경우. ②를 안 보면 부모 ctx 로 판정하는 구현이 통과한다(A-P2)
//	R15  패닉도 죽음이고, 그것이 프로세스를 안 죽인다
//	R18  죽어도 **운영 모드는 안 움직인다** (사용자 결정 11-1)
//
// R3만 있으면 「반환하면 무조건 잠근다」가 통과하고, 그 구현은 SIGTERM 한 번에
// 다음 기동을 운영자 승인 대기로 만든다.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a098TestGate is a bare entry gate — no journal, no broker. What is under test is
// which latch ends up in it, and a real wiring would only add ways to be wrong.
func a098TestGate() *execgw.EntryGate {
	return execgw.NewEntryGate(clock.System(), nil)
}

// a098CountingEscalation records every operating-mode escalation. R18 asserts the
// count is zero, so the fake has to be able to say "none" as a measurement rather
// than as the absence of a fake.
type a098CountingEscalation struct {
	mu    sync.Mutex
	calls int
}

func (e *a098CountingEscalation) EscalateOperatingMode(_ context.Context,
	_, _ string, _ journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	return journal.OperatingModeRecord{}, false, nil
}

func (e *a098CountingEscalation) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.calls
}

// a098DeliveryAux builds the executor the way production does: the stop handler is
// the same function, so a test that passes here is not testing a different rule.
func a098DeliveryAux(gate *execgw.EntryGate, run func(context.Context) error) AuxiliaryExecutor {
	return AuxiliaryExecutor{
		Name: alertDeliveryName,
		Run:  run,
		OnStop: func(_ context.Context, err error) {
			blockEntryOnDeliveryStop(gate, nil, err)
		},
	}
}

// a098RunWithAux runs a runtime with one surviving loop and one auxiliary.
func a098RunWithAux(t *testing.T, aux AuxiliaryExecutor,
	loop SupervisedLoop, esc execgw.ModeEscalation) (context.CancelFunc, <-chan error) {
	t.Helper()
	rt, err := NewRuntime(RuntimeOptions{
		AccountRef: "acct-a098",
		Escalate:   esc,
		Loops:      []SupervisedLoop{loop},
		Auxiliary:  []AuxiliaryExecutor{aux},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	// finished is separate from done on purpose: a test that reads done drains it,
	// and a cleanup waiting on the same channel would then block until its own
	// timeout and report a stall that never happened.
	finished := make(chan struct{})
	go func() {
		done <- rt.Run(ctx)
		close(finished)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Error("Run did not return")
		}
	})
	return cancel, done
}

func a098LiveLoop(name string) SupervisedLoop {
	return SupervisedLoop{
		Name: name,
		Run:  func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
	}
}

// a098WaitForBlock polls the gate until the reason appears or the deadline passes.
func a098WaitForBlock(gate *execgw.EntryGate, reason execgw.ReasonCode, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if _, ok := gate.Blocks()[reason]; ok {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	_, ok := gate.Blocks()[reason]
	return ok
}

// TestAStoppedDelivererLocksTheEntryGate is R3.
func TestAStoppedDelivererLocksTheEntryGate(t *testing.T) {
	gate := a098TestGate()
	esc := &a098CountingEscalation{}
	dead := errors.New("the ledger went away")
	returned := make(chan struct{})

	a098RunWithAux(t, a098DeliveryAux(gate, func(context.Context) error {
		close(returned)
		return dead
	}), a098LiveLoop("exit"), esc)

	<-returned
	if !a098WaitForBlock(gate, execgw.ReasonAlertSenderDown, 5*time.Second) {
		t.Fatalf("the gate holds %v; 보낼 주체가 없는데 진입이 열려 있다", gate.Blocks())
	}
	// 사유가 하나여야 한다 — 전달 실패(ReasonAlertUndelivered)와 합치면 운영자가
	// 밀린 행을 승인하는 순간 「보낼 주체 없음」까지 함께 풀린다(spec, 결정 8-1).
	if _, wrong := gate.Blocks()[execgw.ReasonAlertUndelivered]; wrong {
		t.Error("the stop also latched ReasonAlertUndelivered; 두 사유는 합치면 안 된다")
	}
	// R18: 게이트는 잠기고 모드는 안 움직인다 — 둘을 한 테스트에서 본다.
	if esc.count() != 0 {
		t.Errorf("the stop escalated the operating mode %d time(s); 결정 11-1은 그것을 금지한다",
			esc.count())
	}
}

// TestCancellingTheRuntimeDoesNotLockTheEntryGate is R13 ①.
//
// 정상 종료를 죽음으로 오인하면 **다음 기동이 아무 이유 없이 운영자 승인을 요구한다.**
func TestCancellingTheRuntimeDoesNotLockTheEntryGate(t *testing.T) {
	gate := a098TestGate()
	esc := &a098CountingEscalation{}
	running := make(chan struct{})

	cancel, done := a098RunWithAux(t, a098DeliveryAux(gate, func(ctx context.Context) error {
		close(running)
		<-ctx.Done()
		return ctx.Err()
	}), a098LiveLoop("exit"), esc)

	<-running
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v on cancellation, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Fatalf("a graceful stop latched %v; SIGTERM 은 사건이 아니다", blocks)
	}
	if esc.count() != 0 {
		t.Errorf("a graceful stop escalated the operating mode %d time(s)", esc.count())
	}
}

// TestASupervisedLoopFailingDoesNotLockTheEntryGate is R13 ②, and it is the case
// that separates the two contexts (5라운드 A-P2).
//
// 감독 루프 하나가 실패하면 `Runtime.Run`은 **부모 ctx 는 그대로 둔 채 `loopCtx`만**
// 취소한다. 판정에 부모를 넣으면 `ctx.Err() == nil` 이므로 **아무 잘못도 없는 배달
// 실행자의 정상 반환이 죽음으로 오인되고**, 엔진이 내려가는 그 순간에
// `ReasonAlertSenderDown` 이 걸린다 — 다음 기동이 운영자 승인을 기다리게 된다.
func TestASupervisedLoopFailingDoesNotLockTheEntryGate(t *testing.T) {
	gate := a098TestGate()
	boom := errors.New("the reconciliation driver hit a state nobody wrote")
	running := make(chan struct{})

	rt, err := NewRuntime(RuntimeOptions{
		AccountRef: "acct-a098",
		Loops: []SupervisedLoop{
			{Name: "reconcile", Run: func(context.Context) error {
				<-running // 실행자가 먼저 돌기 시작한 뒤에 죽는다
				return boom
			}},
		},
		Auxiliary: []AuxiliaryExecutor{a098DeliveryAux(gate, func(ctx context.Context) error {
			close(running)
			<-ctx.Done()
			return ctx.Err()
		})},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	// 부모 ctx 는 **한 번도 취소하지 않는다.** 엔진을 내리는 것은 감독 루프의 실패다.
	runErr := rt.Run(context.Background())
	if !errors.Is(runErr, ErrLoopFailed) {
		t.Fatalf("Run returned %v, want ErrLoopFailed", runErr)
	}
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Fatalf("the alert gate latched %v when a *supervised loop* failed; "+
			"판정이 부모 ctx 를 보고 있다 — `loopCtx` 여야 한다", blocks)
	}
}

// TestAPanickingDelivererDoesNotKillTheEngine is R15.
//
// ⛔ 이 테스트는 `recover`가 **있는 상태에서만** 같은 프로세스에서 돌릴 수 있다.
// 없는 상태의 RED 는 테스트 바이너리를 죽이므로 단언에 도달하지 못한다(§3의 B-P6) —
// 그래서 RED 는 **뮤테이션으로** 관측한다: `runAuxiliary`의 `recover`를 지우면
// 이 테스트가 패닉으로 패키지 전체를 죽인다. 그 죽음 자체가 빨강이다.
func TestAPanickingDelivererDoesNotKillTheEngine(t *testing.T) {
	gate := a098TestGate()
	esc := &a098CountingEscalation{}
	panicked := make(chan struct{})
	observerAlive := make(chan struct{})

	_, done := a098RunWithAux(t, a098DeliveryAux(gate, func(context.Context) error {
		close(panicked)
		panic("the publisher dereferenced something that was not there")
	}), SupervisedLoop{
		Name: "exit",
		Run: func(ctx context.Context) error {
			<-ctx.Done()
			close(observerAlive)
			return ctx.Err()
		},
	}, esc)

	<-panicked
	if !a098WaitForBlock(gate, execgw.ReasonAlertSenderDown, 5*time.Second) {
		t.Fatalf("a panic left the gate holding %v; 패닉도 정지다", gate.Blocks())
	}
	// 엔진이 살아 있어야 한다 — 패닉이 프로세스를 죽였다면 여기 도달하지 못한다.
	select {
	case err := <-done:
		t.Fatalf("Run returned %v after the executor panicked; 알림 패닉이 손절 루프를 죽였다", err)
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case <-observerAlive:
		t.Error("the exit observer stopped when the executor panicked")
	default:
	}
	if esc.count() != 0 {
		t.Errorf("a panic escalated the operating mode %d time(s)", esc.count())
	}
}

// TestTheStopDetailNamesNoAlertContent pins 불변식 8 at the one string this change
// puts where every reader of the gate sees it.
//
// 원장의 행은 제목·본문·payload·계좌를 들고 온다(`alertSelect`). 게이트 detail 은
// 게이트 상태가 읽히는 **모든** 자리에서 함께 읽히므로, 그 문자열이 알림 내용을
// 실어 나르면 이 change 가 새는 자리를 하나 만든 것이다.
func TestTheStopDetailNamesNoAlertContent(t *testing.T) {
	gate := a098TestGate()
	secret := "UNRESOLVED_IN_DOUBT: 005930 계좌 12345678"
	blockEntryOnDeliveryStop(gate, nil, errors.New(secret))

	detail, ok := gate.Blocks()[execgw.ReasonAlertSenderDown]
	if !ok {
		t.Fatal("the gate was not latched")
	}
	for _, leak := range []string{"005930", "12345678", "UNRESOLVED_IN_DOUBT"} {
		if strings.Contains(detail, leak) {
			t.Errorf("the gate detail carries %q: %q", leak, detail)
		}
	}
	if detail == "" {
		t.Error("the gate detail is empty; 운영자는 왜 막혔는지 알아야 한다")
	}
}

// TestTheDelivererReportsCancellationAsCancellation is the convention this file
// depends on, asserted at its source.
//
// ⛔ 4.0 은 `Run` 이 취소에서 **nil** 을 반환하게 짰고 그 주석은
// *"a clean stop is not a failure"* 였다. **런타임의 판정과 맞지 않는다**:
// `gracefulStop` 은 `err != nil && errors.Is(err, ctx.Err())` 를 요구하므로
// (`runtime.go`), nil 반환은 **정상 종료를 죽음으로 만든다.**
// 판정을 새로 만들지 않기로 한 이상(design D8.3) 규약은 반환 쪽이 맞춘다.
func TestTheDelivererReportsCancellationAsCancellation(t *testing.T) {
	j := openTestJournal(t)
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	d := &alertDeliverer{
		Journal: j, Publisher: &a098RecordingPublisher{}, Clock: fake,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()

	if !fake.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the deliverer never slept")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v on cancellation, want context.Canceled — "+
				"런타임의 gracefulStop 이 그것을 요구한다", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
	// 원장을 실제로 쓴 적이 없어야 이 테스트가 취소만 재는 것이 된다.
	if n, err := j.UndeliveredCount(context.Background()); err != nil || n != 0 {
		t.Errorf("UndeliveredCount = %d, %v; want 0, nil", n, err)
	}
}
