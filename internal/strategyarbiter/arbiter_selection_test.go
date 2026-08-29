//go:build tossos_testseams

package strategyarbiter

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// 세 짧은 가족이 같은 KR 종목을 제안하면 최고점 하나만 나간다.
func TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation:   400_000,
		strategyrouter.FamilyReversal:       900_000,
		strategyrouter.FamilyBreakoutRetest: 700_000,
	})
	request := fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID, breakoutlane.KRLaneID)
	outcome := Arbitrate(request)
	if outcome.Refusal != RefusalNone {
		t.Fatalf("refusal=%q, want a selection", outcome.Refusal)
	}
	if outcome.Family != strategyrouter.FamilyReversal || outcome.ScorePPM != 900_000 {
		t.Fatalf("selected family=%q score=%d, want REVERSAL/900000", outcome.Family, outcome.ScorePPM)
	}
	selected := request.Proposals[outcome.Selected].Result
	if selected.Lineage.LaneID != reversallane.KRReversalLaneID {
		t.Fatalf("selected lane=%q, want the reversal lane", selected.Lineage.LaneID)
	}
	if outcome.LineageIdentity != selected.Lineage.Identity || outcome.LineageIdentity == "" {
		t.Fatalf("outcome lineage %q does not name the selected proposal", outcome.LineageIdentity)
	}
	if outcome.ExistingOwner {
		t.Fatal("outcome claims an existing owner where the snapshot has none")
	}
}

// 최고점이 동률이면 임의로 고르지 않고 닫는다.
func TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation: 600_000,
		strategyrouter.FamilyReversal:     600_000,
	})
	assertRefusal(t, Arbitrate(fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)),
		RefusalTie, DetailTie)
}

// 동률이 최고점이 아니라 아래쪽이면 최고점 하나는 여전히 유일하다.
//
// 동률 둘을 승자보다 *앞*에 둔다. 레인 이름 순서가
// kr_short_absorption_reversal < kr_short_breakout_retest < kr_short_flow_continuation
// 이므로 아래 배점은 [300k, 300k, 800k] 순서로 들어가고, 동률 수를 승자에서 1로
// 되돌리는 경로가 실제로 돈다. 승자를 앞에 두면 그 경로는 한 번도 실행되지 않는다.
func TestATieBelowTheTopStillLeavesAUniqueWinner(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyReversal:       300_000,
		strategyrouter.FamilyBreakoutRetest: 300_000,
		strategyrouter.FamilyContinuation:   800_000,
	})
	outcome := Arbitrate(fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID, breakoutlane.KRLaneID))
	if outcome.Refusal != RefusalNone || outcome.Family != strategyrouter.FamilyContinuation {
		t.Fatalf("refusal=%q family=%q, want the unique continuation winner", outcome.Refusal, outcome.Family)
	}
	if outcome.ScorePPM != 800_000 {
		t.Fatalf("score=%d, want 800000", outcome.ScorePPM)
	}
}

// 제안이 하나뿐이어도 승인된 채점 권한이 없으면 내보내지 않는다.
func TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 500_000})
	if outcome := Arbitrate(fixture.request(t, continuationlane.KRContinuationLaneID)); outcome.Refusal != RefusalNone {
		t.Fatalf("calibrated singleton refusal=%q, want a selection", outcome.Refusal)
	}
	bare := fixture.request(t, continuationlane.KRContinuationLaneID)
	bare.Proposals[0].Authority = fixture.bareAuthority(t, continuationlane.KRContinuationLaneID)
	assertRefusal(t, Arbitrate(bare), RefusalUncalibrated, DetailCalibration)
}

