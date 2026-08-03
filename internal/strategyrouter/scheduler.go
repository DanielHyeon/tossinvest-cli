package strategyrouter

import (
	"crypto/sha256"
	"errors"
	"sync"
	"time"
)

type MarketRecord struct {
	Market              Market
	Desired             DesiredState
	Effective           DesiredState
	Revision            uint64
	LockID              string
	CalendarGeneration  string
	CalendarDigest      string
	Timezone            string
	SessionScope        string
	ActivationDigest    string
	ActivationExpiresAt time.Time
	ConfigVersion       string
	UpdatedActor        string
	UpdatedAt           time.Time
	Runtime             RuntimeState
	seal                [32]byte
}

type marketRecordInput struct {
	Market              Market
	Desired             DesiredState
	Effective           DesiredState
	Revision            uint64
	LockID              string
	CalendarGeneration  string
	CalendarDigest      string
	Timezone            string
	SessionScope        string
	ActivationDigest    string
	ActivationExpiresAt time.Time
	ConfigVersion       string
	UpdatedActor        string
	UpdatedAt           time.Time
	Runtime             RuntimeState
}

func newMarketRecord(input marketRecordInput) (MarketRecord, error) {
	record := MarketRecord{
		Market: input.Market, Desired: input.Desired, Effective: input.Effective, Revision: input.Revision, LockID: input.LockID,
		CalendarGeneration: input.CalendarGeneration, CalendarDigest: input.CalendarDigest, Timezone: input.Timezone, SessionScope: input.SessionScope,
		ActivationDigest: input.ActivationDigest, ActivationExpiresAt: input.ActivationExpiresAt.UTC(), ConfigVersion: input.ConfigVersion,
		UpdatedActor: input.UpdatedActor, UpdatedAt: input.UpdatedAt.UTC(), Runtime: input.Runtime,
	}
	if record.Runtime == "" {
		record.Runtime = RuntimeUnobserved
	}
	if !validMarket(record.Market) || !validDesiredState(record.Desired) || !validDesiredState(record.Effective) || record.Revision == 0 {
		return MarketRecord{}, errors.New("strategyrouter: invalid market record scope/state")
	}
	if boundedNonEmpty("calendar generation", record.CalendarGeneration) != nil || boundedNonEmpty("calendar digest", record.CalendarDigest) != nil ||
		boundedNonEmpty("session scope", record.SessionScope) != nil || boundedNonEmpty("config version", record.ConfigVersion) != nil ||
		boundedNonEmpty("updated actor", record.UpdatedActor) != nil || record.UpdatedAt.IsZero() {
		return MarketRecord{}, errors.New("strategyrouter: incomplete market record evidence")
	}
	wantTimezone := map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[record.Market]
	if record.Timezone != wantTimezone {
		return MarketRecord{}, errors.New("strategyrouter: market timezone mismatch")
	}
	if record.Desired == StateOn || record.Effective == StateOn {
		if record.Desired != StateOn || record.Effective != StateOn || boundedNonEmpty("activation digest", record.ActivationDigest) != nil || !record.ActivationExpiresAt.After(record.UpdatedAt) {
			return MarketRecord{}, errors.New("strategyrouter: invalid active market record")
		}
	} else if record.ActivationDigest != "" {
		return MarketRecord{}, errors.New("strategyrouter: OFF record carries activation")
	}
	record.seal = marketRecordSeal(record)
	return record, nil
}

func DefaultMarketRecord(market Market) MarketRecord {
	record := MarketRecord{Market: market, Desired: StateOff, Effective: StateOff, Revision: 1, Runtime: RuntimeUnobserved}
	if !validMarket(market) {
		return MarketRecord{}
	}
	record.seal = marketRecordSeal(record)
	return record
}

