package strategyworker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// boundedLane 은 가짜 시계를 든 레인 하나를 만든다.
func boundedLane(t *testing.T, policy RuntimePolicy) (*Lane, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(laneNow)
	lanes := ProductionLanes(fake)
	if len(lanes) == 0 {
		t.Fatal("no production lane was built")
	}
	return newLane(lanes[0].worker, policy, fake), fake
}

// admit 은 투입 한 번을 넣고 그 결과가 받아들여졌는지 확인한다.
func admit(t *testing.T, lane *Lane) {
	t.Helper()
	if got := lane.Offer(); got != TriggerEnqueued {
		t.Fatalf("offer was not enqueued: %s", got)
	}
}

func neverRuns(t *testing.T) Step {
	t.Helper()
	return func(context.Context, Input) (Cycle, error) {
		t.Error("the step ran even though the lane should not have admitted a cycle")
		return Cycle{}, nil
	}
}

func succeeds() Step {
	return func(context.Context, Input) (Cycle, error) { return Cycle{Outcome: OutcomeDormant}, nil }
}

// 투입 칸은 정책이 정한 깊이만큼만 받고, 넘친 것을 **센다**.
//
// 엔진에는 이 계수기가 없다. `enqueueStrategyPoll` 은 `StrategyTriggerFull` 을
// 돌려줄 뿐이고 버린 수를 세는 곳이 없다. 골든 `queue.overflow` 는 "typed refusal
// **and bounded drop counter**" 를 요구하므로 그 빈칸을 레인이 메운다.
func TestTheInboundSlotIsBoundedAndCountsWhatItDropped(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	depth := lane.Policy().QueueDepth()

	for filled := 0; filled < depth; filled++ {
		if got := lane.Offer(); got != TriggerEnqueued {
			t.Fatalf("offer %d of %d was refused: %s", filled+1, depth, got)
		}
	}
	if lane.Dropped() != 0 {
		t.Fatalf("nothing overflowed yet but the drop counter is %d", lane.Dropped())
	}
	if got := lane.Offer(); got != TriggerFull {
		t.Fatalf("the slot is full and the next offer said %s", got)
	}
	if lane.Dropped() != 1 {
		t.Fatalf("one offer was dropped and the counter says %d", lane.Dropped())
	}
	if lane.Pending() != depth {
		t.Fatalf("the slot holds %d, not the policy depth %d", lane.Pending(), depth)
	}
}

// 잠긴 레인은 투입을 아예 받지 않는다.
//
// 엔진도 그렇게 한다 — `enqueueStrategyPoll` 이 `worker.latched` 를 보고
// `StrategyTriggerDisabled` 를 돌려준다. 받아 두고 나중에 버리면 그 사이 쌓인
// 투입이 복구 직후 한꺼번에 터진다.
func TestALatchedLaneRefusesEveryOffer(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	if _, latched := lane.Fail("evidence refresh failed", false); !latched {
		t.Fatal("the production threshold is 1 so the first failure must latch")
	}
	if got := lane.Offer(); got != TriggerDisabled {
		t.Fatalf("a latched lane accepted an offer: %s", got)
	}
	if lane.Pending() != 0 {
		t.Fatalf("a latched lane queued %d triggers", lane.Pending())
	}
}

// 투입이 없으면 사이클도 없다.
func TestALaneWithNoTriggerDoesNotStartACycle(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	got, start := lane.RunBounded(context.Background(), Input{}, neverRuns(t))
	if start != StartNoTrigger {
		t.Fatalf("a lane with an empty slot started a cycle: %s", start)
	}
	if got.Err != nil || got.Abandoned {
		t.Fatalf("a cycle that never ran produced %+v", got)
	}
}

// 단일 비행: 한 사이클이 도는 동안 두 번째는 시작하지 못한다.
//
// 엔진은 이것을 플래그가 아니라 **소비자 goroutine 하나**로 지킨다. 배선으로
// 지켜지는 성질은 드라이버가 둘이 되는 순간 조용히 사라지므로, 레인은 그것을
// 자기 상태로 갖는다.
func TestALaneAdmitsOnlyOneCycleAtATime(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)

	inFlight := make(chan struct{})
	release := make(chan struct{})
	done := make(chan Start, 1)
	go func() {
		_, start := lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			close(inFlight)
			<-release
			return Cycle{Outcome: OutcomeDormant}, nil
		})
		done <- start
	}()
	<-inFlight

	if _, start := lane.RunBounded(context.Background(), Input{}, neverRuns(t)); start != StartInFlight {
		t.Fatalf("a second cycle started while one was in flight: %s", start)
	}
	close(release)
	if start := <-done; start != StartAdmitted {
		t.Fatalf("the first cycle was not admitted: %s", start)
	}
}

