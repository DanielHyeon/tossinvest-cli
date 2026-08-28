package strategyrouter

import (
	"testing"
	"time"
)

// 태스크 4.3.1: RouteSet 은 평가 *전에* 고르지 않는다.
// 자격 있는 가족 후보를 전부 내보내고, 고르는 일은 순수 평가 뒤 조정자가 한다.
func TestRouteSetEmitsEveryEligibleCandidateWithoutRawScorePreselection(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	candidates := []Candidate{
		validCandidate(key, HorizonShort, "kr-continuation", 30),
		validCandidate(key, HorizonShort, "kr-reversal", 20),
		validCandidate(key, HorizonWeekly, "kr-weekly", 10),
		validCandidate(key, HorizonShort, "kr-breakout", 5),
	}
	result := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), candidates))
	if result.Code != RefusalNone {
		t.Fatalf("eligible set refused: %+v", result)
	}
	if len(result.Decisions) != len(candidates) {
		t.Fatalf("decisions=%d, want every eligible candidate (%d)", len(result.Decisions), len(candidates))
	}
	if result.ExistingOwner {
		t.Fatal("no active owner exists, so ExistingOwner must be false")
	}
	seen := map[string]bool{}
	for _, decision := range result.Decisions {
		if decision.Key != key {
			t.Fatalf("decision escaped the owner scope: %+v", decision)
		}
		if decision.ExistingOwner || decision.CampaignID != "" {
			t.Fatalf("a fresh candidate minted campaign authority: %+v", decision)
		}
		if decision.EvidenceDigest == "" || decision.ConfigDigest == "" {
			t.Fatalf("a fresh candidate lost its digests: %+v", decision)
		}
		seen[decision.LaneID] = true
	}
	for _, candidate := range candidates {
		if !seen[candidate.LaneID] {
			t.Fatalf("eligible candidate %s was silently dropped", candidate.LaneID)
		}
	}
	if !result.Valid() {
		t.Fatal("a freshly built route set does not verify its own seal")
	}
}

// 점수가 같다고 거절하면 안 된다 — 동점 판정은 조정자의 몫으로 옮겼다.
func TestRouteSetDoesNotRefuseOnAScoreTie(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 2)
	tie := []Candidate{
		validCandidate(key, HorizonShort, "us-short", 10),
		validCandidate(key, HorizonWeekly, "us-weekly", 10),
	}
	result := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), tie))
	if result.Code != RefusalNone || len(result.Decisions) != 2 {
		t.Fatalf("a score tie was still resolved inside the router: %+v", result)
	}
	// 같은 입력을 Route 로 보내면 여전히 예전처럼 동점 거절이어야 한다(회귀 금지).
	if legacy := Route(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), tie)); legacy.Code != RefusalAmbiguous {
		t.Fatalf("legacy Route changed behaviour: %+v", legacy)
	}
}

// 활성 소유자는 점수 비교보다 언제나 먼저다. 그 경우 집합은 그 하나뿐이다.
func TestRouteSetPreservesTheActiveOwnerAloneBeforeAnyComparison(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	owner := Owner{Key: key, Horizon: HorizonShort, LaneID: "kr-short", LaneVersion: "v1", CampaignID: "campaign-1", Active: true, Desired: StateOn, Effective: StateOn}
	snapshot := mustOwnerSnapshot(t, key, 4, []Owner{owner})
	result := RouteSet(validRouteRequest(t, key, 4, snapshot, []Candidate{
		validCandidate(key, HorizonWeekly, "kr-weekly", 999),
		validCandidate(key, HorizonShort, "kr-breakout", 998),
	}))
	if result.Code != RefusalNone || len(result.Decisions) != 1 || !result.ExistingOwner {
		t.Fatalf("the active owner was not preserved alone: %+v", result)
	}
	if result.Decisions[0].LaneID != owner.LaneID || !result.Decisions[0].ExistingOwner ||
		result.Decisions[0].CampaignID != owner.CampaignID {
		t.Fatalf("the active owner lost its campaign lineage: %+v", result.Decisions[0])
	}
	if result.Decisions[0].EvidenceDigest != "" || result.Decisions[0].ConfigDigest != "" {
		t.Fatalf("an existing owner minted fresh evidence digests: %+v", result.Decisions[0])
	}
}

// 자격이 없는 후보는 조용히 섞이지 않는다.
func TestRouteSetExcludesIneligibleAndDisabledCandidates(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	eligible := validCandidate(key, HorizonShort, "kr-continuation", 30)
	ineligible := validCandidate(key, HorizonShort, "kr-reversal", 99)
	ineligible.Eligible = false
	desiredOff := validCandidate(key, HorizonWeekly, "kr-weekly", 98)
	desiredOff.Desired = StateOff
	effectiveOff := validCandidate(key, HorizonShort, "kr-breakout", 97)
	effectiveOff.Effective = StateOff
	result := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil),
		[]Candidate{eligible, ineligible, desiredOff, effectiveOff}))
	if result.Code != RefusalNone || len(result.Decisions) != 1 || result.Decisions[0].LaneID != eligible.LaneID {
		t.Fatalf("ineligible or OFF candidates leaked into the set: %+v", result)
	}
}

// 아무도 자격이 없으면 빈 성공이 아니라 거절이다.
func TestRouteSetRefusesWhenNothingIsEligible(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 2)
	off := validCandidate(key, HorizonShort, "us-short", 10)
	off.Desired = StateOff
	result := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), []Candidate{off}))
	if result.Code != RefusalDisabled || len(result.Decisions) != 0 {
		t.Fatalf("an empty set was reported as success: %+v", result)
	}
	if result.Valid() {
		t.Fatal("a refused set must not verify as an authority")
	}
}

