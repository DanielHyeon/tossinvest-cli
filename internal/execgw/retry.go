package execgw

// retry.go implements the retry matrix published as
// openspec/changes/harden-execution-base/retry-matrix.md. The two are meant to be
// changed together — TestDefaultsMatchThePublishedMatrix fails if they drift.
//
// The single most important thing in this file is what it does NOT contain:
// there is no entry point that retries a mutation. The official API has no
// idempotency key, so a retried order is a duplicated order. Ambiguity about a
// mutation is resolved by the IN_DOUBT procedure, never by trying again.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// ErrorClass is how a query failure is treated by the matrix.
type ErrorClass string

const (
	// ClassOK is no error.
	ClassOK ErrorClass = "ok"
	// ClassTransient is retryable: transport failures and 5xx. An unrecognised
	// error is treated as transient too — for a *query* that is the safe
	// direction, because the alternative is giving up on data the engine needs.
	ClassTransient ErrorClass = "transient"
	// ClassRateLimited is 429: retryable only within the Retry-After cap.
	ClassRateLimited ErrorClass = "rate_limited"
	// ClassAuthFatal is 401/403: never retried, and it latches the entry gate.
	ClassAuthFatal ErrorClass = "auth_fatal"
	// ClassPermanent is a 4xx describing the request. Retrying cannot change it.
	ClassPermanent ErrorClass = "permanent"
	// ClassCanceled is the caller's own cancellation or deadline.
	ClassCanceled ErrorClass = "canceled"
)

// ClassifyQueryError maps a read error onto the matrix.
func ClassifyQueryError(err error) ErrorClass {
	switch {
	case err == nil:
		return ClassOK
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return ClassCanceled
	case errors.Is(err, official.ErrAuth), errors.Is(err, official.ErrIPNotAllowed):
		return ClassAuthFatal
	case errors.Is(err, official.ErrRateLimited):
		return ClassRateLimited
	case errors.Is(err, official.ErrServer), errors.Is(err, official.ErrTransport):
		return ClassTransient
	}
	var apiErr *official.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Code == http.StatusTooManyRequests:
			return ClassRateLimited
		case apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusForbidden:
			return ClassAuthFatal
		case apiErr.Code >= 500:
			return ClassTransient
		default:
			return ClassPermanent
		}
	}
	return ClassTransient
}

// RetryPolicy is the query half of the matrix. Mutations have no policy because
// they have no retries.
type RetryPolicy struct {
	// MaxAttempts counts the first call: 3 means one call and two retries.
	MaxAttempts int
	// Budget bounds the total wall time of all attempts plus waits.
	Budget time.Duration
	// BaseBackoff is the first wait; each retry doubles it up to MaxBackoff.
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	// MaxRetryAfter caps how long a 429's Retry-After can park a query. A header
	// beyond this aborts the query instead: blocking new entries is better than
	// freezing the engine inside a decision.
	MaxRetryAfter time.Duration
	// JitterFraction is the bounded ± spread applied to every wait, to stop
	// concurrent pollers resynchronising after an outage.
	JitterFraction float64
	// Rand returns a value in [0,1). Tests inject a constant for exact waits.
	Rand func() float64
}

// DefaultRetryPolicy is the conservative provisional default from the matrix.
// The numbers are pending real measurement in verify-execution-capability.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		Budget:         8 * time.Second,
		BaseBackoff:    400 * time.Millisecond,
		MaxBackoff:     3 * time.Second,
		MaxRetryAfter:  30 * time.Second,
		JitterFraction: 0.25,
	}
}

// BackoffFor returns the wait before retry number attempt (1-based) for a base
// delay, jitter included and bounded by MaxBackoff.
func (p RetryPolicy) BackoffFor(attempt int, base time.Duration) time.Duration {
	if base <= 0 {
		base = p.BaseBackoff
	}
	delay := float64(base) * math.Pow(2, float64(attempt-1))
	if p.MaxBackoff > 0 && delay > float64(p.MaxBackoff) {
		delay = float64(p.MaxBackoff)
	}
	return p.jitter(time.Duration(delay))
}

