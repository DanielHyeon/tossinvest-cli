package scheduler

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type incrementingEntropy struct {
	next byte
}

func (r *incrementingEntropy) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.next
		r.next++
	}
	return len(p), nil
}

type failingEntropy struct{}

func (failingEntropy) Read([]byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}

type repeatingEntropy byte

func (r repeatingEntropy) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(r)
	}
	return len(p), nil
}

type uniqueEntropy struct {
	next uint64
}

func (r *uniqueEntropy) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	r.next++
	if len(p) >= 8 {
		binary.BigEndian.PutUint64(p[len(p)-8:], r.next)
	}
	return len(p), nil
}

func deterministicBudgetCoordinator(seed byte) *BudgetCoordinator {
	return deterministicBudgetCoordinatorAt(seed, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
}

func deterministicBudgetCoordinatorAt(seed byte, completionTime time.Time) *BudgetCoordinator {
	return newBudgetCoordinatorWithEntropyAndClock(&incrementingEntropy{next: seed}, func() time.Time {
		return completionTime
	})
}

func budget(at time.Time, remaining int) official.RateBudget {
	return official.RateBudget{
		Path: "/api/v1/rankings", Limit: 100, Remaining: remaining,
		Reset: at.Add(time.Minute), ResetRaw: "60", ResetKind: official.ResetDelta,
		ObservedAt: at, Reported: true,
	}
}

func TestSafetyReserveIsHalfRoundedUpWithFiveCallFloor(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct{ remaining, want int }{
		{100, 50}, {11, 6}, {10, 5}, {9, 5}, {4, 5}, {maxInt, maxInt/2 + 1},
	} {
		if got := SafetyReserve(tc.remaining); got != tc.want {
			t.Errorf("SafetyReserve(%d) = %d, want %d", tc.remaining, got, tc.want)
		}
	}
}

func TestBudgetArithmeticFailsClosedForInvalidBounds(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	maxInt := int(^uint(0) >> 1)
	c := deterministicBudgetCoordinator(1)
	valid := budget(now, maxInt)
	valid.Limit = maxInt
	c.Observe(valid)
	grant := c.TryAcquire(valid.Path, PollCandidate, now)
	if !grant.Allowed || grant.Reserve != maxInt/2+1 || grant.Available != maxInt-(maxInt/2+1) {
		t.Fatalf("MaxInt budget arithmetic overflowed or denied: %+v", grant)
	}

	for _, tc := range []struct {
		name      string
		limit     int
		remaining int
	}{
		{name: "negative limit", limit: -1, remaining: 0},
		{name: "negative remaining", limit: 100, remaining: -1},
		{name: "remaining exceeds limit", limit: 10, remaining: 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coordinator := deterministicBudgetCoordinator(11)
			observation := budget(now, tc.remaining)
			observation.Limit = tc.limit
			coordinator.Observe(observation)
			got := coordinator.TryAcquire(observation.Path, PollCandidate, now)
			if got.Allowed || got.Reason != BudgetInvalidBounds {
				t.Fatalf("invalid bounds were not rejected: %+v", got)
			}
		})
	}
}

func TestCandidateEntryAndAnalyticsNeverSpendReservedBudget(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	c.Observe(budget(now, 11)) // reserve 6, five low-priority calls available
	for i := 0; i < 5; i++ {
		grant := c.TryAcquire("/api/v1/rankings", PollCandidate, now)
		if !grant.Allowed {
			t.Fatalf("grant %d refused: %+v", i, grant)
		}
	}
	if got := c.TryAcquire("/api/v1/rankings", PollEntry, now); got.Allowed || got.Reason != BudgetReserved {
		t.Fatalf("sixth low-priority call = %+v", got)
	}
	if got := c.TryAcquire("/api/v1/rankings", PollAnalytics, now); got.Allowed {
		t.Fatalf("analytics consumed reserve: %+v", got)
	}
}

func TestMissingStaleAndClockSkewBudgetGrantNoAdditionalLowPriorityPoll(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	cases := []struct {
		name string
		b    *official.RateBudget
	}{
		{"missing", nil},
		{"reset missing", func() *official.RateBudget {
			b := budget(now, 20)
			b.Reset = time.Time{}
			b.ResetKind = official.ResetAbsent
			return &b
		}()},
		{"stale", func() *official.RateBudget { b := budget(now.Add(-2*time.Minute), 20); return &b }()},
		{"clock moved backward", func() *official.RateBudget { b := budget(now.Add(time.Minute), 20); return &b }()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewBudgetCoordinator()
			if tc.b != nil {
				c.Observe(*tc.b)
			}
			for _, class := range []PollClass{PollCandidate, PollEntry, PollAnalytics} {
				if got := c.TryAcquire("/api/v1/rankings", class, now); got.Allowed {
					t.Errorf("%s granted with invalid provenance: %+v", class, got)
				}
			}
		})
	}
}

func TestSafetyClassesContinueWithoutBudgetProvenance(t *testing.T) {
	c := NewBudgetCoordinator()
	now := at(t, "2026-08-01T00:00:00Z")
	for _, class := range []PollClass{PollEmergencyExit, PollReconcile, PollFillDetection, PollProtection} {
		if got := c.TryAcquire("missing", class, now); !got.Allowed || got.Reason != BudgetSafetyPriority {
			t.Errorf("%s = %+v", class, got)
		}
	}
}

func TestCoordinatorIsEndpointKeyed(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	a := budget(now, 20)
	b := budget(now, 4)
	b.Path = "/api/v1/prices"
	c.Observe(a)
	c.Observe(b)
	if !c.TryAcquire(a.Path, PollEntry, now).Allowed {
		t.Fatal("healthy endpoint was contaminated by another key")
	}
	if c.TryAcquire(b.Path, PollEntry, now).Allowed {
		t.Fatal("tight endpoint borrowed another key's budget")
	}
}

func TestOlderBudgetObservationCannotReplaceNewerEvidence(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	newer := budget(now, 20)
	newer.Reset = now.Add(10 * time.Minute)
	newer.ResetRaw = "600"
	c.Observe(newer)
	older := budget(now.Add(-time.Minute), 100)
	older.Reset = now.Add(10 * time.Minute)
	older.ResetRaw = "660"
	c.Observe(older)
	grant := c.TryAcquire(newer.Path, PollEntry, now)
	if !grant.Allowed || grant.Remaining != 20 || grant.Reserve != 10 {
		t.Fatalf("out-of-order observation replaced current evidence: %+v", grant)
	}
}

