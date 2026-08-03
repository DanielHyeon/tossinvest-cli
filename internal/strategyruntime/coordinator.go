package strategyruntime

import (
	"crypto/sha256"
	"errors"
	"time"
)

type WorkerState struct {
	Market               Market
	Desired              EntryState
	Effective            EntryState
	Runtime              RuntimeState
	CalendarGeneration   uint64
	CalendarDigest       string
	ActivationGeneration uint64
	ActivationDigest     string
	EvidenceCursor       string
	EvidenceDigest       string
	BudgetKey            string
	FirstRefusal         RefusalCode
	Latched              bool
	RestartAttempt       uint64
	RestartNotBefore     time.Time
	seal                 [32]byte
}

type workerStateInput struct {
	Market               Market
	Desired              EntryState
	Effective            EntryState
	Runtime              RuntimeState
	CalendarGeneration   uint64
	CalendarDigest       string
	ActivationGeneration uint64
	ActivationDigest     string
	EvidenceCursor       string
	EvidenceDigest       string
	BudgetKey            string
}

func defaultWorker(market Market) WorkerState {
	worker := WorkerState{Market: market, Desired: EntryOff, Effective: EntryOff, Runtime: RuntimeUnobserved}
	worker.seal = workerSeal(worker)
	return worker
}

func newWorkerState(input workerStateInput) (WorkerState, error) {
	if !validMarket(input.Market) || input.Desired != EntryOn || input.Effective != EntryOn || input.Runtime != RuntimeObserved ||
		input.CalendarGeneration == 0 || !validDigest(input.CalendarDigest) || input.ActivationGeneration == 0 || !validDigest(input.ActivationDigest) ||
		!validIdentity(input.EvidenceCursor) || !validDigest(input.EvidenceDigest) || !validIdentity(input.BudgetKey) {
		return WorkerState{}, errors.New("strategyruntime: incomplete worker authority")
	}
	worker := WorkerState{Market: input.Market, Desired: input.Desired, Effective: input.Effective, Runtime: input.Runtime,
		CalendarGeneration: input.CalendarGeneration, CalendarDigest: input.CalendarDigest, ActivationGeneration: input.ActivationGeneration,
		ActivationDigest: input.ActivationDigest, EvidenceCursor: input.EvidenceCursor, EvidenceDigest: input.EvidenceDigest, BudgetKey: input.BudgetKey}
	worker.seal = workerSeal(worker)
	return worker, nil
}

