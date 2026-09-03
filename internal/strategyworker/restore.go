package strategyworker

import (
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 이 파일은 태스크 5.3.3 의 절반이다: **레인이 durable 기록에서 태어난다.**
//
// 나머지 절반(기록을 쓰고, 복구 증거를 판정하는 것)은 이 패키지에 올 수 없다.
// `dependency_closure_test.go` 가 `-deps` 와 `-deps-test` 를 걸어 이 패키지의
// 폐포에 쓸 수 있는 저널도, broker mutator 도, 토글 writer 도 없다는 것을
// 증명한다. durable 기록에는 writer 가 필요하므로, writer 를 여기 두면 그
// 증명이 사라진다.
//
// 그래서 이 패키지가 받는 것은 **값**이다. 값은 능력이 아니다.
//
// 그리고 방향은 한쪽뿐이다. 여기 있는 것은 "잠긴 채로 태어나는" 길이고 "잠긴
// 레인을 여는" 길은 없다. 여는 것은 기록을 지우는 일이며 그 판정(더 큰 서명
// 활성화 세대)은 기록이 사는 곳이 한다.

// LatchRestore 는 프로세스보다 오래 살아남은 잠금 하나다.
//
// `Fault` 와 모양이 비슷하지만 방향이 반대다. `Fault` 는 레인이 **내는** 관측이고
// 이것은 레인이 **받는** 과거다. 한 타입으로 합치면 "레인이 낸 것"을 그대로
// 되먹여 레인을 잠글 수 있게 되고, 그 순간 잠금의 출처가 기록인지 이번 실행인지
// 구별되지 않는다.
type LatchRestore struct {
	Key           Key
	LatchID       string
	LatchRevision uint64
	Reason        string
	Abnormal      bool
	ObservedAt    time.Time
}

func (restore LatchRestore) valid() bool {
	return restore.Key.Market != "" && restore.Key.Family != "" &&
		strings.TrimSpace(restore.Key.LaneID) != "" && strings.TrimSpace(restore.Key.LaneVersion) != "" &&
		restore.LatchRevision > 0 && strings.TrimSpace(restore.Reason) != ""
}

// ProductionLanesFrom 은 durable 기록에서 생산 레인 여덟을 세운다.
//
// 두 번째 반환값은 **어느 레인에도 붙지 않은 기록**이다. 조용히 버리지 않는
// 이유: 붙지 않았다는 것은 그 기록이 가리키는 lane_id/lane_version 이 이 빌드에
// 없다는 뜻이고, 그러면 사람이 잠가 둔 것이 아무 일도 하지 않으면서 기록으로만
// 남는다. 그 상태를 볼 수 있게 하는 것은 부르는 쪽의 몫이다.
func ProductionLanesFrom(clk clock.Clock, restored []LatchRestore) (lanes []*Lane, unmatched []LatchRestore) {
	lanes = ProductionLanes(clk)
	byKey := make(map[Key]*Lane, len(lanes))
	for _, lane := range lanes {
		byKey[lane.Key()] = lane
	}
	for _, restore := range restored {
		lane, known := byKey[restore.Key]
		if !known || !restore.valid() {
			unmatched = append(unmatched, restore)
			continue
		}
		lane.restoreLatch(restore)
	}
	return lanes, unmatched
}

// restoreLatch 는 태어나는 레인 하나에 과거의 잠금을 앉힌다.
//
// 내보내지 않는다. 살아 있는 레인의 잠금을 밖에서 세울 수 있으면, 그 자리는
// 잠금을 **지우는** 자리로도 쓰일 수 있다 — 부호만 바꾸면 되기 때문이다.
// 여기서만 부르고, 부르는 곳은 갓 만든 레인뿐이다.
func (lane *Lane) restoreLatch(restore LatchRestore) {
	lane.mu.Lock()
	defer lane.mu.Unlock()
	lane.latched = true
	lane.latchRevision = restore.LatchRevision
	lane.firstFailure = restore.Reason
	lane.firstAbnormal = restore.Abnormal
	// 연속 실패 계수기는 되살리지 않는다. 그것은 "지금 몇 번 연달아 실패하는
	// 중인가"이고, 다시 선 프로세스에서는 아직 한 번도 안 돌았다. 잠금과 달리
	// 계수기를 되살리면 복구 직후의 첫 실패가 곧바로 임계값을 넘는다.
}
