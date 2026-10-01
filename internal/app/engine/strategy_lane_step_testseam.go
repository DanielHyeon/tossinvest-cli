//go:build tossos_testseams

package engine

import (
	"sync"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112 7.5 D1 시험 seam — tossos_testseams 빌드에만 있다(생산 바이너리 밖).
//
// 생산 step 은 순수 메모리 평가라 「멈추는 레인」을 만들 입력이 없다. 시험은 런타임 하나에 훅을 걸어 그 런타임의 레인 step 을 바꾼다.
// 훅이 없는 런타임은 생산과 같은 strategyFamilyLaneStep 을 돈다.
var strategyLaneStepHooks sync.Map // *strategyLaneRuntime → func(*strategyworker.Lane, strategyrouter.FamilyActivation) strategyworker.Step

func (runtime *strategyLaneRuntime) laneStepFor(lane *strategyworker.Lane, promotion strategyrouter.FamilyActivation) strategyworker.Step {
	if hook, ok := strategyLaneStepHooks.Load(runtime); ok {
		return hook.(func(*strategyworker.Lane, strategyrouter.FamilyActivation) strategyworker.Step)(lane, promotion)
	}
	return strategyFamilyLaneStep(lane, promotion)
}

// setStrategyLaneStepHookForTest 는 런타임 하나에 레인 step 훅을 건다. 돌려준 함수가 훅을 걷는다(시험 Cleanup 용).
func setStrategyLaneStepHookForTest(runtime *strategyLaneRuntime,
	hook func(*strategyworker.Lane, strategyrouter.FamilyActivation) strategyworker.Step,
) func() {
	strategyLaneStepHooks.Store(runtime, hook)
	return func() { strategyLaneStepHooks.Delete(runtime) }
}
