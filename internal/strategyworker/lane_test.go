package strategyworker

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

var laneNow = time.Date(2026, 9, 2, 1, 2, 3, 0, time.UTC)

// threeStrikes 는 임계값만 3 으로 바꾼 정책이다.
//
// 생산 임계값은 1 이라 카운터가 0 아니면 1 밖에 안 된다. 그 상태로는 "센다"와
// "첫 실패에 잠근다"를 구별할 수 없으므로, 카운터가 실제로 세는지는 임계값이
// 1 보다 큰 정책으로만 볼 수 있다.
func threeStrikes() RuntimePolicy {
	policy := ProductionRuntimePolicy()
	policy.failureThreshold = 3
	return policy
}

func laneUnder(t *testing.T, policy RuntimePolicy) *Lane {
	t.Helper()
	lanes := ProductionLanes()
	if len(lanes) == 0 {
		t.Fatal("no production lane was built")
	}
	return newLane(lanes[0].worker, policy)
}

// 갓 태어난 레인은 건강하고 잠기지 않았다.
func TestEveryProductionLaneIsBornHealthyAndUnlatched(t *testing.T) {
	lanes := ProductionLanes()
	if len(lanes) != len(ProductionWorkers()) {
		t.Fatalf("there are %d workers but %d lanes", len(ProductionWorkers()), len(lanes))
	}
	for index, lane := range lanes {
		if lane.Health() != LaneHealthy {
			t.Errorf("lane %d is not healthy at birth: %s", index, lane.Health())
		}
		if lane.Latched() || lane.ConsecutiveFailures() != 0 || lane.LatchRevision() != 0 {
			t.Errorf("lane %d is born with fault state: latched=%v failures=%d revision=%d",
				index, lane.Latched(), lane.ConsecutiveFailures(), lane.LatchRevision())
		}
		if !lane.RestartNotBefore().IsZero() {
			t.Errorf("lane %d is born waiting until %v", index, lane.RestartNotBefore())
		}
	}
}

// 보통 실패는 세기만 하고, 임계값에 닿는 그 실패가 잠근다.
func TestAnOrdinaryFailureCountsAndLatchesExactlyAtTheThreshold(t *testing.T) {
	lane := laneUnder(t, threeStrikes())

	for attempt := uint64(1); attempt < lane.Policy().FailureThreshold(); attempt++ {
		fault, latched := lane.Fail(laneNow, "evidence refresh failed", false)
		if latched {
			t.Fatalf("failure %d latched before the threshold %d", attempt, lane.Policy().FailureThreshold())
		}
		if fault != (Fault{}) {
			t.Errorf("failure %d handed out a fault record without latching: %+v", attempt, fault)
		}
		if lane.ConsecutiveFailures() != attempt {
			t.Errorf("after failure %d the counter reads %d", attempt, lane.ConsecutiveFailures())
		}
		if lane.Health() != LaneDegraded {
			t.Errorf("after failure %d the lane is %s, want %s", attempt, lane.Health(), LaneDegraded)
		}
		if lane.Latched() {
			t.Fatalf("after failure %d the lane is latched", attempt)
		}
	}

	fault, latched := lane.Fail(laneNow, "evidence refresh failed again", false)
	if !latched {
		t.Fatalf("failure %d did not latch at threshold %d", lane.Policy().FailureThreshold(), lane.Policy().FailureThreshold())
	}
	if !lane.Latched() || lane.Health() != LaneLatched {
		t.Errorf("the lane did not latch: latched=%v health=%s", lane.Latched(), lane.Health())
	}
	if fault.Key != lane.Key() {
		t.Errorf("the fault names another lane: %v", fault.Key.Parts())
	}
	if fault.LatchRevision != 1 || lane.LatchRevision() != 1 {
		t.Errorf("first latch revision: fault=%d lane=%d", fault.LatchRevision, lane.LatchRevision())
	}
	if fault.LatchID == "" || fault.Reason == "" || fault.ObservedAt.IsZero() {
		t.Errorf("the fault record is incomplete: %+v", fault)
	}
	if fault.Abnormal {
		t.Error("an ordinary failure was recorded as abnormal")
	}
}

