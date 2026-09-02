package strategyworker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrCycleDeadline 는 사이클 하나가 정책의 마감 시한을 넘겼다는 뜻이다.
//
// 엔진의 `ErrStrategyCycleDeadline` 과 같은 자리의 오류이고, 같은 성질을 갖는다:
// 시장이 아니라 **이 레인만** 잠근다.
var ErrCycleDeadline = errors.New("strategyworker: lane cycle deadline exceeded")

// Trigger 는 사이클 투입 한 번의 결과다.
//
// 세 철자는 지어낸 것이 아니라 엔진의 `StrategyTriggerResult`(`:204`–`:207`)에서
// 읽었다. 네 번째 값 `INVALID_MARKET` 은 여기 없다 — 레인은 시장을 고르지 않고
// 자기 시장 하나만 맡으므로 그 거절이 생길 자리가 없다. 두 곳이 갈라지지
// 않는지는 bounded_receipt_test.go 가 엔진 소스를 직접 파싱해 대조한다.
type Trigger string

const (
	// TriggerEnqueued 는 투입이 빈 칸에 들어갔다는 뜻이다.
	TriggerEnqueued Trigger = "ENQUEUED"
	// TriggerDisabled 는 이 레인이 투입을 받을 상태가 아니라는 뜻이다.
	TriggerDisabled Trigger = "DISABLED"
	// TriggerFull 은 칸이 차서 이 투입을 버렸다는 뜻이다.
	TriggerFull Trigger = "FULL"
)

// Start 는 이 호출이 사이클을 열었는지, 못 열었다면 왜인지다.
//
// 이유를 하나로 뭉치지 않는 이유는 운영자가 할 조치가 전부 다르기 때문이다.
// "아직 주기가 안 됐다"는 기다리면 되고, "잠겼다"는 복구 증거가 필요하며,
// "이미 돌고 있다"는 드라이버가 둘이라는 뜻이다.
type Start string

const (
	// StartAdmitted 는 사이클이 실제로 돌았다는 뜻이다.
	StartAdmitted Start = "ADMITTED"
	// StartLatched 는 이 레인의 신규 진입이 잠겨 있다는 뜻이다.
	StartLatched Start = "LATCHED"
	// StartInFlight 는 이 레인에서 사이클 하나가 이미 돌고 있다는 뜻이다.
	StartInFlight Start = "IN_FLIGHT"
	// StartBackoff 는 재시작 기한이 아직 남았다는 뜻이다.
	StartBackoff Start = "BACKOFF"
	// StartTooSoon 는 카덴스가 아직 안 지났다는 뜻이다.
	StartTooSoon Start = "TOO_SOON"
	// StartNoTrigger 는 처리할 투입이 없다는 뜻이다. 거절이 아니다.
	StartNoTrigger Start = "NO_TRIGGER"
)

// Step 은 레인이 사이클 한 번에 실제로 돌릴 일이다.
//
// 이 패키지는 이 함수가 무엇을 하는지 모른다. 마감 시한과 panic 복구를 씌울
// 대상일 뿐이고, 그 대상이 무엇을 만질 수 있는지는 이 패키지의 import 폐포가
// 아니라 **넘겨주는 쪽**이 정한다. 생산 배선(태스크 5.1.2)이 그 자리다.
//
// **오류를 따로 돌려주는 이유** (5.7 이 찾은 것). 첫 판본은 `Cycle` 만
// 돌려줬고, 그러면 설계 고장표(`design.md:198`)의 "보통 오류 — 세고 다시 시도"
// 줄에 **닿는 길이 없었다**: 실패는 panic 과 마감 시한뿐이고 둘 다 비정상이라
// 임계값을 기다리지 않는다. 엔진의 영수증도 이쪽이다 — `StrategyCycle` 은
// `func(context.Context) error` 다. 거절(`OutcomeRefused`)과 오류는 다른 것이라
// 한 값에 담지 않는다: 거절은 정당한 결과이고 오류는 고장이다.
type Step func(context.Context, Input) (Cycle, error)

