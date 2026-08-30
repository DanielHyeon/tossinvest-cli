//go:build tossos_testseams

package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 목록의 순서는 이제 경로가 온 순서가 아니라 조정자가 정한 소유자 범위 순서다.
//
// arbitrationRoutePair 는 005930 을 먼저, 000660 을 나중에 준다. 결과가
// 000660 먼저면 조정자가 순서를 정한 것이고, 005930 먼저면 경로 순서가
// 그대로 새어 나온 것이다. 이 순서가 ProposalSetDigest 에 그대로 들어가므로
// 순서가 흔들리면 같은 시장이 주기마다 다른 다이제스트를 낸다.
func TestEntriesComeBackInOwnerScopeOrderNotRouteOrder(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	pair := collectArbitrated(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID)
	if !pair.kr.snapshot.Ready || len(pair.kr.entries) != 2 {
		t.Fatalf("KR=%+v entries=%d", pair.kr.snapshot, len(pair.kr.entries))
	}
	got := []string{pair.kr.entries[0].route.approved.Symbol(), pair.kr.entries[1].route.approved.Symbol()}
	if got[0] != "000660" || got[1] != "005930" {
		t.Fatalf("entry order=%v, want [000660 005930] — 경로가 온 순서가 그대로 새어 나왔다", got)
	}
	if pair.kr.snapshot.QueueDropCount != 0 {
		t.Fatalf("drops=%d, want 0", pair.kr.snapshot.QueueDropCount)
	}
}

// 큐가 담을 수 있는 것보다 소유자 범위가 많으면 그 시장을 닫는다.
// 자리를 만들려고 하나를 조용히 버리면, 버려진 범위는 아무 기록도 남기지 않고
// 사라지고 나머지는 정상인 척 통과한다.
func TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	symbols := make([]string, 0, strategycoordinator.Capacity+1)
	for index := range strategycoordinator.Capacity + 1 {
		symbols = append(symbols, fmt.Sprintf("A%04d", index))
	}
	pair := collectOverflowing(t, now, symbols)
	if pair.kr.snapshot.Reason != StrategyProposalQueueOverflow {
		t.Fatalf("KR=%+v, want %q", pair.kr.snapshot, StrategyProposalQueueOverflow)
	}
	if pair.kr.snapshot.Ready || len(pair.kr.entries) != 0 {
		t.Fatalf("KR still ready=%v with %d entries", pair.kr.snapshot.Ready, len(pair.kr.entries))
	}
	if pair.kr.snapshot.QueueDropCount == 0 {
		t.Fatalf("drops=%d, want a counted drop — 조용한 유실은 금지다", pair.kr.snapshot.QueueDropCount)
	}
	// 큐 문제는 큐 이름으로 보고한다. 중재 계약 코드를 빌려 쓰지 않는다.
	if pair.kr.snapshot.ArbitrationRefusal != "" {
		t.Fatalf("overflow borrowed the arbitration code %q", pair.kr.snapshot.ArbitrationRefusal)
	}
}

// collectOverflowing 는 KR 에 종목을 원하는 수만큼 두고 각 종목이 지속형
// 한 레인으로만 제안하게 한다.
func collectOverflowing(t *testing.T, now time.Time, symbols []string) strategyProposalAuthorityPair {
	t.Helper()
	scores := familyScoresForTest(strategyrouter.MarketKR)
	evidenceDigest, configDigest := arbitrationLineageDigests(t, StrategyMarketKR, now)
	entries := make([]strategyRouteEntryAuthority, 0, len(symbols))
	for _, symbol := range symbols {
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
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		values := make(map[string][]strategyflow.Result, len(targets))
		for _, target := range targets {
			result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR),
				config.AccountRef, target.Approved.Symbol(), "campaign-KR", 8, "100", "90", "120",
				now.Add(-time.Second), now.Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			values[target.Approved.Symbol()] = []strategyflow.Result{result}
		}
		return strategyproposal.ProductionBatchAuthorityMultiLaneForTest(config.ManifestDigest, values), nil
	}
	return loader.collect(context.Background(), routeReadySchedulePair(now), routes, proposalFXPair(now))
}

// 고른 것을 되돌릴 자리를 못 찾으면 그 종목을 건너뛰지 않고 닫는다.
//
// 건너뛰면 목록만 조용히 짧아진다. 사라진 종목은 아무 기록도 남기지 않고,
// 남은 것들은 정상인 척 다음 관문을 통과한다. 그러면 "시장에 하나" 관문이
// 막으려던 것과 상관없는 종목을 대신 풀어 준다.
func TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList(t *testing.T) {
	arbitration := strategyMarketArbitration{
		outcome: strategycoordinator.Outcome{Selections: []strategycoordinator.Selection{
			{LineageIdentity: "어느 레인도 자기 것이라 하지 않는 계보"}}},
		byIdentity: map[string]strategyProposalEntryAuthority{},
	}
	entries, resolved := arbitration.entries()
	if resolved {
		t.Fatalf("resolved=true entries=%d — 되돌릴 자리가 없는데 통과했다", len(entries))
	}
	if entries != nil {
		t.Fatalf("entries=%v, want nil — 짧아진 목록을 내보내면 안 된다", entries)
	}
}