// jitter spreads d by ±JitterFraction. It is multiplicative and clamped, so the
// result is always within [d(1-f), d(1+f)] — jitter must not be able to invent an
// unbounded wait.
func (p RetryPolicy) jitter(d time.Duration) time.Duration {
	if p.JitterFraction <= 0 || d <= 0 {
		return d
	}
	r := p.Rand
	if r == nil {
		r = rand.Float64
	}
	// (r-0.5)*2 ∈ [-1,1)
	offset := (r() - 0.5) * 2 * p.JitterFraction
	out := time.Duration(float64(d) * (1 + offset))
	lo := time.Duration(float64(d) * (1 - p.JitterFraction))
	hi := time.Duration(float64(d) * (1 + p.JitterFraction))
	switch {
	case out < lo:
		return lo
	case out > hi:
		return hi
	default:
		return out
	}
}

// RequiredQuery names a read whose freshness gates new entries.
type RequiredQuery string

const (
	QueryOpenOrders  RequiredQuery = "open_orders"
	QueryBuyingPower RequiredQuery = "buying_power"
	QueryHoldings    RequiredQuery = "holdings"
	QueryPrice       RequiredQuery = "price"
)

// DefaultStaleness is the provisional staleness table from the matrix.
func DefaultStaleness() map[RequiredQuery]time.Duration {
	return map[RequiredQuery]time.Duration{
		QueryOpenOrders:  20 * time.Second,
		QueryBuyingPower: 45 * time.Second,
		QueryHoldings:    60 * time.Second,
		QueryPrice:       15 * time.Second,
	}
}

// RetryAfterSource supplies the most recent Retry-After the transport observed.
// Reading it consumes it: a stale header must not govern an unrelated later call.
type RetryAfterSource interface {
	RetryAfter() (time.Duration, bool)
}

// RetryAfterTransport captures Retry-After from 429/503 responses.
//
// It exists because internal/official maps a 429 onto a sentinel error and drops
// the response headers, so the only place the header survives is the round
// tripper. Engine wiring installs this on the official client's http.Client.
type RetryAfterTransport struct {
	base http.RoundTripper
	clk  clock.Clock

	mu    sync.Mutex
	delay time.Duration
	seen  bool
}

// NewRetryAfterTransport wraps base. A nil base uses http.DefaultTransport.
func NewRetryAfterTransport(base http.RoundTripper, clk clock.Clock) *RetryAfterTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if clk == nil {
		clk = clock.System()
	}
	return &RetryAfterTransport{base: base, clk: clk}
}

// RoundTrip forwards the request and records any Retry-After it comes back with.
func (t *RetryAfterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
		return resp, nil
	}
	if d, ok := ParseRetryAfter(resp.Header.Get("Retry-After"), t.clk.Now()); ok {
		t.mu.Lock()
		t.delay = d
		t.seen = true
		t.mu.Unlock()
	}
	return resp, nil
}

// RetryAfter returns and consumes the last observed delay.
func (t *RetryAfterTransport) RetryAfter() (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.seen {
		return 0, false
	}
	d := t.delay
	t.seen = false
	t.delay = 0
	return d, true
}

// ParseRetryAfter reads both header forms: delta-seconds and an HTTP date. An
// unparseable or past value reports false, which makes the caller fall back to
// its own backoff rather than to zero.
func ParseRetryAfter(header string, now time.Time) (time.Duration, bool) {
	if header == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(header); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if when, err := http.ParseTime(header); err == nil {
		d := when.Sub(now)
		if d <= 0 {
			return 0, false
		}
		return d.Round(time.Second), true
	}
	return 0, false
}

// Retrier runs required queries under the matrix and reports their outcome to the
// entry gate.
type Retrier struct {
	Policy RetryPolicy
	Clock  clock.Clock
	// Gate is notified of successes (freshness) and auth failures (latch).
	// Optional; nil disables entry-gate bookkeeping.
	Gate *EntryGate
	// RetryAfter supplies the broker's 429 hint. Optional.
	RetryAfter RetryAfterSource
}