func TestEqualTimestampBudgetCorrectionOnlyMovesConservatively(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	first := budget(now, 20)
	c.Observe(first)

	correction := first
	correction.Remaining = 8
	c.Observe(correction)
	higher := first
	higher.Remaining = 100
	c.Observe(higher)

	grant := c.TryAcquire(first.Path, PollEntry, now)
	if !grant.Allowed || grant.Remaining != 8 || grant.Reserve != 5 || grant.Available != 3 {
		t.Fatalf("equal-time conservative correction lost or relaxed: %+v", grant)
	}
}

func TestEqualTimestampBudgetCorrectionWithConflictingProvenanceFailsClosed(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	mutations := []struct {
		name   string
		mutate func(*official.RateBudget)
	}{
		{"reported", func(b *official.RateBudget) { b.Reported = false }},
		{"limit", func(b *official.RateBudget) { b.Limit++ }},
		{"reset", func(b *official.RateBudget) { b.Reset = b.Reset.Add(time.Minute) }},
		{"reset raw", func(b *official.RateBudget) { b.ResetRaw = "conflict" }},
		{"reset kind", func(b *official.RateBudget) { b.ResetKind = official.ResetKind("future-kind") }},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			c := NewBudgetCoordinator()
			first := budget(now, 20)
			first.ResetRaw = "60"
			c.Observe(first)
			correction := first
			tc.mutate(&correction)
			c.Observe(correction)
			if grant := c.TryAcquire(first.Path, PollEntry, now); grant.Allowed || grant.Reason != BudgetConflictingProvenance {
				t.Fatalf("conflicting equal-time correction was masked: %+v", grant)
			}
		})
	}
}

func TestOutstandingCommitmentsSurviveNewObservationInSameResetWindow(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinatorAt(1, now.Add(1500*time.Millisecond))
	first := budget(now, 11) // reserve 6; five discretionary requests
	c.Observe(first)
	grants := make([]BudgetGrant, 3)
	for i := range grants {
		grants[i] = c.TryAcquire(first.Path, PollCandidate, now)
		if !grants[i].Allowed || grants[i].Commitment == (CommitmentToken{}) {
			t.Fatalf("initial grant %d = %+v", i, grants[i])
		}
	}

	newer := first
	newer.ObservedAt = now.Add(time.Second)
	newer.ResetRaw = "59"
	newer.Remaining = 8 // potentially reflects none, some, or all in-flight calls
	c.Observe(newer)
	if grant := c.TryAcquire(first.Path, PollEntry, newer.ObservedAt); grant.Allowed || grant.Reason != BudgetReserved {
		t.Fatalf("same-window observation forgot in-flight commitments: %+v", grant)
	}

	if !c.Complete(first.Path, grants[0].Commitment) {
		t.Fatal("known commitment was not completed")
	}
	if grant := c.TryAcquire(first.Path, PollEntry, newer.ObservedAt); grant.Allowed || grant.Reason != BudgetReserved {
		t.Fatalf("completion without newer authority released capacity: %+v", grant)
	}
	if c.Complete(first.Path, grants[0].Commitment) {
		t.Fatal("commitment completed twice")
	}

	reconciled := newer
	reconciled.ObservedAt = newer.ObservedAt.Add(time.Second)
	reconciled.ResetRaw = "58"
	cycle := c.BeginObservation(first.Path)
	if !c.ObserveCycle(reconciled, cycle) {
		t.Fatal("reconciliation cycle was rejected")
	}
	grant := c.TryAcquire(first.Path, PollEntry, reconciled.ObservedAt)
	if !grant.Allowed || grant.Commitment == (CommitmentToken{}) {
		t.Fatalf("newer authoritative observation did not reconcile completed call: %+v", grant)
	}
}

func TestNewResetWindowReconcilesOldCommitments(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	first := budget(now, 11)
	c.Observe(first)
	old := c.TryAcquire(first.Path, PollCandidate, now)
	if !old.Allowed {
		t.Fatalf("old-window grant = %+v", old)
	}
	if !c.Complete(first.Path, old.Commitment) {
		t.Fatal("old-window completion was not recorded")
	}
	cycle := c.BeginObservation(first.Path)
	newWindow := budget(now.Add(61*time.Second), 11)
	if !c.ObserveCycle(newWindow, cycle) {
		t.Fatal("new-window cycle was rejected")
	}
	if c.Complete(first.Path, old.Commitment) {
		t.Fatal("old-window commitment survived reset reconciliation")
	}
	for i := 0; i < 5; i++ {
		if grant := c.TryAcquire(first.Path, PollCandidate, newWindow.ObservedAt); !grant.Allowed {
			t.Fatalf("new-window grant %d = %+v", i, grant)
		}
	}
}

func TestNewResetWindowReconcilesCommitmentsAfterInvalidInterimObservation(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	first := budget(now, 11)
	c.Observe(first)
	old := c.TryAcquire(first.Path, PollCandidate, now)
	if !old.Allowed {
		t.Fatalf("old-window grant = %+v", old)
	}
	if !c.Complete(first.Path, old.Commitment) {
		t.Fatal("old-window completion was not recorded")
	}
	invalid := first
	invalid.ObservedAt = now.Add(30 * time.Second)
	invalid.Reported = false
	invalid.Reset = time.Time{}
	invalid.ResetKind = official.ResetAbsent
	c.Observe(invalid)
	cycle := c.BeginObservation(first.Path)
	newWindow := budget(now.Add(61*time.Second), 11)
	if !c.ObserveCycle(newWindow, cycle) {
		t.Fatal("new-window cycle was rejected")
	}
	if c.Complete(first.Path, old.Commitment) {
		t.Fatal("invalid interim evidence hid the positive reset boundary")
	}
}

