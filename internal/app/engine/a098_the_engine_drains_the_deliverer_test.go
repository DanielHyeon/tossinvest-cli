package engine

// a098 R7 — 배달 실행자가 붙어도 엔진은 상한 안에 멈춘다.
//
// 이 테스트만 실행자와 런타임을 **둘 다 진짜로** 세운다. 옆의 두 파일은 각각 한쪽만
// 본다 — `a098_the_outbox_gets_emptied_test.go` 는 실행자를 런타임 없이,
// `a098_the_runtime_starts_what_it_does_not_supervise_test.go` 는 런타임을 가짜
// 실행자로. 붙였을 때만 나오는 실패가 있다: **`wg` 가 담는 goroutine 이 하나 늘어
// 못 죽는 경로도 하나 는다**(A-T3). 정지하지 못하는 엔진은 사람이 손절 배선을 고치러
// 들어갈 수 없다.
//
// 가짜 시계를 **한 번도 안 전진시킨다.** `≤ 주기` 로 쓰면 취소 불가능한
// `Sleep(주기)` 가 통과한다(B-P4) — 전진이 없으면 취소 가능한 대기만 깨어난다.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestTheEngineStopsPromptlyWithTheDelivererAttached(t *testing.T) {
	j := openTestJournal(t)
	fake := clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	d := &alertDeliverer{
		Journal:   j,
		Publisher: &a098RecordingPublisher{},
		Clock:     fake,
		Interval:  alertDeliveryInterval,
		Batch:     alertDeliveryBatch,
		Claimant:  "a098-test",
	}

	observerStopped := make(chan struct{})
	rt, err := NewRuntime(RuntimeOptions{
		AccountRef: "acct-a098",
		Loops: []SupervisedLoop{{
			Name: "exit",
			Run: func(ctx context.Context) error {
				<-ctx.Done()
				close(observerStopped)
				return ctx.Err()
			},
		}},
		Auxiliary: []AuxiliaryExecutor{{Name: "alert-delivery", Run: d.Run}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()

	// 가짜 시계에 잠든 것이 있다는 것이 **런타임이 실제로 이 실행자를 기동했다**는
	// 관측이다. 등록만 하고 안 띄우는 구현은 여기서 걸린다.
	if !fake.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the deliverer never slept — the runtime did not start it")
	}

	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v on cancellation, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return within 2s of cancellation — 배달 실행자가 엔진을 붙잡고 있다")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Run took %v to return; 주기 대기가 취소 가능하지 않다", elapsed)
	}
	select {
	case <-observerStopped:
	default:
		t.Error("the exit loop was not drained")
	}

	// 원장이 아직 살아 있어야 한다 — Run 이 반환한 시점에 실행자의 쓰기가 끝나 있다는
	// 것이 wg 계약이고, 그 계약이 깨졌다면 여기서 닫는 것이 경합이 된다.
	if _, err := j.UndeliveredCount(context.Background()); err != nil {
		t.Errorf("the journal is not usable after Run returned: %v", err)
	}
}
