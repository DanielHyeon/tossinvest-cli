// Package brokerstate derives an order's true state from what the official Toss
// Open API actually reports.
//
// # The input is a ten-value enum
//
// docs/migration/openapi.latest.json documents `OrderStatus` as ten values —
// PENDING, PENDING_CANCEL, PENDING_REPLACE, PARTIAL_FILLED, FILLED, CANCELED,
// REJECTED, CANCEL_REJECTED, REPLACE_REJECTED, REPLACED — and instructs clients to
// "unknown code 를 허용하도록" implement. Every per-order row carries one of them,
// in the list endpoint as well as the detail endpoint.
//
// OPEN and CLOSED are *not* order statuses. They are the group labels of the
// `status` query parameter of GET /api/v1/orders: "이 값은 각 주문의 세부
// 상태(`orders[].status`)를 **그룹화한 라벨**이며, `orders[].status` 와 값 체계가
// 다릅니다" (openapi). An earlier build read them as the whole status vocabulary,
// which sent every real payload to UNKNOWN_BROKER_STATE. They survive here only as
// a compatibility shim for recorded fixtures and older journal rows
// (deriveGroupLabel), never as the model.
//
// # Why a priority table and not a switch on status
//
// status alone does not say whether the engine is flat, still exposed, or about to
// double up: CANCELED and REJECTED both carry a possible partial fill
// ("execution.filledQuantity를 통해 부분 체결 여부를 확인할 수 있음" — openapi),
// REPLACED means the exposure moved to another order rather than going away, and
// PENDING_CANCEL/PENDING_REPLACE are live orders with a mutation in flight. The
// state therefore comes from a priority table over
// (status, canceledAt, execution.filledQuantity, quantity, lineage).
//
// # The table's other job is to refuse to guess
//
// Any combination the documentation cannot explain, and any status value this
// build does not know, resolves to UNKNOWN_BROKER_STATE, which the caller must
// treat as fail-closed: block new entries for that symbol and alert. An enum that
// is incomplete because the broker has undocumented states is then safe by
// construction — a new value blocks trading instead of being misread as "fine".
// That is also how the two separate-record statuses are handled: CANCEL_REJECTED
// and REPLACE_REJECTED are "별도 주문 레코드로 생성됨"(openapi) whose concrete
// shape is [미측정 — 2b 2.1], so an unattributed record of either kind blocks
// instead of being classified as somebody else's order.
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
	// StateRejected — terminal, the broker refused the order and nothing filled.
	StateRejected State = "REJECTED"
	// StateRejectedPartiallyFilled — terminal, refused after a partial fill.
	// openapi documents REJECTED as carrying execution.filledQuantity, so a
	// rejection is not automatically "no position".
	StateRejectedPartiallyFilled State = "REJECTED_PARTIALLY_FILLED"
	// StateCancelPending — live: a cancel request is accepted and awaiting the
	// broker's answer (raw PENDING_CANCEL). Not terminal — the order can still
	// fill, and the cancel can still be refused as a CANCEL_REJECTED record.
	StateCancelPending State = "CANCEL_PENDING"
	// StateReplacePending — live: a replace request is awaiting the broker's
	// answer (raw PENDING_REPLACE). Not terminal, same reasoning.
	StateReplacePending State = "REPLACE_PENDING"
	// StateUnknown — the fail-closed outcome. Block new entries for the symbol,
	// alert, and resolve by hand or by the IN_DOUBT procedure.
	StateUnknown State = "UNKNOWN_BROKER_STATE"
)

