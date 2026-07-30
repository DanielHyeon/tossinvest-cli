package official

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// apiOrderExecution mirrors the execution sub-object inside an Order.
// All fields are nullable in the API (null while order is pending).
type apiOrderExecution struct {
	FilledQuantity     string `json:"filledQuantity"`
	AverageFilledPrice string `json:"averageFilledPrice"`
	FilledAmount       string `json:"filledAmount"`
	Commission         string `json:"commission"`
	Tax                string `json:"tax"`
	FilledAt           string `json:"filledAt"`
	SettlementDate     string `json:"settlementDate"`
}

// apiOrder mirrors the Order schema in PaginatedOrderResponse / single-order
// response.
// Endpoint: GET /api/v1/orders and GET /api/v1/orders/{orderId}
// Schema:
//
//	orderId     string — unique order identifier
//	symbol      string — ticker
//	side        string — "BUY" | "SELL"
//	orderType   string — "LIMIT" | "MARKET" etc.
//	timeInForce string — "DAY" etc.
//	status      string — "OPEN" | "CLOSED" etc.
//	quantity    string (decimal) — ordered quantity
//	price       string (decimal, nullable) — limit price; null for market orders
//	currency    string — "KRW" | "USD"
//	orderedAt   string (datetime, nullable) — order submission time (RFC3339)
//	canceledAt  string (datetime, nullable)
//	orderAmount string (decimal, nullable)
//	execution   apiOrderExecution
type apiOrder struct {
	OrderID     string            `json:"orderId"`
	Symbol      string            `json:"symbol"`
	Side        string            `json:"side"`
	OrderType   string            `json:"orderType"`
	TimeInForce string            `json:"timeInForce"`
	Status      string            `json:"status"`
	Quantity    string            `json:"quantity"`
	Price       string            `json:"price"`
	Currency    string            `json:"currency"`
	OrderedAt   string            `json:"orderedAt"`
	CanceledAt  string            `json:"canceledAt"`
	OrderAmount string            `json:"orderAmount"`
	Execution   apiOrderExecution `json:"execution"`
}

// apiOrderPage mirrors the PaginatedOrderResponse envelope.
type apiOrderPage struct {
	Orders     []apiOrder `json:"orders"`
	NextCursor string     `json:"nextCursor"`
	HasNext    bool       `json:"hasNext"`
}

// OrdersFilter specifies optional filters for the Orders list endpoint.
// Zero-value fields are omitted from the request query string.
type OrdersFilter struct {
	// Status filters orders by lifecycle state ("OPEN" | "CLOSED").
	Status string
	// Symbol restricts to a specific ticker.
	Symbol string
	// From is the start date (inclusive) in "YYYY-MM-DD" format.
	From string
	// To is the end date (inclusive) in "YYYY-MM-DD" format.
	To string
	// Cursor is the pagination cursor returned by a previous call's NextCursor.
	Cursor string
	// Limit is the maximum number of orders per page (0 = API default).
	Limit int
}

// Orders fetches the order history for the authenticated account.
// filter may be a zero-value OrdersFilter to use API defaults.
// Requires the X-Tossinvest-Account header; configure via WithAccountSeq.
func (c *Client) Orders(ctx context.Context, filter OrdersFilter) ([]domain.Order, error) {
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
	if filter.Cursor != "" {
		q.Set("cursor", filter.Cursor)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}

	var raw apiOrderPage
	if err := c.getAcct(ctx, "/api/v1/orders", q, &raw); err != nil {
		return nil, err
	}
	return adaptOrders(raw.Orders), nil
}

// OrderByID fetches a single order by its unique ID.
// Requires the X-Tossinvest-Account header; configure via WithAccountSeq.
func (c *Client) OrderByID(ctx context.Context, orderID string) (domain.Order, error) {
	var raw apiOrder
	if err := c.getAcct(ctx, "/api/v1/orders/"+url.PathEscape(orderID), nil, &raw); err != nil {
		return domain.Order{}, err
	}
	return adaptOrder(raw), nil
}

// --- the raw-preserving read (change console-orders-screen, task 1.1-1.4) -------

