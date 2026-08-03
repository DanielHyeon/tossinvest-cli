package strategyrouter

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDefaultRecordsAreIndependentOFFAndUnobserved(t *testing.T) {
	state := NewSchedulerState()
	for _, market := range []Market{MarketKR, MarketUS} {
		record := state.Records[market]
		if record.Market != market || record.Desired != StateOff || record.Effective != StateOff || record.Revision != 1 || record.ActivationDigest != "" || record.Runtime != RuntimeUnobserved {
			t.Fatalf("market=%s default=%+v", market, record)
		}
	}
	if state.SelectedMarket != "" || state.CombinedActivationDigest != "" {
		t.Fatalf("default synthesized market/combined activation: %+v", state)
	}
}

func TestPerMarketCASConflictAndRollbackPreservePeerRevision(t *testing.T) {
	state := NewSchedulerState()
	krUpdate := mustRecordUpdate(t, state.Records[MarketKR], StateOn, "kr-lock", "kr-activation", routerNow.Add(time.Hour))
	next, result := CASMarketRecord(state, MarketKR, 1, "", krUpdate)
	if result.Code != RefusalNone || next.Records[MarketKR].Revision != 2 || next.Records[MarketUS].Revision != 1 {
		t.Fatalf("KR CAS contaminated peer: next=%+v result=%+v", next, result)
	}
	stale, result := CASMarketRecord(next, MarketKR, 1, "", krUpdate)
	if result.Code != RefusalVersionConflict || !reflect.DeepEqual(stale, next) {
		t.Fatalf("stale KR CAS changed state: %+v/%+v", stale, result)
	}
	usUpdate := mustRecordUpdate(t, next.Records[MarketUS], StateOn, "us-lock", "us-activation", routerNow.Add(time.Hour))
	withUS, result := CASMarketRecord(next, MarketUS, 1, "", usUpdate)
	if result.Code != RefusalNone || withUS.Records[MarketUS].Revision != 2 || withUS.Records[MarketKR].Revision != 2 {
		t.Fatalf("US CAS was gated by KR revision: %+v/%+v", withUS, result)
	}
	rolled, result := RollbackMarketRecord(withUS, MarketKR, 2, "kr-lock", state.Records[MarketKR], "rollback", routerNow.Add(time.Minute))
	if result.Code != RefusalNone || rolled.Records[MarketKR].Desired != StateOff || rolled.Records[MarketKR].Revision != 3 || rolled.Records[MarketUS].Revision != 2 {
		t.Fatalf("independent rollback failed: %+v/%+v", rolled, result)
	}
}

func TestConcurrentKRandUSCASCommitIndependently(t *testing.T) {
	store := NewMarketRecordStore(NewSchedulerState())
	var wg sync.WaitGroup
	results := make(chan CASResult, 2)
	for _, market := range []Market{MarketKR, MarketUS} {
		market := market
		wg.Add(1)
		go func() {
			defer wg.Done()
			current := store.Snapshot().Records[market]
			update := mustRecordUpdateNoTest(current, StateOn, string(market)+"-lock", string(market)+"-activation", routerNow.Add(time.Hour))
			results <- store.CAS(market, current.Revision, current.LockID, update)
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		if result.Code != RefusalNone {
			t.Fatalf("peer market CAS conflict=%+v", result)
		}
	}
	snapshot := store.Snapshot()
	if snapshot.Records[MarketKR].Revision != 2 || snapshot.Records[MarketUS].Revision != 2 {
		t.Fatalf("concurrent independent CAS lost update: %+v", snapshot)
	}
}

func TestLegacyMigrationMatrixAndIdempotentRetry(t *testing.T) {
	tests := []struct {
		name      string
		legacy    LegacyState
		wantKR    DesiredState
		wantUS    DesiredState
		wantCode  RefusalCode
		wantUSAct string
	}{
		{name: "disabled", legacy: LegacyState{Disabled: true}, wantKR: StateOff, wantUS: StateOff},
		{name: "verified US", legacy: LegacyState{Enabled: true, SelectedMarket: MarketUS, Verified: true, Record: mustLegacyRecord(t, MarketUS)}, wantKR: StateOff, wantUS: StateOn, wantUSAct: "legacy-US-activation"},
		{name: "combined", legacy: LegacyState{Enabled: true, Combined: true, Verified: true}, wantKR: StateOff, wantUS: StateOff, wantCode: RefusalMigration},
		{name: "unverified", legacy: LegacyState{Enabled: true, SelectedMarket: MarketKR}, wantKR: StateOff, wantUS: StateOff, wantCode: RefusalMigration},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := MigrateLegacy(NewSchedulerState(), test.legacy, "migration-v1")
			if first.State.Records[MarketKR].Desired != test.wantKR || first.State.Records[MarketUS].Desired != test.wantUS || first.Code != test.wantCode || first.State.Records[MarketUS].ActivationDigest != test.wantUSAct || first.State.CombinedActivationDigest != "" {
				t.Fatalf("migration=%+v", first)
			}
			retry := MigrateLegacy(first.State, test.legacy, "migration-v1")
			if !retry.Duplicate || !reflect.DeepEqual(retry.State, first.State) || retry.Code != first.Code {
				t.Fatalf("migration retry diverged: first=%+v retry=%+v", first, retry)
			}
		})
	}
}

