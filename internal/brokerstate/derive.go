// Package brokerstate derives an order's true state from what the official Toss
// Open API actually reports.
//
// The API's raw status field is effectively two values, OPEN and CLOSED. That is
// not enough to act on: "CLOSED" covers fully filled, cancelled, cancelled after a
// partial fill, and replaced-by-an-amendment, and the difference between them
// decides whether the engine is flat, still exposed, or about to double up. The
// state therefore comes from a priority table over
// (status, canceledAt, execution.filledQuantity, quantity, lineage) — never from
// status alone.
//
// The table's other job is to refuse to guess. Any contradiction and any status
// value we have not seen resolves to UNKNOWN_BROKER_STATE, which the caller must
// treat as fail-closed: block new entries for that symbol and alert. An enum that
// is incomplete because the broker has undocumented states is then safe by
// construction — a new value blocks trading instead of being misread as "fine".
package brokerstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// State is the derived order state.
type State string

const (
	// StateOpenUnfilled — live at the broker, nothing filled yet.
	StateOpenUnfilled State = "OPEN_UNFILLED"
	// StateOpenPartiallyFilled — live at the broker with a partial fill.
	StateOpenPartiallyFilled State = "OPEN_PARTIALLY_FILLED"
	// StateFilled — terminal, the whole quantity filled.
	StateFilled State = "FILLED"
	// StateCancelled — terminal, cancelled with nothing filled.
	StateCancelled State = "CANCELLED"
	// StateCancelledPartiallyFilled — terminal, cancelled after a partial fill.
	StateCancelledPartiallyFilled State = "CANCELLED_PARTIALLY_FILLED"
	// StateReplaced — terminal for this order number: an amendment closed it and
	// a successor order carries the remainder. The exposure did not go away, it
	// moved, which is why this is never reported as CANCELLED.
	StateReplaced State = "REPLACED"
	// StateUnknown — the fail-closed outcome. Block new entries for the symbol,
	// alert, and resolve by hand or by the IN_DOUBT procedure.
	StateUnknown State = "UNKNOWN_BROKER_STATE"
)

// IsTerminal reports whether the order can no longer change at the broker.
// UNKNOWN_BROKER_STATE is not terminal: it is an unresolved observation, not an
// end state.
func (s State) IsTerminal() bool {
	switch s {
	case StateFilled, StateCancelled, StateCancelledPartiallyFilled, StateReplaced:
		return true
	default:
		return false
	}
}

// ReasonCode explains a derivation outcome. The strings are a stable contract:
// they are written to the journal, structured logs and operator alerts.
type ReasonCode string

const (
	// ReasonNone is the empty reason of an unambiguous derivation.
	ReasonNone ReasonCode = ""

	ReasonMissingStatus       ReasonCode = "missing_status"
	ReasonUnknownStatus       ReasonCode = "unknown_status"
	ReasonQuantityNotPositive ReasonCode = "quantity_not_positive"
	ReasonFilledNegative      ReasonCode = "filled_negative"
	// ReasonFilledExceedsQuantity: more filled than ordered. Either the payload
	// is wrong or we are looking at the wrong order; both block trading.
	ReasonFilledExceedsQuantity ReasonCode = "filled_exceeds_quantity"
	ReasonNonFiniteQuantity     ReasonCode = "non_finite_quantity"
	// ReasonOpenWithCancelTimestamp: still open yet carries a cancellation time.
	ReasonOpenWithCancelTimestamp ReasonCode = "open_with_cancel_timestamp"
	// ReasonOpenWithSuccessor: an amendment created a successor but this order is
	// still live — that would be double exposure, so it blocks.
	ReasonOpenWithSuccessor ReasonCode = "open_with_successor"
	// ReasonReplacedWithFullFill: a successor exists for an order that has
	// nothing left to replace.
	ReasonReplacedWithFullFill ReasonCode = "replaced_with_full_fill"
	// ReasonCancelledWithFullFill: cancelled, yet the whole quantity is filled.
	ReasonCancelledWithFullFill ReasonCode = "cancelled_with_full_fill"
	// ReasonClosedWithoutFillOrCancel: closed with no fill, no cancellation
	// timestamp and no successor. Expiry would look like this, and so would a
	// rejection we never saw — indeterminate, so it blocks.
	ReasonClosedWithoutFillOrCancel ReasonCode = "closed_without_fill_or_cancel"
	// ReasonClosedWithUnexplainedRemainder: closed with a remainder that no
	// cancellation or successor accounts for.
	ReasonClosedWithUnexplainedRemainder ReasonCode = "closed_with_unexplained_remainder"
	// ReasonMalformedPayload: the broker payload could not be parsed.
	ReasonMalformedPayload ReasonCode = "malformed_payload"
)

