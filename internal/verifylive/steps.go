package verifylive

// steps.go is the procedure itself: one function per entry in the catalogue.
//
// Every one of them follows the same shape. Read what it needs, mutate at most
// what the catalogue says it will, write an Observation for each thing it
// learned, and return an error only for something that stops the step. The
// verdict is decided by stepRun.resolve, so a step never has to remember whether
// a refusal is a failure (it is not) or whether an unexpected broker answer is
// (it is).
//
// The observation keys are a contract. `verify report` groups on them and task
// 1.4 copies them into the attestation's attribute set, so they are dotted,
// stable, and named after the property rather than after the call that produced
// it. report.go lists the ones the report knows about.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// The broker's two order lifecycle groups, and the only two values
// GET /api/v1/orders accepts for its `status` parameter.
//
// openapi.latest.json marks the parameter required, with enum {OPEN, CLOSED} and no
// member meaning "everything": a request that omits it is answered with 400
// invalid-request naming the field, on every cycle, on a real account
// (measurements.md M7). Reading the whole list therefore means reading both groups
// and taking the union — the shape internal/soak's walk was corrected to in the same
// investigation.
//
// The groups are labels over the per-order status and they are not disjoint: the
// spec puts PARTIAL_FILLED in both. The same order coming back from each of them is
// one order, and nothing here may read that as a broken list.
const (
	statusOpen   = "OPEN"
	statusClosed = "CLOSED"
)

// maxFixturePages bounds the history walk. The status fixtures need variety, not
// completeness, and an account with years of orders should not turn step one into
// a rate-limit incident.
//
// It bounds each group separately rather than the two together. CLOSED is the whole
// history of the account — the walk sets no from/to bound — and is what the cap
// exists to contain; OPEN is not paginated at all, since the spec has it ignore
// cursor and limit and answer with every working order at once. A budget shared
// between them would let a long history spend the whole allowance and leave nothing
// for the OPEN read, which would silently drop every working status from the fixture
// set — and the working statuses are precisely the ones CLOSED can never show.
const maxFixturePages = 10

// documentedOrderStatuses is the OrderStatus enum as the API documents it
// (docs/migration/openapi.latest.json, component "OrderStatus"). The read-fixtures
// step records which of them this account has actually produced, and — more
// usefully — which it has not, because those are the ones the state-derivation
// table is still guessing about.
func documentedOrderStatuses() []string {
	return []string{
		"PENDING", "PENDING_CANCEL", "PENDING_REPLACE", "PARTIAL_FILLED", "FILLED",
		"CANCELED", "REJECTED", "CANCEL_REJECTED", "REPLACE_REJECTED", "REPLACED",
	}
}

// dispatch routes a step to its body.
func (r *Runner) dispatch(ctx context.Context, id StepID, sr *stepRun) {
	var err error
	switch id {
	case StepReadFixtures:
		err = r.stepReadFixtures(ctx, sr)
	case StepSellableBaseline:
		err = r.stepSellableBaseline(ctx, sr)
	case StepIdempotency:
		err = r.stepIdempotency(ctx, sr)
	case StepIdempotencyTTLEdge:
		err = r.stepIdempotencyTTLEdge(ctx, sr)
	case StepOrderCancel:
		err = r.stepOrderCancel(ctx, sr)
	case StepOrderAmend:
		err = r.stepOrderAmend(ctx, sr)
	case StepSellBoundary:
		err = r.stepSellBoundary(ctx, sr)
	case StepConditionalRegister:
		err = r.stepConditionalRegister(ctx, sr)
	case StepSellableReserved:
		err = r.stepSellableReserved(ctx, sr)
	case StepConditionalPersist:
		err = r.stepConditionalPersist(ctx, sr)
	case StepConditionalTrigger:
		err = r.stepConditionalTrigger(sr)
	case StepConditionalModify:
		err = r.stepConditionalModify(ctx, sr)
	case StepConditionalCancel:
		err = r.stepConditionalCancel(ctx, sr)
	case StepCosts:
		err = r.stepCosts(ctx, sr)
	default:
		sr.skip("이 빌드에는 해당 단계의 구현이 없다")
		return
	}
	sr.resolve(err)
}

// --- A. read-only groundwork --------------------------------------------------

