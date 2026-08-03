package protectionlifecycle

import (
	"fmt"
	"sort"
)

type fillRecord struct {
	Quantity    uint64
	Fingerprint string
}

type positionState struct {
	Generation         uint64
	ProtectionRevision uint64
	LifecycleRevision  uint64
	Holdings           uint64
	OtherSellClaims    uint64
	Phase              Phase
	EntryOpen          bool
	EntryLatch         RefusalCode
	Desired            ProtectionOrder
	Observed           ProtectionOrder
	Pending            BrokerCommand
	HasPending         bool
	Fills              map[string]fillRecord
}

type State struct {
	positions     map[PositionKey]positionState
	marketLatches map[Market]RefusalCode
	orphans       map[string]BrokerObservation
	seal          [32]byte
}

func newState(seeds []PositionSeed) (State, error) {
	state := State{
		positions: map[PositionKey]positionState{}, marketLatches: map[Market]RefusalCode{},
		orphans: map[string]BrokerObservation{},
	}
	for _, seed := range seeds {
		key := PositionKey{seed.AccountID, seed.PositionID, seed.Market}
		if !validKey(key) || seed.Generation == 0 || seed.Holdings == 0 || seed.OtherSellClaims >= seed.Holdings {
			return State{}, refuse(RefusalInvalidIdentity, "invalid seed for %q/%q/%q", seed.AccountID, seed.PositionID, seed.Market)
		}
		if _, exists := state.positions[key]; exists {
			return State{}, refuse(RefusalInvalidIdentity, "duplicate position")
		}
		state.positions[key] = positionState{
			Generation: seed.Generation, Holdings: seed.Holdings, OtherSellClaims: seed.OtherSellClaims,
			Phase: Unprotected, EntryOpen: false, Fills: map[string]fillRecord{},
		}
	}
	state.reseal()
	return state, nil
}

func validKey(key PositionKey) bool {
	return validIdentity(key.AccountID) && validIdentity(key.PositionID) && validMarket(key.Market)
}

func (state State) view(key PositionKey) PositionView {
	position := state.positions[key]
	return PositionView{
		Key: key, Generation: position.Generation, ProtectionRevision: position.ProtectionRevision,
		LifecycleRevision: position.LifecycleRevision, Holdings: position.Holdings,
		OtherSellClaims: position.OtherSellClaims, Phase: position.Phase,
		EntryOpen: position.EntryOpen,
		Desired:   position.Desired, Observed: position.Observed,
	}
}

func (state State) marketEntryOpen(market Market) bool { return state.marketLatches[market] == "" }

func cloneState(state State) State {
	clone := State{positions: make(map[PositionKey]positionState, len(state.positions)), marketLatches: make(map[Market]RefusalCode, len(state.marketLatches)), orphans: make(map[string]BrokerObservation, len(state.orphans))}
	for key, position := range state.positions {
		position.Fills = cloneFills(position.Fills)
		clone.positions[key] = position
	}
	for market, code := range state.marketLatches {
		clone.marketLatches[market] = code
	}
	for brokerID, observation := range state.orphans {
		clone.orphans[brokerID] = observation
	}
	clone.reseal()
	return clone
}

func cloneFills(source map[string]fillRecord) map[string]fillRecord {
	result := make(map[string]fillRecord, len(source))
	for id, record := range source {
		result[id] = record
	}
	return result
}

func (state *State) reseal() { state.seal = stateSeal(*state) }

func validState(state State) bool {
	if state.positions == nil || state.marketLatches == nil || state.orphans == nil || state.seal != stateSeal(state) {
		return false
	}
	for key, position := range state.positions {
		if !validKey(key) || position.Generation == 0 || position.Holdings < position.OtherSellClaims || position.Fills == nil {
			return false
		}
		if position.Observed.Status == BrokerActive && position.Observed.Quantity+position.OtherSellClaims > position.Holdings {
			return false
		}
		if !validPositionTruth(state, key, position) {
			return false
		}
	}
	for market := range state.marketLatches {
		if !validMarket(market) {
			return false
		}
	}
	for brokerID, observation := range state.orphans {
		if brokerID == "" || brokerID != observation.BrokerOrderID {
			return false
		}
	}
	return true
}

