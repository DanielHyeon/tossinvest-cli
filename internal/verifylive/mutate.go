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
//	2. clear the human gate      the approved batch must carry a line for exactly
//	                             this request (gate → authorise), or, under
//	                             --confirm-each, a typed expiring nonce for it
//	3. send, and time it         the latency feeds the idempotency safety margin
//	4. record                    a Call with digests, and an Artifact if something
//	                             live now exists
//
// The two exceptions are replayPlaceOrder and replayCreateConditional, which are
// commented where they are. They still pass step 2's authorisation — a replay the
// approved list does not carry is a request nobody agreed to, whatever its body.

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

// ErrOutsidePlan means a step tried to send something the approved batch does not
// carry a line for.
//
// It is not a refusal and it is not a measurement: it means the run would have had
// to do something other than what a person read and approved. Nothing is sent and
// the whole run stops, because the alternative — adapting to the new conditions and
// carrying on — is precisely the behaviour a batch approval must not permit.
var ErrOutsidePlan = errors.New("verify: this request is not on the approved list, so the run stops")

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
		fmt.Sprintf("종목             %s (%s)", s.Symbol, s.Market),
		fmt.Sprintf("방향             %s, LIMIT", strings.ToUpper(s.Side)),
		fmt.Sprintf("수량             %s주", trim(s.Quantity)),
		fmt.Sprintf("지정가           %s", trim(s.Price)),
		fmt.Sprintf("명목금액         %s", trim(s.Quantity*s.Price)),
		fmt.Sprintf("가격 규칙        %s", s.Basis),
	}
	if s.Key != "" {
		lines = append(lines, fmt.Sprintf("clientOrderId    %s", s.Key))
	}
	return lines
}

// --- the human gate -------------------------------------------------------------

// request is one live request on its way through the gate.
type request struct {
	Kind     MutationKind
	Symbol   string
	Side     string
	Quantity float64
	// Detail and Reversal are what a per-mutation prompt shows. The batch model
	// showed the same facts before the run started, so they are only rendered
	// under --confirm-each.
	Detail   []string
	Reversal string
}

// gate is what stands between a step and the network.
//
// Under the default batch model it checks the approved plan and nothing else: the
// operator already read this request, in this run, on the list they typed a string
// for. Under --confirm-each it asks for that request's own expiring nonce. There is
// no third branch, and neither branch has an override.
func (r *Runner) gate(sr *stepRun, req request) error {
	if err := r.authorise(sr, req); err != nil {
		return err
	}
	if !r.confirmEach {
		fmt.Fprintf(r.out, "    [승인된 배치] %s\n", req.Kind.VerbKO())
		return nil
	}
	m := NewMutation(sr.step.ID, req.Kind.VerbKO(), r.accountRef, req.Detail, req.Reversal, r.now())
	return r.confirm(m)
}

// authorise refuses anything the approved batch does not carry a line for.
//
// It fails closed in the strongest sense available: with no plan at all — a runner
// that never reached its approval — nothing is authorised.
func (r *Runner) authorise(sr *stepRun, req request) error {
	if r.confirmEach {
		return nil
	}
	if r.plan == nil {
		return fmt.Errorf("%w: this run has no approved batch, so %s may send nothing", ErrOutsidePlan, sr.step.ID)
	}
	if r.plan.Authorises(sr.step.ID, req.Kind, req.Symbol, req.Side, req.Quantity) {
		return nil
	}
	return fmt.Errorf("%w: %s is about to %s for %s %s %s, which is not on the list approved at the start of "+
		"this run. Nothing was sent. Start a new run so the change is listed and approved on its own terms",
		ErrOutsidePlan, sr.step.ID, req.Kind, strings.ToUpper(orNone(req.Side)), trim(req.Quantity),
		orNone(req.Symbol))
}

