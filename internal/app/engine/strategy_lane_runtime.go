package engine

import (
	"context"
	"sync"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.1.2 의 앞 절반이다: 여덟 전략군 레인이 **생산 런타임 안에**
// 서고, 전략군 하나를 평가하는 일이 `*Context` 클로저가 아니게 만든다.
//
// 왜 그것이 안전 문제인가. 오늘 생산이 "worker" 라고 부르는 것은
// `StrategyMarketWorker` 이고 그 `Cycle` 은 `*Context` 를 담은 클로저다
// (`runProductionStrategyMarketCycle`). `*Context` 는 Journal 과 Gateway 를
// 들고 있으므로, 전략군 하나를 평가하는 일이 원장을 쓰고 주문을 낼 수 있는
// 자리에서 일어난다. 스펙이 요구하는 것은 그 반대다 — "Worker dependency
// closure 에는 broker mutator, writable journal, Guardian issuer,
// activation/toggle writer 가 없어야 한다 (MUST NOT)".
//
// **이 로트가 하지 않는 것을 먼저 적는다.** 여덟은 아직 생산 진입의 관문이
// 아니다. 동결 골든 `four-family-runtime-v1.json` 이 여덟 서술자를 전부
// `effective: OFF` 로 얼렸고, 스펙은 "Legacy 3-family approval 은 4-family
// activation 으로 자동 승격되어서는 안 된다 (MUST NOT)" 고 적었다. 그래서
// 오늘 여덟은 모두 DORMANT 를 돌려주고 기존 시장 단위 경로가 그대로 주문을
// 낸다. 관문을 옮기려면 서명된 활성화 매니페스트가 있어야 하고, 그것은 이
// 태스크가 아니라 5.1.2.2 와 8 절의 일이다. 여기서 그것을 유도해 켜면
// 위 MUST NOT 을 어기는 것이다.

// strategyLaneObservation 은 레인 하나가 한 주기에 남긴 읽기 전용 관측이다.
//
// 여기에 진입을 다시 여는 방법이나 상태를 바꾸는 방법은 없다. 관측이 복구
// 수단을 겸하면 "복구에는 증거가 필요하다"가 거짓이 된다.
type strategyLaneObservation struct {
	Key       strategyworker.Key
	Trigger   strategyworker.Trigger
	Start     strategyworker.Start
	Outcome   strategyworker.Outcome
	Health    strategyworker.LaneHealth
	Detail    string
	Failure   string
	Abnormal  bool
	Cancelled bool
	Abandoned bool
	Latched   bool
	Emitted   bool
	// 아래는 a112 7.3 투영이 읽는 관측 시점 값이다. Wave 는 record 가 찍는 시장별 물결 번호(0 = 미관측), Desired · Effective 는 그 물결이
	// 받은 활성화가 이 레인에 대해 말한 값, 두 digest 는 그 물결에 이 레인이 받은 입력(없으면 빈 값)이다.
	Wave                           uint64
	Refusal                        strategyarbiter.Refusal
	Desired, Effective             strategyrouter.DesiredState
	SnapshotDigest, EvidenceDigest string
}

// strategyLaneRuntime 은 프로세스가 사는 내내 **같은** 여덟 레인이다.
//
// 왜 프로세스가 살아 있는 내내 같은 것이어야 하나. 레인의 latch 와 연속 실패
// 계수기는 기억이다. 새로 고침(refresh)마다 레인을 다시 만들면 잠긴 레인이
// 1 초 뒤에 열린 채로 돌아온다 — 잠근 이유는 그대로인데. 그래서 이 값은
// `*Context` 가 한 번만 만들고 계속 들고 있는다(`Context.strategyLanes`).
type strategyLaneRuntime struct {
	// clk 와 ledger·accountRef 는 레인을 **다시 세우는** 데 필요하다. 복구는
	// 잠금을 푸는 것이 아니라 기록 없이 다시 태어나게 하는 것이므로(5.3.3),
	// 런타임은 자기가 무엇으로 레인을 만들었는지 기억해야 한다.
	clk        clock.Clock
	ledger     strategyLaneLedger
	accountRef string

	mu       sync.RWMutex
	lanes    []*strategyworker.Lane
	observed map[strategyworker.Key]strategyLaneObservation
	// restored 는 durable 기록을 이미 한 번 읽었는지다. 두 시장이 같은 런타임을
	// 나눠 쓰므로, 두 번 읽으면 두 번째가 첫 시장이 이번 프로세스에서 만든
	// 잠금을 지운다.
	restored bool
	// latches 는 지금 **열려 있는** 원장 기록이다. 원장의 순번을 들고 있어야
	// 복구를 요청할 수 있다.
	latches map[strategyworker.Key]journal.StrategyLaneLatch
	// waves 는 시장별 evaluate 물결 번호다(a112 7.3, 판정 Q1=(B)). 프로세스 수명이고 record 가 mu 아래에서 올린다.
	waves map[StrategyMarket]uint64
	// unmatched 는 이 빌드에 없는 레인을 가리키는 기록이다. 버리지 않고 들고
	// 있는 이유는 그것이 복구를 요청할 수 있는 유일한 손잡이이기 때문이다.
	unmatched []journal.StrategyLaneLatch
	// 아래는 a112 7.3.1 SHADOW 상태다(전부 mu 아래, 프로세스 수명 — 원장 · 파일에 쓰지 않으므로 재시작이 되살리지 않는다).
	//   shadowCells: 시장별 칸 하나 {wave, batch, activation} — record 가 파도를 올리는 같은 임계 구역에서 덮어쓴다(누적 없음).
	//   shadowEpochs: 실패한 주기마다 오르는 세대 — 게시는 시작 때 복사한 세대 · 파도와 같을 때만(CAS).
	//   shadowObserved: 게시된 관측(시장 → 레인 열쇠). 투영은 shadowObservationUsable 하나로만 쓴다.
	//   shadowInFlight · shadowSkipped: 시장당 단일 비행과 그때 건너뛴 물결 수(시험 관측 전용 — 생산 독자 · 투영 노출 없음).
	shadowCells    map[StrategyMarket]strategyShadowCell
	shadowEpochs   map[StrategyMarket]uint64
	shadowObserved map[StrategyMarket]map[strategyworker.Key]strategyShadowObservation
	shadowInFlight map[StrategyMarket]bool
	shadowSkipped  map[StrategyMarket]uint64
}

// newStrategyLaneRuntime 은 생산 레인 여덟을 세운다.
//
// 목록도 정책도 여기서 고르지 않는다. 여덟이 누구인지는 `strategyworker` 의
// 생산 진입점 하나가 정하고, 그 목록이 동결 골든과 같은지는 그 패키지의
// golden_contract_test.go 가 골든 파일을 직접 읽어 대조한다.
func newStrategyLaneRuntime(clk clock.Clock, ledger strategyLaneLedger, accountRef string) *strategyLaneRuntime {
	if clk == nil {
		return nil
	}
	lanes := strategyworker.ProductionLanes(clk)
	// shadow 맵은 여기서 만든다 — nil 맵 대입은 panic 이고, 그 panic 이 주기 경로(record · 실패 폐기 defer)에서 나면 원래 오류를 덮는다
	// (브리프 v3.3 N5).
	return &strategyLaneRuntime{clk: clk, ledger: ledger, accountRef: accountRef, lanes: lanes,
		observed:       make(map[strategyworker.Key]strategyLaneObservation, len(lanes)),
		latches:        map[strategyworker.Key]journal.StrategyLaneLatch{},
		shadowCells:    make(map[StrategyMarket]strategyShadowCell, 2),
		shadowEpochs:   make(map[StrategyMarket]uint64, 2),
		shadowObserved: make(map[StrategyMarket]map[strategyworker.Key]strategyShadowObservation, 2),
		shadowInFlight: make(map[StrategyMarket]bool, 2),
		shadowSkipped:  make(map[StrategyMarket]uint64, 2)}
}

// productionStrategyLanes 는 이 프로세스의 여덟 레인을 돌려준다. 없으면 만든다.
//
// 게으른 이유는 시계가 첫 생산 주기와 함께 도착하기 때문이고, 공유하는 이유는
// 위에 적은 그대로다 — 레인의 잠금은 새로 고침보다 오래 살아야 한다.
// `strategyDispatchOwner` 가 같은 이유로 같은 모양이다.
//
// **ctx 를 받는 이유는 durable 기록 때문이다**(5.3.3). 레인은 열린 채로 태어난
// 뒤 나중에 잠기는 것이 아니라 기록에서 태어난다. 만들어 놓고 나중에 되살리면
// 그사이에 `Latched()` 가 거짓말을 하고, 그 창을 아무도 못 본다.
func (c *Context) productionStrategyLanes(ctx context.Context, clk clock.Clock) (*strategyLaneRuntime, error) {
	if c == nil || clk == nil {
		return nil, nil
	}
	c.strategyLanesMu.Lock()
	defer c.strategyLanesMu.Unlock()
	if c.strategyLanes == nil {
		// 원장을 인터페이스로 좁혀 넘긴다. nil 저널을 그대로 넘기면 인터페이스가
		// non-nil 이 되어 "원장이 없다" 를 판정할 수 없다.
		var ledger strategyLaneLedger
		if c.Journal != nil {
			ledger = c.Journal
		}
		c.strategyLanes = newStrategyLaneRuntime(clk, ledger, c.AccountRef)
	}
	if err := c.strategyLanes.restoreLatches(ctx); err != nil {
		return nil, err
	}
	return c.strategyLanes, nil
}

// lanesFor 는 이 시장이 맡은 네 레인이다.
//
// 시장은 레인 열쇠에서 읽는다. 여기서 따로 세어 두지 않는 이유는, 세어 두면
// 목록이 바뀌었을 때 두 수가 갈라지고 그 차이를 아무도 보고하지 않기 때문이다.
func (runtime *strategyLaneRuntime) lanesFor(market StrategyMarket) []*strategyworker.Lane {
	if runtime == nil {
		return nil
	}
	routerMarket := strategyRouterMarket(market)
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	selected := make([]*strategyworker.Lane, 0, len(runtime.lanes))
	for _, lane := range runtime.lanes {
		if lane.Key().Market == routerMarket {
			selected = append(selected, lane)
		}
	}
	return selected
}

// strategyFamilyLaneStep 은 레인 하나가 사이클마다 실제로 도는 일이다.
//
// **이 함수가 패키지 수준이고 인자가 레인 하나뿐인 것이 이 태스크의 요점이다.**
// 메서드로 두면 수신자가 무엇이든 될 수 있고, `*Context` 를 수신자로 두는 순간
// 원장과 게이트웨이가 전략군 평가 안으로 들어온다 — 오늘 생산의 시장 주기가
// 정확히 그 모양이다. 여기서는 무엇을 만질 수 있는지가 주석이 아니라 **인자의
// 타입**으로 정해지고, `*strategyworker.Lane` 이 사는 패키지는 자기 import
// 폐포에 broker mutator·writable journal·Guardian issuer 가 없다는 것을
// `-deps` 로 훑어 시험으로 지킨다.
//
// **활성화도 인자다** (태스크 8.7.1). 승격은 이 함수가 정하지 않고 서명된
// 매니페스트가 정하며, 관문이 이 주기에 읽은 그 값을 그대로 받는다. 여기서
// 다시 읽으면 만료가 그 사이에 지나갔을 때 관문과 관측이 갈린다.
//
// 오류를 절대 돌려주지 않는 이유: 평가 실패는 오류가 아니라 **거절**이고,
// 거절은 `Cycle.Outcome` 에 담긴다. 둘을 한 값에 담으면 정당한 거절이 레인의
// 연속 실패 계수기를 올려 결국 진입을 잠근다.
func strategyFamilyLaneStep(lane *strategyworker.Lane,
	activation strategyrouter.FamilyActivation,
) strategyworker.Step {
	return func(_ context.Context, input strategyworker.Input) (strategyworker.Cycle, error) {
		return lane.Run(activation, input), nil
	}
}

// evaluate 는 이 시장의 네 레인에 각자 자기 것인 봉인된 제안을 한 번씩 돌리고
// 그 관측을 기록한다.
//
// **관측은 돌려주지 않는다.** 앞선 판본은 관측과 봉투를 함께 돌려줬는데, 그러면
// 호출자가 그중 하나를 버릴 수 있고 버린 것을 아무도 못 본다. 같은 문제를
// dispatch 주기에서 `Deliver` 로 푼 것과 같은 답이다 — 무시할 수 있는 답을
// 호출자에게 주지 않는다. 결과는 이 런타임 안에 남고 observations 가 읽는다.
//
// **오류는 돌려준다**(5.3.3). 그것은 답이 아니라 실패다: durable latch 를 원장에
// 남기지 못했거나, 이 빌드에 없는 레인을 가리키는 기록이 남아 있다는 뜻이다.
// 조용히 넘기면 다음 재시작이 잠긴 레인을 열고, 그것이 이 로트가 없애려는 바로
// 그 동작이다.
//
// 어느 제안이 어느 레인의 것인지는 여기서 판정하지 않고 레인에게 묻는다
// (`lane.Owns`). 같은 판정을 두 곳에 두면 운영자가 보는 진단이 갈리고, 여기
// 옮겨 적은 사본은 봉인을 먼저 보는 것을 잊기 쉽다.
//
// 자기 제안이 없는 레인도 사이클을 돈다. "이번 물결에 낼 것이 없었다"와
// "레인이 돌지 않았다"는 다른 뜻이고, 돌지 않은 레인은 관측에서 사라진다.
//
// shadow(a112 7.3.1)는 이 물결의 관문 앞 제안 묶음이다. 여기서는 record 로 넘기기만 한다 — 메서드 호출 · 순회 0(AST 핀 ④).
func (runtime *strategyLaneRuntime) evaluate(ctx context.Context, market StrategyMarket,
	activationGeneration uint64, promotion strategyrouter.FamilyActivation, inputs []strategyworker.Input, shadow strategyShadowBatch,
) error {
	if runtime == nil {
		return nil
	}
	// 순서가 계약이다. (기록에서 태어나는 것은 `productionStrategyLanes` 가 이미
	// 했다.) 증거가 있으면 다시 태어나고 → 돌고 → 잠긴 것을 남긴다. 복구를 사이클
	// 뒤로 미루면 증거가 이미 도착한 레인이 한 주기를 더 잠긴 채로 보낸다.
	if err := runtime.recoverMarketLanes(ctx, market, activationGeneration); err != nil {
		return err
	}
	lanes := runtime.lanesFor(market)
	// a112 7.5 D1: 레인 넷은 **동시에** 돌고 여기서 join 된다(시장 주기 안 — 분리된 레인 goroutine 금지, Manager 조건 ①).
	//
	// 왜: 순차로 돌면 앞 레인의 멈춤이 뒤 레인을 그 마감 시한만큼 세운다 — 설계 「느린 worker 때문에 다른 worker 가 기다리지 않게」와
	// 스펙 「한 instance 의 wait · timeout 이 peer 의 evaluation cycle 을 바꾸지 않는다」를 어긴다. 동시에 돌면 이웃은 자기 지연으로
	// 끝나고, 시장 주기의 지연은 합이 아니라 최댓값(각 레인이 RunBounded 마감 시한으로 끊기므로 ≤ 마감 시한 1 회)이다.
	//
	// 안전한 이유: 레인은 서로 상태를 공유하지 않고(Lane 주석 — 이웃 레인을 가리키는 필드 0), goroutine 하나가 레인 하나만 돌며, 관측은
	// 자기 색인 칸에만 쓴다(레인 순서 그대로). 버려진 사이클의 step goroutine 수명은 strategyworker.invokeBounded 가 정한다 — step 이
	// 돌아올 때까지 살고(순수 메모리 step 이라 곧 끝남), 비정상으로 레인이 즉시 잠기므로 레인당 최대 하나다(review 「7.5」 조건 ②).
	//
	// panic: 레인 goroutine 안의 panic(step 밖 — step 의 panic 은 invokeStep 이 이미 실패로 바꾼다)을 삼키지 않고 join 뒤 이 goroutine
	// 에서 다시 던진다. 순차였을 때와 같이 시장 주기의 회복 경로(invokeStrategyCycle)가 받게 하려는 것이다 — 다른 goroutine 의 panic 은
	// 그 경로가 잡을 수 없어 프로세스를 끝낸다.
	observations := make([]strategyLaneObservation, len(lanes))
	panics := make([]any, len(lanes))
	var join sync.WaitGroup
	for index, lane := range lanes {
		input := strategyworker.Input{}
		for _, candidate := range inputs {
			if lane.Owns(candidate.Proposal) {
				input = candidate
				break
			}
		}
		join.Add(1)
		go func(index int, lane *strategyworker.Lane, input strategyworker.Input) {
			defer join.Done()
			defer func() { panics[index] = recover() }()
			observations[index] = runtime.runLane(ctx, lane, promotion, input)
		}(index, lane, input)
	}
	join.Wait()
	for _, recovered := range panics {
		if recovered != nil {
			panic(recovered)
		}
	}
	runtime.record(market, observations, shadow, promotion)
	// 남기지 못한 잠금은 오류다. 조용히 넘기면 다음 재시작이 잠긴 레인을 열고,
	// 그것이 이 태스크가 없애려는 바로 그 동작이다.
	if err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {
		return err
	}
	// 마지막으로, 이 빌드에 없는 레인을 가리키는 기록이 남아 있으면 그것도
	// 오류다 — 사람은 잠갔다고 믿는데 런타임은 열려 있는 상태다.
	return runtime.staleLatchError(market)
}

// strategyLaneInputs 는 이 시장이 이번 물결에 세운 봉인된 제안을 레인 입력으로 바꾼다.
//
// 소유자 범위의 종목은 제안이 스스로 말한 계보가 아니라 **승인된 후보**에서
// 읽는다. 권한이 스스로 신고한 값으로 그 권한을 가리키면 어긋남을 잡으려던
// 자리가 언제나 참이 된다 — `coordinateMarketProposals` 가 같은 이유로 같은
// 선택을 한다.
func strategyLaneInputs(accountRef string, authority strategyProposalMarketAuthority) []strategyworker.Input {
	inputs := make([]strategyworker.Input, 0, len(authority.entries))
	for _, entry := range authority.entries {
		inputs = append(inputs, strategyworker.Input{
			Scope: strategyrouter.OwnerKey{AccountRef: accountRef,
				Market:             strategyRouterMarket(authority.market),
				Symbol:             entry.route.approved.Symbol(),
				PositionGeneration: entry.route.route.Request().Key.PositionGeneration},
			SnapshotDigest: entry.authority.SnapshotDigest(),
			Proposal:       strategyarbiter.Proposal{Result: entry.authority.Proposal(), Authority: entry.route.route},
		})
	}
	return inputs
}

// runLane 은 레인 하나의 유계 사이클 한 번이다.
//
// 투입을 먼저 넣는 이유: 레인의 관문(`RunBounded`)은 투입이 없으면 사이클을
// 열지 않는다. 넣지 못했다면 그 이유가 곧 이번 주기의 관측이다 — 잠겨서
// 못 받았는지(DISABLED), 칸이 차서 버렸는지(FULL)는 운영자가 할 조치가 다르다.
func (runtime *strategyLaneRuntime) runLane(ctx context.Context, lane *strategyworker.Lane,
	promotion strategyrouter.FamilyActivation, input strategyworker.Input,
) strategyLaneObservation {
	// a112 7.3: 투영이 읽을 관측 시점 값 — 활성화가 이 레인 자기 열쇠에 대해 말한 상태와 이번 입력의 digest. 레인 상태는 건드리지 않는다
	// (Desired · Effective 는 worker 값 위임, 입력은 이미 받은 값).
	observation := strategyLaneObservation{Key: lane.Key(), Desired: lane.Desired(promotion), Effective: lane.Effective(promotion),
		SnapshotDigest: input.SnapshotDigest, EvidenceDigest: strategyLaneEvidenceDigest(input)}
	observation.Trigger = lane.Offer()
	if observation.Trigger != strategyworker.TriggerEnqueued {
		observation.Health = lane.Health()
		return observation
	}
	// 레인 안에서 도는 일은 laneStepFor 하나다(a112 7.5 D1). 생산 빌드(태그 없음)의 정의는 strategyFamilyLaneStep 그대로이고 seam 은
	// tossos_testseams 빌드에만 있다 — 생산 바이너리에 함수 필드를 두면 `*Context` 클로저가 레인 안으로 들어올 길이 열린다
	// (TestOnlyThePackageLevelStepEverRunsInsideALane 가 자리 · 본문 · 태그를 못 박음).
	bounded, start := lane.RunBounded(ctx, input, runtime.laneStepFor(lane, promotion))
	observation.Start = start
	observation.Outcome = bounded.Cycle.Outcome
	observation.Detail = bounded.Cycle.Detail
	observation.Refusal = bounded.Cycle.Refusal
	observation.Abnormal = bounded.Abnormal
	observation.Cancelled = bounded.Cancelled
	observation.Abandoned = bounded.Abandoned
	observation.Latched = bounded.Latched
	if bounded.Err != nil {
		observation.Failure = bounded.Err.Error()
	}
	observation.Health = lane.Health()
	observation.Emitted = bounded.Cycle.Outcome == strategyworker.OutcomeEmitted
	return observation
}

// record 는 이번 주기의 관측을 레인 열쇠별로 덮어쓴다.
//
// a112 7.3(판정 Q1=(B)): 이 시장의 물결 번호를 하나 올려 이번 관측들에 찍는다. 같은 잠금 안에서 올리고 찍어야 두 물결이 한 번호를
// 나눠 갖지 않는다. 관측이 없는 호출은 물결이 아니다(번호를 올리지 않음).
//
// a112 7.3.1(⑤): 같은 잠금 · 파도 증가 바로 뒤에 그 시장의 shadow 칸 {wave, batch, activation} 을 한 번에 덮어쓴다 — 그래서 「파도」 는
// 묶음을 실어 온 evaluate 의 파도 번호이고 묶음 · 파도 · 활성화는 구성으로 짝이다. 관측이 없는 호출은 칸도 건드리지 않는다.
func (runtime *strategyLaneRuntime) record(market StrategyMarket, observations []strategyLaneObservation, shadow strategyShadowBatch,
	activation strategyrouter.FamilyActivation,
) {
	if runtime == nil || len(observations) == 0 {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.waves == nil {
		runtime.waves = make(map[StrategyMarket]uint64, 2)
	}
	if runtime.waves[market] < ^uint64(0) {
		runtime.waves[market]++
	}
	runtime.shadowCells[market] = strategyShadowCell{wave: runtime.waves[market], batch: shadow, activation: activation}
	for _, observation := range observations {
		observation.Wave = runtime.waves[market]
		runtime.observed[observation.Key] = observation
	}
}

// observations 는 여덟 레인의 마지막 관측을 생산 목록 순서 그대로 돌려준다.
//
// 아직 한 번도 돌지 않은 레인도 자기 열쇠와 함께 나온다. 목록에서 빠지면
// "안 돌았다"가 "없다"로 보이고, 그러면 레인 하나가 조용히 사라진 것을
// 아무도 못 본다 — 5.4.3 이 고친 것이 정확히 그 혼동이다.
func (runtime *strategyLaneRuntime) observations() []strategyLaneObservation {
	if runtime == nil {
		return nil
	}
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	values := make([]strategyLaneObservation, 0, len(runtime.lanes))
	for _, lane := range runtime.lanes {
		key := lane.Key()
		observation, seen := runtime.observed[key]
		if !seen {
			observation = strategyLaneObservation{Key: key, Health: lane.Health()}
		}
		values = append(values, observation)
	}
	return values
}
