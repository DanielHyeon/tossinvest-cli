//go:build tossos_testseams

package engine

import (
	"context"
	"sync"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
)

// a112 7.3.1 시험 seam — tossos_testseams 빌드에만 있다(생산 바이너리 밖).
//
// 생산 적재기는 파일만 읽어 「멈추는 단계」 · panic · 적재 오류를 만들 입력이 없다. 시험은 런타임 하나에 훅을 걸어 그 런타임의 shadow 적재를
// 바꾼다. 훅이 없는 런타임은 생산과 같은 적재기를 돈다.
var strategyShadowLoadHooks sync.Map // *strategyLaneRuntime → func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error)

func (runtime *strategyLaneRuntime) loadShadow(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
	if hook, ok := strategyShadowLoadHooks.Load(runtime); ok {
		return hook.(func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error))(ctx, config)
	}
	return strategyshadow.LoadProductionFamilyShadow(ctx, config)
}

// setStrategyShadowLoadHookForTest 는 런타임 하나에 shadow 적재 훅을 건다. 돌려준 함수가 훅을 걷는다(시험 Cleanup 용).
func setStrategyShadowLoadHookForTest(runtime *strategyLaneRuntime,
	hook func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error),
) func() {
	strategyShadowLoadHooks.Store(runtime, hook)
	return func() { strategyShadowLoadHooks.Delete(runtime) }
}
