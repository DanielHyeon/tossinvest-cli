// Package reconcile compares what the engine believes about an account with what
// the account says, and decides what that disagreement means.
//
// # The account wins, always
//
// The one rule everything else serves: the Toss account is the authority
// (WORKFLOW 불변 규칙, reconciliation spec "충돌 시 토스 계좌가 항상 우선한다").
// Nothing in this package writes a local belief over an account reading. A
// disagreement is a reason to stop opening positions, never a reason to correct
// the broker's number.
//
// # Why a snapshot is a contract and not just three reads
//
// The three reads happen at three different instants, and an account changes
// between them. A snapshot assembled from an open-order list read before a fill
// and a holdings list read after it describes an account that never existed —
// and the invented discrepancy would block trading for a fill that was perfectly
// normal.
//
// So the contract fixes what can be relied on:
//
//	fixed order   open orders (walked to the last page) → holdings → balances
//	as-of         stamped from the *first* read, because a snapshot is only as
//	              fresh as its oldest part
//	all or none   a partial read is discarded whole; there is no such thing as
//	              most of a snapshot
//	stabilised    a difference counts only when two snapshots at least
//	              MinInterval apart agree
//
// The ordering is the interesting one. Orders first, positions last, means an
// order that fills mid-snapshot appears as still-open in the order list *and*
// as filled in the holdings — an apparent double count, which reads as a
// discrepancy and blocks entries. The other order would show the fill in neither,
// which reads as agreement and lets the engine trade on a stale picture. Between
// a false alarm and a false all-clear on a live account, the false alarm is the
// one to design for; the stabilisation rule is what stops it becoming noise.
package reconcile

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// PositionsReader sweeps the account's holdings.
type PositionsReader interface {
	Positions(ctx context.Context) ([]domain.Position, error)
}

// RawHolding is one holding with the broker's own decimal strings preserved.
type RawHolding struct {
	Symbol string
	// Name is the broker's display name, carried for the operator surfaces only.
	Name     string
	Market   string
	Quantity string
	// AveragePrice is the broker's cost basis exactly as it arrived. Empty means
	// the payload carried none, which is a different fact from zero.
	AveragePrice string
}

// RawPositionsReader is the optional half of PositionsReader: the *same single
// read*, with the decimals unadapted.
//
// It is optional because only the official client can satisfy it, and this
// package is testable — and useful — without one. When the injected reader does
// satisfy it the collector takes this path instead of the float one, so the §0.4
// budget is unchanged: one holdings request either way.
//
// The value it adds is evidence. `position_adoptions.cost_basis` stores the
// broker's `averagePurchasePrice` verbatim (position-ledger: 원문 decimal 문자열
// 보존 SHALL), and a number that has been through float64 is no longer the
// broker's answer to the fee question 2b has to measure.
type RawPositionsReader interface {
	PositionsRaw(ctx context.Context) ([]RawHolding, error)
}

// BuyingPowerReader reads the cash buying power of one currency.
type BuyingPowerReader interface {
	BuyingPower(ctx context.Context, currency string) (float64, error)
}

// BrokerOrder is one open order as the broker reported it.
//
// Quantities and prices stay decimal strings all the way through: this is the
// record a mismatch is judged against, and a float round-trip in the middle of
// that judgement is a mismatch nobody can explain later.
type BrokerOrder struct {
	OrderID        string
	AccountRef     string
	Market         string
	TradingDay     string
	Symbol         string
	Side           string
	Status         string
	Quantity       string
	FilledQuantity string
	Price          string
	// OrderedAt preserves the official payload evidence verbatim. TradingDay is
	// the canonical date used for identity matching when that evidence exists.
	OrderDate string
	OrderedAt string
}

// Holding is one account position.
type Holding struct {
	Symbol       string
	Quantity     string
	AveragePrice string
	Market       string
	// Name is the broker's display name for the instrument, when the holdings read
	// supplied one.
	//
	// Display evidence, on the same footing as CostBasisRaw below: the comparison
	// reads Symbol, Quantity and Market and nothing else, so two holdings that
	// differ only by Name are the same holding to Compare. It rides here rather
	// than in a parallel structure because it arrives in the same response, and a
	// second shape carrying the same read is a second shape that can disagree with
	// the first (a085).
	Name string
	// CostBasisRaw is the broker's cost basis exactly as it arrived, when the
	// holdings reader could supply one (RawPositionsReader). It is empty
	// otherwise, and empty is honest: "no raw string was preserved" rather than
	// a re-rendering of the float beside it.
	//
	// Only the adoption record consumes it, and only to store it. Nothing
	// compares or computes with it — AveragePrice is what the comparison uses,
	// and the R formula uses neither (design A7).
	CostBasisRaw string
}

