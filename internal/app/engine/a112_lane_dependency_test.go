//go:build tossos_testseams

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 설계가 5.1.2 의 **선결 조건**으로 건 것을 값으로 확인한다:
// "여덟 FamilyWorker/2 MarketCoordinator 를 dormant/shadow 로 배선한다.
// existing shared dispatch handoff 는 prerequisite 가 닫힐 때까지 spy 로 막고"
// (design.md, Migration Plan 4).
//
// 여기서 쓰는 제안은 지어낸 것이 아니라 `loader.collect` 가 실제 조정자를
// 지나 세운 **봉인된 제안**이다. 그래야 "레인이 아무것도 안 했다"가 "줄 것이
// 없었다"와 구분된다 — 첫 판본은 dispatch 주기 fixture 를 썼는데, 그 경로의
// 경로 권한에는 가족 점수 행이 없어 여덟 중 아무도 제안을 자기 것으로
// 알아보지 못했다. 그러면 "주문 0 건"은 배선이 조용해서가 아니라 배선이
// 닿지 않아서다.

func laneOwnedInputs(t *testing.T) []strategyworker.Input {
	t.Helper()
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	pair := collectArbitrated(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID)
	inputs := strategyLaneInputs("acct", pair.forMarket(StrategyMarketKR))
	if len(inputs) == 0 {
		t.Fatal("조정을 지난 KR 제안이 0 건이다 — 이 파일의 시험은 아무것도 증명하지 않는다")
	}
	return inputs
}

func laneRuntimeForDependency(t *testing.T) *strategyLaneRuntime {
	t.Helper()
	runtime := newStrategyLaneRuntime(clock.NewFake(time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC)), nil, "")
	if runtime == nil {
		t.Fatal("생산 레인 런타임이 서지 않았다")
	}
	return runtime
}

// TestExactlyOneLaneOwnsEachSealedProposal 은 배선이 실제로 닿는다는 것을 본다.
//
// 아래 두 시험의 전제다. 아무 레인도 제안을 자기 것으로 알아보지 못하면
// "봉투 0 건"과 "게이트웨이 호출 0 건"은 둘 다 공허하게 참이 된다.
func TestExactlyOneLaneOwnsEachSealedProposal(t *testing.T) {
	runtime := laneRuntimeForDependency(t)
	for _, input := range laneOwnedInputs(t) {
		owners := make([]strategyworker.Key, 0, 1)
		for _, lane := range runtime.lanes {
			if lane.Owns(input.Proposal) {
				owners = append(owners, lane.Key())
			}
		}
		if len(owners) != 1 {
			t.Fatalf("%s 를 자기 것이라 한 레인=%v, want 정확히 하나",
				input.Proposal.Result.Lineage.LaneID, owners)
		}
		if owners[0].Market != strategyrouter.MarketKR {
			t.Fatalf("KR 제안을 %s 레인이 가져갔다", owners[0].Market)
		}
	}
}

// TestEveryLaneStaysDormantOnAProposalItActuallyOwns 는 강한 판본의 잠듦 시험이다.
//
// 빈 입력으로 DORMANT 를 받는 것은 쉽다 — 켜진 레인도 줄 것이 없으면 아무것도
// 안 낸다. 여기서는 여덟 중 하나가 **자기 것이라고 인정한** 제안을 주고도
// DORMANT 인지를 본다. 그것이 동결 골든의 `effective: OFF` 와 스펙의
// "Legacy 3-family approval 은 4-family activation 으로 자동 승격되어서는
// 안 된다 (MUST NOT)" 가 값으로 참이라는 뜻이다.
func TestEveryLaneStaysDormantOnAProposalItActuallyOwns(t *testing.T) {
	runtime := laneRuntimeForDependency(t)
	_ = runtime.evaluate(context.Background(), StrategyMarketKR, 0, laneOwnedInputs(t))
	admitted := 0
	for _, observation := range runtime.observations() {
		if observation.Key.Market != strategyrouter.MarketKR {
			continue
		}
		admitted++
		if observation.Outcome != strategyworker.OutcomeDormant {
			t.Fatalf("%v: 결과=%s, want DORMANT — 활성화 매니페스트 없이 레인이 켜졌다",
				observation.Key, observation.Outcome)
		}
		if observation.Emitted {
			t.Fatalf("%v: 잠든 레인이 봉투를 냈다 — 누군가 켜진 worker 를 주조했다", observation.Key)
		}
	}
	if admitted != 4 {
		t.Fatalf("돈 KR 레인=%d, want 4", admitted)
	}
}

// TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes 는 오늘 배선이 조용하다는
// 것을 값으로 남긴다.
//
// **무엇을 증명하지 못하는지 먼저 적는다.** 레인이 켜졌을 때 무엇을 만질 수
// 있는지는 여기서 알 수 없다 — 오늘 여덟은 전부 DORMANT 다. 그것을 지키는
// 것은 실행이 아니라 구조다: `TestOnlyThePackageLevelStepEverRunsInsideALane`
// 이 레인 안에서 도는 값의 목록을 패키지 전체에서 세어 얼리고,
// `strategyworker` 의 dependency_closure_test.go 가 그 값이 사는 패키지의
// `-deps`/`-deps-test` 를 훑어 broker mutator·writable journal·Guardian
// issuer 가 없다는 것을 시험으로 만든다.
func TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes(t *testing.T) {
	_, _, _, spy := pairedStrategyDispatchCycleFixture(t)
	runtime := laneRuntimeForDependency(t)
	_ = runtime.evaluate(context.Background(), StrategyMarketKR, 0, laneOwnedInputs(t))
	spy.mu.Lock()
	places, observed := len(spy.calls), len(spy.observed)
	spy.mu.Unlock()
	if places != 0 {
		t.Fatalf("레인 사이클 뒤 브로커 주문=%d 건, want 0", places)
	}
	if observed != 0 {
		t.Fatalf("레인 사이클이 게이트웨이를 %d 번 읽었다, want 0 — 레인은 게이트웨이를 들지 않는다", observed)
	}
}

// TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane 은 소유 판정이
// 계보가 스스로 하는 말이 아니라 **봉인**에 걸려 있다는 것을 본다.
//
// 봉인을 안 보고 계보의 레인 이름만 읽으면, 위조된 계보가 "네 레인이야" 라고
// 말하는 것을 그대로 믿게 된다. 여기서는 진짜 제안의 계보를 다른 레인 이름으로
// 바꿔 본다 — 봉인이 깨지므로 여덟 중 **아무도** 가져가지 않아야 한다.
func TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane(t *testing.T) {
	runtime := laneRuntimeForDependency(t)
	forged := laneOwnedInputs(t)[0].Proposal
	forged.Result.Lineage.LaneID = "kr_short_absorption_reversal_v1"
	if forged.Result.ValidProposal() {
		t.Fatal("계보를 바꿨는데 봉인이 그대로다 — 봉인이 계보를 덮지 않는다")
	}
	for _, lane := range runtime.lanes {
		if lane.Owns(forged) {
			t.Fatalf("%v 가 이름만 바꾼 계보를 자기 것으로 받았다", lane.Key())
		}
	}
}
