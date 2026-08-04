package strategyrouter

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSharedEndpointExhaustionSpansMarketAndHorizon(t *testing.T) {
	authority := newQuotaAuthority(func() time.Time { return routerNow })
	key := PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset-1"}
	if err := authority.Install(mustQuotaSnapshot(t, key, 7, 5, 2, 2)); err != nil {
		t.Fatal(err)
	}
	first := authority.Acquire(validAcquire(key, MarketKR, HorizonShort, "request-1"))
	second := authority.Acquire(validAcquire(key, MarketKR, HorizonShort, "request-2"))
	if first.Code != RefusalNone || second.Code != RefusalNone {
		t.Fatalf("shared allowance not issued: %+v %+v", first, second)
	}
	for _, scope := range []struct {
		market  Market
		horizon Horizon
	}{{MarketUS, HorizonShort}, {MarketKR, HorizonWeekly}, {MarketUS, HorizonWeekly}} {
		got := authority.Acquire(validAcquire(key, scope.market, scope.horizon, "exhausted-"+string(scope.market)+"-"+string(scope.horizon)))
		if got.Code != RefusalBudgetDeferred || got.Capability.Token != "" {
			t.Fatalf("scope multiplied physical quota: scope=%+v got=%+v", scope, got)
		}
	}
	status := authority.Status(key)
	if status.Issued != 2 || status.Outstanding != 2 || status.SafetyReserve != 5 || status.Available != 0 {
		t.Fatalf("shared counter/reserve changed=%+v", status)
	}
}

func TestDifferentEndpointIsIndependentAndReplayCannotChangeCommitment(t *testing.T) {
	authority := newQuotaAuthority(func() time.Time { return routerNow })
	keyA := PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset-1"}
	keyB := PhysicalQuotaKey{Endpoint: "orders", ResetGeneration: "reset-1"}
	for _, key := range []PhysicalQuotaKey{keyA, keyB} {
		if err := authority.Install(mustQuotaSnapshot(t, key, 2, 1, 1, 1)); err != nil {
			t.Fatal(err)
		}
	}
	capA := authority.Acquire(validAcquire(keyA, MarketKR, HorizonShort, "request-a")).Capability
	if got := authority.Acquire(validAcquire(keyA, MarketUS, HorizonWeekly, "request-b")); got.Code != RefusalBudgetDeferred {
		t.Fatalf("endpoint A not exhausted=%+v", got)
	}
	if got := authority.Acquire(validAcquire(keyB, MarketUS, HorizonWeekly, "request-c")); got.Code != RefusalNone {
		t.Fatalf("endpoint B gated by endpoint A=%+v", got)
	}
	before := authority.Status(keyA)
	if got := authority.Complete(capA, MarketUS, HorizonWeekly); got != RefusalScopeMismatch {
		t.Fatalf("cross-scope capability replay accepted=%s", got)
	}
	wrongReset := capA
	wrongReset.ResetGeneration = "reset-2"
	if got := authority.Complete(wrongReset, MarketKR, HorizonShort); got != RefusalReplay {
		t.Fatalf("reset-generation replay accepted=%s", got)
	}
	if after := authority.Status(keyA); after != before {
		t.Fatalf("replay changed commitment/capacity: before=%+v after=%+v", before, after)
	}
	if got := authority.Complete(capA, MarketKR, HorizonShort); got != RefusalNone {
		t.Fatalf("valid completion refused=%s", got)
	}
	if got := authority.Complete(capA, MarketKR, HorizonShort); got != RefusalReplay {
		t.Fatalf("completion replay accepted=%s", got)
	}
}

func TestConcurrentLastSlotCannotMultiplyQuota(t *testing.T) {
	authority := newQuotaAuthority(func() time.Time { return routerNow })
	key := PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset-concurrent"}
	if err := authority.Install(mustQuotaSnapshot(t, key, 6, 5, 1, 1)); err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			market := MarketKR
			horizon := HorizonShort
			if i%2 == 1 {
				market, horizon = MarketUS, HorizonWeekly
			}
			if got := authority.Acquire(validAcquire(key, market, horizon, "request-"+strconv.Itoa(i))); got.Code == RefusalNone {
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 || authority.Status(key).Issued != 1 {
		t.Fatalf("last slot multiplied: wins=%d status=%+v", wins.Load(), authority.Status(key))
	}
}

func TestSubscopeRetryIsIdempotentButCannotMintCapacity(t *testing.T) {
	authority := newQuotaAuthority(func() time.Time { return routerNow })
	key := PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset-idempotent"}
	if err := authority.Install(mustQuotaSnapshot(t, key, 10, 5, 5, 5)); err != nil {
		t.Fatal(err)
	}
	req := validAcquire(key, MarketKR, HorizonWeekly, "same-request")
	first := authority.Acquire(req)
	retry := authority.Acquire(req)
	if first.Code != RefusalNone || retry.Code != RefusalDuplicate || retry.Capability.Token != first.Capability.Token || authority.Status(key).Issued != 1 {
		t.Fatalf("retry minted quota: first=%+v retry=%+v status=%+v", first, retry, authority.Status(key))
	}
}

func TestCallerTimestampCannotBackdateStaleQuota(t *testing.T) {
	trustedNow := routerNow.Add(2 * time.Minute)
	authority := newQuotaAuthority(func() time.Time { return trustedNow })
	key := PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset-stale"}
	if err := authority.Install(mustQuotaSnapshot(t, key, 10, 5, 5, 5)); err != nil {
		t.Fatal(err)
	}
	request := validAcquire(key, MarketKR, HorizonShort, "backdated")
	request.ObservedAt = routerNow.Add(-time.Second)
	if got := authority.Acquire(request); got.Code != RefusalBudgetDeferred || authority.Status(key).Issued != 0 {
		t.Fatalf("caller backdated stale authority: got=%+v status=%+v", got, authority.Status(key))
	}
}

func mustQuotaSnapshot(t *testing.T, key PhysicalQuotaKey, remaining, reserve, cycleCap, absoluteCap uint64) QuotaSnapshot {
	t.Helper()
	snapshot, err := newQuotaSnapshot(quotaSnapshotInput{Key: key, ReportedRemaining: remaining, SafetyReserve: reserve, ObservationCycleCap: cycleCap, AbsoluteIssuanceCap: absoluteCap,
		ObservedAt: routerNow.Add(-time.Second), FreshUntil: routerNow.Add(time.Minute), Digest: "quota-digest"})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func validAcquire(key PhysicalQuotaKey, market Market, horizon Horizon, requestID string) AcquireRequest {
	return AcquireRequest{Key: key, Market: market, Horizon: horizon, PollClass: PollLowPriority, CoordinatorID: "scheduler-coordinator", RequestID: requestID, ObservedAt: routerNow}
}
