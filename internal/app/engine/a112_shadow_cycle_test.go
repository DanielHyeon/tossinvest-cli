//go:build tossos_testseams

package engine

// a112 7.3.1 SHADOW 실행 자리 · 고장 격리(브리프 v3.3 §5 · §8 · §10).
//
// shadow 단계는 cycle 클로저에서 시장 주기 함수가 **nil 로 돌아온 뒤** 비동기로 돈다(dispatch 뒤 — 주문 경로 지연 0). 단계는 evaluate 가
// 같은 잠금에서 보관한 {wave, batch, activation} 의 복사본 · 매니페스트 파일 · 시계만 읽는다. 여기서 재는 것:
//   - 반사실 투영: OFF 레인이 SHADOW · WOULD_EMIT 으로 보인다(빈 표본 금지 — WOULD_EMIT ≥ 1).
//   - 차등 dispatch: shadow 핀 유/무 두 실행의 Gateway 스파이 궤적 · 원장 행이 같다(미선언 · 부분 ON 두 배치).
//   - fault: 단계 panic · 적재 오류 · 마감 초과 → dispatch 궤적 동일 · 레인 잠금 0 · 원장 행 0 · 주기 무오류 · 투영 UNOBSERVED.
//   - 실패 즉시 폐기(i)~(iv) · nil 런타임(v) · central-integrity 신원 · 늦은 nil · 교차 시장 비잠금 · 재시작 셋.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

const (
	a112ShadowRouteDigest = "sha256:" + "0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
	a112ShadowRiskDigest  = "sha256:" + "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"
	a112ShadowCalibration = "sha256:calibration-shadow-kr-v1"
	a112ShadowCalendar    = "kr-regular-2026.10"
)

type a112ShadowWorld struct {
	c       *Context
	clk     *clock.Fake
	spy     *strategyDispatchGatewaySpy
	journal *journal.Journal
	lanes   *strategyLaneRuntime
}

// a112ShadowManifestConfig 는 shadow 매니페스트를 0400 으로 쓰고 결속 설정을 돌려준다. pinned=false 면 핀이 빈(미선언) 설정이다.
func a112ShadowManifestConfig(t *testing.T, now time.Time, pinned bool, expiresIn time.Duration, families ...strategyrouter.Family) strategyshadow.Config {
	t.Helper()
	dir := t.TempDir()
	config := strategyshadow.Config{ConfigDir: dir, Market: strategyrouter.MarketKR, RouteManifestDigest: a112ShadowRouteDigest,
		CalibrationDigest: a112ShadowCalibration, CalendarVersion: a112ShadowCalendar, BuildDigest: strategyRuntimeBuildDigest(),
		RiskPolicyDigest: a112ShadowRiskDigest}
	if !pinned {
		return config
	}
	data, err := strategyshadow.EncodeProductionFamilyShadow(strategyshadow.Document{Market: strategyrouter.MarketKR, Generation: 1,
		RouteManifestDigest: a112ShadowRouteDigest, CalibrationDigest: a112ShadowCalibration, CalendarVersion: a112ShadowCalendar,
		RiskPolicyDigest: a112ShadowRiskDigest, BuildDigest: strategyRuntimeBuildDigest(), Actor: "operator",
		ApprovedAt: now.Add(-2 * time.Hour), IssuedAt: now.Add(-time.Hour), ExpiresAt: now.Add(expiresIn), Shadow: families})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, strategyshadow.ProductionFamilyShadowFileName(strategyrouter.MarketKR))
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	config.ManifestDigest = "sha256:" + hex.EncodeToString(digest[:])
	return config
}

