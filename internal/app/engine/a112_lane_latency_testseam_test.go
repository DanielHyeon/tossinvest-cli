//go:build tossos_testseams

package engine

// a112 태스크 7.5 D1 — 레인 지연 독립(Manager 판정 D1=(A)). 멈추는 레인은 시험 seam(strategy_lane_step_testseam.go)으로만 만들 수 있어
// 이 파일은 tossos_testseams 빌드다. 머리말 · 나머지 운용성 시험은 a112_lane_operability_test.go.

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// D1: KR continuation · reversal 의 step 이 멈춘다(앞의 두 레인 — 순차였다면 weekly · breakout 이 그 둘의 마감 시한을 차례로 기다린다).
func TestAHungLaneDoesNotDelayItsPeersInTheSameWave(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runtime := newStrategyLaneRuntime(clock.System(), nil, "")
		release := make(chan struct{})
		var mu sync.Mutex
		returned := map[strategyworker.Key]time.Duration{}
		start := time.Now()
		hung := func(key strategyworker.Key) bool {
			return key.Market == strategyrouter.MarketKR &&
				(key.Family == strategyrouter.FamilyContinuation || key.Family == strategyrouter.FamilyReversal)
		}
		defer setStrategyLaneStepHookForTest(runtime, func(lane *strategyworker.Lane, activation strategyrouter.FamilyActivation) strategyworker.Step {
			production, key := strategyFamilyLaneStep(lane, activation), lane.Key()
			return func(ctx context.Context, input strategyworker.Input) (strategyworker.Cycle, error) {
				if hung(key) {
					<-release
				}
				cycle, err := production(ctx, input)
				mu.Lock()
				returned[key] = time.Since(start)
				mu.Unlock()
				return cycle, err
			}
		})()
		deadline := strategyworker.ProductionRuntimePolicy().CycleDeadline()
		if err := runtime.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil); err != nil {
			t.Fatal(err)
		}
		// 시장 지연 = 최댓값 = 마감 시한 1 회(조건 ①). 순차였다면 멈춘 레인 둘의 마감 시한이 더해져 2 회다.
		if elapsed := time.Since(start); elapsed != deadline {
			t.Fatalf("market wave took %s with two hung lanes, want exactly one cycle deadline %s (max, not sum)", elapsed, deadline)
		}
		observed := map[strategyworker.Key]strategyLaneObservation{}
		for _, observation := range runtime.observations() {
			observed[observation.Key] = observation
		}
		for _, lane := range runtime.lanesFor(StrategyMarketKR) {
			key, got := lane.Key(), observed[lane.Key()]
			if hung(key) {
				// 멈춘 레인만 자기 실패 계수가 오르고(비정상 → 즉시 잠금) 버려진다.
				if !got.Abandoned || !got.Abnormal || lane.ConsecutiveFailures() != 1 || lane.Health() != strategyworker.LaneLatched {
					t.Fatalf("hung lane %v observation=%+v failures=%d health=%s, want abandoned · abnormal · 1 failure · LATCHED",
						key, got, lane.ConsecutiveFailures(), lane.Health())
				}
				continue
			}
			// 이웃은 시작 즉시(가상 시각 0) 돌아왔다 — 앞의 멈춘 레인을 기다리지 않았다.
			mu.Lock()
			at, ran := returned[key]
			mu.Unlock()
			if !ran || at != 0 || got.Outcome != strategyworker.OutcomeDormant || got.Abandoned || lane.ConsecutiveFailures() != 0 ||
				lane.Health() != strategyworker.LaneHealthy {
				t.Fatalf("peer lane %v returned=%v at %s observation=%+v, want an immediate DORMANT cycle and a healthy lane", key, ran, at, got)
			}
		}
		// 조건 ②: 버려진 레인의 step goroutine 은 step 이 돌아올 때까지 산다 — 여기서 풀어 주면 끝난다(거품이 모두 끝남을 요구한다).
		close(release)
		synctest.Wait()
	})
}

// D1: 레인 goroutine 안의 panic(step 밖 — step 을 만드는 자리)은 삼키지 않고 join 뒤 시장 주기 goroutine 에서 다시 던진다. 순차였을 때와 같이
// 시장 주기의 회복 경로(invokeStrategyCycle)가 받는다. 다시 던지지 않으면 goroutine 의 panic 이 프로세스를 끝낸다.
func TestAPanicOutsideALaneStepStillReachesTheMarketCycle(t *testing.T) {
	runtime := newStrategyLaneRuntime(clock.System(), nil, "")
	defer setStrategyLaneStepHookForTest(runtime, func(lane *strategyworker.Lane, activation strategyrouter.FamilyActivation) strategyworker.Step {
		if lane.Key().Family == strategyrouter.FamilyWeeklyValue {
			panic("a112 lane step factory fault")
		}
		return strategyFamilyLaneStep(lane, activation)
	})()
	defer func() {
		if recovered := recover(); recovered == nil || !strings.Contains(fmt.Sprint(recovered), "a112 lane step factory fault") {
			t.Fatalf("recovered=%v, want the lane goroutine's panic re-raised on the market cycle goroutine", recovered)
		}
	}()
	_ = runtime.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil)
	t.Fatal("evaluate returned without re-raising the lane panic")
}