// stepReadFixtures collects the status enum fixtures (task 2.1).
func (r *Runner) stepReadFixtures(ctx context.Context, sr *stepRun) error {
	counts := map[string]int{}
	shapes := map[string][]string{}
	var (
		pages  int
		orders int
	)
	// CLOSED then OPEN. Between them they are the whole order list; there is no
	// third call that returns it in one go (see the status constants above), and a
	// walk of one group alone is a fixture set missing every status the other holds.
	//
	// An order that comes back from both groups needs no special handling. What is
	// being collected is the *set* of status values the account has produced, so a
	// PARTIAL_FILLED order that is honestly in each group contributes the one value
	// twice and the set absorbs it; only the per-status count sees the repeat, and
	// occurrences are what that count means.
	for _, group := range []string{statusClosed, statusOpen} {
		cursor := ""
		// The loop is written the same way for both groups on purpose. The spec says
		// CLOSED paginates and OPEN does not, and following whatever the broker hands
		// back is how a group that behaves otherwise gets read rather than truncated
		// on an assumption; the page cap is what keeps that generosity bounded.
		for page := 0; page < maxFixturePages; page++ {
			p, err := readRetry(ctx, r, sr, EndpointReadOrders,
				map[string]string{"status": group, "cursor": cursor},
				func(ctx context.Context) (official.RawOrderPage, error) {
					return r.broker.OrdersPageRaw(ctx, official.OrdersFilter{Status: group}, cursor)
				},
				func(p official.RawOrderPage) any { return len(p.Orders) })
			if err != nil {
				return err
			}
			pages++
			for _, raw := range p.Orders {
				orders++
				var view rawOrderView
				if err := json.Unmarshal(raw, &view); err != nil {
					continue
				}
				status := strings.TrimSpace(view.Status)
				if status == "" {
					status = "(absent)"
				}
				counts[status]++
				if _, seen := shapes[status]; !seen {
					shapes[status] = keySet(raw)
				}
			}
			if !p.HasNext || p.NextCursor == "" || p.NextCursor == cursor {
				break
			}
			cursor = p.NextCursor
		}
	}

	sr.observe("order.history.orders_read", strconv.Itoa(orders),
		fmt.Sprintf("%d page(s) across the CLOSED and OPEN groups", pages))
	if orders == 0 {
		sr.skip("계좌에 주문 이력이 없어 상태 표를 도출할 대상이 아직 없다. " +
			"아래 주문 단계들이 주문을 만드니, 그 뒤에 `--redo read-fixtures`로 이 단계를 다시 실행하라")
		return nil
	}

	observed := sortedKeys(counts)
	sr.observe("order.status.observed", strings.Join(observed, ","), "")
	for _, s := range observed {
		sr.observe("order.status.count."+s, strconv.Itoa(counts[s]), "")
		sr.observe("order.status.shape."+s, strings.Join(shapes[s], ","),
			"field names only — no values from the account are recorded")
	}

	var unobserved []string
	for _, want := range documentedOrderStatuses() {
		if counts[want] == 0 {
			unobserved = append(unobserved, want)
		}
	}
	sr.observe("order.status.documented_unobserved", strings.Join(unobserved, ","),
		"documented values this account has never produced; the state-derivation table is unmeasured for these")

	// The two the attribution rules turn on: does a rejected cancel/replace show
	// up as its own record in the list, and does that record link to the original?
	for _, status := range []string{"CANCEL_REJECTED", "REPLACE_REJECTED"} {
		present := counts[status] > 0
		sr.observe("order.status."+strings.ToLower(status)+".listed", strconv.FormatBool(present), "")
		if present {
			keys := shapes[status]
			sr.observe("order.status."+strings.ToLower(status)+".keys", strings.Join(keys, ","), "")
			sr.observe("order.status."+strings.ToLower(status)+".links_original",
				strconv.FormatBool(hasLineageKey(keys)),
				"whether the separate record carries a field that points back at the original order")
		}
	}
	return nil
}

// stepSellableBaseline snapshots the resting relationship between a holding and
// what the broker says can be sold from it (task 2.8, the "before" half).
func (r *Runner) stepSellableBaseline(ctx context.Context, sr *stepRun) error {
	positions, err := readRetry(ctx, r, sr, EndpointReadHoldings, nil,
		func(ctx context.Context) ([]domain.Position, error) { return r.broker.Holdings(ctx, "") },
		func(p []domain.Position) any { return len(p) })
	if err != nil {
		return err
	}
	sr.observe("sellable.baseline.holdings", strconv.Itoa(len(positions)), "")
	if len(positions) == 0 {
		sr.observe("sellable.baseline.measurable", "false",
			"the account holds nothing, so the reservation semantics cannot be observed on it")
		return nil
	}
	for _, p := range positions {
		sq, err := readRetry(ctx, r, sr, EndpointReadSellable, map[string]string{"symbol": p.Symbol},
			func(ctx context.Context) (domain.SellableQuantity, error) {
				return r.broker.SellableQuantity(ctx, p.Symbol)
			},
			func(s domain.SellableQuantity) any { return s.Quantity })
		if err != nil {
			return err
		}
		sr.observe("sellable.baseline."+p.Symbol+".holding", trim(p.Quantity), "")
		sr.observe("sellable.baseline."+p.Symbol+".sellable", trim(sq.Quantity), "")
		sr.observe("sellable.baseline."+p.Symbol+".reserved", trim(p.Quantity-sq.Quantity),
			"holding minus sellable before this tool placed anything: settlement, collateral and any "+
				"pre-existing sell order are already in here")
	}
	sr.observe("sellable.baseline.measurable", "true", "")
	return nil
}

// --- B. idempotency (task 2.7) ------------------------------------------------

