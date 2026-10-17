//go:build tossos_testseams

package engine

// a112 태스크 2.8 의 빈칸(2.x 대조 감사): 안전 loop 의 **cadence**(살아 있음만이 아니라 제때 도는지)를 진입 쪽 **실패** 아래에서 잰 시험이
// 없었다 — 포화(7.5 C3)만 있었다. 여기서는 레인 국소 · 시장 국소 실패의 종류를 한 판에 모두 일으킨 채 안전 생애 다섯의 자리 표시 loop 가
// 1 초 cadence 로 정확히 도는지 잰다(synctest 가상 시계):
//   - 시장 국소: KR 시장 주기는 매번 오류(재시작 backoff), US 시장 주기는 panic(비정상 — 감독자가 회복).
//   - 레인 국소: KR continuation step 멈춤(마감 시한까지 비행 — 버림), KR reversal step panic(비정상 실패 → 잠금), KR weekly 는 미리 잠김.
// 중앙 결함(감독자 장부 손상 등)은 설계상 안전 loop 까지 세운다(TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt) — 여기 대상이 아니다.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

func TestSafetyLoopsKeepTheirCadenceThroughEveryEntryFailureKind(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		worker := func(market StrategyMarket, cycle StrategyCycle) StrategyMarketWorker {
			return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
				AuthorityExpiresAt: time.Now().Add(24 * time.Hour), EvidenceDigest: "sha256:" + strings.Repeat("e", 64), LatchRevision: 1}
		}
		supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Workers: []StrategyMarketWorker{
			worker(StrategyMarketKR, func(context.Context) error { return errors.New("a112 2.8 market-local evidence failure") }),
			worker(StrategyMarketUS, func(context.Context) error { panic("a112 2.8 market-local panic") }),
		}, QueueDepth: 1, Clock: clock.System()})
		if err != nil {
			t.Fatal(err)
		}
		lanes := newStrategyLaneRuntime(clock.System(), nil, "")
		release := make(chan struct{})
		defer setStrategyLaneStepHookForTest(lanes, func(lane *strategyworker.Lane, activation strategyrouter.FamilyActivation) strategyworker.Step {
			production, key := strategyFamilyLaneStep(lane, activation), lane.Key()
			return func(ctx context.Context, input strategyworker.Input) (strategyworker.Cycle, error) {
				if key.Market == strategyrouter.MarketKR {
					switch key.Family {
					case strategyrouter.FamilyContinuation:
						<-release
					case strategyrouter.FamilyReversal:
						panic("a112 2.8 lane step panic")
					}
				}
				return production(ctx, input)
			}
		})()
		for _, lane := range lanes.lanesFor(StrategyMarketKR) {
			if lane.Key().Family == strategyrouter.FamilyWeeklyValue {
				lane.Fail("a112 2.8 pre-latched lane", true)
			}
		}

		const cadence = time.Second
		var ticks [len(a112SafetyLoopNames)]atomic.Int64
		loops := []SupervisedLoop{supervisor.SupervisedLoop()}
		for index, name := range a112SafetyLoopNames {
			loops = append(loops, SupervisedLoop{Name: name, Run: func(ctx context.Context) error {
				ticker := time.NewTicker(cadence)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						ticks[index].Add(1)
					}
				}
			}})
		}
		runtime, err := NewRuntime(RuntimeOptions{Loops: loops})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- runtime.Run(ctx) }()
		<-supervisor.Ready()
		// 두 시장 주기를 깨운다(실패 · panic 이 실제로 일어나게) — 감독자는 투입으로 주기를 연다.
		for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
			supervisor.Trigger(market)
		}
		// 레인 물결은 시장 주기처럼 따로 돈다 — 멈춘 레인 때문에 마감 시한(30 초)까지 끝나지 않는다.
		laneWave := make(chan error, 1)
		go func() {
			laneWave <- lanes.evaluate(ctx, StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil, strategyShadowBatch{})
		}()
		synctest.Wait()
		for index := range ticks {
			ticks[index].Store(0)
		}
		time.Sleep(10*cadence + cadence/2)
		// 판정을 모은 뒤 정리를 먼저 하고 보고한다 — 거품 안에서 Fatal 로 빠지면 남은 goroutine 이 교착으로 보고된다.
		var failures []string
		for index, name := range a112SafetyLoopNames {
			if got := ticks[index].Load(); got != 10 {
				failures = append(failures, fmt.Sprintf("%s ran %d cycles in 10.5 cadences while every entry failure kind was live, want exactly 10", name, got))
			}
		}
		// 대조: 실패가 실제로 일어났다 — 두 시장 다 실패를 겪었고 레인 물결은 아직 멈춘 레인을 기다린다.
		for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
			if snapshot, _ := supervisor.Snapshot(market); snapshot.RestartAttempt == 0 && !snapshot.Latched {
				failures = append(failures, fmt.Sprintf("%s market never failed in this window: %+v", market, snapshot))
			}
		}
		select {
		case err := <-laneWave:
			failures = append(failures, fmt.Sprintf("the lane wave finished before the hung lane's deadline: %v", err))
			laneWave <- err
		default:
		}
		cancel()
		close(release)
		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			failures = append(failures, fmt.Sprintf("runtime stop=%v", err))
		}
		<-laneWave
		for _, failure := range failures {
			t.Error(failure)
		}
	})
}
