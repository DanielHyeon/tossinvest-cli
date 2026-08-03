package protectionlifecycle

func prepareRegister(state State, key PositionKey, quantity, trigger uint64, capability brokerCapability) (State, BrokerCommand, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, BrokerCommand{}, err
	}
	if !state.marketEntryOpen(key.Market) || position.EntryLatch != "" || (position.Phase != Unprotected && position.Phase != Terminal) {
		return state, BrokerCommand{}, refuse(RefusalEntryLatched, "entry is closed")
	}
	if position.HasPending {
		return state, BrokerCommand{}, refuse(RefusalOperationPending, "operation already pending")
	}
	if position.Observed.Status == BrokerActive {
		return state, BrokerCommand{}, refuse(RefusalOperationPending, "protection already active")
	}
	if !capability.exactOperationLookup {
		return state, BrokerCommand{}, refuse(RefusalInvalidObservation, "exact operation lookup unavailable")
	}
	if quantity == 0 || trigger == 0 || quantity+position.OtherSellClaims != position.Holdings {
		return state, BrokerCommand{}, refuse(RefusalSellClaimExceeded, "protection=%d local=%d holdings=%d", quantity, position.OtherSellClaims, position.Holdings)
	}
	next := cloneState(state)
	position = next.positions[key]
	position.ProtectionRevision++
	command := BrokerCommand{Kind: OperationSubmit, Position: key, Generation: position.Generation, Revision: position.ProtectionRevision, Quantity: quantity, Trigger: trigger}
	command.OperationKey = operationIdentity(key, command.Generation, command.Revision, command.Kind)
	position.Desired = commandOrder(command, "", BrokerUnknown)
	position.Pending, position.HasPending, position.Phase, position.EntryOpen, position.EntryLatch = command, true, SubmitPending, false, RefusalOperationPending
	next.positions[key] = position
	next.reseal()
	return next, command, nil
}

func applySubmitResult(state State, command BrokerCommand, observation BrokerObservation) (State, error) {
	position, err := pendingPosition(state, command, OperationSubmit)
	if err != nil {
		return state, err
	}
	next := cloneState(state)
	position = next.positions[command.Position]
	switch observation.Status {
	case BrokerUnknown:
		if observation.OperationKey != "" && observation.OperationKey != command.OperationKey {
			return state, refuse(RefusalInvalidObservation, "unknown result operation key mismatch")
		}
		position.Phase, position.EntryOpen, position.EntryLatch = SubmitUnknown, false, RefusalInvalidObservation
	case BrokerActive:
		if err := validateAccepted(command, observation); err != nil {
			return state, err
		}
		position.Observed = observedOrder(observation)
		position.Desired = position.Observed
		position.Phase, position.EntryOpen, position.EntryLatch, position.HasPending = Active, true, "", false
	default:
		return state, refuse(RefusalInvalidObservation, "submit result status %q", observation.Status)
	}
	next.positions[command.Position] = position
	next.reseal()
	return next, nil
}

func recoverSubmit(state State, key PositionKey, observation BrokerObservation, capability brokerCapability) (State, *BrokerCommand, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, nil, err
	}
	if !capability.exactOperationLookup {
		return latchMarket(state, key.Market, RefusalInvalidObservation), nil, refuse(RefusalInvalidObservation, "exact operation lookup unavailable")
	}
	if !position.HasPending || (position.Pending.Kind != OperationSubmit) {
		if position.Observed.Status == BrokerActive && exactObserved(key, position.Observed, observation) {
			return state, nil, nil
		}
		return latchMarket(state, key.Market, RefusalInvalidObservation), nil, refuse(RefusalInvalidObservation, "no exact submit pending")
	}
	command := position.Pending
	switch observation.Status {
	case BrokerActive:
		if err := validateAccepted(command, observation); err != nil {
			return latchMarket(state, key.Market, RefusalInvalidObservation), nil, err
		}
		next, err := applySubmitResult(state, command, observation)
		return next, nil, err
	case BrokerNotFound:
		if !exactNotFound(command, observation) {
			return latchMarket(state, key.Market, RefusalInvalidObservation), nil, refuse(RefusalInvalidObservation, "lookup key mismatch")
		}
		if capability.idempotentSubmit {
			retry := command
			return state, &retry, nil
		}
		next := cloneState(state)
		position = next.positions[key]
		position.Phase, position.EntryOpen, position.EntryLatch = ReconcileRequired, false, RefusalIdempotencyAbsent
		next.positions[key] = position
		next.reseal()
		return next, nil, nil
	default:
		return latchMarket(state, key.Market, RefusalInvalidObservation), nil, refuse(RefusalInvalidObservation, "submit recovery status %q", observation.Status)
	}
}

