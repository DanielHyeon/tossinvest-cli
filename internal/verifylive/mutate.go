package verifylive

// mutate.go is the only file in this package that calls a mutating Broker
// method.
//
// That is a structural claim, not a convention: static_test.go asserts that
// PlaceOrder, CancelOrder, ModifyOrder, CreateConditionalOrder,
// ModifyConditionalOrderRef and CancelConditionalOrder appear nowhere else in the
// package outside a comment. A step that wanted to place an order without a
// confirmation, an exposure check and an evidence line would have to add its call
// here, next to the ones that do all three, which is the review a change like
// that deserves.
//
// Each function below does the same four things in the same order:
//
//	1. check the exposure cap    at most one live order from this tool at a time
//	2. ask the human             a typed, expiring nonce (confirm.go)
//	3. send, and time it         the latency feeds the idempotency safety margin
//	4. record                    a Call with digests, and an Artifact if something
//	                             live now exists
//
// The one exception is replayPlaceOrder, and it is commented where it is.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// Endpoints, spelled as internal/soak and internal/app/engine spell theirs.
const (
	EndpointPlaceOrder          = "POST /api/v1/orders"
	EndpointCancelOrder         = "POST /api/v1/orders/{id}/cancel"
	EndpointModifyOrder         = "POST /api/v1/orders/{id}/modify"
	EndpointCreateConditional   = "POST /api/v1/conditional-orders"
	EndpointModifyConditional   = "POST /api/v1/conditional-orders/{id}/modify"
	EndpointCancelConditional   = "DELETE /api/v1/conditional-orders/{id}"
	EndpointReadOrders          = "GET /api/v1/orders"
	EndpointReadOrderByID       = "GET /api/v1/orders/{id}"
	EndpointReadHoldings        = "GET /api/v1/holdings"
	EndpointReadSellable        = "GET /api/v1/sellable-quantity"
	EndpointReadPrices          = "GET /api/v1/prices"
	EndpointReadPriceLimits     = "GET /api/v1/price-limits"
	EndpointReadConditionals    = "GET /api/v1/conditional-orders"
	EndpointReadConditionalByID = "GET /api/v1/conditional-orders/{id}"
)

// MaxLiveOrders is the absolute cap on orders this tool has resting at once.
//
// One. The measurement never needs two, and a cap of one means a crash between
// a place and its cancel leaves exactly one thing to clean up, whose identifier
// is in the record. StepIdempotencyTTLEdge is the single documented exception:
// its whole purpose is to observe a second order being created, and it says so in
// its own step text before anything is sent.
const MaxLiveOrders = 1

// MaxLiveOrdersTTLEdge is the cap while the opt-in validity-window step runs.
const MaxLiveOrdersTTLEdge = 2

// MaxLiveConditionals is the cap on conditional orders. One, for the same reason,
// and the persistence step is why it is a separate budget: the conditional is
// deliberately alive while no order is.
const MaxLiveConditionals = 1

// ErrExposureCap means a step asked for more live exposure than the tool allows.
var ErrExposureCap = errors.New("verify: the live-exposure cap would be exceeded")

// orderSpec is one order this tool wants to place.
type orderSpec struct {
	Symbol   string
	Market   string
	Side     string // "buy" | "sell"
	Quantity float64
	Price    float64
	// Key is the clientOrderId. Empty means the request carries no idempotency
	// key at all, which is the documented "멱등성 미적용" and is what the steps
	// that are not measuring idempotency send.
	Key string
	// Basis explains the price to the operator and to the record.
	Basis string
}

func (s orderSpec) intent() orderintent.PlaceIntent {
	return orderintent.PlaceIntent{
		Symbol:        s.Symbol,
		Market:        s.Market,
		Side:          strings.ToLower(s.Side),
		OrderType:     "limit",
		Quantity:      s.Quantity,
		Price:         s.Price,
		ClientOrderID: s.Key,
	}
}

func (s orderSpec) detail() []string {
	lines := []string{
		fmt.Sprintf("symbol           %s (%s)", s.Symbol, s.Market),
		fmt.Sprintf("side             %s, LIMIT", strings.ToUpper(s.Side)),
		fmt.Sprintf("quantity         %s share(s)", trim(s.Quantity)),
		fmt.Sprintf("limit price      %s", trim(s.Price)),
		fmt.Sprintf("notional         %s", trim(s.Quantity*s.Price)),
		fmt.Sprintf("pricing          %s", s.Basis),
	}
	if s.Key != "" {
		lines = append(lines, fmt.Sprintf("clientOrderId    %s", s.Key))
	}
	return lines
}