// 비정상 종료는 임계값을 기다리지 않는다.
//
// 설계의 고장표가 둘을 나눈 이유다: 마감 초과나 보통 오류는 세고 다시 시도하지만,
// panic 과 예상 밖 반환은 그 레인의 entry 를 그 자리에서 잠근다.
func TestAnAbnormalFailureLatchesAtOnceBelowTheThreshold(t *testing.T) {
	lane := laneUnder(t, threeStrikes())
	if lane.Policy().FailureThreshold() < 2 {
		t.Fatal("this test needs a threshold above one to say anything")
	}

	fault, latched := lane.Fail(laneNow, "strategy evaluation panic", true)
	if !latched {
		t.Fatalf("an abnormal failure did not latch at count 1 of %d", lane.Policy().FailureThreshold())
	}
	if !fault.Abnormal {
		t.Error("the fault record does not say it was abnormal")
	}
	if lane.Health() != LaneLatched {
		t.Errorf("health after an abnormal failure: %s", lane.Health())
	}
}

// 성공은 카운터를 지우지만 latch 는 절대 못 푼다.
//
// entry-only latch 는 복구 증거가 있어야 풀린다. 성공 한 번으로 풀리면 그
// 실패를 만든 조건이 그대로인데 entry 만 다시 열린다.
func TestSuccessClearsTheCounterButNeverTheLatch(t *testing.T) {
	lane := laneUnder(t, threeStrikes())

	lane.Fail(laneNow, "first", false)
	lane.Fail(laneNow, "second", false)
	if lane.ConsecutiveFailures() != 2 {
		t.Fatalf("counter before the success: %d", lane.ConsecutiveFailures())
	}
	lane.Succeed()
	if lane.ConsecutiveFailures() != 0 || lane.Health() != LaneHealthy {
		t.Errorf("a success did not clear the counter: failures=%d health=%s",
			lane.ConsecutiveFailures(), lane.Health())
	}

	for attempt := uint64(0); attempt < lane.Policy().FailureThreshold(); attempt++ {
		lane.Fail(laneNow, "again", false)
	}
	if !lane.Latched() {
		t.Fatal("the lane did not latch after a full run of failures following the success")
	}
	lane.Succeed()
	if !lane.Latched() || lane.Health() != LaneLatched {
		t.Fatalf("a success released the entry latch: latched=%v health=%s", lane.Latched(), lane.Health())
	}
}

// 두 번째 실패가 두 번째 latch 를 만들지는 않는다.
//
// latch 판을 실패마다 새로 찍으면 운영자가 보는 것은 "첫 원인"이 아니라
// "마지막 원인"이 되고, 복구 증거가 무엇을 되돌리는지도 흐려진다.
func TestASecondFailureDoesNotMintASecondLatch(t *testing.T) {
	lane := laneUnder(t, ProductionRuntimePolicy())
	if lane.Policy().FailureThreshold() != 1 {
		t.Fatalf("this test assumes the production threshold of one, got %d", lane.Policy().FailureThreshold())
	}

	first, latched := lane.Fail(laneNow, "the first reason", false)
	if !latched {
		t.Fatal("the production threshold of one did not latch on the first failure")
	}
	second, latchedAgain := lane.Fail(laneNow, "the second reason", true)
	if latchedAgain {
		t.Error("a second failure reported a second latch")
	}
	if second != (Fault{}) {
		t.Errorf("a second failure handed out another fault record: %+v", second)
	}
	if lane.LatchRevision() != first.LatchRevision {
		t.Errorf("the latch revision moved: %d -> %d", first.LatchRevision, lane.LatchRevision())
	}
	if lane.FirstFailure() != "the first reason" {
		t.Errorf("the first reason was overwritten: %q", lane.FirstFailure())
	}
	// 재시도 지연은 계속 자란다 — 그것이 bounded backoff 다.
	if !lane.RestartNotBefore().After(first.RestartNotBefore) {
		t.Errorf("the restart delay did not grow: %v -> %v", first.RestartNotBefore, lane.RestartNotBefore())
	}
}

// 재시도 지연은 정책의 사다리를 그대로 따른다.
func TestTheRestartDelayFollowsThePolicyLadder(t *testing.T) {
	lane := laneUnder(t, threeStrikes())
	policy := lane.Policy()

	for attempt := uint64(1); attempt <= 8; attempt++ {
		lane.Fail(laneNow, "again", false)
		want := laneNow.Add(policy.Backoff(attempt))
		if got := lane.RestartNotBefore(); !got.Equal(want) {
			t.Errorf("attempt %d: restart not before %v, want %v", attempt, got, want)
		}
	}
}

