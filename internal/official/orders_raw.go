package official

// orders_raw.go adds raw-JSON order reads (harden-execution-base task 2.9,
// issues.md Manager decision (c)).
//
// # Why raw, and why a new file
//
// The adapted reads next door lose two things the engine's safety path needs:
//
//   - Orders() returns only the page it was given. Its caller cannot walk the
//     rest, because nextCursor and hasNext are decoded and then dropped.
//   - adaptOrder() has nowhere to put canceledAt, since domain.Order has no such
//     field — and canceledAt is what separates "cancelled" from "filled" inside
//     the API's single CLOSED status.
//
// Both matter for IN_DOUBT resolution: an order on page two that is reported as
// absent is a live position the engine does not know about, and a cancellation
// mistaken for a fill is the same error in the other direction.
//
// This is a new file with new methods rather than a change to Orders() or
// domain.Order: the adapted shape is a serialisation contract for `tossctl ...
// --output json` and the MCP order tools, and widening it to serve an engine
// concern would put a CLI-visible contract at risk for no CLI benefit. Everything
// here reuses the existing send/token/account-header path, so authentication,
// the 401-refresh retry and error classification behave identically.

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// RawOrderPage is one page of the order list, unadapted.
//
// Orders holds each order exactly as the broker sent it, so a consumer can read
// fields the domain type does not carry (canceledAt above all).
// internal/brokerstate.ParseOfficialOrder is the intended reader.
type RawOrderPage struct {
	Orders     []json.RawMessage
	NextCursor string
	HasNext    bool
}

// rawOrderPageBody mirrors the PaginatedOrderResponse envelope, keeping the order
// objects undecoded.
type rawOrderPageBody struct {
	Orders     []json.RawMessage `json:"orders"`
	NextCursor string            `json:"nextCursor"`
	HasNext    bool              `json:"hasNext"`
}

// OrdersPageRaw fetches one page of orders without adapting them.
//
// cursor takes precedence over filter.Cursor, so a pagination loop can hold a
// single filter value and pass the cursor along as it walks. An empty cursor
// argument falls back to the filter's, which keeps the method usable as a direct
// replacement for a one-shot Orders() call.
//
// Requires the X-Tossinvest-Account header, which getAcct resolves (lazily, once)
// exactly as it does for every other account-scoped read.
func (c *Client) OrdersPageRaw(ctx context.Context, filter OrdersFilter, cursor string) (RawOrderPage, error) {
	q := url.Values{}
	if filter.Status != "" {
		q.Set("status", filter.Status)
	}
	if filter.Symbol != "" {
		q.Set("symbol", filter.Symbol)
	}
	if filter.From != "" {
		q.Set("from", filter.From)
	}
	if filter.To != "" {
		q.Set("to", filter.To)
	}
	if effective := firstNonEmptyString(cursor, filter.Cursor); effective != "" {
		q.Set("cursor", effective)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}

	var body rawOrderPageBody
	if err := c.getAcct(ctx, "/api/v1/orders", q, &body); err != nil {
		return RawOrderPage{}, err
	}
	return RawOrderPage{
		Orders:     body.Orders,
		NextCursor: body.NextCursor,
		HasNext:    body.HasNext,
	}, nil
}

// OrderRawByID fetches one order without adapting it.
//
// The cancel/amend IN_DOUBT procedures derive an order's true state from
// (status, canceledAt, filledQuantity, quantity, lineage); OrderByID's
// domain.Order cannot express the second of those.
func (c *Client) OrderRawByID(ctx context.Context, orderID string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.getAcct(ctx, "/api/v1/orders/"+url.PathEscape(orderID), nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
