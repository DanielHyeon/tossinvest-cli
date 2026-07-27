package official

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	trading "github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// ---------------------------------------------------------------------------
// Wire-format types
// ---------------------------------------------------------------------------

// orderCreateV0 is the quantity-based variant of OrderCreateRequest.
// Covers: non-fractional LIMIT, non-fractional MARKET, fractional SELL.
//
// Schema: OrderCreateQuantityBased (openapi.latest.json)
// Required: symbol, side, orderType, quantity.
// Optional: price (LIMIT only), timeInForce (LIMIT only).
// clientOrderId is the idempotency key: "동일 값으로 재요청 시 이전 주문 결과를
// 그대로 재반환합니다 … 멱등성 키는 10분간 유효하며" and "서버는 자동 생성하지
// 않습니다" (openapi.latest.json). omitempty is therefore load-bearing: an absent
// field means "멱등성 미적용", and "" is not a value the documented pattern
// `^[a-zA-Z0-9\-_]+$` accepts.
type orderCreateV0 struct {
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	OrderType             string `json:"orderType"`
	Quantity              string `json:"quantity"`
	Price                 string `json:"price,omitempty"`
	TimeInForce           string `json:"timeInForce,omitempty"`
	ClientOrderID         string `json:"clientOrderId,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

// orderCreateV1 is the amount-based variant of OrderCreateRequest.
// Covers: fractional BUY only.
//
// Schema: OrderCreateAmountBased (openapi.latest.json)
// Required: symbol, side, orderType=MARKET, orderAmount.
type orderCreateV1 struct {
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	OrderType             string `json:"orderType"`
	OrderAmount           string `json:"orderAmount"`
	ClientOrderID         string `json:"clientOrderId,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

// apiOrderCreateResponse mirrors the OrderResponse schema returned by
// POST /api/v1/orders (inside the {"result": ...} envelope).
//
// ClientOrderID is "요청 시 전달한 값 그대로 반환. 미전달 시 null"
// (OrderResponse.clientOrderId, openapi.latest.json) — the pointer distinguishes
// that documented null from an echo of a different key, which is not a response
// about the order this process sent.
type apiOrderCreateResponse struct {
	OrderID       string  `json:"orderId"`
	ClientOrderID *string `json:"clientOrderId"`
}

// apiOrderOperationResponse mirrors the OrderOperationResponse schema returned
// by POST /api/v1/orders/{orderId}/cancel and /modify.
type apiOrderOperationResponse struct {
	OrderID string `json:"orderId"`
}

// orderModifyRequest mirrors the OrderModifyRequest schema.
// Required: orderType.
// Optional: price (LIMIT only), quantity (KR only).
type orderModifyRequest struct {
	OrderType             string `json:"orderType"`
	Price                 string `json:"price,omitempty"`
	Quantity              string `json:"quantity,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

// ---------------------------------------------------------------------------
// Builder
// ---------------------------------------------------------------------------

// buildOrderCreate constructs the correct OrderCreateRequest variant from a
// PlaceIntent.
//
// Variant mapping table:
//
//	Intent                | Variant | orderType | Key fields
//	----------------------|---------|-----------|-------------------------------
//	fractional BUY        | 1       | MARKET    | orderAmount (USD decimal str)
//	fractional SELL       | 0       | MARKET    | quantity (decimal str), no price
//	non-fractional LIMIT  | 0       | LIMIT     | price + quantity + timeInForce=DAY
//	non-fractional MARKET | 0       | MARKET    | quantity only
//
// side is uppercased from intent (intent.Side is lowercase "buy"/"sell").
// All numeric values are formatted as shortest decimal strings.
func buildOrderCreate(intent orderintent.PlaceIntent) (any, error) {
	side := strings.ToUpper(intent.Side)

	if intent.Fractional && intent.Side == "buy" {
		// variant1: amount-based (US MARKET fractional buy)
		if intent.Amount <= 0 {
			return nil, fmt.Errorf("fractional buy requires a positive amount, got %v", intent.Amount)
		}
		return orderCreateV1{
			Symbol:                intent.Symbol,
			Side:                  side,
			OrderType:             "MARKET",
			OrderAmount:           formatDecimal(intent.Amount),
			ClientOrderID:         intent.ClientOrderID,
			ConfirmHighValueOrder: false,
		}, nil
	}

	// variant0: quantity-based
	v0 := orderCreateV0{
		Symbol:                intent.Symbol,
		Side:                  side,
		ClientOrderID:         intent.ClientOrderID,
		ConfirmHighValueOrder: false,
	}

	if intent.Fractional && intent.Side == "sell" {
		// fractional sell: MARKET, decimal quantity, no price
		if intent.Quantity <= 0 {
			return nil, fmt.Errorf("fractional sell requires a positive quantity, got %v", intent.Quantity)
		}
		v0.OrderType = "MARKET"
		v0.Quantity = formatDecimal(intent.Quantity)
		return v0, nil
	}

	if intent.Quantity <= 0 {
		return nil, fmt.Errorf("order requires a positive quantity, got %v", intent.Quantity)
	}

	orderType := strings.ToUpper(intent.OrderType)
	v0.OrderType = orderType
	v0.Quantity = formatDecimal(intent.Quantity)

	if orderType == "LIMIT" {
		if intent.Price <= 0 {
			return nil, fmt.Errorf("limit order requires a positive price, got %v", intent.Price)
		}
		v0.Price = formatDecimal(intent.Price)
		v0.TimeInForce = "DAY"
	}

	return v0, nil
}

// formatDecimal converts a float64 to its shortest decimal string.
// Examples: 100 → "100", 0.5 → "0.5", 150.25 → "150.25".
func formatDecimal(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// acctHeaders resolves the account sequence (lazily if not set via
// WithAccountSeq) and returns the extra-headers map for account-scoped order
// endpoints. An error is returned when lazy resolution fails (e.g. no accounts
// found or the accounts endpoint is unreachable).
func (c *Client) acctHeaders(ctx context.Context) (map[string]string, error) {
	seq, err := c.ensureAccountSeq(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]string{"X-Tossinvest-Account": strconv.Itoa(seq)}, nil
}

// ---------------------------------------------------------------------------
// Mutation methods
// ---------------------------------------------------------------------------

// PlaceOrder submits a new order via the official Open API.
//
// Endpoint: POST /api/v1/orders
// Response: OrderResponse → orderId
// Sends X-Tossinvest-Account when accountSeq is set.
func (c *Client) PlaceOrder(ctx context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error) {
	body, err := buildOrderCreate(intent)
	if err != nil {
		return trading.MutationResult{}, fmt.Errorf("building order body: %w", err)
	}
	hdrs, err := c.acctHeaders(ctx)
	if err != nil {
		return trading.MutationResult{}, err
	}
	var resp apiOrderCreateResponse
	if err := c.postWithHeaders(ctx, "/api/v1/orders", body, hdrs, &resp); err != nil {
		return trading.MutationResult{}, err
	}
	// The echo is read back rather than ignored: the response is documented to
	// return the key "요청 시 전달한 값 그대로" (OrderResponse.clientOrderId), so a
	// different non-null key describes a different order. Accepting it as an ack
	// would attach somebody else's order id to this attempt. A null/absent echo is
	// the documented shape for an unkeyed request and is left alone.
	if intent.ClientOrderID != "" && resp.ClientOrderID != nil && *resp.ClientOrderID != intent.ClientOrderID {
		return trading.MutationResult{}, fmt.Errorf(
			"order response echoed clientOrderId %q but the request sent %q",
			*resp.ClientOrderID, intent.ClientOrderID)
	}
	return trading.MutationResult{
		Kind:     "place",
		Status:   "accepted",
		OrderID:  resp.OrderID,
		Symbol:   intent.Symbol,
		Market:   intent.Market,
		Quantity: intent.Quantity,
		Price:    intent.Price,
	}, nil
}

// CancelOrder cancels an existing order by its orderId via the official Open API.
//
// Endpoint: POST /api/v1/orders/{orderId}/cancel
// Response: OrderOperationResponse → orderId (new identifier issued for the cancel)
// Sends X-Tossinvest-Account when accountSeq is set.
// The orderId path segment is percent-encoded via url.PathEscape.
func (c *Client) CancelOrder(ctx context.Context, orderID string) (trading.MutationResult, error) {
	path := "/api/v1/orders/" + url.PathEscape(orderID) + "/cancel"
	hdrs, err := c.acctHeaders(ctx)
	if err != nil {
		return trading.MutationResult{}, err
	}
	var resp apiOrderOperationResponse
	if err := c.postWithHeaders(ctx, path, struct{}{}, hdrs, &resp); err != nil {
		return trading.MutationResult{}, err
	}
	return trading.MutationResult{
		Kind:    "cancel",
		Status:  "accepted",
		OrderID: resp.OrderID,
	}, nil
}

// ModifyOrder amends an existing order via the official Open API.
//
// Endpoint: POST /api/v1/orders/{orderId}/modify
// Response: OrderOperationResponse → orderId
// Sends X-Tossinvest-Account when accountSeq is set.
//
// orderType inference:
//   - intent.Price != nil → LIMIT (price is required for LIMIT)
//   - intent.Price == nil → MARKET
//
// quantity is forwarded when set (required for KR stocks, forbidden for US).
func (c *Client) ModifyOrder(ctx context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error) {
	body := orderModifyRequest{
		ConfirmHighValueOrder: false,
	}
	if intent.Price != nil {
		body.OrderType = "LIMIT"
		body.Price = formatDecimal(*intent.Price)
	} else {
		body.OrderType = "MARKET"
	}
	if intent.Quantity != nil {
		body.Quantity = formatDecimal(*intent.Quantity)
	}

	path := "/api/v1/orders/" + url.PathEscape(intent.OrderID) + "/modify"
	hdrs, err := c.acctHeaders(ctx)
	if err != nil {
		return trading.MutationResult{}, err
	}
	var resp apiOrderOperationResponse
	if err := c.postWithHeaders(ctx, path, body, hdrs, &resp); err != nil {
		return trading.MutationResult{}, err
	}
	return trading.MutationResult{
		Kind:            "amend",
		Status:          "accepted",
		OriginalOrderID: intent.OrderID,
		CurrentOrderID:  resp.OrderID,
	}, nil
}
