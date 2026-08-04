package strategyrouter

import (
	"testing"
	"time"
)

var routerNow = time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)

func TestOwnerKeyExcludesHorizonAndCanonicalizesSymbol(t *testing.T) {
	shortKey, err := NewOwnerKey("acct", MarketUS, " aapl ", 7)
	if err != nil {
		t.Fatal(err)
	}
	weeklyKey, err := NewOwnerKey("acct", MarketUS, "AAPL", 7)
	if err != nil {
		t.Fatal(err)
	}
	if shortKey != weeklyKey || shortKey.Symbol != "AAPL" {
		t.Fatalf("horizon leaked into owner key or symbol was not canonical: %+v %+v", shortKey, weeklyKey)
	}
	if _, err := NewOwnerKey("", MarketKR, "005930", 1); err == nil {
		t.Fatal("empty account accepted")
	}
}

func TestExistingOwnerIsPreservedAcrossAllHorizons(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	owner := Owner{Key: key, Horizon: HorizonShort, LaneID: "kr-short", LaneVersion: "v1", CampaignID: "campaign-1", Active: true, Desired: StateOn, Effective: StateOn}
	snapshot := mustOwnerSnapshot(t, key, 4, []Owner{owner})
	request := validRouteRequest(t, key, 4, snapshot, []Candidate{
		validCandidate(key, HorizonWeekly, "kr-weekly", 999), validCandidate(key, HorizonShort, "kr-other-short", 1),
	})
	result := Route(request)
	if result.Code != RefusalNone || !result.Decision.ExistingOwner || result.Decision.LaneID != owner.LaneID || result.Decision.Horizon != HorizonShort {
		t.Fatalf("existing owner was stolen by scoring: %+v", result)
	}
}

func TestMultipleOwnersStaleSnapshotAndTieFailClosed(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 2)
	first := Owner{Key: key, Horizon: HorizonShort, LaneID: "us-short", LaneVersion: "v1", CampaignID: "campaign-1", Active: true, Desired: StateOn, Effective: StateOn}
	second := Owner{Key: key, Horizon: HorizonWeekly, LaneID: "us-weekly", LaneVersion: "v1", CampaignID: "campaign-2", Active: true, Desired: StateOn, Effective: StateOn}
	if got := Route(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, []Owner{first, second}), nil)); got.Code != RefusalReconstructionMismatch || got.Decision.LaneID != "" {
		t.Fatalf("multiple owners did not fail closed: %+v", got)
	}
	stale := mustOwnerSnapshot(t, key, 1, nil)
	stale.FreshUntil = routerNow.Add(-time.Nanosecond)
	stale.seal = ownerSnapshotSeal(stale)
	if got := Route(validRouteRequest(t, key, 1, stale, nil)); got.Code != RefusalOwnerSnapshotStale {
		t.Fatalf("stale snapshot accepted: %+v", got)
	}
	tie := []Candidate{validCandidate(key, HorizonShort, "us-short", 10), validCandidate(key, HorizonWeekly, "us-weekly", 10)}
	if got := Route(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), tie)); got.Code != RefusalAmbiguous || got.Decision.LaneID != "" {
		t.Fatalf("unversioned tie selected an owner: %+v", got)
	}
}

func TestGenerationRolloverAndMarketFailureIsolation(t *testing.T) {
	oldKey := mustOwnerKey(t, MarketKR, "005930", 1)
	newKey := mustOwnerKey(t, MarketKR, "005930", 2)
	oldOwner := Owner{Key: oldKey, Horizon: HorizonShort, LaneID: "old", LaneVersion: "v1", CampaignID: "closed", Active: false, Desired: StateOn, Effective: StateOn}
	result := Route(validRouteRequest(t, newKey, 2, mustOwnerSnapshot(t, newKey, 2, []Owner{oldOwner}), []Candidate{validCandidate(newKey, HorizonWeekly, "new-weekly", 5)}))
	if result.Code != RefusalScopeMismatch {
		t.Fatalf("snapshot containing another generation was not refused: %+v", result)
	}
	result = Route(validRouteRequest(t, newKey, 2, mustOwnerSnapshot(t, newKey, 2, nil), []Candidate{validCandidate(newKey, HorizonWeekly, "new-weekly", 5)}))
	if result.Code != RefusalNone || result.Decision.LaneID != "new-weekly" {
		t.Fatalf("new generation was not independently eligible: %+v", result)
	}

	krState, usState := DefaultMarketRecord(MarketKR), DefaultMarketRecord(MarketUS)
	krState.Desired, krState.Effective = StateOff, StateOff
	if got := EvaluateMarketLifecycle(krState, routerNow); got != LifecycleOff {
		t.Fatalf("KR OFF lifecycle=%s", got)
	}
	if got := EvaluateMarketLifecycle(usState, routerNow); got != LifecycleOff {
		t.Fatalf("US peer default lifecycle changed=%s", got)
	}
}

func TestOFFOwnerReturnsDisabledWithoutMutationAuthority(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 1)
	owner := Owner{Key: key, Horizon: HorizonShort, LaneID: "us-short", LaneVersion: "v1", CampaignID: "campaign", Active: true, Desired: StateOff, Effective: StateOff}
	request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, []Owner{owner}), nil)
	request.MarketRecord = DefaultMarketRecord(MarketUS)
	request.ExpectedMarketRevision = 1
	got := Route(request)
	if got.Code != RefusalDisabled || got.Decision.LaneID != "" || !got.CommonSafetyIndependent || got.Mutations != 0 {
		t.Fatalf("OFF owner created routing/mutation authority: %+v", got)
	}
}

