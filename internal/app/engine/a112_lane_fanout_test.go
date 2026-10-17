//go:build tossos_testseams

package engine

// a112 태스크 7.5 C1 — 근거(evidence) fan-out 상한(Manager 판정 2026-10-01).
//
// 여덟 레인이 두 시장 주기에서 평가될 때 근거는 **한 물결의 조립 하나**에서 온다: 레인은 I/O 가 없고(strategyworker 폐포 시험) 입력은 그
// 조립의 봉인된 제안을 레인마다 **자기 것 하나만** 받는다. 그래서 근거 수집 수는 레인 수에 비례하지 않는다 — 1 초 창 안의 두 시장 주기 ·
// 여덟 레인은 새 물결을 하나도 만들지 않고(캐시 그대로), 제안 하나는 정확히 한 레인에게만 간다.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestEightLanesShareOneAuthorityWaveAndEachProposalReachesOneLane(t *testing.T) {
	cycle, proposals, j, _ := pairedStrategyDispatchCycleFixture(t)
	loader := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	clk := loader.clk
	// 레인이 제안을 「자기 것」으로 알아보려면 경로 권한에 그 레인의 봉인된 가족 점수 행이 있어야 한다(strategyarbiter.ProposalFamily —
	// 가족은 계보가 아니라 점수 행에서 유도). 생산 경로 권한은 그 행을 싣고 오므로, fixture 의 빈 경로에 같은 모양을 붙인다.
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		authority := proposals.forMarket(market)
		entries := append([]strategyProposalEntryAuthority(nil), authority.entries...)
		for index := range entries {
			lineage := entries[index].authority.Proposal().Lineage
			family, ok := strategyrouter.ProductionLaneFamily(lineage.Market, lineage.LaneID)
			if !ok {
				t.Fatalf("arrangement: %s lane %s has no production family", market, lineage.LaneID)
			}
			entries[index].route.route = strategyrouter.WithArbitrationScoresForTest(entries[index].route.route,
				strategyrouter.ProductionRouteCalibration{ScoreVersion: "arbitration-score:v1", CalibrationDigest: "sha256:calibration-v1"},
				[]strategyrouter.ProductionRouteFamilyScore{{Family: family, Horizon: lineage.Horizon, LaneID: lineage.LaneID,
					LaneVersion: lineage.LaneVersion, ScorePPM: 500000}})
		}
		authority.entries = entries
		if market == StrategyMarketKR {
			proposals.kr = authority
		} else {
			proposals.us = authority
		}
	}
	c := &Context{Journal: j, AccountRef: "acct-risk-loader"}
	// 이 물결의 조립(캐시) — dispatch 는 비운다: 이 시험은 근거가 레인에 닿는 길만 잰다(주문 경로는 다른 시험들의 몫).
	assembly := &StrategyEntryProductionAssembly{proposals: proposals, schedule: cycle.schedule}
	c.strategyRefresh, c.strategyRefreshAt = assembly, clk.Now()

	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		if err := c.runProductionStrategyMarketCycle(context.Background(), clk, market); err != nil {
			t.Fatalf("%s market cycle: %v", market, err)
		}
	}
	// 물결 0 개: 두 시장 주기 · 여덟 레인이 같은 조립 하나를 읽었다(새 수집이 있었다면 캐시 포인터나 진행 중 물결이 바뀐다).
	if c.strategyRefresh != assembly || c.strategyRefreshWave != nil || !c.strategyRefreshAt.Equal(clk.Now()) {
		t.Fatalf("eight lanes minted a new authority wave: cache=%p (want %p) inflight=%v", c.strategyRefresh, assembly, c.strategyRefreshWave)
	}
	lanes, err := c.productionStrategyLanes(context.Background(), clk)
	if err != nil {
		t.Fatal(err)
	}
	observations := lanes.observations()
	if len(observations) != 8 {
		t.Fatalf("observations=%d, want 8", len(observations))
	}
	fed := map[StrategyMarket]int{}
	for _, observation := range observations {
		if observation.Wave == 0 {
			t.Fatalf("lane %v was not evaluated in its market cycle", observation.Key)
		}
		if observation.SnapshotDigest != "" {
			fed[StrategyMarket(observation.Key.Market)]++
		}
	}
	// 시장마다 봉인된 제안 수 = 입력을 받은 레인 수(제안 하나 → 레인 하나, 복제 없음).
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		if want := len(proposals.forMarket(market).entries); fed[market] != want {
			t.Fatalf("%s: %d lanes received an input, want exactly one per sealed proposal (%d)", market, fed[market], want)
		}
	}
}