// IsTerminal reports whether the order can no longer change at the broker.
// UNKNOWN_BROKER_STATE is not terminal: it is an unresolved observation, not an
// end state.
func (s State) IsTerminal() bool {
	switch s {
	case StateFilled, StateCancelled, StateCancelledPartiallyFilled, StateReplaced,
		StateRejected, StateRejectedPartiallyFilled:
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

	// --- contradictions the ten-value enum makes expressible -----------------

	// ReasonPendingWithFill: PENDING is "체결 대기"(openapi) — an order with a
	// fill is PARTIAL_FILLED or FILLED, so a fill under PENDING means the payload
	// and the enum disagree.
	ReasonPendingWithFill ReasonCode = "pending_with_fill"
	// ReasonPartialFilledWithoutFill: PARTIAL_FILLED with nothing filled.
	ReasonPartialFilledWithoutFill ReasonCode = "partial_filled_without_fill"
	// ReasonPartialFilledWithFullFill: PARTIAL_FILLED with everything filled —
	// that is FILLED, and the difference decides whether the order is still live.
	ReasonPartialFilledWithFullFill ReasonCode = "partial_filled_with_full_fill"
	// ReasonPendingMutationWithFullFill: a cancel or a replace is reported in
	// flight on an order with nothing left to cancel or replace.
	ReasonPendingMutationWithFullFill ReasonCode = "pending_mutation_with_full_fill"
	// ReasonFilledWithRemainder: FILLED is "주문 수량이 전량 체결된
	// 상태"(openapi), so a remainder contradicts the status.
	ReasonFilledWithRemainder ReasonCode = "filled_with_remainder"
	// ReasonCancelledWithSuccessor: the broker says plainly cancelled while our
	// lineage says an amendment moved the exposure to another order. A cancel and
	// a replace have opposite exposure meanings, so this is a stop condition.
	ReasonCancelledWithSuccessor ReasonCode = "cancelled_with_successor"
	// ReasonRejectedWithFullFill: refused, yet the whole quantity is filled.
	ReasonRejectedWithFullFill ReasonCode = "rejected_with_full_fill"
	// ReasonRejectedWithSuccessor: refused, yet lineage says it was replaced.
	ReasonRejectedWithSuccessor ReasonCode = "rejected_with_successor"
	// ReasonReplacedWithoutSuccessor: the broker says the order was replaced but
	// nothing tells us which order carries the remainder. The exposure moved to an
	// order we cannot name, which is exactly what must not be reported as a clean
	// terminal state.
	ReasonReplacedWithoutSuccessor ReasonCode = "replaced_without_successor"
	// ReasonCancelRejectedRecord: a CANCEL_REJECTED row. openapi says the broker
	// creates it as a separate order record and returns the original to its prior
	// state; the record's concrete shape is [미측정 — 2b 2.1], so attribution
	// cannot be automated yet and the row blocks instead of being read as an
	// external order.
	ReasonCancelRejectedRecord ReasonCode = "cancel_rejected_record"
	// ReasonReplaceRejectedRecord: a REPLACE_REJECTED row, same reasoning.
	ReasonReplaceRejectedRecord ReasonCode = "replace_rejected_record"
)

// The documented OrderStatus values (openapi.latest.json,
// components.schemas.OrderStatus).
const (
	rawStatusPending         = "PENDING"
	rawStatusPendingCancel   = "PENDING_CANCEL"
	rawStatusPendingReplace  = "PENDING_REPLACE"
	rawStatusPartialFilled   = "PARTIAL_FILLED"
	rawStatusFilled          = "FILLED"
	rawStatusCanceled        = "CANCELED"
	rawStatusRejected        = "REJECTED"
	rawStatusCancelRejected  = "CANCEL_REJECTED"
	rawStatusReplaceRejected = "REPLACE_REJECTED"
	rawStatusReplaced        = "REPLACED"
)

// The group labels of the GET /api/v1/orders `status` *request parameter*. They
// are not order statuses (openapi: "값 체계가 다릅니다") and are accepted only as
// a compatibility shim — see deriveGroupLabel.
const (
	rawGroupOpen   = "OPEN"
	rawGroupClosed = "CLOSED"
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

// facts is the normalised input the table rows read: the raw fields reduced to
// the predicates the rows are written in terms of.
type facts struct {
	canceled   bool
	successor  string
	filledAll  bool
	filledNone bool
	quantity   float64
	filled     float64
}

// Derive applies the priority table. First match wins:
//
//	#   condition                                                    result
//	--  ----------------------------------------------------------   ---------------------------------
//	 1  quantity or filled not finite                                UNKNOWN non_finite_quantity
//	 2  quantity <= 0                                                UNKNOWN quantity_not_positive
//	 3  filled < 0                                                   UNKNOWN filled_negative
//	 4  filled > quantity                                            UNKNOWN filled_exceeds_quantity
//	 5  status empty                                                 UNKNOWN missing_status
//	    PENDING — 체결 대기 (openapi)
//	 6  + cancellation evidence                                      UNKNOWN open_with_cancel_timestamp
//	 7  + successor                                                  UNKNOWN open_with_successor
//	 8  + filled > 0                                                 UNKNOWN pending_with_fill
//	 9  otherwise                                                    OPEN_UNFILLED
//	    PARTIAL_FILLED — 부분 체결 (openapi; in both list groups)
//	10  + cancellation evidence                                      UNKNOWN open_with_cancel_timestamp
//	11  + successor                                                  UNKNOWN open_with_successor
//	12  + filled == 0                                                UNKNOWN partial_filled_without_fill
//	13  + filled == quantity                                         UNKNOWN partial_filled_with_full_fill
//	14  otherwise                                                    OPEN_PARTIALLY_FILLED
//	    PENDING_CANCEL / PENDING_REPLACE — 취소·정정 대기 (openapi)
//	15  + cancellation evidence                                      UNKNOWN open_with_cancel_timestamp
//	16  + successor                                                  UNKNOWN open_with_successor
//	17  + filled == quantity                                         UNKNOWN pending_mutation_with_full_fill
//	18  otherwise                                                    CANCEL_PENDING | REPLACE_PENDING
//	    FILLED — 전량 체결 (openapi)
//	19  + filled != quantity                                         UNKNOWN filled_with_remainder
//	20  + successor                                                  UNKNOWN replaced_with_full_fill
//	21  + cancellation evidence                                      UNKNOWN cancelled_with_full_fill
//	22  otherwise                                                    FILLED
//	    CANCELED — 취소 완료 (openapi)
//	23  + successor                                                  UNKNOWN cancelled_with_successor
//	24  + filled == quantity                                         UNKNOWN cancelled_with_full_fill
//	25  + filled == 0                                                CANCELLED
//	26  otherwise                                                    CANCELLED_PARTIALLY_FILLED
//	    REJECTED — 거부됨 (openapi; 부분 체결 가능)
//	27  + successor                                                  UNKNOWN rejected_with_successor
//	28  + filled == quantity                                         UNKNOWN rejected_with_full_fill
//	29  + filled == 0                                                REJECTED
//	30  otherwise                                                    REJECTED_PARTIALLY_FILLED
//	    REPLACED — 정정됨 (openapi)
//	31  + filled == quantity                                         UNKNOWN replaced_with_full_fill
//	32  + no successor known                                         UNKNOWN replaced_without_successor
//	33  otherwise                                                    REPLACED
//	34  CANCEL_REJECTED                                              UNKNOWN cancel_rejected_record
//	35  REPLACE_REJECTED                                             UNKNOWN replace_rejected_record
//	36  OPEN / CLOSED (legacy group labels)                          deriveGroupLabel
//	37  anything else                                                UNKNOWN unknown_status
//
// Ordering rationale for the cases that matter most, carried over from the
// two-value table this replaces:
//
//   - A full fill combined with a cancellation, a replacement or a pending
//     mutation (21, 24, 20, 31, 17) is refused rather than reported as FILLED.
//     There is nothing left to cancel or replace in that case, so the payload and
//     our lineage disagree, and disagreement about a live account is a stop
//     condition.
//   - REPLACED without a successor we can name (32) blocks. The status says the
//     exposure moved rather than went away, so reporting it as a clean terminal
//     state would tell the engine it is flat while a live order carries the
//     remainder. The IN_DOUBT amend procedure recognises this reason and goes
//     looking for the successor; everyone else fails closed.
//   - CANCELED with a recorded successor (23) blocks rather than preferring one
//     side. In the two-value table the successor won, because CLOSED could not
//     distinguish a cancel from a replace. The enum can: an amended original is
//     REPLACED, so "plainly cancelled" plus "we recorded a replacement" is a real
//     contradiction, not an ambiguity to resolve by precedence.
//   - A cancellation timestamp contradicts a *live* status (6, 10, 15) but
//     corroborates a terminal one. On REPLACED it is expected — an amendment
//     cancels the original — and on REJECTED it is how an order leaves the book,
//     so neither blocks on it.
func Derive(v OrderView) Derived {
	d := Derived{
		OrderID:          v.OrderID,
		FilledQuantity:   v.FilledQuantity,
		SuccessorOrderID: v.Lineage.SuccessorOrderID,
	}

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

	f := facts{
		canceled:   v.Canceled || v.CanceledAt != nil,
		successor:  strings.TrimSpace(v.Lineage.SuccessorOrderID),
		filledAll:  qtyEqual(v.FilledQuantity, v.Quantity),
		filledNone: qtyEqual(v.FilledQuantity, 0),
		quantity:   v.Quantity,
		filled:     v.FilledQuantity,
	}
	d.RemainingQuantity = v.Quantity - v.FilledQuantity
	if f.filledAll {
		d.RemainingQuantity = 0
	}

	status := strings.ToUpper(strings.TrimSpace(v.RawStatus))
	switch status {
	case "":
		return d.unknown(ReasonMissingStatus, "broker returned no status")

	case rawStatusPending: // 6-9
		switch {
		case f.canceled:
			return d.unknown(ReasonOpenWithCancelTimestamp,
				"status is PENDING but the broker reported a cancellation")
		case f.successor != "":
			return d.unknown(ReasonOpenWithSuccessor,
				"status is PENDING while lineage says order "+f.successor+" replaced it")
		case !f.filledNone:
			return d.unknown(ReasonPendingWithFill, fmt.Sprintf(
				"status is PENDING (체결 대기) with %v of %v already filled", f.filled, f.quantity))
		default:
			return d.settle(StateOpenUnfilled)
		}

	case rawStatusPartialFilled: // 10-14
		switch {
		case f.canceled:
			return d.unknown(ReasonOpenWithCancelTimestamp,
				"status is PARTIAL_FILLED but the broker reported a cancellation")
		case f.successor != "":
			return d.unknown(ReasonOpenWithSuccessor,
				"status is PARTIAL_FILLED while lineage says order "+f.successor+" replaced it")
		case f.filledNone:
			return d.unknown(ReasonPartialFilledWithoutFill,
				"status is PARTIAL_FILLED with nothing filled")
		case f.filledAll:
			return d.unknown(ReasonPartialFilledWithFullFill, fmt.Sprintf(
				"status is PARTIAL_FILLED with all %v filled", f.quantity))
		default:
			return d.settle(StateOpenPartiallyFilled)
		}

	case rawStatusPendingCancel: // 15-18
		return d.pendingMutation(f, "PENDING_CANCEL", StateCancelPending)

	case rawStatusPendingReplace: // 15-18
		return d.pendingMutation(f, "PENDING_REPLACE", StateReplacePending)

	case rawStatusFilled: // 19-22
		switch {
		case !f.filledAll:
			return d.unknown(ReasonFilledWithRemainder, fmt.Sprintf(
				"status is FILLED (전량 체결) with only %v of %v filled", f.filled, f.quantity))
		case f.successor != "":
			return d.unknown(ReasonReplacedWithFullFill,
				"lineage says order "+f.successor+" replaced this one, but it is fully filled")
		case f.canceled:
			return d.unknown(ReasonCancelledWithFullFill,
				"the broker reported a cancellation on a fully filled order")
		default:
			return d.settle(StateFilled)
		}

	case rawStatusCanceled: // 23-26
		switch {
		case f.successor != "":
			return d.unknown(ReasonCancelledWithSuccessor,
				"status is CANCELED while lineage says order "+f.successor+
					" carries the remainder; a cancel and a replace mean opposite things about the exposure")
		case f.filledAll:
			return d.unknown(ReasonCancelledWithFullFill,
				"the broker reported a cancellation on a fully filled order")
		case f.filledNone:
			return d.settle(StateCancelled)
		default:
			return d.settle(StateCancelledPartiallyFilled)
		}

	case rawStatusRejected: // 27-30
		switch {
		case f.successor != "":
			return d.unknown(ReasonRejectedWithSuccessor,
				"status is REJECTED while lineage says order "+f.successor+" replaced it")
		case f.filledAll:
			return d.unknown(ReasonRejectedWithFullFill,
				"the broker refused an order it also reports as fully filled")
		case f.filledNone:
			return d.settle(StateRejected)
		default:
			return d.settle(StateRejectedPartiallyFilled)
		}

	case rawStatusReplaced: // 31-33
		switch {
		case f.filledAll:
			return d.unknown(ReasonReplacedWithFullFill,
				"status is REPLACED but the order is fully filled, so there was nothing to replace")
		case f.successor == "":
			// TODO(2a-4.1): RECONCILE 상태로 전이 — the exposure moved to an order
			// we cannot name, which is the "로컬 파생과 브로커 스냅샷의 불일치"
			// entry condition. Until the state store exists this is the
			// fail-closed UNKNOWN path (entry block + alert).
			return d.unknown(ReasonReplacedWithoutSuccessor, fmt.Sprintf(
				"status is REPLACED with %v of %v filled, and nothing names the order carrying the remainder",
				f.filled, f.quantity))
		default:
			return d.settle(StateReplaced)
		}

	case rawStatusCancelRejected: // 34
		// TODO(2a-4.1): RECONCILE 상태로 전이 — openapi says the broker creates
		// this as a separate order record and returns the original to its prior
		// state. The record's shape is [미측정 — 2b 2.1], so it must not be
		// attributed automatically and must never be classified as an external
		// order. Fail-closed UNKNOWN until the RECONCILE store exists.
		return d.unknown(ReasonCancelRejectedRecord,
			"status is CANCEL_REJECTED: openapi documents this as a separate order record "+
				"whose shape is unmeasured, so it cannot be attributed automatically")

	case rawStatusReplaceRejected: // 35
		// TODO(2a-4.1): RECONCILE 상태로 전이 — same reasoning as CANCEL_REJECTED.
		return d.unknown(ReasonReplaceRejectedRecord,
			"status is REPLACE_REJECTED: openapi documents this as a separate order record "+
				"whose shape is unmeasured, so it cannot be attributed automatically")

	case rawGroupOpen, rawGroupClosed: // 36
		return d.deriveGroupLabel(status, f)

	default: // 37
		return d.unknown(ReasonUnknownStatus,
			"broker status "+strconv.Quote(v.RawStatus)+" is not one this build understands")
	}
}

// pendingMutation is rows 15-18: a cancel or a replace the broker has accepted
// and not yet answered. The order is still live, so the state is deliberately not
// terminal — nothing may release a reservation or stop tracking on it.
func (d Derived) pendingMutation(f facts, status string, state State) Derived {
	switch {
	case f.canceled:
		return d.unknown(ReasonOpenWithCancelTimestamp,
			"status is "+status+" (the mutation is still in flight) but the broker also reported a completed cancellation")
	case f.successor != "":
		return d.unknown(ReasonOpenWithSuccessor,
			"status is "+status+" while lineage already names order "+f.successor+" as its replacement")
	case f.filledAll:
		return d.unknown(ReasonPendingMutationWithFullFill, fmt.Sprintf(
			"status is %s with all %v filled, so there is nothing left to cancel or replace", status, f.quantity))
	default:
		return d.settle(state)
	}
}

// deriveGroupLabel is the compatibility shim for the OPEN/CLOSED group labels
// (row 36). They are the `status` request parameter's grouping, not order
// statuses (openapi), so a payload carrying one is either a recorded fixture, an
// older journal row, or a caller that passed the request filter by mistake. The
// rows below are the two-value table this file replaced, kept verbatim so those
// inputs derive exactly as they did before.
func (d Derived) deriveGroupLabel(label string, f facts) Derived {
	if label == rawGroupOpen {
		switch {
		case f.canceled:
			return d.unknown(ReasonOpenWithCancelTimestamp,
				"status is OPEN but the broker reported a cancellation")
		case f.successor != "":
			return d.unknown(ReasonOpenWithSuccessor,
				"status is OPEN while lineage says order "+f.successor+" replaced it")
		case f.filledNone:
			return d.settle(StateOpenUnfilled)
		default:
			return d.settle(StateOpenPartiallyFilled)
		}
	}
	switch {
	case f.successor != "" && f.filledAll:
		return d.unknown(ReasonReplacedWithFullFill,
			"lineage says order "+f.successor+" replaced this one, but it is fully filled")
	case f.successor != "":
		return d.settle(StateReplaced)
	case f.canceled && f.filledAll:
		return d.unknown(ReasonCancelledWithFullFill,
			"the broker reported a cancellation on a fully filled order")
	case f.canceled && f.filledNone:
		return d.settle(StateCancelled)
	case f.canceled:
		return d.settle(StateCancelledPartiallyFilled)
	case f.filledAll:
		return d.settle(StateFilled)
	case f.filledNone:
		return d.unknown(ReasonClosedWithoutFillOrCancel,
			"status is CLOSED with no fill, no cancellation and no successor")
	default:
		return d.unknown(ReasonClosedWithUnexplainedRemainder,
			fmt.Sprintf("status is CLOSED with %v of %v filled and nothing accounting for the remainder",
				f.filled, f.quantity))
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
