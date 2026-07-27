package brokerstate_test

// derive_status10_test.go is the decision table for the documented ten-value
// OrderStatus enum (extend-execution-contract task 5.1).
//
// The rows are the ones docs/migration/openapi.latest.json documents under
// components.schemas.OrderStatus. derive_test.go keeps the legacy OPEN/CLOSED
// group-label table, which the derivation still accepts as a compatibility shim;
// this file is the table real payloads actually exercise.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
)

func lineage(id string) brokerstate.Lineage {
	return brokerstate.Lineage{SuccessorOrderID: id}
}

// TestDeriveOrderStatusTable walks every documented status against the
// (canceledAt, filledQuantity, quantity, lineage) combinations that can
// accompany it. A combination the broker's own documentation cannot explain is
// UNKNOWN_BROKER_STATE — never a guess.
func TestDeriveOrderStatusTable(t *testing.T) {
	cases := []struct {
		name          string
		view          brokerstate.OrderView
		wantState     brokerstate.State
		wantReason    brokerstate.ReasonCode
		wantTerminal  bool
		wantRemaining float64
	}{
		// --- PENDING: 체결 대기 (openapi) --------------------------------------
		{
			name:          "pending with nothing filled",
			view:          brokerstate.OrderView{RawStatus: "PENDING", Quantity: 10},
			wantState:     brokerstate.StateOpenUnfilled,
			wantRemaining: 10,
		},
		{
			name:       "pending carrying a fill is contradictory",
			view:       brokerstate.OrderView{RawStatus: "PENDING", Quantity: 10, FilledQuantity: 3},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonPendingWithFill,
		},
		{
			name: "pending with a cancellation timestamp is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PENDING", Quantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithCancelTimestamp,
		},
		{
			name: "pending while a successor exists is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PENDING", Quantity: 10, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithSuccessor,
		},

		// --- PARTIAL_FILLED: 부분 체결 (openapi) -------------------------------
		{
			name:          "partially filled",
			view:          brokerstate.OrderView{RawStatus: "PARTIAL_FILLED", Quantity: 10, FilledQuantity: 3},
			wantState:     brokerstate.StateOpenPartiallyFilled,
			wantRemaining: 7,
		},
		{
			name:       "partially filled with nothing filled is contradictory",
			view:       brokerstate.OrderView{RawStatus: "PARTIAL_FILLED", Quantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonPartialFilledWithoutFill,
		},
		{
			name:       "partially filled with everything filled is contradictory",
			view:       brokerstate.OrderView{RawStatus: "PARTIAL_FILLED", Quantity: 10, FilledQuantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonPartialFilledWithFullFill,
		},
		{
			name: "partially filled with a cancellation timestamp is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PARTIAL_FILLED", Quantity: 10, FilledQuantity: 3, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithCancelTimestamp,
		},
		{
			name: "partially filled while a successor exists is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PARTIAL_FILLED", Quantity: 10, FilledQuantity: 3, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithSuccessor,
		},

		// --- PENDING_CANCEL: 취소 대기, still live (openapi) --------------------
		{
			name:          "cancel in flight, nothing filled",
			view:          brokerstate.OrderView{RawStatus: "PENDING_CANCEL", Quantity: 10},
			wantState:     brokerstate.StateCancelPending,
			wantRemaining: 10,
		},
		{
			name:          "cancel in flight after a partial fill",
			view:          brokerstate.OrderView{RawStatus: "PENDING_CANCEL", Quantity: 10, FilledQuantity: 4},
			wantState:     brokerstate.StateCancelPending,
			wantRemaining: 6,
		},
		{
			name:       "cancel in flight on a fully filled order is contradictory",
			view:       brokerstate.OrderView{RawStatus: "PENDING_CANCEL", Quantity: 10, FilledQuantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonPendingMutationWithFullFill,
		},
		{
			name: "cancel in flight with a cancellation timestamp is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PENDING_CANCEL", Quantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithCancelTimestamp,
		},

		// --- PENDING_REPLACE: 정정 대기, still live (openapi) -------------------
		{
			name:          "replace in flight after a partial fill",
			view:          brokerstate.OrderView{RawStatus: "PENDING_REPLACE", Quantity: 10, FilledQuantity: 2},
			wantState:     brokerstate.StateReplacePending,
			wantRemaining: 8,
		},
		{
			name:       "replace in flight on a fully filled order is contradictory",
			view:       brokerstate.OrderView{RawStatus: "PENDING_REPLACE", Quantity: 10, FilledQuantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonPendingMutationWithFullFill,
		},
		{
			name: "replace in flight while a successor exists is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "PENDING_REPLACE", Quantity: 10, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithSuccessor,
		},

		// --- FILLED: 전량 체결 (openapi) ---------------------------------------
		{
			name:         "fully filled",
			view:         brokerstate.OrderView{RawStatus: "FILLED", Quantity: 10, FilledQuantity: 10},
			wantState:    brokerstate.StateFilled,
			wantTerminal: true,
		},
		{
			name:       "filled with a remainder is contradictory",
			view:       brokerstate.OrderView{RawStatus: "FILLED", Quantity: 10, FilledQuantity: 4},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonFilledWithRemainder,
		},
		{
			name: "filled and cancelled is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "FILLED", Quantity: 10, FilledQuantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonCancelledWithFullFill,
		},
		{
			name: "filled with a successor is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "FILLED", Quantity: 10, FilledQuantity: 10, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonReplacedWithFullFill,
		},

		// --- CANCELED: 취소 완료 (openapi) -------------------------------------
		{
			name: "cancelled with nothing filled",
			view: brokerstate.OrderView{
				RawStatus: "CANCELED", Quantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:     brokerstate.StateCancelled,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name:          "cancelled without a timestamp is still cancelled",
			view:          brokerstate.OrderView{RawStatus: "CANCELED", Quantity: 10},
			wantState:     brokerstate.StateCancelled,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name: "cancelled after a partial fill",
			view: brokerstate.OrderView{
				RawStatus: "CANCELED", Quantity: 10, FilledQuantity: 4, CanceledAt: ptr(canceledAt),
			},
			wantState:     brokerstate.StateCancelledPartiallyFilled,
			wantTerminal:  true,
			wantRemaining: 6,
		},
		{
			name: "cancelled with a full fill is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "CANCELED", Quantity: 10, FilledQuantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonCancelledWithFullFill,
		},
		{
			name: "cancelled while lineage says it was replaced is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "CANCELED", Quantity: 10, CanceledAt: ptr(canceledAt),
				Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonCancelledWithSuccessor,
		},

		// --- REJECTED: 거부됨 (openapi — 부분 체결 가능) ------------------------
		{
			name:          "rejected with nothing filled",
			view:          brokerstate.OrderView{RawStatus: "REJECTED", Quantity: 10},
			wantState:     brokerstate.StateRejected,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name:          "rejected after a partial fill",
			view:          brokerstate.OrderView{RawStatus: "REJECTED", Quantity: 10, FilledQuantity: 3},
			wantState:     brokerstate.StateRejectedPartiallyFilled,
			wantTerminal:  true,
			wantRemaining: 7,
		},
		{
			name:       "rejected with a full fill is contradictory",
			view:       brokerstate.OrderView{RawStatus: "REJECTED", Quantity: 10, FilledQuantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonRejectedWithFullFill,
		},
		{
			name: "rejected while a successor exists is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "REJECTED", Quantity: 10, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonRejectedWithSuccessor,
		},

		// --- REPLACED: 정정됨 (openapi) ----------------------------------------
		{
			name: "replaced with a known successor",
			view: brokerstate.OrderView{
				RawStatus: "REPLACED", Quantity: 10, FilledQuantity: 3, Lineage: lineage("order-2"),
			},
			wantState:     brokerstate.StateReplaced,
			wantTerminal:  true,
			wantRemaining: 7,
		},
		{
			name: "replaced with a cancellation timestamp is still replaced",
			view: brokerstate.OrderView{
				RawStatus: "REPLACED", Quantity: 10, CanceledAt: ptr(canceledAt),
				Lineage: lineage("order-2"),
			},
			wantState:     brokerstate.StateReplaced,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name:       "replaced with no successor we can name blocks",
			view:       brokerstate.OrderView{RawStatus: "REPLACED", Quantity: 10, FilledQuantity: 3},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonReplacedWithoutSuccessor,
		},
		{
			name: "replaced with a full fill is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "REPLACED", Quantity: 10, FilledQuantity: 10, Lineage: lineage("order-2"),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonReplacedWithFullFill,
		},

		// --- the two separate-record statuses [형태 미측정 — 2b 2.1] ------------
		{
			name:       "cancel rejected record",
			view:       brokerstate.OrderView{RawStatus: "CANCEL_REJECTED", Quantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonCancelRejectedRecord,
		},
		{
			name:       "replace rejected record",
			view:       brokerstate.OrderView{RawStatus: "REPLACE_REJECTED", Quantity: 10, FilledQuantity: 2},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonReplaceRejectedRecord,
		},

		// --- normalisation and the unknown-code rule ---------------------------
		{
			name:          "case and surrounding space are normalised",
			view:          brokerstate.OrderView{RawStatus: " partial_filled ", Quantity: 10, FilledQuantity: 3},
			wantState:     brokerstate.StateOpenPartiallyFilled,
			wantRemaining: 7,
		},
		{
			name:       "an undocumented status blocks",
			view:       brokerstate.OrderView{RawStatus: "EXPIRED", Quantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonUnknownStatus,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := brokerstate.Derive(tc.view)
			if got.State != tc.wantState {
				t.Fatalf("State = %s (reason %s, detail %q), want %s",
					got.State, got.Reason, got.Detail, tc.wantState)
			}
			if got.Reason != tc.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, tc.wantReason)
			}
			if got.State == brokerstate.StateUnknown {
				if !got.FailClosed {
					t.Error("UNKNOWN_BROKER_STATE must set FailClosed")
				}
				if got.Terminal {
					t.Error("UNKNOWN_BROKER_STATE is a stop sign, not an end state")
				}
				if got.Detail == "" {
					t.Error("UNKNOWN_BROKER_STATE needs a detail for the operator alert")
				}
				return
			}
			if got.FailClosed {
				t.Errorf("State %s must not set FailClosed", got.State)
			}
			if got.Terminal != tc.wantTerminal {
				t.Errorf("Terminal = %v, want %v", got.Terminal, tc.wantTerminal)
			}
			if tc.wantRemaining != 0 && got.RemainingQuantity != tc.wantRemaining {
				t.Errorf("RemainingQuantity = %v, want %v", got.RemainingQuantity, tc.wantRemaining)
			}
			if got.State == brokerstate.StateReplaced && got.SuccessorOrderID != "order-2" {
				t.Errorf("SuccessorOrderID = %q, want order-2", got.SuccessorOrderID)
			}
		})
	}
}

// TestPendingMutationStatesAreNotTerminal pins the property the reservation
// release depends on (design D5): a cancel or a replace that is still in flight
// at the broker is a live order, so nothing may release its reservation or stop
// tracking it.
func TestPendingMutationStatesAreNotTerminal(t *testing.T) {
	for _, s := range []brokerstate.State{
		brokerstate.StateCancelPending, brokerstate.StateReplacePending,
	} {
		if s.IsTerminal() {
			t.Errorf("%s must not be terminal: the order is still live at the broker", s)
		}
	}
	for _, s := range []brokerstate.State{
		brokerstate.StateRejected, brokerstate.StateRejectedPartiallyFilled,
	} {
		if !s.IsTerminal() {
			t.Errorf("%s must be terminal: the broker refused the order", s)
		}
	}
}

// --- payloads in the shape the API documents --------------------------------
//
// Bodies taken from docs/migration/openapi.latest.json (GET /api/v1/orders
// examples "pendingMixed" and "completedWithNextPage"), so the derivation is
// pinned against the documented payload rather than one invented here. Note the
// per-order status values: the list rows carry the ten-value enum, not the
// OPEN/CLOSED group label the request parameter takes.

const fixtureListPending = `{"canceledAt":null,"currency":"KRW","execution":{"averageFilledPrice":null,"commission":null,"filledAmount":null,"filledAt":null,"filledQuantity":"0","settlementDate":null,"tax":null},"orderAmount":null,"orderId":"bAGzNvMOOTa5Uy0xVzYNbxDJ3Qpobwau4jDF3hyZZGWbpHm7wha8CFZc7aXVOWAl","orderType":"LIMIT","orderedAt":"2026-03-29T09:30:00+09:00","price":"70000","quantity":"10","side":"BUY","status":"PENDING","symbol":"005930","timeInForce":"DAY"}`

const fixtureListPartialFilled = `{"canceledAt":null,"currency":"USD","execution":{"averageFilledPrice":"185.25","commission":"0.66","filledAmount":"370.5","filledAt":"2026-03-29T10:00:05+09:00","filledQuantity":"2","settlementDate":null,"tax":"0"},"orderAmount":null,"orderId":"RpP3_wtsiKe9btBvdendaHoBqOIY_Zb_xPkRfYaqCIvf2FXtMDv_mo7VnD7KB-ia","orderType":"LIMIT","orderedAt":"2026-03-29T10:00:00+09:00","price":"185.5","quantity":"5","side":"SELL","status":"PARTIAL_FILLED","symbol":"AAPL","timeInForce":"DAY"}`

const fixtureListFilled = `{"result":{"canceledAt":null,"currency":"KRW","execution":{"averageFilledPrice":"70000","commission":"1400","filledAmount":"700000","filledAt":"2026-03-28T09:31:15+09:00","filledQuantity":"10","settlementDate":"2026-03-30","tax":"0"},"orderAmount":null,"orderId":"0d5QIHjmtksbsmM-hBRAgP-ExI8iodGm9fAR5txelPfnMM8XQ_swoJdwL5RpGWMo","orderType":"LIMIT","orderedAt":"2026-03-28T09:30:00+09:00","price":"70000","quantity":"10","side":"BUY","status":"FILLED","symbol":"005930","timeInForce":"DAY"}}`

func TestDeriveDocumentedPayloads(t *testing.T) {
	cases := []struct {
		name      string
		payload   string
		wantState brokerstate.State
		wantOrder string
	}{
		{
			name: "pending row from the open list", payload: fixtureListPending,
			wantState: brokerstate.StateOpenUnfilled,
			wantOrder: "bAGzNvMOOTa5Uy0xVzYNbxDJ3Qpobwau4jDF3hyZZGWbpHm7wha8CFZc7aXVOWAl",
		},
		{
			name: "partially filled row from the open list", payload: fixtureListPartialFilled,
			wantState: brokerstate.StateOpenPartiallyFilled,
			wantOrder: "RpP3_wtsiKe9btBvdendaHoBqOIY_Zb_xPkRfYaqCIvf2FXtMDv_mo7VnD7KB-ia",
		},
		{
			name: "filled row from the closed list", payload: fixtureListFilled,
			wantState: brokerstate.StateFilled,
			wantOrder: "0d5QIHjmtksbsmM-hBRAgP-ExI8iodGm9fAR5txelPfnMM8XQ_swoJdwL5RpGWMo",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := brokerstate.DeriveOfficialOrder([]byte(tc.payload), brokerstate.Lineage{})
			if got.State != tc.wantState {
				t.Fatalf("State = %s (reason %s, detail %q), want %s",
					got.State, got.Reason, got.Detail, tc.wantState)
			}
			if got.OrderID != tc.wantOrder {
				t.Errorf("OrderID = %q, want %q", got.OrderID, tc.wantOrder)
			}
		})
	}
}

// TestLegacyGroupLabelsStillDerive documents the compatibility shim: OPEN and
// CLOSED are the request parameter's group labels, not per-order statuses
// (openapi), but recorded fixtures and older journals carry them, so they keep
// deriving exactly as they did before the rewrite.
func TestLegacyGroupLabelsStillDerive(t *testing.T) {
	cases := []struct {
		view      brokerstate.OrderView
		wantState brokerstate.State
	}{
		{brokerstate.OrderView{RawStatus: "OPEN", Quantity: 10}, brokerstate.StateOpenUnfilled},
		{brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 10}, brokerstate.StateFilled},
		{
			brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, CanceledAt: ptr(canceledAt)},
			brokerstate.StateCancelled,
		},
		{
			brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, Lineage: lineage("order-2")},
			brokerstate.StateReplaced,
		},
	}
	for _, tc := range cases {
		if got := brokerstate.Derive(tc.view); got.State != tc.wantState {
			t.Errorf("Derive(%+v) = %s, want %s", tc.view, got.State, tc.wantState)
		}
	}
}
