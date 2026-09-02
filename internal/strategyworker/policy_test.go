package strategyworker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 아래 셋은 골든의 세 이름이 **강제되는지**를 실제로 돌려서 본다.
// 값을 들고 있는지 묻는 검사는 아무도 그 값을 쓰지 않아도 초록이다.

// cadenceIsEnforced 는 주기가 지나기 전 두 번째 사이클이 거절되는지 본다.
func cadenceIsEnforced(t *testing.T) bool {
	t.Helper()
	lane, fake := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartAdmitted {
		return false
	}
	admit(t, lane)
	if _, start := lane.RunBounded(context.Background(), Input{}, succeeds()); start != StartTooSoon {
		return false
	}
	fake.Advance(lane.Policy().Cadence())
	_, start := lane.RunBounded(context.Background(), Input{}, succeeds())
	return start == StartAdmitted
}

// theInboundSlotIsBounded 는 정책 깊이를 넘긴 투입이 거절되고 세어지는지 본다.
func theInboundSlotIsBounded(t *testing.T) bool {
	t.Helper()
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	for filled := 0; filled < lane.Policy().QueueDepth(); filled++ {
		if lane.Offer() != TriggerEnqueued {
			return false
		}
	}
	return lane.Offer() == TriggerFull && lane.Dropped() == 1
}

// theDeadlineIsEnforced 는 마감 시한을 넘긴 사이클이 실제로 버려지는지 본다.
func theDeadlineIsEnforced(t *testing.T) bool {
	t.Helper()
	lane, fake := boundedLane(t, ProductionRuntimePolicy())
	admit(t, lane)
	release := make(chan struct{})
	defer close(release)
	done := make(chan BoundedCycle, 1)
	go func() {
		bounded, _ := lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			<-release
			return Cycle{}, nil
		})
		done <- bounded
	}()
	if !fake.WaitForSleepers(1, 2*time.Second) {
		return false
	}
	fake.Advance(lane.Policy().CycleDeadline())
	bounded := <-done
	return bounded.Abandoned && errors.Is(bounded.Err, ErrCycleDeadline)
}

// 골든의 `runtime_policy` 절을 읽는다. golden_contract_test.go 의 goldenFile 은
// descriptors 만 담고 있어 이 절을 못 본다.
type goldenRuntimePolicy struct {
	RuntimePolicy struct {
		EachWorkerOwns           []string `json:"each_worker_owns"`
		AllValues                string   `json:"all_values"`
		PeerLaneStateMutationBan bool     `json:"peer_lane_state_mutation_forbidden"`
	} `json:"runtime_policy"`
}

func readGoldenRuntimePolicy(t *testing.T) goldenRuntimePolicy {
	t.Helper()
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read the frozen golden: %v", err)
	}
	var golden goldenRuntimePolicy
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode the frozen golden: %v", err)
	}
	if len(golden.RuntimePolicy.EachWorkerOwns) == 0 || golden.RuntimePolicy.AllValues == "" {
		t.Fatal("the golden's runtime_policy decoded empty, so every comparison below would pass for the wrong reason")
	}
	return golden
}

