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
//	completeness  does the order list paginate correctly, does every order it
//	              returns carry an identifier, does a quote request return a
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

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
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
	// OrdersPage reads one page of the order list. status is the broker's
	// lifecycle group and is required — "OPEN" or "CLOSED", never empty; cursor
	// is empty for the first page and is only honoured for "CLOSED".
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

// The broker's two order lifecycle groups, and the only two values GET
// /api/v1/orders accepts for its `status` parameter.
//
// openapi.latest.json marks the parameter required, with enum {OPEN, CLOSED} and
// no member meaning "everything": a request that omits it is answered with 400
// invalid-request naming the field. Walking the whole list therefore means
// walking both groups and taking the union — the same shape
// execgw.Resolver.scanBoth already uses for IN_DOUBT resolution.
//
// The groups are labels over the per-order status, and they are not disjoint:
// the spec puts PARTIAL_FILLED in both. An order in both lists is one order, and
// nothing here may read that overlap as a broken list.
const (
	statusOpen   = "OPEN"
	statusClosed = "CLOSED"
)

// defaultMaxOrderPages bounds the order walk.
//
// A cap is needed because the CLOSED group is the account's whole history — the
// walk sets no from/to bound — and an account with years behind it would spend
// the entire cycle paging through it. (The OPEN group is not paginated at all:
// the spec has it ignore cursor and limit and return every working order in one
// response, so the cap is effectively a bound on the CLOSED walk.)
//
// Reaching the cap does not fail the completeness check. The pagination contract
// was still observed to hold across every page that was walked; what is beyond
// the cap is unread, not wrong.
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

	// PauseWhile reports that something else needs this account's rate budget,
	// and why. It is consulted before every cycle; while it says yes the cycle is
	// delayed rather than run.
	//
	// The one caller that sets it is cmd/tossctl, and the one thing it reports on
	// is a live verification in progress (internal/runlock). The survey yields
	// because the two are not worth the same: a soak cycle missed is a cycle, and
	// a verification step lost to a 429 is a live order that has to be placed
	// again with a person watching (measurements.md M4).
	//
	// It is a function rather than a path so this package keeps reaching for
	// nothing but its own record, and so a test can drive the pause without a
	// filesystem. Nil never pauses.
	PauseWhile func() (paused bool, reason string)

	// Binary fingerprints the executable this process was started from.
	//
	// It is read once when the Runner is built — that is what this process IS —
	// and again at each cycle boundary, which is how a survey that has been
	// running for two days notices that the binary under it was reinstalled. The
	// reading is stamped on every recorded cycle so a reader of the record can
	// tell which build produced it.
	//
	// Nil means the survey does not fingerprint itself: no stamp on the record, no
	// self-upgrade, and no warning anywhere. That is the pre-1.8 behaviour and it
	// is still a valid way to run.
	Binary func() (binstamp.Stamp, error)

	// ReExec replaces this process with the installed binary, preserving the
	// arguments it was started with.
	//
	// It is consulted only at a cycle boundary: a cycle in flight is finished and
	// recorded first, so an upgrade can never cost a measurement. An
	// implementation does not return on success; a return of any kind ends the
	// loop (with ErrUpgraded when there was no error), and an error is logged and
	// the survey carries on with the build it has.
	//
	// The record is unaffected either way. It is append-only on disk, every append
	// is synced before it returns, and the new process opens the same path — the
	// streak is a property of the file, not of the process.
	ReExec func() error

	// Progress receives a human-readable line per cycle. Nil is silent.
	Progress io.Writer
}

// ErrUpgraded ends a run because the binary under it was replaced and the process
// handed over to the new one. Everything recorded so far is on disk and the
// successor appends to the same file.
var ErrUpgraded = errors.New("soak: the installed binary changed and this process handed over to it")

// Runner executes the survey.
type Runner struct {
	opts Options

	// lastTokenExpiry is the previous cycle's observation, held so a forward
	// move can be recognised as a refresh.
	lastTokenExpiry time.Time
	haveTokenExpiry bool

	// startedWith is the executable this process was loaded from. Every cycle is
	// stamped with it and every boundary is compared against it.
	startedWith binstamp.Stamp
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
	r := &Runner{opts: opts}
	if opts.Binary != nil {
		// A fingerprint that cannot be taken is not a reason to refuse to survey.
		// It leaves the stamp zero, which every reader treats as "not known".
		r.startedWith, _ = opts.Binary()
	}
	return r, nil
}