func TestCommitmentCapabilityIsOpaqueAndBoundToCoordinatorKeyClassAndGeneration(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	observation := budget(now, 20)
	c.Observe(observation)
	grant := c.TryAcquire(observation.Path, PollCandidate, now)
	if !grant.Allowed || grant.Commitment == (CommitmentToken{}) {
		t.Fatalf("grant did not carry a capability: %+v", grant)
	}
	if c.Complete(observation.Path, CommitmentToken{}) {
		t.Fatal("zero capability completed a commitment")
	}
	var nilCoordinator *BudgetCoordinator
	if nilCoordinator.Complete(observation.Path, grant.Commitment) {
		t.Fatal("nil coordinator completed a commitment")
	}

	tokenType := reflect.TypeOf(grant.Commitment)
	for i := 0; i < tokenType.NumField(); i++ {
		if tokenType.Field(i).PkgPath == "" {
			t.Fatalf("commitment authority leaks exported field %q", tokenType.Field(i).Name)
		}
	}

	forged := grant.Commitment
	forged.capability[0] ^= 0xff
	if c.Complete(observation.Path, forged) {
		t.Fatal("forged capability completed a commitment")
	}
	crossClass := grant.Commitment
	crossClass.class = PollEntry
	if c.Complete(observation.Path, crossClass) {
		t.Fatal("cross-class capability completed a commitment")
	}
	if c.Complete("/api/v1/other", grant.Commitment) {
		t.Fatal("cross-key capability completed a commitment")
	}
	other := deterministicBudgetCoordinator(101)
	other.Observe(observation)
	if other.Complete(observation.Path, grant.Commitment) {
		t.Fatal("cross-coordinator capability completed a commitment")
	}

	if !c.Complete(observation.Path, grant.Commitment) {
		t.Fatal("commitment was not completed before generation transition")
	}
	cycle := c.BeginObservation(observation.Path)
	newWindow := budget(now.Add(61*time.Second), 20)
	if !c.ObserveCycle(newWindow, cycle) {
		t.Fatal("new-window cycle was rejected")
	}
	if c.Complete(observation.Path, grant.Commitment) {
		t.Fatal("cross-generation capability completed a commitment")
	}
}

func TestResetGenerationExhaustionFailsClosedWithoutWrapping(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	first := budget(now, 20)
	c.Observe(first)
	state := c.endpoints[first.Path]
	state.generation = ^uint64(0)
	c.endpoints[first.Path] = state

	cycle := c.BeginObservation(first.Path)
	newWindow := budget(now.Add(61*time.Second), 20)
	if !c.ObserveCycle(newWindow, cycle) {
		t.Fatal("generation-exhaustion cycle was rejected")
	}
	grant := c.TryAcquire(first.Path, PollCandidate, newWindow.ObservedAt)
	if grant.Allowed || grant.Reason != BudgetTokenUnavailable {
		t.Fatalf("exhausted generation wrapped or granted: %+v", grant)
	}
}

func TestCommitmentEntropyFailureFailsClosedWithoutBlockingSafety(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropy(failingEntropy{})
	c.Observe(budget(now, 20))
	if grant := c.TryAcquire("/api/v1/rankings", PollCandidate, now); grant.Allowed || grant.Reason != BudgetTokenUnavailable {
		t.Fatalf("low-priority grant survived entropy failure: %+v", grant)
	}
	if grant := c.TryAcquire("missing", PollProtection, now); !grant.Allowed || grant.Reason != BudgetSafetyPriority {
		t.Fatalf("safety grant was blocked by entropy failure: %+v", grant)
	}
}

func TestCompletionClockFailureLeavesCommitmentInFlight(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	clocks := []struct {
		name string
		now  func() time.Time
	}{
		{name: "nil", now: nil},
		{name: "zero", now: func() time.Time { return time.Time{} }},
	}
	for seed, tc := range clocks {
		t.Run(tc.name, func(t *testing.T) {
			c := newBudgetCoordinatorWithEntropyAndClock(&incrementingEntropy{next: byte(seed + 1)}, tc.now)
			observation := budget(now, 6)
			c.Observe(observation)
			grant := c.TryAcquire(observation.Path, PollCandidate, now)
			if !grant.Allowed {
				t.Fatalf("initial grant failed: %+v", grant)
			}
			if c.Complete(observation.Path, grant.Commitment) {
				t.Fatal("completion succeeded without a trustworthy clock")
			}
			if next := c.TryAcquire(observation.Path, PollCandidate, now); next.Allowed {
				t.Fatalf("clock failure released in-flight capacity: %+v", next)
			}
		})
	}
}

func TestCommitmentCapabilityIsNeverReissuedWithinGeneration(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropyAndClock(repeatingEntropy(7), func() time.Time { return now })
	observation := budget(now, 20)
	c.Observe(observation)
	first := c.TryAcquire(observation.Path, PollCandidate, now)
	if !first.Allowed || !c.Complete(observation.Path, first.Commitment) {
		t.Fatalf("first capability lifecycle failed: %+v", first)
	}
	newer := observation
	newer.ObservedAt = now.Add(time.Second)
	newer.ResetRaw = "59"
	c.Observe(newer)
	if grant := c.TryAcquire(observation.Path, PollCandidate, newer.ObservedAt); grant.Allowed || grant.Reason != BudgetTokenUnavailable {
		t.Fatalf("repeated random capability was reissued: %+v", grant)
	}
}

func TestCompleteNeverReopensCapacityWithoutAuthoritativeObservation(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	observation := budget(now, 11)
	c.Observe(observation)
	for i := 0; i < 5; i++ {
		grant := c.TryAcquire(observation.Path, PollCandidate, now)
		if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
			t.Fatalf("acquire/complete %d failed: %+v", i, grant)
		}
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, now); grant.Allowed || grant.Reason != BudgetReserved {
		t.Fatalf("repeated completion reused stale capacity: %+v", grant)
	}
}

func TestCompletionOutcomeAlwaysRemainsConsumedWithoutAuthority(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	for seed, outcome := range []string{"success", "error", "cancellation"} {
		t.Run(outcome, func(t *testing.T) {
			c := deterministicBudgetCoordinator(byte(seed + 1))
			observation := budget(now, 6) // reserve 5; one discretionary call
			c.Observe(observation)
			grant := c.TryAcquire(observation.Path, PollCandidate, now)
			if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
				t.Fatalf("%s completion was not recorded: %+v", outcome, grant)
			}
			if next := c.TryAcquire(observation.Path, PollCandidate, now); next.Allowed || next.Reason != BudgetReserved {
				t.Fatalf("%s completion restored stale capacity: %+v", outcome, next)
			}
		})
	}
}