func (r *Runner) stepIdempotency(ctx context.Context, sr *stepRun) error {
	spec, err := r.farBuySpec(ctx, sr, newToken("VERIFY"))
	if err != nil {
		return err
	}

	before, err := r.openOrderIDs(ctx, sr)
	if err != nil {
		return err
	}

	orderID, err := r.placeOrder(ctx, sr, spec, "cancelled at the end of this step")
	if err != nil {
		return err
	}

	// 1. the identity claim: same key, same body.
	replayID, replayErr := r.replayPlaceOrder(ctx, sr, spec,
		"the documented claim is that this returns the first order rather than creating a second")
	switch {
	case isGateError(replayErr):
		return replayErr
	case replayErr != nil:
		sr.observe("idempotency.replay_returns_same_order_id", "false",
			"the replay failed outright: "+truncateError(replayErr))
	case replayID == orderID:
		sr.observe("idempotency.replay_returns_same_order_id", "true", "orderId "+orderID+" came back again")
	default:
		sr.observe("idempotency.replay_returns_same_order_id", "false",
			"the replay returned "+replayID+" where the first call returned "+orderID)
		if replayID != "" {
			sr.created("order", replayID, spec.Symbol, r.now(),
				"created by the idempotent replay; the key was not honoured")
		}
	}

	// 2. the count claim: did the account actually grow?
	after, err := r.openOrderIDs(ctx, sr)
	if err != nil {
		return err
	}
	delta := len(after) - len(before)
	sr.observe("idempotency.open_orders_delta", strconv.Itoa(delta),
		"one is the single order this step placed; two means the replay created a second")
	sr.observe("idempotency.no_second_order_created", strconv.FormatBool(delta <= 1), "")

	// 3. drain before the next probe. The key stays valid for its own window
	// whether or not the order it created is still working, so cancelling here
	// costs the measurement nothing and keeps the live-order count at one.
	if err := r.cancelLiveOrders(ctx, sr, spec.Symbol, "the replay measurement is complete"); err != nil {
		return err
	}

	// 4. the conflict claim: same key, different body.
	conflictSpec := spec
	conflictSpec.Price = OneTickFurther(spec.Price, "buy", spec.Market)
	conflictSpec.Basis = spec.Basis + " (one tick further, to make the body differ)"
	conflictID, conflictErr := r.conflictProbe(ctx, sr, conflictSpec)
	switch {
	case errors.Is(conflictErr, ErrOutsidePlan):
		return conflictErr
	case conflictErr != nil && isRefusal(conflictErr):
		sr.observe("idempotency.conflict_probe", "refused-by-operator", conflictErr.Error())
	case conflictErr != nil:
		code := brokerErrorCode(conflictErr)
		if code == "" {
			code = "(no code in the error envelope)"
		}
		sr.observe("idempotency.conflict_error_code", code, truncateError(conflictErr))
		sr.observe("idempotency.conflict_rejected", "true", "")
	default:
		sr.observe("idempotency.conflict_error_code", "none", "the broker accepted a different body under the same key")
		sr.observe("idempotency.conflict_rejected", "false", "order "+conflictID+" was created and is cancelled below")
	}

	// 5. the scope claim. One set of credentials resolves to one account, so this
	// tool cannot observe a key crossing accounts. Saying so is the honest
	// outcome; guessing would put an unmeasured attribute into an attestation.
	sr.observe("idempotency.key_scope", "unverified",
		"this tool runs against one account, so a key's behaviour across accounts cannot be observed here")

	if maxMS, n := sr.latencyStats(EndpointPlaceOrder); n > 0 {
		sr.observe("idempotency.place_round_trip_ms_max", strconv.FormatInt(maxMS, 10),
			fmt.Sprintf("worst of %d placement round trip(s); the replay safety margin is derived from this", n))
	}

	// 6. leave nothing behind — the conflict probe creates an order whenever the
	// broker does not enforce the key.
	if err := r.cancelLiveOrders(ctx, sr, spec.Symbol, "the idempotency check is complete"); err != nil {
		return err
	}

	if delta > 1 || (replayErr == nil && replayID != orderID) {
		sr.fail("재전송이 첫 주문을 돌려주지 않았다. 멱등 재생은 비활성으로 두고 응답 유실 해소는 " +
			"조회 기반 경로만 써야 한다 (execution-verification: \"재생이 새 주문을 만드는 경우\")")
	}
	return nil
}

// stepIdempotencyTTLEdge probes the validity window. Opt-in: it is the only step
// that deliberately leaves two live orders at once.
func (r *Runner) stepIdempotencyTTLEdge(ctx context.Context, sr *stepRun) error {
	spec, err := r.farBuySpec(ctx, sr, newToken("VERIFY"))
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "    주의: 이 단계는 의도적으로 두 번째 라이브 주문을 만든다. 둘 다 아래에서 취소된다.\n")

	first, err := r.placeOrder(ctx, sr, spec, "cancelled at the end of this step, together with the second order")
	if err != nil {
		return err
	}
	sr.observe("idempotency.ttl_wait_seconds", strconv.FormatInt(int64(r.ttlWait/time.Second), 10),
		"the document says the key is valid for ten minutes")

	fmt.Fprintf(r.out, "    문서상 키 유효 창이 닫히기를 %s 기다린다…\n", r.ttlWait)
	if err := r.sleep(ctx, r.ttlWait); err != nil {
		// Cancel what exists before returning: an interrupted wait must not leave
		// the order behind silently.
		_ = r.cancelLiveOrders(ctx, sr, spec.Symbol, "the validity-window wait was interrupted")
		return err
	}

	second, replayErr := r.replayPlaceOrder(ctx, sr, spec, "the window should have closed by now")
	switch {
	case isGateError(replayErr):
		_ = r.cancelLiveOrders(ctx, sr, spec.Symbol, "the validity-window replay was not authorised")
		return replayErr
	case replayErr != nil:
		sr.observe("idempotency.ttl_window_closed", "unknown", "the replay failed: "+truncateError(replayErr))
	case second == first:
		sr.observe("idempotency.ttl_window_closed", "false",
			fmt.Sprintf("after %s the same key still returned %s", r.ttlWait, first))
	default:
		sr.observe("idempotency.ttl_window_closed", "true",
			fmt.Sprintf("after %s the same key produced a new order %s", r.ttlWait, second))
		sr.created("order", second, spec.Symbol, r.now(), "the second order this step exists to observe")
	}

	return r.cancelLiveOrders(ctx, sr, spec.Symbol, "the validity-window check is complete")
}

// --- C. order path (task 2.2) --------------------------------------------------

func (r *Runner) stepOrderCancel(ctx context.Context, sr *stepRun) error {
	spec, err := r.farBuySpec(ctx, sr, "")
	if err != nil {
		return err
	}
	orderID, err := r.placeOrder(ctx, sr, spec, "cancelled at the end of this step")
	if err != nil {
		return err
	}
	sr.observe("order.place.ok", "true", EndpointPlaceOrder+" accepted a minimum-quantity KR limit order")

	if view, err := r.readOrder(ctx, sr, orderID); err == nil {
		sr.observe("order.status.after_place", orDash(view.Status), "")
		sr.observe("order.time_in_force.after_place", orDash(view.TimeInForce), "")
	}

	if err := r.cancelOrder(ctx, sr, orderID, spec.Symbol, "the cancel path is what this step measures"); err != nil {
		return err
	}
	sr.observe("order.cancel.ok", "true", EndpointCancelOrder+" accepted")

	if view, err := r.readOrder(ctx, sr, orderID); err == nil {
		sr.observe("order.status.after_cancel", orDash(view.Status), "")
		sr.observe("order.canceled_at.present", strconv.FormatBool(strings.TrimSpace(view.CanceledAt) != ""),
			"canceledAt is what separates a cancel from a fill inside one status")
		sr.observe("order.filled_quantity.after_cancel", orDash(view.Execution.FilledQuantity), "")
	}
	return nil
}

