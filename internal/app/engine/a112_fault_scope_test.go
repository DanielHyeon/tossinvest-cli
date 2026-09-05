package engine_test

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 이 파일이 재는 것은 하나다: **어떤 고장이 어디까지 번지는가.**
//
// a112 태스크 5.6 은 "중앙 무결성 처리를 보존한다 — journal/Gateway/fence/
// multiple-owner 고장은 모든 신규 진입을 막고, lane/market 고장은 국소로 남으며,
// safety loop 는 독립 context 를 유지한다"고 쓰여 있다. 보존하려면 **먼저 재야
// 한다.** 재 보니 이 감독자가 표현할 수 있는 고장 중 여섯 갈래는 패키지 전체
// 스위트에서도 `count=0` 이었다(두 FLM 번들의 branch-test-map). 그 여섯은 우연히
// 지루한 갈래가 아니라 **정확히 중앙으로 번지는 갈래**다 — 즉 오늘 이 저장소는
// "전략 고장이 엔진을 세우는가"를 한 번도 실행해 본 적이 없었다.
//
// 왜 그것이 안전 문제인가. 중앙 무결성 오류는 `Run` 을 반환시키고, Runtime 은
// 첫 정지를 받는 즉시 **모든 loop 를 취소한다**(runtime.go 의 "부분 생존 금지").
// 그 모든 loop 에는 fill detection·reconcile·exit observation 이 들어 있고,
// 엔진이 서면 손절을 놓는 주체가 사라진다. 그러므로 "진입 실패"가 중앙으로
// 번하는 것은 보수적인 선택이 아니라 **더 위험한** 선택이다. 아래 시험들은 그
// 경계가 지금 어디에 그어져 있는지를 값으로 고정한다.
//
// 좌표의 출처는 AST 산출물이다(손으로 읽은 것이 아니다):
// `analysis/function-logic/internal-app-engine--strategyentrysupervisor.latchmarket/ast.json`
// 의 B4(`928:2`)·B5(`932:2`)·B9(`961:2`), 같은 디렉터리
// `…--strategyentrysupervisor.runmarket/ast.json` 의 B5(`787:5`)·B8(`800:4`)·
// B12(`813:4`)·B14(`821:4`), 그리고
// `…--strategyentrysupervisor.waitmarketrestart/ast.json` 의 B3(`845:2`).

// stagedClock 은 시험이 정한 호출 번째부터 **망가진 시각**을 보고한다.
//
// 왜 필요한가. 감독자의 중앙 확대 경로 넷은 전부 "시계나 카운터가 스스로
// 깨졌을 때"에만 열린다. 생성자가 `now.IsZero()` 를 이미 거부하므로 처음부터
// 0 을 주는 시계로는 감독자를 **세울 수조차 없다** — 고장은 세운 뒤에
// 시작되어야 한다. 호출 번째로 트립하는 이유도 같다: 이 시계는 감독자의
// 어느 호출이 깨진 값을 받는지를 시험이 고르게 해 준다.
//
// 번째가 어긋나면 시험은 조용히 다른 것을 재게 된다. 그래서 각 시험은 기대하는
// **원인 문구**까지 확인한다 — 트립이 목표 지점에 닿지 않으면 그 문구가 안 나온다.
type stagedClock struct {
	fake      *clock.Fake
	mu        sync.Mutex
	calls     int
	tripAfter int
	broken    time.Time
}

func newStagedClock(base time.Time, tripAfter int, broken time.Time) *stagedClock {
	return &stagedClock{fake: clock.NewFake(base), tripAfter: tripAfter, broken: broken}
}

func (c *stagedClock) Now() time.Time {
	c.mu.Lock()
	c.calls++
	tripped := c.calls > c.tripAfter
	c.mu.Unlock()
	if tripped {
		return c.broken
	}
	return c.fake.Now()
}

func (c *stagedClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }

func (c *stagedClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.fake.Sleep(ctx, d)
}

func (c *stagedClock) Advance(d time.Duration) { c.fake.Advance(d) }

var _ clock.Clock = (*stagedClock)(nil)

// TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping
// 은 중앙으로 번지는 네 갈래를 **각각 한 번씩 실행**한다. 여섯 빈칸 중 넷이
// 여기서 채워진다(`787-790`, `795-796`, `821-824`, `829-830`).
//
// 네 갈래의 공통점을 이름으로 남긴다: 넷 다 **평가가 실패한 것이 아니라 감독자
// 자신의 장부가 깨진 것**이다 — 관측 시각이 없다, latch revision 이 소진됐다,
// 재시작 기한이 계약 밖이다. 평가 실패(보통 오류·panic·마감 시한)는 이 목록에
// 없고, 그것이 "lane/market 고장은 국소" 절의 실질이다.
func TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	failingCycle := func(context.Context) error { return errors.New("evaluation failed") }

	cases := []struct {
		name   string
		branch string
		build  func() (*engine.StrategyEntrySupervisor, func())
		cause  string
	}{
		{
			name:   "관측 시각이 없으면 잠금을 기록할 수 없다",
			branch: "latchMarket B4 (928:2) → runMarket B14 (821:4)",
			// 호출 순서: 생성자 1, evaluationState 1, latchMarket 이 세 번째.
			build: func() (*engine.StrategyEntrySupervisor, func()) {
				clk := newStagedClock(base, 2, time.Time{})
				kr := activeStrategyWorker(engine.StrategyMarketKR, failingCycle)
				return mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: clk, CycleLimit: engine.MaximumStrategyCycleLimit,
					Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
				}), func() {}
			},
			cause: "strategy fault observation time is unavailable",
		},
		{
			name:   "latch revision 이 소진되면 같은 ID 를 재사용하지 않는다",
			branch: "latchMarket B5 (932:2) → runMarket B14 (821:4)",
			build: func() (*engine.StrategyEntrySupervisor, func()) {
				kr := activeStrategyWorker(engine.StrategyMarketKR, failingCycle)
				kr.LatchRevision = math.MaxUint64
				return mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: clock.NewFake(base), CycleLimit: engine.MaximumStrategyCycleLimit,
					Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
				}), func() {}
			},
			cause: "strategy latch revision exhausted",
		},
		{
			name:   "권한 만료의 잠금도 같은 장부를 쓴다",
			branch: "latchMarket B5 (932:2) → runMarket B5 (787:5)",
			build: func() (*engine.StrategyEntrySupervisor, func()) {
				fake := clock.NewFake(base)
				kr := activeStrategyWorker(engine.StrategyMarketKR, failingCycle)
				kr.AuthorityExpiresAt = base.Add(time.Second)
				kr.LatchRevision = math.MaxUint64
				supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
					Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
				})
				return supervisor, func() { fake.Advance(time.Second) }
			},
			cause: "strategy latch revision exhausted",
		},
		{
			name:   "재시작 기한이 계약 밖이면 기다리지 않는다",
			branch: "waitMarketRestart B3 (845:2) → runMarket B15 (825:4) 의 비취소 갈래 (829-830)",
			// 호출 순서: 생성자 1, evaluationState 1, latchMarket 1, 네 번째가
			// waitMarketRestart 다. 시계가 한 시간 뒤로 물러나면 이미 계산된
			// 절대 기한까지의 남은 시간이 30 초 상한을 넘는다.
			build: func() (*engine.StrategyEntrySupervisor, func()) {
				clk := newStagedClock(base, 3, base.Add(-time.Hour))
				kr := activeStrategyWorker(engine.StrategyMarketKR, failingCycle)
				return mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: clk, CycleLimit: engine.MaximumStrategyCycleLimit,
					Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
				}), func() {}
			},
			cause: "strategy market restart delay is outside the bounded contract",
		},
		{
			name:   "만료 뒤의 재시작 대기도 같은 계약을 쓴다",
			branch: "waitMarketRestart B3 (845:2) → runMarket B6 (791:5) 의 비취소 갈래 (795-796)",
			build: func() (*engine.StrategyEntrySupervisor, func()) {
				clk := newStagedClock(base, 3, base.Add(-time.Hour))
				kr := activeStrategyWorker(engine.StrategyMarketKR, failingCycle)
				kr.AuthorityExpiresAt = base.Add(time.Second)
				supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: clk, CycleLimit: engine.MaximumStrategyCycleLimit,
					Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
				})
				return supervisor, func() { clk.Advance(time.Second) }
			},
			cause: "strategy market restart delay is outside the bounded contract",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			supervisor, arm := testCase.build()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- supervisor.Run(ctx) }()
			waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
			arm()
			if result := supervisor.Trigger(engine.StrategyMarketKR); result != engine.StrategyTriggerEnqueued {
				t.Fatalf("KR trigger=%s", result)
			}
			select {
			case err := <-done:
				if !errors.Is(err, engine.ErrStrategyCentralIntegrity) {
					t.Fatalf("%s: Run=%v, want a central integrity failure", testCase.branch, err)
				}
				if !strings.Contains(err.Error(), testCase.cause) {
					t.Fatalf("%s: Run=%v\n원인 문구가 %q 가 아니다 — 트립이 목표 분기에 닿지 않았을 수 있다."+
						" 이 시험은 호출 번째로 시계를 망가뜨리므로, 감독자가 Now 를 부르는 자리가"+
						" 바뀌면 여기서 조용히 다른 것을 재게 된다.", testCase.branch, err, testCase.cause)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("%s: 중앙 확대가 일어나지 않았다", testCase.branch)
			}
		})
	}
}

// TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt 은 위 넷이
// **무엇을 의미하는지**를 하나의 값으로 잇는다. 위 시험은 `Run` 이 무엇을
// 돌려주는지만 보고, 이 시험은 그 반환이 fill/exit/reconcile 을 함께 죽인다는
// 것을 본다. 둘을 나눈 이유는 앞의 것이 다섯 갈래를 재고 이것이 그 결과의
// **대가**를 재기 때문이다.
//
// 이 시험이 초록이라는 것은 "괜찮다"가 아니라 **"이 네 갈래는 손절을 놓는
// loop 까지 끈다"** 는 뜻이다. 5.1.2 가 여덟 lane 을 실제로 돌리기 시작하면
// 확대 경로가 늘어나는지 여기서 다시 본다.
func TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
		return errors.New("evaluation failed")
	})
	kr.LatchRevision = math.MaxUint64
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: clock.NewFake(base), CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
	})
	safetyStopped := make(chan struct{})
	runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: []engine.SupervisedLoop{
		supervisor.SupervisedLoop(),
		{Name: "safety-proof", Run: func(ctx context.Context) error { <-ctx.Done(); close(safetyStopped); return ctx.Err() }},
	}})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- runtime.Run(context.Background()) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if result := supervisor.Trigger(engine.StrategyMarketKR); result != engine.StrategyTriggerEnqueued {
		t.Fatalf("KR trigger=%s", result)
	}
	if err := <-done; !errors.Is(err, engine.ErrLoopFailed) {
		t.Fatalf("runtime=%v, want ErrLoopFailed", err)
	}
	waitClosed(t, safetyStopped, "safety drain")
}

// TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue 는 빈칸
// `800-801`(runMarket B8)을 채운다.
//
// 이 갈래가 없으면 이미 큐에 들어간 요청이 잠긴 시장을 다시 돌린다. 잠금은
// **평가 전에** 읽히고 큐는 그 뒤에 비워진다는 순서가 계약이다 — 5.3.2 가
// `strategyworker.Lane` 에서 같은 순서를 다시 세운 이유이기도 하다.
func TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var calls atomic.Int32
	kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
		calls.Add(1)
		return errors.New("evaluation failed")
	})
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit, QueueDepth: 2,
		Workers: []engine.StrategyMarketWorker{kr, {Market: engine.StrategyMarketUS}},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")

	// 두 요청을 **잠기기 전에** 넣는다. Trigger 는 잠긴 시장을 거부하므로,
	// 잠긴 뒤에 넣어서는 이 갈래에 닿을 수 없다.
	for i := 0; i < 2; i++ {
		if result := supervisor.Trigger(engine.StrategyMarketKR); result != engine.StrategyTriggerEnqueued {
			t.Fatalf("KR trigger %d=%s", i, result)
		}
	}
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketKR {
		t.Fatalf("fault=%+v", fault)
	}
	// 잠금 뒤의 재시작 대기를 깨워야 두 번째 요청이 읽힌다.
	deadline := time.Now().Add(2 * time.Second)
	drained := false
	for !drained && time.Now().Before(deadline) {
		fake.Advance(engine.MaximumStrategyRestartBackoff)
		snapshot, ok := supervisor.Snapshot(engine.StrategyMarketKR)
		drained = ok && snapshot.Latched && snapshot.QueueDepth == 0
		if !drained {
			time.Sleep(time.Millisecond)
		}
	}
	if !drained {
		t.Fatal("잠긴 시장이 큐에 남은 요청을 소비하지 않았다")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("cycle calls=%d, want 1 — 잠긴 시장이 두 번째 요청으로 다시 평가했다", got)
	}
	select {
	case err := <-done:
		t.Fatalf("시장 국소 잠금이 감독자를 세웠다: %v", err)
	default:
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError 는 빈칸
// `813-814`(runMarket B12)를 채운다. 이 갈래는 **오늘 생산이 실제로 도는
// 유일한 구성**이다: `NewRefreshingPairedStrategyEntrySupervisor` 가 만드는 두
// worker 는 `Effective=false, RefreshesAuthority=true` 이고, `cmd/tossctl` 이
// `SupervisedLoop()` 로 Runtime 에 넣는 것은 그 감독자다.
//
// 그래서 오늘 journal/Gateway 고장이 전략 사이클에서 나와도 시장은 잠기지
// 않는다. 다음 poll 이 다시 시도한다. 이것이 "신규 진입 fail-closed" 를
// 만족하는 방식은 **거절**이지 잠금이 아니다 — dispatch 가 오류를 냈으므로
// 주문은 나가지 않았다.
func TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var calls atomic.Int32
	refreshOnly := func(market engine.StrategyMarket, cycle engine.StrategyCycle) engine.StrategyMarketWorker {
		return engine.StrategyMarketWorker{Market: market, Cycle: cycle,
			PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true}
	}
	kr := refreshOnly(engine.StrategyMarketKR, func(context.Context) error {
		calls.Add(1)
		return errors.New("journal read failed")
	})
	us := refreshOnly(engine.StrategyMarketUS, func(context.Context) error { return nil })
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{kr, us},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")

	deadline := time.Now().Add(3 * time.Second)
	for calls.Load() < 2 && time.Now().Before(deadline) {
		fake.Advance(engine.DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
	if got := calls.Load(); got < 2 {
		t.Fatalf("refresh cycles=%d, want ≥2 — 오류가 poller 를 멈췄다", got)
	}
	snapshot, ok := supervisor.Snapshot(engine.StrategyMarketKR)
	if !ok || snapshot.Latched || snapshot.FirstFailure != "" || snapshot.FirstRefusal != engine.StrategyWorkerRefusalNone {
		t.Fatalf("refresh-only worker latched on a cycle error: %+v", snapshot)
	}
	select {
	case fault := <-supervisor.Faults():
		t.Fatalf("refresh-only worker 가 fault 를 냈다: %+v", fault)
	default:
	}
	select {
	case err := <-done:
		t.Fatalf("refresh-only 사이클 오류가 감독자를 세웠다: %v", err)
	default:
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo 는 위 갈래의
// **날카로운 쪽**이다. `runMarket` 의 순서는 `refreshOnly` 판정이
// `isCentralStrategyIntegrity` 판정보다 **앞**이다(813:4 가 816:4 보다 앞).
// 그래서 오늘 생산이 실제로 돌리는 유일한 구성에서는 중앙 무결성 오류조차
// 삼켜진다.
//
// **이것은 사람이 정할 것이지 리뷰가 정할 것이 아니다.** 두 권위가 갈린다:
//   - `design.md:198` 의 고장표는 "journal/Gateway/fence/owner integrity fault →
//     모든 신규 entry fail-closed" 라고 쓴다.
//   - 같은 절의 "lane context 와 safety context 를 분리한다" 와 spec 의
//     "lane worker 가 safety loop 를 취소해서는 안 된다 (MUST NOT)" 는 반대쪽을 가리킨다.
//
// 오늘 코드는 뒤쪽을 따른다. 순서를 뒤집으면 전략 평가 하나가 엔진을 세우고
// **손절을 놓는 loop 까지 끈다.** 그래서 이 시험은 현재 순서를 값으로 고정하고,
// 바꾸려면 사람의 승인이 필요하다는 것을 실패 문구에 적어 둔다.
func TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var calls atomic.Int32
	kr := engine.StrategyMarketWorker{Market: engine.StrategyMarketKR,
		PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true,
		Cycle: func(context.Context) error {
			calls.Add(1)
			return engine.StrategyCentralIntegrityFailure(errors.New("owner fence CAS failed"))
		}}
	us := engine.StrategyMarketWorker{Market: engine.StrategyMarketUS,
		PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true,
		Cycle: func(context.Context) error { return nil }}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{kr, us},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")

	deadline := time.Now().Add(3 * time.Second)
	for calls.Load() < 2 && time.Now().Before(deadline) {
		fake.Advance(engine.DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
	if got := calls.Load(); got < 2 {
		t.Fatalf("refresh cycles=%d, want ≥2", got)
	}
	select {
	case err := <-done:
		t.Fatalf("dormant refresh worker 의 중앙 무결성 오류가 엔진을 세웠다: %v\n\n"+
			"이 시험이 고정하는 것은 runMarket 의 **판정 순서**다: refreshOnly(813:4) 가"+
			" isCentralStrategyIntegrity(816:4) 보다 앞이라, 진입 권한이 없는 worker 의"+
			" 중앙 오류는 삼켜진다. 그 순서를 뒤집으면 전략 평가 하나가 fill/exit/reconcile"+
			" loop 를 함께 끄고, 엔진이 서면 손절을 놓는 주체가 사라진다. design.md:198 의"+
			" 고장표와 같은 절의 context 분리 요구가 갈리는 자리이므로, 바꾸려면"+
			" 사람의 승인이 먼저 필요하다.", err)
	default:
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone 은 5.6 의
// "lane/market 고장은 국소로 남는다" 절을 **두 시장이 함께 선 상태에서** 값으로
// 확인한다. 기존 스위트에도 비슷한 시험이 있지만, 그것들은 peer 가 dormant 인
// 구성이다. 여기서는 둘 다 effective 라 "국소"가 실제로 고를 수 있는 주장이 된다.
func TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var usCalls atomic.Int32
	kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
		return errors.New("KR evaluation failed")
	})
	us := activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error {
		usCalls.Add(1)
		return nil
	})
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{kr, us},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if result := supervisor.Trigger(engine.StrategyMarketKR); result != engine.StrategyTriggerEnqueued {
		t.Fatalf("KR trigger=%s", result)
	}
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketKR || fault.FirstRefusal != engine.StrategyWorkerRefusalFailure {
		t.Fatalf("fault=%+v", fault)
	}
	if result := supervisor.Trigger(engine.StrategyMarketKR); result != engine.StrategyTriggerDisabled {
		t.Fatalf("잠긴 KR 이 새 요청을 받았다: %s", result)
	}
	if result := supervisor.Trigger(engine.StrategyMarketUS); result != engine.StrategyTriggerEnqueued {
		t.Fatalf("US trigger=%s — peer 가 KR 잠금에 끌려갔다", result)
	}
	eventually(t, func() bool { return usCalls.Load() == 1 }, "US 사이클")
	usSnapshot, ok := supervisor.Snapshot(engine.StrategyMarketUS)
	if !ok || usSnapshot.Latched || !usSnapshot.Effective {
		t.Fatalf("US snapshot=%+v", usSnapshot)
	}
	select {
	case err := <-done:
		t.Fatalf("시장 국소 고장이 감독자를 세웠다: %v", err)
	default:
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// 삼킨 오류가 **세어져서 보인다** (태스크 8.8.4, 항목 7).
//
// 왜 이 시험이 따로 있어야 하는가. 위 두 시험은 삼킴이 **잠그지 않는다**는 것을
// 계약으로 못 박는다 — 그것은 맞고 그대로 둔다. 이 시험이 다루는 것은 그 다음
// 질문이다: 삼킨 뒤에 **아무도 그 일을 모른다.**
//
// 오늘 생산에서 이것이 실제로 일어나는 경로는 이렇다. 레인은 제안 수집 **앞**에
// 서고(5.1.2.2), 레인 세우기는 원장에서 durable latch 를 읽는다(5.3.3). 그래서
// `OpenStrategyLaneLatches` 의 **일시** 실패가 paired assembly 를 통째로 중단시키고,
// 그 assembly 는 두 시장이 나눠 타는 한 파도이므로(5.2.1) **KR·US 가 함께** 오류를
// 받는다. 그 오류가 여기 도착하면 `refreshOnly` 가 `continue` 로 버린다 —
// 카운터도, 스냅샷 필드도, fault 도 없다. 운영자가 볼 수 있는 것이 하나도 없다.
//
// **전파 자체는 고치지 않는다.** 잠금 표는 두 시장을 함께 덮으므로, 그것을 못
// 읽었을 때 두 시장 다 닫는 것이 안전한 방향이다. 고치는 것은 보이지 않는다는 것
// 하나이고, 그래서 이 시험은 `Latched`·`FirstFailure`·fault 부재를 **함께** 단언한다
// — 관측을 더하려다 잠금 자세를 바꾸면 그것이 여기서 빨개져야 한다.
func TestARefreshOnlyWorkerCountsTheCycleErrorsItSwallows(t *testing.T) {
	base := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var calls atomic.Int32
	// 원장이 낸 오류를 그대로 흉내 낸다 — 이 문자열이 스냅샷까지 살아 와야 한다.
	ledgerFailure := errors.New("engine: reading durable strategy lane latches: database is locked")
	// **두 번째부터 다른 오류를 낸다.** 매번 같은 오류를 내면 "첫 값을 지킨다" 와
	// "마지막으로 덮어쓴다" 가 같은 답을 내서 이 시험이 그 축을 아예 못 잰다 —
	// 실제로 첫 판이 그렇게 쓰였고 덮어쓰기 변이가 살아남았다.
	laterFailure := errors.New("engine: a later and different failure")
	kr := engine.StrategyMarketWorker{Market: engine.StrategyMarketKR,
		PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true,
		Cycle: func(context.Context) error {
			if calls.Add(1) == 1 {
				return ledgerFailure
			}
			return laterFailure
		}}
	us := engine.StrategyMarketWorker{Market: engine.StrategyMarketUS,
		PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true,
		Cycle: func(context.Context) error { return nil }}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{kr, us},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")

	deadline := time.Now().Add(3 * time.Second)
	for calls.Load() < 2 && time.Now().Before(deadline) {
		fake.Advance(engine.DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
	if got := calls.Load(); got < 2 {
		t.Fatalf("refresh cycles=%d, want >=2", got)
	}
	// 세어진 수가 보여야 한다. 카운터를 기다리는 것은 사이클 수와 별개다 —
	// 사이클이 돈 것과 그 결과가 기록된 것은 다른 사건이고, 뒤엣것이 이 시험의 대상이다.
	var snapshot engine.StrategyWorkerSnapshot
	for time.Now().Before(deadline) {
		got, ok := supervisor.Snapshot(engine.StrategyMarketKR)
		if ok && got.SwallowedCycleErrors >= 2 {
			snapshot = got
			break
		}
		fake.Advance(engine.DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
	if snapshot.SwallowedCycleErrors < 2 {
		t.Fatalf("삼킨 오류가 세어지지 않았다: SwallowedCycleErrors=%d, want >=2.\n"+
			"이 갈래는 오늘 생산이 실제로 도는 구성이고, 세지 않으면 원장 고장이 "+
			"두 시장의 진입을 조용히 멈춘 채 아무 데도 나타나지 않는다",
			snapshot.SwallowedCycleErrors)
	}
	if snapshot.FirstSwallowedFailure != ledgerFailure.Error() {
		t.Fatalf("첫 원인이 보존되지 않았다\n  got  = %q\n  want = %q\n"+
			"마지막 것을 덮어쓰면 운영자가 **처음** 무엇이 깨졌는지 잃는다 — "+
			"두 번째 사이클부터는 %q 를 내므로 이 칸이 그것이면 덮어쓴 것이다",
			snapshot.FirstSwallowedFailure, ledgerFailure.Error(), laterFailure.Error())
	}
	// 삼킴 자세는 그대로다. 관측을 더하려다 이것이 바뀌면 여기서 잡힌다.
	if snapshot.Latched || snapshot.FirstFailure != "" ||
		snapshot.FirstRefusal != engine.StrategyWorkerRefusalNone || snapshot.RestartAttempt != 0 {
		t.Fatalf("세기를 더하면서 잠금 자세가 바뀌었다: %+v", snapshot)
	}
	select {
	case fault := <-supervisor.Faults():
		t.Fatalf("refresh-only worker 가 fault 를 냈다: %+v", fault)
	default:
	}
	// 성공하는 이웃 시장은 0 이어야 한다 — 세기가 시장별인지 본다.
	peer, ok := supervisor.Snapshot(engine.StrategyMarketUS)
	if !ok || peer.SwallowedCycleErrors != 0 || peer.FirstSwallowedFailure != "" {
		t.Fatalf("이웃 시장이 남의 오류를 세었다: %+v", peer)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}