func stateSeal(state State) [32]byte {
	parts := []string{"protection-lifecycle/v1"}
	keys := make([]PositionKey, 0, len(state.positions))
	for key := range state.positions {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].AccountID != keys[j].AccountID {
			return keys[i].AccountID < keys[j].AccountID
		}
		if keys[i].PositionID != keys[j].PositionID {
			return keys[i].PositionID < keys[j].PositionID
		}
		return keys[i].Market < keys[j].Market
	})
	for _, key := range keys {
		position := state.positions[key]
		parts = append(parts, key.AccountID, key.PositionID, string(key.Market), fmt.Sprint(position.Generation), fmt.Sprint(position.ProtectionRevision), fmt.Sprint(position.LifecycleRevision), fmt.Sprint(position.Holdings), fmt.Sprint(position.OtherSellClaims), string(position.Phase), fmt.Sprint(position.EntryOpen), string(position.EntryLatch))
		parts = append(parts, orderParts(position.Desired)...)
		parts = append(parts, orderParts(position.Observed)...)
		parts = append(parts, commandParts(position.Pending)...)
		parts = append(parts, fmt.Sprint(position.HasPending))
		fillIDs := make([]string, 0, len(position.Fills))
		for fillID := range position.Fills {
			fillIDs = append(fillIDs, fillID)
		}
		sort.Strings(fillIDs)
		for _, fillID := range fillIDs {
			record := position.Fills[fillID]
			parts = append(parts, fillID, fmt.Sprint(record.Quantity), record.Fingerprint)
		}
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		parts = append(parts, string(market), string(state.marketLatches[market]))
	}
	orphanIDs := make([]string, 0, len(state.orphans))
	for brokerID := range state.orphans {
		orphanIDs = append(orphanIDs, brokerID)
	}
	sort.Strings(orphanIDs)
	for _, brokerID := range orphanIDs {
		parts = append(parts, brokerID)
		parts = append(parts, observationParts(state.orphans[brokerID])...)
	}
	return hashParts(parts...)
}

func validPositionTruth(state State, key PositionKey, position positionState) bool {
	switch position.Phase {
	case Unprotected:
		if position.HasPending || position.Observed.Status != "" {
			return false
		}
	case SubmitPending, SubmitUnknown:
		if !validPending(key, position, OperationSubmit) || position.Observed.Status == BrokerActive {
			return false
		}
	case Active:
		if position.HasPending || position.Observed.Status != BrokerActive {
			return false
		}
	case ReplacePending, ReplaceUnknown:
		if !validPending(key, position, OperationReplace) || position.Observed.Status != BrokerActive || position.Pending.BrokerOrderID != position.Observed.BrokerOrderID {
			return false
		}
	case CancelPending, CancelUnknown:
		if !validPending(key, position, OperationCancel) || position.Observed.Status != BrokerActive || position.Pending.BrokerOrderID != position.Observed.BrokerOrderID {
			return false
		}
	case ReconcileRequired:
		// An unknown broker outcome may retain either a pending operation or exact ACTIVE coverage.
	case Terminal:
		if position.HasPending || (position.Observed.Status != BrokerCanceled && position.Observed.Status != BrokerFilled) {
			return false
		}
	default:
		return false
	}
	eligible := position.Phase == Active && !position.HasPending && position.Observed.Status == BrokerActive &&
		position.Observed.Quantity+position.OtherSellClaims == position.Holdings && position.EntryLatch == "" && state.marketLatches[key.Market] == ""
	return position.EntryOpen == eligible
}

func validPending(key PositionKey, position positionState, kind OperationKind) bool {
	command := position.Pending
	if !position.HasPending || command.Kind != kind || command.Position != key || command.Generation != position.Generation || command.Revision != position.ProtectionRevision || command.OperationKey != operationIdentity(key, command.Generation, command.Revision, command.Kind) {
		return false
	}
	if kind == OperationSubmit || kind == OperationReplace {
		return command.Quantity > 0 && command.Trigger > 0 && command.Quantity+position.OtherSellClaims == position.Holdings
	}
	return validIdentity(command.BrokerOrderID)
}

func orderParts(order ProtectionOrder) []string {
	return []string{fmt.Sprint(order.Generation), fmt.Sprint(order.Revision), order.OperationKey, order.BrokerOrderID, string(order.Status), fmt.Sprint(order.Quantity), fmt.Sprint(order.Trigger)}
}

func commandParts(command BrokerCommand) []string {
	return []string{string(command.Kind), command.Position.AccountID, command.Position.PositionID, string(command.Position.Market), fmt.Sprint(command.Generation), fmt.Sprint(command.Revision), command.OperationKey, command.BrokerOrderID, fmt.Sprint(command.Quantity), fmt.Sprint(command.Trigger)}
}

func observationParts(observation BrokerObservation) []string {
	return []string{observation.AccountID, observation.PositionID, string(observation.Market), fmt.Sprint(observation.Generation), fmt.Sprint(observation.Revision), observation.OperationKey, observation.BrokerOrderID, string(observation.Status), fmt.Sprint(observation.Quantity), fmt.Sprint(observation.Trigger)}
}
