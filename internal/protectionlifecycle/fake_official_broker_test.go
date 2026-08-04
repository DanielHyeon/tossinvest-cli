package protectionlifecycle

type fakeOfficialBroker struct {
	byOperation  map[string]BrokerObservation
	byBrokerID   map[string]BrokerObservation
	submitCount  int
	replaceCount int
	cancelCount  int
}

func newFakeOfficialBroker() *fakeOfficialBroker {
	return &fakeOfficialBroker{byOperation: map[string]BrokerObservation{}, byBrokerID: map[string]BrokerObservation{}}
}

func (broker *fakeOfficialBroker) submit(command BrokerCommand) BrokerObservation {
	broker.submitCount++
	observation := acceptedObservation(command, "broker-"+command.OperationKey[:12])
	broker.byOperation[command.OperationKey] = observation
	broker.byBrokerID[observation.BrokerOrderID] = observation
	return observation
}

func (broker *fakeOfficialBroker) acceptWithoutResponse(command BrokerCommand) {
	_ = broker.submit(command)
}

func (broker *fakeOfficialBroker) lookupOperation(operationKey string) BrokerObservation {
	if observation, ok := broker.byOperation[operationKey]; ok {
		return observation
	}
	return BrokerObservation{OperationKey: operationKey, Status: BrokerNotFound}
}

func (broker *fakeOfficialBroker) lookupBrokerID(brokerID string) BrokerObservation {
	if observation, ok := broker.byBrokerID[brokerID]; ok {
		return observation
	}
	return BrokerObservation{BrokerOrderID: brokerID, Status: BrokerNotFound}
}

func (broker *fakeOfficialBroker) replaceWithoutResponse(command BrokerCommand) {
	broker.replaceCount++
	observation := acceptedObservation(command, command.BrokerOrderID)
	broker.byOperation[command.OperationKey] = observation
	broker.byBrokerID[command.BrokerOrderID] = observation
}

func (broker *fakeOfficialBroker) cancelWithoutResponse(command BrokerCommand, canceled bool) BrokerObservation {
	broker.cancelCount++
	observation := broker.byBrokerID[command.BrokerOrderID]
	if canceled {
		observation.Status = BrokerCanceled
		broker.byBrokerID[command.BrokerOrderID] = observation
	}
	return BrokerObservation{BrokerOrderID: command.BrokerOrderID, Status: BrokerUnknown}
}

func acceptedObservation(command BrokerCommand, brokerID string) BrokerObservation {
	return BrokerObservation{AccountID: command.Position.AccountID, PositionID: command.Position.PositionID, Market: command.Position.Market, Generation: command.Generation, Revision: command.Revision, OperationKey: command.OperationKey, BrokerOrderID: brokerID, Status: BrokerActive, Quantity: command.Quantity, Trigger: command.Trigger}
}

func unknownObservation(command BrokerCommand) BrokerObservation {
	return BrokerObservation{OperationKey: command.OperationKey, Status: BrokerUnknown}
}

func notFoundOperation(command BrokerCommand) BrokerObservation {
	return BrokerObservation{AccountID: command.Position.AccountID, PositionID: command.Position.PositionID, Market: command.Position.Market, Generation: command.Generation, Revision: command.Revision, OperationKey: command.OperationKey, Status: BrokerNotFound}
}
