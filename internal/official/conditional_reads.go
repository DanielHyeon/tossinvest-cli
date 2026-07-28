package official

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// apiConditionalCondition mirrors ConditionalOrderCondition (nullable numeric
// fields arrive as "" when JSON null).
type apiConditionalCondition struct {
	Type             string `json:"type"`
	Status           string `json:"status"`
	TriggerPrice     string `json:"triggerPrice"`
	TargetProfitRate string `json:"targetProfitRate"`
	OrderPrice       string `json:"orderPrice"`
	TriggeredOrderID string `json:"triggeredOrderId"`
}

// apiConditionalOrder mirrors ConditionalOrderDetailResponse. second is a
// pointer so a JSON null (SINGLE) stays nil rather than a zero-value leg.
type apiConditionalOrder struct {
	ConditionalOrderID string                   `json:"conditionalOrderId"`
	Type               string                   `json:"type"`
	Status             string                   `json:"status"`
	Symbol             string                   `json:"symbol"`
	Market             string                   `json:"market"`
	Quantity           string                   `json:"quantity"`
	OrderType          string                   `json:"orderType"`
	ExpireDate         string                   `json:"expireDate"`
	First              apiConditionalCondition  `json:"first"`
	Second             *apiConditionalCondition `json:"second"`
	CreatedAt          string                   `json:"createdAt"`
}

// apiConditionalOrderPage mirrors PaginatedConditionalOrderResponse.
type apiConditionalOrderPage struct {
	ConditionalOrders []apiConditionalOrder `json:"conditionalOrders"`
	NextCursor        string                `json:"nextCursor"`
	HasNext           bool                  `json:"hasNext"`
}

// ConditionalOrders lists conditional orders (account-scoped, official-only).
// status/symbol/cursor/limit are optional filters (empty/0 → omitted).
func (c *Client) ConditionalOrders(ctx context.Context, status, symbol, cursor string, limit int) (domain.ConditionalOrderList, error) {
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
		return domain.ConditionalOrderList{}, err
	}
	orders := make([]domain.ConditionalOrder, 0, len(raw.ConditionalOrders))
	for _, o := range raw.ConditionalOrders {
		orders = append(orders, adaptConditionalOrder(o))
	}
	return domain.ConditionalOrderList{
		Orders:     orders,
		NextCursor: raw.NextCursor,
		HasNext:    raw.HasNext,
		FetchedAt:  time.Now().UTC(),
	}, nil
}

// --- the raw-preserving read (change console-orders-screen, task 1.5) -----------

// RawConditionalOrder is one conditional order with the broker's own decimal
// strings preserved.
//
// It is RawOrder's argument (orders_reads.go) on the other endpoint, and it is
// needed for the same screen rather than for a different one: internal/verifylive
// counts a leftover conditional and a leftover plain order as the same thing —
// something filling the live-exposure cap so nothing else can be sent — and M18
// measured that a conditional survives the process that registered it. In this
// product a conditional order is the durable artefact, not the corner case.
//
// The nullable numbers are the same trap. `orderPrice` is null on a MARKET-type
// conditional and `targetProfitRate` is null on a STOP leg, so
// [Client.ConditionalOrders]' parseDecimal renders both as a trigger of zero.
//
// Only the first leg's numbers are carried. The screen shows one row per
// conditional order and OCO's second leg needs a shape of its own; adding it
// later is additive again, and inventing an aggregate of two legs now would be a
// number nobody sent.
//
// It is additive: ConditionalOrders, ConditionalOrder and the two adapters are
// untouched, and this reads the same endpoint through the same getAcct path.
type RawConditionalOrder struct {
	ID     string
	Symbol string
	// Market is the payload's own "KR"/"US" — unlike the plain order endpoint,
	// this one carries it, so nothing is derived here.
	Market string
	// Type is "SINGLE" | "OCO"; Status is the order's own status ("WATCHING"),
	// which is a different vocabulary from the OPEN/CLOSED request filter.
	Type   string
	Status string
	// OrderType is the order this conditional would place ("LIMIT" | "MARKET").
	OrderType string
	// Quantity, TriggerPrice, TargetProfitRate and OrderPrice are the payload's
	// strings, untouched. Empty means the broker sent no value.
	Quantity         string
	TriggerPrice     string
	TargetProfitRate string
	OrderPrice       string
	// ConditionType is the first leg's own type ("STOP" | "PROFIT_RATE" | …).
	ConditionType string
	// TriggeredOrderID names the plain order this conditional turned into, empty
	// while it is still watching.
	TriggeredOrderID string
	ExpireDate       string
	CreatedAt        string
}