// newA112ShadowWorld 는 dispatch fixture(Gateway 스파이 — 실주문 아님) 위에 KR shadow 묶음(005930 reversal · 000660 continuation 제안)과
// shadow 매니페스트(continuation · reversal 을 shadow)를 얹은 Context 다. activation 이 검증된 값이면 dispatch 쪽 KR 권한에도 같은 활성화를 단다.
func newA112ShadowWorld(t *testing.T, pinned bool, activation strategyrouter.FamilyActivation) *a112ShadowWorld {
	t.Helper()
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	loader := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	fake, ok := loader.clk.(*clock.Fake)
	if !ok {
		t.Fatalf("arrangement: first-leg clock is %T", loader.clk)
	}
	if activation.Verified() {
		kr := proposals.kr
		kr.activation = activation
		// 활성화된 시장의 handoff 는 조립이 중재 때 적은 제안 집합 digest 와 대조된다(strategy_dispatch_handoff.go 가독 계약) — fixture 도
		// 조립과 같은 식으로 적어야 dispatch 가 실제로 돈다(0.5 리뷰 시험#1: 이것이 빠져 부분 ON 배치가 0 == 0 으로 통과했다).
		kr.snapshot.ProposalSetDigest = strategyProposalSetDigest(kr.entries)
		proposals.kr, cycle.proposals.kr = kr, kr
		first := loader.proposals.kr
		first.activation = activation
		loader.proposals.kr = first
	}
	_, batch := a112ShadowCoordinated(t, strategyFamilyGate{}, reversallane.KRReversalLaneID)
	batch = batch.boundTo(a112ShadowManifestConfig(t, fake.Now(), pinned, time.Hour, strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal))
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(fake.Now()))
	if err != nil {
		t.Fatal(err)
	}
	c := &Context{Journal: j, AccountRef: "acct-risk-loader", strategyProjection: store}
	c.strategyRefresh = &StrategyEntryProductionAssembly{dispatch: cycle, proposals: proposals, schedule: cycle.schedule,
		shadow: strategyShadowPair{kr: batch}}
	c.strategyRefreshAt = fake.Now()
	lanes, err := c.productionStrategyLanes(context.Background(), fake)
	if err != nil || lanes == nil {
		t.Fatalf("arrangement: lane runtime: %v", err)
	}
	return &a112ShadowWorld{c: c, clk: fake, spy: spy, journal: j, lanes: lanes}
}

// cycle 은 생산 cycle 클로저(productionStrategyCycle) 하나를 돌리고 shadow 단계가 끝날 때까지 기다린다.
func (world *a112ShadowWorld) cycle(t *testing.T, market StrategyMarket) error {
	t.Helper()
	err := world.c.productionStrategyCycle(world.clk, market)(context.Background())
	a112WaitShadowIdle(t, world.lanes, market)
	return err
}

// a112WaitShadowIdle 은 단계 goroutine(단일 비행 표시)과 감독 goroutine(게시 · 마감 판정) **둘 다** 끝날 때까지 기다림.
// 표시만 보면 감독의 게시 전에 돌아올 수 있음(0.5 리뷰 시험#3) — 감독 종료는 seam 의 stepCtx 닫힘으로 잼.
func a112WaitShadowIdle(t *testing.T, runtime *strategyLaneRuntime, market StrategyMarket) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.mu.RLock()
		busy := runtime.shadowInFlight[market]
		runtime.mu.RUnlock()
		if !busy && strategyShadowSupervisionSettledForTest(runtime) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("the %s shadow step did not finish", market)
}

// a112WaitShadowSupervisionSettled 는 감독 goroutine 만 끝날 때까지 기다림(멈춘 단계를 풀기 **전** — 마감 판정이 끝났음을 확정).
func a112WaitShadowSupervisionSettled(t *testing.T, runtime *strategyLaneRuntime) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if strategyShadowStepsStartedForTest(runtime) > 0 && strategyShadowSupervisionSettledForTest(runtime) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the shadow supervisor did not settle")
}

// a112ShadowLanes 는 KR 레인 넷의 투영(생산 순서)이다.
func a112ShadowLanes(t *testing.T, c *Context) map[string]strategyprojection.LaneRuntimeProjection {
	t.Helper()
	out := map[string]strategyprojection.LaneRuntimeProjection{}
	for _, lane := range a112Read(t, c).Lanes {
		if lane.Market == strategyprojection.MarketKR {
			out[lane.LaneID] = lane
		}
	}
	return out
}

func a112ShadowCount(lanes map[string]strategyprojection.LaneRuntimeProjection) int {
	n := 0
	for _, lane := range lanes {
		if lane.Runtime == strategyprojection.LaneRuntimeShadow && lane.ShadowOutcome != nil &&
			*lane.ShadowOutcome == strategyprojection.LaneShadowWouldEmit {
			n++
		}
	}
	return n
}

type a112DispatchTrace struct {
	observed                             int
	placed                               int
	leases, laneLatches, latchRecoveries int
}