func TestCompletedCommitmentCannotUseStaleObservation(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	observation := budget(now, 20)
	observation.Reset = now.Add(time.Hour)
	observation.ResetRaw = "3600"
	c.Observe(observation)
	grant := c.TryAcquire(observation.Path, PollCandidate, now)
	if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
		t.Fatalf("initial lifecycle failed: %+v", grant)
	}
	if next := c.TryAcquire(observation.Path, PollCandidate, now.Add(budgetObservationMaxAge)); next.Allowed || next.Reason != BudgetStale {
		t.Fatalf("stale observation restored capacity: %+v", next)
	}
}

func TestOlderResponseBudgetCannotReconcileCompletedCommitments(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	observation := budget(now, 11)
	c.Observe(observation)
	for i := 0; i < 5; i++ {
		grant := c.TryAcquire(observation.Path, PollCandidate, now)
		if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
			t.Fatalf("acquire/complete %d failed: %+v", i, grant)
		}
	}
	older := observation
	older.ObservedAt = now.Add(-time.Second)
	older.ResetRaw = "61"
	older.Remaining = 100
	c.Observe(older)
	if grant := c.TryAcquire(observation.Path, PollCandidate, now); grant.Allowed || grant.Reason != BudgetReserved {
		t.Fatalf("older response reconciled completed commitments: %+v", grant)
	}
}

func TestMissingOrInvalidResponseBudgetCannotReconcileCompletedCommitment(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinator(1)
	observation := budget(now, 11)
	c.Observe(observation)
	for i := 0; i < 5; i++ {
		grant := c.TryAcquire(observation.Path, PollCandidate, now)
		if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
			t.Fatalf("acquire/complete %d failed: %+v", i, grant)
		}
	}
	missing := observation
	missing.ObservedAt = now.Add(time.Second)
	missing.Reported = false
	missing.Reset = time.Time{}
	missing.ResetKind = official.ResetAbsent
	c.Observe(missing)
	if grant := c.TryAcquire(observation.Path, PollCandidate, missing.ObservedAt); grant.Allowed || grant.Reason != BudgetMissing {
		t.Fatalf("missing response budget reconciled or granted: %+v", grant)
	}
	valid := observation
	valid.ObservedAt = now.Add(2 * time.Second)
	valid.ResetRaw = "58"
	cycle := c.BeginObservation(observation.Path)
	if !c.ObserveCycle(valid, cycle) {
		t.Fatal("valid response cycle was rejected")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, valid.ObservedAt); !grant.Allowed {
		t.Fatalf("newer authoritative response did not reconcile completed calls: %+v", grant)
	}
}

func TestObservationBeforeCompletionDoesNotReleaseUntilFollowingAuthority(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinatorAt(1, now.Add(2*time.Second))
	observation := budget(now, 11)
	c.Observe(observation)
	grants := make([]BudgetGrant, 5)
	for i := range grants {
		grants[i] = c.TryAcquire(observation.Path, PollCandidate, now)
	}
	newer := observation
	newer.ObservedAt = now.Add(time.Second)
	newer.ResetRaw = "59"
	heldCycle := c.BeginObservation(observation.Path)
	if !c.ObserveCycle(newer, heldCycle) {
		t.Fatal("pre-completion observation cycle was rejected")
	}
	if !c.Complete(observation.Path, grants[0].Commitment) {
		t.Fatal("completion after observation was not recorded")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, newer.ObservedAt); grant.Allowed {
		t.Fatalf("observation-before-completion race released capacity: %+v", grant)
	}
	following := newer
	following.ObservedAt = now.Add(3 * time.Second)
	following.ResetRaw = "57"
	followingCycle := c.BeginObservation(observation.Path)
	if !c.ObserveCycle(following, followingCycle) {
		t.Fatal("post-completion observation cycle was rejected")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, following.ObservedAt); !grant.Allowed {
		t.Fatalf("following authority did not reconcile completion: %+v", grant)
	}
}

func TestEarlierResponseObservedAfterCompletionCannotReconcile(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := deterministicBudgetCoordinatorAt(1, now.Add(2*time.Second))
	observation := budget(now, 11)
	c.Observe(observation)
	grants := make([]BudgetGrant, 5)
	for i := range grants {
		grants[i] = c.TryAcquire(observation.Path, PollCandidate, now)
	}
	heldCycle := c.BeginObservation(observation.Path)
	if !c.Complete(observation.Path, grants[0].Commitment) {
		t.Fatal("completion was not recorded")
	}
	heldEarlierResponse := observation
	heldEarlierResponse.ObservedAt = now.Add(time.Second)
	heldEarlierResponse.ResetRaw = "59"
	if !c.ObserveCycle(heldEarlierResponse, heldCycle) {
		t.Fatal("held response cycle was rejected")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, heldEarlierResponse.ObservedAt); grant.Allowed {
		t.Fatalf("earlier response processed later reconciled completion: %+v", grant)
	}
	following := heldEarlierResponse
	following.ObservedAt = now.Add(3 * time.Second)
	following.ResetRaw = "57"
	followingCycle := c.BeginObservation(observation.Path)
	if !c.ObserveCycle(following, followingCycle) {
		t.Fatal("following response cycle was rejected")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, following.ObservedAt); !grant.Allowed {
		t.Fatalf("response completed after completion did not reconcile it: %+v", grant)
	}
}

func TestUnknownPollClassFailsClosed(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	c.Observe(budget(now, 20))
	if grant := c.TryAcquire("/api/v1/rankings", PollClass("future-class"), now); grant.Allowed {
		t.Fatalf("unknown poll class was granted: %+v", grant)
	}
}

func TestUnknownResetKindFailsClosed(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	observation := budget(now, 20)
	observation.ResetKind = official.ResetKind("future-kind")
	c.Observe(observation)
	if grant := c.TryAcquire(observation.Path, PollEntry, now); grant.Allowed {
		t.Fatalf("unknown reset kind was granted: %+v", grant)
	}
}

func TestOldObservationIsStaleEvenWhenResetIsStillInFuture(t *testing.T) {
	now := at(t, "2026-08-01T00:10:00Z")
	c := NewBudgetCoordinator()
	old := budget(now.Add(-2*time.Minute), 100)
	old.Reset = now.Add(time.Hour)
	old.ResetRaw = "3720"
	c.Observe(old)
	if grant := c.TryAcquire(old.Path, PollCandidate, now); grant.Allowed || grant.Reason != BudgetStale {
		t.Fatalf("old provenance = %+v", grant)
	}
}