// 두 제안의 채점 버전이 다르면 견줄 근거가 없다.
func TestProposalsUnderDifferentScoreVersionsAreIncomparable(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation: 400_000,
		strategyrouter.FamilyReversal:     900_000,
	})
	for name, other := range map[string]strategyrouter.ProductionRouteCalibration{
		"score version":      {ScoreVersion: "arbitration-score:v2", CalibrationDigest: fixtureCalibration.CalibrationDigest},
		"calibration digest": {ScoreVersion: fixtureCalibration.ScoreVersion, CalibrationDigest: "sha256:calibration-v2"},
	} {
		request := fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
		request.Proposals[1].Authority = strategyrouter.ProductionRouteAuthorityFromRequestForTest(
			request.Proposals[1].Authority.Request(), other, request.Proposals[1].Authority.FamilyScores())
		t.Run(name, func(t *testing.T) {
			assertRefusal(t, Arbitrate(request), RefusalUncalibrated, DetailCalibration)
		})
	}
}

// 활성 주간가치 소유자가 있으면 더 높은 점수의 돌파 제안도 소유자를 갈아치우지 못한다.
func TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyWeeklyValue:    100_000,
		strategyrouter.FamilyBreakoutRetest: 999_999,
	})
	owned := fixture.ownedRequest(t, weeklyvaluelane.KRWeeklyLaneID)
	outcome := Arbitrate(owned)
	if outcome.Refusal != RefusalNone || outcome.Family != strategyrouter.FamilyWeeklyValue {
		t.Fatalf("owner outcome refusal=%q family=%q, want the weekly owner kept", outcome.Refusal, outcome.Family)
	}
	if !outcome.ExistingOwner || outcome.ScorePPM != 100_000 {
		t.Fatalf("outcome=%+v does not record the preserved owner", outcome)
	}
	owned.Proposals = append(owned.Proposals, Proposal{Result: fixture.result(t, breakoutlane.KRLaneID),
		Authority: owned.Proposals[0].Authority})
	assertRefusal(t, Arbitrate(owned), RefusalMultipleOwner, DetailOwnerLane)
}

// 소유자가 잡은 캠페인과 다른 캠페인의 제안은 소유자를 이어받지 못한다.
func TestAProposalFromAnotherCampaignDoesNotInheritTheOwner(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyWeeklyValue: 100_000})
	owned := fixture.ownedRequestWithCampaign(t, weeklyvaluelane.KRWeeklyLaneID, "campaign-OTHER")
	assertRefusal(t, Arbitrate(owned), RefusalMultipleOwner, DetailOwnerLane)
}

// 봉인된 자격 집합에 없는 레인의 제안은 비교 대상이 아니다.
func TestALaneOutsideTheSealedEligibleSetIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation: 400_000,
		strategyrouter.FamilyReversal:     900_000,
	})
	request := fixture.request(t, continuationlane.KRContinuationLaneID)
	request.Proposals = append(request.Proposals, Proposal{Result: fixture.result(t, reversallane.KRReversalLaneID),
		Authority: request.Proposals[0].Authority})
	assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailIneligibleLane)
}

// 후보 유효 기한이 지난 제안은 이 범위 전체를 닫는다.
func TestAStaleProposalClosesTheWholeScope(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation: 400_000,
		strategyrouter.FamilyReversal:     900_000,
	})
	request := fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	request.ObservedAt = time.Unix(0, request.Proposals[0].Result.Lineage.CandidateValidUntilNS).UTC()
	assertRefusal(t, Arbitrate(request), RefusalStaleEnvelope, DetailStaleCandidate)
	request.ObservedAt = time.Unix(0, request.Proposals[0].Result.Lineage.CandidateValidUntilNS-1).UTC()
	if outcome := Arbitrate(request); outcome.Refusal != RefusalNone {
		t.Fatalf("one nanosecond before expiry refusal=%q, want a selection", outcome.Refusal)
	}
}

// 어느 가족 점수 행에도 붙지 않는 레인은 모르는 가족이다.
func TestALaneWithNoFamilyScoreRowIsUnknown(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyReversal: 400_000})
	request := fixture.request(t, continuationlane.KRContinuationLaneID)
	assertRefusal(t, Arbitrate(request), RefusalUncalibrated, DetailUnknownFamily)
}