func marketRecordSeal(record MarketRecord) [32]byte {
	h := sha256.New()
	for _, value := range []string{
		string(record.Market), string(record.Desired), string(record.Effective), record.LockID, record.CalendarGeneration,
		record.CalendarDigest, record.Timezone, record.SessionScope, record.ActivationDigest, record.ActivationExpiresAt.UTC().Format(time.RFC3339Nano),
		record.ConfigVersion, record.UpdatedActor, record.UpdatedAt.UTC().Format(time.RFC3339Nano), string(record.Runtime),
	} {
		writeString(h, value)
	}
	writeUint64(h, record.Revision)
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

func validMarketRecord(record MarketRecord) bool {
	if !validMarket(record.Market) || !validDesiredState(record.Desired) || !validDesiredState(record.Effective) || record.Revision == 0 || record.Runtime != RuntimeUnobserved || record.seal != marketRecordSeal(record) {
		return false
	}
	if record.Desired == StateOn || record.Effective == StateOn {
		return record.Desired == StateOn && record.Effective == StateOn && record.ActivationDigest != "" && record.ActivationExpiresAt.After(record.UpdatedAt) &&
			record.CalendarGeneration != "" && record.CalendarDigest != "" && record.ConfigVersion != "" && record.UpdatedActor != "" && record.SessionScope != ""
	}
	return record.Desired == StateOff && record.Effective == StateOff && record.ActivationDigest == ""
}

type SchedulerState struct {
	Records                  map[Market]MarketRecord
	SelectedMarket           Market
	CombinedActivationDigest string
	MigrationVersion         string
	MigrationCode            RefusalCode
}

func NewSchedulerState() SchedulerState {
	return SchedulerState{Records: map[Market]MarketRecord{MarketKR: DefaultMarketRecord(MarketKR), MarketUS: DefaultMarketRecord(MarketUS)}}
}

func cloneSchedulerState(state SchedulerState) SchedulerState {
	clone := state
	clone.Records = make(map[Market]MarketRecord, len(state.Records))
	for market, record := range state.Records {
		clone.Records[market] = record
	}
	return clone
}

func ValidSchedulerState(state SchedulerState) bool {
	if len(state.Records) != 2 || state.SelectedMarket != "" || state.CombinedActivationDigest != "" {
		return false
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		record, ok := state.Records[market]
		if !ok || record.Market != market || !validMarketRecord(record) {
			return false
		}
	}
	return true
}

type CASResult struct {
	Code RefusalCode
}

func CASMarketRecord(state SchedulerState, market Market, expectedRevision uint64, expectedLock string, candidate MarketRecord) (SchedulerState, CASResult) {
	if !ValidSchedulerState(state) || !validMarket(market) || !validMarketRecord(candidate) || candidate.Market != market {
		return cloneSchedulerState(state), CASResult{Code: RefusalInvalid}
	}
	current := state.Records[market]
	if current.Revision != expectedRevision || current.LockID != expectedLock || candidate.Revision != current.Revision+1 {
		return cloneSchedulerState(state), CASResult{Code: RefusalVersionConflict}
	}
	next := cloneSchedulerState(state)
	next.Records[market] = candidate
	return next, CASResult{}
}

func RollbackMarketRecord(state SchedulerState, market Market, expectedRevision uint64, expectedLock string, target MarketRecord, actor string, at time.Time) (SchedulerState, CASResult) {
	if !ValidSchedulerState(state) || !validMarketRecord(target) || target.Market != market || boundedNonEmpty("rollback actor", actor) != nil || at.IsZero() {
		return cloneSchedulerState(state), CASResult{Code: RefusalInvalid}
	}
	if target.Desired != StateOff || target.Effective != StateOff || target.ActivationDigest != "" {
		return cloneSchedulerState(state), CASResult{Code: RefusalInvalid}
	}
	current := state.Records[market]
	if current.Revision != expectedRevision || current.LockID != expectedLock {
		return cloneSchedulerState(state), CASResult{Code: RefusalVersionConflict}
	}
	candidate := target
	candidate.Revision = current.Revision + 1
	candidate.UpdatedActor = actor
	candidate.UpdatedAt = at.UTC()
	candidate.Runtime = RuntimeUnobserved
	candidate.seal = marketRecordSeal(candidate)
	return CASMarketRecord(state, market, expectedRevision, expectedLock, candidate)
}

type MarketRecordStore struct {
	mu    sync.Mutex
	state SchedulerState
}

func NewMarketRecordStore(state SchedulerState) *MarketRecordStore {
	if !ValidSchedulerState(state) {
		state = NewSchedulerState()
	}
	return &MarketRecordStore{state: cloneSchedulerState(state)}
}

func (store *MarketRecordStore) Snapshot() SchedulerState {
	store.mu.Lock()
	defer store.mu.Unlock()
	return cloneSchedulerState(store.state)
}

func (store *MarketRecordStore) CAS(market Market, expectedRevision uint64, expectedLock string, candidate MarketRecord) CASResult {
	store.mu.Lock()
	defer store.mu.Unlock()
	next, result := CASMarketRecord(store.state, market, expectedRevision, expectedLock, candidate)
	if result.Code == RefusalNone {
		store.state = next
	}
	return result
}

type MarketLifecycle string

const (
	LifecycleOff   MarketLifecycle = "OFF"
	LifecycleReady MarketLifecycle = "READY"
	LifecycleStale MarketLifecycle = "STALE"
)

func EvaluateMarketLifecycle(record MarketRecord, at time.Time) MarketLifecycle {
	if !validMarketRecord(record) || record.Desired != StateOn || record.Effective != StateOn {
		return LifecycleOff
	}
	if at.IsZero() || at.Before(record.UpdatedAt) || !at.Before(record.ActivationExpiresAt) {
		return LifecycleStale
	}
	return LifecycleReady
}