// BoundedCycle 은 유계 사이클 한 번의 결과다.
//
// 넷을 따로 두는 이유는 엔진 `invokeBoundedStrategyCycle` 이 정확히 그 넷을
// 돌려주기 때문이다(`err, abnormal, cancelled, abandoned`). 하나로 뭉치면
// "종료라서 멈췄다"와 "마감 시한을 넘겨서 버렸다"가 같은 값이 된다.
type BoundedCycle struct {
	Cycle     Cycle
	Err       error
	Abnormal  bool
	Cancelled bool
	Abandoned bool
	Fault     Fault
	Latched   bool
}

// Offer 는 사이클 투입 하나를 이 레인의 칸에 넣는다. 막지 않는다.
//
// 잠긴 레인이 받아 두지 않는 이유: 받아 두면 복구 직후에 쌓인 투입이 한꺼번에
// 터진다. 엔진도 같은 자리에서 같은 선택을 한다(`enqueueStrategyPoll` 이
// `worker.latched` 를 보고 DISABLED 를 돌려준다).
func (lane *Lane) Offer() Trigger {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	if lane.latched {
		return TriggerDisabled
	}
	if lane.pending >= lane.policy.queueDepth {
		if lane.dropped < maxFailureCount {
			lane.dropped++
		}
		return TriggerFull
	}
	lane.pending++
	return TriggerEnqueued
}

// Pending 은 아직 처리하지 않은 투입 수다.
func (lane *Lane) Pending() int {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.pending
}

// Dropped 는 칸이 차서 버린 투입 수다. 유계 계수기다.
//
// **엔진에는 이 계수기가 없다.** `enqueueStrategyPoll` 은 `StrategyTriggerFull`
// 을 돌려줄 뿐이고 버린 수를 세는 곳이 어디에도 없다. 골든 `queue.overflow` 는
// "typed refusal **and bounded drop counter**" 를 요구하므로 여기서 메운다.
func (lane *Lane) Dropped() uint64 {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.dropped
}

// Abandoned 는 기다리기를 포기한 사이클 수다.
//
// "버렸다"는 "멈췄다"가 아니다. 엔진도 마감 시한을 넘긴 사이클 goroutine 을
// 취소하지 않고 결과만 안 읽는다. 그래서 이 수는 "몇 번 늦었는가"이지
// "몇 번 죽였는가"가 아니다.
func (lane *Lane) Abandoned() uint64 {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.abandonedCycles
}

// NextDue 는 이 레인이 다음 사이클을 열 수 있는 가장 이른 시각이다.
func (lane *Lane) NextDue() time.Time {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.nextDue
}

// RunBounded 는 투입 하나를 단일 비행·카덴스·마감 시한 아래 돌린다.
//
// 관문 순서가 곧 계약이다. 잠금 → 비행 중 → backoff → 카덴스 → 투입 유무.
// 잠금을 맨 앞에 두는 이유는 잠긴 레인이 투입을 **먹어 버리면** 안 되기
// 때문이다. 엔진은 반대로 큐에서 먼저 꺼내고 나중에 버리는데(`:779` → `:800`),
// 엔진의 투입은 빈 신호라 잃을 것이 없고 이쪽은 그렇지 않다.
//
// 카덴스 기준점은 사이클 **시작**이다. 완료 기준으로 재면 느린 사이클 하나가
// 뒤따르는 모든 주기를 밀어낸다. 엔진의 `runStrategyPoller` 도 투입을 시도한
// 직후에 자므로 같은 기준이다.
func (lane *Lane) RunBounded(ctx context.Context, input Input, step Step) (BoundedCycle, Start) {
	if start := lane.admit(); start != StartAdmitted {
		return BoundedCycle{}, start
	}
	return lane.settle(lane.invokeBounded(ctx, input, step)), StartAdmitted
}

