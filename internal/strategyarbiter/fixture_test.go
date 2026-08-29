//go:build tossos_testseams

package strategyarbiter

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

const (
	fixtureAccount  = "acct"
	fixtureSymbol   = "005930"
	fixtureCampaign = "campaign-KR"
)

var fixtureCalibration = strategyrouter.ProductionRouteCalibration{
	ScoreVersion: "arbitration-score:v1", CalibrationDigest: "sha256:calibration-v1"}

// krFamilyLane 은 KR 시장의 네 가족이 각각 어느 레인에 묶여 있는지다.
var krFamilyLane = map[strategyrouter.Family]struct {
	horizon strategyrouter.Horizon
	laneID  string
	version string
}{
	strategyrouter.FamilyContinuation:   {strategyrouter.HorizonShort, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1},
	strategyrouter.FamilyReversal:       {strategyrouter.HorizonShort, reversallane.KRReversalLaneID, reversallane.LaneVersionV1},
	strategyrouter.FamilyWeeklyValue:    {strategyrouter.HorizonWeekly, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1},
	strategyrouter.FamilyBreakoutRetest: {strategyrouter.HorizonShort, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1},
}

type krFixture struct {
	key    strategyrouter.OwnerKey
	now    time.Time
	scores []strategyrouter.ProductionRouteFamilyScore
}

// newKRFixture 는 주어진 가족별 점수로 KR 한 종목의 중재 재료를 만든다.
func newKRFixture(t *testing.T, ppm map[strategyrouter.Family]uint32) krFixture {
	t.Helper()
	key, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketKR, fixtureSymbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	scores := make([]strategyrouter.ProductionRouteFamilyScore, 0, len(ppm))
	for family, score := range ppm {
		lane := krFamilyLane[family]
		scores = append(scores, strategyrouter.ProductionRouteFamilyScore{Family: family, Horizon: lane.horizon,
			LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: score})
	}
	return krFixture{key: key, now: time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC), scores: scores}
}

func (fixture krFixture) descriptor(t *testing.T, laneID string) strategyflow.Descriptor {
	t.Helper()
	for _, value := range strategyflow.Descriptors() {
		if value.Market == strategyrouter.MarketKR && value.LaneID == laneID {
			return value
		}
	}
	t.Fatalf("no KR descriptor for lane %q", laneID)
	return strategyflow.Descriptor{}
}

func (fixture krFixture) result(t *testing.T, laneID string) strategyflow.Result {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(fixture.descriptor(t, laneID), fixtureAccount, fixtureSymbol,
		fixtureCampaign, 8, "100", "90", "120", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// candidate 는 그 레인의 자격 후보다. 증거·설정 다이제스트를 손으로 적지 않고
// 실제로 만들어질 제안의 계보에서 읽는다 — 값을 두 곳에 적으면 픽스처끼리
// 어긋나고, 그 어긋남을 잡으라고 있는 검사가 오히려 늘 실패한다.
func (fixture krFixture) candidate(t *testing.T, laneID string) strategyrouter.Candidate {
	t.Helper()
	descriptor := fixture.descriptor(t, laneID)
	lineage := fixture.result(t, laneID).Lineage
	return strategyrouter.Candidate{Horizon: descriptor.Horizon, LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion,
		EvidenceDigest: lineage.RouterEvidenceDigest, ConfigDigest: lineage.ConfigDigest}
}

// authority 는 주어진 레인들이 모두 자격을 갖춘 봉인된 경로 권한이다.
func (fixture krFixture) authority(t *testing.T, laneIDs ...string) strategyrouter.ProductionRouteAuthority {
	t.Helper()
	candidates := make([]strategyrouter.Candidate, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		candidates = append(candidates, fixture.candidate(t, laneID))
	}
	request, err := strategyrouter.MultiCandidateRouteFixture(fixture.key, fixture.now, candidates...)
	if err != nil {
		t.Fatal(err)
	}
	return strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, fixtureCalibration, fixture.scores)
}

// request 는 자격 있는 레인들이 각각 제안을 낸 중재 요청이다.
func (fixture krFixture) request(t *testing.T, laneIDs ...string) Request {
	t.Helper()
	authority := fixture.authority(t, laneIDs...)
	proposals := make([]Proposal, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		proposals = append(proposals, Proposal{Result: fixture.result(t, laneID), Authority: authority})
	}
	return fixture.scope(proposals)
}

// ownedRequest 는 활성 소유자가 이미 그 레인을 잡고 있는 중재 요청이다.
func (fixture krFixture) ownedRequest(t *testing.T, laneID string) Request {
	t.Helper()
	descriptor := fixture.descriptor(t, laneID)
	routed, err := strategyrouter.OwnedRouteFixture(fixture.key, descriptor.Horizon, descriptor.LaneID,
		descriptor.LaneVersion, fixtureCampaign, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, laneID), Authority: authority}})
}