func (r *Runner) stepOrderAmend(ctx context.Context, sr *stepRun) error {
	spec, err := r.farBuySpec(ctx, sr, "")
	if err != nil {
		return err
	}
	orderID, err := r.placeOrder(ctx, sr, spec, "cancelled at the end of this step")
	if err != nil {
		return err
	}

	newPrice := OneTickFurther(spec.Price, "buy", spec.Market)
	currentID, err := r.amendOrder(ctx, sr, orderID, spec.Symbol, newPrice, MinQuantity)
	if err != nil {
		return err
	}
	sr.observe("order.amend.ok", "true", EndpointModifyOrder+" accepted a KR price+quantity amend")
	sr.observe("order.amend.issues_new_id", strconv.FormatBool(currentID != orderID),
		"original "+orderID+" -> current "+currentID)

	if currentID != orderID {
		if view, err := r.readOrder(ctx, sr, orderID); err == nil {
			sr.observe("order.amend.original_status", orDash(view.Status),
				"what the superseded identifier reads back as — the input to the attribution rules")
		} else {
			sr.observe("order.amend.original_status", "unreadable", truncateError(err))
		}
	}
	if view, err := r.readOrder(ctx, sr, currentID); err == nil {
		sr.observe("order.amend.current_status", orDash(view.Status), "")
		sr.observe("order.amend.current_price", orDash(view.Price), "")
	}

	return r.cancelLiveOrders(ctx, sr, spec.Symbol, "the amend check is complete")
}

func (r *Runner) stepSellBoundary(ctx context.Context, sr *stepRun) error {
	symbol := r.holdingSymbol
	sellable, err := r.readSellable(ctx, sr, symbol)
	if err != nil {
		return err
	}
	sr.observe("sell.boundary.sellable_at_start", trim(sellable), "")
	if sellable < MinQuantity {
		sr.skip(fmt.Sprintf("%s의 매도가능수량이 %s주다. 매도 경계 측정에는 최소 1주가 필요하다",
			symbol, trim(sellable)))
		return nil
	}

	price, err := r.farSellPrice(ctx, sr, symbol)
	if err != nil {
		return err
	}
	base := orderSpec{Symbol: symbol, Market: MarketOf(symbol), Side: "sell", Price: price.Price, Basis: price.Basis}

	// partial — one share out of a larger holding.
	if sellable > MinQuantity {
		partial := base
		partial.Quantity = MinQuantity
		id, err := r.placeOrder(ctx, sr, partial, "cancelled before the next boundary is probed")
		if err != nil {
			return err
		}
		sr.observe("sell.boundary.partial_accepted", "true", "")
		if reserved, err := r.readSellable(ctx, sr, symbol); err == nil {
			sr.observe("sell.boundary.sellable_with_resting_sell", trim(reserved), "")
			sr.observe("sell.reservation.resting_sell_reserves", strconv.FormatBool(reserved < sellable),
				fmt.Sprintf("%s sellable before, %s while one share rests", trim(sellable), trim(reserved)))
		}
		if err := r.cancelOrder(ctx, sr, id, symbol, "the partial-sell boundary is measured"); err != nil {
			return err
		}
	} else {
		sr.observe("sell.boundary.partial_accepted", "not-applicable",
			"the holding is one share, so a partial sell and a whole-holding sell are the same order")
	}

	// whole holding — bounded, because "sell the entire position" is a real
	// resting order for the entire position.
	if sellable <= r.maxSellQuantity {
		full := base
		full.Quantity = sellable
		id, err := r.placeOrder(ctx, sr, full, "cancelled before the next boundary is probed")
		if err != nil {
			return err
		}
		sr.observe("sell.boundary.full_accepted", "true", trim(sellable)+" share(s)")
		if err := r.cancelOrder(ctx, sr, id, symbol, "the whole-holding boundary is measured"); err != nil {
			return err
		}
	} else {
		sr.observe("sell.boundary.full_accepted", "unverified",
			fmt.Sprintf("the holding is %s share(s), above this tool's exposure cap of %s; a whole-holding "+
				"sell was not placed", trim(sellable), trim(r.maxSellQuantity)))
	}

	// over-holding — the smallest possible oversell, which is what makes it safe
	// to ask the question at all.
	over := base
	over.Quantity = sellable + 1
	overID, overErr := r.placeOrder(ctx, sr, over, "if the broker accepts it, it is cancelled immediately")
	switch {
	case isGateError(overErr):
		return overErr
	case overErr != nil:
		code := brokerErrorCode(overErr)
		if code == "" {
			code = "(no code in the error envelope)"
		}
		sr.observe("sell.boundary.over_holding_rejected", "true", code+": "+truncateError(overErr))
	default:
		sr.observe("sell.boundary.over_holding_rejected", "false",
			"the broker accepted a sell for more than the holding (order "+overID+"); it is cancelled now")
		if err := r.cancelOrder(ctx, sr, overID, symbol, "an accepted oversell must not be left resting"); err != nil {
			return err
		}
		sr.fail("브로커가 보유보다 많은 수량의 매도를 받아들였다 — 청산 예약 공식은 브로커가 " +
			"한도를 걸어 준다고 믿을 수 없다")
	}
	return nil
}

// --- D. conditional orders (task 2.5) -----------------------------------------

