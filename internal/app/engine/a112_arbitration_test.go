//go:build tossos_testseams

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

var arbitrationCalibrationForTest = strategyrouter.ProductionRouteCalibration{
	ScoreVersion: "arbitration-score:v1", CalibrationDigest: "sha256:calibration-v1"}

// arbitrationLineageDigests 는 제안 계보가 실제로 들고 오는 라우터 증거·설정
// 다이제스트다. 중재자는 자격 결정의 두 값과 이 두 값이 같은지 본다. 값을 손으로
// 적으면 두 픽스처가 어긋나고, 어긋남을 잡으라고 있는 검사가 늘 발화한다.
func arbitrationLineageDigests(t *testing.T, market StrategyMarket, now time.Time) (string, string) {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, market), "acct", "005930",
		"campaign-digest-probe", 8, "100", "90", "120", now.Add(-time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return result.Lineage.RouterEvidenceDigest, result.Lineage.ConfigDigest
}

// familyScoresForTest 는 한 시장의 네 가족 점수표다. 매니페스트가 싣는 것과
// 같은 모양이며, 기본 점수는 모두 다르게 두어 동률이 우연히 생기지 않게 한다.
func familyScoresForTest(market strategyrouter.Market) []strategyrouter.ProductionRouteFamilyScore {
	lanes := map[strategyrouter.Family]struct {
		horizon strategyrouter.Horizon
		laneID  string
		version string
		score   uint32
	}{
		strategyrouter.FamilyContinuation:   {strategyrouter.HorizonShort, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, 400_000},
		strategyrouter.FamilyReversal:       {strategyrouter.HorizonShort, reversallane.KRReversalLaneID, reversallane.LaneVersionV1, 900_000},
		strategyrouter.FamilyWeeklyValue:    {strategyrouter.HorizonWeekly, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, 300_000},
		strategyrouter.FamilyBreakoutRetest: {strategyrouter.HorizonShort, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, 700_000},
	}
	if market == strategyrouter.MarketUS {
		lanes = map[strategyrouter.Family]struct {
			horizon strategyrouter.Horizon
			laneID  string
			version string
			score   uint32
		}{
			strategyrouter.FamilyContinuation:   {strategyrouter.HorizonShort, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, 400_000},
			strategyrouter.FamilyReversal:       {strategyrouter.HorizonShort, reversallane.USReversalLaneID, reversallane.LaneVersionV1, 900_000},
			strategyrouter.FamilyWeeklyValue:    {strategyrouter.HorizonWeekly, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, 300_000},
			strategyrouter.FamilyBreakoutRetest: {strategyrouter.HorizonShort, breakoutlane.USLaneID, breakoutlane.LaneVersionV1, 700_000},
		}
	}
	scores := make([]strategyrouter.ProductionRouteFamilyScore, 0, len(lanes))
	for family, lane := range lanes {
		scores = append(scores, strategyrouter.ProductionRouteFamilyScore{Family: family, Horizon: lane.horizon,
			LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: lane.score})
	}
	return scores
}