// Balance is one currency's cash buying power.
type Balance struct {
	Currency        string
	CashBuyingPower string
}

// Snapshot is one complete reading of the account.
type Snapshot struct {
	// AsOf is when the first read started. It is deliberately the *oldest*
	// instant in the snapshot: freshness claims have to be true of the whole
	// thing, and the open-order list is what a staleness threshold cares about.
	AsOf time.Time
	// CompletedAt is when the last read returned.
	CompletedAt time.Time
	AccountRef  string

	OpenOrders []BrokerOrder
	Holdings   []Holding
	Balances   []Balance
}

// Age reports how old the snapshot is, measured from its oldest part.
func (s Snapshot) Age(now time.Time) time.Duration { return now.Sub(s.AsOf) }

// Empty reports whether this is the zero snapshot (which is what a discarded
// partial read returns).
func (s Snapshot) Empty() bool { return s.AsOf.IsZero() && s.CompletedAt.IsZero() }

// Digest is a canonical rendering of everything the snapshot says about the
// account, excluding when it was taken.
//
// Excluding the timestamps is the point: stabilisation asks "did the account stop
// moving", and two readings a second apart always differ in their clocks.
func (s Snapshot) Digest() string {
	var b strings.Builder
	b.WriteString("account=" + s.AccountRef + "\n")

	orders := append([]BrokerOrder(nil), s.OpenOrders...)
	sort.Slice(orders, func(i, j int) bool {
		return brokerOrderIdentity(orders[i], s.AccountRef).less(brokerOrderIdentity(orders[j], s.AccountRef))
	})
	for _, o := range orders {
		identity := brokerOrderIdentity(o, s.AccountRef)
		fmt.Fprintf(&b, "order\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			identity.AccountRef, identity.Market, identity.TradingDay, identity.Symbol,
			identity.Side, identity.OrderID, o.Status, o.OrderedAt,
			canonicalDecimal(o.Quantity), canonicalDecimal(o.FilledQuantity),
			canonicalDecimal(o.Price))
	}

	holdings := append([]Holding(nil), s.Holdings...)
	sort.Slice(holdings, func(i, j int) bool { return holdings[i].Symbol < holdings[j].Symbol })
	for _, h := range holdings {
		fmt.Fprintf(&b, "holding\t%s\t%s\t%s\n",
			h.Symbol, canonicalDecimal(h.Quantity), canonicalDecimal(h.AveragePrice))
	}

	balances := append([]Balance(nil), s.Balances...)
	sort.Slice(balances, func(i, j int) bool { return balances[i].Currency < balances[j].Currency })
	for _, bal := range balances {
		fmt.Fprintf(&b, "balance\t%s\t%s\n", bal.Currency, canonicalDecimal(bal.CashBuyingPower))
	}
	return b.String()
}

// Collector performs the three reads in the contracted order.
type Collector struct {
	// Orders walks the open-order list. Required.
	Orders execgw.OrderPager
	// Positions sweeps the holdings. Required.
	Positions PositionsReader
	// Balance reads buying power. Required — a snapshot without cash is not the
	// snapshot this contract describes.
	Balance BuyingPowerReader
	// Currencies are the balances to read, in the order given.
	Currencies []string
	// AccountRef identifies the account being read.
	AccountRef string
	// Clock stamps as-of. Defaults to clock.System().
	Clock clock.Clock
	// MaxPages bounds the order-list walk.
	MaxPages int
}

// ErrPartialSnapshot means a read failed and the snapshot was discarded.
var ErrPartialSnapshot = errors.New("reconcile: snapshot discarded after a partial read")