func TestAnalyticsCannotConsumeCandidateAndEntryShare(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	observation := budget(now, 20) // safety 10; discretionary 10
	c.Observe(observation)
	for i := 0; i < 5; i++ {
		if grant := c.TryAcquire(observation.Path, PollAnalytics, now); !grant.Allowed {
			t.Fatalf("analytics %d = %+v", i, grant)
		}
	}
	if grant := c.TryAcquire(observation.Path, PollAnalytics, now); grant.Allowed || grant.Reason != BudgetEntryPriority {
		t.Fatalf("analytics crossed lower-priority share: %+v", grant)
	}
	for i := 0; i < 5; i++ {
		if grant := c.TryAcquire(observation.Path, PollEntry, now); !grant.Allowed {
			t.Fatalf("entry share %d was consumed by analytics: %+v", i, grant)
		}
	}
}

func TestAnalyticsCommitmentsSurviveNewObservationInSameResetWindow(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := NewBudgetCoordinator()
	observation := budget(now, 20)
	c.Observe(observation)
	for i := 0; i < 5; i++ {
		if grant := c.TryAcquire(observation.Path, PollAnalytics, now); !grant.Allowed {
			t.Fatalf("analytics %d = %+v", i, grant)
		}
	}
	newer := observation
	newer.ObservedAt = now.Add(time.Second)
	newer.ResetRaw = "59"
	c.Observe(newer)
	if grant := c.TryAcquire(observation.Path, PollAnalytics, newer.ObservedAt); grant.Allowed || grant.Reason != BudgetEntryPriority {
		t.Fatalf("same-window observation forgot analytics commitments: %+v", grant)
	}
}

func TestPollPriorityOrderIsExplicit(t *testing.T) {
	ordered := []PollClass{PollEmergencyExit, PollReconcile, PollFillDetection, PollProtection, PollCandidate, PollAnalytics}
	for i := 1; i < len(ordered); i++ {
		if ordered[i-1].Priority() <= ordered[i].Priority() {
			t.Fatalf("priority %s=%d is not above %s=%d", ordered[i-1], ordered[i-1].Priority(), ordered[i], ordered[i].Priority())
		}
	}
	if PollCandidate.Priority() != PollEntry.Priority() {
		t.Fatalf("candidate=%d entry=%d", PollCandidate.Priority(), PollEntry.Priority())
	}
}

func TestHeldPreCompletionResponseCannotReconcileAfterWallClockRollback(t *testing.T) {
	base := at(t, "2026-08-01T00:00:00Z")
	wall := base.Add(10 * time.Second)
	c := newBudgetCoordinatorWithEntropyAndClock(&incrementingEntropy{next: 1}, func() time.Time { return wall })
	initial := budget(base, 11)
	initial.ResetRaw = "60"
	c.Observe(initial)

	grants := make([]BudgetGrant, 5)
	for i := range grants {
		grants[i] = c.TryAcquire(initial.Path, PollCandidate, base)
		if !grants[i].Allowed {
			t.Fatalf("initial grant %d = %+v", i, grants[i])
		}
	}
	heldCycle := c.BeginObservation(initial.Path) // request starts before completion
	if heldCycle == (ObservationCycle{}) {
		t.Fatal("held request did not get an observation cycle")
	}

	wall = base.Add(-time.Minute) // the wall clock rolls behind both requests
	if !c.Complete(initial.Path, grants[0].Commitment) {
		t.Fatal("completion after rollback was not recorded")
	}
	heldResponse := initial
	heldResponse.ObservedAt = base.Add(20 * time.Second) // wall timestamp alone looks later
	heldResponse.ResetRaw = "40"
	if !c.ObserveCycle(heldResponse, heldCycle) {
		t.Fatal("valid held response cycle was rejected")
	}
	if grant := c.TryAcquire(initial.Path, PollCandidate, heldResponse.ObservedAt); grant.Allowed {
		t.Fatalf("pre-completion request reconciled a later completion: %+v", grant)
	}

	followingCycle := c.BeginObservation(initial.Path) // begins after completion
	following := heldResponse
	following.ObservedAt = base.Add(21 * time.Second)
	following.ResetRaw = "39"
	if !c.ObserveCycle(following, followingCycle) {
		t.Fatal("following response cycle was rejected")
	}
	if grant := c.TryAcquire(initial.Path, PollCandidate, following.ObservedAt); !grant.Allowed {
		t.Fatalf("post-completion request did not reconcile completion: %+v", grant)
	}
}

func TestObservationCycleIsOpaqueOneShotAndScopeBound(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	var nilCoordinator *BudgetCoordinator
	if cycle := nilCoordinator.BeginObservation("/api/v1/rankings"); cycle != (ObservationCycle{}) {
		t.Fatalf("nil coordinator minted observation cycle: %+v", cycle)
	}
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	observation := budget(now, 6)
	observation.ResetRaw = "60"
	c.Observe(observation)
	grant := c.TryAcquire(observation.Path, PollCandidate, now)
	if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
		t.Fatalf("initial lifecycle failed: %+v", grant)
	}

	cycle := c.BeginObservation(observation.Path)
	cycleType := reflect.TypeOf(cycle)
	for i := 0; i < cycleType.NumField(); i++ {
		if cycleType.Field(i).PkgPath == "" {
			t.Fatalf("observation cycle leaks exported field %q", cycleType.Field(i).Name)
		}
	}
	newer := observation
	newer.ObservedAt = now.Add(time.Second)
	newer.ResetRaw = "59"

	forged := cycle
	forged.capability[0] ^= 0xff
	if c.ObserveCycle(newer, forged) {
		t.Fatal("forged observation cycle was accepted")
	}
	otherKey := newer
	otherKey.Path = "/api/v1/other"
	if c.ObserveCycle(otherKey, cycle) {
		t.Fatal("cross-key observation cycle was accepted")
	}
	other := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	other.Observe(observation)
	if other.ObserveCycle(newer, cycle) {
		t.Fatal("cross-coordinator observation cycle was accepted")
	}
	if !c.ObserveCycle(newer, cycle) {
		t.Fatal("valid observation cycle was rejected")
	}
	if c.ObserveCycle(newer, cycle) {
		t.Fatal("observation cycle replay was accepted")
	}
	staleGeneration := c.BeginObservation(observation.Path)
	advanceCycle := c.BeginObservation(observation.Path)
	newWindow := budget(now.Add(61*time.Second), 6)
	if !c.ObserveCycle(newWindow, advanceCycle) {
		t.Fatal("new-window cycle was rejected")
	}
	if c.ObserveCycle(newWindow, staleGeneration) {
		t.Fatal("cross-generation observation cycle was accepted")
	}
}

