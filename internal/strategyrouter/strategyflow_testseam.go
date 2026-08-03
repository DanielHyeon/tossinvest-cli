//go:build tossos_testseams

package strategyrouter

import "time"

// StrategyflowRouteFixture is compiled only under the repository's explicit
// test-seam build tag. It gives cross-package integration tests a real sealed
// router request without exporting an authority constructor in production.
func StrategyflowRouteFixture(key OwnerKey, horizon Horizon, laneID, laneVersion, evidenceDigest, configDigest string, evaluatedAt time.Time) (RouteRequest, error) {
	snapshot, err := newOwnerSnapshot(key, 1, "strategyflow-owner-fixture", evaluatedAt.Add(-time.Second), evaluatedAt.Add(time.Minute), nil)
	if err != nil {
		return RouteRequest{}, err
	}
	record, err := newMarketRecord(marketRecordInput{
		Market: key.Market, Desired: StateOn, Effective: StateOn, Revision: 2, LockID: "strategyflow-" + string(key.Market),
		CalendarGeneration: "strategyflow-calendar-v1", CalendarDigest: "strategyflow-calendar-digest", Timezone: map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[key.Market],
		SessionScope: "REGULAR", ActivationDigest: "strategyflow-human-approved-fixture", ActivationExpiresAt: evaluatedAt.Add(time.Hour),
		ConfigVersion: "strategyflow-config-v1", UpdatedActor: "strategyflow-integration-test", UpdatedAt: evaluatedAt.Add(-time.Minute), Runtime: RuntimeUnobserved,
	})
	if err != nil {
		return RouteRequest{}, err
	}
	return RouteRequest{Key: key, ExpectedOwnerRevision: 1, ExpectedMarketRevision: record.Revision, EvaluatedAt: evaluatedAt,
		Snapshot: snapshot, MarketRecord: record, Candidates: []Candidate{{Key: key, Horizon: horizon, LaneID: laneID, LaneVersion: laneVersion,
			Score: 1, Eligible: true, Desired: StateOn, Effective: StateOn, EvidenceDigest: evidenceDigest, ConfigDigest: configDigest}}}, nil
}
