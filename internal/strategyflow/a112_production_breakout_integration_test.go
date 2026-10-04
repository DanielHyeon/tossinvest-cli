//go:build tossos_testseams

package strategyflow

// a112 4.5(-a · -b) 감사 보강(Manager 판정 2026-10-04): 생산 Evaluate · Propose(실 RouteSet 라우터 + 실 레지스트리)를 breakout KR · US 로도
// 돌려, 이 파일 옆의 여섯 레인 통합 시험(production_integration_test.go)과 합쳐 **8/8 서술자** 를 벽 아래까지 잰다. 벽(결정 49 —
// strategyproposal buildLaneInput 이 breakout 에 ErrBreakoutEvidenceUnavailable) 너머의 생산 적재는 이 시험의 범위가 아니다(이연 —
// audit-3.8-4.5.md). 계보는 정확히: 라우터 · 레인 증거 digest = 순수 코어 스냅숏 digest, 설정 digest = 순수 코어 설정 digest, 후보 증거는
// 따로, 권한 계수 0.

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestProductionEvaluateAndProposeCarryExactBreakoutLineageInBothMarkets(t *testing.T) {
	evaluatedAt := time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC)
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			laneID, wrap, coreMarket := breakoutlane.KRLaneID, BreakoutKR, breakoutlane.MarketKR
			if market == strategyrouter.MarketUS {
				laneID, wrap, coreMarket = breakoutlane.USLaneID, BreakoutUS, breakoutlane.MarketUS
			}
			descriptor := descriptorByID(t, laneID)
			approved := approvedFixture(t, market)
			key, err := strategyrouter.NewOwnerKey("acct", market, approved.Symbol(), 1)
			if err != nil {
				t.Fatal(err)
			}
			request := breakoutFixtureRequest(t, coreMarket)
			request.CandidateID, request.PositionGeneration = approved.CandidateLifeID(), key.PositionGeneration
			snapshot, err := breakoutlane.NewEvidenceSnapshot(request.Evidence)
			if err != nil {
				t.Fatal(err)
			}
			core := breakoutlane.Evaluate(snapshot, nil)
			routerRequest, err := strategyrouter.StrategyflowRouteFixture(key, descriptor.Horizon, laneID, breakoutlane.LaneVersionV1,
				core.SnapshotDigest(), core.ConfigDigest(), evaluatedAt)
			if err != nil {
				t.Fatal(err)
			}
			flowRequest := Request{Approved: approved, Router: routerRequest, Lane: wrap(request)}

			result := Evaluate(flowRequest)
			if result.Code != RefusalNone || result.Quantity == 0 || !result.Lineage.Complete || !result.Lineage.Valid() {
				t.Fatalf("production breakout evaluate refused: %+v", result)
			}
			lineage := result.Lineage
			if lineage.Market != market || lineage.RouterID != strategyrouter.RouterID || lineage.RouterRelease != strategyrouter.RouterRelease ||
				lineage.Horizon != descriptor.Horizon || lineage.LaneID != laneID || lineage.LaneVersion != breakoutlane.LaneVersionV1 ||
				lineage.LaneRelease != descriptor.Release || lineage.RouterEvidenceDigest != core.SnapshotDigest() ||
				lineage.LaneEvidenceDigest != core.SnapshotDigest() || lineage.ConfigDigest != core.ConfigDigest() ||
				lineage.CandidateEvidenceDigest != approved.EvidenceDigest() || lineage.CampaignID != request.CampaignID ||
				lineage.LegOrdinal != request.LegOrdinal || lineage.RiskBudgetDigest != request.RiskBudgetDigest {
				t.Fatalf("production breakout lineage mismatch: %+v", lineage)
			}
			if result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 || result.ValidProposal() {
				t.Fatalf("production pure breakout flow acquired authority or a proposal seal: %+v", result)
			}

			proposal := Propose(flowRequest)
			if proposal.Code != RefusalNone || !proposal.ValidProposal() || proposal.Quantity == 0 ||
				proposal.Lineage.LaneID != laneID || proposal.Lineage.LaneEvidenceDigest != core.SnapshotDigest() ||
				proposal.Lineage.CampaignID != lineage.CampaignID {
				t.Fatalf("sealed production breakout proposal mismatch: %+v", proposal)
			}
			// 결정 48: 제안은 상한 없는 q_candidate, 평가는 FinalCap — 둘이 같은 계보를 싣는다.
			if proposal.Quantity <= result.Quantity {
				t.Fatalf("proposal q_candidate=%d must exceed the capped evaluate quantity %d for this fixture", proposal.Quantity, result.Quantity)
			}
		})
	}
}