func TestObservationCycleIssuanceIsBoundedAndCollisionSafe(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	observation := budget(now, 20)

	collision := newBudgetCoordinatorWithEntropyAndClock(repeatingEntropy(7), func() time.Time { return now })
	collision.Observe(observation)
	if cycle := collision.BeginObservation(observation.Path); cycle == (ObservationCycle{}) {
		t.Fatal("first repeated capability was not issued")
	}
	if cycle := collision.BeginObservation(observation.Path); cycle != (ObservationCycle{}) {
		t.Fatalf("repeated observation capability was reissued: %+v", cycle)
	}

	bounded := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	bounded.Observe(observation)
	state := bounded.endpoints[observation.Path]
	for i := 0; i < maxObservationCyclesPerGeneration; i++ {
		var capability [32]byte
		binary.BigEndian.PutUint64(capability[24:], uint64(i+1))
		state.issuedObservationCycles[capability] = struct{}{}
	}
	bounded.endpoints[observation.Path] = state
	if cycle := bounded.BeginObservation(observation.Path); cycle != (ObservationCycle{}) {
		t.Fatalf("observation cycle cap was bypassed: %+v", cycle)
	}
	if grant := bounded.TryAcquire(observation.Path, PollEmergencyExit, now); !grant.Allowed || grant.Reason != BudgetSafetyPriority {
		t.Fatalf("observation cycle cap blocked safety class: %+v", grant)
	}
}

func TestManualObserveNeverGainsReconciliationAuthority(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	observation := budget(now, 6)
	observation.ResetRaw = "60"
	c.Observe(observation)
	grant := c.TryAcquire(observation.Path, PollCandidate, now)
	if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
		t.Fatalf("initial lifecycle failed: %+v", grant)
	}
	manual := observation
	manual.ObservedAt = now.Add(time.Second)
	manual.ResetRaw = "59"
	c.Observe(manual)
	if next := c.TryAcquire(observation.Path, PollCandidate, manual.ObservedAt); next.Allowed {
		t.Fatalf("manual observation reconciled a commitment: %+v", next)
	}
	manualNewWindow := budget(now.Add(61*time.Second), 6)
	c.Observe(manualNewWindow)
	if state := c.endpoints[observation.Path]; state.generation != 1 {
		t.Fatalf("manual observation advanced generation to %d", state.generation)
	}
	if next := c.TryAcquire(observation.Path, PollCandidate, manualNewWindow.ObservedAt); next.Allowed {
		t.Fatalf("manual new-window observation reconciled a commitment: %+v", next)
	}
}

func TestCompletionSequenceExhaustionFailsClosed(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	observation := budget(now, 6)
	c.Observe(observation)
	grant := c.TryAcquire(observation.Path, PollCandidate, now)
	if !grant.Allowed {
		t.Fatalf("initial grant failed: %+v", grant)
	}
	c.completionSequence = ^uint64(0)
	if c.Complete(observation.Path, grant.Commitment) {
		t.Fatal("completion sequence wrapped")
	}
	if next := c.TryAcquire(observation.Path, PollCandidate, now); next.Allowed || next.Reason != BudgetReserved {
		t.Fatalf("sequence exhaustion released in-flight capacity: %+v", next)
	}
	if safety := c.TryAcquire(observation.Path, PollProtection, now); !safety.Allowed || safety.Reason != BudgetSafetyPriority {
		t.Fatalf("sequence exhaustion blocked safety class: %+v", safety)
	}
}

func TestCommitmentIssueCapIsAbsoluteAndResetScoped(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	reset := now.Add(time.Hour)
	maxInt := int(^uint(0) >> 1)
	observation := official.RateBudget{
		Path: "/api/v1/rankings", Limit: maxInt, Remaining: maxInt,
		Reset: reset, ResetRaw: strconv.FormatInt(reset.Unix(), 10), ResetKind: official.ResetEpoch,
		ObservedAt: now, Reported: true,
	}
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	c.Observe(observation)
	for i := 0; i < maxIssuedCommitmentsPerGeneration; i++ {
		grant := c.TryAcquire(observation.Path, PollCandidate, observation.ObservedAt)
		if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
			t.Fatalf("lifecycle %d failed: %+v", i, grant)
		}
		cycle := c.BeginObservation(observation.Path)
		observation.ObservedAt = observation.ObservedAt.Add(time.Microsecond)
		if !c.ObserveCycle(observation, cycle) {
			t.Fatalf("reconciliation cycle %d failed", i)
		}
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, observation.ObservedAt); grant.Allowed || grant.Reason != BudgetTokenUnavailable {
		t.Fatalf("reported MaxInt bypassed absolute issue cap: %+v", grant)
	}
	if grant := c.TryAcquire(observation.Path, PollEmergencyExit, observation.ObservedAt); !grant.Allowed || grant.Reason != BudgetSafetyPriority {
		t.Fatalf("issue cap blocked safety class: %+v", grant)
	}

	cycle := c.BeginObservation(observation.Path)
	newReset := reset.Add(time.Hour)
	newWindow := observation
	newWindow.ObservedAt = reset.Add(time.Second)
	newWindow.Reset = newReset
	newWindow.ResetRaw = strconv.FormatInt(newReset.Unix(), 10)
	if !c.ObserveCycle(newWindow, cycle) {
		t.Fatal("new-window cycle was rejected")
	}
	if grant := c.TryAcquire(observation.Path, PollCandidate, newWindow.ObservedAt); !grant.Allowed {
		t.Fatalf("proven reset did not clear the generation cap: %+v", grant)
	}
}