// Collect takes one snapshot, or none.
//
// "Or none" is the contract: every error path returns the zero Snapshot. A caller
// cannot accidentally compare against three-quarters of an account, because
// three-quarters of an account is not a value this function can produce.
//
// # Why the cause is wrapped and not printed
//
// Every failure below wraps two things with %w: ErrPartialSnapshot, which says
// what happened to the snapshot, and the cause, which says what the broker said.
// The second one used to be printed with %v, and that difference cost a restart:
// a boot-time 429 arrived here as prose, so the recovery upstream could not tell
// "the broker is throttling us for a few seconds" from "the account cannot be
// read", and it treated the throttle as permanent (2026-08-13 02:03:30.545Z).
//
// Wrapping only widens what errors.Is can answer — the ErrPartialSnapshot
// judgement every existing consumer makes is unchanged (a102 D1b).

func (c *Collector) Collect(ctx context.Context) (Snapshot, error) {
	if err := c.validate(); err != nil {
		return Snapshot{}, err
	}
	clk := c.clock()
	maxPages := c.MaxPages
	if maxPages <= 0 {
		maxPages = 50
	}

	snap := Snapshot{AsOf: clk.Now(), AccountRef: c.AccountRef}

	// 1. Open orders, walked to the last page. A partial list would make every
	//    order past page one look external.
	raws, err := execgw.ScanOrders(ctx, c.Orders, execgw.OrderQuery{Status: "OPEN"}, maxPages)
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: walking the open-order list: %w", ErrPartialSnapshot, err)
	}
	for _, raw := range raws {
		order, perr := parseBrokerOrder(raw)
		if perr != nil {
			return Snapshot{}, fmt.Errorf("%w: %w", ErrPartialSnapshot, perr)
		}
		snap.OpenOrders = append(snap.OpenOrders, order)
	}

	// 2. Holdings. One request either way: the raw path is the same endpoint with
	//    the decimals unadapted, so preferring it costs nothing in the §0.4
	//    budget and preserves the evidence the adoption record stores.
	holdings, err := c.holdings(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: sweeping the holdings: %w", ErrPartialSnapshot, err)
	}
	snap.Holdings = holdings

	// 3. Balances.
	for _, currency := range c.Currencies {
		bp, err := c.Balance.BuyingPower(ctx, currency)
		if err != nil {
			return Snapshot{}, fmt.Errorf("%w: reading the %s buying power: %w",
				ErrPartialSnapshot, currency, err)
		}
		snap.Balances = append(snap.Balances, Balance{
			Currency:        strings.ToUpper(strings.TrimSpace(currency)),
			CashBuyingPower: decimalString(bp),
		})
	}

	snap.CompletedAt = clk.Now()
	return snap, nil
}

// holdings performs the one holdings read, preferring the raw path.
//
// The two paths produce the same Holding fields; the raw one additionally fills
// CostBasisRaw. Nothing downstream branches on which path ran — a snapshot with
// an empty CostBasisRaw simply cannot supply an adoption's cost basis, and the
// adoption record says ABSENT for it, which is the honest answer.
func (c *Collector) holdings(ctx context.Context) ([]Holding, error) {
	if raw, ok := c.Positions.(RawPositionsReader); ok {
		items, err := raw.PositionsRaw(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]Holding, 0, len(items))
		for _, h := range items {
			out = append(out, Holding{
				Symbol:       strings.ToUpper(strings.TrimSpace(h.Symbol)),
				Quantity:     canonicalDecimal(h.Quantity),
				AveragePrice: canonicalDecimal(h.AveragePrice),
				Market:       strings.ToLower(strings.TrimSpace(h.Market)),
				CostBasisRaw: strings.TrimSpace(h.AveragePrice),
				Name:         strings.TrimSpace(h.Name),
			})
		}
		return out, nil
	}

	positions, err := c.Positions.Positions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Holding, 0, len(positions))
	for _, p := range positions {
		out = append(out, Holding{
			Symbol:       strings.ToUpper(strings.TrimSpace(p.Symbol)),
			Quantity:     decimalString(p.Quantity),
			AveragePrice: decimalString(p.AveragePrice),
			Market:       strings.ToLower(strings.TrimSpace(p.MarketType)),
			Name:         strings.TrimSpace(p.Name),
		})
	}
	return out, nil
}

func (c *Collector) validate() error {
	switch {
	case c.Orders == nil:
		return errors.New("reconcile: an order pager is required")
	case c.Positions == nil:
		return errors.New("reconcile: a positions reader is required")
	case c.Balance == nil:
		return errors.New("reconcile: a buying-power reader is required")
	case len(c.Currencies) == 0:
		return errors.New("reconcile: at least one currency is required — " +
			"a snapshot with no cash reading cannot answer whether an order was affordable")
	}
	return nil
}