// RawOrder is one order with the broker's own decimal strings preserved.
//
// # Why this exists beside Orders
//
// [Client.Orders] sends every number through parseDecimal, which answers 0 for
// an empty string and 0 for a string it cannot parse. That is the right shape
// for a CLI printing a portfolio and every existing caller depends on it — but
// it means that by the time an order reaches domain.Order, "the broker sent no
// value", "the broker sent something unparseable" and "the broker sent zero" are
// one number.
//
// That is not an edge case here. The API sends `price` as null for a market
// order and the whole `execution` object as null while an order is still alive,
// so EVERY open order arrives with a filled quantity and an average fill price of
// zero. A screen built on that renders every live order as fully filled at a
// price of 0.
//
// It is RawHolding's argument (asset_reads.go) applied to orders: a value that
// has been through float64 and back has lost the evidence of what the broker
// actually said, and here the evidence is the difference between "not filled" and
// "filled at nothing". An empty string means the payload carried no value for
// that field, which is a different fact from "0".
//
// It is additive. Orders, OrderByID, adaptOrder and adaptOrders are untouched,
// this reads the same endpoint through the same getAcct path — so the token
// refresh, the account header and the error classification are the same ones —
// and no existing caller changes.
type RawOrder struct {
	// ID is orderId, the number the screen shows and the ledger joins on.
	ID     string
	Symbol string
	// Side is "BUY" | "SELL".
	Side string
	// OrderType is "LIMIT" | "MARKET" | ….
	OrderType string
	// Status is the broker's own status string, unmodified. Deriving a state
	// from it is internal/brokerstate's job and it needs the broker's own word,
	// including a value this build has never seen.
	Status string
	// Currency is "KRW" | "USD" as the payload spelled it.
	Currency string
	// Market is "KR" | "US", derived from Currency, and empty for a currency this
	// build does not recognise. The /orders response carries no market field and
	// no name field at all; Orders decodes the currency and drops it, so a market
	// column is either derived here or it does not exist.
	Market string

	// Quantity, Price, FilledQuantity and AverageFilledPrice are the payload's
	// strings, untouched. Empty means the broker sent no value for that field.
	Quantity           string
	Price              string
	FilledQuantity     string
	AverageFilledPrice string

	// OrderedAt and CanceledAt are the payload's timestamps, untouched. OrderedAt
	// is the actual instant (RFC3339) rather than domain.Order.OrderDate's first
	// ten characters, and CanceledAt has no home in domain.Order at all.
	OrderedAt  string
	CanceledAt string
}

// RawOrderList is one page of RawOrder, with the page boundary kept.
//
// [Client.Orders] decodes nextCursor and hasNext and then drops both, so its
// caller cannot tell a complete list from the first page of a longer one. A count
// taken from a truncated page is a confidently short number.
type RawOrderList struct {
	Orders     []RawOrder
	NextCursor string
	HasNext    bool
}

// ErrOrderStatusRequired is what the raw order reads answer when the caller did
// not name a lifecycle group.
//
// # Why this is a refusal and not a default
//
// `status` is documented `required: true` on both order list endpoints, and it is
// not a filter like `symbol` or `limit`: it selects between two differently shaped
// answers. Picking one here would be this client deciding which question the
// caller meant, and the wrong choice is silent — a plain read defaulted to CLOSED
// would answer with one page of old orders and no indication that every live one
// is missing.
//
// The refusal is here, in the client, rather than in a caller's tests, because the
// failure it prevents was a caller forgetting. `/orders` shipped composing
// `?limit=100` with no status, which is either a rejected request or one page over
// the account's whole history — and in the second case a live order past row 100
// is absent from the screen and from its counts, rendered as a floor nobody can
// resolve. A rule that lives only at the composition site is forgotten again at
// the next composition site, by a change nobody is reviewing for this.
//
// It is not classified as retryable and ShouldFallback answers false for it, which
// is right: a request this client will not send is not a reason to reach for the
// web session.
var ErrOrderStatusRequired = errors.New("official: an order list read must name a status group")