// 봉인된 뒤 값이 바뀐 제안은 거절한다.
func TestAProposalMutatedAfterSealingIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	request := fixture.request(t, continuationlane.KRContinuationLaneID)
	request.Proposals[0].Result.Quantity++
	assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailProposalSeal)
}

// 기대한 것과 다른 종목·계좌·세대의 제안이 섞여 들어오면 거절한다.
func TestAProposalOutsideTheExpectedScopeIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	for name, mutate := range map[string]func(*Request){
		"symbol":     func(r *Request) { r.Symbol = "000660" },
		"account":    func(r *Request) { r.AccountRef = "other" },
		"generation": func(r *Request) { r.PositionGeneration = 2 },
	} {
		request := fixture.request(t, continuationlane.KRContinuationLaneID)
		mutate(&request)
		t.Run(name, func(t *testing.T) {
			assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailScope)
		})
	}
}

// 같은 레인이 두 번 들어오면 자기 자신과 겨루게 되므로 닫는다.
func TestTheSameLaneTwiceIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	request := fixture.request(t, continuationlane.KRContinuationLaneID)
	request.Proposals = append(request.Proposals, request.Proposals[0])
	assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailDuplicateLane)
}

// 상한을 넘는 점수는 승인된 범위 밖이라 견줄 수 없다.
func TestAScoreAboveTheApprovedCeilingIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: strategyrouter.ScorePPMMax})
	if outcome := Arbitrate(fixture.request(t, continuationlane.KRContinuationLaneID)); outcome.Refusal != RefusalNone {
		t.Fatalf("a score exactly at the ceiling refusal=%q, want a selection", outcome.Refusal)
	}
	over := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: strategyrouter.ScorePPMMax + 1})
	assertRefusal(t, Arbitrate(over.request(t, continuationlane.KRContinuationLaneID)), RefusalUncalibrated, DetailScoreCeiling)
}

// 서로 다른 자격 집합에서 온 제안을 한 범위로 묶어서는 안 된다.
func TestProposalsBoundToDifferentRouteSetsAreRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation: 400_000,
		strategyrouter.FamilyReversal:     900_000,
	})
	request := fixture.request(t, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	request.Proposals[1].Authority = fixture.authority(t, reversallane.KRReversalLaneID)
	assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailRouteSetDisagreement)
}

// 열거에 없는 가족 이름을 단 점수 행으로는 아무것도 견줄 수 없다.
func TestAScoreRowNamingAnUnapprovedFamilyIsUnknown(t *testing.T) {
	fixture := newKRFixture(t, nil)
	lane := krFamilyLane[strategyrouter.FamilyContinuation]
	rows := []strategyrouter.ProductionRouteFamilyScore{{Family: "MOMENTUM", Horizon: lane.horizon,
		LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: 400_000}}
	request := fixture.requestWithScores(t, rows, continuationlane.KRContinuationLaneID)
	assertRefusal(t, Arbitrate(request), RefusalUncalibrated, DetailUnknownFamily)
}

// 한 레인에 점수 행이 둘 붙어 있으면 어느 쪽이 그 레인의 점수인지 알 수 없다.
func TestALaneMatchingTwoScoreRowsIsUnknown(t *testing.T) {
	fixture := newKRFixture(t, nil)
	lane := krFamilyLane[strategyrouter.FamilyContinuation]
	rows := []strategyrouter.ProductionRouteFamilyScore{
		{Family: strategyrouter.FamilyContinuation, Horizon: lane.horizon, LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: 400_000},
		{Family: strategyrouter.FamilyReversal, Horizon: lane.horizon, LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: 900_000},
	}
	request := fixture.requestWithScores(t, rows, continuationlane.KRContinuationLaneID)
	assertRefusal(t, Arbitrate(request), RefusalUncalibrated, DetailUnknownFamily)
}

