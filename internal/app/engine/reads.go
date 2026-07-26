package engine

// reads.go seals the engine's official client behind a read-only interface
// (harden-execution-base task 4.2; issues.md 2026-07-26 "잔존 리스크").
//
// # What was still open after task 2.5
//
// Task 2.5 made the order mutators unexported, so no caller holding the engine
// wiring could place an order without a GuardianDecision. It left one door: the
// field `Context.Official *official.Client`, exported for queries — and
// *official.Client also has PlaceOrder, CancelOrder, ModifyOrder and the three
// conditional verbs. A caller with the wiring could therefore still reach the
// broker directly, with no journal record, no Guardian decision and no IN_DOUBT
// handling. The seal was real for one type and cosmetic for the other.
//
// The field is now typed as OfficialReads, an interface that simply does not
// declare those six methods. "Submit without the gateway" is not a thing the
// engine's API can express any more — a caller cannot even name the method, let
// alone call it (engine-safety: "raw mutator 미노출").
//
// # Why an interface and not a wrapper struct
//
// A wrapper would need a hand-written method per call and would have to be
// updated every time the engine reads something new, which makes the seal a
// maintenance tax that someone eventually pays by widening it. An interface the
// concrete client already satisfies costs nothing to extend with a further *read*
// and cannot be extended with a write without that write appearing, in this file,
// under this comment.

import (
	"context"
	"encoding/json"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// OfficialReads is the engine's read-only view of the official Open API client.
//
// Every method here is a GET. Adding a mutating method to this interface defeats
// its only purpose; if the engine needs a new broker write, it goes through
// internal/execgw.Gateway like every other one.
type OfficialReads interface {
	// --- account -----------------------------------------------------------

	Accounts(ctx context.Context) ([]domain.Account, error)
	BuyingPower(ctx context.Context, currency string) (domain.BuyingPower, error)
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
	// HoldingsRaw is the same single read as Holdings with the broker's decimals
	// unadapted. It joined the interface with the runtime (add-engine-runtime):
	// the reconciliation snapshot prefers the raw path so an adoption's
	// `cost_basis` is the broker's own string rather than a re-rendered float
	// (position-ledger: 원문 decimal 문자열 보존 SHALL), and the engine had no
	// production caller for either path before there was a loop to drive one.
	HoldingsRaw(ctx context.Context, symbol string) ([]official.RawHolding, error)
	SellableQuantity(ctx context.Context, symbol string) (domain.SellableQuantity, error)

	// --- orders ------------------------------------------------------------

	Orders(ctx context.Context, filter official.OrdersFilter) ([]domain.Order, error)
	OrderByID(ctx context.Context, orderID string) (domain.Order, error)
	// OrdersPageRaw and OrderRawByID are the unadapted reads the state
	// derivation needs (canceledAt, pagination cursors).
	OrdersPageRaw(ctx context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error)
	OrderRawByID(ctx context.Context, orderID string) (json.RawMessage, error)
	ConditionalOrders(ctx context.Context, status, symbol, cursor string, limit int) (domain.ConditionalOrderList, error)
	ConditionalOrder(ctx context.Context, id string) (domain.ConditionalOrder, error)

	// --- market ------------------------------------------------------------

	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
	PriceLimits(ctx context.Context, symbol string) (domain.PriceLimits, error)
	Orderbook(ctx context.Context, symbol string) (domain.OrderBook, error)
	ExchangeRate(ctx context.Context, base, quote string) (domain.ExchangeRate, error)

	// --- diagnostics -------------------------------------------------------

	BaseURL() string
}

// Compile-time proof that the concrete client still satisfies the read surface.
// A signature drift in internal/official fails the build here rather than at
// wiring time.
var _ OfficialReads = (*official.Client)(nil)
