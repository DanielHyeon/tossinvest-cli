package engine

import (
	"context"
	"math"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112 7.3.1 SHADOW 실행 자리 · 고장 격리 · 소거(브리프 v3.3 §5 · §5.1 — freeze 종결 2026-10-05).
//
// 무엇이 어디서 도는가:
//   - 주기 경로(dispatch 앞)에서 하는 shadow 일은 **값 보관 하나**다: evaluate 인자로 받은 묶음을 record 가 파도 증가와 같은 잠금에서
//     시장 칸 {wave, batch, activation} 에 덮어쓴다(strategy_lane_runtime.go).
//   - 반사실 판정은 cycle 클로저가 주기 함수의 **nil 반환 뒤** 시작하는 비동기 단계다 — 시장당 단일 비행, 자체 상수 마감, 자가 recover,
//     무오류. 단계가 읽는 것은 시작 때 복사한 칸 · 매니페스트 파일 · 시계뿐이다(공유 refresh 캐시 · 조립 · 활성화 적재를 부르지 않는다 —
//     원격 I/O 0, 다른 시장의 권한 스냅숏 불변).
//   - 실패한 주기(오류 · panic · 마감 경과)는 그 시장의 관측을 즉시 지우고 세대를 올린다. 게시는 시작 때 복사한 세대 · 파도와 같을 때만
//     (CAS) — 실패 뒤 늦게 끝난 단계가 관측을 되살리지 못한다.

const (
	// strategyShadowStepDeadline 은 shadow 단계 하나의 상수 마감이다(레인 마감 · 주기 한도와 독립). 넘으면 그 물결은 「관측 없음」.
	strategyShadowStepDeadline = 2 * time.Second
	// strategyShadowObservationMaxAge 는 파도가 멈춘 채(실패 없음 — supervisor 정지 등) 철회 · 폐기가 가려지는 상한이다. 값은 세 상수에서
	// 유도한다: 건강한 두 게시 사이 최대 간격 = 폴 간격 + 주기 한도 + 단계 마감(37s), 그 두 배(74s). 2×폴 간격(10s)은 10s 넘는 건강한
	// 주기를 거부한다 — fail-closed 는 무엇을 거부하는지 말해야 한다(Manager 승인 2026-10-05).
	strategyShadowObservationMaxAge = 2 * (DefaultStrategyCycleLimit + MaximumStrategyCycleLimit + strategyShadowStepDeadline)
)

// strategyShadowCell 은 한 시장의 마지막 물결이 evaluate 로 실어 온 값이다.
type strategyShadowCell struct {
	wave       uint64
	batch      strategyShadowBatch
	activation strategyrouter.FamilyActivation
}

// strategyShadowObservation 은 한 레인의 게시된 반사실 관측이다.
type strategyShadowObservation struct {
	wave       uint64
	outcome    strategyworker.ShadowOutcome
	expiresAt  time.Time
	observedAt time.Time
}

// shadowObservationUsable 은 투영이 관측을 쓰는 **유일한** 판정이다: 파도 등식(최신 evaluate 물결) ∧ 매니페스트 미만료 ∧ 나이 상한 미만.
// 등호 쪽이 거부다(만료 시각 정각 · 나이 정확히 상한).
func shadowObservationUsable(now time.Time, latestWave uint64, observation strategyShadowObservation) bool {
	return observation.wave == latestWave && now.Before(observation.expiresAt) && now.Sub(observation.observedAt) < strategyShadowObservationMaxAge
}

// productionStrategyCycle 은 두 생산 자리(새로 고침 supervisor · 조립 worker)가 쓰는 cycle 클로저다 — 주기 함수를 run 으로 넘긴다.
func (c *Context) productionStrategyCycle(clk clock.Clock, market StrategyMarket) StrategyCycle {
	return strategyCycleWithShadow(c, clk, market, func(cycleCtx context.Context) error {
		return c.runProductionStrategyMarketCycle(cycleCtx, clk, market)
	})
}

// strategyCycleWithShadow 는 run(시장 주기) 하나를 감싼다.
//
// 실패 판정은 **성공 플래그**다(recover 로 판정하지 않음 — panic 은 그대로 invokeStrategyCycle 의 회복 경로로 간다): run 이 nil 로 돌아오고
// 경과가 MaximumStrategyCycleLimit 미만일 때만 성공이다. 경과가 한도 이상이면 supervisor 가 그 주기를 이미 버렸을 수 있으므로(감시견은
// 주기의 ctx 를 취소하지 않는다) nil 이어도 실패로 친다. 실패면 defer 가 그 시장의 shadow 관측을 지우고 세대를 올린다 — 런타임은
// strategyLanesMu 를 잡는 nil-안전 접근자로 얻는다(부팅 뒤 refresh 가 매번 실패하면 런타임이 없다). 반환값은 run 의 오류 그대로다.
func strategyCycleWithShadow(c *Context, clk clock.Clock, market StrategyMarket, run StrategyCycle) StrategyCycle {
	return func(cycleCtx context.Context) error {
		var started time.Time
		if clk != nil {
			started = clk.Now()
		}
		returnedNil := false
		defer func() {
			if !returnedNil {
				c.strategyLaneRuntimeIfAny().invalidateShadow(market)
			}
		}()
		if err := run(cycleCtx); err != nil {
			return err
		}
		if clk == nil || clk.Now().Sub(started) >= MaximumStrategyCycleLimit {
			return nil
		}
		returnedNil = true
		c.strategyLaneRuntimeIfAny().startShadowStep(cycleCtx, clk, market)
		return nil
	}
}

// strategyLaneRuntimeIfAny 는 이 프로세스의 레인 런타임을 만들지 않고 읽기만 한다(없으면 nil). 잠금 순서 strategyLanesMu → runtime.mu 는
// productionStrategyLanes 와 같다.
func (c *Context) strategyLaneRuntimeIfAny() *strategyLaneRuntime {
	if c == nil {
		return nil
	}
	c.strategyLanesMu.Lock()
	defer c.strategyLanesMu.Unlock()
	return c.strategyLanes
}

// invalidateShadow 는 그 시장의 shadow 관측을 지우고 세대를 올린다. nil 수신자 가드는 런타임의 기존 관례다.
func (runtime *strategyLaneRuntime) invalidateShadow(market StrategyMarket) {
	if runtime == nil {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.shadowEpochs[market]++
	delete(runtime.shadowObserved, market)
}

// startShadowStep 은 그 시장의 칸 · 세대를 **지금 동기로** 복사하고(스탬프는 시작 시점) 단계를 비동기로 띄운다. 이전 단계가 아직 돌면
// 이번 물결은 건너뛰고 센다(단일 비행). 칸이 부재 값이면(조정 앞 닫힘 · 계보 충돌) 아무것도 하지 않는다 — 신선도 규칙이 이전 관측을 버린다.
func (runtime *strategyLaneRuntime) startShadowStep(ctx context.Context, clk clock.Clock, market StrategyMarket) {
	if runtime == nil || clk == nil {
		return
	}
	runtime.mu.Lock()
	cell, epoch := runtime.shadowCells[market], runtime.shadowEpochs[market]
	if runtime.shadowInFlight[market] {
		if runtime.shadowSkipped[market] < math.MaxUint64 {
			runtime.shadowSkipped[market]++
		}
		runtime.mu.Unlock()
		return
	}
	if !cell.batch.observed {
		runtime.mu.Unlock()
		return
	}
	runtime.shadowInFlight[market] = true
	runtime.mu.Unlock()
	go runtime.superviseShadowStep(ctx, clk, market, cell, epoch)
}

// strategyShadowStepResult 는 단계 하나의 결과다. ok=false 는 「관측 없음」(panic).
type strategyShadowStepResult struct {
	ok           bool
	observations map[strategyworker.Key]strategyShadowObservation
}

// superviseShadowStep 은 단계를 상수 마감으로 감싸고, 마감 안에 끝난 결과만 CAS 로 게시한다. 마감을 넘긴 결과는 버린다(관측 없음).
// 단일 비행 표시는 단계 goroutine 이 **끝날 때** 내린다 — 멈춘 단계가 있으면 다음 물결은 건너뛴다(goroutine 이 쌓이지 않음).
func (runtime *strategyLaneRuntime) superviseShadowStep(ctx context.Context, clk clock.Clock, market StrategyMarket,
	cell strategyShadowCell, epoch uint64,
) {
	stepCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan strategyShadowStepResult, 1)
	go func() {
		defer func() {
			runtime.mu.Lock()
			runtime.shadowInFlight[market] = false
			runtime.mu.Unlock()
		}()
		result <- runtime.runShadowStep(stepCtx, clk, market, cell)
	}()
	deadline := make(chan error, 1)
	go func() { deadline <- clk.Sleep(stepCtx, strategyShadowStepDeadline) }()
	select {
	case outcome := <-result:
		if outcome.ok {
			runtime.publishShadow(market, cell.wave, epoch, clk.Now(), outcome.observations)
		}
	case <-deadline:
	}
}

// runShadowStep 은 매니페스트를 한 번 읽고 레인마다 반사실을 판정한다. 자기 panic 을 회복한다(그 물결 「관측 없음」). 판정 대상은
// 매니페스트가 shadow 로 둔 레인 ∩ 같은 물결의 활성화가 OFF/OFF 로 둔 레인(술어 하나 — strategyworker.ShadowEligible).
func (runtime *strategyLaneRuntime) runShadowStep(ctx context.Context, clk clock.Clock, market StrategyMarket,
	cell strategyShadowCell,
) (result strategyShadowStepResult) {
	defer func() {
		if recover() != nil {
			result = strategyShadowStepResult{}
		}
	}()
	config := cell.batch.config
	config.ObservedAt = clk.Now()
	shadow, err := runtime.loadShadow(ctx, config)
	observations := map[strategyworker.Key]strategyShadowObservation{}
	// 미선언(오늘 생산) · 쓸 수 없는 매니페스트는 빈 관측을 게시한다. 이전 물결의 관측은 이미 신선도 규칙(파도 등식)이 버리므로 결론은
	// 게시하지 않는 것과 같다 — 변이 S21 은 그 이유로 동등(SURVIVED)이고, 지키는 쪽인 파도 등식은 변이 S25 가 잰다.
	if err != nil || !shadow.Verified() {
		return strategyShadowStepResult{ok: true, observations: observations}
	}
	for _, lane := range runtime.lanesFor(market) {
		key := lane.Key()
		if !shadow.Shadowed(key.Market, key.Family, key.LaneID, key.LaneVersion) || !lane.ShadowEligible(cell.activation) {
			continue
		}
		observations[key] = strategyShadowObservation{wave: cell.wave, outcome: lane.ShadowOutcomeOver(cell.batch.inputs),
			expiresAt: shadow.ExpiresAt()}
	}
	return strategyShadowStepResult{ok: true, observations: observations}
}

// publishShadow 는 CAS 게시다: 시작 때 복사한 세대와 파도가 지금도 그대로일 때만 그 시장의 관측을 바꾼다.
func (runtime *strategyLaneRuntime) publishShadow(market StrategyMarket, wave, epoch uint64, now time.Time,
	observations map[strategyworker.Key]strategyShadowObservation,
) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.shadowEpochs[market] != epoch || runtime.shadowCells[market].wave != wave {
		return
	}
	for key, observation := range observations {
		observation.observedAt = now
		observations[key] = observation
	}
	runtime.shadowObserved[market] = observations
}