func (lane *Lane) admit() Start {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	if lane.latched {
		return StartLatched
	}
	if lane.inFlight {
		return StartInFlight
	}
	now := lane.clk.Now()
	if now.Before(lane.restartNotBefore) {
		return StartBackoff
	}
	if now.Before(lane.nextDue) {
		return StartTooSoon
	}
	if lane.pending == 0 {
		return StartNoTrigger
	}
	lane.pending--
	lane.inFlight = true
	lane.nextDue = now.Add(lane.policy.cadence)
	return StartAdmitted
}

// invokeBounded 는 엔진 `invokeBoundedStrategyCycle`(`:879`–`:899`)과 같은
// 세 갈래 감시견이다. 그 함수를 이 패키지에서 부를 수는 없다 —
// `internal/app/engine` 을 들여오면 원장과 게이트웨이가 이 패키지의 폐포에
// 들어와 dependency_closure_test.go 가 지키는 것이 사라진다. 그래서 옮겨 적었고,
// 두 벌이 갈라지지 않는지는 bounded_receipt_test.go 가 엔진 소스를 파싱해 본다.
func (lane *Lane) invokeBounded(ctx context.Context, input Input, step Step) BoundedCycle {
	result := make(chan BoundedCycle, 1)
	go func() { result <- invokeStep(ctx, input, step) }()
	watchdogCtx, cancelWatchdog := context.WithCancel(ctx)
	defer cancelWatchdog()
	deadline := make(chan error, 1)
	go func() { deadline <- lane.clk.Sleep(watchdogCtx, lane.policy.cycleDeadline) }()
	select {
	case <-ctx.Done():
		return BoundedCycle{Err: ctx.Err(), Cancelled: true, Abandoned: true}
	case outcome := <-result:
		return outcome
	case <-deadline:
		if ctx.Err() != nil {
			return BoundedCycle{Err: ctx.Err(), Cancelled: true, Abandoned: true}
		}
		// 엔진과 같이 abnormal 이다. 설계 고장표(`design.md:198`)는 deadline 을
		// 보통 오류 줄에 두었으므로 두 정본이 갈리지만, 생산 임계값이 1 이라
		// 오늘은 결과가 같고, 갈리는 방향에서는 이쪽이 더 보수적이다.
		return BoundedCycle{Err: ErrCycleDeadline, Abnormal: true, Abandoned: true}
	}
}

// invokeStep 은 엔진 `invokeStrategyCycle`(`:901`)과 같이 panic 을 비정상
// 실패로 바꾼다. 엔진이 거기서 하는 중앙 무결성 오류 되살리기는 여기 없다 —
// 그 판정은 `internal/app/engine` 의 것이고 이 패키지의 폐포 밖이다(태스크 5.6).
func invokeStep(ctx context.Context, input Input, step Step) (outcome BoundedCycle) {
	defer func() {
		if recovered := recover(); recovered != nil {
			outcome.Abnormal = true
			outcome.Err = fmt.Errorf("strategyworker: lane cycle panic: %v", recovered)
		}
	}()
	outcome.Cycle, outcome.Err = step(ctx, input)
	return outcome
}

// settle 은 사이클 하나의 결과를 레인 상태에 반영한다.
//
// 취소는 실패가 아니다. 엔진도 `if cancelled { return }` 을 오류 처리보다 먼저
// 둔다(`:807`). 종료를 실패로 세면 다음 기동이 잠긴 채로 시작한다.
func (lane *Lane) settle(bounded BoundedCycle) BoundedCycle {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	lane.inFlight = false
	if bounded.Abandoned && lane.abandonedCycles < maxFailureCount {
		lane.abandonedCycles++
	}
	if bounded.Cancelled {
		return bounded
	}
	if bounded.Err == nil {
		lane.consecutiveFailures = 0
		return bounded
	}
	bounded.Fault, bounded.Latched = lane.failLocked(bounded.Err.Error(), bounded.Abnormal)
	return bounded
}
