package strategyruntime

import (
	"math"
	"reflect"
	"sync"
	"testing"
	"time"
)

var runtimeNow = time.Date(2026, 8, 4, 7, 0, 0, 0, time.UTC)

func TestDefaultCoordinatorShipsPairedMarketsOFF(t *testing.T) {
	state := NewCoordinatorState()
	for _, market := range []Market{MarketKR, MarketUS} {
		worker := state.Worker(market)
		if worker.Market != market || worker.Desired != EntryOff || worker.Effective != EntryOff || worker.Runtime != RuntimeUnobserved || worker.Latched {
			t.Fatalf("market=%s worker=%+v", market, worker)
		}
	}
	if state.CombinedAuthorityDigest() != "" || !state.Safety().AllEnabled() {
		t.Fatalf("default combined/safety state=%+v", state)
	}
}

func TestKRWaitAndUSAllowedAreEvaluatedIndependentlySameCycle(t *testing.T) {
	state := readyCoordinator(t)
	result := ApplyWorkerCycle(state, newTrustedTime(runtimeNow), map[Market]WorkerCycle{
		MarketKR: {Outcome: OutcomeWaitMarket, Refusal: RefusalWaitMarket},
		MarketUS: {Outcome: OutcomeEntryAllowed},
	})
	if result.Decisions[MarketKR].EntryAllowed || result.Decisions[MarketKR].Code != RefusalWaitMarket {
		t.Fatalf("KR wait=%+v", result.Decisions[MarketKR])
	}
	if !result.Decisions[MarketUS].EntryAllowed || result.Decisions[MarketUS].Code != RefusalNone {
		t.Fatalf("KR wait gated US=%+v", result.Decisions[MarketUS])
	}
	if !result.State.Safety().AllEnabled() || result.EntryMutations != 0 {
		t.Fatalf("cycle changed safety/mutated entry=%+v", result)
	}
}