func prepareReplace(state State, key PositionKey, quantity, trigger uint64, capability brokerCapability) (State, BrokerCommand, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, BrokerCommand{}, err
	}
	if position.Observed.Status != BrokerActive || position.Phase != Active || position.HasPending {
		return state, BrokerCommand{}, refuse(RefusalNotActive, "exact active protection required")
	}
	if !capability.continuousReplace || !capability.exactBrokerIDLookup {
		return state, BrokerCommand{}, refuse(RefusalContinuousCoverage, "atomic or continuous replacement unattested")
	}
	if trigger < position.Observed.Trigger {
		return state, BrokerCommand{}, refuse(RefusalTriggerRetreat, "trigger %d below active %d", trigger, position.Observed.Trigger)
	}
	if quantity == 0 || quantity+position.OtherSellClaims != position.Holdings {
		return state, BrokerCommand{}, refuse(RefusalSellClaimExceeded, "replacement claim exceeds holdings")
	}
	next := cloneState(state)
	position = next.positions[key]
	position.ProtectionRevision++
	command := BrokerCommand{Kind: OperationReplace, Position: key, Generation: position.Generation, Revision: position.ProtectionRevision, BrokerOrderID: position.Observed.BrokerOrderID, Quantity: quantity, Trigger: trigger}
	command.OperationKey = operationIdentity(key, command.Generation, command.Revision, command.Kind)
	position.Desired = commandOrder(command, command.BrokerOrderID, BrokerUnknown)
	position.Pending, position.HasPending, position.Phase, position.EntryOpen, position.EntryLatch = command, true, ReplacePending, false, RefusalOperationPending
	next.positions[key] = position
	next.reseal()
	return next, command, nil
}

func applyReplaceResult(state State, command BrokerCommand, observation BrokerObservation) (State, error) {
	position, err := pendingPosition(state, command, OperationReplace)
	if err != nil {
		return state, err
	}
	next := cloneState(state)
	position = next.positions[command.Position]
	if observation.Status == BrokerUnknown {
		position.Phase, position.EntryOpen, position.EntryLatch = ReplaceUnknown, false, RefusalInvalidObservation
	} else {
		if err := validateAccepted(command, observation); err != nil {
			return state, err
		}
		if observation.BrokerOrderID != command.BrokerOrderID {
			return state, refuse(RefusalInvalidObservation, "replacement broker ID changed")
		}
		position.Observed, position.Desired = observedOrder(observation), observedOrder(observation)
		position.Phase, position.EntryOpen, position.EntryLatch, position.HasPending = Active, true, "", false
	}
	next.positions[command.Position] = position
	next.reseal()
	return next, nil
}

func recoverReplace(state State, key PositionKey, observation BrokerObservation, capability brokerCapability) (State, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, err
	}
	if !position.HasPending || position.Pending.Kind != OperationReplace || !capability.exactBrokerIDLookup {
		return latchMarket(state, key.Market, RefusalInvalidObservation), refuse(RefusalInvalidObservation, "exact replace pending unavailable")
	}
	if observation.Status != BrokerActive || observation.BrokerOrderID != position.Pending.BrokerOrderID {
		return latchPosition(state, key, RefusalInvalidObservation), refuse(RefusalInvalidObservation, "replacement not exactly active")
	}
	return applyReplaceResult(state, position.Pending, observation)
}

func prepareCancel(state State, key PositionKey, capability brokerCapability) (State, BrokerCommand, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, BrokerCommand{}, err
	}
	if position.Observed.Status != BrokerActive || position.HasPending {
		return state, BrokerCommand{}, refuse(RefusalNotActive, "exact active protection required")
	}
	if !capability.exactBrokerIDLookup || !capability.cancelResultQuery {
		return state, BrokerCommand{}, refuse(RefusalInvalidObservation, "exact cancel query unavailable")
	}
	next := cloneState(state)
	position = next.positions[key]
	position.ProtectionRevision++
	command := BrokerCommand{Kind: OperationCancel, Position: key, Generation: position.Generation, Revision: position.ProtectionRevision, BrokerOrderID: position.Observed.BrokerOrderID, Quantity: position.Observed.Quantity, Trigger: position.Observed.Trigger}
	command.OperationKey = operationIdentity(key, command.Generation, command.Revision, command.Kind)
	position.Pending, position.HasPending, position.Phase, position.EntryOpen, position.EntryLatch = command, true, CancelPending, false, RefusalOperationPending
	next.positions[key] = position
	next.reseal()
	return next, command, nil
}

