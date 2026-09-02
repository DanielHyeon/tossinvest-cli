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
	fixtureAccount = "acct"
	// 종목을 시장마다 따로 두는 이유는 리허설이 두 조정자를 **함께** 돌리기
	// 때문이다. 한 종목을 두 시장이 나눠 쓰면 소유자 범위가 겹쳐, KR 의 결함이
	// US 를 닫았는지 아니면 원래 같은 범위였는지 구별할 수 없다.
	fixtureSymbol   = "000660"
	fixtureSymbolUS = "AAPL"
	// 둘째 종목은 "시장 하나에 제안 하나"라는 옛 가정이 되살아나면 바로 깨지게
	// 한다. 사전순으로 000660 이 005930 보다 앞선다.
	fixtureSymbolSecond = "005930"
)

var fixtureCalibration = strategyrouter.ProductionRouteCalibration{
	ScoreVersion: "arbitration-score:v1", CalibrationDigest: "sha256:calibration-v1"}

var usFamilyLane = map[strategyrouter.Family]struct {
	horizon strategyrouter.Horizon
	laneID  string
	version string
}{
	strategyrouter.FamilyContinuation:   {strategyrouter.HorizonShort, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1},
	strategyrouter.FamilyReversal:       {strategyrouter.HorizonShort, reversallane.USReversalLaneID, reversallane.LaneVersionV1},
	strategyrouter.FamilyWeeklyValue:    {strategyrouter.HorizonWeekly, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1},
	strategyrouter.FamilyBreakoutRetest: {strategyrouter.HorizonShort, breakoutlane.USLaneID, breakoutlane.LaneVersionV1},
}

// familyLane 은 그 시장의 가족→레인 표다.
func familyLane(market strategyrouter.Market) map[strategyrouter.Family]struct {
	horizon strategyrouter.Horizon
	laneID  string
	version string
} {
	if market == strategyrouter.MarketUS {
		return usFamilyLane
	}
	return krFamilyLane
}

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
	market strategyrouter.Market
	symbol string
	scores []strategyrouter.ProductionRouteFamilyScore
}

// newFixture 는 주어진 가족들만 점수 행에 담은 KR 재료를 만든다.
//
// 어떤 가족을 **빼는** 것이 이 재료의 쓸모다. 가족 점수 행이 없으면
// strategyarbiter 가 가족을 유도하지 못하고, 그때 worker 가 제안을 자기 것으로
// 세우지 못하는지 볼 수 있다.
func newFixture(families ...strategyrouter.Family) fixture {
	return newMarketFixture(strategyrouter.MarketKR, families...)
}

// newMarketFixture 는 시장을 골라 같은 재료를 만든다. 리허설(5.7)이 두 조정자를
// 함께 돌리려면 US 재료가 있어야 하고, 그전까지 이 패키지의 재료는 KR 뿐이었다.
func newMarketFixture(market strategyrouter.Market, families ...strategyrouter.Family) fixture {
	table := familyLane(market)
	scores := make([]strategyrouter.ProductionRouteFamilyScore, 0, len(families))
	for index, family := range families {
		lane := table[family]
		scores = append(scores, strategyrouter.ProductionRouteFamilyScore{Family: family, Horizon: lane.horizon,
			LaneID: lane.laneID, LaneVersion: lane.version, ScorePPM: uint32(1000 * (index + 1))})
	}
	symbol := fixtureSymbol
	if market == strategyrouter.MarketUS {
		symbol = fixtureSymbolUS
	}
	return fixture{now: time.Date(2026, 9, 2, 1, 2, 3, 0, time.UTC),
		market: market, symbol: symbol, scores: scores}
}

// withSymbol 은 같은 재료를 다른 종목으로 복사한다.
func (f fixture) withSymbol(symbol string) fixture {
	f.symbol = symbol
	return f
}

func allKRFamilies() []strategyrouter.Family {
	return []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal,
		strategyrouter.FamilyWeeklyValue, strategyrouter.FamilyBreakoutRetest}
}

func (f fixture) scope(t *testing.T) strategyrouter.OwnerKey {
	t.Helper()
	key, err := strategyrouter.NewOwnerKey(fixtureAccount, f.market, f.symbol, 1)
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
	result, err := strategyflow.AcceptedResultForAuthorityTest(f.descriptor(t, f.market, laneID),
		fixtureAccount, f.symbol, "campaign-"+string(f.market), 8, "100", "90", "120",
		f.now.Add(-time.Second), f.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func (f fixture) candidate(t *testing.T, laneID string) strategyrouter.Candidate {
	t.Helper()
	descriptor := f.descriptor(t, f.market, laneID)
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
		laneIDs = allLanes(f.market)
	}
	return Input{Scope: f.scope(t), SnapshotDigest: "sha256:snapshot-" + f.symbol + "-" + laneID,
		Proposal: strategyarbiter.Proposal{Result: f.result(t, laneID), Authority: f.authority(t, laneIDs...)}}
}

func allKRLanes() []string { return allLanes(strategyrouter.MarketKR) }

// allLanes 는 그 시장의 네 레인 ID 다.
func allLanes(market strategyrouter.Market) []string {
	table := familyLane(market)
	ids := make([]string, 0, len(table))
	for _, family := range allKRFamilies() {
		ids = append(ids, table[family].laneID)
	}
	return ids
}

func allMarkets() []strategyrouter.Market {
	return []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS}
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
