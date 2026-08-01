package official

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// ConditionalOrderRaw reads one conditional order without converting broker
// decimal strings through float64. Protection ownership and oversell checks are
// integer-exact and therefore cannot consume the lossy domain adapter.
func (c *Client) ConditionalOrderRaw(ctx context.Context, id string) (RawConditionalOrder, error) {
	var o apiConditionalOrder
	if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &o); err != nil {
		return RawConditionalOrder{}, err
	}
	return adaptProtectionConditional(o), nil
}

// ProtectionConditionalOrdersRaw is a protection-specific list preserving the
// broker client identity needed to recover a response-lost create. The general
// orders-screen reader intentionally keeps its pre-a045 shape unchanged.
func (c *Client) ProtectionConditionalOrdersRaw(ctx context.Context, status, symbol, cursor string, limit int) (RawConditionalOrderList, error) {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if symbol != "" {
		q.Set("symbol", symbol)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var raw apiConditionalOrderPage
	if err := c.getAcct(ctx, "/api/v1/conditional-orders", q, &raw); err != nil {
		return RawConditionalOrderList{}, err
	}
	out := RawConditionalOrderList{NextCursor: raw.NextCursor, HasNext: raw.HasNext, Orders: make([]RawConditionalOrder, 0, len(raw.ConditionalOrders))}
	for _, item := range raw.ConditionalOrders {
		out.Orders = append(out.Orders, adaptProtectionConditional(item))
	}
	return out, nil
}

func adaptProtectionConditional(o apiConditionalOrder) RawConditionalOrder {
	return RawConditionalOrder{
		ID: o.ConditionalOrderID, ClientOrderID: o.ClientOrderID, Symbol: o.Symbol, Market: o.Market,
		Type: o.Type, Status: o.Status, OrderType: o.OrderType, Quantity: o.Quantity,
		TriggerPrice: o.First.TriggerPrice, TargetProfitRate: o.First.TargetProfitRate,
		OrderPrice: o.First.OrderPrice, ConditionType: o.First.Type,
		TriggeredOrderID: o.First.TriggeredOrderID, ExpireDate: o.ExpireDate, CreatedAt: o.CreatedAt,
	}
}

// SellableQuantityRaw preserves the official decimal string for protection and
// flatten oversell checks. It uses the same account-scoped endpoint and adds no
// fallback transport.
func (c *Client) SellableQuantityRaw(ctx context.Context, symbol string) (string, time.Time, error) {
	q := url.Values{}
	q.Set("symbol", symbol)
	var raw apiSellableQuantity
	if err := c.getAcct(ctx, "/api/v1/sellable-quantity", q, &raw); err != nil {
		return "", time.Time{}, err
	}
	return raw.SellableQuantity, time.Now().UTC(), nil
}