func (r *Runner) stepConditionalRegister(ctx context.Context, sr *stepRun) error {
	symbol := r.holdingSymbol
	sellable, err := r.readSellable(ctx, sr, symbol)
	if err != nil {
		return err
	}
	if sellable < MinQuantity {
		sr.skip(fmt.Sprintf("%s의 매도가능수량이 %s주다. 보호 손절에는 최소 1주가 필요하다",
			symbol, trim(sellable)))
		return nil
	}

	last, lower, _, err := r.marketFrame(ctx, sr, symbol)
	if err != nil {
		return err
	}
	trigger, err := FarStopTrigger(last, lower, r.offset, MarketOf(symbol))
	if err != nil {
		return err
	}

	body := official.ConditionalCreateBody{
		Symbol:        symbol,
		Type:          "SINGLE",
		Quantity:      trim(MinQuantity),
		OrderType:     "MARKET",
		ClientOrderID: newToken("VERIFY"),
		ExpireDate:    r.expireDate(),
		First:         official.ConditionLegBody{OrderSide: "SELL", TriggerPrice: trim(trigger.Price)},
	}

	id, err := r.createConditional(ctx, sr, body, trigger.Basis)
	if err != nil {
		return err
	}
	sr.observe("conditional.register.ok", "true", "SINGLE + MARKET + SELL registered against a holding")
	sr.observe("conditional.register.type", "SINGLE", "")
	sr.observe("conditional.register.order_type", "MARKET",
		"OCO and OTO are LIMIT-only per openapi, so a MARKET stop can only be SINGLE")
	sr.observe("conditional.register.market", MarketOf(symbol), "")
	sr.observe("conditional.register.session", r.sessionLabel(),
		"which session the registration was accepted in — a registration is not a trigger")

	// The conditional half of task 2.7: the same key, the same body, once.
	replayID, replayErr := r.replayCreateConditional(ctx, sr, body)
	switch {
	case isGateError(replayErr):
		return replayErr
	case replayErr != nil:
		sr.observe("idempotency.conditional_replay_returns_same_id", "false",
			"the replay failed outright: "+truncateError(replayErr))
	case replayID == id:
		sr.observe("idempotency.conditional_replay_returns_same_id", "true", "conditionalOrderId "+id+" came back again")
	default:
		sr.observe("idempotency.conditional_replay_returns_same_id", "false",
			"the replay returned "+replayID)
		if replayID != "" {
			sr.created("conditional-order", replayID, symbol, r.now(),
				"created by the idempotent replay; the key was not honoured")
			if err := r.cancelConditional(ctx, sr, replayID, symbol,
				"a duplicate conditional created by the replay must not be left registered"); err != nil {
				return err
			}
		}
	}

	// read it back, by id and through the list.
	if co, err := r.readConditional(ctx, sr, id); err == nil {
		sr.observe("conditional.read_by_id.ok", "true", "")
		sr.observe("conditional.status.after_register", orDash(co.Status), "")
		sr.observe("conditional.first_leg_status", orDash(co.First.Status), "")
		sr.observe("conditional.triggered_order_id.at_register",
			orNone(co.First.TriggeredOrderID),
			"null before a trigger is the documented shape")
	} else {
		sr.observe("conditional.read_by_id.ok", "false", truncateError(err))
	}

	// OPEN, not WATCHING: the filter's vocabulary is not the conditional's own
	// status. "OPEN: 진행 중(감시 중·일시중지·주문 진행 중 포함)" and anything outside
	// {OPEN, CLOSED} is 400 invalid-request with allowedValues (measured, M17).
	list, err := readRetry(ctx, r, sr, EndpointReadConditionals,
		map[string]string{"status": ConditionalStatusOpen, "symbol": symbol},
		func(ctx context.Context) (domain.ConditionalOrderList, error) {
			return r.broker.ConditionalOrders(ctx, ConditionalStatusOpen, symbol, "", 50)
		},
		func(l domain.ConditionalOrderList) any { return len(l.Orders) })
	if err != nil {
		sr.observe("conditional.list_by_status.ok", "false", truncateError(err))
	} else {
		sr.observe("conditional.list_by_status.ok", "true", "")
		sr.observe("conditional.list_by_status.contains_new", strconv.FormatBool(containsConditional(list.Orders, id)), "")
	}

	sr.markDeliberate("conditional-order", id,
		"left registered on purpose: the next step proves the broker holds it after this process exits")
	return nil
}

func (r *Runner) stepSellableReserved(ctx context.Context, sr *stepRun) error {
	id, symbol, ok := r.liveConditional()
	if !ok {
		sr.skip("이 검증이 만든 살아 있는 조건주문이 없다")
		return nil
	}
	sellable, err := r.readSellable(ctx, sr, symbol)
	if err != nil {
		return err
	}
	baseline, found := r.baselineSellable(symbol)
	sr.observe("sellable.with_conditional."+symbol, trim(sellable), "conditional "+id+" is registered")
	if !found {
		sr.observe("conditional.reserves_sellable_quantity", "unknown",
			"no baseline was recorded for "+symbol+", so the difference cannot be attributed")
		return nil
	}
	sr.observe("sellable.baseline_recall."+symbol, trim(baseline), "")
	sr.observe("conditional.reserves_sellable_quantity", strconv.FormatBool(sellable < baseline),
		fmt.Sprintf("%s sellable before the conditional, %s with it registered — the missing term in the "+
			"liquidation reservation formula", trim(baseline), trim(sellable)))
	return nil
}

func (r *Runner) stepConditionalPersist(ctx context.Context, sr *stepRun) error {
	id, symbol, ok := r.liveConditional()
	if !ok {
		sr.skip("이 검증이 만든 살아 있는 조건주문이 없다")
		return nil
	}
	registrar, found := r.registeringProcess(id)
	if found && registrar == r.process.InstanceID {
		sr.awaitRestart(fmt.Sprintf(
			"%s에 걸린 조건주문 %s가 등록되어 있고, 등록한 프로세스가 바로 이 프로세스다. "+
				"프로세스 종료를 넘긴 존속은 그 프로세스 안에서는 관측할 수 없다: 여기서 멈추고 "+
				"새 프로세스에서 `tossctl verify run --resume`을 실행해 다시 읽어 마무리하라 "+
				"(콘솔에서는 [콘솔 재시작] 버튼. 남은 단계들이 이 조건주문을 취소한다)", symbol, id))
		fmt.Fprintf(r.out, "\n  계좌에 살아 있는 조건주문이 남아 있다: %s (%s).\n", id, symbol)
		fmt.Fprintf(r.out, "  이어가려면 `tossctl verify run --resume` — 그 실행이 이것을 취소한다.\n")
		fmt.Fprintf(r.out, "  지금 멈추려면 `tossctl order conditional cancel %s`로 직접 취소하라.\n", id)
		return nil
	}

	co, err := r.readConditional(ctx, sr, id)
	if err != nil {
		sr.observe("conditional.survives_process_exit", "false", truncateError(err))
		sr.fail("새 프로세스에서 조건주문 %s를 다시 읽을 수 없었다: %s", id, truncateError(err))
		return nil
	}
	sr.observe("conditional.survives_process_exit", "true",
		fmt.Sprintf("registered by process %s, read back by %s as status %s",
			shortID(registrar), shortID(r.process.InstanceID), orDash(co.Status)))
	sr.observe("conditional.status.after_restart", orDash(co.Status), "")
	sr.markDeliberate("conditional-order", id, "still live; the cancel step below removes it")
	return nil
}