// placeOrder is the confirmed order placement.
func (r *Runner) placeOrder(ctx context.Context, sr *stepRun, spec orderSpec, reversal string) (string, error) {
	if err := r.checkOrderCap(sr); err != nil {
		return "", err
	}
	m := NewMutation(sr.step.ID, "place a live LIMIT order", r.accountRef, spec.detail(), reversal, r.now())
	if err := r.confirm(m); err != nil {
		return "", err
	}

	intent := spec.intent()
	started := r.now()
	res, err := r.broker.PlaceOrder(ctx, intent)
	sr.logCall(EndpointPlaceOrder, intent, started, r.now(), res, err)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(res.OrderID) == "" {
		return "", fmt.Errorf("verify: the broker accepted the order but returned no orderId")
	}
	sr.created("order", res.OrderID, spec.Symbol, r.now(), "")
	fmt.Fprintf(r.out, "    placed order %s (%s %s x%s @ %s)\n",
		res.OrderID, strings.ToUpper(spec.Side), spec.Symbol, trim(spec.Quantity), trim(spec.Price))
	return res.OrderID, nil
}

// replayPlaceOrder re-sends a body that has already been confirmed, under the
// same idempotency key.
//
// It is the one mutating call in this package that is not separately confirmed,
// and the reason is that confirming it would destroy the measurement. The claim
// under test is "동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다"
// (OrderCreateRequest.clientOrderId, docs/migration/openapi.latest.json): a
// request that is byte-identical to one the operator already approved, sent under
// the key they already saw, either returns the same order or exposes the
// documentation as wrong. A second prompt would be asking permission for the
// same order twice.
//
// Three things keep it honest. The body must be identical to the one that was
// confirmed — the caller passes the same orderSpec, and a mismatch is the
// conflict probe, which does not come through here. It is printed before it is
// sent, so nothing is hidden. And whatever comes back is registered as an
// artifact, so if the broker does create a second order it is cancelled like any
// other.
func (r *Runner) replayPlaceOrder(ctx context.Context, sr *stepRun, spec orderSpec, why string) (string, error) {
	fmt.Fprintf(r.out, "    replaying the identical body under key %s — %s\n", spec.Key, why)
	intent := spec.intent()
	started := r.now()
	res, err := r.broker.PlaceOrder(ctx, intent)
	sr.logCall(EndpointPlaceOrder, intent, started, r.now(), res, err)
	if err != nil {
		return "", err
	}
	return res.OrderID, nil
}

// conflictProbe sends a *different* body under a key that has already been used.
//
// It is confirmed, because it is a different order: if the broker does not
// enforce the key it will create one.
func (r *Runner) conflictProbe(ctx context.Context, sr *stepRun, spec orderSpec) (string, error) {
	if err := r.checkOrderCap(sr); err != nil {
		return "", err
	}
	detail := append(spec.detail(),
		"NOTE             this deliberately re-uses the key above with a DIFFERENT price",
		"expected         the broker refuses with idempotency-key-conflict and creates nothing")
	m := NewMutation(sr.step.ID, "probe the idempotency key with a conflicting body", r.accountRef, detail,
		"if the broker creates an order instead of refusing, it is cancelled immediately", r.now())
	if err := r.confirm(m); err != nil {
		return "", err
	}

	intent := spec.intent()
	started := r.now()
	res, err := r.broker.PlaceOrder(ctx, intent)
	sr.logCall(EndpointPlaceOrder, intent, started, r.now(), res, err)
	if err != nil {
		return "", err
	}
	if id := strings.TrimSpace(res.OrderID); id != "" {
		sr.created("order", id, spec.Symbol, r.now(), "created by the conflicting-body probe; the key was not enforced")
	}
	return res.OrderID, nil
}

// cancelOrder is the confirmed cancel.
func (r *Runner) cancelOrder(ctx context.Context, sr *stepRun, orderID, symbol, why string) error {
	detail := []string{
		fmt.Sprintf("order            %s", orderID),
		fmt.Sprintf("symbol           %s", symbol),
		fmt.Sprintf("why              %s", why),
		"direction        this REDUCES exposure — it removes a resting order",
	}
	m := NewMutation(sr.step.ID, "cancel a live order", r.accountRef, detail,
		"the account returns to how this step found it", r.now())
	if err := r.confirm(m); err != nil {
		return err
	}

	started := r.now()
	res, err := r.broker.CancelOrder(ctx, orderID)
	sr.logCall(EndpointCancelOrder, map[string]string{"orderId": orderID}, started, r.now(), res, err)
	if err != nil {
		return err
	}
	sr.cancelled("order", orderID, symbol, r.now(), "")
	fmt.Fprintf(r.out, "    cancelled order %s\n", orderID)
	if id := strings.TrimSpace(res.CurrentOrderID); id != "" && id != orderID {
		sr.observe("order.cancel.response_order_id_differs", "true",
			"the cancel returned "+id+" for original "+orderID)
	}
	return nil
}