func (world *a112ShadowWorld) trace(t *testing.T) a112DispatchTrace {
	t.Helper()
	world.spy.mu.Lock()
	trace := a112DispatchTrace{observed: world.spy.observed["protection-kr"], placed: len(world.spy.calls)}
	world.spy.mu.Unlock()
	db, err := sql.Open("sqlite", "file:"+world.journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for table, into := range map[string]*int{"strategy_dispatch_leases": &trace.leases, "strategy_lane_latches": &trace.laneLatches,
		"strategy_lane_latch_recoveries": &trace.latchRecoveries} {
		if err := db.QueryRow(`SELECT count(*) FROM ` + table).Scan(into); err != nil {
			t.Fatal(err)
		}
	}
	return trace
}

func TestTheShadowStepProjectsWouldEmitForOffLanesAfterTheCycle(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	if err := world.cycle(t, StrategyMarketKR); err != nil {
		t.Fatalf("cycle err=%v", err)
	}
	lanes := a112ShadowLanes(t, world.c)
	for laneID, want := range map[string]bool{continuationlane.KRContinuationLaneID: true, reversallane.KRReversalLaneID: true} {
		lane := lanes[laneID]
		if want && (lane.Runtime != strategyprojection.LaneRuntimeShadow || lane.ShadowOutcome == nil ||
			*lane.ShadowOutcome != strategyprojection.LaneShadowWouldEmit) {
			t.Fatalf("lane %s runtime=%s outcome=%v, want SHADOW/WOULD_EMIT", laneID, lane.Runtime, lane.ShadowOutcome)
		}
		if lane.Desired != strategyprojection.StateOff || lane.Effective != strategyprojection.StateOff {
			t.Fatalf("SHADOW lane %s is %s/%s — SHADOW must never promote", laneID, lane.Desired, lane.Effective)
		}
	}
	for laneID, lane := range lanes {
		if laneID != continuationlane.KRContinuationLaneID && laneID != reversallane.KRReversalLaneID &&
			(lane.Runtime != strategyprojection.LaneRuntimeUnobserved || lane.ShadowOutcome != nil) {
			t.Fatalf("an unshadowed lane %s is %s/%v", laneID, lane.Runtime, lane.ShadowOutcome)
		}
	}
}

// 0.5 리뷰 시험#2 행동 짝: 성공한 단계 전후로 8 레인 Status() 가 같음. 구조 핀(허용 목록)이 막는 「단계가 레인 상태를 바꾼다」
// (`lane.Offer` 변이 — OFF 레인 pending 을 올려 나중에 켜졌을 때 진짜 투입이 queueDepth 에서 버려짐)를 행동으로도 잼.
// 주기가 바꾼 상태는 기준에 넣고 단계만 다시 띄움: 관측을 지워 세대를 올린 뒤 같은 칸으로 단계를 시작함.
func TestASuccessfulShadowStepLeavesEveryLaneStatusUnchanged(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	if err := world.cycle(t, StrategyMarketKR); err != nil {
		t.Fatalf("cycle err=%v", err)
	}
	statuses := func() map[strategyworker.Key]strategyworker.LaneStatus {
		out := map[strategyworker.Key]strategyworker.LaneStatus{}
		for _, lane := range world.lanes.lanes {
			out[lane.Key()] = lane.Status()
		}
		return out
	}
	world.lanes.invalidateShadow(StrategyMarketKR)
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("arrangement: invalidate left %d SHADOW lanes", n)
	}
	before := statuses()
	if len(before) != 8 {
		t.Fatalf("arrangement: %d lanes, want 8", len(before))
	}
	world.lanes.startShadowStep(context.Background(), world.clk, StrategyMarketKR)
	a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
	// 전제(빈 표본 금지): 단계가 실제로 성공해 WOULD_EMIT 를 게시했어야 「불변」 이 무언가를 잼.
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n < 1 {
		t.Fatal("precondition: the re-run step published no WOULD_EMIT")
	}
	after := statuses()
	for key, want := range before {
		if got := after[key]; got != want {
			t.Errorf("lane %v status changed across a successful shadow step:\n before %+v\n  after %+v", key, want, got)
		}
	}
}