// 카덴스: 주기가 지나기 전에는 다음 사이클을 열지 않는다.
func TestALaneRefusesASecondCycleBeforeItsCadenceElapsed(t *testing.T) {
	lane, fake := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatal("the first cycle was refused")
	}
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, neverRuns(t)); start != StartTooSoon {
		t.Fatalf("a second cycle opened inside the cadence: %s", start)
	}
	fake.Advance(lane.Policy().Cadence())
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatalf("the cadence elapsed and the cycle was still refused: %s", start)
	}
}

// 카덴스의 기준점은 사이클 **시작**이지 완료가 아니다.
//
// 영수증은 `runStrategyPoller` (`strategy_entry_supervisor.go:719`) 다 — 투입을
// 시도한 **직후** 자고, 사이클이 끝나기를 기다리지 않는다. 완료 기준으로 재면
// 느린 사이클 하나가 그 뒤의 모든 주기를 밀어낸다.
func TestTheCadenceIsMeasuredFromTheCycleStartNotItsEnd(t *testing.T) {
	policy := ProductionRuntimePolicy()
	lane, fake := boundedLane(t, policy)
	admit(t, lane)

	slow := policy.Cadence() / 2
	if _, start := lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
		fake.Advance(slow)
		return Cycle{Outcome: OutcomeDormant}, nil
	}); start != StartAdmitted {
		t.Fatal("the slow cycle was refused")
	}
	// 시작 기준이면 남은 대기는 cadence - slow 다. 완료 기준이면 cadence 다.
	fake.Advance(policy.Cadence() - slow)
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatalf("the cadence is being measured from the cycle's end, not its start: %s", start)
	}
}

// 마감 시한을 넘긴 사이클은 버려지고 실패로 센다.
//
// 엔진 `invokeBoundedStrategyCycle` (`:879`–`:899`) 의 세 번째 갈래 그대로다:
// `ErrStrategyCycleDeadline`, abnormal=true, abandoned=true.
func TestACycleThatCrossesTheDeadlineIsAbandonedAndCountedAsAFailure(t *testing.T) {
	lane, fake := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)

	release := make(chan struct{})
	defer close(release)
	type result struct {
		bounded BoundedCycle
		start   Start
	}
	done := make(chan result, 1)
	go func() {
		bounded, start := lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			<-release
			return Cycle{Outcome: OutcomeDormant}, nil
		})
		done <- result{bounded, start}
	}()
	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("no watchdog registered a sleep, so this lane has no deadline")
	}
	fake.Advance(lane.Policy().CycleDeadline())

	got := <-done
	if got.start != StartAdmitted {
		t.Fatalf("the cycle was not admitted: %s", got.start)
	}
	if !errors.Is(got.bounded.Err, ErrCycleDeadline) {
		t.Fatalf("the deadline did not produce ErrCycleDeadline: %v", got.bounded.Err)
	}
	if !got.bounded.Abandoned || !got.bounded.Abnormal || got.bounded.Cancelled {
		t.Fatalf("the engine's deadline branch is abandoned+abnormal, this one is %+v", got.bounded)
	}
	if lane.Abandoned() != 1 {
		t.Fatalf("one cycle was abandoned and the counter says %d", lane.Abandoned())
	}
	if !lane.Latched() {
		t.Fatal("the production threshold is 1 so the abandoned cycle must latch the lane")
	}
}

// 버려진 사이클의 goroutine 은 취소되지 않는다 — 늦게 끝나도 아무 일이 없어야 한다.
//
// 엔진이 그렇게 한다(`abandoned` 는 "일을 멈춘다"가 아니라 "기다리기를 멈춘다").
// 늦은 결과가 레인 상태를 뒤늦게 바꾸면, 잠긴 레인이 스스로 풀리는 길이 생긴다.
func TestALateResultFromAnAbandonedCycleChangesNothing(t *testing.T) {
	lane, fake := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)

	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			<-release
			return Cycle{Outcome: OutcomeEmitted}, nil
		})
		close(done)
	}()
	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("no watchdog registered a sleep")
	}
	fake.Advance(lane.Policy().CycleDeadline())
	<-done

	revision, health := lane.LatchRevision(), lane.Health()
	close(release) // 늦은 결과가 이제 도착한다
	time.Sleep(20 * time.Millisecond)
	if lane.LatchRevision() != revision || lane.Health() != health {
		t.Fatalf("a late result changed the lane from (%d, %s) to (%d, %s)",
			revision, health, lane.LatchRevision(), lane.Health())
	}
}

