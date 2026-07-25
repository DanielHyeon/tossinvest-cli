package execgw

// orders_source.go binds the official client to the read interfaces the
// resolution procedures declare (harden-execution-base task 2.9).
//
// The interfaces live next to their users (indoubt.go) rather than here, so the
// resolver can be tested against fakes with no HTTP in sight. This file is the one
// place that knows both sides, and it is deliberately thin: no retry logic, no
// caching, no error rewriting. Retries belong to the Retrier (retry matrix), and
// rewriting the official sentinels here would break the classification the matrix
// depends on.

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// RawOrderClient is the read surface OfficialOrders needs.
//
// It is an interface rather than *official.Client so that engine wiring can pass
// its read-only view of the client (engine.OfficialReads, task 4.2) without
// re-widening it back to the concrete type — which would hand a mutator to code
// that is only supposed to be reading. Every existing caller passes
// *official.Client and still compiles.
type RawOrderClient interface {
	OrdersPageRaw(ctx context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error)
	OrderRawByID(ctx context.Context, orderID string) (json.RawMessage, error)
}

// AccountClient is the read surface OfficialAccount needs. Same reasoning as
// RawOrderClient.
type AccountClient interface {
	BuyingPower(ctx context.Context, currency string) (domain.BuyingPower, error)
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
}

// OfficialOrders adapts the official client to OrderPager and OrderReader.
type OfficialOrders struct {
	Client RawOrderClient
}

// OrdersPageRaw fetches one page of orders.
func (o OfficialOrders) OrdersPageRaw(ctx context.Context, q OrderQuery, cursor string) (OrderPage, error) {
	page, err := o.Client.OrdersPageRaw(ctx, official.OrdersFilter{
		Status: q.Status,
		Symbol: q.Symbol,
		From:   q.From,
		To:     q.To,
		Limit:  q.Limit,
	}, cursor)
	if err != nil {
		return OrderPage{}, err
	}
	return OrderPage{
		Orders:     page.Orders,
		NextCursor: page.NextCursor,
		HasNext:    page.HasNext,
	}, nil
}

// OrderRaw fetches a single order.
func (o OfficialOrders) OrderRaw(ctx context.Context, orderID string) (json.RawMessage, error) {
	return o.Client.OrderRawByID(ctx, orderID)
}

// OfficialAccount adapts the official client to AccountReader, the absence
// cross-check's input.
type OfficialAccount struct {
	Client AccountClient
}

// BuyingPower returns the cash buying power in the given currency.
func (a OfficialAccount) BuyingPower(ctx context.Context, currency string) (float64, error) {
	bp, err := a.Client.BuyingPower(ctx, currency)
	if err != nil {
		return 0, err
	}
	return bp.CashBuyingPower, nil
}

// HoldingQuantity returns the held quantity of one symbol.
//
// A symbol that is not in the holdings list is a zero holding, not an error: "you
// own none of it" is exactly what an empty response means, and turning it into an
// error would make the cross-check fail closed on the common case of a first
// purchase.
func (a OfficialAccount) HoldingQuantity(ctx context.Context, symbol string) (float64, error) {
	positions, err := a.Client.Holdings(ctx, symbol)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, p := range positions {
		if equalFoldTrimmed(p.Symbol, symbol) {
			total += p.Quantity
		}
	}
	return total, nil
}

// Compile-time proof that the adapters satisfy what the resolver asks for. A
// signature drift on either side fails the build here rather than at wiring time.
var (
	_ OrderPager     = OfficialOrders{}
	_ OrderReader    = OfficialOrders{}
	_ AccountReader  = OfficialAccount{}
	_ RawOrderClient = (*official.Client)(nil)
	_ AccountClient  = (*official.Client)(nil)
)

// equalFoldTrimmed compares symbols the way the broker writes them: trimmed and
// case-insensitive.
func equalFoldTrimmed(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