// raw statuses the official API is observed to return.
const (
	rawStatusOpen   = "OPEN"
	rawStatusClosed = "CLOSED"
)

// Lineage is what the journal knows about this order's replacement chain.
type Lineage struct {
	// SuccessorOrderID is the order number an AMEND created to carry this
	// order's remainder, empty when there is none.
	SuccessorOrderID string
}

// OrderView is the derivation's input: exactly the fields the priority table reads.
//
// It is a separate type rather than domain.Order because domain.Order has no
// canceledAt field — the upstream official adapter drops it — and because the
// lineage comes from the journal, not from the broker payload.
type OrderView struct {
	OrderID string
	// RawStatus is the broker's status string, unmodified.
	RawStatus string
	// Canceled reports that the broker signalled a cancellation. It is set
	// independently of CanceledAt so an unparseable timestamp still counts as
	// evidence.
	Canceled bool
	// CanceledAt is the parsed cancellation time when available.
	CanceledAt *time.Time
	// Quantity is the ordered quantity.
	Quantity float64
	// FilledQuantity is the broker's cumulative filled quantity (there is no
	// per-fill id in this API).
	FilledQuantity float64
	// Lineage carries the journal's replacement knowledge.
	Lineage Lineage
}

// Derived is the outcome of the priority table.
type Derived struct {
	OrderID string
	State   State
	// Terminal mirrors State.IsTerminal().
	Terminal bool
	// FailClosed is true exactly when State is UNKNOWN_BROKER_STATE. The caller
	// must block new entries for the symbol and alert.
	FailClosed bool
	Reason     ReasonCode
	// Detail is a human-readable explanation for logs and alerts.
	Detail string

	FilledQuantity    float64
	RemainingQuantity float64
	SuccessorOrderID  string
}

// qtyEpsilon is the tolerance for decimal quantities that arrive as strings and
// are compared as float64. It is relative so it works for both a 1,000,000-share
// KR order and a 0.0001 fractional US share.
const qtyEpsilon = 1e-9

func qtyEqual(a, b float64) bool {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= qtyEpsilon*scale
}

func qtyGreater(a, b float64) bool { return a > b && !qtyEqual(a, b) }