// 차등 dispatch(§8): 두 배치(미선언 · 부분 ON)에서 shadow 핀 유/무의 궤적이 같다. 전제: 핀 쪽 WOULD_EMIT ≥ 1(빈 표본 통과 금지).
func TestTheShadowPinLeavesTheDispatchTraceUnchanged(t *testing.T) {
	partial := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, map[string]bool{continuationlane.KRContinuationLaneID: true})
	for name, activation := range map[string]strategyrouter.FamilyActivation{"undeclared": {}, "partial ON": partial} {
		t.Run(name, func(t *testing.T) {
			with, without := newA112ShadowWorld(t, true, activation), newA112ShadowWorld(t, false, activation)
			errWith, errWithout := with.cycle(t, StrategyMarketKR), without.cycle(t, StrategyMarketKR)
			if (errWith == nil) != (errWithout == nil) || errWith != nil && errWith.Error() != errWithout.Error() {
				t.Fatalf("cycle errors differ: with=%v without=%v", errWith, errWithout)
			}
			if a112ShadowCount(a112ShadowLanes(t, with.c)) < 1 {
				t.Fatal("precondition: the pinned run projected no WOULD_EMIT — the differential would pass on an empty sample")
			}
			if a112ShadowCount(a112ShadowLanes(t, without.c)) != 0 {
				t.Fatal("the unpinned run projected SHADOW")
			}
			// 전제(빈 표본 금지 — 0.5 리뷰 시험#1): 비교하는 궤적에 실제 주문 하나 이상이 있어야 「같다」 가 무언가를 잰다.
			if placed := without.trace(t).placed; placed < 1 {
				t.Fatalf("precondition: the baseline dispatched %d orders — the differential would compare two empty traces", placed)
			}
			if got, want := with.trace(t), without.trace(t); got != want {
				t.Fatalf("dispatch trace with the shadow pin %+v, without %+v — SHADOW must not change dispatch", got, want)
			}
		})
	}
}

// fault(§5): 단계 panic · 적재 오류 · 마감 초과 → 주기 무오류 · dispatch 궤적 동일 · 레인 잠금 0 · 투영 UNOBSERVED.
func TestAFailingShadowStepChangesNothingButItsOwnObservation(t *testing.T) {
	baseline := newA112ShadowWorld(t, false, strategyrouter.FamilyActivation{})
	if err := baseline.cycle(t, StrategyMarketKR); err != nil {
		t.Fatal(err)
	}
	want := baseline.trace(t)
	for _, tc := range []struct {
		name string
		hook func(world *a112ShadowWorld, release chan struct{}) func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error)
	}{
		{"panic", func(*a112ShadowWorld, chan struct{}) func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
			return func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
				panic("a112 shadow step panic")
			}
		}},
		{"load error", func(*a112ShadowWorld, chan struct{}) func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
			return func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
				return strategyshadow.FamilyShadow{}, strategyshadow.ErrProductionFamilyShadowUnavailable
			}
		}},
		{"deadline", func(world *a112ShadowWorld, release chan struct{}) func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
			return func(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
				<-release
				return strategyshadow.LoadProductionFamilyShadow(context.Background(), config)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
			release := make(chan struct{})
			t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, tc.hook(world, release)))
			err := world.c.productionStrategyCycle(world.clk, StrategyMarketKR)(context.Background())
			if err != nil {
				t.Fatalf("cycle err=%v, want nil — the shadow step returns no error", err)
			}
			if tc.name == "deadline" {
				if !world.clk.WaitForSleepers(1, 5*time.Second) {
					t.Fatal("the shadow watchdog never slept")
				}
				world.clk.Advance(strategyShadowStepDeadline + time.Nanosecond)
				// 감시견이 마감을 판정한 뒤 늦은 결과를 낸다 — 버려져야 한다. 실시간 sleep 대신 감독 종료를 seam 으로 확정한 뒤
				// 풂(0.5 리뷰 시험#4 — 부하에서 결과 · 마감이 동시에 준비되면 select 가 무작위로 고르던 경합).
				a112WaitShadowSupervisionSettled(t, world.lanes)
				close(release)
			}
			a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
			if got := world.trace(t); got != want {
				t.Fatalf("dispatch trace %+v, baseline %+v", got, want)
			}
			for _, lane := range world.lanes.lanesFor(StrategyMarketKR) {
				if lane.Latched() {
					t.Fatalf("lane %v latched by a shadow fault", lane.Key())
				}
			}
			if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
				t.Fatalf("a failed shadow step left %d SHADOW lanes", n)
			}
		})
	}
}