// amendOrder is the confirmed amend.
func (r *Runner) amendOrder(ctx context.Context, sr *stepRun, orderID, symbol string, price, quantity float64) (string, error) {
	detail := []string{
		fmt.Sprintf("order            %s", orderID),
		fmt.Sprintf("symbol           %s", symbol),
		fmt.Sprintf("new limit price  %s (further from the market, still un-fillable)", trim(price)),
		fmt.Sprintf("new quantity     %s share(s) — KR requires quantity on a modify", trim(quantity)),
	}
	m := NewMutation(sr.step.ID, "amend a live order", r.accountRef, detail,
		"whichever identifier is live afterwards is cancelled in this same step", r.now())
	if err := r.confirm(m); err != nil {
		return "", err
	}

	intent := orderintent.AmendIntent{OrderID: orderID, Price: &price, Quantity: &quantity}
	started := r.now()
	res, err := r.broker.ModifyOrder(ctx, intent)
	sr.logCall(EndpointModifyOrder, intent, started, r.now(), res, err)
	if err != nil {
		return "", err
	}
	current := strings.TrimSpace(res.CurrentOrderID)
	if current == "" {
		current = orderID
	}
	if current != orderID {
		// Superseded first, then created. The broker replaced one order with
		// another in a single call; recording the creation first would make a
		// sequential read of the record show two live orders at once, which is
		// the invariant the exposure cap exists to keep true.
		sr.cancelled("order", orderID, symbol, r.now(), "replaced by the amend; superseded by "+current)
		sr.created("order", current, symbol, r.now(), "issued by the amend, replacing "+orderID)
	}
	fmt.Fprintf(r.out, "    amended order %s -> %s\n", orderID, current)
	return current, nil
}

// createConditional is the confirmed conditional registration.
func (r *Runner) createConditional(ctx context.Context, sr *stepRun, body official.ConditionalCreateBody, basis string) (string, error) {
	if err := r.checkConditionalCap(sr); err != nil {
		return "", err
	}
	detail := conditionalDetail(body, basis)
	m := NewMutation(sr.step.ID, "register a live conditional order", r.accountRef, detail,
		"cancelled by the conditional-cancel step; it deliberately outlives this process first", r.now())
	if err := r.confirm(m); err != nil {
		return "", err
	}

	started := r.now()
	ref, err := r.broker.CreateConditionalOrder(ctx, body)
	sr.logCall(EndpointCreateConditional, body, started, r.now(), ref, err)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ref.ID) == "" {
		return "", fmt.Errorf("verify: the broker accepted the conditional order but returned no conditionalOrderId")
	}
	sr.created("conditional-order", ref.ID, body.Symbol, r.now(), "")
	fmt.Fprintf(r.out, "    registered conditional %s (%s %s trigger %s)\n",
		ref.ID, body.Type, body.Symbol, body.First.TriggerPrice)
	return ref.ID, nil
}

