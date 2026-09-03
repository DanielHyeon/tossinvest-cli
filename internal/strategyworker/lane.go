package strategyworker

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
)

// LaneHealth 는 레인 하나의 관측 가능한 상태다. 세 값을 따로 두는 이유는
// "아직 실패가 없다"와 "실패가 쌓였지만 아직 열려 있다"가 운영자에게 다른
// 뜻이기 때문이다. 하나로 뭉치면 잠기기 직전의 레인이 건강해 보인다.
type LaneHealth string

const (
	// LaneHealthy 는 연속 실패가 하나도 없다는 뜻이다.
	LaneHealthy LaneHealth = "HEALTHY"
	// LaneDegraded 는 실패가 쌓였지만 아직 임계값에 못 닿았다는 뜻이다.
	LaneDegraded LaneHealth = "DEGRADED"
	// LaneLatched 는 이 레인의 신규 진입이 잠겼다는 뜻이다.
	LaneLatched LaneHealth = "LATCHED"
)

// DetailLatched 는 잠긴 레인이 내는 진단이다. 계약이 아니라 진단이다.
const DetailLatched = "this lane's entry is latched OFF and needs recovery evidence"

// DetailUnexplainedFailure 는 이유 없이 들어온 실패에 붙이는 이유다.
const DetailUnexplainedFailure = "the lane failed without a reason"

// maxFailureCount 는 카운터가 한 바퀴 도는 것을 막는 상한이다. 0 으로 넘어가면
// 잠긴 레인이 건강해 보인다.
const maxFailureCount = ^uint64(0)

// Fault 는 레인 하나가 신규 진입을 잠글 때 남기는 읽기 전용 기록이다.
//
// 이것은 복구 영수증이 **아니다**. 이 값에는 entry 를 다시 여는 방법이 없고,
// 그 방법을 이 패키지가 갖는 순간 "복구는 증거가 있어야 한다"가 거짓이 된다.
type Fault struct {
	Key              Key
	LatchID          string
	LatchRevision    uint64
	Reason           string
	Abnormal         bool
	ObservedAt       time.Time
	RestartAttempt   uint64
	RestartNotBefore time.Time
}

// Lane 은 worker 하나의 레인-로컬 고장 상태 기계다.
//
// 왜 worker 와 따로 있나. `FamilyWorker` 는 값이고 아무것도 기억하지 않는다 —
// 같은 입력에 언제나 같은 답을 낸다. 연속 실패와 latch 는 기억이라 값으로는
// 담을 수 없다. 둘을 한 타입에 합치면 worker 를 복사하는 모든 자리가 조용히
// 고장 상태까지 복사하게 된다.
//
// 상태는 이 레인 안에만 있다. 이웃 레인을 가리키는 필드가 하나도 없으므로
// 골든의 `peer_lane_state_mutation_forbidden` 은 이 타입에서 구조적으로 참이고,
// lane_test.go 가 그것을 실제로 재어 본다.
type Lane struct {
	worker FamilyWorker
	policy RuntimePolicy
	// clk 는 이 레인의 **유일한** 시간 출처다. 예전 판본은 `now` 를 인자로 받았고
	// 그 편이 능력을 안 늘려서 좋았지만, 마감 시한 감시견은 읽는 것이 아니라
	// **자야** 하므로 인자로는 만들 수 없다. 출처가 둘이면(호출자의 now 와 레인의
	// 시계) 같은 사이클의 backoff 기한과 마감 시한이 서로 다른 시간축에 놓인다.
	clk clock.Clock
	// mu 는 아래 전부를 지킨다. 레인은 이제 사이클 goroutine 과 투입하는 쪽이
	// 함께 만지므로, 상태를 잠금 없이 두면 "단일 비행" 자체가 경합이 된다.
	mu                  sync.Mutex
	consecutiveFailures uint64
	latched             bool
	latchRevision       uint64
	firstFailure        string
	firstAbnormal       bool
	restartAttempt      uint64
	restartNotBefore    time.Time
	pending             int
	dropped             uint64
	abandonedCycles     uint64
	inFlight            bool
	nextDue             time.Time
}