// arbitrationRoutePair 는 KR 에 종목 두 개를 두고, 첫 종목만 여러 가족이
// 자격을 갖춘 경로 쌍이다. 두 종목을 두는 이유는 한 종목이 닫혔을 때
// 나머지 종목이 대신 풀리는지를 볼 수 있어야 하기 때문이다.
func arbitrationRoutePair(t *testing.T, now time.Time, scores []strategyrouter.ProductionRouteFamilyScore, multiSymbol string, multiLanes ...string) strategyRouteAuthorityPair {
	t.Helper()
	entry := func(market StrategyMarket, symbol string, laneIDs ...string) strategyRouteEntryAuthority {
		routerMarket, marketScores := strategyrouter.MarketKR, scores
		if market == StrategyMarketUS {
			// 미국 시장은 이 테스트가 조작하는 대상이 아니다. 자기 시장의
			// 정상 점수표를 준다 — 한쪽을 망가뜨리는 실험이 반대쪽까지
			// 망가뜨리면 "거절이 시장 안에 머문다"를 증명할 수 없다.
			routerMarket, marketScores = strategyrouter.MarketUS, familyScoresForTest(strategyrouter.MarketUS)
		}
		key, err := strategyrouter.NewOwnerKey("acct", routerMarket, symbol, 1)
		if err != nil {
			t.Fatal(err)
		}
		evidenceDigest, configDigest := arbitrationLineageDigests(t, market, now)
		candidates := make([]strategyrouter.Candidate, 0, len(laneIDs))
		for _, laneID := range laneIDs {
			candidates = append(candidates, strategyrouter.Candidate{Horizon: strategyrouter.HorizonShort,
				LaneID: laneID, LaneVersion: continuationlane.LaneVersionV1,
				EvidenceDigest: evidenceDigest, ConfigDigest: configDigest})
		}
		request, err := strategyrouter.MultiCandidateRouteFixture(key, now, candidates...)
		if err != nil {
			t.Fatal(err)
		}
		return strategyRouteEntryAuthority{approved: strategy.ApprovedSnapshotForTest(string(market), symbol, now),
			route: strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, arbitrationCalibrationForTest, marketScores)}
	}
	kr := strategyRouteMarketAuthority{market: StrategyMarketKR,
		entries: []strategyRouteEntryAuthority{entry(StrategyMarketKR, multiSymbol, multiLanes...),
			entry(StrategyMarketKR, "000660", continuationlane.KRContinuationLaneID)},
		snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketKR, Ready: true, Reason: StrategyRouteReady, RoutedCount: 2, ManifestDigest: "route-KR"}}
	us := strategyRouteMarketAuthority{market: StrategyMarketUS,
		entries:  []strategyRouteEntryAuthority{entry(StrategyMarketUS, "AAPL", continuationlane.USContinuationLaneID)},
		snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketUS, Ready: true, Reason: StrategyRouteReady, RoutedCount: 1, ManifestDigest: "route-US"}}
	return strategyRouteAuthorityPair{observedAt: now, kr: kr, us: us}
}

// arbitrationBatch 는 지정한 종목이 여러 레인으로 제안한 배치를 만든다.
func arbitrationBatch(t *testing.T, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget,
	now time.Time, multiSymbol string, multiLanes []string,
) strategyproposal.ProductionBatchAuthority {
	t.Helper()
	descriptorFor := func(laneID string) strategyflow.Descriptor {
		for _, descriptor := range strategyflow.Descriptors() {
			if descriptor.LaneID == laneID {
				return descriptor
			}
		}
		t.Fatalf("no descriptor for lane %q", laneID)
		return strategyflow.Descriptor{}
	}
	values := make(map[string][]strategyflow.Result, len(targets))
	for _, target := range targets {
		symbol := target.Approved.Symbol()
		lanes := []string{riskLoaderDescriptor(t, StrategyMarket(config.Market)).LaneID}
		if symbol == multiSymbol && config.Market == strategyrouter.MarketKR {
			lanes = multiLanes
		}
		for _, laneID := range lanes {
			result, err := strategyflow.AcceptedResultForAuthorityTest(descriptorFor(laneID), config.AccountRef, symbol,
				"campaign-"+string(config.Market), 8, "100", "90", "120", now.Add(-time.Second), now.Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			values[symbol] = append(values[symbol], result)
		}
	}
	return strategyproposal.ProductionBatchAuthorityMultiLaneForTest(config.ManifestDigest, values)
}

func collectArbitrated(t *testing.T, now time.Time, scores []strategyrouter.ProductionRouteFamilyScore,
	multiSymbol string, multiLanes ...string,
) strategyProposalAuthorityPair {
	t.Helper()
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		return arbitrationBatch(t, config, targets, now, multiSymbol, multiLanes), nil
	}
	return loader.collect(context.Background(), routeReadySchedulePair(now),
		arbitrationRoutePair(t, now, scores, multiSymbol, multiLanes...), proposalFXPair(now))
}