func applyCancelResult(state State, command BrokerCommand, observation BrokerObservation) (State, error) {
	position, err := pendingPosition(state, command, OperationCancel)
	if err != nil {
		return state, err
	}
	if observation.BrokerOrderID != command.BrokerOrderID {
		return state, refuse(RefusalInvalidObservation, "cancel broker ID mismatch")
	}
	next := cloneState(state)
	position = next.positions[command.Position]
	switch observation.Status {
	case BrokerUnknown:
		position.Phase, position.EntryOpen, position.EntryLatch = CancelUnknown, false, RefusalInvalidObservation
	case BrokerActive:
		if !exactObserved(command.Position, position.Observed, observation) {
			return state, refuse(RefusalInvalidObservation, "active cancel result does not match protected order")
		}
		position.Phase, position.EntryOpen, position.EntryLatch, position.HasPending = Active, false, RefusalInvalidObservation, false
	case BrokerCanceled, BrokerFilled:
		position.Observed.Status, position.Phase, position.EntryOpen, position.EntryLatch, position.HasPending = observation.Status, Terminal, false, "", false
	default:
		return state, refuse(RefusalInvalidObservation, "cancel result status %q", observation.Status)
	}
	next.positions[command.Position] = position
	next.reseal()
	return next, nil
}

func recoverCancel(state State, key PositionKey, observation BrokerObservation, capability brokerCapability) (State, error) {
	position, err := mutablePosition(state, key, capability)
	if err != nil {
		return state, err
	}
	if !position.HasPending || position.Pending.Kind != OperationCancel || !capability.exactBrokerIDLookup || !capability.cancelResultQuery {
		return latchPosition(state, key, RefusalInvalidObservation), refuse(RefusalInvalidObservation, "exact cancel pending unavailable")
	}
	if observation.BrokerOrderID != position.Pending.BrokerOrderID {
		return latchPosition(state, key, RefusalInvalidObservation), refuse(RefusalInvalidObservation, "cancel lookup ID mismatch")
	}
	return applyCancelResult(state, position.Pending, observation)
}

func applyFill(state State, key PositionKey, fill Fill) (State, FillResult, error) {
	position, err := mutablePosition(state, key, brokerCapability{})
	if err != nil && errorCode(err) != RefusalInvalidObservation {
		return state, FillResult{PreserveExit: true}, err
	}
	if !validState(state) {
		return state, FillResult{PreserveExit: true}, refuse(RefusalInvalidState, "state seal invalid")
	}
	position = state.positions[key]
	if !validIdentity(fill.FillID) || !validIdentity(fill.Fingerprint) || fill.BrokerOrderID == "" || fill.BrokerOrderID != position.Observed.BrokerOrderID {
		return state, FillResult{PreserveExit: true}, refuse(RefusalInvalidObservation, "fill identity mismatch")
	}
	if previous, exists := position.Fills[fill.FillID]; exists {
		if previous.Quantity == fill.Quantity && previous.Fingerprint == fill.Fingerprint {
			return state, FillResult{Duplicate: true, PreserveExit: true}, nil
		}
		next := latchPosition(state, key, RefusalConflictingFill)
		return next, FillResult{PreserveExit: true}, refuse(RefusalConflictingFill, "fill ID reused with different content")
	}
	if fill.Quantity == 0 || fill.Quantity > position.Observed.Quantity || fill.Quantity > position.Holdings {
		return state, FillResult{PreserveExit: true}, refuse(RefusalFillExceeded, "fill quantity exceeds claim")
	}
	next := cloneState(state)
	position = next.positions[key]
	position.Fills[fill.FillID] = fillRecord{fill.Quantity, fill.Fingerprint}
	position.Holdings -= fill.Quantity
	position.Observed.Quantity -= fill.Quantity
	position.Desired.Quantity = position.Observed.Quantity
	position.LifecycleRevision++
	if position.Observed.Quantity == 0 {
		position.Observed.Status, position.Desired.Status, position.Phase, position.HasPending = BrokerFilled, BrokerFilled, Terminal, false
	}
	position.EntryOpen = position.Phase == Active && position.EntryLatch == "" && next.marketLatches[key.Market] == "" && position.Observed.Status == BrokerActive && position.Observed.Quantity+position.OtherSellClaims == position.Holdings
	next.positions[key] = position
	next.reseal()
	return next, FillResult{Applied: true, PreserveExit: true}, nil
}

