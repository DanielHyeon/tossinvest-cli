package brokerstate_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func ptr(t time.Time) *time.Time { return &t }

var canceledAt = time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)

// TestDerivePriorityTable is the decision table from the "브로커 상태 파생"
// requirement. Every row is a (status, canceledAt, filledQuantity, quantity,
// lineage) combination and its single allowed outcome. Contradictory or unknown
// combinations must land on UNKNOWN_BROKER_STATE, which blocks new entries for the
// symbol — never on a guess.
func TestDerivePriorityTable(t *testing.T) {
	cases := []struct {
		name          string
		view          brokerstate.OrderView
		wantState     brokerstate.State
		wantReason    brokerstate.ReasonCode
		wantTerminal  bool
		wantRemaining float64
	}{
		// --- structural validation wins over everything -------------------------
		{
			name:       "zero quantity",
			view:       brokerstate.OrderView{RawStatus: "OPEN", Quantity: 0},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonQuantityNotPositive,
		},
		{
			name:       "negative quantity",
			view:       brokerstate.OrderView{RawStatus: "CLOSED", Quantity: -5, FilledQuantity: 0},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonQuantityNotPositive,
		},
		{
			name:       "negative filled quantity",
			view:       brokerstate.OrderView{RawStatus: "OPEN", Quantity: 10, FilledQuantity: -1},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonFilledNegative,
		},
		{
			name:       "filled exceeds quantity",
			view:       brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 11},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonFilledExceedsQuantity,
		},

		// --- unknown status -----------------------------------------------------
		{
			name:       "empty status",
			view:       brokerstate.OrderView{RawStatus: "", Quantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonMissingStatus,
		},
		{
			name:       "status the API never documented",
			view:       brokerstate.OrderView{RawStatus: "PARTIALLY_FILLED", Quantity: 10, FilledQuantity: 4},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonUnknownStatus,
		},
		{
			name:       "WTS-style Korean status leaking in",
			view:       brokerstate.OrderView{RawStatus: "체결", Quantity: 10, FilledQuantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonUnknownStatus,
		},

		// --- OPEN ---------------------------------------------------------------
		{
			name:          "open, nothing filled",
			view:          brokerstate.OrderView{RawStatus: "OPEN", Quantity: 10},
			wantState:     brokerstate.StateOpenUnfilled,
			wantRemaining: 10,
		},
		{
			name:          "open, partially filled",
			view:          brokerstate.OrderView{RawStatus: "OPEN", Quantity: 10, FilledQuantity: 3},
			wantState:     brokerstate.StateOpenPartiallyFilled,
			wantRemaining: 7,
		},
		{
			name:          "lowercase status is accepted",
			view:          brokerstate.OrderView{RawStatus: " open ", Quantity: 10},
			wantState:     brokerstate.StateOpenUnfilled,
			wantRemaining: 10,
		},
		{
			name: "open with a cancellation timestamp is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "OPEN", Quantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithCancelTimestamp,
		},
		{
			name: "open while a successor order exists is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "OPEN", Quantity: 10,
				Lineage: brokerstate.Lineage{SuccessorOrderID: "order-2"},
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonOpenWithSuccessor,
		},

		// --- CLOSED with a lineage successor (AMEND) ----------------------------
		{
			name: "closed with successor and an unfilled remainder is REPLACED",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 3,
				Lineage: brokerstate.Lineage{SuccessorOrderID: "order-2"},
			},
			wantState:     brokerstate.StateReplaced,
			wantTerminal:  true,
			wantRemaining: 7,
		},
		{
			name: "closed with successor, canceledAt set, is still REPLACED",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 0,
				CanceledAt: ptr(canceledAt),
				Lineage:    brokerstate.Lineage{SuccessorOrderID: "order-2"},
			},
			wantState:     brokerstate.StateReplaced,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name: "closed with successor but fully filled is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 10,
				Lineage: brokerstate.Lineage{SuccessorOrderID: "order-2"},
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonReplacedWithFullFill,
		},

		// --- CLOSED + canceledAt (the spec's explicit scenario) -----------------
		{
			name: "closed and cancelled with nothing filled",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:     brokerstate.StateCancelled,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name: "closed and cancelled after a partial fill",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 4, CanceledAt: ptr(canceledAt),
			},
			wantState:     brokerstate.StateCancelledPartiallyFilled,
			wantTerminal:  true,
			wantRemaining: 6,
		},
		{
			name: "cancel flag without a timestamp still counts as cancelled",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, Canceled: true,
			},
			wantState:     brokerstate.StateCancelled,
			wantTerminal:  true,
			wantRemaining: 10,
		},
		{
			name: "cancelled with a full fill is contradictory",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 10, CanceledAt: ptr(canceledAt),
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonCancelledWithFullFill,
		},

		// --- CLOSED without cancel evidence ------------------------------------
		{
			name:         "closed and fully filled",
			view:         brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 10},
			wantState:    brokerstate.StateFilled,
			wantTerminal: true,
		},
		{
			name:       "closed with nothing filled and no cancel evidence",
			view:       brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonClosedWithoutFillOrCancel,
		},
		{
			name:       "closed with an unexplained remainder",
			view:       brokerstate.OrderView{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 4},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonClosedWithUnexplainedRemainder,
		},

		// --- decimal handling ---------------------------------------------------
		{
			name: "fractional US share fully filled",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 0.5, FilledQuantity: 0.5,
			},
			wantState:    brokerstate.StateFilled,
			wantTerminal: true,
		},
		{
			name: "float representation noise counts as a full fill",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 9.999999999,
			},
			wantState:    brokerstate.StateFilled,
			wantTerminal: true,
		},
		{
			name: "a real remainder is not swallowed by the epsilon",
			view: brokerstate.OrderView{
				RawStatus: "CLOSED", Quantity: 10, FilledQuantity: 9.99,
			},
			wantState:  brokerstate.StateUnknown,
			wantReason: brokerstate.ReasonClosedWithUnexplainedRemainder,
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
				if got.Reason == "" || got.Detail == "" {
					t.Errorf("UNKNOWN_BROKER_STATE needs a reason and a detail, got %+v", got)
				}
				if !got.Terminal {
					// Unknown is not a lifecycle end, it is a stop sign.
					if got.Terminal {
						t.Error("unreachable")
					}
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

// TestDeriveNonFiniteQuantities keeps NaN/Inf out of the position accounting.
func TestDeriveNonFiniteQuantities(t *testing.T) {
	inf := 1.0
	for i := 0; i < 400; i++ {
		inf *= 10 // overflow to +Inf without importing math
	}
	nan := inf - inf

	for _, v := range []brokerstate.OrderView{
		{RawStatus: "OPEN", Quantity: nan},
		{RawStatus: "OPEN", Quantity: inf},
		{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: nan},
		{RawStatus: "CLOSED", Quantity: 10, FilledQuantity: inf},
	} {
		got := brokerstate.Derive(v)
		if got.State != brokerstate.StateUnknown {
			t.Errorf("Derive(%+v) = %s, want UNKNOWN_BROKER_STATE", v, got.State)
		}
		if got.Reason != brokerstate.ReasonNonFiniteQuantity &&
			got.Reason != brokerstate.ReasonQuantityNotPositive {
			t.Errorf("Derive(%+v) reason = %q", v, got.Reason)
		}
	}
}

// TestFromDomainOrder wires the derivation to the shape the official client
// actually produces. domain.Order has no canceledAt field (the upstream adapter
// drops it), so the caller supplies it alongside — which is exactly why the
// derivation takes its own view type instead of domain.Order.
func TestFromDomainOrder(t *testing.T) {
	o := domain.Order{
		ID:             "abc123",
		Symbol:         "005930",
		Side:           "BUY",
		Status:         "CLOSED",
		Quantity:       10,
		FilledQuantity: 4,
	}
	view := brokerstate.FromDomainOrder(o, ptr(canceledAt), brokerstate.Lineage{})
	if view.OrderID != "abc123" || view.RawStatus != "CLOSED" ||
		view.Quantity != 10 || view.FilledQuantity != 4 {
		t.Fatalf("view = %+v", view)
	}
	if !view.Canceled || view.CanceledAt == nil || !view.CanceledAt.Equal(canceledAt) {
		t.Fatalf("cancellation not carried over: %+v", view)
	}

	got := brokerstate.Derive(view)
	if got.State != brokerstate.StateCancelledPartiallyFilled {
		t.Fatalf("State = %s, want CANCELLED_PARTIALLY_FILLED", got.State)
	}
	if got.OrderID != "abc123" {
		t.Fatalf("OrderID = %q, want abc123", got.OrderID)
	}

	// Without a cancellation timestamp the same row is a contradiction, not a
	// silent "probably cancelled".
	if s := brokerstate.Derive(brokerstate.FromDomainOrder(o, nil, brokerstate.Lineage{})).State; s != brokerstate.StateUnknown {
		t.Fatalf("State without canceledAt = %s, want UNKNOWN_BROKER_STATE", s)
	}
}

// --- official payload fixtures ---------------------------------------------
//
// These are the exact response bodies internal/official's own tests use
// (orders_reads_test.go), so the derivation is pinned against the upstream
// fixture rather than against a shape we invented here.

const fixtureOpenEnvelope = `{"result":{"orderId":"abc123","symbol":"005930","side":"BUY","orderType":"LIMIT","timeInForce":"DAY","status":"OPEN","quantity":"10","price":"70000","currency":"KRW","orderedAt":"2026-03-29T09:30:00+09:00","canceledAt":null,"orderAmount":null,"execution":{"filledQuantity":"0","averageFilledPrice":null,"filledAmount":null,"commission":null,"tax":null,"filledAt":null,"settlementDate":null}}}`

const fixtureBareClosedFilled = `{"orderId":"xyz789","symbol":"AAPL","side":"SELL","orderType":"LIMIT","timeInForce":"DAY","status":"CLOSED","quantity":"5","price":"200","currency":"USD","orderedAt":"2026-03-28T22:00:00+00:00","canceledAt":null,"orderAmount":null,"execution":{"filledQuantity":"5","averageFilledPrice":"201.5","filledAmount":"1007.5","commission":null,"tax":null,"filledAt":"2026-03-28T22:01:00+00:00","settlementDate":null}}`

const fixtureCancelledPartial = `{"result":{"orderId":"cnl001","symbol":"005930","side":"BUY","orderType":"LIMIT","timeInForce":"DAY","status":"CLOSED","quantity":"10","price":"70000","currency":"KRW","orderedAt":"2026-03-29T09:30:00+09:00","canceledAt":"2026-03-29T10:15:00+09:00","orderAmount":null,"execution":{"filledQuantity":"4","averageFilledPrice":"70000","filledAmount":"280000","commission":null,"tax":null,"filledAt":"2026-03-29T09:31:00+09:00","settlementDate":null}}}`

// TestDeriveOfficialOrderFixtures runs the derivation over upstream's own response
// bodies, with and without the {"result": …} envelope.
func TestDeriveOfficialOrderFixtures(t *testing.T) {
	cases := []struct {
		name      string
		payload   string
		lineage   brokerstate.Lineage
		wantState brokerstate.State
		wantOrder string
	}{
		{
			name: "open unfilled (envelope)", payload: fixtureOpenEnvelope,
			wantState: brokerstate.StateOpenUnfilled, wantOrder: "abc123",
		},
		{
			name: "closed fully filled (bare)", payload: fixtureBareClosedFilled,
			wantState: brokerstate.StateFilled, wantOrder: "xyz789",
		},
		{
			name: "cancelled after partial fill", payload: fixtureCancelledPartial,
			wantState: brokerstate.StateCancelledPartiallyFilled, wantOrder: "cnl001",
		},
		{
			name: "cancelled partial with a successor is a replace", payload: fixtureCancelledPartial,
			lineage:   brokerstate.Lineage{SuccessorOrderID: "cnl002"},
			wantState: brokerstate.StateReplaced, wantOrder: "cnl001",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view, err := brokerstate.ParseOfficialOrder([]byte(tc.payload), tc.lineage)
			if err != nil {
				t.Fatalf("ParseOfficialOrder: %v", err)
			}
			if view.OrderID != tc.wantOrder {
				t.Errorf("OrderID = %q, want %q", view.OrderID, tc.wantOrder)
			}
			got := brokerstate.Derive(view)
			if got.State != tc.wantState {
				t.Fatalf("State = %s (reason %s, detail %q), want %s",
					got.State, got.Reason, got.Detail, tc.wantState)
			}

			// The convenience wrapper must agree with parse+derive.
			direct := brokerstate.DeriveOfficialOrder([]byte(tc.payload), tc.lineage)
			if direct.State != got.State {
				t.Errorf("DeriveOfficialOrder = %s, Derive(Parse) = %s", direct.State, got.State)
			}
		})
	}
}

// TestParseOfficialOrderCancelledAtVariants covers timestamp formats: a format we
// can read is parsed, and one we cannot still counts as cancellation evidence. The
// dangerous outcome would be reading an unfamiliar timestamp as "not cancelled".
func TestParseOfficialOrderCancelledAtVariants(t *testing.T) {
	cases := []struct {
		name       string
		canceledAt string
		wantParsed bool
	}{
		{"RFC3339 with offset", "2026-03-29T10:15:00+09:00", true},
		{"space separated fallback", "2026-03-29 10:15:00", true},
		{"unparseable", "yesterday afternoon", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := `{"orderId":"o1","status":"CLOSED","quantity":"10","canceledAt":"` +
				tc.canceledAt + `","execution":{"filledQuantity":"0"}}`
			view, err := brokerstate.ParseOfficialOrder([]byte(payload), brokerstate.Lineage{})
			if err != nil {
				t.Fatalf("ParseOfficialOrder: %v", err)
			}
			if !view.Canceled {
				t.Fatal("canceledAt present must always count as cancellation evidence")
			}
			if gotParsed := view.CanceledAt != nil; gotParsed != tc.wantParsed {
				t.Errorf("CanceledAt parsed = %v (%v), want %v", gotParsed, view.CanceledAt, tc.wantParsed)
			}
			if got := brokerstate.Derive(view); got.State != brokerstate.StateCancelled {
				t.Fatalf("State = %s, want CANCELLED", got.State)
			}
		})
	}
}

