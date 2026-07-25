// Package soak is the read-only capability survey that a TossOS capability
// attestation is built from (openspec change verify-execution-capability, task
// 1.1).
//
// # What it is for
//
// internal/attest describes the file the engine's automation gate is interlocked
// on, and says plainly that the writer of that file "must contain no mutation
// transport at all". This package is that writer's engine. It walks the official
// Open API's read endpoints on a timer for days at a stretch and records, in a
// durable local file, three things nobody can establish from a build:
//
//	credentials   did the access token renew itself with nobody watching, every
//	              day, for N consecutive days
//	endpoints     what fraction of calls to each endpoint succeeded, how long
//	              they took, and how often the broker throttled us
//	completeness  does the order list paginate correctly, does it contain every
//	              order the broker reports as open, does a quote request return a
//	              quote for every symbol asked for
//
// When all of that passes, BuildAttestation turns the record into an
// attest.Attestation. When it does not, it refuses and says why. There is no
// third outcome and no partial attestation: the engine reads this file as
// permission to trade unattended.
//
// # Why the broker arrives as an interface
//
// Reads has six methods and every one of them is a GET. The type is not a
// convenience — it is the mutation exclusion. A package that never imports
// internal/official cannot call PlaceOrder no matter what a future edit does to
// it, and the import-graph assertion in static_test.go keeps it that way. The
// wiring that owns a real client lives in cmd/tossctl/soak.go, one file, checked
// by the same test.
//
// Note what the interface returns: counts, identifiers and errors. No prices, no
// quantities, no cash balances. The record is written once per cycle for days and
// ends up in a support transcript sooner or later; it should carry evidence about
// the API, not a picture of somebody's portfolio.
package soak

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// Endpoints the survey exercises, spelled exactly as the engine's interlock
// spells them (internal/app/engine.RequiredEndpoints) so the two sets compare
// without a translation table in between.
const (
	EndpointAccounts    = "GET /api/v1/accounts"
	EndpointBuyingPower = "GET /api/v1/buying-power"
	EndpointHoldings    = "GET /api/v1/holdings"
	EndpointOrders      = "GET /api/v1/orders"
	EndpointOrderByID   = "GET /api/v1/orders/{id}"
	EndpointPrices      = "GET /api/v1/prices"
)

// RequiredEndpoints are the reads a completed soak has to have proven.
//
// Every one of them is a GET, and TestRequiredEndpointsAreAllReads keeps it that
// way: a read-only tool that required a write would be a tool that can never
// finish.
func RequiredEndpoints() []string {
	return []string{
		EndpointAccounts,
		EndpointBuyingPower,
		EndpointHoldings,
		EndpointOrders,
		EndpointOrderByID,
		EndpointPrices,
	}
}

// LiveOnlyEndpoints are the calls the engine's interlock also requires but that
// this tool structurally cannot make.
//
// They are reported, never attested. They come from the supervised one-off live
// verification in the same change (task 2.2: minimum quantity, limit-only,
// cancelled immediately), and until that has run, an attestation written here
// covers the reads and nothing else — which is exactly what the engine will then
// refuse to start on, by design.
func LiveOnlyEndpoints() []string {
	return []string{
		"POST /api/v1/orders",
		"POST /api/v1/orders/{id}/cancel",
	}
}

// OrderPage is one page of the order list: the identifiers on it and the cursor
// that leads to the next one.
type OrderPage struct {
	// IDs are the order identifiers on this page, in the order the broker
	// returned them.
	IDs []string
	// NextCursor is the cursor for the following page.
	NextCursor string
	// HasNext is the broker's own claim that another page exists.
	HasNext bool
}