// 같은 종목에 세 가족이 제안하면 시장을 닫는 대신 점수가 높은 하나를 고른다.
//
// 승자를 일부러 레인 이름 순서의 *가운데*에 둔다. 목록은
// kr_short_absorption_reversal < kr_short_breakout_retest < kr_short_flow_continuation
// 순서로 오므로, 늘 첫 번째를 집거나 늘 마지막을 집는 구현은 여기서 걸린다.
func TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	scores := familyScoresForTest(strategyrouter.MarketKR)
	for index := range scores {
		switch scores[index].Family {
		case strategyrouter.FamilyBreakoutRetest:
			scores[index].ScorePPM = 900_000
		case strategyrouter.FamilyReversal:
			scores[index].ScorePPM = 500_000
		case strategyrouter.FamilyContinuation:
			scores[index].ScorePPM = 400_000
		}
	}
	pair := collectArbitrated(t, now, scores, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID, breakoutlane.KRLaneID)
	if !pair.kr.snapshot.Ready || pair.kr.snapshot.Reason != StrategyProposalReady {
		t.Fatalf("KR=%+v, want a ready market", pair.kr.snapshot)
	}
	if pair.kr.snapshot.ProposedCount != 2 || len(pair.kr.entries) != 2 {
		t.Fatalf("KR proposed=%d entries=%d, want one selected proposal per symbol", pair.kr.snapshot.ProposedCount, len(pair.kr.entries))
	}
	var multi strategyProposalEntryAuthority
	for _, entry := range pair.kr.entries {
		if entry.route.approved.Symbol() == "005930" {
			multi = entry
		}
	}
	if lane := multi.authority.Proposal().Lineage.LaneID; lane != breakoutlane.KRLaneID {
		t.Fatalf("selected lane=%q, want the highest-scored breakout lane", lane)
	}
}

// 중재가 닫히면 그 종목만 빼는 것이 아니라 시장 전체가 닫힌다.
// 하나만 빼면 남은 목록이 하나가 되어 아래 관문이 오히려 만족된다.
func TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	tied := familyScoresForTest(strategyrouter.MarketKR)
	for index := range tied {
		if tied[index].Family == strategyrouter.FamilyReversal {
			tied[index].ScorePPM = 400_000 // 지속형과 같은 점수 — 최고점 동률
		}
	}
	pair := collectArbitrated(t, now, tied, "005930", continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if pair.kr.snapshot.Ready || pair.kr.snapshot.Reason != StrategyProposalArbitrationRefused {
		t.Fatalf("KR=%+v, want the whole market closed", pair.kr.snapshot)
	}
	if pair.kr.snapshot.ArbitrationRefusal != string(strategyarbiter.RefusalTie) ||
		pair.kr.snapshot.ArbitrationDetail != strategyarbiter.DetailTie {
		t.Fatalf("KR refusal=%q detail=%q, want %q / %q", pair.kr.snapshot.ArbitrationRefusal,
			pair.kr.snapshot.ArbitrationDetail, strategyarbiter.RefusalTie, strategyarbiter.DetailTie)
	}
	if len(pair.kr.entries) != 0 || pair.ResultAuthority().kr.ready {
		t.Fatalf("KR still carries %d entries and ready=%v", len(pair.kr.entries), pair.ResultAuthority().kr.ready)
	}
	// 다른 시장은 그대로 진행한다 — 거절은 시장 안에 머문다.
	if !pair.us.snapshot.Ready || pair.us.snapshot.ProposedCount != 1 {
		t.Fatalf("US=%+v, want the other market untouched", pair.us.snapshot)
	}
}

// 승인된 채점 기준이 없으면 제안이 하나뿐인 종목도 내보내지 않는다.
func TestAnUncalibratedMarketRefusesEvenASingleProposal(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	pair := collectArbitrated(t, now, nil, "005930", continuationlane.KRContinuationLaneID)
	if pair.kr.snapshot.Ready || pair.kr.snapshot.ArbitrationRefusal != string(strategyarbiter.RefusalUncalibrated) ||
		pair.kr.snapshot.ArbitrationDetail != strategyarbiter.DetailCalibration {
		t.Fatalf("KR=%+v, want %q / %q", pair.kr.snapshot, strategyarbiter.RefusalUncalibrated, strategyarbiter.DetailCalibration)
	}
}