// 교차 시장(§5 v3.1): KR shadow 단계가 마감 초과로 멈춘 동안 US 주기 → US 잠기지 않음 · refresh 파도를 일으키지 않음 · US 주기의 결과와
// dispatch 궤적이 shadow 없는 같은 순서(KR → US)와 같다.
//
// 0.5 리뷰 시험#7: 앞 판은 KR 주기가 주문을 내 공유 위험 버킷 원장이 움직였고, 캐시된 US 위험 스냅숏이 낡아(BUCKET_USAGE_STALE) US 가
// 기준 · 실험 둘 다 같은 실패로 끝났음 — 「US dispatch 불변」 이 「같은 오류」 만 쟀음. 그래서 KR 주기는 dispatch 없이 돌림(평가 · record ·
// shadow 칸은 그대로 서고 단계도 시작됨 — dispatch 는 이 시험의 KR 쪽 측정 대상이 아님). US 기준은 주문 ≥ 1 을 전제로 단언함.
func TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket(t *testing.T) {
	krWithoutDispatch := func(world *a112ShadowWorld) error {
		saved := world.c.strategyRefresh.dispatch
		world.c.strategyRefresh.dispatch = nil
		defer func() { world.c.strategyRefresh.dispatch = saved }()
		return world.c.productionStrategyCycle(world.clk, StrategyMarketKR)(context.Background())
	}
	baseline := newA112ShadowWorld(t, false, strategyrouter.FamilyActivation{})
	baselineKR := krWithoutDispatch(baseline)
	a112WaitShadowIdle(t, baseline.lanes, StrategyMarketKR)
	baselineUS := baseline.cycle(t, StrategyMarketUS)
	baseline.spy.mu.Lock()
	baselineUSObserved, baselinePlaced := baseline.spy.observed["protection-us"], len(baseline.spy.calls)
	baseline.spy.mu.Unlock()
	// 전제(빈 표본 금지): US 기준 주기가 실제로 주문을 냈어야 「같다」 가 US dispatch 를 잼.
	if baselineUS != nil || baselinePlaced < 1 || baselineUSObserved < 1 {
		t.Fatalf("precondition: the US baseline err=%v placed=%d observed=%d — want a clean US dispatch ≥ 1", baselineUS, baselinePlaced, baselineUSObserved)
	}

	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	release := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, func(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
		<-release
		return strategyshadow.FamilyShadow{}, nil
	}))
	errKR := krWithoutDispatch(world)
	if (errKR == nil) != (baselineKR == nil) {
		t.Fatalf("KR cycle err=%v, baseline %v", errKR, baselineKR)
	}
	if strategyShadowStepsStartedForTest(world.lanes) < 1 {
		// 단계가 적재기에 닿기 전일 수 있음 — 닿을 때까지 기다림(멈춘 단계가 있어야 이 시험이 무언가를 잼).
		deadline := time.Now().Add(5 * time.Second)
		for strategyShadowStepsStartedForTest(world.lanes) < 1 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if strategyShadowStepsStartedForTest(world.lanes) < 1 {
			t.Fatal("precondition: the KR shadow step never reached its loader — nothing is stuck")
		}
	}
	cached, cachedAt := world.c.strategyRefresh, world.c.strategyRefreshAt
	errUS := world.c.productionStrategyCycle(world.clk, StrategyMarketUS)(context.Background())
	if (errUS == nil) != (baselineUS == nil) || errUS != nil && errUS.Error() != baselineUS.Error() {
		t.Fatalf("US cycle err=%v while the KR shadow step is stuck, baseline %v", errUS, baselineUS)
	}
	world.spy.mu.Lock()
	usObserved, placed := world.spy.observed["protection-us"], len(world.spy.calls)
	world.spy.mu.Unlock()
	if usObserved != baselineUSObserved || placed != baselinePlaced {
		t.Fatalf("US dispatch observed=%d placed=%d, baseline %d/%d", usObserved, placed, baselineUSObserved, baselinePlaced)
	}
	if world.c.strategyRefresh != cached || !world.c.strategyRefreshAt.Equal(cachedAt) || world.c.strategyRefreshWave != nil {
		t.Fatal("the shadow step caused a refresh wave or replaced the shared snapshot")
	}
	for _, lane := range world.lanes.lanesFor(StrategyMarketUS) {
		if lane.Latched() {
			t.Fatalf("US lane %v latched while KR shadow was stuck", lane.Key())
		}
	}
	close(release)
	a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
}