func (fixture krFixture) scope(proposals []Proposal) Request {
	return Request{AccountRef: fixtureAccount, Market: strategyrouter.MarketKR, Symbol: fixtureSymbol,
		PositionGeneration: 1, ObservedAt: fixture.now, Proposals: proposals}
}

// bareAuthority 는 보정도 가족 점수도 붙지 않은 경로 권한이다.
// 매니페스트가 채점 기준을 싣지 않은 상태를 그대로 흉내 낸다.
func (fixture krFixture) bareAuthority(t *testing.T, laneID string) strategyrouter.ProductionRouteAuthority {
	t.Helper()
	descriptor := fixture.descriptor(t, laneID)
	authority, err := strategyrouter.ProductionRouteAuthorityForTest(fixture.key, descriptor.Horizon, descriptor.LaneID,
		descriptor.LaneVersion, "lane-evidence", "lane-config", fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	return authority
}

// ownedRequestWithCampaign 은 소유자가 잡은 캠페인을 따로 지정한다.
func (fixture krFixture) ownedRequestWithCampaign(t *testing.T, laneID, ownerCampaign string) Request {
	t.Helper()
	descriptor := fixture.descriptor(t, laneID)
	routed, err := strategyrouter.OwnedRouteFixture(fixture.key, descriptor.Horizon, descriptor.LaneID,
		descriptor.LaneVersion, ownerCampaign, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, laneID), Authority: authority}})
}

// requestWithScores 는 가족 점수 표를 직접 지정한 중재 요청이다.
func (fixture krFixture) requestWithScores(t *testing.T, scores []strategyrouter.ProductionRouteFamilyScore, laneIDs ...string) Request {
	t.Helper()
	candidates := make([]strategyrouter.Candidate, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		candidates = append(candidates, fixture.candidate(t, laneID))
	}
	routed, err := strategyrouter.MultiCandidateRouteFixture(fixture.key, fixture.now, candidates...)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, scores)
	proposals := make([]Proposal, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		proposals = append(proposals, Proposal{Result: fixture.result(t, laneID), Authority: authority})
	}
	return fixture.scope(proposals)
}

// crossScopeRequest 는 권한의 소유자 열쇠는 기대와 맞지만 제안의 계보만
// 다른 범위를 가리키는 요청이다. 두 값은 서로 다른 곳에서 오므로, 하나가
// 맞다고 해서 다른 하나가 맞다는 보장은 없다.
func (fixture krFixture) crossScopeRequest(t *testing.T, keyGeneration uint64, keySymbol,
	lineageAccount, lineageSymbol string, lineageMarket strategyrouter.Market,
) Request {
	t.Helper()
	key, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketKR, keySymbol, keyGeneration)
	if err != nil {
		t.Fatal(err)
	}
	laneID := continuationlane.KRContinuationLaneID
	if lineageMarket == strategyrouter.MarketUS {
		laneID = continuationlane.USContinuationLaneID
	}
	routed, err := strategyrouter.MultiCandidateRouteFixture(key, fixture.now,
		strategyrouter.Candidate{Horizon: strategyrouter.HorizonShort, LaneID: continuationlane.KRContinuationLaneID,
			LaneVersion: continuationlane.LaneVersionV1})
	if err != nil {
		t.Fatal(err)
	}
	var descriptor strategyflow.Descriptor
	for _, value := range strategyflow.Descriptors() {
		if value.LaneID == laneID {
			descriptor = value
		}
	}
	result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, lineageAccount, lineageSymbol,
		fixtureCampaign, 8, "100", "90", "120", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return Request{AccountRef: fixtureAccount, Market: strategyrouter.MarketKR, Symbol: keySymbol,
		PositionGeneration: keyGeneration, ObservedAt: fixture.now,
		Proposals: []Proposal{{Result: result, Authority: authority}}}
}

