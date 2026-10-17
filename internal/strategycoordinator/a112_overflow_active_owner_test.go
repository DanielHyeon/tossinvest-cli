//go:build tossos_testseams

package strategycoordinator

// a112 태스크 2.6.1 의 마지막 조각 — 「queue overflow cannot silently drop the active-owner scope」(2.x 대조 감사가 찾은 빈칸: 넘침 시험에
// 활성 소유자 범위가 든 fixture 가 없었다).
//
// 두 순서를 잰다. 어느 쪽이든 넘침은 **시장을 닫고 보고한다**(Overflow · 선택 0 · 버림 수) — 활성 소유자 범위만 빠진 선택 목록으로 조용히
// 진행하는 결과가 없어야 한다.
//   ① 소유자 범위가 먼저 큐에 있고 다른 범위가 넘친다 → 소유자 봉투는 쫓겨나지 않는다(depth 그대로).
//   ② 큐가 다른 범위로 차 있을 때 소유자 범위가 온다 → 소유자 봉투는 거절되지만(넘침) 시장 전체가 닫혀 그 거절이 선택 목록 밖으로 새지 않는다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// ownerEnvelope 은 활성 소유자(주간 레인 · fixture 캠페인)를 가진 경로 권한 위의 봉투다.
func (f fixture) ownerEnvelope(t *testing.T, symbol string) Envelope {
	t.Helper()
	descriptor := f.descriptor(t, weeklyvaluelane.KRWeeklyLaneID)
	routed, err := strategyrouter.OwnedRouteFixture(f.key(t, symbol), descriptor.Horizon, descriptor.LaneID, descriptor.LaneVersion,
		fixtureCampaign, f.now)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, f.scores)
	return Envelope{Scope: f.key(t, symbol), SnapshotDigest: snapshotDigest(symbol, descriptor.LaneID),
		Proposal: strategyarbiter.Proposal{Result: f.result(t, symbol, descriptor.LaneID), Authority: authority}}
}

func TestOverflowNeverSilentlyDropsAnActiveOwnerScope(t *testing.T) {
	scores := map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 300_000, strategyrouter.FamilyReversal: 200_000,
		strategyrouter.FamilyWeeklyValue: 100_000}
	t.Run("owner queued first, another scope overflows", func(t *testing.T) {
		f := newFixture(t, scores)
		coordinator := newMarketCoordinator(strategyrouter.MarketKR, f.now, 2)
		if admission := coordinator.Submit(f.ownerEnvelope(t, fixtureSymbolFirst)); !admission.Admitted {
			t.Fatalf("arrangement: the active-owner envelope was not admitted: %+v", admission)
		}
		f.submit(t, coordinator, fixtureSymbolSecond, continuationlane.KRContinuationLaneID)
		if coordinator.Depth() != 2 {
			t.Fatalf("arrangement: depth=%d, want 2", coordinator.Depth())
		}
		overflow := f.submit(t, coordinator, fixtureSymbolSecond, reversallane.KRReversalLaneID)
		if overflow[0].Admitted || !overflow[0].Overflow {
			t.Fatalf("admission=%+v, want an overflow refusal", overflow[0])
		}
		if coordinator.Depth() != 2 {
			t.Fatalf("depth=%d — the overflow evicted something (the active-owner envelope must stay)", coordinator.Depth())
		}
		assertClosedByOverflow(t, coordinator.Arbitrate())
	})
	t.Run("queue full of other scopes, the owner arrives", func(t *testing.T) {
		f := newFixture(t, scores)
		coordinator := newMarketCoordinator(strategyrouter.MarketKR, f.now, 2)
		f.submit(t, coordinator, fixtureSymbolSecond, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
		admission := coordinator.Submit(f.ownerEnvelope(t, fixtureSymbolFirst))
		if admission.Admitted || !admission.Overflow {
			t.Fatalf("owner admission=%+v, want an overflow refusal", admission)
		}
		assertClosedByOverflow(t, coordinator.Arbitrate())
	})
}

// assertClosedByOverflow: 시장이 닫혔고(선택 0) 그 사유가 넘침이며 버림이 세어졌다 — 일부 범위만 고른 결과로 진행하지 않는다.
func assertClosedByOverflow(t *testing.T, outcome Outcome) {
	t.Helper()
	if !outcome.Overflow || !outcome.Closed() || len(outcome.Selections) != 0 || outcome.Drops == 0 {
		t.Fatalf("outcome=%+v, want an overflow-closed market with no selection and a counted drop", outcome)
	}
	if outcome.Refusal != strategyarbiter.RefusalNone {
		t.Fatalf("overflow reported as arbitration code %q", outcome.Refusal)
	}
}
