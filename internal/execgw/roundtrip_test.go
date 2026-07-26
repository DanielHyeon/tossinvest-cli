package execgw_test

// roundtrip_test.go covers the opaque-identifier rules (extend-execution-contract
// task 5.2): identifiers are stored exactly as received, compared byte-for-byte,
// and a created order is confirmed to exist before its attempt settles.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// newGatewayWithOrders wires a gateway that can read orders back.
func newGatewayWithOrders(t *testing.T, broker trading.Broker, reader execgw.OrderReader) (*execgw.Gateway, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
		Orders:     reader,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk
}

// orderDetailJSON is the detail-read payload, in the envelope shape the official
// client can hand back.
func orderDetailJSON(id, symbol string) string {
	return `{"result":{"orderId":"` + id + `","symbol":"` + symbol + `","side":"BUY",` +
		`"status":"PENDING","quantity":"2","price":"70000","currency":"KRW",` +
		`"orderedAt":"2026-03-30T10:30:00+09:00","canceledAt":null,` +
		`"execution":{"filledQuantity":"0"}}}`
}

// TestPlaceRoundTripConfirmsTheCreatedOrder is the happy path: the ack names an
// order, the read-back finds exactly that order on exactly that symbol, and only
// then does the attempt settle.
func TestPlaceRoundTripConfirmsTheCreatedOrder(t *testing.T) {
	const orderID = "0d5QIHjmtksbsmM-hBRAgP-ExI8iodGm9fAR5txelPfnMM8XQ_swoJdwL5RpGWMo"
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: orderID}}
	reader := newOrderReader()
	reader.set(orderID, orderDetailJSON(orderID, "005930"))
	gw, j, clk := newGatewayWithOrders(t, broker, reader)

	out, err := gw.Place(context.Background(), placeRequest(t, clk))
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.BrokerOrderID != orderID {
		t.Errorf("broker order id: got %q, want %q", rec.BrokerOrderID, orderID)
	}
}

// TestPlaceRoundTripFailuresLeaveInDoubt is the rule the spec states outright:
// a created identifier that cannot be confirmed does not settle as CONFIRMED.
func TestPlaceRoundTripFailuresLeaveInDoubt(t *testing.T) {
	const acked = "O-1"
	cases := []struct {
		name string
		// detail is what the reader answers with; empty means the order is
		// absent, so the read errors.
		detail     string
		wantDetail string
	}{
		{
			name:       "the order cannot be read back at all",
			wantDetail: "reading order",
		},
		{
			name:       "the read-back names a different order",
			detail:     orderDetailJSON("O-2", "005930"),
			wantDetail: "read-back names",
		},
		{
			name:       "the read-back differs only by whitespace",
			detail:     orderDetailJSON(" O-1", "005930"),
			wantDetail: "read-back names",
		},
		{
			name:       "the identifier turns up on another symbol",
			detail:     orderDetailJSON(acked, "000660"),
			wantDetail: "conflicting context",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: acked}}
			reader := newOrderReader()
			if tc.detail != "" {
				reader.set(acked, tc.detail)
			}
			gw, j, clk := newGatewayWithOrders(t, broker, reader)

			out, err := gw.Place(context.Background(), placeRequest(t, clk))
			if err == nil {
				t.Fatal("an unconfirmed creation must be reported as an error, not as success")
			}
			if out.State != journal.StateInDoubt {
				t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
			}
			if out.Reason != execgw.ReasonCode(journal.ReasonAckRoundTripUnconfirmed) {
				t.Errorf("reason: got %q, want %q", out.Reason, journal.ReasonAckRoundTripUnconfirmed)
			}
			if !strings.Contains(out.Detail, tc.wantDetail) {
				t.Errorf("detail %q does not mention %q", out.Detail, tc.wantDetail)
			}

			ctx := context.Background()
			rec, err := j.LookupAttempt(ctx, out.AttemptID)
			if err != nil {
				t.Fatalf("LookupAttempt: %v", err)
			}
			if rec.State != journal.StateInDoubt {
				t.Errorf("stored state: got %s, want IN_DOUBT", rec.State)
			}
			// The acked id is still durably recorded: an unconfirmed creation is
			// exactly the case where the resolution procedure needs it.
			if rec.BrokerOrderID != acked {
				t.Errorf("stored broker order id: got %q, want %q", rec.BrokerOrderID, acked)
			}
			// The broker was called once. A failed read-back is not a reason to
			// place anything again.
			if places, _, _ := broker.totals(); places != 1 {
				t.Errorf("broker place calls: got %d, want exactly 1", places)
			}
		})
	}
}

