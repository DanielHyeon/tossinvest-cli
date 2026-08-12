package engine

// a098 R14 — 보조 실행자를 다 기다린 뒤에 원장이 닫힌다.
//
// `Runtime.Run` 의 doc 이 계약으로 적어 둔 문장이 있다 — *"when this returns, every
// goroutine it started has returned too … without racing a loop that is still
// writing"*(`runtime.go:258-260`). 배달 실행자는 **원장에 쓴다.** 그러니 이 계약이
// 깨지면, `Run` 이 반환한 직후 호출자가 닫는 원장을 **아직 쓰고 있는 실행자**가 만난다.
//
// 답은 새 WaitGroup 이 아니라 **기존 `wg` 재사용**이었다(design D8.3). 위험이 0 이 된 게
// 아니라 **잊을 수 있는 자리가 하나(`wg.Add`)로 줄었을 뿐**이고, 이 파일이 그 한 자리를 잰다.
//
// # 왜 하필 「취소된 뒤에도 남는 쓰기」를 골랐나
//
// 취소된 ctx 로는 원장 쓰기가 거의 다 실패한다. 실패하는 쓰기를 기다리는 것은 계약을
// 안 재고도 초록이 된다. 예외가 딱 하나 있고 그것은 **일부러** 분리돼 있다 —
// `alertDeliverer.release` 의 `context.WithoutCancel`(`alertdelivery.go:250`).
// 취소된 주기가 임차를 쥔 채 사라지면 만료(81s)까지 아무도 그 행을 못 보내기 때문이다.
// 즉 **취소 이후에 반드시 일어나는 원장 쓰기**가 저기 하나 있고, 그것이 `Run` 반환보다
// 먼저 끝나는가가 이 계약 그 자체다.
//
// # 두 방향을 다 본다
//
//	① 쓰기가 안 풀리면 `Run` 이 **안 돌아온다**
//	② 풀리면 **고정 상한 안에** 돌아온다
//
// ①만 보면 **영원히 안 돌아오는 `wg.Wait()` 가 통과한다**(6라운드 P2).
// 그리고 ③으로 그 쓰기가 **원장에 실제로 닿았는지**까지 본다 — ①·②는 순서만 재므로,
// 닿지도 않은 쓰기를 기다린 구현도 순서로는 옳아 보인다.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// a098LedgerCloseNow 는 가짜 시계의 출발점이다. 이 테스트는 시계를 **한 번도 안
// 전진시킨다** — 전진이 없으면 취소 가능한 대기만 깨어나므로, `Sleep` 으로 버티는
// 구현이 상한 안에 들어오는 척할 수 없다.
var a098LedgerCloseNow = time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)

func TestTheRuntimeWaitsForTheExecutorsLedgerWriteBeforeReturning(t *testing.T) {
	j := openTestJournal(t)
	id := a098ParkedAlert(t, j, "a098-r14")

	// 전송은 **실패**시킨다. 성공 경로의 정산(`MarkAlertDelivered`)은 주기의 ctx 를
	// 쓰므로 취소 뒤에는 실패하고 아무 쓰기도 안 남긴다 — 그러면 잴 것이 없어진다.
	pub := &a098BlockingPublisher{
		entered: make(chan struct{}),
		release: make(chan struct{}),
		fail:    errors.New("transport down"),
	}
	fake := clock.NewFake(a098LedgerCloseNow)
	d := &alertDeliverer{
		Journal:   j,
		Publisher: pub,
		Clock:     fake,
		Interval:  alertDeliveryInterval,
		Batch:     alertDeliveryBatch,
		Claimant:  "a098-r14",
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

	// 여기 닿았다는 것은 실행자가 이 행의 임차를 **이미 쥐었다**는 뜻이고,
	// 아직 안 끝난 원장 쓰기(임차 반납) 하나가 그 앞에 남아 있다는 뜻이다.
	select {
	case <-pub.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("실행자가 전송에 진입하지 못했다 — 런타임이 실행자를 안 띄웠다")
	}

	cancel()

	// ① 먼저 취소가 런타임까지 갔는지 본다. 감독 루프가 안 빠져나온 상태라면
	//    「Run 이 아직 안 돌아왔다」는 아무것도 안 재고 참이 된다.
	select {
	case <-observerStopped:
	case <-time.After(5 * time.Second):
		t.Fatal("취소가 감독 루프까지 안 갔다")
	}
	select {
	case err := <-done:
		t.Fatalf("Run 이 실행자의 원장 쓰기를 안 기다리고 반환했다 (err=%v) — "+
			"호출자가 이 시점에 원장을 닫으면 실행자의 쓰기와 경합한다", err)
	case <-time.After(250 * time.Millisecond):
	}

	// ② 이제 풀어 준다. 상한은 「유한」이 아니라 **고정 값**이어야 한다 —
	//    영원히 안 돌아오는 `wg.Wait()` 도 ①은 통과하기 때문이다.
	released := time.Now()
	close(pub.release)
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run 이 취소에 %v 를 냈다; nil 이어야 한다", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("쓰기가 풀린 뒤에도 Run 이 2초 안에 안 돌아왔다")
	}
	if waited := time.Since(released); waited > time.Second {
		t.Errorf("쓰기가 풀린 뒤 Run 반환까지 %v 걸렸다 — 대기가 상한을 넘는다", waited)
	}

	// ③ 그 쓰기가 원장에 실제로 닿았는가. 취소된 주기가 임차를 쥔 채 사라졌다면
	//    이 행은 만료까지 아무도 못 보낸다.
	row, err := j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if row.ClaimedBy != "" || row.ClaimExpiresAt != nil {
		t.Errorf("임차가 아직 %q 에게 잡혀 있다 (만료=%v) — 취소 뒤 반납 쓰기가 원장에 안 닿았다",
			row.ClaimedBy, row.ClaimExpiresAt)
	}

	// 호출자가 `Close` 를 부르는 자리가 바로 여기다. 원장이 아직 성해야 한다.
	if _, err := j.UndeliveredCount(context.Background()); err != nil {
		t.Errorf("Run 반환 뒤 원장을 못 쓴다: %v", err)
	}
}