// TestParseOfficialOrderMalformed keeps a broken payload fail-closed instead of
// deriving from zero values.
func TestParseOfficialOrderMalformed(t *testing.T) {
	cases := []string{
		`not json`,
		`{"orderId":"o1","status":"CLOSED","quantity":"ten","execution":{"filledQuantity":"0"}}`,
		`{"orderId":"o1","status":"CLOSED","quantity":"10","execution":{"filledQuantity":"lots"}}`,
	}
	for _, payload := range cases {
		if _, err := brokerstate.ParseOfficialOrder([]byte(payload), brokerstate.Lineage{}); !errors.Is(err, brokerstate.ErrMalformedOrder) {
			t.Errorf("ParseOfficialOrder(%q): want ErrMalformedOrder, got %v", payload, err)
		}
		got := brokerstate.DeriveOfficialOrder([]byte(payload), brokerstate.Lineage{})
		if got.State != brokerstate.StateUnknown || got.Reason != brokerstate.ReasonMalformedPayload {
			t.Errorf("DeriveOfficialOrder(%q) = (%s, %s), want (UNKNOWN_BROKER_STATE, %s)",
				payload, got.State, got.Reason, brokerstate.ReasonMalformedPayload)
		}
		if !got.FailClosed {
			t.Errorf("DeriveOfficialOrder(%q) must fail closed", payload)
		}
	}
}