// panic 은 프로세스를 죽이지 않고 비정상 실패가 된다.
func TestAPanickingStepIsAnAbnormalFailureRatherThanACrash(t *testing.T) {
	lane, _ := boundedLane(t, threeStrikes())
	admit(t, lane)
	bounded, start := lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
		panic("evidence replay exploded")
	})
	if start != StartAdmitted {
		t.Fatalf("the cycle was not admitted: %s", start)
	}
	if bounded.Err == nil || !bounded.Abnormal {
		t.Fatalf("a panic did not become an abnormal failure: %+v", bounded)
	}
	if !lane.Latched() {
		t.Fatal("an abnormal failure latches at once even below the threshold")
	}
}

// 취소는 실패가 아니다.
//
// 엔진 `runMarket` 이 `if cancelled { return }` 을 오류 처리보다 **먼저** 둔다
// (`:807`). 종료를 실패로 세면 다음 기동이 잠긴 채로 시작한다.
func TestCancellationIsNotAFailure(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bounded, start := lane.RunBounded(ctx, Input{}, func(context.Context, Input) (Cycle, error) {
		<-make(chan struct{}) // 취소가 이기도록 영원히 막는다
		return Cycle{}, nil
	})
	if start != StartAdmitted {
		t.Fatalf("the cycle was not admitted: %s", start)
	}
	if !bounded.Cancelled {
		t.Fatalf("a cancelled cycle was not reported as cancelled: %+v", bounded)
	}
	if lane.Latched() || lane.ConsecutiveFailures() != 0 {
		t.Fatalf("cancellation counted as a failure: latched=%v failures=%d",
			lane.Latched(), lane.ConsecutiveFailures())
	}
}

// backoff 기한 전에는 사이클을 열지 않는다.
func TestALaneInBackoffWaitsItsRestartDeadline(t *testing.T) {
	lane, fake := boundedLane(t, threeStrikes())
	if _, latched := lane.Fail("evidence refresh failed", false); latched {
		t.Fatal("one failure must not latch under a threshold of three")
	}
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, neverRuns(t)); start != StartBackoff {
		t.Fatalf("a lane inside its backoff opened a cycle: %s", start)
	}
	fake.Advance(lane.Policy().Backoff(1))
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatalf("the backoff elapsed and the cycle was still refused: %s", start)
	}
}

// 성공한 사이클은 카운터를 지운다.
func TestASuccessfulBoundedCycleClearsTheFailureCounter(t *testing.T) {
	lane, fake := boundedLane(t, threeStrikes())
	if _, latched := lane.Fail("evidence refresh failed", false); latched {
		t.Fatal("one failure must not latch under a threshold of three")
	}
	fake.Advance(lane.Policy().Backoff(1))
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatal("the cycle was refused")
	}
	if lane.ConsecutiveFailures() != 0 {
		t.Fatalf("a successful cycle left %d failures", lane.ConsecutiveFailures())
	}
}

// 여덟 레인은 투입 칸과 비행 상태를 서로 나누지 않는다.
func TestEveryProductionLaneKeepsItsOwnSlotAndFlight(t *testing.T) {
	fake := clock.NewFake(laneNow)
	lanes := ProductionLanes(fake)
	admit(t, lanes[0])
	for index, lane := range lanes {
		want := 0
		if index == 0 {
			want = 1
		}
		if lane.Pending() != want {
			t.Errorf("lane %d holds %d triggers, want %d", index, lane.Pending(), want)
		}
	}
}

// 잠긴 레인은 사이클도 열지 않는다.
//
// 투입 거절(위)만으로는 부족하다. 투입은 잠기기 **전에** 들어와 있을 수 있고,
// 그 사이 레인이 잠기면 그 투입은 돌면 안 된다.
func TestALatchedLaneDoesNotOpenACycleItAlreadyHeld(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	if _, latched := lane.Fail("evidence refresh failed", false); !latched {
		t.Fatal("the production threshold is 1 so the first failure must latch")
	}
	if _, start := lane.RunBounded(context.Background(), Input{}, neverRuns(t)); start != StartLatched {
		t.Fatalf("a latched lane opened the cycle it was already holding: %s", start)
	}
	if lane.Pending() != 1 {
		t.Fatalf("the refused cycle consumed the trigger; %d left", lane.Pending())
	}
}

