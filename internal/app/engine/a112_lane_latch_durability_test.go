package engine

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.3.3 이 닫는 구멍을 값으로 잰다: **잠긴 레인이 재시작을
// 견디는가.**
//
// 5.3.1 이 만든 latch 는 `*Lane` 의 bool 이다. 프로세스가 다시 서면 그 값이
// 사라지고, 잠근 이유가 그대로여도 레인은 열린 채로 돌아온다.
// `internal/strategyworker/rehearsal_test.go` 의
// `TestALaneBuiltWithoutADurableRecordIsBornUnlatched`(5.7 이 쓸 때 이름은
// `TestRestartForgetsALatchedLaneWhichIsExactlyWhatTask533MustFix` 였다)가 그 구멍을
// 이미 값으로 못 박아 두었다. 이 로트는 그 시험을 지우지 않는다 — 그 단언은 여전히
// 옳고 뜻만 바뀐다: 기록 없이 세운 레인은 열려 있다.

func laneLatchFixture(t *testing.T) (*Context, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(time.Date(2026, 9, 3, 4, 0, 0, 0, time.UTC))
	opened, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		Clock:    fake,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	t.Cleanup(func() { _ = opened.Close() })
	return &Context{Journal: opened, AccountRef: "acct-lane-latch"}, fake
}

// laneLatchFailingStep 은 레인 하나를 실제로 잠그는 유일한 방법이다.
//
// 생산 단계(`strategyFamilyLaneStep`)로는 잠글 수 없다 — 여덟 레인이 전부
// DORMANT 라 `Run` 이 아무것도 보기 전에 돌아오고 오류가 없다. 그 사실이 이
// 로트가 오늘 생산에 쓰는 행이 0인 이유이기도 하다.
func laneLatchFailingStep(reason string) strategyworker.Step {
	return func(context.Context, strategyworker.Input) (strategyworker.Cycle, error) {
		return strategyworker.Cycle{}, errors.New(reason)
	}
}

func latchOneLane(t *testing.T, lane *strategyworker.Lane, reason string) {
	t.Helper()
	if trigger := lane.Offer(); trigger != strategyworker.TriggerEnqueued {
		t.Fatalf("레인이 사이클을 받지 않았다: %s", trigger)
	}
	bounded, start := lane.RunBounded(context.Background(), strategyworker.Input{}, laneLatchFailingStep(reason))
	if start != strategyworker.StartAdmitted {
		t.Fatalf("레인이 사이클을 시작하지 않았다: %s", start)
	}
	if !bounded.Latched {
		t.Fatalf("실패 하나로 레인이 잠기지 않았다 — 생산 임계값이 1 이 아니다")
	}
	if !lane.Latched() {
		t.Fatal("레인이 잠겼다고 보고하지 않는다")
	}
}

// mustProductionStrategyLanes 는 이제 durable 기록까지 읽으므로 오류를 낼 수
// 있다. 시험에서 그 오류는 언제나 실패다.
func mustProductionStrategyLanes(t *testing.T, c *Context, clk clock.Clock) *strategyLaneRuntime {
	t.Helper()
	runtime, err := c.productionStrategyLanes(context.Background(), clk)
	if err != nil {
		t.Fatalf("레인 런타임을 세우지 못했다: %v", err)
	}
	return runtime
}

func laneByKey(t *testing.T, runtime *strategyLaneRuntime, key strategyworker.Key) *strategyworker.Lane {
	t.Helper()
	for _, lane := range runtime.lanes {
		if lane.Key() == key {
			return lane
		}
	}
	t.Fatalf("레인 %v 가 런타임에 없다", key)
	return nil
}