func discoverOrphan(state State, observation BrokerObservation) (State, *BrokerCommand, error) {
	if !validState(state) || !validKey(PositionKey{observation.AccountID, observation.PositionID, observation.Market}) || !validIdentity(observation.BrokerOrderID) {
		return state, nil, refuse(RefusalInvalidObservation, "invalid orphan observation")
	}
	key := PositionKey{observation.AccountID, observation.PositionID, observation.Market}
	position, exists := state.positions[key]
	if exists && position.Observed.BrokerOrderID == observation.BrokerOrderID && exactObserved(key, position.Observed, observation) {
		return state, nil, nil
	}
	if first, recorded := state.orphans[observation.BrokerOrderID]; recorded {
		if first == observation {
			return state, nil, refuse(RefusalUnownedOrphan, "broker order remains unowned")
		}
		next := latchMarket(state, observation.Market, RefusalOrphanConflict)
		return next, nil, refuse(RefusalOrphanConflict, "broker order observation conflicts with first evidence")
	}
	next := cloneState(state)
	next.orphans[observation.BrokerOrderID] = observation
	next.marketLatches[observation.Market] = RefusalUnownedOrphan
	next.reseal()
	return next, nil, refuse(RefusalUnownedOrphan, "broker order has no exact durable owner")
}

func mutablePosition(state State, key PositionKey, capability brokerCapability) (positionState, error) {
	if !validState(state) {
		return positionState{}, refuse(RefusalInvalidState, "state seal invalid")
	}
	if capability.seal != ([32]byte{}) && !validCapability(capability) {
		return positionState{}, refuse(RefusalInvalidObservation, "capability seal invalid")
	}
	position, ok := state.positions[key]
	if !ok || !validKey(key) {
		return positionState{}, refuse(RefusalInvalidIdentity, "position missing")
	}
	return position, nil
}

func pendingPosition(state State, command BrokerCommand, kind OperationKind) (positionState, error) {
	position, err := mutablePosition(state, command.Position, brokerCapability{})
	if err != nil {
		return positionState{}, err
	}
	if !position.HasPending || position.Pending != command || command.Kind != kind || command.OperationKey != operationIdentity(command.Position, command.Generation, command.Revision, command.Kind) {
		return positionState{}, refuse(RefusalInvalidObservation, "command is not exact durable pending operation")
	}
	return position, nil
}

func validateAccepted(command BrokerCommand, observation BrokerObservation) error {
	if observation.Status != BrokerActive || !validIdentity(observation.BrokerOrderID) || observation.AccountID != command.Position.AccountID || observation.PositionID != command.Position.PositionID || observation.Market != command.Position.Market || observation.Generation != command.Generation || observation.Revision != command.Revision || observation.OperationKey != command.OperationKey || observation.Quantity != command.Quantity || observation.Trigger != command.Trigger {
		return refuse(RefusalInvalidObservation, "broker response does not exactly match command")
	}
	return nil
}

func exactNotFound(command BrokerCommand, observation BrokerObservation) bool {
	return observation.Status == BrokerNotFound && observation.AccountID == command.Position.AccountID && observation.PositionID == command.Position.PositionID && observation.Market == command.Position.Market && observation.Generation == command.Generation && observation.Revision == command.Revision && observation.OperationKey == command.OperationKey && observation.BrokerOrderID == ""
}

func exactObserved(key PositionKey, order ProtectionOrder, observation BrokerObservation) bool {
	return key.AccountID == observation.AccountID && key.PositionID == observation.PositionID && key.Market == observation.Market && order.Generation == observation.Generation && order.Revision == observation.Revision && order.OperationKey == observation.OperationKey && order.BrokerOrderID == observation.BrokerOrderID && order.Status == observation.Status && order.Quantity == observation.Quantity && order.Trigger == observation.Trigger
}

func commandOrder(command BrokerCommand, brokerID string, status BrokerStatus) ProtectionOrder {
	return ProtectionOrder{Generation: command.Generation, Revision: command.Revision, OperationKey: command.OperationKey, BrokerOrderID: brokerID, Status: status, Quantity: command.Quantity, Trigger: command.Trigger}
}

func observedOrder(observation BrokerObservation) ProtectionOrder {
	return ProtectionOrder{Generation: observation.Generation, Revision: observation.Revision, OperationKey: observation.OperationKey, BrokerOrderID: observation.BrokerOrderID, Status: observation.Status, Quantity: observation.Quantity, Trigger: observation.Trigger}
}

func latchPosition(state State, key PositionKey, code RefusalCode) State {
	next := cloneState(state)
	position := next.positions[key]
	position.EntryOpen, position.EntryLatch, position.Phase = false, code, ReconcileRequired
	next.positions[key] = position
	next.reseal()
	return next
}

func latchMarket(state State, market Market, code RefusalCode) State {
	next := cloneState(state)
	next.marketLatches[market] = code
	for key, position := range next.positions {
		if key.Market == market {
			position.EntryOpen = false
			if position.EntryLatch == "" {
				position.EntryLatch = code
			}
			next.positions[key] = position
		}
	}
	next.reseal()
	return next
}