// newLane 은 정책까지 지정해 레인 하나를 만든다.
//
// 내보내지 않는다. 정책을 아무나 고를 수 있으면 "server-owned" 가 거짓이 된다.
// 이 패키지의 시험만 다른 임계값을 가진 레인을 만들 수 있고, 생산 진입점은
// 아래 ProductionLanes 하나뿐이다.
func newLane(worker FamilyWorker, policy RuntimePolicy, clk clock.Clock) *Lane {
	return &Lane{worker: worker, policy: policy, clk: clk}
}

// ProductionLanes 는 생산 worker 여덟에 서버 정책을 붙인 레인 여덟이며, 전부
// **열린 채로** 태어난다.
//
// 부를 때마다 새로 만든다. 패키지 수준에 한 벌을 두고 나눠 주면 한 번 잠긴
// 레인이 프로세스가 사는 내내 모든 호출자에게 잠긴 채로 건네진다.
//
// 생산 배선은 태스크 5.1.2.1 에서 섰다: `internal/app/engine` 의
// `newStrategyLaneRuntime` 이 여덟을 세우고 시장 주기가 자기 넷을 돌린다.
// 여덟은 전부 DORMANT 이므로 사이클은 아무것도 보지 않는다.
//
// **durable 기록에서 태어나야 하는 경우는 `ProductionLanesFrom` 이다**(5.3.3).
// 이 함수는 그것의 "기록 없음" 형태다 — 기록이 없으면 열린 채로 태어나는 것이
// 맞고, 그 판정은 기록이 사는 곳이 한다.
func ProductionLanes(clk clock.Clock) []*Lane {
	workers := ProductionWorkers()
	lanes := make([]*Lane, 0, len(workers))
	policy := ProductionRuntimePolicy()
	for _, worker := range workers {
		lanes = append(lanes, newLane(worker, policy, clk))
	}
	return lanes
}

// Key 는 이 레인이 맡은 worker 의 열쇠다.
func (lane *Lane) Key() Key { return lane.worker.Key() }

// Owns 는 봉인된 제안이 이 레인의 것인지다.
//
// 판정은 여기서 하지 않고 worker 하나에 있다. 배선하는 쪽(엔진)이 "어느 레인의
// 제안인가"를 스스로 다시 적으면 그 사본은 봉인을 먼저 보는 것을 잊기 쉽고,
// 두 판정이 갈리면 잘못된 레인이 남의 제안을 자기 것으로 세운다.
//
// 잠금 상태를 보지 않는 이유: 소유는 고장과 무관한 성질이다. 잠긴 레인도
// 자기 제안을 알아봐야 그 제안이 "주인 없는 것"으로 새지 않는다.
func (lane *Lane) Owns(proposal strategyarbiter.Proposal) bool {
	return lane.worker.owns(proposal)
}

// Policy 는 이 레인이 든 서버 소유 정책이다.
func (lane *Lane) Policy() RuntimePolicy { return lane.policy }

// ConsecutiveFailures 는 마지막 성공 이후 쌓인 실패 수다.
func (lane *Lane) ConsecutiveFailures() uint64 {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.consecutiveFailures
}

// Latched 는 이 레인의 신규 진입이 잠겼는지다.
func (lane *Lane) Latched() bool {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.latched
}

// LatchRevision 는 이 레인이 잠긴 횟수다. 한 번 잠기면 더 오르지 않는다.
func (lane *Lane) LatchRevision() uint64 {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.latchRevision
}

// FirstFailure 는 이 레인을 잠근 **첫** 이유다. 뒤따르는 실패가 덮어쓰지 않는다.
func (lane *Lane) FirstFailure() string {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.firstFailure
}

// FirstAbnormal 은 이 레인을 잠근 **첫** 실패가 비정상이었는지다.
//
// 설계의 고장표는 비정상(패닉·예기치 않은 반환·마감 시한)과 보통 오류를 다르게
// 다루므로, durable 기록이 그 구별을 잃으면 다시 세운 레인이 왜 잠겼는지 운영자가
// 알 수 없다. 읽기만 하고 아무것도 바꾸지 않는다.
func (lane *Lane) FirstAbnormal() bool {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.firstAbnormal
}

