//go:build tossos_testseams

package engine

// a124 tasks 2.6 (a) — 격리 단언(행동). 배달 실행자가 게이트 잠금(g.mu)을 기다리는 동안 — 전략 진입
// dispatch 가 브로커 전송 동안 그 잠금을 쥔 모양 — 손절 쪽의 동기 Notify 는 실행자를 기다리지 않는다.
// 실행자는 Notifier 의 뮤텍스(n.mu)를 잡지 않기 때문이다(design D7, freeze Q1). 구조 쪽 단언은 태그 없는
// `a124_the_executor_holds_no_stop_lock_internal_test.go` 에 있다.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestAStopNotifyDoesNotWaitWhileTheExecutorWaitsForTheGate(t *testing.T) {
	f := a124Setup(t, a124Failing())
	f.row(t, "a124-2.6a")
	f.cycles(t, alertAttemptLimit-1)

	waiting := make(chan struct{})
	a124AtStage(f.d, alertStageSettled, func() { close(waiting) })

	release := execgw.HoldEntryGateLockForTest(f.gate) // 전략 dispatch 가 전송 동안 g.mu 를 쥔다
	defer release()
	executorDone := make(chan struct{})
	go func() { defer close(executorDone); _ = f.d.cycle(context.Background()) }()
	select {
	case <-waiting:
	case <-time.After(10 * time.Second):
		t.Fatal("the executor never reached its judgement")
	}
	// 실행자는 이제 세대 읽기에서 g.mu 를 기다린다(판정 직전).

	// 손절 쪽의 동기 알림 — 같은 원장 · 같은 게이트를 쓰는 Notifier, 즉답하는 전송 수단.
	stop := &obs.Notifier{Publisher: &a098RecordingPublisher{}, Journal: f.j, Gate: f.gate,
		AccountRef: a124Account, Clock: f.clk, Attempts: 1}
	notified := make(chan error, 1)
	start := time.Now()
	go func() {
		notified <- stop.Notify(context.Background(), obs.Event{
			Type: obs.EventExitProposalRefused, Key: "a124-2.6a-stop",
			Title: "청산 주문이 제출되지 않았다", Body: "손절이 나가지 않았다",
		})
	}()
	select {
	case err := <-notified:
		if err != nil {
			t.Fatalf("stop Notify: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the stop alert waited on the executor, which is itself waiting on the gate lock")
	}
	t.Logf("stop Notify returned in %v while the executor waited on the gate lock", time.Since(start))
	select {
	case <-executorDone:
		t.Fatal("the executor finished while the gate lock was held — it was not waiting where this test says")
	default:
	}
	release()
	select {
	case <-executorDone:
	case <-time.After(10 * time.Second):
		t.Fatal("the executor did not finish after the gate lock was released")
	}
	if !f.latched() {
		t.Fatal("the delayed judgement did not latch after the lock came back")
	}
}

// R6 · Y5 (2.12): 전달 정산 오류 경로는 임차를 쥔 채 세대를 읽는다. 그 읽기가 게이트 잠금을 기다리는 동안
// 임차가 만료되면 다른 발송자가 가져가 다시 보낼 수 있다 — 허용되고 기록된 성질이다(재발행 억제는 임차가 살아
// 있고 교체되지 않은 동안뿐). 판정은 잠금이 돌아온 뒤 그대로 선다.
func TestALeaseCanLapseWhileTheJudgementWaitsForTheGate(t *testing.T) {
	f := a124Setup(t, &a098RecordingPublisher{})
	id := f.row(t, "a124-2.12-lapse")
	f.led.failDeliver = 1
	waiting := make(chan struct{})
	a124AtStage(f.d, alertStageSettled, func() { close(waiting) })

	release := execgw.HoldEntryGateLockForTest(f.gate)
	defer release()
	done := make(chan struct{})
	go func() { defer close(done); _ = f.d.cycle(context.Background()) }()
	select {
	case <-waiting:
	case <-time.After(10 * time.Second):
		t.Fatal("the executor never reached its judgement")
	}
	f.clk.Advance(journal.DefaultAlertLease + time.Second)
	claim, err := f.j.ClaimAlertByID(context.Background(), id, "a124-other-sender")
	if err != nil || claim.Disposition != journal.ClaimAcquired || !claim.Stole {
		t.Fatalf("after the lease lapsed another sender could not take the row: %+v %v", claim, err)
	}
	// 그 발송자가 실제로 다시 보내고 정산한다 — 허용된 재발행이 일어나고 원장에 남는다(codex 3회차 T5).
	pub := f.d.Publisher.(*a098RecordingPublisher)
	before := pub.count()
	if err := pub.Publish(context.Background(), obs.Notification{Title: "t"}); err != nil {
		t.Fatalf("republish: %v", err)
	}
	if res, err := f.j.MarkAlertDelivered(context.Background(), id, claim.Token); err != nil || res.Outcome != journal.SettleApplied {
		t.Fatalf("the republication was not recorded: %+v %v", res, err)
	}
	if pub.count() != before+1 {
		t.Fatalf("publishes %d → %d, want one republication", before, pub.count())
	}
	release()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the executor did not finish")
	}
	if !f.latched() {
		t.Fatal("the delayed judgement did not latch")
	}
	if f.mode(t) != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s — the delayed judgement did not escalate", f.mode(t))
	}
}