// Query performs a read under the retry budget.
//
// The bookkeeping is as important as the retrying: a success stamps the query's
// freshness (which is what auto-clears a staleness block) and a 401/403 latches
// the entry gate before the error is even returned.
func (r *Retrier) Query(ctx context.Context, kind RequiredQuery, fn func(context.Context) error) error {
	policy := r.Policy
	if policy.MaxAttempts <= 0 {
		policy = DefaultRetryPolicy()
	}
	clk := r.Clock
	if clk == nil {
		clk = clock.System()
	}
	deadline := clk.Now().Add(policy.Budget)

	var lastErr error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		err := fn(ctx)
		lastErr = err
		switch ClassifyQueryError(err) {
		case ClassOK:
			if r.Gate != nil {
				r.Gate.RecordSuccess(kind)
			}
			return nil

		case ClassAuthFatal:
			// Immediate, no retry: a rejected credential does not improve by
			// being presented again, and a live engine with a dead credential
			// must stop opening positions right now.
			if r.Gate != nil {
				r.Gate.Block(ReasonBrokerAuthRejected,
					fmt.Sprintf("%s was refused by the broker: %v", kind, err))
			}
			return err

		case ClassPermanent, ClassCanceled:
			return err

		case ClassRateLimited:
			wait, ok := r.rateLimitWait(policy, attempt)
			if !ok {
				return fmt.Errorf("%w (Retry-After beyond the %s cap)", err, policy.MaxRetryAfter)
			}
			if !r.sleepWithin(ctx, clk, wait, deadline) {
				return err
			}

		default: // ClassTransient
			if attempt == policy.MaxAttempts {
				return err
			}
			if !r.sleepWithin(ctx, clk, policy.BackoffFor(attempt, policy.BaseBackoff), deadline) {
				return err
			}
		}
	}
	return lastErr
}

// rateLimitWait honours the broker's hint when there is one, capped; otherwise it
// falls back to the ordinary backoff. A hint beyond the cap reports false, which
// aborts the query.
func (r *Retrier) rateLimitWait(policy RetryPolicy, attempt int) (time.Duration, bool) {
	if r.RetryAfter != nil {
		if d, ok := r.RetryAfter.RetryAfter(); ok {
			if policy.MaxRetryAfter > 0 && d > policy.MaxRetryAfter {
				return 0, false
			}
			return d, true
		}
	}
	return policy.BackoffFor(attempt, policy.BaseBackoff), true
}

// sleepWithin waits, unless the wait would spend more than the remaining budget.
// Reporting false ends the retry loop with the error already in hand.
func (r *Retrier) sleepWithin(ctx context.Context, clk clock.Clock, d time.Duration, deadline time.Time) bool {
	if d <= 0 {
		return true
	}
	if clk.Now().Add(d).After(deadline) {
		return false
	}
	return clk.Sleep(ctx, d) == nil
}

// EntryGate decides whether new exposure may be opened.
//
// Two kinds of block live here and they behave differently on purpose:
//
//   - staleness: derived from the last successful poll of each required query, so
//     it clears itself the moment the data is fresh again (the spec's "조회 복구 후
//     자동 해제").
//   - latches: set by an event that a successful poll does not undo — a rejected
//     credential, an unresolved IN_DOUBT. Only an explicit Clear reopens those.
//
// Blocks apply to new entries only. Cancels and liquidations are never gated:
// trapping the engine in a position it cannot exit is a worse failure than the
// one being guarded against (§0.3).
type EntryGate struct {
	clk clock.Clock

	mu         sync.Mutex
	thresholds map[RequiredQuery]time.Duration
	lastOK     map[RequiredQuery]time.Time
	latches    map[ReasonCode]string
	// symbolLatches are the narrower, per-symbol blocks (symbolgate.go, task
	// 4.2). Lazily created, so NewEntryGate is unchanged.
	symbolLatches map[string]SymbolBlock
}