// 권한의 소유자 열쇠는 맞는데 제안의 계보만 다른 범위를 가리키면 거절한다.
func TestAProposalWhoseLineageLeavesTheScopeIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	cases := map[string]Request{
		"generation": fixture.crossScopeRequest(t, 2, fixtureSymbol, fixtureAccount, fixtureSymbol, strategyrouter.MarketKR),
		"symbol":     fixture.crossScopeRequest(t, 1, fixtureSymbol, fixtureAccount, "000660", strategyrouter.MarketKR),
		"account":    fixture.crossScopeRequest(t, 1, fixtureSymbol, "other-account", fixtureSymbol, strategyrouter.MarketKR),
		"market":     fixture.crossScopeRequest(t, 1, "AAPL", fixtureAccount, "AAPL", strategyrouter.MarketUS),
	}
	for name, request := range cases {
		t.Run(name, func(t *testing.T) {
			assertRefusal(t, Arbitrate(request), RefusalSealMismatch, DetailScope)
		})
	}
}

// 다른 종목의 경로 권한으로 이 종목의 제안을 재서는 안 된다.
func TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	assertRefusal(t, Arbitrate(fixture.crossAuthorityRequest(t, "000660")), RefusalSealMismatch, DetailScope)
}

// 활성 소유자의 레인이 가족 점수표에 없으면 소유자라도 이어 가지 못한다.
func TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	owned := fixture.ownedRequest(t, weeklyvaluelane.KRWeeklyLaneID)
	assertRefusal(t, Arbitrate(owned), RefusalUncalibrated, DetailUnknownFamily)
}

// 소유자 스냅샷 리비전이 기대와 어긋나면 골든이 요구하는 대로 STALE_OWNER 다.
func TestAStaleOwnerRevisionIsItsOwnRefusal(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyWeeklyValue: 100_000})
	assertRefusal(t, Arbitrate(fixture.staleOwnerRequest(t, weeklyvaluelane.KRWeeklyLaneID)),
		RefusalStaleOwner, DetailStaleOwnerRevision)
}

// 활성 소유자가 둘이면 골든이 요구하는 대로 MULTIPLE_OWNER 다.
// 둘 다 RouteSet 에서는 같은 코드로 돌아오므로, 코드를 되짚지 않고 스냅샷을 본다.
func TestTwoActiveOwnersAreItsOwnRefusal(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyWeeklyValue:  100_000,
		strategyrouter.FamilyContinuation: 400_000,
	})
	assertRefusal(t, Arbitrate(fixture.twoOwnersRequest(t, weeklyvaluelane.KRWeeklyLaneID, continuationlane.KRContinuationLaneID)),
		RefusalMultipleOwner, DetailMultipleOwners)
}

// 제안이 근거로 삼은 증거·설정이 지금 자격 집합의 결정과 다르면 거절한다.
//
// 레인 이름만 맞으면 통과시키면, 지금이 아닌 어떤 증거로 만들어진 옛 제안이
// 현재 권한을 타고 들어온다. Propose 는 그때의 결정에서 두 값을 계보에 박는다.
func TestAProposalBuiltFromOtherEvidenceIsRefused(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	if outcome := Arbitrate(fixture.mismatchedEvidenceRequest(t, continuationlane.KRContinuationLaneID, false, false)); outcome.Refusal != RefusalNone {
		t.Fatalf("matching digests refusal=%q, want a selection", outcome.Refusal)
	}
	for name, mismatch := range map[string][2]bool{
		"evidence digest": {true, false},
		"config digest":   {false, true},
		"both":            {true, true},
	} {
		t.Run(name, func(t *testing.T) {
			assertRefusal(t, Arbitrate(fixture.mismatchedEvidenceRequest(t, continuationlane.KRContinuationLaneID, mismatch[0], mismatch[1])),
				RefusalSealMismatch, DetailEvidenceBinding)
		})
	}
}

// 소유자 스냅샷이 신선도 창을 벗어났으면 그것도 묵은 소유자다.
func TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner(t *testing.T) {
	fixture := newKRFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyContinuation: 400_000})
	assertRefusal(t, Arbitrate(fixture.staleSnapshotRequest(t, continuationlane.KRContinuationLaneID)),
		RefusalStaleOwner, DetailOwnerSnapshotStale)
}