// RawConditionalOrderList is one page of RawConditionalOrder with the page
// boundary kept, for the same reason RawOrderList keeps it.
type RawConditionalOrderList struct {
	Orders     []RawConditionalOrder
	NextCursor string
	HasNext    bool
}

// ConditionalOrdersRaw fetches the same page [Client.ConditionalOrders] does,
// without adapting the decimals.
//
// The parameters are ConditionalOrders' exactly, so the two are interchangeable
// at the call site and the caller picks the shape it needs rather than calling
// both and spending two requests.
//
// status is required (see [ErrOrderStatusRequired]) and its value is passed
// verbatim. Unlike the plain endpoint, both groups here paginate — `limit`
// defaults to 20 and caps at 100 for OPEN as well — so naming the group buys a
// defined question rather than a whole answer, and the caller still has to read
// HasNext.
func (c *Client) ConditionalOrdersRaw(ctx context.Context, status, symbol, cursor string,
	limit int) (RawConditionalOrderList, error) {
	if strings.TrimSpace(status) == "" {
		return RawConditionalOrderList{}, fmt.Errorf("%w: GET /api/v1/conditional-orders documents "+
			"it as required, and the two groups are different questions — OPEN is "+
			"{WATCHING, PAUSED, ORDERING, ORDERED}, which is exactly the set filling the "+
			"live-exposure cap, and CLOSED is {COMPLETED, EXPIRED}. Without it the request is one "+
			"the broker is entitled to refuse, and a refused leftover read renders as no leftovers",
			ErrOrderStatusRequired)
	}

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
	out := RawConditionalOrderList{
		Orders:     make([]RawConditionalOrder, 0, len(raw.ConditionalOrders)),
		NextCursor: raw.NextCursor,
		HasNext:    raw.HasNext,
	}
	for _, o := range raw.ConditionalOrders {
		out.Orders = append(out.Orders, RawConditionalOrder{
			ID:        o.ConditionalOrderID,
			Symbol:    o.Symbol,
			Market:    o.Market,
			Type:      o.Type,
			Status:    o.Status,
			OrderType: o.OrderType,
			// Verbatim: parseDecimal here would put back the zero this type
			// exists to keep out.
			Quantity:         o.Quantity,
			TriggerPrice:     o.First.TriggerPrice,
			TargetProfitRate: o.First.TargetProfitRate,
			OrderPrice:       o.First.OrderPrice,
			ConditionType:    o.First.Type,
			TriggeredOrderID: o.First.TriggeredOrderID,
			ExpireDate:       o.ExpireDate,
			CreatedAt:        o.CreatedAt,
		})
	}
	return out, nil
}

// ConditionalOrder fetches one conditional order by id (account-scoped).
func (c *Client) ConditionalOrder(ctx context.Context, id string) (domain.ConditionalOrder, error) {
	var raw apiConditionalOrder
	if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &raw); err != nil {
		return domain.ConditionalOrder{}, err
	}
	return adaptConditionalOrder(raw), nil
}

// adaptCondition converts one raw watch leg to domain.
func adaptCondition(a apiConditionalCondition) domain.ConditionalOrderCondition {
	return domain.ConditionalOrderCondition{
		Type:             a.Type,
		Status:           a.Status,
		TriggerPrice:     parseDecimal(a.TriggerPrice),
		TargetProfitRate: parseDecimal(a.TargetProfitRate),
		OrderPrice:       parseDecimal(a.OrderPrice),
		TriggeredOrderID: a.TriggeredOrderID,
	}
}

// adaptConditionalOrder converts the official detail response to domain.
// second stays nil for SINGLE (raw pointer nil).
func adaptConditionalOrder(o apiConditionalOrder) domain.ConditionalOrder {
	out := domain.ConditionalOrder{
		ID:         o.ConditionalOrderID,
		Type:       o.Type,
		Status:     o.Status,
		Symbol:     o.Symbol,
		Market:     o.Market,
		Quantity:   parseDecimal(o.Quantity),
		OrderType:  o.OrderType,
		ExpireDate: o.ExpireDate,
		First:      adaptCondition(o.First),
		CreatedAt:  o.CreatedAt,
	}
	if o.Second != nil {
		s := adaptCondition(*o.Second)
		out.Second = &s
	}
	return out
}