func (c *Collector) clock() clock.Clock {
	if c.Clock == nil {
		return clock.System()
	}
	return c.Clock
}

// DefaultStabilisationInterval is the minimum gap between two snapshots that are
// allowed to corroborate each other (reconciliation spec: 기본 2초).
const DefaultStabilisationInterval = 2 * time.Second

// DefaultStabilisationCount is how many consecutive agreeing snapshots make a
// reading stable.
const DefaultStabilisationCount = 2

// Stabiliser decides when the account has stopped moving.
//
// Two snapshots taken microseconds apart agree about almost anything, including
// about a fill that is halfway through settling. The minimum interval is what
// makes agreement mean "nothing is in flight" rather than "we looked twice very
// quickly".
type Stabiliser struct {
	// MinInterval is the minimum age difference between two counted snapshots.
	// Zero takes DefaultStabilisationInterval.
	MinInterval time.Duration
	// Required is how many consecutive agreeing snapshots are needed. Zero takes
	// DefaultStabilisationCount.
	Required int

	last     Snapshot
	lastSeen bool
	digest   string
	streak   int
}

// StabilityResult is the verdict on one offered snapshot.
type StabilityResult struct {
	// Counted reports whether the snapshot was old enough to be considered.
	Counted bool
	// Streak is how many consecutive agreeing snapshots have now been seen.
	Streak int
	// Stable reports Streak >= Required.
	Stable bool
	// Why explains a rejection or a reset, for logs.
	Why string
}

// Offer presents a snapshot.
//
// A snapshot taken too soon after the previous one is not an error and not a
// disagreement: it is simply not evidence, so it is neither counted nor allowed
// to break an existing streak.
func (s *Stabiliser) Offer(snap Snapshot) StabilityResult {
	interval := s.MinInterval
	if interval == 0 {
		interval = DefaultStabilisationInterval
	}
	required := s.Required
	if required <= 0 {
		required = DefaultStabilisationCount
	}

	if snap.Empty() {
		return StabilityResult{Streak: s.streak, Why: "a discarded snapshot is not evidence"}
	}
	if s.lastSeen && snap.AsOf.Sub(s.last.AsOf) < interval {
		return StabilityResult{
			Streak: s.streak,
			Stable: s.streak >= required,
			Why: fmt.Sprintf("taken %s after the previous snapshot, inside the %s stabilisation interval",
				snap.AsOf.Sub(s.last.AsOf).Round(time.Millisecond), interval),
		}
	}

	digest := snap.Digest()
	switch {
	case !s.lastSeen:
		s.streak = 1
	case digest == s.digest:
		s.streak++
	default:
		s.streak = 1
	}
	s.last = snap
	s.lastSeen = true
	s.digest = digest

	return StabilityResult{
		Counted: true,
		Streak:  s.streak,
		Stable:  s.streak >= required,
	}
}

// Reset forgets the streak. A reconciliation that acted on a stable reading must
// start the next one from scratch rather than inherit its own corroboration.
func (s *Stabiliser) Reset() {
	s.lastSeen = false
	s.digest = ""
	s.streak = 0
	s.last = Snapshot{}
}

// Last returns the most recently counted snapshot.
func (s *Stabiliser) Last() (Snapshot, bool) { return s.last, s.lastSeen }

// --- payload reading --------------------------------------------------------

// officialOrder is the subset of the broker's order payload a snapshot keeps.
// Decimal fields stay strings on purpose (see BrokerOrder).
type officialOrder struct {
	OrderID    string  `json:"orderId"`
	AccountRef string  `json:"accountRef"`
	Market     string  `json:"market"`
	OrderDate  string  `json:"orderDate"`
	OrderedAt  string  `json:"orderedAt"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Status     string  `json:"status"`
	Quantity   *string `json:"quantity"`
	Price      *string `json:"price"`
	Execution  *struct {
		FilledQuantity *string `json:"filledQuantity"`
	} `json:"execution"`
}

func canonicalDecimal(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "0"
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		// Not a decimal: keep it verbatim rather than inventing a zero. A value
		// we cannot read must stay visible in the digest so it can differ.
		return trimmed
	}
	return decimalString(v)
}

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