// placeOrder is the gated order placement.
func (r *Runner) placeOrder(ctx context.Context, sr *stepRun, spec orderSpec, reversal string) (string, error) {
	if err := r.checkOrderCap(sr); err != nil {
		return "", err
	}
	if err := r.gate(sr, request{
		Kind: MutatePlaceOrder, Symbol: spec.Symbol, Side: spec.Side, Quantity: spec.Quantity,
		Detail: spec.detail(), Reversal: reversal,
	}); err != nil {
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
		return "", fmt.Errorf("verify: 브로커가 주문을 받아들였지만 orderId를 돌려주지 않았다")
	}
	sr.created("order", res.OrderID, spec.Symbol, r.now(), "")
	fmt.Fprintf(r.out, "    주문 접수 %s (%s %s x%s @ %s)\n",
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
// Four things keep it honest. The approved batch carries its own line for it, so a
// person read that a replay would be sent before the run started. The body must be
// identical to the one that was confirmed — the caller passes the same orderSpec,
// and a mismatch is the conflict probe, which does not come through here. It is
// printed before it is sent, so nothing is hidden. And whatever comes back is
// registered as an artifact, so if the broker does create a second order it is
// cancelled like any other.
func (r *Runner) replayPlaceOrder(ctx context.Context, sr *stepRun, spec orderSpec, why string) (string, error) {
	if err := r.authorise(sr, request{
		Kind: MutateReplayOrder, Symbol: spec.Symbol, Side: spec.Side, Quantity: spec.Quantity,
	}); err != nil {
		return "", err
	}
	fmt.Fprintf(r.out, "    동일 본문을 키 %s로 재전송 — %s\n", spec.Key, why)
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
// It has its own line on the approved list (and its own prompt under
// --confirm-each), because it is a different order: if the broker does not honour
// the key it will create one.
func (r *Runner) conflictProbe(ctx context.Context, sr *stepRun, spec orderSpec) (string, error) {
	if err := r.checkOrderCap(sr); err != nil {
		return "", err
	}
	detail := append(spec.detail(),
		"주의             위 키를 의도적으로 재사용하되 가격을 다르게 한다",
		"expected         the broker refuses with idempotency-key-conflict and creates nothing")
	if err := r.gate(sr, request{
		Kind: MutateConflictProbe, Symbol: spec.Symbol, Side: spec.Side, Quantity: spec.Quantity,
		Detail:   detail,
		Reversal: "브로커가 거부하지 않고 주문을 만들면 즉시 취소한다",
	}); err != nil {
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

// cancelOrder is the gated cancel.
func (r *Runner) cancelOrder(ctx context.Context, sr *stepRun, orderID, symbol, why string) error {
	detail := []string{
		fmt.Sprintf("주문             %s", orderID),
		fmt.Sprintf("종목             %s", symbol),
		fmt.Sprintf("사유             %s", why),
		"방향             노출을 줄인다 — 미체결 주문을 제거한다",
	}
	if err := r.gate(sr, request{
		Kind: MutateCancelOrder, Symbol: symbol, Detail: detail,
		Reversal: "계좌는 이 단계가 발견한 상태로 되돌아간다",
	}); err != nil {
		return err
	}

	started := r.now()
	res, err := r.broker.CancelOrder(ctx, orderID)
	sr.logCall(EndpointCancelOrder, map[string]string{"orderId": orderID}, started, r.now(), res, err)
	if err != nil {
		return err
	}
	sr.cancelled("order", orderID, symbol, r.now(), "")
	fmt.Fprintf(r.out, "    주문 취소 %s\n", orderID)
	if id := strings.TrimSpace(res.CurrentOrderID); id != "" && id != orderID {
		sr.observe("order.cancel.response_order_id_differs", "true",
			"취소가 원주문 "+orderID+"에 대해 "+id+"를 돌려주었다")
	}
	return nil
}

// amendOrder is the gated amend.
func (r *Runner) amendOrder(ctx context.Context, sr *stepRun, orderID, symbol string, price, quantity float64) (string, error) {
	detail := []string{
		fmt.Sprintf("주문             %s", orderID),
		fmt.Sprintf("종목             %s", symbol),
		fmt.Sprintf("새 지정가        %s (시장에서 더 멀고 여전히 체결 불가)", trim(price)),
		fmt.Sprintf("새 수량          %s주 — KR은 정정에 수량을 요구한다", trim(quantity)),
	}
	if err := r.gate(sr, request{
		Kind: MutateAmendOrder, Symbol: symbol, Quantity: quantity, Detail: detail,
		Reversal: "이후 살아 있는 쪽 식별자를 같은 단계에서 취소한다",
	}); err != nil {
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
	fmt.Fprintf(r.out, "    주문 정정 %s -> %s\n", orderID, current)
	return current, nil
}

// createConditional is the gated conditional registration.
func (r *Runner) createConditional(ctx context.Context, sr *stepRun, body official.ConditionalCreateBody, basis string) (string, error) {
	if err := r.checkConditionalCap(sr); err != nil {
		return "", err
	}
	detail := conditionalDetail(body, basis)
	if err := r.gate(sr, request{
		Kind: MutateRegisterConditional, Symbol: body.Symbol, Side: body.First.OrderSide,
		Quantity: parseDecimal(body.Quantity), Detail: detail,
		Reversal: "conditional-cancel 단계가 취소한다. 그 전에 의도적으로 이 프로세스보다 오래 산다",
	}); err != nil {
		return "", err
	}

	started := r.now()
	ref, err := r.broker.CreateConditionalOrder(ctx, body)
	sr.logCall(EndpointCreateConditional, body, started, r.now(), ref, err)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ref.ID) == "" {
		return "", fmt.Errorf("verify: 브로커가 조건주문을 받아들였지만 conditionalOrderId를 돌려주지 않았다")
	}
	sr.created("conditional-order", ref.ID, body.Symbol, r.now(), "")
	fmt.Fprintf(r.out, "    조건주문 등록 %s (%s %s 발동가 %s)\n",
		ref.ID, body.Type, body.Symbol, body.First.TriggerPrice)
	return ref.ID, nil
}

// replayCreateConditional is createConditional's unprompted replay, for the same
// reason and under the same four protections as replayPlaceOrder: its own line on
// the approved list, an identical body, printed before it is sent, and anything it
// creates is an artifact.
func (r *Runner) replayCreateConditional(ctx context.Context, sr *stepRun, body official.ConditionalCreateBody) (string, error) {
	if err := r.authorise(sr, request{
		Kind: MutateReplayConditional, Symbol: body.Symbol, Side: body.First.OrderSide,
		Quantity: parseDecimal(body.Quantity),
	}); err != nil {
		return "", err
	}
	fmt.Fprintf(r.out, "    동일 조건주문 본문을 키 %s로 재전송\n", body.ClientOrderID)
	started := r.now()
	ref, err := r.broker.CreateConditionalOrder(ctx, body)
	sr.logCall(EndpointCreateConditional, body, started, r.now(), ref, err)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

// modifyConditional is the gated conditional modify.
func (r *Runner) modifyConditional(ctx context.Context, sr *stepRun, id, symbol string, body official.ConditionalModifyBody, basis string) (string, error) {
	detail := []string{
		fmt.Sprintf("조건주문         %s", id),
		fmt.Sprintf("종목             %s", symbol),
		fmt.Sprintf("유형             %s / %s", body.Type, body.OrderType),
		fmt.Sprintf("새 발동가        %s (%s)", body.First.TriggerPrice, basis),
		fmt.Sprintf("수량             %s주", body.Quantity),
		"주의             openapi에 따르면 정정은 취소 후 재생성이다 — 새 id가 발급되고 기존 id는 무효화된다",
	}
	if err := r.gate(sr, request{
		Kind: MutateModifyConditional, Symbol: symbol, Side: body.First.OrderSide,
		Quantity: parseDecimal(body.Quantity), Detail: detail,
		Reversal: "이후 살아 있는 쪽 식별자는 conditional-cancel 단계가 취소한다",
	}); err != nil {
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
	fmt.Fprintf(r.out, "    조건주문 정정 %s -> %s\n", id, current)
	return current, nil
}

// cancelConditional is the gated conditional cancel.
func (r *Runner) cancelConditional(ctx context.Context, sr *stepRun, id, symbol, why string) error {
	detail := []string{
		fmt.Sprintf("조건주문         %s", id),
		fmt.Sprintf("종목             %s", symbol),
		fmt.Sprintf("사유             %s", why),
		"방향             계좌에서 보호 주문을 제거한다",
	}
	if err := r.gate(sr, request{
		Kind: MutateCancelConditional, Symbol: symbol, Detail: detail,
		Reversal: "계좌는 이 검증이 발견한 상태로 되돌아간다",
	}); err != nil {
		return err
	}

	started := r.now()
	err := r.broker.CancelConditionalOrder(ctx, id)
	sr.logCall(EndpointCancelConditional, map[string]string{"conditionalOrderId": id}, started, r.now(), nil, err)
	if err != nil {
		return err
	}
	sr.cancelled("conditional-order", id, symbol, r.now(), "")
	fmt.Fprintf(r.out, "    조건주문 취소 %s\n", id)
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
		fmt.Sprintf("종목             %s", body.Symbol),
		fmt.Sprintf("유형             %s / %s", body.Type, body.OrderType),
		fmt.Sprintf("감시             가격이 %s에 닿으면 %s", body.First.TriggerPrice, body.First.OrderSide),
		fmt.Sprintf("수량             %s주", body.Quantity),
		fmt.Sprintf("만료             %s", body.ExpireDate),
		fmt.Sprintf("가격 규칙        %s", basis),
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
