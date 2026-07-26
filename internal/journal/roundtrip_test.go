package journal

// roundtrip_test.go pins where the created-order confirmation sits in the
// dispatch sequence (extend-execution-contract task 5.2). The placement is the
// property: after MarkAcked, so the identifier is already durable when the check
// runs, and before Settle, so a failed check can still produce IN_DOUBT instead
// of CONFIRMED.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func ackDispatch(orderID string) DispatchFunc {
	return func(context.Context, *Attempt) DispatchOutcome {
		return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: orderID}
	}
}

// TestDispatchVerifiedRunsAfterMarkAckedAndBeforeSettle observes the journal from
// inside the check itself, which is the only way to prove the ordering rather
// than assume it.
func TestDispatchVerifiedRunsAfterMarkAckedAndBeforeSettle(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	var stateDuringCheck AttemptState
	var idDuringCheck string
	verify := func(ctx context.Context, brokerOrderID string) error {
		rec, err := j.LookupAttempt(ctx, "attempt-1")
		if err != nil {
			t.Errorf("LookupAttempt inside the check: %v", err)
			return nil
		}
		stateDuringCheck = rec.State
		idDuringCheck = rec.BrokerOrderID
		return nil
	}

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	res, err := attempt.DispatchVerified(ctx, ackDispatch("order-42"), verify)
	if err != nil {
		t.Fatalf("DispatchVerified: %v", err)
	}
	if res.Final != StateConfirmed {
		t.Fatalf("Final = %s, want CONFIRMED", res.Final)
	}
	if stateDuringCheck != StateAcked {
		t.Errorf("state during the check = %s, want ACKED (the id must be durable before it is checked)",
			stateDuringCheck)
	}
	if idDuringCheck != "order-42" {
		t.Errorf("broker order id during the check = %q, want order-42", idDuringCheck)
	}
}

// TestDispatchVerifiedFailureLeavesInDoubt: an unconfirmed creation is IN_DOUBT,
// keeps the identifier, and stays blocking.
func TestDispatchVerifiedFailureLeavesInDoubt(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	checkErr := errors.New("the order is not there")

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	res, err := attempt.DispatchVerified(ctx, ackDispatch("order-42"),
		func(context.Context, string) error { return checkErr })
	if err != nil {
		t.Fatalf("DispatchVerified: %v", err)
	}
	if res.Final != StateInDoubt {
		t.Fatalf("Final = %s, want IN_DOUBT", res.Final)
	}
	if res.ReasonCode != ReasonAckRoundTripUnconfirmed {
		t.Errorf("ReasonCode = %q, want %q", res.ReasonCode, ReasonAckRoundTripUnconfirmed)
	}
	if !errors.Is(res.Err, checkErr) {
		t.Errorf("Err = %v, want the check's error", res.Err)
	}
	if !res.Blocking() {
		t.Error("an unconfirmed creation must keep the symbol blocked")
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateInDoubt {
		t.Errorf("stored state = %s, want IN_DOUBT", rec.State)
	}
	if rec.BrokerOrderID != "order-42" {
		t.Errorf("stored broker order id = %q, want order-42 — the resolution needs it", rec.BrokerOrderID)
	}
}

// TestDispatchIsDispatchVerifiedWithoutACheck keeps the released Dispatch
// contract: no reader, no round trip, same outcome as before.
func TestDispatchIsDispatchVerifiedWithoutACheck(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	res, err := attempt.Dispatch(ctx, ackDispatch("order-42"))
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if res.Final != StateConfirmed || res.BrokerOrderID != "order-42" {
		t.Errorf("res = %+v, want a CONFIRMED order-42", res)
	}
}

// TestAckIdentifiersAreStoredVerbatim: `orderId` is opaque (openapi contracts no
// shape), so emptiness is judged after trimming and the value is stored as
// received.
func TestAckIdentifiersAreStoredVerbatim(t *testing.T) {
	cases := []struct {
		name      string
		orderID   string
		wantState AttemptState
		wantStore string
	}{
		{name: "padded id is stored as sent", orderID: " order-42\t",
			wantState: StateConfirmed, wantStore: " order-42\t"},
		{name: "blank id is not a name", orderID: "  \t ",
			wantState: StateInDoubt},
		{name: "empty id is not a name", orderID: "",
			wantState: StateInDoubt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			attempt, err := j.Prepare(ctx, testRequest())
			if err != nil {
				t.Fatalf("Prepare: %v", err)
			}
			res, err := attempt.Dispatch(ctx, ackDispatch(tc.orderID))
			if err != nil {
				t.Fatalf("Dispatch: %v", err)
			}
			if res.Final != tc.wantState {
				t.Fatalf("Final = %s, want %s", res.Final, tc.wantState)
			}
			rec, err := j.LookupAttempt(ctx, "attempt-1")
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantState == StateInDoubt {
				if res.ReasonCode != ReasonAckWithoutOrderID {
					t.Errorf("ReasonCode = %q, want %q", res.ReasonCode, ReasonAckWithoutOrderID)
				}
				if strings.TrimSpace(rec.BrokerOrderID) != "" {
					t.Errorf("a blank ack must not record an addressable id, got %q", rec.BrokerOrderID)
				}
				return
			}
			if rec.BrokerOrderID != tc.wantStore {
				t.Errorf("stored broker order id = %q, want %q", rec.BrokerOrderID, tc.wantStore)
			}
		})
	}
}

// TestOperatorResolveStoresTheIdentifierVerbatim covers the other write path an
// identifier reaches the journal by.
func TestOperatorResolveStoresTheIdentifierVerbatim(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkInDoubt(ctx, "test", "parked for the test"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	if err := attempt.ResolveUnresolved(ctx, ReasonResolutionExhausted, "nothing could be proven"); err != nil {
		t.Fatalf("ResolveUnresolved: %v", err)
	}

	const padded = " order-42 "
	if err := j.OperatorResolve(ctx, "attempt-1", StateConfirmed, "daniel", padded,
		"found it in the app"); err != nil {
		t.Fatalf("OperatorResolve: %v", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.BrokerOrderID != padded {
		t.Errorf("stored broker order id = %q, want %q", rec.BrokerOrderID, padded)
	}
}