// crossAuthorityRequest 는 제안의 계보는 기대한 범위와 맞는데 그 제안을
// 재는 경로 권한만 다른 종목의 것인 요청이다. 남의 종목 자격 집합으로
// 내 종목 제안을 재면 레인 이름이 같다는 이유만으로 통과할 수 있다.
func (fixture krFixture) crossAuthorityRequest(t *testing.T, authoritySymbol string) Request {
	t.Helper()
	other, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketKR, authoritySymbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	routed, err := strategyrouter.MultiCandidateRouteFixture(other, fixture.now,
		fixture.candidate(t, continuationlane.KRContinuationLaneID))
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, continuationlane.KRContinuationLaneID), Authority: authority}})
}

// assertRefusal 는 계약 코드와 진단 문자열을 함께 확인한다. 코드만 보면
// 서로 다른 이유가 같은 코드로 뭉쳐 있어 무엇이 발화했는지 알 수 없다.
func assertRefusal(t *testing.T, outcome Outcome, code Refusal, detail string) {
	t.Helper()
	if outcome.Refusal != code || outcome.Detail != detail {
		t.Fatalf("refusal=%q detail=%q, want %q / %q", outcome.Refusal, outcome.Detail, code, detail)
	}
	if outcome.Selected != -1 {
		t.Fatalf("refused outcome still names index %d", outcome.Selected)
	}
}

// staleOwnerRequest 는 소유자 스냅샷 리비전이 기대와 어긋난 요청이다.
func (fixture krFixture) staleOwnerRequest(t *testing.T, laneID string) Request {
	t.Helper()
	descriptor := fixture.descriptor(t, laneID)
	routed, err := strategyrouter.StaleOwnerRouteFixture(fixture.key, descriptor.Horizon, descriptor.LaneID,
		descriptor.LaneVersion, fixtureCampaign, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, laneID), Authority: authority}})
}

// twoOwnersRequest 는 활성 소유자가 둘인 손상된 요청이다.
func (fixture krFixture) twoOwnersRequest(t *testing.T, firstLane, secondLane string) Request {
	t.Helper()
	owner := func(laneID, campaign string) strategyrouter.Owner {
		descriptor := fixture.descriptor(t, laneID)
		return strategyrouter.Owner{Horizon: descriptor.Horizon, LaneID: descriptor.LaneID,
			LaneVersion: descriptor.LaneVersion, CampaignID: campaign}
	}
	routed, err := strategyrouter.TwoActiveOwnersRouteFixture(fixture.key,
		owner(firstLane, fixtureCampaign), owner(secondLane, "campaign-SECOND"), fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, firstLane), Authority: authority}})
}

// mismatchedEvidenceRequest 는 자격 집합의 결정과 제안 계보의 증거·설정
// 다이제스트가 어긋난 요청이다.
func (fixture krFixture) mismatchedEvidenceRequest(t *testing.T, laneID string, evidence, config bool) Request {
	t.Helper()
	candidate := fixture.candidate(t, laneID)
	if evidence {
		candidate.EvidenceDigest = "sha256:" + strings.Repeat("e", 64)
	}
	if config {
		candidate.ConfigDigest = "sha256:" + strings.Repeat("c", 64)
	}
	routed, err := strategyrouter.MultiCandidateRouteFixture(fixture.key, fixture.now, candidate)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, laneID), Authority: authority}})
}

// staleSnapshotRequest 는 소유자 스냅샷의 신선도 창을 벗어난 시점에 평가된 요청이다.
func (fixture krFixture) staleSnapshotRequest(t *testing.T, laneID string) Request {
	t.Helper()
	routed, err := strategyrouter.MultiCandidateRouteFixture(fixture.key, fixture.now, fixture.candidate(t, laneID))
	if err != nil {
		t.Fatal(err)
	}
	routed.EvaluatedAt = routed.Snapshot.FreshUntil
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(routed, fixtureCalibration, fixture.scores)
	return fixture.scope([]Proposal{{Result: fixture.result(t, laneID), Authority: authority}})
}
