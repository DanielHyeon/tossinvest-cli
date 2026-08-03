package protectionlifecycle

import "testing"

func baseState(t *testing.T) State {
	t.Helper()
	state, err := newState([]PositionSeed{
		{AccountID: "acct", PositionID: "kr-pos", Market: MarketKR, Generation: 7, Holdings: 10, OtherSellClaims: 2},
		{AccountID: "acct", PositionID: "us-pos", Market: MarketUS, Generation: 9, Holdings: 20, OtherSellClaims: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func fullCapability() brokerCapability {
	return newBrokerCapability(true, true, true, true, true)
}

func TestRegisterResponseCrashRecoversExactlyOnce(t *testing.T) {
	state := baseState(t)
	if state.view(PositionKey{"acct", "kr-pos", MarketKR}).EntryOpen {
		t.Fatal("unprotected seed opened entry")
	}
	next, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if command.OperationKey == "" || command.Generation != 7 || command.Revision != 1 {
		t.Fatalf("unstable command=%+v", command)
	}
	if next.view(command.Position).EntryOpen {
		t.Fatal("submit pending opened entry")
	}

	broker := newFakeOfficialBroker()
	broker.acceptWithoutResponse(command)
	recovered, retry, err := recoverSubmit(next, command.Position, broker.lookupOperation(command.OperationKey), fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if retry != nil || broker.submitCount != 1 {
		t.Fatalf("duplicate submit retry=%+v count=%d", retry, broker.submitCount)
	}
	view := recovered.view(command.Position)
	if view.Observed.Status != BrokerActive || view.Observed.BrokerOrderID == "" || !view.EntryOpen {
		t.Fatalf("recovered=%+v", view)
	}
	if _, retry, err = recoverSubmit(recovered, command.Position, broker.lookupOperation(command.OperationKey), fullCapability()); err != nil || retry != nil {
		t.Fatalf("repeated recovery err=%v retry=%+v", err, retry)
	}
}

func TestOperationIdentityBindsMarketGenerationRevisionAndKind(t *testing.T) {
	base := PositionKey{"acct", "position", MarketKR}
	original := operationIdentity(base, 3, 4, OperationSubmit)
	cases := []string{
		operationIdentity(PositionKey{"acct", "position", MarketUS}, 3, 4, OperationSubmit),
		operationIdentity(base, 5, 4, OperationSubmit),
		operationIdentity(base, 3, 5, OperationSubmit),
		operationIdentity(base, 3, 4, OperationReplace),
	}
	for _, candidate := range cases {
		if candidate == original {
			t.Fatalf("identity collision %q", candidate)
		}
	}
	if original != operationIdentity(base, 3, 4, OperationSubmit) {
		t.Fatal("operation identity is not deterministic")
	}
}

func TestUnknownSubmitWithoutIdempotencyNeverResubmits(t *testing.T) {
	state := baseState(t)
	capability := newBrokerCapability(true, true, true, true, false)
	next, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, capability)
	if err != nil {
		t.Fatal(err)
	}
	next, err = applySubmitResult(next, command, unknownObservation(command))
	if err != nil {
		t.Fatal(err)
	}
	next, retry, err := recoverSubmit(next, command.Position, notFoundOperation(command), capability)
	if err != nil {
		t.Fatal(err)
	}
	if retry != nil || next.view(command.Position).Phase != ReconcileRequired || next.view(command.Position).EntryOpen {
		t.Fatalf("unsafe resubmit state=%+v retry=%+v", next.view(command.Position), retry)
	}
}

func TestRegisterRequiresFullAvailableCoverage(t *testing.T) {
	state := baseState(t)
	if _, _, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 7, 100, fullCapability()); errorCode(err) != RefusalSellClaimExceeded {
		t.Fatalf("partial protection err=%v", err)
	}
}

func TestUnknownSubmitWithAttestedIdempotencyReusesSameOperationKey(t *testing.T) {
	state := baseState(t)
	next, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	next, err = applySubmitResult(next, command, unknownObservation(command))
	if err != nil {
		t.Fatal(err)
	}
	_, retry, err := recoverSubmit(next, command.Position, notFoundOperation(command), fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if retry == nil || retry.OperationKey != command.OperationKey || retry.Revision != command.Revision {
		t.Fatalf("retry changed identity original=%+v retry=%+v", command, retry)
	}
	broker := newFakeOfficialBroker()
	first := broker.submit(command)
	second := broker.submit(*retry)
	if first.BrokerOrderID != second.BrokerOrderID || len(broker.byOperation) != 1 {
		t.Fatalf("same-key retry created another order first=%+v second=%+v", first, second)
	}
}

func TestSaferReplaceIsAtomicNonRetreatingAndClaimBounded(t *testing.T) {
	state, command, broker := registeredKR(t)
	active := state.view(command.Position)

	if _, _, err := prepareReplace(state, command.Position, 8, 99, fullCapability()); errorCode(err) != RefusalTriggerRetreat {
		t.Fatalf("retreat err=%v", err)
	}
	if got := state.view(command.Position); got.Observed != active.Observed {
		t.Fatalf("retreat weakened active before=%+v after=%+v", active.Observed, got.Observed)
	}
	if _, _, err := prepareReplace(state, command.Position, 9, 101, fullCapability()); errorCode(err) != RefusalSellClaimExceeded {
		t.Fatalf("oversubscription err=%v", err)
	}
	capability := newBrokerCapability(true, true, true, false, true)
	if _, _, err := prepareReplace(state, command.Position, 8, 101, capability); errorCode(err) != RefusalContinuousCoverage {
		t.Fatalf("coverage err=%v", err)
	}

	pending, replace, err := prepareReplace(state, command.Position, 8, 101, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if replace.Kind != OperationReplace || replace.BrokerOrderID != active.Observed.BrokerOrderID {
		t.Fatalf("replace not atomic/exact=%+v", replace)
	}
	if pending.view(command.Position).EntryOpen {
		t.Fatal("replace pending opened entry")
	}
	broker.replaceWithoutResponse(replace)
	if got := pending.view(command.Position); got.Observed.Status != BrokerActive || got.Observed.Trigger != 100 {
		t.Fatalf("old coverage not preserved=%+v", got)
	}
	recovered, err := recoverReplace(pending, command.Position, broker.lookupBrokerID(replace.BrokerOrderID), fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if got := recovered.view(command.Position); got.Observed.Trigger != 101 || got.Observed.Quantity != 8 || !got.EntryOpen {
		t.Fatalf("replacement not converged=%+v", got)
	}
}

func TestCancelUnknownRequiresExactBrokerIDAndKeepsCoverage(t *testing.T) {
	state, command, broker := registeredKR(t)
	pending, cancel, err := prepareCancel(state, command.Position, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if pending.view(command.Position).EntryOpen {
		t.Fatal("cancel pending opened entry")
	}
	unknown := broker.cancelWithoutResponse(cancel, false)
	pending, err = applyCancelResult(pending, cancel, unknown)
	if err != nil {
		t.Fatal(err)
	}
	if got := pending.view(command.Position); got.Observed.Status != BrokerActive || got.EntryOpen {
		t.Fatalf("cancel unknown inferred terminal=%+v", got)
	}
	recovered, err := recoverCancel(pending, command.Position, broker.lookupBrokerID(cancel.BrokerOrderID), fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if got := recovered.view(command.Position); got.Observed.Status != BrokerActive || got.EntryOpen {
		t.Fatalf("active exact recovery weakened latch=%+v", got)
	}
}

func TestExactCancelResultConvergesTerminalAndKeepsEntryClosed(t *testing.T) {
	state, command, broker := registeredKR(t)
	pending, cancel, err := prepareCancel(state, command.Position, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	_ = broker.cancelWithoutResponse(cancel, true)
	next, err := recoverCancel(pending, command.Position, broker.lookupBrokerID(cancel.BrokerOrderID), fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	if got := next.view(command.Position); got.Phase != Terminal || got.Observed.Status != BrokerCanceled || got.EntryOpen {
		t.Fatalf("cancel did not converge safely=%+v", got)
	}
}

func TestNotFoundProofMustMatchCompleteOperationScope(t *testing.T) {
	state := baseState(t)
	pending, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	wrong := notFoundOperation(command)
	wrong.Market = MarketUS
	next, retry, err := recoverSubmit(pending, command.Position, wrong, fullCapability())
	if errorCode(err) != RefusalInvalidObservation || retry != nil || next.marketEntryOpen(MarketKR) {
		t.Fatalf("cross-scope NOT_FOUND accepted next=%+v retry=%+v err=%v", next.view(command.Position), retry, err)
	}
}

func TestDuplicateAndPartialFillConvergeOnce(t *testing.T) {
	state, command, _ := registeredKR(t)
	brokerID := state.view(command.Position).Observed.BrokerOrderID
	fill := Fill{FillID: "fill-1", BrokerOrderID: brokerID, Quantity: 3, Fingerprint: "trade-42"}
	next, result, err := applyFill(state, command.Position, fill)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || !result.PreserveExit || next.view(command.Position).Holdings != 7 || next.view(command.Position).Observed.Quantity != 5 {
		t.Fatalf("partial fill=%+v result=%+v", next.view(command.Position), result)
	}
	revision := next.view(command.Position).LifecycleRevision
	again, duplicate, err := applyFill(next, command.Position, fill)
	if err != nil || !duplicate.Duplicate || again.view(command.Position).LifecycleRevision != revision || again.view(command.Position).Holdings != 7 {
		t.Fatalf("duplicate changed state=%+v result=%+v err=%v", again.view(command.Position), duplicate, err)
	}
	conflict := fill
	conflict.Quantity = 2
	conflicted, _, err := applyFill(again, command.Position, conflict)
	if errorCode(err) != RefusalConflictingFill || conflicted.view(command.Position).EntryOpen {
		t.Fatalf("conflicting duplicate not latched state=%+v err=%v", conflicted.view(command.Position), err)
	}
}

func TestOrphanIsNeverGuessedCanceledOrAdopted(t *testing.T) {
	state := baseState(t)
	orphan := BrokerObservation{AccountID: "acct", PositionID: "kr-pos", Market: MarketKR, BrokerOrderID: "orphan-1", Status: BrokerActive, Quantity: 8, Trigger: 100}
	next, action, err := discoverOrphan(state, orphan)
	if errorCode(err) != RefusalUnownedOrphan || action != nil {
		t.Fatalf("orphan action=%+v err=%v", action, err)
	}
	if next.marketEntryOpen(MarketKR) || !next.marketEntryOpen(MarketUS) {
		t.Fatalf("market isolation kr=%v us=%v", next.marketEntryOpen(MarketKR), next.marketEntryOpen(MarketUS))
	}
}

func TestConflictingOrphanReobservationPreservesFirstEvidence(t *testing.T) {
	state := baseState(t)
	first := BrokerObservation{AccountID: "acct", PositionID: "kr-pos", Market: MarketKR, BrokerOrderID: "orphan-1", Status: BrokerActive, Quantity: 8, Trigger: 100}
	state, _, _ = discoverOrphan(state, first)
	second := first
	second.Quantity = 7
	next, action, err := discoverOrphan(state, second)
	if errorCode(err) != RefusalOrphanConflict || action != nil {
		t.Fatalf("conflict action=%+v err=%v", action, err)
	}
	if got := next.orphans[first.BrokerOrderID]; got != first {
		t.Fatalf("first evidence overwritten got=%+v want=%+v", got, first)
	}
}

func TestKRFailureDoesNotMutateUSProtection(t *testing.T) {
	state, _, _ := registeredKR(t)
	state, usCommand, err := prepareRegister(state, PositionKey{"acct", "us-pos", MarketUS}, 19, 200, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	state, err = applySubmitResult(state, usCommand, acceptedObservation(usCommand, "us-broker"))
	if err != nil {
		t.Fatal(err)
	}
	beforeUS := state.view(usCommand.Position)
	failed, _, err := recoverSubmit(state, PositionKey{"acct", "kr-pos", MarketKR}, BrokerObservation{Status: BrokerUnknown}, fullCapability())
	if err == nil {
		t.Fatal("expected exact recovery refusal")
	}
	afterUS := failed.view(usCommand.Position)
	if beforeUS != afterUS || !afterUS.EntryOpen || afterUS.Observed.Status != BrokerActive {
		t.Fatalf("KR failure changed US before=%+v after=%+v", beforeUS, afterUS)
	}
	krKey := PositionKey{"acct", "kr-pos", MarketKR}
	krBrokerID := failed.view(krKey).Observed.BrokerOrderID
	failed, krFill, err := applyFill(failed, krKey, Fill{FillID: "kr-after-failure", BrokerOrderID: krBrokerID, Quantity: 1, Fingerprint: "kr-trade"})
	if err != nil || !krFill.Applied || !krFill.PreserveExit {
		t.Fatalf("KR fill stopped by recovery latch result=%+v err=%v", krFill, err)
	}
	failed, usFill, err := applyFill(failed, usCommand.Position, Fill{FillID: "us-after-kr-failure", BrokerOrderID: beforeUS.Observed.BrokerOrderID, Quantity: 1, Fingerprint: "us-trade"})
	if err != nil || !usFill.Applied || !usFill.PreserveExit || failed.view(usCommand.Position).Observed.Quantity != beforeUS.Observed.Quantity-1 {
		t.Fatalf("US fill stopped by KR failure result=%+v state=%+v err=%v", usFill, failed.view(usCommand.Position), err)
	}
}

func TestExactBrokerIDRecoveryRejectsMismatchAndPreservesActive(t *testing.T) {
	state, command, _ := registeredKR(t)
	pending, _, err := prepareCancel(state, command.Position, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	next, err := recoverCancel(pending, command.Position, BrokerObservation{BrokerOrderID: "different", Status: BrokerCanceled}, fullCapability())
	if errorCode(err) != RefusalInvalidObservation {
		t.Fatalf("mismatch err=%v", err)
	}
	if got := next.view(command.Position); got.Observed.Status != BrokerActive || got.EntryOpen {
		t.Fatalf("mismatch weakened active protection=%+v", got)
	}
}

func TestSealedStateAndCapabilityTamperFailClosed(t *testing.T) {
	state := baseState(t)
	tampered := cloneState(state)
	position := tampered.positions[PositionKey{"acct", "kr-pos", MarketKR}]
	position.Holdings++
	tampered.positions[PositionKey{"acct", "kr-pos", MarketKR}] = position
	if _, _, err := prepareRegister(tampered, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability()); errorCode(err) != RefusalInvalidState {
		t.Fatalf("state tamper err=%v", err)
	}
	capability := fullCapability()
	capability.idempotentSubmit = false
	if _, _, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, capability); errorCode(err) != RefusalInvalidObservation {
		t.Fatalf("capability tamper err=%v", err)
	}
}

func TestStateTruthTableRejectsResealedEntryDuringPending(t *testing.T) {
	state := baseState(t)
	pending, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	position := pending.positions[command.Position]
	position.EntryOpen = true
	pending.positions[command.Position] = position
	pending.reseal()
	if validState(pending) {
		t.Fatal("resealed SUBMIT_PENDING state opened entry")
	}
	if _, err := applySubmitResult(pending, command, acceptedObservation(command, "broker")); errorCode(err) != RefusalInvalidState {
		t.Fatalf("truth-table corruption err=%v", err)
	}
}

func registeredKR(t *testing.T) (State, BrokerCommand, *fakeOfficialBroker) {
	t.Helper()
	state := baseState(t)
	pending, command, err := prepareRegister(state, PositionKey{"acct", "kr-pos", MarketKR}, 8, 100, fullCapability())
	if err != nil {
		t.Fatal(err)
	}
	broker := newFakeOfficialBroker()
	observation := broker.submit(command)
	active, err := applySubmitResult(pending, command, observation)
	if err != nil {
		t.Fatal(err)
	}
	return active, command, broker
}
