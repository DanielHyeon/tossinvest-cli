// Package strategyworker 는 전략군 레인 하나를 맡는 생산 worker 다.
//
// 왜 별도 패키지인가. 오늘 생산 경로에서 "worker" 라고 부르는 것은
// `internal/app/engine` 의 `StrategyMarketWorker` 이고, 그 `Cycle` 은 `*Context`
// 를 담은 클로저다. `*Context` 는 Journal·Gateway·Guardian 을 들고 있으므로 그
// worker 는 원장을 쓰고 주문을 낼 수 있다. 스펙이 요구하는 것은 그 반대다 —
// "Worker dependency closure 에는 broker mutator, writable journal, Guardian
// issuer, activation/toggle writer 가 없어야 한다".
//
// 능력이 없다는 것은 주석으로 약속할 수 없다. import 목록이 지켜야 한다.
// 그래서 이 타입은 엔진 밖 자기 패키지에 있고, dependency_closure_test.go 가
// 직접 import 와 전이 의존 전체를 훑어 그 약속을 시험으로 만든다.
//
// 이 패키지는 아직 생산 배선이 없다(dormant). 설계가 정한 순서가 그것이다 —
// 여덟 worker 와 두 조정자를 먼저 dormant 로 세우고, 공유 dispatch 로 가는
// 길은 선결 조건이 닫힐 때까지 열지 않는다.
package strategyworker

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// Key 는 동결 골든 `worker_key_fields` 가 정한 정확히 네 필드의 열쇠다.
//
// 조정자 큐의 여덟 필드 열쇠(`strategycoordinator.Key`)와 다른 것이다. 그쪽은
// 봉투를 접기 위한 열쇠라 종목·세대·스냅샷까지 들어가고, 이쪽은 worker 를
// 가리키는 이름이라 종목이 없다. 두 열쇠를 하나로 합치면 worker 가 종목마다
// 생겨 여덟이라는 수가 깨진다.
type Key struct {
	Market      strategyrouter.Market
	Family      strategyrouter.Family
	LaneID      string
	LaneVersion string
}

// KeyFields 는 골든이 적은 네 필드 이름을 그 순서대로 돌려준다.
func KeyFields() []string {
	return []string{"market", "family", "lane_id", "lane_version"}
}

// Parts 는 이 열쇠의 네 값이다. KeyFields 와 길이·순서가 언제나 같다.
func (key Key) Parts() []any {
	return []any{key.Market, key.Family, key.LaneID, key.LaneVersion}
}

// Outcome 은 사이클 한 번의 결과 종류다.
//
// 세 가지를 **따로** 둔다. "봉투 없음" 하나로 뭉치면 잠든 worker 와 거절당한
// 제안이 같은 값이 되고, 그러면 한 레인이 조용히 빠진 것을 아무도 못 본다 —
// 5.4.3 이 고친 것이 정확히 그 혼동(사라진 제안과 없던 제안)이다.
type Outcome string

const (
	// OutcomeEmitted 는 조정자에 넣을 봉투가 나왔다는 뜻이다.
	OutcomeEmitted Outcome = "EMITTED"
	// OutcomeDormant 는 이 worker 가 아직 켜지지 않았다는 뜻이다. 거절이 아니다.
	OutcomeDormant Outcome = "DORMANT"
	// OutcomeRefused 는 이 worker 가 제안을 자기 것으로 세울 수 없었다는 뜻이다.
	OutcomeRefused Outcome = "REFUSED"
	// OutcomeLatched 는 이 레인의 신규 진입이 고장으로 잠겼다는 뜻이다.
	//
	// DORMANT 와 합치면 안 된다. 잠긴 레인을 "아직 안 켰다"로 읽으면 운영자는
	// 아무 조치도 하지 않는데, 실제로는 복구 증거가 필요한 상태다.
	OutcomeLatched Outcome = "LATCHED"
)

// 아래 두 상수는 진단이지 계약이 아니다. 계약인 거절 코드는 골든이 정한
// `refusal_enums.arbitration` 여섯 개뿐이고 그 이름은 strategyarbiter 가 갖고 있다.
const (
	DetailDormant     = "this worker's effective state is not ON"
	DetailNotThisLane = "the sealed proposal does not establish this worker's lane"
)

// FamilyWorker 는 (시장, 가족, 레인, 레인 버전) 하나를 맡는 worker 다.
//
// 상태는 골든이 정한 대로 desired/effective OFF, runtime UNOBSERVED 로 태어난다.
// 켜는 일은 이 타입의 일이 아니다 — 서명된 활성화 권한이 하는 일이고, 그때까지
// 모든 사이클은 DORMANT 다.
type FamilyWorker struct {
	key       Key
	horizon   strategyrouter.Horizon
	desired   strategyrouter.DesiredState
	effective strategyrouter.DesiredState
	runtime   strategyrouter.RuntimeState
}

func (worker FamilyWorker) Key() Key                             { return worker.key }
func (worker FamilyWorker) Horizon() strategyrouter.Horizon      { return worker.horizon }
func (worker FamilyWorker) Desired() strategyrouter.DesiredState { return worker.desired }
func (worker FamilyWorker) Effective() strategyrouter.DesiredState {
	return worker.effective
}
func (worker FamilyWorker) Runtime() strategyrouter.RuntimeState { return worker.runtime }