// TestALatchedLaneComesBackLatchedAfterTheProcessRestarts 는 이 로트의 RED 다.
//
// 고치기 전에는 재시작이 잠긴 레인을 연다. 잠근 이유(증거 저장소 고장, 패닉)는
// 재시작으로 사라지지 않는데 잠금만 사라진다.
//
// 이 시험은 **오늘의 API 만** 쓴다. 그래서 "새 함수가 없어서 컴파일이 안 된다"가
// 아니라 실제 행동으로 빨갛다.
func TestALatchedLaneComesBackLatchedAfterTheProcessRestarts(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()

	runtime := mustProductionStrategyLanes(t, c, fake)
	if runtime == nil {
		t.Fatal("레인 런타임이 서지 않았다")
	}
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	key := lane.Key()
	latchOneLane(t, lane, "breakout evidence store unavailable")

	// 시장 주기가 한 번 돈다. 관측자는 이때 잠긴 레인을 원장에 남겨야 한다.
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("시장 주기가 잠금을 남기지 못했다: %v", err)
	}

	// **프로세스가 다시 선다.** 레인 런타임은 프로세스 수명이므로, 그것을 버리고
	// 다시 만드는 것이 재시작이다.
	c.strategyLanes = nil
	restarted := mustProductionStrategyLanes(t, c, fake)
	if !laneByKey(t, restarted, key).Latched() {
		t.Fatal("재시작이 잠긴 레인을 열었다 — 잠근 이유는 그대로인데 잠금만 사라졌다")
	}
}

// TestALatchOnlyReopensForAStrictlyNewerSignedActivation 은 복구 조건을 엔진
// 경로로 확인한다.
//
// 재시작은 복구가 아니다. 세대를 그대로 들고 다시 서면 레인은 잠긴 채로 온다.
// 세대가 오르면 — 사람이 서명된 매니페스트를 바꿔야만 오르는 수 — 레인이 기록
// 없이 다시 태어난다.
func TestALatchOnlyReopensForAStrictlyNewerSignedActivation(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	key := lane.Key()
	latchOneLane(t, lane, "breakout evidence store unavailable")
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("잠금을 남기지 못했다: %v", err)
	}

	// 같은 세대로 몇 주기를 더 돌아도 열리지 않는다.
	for cycle := 0; cycle < 3; cycle++ {
		if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
			t.Fatalf("주기 %d: %v", cycle, err)
		}
		if !laneByKey(t, runtime, key).Latched() {
			t.Fatal("같은 서명 세대가 잠금을 열었다 — 잠긴 그 상태는 증거가 아니다")
		}
	}

	// 세대가 오른다. 사람이 서명 매니페스트를 바꾼 것이다.
	if err := runtime.evaluate(ctx, StrategyMarketKR, 8, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("복구 주기: %v", err)
	}
	if laneByKey(t, runtime, key).Latched() {
		t.Fatal("더 큰 서명 세대인데 레인이 잠긴 채로 남았다")
	}

	// 그리고 그 복구는 원장에도 남는다 — 다시 서도 열려 있다.
	c.strategyLanes = nil
	restarted := mustProductionStrategyLanes(t, c, fake)
	if laneByKey(t, restarted, key).Latched() {
		t.Fatal("복구된 레인이 재시작에서 다시 잠겼다")
	}
}

// TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart 는 기록이 **첫** 원인을
// 지킨다는 것이다. 마지막 원인을 들고 오면 운영자가 보는 것은 결과이지 원인이
// 아니다.
func TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	key := lane.Key()
	latchOneLane(t, lane, "the first cause")
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("첫 잠금: %v", err)
	}
	// 잠긴 레인에 두 번째 실패를 먹인다. 레인은 두 번 잠기지 않고, 기록도
	// 덮이지 않아야 한다.
	lane.Fail("a later and less interesting failure", false)
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("두 번째 주기: %v", err)
	}

	c.strategyLanes = nil
	restarted := mustProductionStrategyLanes(t, c, fake)
	if got := laneByKey(t, restarted, key).FirstFailure(); got != "the first cause" {
		t.Fatalf("재시작 뒤 첫 원인이 %q 다", got)
	}
}

// TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart 는 골든의
// `peer_lane_state_mutation_forbidden` 이 프로세스 경계를 건너서도 참인지 본다.
func TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	key := lane.Key()
	latchOneLane(t, lane, "one lane, one fault")
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		if err := runtime.evaluate(ctx, market, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
			t.Fatalf("%s: %v", market, err)
		}
	}

	c.strategyLanes = nil
	restarted := mustProductionStrategyLanes(t, c, fake)
	latched := 0
	for _, peer := range restarted.lanes {
		if peer.Latched() {
			latched++
			if peer.Key() != key {
				t.Fatalf("이웃 레인 %v 가 함께 잠긴 채로 돌아왔다", peer.Key().Parts())
			}
		}
	}
	if latched != 1 {
		t.Fatalf("재시작 뒤 잠긴 레인 %d 개 — 하나여야 한다", latched)
	}
}

// TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed 은
// 조용한 실패를 없애되 **나가는 문을 남긴다.**
//
// lane_id 나 버전이 바뀌면 사람이 잠가 둔 것이 아무 일도 하지 않으면서 기록으로만
// 남는다. 그 상태를 조용히 지나가면 운영자는 잠겨 있다고 믿고 런타임은 열려 있다.
// 그래서 주기가 오류를 낸다.
//
// 그리고 그 오류에는 문이 있어야 한다. 첫 판본은 붙지 않은 기록을 버렸고, 버린
// 순간 그것을 닫을 방법이 사라졌다 — 진입이 영원히 멈추고 빠져나갈 길이 없는
// fail-closed 다. 지금은 기록을 들고 있다가 복구도 함께 시도하므로, 더 큰 서명
// 활성화 세대가 오면 닫힌다.
func TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	stale := journal.StrategyLaneLatch{AccountRef: c.AccountRef, Market: "KR", Family: "BREAKOUT_RETEST",
		LaneID: "kr_short_breakout_retest_v9", LaneVersion: "v9",
		LatchID: "lane-latch:KR:BREAKOUT_RETEST:kr_short_breakout_retest_v9:v9:1", LatchRevision: 1,
		Reason: "a lane this build does not have", ActivationGeneration: 7, ObservedAt: fake.Now()}
	if _, err := c.Journal.RecordStrategyLaneLatch(ctx, stale); err != nil {
		t.Fatalf("record: %v", err)
	}
	runtime := mustProductionStrategyLanes(t, c, fake)

	// 같은 세대로는 주기가 계속 빨갛다.
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err == nil {
		t.Fatal("이 빌드에 없는 레인의 잠금을 조용히 넘겼다")
	}
	// 다른 시장은 그 기록에 걸리지 않는다.
	if err := runtime.evaluate(ctx, StrategyMarketUS, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("KR 의 낡은 기록이 US 주기를 막았다: %v", err)
	}
	// 세대가 오르면 닫힌다 — 이것이 나가는 문이다.
	if err := runtime.evaluate(ctx, StrategyMarketKR, 8, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("더 큰 서명 세대인데 낡은 기록이 닫히지 않았다: %v", err)
	}
	open, err := c.Journal.OpenStrategyLaneLatches(ctx, c.AccountRef)
	if err != nil {
		t.Fatalf("open latches: %v", err)
	}
	if len(open) != 0 {
		t.Fatalf("닫혔다고 했는데 열린 기록이 %d 개 남았다", len(open))
	}
}

// TestALedgerThatCannotTakeTheLatchStopsTheCycle 은 쓰기 실패가 오류로 올라오는지
// 본다. 조용히 넘기면 다음 재시작이 잠긴 레인을 열고, 그것이 이 태스크가 없애려는
// 바로 그 동작이다.
func TestALedgerThatCannotTakeTheLatchStopsTheCycle(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	latchOneLane(t, lane, "breakout evidence store unavailable")
	if err := c.Journal.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err == nil {
		t.Fatal("원장이 잠금을 받지 못했는데 주기가 성공했다")
	}
}