// (i)~(iii) 실패 즉시 폐기 + epoch CAS. 실패는 cycle 클로저의 run 인자로 넣는다 — 레인 런타임이 보기에 record 전 실패(파도 불변)와 같은 모양이다.
func TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	// 이 시험은 dispatch 를 재지 않는다 — 같은 제안을 주기마다 다시 내지 않게 dispatch 없는 조립(주기 함수가 evaluate 뒤 nil)으로 돈다.
	world.c.strategyRefresh.dispatch = nil
	if err := world.cycle(t, StrategyMarketKR); err != nil || a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
		t.Fatalf("arrangement: no SHADOW at wave W (err=%v)", err)
	}
	failing := errors.New("a112 pre-record failure")
	run := strategyCycleWithShadow(world.c, world.clk, StrategyMarketKR, func(context.Context) error { return failing })
	if err := run(context.Background()); !errors.Is(err, failing) {
		t.Fatalf("(i) err=%v, want the cycle's own error unchanged", err)
	}
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("(i) after a pre-record error %d lanes still show SHADOW", n)
	}
	// (ii) record 전 panic: 같은 값이 위로 가고 관측은 지워진다.
	if err := world.cycle(t, StrategyMarketKR); err != nil || a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
		t.Fatalf("arrangement: no SHADOW before the panic case (err=%v)", err)
	}
	panicking := strategyCycleWithShadow(world.c, world.clk, StrategyMarketKR, func(context.Context) error { panic("a112 pre-record panic") })
	if recovered := func() (value any) {
		defer func() { value = recover() }()
		_ = panicking(context.Background())
		return nil
	}(); recovered != "a112 pre-record panic" {
		t.Fatalf("(ii) recovered=%v, want the same panic value propagated", recovered)
	}
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("(ii) after a pre-record panic %d lanes still show SHADOW", n)
	}
	// (iii) W 의 in-flight 단계가 실패 **뒤** 끝나도 되살아나지 않는다.
	release := make(chan struct{})
	t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, func(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
		<-release
		return strategyshadow.LoadProductionFamilyShadow(context.Background(), config)
	}))
	if err := world.c.productionStrategyCycle(world.clk, StrategyMarketKR)(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background()); !errors.Is(err, failing) {
		t.Fatal(err)
	}
	close(release)
	a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("(iii) a late in-flight step revived %d SHADOW lanes after the failure", n)
	}
}

// (iv) 늦은 nil: 주기가 nil 이어도 경과 ≥ MaximumStrategyCycleLimit 이면 실패 취급(폐기 · 단계 미기동). 한도 − 1ns 는 시작한다.
func TestALateNilCycleIsAFailureAtTheLimit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		elapsed time.Duration
		shadow  bool
	}{
		{"limit − 1ns starts the step", MaximumStrategyCycleLimit - time.Nanosecond, true},
		{"exactly the limit is a failure", MaximumStrategyCycleLimit, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
			run := strategyCycleWithShadow(world.c, world.clk, StrategyMarketKR, func(ctx context.Context) error {
				err := world.c.runProductionStrategyMarketCycle(ctx, world.clk, StrategyMarketKR)
				world.clk.Advance(tc.elapsed)
				return err
			})
			if err := run(context.Background()); err != nil {
				t.Fatalf("cycle err=%v", err)
			}
			a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
			if got := a112ShadowCount(a112ShadowLanes(t, world.c)) > 0; got != tc.shadow {
				t.Fatalf("SHADOW shown=%v, want %v", got, tc.shadow)
			}
		})
	}
}

// (v) 부팅 첫 주기에 refresh 오류 · 레인 런타임 없음 → 반환 오류 동일 · panic 0. (삼킴 갈래의 FirstSwallowedFailure 는 err.Error() 이므로
// 오류가 같으면 같다.)
func TestTheShadowDeferNeverRewritesAFirstCycleError(t *testing.T) {
	c := &Context{}
	direct := c.runProductionStrategyMarketCycle(context.Background(), nil, StrategyMarketKR)
	if direct == nil {
		t.Fatal("arrangement: the direct cycle must fail without a clock")
	}
	var got error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("the shadow defer panicked: %v", recovered)
			}
		}()
		got = c.productionStrategyCycle(nil, StrategyMarketKR)(context.Background())
	}()
	if got == nil || got.Error() != direct.Error() {
		t.Fatalf("wrapped err=%v, direct err=%v — the defer must not rewrite it", got, direct)
	}
	if c.strategyLanes != nil {
		t.Fatal("the defer created a lane runtime")
	}
}

