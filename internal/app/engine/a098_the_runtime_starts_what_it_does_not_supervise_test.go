package engine_test

// a098 4.1 — 런타임은 감독하지 않는 일도 기동하고 배수한다.
//
// 배달 실행자를 `Loops`에 넣으면 그 반환이 **엔진 전체를 내린다** — 방어적 종료 계약이
// 그렇게 하도록 되어 있고, 그것이 옳은 계약이다(감독 루프에 대해서는). 알림 배달은
// 그런 실패가 아니다: 배달이 멈춘 동안에도 손절·비상 청산은 계속되어야 한다
// (안전 불변식 4 · 사용자 결정 9-2).
//
// 그래서 등록 자리를 바꾼다. **판정으로 거르지 않는다** — 감독 루프로 등록한 뒤
// "이 이름이면 다르게 다룬다"로 피하는 구성은 판정이 틀리면 무너진다.
// `stops` 채널에 아무것도 안 보내면 **틀릴 판정이 없다.**

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a098SurvivingLoop is a supervised loop that returns only when cancelled. It is
// the observer that must still be alive after an auxiliary executor dies.
func a098SurvivingLoop(name string, stopped chan<- struct{}) engine.SupervisedLoop {
	return engine.SupervisedLoop{
		Name: name,
		Run: func(ctx context.Context) error {
			<-ctx.Done()
			close(stopped)
			return ctx.Err()
		},
	}
}

