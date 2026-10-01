package strategyworker

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// LaneStatus 는 레인 한 행의 상태를 **한 잠금 안에서** 읽은 값이다(a112 7.5 D2).
//
// 개별 접근자(Health · LatchRevision · FirstFailure …)는 각자 잠금을 잡으므로, 그 사이에 레인이 잠기면 「LATCHED 인데 revision 0」 같은
// 찢긴 행이 나온다. 투영처럼 한 행을 보여 주는 쪽은 이것 하나만 읽어야 한다(엔진 시험이 AST 로 못 박음). 값이고 아무것도 바꾸지 않는다.
type LaneStatus struct {
	Health              LaneHealth
	ConsecutiveFailures uint64
	LatchRevision       uint64
	FirstFailure        string
	Pending             int
	Dropped             uint64
	Abandoned           uint64
	NextDue             time.Time
	RestartNotBefore    time.Time
}

// Status 는 레인 한 행을 한 잠금으로 읽는다.
func (lane *Lane) Status() LaneStatus {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	return LaneStatus{Health: lane.healthLocked(), ConsecutiveFailures: lane.consecutiveFailures, LatchRevision: lane.latchRevision,
		FirstFailure: lane.firstFailure, Pending: lane.pending, Dropped: lane.dropped, Abandoned: lane.abandonedCycles,
		NextDue: lane.nextDue, RestartNotBefore: lane.restartNotBefore}
}

// a112 태스크 7.3 — 투영이 레인 하나의 worker 서술을 읽는 접근자(Manager 판정: additive 읽기 접근자).
//
// 레인은 worker 를 숨긴다(worker 를 복사하는 자리가 고장 상태까지 복사하지 않도록 — Lane 주석). 투영은 그 worker 의 horizon · runtime 과
// 「이 활성화가 이 레인에 대해 말한 desired/effective」만 필요하므로 그 넷을 위임으로 연다. 아무것도 바꾸지 않고 잠금도 잡지 않는다 —
// worker 는 값이고 레인 수명 동안 바뀌지 않는다.

// Horizon 은 이 레인 worker 의 보유 기간이다.
func (lane *Lane) Horizon() strategyrouter.Horizon { return lane.worker.Horizon() }

// Runtime 은 이 레인 worker 의 runtime 상태다(현재 언제나 UNOBSERVED — 판정 Q3).
func (lane *Lane) Runtime() strategyrouter.RuntimeState { return lane.worker.Runtime() }

// Desired 와 Effective 는 활성화가 이 레인 **자기 열쇠**에 대해 말한 상태다(호출자가 시장 · 가족을 고르지 않음).
func (lane *Lane) Desired(activation strategyrouter.FamilyActivation) strategyrouter.DesiredState {
	return lane.worker.Desired(activation)
}

func (lane *Lane) Effective(activation strategyrouter.FamilyActivation) strategyrouter.DesiredState {
	return lane.worker.Effective(activation)
}