// Derive applies the priority table. First match wins:
//
//	#   condition                                                  result
//	--  --------------------------------------------------------   ---------------------------------
//	 1  quantity or filled not finite                              UNKNOWN non_finite_quantity
//	 2  quantity <= 0                                              UNKNOWN quantity_not_positive
//	 3  filled < 0                                                 UNKNOWN filled_negative
//	 4  filled > quantity                                          UNKNOWN filled_exceeds_quantity
//	 5  status empty                                               UNKNOWN missing_status
//	 6  status not OPEN/CLOSED                                     UNKNOWN unknown_status
//	 7  OPEN + cancellation evidence                               UNKNOWN open_with_cancel_timestamp
//	 8  OPEN + successor                                           UNKNOWN open_with_successor
//	 9  OPEN + filled == 0                                         OPEN_UNFILLED
//	10  OPEN                                                       OPEN_PARTIALLY_FILLED
//	11  CLOSED + successor + filled == quantity                    UNKNOWN replaced_with_full_fill
//	12  CLOSED + successor                                         REPLACED
//	13  CLOSED + cancelled + filled == quantity                    UNKNOWN cancelled_with_full_fill
//	14  CLOSED + cancelled + filled == 0                           CANCELLED
//	15  CLOSED + cancelled                                         CANCELLED_PARTIALLY_FILLED
//	16  CLOSED + filled == quantity                                FILLED
//	17  CLOSED + filled == 0                                       UNKNOWN closed_without_fill_or_cancel
//	18  CLOSED                                                     UNKNOWN closed_with_unexplained_remainder
//
// Ordering rationale for the two cases that matter most:
//
//   - The successor check (11/12) is evaluated before the cancellation check
//     (13-15). An amendment cancels the original *and* creates a replacement, so a
//     row with both a canceledAt and a successor is a replace. Reporting it as
//     CANCELLED would tell the engine it is flat while a live order still carries
//     the remainder.
//   - A full fill combined with a cancellation or a successor (13, 11) is refused
//     rather than reported as FILLED. There is nothing left to cancel or replace in
//     that case, so the payload and our lineage disagree, and disagreement about a
//     live account is a stop condition.
func Derive(v OrderView) Derived {
	d := Derived{
		OrderID:          v.OrderID,
		FilledQuantity:   v.FilledQuantity,
		SuccessorOrderID: v.Lineage.SuccessorOrderID,
	}
	canceled := v.Canceled || v.CanceledAt != nil
	successor := strings.TrimSpace(v.Lineage.SuccessorOrderID)

	// 1-4: structural validation.
	switch {
	case isNotFinite(v.Quantity) || isNotFinite(v.FilledQuantity):
		return d.unknown(ReasonNonFiniteQuantity,
			fmt.Sprintf("quantity=%v filledQuantity=%v is not a finite number", v.Quantity, v.FilledQuantity))
	case v.Quantity <= 0:
		return d.unknown(ReasonQuantityNotPositive,
			fmt.Sprintf("ordered quantity %v is not positive", v.Quantity))
	case v.FilledQuantity < 0 && !qtyEqual(v.FilledQuantity, 0):
		return d.unknown(ReasonFilledNegative,
			fmt.Sprintf("filled quantity %v is negative", v.FilledQuantity))
	case qtyGreater(v.FilledQuantity, v.Quantity):
		return d.unknown(ReasonFilledExceedsQuantity,
			fmt.Sprintf("filled %v exceeds ordered %v", v.FilledQuantity, v.Quantity))
	}

	filledAll := qtyEqual(v.FilledQuantity, v.Quantity)
	filledNone := qtyEqual(v.FilledQuantity, 0)
	d.RemainingQuantity = v.Quantity - v.FilledQuantity
	if filledAll {
		d.RemainingQuantity = 0
	}

	// 5-6: status must be one we know.
	status := strings.ToUpper(strings.TrimSpace(v.RawStatus))
	switch status {
	case "":
		return d.unknown(ReasonMissingStatus, "broker returned no status")
	case rawStatusOpen:
		// 7-10.
		switch {
		case canceled:
			return d.unknown(ReasonOpenWithCancelTimestamp,
				"status is OPEN but the broker reported a cancellation")
		case successor != "":
			return d.unknown(ReasonOpenWithSuccessor,
				"status is OPEN while lineage says order "+successor+" replaced it")
		case filledNone:
			return d.settle(StateOpenUnfilled)
		default:
			return d.settle(StateOpenPartiallyFilled)
		}
	case rawStatusClosed:
		// 11-18.
		switch {
		case successor != "" && filledAll:
			return d.unknown(ReasonReplacedWithFullFill,
				"lineage says order "+successor+" replaced this one, but it is fully filled")
		case successor != "":
			return d.settle(StateReplaced)
		case canceled && filledAll:
			return d.unknown(ReasonCancelledWithFullFill,
				"the broker reported a cancellation on a fully filled order")
		case canceled && filledNone:
			return d.settle(StateCancelled)
		case canceled:
			return d.settle(StateCancelledPartiallyFilled)
		case filledAll:
			return d.settle(StateFilled)
		case filledNone:
			return d.unknown(ReasonClosedWithoutFillOrCancel,
				"status is CLOSED with no fill, no cancellation and no successor")
		default:
			return d.unknown(ReasonClosedWithUnexplainedRemainder,
				fmt.Sprintf("status is CLOSED with %v of %v filled and nothing accounting for the remainder",
					v.FilledQuantity, v.Quantity))
		}
	default:
		return d.unknown(ReasonUnknownStatus,
			"broker status "+strconv.Quote(v.RawStatus)+" is not one this build understands")
	}
}

func (d Derived) settle(state State) Derived {
	d.State = state
	d.Terminal = state.IsTerminal()
	d.FailClosed = false
	d.Reason = ReasonNone
	if state != StateReplaced {
		d.SuccessorOrderID = ""
	}
	return d
}