// 골든이 "레인마다 자기 것으로 갖는다"고 적은 일곱은 여전히 그 일곱이어야 한다.
//
// 이 시험이 하는 일은 **관측 가능한 것과의 짝을 강제**하는 것이다. 이름이
// 소스 어딘가에 있는지 묻는 것이 아니라(그런 검사는 언제나 뚫린다), 일곱 개
// 각각에 대해 이 패키지가 내놓는 값을 실제로 불러 본다. 골든에 여덟 번째가
// 생기면 아래 표에 짝이 없어 빨개지고, 그때 사람이 그것을 구현해야 한다.
func TestTheGoldenSevenAreEachSomethingThisLaneCanShow(t *testing.T) {
	golden := readGoldenRuntimePolicy(t)
	lane, _ := boundedLane(t, ProductionRuntimePolicy())
	policy := lane.Policy()

	// 값이 관측되는지를 실제로 확인하는 짝. bool 을 돌려주는 함수로 두는 이유는
	// "이름이 있다"가 아니라 "불러서 값이 나온다"를 요구하기 위해서다.
	//
	// 앞의 세 짝은 5.3.2 에서 바뀌었다. 예전에는 `policy.Cadence() > 0` 처럼
	// **들고 있는 수**를 물었는데, 그러면 아무도 그 수를 강제하지 않아도 초록이다.
	// 이제 셋 다 레인을 실제로 돌려서 그 수가 **행동을 바꾸는지**를 묻는다.
	shown := map[string]func() bool{
		"cadence":                     func() bool { return cadenceIsEnforced(t) },
		"bounded_queue":               func() bool { return theInboundSlotIsBounded(t) },
		"monotonic_deadline":          func() bool { return theDeadlineIsEnforced(t) },
		"health":                      func() bool { return lane.Health() == LaneHealthy },
		"consecutive_failure_counter": func() bool { return lane.ConsecutiveFailures() == 0 },
		"entry_only_latch":            func() bool { return !lane.Latched() },
		"versioned_backoff":           func() bool { return policy.Version() != "" && policy.Backoff(1) > 0 },
	}
	if len(shown) != len(golden.RuntimePolicy.EachWorkerOwns) {
		t.Fatalf("the golden names %d things a worker owns, this test pairs %d: %v",
			len(golden.RuntimePolicy.EachWorkerOwns), len(shown), golden.RuntimePolicy.EachWorkerOwns)
	}
	for _, name := range golden.RuntimePolicy.EachWorkerOwns {
		show, paired := shown[name]
		if !paired {
			t.Errorf("the golden says a worker owns %q and nothing in this package shows it", name)
			continue
		}
		if !show() {
			t.Errorf("%q is paired but a freshly born lane does not show it", name)
		}
	}
}

// 정책 값은 전부 양의 유한 값이어야 한다 — 골든 `all_values` 가 그렇게 적었다.
//
// 0 은 특히 위험하다. cadence 0 은 "안 돈다", deadline 0 은 "감시가 없다",
// queue depth 0 은 "아무것도 못 받는다"인데 셋 다 조용하다.
func TestEveryProductionPolicyValueIsPositiveAndFinite(t *testing.T) {
	policy := ProductionRuntimePolicy()
	durations := map[string]time.Duration{
		"cadence":         policy.Cadence(),
		"cycle deadline":  policy.CycleDeadline(),
		"backoff step":    policy.BackoffStep(),
		"backoff ceiling": policy.BackoffCeiling(),
	}
	for name, value := range durations {
		if value <= 0 {
			t.Errorf("%s is not positive: %v", name, value)
		}
		if value > 24*time.Hour {
			t.Errorf("%s is not a finite runtime value: %v", name, value)
		}
	}
	if policy.QueueDepth() < 1 {
		t.Errorf("queue depth is not positive: %d", policy.QueueDepth())
	}
	if policy.FailureThreshold() < 1 {
		t.Errorf("failure threshold is not positive: %d", policy.FailureThreshold())
	}
	if policy.BackoffStep() > policy.BackoffCeiling() {
		t.Errorf("the backoff step %v is larger than its ceiling %v", policy.BackoffStep(), policy.BackoffCeiling())
	}
	if policy.Version() == "" {
		t.Error("the policy carries no version, so `versioned runtime manifest` is not true of it")
	}
}

// 여덟 레인은 같은 판본의 정책을 든다.
//
// 레인마다 다른 값을 들 수 있게 두면 그 순간 "서버가 정한다"가 거짓이 된다 —
// 어느 레인이 어떤 값을 들었는지 사람이 세어 봐야만 알 수 있게 되기 때문이다.
func TestEveryProductionLaneCarriesTheSameServerOwnedPolicy(t *testing.T) {
	want := ProductionRuntimePolicy()
	lanes := ProductionLanes(clock.NewFake(laneNow))
	if len(lanes) != len(ProductionWorkers()) {
		t.Fatalf("there are %d workers but %d lanes", len(ProductionWorkers()), len(lanes))
	}
	for index, lane := range lanes {
		if lane.Policy() != want {
			t.Errorf("lane %d (%v) carries a different policy: %+v", index, lane.Key().Parts(), lane.Policy())
		}
	}
}