func TestDurableMarketRecordCannotBeBypassedByONCandidate(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	request := RouteRequest{
		Key: key, ExpectedOwnerRevision: 1, ExpectedMarketRevision: 1, EvaluatedAt: routerNow,
		Snapshot: mustOwnerSnapshot(t, key, 1, nil), MarketRecord: DefaultMarketRecord(MarketKR),
		Candidates: []Candidate{validCandidate(key, HorizonShort, "kr-short", 100)},
	}
	got := Route(request)
	if got.Code != RefusalDisabled || got.Decision.LaneID != "" || got.Mutations != 0 {
		t.Fatalf("candidate bypassed durable OFF record: %+v", got)
	}
}

func TestMarketRecordBindingRequiresExactMarketAndRevision(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), []Candidate{validCandidate(key, HorizonShort, "kr-short", 1)})
	request.ExpectedMarketRevision++
	if got := Route(request); got.Code != RefusalVersionConflict {
		t.Fatalf("market revision drift accepted: %+v", got)
	}
	request = validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), []Candidate{validCandidate(key, HorizonShort, "kr-short", 1)})
	request.MarketRecord = mustActiveRecord(t, MarketUS)
	if got := Route(request); got.Code != RefusalScopeMismatch {
		t.Fatalf("cross-market binding accepted: %+v", got)
	}
}

func TestKROFFDoesNotGateEligibleUSRoute(t *testing.T) {
	krKey := mustOwnerKey(t, MarketKR, "005930", 1)
	krRequest := RouteRequest{
		Key: krKey, ExpectedOwnerRevision: 1, ExpectedMarketRevision: 1, EvaluatedAt: routerNow,
		Snapshot: mustOwnerSnapshot(t, krKey, 1, nil), MarketRecord: DefaultMarketRecord(MarketKR),
		Candidates: []Candidate{validCandidate(krKey, HorizonShort, "kr-short", 1)},
	}
	if got := Route(krRequest); got.Code != RefusalDisabled {
		t.Fatalf("KR OFF route=%+v", got)
	}

	usKey := mustOwnerKey(t, MarketUS, "AAPL", 1)
	usRequest := validRouteRequest(t, usKey, 1, mustOwnerSnapshot(t, usKey, 1, nil), []Candidate{validCandidate(usKey, HorizonWeekly, "us-weekly", 1)})
	if got := Route(usRequest); got.Code != RefusalNone || got.Decision.LaneID != "us-weekly" {
		t.Fatalf("KR OFF gated independent US route: %+v", got)
	}
}

func TestInactiveRowsDoNotCompeteButAnyCrossScopeRowCorruptsSnapshot(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 2)
	inactive := []Owner{
		{Key: key, Horizon: HorizonShort, LaneID: "old-short", LaneVersion: "v1", CampaignID: "closed-1", Desired: StateOn, Effective: StateOn},
		{Key: key, Horizon: HorizonWeekly, LaneID: "old-weekly", LaneVersion: "v1", CampaignID: "closed-2", Desired: StateOn, Effective: StateOn},
	}
	request := validRouteRequest(t, key, 3, mustOwnerSnapshot(t, key, 3, inactive), []Candidate{validCandidate(key, HorizonWeekly, "new", 5)})
	if got := Route(request); got.Code != RefusalNone || got.Decision.LaneID != "new" {
		t.Fatalf("closed rows incorrectly competed: %+v", got)
	}
	other := mustOwnerKey(t, MarketUS, "AAPL", 1)
	inactive = append(inactive, Owner{Key: other, Horizon: HorizonShort, LaneID: "wrong-generation", LaneVersion: "v1", CampaignID: "closed-3", Desired: StateOn, Effective: StateOn})
	request = validRouteRequest(t, key, 3, mustOwnerSnapshot(t, key, 3, inactive), []Candidate{validCandidate(key, HorizonWeekly, "new", 5)})
	if got := Route(request); got.Code != RefusalScopeMismatch {
		t.Fatalf("cross-scope inactive row did not corrupt snapshot: %+v", got)
	}
}

func mustOwnerKey(t *testing.T, market Market, symbol string, generation uint64) OwnerKey {
	t.Helper()
	key, err := NewOwnerKey("acct", market, symbol, generation)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func mustOwnerSnapshot(t *testing.T, key OwnerKey, revision uint64, owners []Owner) OwnerSnapshot {
	t.Helper()
	snapshot, err := newOwnerSnapshot(key, revision, "owner-digest", routerNow.Add(-time.Second), routerNow.Add(time.Minute), owners)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func validRouteRequest(t *testing.T, key OwnerKey, ownerRevision uint64, snapshot OwnerSnapshot, candidates []Candidate) RouteRequest {
	t.Helper()
	record := mustActiveRecord(t, key.Market)
	return RouteRequest{Key: key, ExpectedOwnerRevision: ownerRevision, ExpectedMarketRevision: record.Revision, EvaluatedAt: routerNow, Snapshot: snapshot, MarketRecord: record, Candidates: candidates}
}

func mustActiveRecord(t *testing.T, market Market) MarketRecord {
	t.Helper()
	return mustRecordUpdateNoTest(DefaultMarketRecord(market), StateOn, string(market)+"-lock", string(market)+"-activation", routerNow.Add(time.Hour))
}

func validCandidate(key OwnerKey, horizon Horizon, laneID string, score int64) Candidate {
	return Candidate{Key: key, Horizon: horizon, LaneID: laneID, LaneVersion: "v1", Score: score, Eligible: true, Desired: StateOn, Effective: StateOn, EvidenceDigest: "evidence", ConfigDigest: "config"}
}
