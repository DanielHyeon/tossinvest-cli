package strategyworker

import "time"

// RuntimePolicyVersion 는 이 런타임 정책 매니페스트의 판본이다.
//
// 골든은 값들이 "versioned runtime manifest" 에 있어야 한다고만 적었고 판본
// 문자열은 정하지 않았다. 그래서 이 이름은 읽어 온 값이 아니라 이 매니페스트가
// 스스로 붙인 첫 판본이다. 값이 하나라도 바뀌면 이 문자열도 함께 올린다 —
// 그래야 관측 기록에서 "어느 판본의 값으로 돈 레인인가"를 나중에 가를 수 있다.
const RuntimePolicyVersion = "worker-runtime-policy:v1"

// RuntimePolicy 는 레인 하나가 자기 것으로 갖는 런타임 정책이다.
//
// 필드를 내보내지 않는 이유는 골든의 `all_values` 가 "server-owned" 이기
// 때문이다. 내보내면 이 패키지 밖 아무나 자기 값을 담은 정책을 만들어 레인에
// 꽂을 수 있고, 그 순간 "서버가 정한다"는 문장이 거짓이 된다.
//
// 값들은 고른 수가 아니라 오늘 생산이 쓰는 수다. 어디서 읽었는지는 아래
// ProductionRuntimePolicy 에 적었고, 두 곳이 갈라지지 않는지는
// policy_receipt_test.go 가 엔진 소스를 직접 파싱해 대조한다.
type RuntimePolicy struct {
	version          string
	cadence          time.Duration
	queueDepth       int
	cycleDeadline    time.Duration
	failureThreshold uint64
	backoffStep      time.Duration
	backoffCeiling   time.Duration
}

// Version 는 이 정책의 판본이다.
func (policy RuntimePolicy) Version() string { return policy.version }

// Cadence 는 이 레인이 다음 사이클을 여는 간격이다.
func (policy RuntimePolicy) Cadence() time.Duration { return policy.cadence }

// QueueDepth 는 이 레인이 받아 두는 스냅샷 칸 수다.
func (policy RuntimePolicy) QueueDepth() int { return policy.queueDepth }

// CycleDeadline 는 사이클 한 번이 넘으면 안 되는 시간이다.
func (policy RuntimePolicy) CycleDeadline() time.Duration { return policy.cycleDeadline }

// FailureThreshold 는 연속 실패가 몇 번이면 entry 를 잠그는지다.
func (policy RuntimePolicy) FailureThreshold() uint64 { return policy.failureThreshold }

// BackoffStep 는 재시도 지연의 한 계단이다.
func (policy RuntimePolicy) BackoffStep() time.Duration { return policy.backoffStep }

// BackoffCeiling 는 재시도 지연의 천장이다.
func (policy RuntimePolicy) BackoffCeiling() time.Duration { return policy.backoffCeiling }

// Backoff 는 n 번째 실패 뒤 기다릴 시간이다.
//
// 계산은 엔진의 `strategyRestartBackoff` 와 같은 사다리다. 두 사다리가 갈라지지
// 않는지는 policy_receipt_test.go 가 엔진 함수의 본문을 구조로 대조한다.
// 0 번째는 0 이다 — 실패가 없으면 기다림도 없다.
func (policy RuntimePolicy) Backoff(attempt uint64) time.Duration {
	steps := uint64(policy.backoffCeiling / policy.backoffStep)
	if attempt >= steps {
		return policy.backoffCeiling
	}
	return time.Duration(attempt) * policy.backoffStep
}

// ProductionRuntimePolicy 는 여덟 레인이 모두 드는 서버 소유 정책이다.
//
// 값의 영수증(전부 internal/app/engine/strategy_entry_supervisor.go):
//
//   - cadence — 생산 worker 서술자 셋이 모두 `PollInterval: DefaultStrategyCycleLimit`
//     로 대입한다(`:346`, `:391`, `:419`). 영수증은 상수 이름이 아니라 이 대입이다.
//   - queue depth — 생산은 `QueueDepth` 를 한 번도 지정하지 않으므로 생성자가
//     `DefaultStrategyQueueDepth` 로 채운다(`:510`–`:512`).
//   - cycle deadline — 생산 supervisor 둘이 `CycleLimit: MaximumStrategyCycleLimit`
//     으로 짓는다(`:318`, `:352`).
//   - backoff step/ceiling — `DefaultStrategyRestartStep`, `MaximumStrategyRestartBackoff`.
//   - failure threshold — 오늘 엔진은 실패를 세지 않고 첫 실패에 잠근다. 1 보다
//     크게 잡으면 오늘이라면 잠겼을 레인이 한 번 더 진입을 시도하게 되므로,
//     그 완화는 서명된 매니페스트가 정할 일이고 여기서는 오늘 동작을 그대로 둔다.
func ProductionRuntimePolicy() RuntimePolicy {
	return RuntimePolicy{
		version:          RuntimePolicyVersion,
		cadence:          5 * time.Second,
		queueDepth:       1,
		cycleDeadline:    30 * time.Second,
		failureThreshold: 1,
		backoffStep:      5 * time.Second,
		backoffCeiling:   30 * time.Second,
	}
}