// 사이클 하나는 투입 하나를 쓴다.
func TestAnAdmittedCycleConsumesExactlyOneTrigger(t *testing.T) {
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	if lane.Pending() != 1 {
		t.Fatalf("the slot holds %d after one offer", lane.Pending())
	}
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		t.Fatal("the cycle was refused")
	}
	if lane.Pending() != 0 {
		t.Fatalf("the admitted cycle left %d triggers in the slot", lane.Pending())
	}
}

// 감시견이 자는 시간은 **마감 시한**이지 다른 정책 값이 아니다.
//
// 이 시험이 따로 있는 이유. 위의 마감 시한 시험은 `CycleDeadline()` 만큼 시계를
// 밀고 감시견이 깨어나는지 본다. 그런데 가짜 시계를 30 초 밀면 5 초짜리 잠도
// 함께 깨므로, 감시견이 카덴스만큼 자도 그 시험은 초록이다. 결함이 사는 차원은
// "깨어나는가"가 아니라 "**언제** 깨어나는가"이고, 그것은 마감 시한보다 **짧게**
// 밀어 봐야만 갈린다.
func TestTheWatchdogSleepsForTheDeadlineAndNotSomeOtherPolicyValue(t *testing.T) {
	policy := ProductionRuntimePolicy()
	if policy.Cadence() >= policy.CycleDeadline() {
		t.Fatalf("this test separates two values that are no longer different: cadence %v, deadline %v",
			policy.Cadence(), policy.CycleDeadline())
	}
	lane, fake := boundedLane(t, policy)
	admit(t, lane)
	release := make(chan struct{})
	defer close(release)
	done := make(chan struct{})
	go func() {
		lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			<-release
			return Cycle{}, nil
		})
		close(done)
	}()
	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("no watchdog registered a sleep")
	}

	fake.Advance(policy.Cadence())
	time.Sleep(20 * time.Millisecond)
	select {
	case <-done:
		t.Fatalf("the watchdog woke after the cadence %v, so it is not sleeping for the deadline %v",
			policy.Cadence(), policy.CycleDeadline())
	default:
	}

	fake.Advance(policy.CycleDeadline() - policy.Cadence())
	<-done
	if lane.Abandoned() != 1 {
		t.Fatalf("the deadline elapsed and %d cycles were abandoned", lane.Abandoned())
	}
}

// 보통 오류는 세고 다시 시도하며, 비정상과 다른 값으로 기록된다.
//
// **5.7 이 이 자리를 찾았다.** `Step` 이 `Cycle` 만 돌려주던 판본에서는 설계
// 고장표(`design.md:198`)의 "보통 오류" 줄에 닿는 길이 아예 없었다 — 실패는
// panic 과 마감 시한뿐이고 둘 다 비정상이라 임계값을 기다리지 않는다. 그래서
// 임계값이라는 개념 자체가 생산 배선에서는 죽은 값이었다.
func TestAnOrdinaryStepErrorCountsAndDoesNotLatchBelowTheThreshold(t *testing.T) {
	lane, fake := boundedLane(t, threeStrikes())
	failing := func(context.Context, Input) (Cycle, error) {
		return Cycle{}, errors.New("evidence replay failed")
	}

	for attempt := uint64(1); attempt < lane.Policy().FailureThreshold(); attempt++ {
		admit(t, lane)
		bounded, start := lane.RunBounded(context.Background(), Input{}, failing)
		if start != StartAdmitted {
			t.Fatalf("attempt %d was refused: %s", attempt, start)
		}
		if bounded.Abnormal {
			t.Fatalf("attempt %d was recorded as abnormal", attempt)
		}
		if bounded.Latched {
			t.Fatalf("attempt %d latched before the threshold %d", attempt, lane.Policy().FailureThreshold())
		}
		if lane.ConsecutiveFailures() != attempt {
			t.Fatalf("attempt %d left the counter at %d", attempt, lane.ConsecutiveFailures())
		}
		fake.Advance(lane.Policy().Backoff(attempt) + lane.Policy().Cadence())
	}

	admit(t, lane)
	bounded, _ := lane.RunBounded(context.Background(), Input{}, failing)
	if !bounded.Latched || bounded.Fault.Abnormal {
		t.Fatalf("the threshold failure did not latch as an ordinary fault: %+v", bounded)
	}
	if lane.FirstFailure() != "evidence replay failed" {
		t.Fatalf("the latch kept %q rather than the step's own error", lane.FirstFailure())
	}
}
