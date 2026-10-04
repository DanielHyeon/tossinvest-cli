package strategyflow

// a112 3.8(-a · -b) 감사 보강(Manager 판정 2026-10-04): Toss 순위 · 거래량 · 흐름은 **발견 증거** 일 뿐이다 — 공식 마감 봉을 대신할 수 없다.
//
// 승인 후보(approvedFixture — 순위 백분위 · 상승폭 · 고가 거리 지표를 실은 실제 candidate 판정)가 있고, 라우터가 breakout 레인을 자격
// 있다고 답해도, breakout 레인 입력에 공식 봉 증거가 없으면 레인이 **타입 있는 거절** 을 낸다(spec breakout-retest-strategy-lane 「Toss
// rank 만 존재」). 후보 증거는 계보에 따로 남고 레인 증거 자리로 옮겨 가지 않는다. 대조(differential): 같은 후보 · 같은 라우팅에 공식 봉
// 증거(breakoutFixtureRequest)를 주면 레인 거절이 사라진다 — 거절의 원인이 「봉 증거 없음」 하나임을 보인다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestTossRankEvidenceAloneCannotStandInForOfficialClosedBars(t *testing.T) {
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			laneID, wrap, coreMarket := breakoutlane.KRLaneID, BreakoutKR, breakoutlane.MarketKR
			if market == strategyrouter.MarketUS {
				laneID, wrap, coreMarket = breakoutlane.USLaneID, BreakoutUS, breakoutlane.MarketUS
			}
			descriptor := descriptorByID(t, laneID)
			approved := approvedFixture(t, market)
			if approved.EvidenceDigest() == "" {
				t.Fatal("arrangement: the approved candidate carries no discovery evidence")
			}
			key, err := strategyrouter.NewOwnerKey("acct", market, approved.Symbol(), 7)
			if err != nil {
				t.Fatal(err)
			}
			evaluate := func(request BreakoutRequest, decision strategyrouter.RouteDecision) Result {
				route := func(strategyrouter.RouteRequest) strategyrouter.RouteSetResult {
					return strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{decision}}
				}
				return evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: wrap(request)}, route, proposalRegistry())
			}

			// 순위 증거뿐(공식 봉 없음): 레인이 이유를 들고 거절한다.
			refused := evaluate(BreakoutRequest{}, routeDecision(descriptor, key, approved))
			if refused.Code != RefusalLane || refused.NativeCode == "" {
				t.Fatalf("rank-only: code=%s native=%q, want a typed lane refusal", refused.Code, refused.NativeCode)
			}
			if refused.ValidProposal() || refused.Quantity != 0 {
				t.Fatalf("rank-only produced a proposal or a quantity: %+v", refused)
			}
			if refused.GuardianCalls != 0 || refused.BrokerCalls != 0 || refused.Mutations != 0 {
				t.Fatalf("rank-only touched an authority: guardian=%d broker=%d mutations=%d", refused.GuardianCalls, refused.BrokerCalls, refused.Mutations)
			}
			// 후보 증거는 그 자리에 남고 레인 증거 자리로 옮겨 가지 않는다.
			if refused.Lineage.CandidateEvidenceDigest != approved.EvidenceDigest() {
				t.Fatalf("candidate evidence digest=%q, want the approved candidate's %q", refused.Lineage.CandidateEvidenceDigest, approved.EvidenceDigest())
			}
			if refused.Lineage.LaneEvidenceDigest != "" {
				t.Fatalf("rank-only filled the lane evidence slot with %q", refused.Lineage.LaneEvidenceDigest)
			}

			// 대조: 같은 후보 · 같은 레인 라우팅에 공식 봉 증거를 주면 제안이 선다. 라우터의 증거 · 설정 digest 는 레인 증거(순수 코어의
			// 스냅숏 · 설정 digest)다 — 후보 digest 가 아니다(matchingLaneLineage 가 둘을 대조).
			request := breakoutFixtureRequest(t, coreMarket)
			request.CandidateID, request.PositionGeneration = approved.CandidateLifeID(), key.PositionGeneration
			snapshot, err := breakoutlane.NewEvidenceSnapshot(request.Evidence)
			if err != nil {
				t.Fatalf("arrangement: the bar evidence does not seal: %v", err)
			}
			core := breakoutlane.Evaluate(snapshot, nil)
			decision := routeDecision(descriptor, key, approved)
			decision.EvidenceDigest, decision.ConfigDigest = core.SnapshotDigest(), core.ConfigDigest()
			withBars := evaluate(request, decision)
			if withBars.Code != RefusalNone || withBars.Quantity == 0 {
				t.Fatalf("with official bars: code=%s native=%q quantity=%d — the differential does not isolate the missing bars", withBars.Code, withBars.NativeCode, withBars.Quantity)
			}
			if withBars.Lineage.LaneEvidenceDigest != core.SnapshotDigest() || withBars.Lineage.LaneEvidenceDigest == approved.EvidenceDigest() {
				t.Fatalf("lane evidence digest=%q, want the bar snapshot's %q (never the candidate's)", withBars.Lineage.LaneEvidenceDigest, core.SnapshotDigest())
			}
			if withBars.Lineage.CandidateEvidenceDigest != approved.EvidenceDigest() {
				t.Fatal("the candidate evidence digest was lost when bar evidence was present")
			}
		})
	}
}