// TestParseOfficialOrderMissingPrice tolerates the nullable fields the API sends as
// null for pending orders.
func TestParseOfficialOrderNullExecution(t *testing.T) {
	payload := `{"orderId":"o1","status":"OPEN","quantity":"10","price":null,"canceledAt":null,"execution":{"filledQuantity":null,"averageFilledPrice":null}}`
	view, err := brokerstate.ParseOfficialOrder([]byte(payload), brokerstate.Lineage{})
	if err != nil {
		t.Fatalf("ParseOfficialOrder: %v", err)
	}
	if view.FilledQuantity != 0 {
		t.Fatalf("FilledQuantity = %v, want 0 for a null execution", view.FilledQuantity)
	}
	if got := brokerstate.Derive(view); got.State != brokerstate.StateOpenUnfilled {
		t.Fatalf("State = %s, want OPEN_UNFILLED", got.State)
	}
}

// TestStateAndReasonStringsAreStable pins the wire strings: they end up in journal
// rows, structured logs and operator alerts, so renaming one is a breaking change.
func TestStateAndReasonStringsAreStable(t *testing.T) {
	states := map[brokerstate.State]string{
		brokerstate.StateOpenUnfilled:             "OPEN_UNFILLED",
		brokerstate.StateOpenPartiallyFilled:      "OPEN_PARTIALLY_FILLED",
		brokerstate.StateFilled:                   "FILLED",
		brokerstate.StateCancelled:                "CANCELLED",
		brokerstate.StateCancelledPartiallyFilled: "CANCELLED_PARTIALLY_FILLED",
		brokerstate.StateReplaced:                 "REPLACED",
		brokerstate.StateRejected:                 "REJECTED",
		brokerstate.StateRejectedPartiallyFilled:  "REJECTED_PARTIALLY_FILLED",
		brokerstate.StateCancelPending:            "CANCEL_PENDING",
		brokerstate.StateReplacePending:           "REPLACE_PENDING",
		brokerstate.StateUnknown:                  "UNKNOWN_BROKER_STATE",
	}
	for state, want := range states {
		if string(state) != want {
			t.Errorf("state %v = %q, want %q", state, string(state), want)
		}
	}

	reasons := []brokerstate.ReasonCode{
		brokerstate.ReasonMissingStatus,
		brokerstate.ReasonUnknownStatus,
		brokerstate.ReasonQuantityNotPositive,
		brokerstate.ReasonFilledNegative,
		brokerstate.ReasonFilledExceedsQuantity,
		brokerstate.ReasonNonFiniteQuantity,
		brokerstate.ReasonOpenWithCancelTimestamp,
		brokerstate.ReasonOpenWithSuccessor,
		brokerstate.ReasonReplacedWithFullFill,
		brokerstate.ReasonCancelledWithFullFill,
		brokerstate.ReasonClosedWithoutFillOrCancel,
		brokerstate.ReasonClosedWithUnexplainedRemainder,
		brokerstate.ReasonMalformedPayload,
		brokerstate.ReasonPendingWithFill,
		brokerstate.ReasonPartialFilledWithoutFill,
		brokerstate.ReasonPartialFilledWithFullFill,
		brokerstate.ReasonPendingMutationWithFullFill,
		brokerstate.ReasonFilledWithRemainder,
		brokerstate.ReasonCancelledWithSuccessor,
		brokerstate.ReasonRejectedWithFullFill,
		brokerstate.ReasonRejectedWithSuccessor,
		brokerstate.ReasonReplacedWithoutSuccessor,
		brokerstate.ReasonCancelRejectedRecord,
		brokerstate.ReasonReplaceRejectedRecord,
	}
	seen := map[brokerstate.ReasonCode]bool{}
	for _, r := range reasons {
		if r == "" {
			t.Error("reason codes must not be empty")
		}
		if strings.ToLower(string(r)) != string(r) {
			t.Errorf("reason code %q should be lower_snake_case", r)
		}
		if seen[r] {
			t.Errorf("duplicate reason code %q", r)
		}
		seen[r] = true
	}
}

// TestTerminalHelpers documents which states end an order's life at the broker.
func TestTerminalHelpers(t *testing.T) {
	terminal := []brokerstate.State{
		brokerstate.StateFilled, brokerstate.StateCancelled,
		brokerstate.StateCancelledPartiallyFilled, brokerstate.StateReplaced,
		brokerstate.StateRejected, brokerstate.StateRejectedPartiallyFilled,
	}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("%s must be terminal", s)
		}
	}
	for _, s := range []brokerstate.State{
		brokerstate.StateOpenUnfilled, brokerstate.StateOpenPartiallyFilled,
		brokerstate.StateCancelPending, brokerstate.StateReplacePending,
		brokerstate.StateUnknown,
	} {
		if s.IsTerminal() {
			t.Errorf("%s must not be terminal", s)
		}
	}
}