func TestManualNewWindowCannotResetIssuedCapWhenCommitmentsAreEmpty(t *testing.T) {
	now := at(t, "2026-08-01T00:00:00Z")
	reset := now.Add(time.Hour)
	maxInt := int(^uint(0) >> 1)
	observation := official.RateBudget{
		Path: "/api/v1/rankings", Limit: maxInt, Remaining: maxInt,
		Reset: reset, ResetRaw: strconv.FormatInt(reset.Unix(), 10), ResetKind: official.ResetEpoch,
		ObservedAt: now, Reported: true,
	}
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return now })
	c.Observe(observation)
	for i := 0; i < maxIssuedCommitmentsPerGeneration; i++ {
		grant := c.TryAcquire(observation.Path, PollCandidate, observation.ObservedAt)
		if !grant.Allowed || !c.Complete(observation.Path, grant.Commitment) {
			t.Fatalf("lifecycle %d failed: %+v", i, grant)
		}
		cycle := c.BeginObservation(observation.Path)
		observation.ObservedAt = observation.ObservedAt.Add(time.Microsecond)
		if !c.ObserveCycle(observation, cycle) {
			t.Fatalf("reconciliation cycle %d failed", i)
		}
	}
	if state := c.endpoints[observation.Path]; len(state.commitments) != 0 || len(state.issued) != maxIssuedCommitmentsPerGeneration {
		t.Fatalf("precondition commitments=%d issued=%d", len(state.commitments), len(state.issued))
	}
	var authoritativeCycle ObservationCycle
	for len(c.endpoints[observation.Path].issuedObservationCycles) < maxObservationCyclesPerGeneration {
		authoritativeCycle = c.BeginObservation(observation.Path)
		if authoritativeCycle == (ObservationCycle{}) {
			t.Fatal("observation cycle cap reached before its documented bound")
		}
	}

	newReset := reset.Add(time.Hour)
	manual := observation
	manual.ObservedAt = reset.Add(time.Second)
	manual.Reset = newReset
	manual.ResetRaw = strconv.FormatInt(newReset.Unix(), 10)
	c.Observe(manual)
	state := c.endpoints[observation.Path]
	if state.generation != 1 || len(state.issued) != maxIssuedCommitmentsPerGeneration ||
		len(state.issuedObservationCycles) != maxObservationCyclesPerGeneration {
		t.Fatalf("manual new-window observation reset authority: generation=%d issued=%d cycles=%d",
			state.generation, len(state.issued), len(state.issuedObservationCycles))
	}
	if extra := c.BeginObservation(observation.Path); extra != (ObservationCycle{}) {
		t.Fatalf("manual observation cleared cycle cap: %+v", extra)
	}
	if safety := c.TryAcquire(observation.Path, PollEmergencyExit, manual.ObservedAt); !safety.Allowed || safety.Reason != BudgetSafetyPriority {
		t.Fatalf("manual denial blocked safety class: %+v", safety)
	}

	authoritative := manual
	authoritative.ObservedAt = manual.ObservedAt.Add(time.Second)
	if !c.ObserveCycle(authoritative, authoritativeCycle) {
		t.Fatal("authoritative new-window cycle was rejected")
	}
	if state := c.endpoints[observation.Path]; state.generation != 2 || len(state.issued) != 0 {
		t.Fatalf("causal cycle did not reset authority: generation=%d issued=%d", state.generation, len(state.issued))
	}
}

func TestResetSemanticsMatchOfficialParserBoundaries(t *testing.T) {
	observedAt := at(t, "2026-08-01T00:00:00Z")
	validEpoch := observedAt.Add(time.Hour)
	tests := []struct {
		name      string
		raw       string
		kind      official.ResetKind
		reset     time.Time
		observed  time.Time
		wantValid bool
	}{
		{name: "valid delta", raw: "60", kind: official.ResetDelta, reset: observedAt.Add(time.Minute), observed: observedAt, wantValid: true},
		{name: "delta max ahead inclusive", raw: "86400", kind: official.ResetDelta, reset: observedAt.Add(24 * time.Hour), observed: observedAt, wantValid: true},
		{name: "valid epoch", raw: strconv.FormatInt(validEpoch.Unix(), 10), kind: official.ResetEpoch, reset: validEpoch, observed: observedAt, wantValid: true},
		{name: "epoch max behind inclusive", raw: strconv.FormatInt(observedAt.Add(-time.Minute).Unix(), 10), kind: official.ResetEpoch, reset: observedAt.Add(-time.Minute), observed: observedAt, wantValid: true},
		{name: "wrapping delta", raw: "36028797018963969", kind: official.ResetDelta, reset: observedAt.Add(time.Second), observed: observedAt},
		{name: "threshold mislabeled delta", raw: "1000000000", kind: official.ResetDelta, reset: observedAt.Add(time.Duration(1_000_000_000) * time.Second), observed: observedAt},
		{name: "delta raw mislabeled epoch", raw: "60", kind: official.ResetEpoch, reset: time.Unix(60, 0).UTC(), observed: time.Unix(60, 0).UTC()},
		{name: "noncanonical raw", raw: " 60 ", kind: official.ResetDelta, reset: observedAt.Add(time.Minute), observed: observedAt},
		{name: "delta beyond max ahead", raw: "86401", kind: official.ResetDelta, reset: observedAt.Add(24*time.Hour + time.Second), observed: observedAt},
		{name: "implausible epoch behind", raw: "1000000000", kind: official.ResetEpoch, reset: time.Unix(1_000_000_000, 0).UTC(), observed: observedAt},
		{name: "epoch beyond max ahead", raw: strconv.FormatInt(observedAt.Add(24*time.Hour+time.Second).Unix(), 10), kind: official.ResetEpoch, reset: observedAt.Add(24*time.Hour + time.Second), observed: observedAt},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			budget := official.RateBudget{ResetRaw: tc.raw, ResetKind: tc.kind, Reset: tc.reset, ObservedAt: tc.observed}
			if got := validResetSemantics(budget); got != tc.wantValid {
				t.Fatalf("validResetSemantics() = %v, want %v for raw=%q kind=%q reset=%s", got, tc.wantValid, tc.raw, tc.kind, tc.reset)
			}
		})
	}
}

func TestDeltaWindowToleranceComparisonIsMinIntSafe(t *testing.T) {
	anchor := at(t, "2026-08-01T00:00:00Z")
	state := endpointBudget{
		trustedResetAnchor: anchor,
		trustedResetKind:   official.ResetDelta,
		hasTrustedReset:    true,
	}
	observation := official.RateBudget{
		ObservedAt: anchor.Add(-time.Second),
		Reset:      anchor.Add(time.Duration(-1 << 63)),
		ResetKind:  official.ResetDelta,
	}
	if got := classifyBudgetWindow(state, observation); got != budgetWindowConflict {
		t.Fatalf("MinInt delta difference classified as %v, want conflict", got)
	}
}