func (r *Runner) stepConditionalTrigger(sr *stepRun) error {
	sr.observe("conditional.trigger_observed", "false", sr.step.Deferred)
	sr.observe("conditional.triggered_order_id_exposed", "unverified",
		"triggeredOrderId links a conditional to the order it generated; it can only be read after a trigger")
	sr.observe("conditional.triggered_order_latency", "unverified", "")
	// The catalogue's own Deferred text already opens with the reason, so the
	// prefix that used to be added here would say it twice.
	sr.deferStep(sr.step.Deferred)
	return nil
}

func (r *Runner) stepConditionalModify(ctx context.Context, sr *stepRun) error {
	id, symbol, ok := r.liveConditional()
	if !ok {
		sr.skip("이 검증이 만든 살아 있는 조건주문이 없다")
		return nil
	}
	before, err := r.readConditional(ctx, sr, id)
	if err != nil {
		return err
	}
	newTrigger := OneTickFurther(before.First.TriggerPrice, "buy", MarketOf(symbol))
	if newTrigger <= 0 {
		sr.skip("살아 있는 조건주문에 옮길 발동가가 없다")
		return nil
	}

	body := official.ConditionalModifyBody{
		Type:       "SINGLE",
		Quantity:   trim(MinQuantity),
		OrderType:  "MARKET",
		ExpireDate: r.expireDate(),
		First:      official.ConditionLegBody{OrderSide: "SELL", TriggerPrice: trim(newTrigger)},
	}
	newID, err := r.modifyConditional(ctx, sr, id, symbol, body, "one tick further below the market")
	if err != nil {
		return err
	}
	sr.observe("conditional.modify.ok", "true", EndpointModifyConditional+" accepted")
	sr.observe("conditional.modify_issues_new_id", strconv.FormatBool(newID != id), id+" -> "+newID)

	if newID != id {
		if _, err := r.readConditional(ctx, sr, id); err != nil {
			sr.observe("conditional.modify_invalidates_old_id", "true",
				"the old identifier no longer reads back: "+truncateError(err))
		} else {
			sr.observe("conditional.modify_invalidates_old_id", "false",
				"the old identifier still reads back after the modify — two live claims on the same shares")
			sr.fail("정정이 기존 조건주문 %s를 새 %s와 나란히 읽히는 상태로 남겼다. 이 동작이 이해되기 "+
				"전까지 보호 수량 변경은 취소 후 재등록으로만 해야 한다", id, newID)
		}
	} else {
		sr.observe("conditional.modify_invalidates_old_id", "not-applicable", "the identifier did not change")
	}

	if co, err := r.readConditional(ctx, sr, newID); err == nil {
		sr.observe("conditional.status.after_modify", orDash(co.Status), "")
		sr.observe("conditional.trigger_after_modify", trim(co.First.TriggerPrice), "")
	}
	sr.markDeliberate("conditional-order", newID, "still live; the cancel step below removes it")
	return nil
}

func (r *Runner) stepConditionalCancel(ctx context.Context, sr *stepRun) error {
	id, symbol, ok := r.liveConditional()
	if !ok {
		sr.skip("이 검증이 만든 살아 있는 조건주문이 없다")
		return nil
	}
	if err := r.cancelConditional(ctx, sr, id, symbol, "the verification is finished with it"); err != nil {
		return err
	}
	sr.observe("conditional.cancel.ok", "true", EndpointCancelConditional+" accepted")

	if co, err := r.readConditional(ctx, sr, id); err != nil {
		sr.observe("conditional.cancel.gone_after", "true", "the identifier no longer reads back")
	} else {
		sr.observe("conditional.cancel.gone_after", "false", "status is still "+orDash(co.Status))
	}
	return nil
}

// --- E. costs (task 2.9's input) ----------------------------------------------

func (r *Runner) stepCosts(ctx context.Context, sr *stepRun) error {
	ids := r.createdOrderIDs()
	sr.observe("costs.orders_examined", strconv.Itoa(len(ids)), "")
	if len(ids) == 0 {
		sr.observe("costs.collected", "false", "this verification created no order to read costs from")
		return nil
	}

	var (
		filled     int
		commission float64
		tax        float64
	)
	for _, id := range ids {
		view, err := r.readOrder(ctx, sr, id)
		if err != nil {
			sr.observe("costs.order."+id+".readable", "false", truncateError(err))
			continue
		}
		// Doubling as fixture collection: these are statuses this account has now
		// certainly produced, which is what task 2.1 is after.
		sr.observe("order.status.verification_order."+id, orDash(view.Status), "")
		q := parseDecimal(view.Execution.FilledQuantity)
		if q <= 0 {
			continue
		}
		filled++
		commission += parseDecimal(view.Execution.Commission)
		tax += parseDecimal(view.Execution.Tax)
		sr.observe("costs.order."+id+".filled_quantity", trim(q), "")
		sr.observe("costs.order."+id+".commission", trim(parseDecimal(view.Execution.Commission)), "")
		sr.observe("costs.order."+id+".tax", trim(parseDecimal(view.Execution.Tax)), "")
		sr.observe("costs.order."+id+".filled_amount", trim(parseDecimal(view.Execution.FilledAmount)), "")
	}

	sr.observe("costs.orders_filled", strconv.Itoa(filled), "")
	if filled == 0 {
		sr.observe("costs.collected", "false",
			"no verification order filled — the far-from-market limits are designed not to. The Toss cost "+
				"model has no measured input from this run; task 2.9 still needs one")
		return nil
	}
	sr.observe("costs.collected", "true", "")
	sr.observe("costs.commission_total", trim(commission), "")
	sr.observe("costs.tax_total", trim(tax), "")
	return nil
}