// NewEntryGate builds a gate. A nil threshold map uses DefaultStaleness(); an
// empty (non-nil) map means "no staleness requirements", which is what unit tests
// of the latch behaviour want.
func NewEntryGate(clk clock.Clock, thresholds map[RequiredQuery]time.Duration) *EntryGate {
	if clk == nil {
		clk = clock.System()
	}
	if thresholds == nil {
		thresholds = DefaultStaleness()
	}
	copied := make(map[RequiredQuery]time.Duration, len(thresholds))
	for k, v := range thresholds {
		copied[k] = v
	}
	return &EntryGate{
		clk:        clk,
		thresholds: copied,
		lastOK:     make(map[RequiredQuery]time.Time, len(copied)),
		latches:    make(map[ReasonCode]string),
	}
}

// RecordSuccess stamps a required query as fresh.
func (g *EntryGate) RecordSuccess(kind RequiredQuery) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lastOK[kind] = g.clk.Now()
}

// Block latches the gate shut with a reason.
func (g *EntryGate) Block(reason ReasonCode, detail string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.latches[reason]; !exists {
		g.latches[reason] = detail
	}
}

// Clear releases a latch. It is the operator action the spec requires for auth
// failures and unresolved doubt; nothing automatic calls it.
func (g *EntryGate) Clear(reason ReasonCode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.latches, reason)
}

// CheckEntry reports why new exposure is refused anywhere on the account, or nil
// when it is allowed.
//
// It answers the account-wide question only. Per-symbol blocks are deliberately
// not consulted: a symbol that is blocked is not a reason to stop trading the
// rest of the account, and conflating the two is what made this gate wider than
// the reconciliation state table said it should be (issues.md, 2026-07-26).
// CheckEntryFor is the symbol-aware question, and it is what the gateway asks.
func (g *EntryGate) CheckEntry() *RejectedError {
	return g.CheckEntryFor("", "")
}

// checkAccountEntry holds the account-wide half. Callers must not hold g.mu.
func (g *EntryGate) checkAccountEntry() *RejectedError {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Latches first: they are the deliberate, operator-visible stops.
	for _, reason := range latchOrder {
		if detail, blocked := g.latches[reason]; blocked {
			return &RejectedError{Reason: reason, Detail: detail}
		}
	}
	for reason, detail := range g.latches {
		return &RejectedError{Reason: reason, Detail: detail}
	}

	now := g.clk.Now()
	for _, kind := range staleOrder {
		threshold, required := g.thresholds[kind]
		if !required {
			continue
		}
		last, seen := g.lastOK[kind]
		if !seen {
			return reject(ReasonQueryStale,
				"%s has never been observed successfully; new entries stay blocked until it is", kind)
		}
		if age := now.Sub(last); age > threshold {
			return reject(ReasonQueryStale,
				"%s is %s old, past the %s threshold", kind, age.Round(time.Second), threshold)
		}
	}
	return nil
}

// Blocks lists every active latch, for status output and alerting.
func (g *EntryGate) Blocks() map[ReasonCode]string {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make(map[ReasonCode]string, len(g.latches))
	for k, v := range g.latches {
		out[k] = v
	}
	return out
}

// latchOrder and staleOrder make CheckEntry deterministic: map iteration order
// would otherwise make the reported reason arbitrary when several apply, and an
// operator chasing a flapping reason code is an operator who stops trusting it.
// New reasons are appended, never inserted: the relative precedence of the
// original three is itself a contract with the operator runbooks.
var latchOrder = []ReasonCode{
	ReasonBrokerAuthRejected,
	ReasonUnresolvedInDoubt,
	ReasonAlertUndelivered,
	ReasonRecoveryIncomplete,
	ReasonReconcilePermanent,
	ReasonReconcileMismatch,
	ReasonBrokerStateUnknown,
	ReasonFillDetectionSLO,
}

var staleOrder = []RequiredQuery{
	QueryOpenOrders,
	QueryBuyingPower,
	QueryHoldings,
	QueryPrice,
}