// Run executes cycles until the count is reached or the context is cancelled.
//
// A cancellation between cycles returns ctx.Err() with everything already
// recorded still on disk — the soak is a multi-day process and Ctrl-C is a
// normal way to end it.
func (r *Runner) Run(ctx context.Context) error {
	for i := 0; r.opts.Cycles == 0 || i < r.opts.Cycles; i++ {
		if err := r.awaitBudget(ctx); err != nil {
			return err
		}
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

		// The cycle boundary, and the only place an upgrade may happen: the cycle
		// above is finished, recorded and synced, so handing over here cannot cost
		// a measurement.
		if err := r.upgradeIfReplaced(); err != nil {
			return err
		}

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

// upgradeIfReplaced hands over to a newly installed binary, at a cycle boundary.
//
// Everything about it fails towards "keep surveying". A fingerprint that cannot be
// taken, a binary that has not moved, a re-exec that refuses — each of them returns
// nil and the loop continues with the process it has. The survey's job is to run
// for days; being one build behind is a note on a dashboard, and stopping would be
// a lost day.
func (r *Runner) upgradeIfReplaced() error {
	if r.opts.Binary == nil || r.opts.ReExec == nil || !r.startedWith.Known() {
		return nil
	}
	installed, err := r.opts.Binary()
	if err != nil {
		r.writeLine("설치된 바이너리를 확인할 수 없다 (%v) — 이대로 계속한다", err)
		return nil
	}
	if r.startedWith.Same(installed) {
		return nil
	}

	r.writeLine("설치된 바이너리가 바뀌었다 (%s → %s). 사이클 경계에서 새 바이너리로 자기 재실행한다 — "+
		"기록 파일은 그대로 이어진다", r.startedWith.Describe(), installed.Describe())
	if err := r.opts.ReExec(); err != nil {
		r.writeLine("자기 재실행 실패 (%v) — 기존 바이너리로 서베이를 계속한다", err)
		return nil
	}
	// A real re-exec never gets here. A seam that does is telling us the handover
	// happened in a way this process cannot observe, so the loop ends.
	return ErrUpgraded
}

// PausePoll is how often a paused survey asks again.
//
// Half a minute: short enough that the soak resumes promptly once the
// verification finishes, long enough that a paused survey is not a spin loop with
// a timer on it.
const PausePoll = 30 * time.Second

// awaitBudget delays a cycle while something else needs the rate budget.
//
// It logs the pause and the resumption once each rather than per poll, because a
// soak runs for days into a file somebody reads later and a wall of identical
// lines is the same as no line at all.
func (r *Runner) awaitBudget(ctx context.Context) error {
	if r.opts.PauseWhile == nil {
		return ctx.Err()
	}
	announced := false
	for {
		paused, reason := r.opts.PauseWhile()
		if !paused {
			if announced {
				r.writeLine("resuming — %s", orUnstated(reason, "the pause has cleared"))
			}
			return ctx.Err()
		}
		if !announced {
			r.writeLine("paused — %s", orUnstated(reason, "something else needs this account's rate budget"))
			announced = true
		}
		if err := r.opts.Clock.Sleep(ctx, PausePoll); err != nil {
			return err
		}
	}
}

func (r *Runner) writeLine(format string, a ...any) {
	if r.opts.Progress == nil {
		return
	}
	fmt.Fprintf(r.opts.Progress, "%s  %s\n",
		r.opts.Clock.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

func orUnstated(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
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
		Binary:        r.startedWith,
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

	// 3. orders: the closed group, then the open one, then one order by id out of
	//    the two together.
	closedWalk, openWalk, ordersResult := r.probeOrders(ctx)
	cycle.Endpoints = append(cycle.Endpoints, ordersResult)
	cycle.Endpoints = append(cycle.Endpoints, r.probeOrderByID(ctx, closedWalk, openWalk))

	// 4. quotes.
	quotes, pricesResult := r.probePrices(ctx)
	cycle.Endpoints = append(cycle.Endpoints, pricesResult)

	cycle.Completeness = completenessOf(closedWalk, openWalk, ordersResult.OK, positions, len(r.opts.Symbols), quotes)
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

// probeOrders walks the closed group and then the open one. Between them they
// are the whole order list; there is no third call that would return it in one
// go (see the status constants above).
//
// Both walks are recorded against GET /api/v1/orders, because that is the
// endpoint an attestation would name, and one failed walk fails the endpoint:
// half the list is not a read of the list.
func (r *Runner) probeOrders(ctx context.Context) (closed, open orderWalk, result EndpointResult) {
	started := r.opts.Clock.Now()

	closed = r.walkOrders(ctx, statusClosed)
	open = r.walkOrders(ctx, statusOpen)

	result = EndpointResult{
		Endpoint:  EndpointOrders,
		Requests:  closed.requests + open.requests,
		LatencyMS: r.opts.Clock.Since(started).Milliseconds(),
	}
	switch {
	case closed.err != nil:
		result.Class = r.classify(closed.err)
		result.Error = closed.err.Error()
	case open.err != nil:
		result.Class = r.classify(open.err)
		result.Error = open.err.Error()
	default:
		result.OK = true
		result.Class = ClassOK
	}
	return closed, open, result
}

// walkOrders follows the cursor to the end of one status group, or to the first
// sign that the cursor cannot be followed.
//
// It is written the same way for both groups on purpose. The spec says CLOSED
// paginates and OPEN does not, and the way to find out whether that is true is
// to follow whatever the broker hands back rather than to assume it: an OPEN
// response that claims a next page gets walked, and if the cursor is ignored the
// repeat lands in the loop and duplicate detection below.
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

// probeOrderByID reads the first order the two walks found between them.
//
// Closed orders come first only because a live account is far likelier to have
// history than to have something working right now; either group is an equally
// good id to prove the endpoint with, and an account with nothing in the closed
// group falls through to the open one.
//
// With no orders at all there is nothing to read, and the result is a recorded
// skip rather than a success — Evaluate then refuses to complete the soak until
// the endpoint has actually been exercised, which is the honest answer: the
// engine calls it, and nobody has seen it work.
func (r *Runner) probeOrderByID(ctx context.Context, closed, open orderWalk) EndpointResult {
	ids := append(append([]string(nil), closed.ids...), open.ids...)
	if len(ids) == 0 {
		reason := "the account has no orders, so there is no id to read"
		if closed.err != nil || open.err != nil {
			reason = "the order list could not be read, so no id was available"
		}
		return EndpointResult{Endpoint: EndpointOrderByID, Skipped: true, SkipReason: reason}
	}
	id := ids[0]
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
//
// It takes the two walks apart rather than one merged list because the groups
// overlap by design: the spec puts PARTIAL_FILLED in both OPEN and CLOSED, so one
// order can honestly come back from both walks. Merging first and then counting
// repeats would turn that documented overlap into a duplicate-id failure and
// block the attestation for something the broker is entitled to do — the same
// artefact execgw.Resolver.scanBoth removes when it asks the same question.
func completenessOf(closed, open orderWalk, ordersOK bool, positions, symbolsAsked, quotes int) Completeness {
	walks := []orderWalk{closed, open}
	c := Completeness{
		OrderPages:          closed.pages + open.pages,
		OrderIDs:            len(closed.ids) + len(open.ids),
		Positions:           positions,
		OpenOrders:          len(open.ids),
		QuotesRequested:     symbolsAsked,
		QuotesReturned:      quotes,
		CursorLoop:          closed.cursorLoop || open.cursorLoop,
		TruncatedPagination: closed.truncated || open.truncated,
		PageLimitReached:    closed.pageLimit || open.pageLimit,
	}
	if !ordersOK {
		// A walk did not finish, so there is nothing to conclude. The failure is
		// already recorded against the endpoint; recording it a second time as a
		// completeness failure would let one dropped connection block the
		// attestation for the rest of the window.
		c.Detail = "not evaluated: the order list could not be read"
		return c
	}
	c.Evaluated = true

	var problems []string
	// Duplicates are counted inside each walk, never across the two. A repeat
	// within one walk means the cursor handed the same order out twice and the
	// walk cannot be trusted to have seen every order exactly once; a repeat
	// across the walks is the OPEN/CLOSED overlap, which is counted separately
	// and reported rather than failed.
	//
	// The empty id is excluded from both counts. Two entries with no identifier
	// are two orders nobody can name, not one order seen twice, and calling them
	// a duplicate would describe the wrong fault; the identifier check below is
	// where they land.
	for _, w := range walks {
		for id, n := range w.seen {
			if id != "" && n > 1 {
				c.DuplicateOrderIDs++
				if len(problems) < 4 {
					problems = append(problems, fmt.Sprintf("order %s appeared %d times in one status group", id, n))
				}
			}
		}
	}
	for id := range open.seen {
		if id != "" && closed.seen[id] > 0 {
			c.OrdersInBothStatuses++
		}
	}

	// This is what is left of the open-order coverage check.
	//
	// That check asked whether every open order was also present in the unfiltered
	// list. There is no unfiltered list to ask about any more, and the list is now
	// *defined* as CLOSED ∪ OPEN, so the question can only ever answer yes: keeping
	// it would be reporting a check that cannot fail. What it could genuinely
	// catch is an entry the broker returned with no identifier on it — the case
	// cmd/tossctl's adapter deliberately surfaces as an empty id rather than
	// dropping the page — because an order nothing can name is an order no
	// reconciliation can ever match against local state. That is checked here, over
	// both groups, and it does fail.
	for _, w := range walks {
		c.BlankOrderIDs += w.seen[""]
	}
	if c.BlankOrderIDs > 0 {
		problems = append(problems, fmt.Sprintf(
			"%d order(s) came back with no identifier on them", c.BlankOrderIDs))
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
	if closed.pageLimit {
		problems = append(problems, fmt.Sprintf(
			"the CLOSED walk stopped at the %d-page limit, so the history behind it was not read", closed.pages))
	}
	if open.pageLimit {
		problems = append(problems, fmt.Sprintf(
			"the OPEN walk stopped at the %d-page limit, although the spec has that group return every "+
				"working order in one response", open.pages))
	}

	c.OK = c.DuplicateOrderIDs == 0 && c.BlankOrderIDs == 0 &&
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
