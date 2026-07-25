package engine

// precheck.go implements the cancel/amend pre-check the official Open API does
// not offer as an endpoint (harden-execution-base task 4.1).
//
// # What upstream expects and why it cannot be met directly
//
// trading.Service calls Broker.GetOrderAvailableActions before every cancel and
// every amend. The web-session client answers it from a WTS endpoint, and the
// hybrid broker therefore routes it to WTS *always* (internal/hybrid/broker.go).
// For the engine that would be fatal in the literal sense: every cancel — the
// exit path — would depend on a scraped browser session that expires without
// warning. engine-safety says so directly: "WTS 세션이 없거나 만료된 상태에서도
// 엔진의 cancel/amend는 동작해야 한다".
//
// So the engine derives the answer instead, from the one authoritative read it
// does have: OrderRawByID → brokerstate.Derive (the priority table over
// status/canceledAt/filledQuantity/quantity). Raw JSON rather than OrderByID
// because domain.Order drops canceledAt, and canceledAt is the field that
// separates "cancelled" from "filled" inside the API's single CLOSED status
// (issues.md, 2026-07-26).
//
// # Why this function cannot fail
//
// It returns no error. Ever. That is not laziness, it is WORKFLOW §0.3: the
// return value of GetOrderAvailableActions is discarded by trading.Service,
// which reacts only to the *error* — so any error here aborts the cancel. A
// pre-check that can abort a cancel is a pre-check that can trap the engine in a
// position during an outage, which is precisely the failure §0.3 forbids adding
// to the exit path. A broker read that times out, 500s or 404s therefore reports
// "not checked" and lets the mutation proceed; the cancel call's own response is
// the authoritative answer anyway, and it is one the broker cannot get wrong.
//
// The booleans in the returned map are advisory for the same reason and are
// documented as such: nothing may gate an exit on them.
//
// # Budget (§0.4)
//
// One GET /api/v1/orders/{id} per cancel or amend. The retry matrix already
// accounts for single-order reads at "in-flight 주문당 3s마다 1회"; a mutation
// happens far less often than that, so this adds no new rate-limit line. The call
// is bounded by precheckTimeout so a hanging broker delays an exit by at most
// that, rather than by the caller's whole context.

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
)

// precheckTimeout bounds the pre-check read.
//
// Two seconds is a deliberately small number: this call sits in front of a
// cancel, and a cancel that is late is a cancel that failed. The value is not
// configurable because making it configurable would make it possible to
// configure the exit path into being slow.
const precheckTimeout = 2 * time.Second

// Keys of the map GetOrderAvailableActions returns. They are named constants
// because the map is read by operators in logs and by tests, and a typo in a
// string literal is not a compile error.
const (
	// ActionKeyChecked reports whether the broker read succeeded. False means
	// every other key is a default, not an observation.
	ActionKeyChecked = "checked"
	// ActionKeyOrderID echoes the order the pre-check was asked about.
	ActionKeyOrderID = "order_id"
	// ActionKeyState is the derived brokerstate.State.
	ActionKeyState = "state"
	// ActionKeyTerminal reports that the order can no longer change.
	ActionKeyTerminal = "terminal"
	// ActionKeyFailClosed reports UNKNOWN_BROKER_STATE.
	ActionKeyFailClosed = "fail_closed"
	// ActionKeyReason is the brokerstate.ReasonCode of the derivation.
	ActionKeyReason = "reason"
	// ActionKeyDetail is the human-readable explanation.
	ActionKeyDetail = "detail"
	// ActionKeyFilled is the cumulative filled quantity the broker reported.
	ActionKeyFilled = "filled_quantity"
	// ActionKeyRemaining is quantity - filled.
	ActionKeyRemaining = "remaining_quantity"
	// ActionKeySuccessor is the order id an amendment created, when known.
	ActionKeySuccessor = "successor_order_id"

	// ActionKeyCancel is ADVISORY ONLY: true unless the order is provably
	// terminal. Nothing may refuse a cancel because this is false — see §0.3
	// above. It exists so an operator reading a log can see what the engine knew.
	ActionKeyCancel = "cancel"
	// ActionKeyAmend is ADVISORY ONLY: true only when the order is derivably
	// still open. An amend can raise exposure, so unlike cancel it is not
	// optimistic about states it could not read.
	ActionKeyAmend = "amend"
)

// availableActions derives the pre-check answer. It never returns an error; see
// the file comment.
func (b *officialBroker) availableActions(ctx context.Context, orderID string) map[string]any {
	pctx, cancel := context.WithTimeout(ctx, precheckTimeout)
	defer cancel()

	raw, err := b.off.OrderRawByID(pctx, orderID)
	if err != nil {
		return inconclusiveActions(orderID, "the order could not be read: "+err.Error())
	}

	// No lineage is supplied: the journal is the only place a successor is
	// recorded and the broker adapter deliberately has no journal handle (it is
	// wrapped by execgw, which does). The cost is bounded and lands in the safe
	// direction — an amended order reads as CLOSED with an unexplained remainder,
	// which derives to UNKNOWN_BROKER_STATE, which reports amend=false.
	d := brokerstate.DeriveOfficialOrder(raw, brokerstate.Lineage{})

	return map[string]any{
		ActionKeyChecked:    true,
		ActionKeyOrderID:    orderID,
		ActionKeyState:      string(d.State),
		ActionKeyTerminal:   d.Terminal,
		ActionKeyFailClosed: d.FailClosed,
		ActionKeyReason:     string(d.Reason),
		ActionKeyDetail:     d.Detail,
		ActionKeyFilled:     d.FilledQuantity,
		ActionKeyRemaining:  d.RemainingQuantity,
		ActionKeySuccessor:  d.SuccessorOrderID,
		ActionKeyCancel:     !d.Terminal,
		ActionKeyAmend:      isOpen(d.State),
	}
}

// inconclusiveActions is what an unreadable order reports: the mutation is
// allowed to proceed, and the record says the engine did not know.
func inconclusiveActions(orderID, detail string) map[string]any {
	return map[string]any{
		ActionKeyChecked:    false,
		ActionKeyOrderID:    orderID,
		ActionKeyState:      "",
		ActionKeyTerminal:   false,
		ActionKeyFailClosed: false,
		ActionKeyReason:     "",
		ActionKeyDetail:     detail,
		ActionKeyFilled:     0.0,
		ActionKeyRemaining:  0.0,
		ActionKeySuccessor:  "",
		// Optimistic for the exit, pessimistic for the entry.
		ActionKeyCancel: true,
		ActionKeyAmend:  false,
	}
}

func isOpen(s brokerstate.State) bool {
	return s == brokerstate.StateOpenUnfilled || s == brokerstate.StateOpenPartiallyFilled
}
