package protectionlifecycle

import "testing"

func FuzzOperationIdentitySeparation(f *testing.F) {
	f.Add("acct", "position", uint64(1), uint64(1))
	f.Add("a:b", "c", uint64(9), uint64(11))
	f.Fuzz(func(t *testing.T, account, position string, generation, revision uint64) {
		key := PositionKey{account, position, MarketKR}
		if !validIdentity(account) || !validIdentity(position) || generation == 0 || revision == 0 {
			t.Skip()
		}
		original := operationIdentity(key, generation, revision, OperationSubmit)
		if original == operationIdentity(PositionKey{account, position, MarketUS}, generation, revision, OperationSubmit) {
			t.Fatal("market not bound")
		}
		if original == operationIdentity(key, generation, revision, OperationCancel) {
			t.Fatal("operation kind not bound")
		}
		if original != operationIdentity(key, generation, revision, OperationSubmit) {
			t.Fatal("identity is not stable")
		}
	})
}

func FuzzDuplicateFillNeverDoubleDecrements(f *testing.F) {
	f.Add(uint64(1), "fill", "fingerprint")
	f.Fuzz(func(t *testing.T, quantity uint64, fillID, fingerprint string) {
		if quantity == 0 || quantity > 8 || !validIdentity(fillID) || !validIdentity(fingerprint) {
			t.Skip()
		}
		state, err := newState([]PositionSeed{{AccountID: "acct", PositionID: "pos", Market: MarketKR, Generation: 1, Holdings: 10, OtherSellClaims: 2}})
		if err != nil {
			t.Fatal(err)
		}
		pending, command, err := prepareRegister(state, PositionKey{"acct", "pos", MarketKR}, 8, 100, fullCapability())
		if err != nil {
			t.Fatal(err)
		}
		state, err = applySubmitResult(pending, command, acceptedObservation(command, "broker"))
		if err != nil {
			t.Fatal(err)
		}
		fill := Fill{FillID: fillID, BrokerOrderID: "broker", Quantity: quantity, Fingerprint: fingerprint}
		once, _, err := applyFill(state, command.Position, fill)
		if err != nil {
			t.Fatal(err)
		}
		twice, result, err := applyFill(once, command.Position, fill)
		if err != nil || !result.Duplicate || twice.view(command.Position).Holdings != once.view(command.Position).Holdings || twice.view(command.Position).LifecycleRevision != once.view(command.Position).LifecycleRevision {
			t.Fatalf("duplicate fill changed state result=%+v err=%v", result, err)
		}
	})
}
