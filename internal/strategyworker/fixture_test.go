//go:build tossos_testseams

package strategyworker

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

// 재료는 strategycoordinator/fixture_test.go 와 같은 방식으로 만든다. 봉인된
// 제안을 손으로 지어낼 수는 없고, strategyflow 의 시험 seam 이 봉인까지 해 준다.
const (
	fixtureAccount  = "acct"
	fixtureCampaign = "campaign-KR"
	fixtureSymbol   = "000660"
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

// newFixture 는 주어진 가족들만 점수 행에 담은 KR 재료를 만든다.
//
// 어떤 가족을 **빼는** 것이 이 재료의 쓸모다. 가족 점수 행이 없으면
// strategyarbiter 가 가족을 유도하지 못하고, 그때 worker 가 제안을 자기 것으로
// 세우지 못하는지 볼 수 있다.
func newFixture(families ...strategyrouter.Family) fixture {
	scores := make([]strategyrouter.ProductionRouteFamilyScore, 0, len(families))
	for index, family := range families {
		lane := krFamilyLane[family]
		scores = append(scores, strategyrouter.ProductionRouteFamilyScore{Family: family, Horizon: lane.horizon,
			LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: uint32(1000 * (index + 1))})
	}
	return fixture{now: time.Date(2026, 9, 2, 1, 2, 3, 0, time.UTC), scores: scores}
}

func allKRFamilies() []strategyrouter.Family {
	return []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal,
		strategyrouter.FamilyWeeklyValue, strategyrouter.FamilyBreakoutRetest}
}

func (f fixture) scope(t *testing.T) strategyrouter.OwnerKey {
	t.Helper()
	key, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketKR, fixtureSymbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func (f fixture) descriptor(t *testing.T, market strategyrouter.Market, laneID string) strategyflow.Descriptor {
	t.Helper()
	for _, value := range strategyflow.Descriptors() {
		if value.Market == market && value.LaneID == laneID {
			return value
		}
	}
	t.Fatalf("no %s descriptor for lane %q", market, laneID)
	return strategyflow.Descriptor{}
}

func (f fixture) result(t *testing.T, laneID string) strategyflow.Result {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(f.descriptor(t, strategyrouter.MarketKR, laneID),
		fixtureAccount, fixtureSymbol, fixtureCampaign, 8, "100", "90", "120",
		f.now.Add(-time.Second), f.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func (f fixture) candidate(t *testing.T, laneID string) strategyrouter.Candidate {
	t.Helper()
	descriptor := f.descriptor(t, strategyrouter.MarketKR, laneID)
	lineage := f.result(t, laneID).Lineage
	return strategyrouter.Candidate{Horizon: descriptor.Horizon, LaneID: descriptor.LaneID,
		LaneVersion: descriptor.LaneVersion, EvidenceDigest: lineage.RouterEvidenceDigest,
		ConfigDigest: lineage.ConfigDigest}
}

func (f fixture) authority(t *testing.T, laneIDs ...string) strategyrouter.ProductionRouteAuthority {
	t.Helper()
	candidates := make([]strategyrouter.Candidate, 0, len(laneIDs))
	for _, laneID := range laneIDs {
		candidates = append(candidates, f.candidate(t, laneID))
	}
	request, err := strategyrouter.MultiCandidateRouteFixture(f.scope(t), f.now, candidates...)
	if err != nil {
		t.Fatal(err)
	}
	return strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, fixtureCalibration, f.scores)
}

// input 은 한 레인의 제안을 담은 사이클 입력이다.
func (f fixture) input(t *testing.T, laneID string, laneIDs ...string) Input {
	t.Helper()
	if len(laneIDs) == 0 {
		laneIDs = allKRLanes()
	}
	return Input{Scope: f.scope(t), SnapshotDigest: "sha256:snapshot-" + fixtureSymbol + "-" + laneID,
		Proposal: strategyarbiter.Proposal{Result: f.result(t, laneID), Authority: f.authority(t, laneIDs...)}}
}

func allKRLanes() []string {
	return []string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID,
		weeklyvaluelane.KRWeeklyLaneID, breakoutlane.KRLaneID}
}

// effective 는 생산 목록의 그 worker 를 켠 사본이다.
//
// 생산 진입점은 여덟 개를 언제나 OFF 로 준다. 켜는 seam 을 내보내면 그 약속이
// 권고가 되므로, 켜진 worker 는 이 패키지 안에서만 만들 수 있다.
func effective(t *testing.T, market strategyrouter.Market, family strategyrouter.Family) FamilyWorker {
	t.Helper()
	for _, worker := range ProductionWorkers() {
		if worker.Key().Market == market && worker.Key().Family == family {
			return newWorker(worker.Key(), worker.Horizon(),
				strategyrouter.StateOn, strategyrouter.StateOn, worker.Runtime())
		}
	}
	t.Fatalf("no production worker for %s/%s", market, family)
	return FamilyWorker{}
}