func TestCrashBoundaryRestoresOldOrCompleteNewRecord(t *testing.T) {
	old := NewSchedulerState()
	update := mustRecordUpdate(t, old.Records[MarketKR], StateOn, "kr-lock", "kr-act", routerNow.Add(time.Hour))
	newState, result := CASMarketRecord(old, MarketKR, 1, "", update)
	if result.Code != RefusalNone {
		t.Fatal(result.Code)
	}
	for _, recovered := range []SchedulerState{old, newState} {
		if !ValidSchedulerState(recovered) || recovered.Records[MarketUS].Revision != 1 {
			t.Fatalf("partial/crash-contaminated record accepted: %+v", recovered)
		}
	}
}

func mustRecordUpdate(t *testing.T, current MarketRecord, desired DesiredState, lock, activation string, expiry time.Time) MarketRecord {
	t.Helper()
	return mustRecordUpdateNoTest(current, desired, lock, activation, expiry)
}

func mustRecordUpdateNoTest(current MarketRecord, desired DesiredState, lock, activation string, expiry time.Time) MarketRecord {
	record, err := newMarketRecord(marketRecordInput{Market: current.Market, Desired: desired, Effective: desired, Revision: current.Revision + 1, LockID: lock,
		CalendarGeneration: "calendar-v1", CalendarDigest: "calendar-digest", Timezone: map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[current.Market],
		SessionScope: "REGULAR", ActivationDigest: activation, ActivationExpiresAt: expiry, ConfigVersion: "config-v1", UpdatedActor: "human-approved-fixture", UpdatedAt: routerNow})
	if err != nil {
		panic(err)
	}
	return record
}

func TestRollbackCannotReissueONAuthority(t *testing.T) {
	state := NewSchedulerState()
	first := mustRecordUpdate(t, state.Records[MarketKR], StateOn, "kr-lock", "kr-activation", routerNow.Add(time.Hour))
	active, result := CASMarketRecord(state, MarketKR, 1, "", first)
	if result.Code != RefusalNone {
		t.Fatal(result.Code)
	}
	newer := mustRecordUpdate(t, active.Records[MarketKR], StateOn, "kr-lock-2", "kr-activation-2", routerNow.Add(2*time.Hour))
	active, result = CASMarketRecord(active, MarketKR, 2, "kr-lock", newer)
	if result.Code != RefusalNone {
		t.Fatal(result.Code)
	}
	before := active
	after, result := RollbackMarketRecord(active, MarketKR, 3, "kr-lock-2", first, "rollback", routerNow.Add(time.Minute))
	if result.Code != RefusalInvalid || !reflect.DeepEqual(after, before) {
		t.Fatalf("rollback reissued old ON authority: after=%+v result=%+v", after, result)
	}
}

func mustLegacyRecord(t *testing.T, market Market) MarketRecord {
	t.Helper()
	base := DefaultMarketRecord(market)
	return mustRecordUpdateNoTest(base, StateOn, "legacy-lock", "legacy-"+string(market)+"-activation", routerNow.Add(time.Hour))
}
