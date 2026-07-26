package execgw

// roundtrip.go confirms that an order the broker said it created actually exists
// (extend-execution-contract task 5.2).
//
// # Why an ack is not proof
//
// A place answers with an order id. That id is an opaque token — openapi
// contracts no shape, no prefix and no length for `orderId` — so there is nothing
// about the value itself that can be checked, and "the response parsed" is not
// the same fact as "an order exists". Settling the attempt as CONFIRMED on the
// strength of the ack alone would put an unverifiable id into the journal as the
// engine's handle on a live position.
//
// So the created id is read back through the detail endpoint before the attempt
// settles. The order-execution spec states the outcome of a failure directly:
// "생성 응답의 식별자는 상세조회 round-trip으로 실재를 확인하며(SHALL),
// round-trip 실패·부재 시 CONFIRMED로 종결하지 않고 IN_DOUBT로 남겨 해소 절차가
// 확정한다(SHALL)". IN_DOUBT is not an error state here — it is the honest one,
// and the resolution procedure is what turns it into an answer.
//
// # What is checked, and what is deliberately not
//
// Checked: the read succeeds, the record it returns carries the *same bytes* as
// the id we were given, and its symbol is the one this mutation was for. Not
// checked: anything about the shape of the id. No regex, no prefix, no length —
// adding one would invent a contract the broker never made and would fail closed
// on the day the broker changes its id format (openapi's example is a 64-char
// random token, which is a fact about today, not a promise).
//
// # Budget (§0.4)
//
// One GET /api/v1/orders/{orderId} per place. Places are rare compared to the
// polling reads the retry matrix budgets for, and the call is bounded by
// roundTripTimeout so a hanging broker delays a *place* — never an exit — by at
// most that.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// roundTripTimeout bounds the read-back.
//
// A place that has already been acked is not made safer by waiting longer: an
// unanswered read leaves the attempt IN_DOUBT, which is where a slow broker
// should land it anyway.
const roundTripTimeout = 3 * time.Second

// errNoOrderReader is what a gateway without an order reader would produce. It is
// never returned in practice — the verifier is only installed when a reader is
// wired — and exists so the failure is named rather than silent.
var errNoOrderReader = errors.New("execgw: no order reader is configured, so a created order cannot be read back")

// roundTripFor returns the existence check for one mutation, or nil when there
// is nothing to confirm.
//
// Only a place is checked. A place is the mutation that *creates* the record the
// engine will then reason about, and its id is the handle every later cancel,
// fill and reconciliation hangs off. A cancel creates no record of its own, and
// an amend's successor id is confirmed by the amend resolution procedure against
// the original's lineage — its record shape is [미측정 — 2b 2.1], so a naive
// read-back would be checking something the documentation does not yet describe.
//
// §0.3: this never runs in front of an exit. Cancels are untouched.
func (g *Gateway) roundTripFor(plan mutationPlan) journal.ExistenceCheck {
	if g.orders == nil || plan.kind != journal.KindPlace {
		return nil
	}
	return func(ctx context.Context, brokerOrderID string) error {
		return g.confirmCreatedOrder(ctx, plan, brokerOrderID)
	}
}

// confirmCreatedOrder reads back the order the broker said it created.
//
// A nil return means confirmed. Any error means "not confirmed", which the
// journal turns into IN_DOUBT rather than CONFIRMED — including the case where
// the record exists but belongs somewhere else.
func (g *Gateway) confirmCreatedOrder(ctx context.Context, plan mutationPlan, brokerOrderID string) error {
	if g.orders == nil {
		return errNoOrderReader
	}
	rctx, cancel := context.WithTimeout(ctx, roundTripTimeout)
	defer cancel()

	raw, err := g.orders.OrderRaw(rctx, brokerOrderID)
	if err != nil {
		return fmt.Errorf("reading order %q back: %w", brokerOrderID, err)
	}
	facts, err := parseOrderFacts(unwrapOrderEnvelope(raw))
	if err != nil {
		return fmt.Errorf("the read-back of order %q could not be read: %w", brokerOrderID, err)
	}

	// Byte-exact. An id that differs only in whitespace or case is a different
	// id: we have no rule from the broker that says otherwise, and the safe
	// reading of "the record I got back is not the one I asked for" is doubt.
	if facts.OrderID != brokerOrderID {
		return fmt.Errorf("the broker acked order %q but the read-back names %q",
			brokerOrderID, facts.OrderID)
	}

	// TODO(2a-4.1): RECONCILE 상태로 전이 — "같은 브로커 식별자가 상충하는
	// 계좌·심볼 컨텍스트에 출현"은 RECONCILE 진입 조건이다. 상태 저장소가 생기기
	// 전까지는 fail-closed 경로(IN_DOUBT + 심볼 차단)로 기록한다.
	//
	// Account scope is not compared here because the read is already
	// account-scoped: OrderRawByID sends the X-Tossinvest-Account header, so a
	// record belonging to another account cannot come back through it. The symbol
	// is the conflicting-context dimension this read can actually see.
	if facts.Symbol != "" && plan.symbol != "" && facts.Symbol != plan.symbol {
		return fmt.Errorf(
			"order %q was acked for %s but the broker reports it on %s — the same identifier appears in a conflicting context",
			brokerOrderID, plan.symbol, facts.Symbol)
	}
	return nil
}

// unwrapOrderEnvelope returns the inner object of an official `{"result": …}`
// response, or the input unchanged when there is no envelope. The client unwraps
// it on the read paths that decode into a struct, and passes it through on the
// raw ones, so both shapes reach here.
func unwrapOrderEnvelope(data json.RawMessage) json.RawMessage {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return data
	}
	if len(envelope.Result) == 0 || strings.TrimSpace(string(envelope.Result)) == "null" {
		return data
	}
	return envelope.Result
}