// TestTheRecoveryGenerationComesFromTheSignedActivationAndNothingElse 는 이
// 로트에서 **가장 우회하기 쉬운 자리**를 얼린다.
//
// durable latch 의 복구 조건은 "더 큰 수"가 아니라 "더 큰 **서명된** 활성화
// 세대"다. 그 자리에 다른 수(주기 계수기, 시각, 달력 판번호)를 넣으면 트리거는
// 그대로 통과하고 잠금은 저절로 열린다 — 그리고 어떤 행동 시험도 그것을 보지
// 못한다. 두 수 다 그냥 커지기 때문이다.
//
// 그래서 인자로 넘어가는 **식 자체**를 센다. 이 change 가 5.1.2.1 에서
// 소유자 범위의 종목에 쓴 것과 같은 방법이다.
func TestTheRecoveryGenerationComesFromTheSignedActivationAndNothingElse(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read engine package: %v", err)
	}
	fset := token.NewFileSet()
	scanned := 0
	sites := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		for _, decl := range file.Decls {
			fn, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "evaluate" || len(call.Args) < 3 {
					return true
				}
				sites = append(sites, fmt.Sprintf("%s: %s", fn.Name.Name, types.ExprString(call.Args[2])))
				return true
			})
		}
	}
	if scanned < 10 {
		t.Fatalf("비시험 파일 %d 개만 훑었다 — 셈의 범위가 무너졌다", scanned)
	}
	sort.Strings(sites)
	want := []string{
		"runProductionStrategyMarketCycle: fresh.schedule.forMarket(market).restore.Activation.Generation()",
	}
	if strings.Join(sites, "\n") != strings.Join(want, "\n") {
		t.Fatalf("복구 세대를 넘기는 식이 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
			"이 자리에 서명과 무관한 수가 들어가면 잠금은 저절로 열린다. 그 수는 그냥"+
			" 커지기 때문에 어떤 행동 시험도 차이를 못 본다.",
			strings.Join(sites, "\n  "), strings.Join(want, "\n  "))
	}
}

// TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty 는 이 로트가 **오늘
// 생산에서 아무것도 쓰지 않는다**는 주장을 값으로 만든다.
//
// 여덟 레인은 전부 DORMANT 라 `Run` 이 아무것도 보기 전에 돌아오고 오류가 없다.
// 오류가 없으면 잠기지 않고, 잠기지 않으면 원장에 행이 생기지 않는다. 그 사슬을
// 산문으로만 적어 두면 다음 사람이 "쓰기가 0" 을 믿을 근거가 없다.
func TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	for cycle := 0; cycle < 5; cycle++ {
		for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
			if err := runtime.evaluate(ctx, market, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
				t.Fatalf("주기 %d %s: %v", cycle, market, err)
			}
		}
		fake.Advance(10 * time.Second)
	}
	open, err := c.Journal.OpenStrategyLaneLatches(ctx, c.AccountRef)
	if err != nil {
		t.Fatalf("open latches: %v", err)
	}
	if len(open) != 0 {
		t.Fatalf("생산 단계가 레인을 %d 개 잠갔다 — 이 로트는 오늘 쓰기 0 이어야 한다: %+v", len(open), open)
	}
	for _, lane := range runtime.lanes {
		if lane.Latched() {
			t.Fatalf("레인 %v 가 생산 단계로 잠겼다", lane.Key().Parts())
		}
	}
}

// TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther 는
// 이 로트가 공유 런타임에 **가변 상태**를 더했기 때문에 필요하다.
//
// 5.1.2.1 까지 `strategyLaneRuntime.lanes` 는 만든 뒤 바뀌지 않았다. 이제 복구가
// 레인 하나를 갈아 끼우고, 두 시장의 주기 goroutine 이 같은 런타임을 나눠 쓴다.
func TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther(t *testing.T) {
	c, fake := laneLatchFixture(t)
	ctx := context.Background()
	runtime := mustProductionStrategyLanes(t, c, fake)
	kr := runtime.lanesFor(StrategyMarketKR)[0]
	latchOneLane(t, kr, "one market, one fault")
	if err := runtime.evaluate(ctx, StrategyMarketKR, 7, strategyrouter.FamilyActivation{}, nil); err != nil {
		t.Fatalf("잠금을 남기지 못했다: %v", err)
	}

	start := make(chan struct{})
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		wg.Add(1)
		go func(market StrategyMarket) {
			defer wg.Done()
			<-start
			for cycle := 0; cycle < 20; cycle++ {
				// KR 은 8 세대로 복구되고 US 는 잠긴 적이 없다. 둘이 같은
				// 런타임의 목록을 동시에 읽고 고친다.
				if err := runtime.evaluate(ctx, market, 8, strategyrouter.FamilyActivation{}, nil); err != nil {
					failures <- err
					return
				}
			}
		}(market)
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Fatalf("동시 주기가 실패했다: %v", err)
	}
	if laneByKey(t, runtime, kr.Key()).Latched() {
		t.Fatal("더 큰 서명 세대인데 레인이 잠긴 채로 남았다")
	}
	if got := len(runtime.lanes); got != 8 {
		t.Fatalf("동시 복구 뒤 레인 %d 개", got)
	}
}
