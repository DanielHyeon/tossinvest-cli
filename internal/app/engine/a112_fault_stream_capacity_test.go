package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// latchMarket 의 마지막 갈래(B9, `961:2`)에는 `default` 팔이 있다: fault 를
// 건네지 못하면 잠금을 **오류로 되돌리고**, runMarket 이 그것을 중앙 무결성으로
// 올려 엔진을 세운다(그리고 손절을 놓는 loop 까지 함께 끈다).
//
// 측정 결과 그 팔은 `count=0` 이고, **오늘은 도달할 수 없다.** 이유는 우연한
// 균형이다: 스트림 용량이 2 이고, 잠금은 시장당 최대 한 번이며(잠긴 시장은
// `evaluationState` 가 거부한다), 시장은 정확히 둘이다. 2 = 2 라서 넘칠 수 없다.
//
// "도달할 수 없으니 괜찮다"로 끝내면 안 되는 이유가 5.1.2 다. 그 태스크는 시장
// 둘을 lane 여덟으로 바꾼다. 같은 용량 2 를 그대로 두면 **세 번째 lane 이
// 잠기는 순간 엔진이 서고 safety loop 가 함께 죽는다** — 그것도 "lane 고장은
// 국소로 남는다"를 지키려던 코드에서.
//
// 그래서 이 시험은 균형을 말이 아니라 값으로 못 박는다: fault 스트림의 용량은
// 잠길 수 있는 worker 수와 같아야 한다. 이것은 오늘의 동작을 바꾸지 않는다
// (둘 다 2 다). 바꾸는 것은 **누군가 worker 를 늘렸을 때 무슨 일이 일어나는가**다.
func TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch(t *testing.T) {
	supervisor := newFaultStreamSupervisor(t, func(context.Context) error { return nil })
	if got, want := cap(supervisor.Faults()), len(supervisor.workers); got != want {
		t.Fatalf("fault stream capacity=%d, workers=%d\n\n"+
			"둘이 같아야 한다. 작으면 마지막 worker 의 잠금이 fault 를 건네지 못하고,"+
			" latchMarket 의 default 팔(961:2)이 그 잠금을 오류로 되돌린다. runMarket 은"+
			" 그것을 중앙 무결성으로 올리고 Runtime 은 모든 loop 를 취소한다 —"+
			" fill/exit/reconcile 을 포함해서. 즉 **진입 쪽 고장 하나가 손절을 끈다.**"+
			" 5.1.2 가 시장 둘을 lane 여덟으로 바꿀 때 이 줄을 함께 옮기지 않으면"+
			" 세 번째 잠금에서 정확히 그 일이 일어난다.", got, want)
	}
}

// TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining 은 위 등식이 실제로
// 무엇을 사는지 보여 준다: **아무도 읽지 않아도** 모든 worker 가 자기 잠금을
// 건넬 수 있다. 생산에는 `Faults()` 소비자가 없으므로(이 저장소 전체에서 읽는
// 것은 시험뿐이다) 이것이 실제 운영 조건이다.
func TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining(t *testing.T) {
	supervisor := newFaultStreamSupervisor(t, func(context.Context) error {
		return errors.New("evaluation failed")
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	select {
	case <-supervisor.Ready():
	case <-time.After(time.Second):
		t.Fatal("strategy supervisor readiness")
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		if result := supervisor.Trigger(market); result != StrategyTriggerEnqueued {
			t.Fatalf("%s trigger=%s", market, result)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(supervisor.Faults()) < len(supervisor.workers) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := len(supervisor.Faults()); got != len(supervisor.workers) {
		t.Fatalf("buffered faults=%d, want %d", got, len(supervisor.workers))
	}
	select {
	case err := <-done:
		t.Fatalf("모든 시장이 잠겼는데 감독자가 섰다: %v — 잠금은 진입만 닫아야 한다", err)
	default:
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func newFaultStreamSupervisor(t *testing.T, cycle StrategyCycle) *StrategyEntrySupervisor {
	t.Helper()
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	worker := func(market StrategyMarket) StrategyMarketWorker {
		return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
			AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			LatchRevision:      1}
	}
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{
		Clock: clock.NewFake(base), CycleLimit: MaximumStrategyCycleLimit,
		Workers: []StrategyMarketWorker{worker(StrategyMarketKR), worker(StrategyMarketUS)},
	})
	if err != nil {
		t.Fatalf("NewStrategyEntrySupervisor: %v", err)
	}
	return supervisor
}