// 출력 순서는 입력 순서에 의존하지 않는다.
func TestRouteSetOrderIsDeterministicAndInputOrderIndependent(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	forward := []Candidate{
		validCandidate(key, HorizonShort, "kr-continuation", 1),
		validCandidate(key, HorizonShort, "kr-reversal", 2),
		validCandidate(key, HorizonWeekly, "kr-weekly", 3),
		validCandidate(key, HorizonShort, "kr-breakout", 4),
	}
	reversed := make([]Candidate, 0, len(forward))
	for index := len(forward) - 1; index >= 0; index-- {
		reversed = append(reversed, forward[index])
	}
	first := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), forward))
	second := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), reversed))
	if first.Code != RefusalNone || second.Code != RefusalNone {
		t.Fatalf("deterministic-order fixture refused: %+v %+v", first, second)
	}
	if len(first.Decisions) != len(second.Decisions) {
		t.Fatalf("different lengths: %d vs %d", len(first.Decisions), len(second.Decisions))
	}
	for index := range first.Decisions {
		if first.Decisions[index] != second.Decisions[index] {
			t.Fatalf("order depended on input order at %d: %+v vs %+v", index, first.Decisions[index], second.Decisions[index])
		}
	}
	if first.SetDigest() != second.SetDigest() {
		t.Fatalf("set digest depended on input order: %q vs %q", first.SetDigest(), second.SetDigest())
	}
}

// 봉인은 사후 변경을 잡아야 한다.
func TestRouteSetSealDetectsPostBuildMutation(t *testing.T) {
	key := mustOwnerKey(t, MarketKR, "005930", 1)
	result := RouteSet(validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil),
		[]Candidate{validCandidate(key, HorizonShort, "kr-continuation", 1)}))
	if !result.Valid() {
		t.Fatal("valid set did not verify")
	}
	mutated := result
	mutated.Decisions = append([]RouteDecision(nil), result.Decisions...)
	mutated.Decisions[0].LaneID = "kr-swapped"
	if mutated.Valid() {
		t.Fatal("a swapped lane id survived the seal")
	}
	appended := result
	appended.Decisions = append(append([]RouteDecision(nil), result.Decisions...),
		RouteDecision{Key: key, Horizon: HorizonShort, LaneID: "kr-injected", LaneVersion: "v1", EvidenceDigest: "e", ConfigDigest: "c"})
	if appended.Valid() {
		t.Fatal("an injected decision survived the seal")
	}
	zero := RouteSetResult{}
	if zero.Valid() {
		t.Fatal("the zero value verified as an authority")
	}
}

// 전단 검증은 Route 와 같은 실패를 같은 코드로 내야 한다.
func TestRouteSetFailsClosedExactlyLikeRouteOnPreludeFaults(t *testing.T) {
	key := mustOwnerKey(t, MarketUS, "AAPL", 2)
	candidates := []Candidate{validCandidate(key, HorizonShort, "us-short", 10)}

	t.Run("multiple-active-owners", func(t *testing.T) {
		first := Owner{Key: key, Horizon: HorizonShort, LaneID: "us-short", LaneVersion: "v1", CampaignID: "c1", Active: true, Desired: StateOn, Effective: StateOn}
		second := Owner{Key: key, Horizon: HorizonWeekly, LaneID: "us-weekly", LaneVersion: "v1", CampaignID: "c2", Active: true, Desired: StateOn, Effective: StateOn}
		request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, []Owner{first, second}), candidates)
		if got := RouteSet(request); got.Code != RefusalReconstructionMismatch || len(got.Decisions) != 0 {
			t.Fatalf("multiple owners did not fail closed: %+v", got)
		}
	})

	t.Run("stale-owner-snapshot", func(t *testing.T) {
		stale := mustOwnerSnapshot(t, key, 1, nil)
		stale.FreshUntil = routerNow.Add(-time.Nanosecond)
		stale.seal = ownerSnapshotSeal(stale)
		if got := RouteSet(validRouteRequest(t, key, 1, stale, candidates)); got.Code != RefusalOwnerSnapshotStale {
			t.Fatalf("stale snapshot accepted: %+v", got)
		}
	})

	t.Run("scope-mismatch", func(t *testing.T) {
		other := mustOwnerKey(t, MarketUS, "MSFT", 2)
		request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), []Candidate{validCandidate(other, HorizonShort, "us-short", 10)})
		if got := RouteSet(request); got.Code != RefusalScopeMismatch || len(got.Decisions) != 0 {
			t.Fatalf("a foreign owner scope was routed: %+v", got)
		}
	})

	t.Run("invalid-request", func(t *testing.T) {
		request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), candidates)
		request.ExpectedOwnerRevision = 0
		if got := RouteSet(request); got.Code != RefusalInvalid {
			t.Fatalf("a request without an expected revision was accepted: %+v", got)
		}
	})

	t.Run("market-revision-conflict", func(t *testing.T) {
		request := validRouteRequest(t, key, 1, mustOwnerSnapshot(t, key, 1, nil), candidates)
		request.ExpectedMarketRevision = request.MarketRecord.Revision + 1
		if got := RouteSet(request); got.Code != RefusalVersionConflict {
			t.Fatalf("a market revision conflict was accepted: %+v", got)
		}
	})
}