// OrdersRaw fetches the same page [Client.Orders] does, without adapting the
// decimals and without discarding the page boundary.
//
// It is one request to `GET /api/v1/orders`, the same one Orders makes, so a
// caller that needs both shapes should call this one and adapt rather than
// spending two requests out of the §0.4 budget.
//
// filter.Status is required (see [ErrOrderStatusRequired]) and its value is passed
// verbatim. The broker is the authority on which groups exist; this read refuses
// the absence of an answer, not an answer it has not seen before.
func (c *Client) OrdersRaw(ctx context.Context, filter OrdersFilter) (RawOrderList, error) {
	if strings.TrimSpace(filter.Status) == "" {
		return RawOrderList{}, fmt.Errorf("%w: GET /api/v1/orders documents it as required, and the "+
			"two groups are different questions — status=OPEN returns every pending order and "+
			"ignores limit and cursor, status=CLOSED returns one page and spans the whole account "+
			"history when no from/to is given. Without it the request is either refused or silently "+
			"bounded, and a silently bounded live read is a leftover the caller cannot see",
			ErrOrderStatusRequired)
	}

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
	if filter.Cursor != "" {
		q.Set("cursor", filter.Cursor)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}

	var raw apiOrderPage
	if err := c.getAcct(ctx, "/api/v1/orders", q, &raw); err != nil {
		return RawOrderList{}, err
	}

	out := RawOrderList{
		Orders:     make([]RawOrder, 0, len(raw.Orders)),
		NextCursor: raw.NextCursor,
		HasNext:    raw.HasNext,
	}
	for _, o := range raw.Orders {
		out.Orders = append(out.Orders, RawOrder{
			ID:        o.OrderID,
			Symbol:    o.Symbol,
			Side:      o.Side,
			OrderType: o.OrderType,
			Status:    o.Status,
			Currency:  o.Currency,
			Market:    marketFromCurrency(o.Currency),
			// Verbatim, deliberately. parseDecimal here would put the whole
			// point of this type back: an empty string would become 0 and the
			// screen would render a pending order as filled at zero.
			Quantity:           o.Quantity,
			Price:              o.Price,
			FilledQuantity:     o.Execution.FilledQuantity,
			AverageFilledPrice: o.Execution.AverageFilledPrice,
			OrderedAt:          o.OrderedAt,
			CanceledAt:         o.CanceledAt,
		})
	}
	return out, nil
}

// marketFromCurrency names the market a trading currency implies.
//
// The /orders response has no market field. An unrecognised currency answers
// empty rather than guessing: a wrong market on a row is worse than no market,
// because the operator filters by it.
func marketFromCurrency(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "KRW":
		return "KR"
	case "USD":
		return "US"
	default:
		return ""
	}
}

// adaptOrders converts a slice of official Order records to []domain.Order.
func adaptOrders(raw []apiOrder) []domain.Order {
	out := make([]domain.Order, 0, len(raw))
	for _, o := range raw {
		out = append(out, adaptOrder(o))
	}
	return out
}

// adaptOrder converts a single official Order to domain.Order.
//
// Mapping rationale (cross-referenced with internal/client/order.go WTS adapter):
//
//   - orderId → ID: direct pass-through.
//
//   - symbol → Symbol: direct pass-through.
//
//   - side → Side: "BUY" | "SELL", same as WTS.
//
//   - status → Status: "OPEN" | "CLOSED" etc.
//
//   - quantity (decimal string) → Quantity: parseDecimal.
//
//   - price (nullable decimal string) → Price: parseDecimal (0 for null/market).
//
//   - orderedAt → OrderDate: date portion (first 10 chars of RFC3339 string).
//     Also parsed to *time.Time for SubmittedAt if valid RFC3339.
//
//   - execution.filledQuantity → FilledQuantity: parseDecimal.
//
//   - execution.averageFilledPrice (nullable) → AverageExecutionPrice: parseDecimal.
//
//   - Name, Market: not available from this endpoint; left empty.
func adaptOrder(raw apiOrder) domain.Order {
	var orderDate string
	var submittedAt *time.Time
	if raw.OrderedAt != "" {
		if len(raw.OrderedAt) >= 10 {
			orderDate = raw.OrderedAt[:10]
		} else {
			orderDate = raw.OrderedAt
		}
		if t, err := time.Parse(time.RFC3339, raw.OrderedAt); err == nil {
			submittedAt = &t
		}
	}

	return domain.Order{
		ID:                    raw.OrderID,
		Symbol:                raw.Symbol,
		Side:                  raw.Side,
		Status:                raw.Status,
		Quantity:              parseDecimal(raw.Quantity),
		Price:                 parseDecimal(raw.Price),
		FilledQuantity:        parseDecimal(raw.Execution.FilledQuantity),
		AverageExecutionPrice: parseDecimal(raw.Execution.AverageFilledPrice),
		OrderDate:             orderDate,
		SubmittedAt:           submittedAt,
		// Name, Market — not available from /orders response
	}
}