// central-integrity 신원: 생산 클로저를 지난 오류 · panic 이 invokeStrategyCycle 뒤에도 중앙 무결성으로 읽힌다.
func TestCentralIntegrityKeepsItsIdentityThroughTheShadowDefer(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	cause := StrategyCentralIntegrityFailure(errors.New("a112 central invariant"))
	for name, run := range map[string]StrategyCycle{
		"returned": func(context.Context) error { return cause },
		"panicked": func(context.Context) error { panic(cause) },
	} {
		outcome := invokeStrategyCycle(context.Background(), strategyCycleWithShadow(world.c, world.clk, StrategyMarketKR, run))
		if !isCentralStrategyIntegrity(outcome.err) {
			t.Errorf("%s: err=%v lost its central-integrity identity", name, outcome.err)
		}
	}
}

// 재시작 셋(§10): 앞 프로세스가 전체 주기에서 WOULD_EMIT ≥ 1 → 같은 원장 · 새 Context 가 전체 주기를 1 회 이상 돈 뒤 잰다.
func TestARestartNeverRestoresShadow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config func(t *testing.T, now time.Time) strategyshadow.Config
		shadow bool
	}{
		{"① no pin", func(t *testing.T, now time.Time) strategyshadow.Config {
			return a112ShadowManifestConfig(t, now, false, time.Hour)
		}, false},
		{"② pin without a usable file", func(t *testing.T, now time.Time) strategyshadow.Config {
			config := a112ShadowManifestConfig(t, now, false, time.Hour)
			config.ManifestDigest = "sha256:" + strings.Repeat("e", 64)
			return config
		}, false},
		{"② pin of an expired manifest", func(t *testing.T, now time.Time) strategyshadow.Config {
			return a112ShadowManifestConfig(t, now, true, 0, strategyrouter.FamilyContinuation)
		}, false},
		{"③ a valid pin re-reads on the first wave", func(t *testing.T, now time.Time) strategyshadow.Config {
			return a112ShadowManifestConfig(t, now, true, time.Hour, strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal)
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
			if err := before.cycle(t, StrategyMarketKR); err != nil || a112ShadowCount(a112ShadowLanes(t, before.c)) < 1 {
				t.Fatalf("precondition: the first process saw no WOULD_EMIT (err=%v)", err)
			}
			// 재시작: 같은 원장 · 새 Context · 새 레인 런타임 · 같은 dispatch 조립 — 묶음의 설정만 이 경우의 것.
			assembly := *before.c.strategyRefresh
			assembly.dispatch = nil // 재시작 뒤 주기는 dispatch 를 재지 않는다(같은 제안 재발행 없음) — 원장 행 0 을 잰다.
			batch := assembly.shadow.kr
			batch.config = tc.config(t, before.clk.Now())
			assembly.shadow = strategyShadowPair{kr: batch}
			store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(before.clk.Now()))
			if err != nil {
				t.Fatal(err)
			}
			restarted := &Context{Journal: before.journal, AccountRef: before.c.AccountRef, strategyProjection: store}
			restarted.strategyRefresh, restarted.strategyRefreshAt = &assembly, before.clk.Now()
			lanes, err := restarted.productionStrategyLanes(context.Background(), before.clk)
			if err != nil {
				t.Fatal(err)
			}
			// 첫 물결 전: 여덟 UNOBSERVED(관측 메모리 비복원).
			for _, lane := range a112Read(t, restarted).Lanes {
				if lane.Runtime != strategyprojection.LaneRuntimeUnobserved || lane.ShadowOutcome != nil {
					t.Fatalf("before the first wave lane %s is %s/%v", lane.LaneID, lane.Runtime, lane.ShadowOutcome)
				}
			}
			world := &a112ShadowWorld{c: restarted, clk: before.clk, spy: before.spy, journal: before.journal, lanes: lanes}
			if err := world.cycle(t, StrategyMarketKR); err != nil {
				t.Fatal(err)
			}
			if got := a112ShadowCount(a112ShadowLanes(t, restarted)) > 0; got != tc.shadow {
				t.Fatalf("after the first wave SHADOW shown=%v, want %v", got, tc.shadow)
			}
			trace := world.trace(t)
			if trace.laneLatches != 0 || trace.latchRecoveries != 0 {
				t.Fatalf("ledger after the restart %+v — SHADOW must write nothing", trace)
			}
		})
	}
}