func (d Derived) unknown(reason ReasonCode, detail string) Derived {
	d.State = StateUnknown
	d.Terminal = false
	d.FailClosed = true
	d.Reason = reason
	d.Detail = detail
	return d
}

func isNotFinite(f float64) bool { return math.IsNaN(f) || math.IsInf(f, 0) }

// FromDomainOrder builds a view from the canonical order type.
//
// canceledAt has to be passed separately: the official adapter
// (internal/official.adaptOrder) does not carry the API's canceledAt into
// domain.Order, and lineage is journal knowledge rather than broker payload.
func FromDomainOrder(o domain.Order, canceledAt *time.Time, lineage Lineage) OrderView {
	return OrderView{
		OrderID:        o.ID,
		RawStatus:      o.Status,
		Canceled:       canceledAt != nil,
		CanceledAt:     canceledAt,
		Quantity:       o.Quantity,
		FilledQuantity: o.FilledQuantity,
		Lineage:        lineage,
	}
}

// ErrMalformedOrder means a broker payload could not be turned into a view.
var ErrMalformedOrder = errors.New("brokerstate: malformed order payload")

// officialOrderPayload mirrors the subset of the official Order schema the
// derivation needs. It is intentionally its own struct: internal/official keeps its
// adapter private and drops canceledAt, and this package must not change upstream
// files to read a field it needs.
type officialOrderPayload struct {
	OrderID    string  `json:"orderId"`
	Status     string  `json:"status"`
	Quantity   *string `json:"quantity"`
	CanceledAt *string `json:"canceledAt"`
	Execution  *struct {
		FilledQuantity *string `json:"filledQuantity"`
	} `json:"execution"`
}

// cancelTimeLayouts are tried in order. The API is documented as RFC3339; the
// others exist so a format surprise degrades into "cancelled at an unknown time"
// rather than into "not cancelled".
var cancelTimeLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// ParseOfficialOrder builds a view from a single-order JSON payload, with or
// without the `{"result": …}` envelope the official client unwraps.
func ParseOfficialOrder(data []byte, lineage Lineage) (OrderView, error) {
	body := unwrapResult(data)

	var raw officialOrderPayload
	if err := json.Unmarshal(body, &raw); err != nil {
		return OrderView{}, fmt.Errorf("%w: %v", ErrMalformedOrder, err)
	}

	quantity, err := parseDecimal(raw.Quantity)
	if err != nil {
		return OrderView{}, fmt.Errorf("%w: quantity: %v", ErrMalformedOrder, err)
	}
	var filled float64
	if raw.Execution != nil {
		filled, err = parseDecimal(raw.Execution.FilledQuantity)
		if err != nil {
			return OrderView{}, fmt.Errorf("%w: execution.filledQuantity: %v", ErrMalformedOrder, err)
		}
	}

	view := OrderView{
		OrderID:        raw.OrderID,
		RawStatus:      raw.Status,
		Quantity:       quantity,
		FilledQuantity: filled,
		Lineage:        lineage,
	}
	if raw.CanceledAt != nil && strings.TrimSpace(*raw.CanceledAt) != "" {
		view.Canceled = true
		for _, layout := range cancelTimeLayouts {
			if ts, err := time.Parse(layout, strings.TrimSpace(*raw.CanceledAt)); err == nil {
				utc := ts.UTC()
				view.CanceledAt = &utc
				break
			}
		}
	}
	return view, nil
}

// DeriveOfficialOrder parses and derives in one step, turning a malformed payload
// into UNKNOWN_BROKER_STATE so the caller cannot forget to fail closed.
func DeriveOfficialOrder(data []byte, lineage Lineage) Derived {
	view, err := ParseOfficialOrder(data, lineage)
	if err != nil {
		return Derived{}.unknown(ReasonMalformedPayload, err.Error())
	}
	return Derive(view)
}

// unwrapResult returns the inner object of an official response envelope, or the
// input unchanged when there is no envelope.
func unwrapResult(data []byte) []byte {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return data
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return data
	}
	return envelope.Result
}

// parseDecimal converts a nullable decimal string. A null or empty value is 0,
// which is what the API means for a pending order's execution; anything
// non-numeric is an error rather than a silent zero.
func parseDecimal(s *string) (float64, error) {
	if s == nil {
		return 0, nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a decimal", trimmed)
	}
	return v, nil
}