// RestartNotBefore 는 이 시각 전에는 다시 시도하지 않는다는 뜻이다.
func (lane *Lane) RestartNotBefore() time.Time {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.restartNotBefore
}

// Health 는 운영자가 보는 이 레인의 상태다.
func (lane *Lane) Health() LaneHealth {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	if lane.latched {
		return LaneLatched
	}
	if lane.consecutiveFailures > 0 {
		return LaneDegraded
	}
	return LaneHealthy
}

// Run 은 잠금을 먼저 보고 worker 사이클로 넘긴다.
//
// 잠금이 먼저인 이유: 잠긴 레인을 DORMANT 로 보고하면 운영자는 "아직 안 켰다"로
// 읽는다. 실제로는 고장으로 닫힌 것이고, 그 둘은 필요한 조치가 다르다.
func (lane *Lane) Run(input Input) Cycle {
	lane.mu.Lock()
	latched := lane.latched
	lane.mu.Unlock()
	if latched {
		return Cycle{Outcome: OutcomeLatched, Detail: DetailLatched}
	}
	return lane.worker.Run(input)
}

// Succeed 는 연속 실패 카운터를 지운다.
//
// latch 는 **풀지 않는다.** entry-only latch 는 복구 증거가 있어야 풀리고, 이
// 패키지에는 그 증거를 만들 방법이 없다. 성공 한 번으로 풀면 실패를 만든 조건이
// 그대로인데 진입만 다시 열린다.
func (lane *Lane) Succeed() {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	lane.consecutiveFailures = 0
}

// Fail 은 실패 하나를 기록하고, 필요하면 이 레인의 신규 진입을 잠근다.
//
// 두 번째 반환값은 "**이 호출이** 잠갔는가"다. 이미 잠긴 레인에서는 false 이고
// 고장 기록도 빈 값이다 — 실패마다 새 latch 판을 찍으면 운영자가 보는 것은
// 첫 원인이 아니라 마지막 원인이 된다.
//
// 비정상(panic·예상 밖 반환)은 임계값을 기다리지 않는다. 설계의 고장표가
// 보통 오류(세고 다시 시도)와 비정상(그 자리에서 잠금)을 나눈 그대로다.
func (lane *Lane) Fail(reason string, abnormal bool) (Fault, bool) {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return lane.failLocked(reason, abnormal)
}

// failLocked 는 Fail 의 본문이다. 사이클 정산(settle)이 같은 잠금 안에서
// 불러야 하므로 따로 두었다 — 잠금을 놓았다 다시 잡으면 그 틈에 다른 goroutine
// 이 곧 잠길 레인에 사이클을 하나 더 넣을 수 있다.
func (lane *Lane) failLocked(reason string, abnormal bool) (Fault, bool) {
	now := lane.clk.Now()
	if lane.consecutiveFailures < maxFailureCount {
		lane.consecutiveFailures++
	}
	if lane.restartAttempt < maxFailureCount {
		lane.restartAttempt++
	}
	lane.restartNotBefore = now.Add(lane.policy.Backoff(lane.restartAttempt))
	if lane.latched || !(abnormal || lane.consecutiveFailures >= lane.policy.failureThreshold) {
		return Fault{}, false
	}
	return lane.latch(now, reason, abnormal), true
}

func (lane *Lane) latch(now time.Time, reason string, abnormal bool) Fault {
	if strings.TrimSpace(reason) == "" {
		reason = DetailUnexplainedFailure
	}
	lane.latched = true
	lane.latchRevision++
	lane.firstFailure = reason
	lane.firstAbnormal = abnormal
	key := lane.worker.Key()
	return Fault{
		Key:              key,
		LatchID:          "lane-latch:" + string(key.Market) + ":" + string(key.Family) + ":" + key.LaneID + ":" + key.LaneVersion + ":" + strconv.FormatUint(lane.latchRevision, 10),
		LatchRevision:    lane.latchRevision,
		Reason:           reason,
		Abnormal:         abnormal,
		ObservedAt:       now.UTC(),
		RestartAttempt:   lane.restartAttempt,
		RestartNotBefore: lane.restartNotBefore,
	}
}