// newWorker 는 상태까지 지정해 worker 하나를 만든다.
//
// 내보내지 않는다. 켜진 worker 를 아무나 만들 수 있으면 OFF 기본값은 약속이
// 아니라 권고가 된다. 이 패키지의 시험만 켜진 worker 를 만들 수 있고, 생산
// 진입점은 아래 ProductionWorkers 하나뿐이다.
func newWorker(key Key, horizon strategyrouter.Horizon,
	desired, effective strategyrouter.DesiredState, runtime strategyrouter.RuntimeState,
) FamilyWorker {
	return FamilyWorker{key: key, horizon: horizon, desired: desired, effective: effective, runtime: runtime}
}

// ProductionWorkers 는 골든이 정한 여덟 worker 를 그 순서 그대로 돌려준다.
//
// 레인 ID 와 버전은 각 레인 패키지가 가진 상수에서 읽는다. 골든에서 문자열을
// 옮겨 적지 않는 이유는, 옮겨 적으면 생산 상수가 바뀌어도 이 목록은 조용히
// 옛 값을 들고 있기 때문이다. 두 곳이 같은지는 golden_contract_test.go 가
// 골든 파일을 직접 읽어 대조한다.
func ProductionWorkers() []FamilyWorker {
	return []FamilyWorker{
		dormant(strategyrouter.MarketKR, strategyrouter.FamilyContinuation,
			continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort),
		dormant(strategyrouter.MarketUS, strategyrouter.FamilyContinuation,
			continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort),
		dormant(strategyrouter.MarketKR, strategyrouter.FamilyReversal,
			reversallane.KRReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort),
		dormant(strategyrouter.MarketUS, strategyrouter.FamilyReversal,
			reversallane.USReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort),
		dormant(strategyrouter.MarketKR, strategyrouter.FamilyWeeklyValue,
			weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly),
		dormant(strategyrouter.MarketUS, strategyrouter.FamilyWeeklyValue,
			weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly),
		dormant(strategyrouter.MarketKR, strategyrouter.FamilyBreakoutRetest,
			breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort),
		dormant(strategyrouter.MarketUS, strategyrouter.FamilyBreakoutRetest,
			breakoutlane.USLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort),
	}
}

func dormant(market strategyrouter.Market, family strategyrouter.Family,
	laneID, laneVersion string, horizon strategyrouter.Horizon,
) FamilyWorker {
	return newWorker(Key{Market: market, Family: family, LaneID: laneID, LaneVersion: laneVersion},
		horizon, strategyrouter.StateOff, strategyrouter.StateOff, strategyrouter.RuntimeUnobserved)
}

// Input 은 사이클 한 번의 입력이다. 조정자 봉투가 필요로 하는 것과 정확히 같다.
type Input struct {
	Scope          strategyrouter.OwnerKey
	SnapshotDigest string
	Proposal       strategyarbiter.Proposal
}

// Cycle 은 사이클 한 번의 결과다.
type Cycle struct {
	Outcome  Outcome
	Envelope strategycoordinator.Envelope
	Refusal  strategyarbiter.Refusal
	Detail   string
}

// Run 은 봉인된 제안 하나를 조정자 봉투로 만든다.
//
// 여기서 하는 판정은 **하나**다: 이 제안이 이 worker 의 레인인가. 범위·봉인·
// 용량은 조정자가 이미 판정하므로 다시 세지 않는다 — 같은 판정을 두 곳에 두면
// 운영자가 보는 진단이 갈린다(조정자 Submit 이 가족을 읽는 자리에 같은 주석이 있다).
//
// 이 함수는 아무것도 바꾸지 않는다. 그 약속은 이 자리에서 확인할 수 없고
// 패키지의 import 폐포가 확인한다.
func (worker FamilyWorker) Run(input Input) Cycle {
	if worker.effective != strategyrouter.StateOn {
		return Cycle{Outcome: OutcomeDormant, Detail: DetailDormant}
	}
	if !worker.owns(input.Proposal) {
		return Cycle{Outcome: OutcomeRefused,
			Refusal: strategyarbiter.RefusalSealMismatch, Detail: DetailNotThisLane}
	}
	return Cycle{Outcome: OutcomeEmitted, Envelope: strategycoordinator.Envelope{
		Scope: input.Scope, SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal}}
}

// owns 는 봉인된 제안이 이 worker 의 레인인지 본다.
//
// 봉인부터 확인하는 이유: 계보는 봉인이 성립할 때만 값이다. 봉인을 안 보고
// 계보의 레인 이름만 읽으면, 위조된 계보가 "네 레인이야" 라고 말하는 것을
// 그대로 믿게 된다.
//
// 가족은 계보가 말하는 것을 받지 않고 봉인된 권한의 가족 점수 행에서 **유도**
// 한다. 제안이 스스로 신고한 가족으로 제안을 검사하면 그 검사는 언제나 참이다.
func (worker FamilyWorker) owns(proposal strategyarbiter.Proposal) bool {
	if !proposal.Result.ValidProposal() {
		return false
	}
	lineage := proposal.Result.Lineage
	return lineage.Market == worker.key.Market &&
		lineage.LaneID == worker.key.LaneID &&
		lineage.LaneVersion == worker.key.LaneVersion &&
		lineage.Horizon == worker.horizon &&
		strategyarbiter.ProposalFamily(proposal) == worker.key.Family
}
