package strategyflow

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// 이 변경의 핵심은 "자격 있는 가족이 여럿일 수 있다"이다. 그래서 고르는 함수는
// 반드시 결정이 둘 이상인 상태에서 검사해야 한다. 하나짜리만 넣고 통과시키면
// "첫 번째만 본다"로 바꿔도 아무 검사가 안 걸린다.

func decisionForTest(laneID, laneVersion string, horizon strategyrouter.Horizon) strategyrouter.RouteDecision {
	return strategyrouter.RouteDecision{
		Key:         strategyrouter.OwnerKey{Market: strategyrouter.MarketKR},
		Horizon:     horizon,
		LaneID:      laneID,
		LaneVersion: laneVersion,
	}
}

func continuationDecisionForTest() strategyrouter.RouteDecision {
	return decisionForTest(continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort)
}

func reversalDecisionForTest() strategyrouter.RouteDecision {
	return decisionForTest(reversallane.KRReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort)
}

func weeklyDecisionForTest() strategyrouter.RouteDecision {
	return decisionForTest(weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly)
}

// 요청이 들고 온 가족이 목록의 첫 번째가 아니어도 그것을 찾아내야 한다.
func TestSelectRouteDecisionFindsTheRequestedLaneEvenWhenItIsNotFirst(t *testing.T) {
	routed := strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{
		reversalDecisionForTest(), weeklyDecisionForTest(), continuationDecisionForTest(),
	}}
	decision, descriptor, matched, found := selectRouteDecision(routed, Request{Lane: LaneInput{kind: laneContinuationKR}})
	if !found || !matched {
		t.Fatalf("세 번째에 있는 요청 가족을 못 찾았다: matched=%v found=%v", matched, found)
	}
	if decision.LaneID != continuationlane.KRContinuationLaneID || descriptor.LaneID != continuationlane.KRContinuationLaneID {
		t.Fatalf("다른 가족을 골랐다: decision=%q descriptor=%q", decision.LaneID, descriptor.LaneID)
	}
}

// 정본에 없는 결정은 건너뛰고 계속 봐야 한다. 첫 항목에서 멈추면 안 된다.
func TestSelectRouteDecisionSkipsUnknownDecisionsAndKeepsLooking(t *testing.T) {
	routed := strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{
		decisionForTest("kr_short_not_a_real_lane", "v1", strategyrouter.HorizonShort),
		reversalDecisionForTest(),
	}}
	decision, _, matched, found := selectRouteDecision(routed, Request{Lane: LaneInput{kind: laneReversalKR}})
	if !found || !matched || decision.LaneID != reversallane.KRReversalLaneID {
		t.Fatalf("정본에 없는 첫 항목 뒤의 가족을 못 찾았다: matched=%v found=%v lane=%q", matched, found, decision.LaneID)
	}
}

// 요청한 가족이 목록에 없으면 고르지 않았다고(matched=false) 알리되,
// 거절 계보를 적을 수 있도록 첫 번째 정본 결정을 함께 돌려준다.
func TestSelectRouteDecisionReportsNoMatchButKeepsTheFirstCanonicalDecision(t *testing.T) {
	routed := strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{
		reversalDecisionForTest(), continuationDecisionForTest(),
	}}
	decision, descriptor, matched, found := selectRouteDecision(routed, Request{Lane: LaneInput{kind: laneWeeklyKR}})
	if matched {
		t.Fatal("요청한 가족이 없는데 골랐다고 했다")
	}
	if !found {
		t.Fatal("정본 결정이 있는데 하나도 못 찾았다고 했다")
	}
	if decision.LaneID != reversallane.KRReversalLaneID || descriptor.LaneID != reversallane.KRReversalLaneID {
		t.Fatalf("첫 번째 정본 결정이 아닌 것을 돌려줬다: decision=%q descriptor=%q", decision.LaneID, descriptor.LaneID)
	}
}

// 정본 결정이 하나도 없으면 아무것도 못 찾았다고 해야 한다.
func TestSelectRouteDecisionFindsNothingWhenNoDecisionIsCanonical(t *testing.T) {
	routed := strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{
		decisionForTest("kr_short_not_a_real_lane", "v1", strategyrouter.HorizonShort),
		decisionForTest(continuationlane.KRContinuationLaneID, "v-does-not-exist", strategyrouter.HorizonShort),
	}}
	if _, _, matched, found := selectRouteDecision(routed, Request{Lane: LaneInput{kind: laneContinuationKR}}); matched || found {
		t.Fatalf("정본이 하나도 없는데 찾았다고 했다: matched=%v found=%v", matched, found)
	}
}

// 빈 목록도 조용히 아무것도 못 찾았다고 해야 한다.
func TestSelectRouteDecisionFindsNothingWhenThereAreNoDecisions(t *testing.T) {
	if _, _, matched, found := selectRouteDecision(strategyrouter.RouteSetResult{}, Request{Lane: LaneInput{kind: laneContinuationKR}}); matched || found {
		t.Fatalf("결정이 없는데 찾았다고 했다: matched=%v found=%v", matched, found)
	}
}