// 한 레인의 고장은 나머지 일곱을 건드리지 않는다.
//
// 골든이 `peer_lane_state_mutation_forbidden` 으로 적어 둔 것이고, 이 change 의
// 이유 그 자체다 — 오늘은 한 시장이 잠기면 그 시장의 네 전략이 함께 멈춘다.
func TestAFaultOnOneLaneChangesNothingOnItsPeers(t *testing.T) {
	lanes := ProductionLanes()
	if len(lanes) < 2 {
		t.Fatalf("peer isolation needs more than one lane, got %d", len(lanes))
	}
	if _, latched := lanes[0].Fail(laneNow, "this lane only", true); !latched {
		t.Fatal("the first lane did not latch, so this test proves nothing about its peers")
	}
	for index, peer := range lanes[1:] {
		if peer.Latched() || peer.ConsecutiveFailures() != 0 || peer.Health() != LaneHealthy {
			t.Errorf("peer %d moved with its neighbour: latched=%v failures=%d health=%s",
				index+1, peer.Latched(), peer.ConsecutiveFailures(), peer.Health())
		}
	}
}

// 생산 진입점은 부를 때마다 새 레인을 준다.
//
// 패키지 수준에 여덟을 한 벌 만들어 두고 그 포인터를 나눠 주면, 한 번 잠긴
// 레인이 프로세스가 사는 내내 잠긴 채로 모든 호출자에게 건네진다.
func TestProductionLanesHandsOutFreshLanesEveryTime(t *testing.T) {
	first := ProductionLanes()
	if _, latched := first[0].Fail(laneNow, "latched in this call", true); !latched {
		t.Fatal("the first lane did not latch, so a shared lane would not be visible")
	}
	for index, lane := range ProductionLanes() {
		if lane.Latched() || lane.ConsecutiveFailures() != 0 {
			t.Fatalf("lane %d came back carrying an earlier call's fault", index)
		}
	}
}

// 잠긴 레인은 켜져 있어도 봉투를 내지 않는다.
//
// 사이클의 판정 순서가 여기서 정해진다: latch 가 먼저다. 잠긴 레인을 그냥
// DORMANT 로 보고하면 운영자는 "아직 안 켰다"로 읽는데, 실제로는 고장으로
// 닫힌 것이다.
func TestALatchedLaneReportsTheLatchRatherThanDormancy(t *testing.T) {
	lane := laneUnder(t, ProductionRuntimePolicy())
	if lane.Run(Input{}).Outcome != OutcomeDormant {
		t.Fatal("a dormant lane must report dormancy before it is latched")
	}
	if _, latched := lane.Fail(laneNow, "boom", true); !latched {
		t.Fatal("the lane did not latch")
	}
	cycle := lane.Run(Input{})
	if cycle.Outcome != OutcomeLatched {
		t.Fatalf("a latched lane reported %s", cycle.Outcome)
	}
	if cycle.Detail != DetailLatched {
		t.Errorf("detail: want %q, got %q", DetailLatched, cycle.Detail)
	}
	if !emptyEnvelopeValue(cycle.Envelope) {
		t.Error("a latched cycle carried an envelope")
	}
}

// 이유 없는 실패도 이유를 남긴다.
//
// 빈 이유를 그대로 저장하면 운영자가 보는 것은 "고장났다"뿐이고, 그것으로는
// 복구 증거를 만들 수 없다.
func TestAFailureWithoutAReasonStillRecordsOne(t *testing.T) {
	lane := laneUnder(t, ProductionRuntimePolicy())
	fault, latched := lane.Fail(laneNow, "   ", false)
	if !latched {
		t.Fatal("the lane did not latch")
	}
	if fault.Reason == "" || lane.FirstFailure() == "" {
		t.Errorf("a blank reason was stored as blank: %q", fault.Reason)
	}
}

// emptyEnvelopeValue 는 봉투가 비었는지 본다. cycle_test.go 의 같은 검사는
// seam 태그 아래에 있어 이 파일에서 쓸 수 없다 — 이 파일의 시험은 기본 빌드에서
// 돌아야 한다. 레인의 고장 상태 기계는 태그를 켜야만 검사되면 안 된다.
func emptyEnvelopeValue(envelope strategycoordinator.Envelope) bool {
	return envelope.Scope == (strategyrouter.OwnerKey{}) && envelope.SnapshotDigest == "" &&
		envelope.Proposal.Result.Lineage.Identity == ""
}