// Reads is the read-only broker surface the survey drives.
//
// It is deliberately narrower than the official client: no method here returns a
// tradeable object or accepts an intent, so there is nothing an implementation
// could be asked to do that would touch an account.
type Reads interface {
	// Accounts returns the account references the credentials resolve to. The
	// first non-empty entry becomes the account the attestation names.
	Accounts(ctx context.Context) ([]string, error)
	// BuyingPower reads the account's buying power in currency. Only the success
	// of the read is recorded; the amount is not.
	BuyingPower(ctx context.Context, currency string) error
	// Holdings reads every position on the account and returns how many there
	// are.
	Holdings(ctx context.Context) (int, error)
	// OrdersPage reads one page of the order list. status is the broker's filter
	// ("" for every order, "OPEN" for the working ones); cursor is empty for the
	// first page.
	OrdersPage(ctx context.Context, status, cursor string) (OrderPage, error)
	// Order reads a single order by identifier.
	Order(ctx context.Context, id string) error
	// Prices reads quotes for symbols and returns how many came back.
	Prices(ctx context.Context, symbols []string) (int, error)
}

// Class is why a read failed. The distinction that matters is auth versus
// everything else: the consecutive-day streak is a claim about credentials, and
// a dropped TCP connection is not a credential problem.
type Class string

const (
	// ClassOK is recorded on a successful read.
	ClassOK Class = "ok"
	// ClassAuth is a refused token: the thing the soak exists to detect.
	ClassAuth Class = "auth"
	// ClassRateLimited is a 429.
	ClassRateLimited Class = "rate_limited"
	// ClassTransport is a network failure.
	ClassTransport Class = "transport"
	// ClassServer is a 5xx.
	ClassServer Class = "server"
	// ClassOther is anything else, including an unclassified error.
	ClassOther Class = "other"
)

// statusOpen is the broker's filter value for orders that are still working.
const statusOpen = "OPEN"

// defaultMaxOrderPages bounds the order walk.
//
// A cap is needed because the list has no date filter here and an account with
// years of history would spend the whole cycle paging through it. Reaching the
// cap does not fail the completeness check — the pagination contract was still
// observed to hold across every page walked — but it does suppress the
// open-order coverage check, because an open order could be sitting on a page
// beyond the cap and calling that "missing" would be a false alarm.
const defaultMaxOrderPages = 25

// Options configures a Runner.
type Options struct {
	// Reads is the broker surface. Required.
	Reads Reads
	// Recorder receives one line per cycle. Nil runs the survey without
	// persisting it, which is only useful to a test.
	Recorder *Recorder
	// Clock drives the interval and every timestamp. Nil uses the system clock.
	Clock clock.Clock

	// Interval is the wait between cycles. Zero runs them back to back.
	Interval time.Duration
	// Cycles bounds the run. Zero runs until the context is cancelled, which is
	// how the multi-day soak is meant to be used.
	Cycles int

	// Symbols are the tickers the quote read asks for.
	Symbols []string
	// Currency is the buying-power currency. Empty uses KRW.
	Currency string
	// MaxOrderPages bounds the order walk. Zero uses the default.
	MaxOrderPages int

	// Classify maps a broker error to a Class. Nil classifies everything as
	// ClassOther, which is safe but blunt: without it an expired credential looks
	// like a network glitch and the streak would not reset.
	Classify func(error) Class

	// TokenExpiry reports the cached access token's expiry. A value that moves
	// forward between cycles is the observation that proves an unattended
	// refresh happened. Nil means the expiry cannot be seen, which Evaluate
	// treats as "unproven" rather than "fine".
	TokenExpiry func() (time.Time, bool)

	// Progress receives a human-readable line per cycle. Nil is silent.
	Progress io.Writer
}

// Runner executes the survey.
type Runner struct {
	opts Options

	// lastTokenExpiry is the previous cycle's observation, held so a forward
	// move can be recognised as a refresh.
	lastTokenExpiry time.Time
	haveTokenExpiry bool
}

// New validates the options and returns a Runner.
func New(opts Options) (*Runner, error) {
	if opts.Reads == nil {
		return nil, errors.New("soak: no read surface was supplied; there is nothing to survey")
	}
	if opts.Clock == nil {
		opts.Clock = clock.System()
	}
	if strings.TrimSpace(opts.Currency) == "" {
		opts.Currency = "KRW"
	}
	if opts.MaxOrderPages <= 0 {
		opts.MaxOrderPages = defaultMaxOrderPages
	}
	if len(opts.Symbols) == 0 {
		return nil, errors.New("soak: no symbols to quote; the quote endpoint cannot be surveyed without one")
	}
	if opts.Cycles < 0 {
		return nil, fmt.Errorf("soak: a negative cycle count (%d) is not a run", opts.Cycles)
	}
	return &Runner{opts: opts}, nil
}