// 승인된 후보와 경로 권한이 서로 다른 종목을 가리키면 그 제안은 나가지 못한다.
//
// 기대 범위의 종목을 경로 권한에서 읽으면 이 어긋남이 스스로를 정당화한다 —
// 권한이 "나는 000660 이다"라고 말하고, 그 말로 자기를 검사하면 언제나 통과한다.
// 그래서 승인된 후보에서 읽는다.
func TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	const approvedSymbol, authoritySymbol = "005930", "000660"
	key, err := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, authoritySymbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	evidenceDigest, configDigest := arbitrationLineageDigests(t, StrategyMarketKR, now)
	request, err := strategyrouter.MultiCandidateRouteFixture(key, now, strategyrouter.Candidate{
		Horizon: strategyrouter.HorizonShort, LaneID: continuationlane.KRContinuationLaneID,
		LaneVersion: continuationlane.LaneVersionV1, EvidenceDigest: evidenceDigest, ConfigDigest: configDigest})
	if err != nil {
		t.Fatal(err)
	}
	entry := strategyRouteEntryAuthority{approved: strategy.ApprovedSnapshotForTest("KR", approvedSymbol, now),
		route: strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, arbitrationCalibrationForTest,
			familyScoresForTest(strategyrouter.MarketKR))}
	routes := strategyRouteAuthorityPair{observedAt: now,
		kr: strategyRouteMarketAuthority{market: StrategyMarketKR, entries: []strategyRouteEntryAuthority{entry},
			snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketKR, Ready: true, Reason: StrategyRouteReady, RoutedCount: 1, ManifestDigest: "route-KR"}},
		us: strategyRouteMarketAuthority{market: StrategyMarketUS, snapshot: StrategyRouteMarketSnapshot{Market: StrategyMarketUS, Reason: StrategyRouteNoCandidate}}}

	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, _ []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		// 제안은 권한이 말하는 종목(000660)의 계보를 달고 있지만, 승인된
		// 후보의 이름(005930) 아래 담긴다.
		result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR),
			config.AccountRef, authoritySymbol, "campaign-KR", 8, "100", "90", "120", now.Add(-time.Second), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		return strategyproposal.ProductionBatchAuthorityMultiLaneForTest(config.ManifestDigest,
			map[string][]strategyflow.Result{approvedSymbol: {result}}), nil
	}
	pair := loader.collect(context.Background(), routeReadySchedulePair(now), routes, proposalFXPair(now))
	if pair.kr.snapshot.Ready || pair.kr.snapshot.ArbitrationRefusal != string(strategyarbiter.RefusalSealMismatch) ||
		pair.kr.snapshot.ArbitrationDetail != strategyarbiter.DetailScope {
		t.Fatalf("KR=%+v, want %q / %q", pair.kr.snapshot, strategyarbiter.RefusalSealMismatch, strategyarbiter.DetailScope)
	}
}

// 제안을 하나도 내지 않은 종목은 중재 대상이 아니라 그냥 거절 수에 든다.
// 두 종목 다 그러면 목록이 비고 시장은 "받아들인 제안 없음"으로 닫힌다.
func TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, _ []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		return strategyproposal.ProductionBatchAuthorityMultiLaneForTest(config.ManifestDigest, nil), nil
	}
	pair := loader.collect(context.Background(), routeReadySchedulePair(now),
		arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID),
		proposalFXPair(now))
	if pair.kr.snapshot.Ready || pair.kr.snapshot.Reason != StrategyProposalNoAcceptedScope {
		t.Fatalf("KR=%+v, want NO_ACCEPTED_PROPOSAL", pair.kr.snapshot)
	}
	if pair.kr.snapshot.RefusedCount != 2 || pair.kr.snapshot.ArbitrationRefusal != "" {
		t.Fatalf("KR refused=%d arbitration=%q, want 2 refused with no arbitration refusal",
			pair.kr.snapshot.RefusedCount, pair.kr.snapshot.ArbitrationRefusal)
	}
}