// backoff 사다리는 엔진이 오늘 쓰는 것과 같은 계단이어야 한다.
//
// 값은 policy_receipt_test.go 가 엔진 소스에서 읽어 대조하고, 여기서는 그 두
// 값으로 만들어지는 **사다리 모양**을 본다. 0 번째 시도가 0 인 것이 중요하다 —
// 실패가 없으면 기다림도 없다.
func TestTheBackoffLadderSaturatesAtItsCeiling(t *testing.T) {
	policy := ProductionRuntimePolicy()
	step := policy.BackoffStep()
	ceiling := policy.BackoffCeiling()
	steps := uint64(ceiling / step)

	if got := policy.Backoff(0); got != 0 {
		t.Errorf("attempt 0 must not wait, got %v", got)
	}
	for attempt := uint64(1); attempt < steps; attempt++ {
		want := time.Duration(attempt) * step
		if got := policy.Backoff(attempt); got != want {
			t.Errorf("attempt %d: want %v, got %v", attempt, want, got)
		}
	}
	for _, attempt := range []uint64{steps, steps + 1, steps + 100, ^uint64(0)} {
		if got := policy.Backoff(attempt); got != ceiling {
			t.Errorf("attempt %d must saturate at %v, got %v", attempt, ceiling, got)
		}
	}
}

// 천장이 계단의 배수가 **아닐** 때도 천장에서 멈춘다.
//
// 이 시험이 따로 있는 이유를 적어 둔다. 생산 값은 30초 천장에 5초 계단이라
// 정확히 나누어떨어지고, 그러면 마지막 계단(6번째 시도)이 마침 천장과 같은
// 값이라 `attempt >= steps` 를 `attempt > steps` 로 바꿔도 답이 안 변한다.
// 실제로 그 뮤테이션은 위 시험을 통과했다.
//
// 결함이 사는 차원은 시도 횟수의 **크기**가 아니라 천장과 계단의 **가분성**이다.
// 그래서 나누어떨어지지 않는 정책을 하나 심어 둔다. 7초 계단에 30초 천장이면
// 계단 수는 4(정수 나눗셈)이고, 4번째 시도는 28초가 아니라 천장인 30초여야 한다.
// 그 한 칸이 두 비교를 가른다.
//
// 사다리를 계산으로 다시 유도하지 않고 손으로 적은 이유: 계산으로 유도하면
// 이 시험은 구현을 다시 쓴 것이 되어 무엇도 반증하지 못한다.
func TestTheBackoffLadderSaturatesEvenWhenTheCeilingIsNotAMultipleOfTheStep(t *testing.T) {
	policy := ProductionRuntimePolicy()
	policy.backoffStep = 7 * time.Second
	policy.backoffCeiling = 30 * time.Second

	want := []time.Duration{
		0,
		7 * time.Second,
		14 * time.Second,
		21 * time.Second,
		30 * time.Second, // 28초가 아니다 — 계단 수가 4 라 여기서 이미 천장이다.
		30 * time.Second,
		30 * time.Second,
	}
	for attempt, expected := range want {
		if got := policy.Backoff(uint64(attempt)); got != expected {
			t.Errorf("attempt %d: want %v, got %v", attempt, expected, got)
		}
	}
}

// 카운터는 한 바퀴 돌지 않는다.
//
// 되돌아 0 이 되면 잠긴 레인이 건강해 보이고, 재시도 지연도 처음부터 다시
// 시작한다. 엔진도 같은 자리에 같은 상한을 두고 있다.
func TestTheFailureCountersSaturateRatherThanWrap(t *testing.T) {
	lane := newLane(ProductionWorkers()[0], ProductionRuntimePolicy(), clock.NewFake(laneNow))
	lane.consecutiveFailures = maxFailureCount
	lane.restartAttempt = maxFailureCount

	lane.Fail("one more", false)
	if lane.ConsecutiveFailures() != maxFailureCount {
		t.Errorf("the failure counter wrapped to %d", lane.ConsecutiveFailures())
	}
	if lane.restartAttempt != maxFailureCount {
		t.Errorf("the restart attempt counter wrapped to %d", lane.restartAttempt)
	}
}