// TestTheDelivererAsASupervisedLoopStopsTheEngine is R2's **대조군**.
//
// 이것이 오늘 통과한다. 그 통과가 R2의 관측을 의미 있게 만든다 — 아래 테스트가
// "안 내려간다"를 보는데, 그 성질이 **등록 자리 때문**이라는 것은 같은 함수가
// 감독 루프일 때 **실제로 내려가는 것**을 봐야만 말할 수 있다. 대조군이 없으면
// 「원래 안 내려가는 구성」도 초록이다(B-P3).
func TestTheDelivererAsASupervisedLoopStopsTheEngine(t *testing.T) {
	alerts := &recordingAlerts{}
	dead := errors.New("the delivery executor stopped")
	observerStopped := make(chan struct{})

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops: []engine.SupervisedLoop{
			{Name: "alert-delivery", Run: func(context.Context) error { return dead }},
			a098SurvivingLoop("exit", observerStopped),
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	runErr := rt.Run(context.Background())
	if !errors.Is(runErr, engine.ErrLoopFailed) {
		t.Fatalf("Run returned %v, want ErrLoopFailed — 감독 루프의 자발적 반환은 엔진을 내린다", runErr)
	}
	if !strings.Contains(runErr.Error(), "alert-delivery") {
		t.Errorf("the failure %q does not name the executor", runErr)
	}
	select {
	case <-observerStopped:
	default:
		t.Error("the exit observer was not stopped; 대조군이 대조군이 아니다")
	}
	if n := alerts.count(obs.EventEngineLoopFailed); n != 1 {
		t.Errorf("the control raised %d loop-failed alerts, want 1", n)
	}
}

// TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine is R2.
//
// 위 대조군과 **같은 함수**를 `Auxiliary`로 옮긴다. 다른 것은 등록 자리 하나뿐이다.
func TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine(t *testing.T) {
	alerts := &recordingAlerts{}
	dead := errors.New("the delivery executor stopped")
	observerStopped := make(chan struct{})
	returned := make(chan struct{})

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     alerts,
		Loops:      []engine.SupervisedLoop{a098SurvivingLoop("exit", observerStopped)},
		Auxiliary: []engine.AuxiliaryExecutor{{
			Name: "alert-delivery",
			Run: func(context.Context) error {
				close(returned)
				return dead
			},
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	select {
	case <-returned:
	case <-time.After(5 * time.Second):
		t.Fatal("the auxiliary executor never ran — 등록이 아무것도 기동하지 않았다")
	}

	// 실행자가 죽은 뒤에도 엔진이 살아 있어야 한다. 100ms 는 상한이 아니라
	// **아직 안 죽었다**를 보는 창이다 — 죽는다면 `stops` 수신이 즉시 깨어난다.
	select {
	case err := <-done:
		t.Fatalf("Run returned %v when the auxiliary executor died; 알림 버그가 손절 부재가 됐다", err)
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case <-observerStopped:
		t.Fatal("the exit observer was cancelled by an auxiliary executor's death")
	default:
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v on cancellation, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
	if n := alerts.count(obs.EventEngineLoopFailed); n != 0 {
		t.Errorf("an auxiliary executor's death raised %d loop-failed alerts, want 0 — "+
			"그것은 감독 루프의 사건이다", n)
	}
}

// TestTheRuntimeDrainsAnAuxiliaryExecutorBeforeItReturns is R14.
//
// `Runtime.Run`의 doc 이 계약을 적는다 — *"when this returns, every goroutine it
// started has returned too. That is what lets the caller close the journal
// immediately afterwards without racing a loop that is still writing."*
// 배달 실행자는 **원장에 쓴다**(claim·정산). 그래서 이 계약이 실행자에게도 걸려야 한다.
//
// 둘 다 본다. ①만 보면 **영원히 안 돌아오는 `wg.Wait()`도 통과하고**,
// ②만 보면 **아예 안 기다리는 구현이 통과한다**(6라운드 P2).
func TestTheRuntimeDrainsAnAuxiliaryExecutorBeforeItReturns(t *testing.T) {
	writing := make(chan struct{})
	release := make(chan struct{})
	observerStopped := make(chan struct{})

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     &recordingAlerts{},
		Loops:      []engine.SupervisedLoop{a098SurvivingLoop("exit", observerStopped)},
		Auxiliary: []engine.AuxiliaryExecutor{{
			Name: "alert-delivery",
			Run: func(context.Context) error {
				close(writing)
				// 취소를 일부러 안 본다: 원장 쓰기 한가운데는 되돌릴 수 없는 자리이고,
				// 계약은 "취소되면 즉시 죽는다"가 아니라 "끝날 때까지 기다린다"이다.
				<-release
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	select {
	case <-writing:
	case <-time.After(5 * time.Second):
		t.Fatal("the auxiliary executor never started")
	}
	cancel()

	// ① 쓰기가 안 풀렸는데 Run 이 반환하면 원장이 그 아래에서 닫힌다.
	select {
	case err := <-done:
		t.Fatalf("Run returned %v while the executor was still writing to the journal; "+
			"호출자는 그 직후 원장을 닫는다", err)
	case <-time.After(200 * time.Millisecond):
	}

	// ② 풀린 뒤에는 고정 상한 안에 반환해야 한다 — 안 그러면 엔진이 못 죽는다.
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return within 2s of the executor finishing; 엔진이 정지하지 못한다")
	}
}

// TestAnAuxiliaryExecutorIsNotInTheSupervisedLoopSet pins the spec Scenario's last
// clause: 배달 실행자의 이름은 감독 루프 집합에 **나타나지 않는다**.
func TestAnAuxiliaryExecutorIsNotInTheSupervisedLoopSet(t *testing.T) {
	stopped := make(chan struct{})
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Loops:      []engine.SupervisedLoop{a098SurvivingLoop("exit", stopped)},
		Auxiliary: []engine.AuxiliaryExecutor{{
			Name: "alert-delivery",
			Run:  func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	names := rt.LoopNames()
	if len(names) != 1 || names[0] != "exit" {
		t.Fatalf("LoopNames() = %v, want [exit] — 보조 실행자가 감독 집합에 들어갔다", names)
	}
}

// TestTheRuntimeRefusesAMisWiredAuxiliaryExecutor covers the four refusals.
//
// 중복 이름을 **감독 루프와 함께** 본다: 이름은 정지 기록이 무엇을 가리키는지를 정하고,
// 두 집합이 이름을 나눠 쓰면 그 기록이 어느 쪽인지 말할 수 없다.
func TestTheRuntimeRefusesAMisWiredAuxiliaryExecutor(t *testing.T) {
	live := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	loops := []engine.SupervisedLoop{{Name: "exit", Run: live}}

	cases := []struct {
		name string
		aux  []engine.AuxiliaryExecutor
		want string
	}{
		{
			name: "이름이 없다",
			aux:  []engine.AuxiliaryExecutor{{Name: "  ", Run: live}},
			want: "no name",
		},
		{
			name: "Run 이 없다",
			aux:  []engine.AuxiliaryExecutor{{Name: "alert-delivery"}},
			want: "no Run function",
		},
		{
			name: "보조끼리 이름이 겹친다",
			aux: []engine.AuxiliaryExecutor{
				{Name: "alert-delivery", Run: live},
				{Name: "alert-delivery", Run: live},
			},
			want: "alert-delivery",
		},
		{
			name: "감독 루프와 이름이 겹친다",
			aux:  []engine.AuxiliaryExecutor{{Name: "exit", Run: live}},
			want: "exit",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := engine.NewRuntime(engine.RuntimeOptions{
				AccountRef: runtimeAccount,
				Loops:      loops,
				Auxiliary:  tc.aux,
			})
			if err == nil {
				t.Fatal("NewRuntime accepted the wiring; 조립 시점에 거절해야 한다")
			}
			if !errors.Is(err, engine.ErrRuntimeUnavailable) {
				t.Errorf("error %v is not ErrRuntimeUnavailable", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal %q does not say %q", err, tc.want)
			}
		})
	}
}

// a098 컴파일 시점 확인: 보조 실행자 슬라이스가 여러 개를 담고 전부 기동된다.
func TestEveryAuxiliaryExecutorIsStarted(t *testing.T) {
	var mu sync.Mutex
	started := map[string]bool{}
	stopped := make(chan struct{})
	ready := make(chan struct{}, 2)

	mark := func(name string) func(context.Context) error {
		return func(ctx context.Context) error {
			mu.Lock()
			started[name] = true
			mu.Unlock()
			ready <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}
	}

	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Loops:      []engine.SupervisedLoop{a098SurvivingLoop("exit", stopped)},
		Auxiliary: []engine.AuxiliaryExecutor{
			{Name: "alert-delivery", Run: mark("alert-delivery")},
			{Name: "second", Run: mark("second")},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	for range 2 {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			t.Fatal("not every auxiliary executor was started")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
	mu.Lock()
	defer mu.Unlock()
	if !started["alert-delivery"] || !started["second"] {
		t.Fatalf("started = %v", started)
	}
}
