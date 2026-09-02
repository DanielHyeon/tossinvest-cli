package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 이 파일은 남은 빈칸 하나를 채운다: `invokeBoundedStrategyCycle` 의 B2
// (`analysis/function-logic/internal-app-engine--invokeboundedstrategycycle/ast.json`
// 의 `894:3`, block `894-896`). 패키지 전체 프로파일에서도 `count=0` 이었다.
//
// 왜 아무도 안 밟았나. 실제 취소에서는 `ctx.Done()` 갈래와 마감 시한 갈래가
// **둘 다 준비된다.** Go 는 준비된 갈래 중 하나를 무작위로 고르므로 시험은
// 어느 쪽이 돌았는지 고를 수 없고, 두 갈래의 반환값이 같아서 결과로도 구별할
// 수 없다. 그래서 이 블록은 "돌지 않는" 것이 아니라 **관측할 수 없는** 것이었다.
//
// 여기서는 `Done()` 이 절대 준비되지 않으면서 `Err()` 는 취소를 보고하는
// context 를 넣어 마감 시한 갈래만 열어 둔다. 그러면 B2 가 결정적으로 돌고,
// 이 갈래가 하는 일이 값으로 드러난다: **자기 타이머를 믿지 않고 취소를 다시
// 확인한다.** 확인을 지우면 취소가 `abnormal` 로 승격되고, 오늘의 임계값 1
// 아래에서 그것은 "정상 종료가 시장을 잠근다"가 된다.

// cancelledWithoutDoneChannel 은 취소를 **보고만 하고 신호하지 않는** context 다.
// `context.Context` 의 계약상 `Done()` 은 취소될 수 없는 context 에서 nil 일 수
// 있고, nil 채널의 수신은 영원히 막힌다. 그 둘을 합치면 select 의 첫 갈래가
// 닫히고 마감 시한 갈래만 남는다.
type cancelledWithoutDoneChannel struct{ context.Context }

func (cancelledWithoutDoneChannel) Done() <-chan struct{} { return nil }
func (cancelledWithoutDoneChannel) Err() error            { return context.Canceled }

func TestTheWatchdogRechecksCancellationInsteadOfTrustingItsOwnTimer(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	limit := 5 * time.Second

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	// 취소를 무시하는 사이클. 마감 시한이 울릴 때까지 반환하지 않으므로 select 가
	// 결과 갈래를 고를 수 없다.
	stubborn := func(context.Context) error { <-release; return nil }

	cases := []struct {
		name          string
		ctx           context.Context
		wantErr       error
		wantAbnormal  bool
		wantCancelled bool
	}{
		{
			// B2 가 참인 갈래 (block 894-896).
			name:          "이미 취소된 상위 아래에서 울린 마감 시한은 취소다",
			ctx:           cancelledWithoutDoneChannel{Context: context.Background()},
			wantErr:       context.Canceled,
			wantAbnormal:  false,
			wantCancelled: true,
		},
		{
			// B2 가 거짓인 갈래 — 대조군. 같은 타이머, 다른 상위 상태.
			name:          "살아 있는 상위 아래에서 울린 마감 시한은 비정상이다",
			ctx:           context.Background(),
			wantErr:       ErrStrategyCycleDeadline,
			wantAbnormal:  true,
			wantCancelled: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fake := clock.NewFake(base)
			type outcome struct {
				err                            error
				abnormal, cancelled, abandoned bool
			}
			done := make(chan outcome, 1)
			go func() {
				err, abnormal, cancelled, abandoned := invokeBoundedStrategyCycle(testCase.ctx, fake, limit, stubborn)
				done <- outcome{err, abnormal, cancelled, abandoned}
			}()
			if !fake.WaitForSleepers(1, time.Second) {
				t.Fatal("마감 시한 감시자가 잠들지 않았다")
			}
			fake.Advance(limit)
			select {
			case got := <-done:
				if !errors.Is(got.err, testCase.wantErr) || got.abnormal != testCase.wantAbnormal {
					t.Fatalf("err=%v abnormal=%v, want %v/%v\n\n"+
						"이 두 칸을 가르는 것은 894:3 의 재확인 하나다. 그것을 지우면 취소가"+
						" abnormal 로 올라가고, 생산 임계값이 1 이므로 정상 종료가 시장을 잠근다.",
						got.err, got.abnormal, testCase.wantErr, testCase.wantAbnormal)
				}
				// 두 갈래 모두 늦게 오는 결과를 버린다. 그것이 "abandoned" 의 뜻이다.
				if !got.abandoned {
					t.Fatalf("abandoned=false — 마감 시한을 넘긴 사이클이 버려진 것으로 표시되지 않았다")
				}
				if got.cancelled != testCase.wantCancelled {
					t.Fatalf("cancelled=%v, want %v — 취소와 비정상은 같은 사이클에서 함께 참일 수 없다",
						got.cancelled, testCase.wantCancelled)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("마감 시한 갈래가 반환하지 않았다")
			}
		})
	}
}
