//go:build !tossos_testseams

package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// laneStepFor 는 레인 안에서 도는 일이다 — 생산 빌드에서는 **언제나** strategyFamilyLaneStep 하나다(a112 7.5 D1).
//
// 이 본문은 한 줄이어야 한다. 레인 안에 무엇이 들어가는지가 이 패키지의 경계(레인 평가에 Journal · Gateway 없음)이고, 그 경계를 지키는
// TestOnlyThePackageLevelStepEverRunsInsideALane 가 이 정의의 본문 · 빌드 태그 · 정의 수를 AST 로 못 박는다. 시험용 seam 은
// strategy_lane_step_testseam.go(tossos_testseams 빌드)에만 있다.
func (runtime *strategyLaneRuntime) laneStepFor(lane *strategyworker.Lane, promotion strategyrouter.FamilyActivation) strategyworker.Step {
	return strategyFamilyLaneStep(lane, promotion)
}
