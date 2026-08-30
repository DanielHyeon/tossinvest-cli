//go:build tossos_testseams

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 한 종목의 제안이 고장으로 사라지면 그 종목만 빠지는 것이 아니라 시장이 닫힌다.
//
// 빠지기만 하면 목록이 둘에서 하나로 줄고, 줄어든 목록이 아래 파이프라인의
// `len(entries) != 1` 관문을 **오히려 만족시킨다.** 그래서 005930 의 제안이
// 사라진 덕분에 000660 이 풀린다 — 고장 하나가 시스템을 더 관대하게 만드는 것이다.
// 이것이 5.4.2 적대적 리뷰가 시연으로 보여 준 P1 이고, 태스크 5.4.3 이다.
func TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	pair := collectWithLostProposal(t, now, strategyproposal.ProductionAbsence{
		Symbol: "005930", LaneID: continuationlane.KRContinuationLaneID,
		Reason: strategyproposal.ProductionAbsenceEvidenceReplay})

	if pair.kr.snapshot.Ready {
		t.Fatalf("KR=%+v — 제안 하나가 고장으로 사라졌는데 시장이 열려 있다", pair.kr.snapshot)
	}
	if pair.kr.snapshot.Reason != StrategyProposalProductionFault {
		t.Fatalf("reason=%q, want %q", pair.kr.snapshot.Reason, StrategyProposalProductionFault)
	}
	if len(pair.kr.entries) != 0 {
		t.Fatalf("entries=%d — 닫힌 시장에서 항목이 새어 나왔다", len(pair.kr.entries))
	}
	// 어느 종목의 어느 레인이 사라졌는지가 남아야 운영자가 찾을 수 있다.
	if pair.kr.snapshot.ProductionFault == "" {
		t.Fatal("사라진 자리를 아무 데도 안 적었다")
	}
	// 닫힌 시장은 경로에 오른 전부가 거절이다.
	if pair.kr.snapshot.RefusedCount != pair.kr.snapshot.RoutedCount {
		t.Fatalf("refused=%d routed=%d", pair.kr.snapshot.RefusedCount, pair.kr.snapshot.RoutedCount)
	}
}

// collectWithLostProposal 은 두 종목이 경로에 오르고, 그중 000660 만 제안을 내며,
// 배치가 005930 의 고장을 들고 있는 상태를 만든다.
func collectWithLostProposal(t *testing.T, now time.Time,
	absence strategyproposal.ProductionAbsence,
) strategyProposalAuthorityPair {
	t.Helper()
	scores := familyScoresForTest(strategyrouter.MarketKR)
	evidenceDigest, configDigest := arbitrationLineageDigests(t, StrategyMarketKR, now)
	entries := make([]strategyRouteEntryAuthority, 0, 2)
	for _, symbol := range []string{"005930", "000660"} {
		key, err := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, symbol, 1)
		if err != nil {
			t.Fatal(err)
		}
		request, err := strategyrouter.MultiCandidateRouteFixture(key, now, strategyrouter.Candidate{
			Horizon: strategyrouter.HorizonShort, LaneID: continuationlane.KRContinuationLaneID,
			LaneVersion: continuationlane.LaneVersionV1, EvidenceDigest: evidenceDigest, ConfigDigest: configDigest})
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, strategyRouteEntryAuthority{
			approved: strategy.ApprovedSnapshotForTest("KR", symbol, now),
			route:    strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, arbitrationCalibrationForTest, scores)})
	}
	routes := strategyRouteAuthorityPair{observedAt: now,
		kr: strategyRouteMarketAuthority{market: StrategyMarketKR, entries: entries,
			snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketKR, Ready: true, Reason: StrategyRouteReady,
				RoutedCount: len(entries), ManifestDigest: "route-KR"}},
		us: strategyRouteMarketAuthority{market: StrategyMarketUS,
			snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketUS, Reason: StrategyRouteNoCandidate}}}

	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, _ []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		// 000660 만 제안을 낸다. 005930 의 제안은 고장으로 사라졌다.
		result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR),
			config.AccountRef, "000660", "campaign-KR", 8, "100", "90", "120", now.Add(-time.Second), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		return strategyproposal.ProductionBatchAuthorityWithFaultForTest(config.ManifestDigest,
			map[string][]strategyflow.Result{"000660": {result}}, absence), nil
	}
	return loader.collect(context.Background(), routeReadySchedulePair(now), routes, proposalFXPair(now))
}