// --- shared reads -------------------------------------------------------------

// farBuySpec prices the standard probe order: one share, limit, far below.
func (r *Runner) farBuySpec(ctx context.Context, sr *stepRun, key string) (orderSpec, error) {
	symbol := r.symbol
	last, lower, _, err := r.marketFrame(ctx, sr, symbol)
	if err != nil {
		return orderSpec{}, err
	}
	price, err := FarBuyLimit(last, lower, r.offset, MarketOf(symbol))
	if err != nil {
		return orderSpec{}, err
	}
	return orderSpec{
		Symbol:   symbol,
		Market:   MarketOf(symbol),
		Side:     "buy",
		Quantity: MinQuantity,
		Price:    price.Price,
		Key:      key,
		Basis:    price.Basis,
	}, nil
}

func (r *Runner) farSellPrice(ctx context.Context, sr *stepRun, symbol string) (SafePrice, error) {
	last, _, upper, err := r.marketFrame(ctx, sr, symbol)
	if err != nil {
		return SafePrice{}, err
	}
	return FarSellLimit(last, upper, r.offset, MarketOf(symbol))
}

// marketFrame reads the last trade and the day's band.
func (r *Runner) marketFrame(ctx context.Context, sr *stepRun, symbol string) (last, lower, upper float64, err error) {
	quotes, err := readRetry(ctx, r, sr, EndpointReadPrices, map[string]string{"symbol": symbol},
		func(ctx context.Context) ([]domain.Quote, error) { return r.broker.Prices(ctx, []string{symbol}) },
		func(q []domain.Quote) any { return len(q) })
	if err != nil {
		return 0, 0, 0, err
	}
	for _, q := range quotes {
		if strings.EqualFold(q.Symbol, symbol) && q.Last > 0 {
			last = q.Last
			break
		}
	}
	if last == 0 && len(quotes) > 0 {
		last = quotes[0].Last
	}
	if last <= 0 {
		return 0, 0, 0, fmt.Errorf("verify: no last price for %s, so no order can be priced safely", symbol)
	}

	limits, err := readRetry(ctx, r, sr, EndpointReadPriceLimits, map[string]string{"symbol": symbol},
		func(ctx context.Context) (domain.PriceLimits, error) { return r.broker.PriceLimits(ctx, symbol) },
		func(l domain.PriceLimits) any { return l })
	if err != nil {
		// A missing band is not fatal — US has none — but it does mean the clamp
		// cannot run, so it is recorded rather than swallowed.
		sr.observe("pricing."+symbol+".price_limits_readable", "false", truncateError(err))
		return last, 0, 0, nil
	}
	return last, limits.LowerLimit, limits.UpperLimit, nil
}

func (r *Runner) readSellable(ctx context.Context, sr *stepRun, symbol string) (float64, error) {
	sq, err := readRetry(ctx, r, sr, EndpointReadSellable, map[string]string{"symbol": symbol},
		func(ctx context.Context) (domain.SellableQuantity, error) {
			return r.broker.SellableQuantity(ctx, symbol)
		},
		func(s domain.SellableQuantity) any { return s.Quantity })
	if err != nil {
		return 0, err
	}
	return sq.Quantity, nil
}

func (r *Runner) readOrder(ctx context.Context, sr *stepRun, id string) (rawOrderView, error) {
	raw, err := readRetry(ctx, r, sr, EndpointReadOrderByID, map[string]string{"orderId": id},
		func(ctx context.Context) (json.RawMessage, error) { return r.broker.OrderRawByID(ctx, id) },
		func(b json.RawMessage) any { return DigestBytes(b) })
	if err != nil {
		return rawOrderView{}, err
	}
	var view rawOrderView
	if err := json.Unmarshal(raw, &view); err != nil {
		return rawOrderView{}, fmt.Errorf("verify: order %s did not decode: %w", id, err)
	}
	return view, nil
}

func (r *Runner) readConditional(ctx context.Context, sr *stepRun, id string) (conditionalView, error) {
	co, err := readRetry(ctx, r, sr, EndpointReadConditionalByID,
		map[string]string{"conditionalOrderId": id},
		func(ctx context.Context) (domain.ConditionalOrder, error) { return r.broker.ConditionalOrder(ctx, id) },
		func(c domain.ConditionalOrder) any { return c.Status })
	if err != nil {
		return conditionalView{}, err
	}
	return conditionalView{
		Status: co.Status,
		First: conditionalLegView{
			Status:           co.First.Status,
			TriggerPrice:     co.First.TriggerPrice,
			TriggeredOrderID: co.First.TriggeredOrderID,
		},
	}, nil
}

// openOrderIDs is the count the idempotency step compares before and after.
func (r *Runner) openOrderIDs(ctx context.Context, sr *stepRun) (map[string]bool, error) {
	ids := map[string]bool{}
	cursor := ""
	for page := 0; page < maxFixturePages; page++ {
		p, err := readRetry(ctx, r, sr, EndpointReadOrders,
			map[string]string{"status": statusOpen, "cursor": cursor},
			func(ctx context.Context) (official.RawOrderPage, error) {
				return r.broker.OrdersPageRaw(ctx, official.OrdersFilter{Status: statusOpen}, cursor)
			},
			func(page official.RawOrderPage) any { return len(page.Orders) })
		if err != nil {
			return nil, err
		}
		for _, raw := range p.Orders {
			var view rawOrderView
			if err := json.Unmarshal(raw, &view); err == nil && view.OrderID != "" {
				ids[view.OrderID] = true
			}
		}
		if !p.HasNext || p.NextCursor == "" || p.NextCursor == cursor {
			break
		}
		cursor = p.NextCursor
	}
	return ids, nil
}