// 단일 비행: 단계가 아직 돌면 다음 물결의 단계는 건너뛰고 센다(적재는 한 번).
func TestOnlyOneShadowStepRunsPerMarket(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	world.c.strategyRefresh.dispatch = nil
	release, loads := make(chan struct{}), 0
	var mu sync.Mutex
	t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, func(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
		mu.Lock()
		loads++
		mu.Unlock()
		<-release
		return strategyshadow.LoadProductionFamilyShadow(context.Background(), config)
	}))
	for range 2 {
		if err := world.c.productionStrategyCycle(world.clk, StrategyMarketKR)(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	close(release)
	a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
	world.lanes.mu.RLock()
	skipped := world.lanes.shadowSkipped[StrategyMarketKR]
	world.lanes.mu.RUnlock()
	mu.Lock()
	defer mu.Unlock()
	if loads != 1 || skipped != 1 {
		t.Fatalf("loads=%d skipped=%d, want one step and one skipped wave", loads, skipped)
	}
}

// 술어의 두 자리: 단계는 활성화 ON 레인을 관측에 넣지 않고(ShadowEligible), 투영은 ON 관측 값을 SHADOW 로 올리지 않는다(값 형태).
func TestNeitherTheStepNorTheProjectionShadowsAnOnLane(t *testing.T) {
	partial := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, map[string]bool{continuationlane.KRContinuationLaneID: true})
	world := newA112ShadowWorld(t, true, partial)
	world.c.strategyRefresh.dispatch = nil
	if err := world.cycle(t, StrategyMarketKR); err != nil {
		t.Fatal(err)
	}
	world.lanes.mu.RLock()
	observed := world.lanes.shadowObserved[StrategyMarketKR]
	world.lanes.mu.RUnlock()
	for key := range observed {
		if key.LaneID == continuationlane.KRContinuationLaneID {
			t.Fatal("the step observed a lane the activation turned ON")
		}
	}
	if len(observed) == 0 {
		t.Fatal("precondition: the step observed no OFF lane")
	}
	lane := world.lanes.lanesFor(StrategyMarketKR)[0]
	on := strategyLaneObservation{Key: lane.Key(), Wave: 1, Trigger: strategyworker.TriggerDisabled,
		Desired: strategyrouter.StateOn, Effective: strategyrouter.StateOn}
	shadow := &strategyShadowObservation{wave: 1, outcome: strategyworker.ShadowWouldEmit}
	if got := strategyLaneProjection(lane, on, true, shadow); got.Runtime != strategyprojection.LaneRuntimeUnobserved || got.ShadowOutcome != nil {
		t.Fatalf("an ON observation was projected as %s/%v", got.Runtime, got.ShadowOutcome)
	}
}

// 0.5 리뷰 유지#3: 투영의 OFF/OFF 재확인(strategy_lane_projection.go)은 단계의 ShadowEligible 뒤라 생산에서는 닿지 않고, 위 시험은
// ON/ON 만 넣어 `Desired` 항을 지운 변이(M2)가 살아남았음. 직접 호출로 네 조합을 다 넣어 두 항을 각각 못 박음 —
// desired ON · effective OFF(켜기로 했지만 아직 안 켜진 레인)도 SHADOW 가 아님. OFF/OFF 는 양성 대조(SHADOW 로 보여야 함).
func TestTheProjectionShadowsOnlyAnOffOffObservation(t *testing.T) {
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	lane := world.lanes.lanesFor(StrategyMarketKR)[0]
	shadow := &strategyShadowObservation{wave: 1, outcome: strategyworker.ShadowWouldEmit}
	for _, tc := range []struct {
		desired, effective strategyrouter.DesiredState
		shadowed           bool
	}{
		{strategyrouter.StateOff, strategyrouter.StateOff, true},
		{strategyrouter.StateOn, strategyrouter.StateOff, false},
		{strategyrouter.StateOff, strategyrouter.StateOn, false},
		{strategyrouter.StateOn, strategyrouter.StateOn, false},
	} {
		observation := strategyLaneObservation{Key: lane.Key(), Wave: 1, Trigger: strategyworker.TriggerDisabled,
			Desired: tc.desired, Effective: tc.effective}
		got := strategyLaneProjection(lane, observation, true, shadow)
		isShadow := got.Runtime == strategyprojection.LaneRuntimeShadow && got.ShadowOutcome != nil
		if isShadow != tc.shadowed {
			t.Errorf("desired=%s effective=%s projected %s/%v, want shadowed=%v", tc.desired, tc.effective, got.Runtime, got.ShadowOutcome, tc.shadowed)
		}
		if !tc.shadowed && (got.Runtime != strategyprojection.LaneRuntimeUnobserved || got.ShadowOutcome != nil) {
			t.Errorf("desired=%s effective=%s projected %s/%v, want UNOBSERVED with no outcome", tc.desired, tc.effective, got.Runtime, got.ShadowOutcome)
		}
	}
}