// TestCancelIsNotRoundTripped keeps the exit path free of the extra read (§0.3).
func TestCancelIsNotRoundTripped(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-1"}}
	// The reader knows nothing, so a round trip on this path would fail.
	gw, _, clk := newGatewayWithOrders(t, broker, newOrderReader())

	intent := cancelIntentFixture()
	out, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   intent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: goodDecision(t, execgw.CancelHash(intent), clk),
	})
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// TestBrokerOrderIDIsStoredVerbatim: the broker's identifier is opaque, so what
// goes into the journal is what came back — byte for byte, whitespace included.
// Normalising it would store an id the broker never issued.
func TestBrokerOrderIDIsStoredVerbatim(t *testing.T) {
	const padded = "  O-1\t"
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: padded}}
	gw, j, clk := newGateway(t, broker) // no reader: no round trip

	out, err := gw.Place(context.Background(), placeRequest(t, clk))
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if out.BrokerOrderID != padded {
		t.Errorf("outcome order id: got %q, want %q", out.BrokerOrderID, padded)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.BrokerOrderID != padded {
		t.Errorf("stored order id: got %q, want %q — identifiers are stored as received", rec.BrokerOrderID, padded)
	}
}

// TestBlankBrokerOrderIDIsNotAnAck: whitespace is not a name. An ack whose id is
// blank after trimming is an ack we cannot address, which is IN_DOUBT.
func TestBlankBrokerOrderIDIsNotAnAck(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "   "}}
	gw, j, clk := newGateway(t, broker)

	out, err := gw.Place(context.Background(), placeRequest(t, clk))
	if err == nil {
		t.Fatal("an unaddressable ack must not be reported as success")
	}
	if out.State != journal.StateInDoubt {
		t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
	}
	if out.Reason != execgw.ReasonCode(journal.ReasonAckWithoutOrderID) {
		t.Errorf("reason: got %q, want %q", out.Reason, journal.ReasonAckWithoutOrderID)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateInDoubt {
		t.Errorf("stored state: got %s, want IN_DOUBT", rec.State)
	}
}

// TestNoIdentifierShapeValidation: an identifier that looks nothing like the
// broker's usual token is still a valid identifier. openapi contracts no shape
// for `orderId`, so this build must not invent one — a rejection here would be a
// self-inflicted outage the day the broker changes its format.
func TestNoIdentifierShapeValidation(t *testing.T) {
	for _, id := range []string{
		"1",
		"주문-1",
		"a/b+c=",
		strings.Repeat("x", 512),
	} {
		broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: id}}
		reader := newOrderReader()
		reader.set(id, orderDetailJSON(id, "005930"))
		gw, _, clk := newGatewayWithOrders(t, broker, reader)

		out, err := gw.Place(context.Background(), placeRequest(t, clk))
		if err != nil {
			t.Errorf("Place with order id %q: %v", id, err)
			continue
		}
		if out.State != journal.StateConfirmed {
			t.Errorf("order id %q: state %s, want CONFIRMED (%s)", id, out.State, out.Detail)
		}
	}
}

// TestResolveConfirmedStoresTheIdentifierVerbatim pins the same rule on the
// resolution write path, which is the other place an identifier enters the
// journal.
func TestResolveConfirmedStoresTheIdentifierVerbatim(t *testing.T) {
	const padded = " O-42 "
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, nil)
	ctx := context.Background()

	attempt, err := f.journal.Resume(ctx, attemptID)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if err := attempt.ResolveConfirmed(ctx, padded, journal.ReasonResolvedFound, "found by hand"); err != nil {
		t.Fatalf("ResolveConfirmed: %v", err)
	}
	rec, err := f.journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.BrokerOrderID != padded {
		t.Errorf("stored order id: got %q, want %q", rec.BrokerOrderID, padded)
	}

	// Emptiness is still rejected, and it is still judged after trimming.
	f2 := newDoubtFixture(t)
	attempt2, err := f2.journal.Resume(ctx, f2.placeInDoubt(t, nil))
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if err := attempt2.ResolveConfirmed(ctx, "   ", journal.ReasonResolvedFound, "blank"); !errors.Is(err, journal.ErrInvalidRequest) {
		t.Errorf("a blank identifier must be refused, got %v", err)
	}
}