func workerSeal(worker WorkerState) [32]byte {
	hash := sha256.New()
	for _, value := range []string{string(worker.Market), string(worker.Desired), string(worker.Effective), string(worker.Runtime), worker.CalendarDigest,
		worker.ActivationDigest, worker.EvidenceCursor, worker.EvidenceDigest, worker.BudgetKey, string(worker.FirstRefusal), formatTime(worker.RestartNotBefore)} {
		writeString(hash, value)
	}
	for _, value := range []uint64{worker.CalendarGeneration, worker.ActivationGeneration, worker.RestartAttempt} {
		writeUint64(hash, value)
	}
	if worker.Latched {
		writeString(hash, "latched")
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validWorker(worker WorkerState) bool {
	if !validMarket(worker.Market) || worker.seal != workerSeal(worker) {
		return false
	}
	if worker.Desired == EntryOff && worker.Effective == EntryOff && worker.Runtime == RuntimeUnobserved {
		return worker.CalendarGeneration == 0 && worker.CalendarDigest == "" && worker.ActivationGeneration == 0 && worker.ActivationDigest == "" &&
			worker.EvidenceCursor == "" && worker.EvidenceDigest == "" && worker.BudgetKey == "" && worker.FirstRefusal == RefusalNone &&
			!worker.Latched && worker.RestartAttempt == 0 && worker.RestartNotBefore.IsZero()
	}
	effectiveMatchesLatch := (!worker.Latched && worker.Effective == EntryOn) || (worker.Latched && worker.Effective == EntryOff)
	latchedRefusalValid := !worker.Latched || (worker.FirstRefusal != RefusalNone && validRefusalCode(worker.FirstRefusal))
	return worker.Desired == EntryOn && effectiveMatchesLatch && latchedRefusalValid && worker.Runtime == RuntimeObserved &&
		worker.CalendarGeneration > 0 && validDigest(worker.CalendarDigest) && worker.ActivationGeneration > 0 && validDigest(worker.ActivationDigest) &&
		validIdentity(worker.EvidenceCursor) && validDigest(worker.EvidenceDigest) && validIdentity(worker.BudgetKey) && validRefusalCode(worker.FirstRefusal)
}

type SafetyState struct {
	FillDetection      bool
	Reconciliation     bool
	Protection         bool
	ReduceOnlyExit     bool
	EmergencyReduction bool
}

func fullSafetyState() SafetyState {
	return SafetyState{FillDetection: true, Reconciliation: true, Protection: true, ReduceOnlyExit: true, EmergencyReduction: true}
}

func (state SafetyState) AllEnabled() bool {
	return state.FillDetection && state.Reconciliation && state.Protection && state.ReduceOnlyExit && state.EmergencyReduction
}

type CoordinatorState struct {
	workers                 map[Market]WorkerState
	safety                  SafetyState
	combinedAuthorityDigest string
	seal                    [32]byte
}

func NewCoordinatorState() CoordinatorState {
	state := CoordinatorState{workers: map[Market]WorkerState{MarketKR: defaultWorker(MarketKR), MarketUS: defaultWorker(MarketUS)}, safety: fullSafetyState()}
	state.seal = coordinatorSeal(state)
	return state
}

func (state CoordinatorState) Worker(market Market) WorkerState { return state.workers[market] }
func (state CoordinatorState) Safety() SafetyState              { return state.safety }
func (state CoordinatorState) CombinedAuthorityDigest() string  { return state.combinedAuthorityDigest }

func cloneCoordinator(state CoordinatorState) CoordinatorState {
	clone := state
	clone.workers = make(map[Market]WorkerState, len(state.workers))
	for market, worker := range state.workers {
		clone.workers[market] = worker
	}
	return clone
}

func replaceWorkerForTest(state CoordinatorState, worker WorkerState) CoordinatorState {
	if !ValidCoordinatorState(state) || !validWorker(worker) {
		return state
	}
	next := cloneCoordinator(state)
	next.workers[worker.Market] = worker
	next.seal = coordinatorSeal(next)
	return next
}

func coordinatorSeal(state CoordinatorState) [32]byte {
	hash := sha256.New()
	writeString(hash, state.combinedAuthorityDigest)
	for _, market := range []Market{MarketKR, MarketUS} {
		worker := state.workers[market]
		writeString(hash, string(worker.seal[:]))
	}
	if state.safety.AllEnabled() {
		writeString(hash, "safety-all")
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func ValidCoordinatorState(state CoordinatorState) bool {
	if len(state.workers) != 2 || state.combinedAuthorityDigest != "" || !state.safety.AllEnabled() || state.seal != coordinatorSeal(state) {
		return false
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		worker, ok := state.workers[market]
		if !ok || worker.Market != market || !validWorker(worker) {
			return false
		}
	}
	return true
}

type WorkerOutcome string

const (
	OutcomeEntryAllowed   WorkerOutcome = "ENTRY_ALLOWED"
	OutcomeWaitMarket     WorkerOutcome = "WAIT_MARKET"
	OutcomeBudgetDeferred WorkerOutcome = "BUDGET_DEFERRED"
	OutcomeCycleFailure   WorkerOutcome = "CYCLE_FAILURE"
	OutcomeAbnormalReturn WorkerOutcome = "ABNORMAL_RETURN"
)

type WorkerCycle struct {
	Outcome WorkerOutcome
	Refusal RefusalCode
}

type WorkerDecision struct {
	Market       Market
	EntryAllowed bool
	Code         RefusalCode
}

type CycleResult struct {
	State          CoordinatorState
	Decisions      map[Market]WorkerDecision
	EntryMutations uint64
}

func ApplyWorkerCycle(state CoordinatorState, observed trustedTime, cycles map[Market]WorkerCycle) CycleResult {
	result := CycleResult{State: cloneCoordinator(state), Decisions: make(map[Market]WorkerDecision, 2)}
	if !ValidCoordinatorState(state) || !validTrustedTime(observed) {
		for _, market := range []Market{MarketKR, MarketUS} {
			result.Decisions[market] = WorkerDecision{Market: market, Code: RefusalInvalid}
		}
		return result
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		cycle, present := cycles[market]
		worker := result.State.workers[market]
		decision := WorkerDecision{Market: market, Code: RefusalDisabled}
		if !present {
			result.Decisions[market] = decision
			continue
		}
		switch cycle.Outcome {
		case OutcomeEntryAllowed:
			if worker.Desired == EntryOn && worker.Effective == EntryOn && !worker.Latched {
				decision.EntryAllowed, decision.Code = true, RefusalNone
			}
		case OutcomeWaitMarket:
			decision.Code = RefusalWaitMarket
		case OutcomeBudgetDeferred:
			decision.Code = RefusalBudgetDeferred
		case OutcomeCycleFailure, OutcomeAbnormalReturn:
			worker.Effective, worker.Latched = EntryOff, true
			if worker.RestartAttempt < ^uint64(0) {
				worker.RestartAttempt++
			}
			backoff := restartBackoff(worker.RestartAttempt)
			worker.RestartNotBefore = restartNotBefore(observed.now, backoff)
			if cycle.Outcome == OutcomeAbnormalReturn {
				decision.Code = RefusalWorkerAbnormal
			} else if validCycleFailureRefusal(cycle.Refusal) {
				decision.Code = cycle.Refusal
			} else {
				decision.Code = RefusalInvalid
			}
			if worker.FirstRefusal == RefusalNone {
				worker.FirstRefusal = decision.Code
			}
			worker.seal = workerSeal(worker)
			result.State.workers[market] = worker
		default:
			decision.Code = RefusalInvalid
		}
		result.Decisions[market] = decision
	}
	result.State.seal = coordinatorSeal(result.State)
	return result
}

func validCycleFailureRefusal(code RefusalCode) bool {
	return validRefusalCode(code) && code != RefusalNone && code != RefusalInvalid && code != RefusalWorkerAbnormal
}

func restartBackoff(attempt uint64) time.Duration {
	const step = 5 * time.Second
	maximumSteps := uint64(MaximumWorkerBackoff / step)
	if attempt >= maximumSteps {
		return MaximumWorkerBackoff
	}
	return time.Duration(attempt) * step
}

func restartNotBefore(observed time.Time, backoff time.Duration) time.Time {
	maximumTime := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	if observed.After(maximumTime.Add(-backoff)) {
		return maximumTime
	}
	return observed.Add(backoff)
}