// Run executes cycles until the count is reached or the context is cancelled.
//
// A cancellation between cycles returns ctx.Err() with everything already
// recorded still on disk — the soak is a multi-day process and Ctrl-C is a
// normal way to end it.
func (r *Runner) Run(ctx context.Context) error {
	for i := 0; r.opts.Cycles == 0 || i < r.opts.Cycles; i++ {
		cycle, err := r.RunCycle(ctx)
		if err != nil {
			return err
		}
		if r.opts.Recorder != nil {
			if err := r.opts.Recorder.Append(cycle); err != nil {
				// A cycle we cannot record is a cycle that did not happen, as far
				// as any later judgement is concerned. Stopping is the honest
				// response: continuing would build a streak out of gaps.
				return err
			}
		}
		r.writeProgress(cycle, i+1)

		last := r.opts.Cycles != 0 && i == r.opts.Cycles-1
		if last {
			return nil
		}
		if r.opts.Interval > 0 {
			if err := r.opts.Clock.Sleep(ctx, r.opts.Interval); err != nil {
				return err
			}
		} else if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

// RunCycle performs one pass over every surveyed endpoint.
//
// It returns an error only when the context ended; a failed *read* is part of
// the measurement and is recorded in the returned Cycle rather than raised.
func (r *Runner) RunCycle(ctx context.Context) (Cycle, error) {
	if err := ctx.Err(); err != nil {
		return Cycle{}, err
	}

	cycle := Cycle{
		FormatVersion: RecordFormatVersion,
		Kind:          KindCycle,
		StartedAt:     r.opts.Clock.Now(),
	}

	// 1. accounts — also the credential probe. It is the one read that must
	//    happen first: without an account reference the cycle cannot say which
	//    portfolio it measured.
	accounts, accountsResult := r.probeAccounts(ctx)
	cycle.Endpoints = append(cycle.Endpoints, accountsResult)
	for _, a := range accounts {
		if strings.TrimSpace(a) != "" {
			cycle.AccountRef = strings.TrimSpace(a)
			break
		}
	}
	cycle.Credential = r.observeCredential(accountsResult)

	// 2. balances.
	cycle.Endpoints = append(cycle.Endpoints, r.probe(ctx, EndpointBuyingPower, func(ctx context.Context) error {
		return r.opts.Reads.BuyingPower(ctx, r.opts.Currency)
	}))

	positions, holdingsResult := r.probeHoldings(ctx)
	cycle.Endpoints = append(cycle.Endpoints, holdingsResult)

	// 3. orders: the full walk, then the open subset, then one order by id.
	walk, openWalk, ordersResult := r.probeOrders(ctx)
	cycle.Endpoints = append(cycle.Endpoints, ordersResult)
	cycle.Endpoints = append(cycle.Endpoints, r.probeOrderByID(ctx, walk))

	// 4. quotes.
	quotes, pricesResult := r.probePrices(ctx)
	cycle.Endpoints = append(cycle.Endpoints, pricesResult)

	cycle.Completeness = completenessOf(walk, openWalk, ordersResult.OK, positions, len(r.opts.Symbols), quotes)
	cycle.FinishedAt = r.opts.Clock.Now()
	return cycle, nil
}

// --- individual probes ------------------------------------------------------

// probe times a single read and classifies its outcome.
func (r *Runner) probe(ctx context.Context, endpoint string, call func(context.Context) error) EndpointResult {
	started := r.opts.Clock.Now()
	err := call(ctx)
	result := EndpointResult{
		Endpoint:  endpoint,
		Requests:  1,
		LatencyMS: r.opts.Clock.Since(started).Milliseconds(),
	}
	if err != nil {
		result.Class = r.classify(err)
		result.Error = err.Error()
		return result
	}
	result.OK = true
	result.Class = ClassOK
	return result
}

func (r *Runner) probeAccounts(ctx context.Context) ([]string, EndpointResult) {
	var accounts []string
	result := r.probe(ctx, EndpointAccounts, func(ctx context.Context) error {
		var err error
		accounts, err = r.opts.Reads.Accounts(ctx)
		return err
	})
	return accounts, result
}

func (r *Runner) probeHoldings(ctx context.Context) (int, EndpointResult) {
	var positions int
	result := r.probe(ctx, EndpointHoldings, func(ctx context.Context) error {
		var err error
		positions, err = r.opts.Reads.Holdings(ctx)
		return err
	})
	return positions, result
}

func (r *Runner) probePrices(ctx context.Context) (int, EndpointResult) {
	var quotes int
	result := r.probe(ctx, EndpointPrices, func(ctx context.Context) error {
		var err error
		quotes, err = r.opts.Reads.Prices(ctx, r.opts.Symbols)
		return err
	})
	return quotes, result
}

// orderWalk is what one pass over the order list observed.
type orderWalk struct {
	ids        []string
	seen       map[string]int
	pages      int
	requests   int
	cursorLoop bool
	truncated  bool
	pageLimit  bool
	err        error
}

// probeOrders walks the whole list and then the open subset. Both walks are
// recorded against GET /api/v1/orders, because that is the endpoint an
// attestation would name.
func (r *Runner) probeOrders(ctx context.Context) (full, open orderWalk, result EndpointResult) {
	started := r.opts.Clock.Now()

	full = r.walkOrders(ctx, "")
	open = r.walkOrders(ctx, statusOpen)

	result = EndpointResult{
		Endpoint:  EndpointOrders,
		Requests:  full.requests + open.requests,
		LatencyMS: r.opts.Clock.Since(started).Milliseconds(),
	}
	switch {
	case full.err != nil:
		result.Class = r.classify(full.err)
		result.Error = full.err.Error()
	case open.err != nil:
		result.Class = r.classify(open.err)
		result.Error = open.err.Error()
	default:
		result.OK = true
		result.Class = ClassOK
	}
	return full, open, result
}

// walkOrders follows the cursor to the end of the list, or to the first sign
// that the cursor cannot be followed.
func (r *Runner) walkOrders(ctx context.Context, status string) orderWalk {
	walk := orderWalk{seen: map[string]int{}}
	cursor := ""
	visited := map[string]bool{}

	for {
		if walk.pages >= r.opts.MaxOrderPages {
			walk.pageLimit = true
			return walk
		}
		page, err := r.opts.Reads.OrdersPage(ctx, status, cursor)
		walk.requests++
		if err != nil {
			walk.err = err
			return walk
		}
		walk.pages++
		for _, id := range page.IDs {
			walk.seen[id]++
			walk.ids = append(walk.ids, id)
		}

		if !page.HasNext {
			return walk
		}
		if strings.TrimSpace(page.NextCursor) == "" {
			// The broker says there is more and gives us no way to reach it.
			walk.truncated = true
			return walk
		}
		if visited[page.NextCursor] {
			walk.cursorLoop = true
			return walk
		}
		visited[page.NextCursor] = true
		cursor = page.NextCursor
	}
}

// probeOrderByID reads the first order the walk found.
//
// With no orders on the account there is nothing to read, and the result is a
// recorded skip rather than a success — Evaluate then refuses to complete the
// soak until the endpoint has actually been exercised, which is the honest
// answer: the engine calls it, and nobody has seen it work.
func (r *Runner) probeOrderByID(ctx context.Context, walk orderWalk) EndpointResult {
	if len(walk.ids) == 0 {
		reason := "the account has no orders, so there is no id to read"
		if walk.err != nil {
			reason = "the order list could not be read, so no id was available"
		}
		return EndpointResult{Endpoint: EndpointOrderByID, Skipped: true, SkipReason: reason}
	}
	id := walk.ids[0]
	return r.probe(ctx, EndpointOrderByID, func(ctx context.Context) error {
		return r.opts.Reads.Order(ctx, id)
	})
}

// --- credentials ------------------------------------------------------------

// observeCredential turns the authenticated read's outcome, plus the token
// expiry, into the cycle's credential record.
func (r *Runner) observeCredential(accounts EndpointResult) Credential {
	cred := Credential{OK: accounts.OK, Class: accounts.Class}
	if r.opts.TokenExpiry == nil {
		return cred
	}
	expiry, ok := r.opts.TokenExpiry()
	if !ok || expiry.IsZero() {
		return cred
	}
	cred.Observed = true
	cred.TokenExpiresAt = expiry.UTC()
	if r.haveTokenExpiry && expiry.After(r.lastTokenExpiry) {
		cred.Refreshed = true
	}
	r.lastTokenExpiry = expiry
	r.haveTokenExpiry = true
	return cred
}

func (r *Runner) classify(err error) Class {
	if err == nil {
		return ClassOK
	}
	if r.opts.Classify == nil {
		return ClassOther
	}
	if c := r.opts.Classify(err); c != "" {
		return c
	}
	return ClassOther
}

// --- completeness -----------------------------------------------------------

// completenessOf answers the question the engine actually needs answered: if it
// reads this API, does it see everything that exists?
func completenessOf(full, open orderWalk, ordersOK bool, positions, symbolsAsked, quotes int) Completeness {
	c := Completeness{
		OrderPages:          full.pages,
		OrderIDs:            len(full.ids),
		Positions:           positions,
		OpenOrders:          len(open.ids),
		QuotesRequested:     symbolsAsked,
		QuotesReturned:      quotes,
		CursorLoop:          full.cursorLoop || open.cursorLoop,
		TruncatedPagination: full.truncated || open.truncated,
		PageLimitReached:    full.pageLimit || open.pageLimit,
	}
	if !ordersOK {
		// The walk did not finish, so there is nothing to conclude. The failure is
		// already recorded against the endpoint; recording it a second time as a
		// completeness failure would let one dropped connection block the
		// attestation for the rest of the window.
		c.Detail = "not evaluated: the order list could not be read"
		return c
	}
	c.Evaluated = true

	var problems []string
	for id, n := range full.seen {
		if n > 1 {
			c.DuplicateOrderIDs++
			if len(problems) < 4 {
				problems = append(problems, fmt.Sprintf("order %s appeared %d times in the list", id, n))
			}
		}
	}
	if !c.PageLimitReached {
		for _, id := range open.ids {
			if full.seen[id] == 0 {
				c.OpenOrdersMissing++
				if len(problems) < 8 {
					problems = append(problems, fmt.Sprintf("order %s was open but absent from the list", id))
				}
			}
		}
	}
	if c.CursorLoop {
		problems = append(problems, "the pagination cursor pointed back at a page already read")
	}
	if c.TruncatedPagination {
		problems = append(problems, "the broker reported another page and returned no cursor for it")
	}
	if quotes != symbolsAsked {
		problems = append(problems, fmt.Sprintf("%d quote(s) returned for %d symbol(s)", quotes, symbolsAsked))
	}
	if c.PageLimitReached {
		problems = append(problems, fmt.Sprintf(
			"the order walk stopped at the %d-page limit, so open-order coverage was not checked", full.pages))
	}

	c.OK = c.DuplicateOrderIDs == 0 && c.OpenOrdersMissing == 0 &&
		!c.CursorLoop && !c.TruncatedPagination && quotes == symbolsAsked
	c.Detail = strings.Join(problems, "; ")
	return c
}

// --- progress ---------------------------------------------------------------

func (r *Runner) writeProgress(c Cycle, n int) {
	if r.opts.Progress == nil {
		return
	}
	ok, total := 0, 0
	for _, e := range c.Endpoints {
		if e.Skipped {
			continue
		}
		total++
		if e.OK {
			ok++
		}
	}
	credentials := "ok"
	if !c.Credential.OK {
		credentials = "FAILED (" + string(c.Credential.Class) + ")"
	}
	if c.Credential.Refreshed {
		credentials += ", token refreshed"
	}
	completeness := "not evaluated"
	if c.Completeness.Evaluated {
		completeness = "ok"
		if !c.Completeness.OK {
			completeness = "FAILED: " + c.Completeness.Detail
		}
	}
	fmt.Fprintf(r.opts.Progress, "%s  cycle %d  credentials %s  endpoints %d/%d  orders %d on %d page(s)  completeness %s\n",
		c.StartedAt.UTC().Format(time.RFC3339), n, credentials, ok, total,
		c.Completeness.OrderIDs, c.Completeness.OrderPages, completeness)
}