// replayCreateConditional is createConditional's unconfirmed replay, for the same
// reason and under the same three protections as replayPlaceOrder: identical
// body, printed before it is sent, and anything it creates is an artifact.
func (r *Runner) replayCreateConditional(ctx context.Context, sr *stepRun, body official.ConditionalCreateBody) (string, error) {
	fmt.Fprintf(r.out, "    replaying the identical conditional body under key %s\n", body.ClientOrderID)
	started := r.now()
	ref, err := r.broker.CreateConditionalOrder(ctx, body)
	sr.logCall(EndpointCreateConditional, body, started, r.now(), ref, err)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

// modifyConditional is the confirmed conditional modify.
func (r *Runner) modifyConditional(ctx context.Context, sr *stepRun, id, symbol string, body official.ConditionalModifyBody, basis string) (string, error) {
	detail := []string{
		fmt.Sprintf("conditional      %s", id),
		fmt.Sprintf("symbol           %s", symbol),
		fmt.Sprintf("type             %s / %s", body.Type, body.OrderType),
		fmt.Sprintf("new trigger      %s (%s)", body.First.TriggerPrice, basis),
		fmt.Sprintf("quantity         %s share(s)", body.Quantity),
		"NOTE             openapi says a modify cancels and recreates: a NEW id is issued and the old one is invalidated",
	}
	m := NewMutation(sr.step.ID, "modify a live conditional order", r.accountRef, detail,
		"whichever identifier is live afterwards is cancelled by the conditional-cancel step", r.now())
	if err := r.confirm(m); err != nil {
		return "", err
	}

	started := r.now()
	ref, err := r.broker.ModifyConditionalOrderRef(ctx, id, body)
	sr.logCall(EndpointModifyConditional, body, started, r.now(), ref, err)
	if err != nil {
		return "", err
	}
	current := strings.TrimSpace(ref.ID)
	if current == "" || current == id {
		// The broker kept the identifier. Nothing new exists; the old artifact
		// stays live and the observation records what happened.
		return id, nil
	}
	// Superseded first, then created — see the note in amendOrder.
	sr.cancelled("conditional-order", id, symbol, r.now(), "invalidated by the modify; superseded by "+current)
	sr.created("conditional-order", current, symbol, r.now(), "issued by the modify, replacing "+id)
	fmt.Fprintf(r.out, "    modified conditional %s -> %s\n", id, current)
	return current, nil
}

// cancelConditional is the confirmed conditional cancel.
func (r *Runner) cancelConditional(ctx context.Context, sr *stepRun, id, symbol, why string) error {
	detail := []string{
		fmt.Sprintf("conditional      %s", id),
		fmt.Sprintf("symbol           %s", symbol),
		fmt.Sprintf("why              %s", why),
		"direction        this REMOVES a protective order from the account",
	}
	m := NewMutation(sr.step.ID, "cancel a live conditional order", r.accountRef, detail,
		"the account returns to how this verification found it", r.now())
	if err := r.confirm(m); err != nil {
		return err
	}

	started := r.now()
	err := r.broker.CancelConditionalOrder(ctx, id)
	sr.logCall(EndpointCancelConditional, map[string]string{"conditionalOrderId": id}, started, r.now(), nil, err)
	if err != nil {
		return err
	}
	sr.cancelled("conditional-order", id, symbol, r.now(), "")
	fmt.Fprintf(r.out, "    cancelled conditional %s\n", id)
	return nil
}

// --- exposure ---------------------------------------------------------------

// checkOrderCap refuses a placement that would breach MaxLiveOrders.
func (r *Runner) checkOrderCap(sr *stepRun) error {
	limit := MaxLiveOrders
	if sr.step.ID == StepIdempotencyTTLEdge {
		limit = MaxLiveOrdersTTLEdge
	}
	if n := r.liveCount("order", sr); n >= limit {
		return fmt.Errorf("%w: %d live order(s) from this tool already, and the cap is %d — "+
			"cancel them before continuing (`tossctl verify status` lists them)", ErrExposureCap, n, limit)
	}
	return nil
}

func (r *Runner) checkConditionalCap(sr *stepRun) error {
	if n := r.liveCount("conditional-order", sr); n >= MaxLiveConditionals {
		return fmt.Errorf("%w: %d live conditional order(s) from this tool already, and the cap is %d",
			ErrExposureCap, n, MaxLiveConditionals)
	}
	return nil
}

// liveCount counts outstanding artifacts of a kind across the record and the
// step in progress. The step in progress matters: within one step the tool can
// place and cancel more than once, and only the ones still live count.
func (r *Runner) liveCount(kind string, sr *stepRun) int {
	entries := append([]Entry(nil), r.prior...)
	entries = append(entries, r.written...)
	if sr != nil {
		entries = append(entries, Entry{StepID: sr.step.ID, Artifacts: sr.artifacts})
	}
	n := 0
	for _, a := range Outstanding(entries) {
		if a.Kind == kind {
			n++
		}
	}
	return n
}

// --- error decoding -----------------------------------------------------------

// brokerErrorCode digs the broker's own error code out of a failure.
//
// The official client hands 4xx back as *official.APIError with the body
// attached, and the body is the documented ErrorResponse envelope. The code is
// the thing task 2.7 is looking for by name (`idempotency-key-conflict`), so it
// is lifted onto its own field rather than left inside a message somebody has to
// eyeball.
func brokerErrorCode(err error) string {
	var apiErr *official.APIError
	if !errors.As(err, &apiErr) {
		return ""
	}
	var env struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal([]byte(apiErr.Body), &env); jsonErr != nil {
		return ""
	}
	return strings.TrimSpace(env.Error.Code)
}

func conditionalDetail(body official.ConditionalCreateBody, basis string) []string {
	lines := []string{
		fmt.Sprintf("symbol           %s", body.Symbol),
		fmt.Sprintf("type             %s / %s", body.Type, body.OrderType),
		fmt.Sprintf("watch            %s when the price reaches %s", body.First.OrderSide, body.First.TriggerPrice),
		fmt.Sprintf("quantity         %s share(s)", body.Quantity),
		fmt.Sprintf("expires          %s", body.ExpireDate),
		fmt.Sprintf("pricing          %s", basis),
	}
	if body.ClientOrderID != "" {
		lines = append(lines, fmt.Sprintf("clientOrderId    %s", body.ClientOrderID))
	}
	return lines
}

// truncateError bounds what a failure can put into the record. The official
// client's errors carry the broker's error envelope, which is a code, a message
// and a request id — no credentials — but an unbounded string in an append-only
// file is still an unbounded string.
func truncateError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	const max = 400
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

// sleepFor waits, cancellably. The validity-window step is the only caller and it
// waits for minutes, so a Ctrl-C during the wait has to come back promptly with
// the order still recorded.
func sleepFor(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