func TestConcurrentDifferentMarketCyclesDoNotShareFailureState(t *testing.T) {
	state := readyCoordinator(t)
	var wg sync.WaitGroup
	results := make(chan CycleResult, 2)
	for _, market := range []Market{MarketKR, MarketUS} {
		market := market
		wg.Add(1)
		go func() {
			defer wg.Done()
			outcome := OutcomeEntryAllowed
			code := RefusalNone
			if market == MarketKR {
				outcome, code = OutcomeAbnormalReturn, RefusalWorkerAbnormal
			}
			results <- ApplyWorkerCycle(state, newTrustedTime(runtimeNow), map[Market]WorkerCycle{market: {Outcome: outcome, Refusal: code}})
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		if !result.State.Safety().AllEnabled() || !ValidCoordinatorState(result.State) {
			t.Fatalf("concurrent pure result invalid=%+v", result)
		}
	}
}

func TestAbnormalWorkerLatchesOnlyThatMarketWithBoundedRestart(t *testing.T) {
	state := readyCoordinator(t)
	beforeUS := state.Worker(MarketUS)
	result := ApplyWorkerCycle(state, newTrustedTime(runtimeNow), map[Market]WorkerCycle{MarketKR: {Outcome: OutcomeAbnormalReturn, Refusal: RefusalWorkerAbnormal}})
	kr := result.State.Worker(MarketKR)
	if kr.Effective != EntryOff || !kr.Latched || kr.RestartAttempt != 1 || kr.RestartNotBefore.Sub(runtimeNow) > MaximumWorkerBackoff {
		t.Fatalf("KR fault not bounded/latching=%+v", kr)
	}
	if got := result.State.Worker(MarketUS); !reflect.DeepEqual(got, beforeUS) {
		t.Fatalf("KR fault contaminated US: before=%+v after=%+v", beforeUS, got)
	}
	if !result.State.Safety().AllEnabled() || result.Decisions[MarketKR].EntryAllowed {
		t.Fatalf("fault stopped safety or allowed entry=%+v", result)
	}
}

func TestMalformedSealedWorkerLatchEffectiveInvariantIsRejected(t *testing.T) {
	state := readyCoordinator(t)
	worker := state.Worker(MarketKR)
	worker.Latched = true
	worker.Effective = EntryOn
	worker.seal = workerSeal(worker)
	if validWorker(worker) {
		t.Fatal("latched worker remained effectively ON")
	}

	worker.Latched = false
	worker.Effective = EntryOff
	worker.seal = workerSeal(worker)
	if validWorker(worker) {
		t.Fatal("unlatched observed worker remained effectively OFF")
	}

	worker = state.Worker(MarketKR)
	worker.Latched = true
	worker.Effective = EntryOff
	worker.FirstRefusal = RefusalNone
	worker.seal = workerSeal(worker)
	if validWorker(worker) {
		t.Fatal("latched worker without a typed first refusal became valid")
	}

	worker = defaultWorker(MarketKR)
	worker.BudgetKey = "noncanonical budget"
	worker.seal = workerSeal(worker)
	if validWorker(worker) {
		t.Fatal("dormant worker accepted non-empty adapter identity")
	}
}

func TestWorkerFailurePreservesFirstTypedRefusalAndSaturatesRestart(t *testing.T) {
	state := readyCoordinator(t)
	first := ApplyWorkerCycle(state, newTrustedTime(runtimeNow), map[Market]WorkerCycle{
		MarketKR: {Outcome: OutcomeCycleFailure, Refusal: RefusalAuthorityDrift},
	})
	if got := first.State.Worker(MarketKR); !got.Latched || got.Effective != EntryOff || got.FirstRefusal != RefusalAuthorityDrift {
		t.Fatalf("first typed refusal not preserved=%+v", got)
	}
	second := ApplyWorkerCycle(first.State, newTrustedTime(runtimeNow.Add(time.Second)), map[Market]WorkerCycle{
		MarketKR: {Outcome: OutcomeAbnormalReturn, Refusal: RefusalBudgetDeferred},
	})
	if got := second.State.Worker(MarketKR); got.FirstRefusal != RefusalAuthorityDrift || second.Decisions[MarketKR].Code != RefusalWorkerAbnormal {
		t.Fatalf("later abnormal overwrote first refusal=%+v decision=%+v", got, second.Decisions[MarketKR])
	}

	worker := state.Worker(MarketKR)
	worker.RestartAttempt = math.MaxUint64
	worker.seal = workerSeal(worker)
	state = replaceWorkerForTest(state, worker)
	saturated := ApplyWorkerCycle(state, newTrustedTime(runtimeNow), map[Market]WorkerCycle{
		MarketKR: {Outcome: OutcomeCycleFailure, Refusal: RefusalWorkerFailure},
	})
	got := saturated.State.Worker(MarketKR)
	if got.RestartAttempt != math.MaxUint64 || got.RestartNotBefore.Before(runtimeNow) || got.RestartNotBefore.Sub(runtimeNow) > MaximumWorkerBackoff {
		t.Fatalf("restart saturation failed=%+v", got)
	}
}

func TestInvalidWorkerFailureRefusalAndRestartTimeOverflowFailClosed(t *testing.T) {
	state := readyCoordinator(t)
	peer := state.Worker(MarketUS)
	nearMaximumTime := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	result := ApplyWorkerCycle(state, newTrustedTime(nearMaximumTime), map[Market]WorkerCycle{
		MarketKR: {Outcome: OutcomeCycleFailure, Refusal: RefusalNone},
	})
	worker := result.State.Worker(MarketKR)
	if !worker.Latched || worker.Effective != EntryOff || worker.FirstRefusal != RefusalInvalid || result.Decisions[MarketKR].Code != RefusalInvalid || worker.RestartNotBefore.Before(nearMaximumTime) {
		t.Fatalf("invalid refusal/time overflow did not fail closed=%+v decision=%+v", worker, result.Decisions[MarketKR])
	}
	if got := result.State.Worker(MarketUS); !reflect.DeepEqual(got, peer) || !result.State.Safety().AllEnabled() {
		t.Fatalf("fail-closed worker contaminated peer/safety=%+v", result)
	}
}

func readyCoordinator(t *testing.T) CoordinatorState {
	t.Helper()
	state := NewCoordinatorState()
	for _, market := range []Market{MarketKR, MarketUS} {
		worker, err := newWorkerState(workerStateInput{Market: market, Desired: EntryOn, Effective: EntryOn, Runtime: RuntimeObserved,
			CalendarGeneration: 1, CalendarDigest: digest("calendar-" + string(market)), ActivationGeneration: 1, ActivationDigest: digest("activation-" + string(market)),
			EvidenceCursor: "cursor-" + string(market), EvidenceDigest: digest("evidence-" + string(market)), BudgetKey: "quotes:" + string(market)})
		if err != nil {
			t.Fatal(err)
		}
		state = replaceWorkerForTest(state, worker)
	}
	return state
}
