//go:build tossos_testseams

package strategycoordinator

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

const (
	fixtureAccount  = "acct"
	fixtureCampaign = "campaign-KR"
	// 두 종목을 쓰는 이유는 "시장 하나에 제안 하나"라는 옛 가정이 되살아나면
	// 바로 깨지게 하기 위해서다. 사전순으로 000660 이 005930 보다 앞선다.
	fixtureSymbolFirst  = "000660"
	fixtureSymbolSecond = "005930"
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

type fixture struct {
	now    time.Time
	scores []strategyrouter.ProductionRouteFamilyScore
}

// newFixture 는 가족별 점수를 고정한 KR 재료를 만든다.
func newFixture(t *testing.T, ppm map[strategyrouter.Family]uint32) fixture {
	t.Helper()
	scores := make([]strategyrouter.ProductionRouteFamilyScore, 0, len(ppm))
	for family, score := range ppm {
		lane := krFamilyLane[family]
		scores = append(scores, strategyrouter.ProductionRouteFamilyScore{Family: family, Horizon: lane.horizon,
			LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: score})
	}
	return fixture{now: time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC), scores: scores}
}

func (f fixture) key(t *testing.T, symbol string) strategyrouter.OwnerKey {
	t.Helper()
	key, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketKR, symbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func (f fixture) descriptor(t *testing.T, laneID string) strategyflow.Descriptor {
	t.Helper()
	for _, value := range strategyflow.Descriptors() {
		if value.Market == strategyrouter.MarketKR && value.LaneID == laneID {
			return value
		}
	}
	t.Fatalf("no KR descriptor for lane %q", laneID)
	return strategyflow.Descriptor{}
}

func (f fixture) result(t *testing.T, symbol, laneID string) strategyflow.Result {
	t.Helper()
	return f.resultValidUntil(t, symbol, laneID, f.now.Add(time.Minute))
}

// resultValidUntil 은 후보 유효기한만 달리한 제안이다. 기한이 계보 신원에
// 들어가므로, 기한이 다르면 같은 레인이라도 다른 스냅샷이 된다.
func (f fixture) resultValidUntil(t *testing.T, symbol, laneID string, validUntil time.Time) strategyflow.Result {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(f.descriptor(t, laneID), fixtureAccount, symbol,
		fixtureCampaign, 8, "100", "90", "120", f.now.Add(-time.Second), validUntil)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// candidate 는 그 레인의 자격 후보다. 증거·설정 다이제스트를 손으로 적지 않고
// 실제 제안의 계보에서 읽는다.
func (f fixture) candidate(t *testing.T, symbol, laneID string) strategyrouter.Candidate {
	t.Helper()
	descriptor := f.descriptor(t, laneID)
	lineage := f.result(t, symbol, laneID).Lineage
	return strategyrouter.Candidate{Horizon: descriptor.Horizon, LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion,
		EvidenceDigest: lineage.RouterEvidenceDigest, ConfigDigest: lineage.ConfigDigest}
}

func (f fixture) authority(t *testing.T, symbol string, laneIDs ...string) strategyrouter.ProductionRouteAuthority {
	t.Helper()
	candidates := make([]strategyrouter.Candidate, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		candidates = append(candidates, f.candidate(t, symbol, laneID))
	}
	request, err := strategyrouter.MultiCandidateRouteFixture(f.key(t, symbol), f.now, candidates...)
	if err != nil {
		t.Fatal(err)
	}
	return strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, fixtureCalibration, f.scores)
}

// snapshotDigest 는 그 레인이 재생한 스냅샷의 다이제스트를 흉내 낸다.
// 생산에서는 strategyproposal.ProductionAuthority.SnapshotDigest() 가 준다.
func snapshotDigest(symbol, laneID string) string {
	return "sha256:snapshot-" + symbol + "-" + laneID
}

// submit 은 한 종목의 여러 레인 제안을 모두 큐에 넣는다.
func (f fixture) submit(t *testing.T, coordinator *MarketCoordinator, symbol string, laneIDs ...string) []Admission {
	t.Helper()
	authority := f.authority(t, symbol, laneIDs...)
	admissions := make([]Admission, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		admissions = append(admissions, coordinator.Submit(Envelope{Scope: f.key(t, symbol),
			SnapshotDigest: snapshotDigest(symbol, laneID),
			Proposal:       strategyarbiter.Proposal{Result: f.result(t, symbol, laneID), Authority: authority}}))
	}
	return admissions
}

// allKRLanes 는 KR 네 레인 전부다.
func allKRLanes() []string {
	return []string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID,
		weeklyvaluelane.KRWeeklyLaneID, breakoutlane.KRLaneID}
}
