package engine

import (
	"context"
	"sync"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
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
}

// strategyLaneRuntime 은 프로세스가 사는 내내 **같은** 여덟 레인이다.
//
// 왜 프로세스가 살아 있는 내내 같은 것이어야 하나. 레인의 latch 와 연속 실패
// 계수기는 기억이다. 새로 고침(refresh)마다 레인을 다시 만들면 잠긴 레인이
// 1 초 뒤에 열린 채로 돌아온다 — 잠근 이유는 그대로인데. 그래서 이 값은
// `*Context` 가 한 번만 만들고 계속 들고 있는다(`Context.strategyLanes`).
type strategyLaneRuntime struct {
	lanes []*strategyworker.Lane

	mu       sync.RWMutex
	observed map[strategyworker.Key]strategyLaneObservation
}

// newStrategyLaneRuntime 은 생산 레인 여덟을 세운다.
//
// 목록도 정책도 여기서 고르지 않는다. 여덟이 누구인지는 `strategyworker` 의
// 생산 진입점 하나가 정하고, 그 목록이 동결 골든과 같은지는 그 패키지의
// golden_contract_test.go 가 골든 파일을 직접 읽어 대조한다.
func newStrategyLaneRuntime(clk clock.Clock) *strategyLaneRuntime {
	if clk == nil {
		return nil
	}
	lanes := strategyworker.ProductionLanes(clk)
	return &strategyLaneRuntime{lanes: lanes,
		observed: make(map[strategyworker.Key]strategyLaneObservation, len(lanes))}
}

// productionStrategyLanes 는 이 프로세스의 여덟 레인을 돌려준다. 없으면 만든다.
//
// 게으른 이유는 시계가 첫 생산 주기와 함께 도착하기 때문이고, 공유하는 이유는
// 위에 적은 그대로다 — 레인의 잠금은 새로 고침보다 오래 살아야 한다.
// `strategyDispatchOwner` 가 같은 이유로 같은 모양이다.
func (c *Context) productionStrategyLanes(clk clock.Clock) *strategyLaneRuntime {
	if c == nil || clk == nil {
		return nil
	}
	c.strategyLanesMu.Lock()
	defer c.strategyLanesMu.Unlock()
	if c.strategyLanes == nil {
		c.strategyLanes = newStrategyLaneRuntime(clk)
	}
	return c.strategyLanes
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
// 오류를 절대 돌려주지 않는 이유: 평가 실패는 오류가 아니라 **거절**이고,
// 거절은 `Cycle.Outcome` 에 담긴다. 둘을 한 값에 담으면 정당한 거절이 레인의
// 연속 실패 계수기를 올려 결국 진입을 잠근다.
func strategyFamilyLaneStep(lane *strategyworker.Lane) strategyworker.Step {
	return func(_ context.Context, input strategyworker.Input) (strategyworker.Cycle, error) {
		return lane.Run(input), nil
	}
}

// evaluate 는 이 시장의 네 레인에 각자 자기 것인 봉인된 제안을 한 번씩 돌리고
// 그 관측을 기록한다.
//
// **돌려주는 값이 없다.** 앞선 판본은 관측과 봉투를 함께 돌려줬는데, 그러면
// 호출자가 그중 하나를 버릴 수 있고 버린 것을 아무도 못 본다. 같은 문제를
// dispatch 주기에서 `Deliver` 로 푼 것과 같은 답이다 — 무시할 수 있는 답을
// 호출자에게 주지 않는다. 결과는 이 런타임 안에 남고 observations 가 읽는다.
//
// 어느 제안이 어느 레인의 것인지는 여기서 판정하지 않고 레인에게 묻는다
// (`lane.Owns`). 같은 판정을 두 곳에 두면 운영자가 보는 진단이 갈리고, 여기
// 옮겨 적은 사본은 봉인을 먼저 보는 것을 잊기 쉽다.
//
// 자기 제안이 없는 레인도 사이클을 돈다. "이번 물결에 낼 것이 없었다"와
// "레인이 돌지 않았다"는 다른 뜻이고, 돌지 않은 레인은 관측에서 사라진다.
func (runtime *strategyLaneRuntime) evaluate(ctx context.Context, market StrategyMarket,
	inputs []strategyworker.Input,
) {
	if runtime == nil {
		return
	}
	lanes := runtime.lanesFor(market)
	observations := make([]strategyLaneObservation, 0, len(lanes))
	for _, lane := range lanes {
		input := strategyworker.Input{}
		for _, candidate := range inputs {
			if lane.Owns(candidate.Proposal) {
				input = candidate
				break
			}
		}
		observations = append(observations, runtime.runLane(ctx, lane, input))
	}
	runtime.record(observations)
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
	input strategyworker.Input,
) strategyLaneObservation {
	observation := strategyLaneObservation{Key: lane.Key(), Trigger: lane.Offer()}
	if observation.Trigger != strategyworker.TriggerEnqueued {
		observation.Health = lane.Health()
		return observation
	}
	bounded, start := lane.RunBounded(ctx, input, strategyFamilyLaneStep(lane))
	observation.Start = start
	observation.Outcome = bounded.Cycle.Outcome
	observation.Detail = bounded.Cycle.Detail
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
func (runtime *strategyLaneRuntime) record(observations []strategyLaneObservation) {
	if runtime == nil || len(observations) == 0 {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	for _, observation := range observations {
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