func TestDeltaWindowToleranceBoundariesAreInclusiveAndFixed(t *testing.T) {
	anchor := at(t, "2026-08-01T00:00:00Z")
	state := endpointBudget{
		trustedResetAnchor: anchor,
		trustedResetKind:   official.ResetDelta,
		hasTrustedReset:    true,
	}
	tests := []struct {
		name  string
		reset time.Time
		want  budgetWindowRelation
	}{
		{name: "lower inclusive", reset: anchor.Add(-deltaResetTolerance), want: budgetWindowSame},
		{name: "upper inclusive", reset: anchor.Add(deltaResetTolerance), want: budgetWindowSame},
		{name: "below lower", reset: anchor.Add(-deltaResetTolerance - time.Nanosecond), want: budgetWindowConflict},
		{name: "above upper", reset: anchor.Add(deltaResetTolerance + time.Nanosecond), want: budgetWindowConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			observation := official.RateBudget{
				ObservedAt: anchor.Add(-time.Second),
				Reset:      tc.reset,
				ResetKind:  official.ResetDelta,
			}
			if got := classifyBudgetWindow(state, observation); got != tc.want {
				t.Fatalf("classifyBudgetWindow(reset=%s) = %v, want %v", tc.reset, got, tc.want)
			}
		})
	}
}

func TestDeltaResetDriftStaysInOneWindowWithEarliestDeadline(t *testing.T) {
	base := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return base })
	first := budget(base.Add(100*time.Millisecond), 6)
	first.ResetRaw = "60" // derived boundary 00:01:00.100
	c.Observe(first)
	grant := c.TryAcquire(first.Path, PollCandidate, first.ObservedAt)
	if !grant.Allowed || !c.Complete(first.Path, grant.Commitment) {
		t.Fatalf("initial lifecycle failed: %+v", grant)
	}

	cycle := c.BeginObservation(first.Path)
	second := first
	second.ObservedAt = base.Add(900 * time.Millisecond)
	second.ResetRaw = "59"
	second.Reset = second.ObservedAt.Add(59 * time.Second) // 00:00:59.900
	if !c.ObserveCycle(second, cycle) {
		t.Fatal("compatible delta cycle was rejected")
	}
	state := c.endpoints[first.Path]
	if state.generation != 1 {
		t.Fatalf("subsecond delta drift created generation %d", state.generation)
	}
	if !state.trustedReset.Equal(second.Reset) {
		t.Fatalf("trusted reset = %s, want conservative earliest %s", state.trustedReset, second.Reset)
	}
	if next := c.TryAcquire(first.Path, PollCandidate, second.ObservedAt); !next.Allowed || !next.Reset.Equal(second.Reset) {
		t.Fatalf("same-window delta did not reconcile at earliest deadline: %+v", next)
	}
}

func TestDeltaResetStartsNewGenerationOnlyAfterPriorBoundary(t *testing.T) {
	base := at(t, "2026-08-01T00:00:00Z")
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return base })
	first := budget(base.Add(100*time.Millisecond), 20)
	first.ResetRaw = "60"
	c.Observe(first)

	preBoundaryCycle := c.BeginObservation(first.Path)
	preBoundary := first
	preBoundary.ObservedAt = base.Add(900 * time.Millisecond)
	preBoundary.ResetRaw = "60"
	preBoundary.Reset = preBoundary.ObservedAt.Add(time.Minute) // 800ms later derived instant
	if !c.ObserveCycle(preBoundary, preBoundaryCycle) {
		t.Fatal("compatible pre-boundary delta was rejected")
	}
	if got := c.endpoints[first.Path].generation; got != 1 {
		t.Fatalf("pre-boundary drift created generation %d", got)
	}

	boundaryCycle := c.BeginObservation(first.Path)
	newWindow := preBoundary
	newWindow.ObservedAt = base.Add(62 * time.Second)
	newWindow.ResetRaw = "60"
	newWindow.Reset = newWindow.ObservedAt.Add(time.Minute)
	if !c.ObserveCycle(newWindow, boundaryCycle) {
		t.Fatal("new-window delta was rejected")
	}
	if got := c.endpoints[first.Path].generation; got != 2 {
		t.Fatalf("proven boundary did not create generation 2: %d", got)
	}
}

func TestResetWindowConflictingProvenanceFailsClosed(t *testing.T) {
	base := at(t, "2026-08-01T00:00:00Z")
	for _, tc := range []struct {
		name   string
		mutate func(*official.RateBudget)
	}{
		{name: "delta raw outside tolerance", mutate: func(b *official.RateBudget) {
			b.ResetRaw = "120"
			b.Reset = b.ObservedAt.Add(120 * time.Second)
		}},
		{name: "reset kind changes before boundary", mutate: func(b *official.RateBudget) {
			b.ResetKind = official.ResetEpoch
			b.Reset = base.Add(61 * time.Second)
			b.ResetRaw = strconv.FormatInt(b.Reset.Unix(), 10)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return base })
			first := budget(base, 20)
			first.ResetRaw = "60"
			c.Observe(first)
			conflict := first
			conflict.ObservedAt = base.Add(time.Second)
			tc.mutate(&conflict)
			c.Observe(conflict)
			if grant := c.TryAcquire(first.Path, PollCandidate, conflict.ObservedAt); grant.Allowed || grant.Reason != BudgetConflictingProvenance {
				t.Fatalf("conflicting reset provenance was accepted: %+v", grant)
			}
		})
	}
}

func TestEpochResetWindowIdentityRemainsExact(t *testing.T) {
	base := at(t, "2026-08-01T00:00:00Z")
	reset := base.Add(time.Minute)
	first := official.RateBudget{
		Path: "/api/v1/rankings", Limit: 100, Remaining: 20,
		Reset: reset, ResetRaw: strconv.FormatInt(reset.Unix(), 10), ResetKind: official.ResetEpoch,
		ObservedAt: base, Reported: true,
	}
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return base })
	c.Observe(first)
	conflict := first
	conflict.ObservedAt = base.Add(time.Second)
	conflict.Reset = reset.Add(time.Second)
	conflict.ResetRaw = strconv.FormatInt(conflict.Reset.Unix(), 10)
	c.Observe(conflict)
	if grant := c.TryAcquire(first.Path, PollCandidate, conflict.ObservedAt); grant.Allowed || grant.Reason != BudgetConflictingProvenance {
		t.Fatalf("different pre-boundary epoch was treated as the same window: %+v", grant)
	}
}
