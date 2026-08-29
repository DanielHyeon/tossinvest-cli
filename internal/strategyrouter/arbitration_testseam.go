//go:build tossos_testseams

package strategyrouter

import (
	"sort"
	"time"
)

// WithArbitrationScoresForTest 는 봉인된 보정과 가족 점수를 권한에 붙인다.
//
// 봉인은 생산 경로가 쓰는 바로 그 함수로 다시 계산한다. 봉인을 흉내만 내면
// SealsValid 가 통과하지 않고, 그러면 테스트는 증명하려던 것과 다른 것을 증명한다.
func WithArbitrationScoresForTest(authority ProductionRouteAuthority, calibration ProductionRouteCalibration,
	scores []ProductionRouteFamilyScore,
) ProductionRouteAuthority {
	values := append([]ProductionRouteFamilyScore(nil), scores...)
	sort.Slice(values, func(i, j int) bool { return values[i].Family < values[j].Family })
	market := authority.request.Key.Market
	authority.calibration = calibration
	authority.scores = values
	authority.seals = ProductionRouteSeals{Family: productionRouteFamilySeal(market, values),
		Scoring: productionRouteScoringSeal(calibration, values), Calibration: productionRouteCalibrationSeal(market, calibration)}
	return authority
}

// ProductionRouteAuthorityFromRequestForTest 는 이미 봉인된 경로 요청을 그대로
// 안고 있는 권한을 만든다. 요청과 자격 집합이 어긋난 권한은 생산에서 생길 수
// 없으므로 테스트에서도 만들지 않는다.
func ProductionRouteAuthorityFromRequestForTest(request RouteRequest, calibration ProductionRouteCalibration,
	scores []ProductionRouteFamilyScore,
) ProductionRouteAuthority {
	return WithArbitrationScoresForTest(ProductionRouteAuthority{request: request,
		manifestDigest: "sha256:test-manifest", ownerDigest: request.Snapshot.Digest}, calibration, scores)
}

// MultiCandidateRouteFixture 는 같은 종목에 여러 가족이 자격을 갖춘 봉인된
// 경로 요청이다. 활성 소유자는 없다.
func MultiCandidateRouteFixture(key OwnerKey, evaluatedAt time.Time, candidates ...Candidate) (RouteRequest, error) {
	if len(candidates) == 0 {
		return RouteRequest{}, errNoFixtureCandidate
	}
	request, err := StrategyflowRouteFixture(key, candidates[0].Horizon, candidates[0].LaneID, candidates[0].LaneVersion,
		"lane-evidence", "lane-config", evaluatedAt)
	if err != nil {
		return RouteRequest{}, err
	}
	values := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Key = key
		candidate.Score = 1
		candidate.Eligible, candidate.Desired, candidate.Effective = true, StateOn, StateOn
		// 호출자가 준 증거·설정 다이제스트를 덮어쓰지 않는다. 중재자는 제안 계보의
		// 두 값이 이 결정의 두 값과 같은지 본다 — 픽스처가 임의 값으로 덮으면
		// 그 검사는 픽스처 사정 때문에 늘 실패하거나 늘 통과한다.
		if candidate.EvidenceDigest == "" {
			candidate.EvidenceDigest = "lane-evidence"
		}
		if candidate.ConfigDigest == "" {
			candidate.ConfigDigest = "lane-config"
		}
		values = append(values, candidate)
	}
	request.Candidates = values
	return request, nil
}

// OwnedRouteFixture 는 활성 소유자가 정확히 하나 있는 봉인된 경로 요청이다.
func OwnedRouteFixture(key OwnerKey, horizon Horizon, laneID, laneVersion, campaignID string, evaluatedAt time.Time) (RouteRequest, error) {
	request, err := StrategyflowRouteFixture(key, horizon, laneID, laneVersion, "lane-evidence", "lane-config", evaluatedAt)
	if err != nil {
		return RouteRequest{}, err
	}
	snapshot, err := newOwnerSnapshot(key, 1, "strategyflow-owner-fixture", evaluatedAt.Add(-time.Second), evaluatedAt.Add(time.Minute),
		[]Owner{{Key: key, Horizon: horizon, LaneID: laneID, LaneVersion: laneVersion, CampaignID: campaignID,
			Active: true, Desired: StateOn, Effective: StateOn}})
	if err != nil {
		return RouteRequest{}, err
	}
	request.Snapshot = snapshot
	return request, nil
}

var errNoFixtureCandidate = errFixture("strategyrouter: route fixture needs at least one candidate")

type errFixture string

func (e errFixture) Error() string { return string(e) }

// StaleOwnerRouteFixture 는 소유자 스냅샷 리비전이 기대와 어긋난 경로 요청이다.
func StaleOwnerRouteFixture(key OwnerKey, horizon Horizon, laneID, laneVersion, campaignID string, evaluatedAt time.Time) (RouteRequest, error) {
	request, err := OwnedRouteFixture(key, horizon, laneID, laneVersion, campaignID, evaluatedAt)
	if err != nil {
		return RouteRequest{}, err
	}
	request.ExpectedOwnerRevision = request.Snapshot.Revision + 1
	return request, nil
}

// TwoActiveOwnersRouteFixture 는 활성 소유자가 둘인 손상된 경로 요청이다.
func TwoActiveOwnersRouteFixture(key OwnerKey, first, second Owner, evaluatedAt time.Time) (RouteRequest, error) {
	request, err := StrategyflowRouteFixture(key, first.Horizon, first.LaneID, first.LaneVersion, "lane-evidence", "lane-config", evaluatedAt)
	if err != nil {
		return RouteRequest{}, err
	}
	first.Key, second.Key = key, key
	first.Active, second.Active = true, true
	first.Desired, first.Effective = StateOn, StateOn
	second.Desired, second.Effective = StateOn, StateOn
	snapshot, err := newOwnerSnapshot(key, 1, "strategyflow-owner-fixture", evaluatedAt.Add(-time.Second), evaluatedAt.Add(time.Minute),
		[]Owner{first, second})
	if err != nil {
		return RouteRequest{}, err
	}
	request.Snapshot = snapshot
	return request, nil
}