// cancelLiveOrders cancels everything this tool still has resting.
//
// A step that placed more than one order — the conflict probe that was not
// refused, the validity-window step's second order — ends by draining them all,
// so the exposure cap is honest at the start of the next step.
func (r *Runner) cancelLiveOrders(ctx context.Context, sr *stepRun, symbol, why string) error {
	for {
		live := ""
		entries := append(r.allEntries(), Entry{StepID: sr.step.ID, Artifacts: sr.artifacts})
		for _, a := range Outstanding(entries) {
			if a.Kind == "order" {
				live = a.ID
				break
			}
		}
		if live == "" {
			return nil
		}
		if err := r.cancelOrder(ctx, sr, live, symbol, why); err != nil {
			return err
		}
	}
}

// --- record queries -----------------------------------------------------------

// liveConditional finds the conditional order this verification still has
// registered, across the record and this run.
func (r *Runner) liveConditional() (id, symbol string, ok bool) {
	entries := r.allEntries()
	for _, a := range Outstanding(entries) {
		if a.Kind == "conditional-order" {
			return a.ID, a.Symbol, true
		}
	}
	return "", "", false
}

// registeringProcess reports which process created a conditional order. It is
// what makes conditional-persist enforceable rather than a promise.
func (r *Runner) registeringProcess(id string) (string, bool) {
	for _, e := range r.allEntries() {
		for _, a := range e.Artifacts {
			if a.Kind == "conditional-order" && a.ID == id && !a.Cancelled {
				return e.Process.InstanceID, true
			}
		}
	}
	return "", false
}

// createdOrderIDs lists every order this verification placed, cancelled or not.
func (r *Runner) createdOrderIDs() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range r.allEntries() {
		for _, a := range e.Artifacts {
			if a.Kind != "order" || a.CreatedAt.IsZero() || seen[a.ID] {
				continue
			}
			seen[a.ID] = true
			out = append(out, a.ID)
		}
	}
	return out
}

// baselineSellable recalls what the baseline step measured for a symbol.
func (r *Runner) baselineSellable(symbol string) (float64, bool) {
	key := "sellable.baseline." + symbol + ".sellable"
	for _, e := range r.allEntries() {
		if e.StepID != StepSellableBaseline {
			continue
		}
		for _, o := range e.Observations {
			if o.Key == key {
				return parseDecimal(o.Value), true
			}
		}
	}
	return 0, false
}

// --- small helpers ------------------------------------------------------------

// rawOrderView is the subset of the Order schema the steps read. It decodes the
// raw JSON rather than domain.Order because domain.Order has nowhere to put
// canceledAt or execution.commission, which are two of the things being measured.
type rawOrderView struct {
	OrderID     string `json:"orderId"`
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	OrderType   string `json:"orderType"`
	TimeInForce string `json:"timeInForce"`
	Status      string `json:"status"`
	Quantity    string `json:"quantity"`
	Price       string `json:"price"`
	Currency    string `json:"currency"`
	OrderedAt   string `json:"orderedAt"`
	CanceledAt  string `json:"canceledAt"`
	Execution   struct {
		FilledQuantity     string `json:"filledQuantity"`
		AverageFilledPrice string `json:"averageFilledPrice"`
		FilledAmount       string `json:"filledAmount"`
		Commission         string `json:"commission"`
		Tax                string `json:"tax"`
		FilledAt           string `json:"filledAt"`
		SettlementDate     string `json:"settlementDate"`
	} `json:"execution"`
}

type conditionalLegView struct {
	Status           string
	TriggerPrice     float64
	TriggeredOrderID string
}

type conditionalView struct {
	Status string
	First  conditionalLegView
}

// expireDate is a week out, in KST, which is the timezone the conditional order's
// expireDate is documented in ("조건주문 만료일 (YYYY-MM-DD, 필수)", and createdAt
// is returned as KST).
//
// A fixed zone rather than time.LoadLocation: a machine without tzdata would
// otherwise fall back to UTC silently, and near midnight that is the wrong day.
func (r *Runner) expireDate() string {
	kst := time.FixedZone("KST", 9*60*60)
	return r.now().In(kst).AddDate(0, 0, 7).Format("2006-01-02")
}

// sessionLabel lives in hours.go, next to the advisory that shares its window.

func keySet(raw json.RawMessage) []string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return sortedKeysOf(m)
}

// hasLineageKey reports whether a record carries a field that could point back at
// an original order. The names are guesses on purpose — the point of the fixture
// is to find out what the field is actually called.
func hasLineageKey(keys []string) bool {
	for _, k := range keys {
		lower := strings.ToLower(k)
		for _, hint := range []string{"original", "parent", "source", "related", "linked", "root"} {
			if strings.Contains(lower, hint) {
				return true
			}
		}
	}
	return false
}

func containsConditional(orders []domain.ConditionalOrder, id string) bool {
	for _, o := range orders {
		if o.ID == id {
			return true
		}
	}
	return false
}

// isRefusal reports a confirmation the operator declined, as opposed to a broker
// answer. The two must not be conflated: a refusal is a decision to record and a
// broker rejection is a measurement.
func isRefusal(err error) bool {
	return errors.Is(err, ErrRefused) ||
		errors.Is(err, ErrConfirmationExpired) ||
		errors.Is(err, ErrNotATerminal)
}

// isGateError reports an error that came from the human gate rather than from the
// broker: a declined confirmation, an expired one, no terminal to type at, or a
// request the approved batch does not cover.
//
// Every step that interprets a mutation's error has to check this first. Nothing
// was sent in any of these cases, so recording one as "the broker refused" would
// put an observation into the evidence that no broker ever produced — and the
// evidence is the entire point of the exercise.
func isGateError(err error) bool {
	return isRefusal(err) || errors.Is(err, ErrOutsidePlan)
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func parseDecimal(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func shortID(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12] + "…"
}

func maskedAccount(ref string) string { return attest.Mask(ref) }
