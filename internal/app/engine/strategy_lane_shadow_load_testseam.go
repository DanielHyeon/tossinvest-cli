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

// strategyShadowStepContexts 는 런타임별로 단계가 받은 stepCtx 를 모음(0.5 리뷰 시험#3 · #4 — 감독 종료 신호).
// 감독 goroutine 은 `defer cancel()` 로 끝나므로 stepCtx 닫힘 = 감독 종료(게시 · 마감 판정이 끝난 뒤). 생산 파일은 바꾸지 않음.
var strategyShadowStepContexts sync.Map // *strategyLaneRuntime → *strategyShadowStepTrack

type strategyShadowStepTrack struct {
	mu       sync.Mutex
	contexts []context.Context
}

func (runtime *strategyLaneRuntime) loadShadow(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
	value, _ := strategyShadowStepContexts.LoadOrStore(runtime, &strategyShadowStepTrack{})
	track := value.(*strategyShadowStepTrack)
	track.mu.Lock()
	track.contexts = append(track.contexts, ctx)
	track.mu.Unlock()
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

// strategyShadowSupervisionSettledForTest 는 이 런타임에서 적재까지 간 모든 단계의 감독 goroutine 이 끝났는지 답함.
// 단일 비행 표시(shadowInFlight)는 안쪽 goroutine 이 결과를 보낸 직후 내리고 게시는 그 뒤 감독이 하므로, 표시만 보는 대기는
// 게시 전에 돌아올 수 있음 — 그 경합을 닫는 신호.
func strategyShadowSupervisionSettledForTest(runtime *strategyLaneRuntime) bool {
	value, ok := strategyShadowStepContexts.Load(runtime)
	if !ok {
		return true
	}
	track := value.(*strategyShadowStepTrack)
	track.mu.Lock()
	defer track.mu.Unlock()
	for _, ctx := range track.contexts {
		if ctx.Err() == nil {
			return false
		}
	}
	return true
}

// strategyShadowStepsStartedForTest 는 이 런타임에서 적재까지 간 단계 수임(대기 helper 의 하한 단언용).
func strategyShadowStepsStartedForTest(runtime *strategyLaneRuntime) int {
	value, ok := strategyShadowStepContexts.Load(runtime)
	if !ok {
		return 0
	}
	track := value.(*strategyShadowStepTrack)
	track.mu.Lock()
	defer track.mu.Unlock()
	return len(track.contexts)
}
