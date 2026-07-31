package official

// ratebudget.go keeps what a response said about the caller's remaining rate
// allowance.
//
// # Why this exists
//
// Every documented endpoint group has a per-second limit, and the responses carry
// the current budget in X-RateLimit-Limit / -Remaining / -Reset. Until this file,
// nothing in TossOS read those headers: doRequest returned (status, body) and the
// headers went out of scope, and a 429 was mapped onto ErrRateLimited with the
// response discarded. internal/execgw's RetryAfterTransport says the same thing
// about Retry-After and exists for exactly that reason.
//
// The consequence is that the budget was only ever learned by exhausting it.
// That is late for any caller, and it is late by a whole data source for
// discovery's official ranking, which — unlike the price reads — has no WTS
// fallback to absorb the 429.
//
// # Additive, deliberately
//
// Nothing here changes a signature, a return value or an error. Callers that do
// not ask keep behaving exactly as before, which is what makes this safe to add
// to a package the engine and the live-verification path both depend on.
//
// # The Reset header is not guessed
//
// Rate-limit resets are spelled two ways in the wild — epoch seconds, or seconds
// from now — and the Toss documentation does not say which. Rather than pick one
// and be quietly wrong, this keeps the raw value alongside a best-effort parse and
// says which reading it took. The point of recording these headers at all is to
// replace an assumption with a measurement; starting with a new assumption would
// defeat it.

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PathRankings is the rankings endpoint's request path.
//
// Exported because the budget map is keyed by path, so a caller asking for the
// rankings budget has to name the same string this package requested with. A
// second copy that drifted would make the lookup miss, and a miss is reported as
// "no headers were sent" — the measurement would go dark without failing.
const PathRankings = "/api/v1/rankings"

// Rate-limit header names, as the Toss Open API sends them.
const (
	HeaderRateLimit     = "X-RateLimit-Limit"
	HeaderRateRemaining = "X-RateLimit-Remaining"
	HeaderRateReset     = "X-RateLimit-Reset"
)

// ResetKind is how the X-RateLimit-Reset value was read.
type ResetKind string

const (
	// ResetAbsent: the header was not sent.
	ResetAbsent ResetKind = ""
	// ResetEpoch: read as Unix seconds.
	ResetEpoch ResetKind = "epoch"
	// ResetDelta: read as seconds from the moment of the response.
	ResetDelta ResetKind = "delta"
	// ResetUnparsed: the header was sent and could not be read as either.
	ResetUnparsed ResetKind = "unparsed"
)

// resetEpochThreshold is where a reset value stops being plausible as a delay and
// starts being plausible as a timestamp. 10^9 seconds is 2001-09-09, so any real
// epoch is above it and no sane retry delay is.
const resetEpochThreshold = 1_000_000_000

// resetMaxAhead and resetMaxBehind bound a parsed reset to something a rate-limit
// window could plausibly be.
//
// Without them the heuristic produced confident nonsense: epoch *milliseconds* —
// a very common encoding — read as epoch seconds and landed in the year 58541,
// and a nanosecond delay read as an epoch landed in 2001. Both were labelled
// ResetEpoch, i.e. presented as a determination. A caller waiting for a reset in
// 2001 does not wait at all and one waiting for 58541 never stops.
//
// Out-of-range values are ResetUnparsed with the raw string kept. This file
// exists to replace an assumption with a measurement; an unbounded parse smuggles
// a new assumption back in.
const (
	resetMaxAhead  = 24 * time.Hour
	resetMaxBehind = time.Minute
)

// budgetKey collapses the identifying segments of a path so the map is keyed by
// endpoint rather than by instance.
//
// Order cancels, order modifies and per-symbol candles all embed an id in the
// path. Keyed raw, a long-lived engine client accumulates one entry per order id
// for the life of the process, and — worse for the purpose this serves — the
// candle budget is split across every symbol, so the per-group allowance the
// server actually meters is never observable. The limits are per group; the path
// template is the closest thing to a group that is visible from here.
func budgetKey(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if identifierSegment(seg) {
			segments[i] = "{id}"
		}
	}
	return strings.Join(segments, "/")
}

// identifierSegment reports a path segment that names one object rather than one
// endpoint. Endpoint segments in this API are lowercase words and hyphens; ids
// and symbols carry digits or uppercase.
func identifierSegment(seg string) bool {
	if seg == "" || seg == "api" {
		return false
	}
	if len(seg) > 1 && seg[0] == 'v' && allDigits(seg[1:]) {
		return false // version segment
	}
	for _, r := range seg {
		if r >= '0' && r <= '9' {
			return true
		}
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// RateBudget is one endpoint's remaining allowance, as its last response
// reported it.
type RateBudget struct {
	// Path is the request path this budget was observed on, with identifying
	// segments collapsed to {id} — see budgetKey. The limits are per group and the
	// group is not in the response, so the path template is the closest visible
	// approximation; the raw path would be finer than the thing being metered and
	// would grow the map without bound.
	Path string
	// Limit and Remaining are the window's size and what is left. Meaningful only
	// when Reported is true.
	Limit, Remaining int
	// Reset is the best-effort instant the window rolls over. ResetRaw is the
	// trimmed header value, kept because the encoding is undocumented and a later
	// reader may be able to tell what this one could not.
	Reset    time.Time
	ResetRaw string
	// ResetKind says how Reset was derived, including that it could not be.
	ResetKind ResetKind
	// ObservedAt is when the response arrived.
	ObservedAt time.Time
	// Reported is false when the response carried no rate headers at all.
	//
	// It is the absent-versus-zero flag and it is load-bearing: a caller that
	// read a missing header as "0 remaining" would back off from an endpoint that
	// never asked it to.
	Reported bool
}

// Exhausted reports that the allowance is spent. An unreported budget is never
// exhausted — not because it is known to be fine, but because nothing was said.
func (b RateBudget) Exhausted() bool { return b.Reported && b.Remaining <= 0 }

// Tight reports that the remaining allowance is at or below floor.
func (b RateBudget) Tight(floor int) bool { return b.Reported && b.Remaining <= floor }

// rateBudgets is the per-path store on a Client.
type rateBudgets struct {
	mu sync.RWMutex
	by map[string]RateBudget
}

func (r *rateBudgets) record(b RateBudget) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.by == nil {
		r.by = make(map[string]RateBudget, 8)
	}
	r.by[b.Path] = b
}

func (r *rateBudgets) get(path string) RateBudget {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.by[path]
}

func (r *rateBudgets) all() map[string]RateBudget {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]RateBudget, len(r.by))
	for k, v := range r.by {
		out[k] = v
	}
	return out
}

// RateBudget returns the last observed allowance for a request path.
//
// A path never seen, or seen without rate headers, comes back with Reported
// false. Callers must not read that as zero remaining.
func (c *Client) RateBudget(path string) RateBudget { return c.rates.get(budgetKey(path)) }

// RateBudgets returns a copy of every observed allowance, keyed by request path.
func (c *Client) RateBudgets() map[string]RateBudget { return c.rates.all() }

// readRateBudget extracts the budget from a response.
//
// It is called for every response, including error ones: a 429 is the single most
// informative moment for this reading, and it is exactly the one the error
// mapping used to throw away.
func readRateBudget(path string, header http.Header, now time.Time) RateBudget {
	b := RateBudget{Path: budgetKey(path), ObservedAt: now.UTC()}

	limit, hasLimit := headerInt(header, HeaderRateLimit)
	remaining, hasRemaining := headerInt(header, HeaderRateRemaining)
	raw, reset, resetKind := ParseRateBudgetReset(header.Get(HeaderRateReset), now)

	if !hasLimit && !hasRemaining && raw == "" {
		return b
	}
	b.Reported = true
	b.Limit, b.Remaining = limit, remaining
	b.ResetRaw = raw
	b.Reset = reset
	b.ResetKind = resetKind
	return b
}

// ParseRateBudgetReset applies the official client's exact reset-header
// semantics without mutating client state. The canonical raw value is trimmed
// exactly as readRateBudget stores it, allowing downstream admission code to
// validate a RateBudget without copying the threshold, plausibility bounds, or
// overflow-sensitive delta arithmetic.
func ParseRateBudgetReset(raw string, observedAt time.Time) (canonicalRaw string, reset time.Time, kind ResetKind) {
	canonicalRaw = strings.TrimSpace(raw)
	if canonicalRaw == "" {
		return canonicalRaw, time.Time{}, ResetAbsent
	}
	seconds, err := strconv.ParseInt(canonicalRaw, 10, 64)
	if err != nil || seconds < 0 || observedAt.IsZero() {
		return canonicalRaw, time.Time{}, ResetUnparsed
	}
	if seconds >= resetEpochThreshold {
		reset, kind = time.Unix(seconds, 0).UTC(), ResetEpoch
	} else {
		var ok bool
		reset, ok = addResetDelta(observedAt, seconds)
		if !ok {
			return canonicalRaw, time.Time{}, ResetUnparsed
		}
		kind = ResetDelta
	}
	if !plausibleReset(reset, observedAt) {
		return canonicalRaw, time.Time{}, ResetUnparsed
	}
	return canonicalRaw, reset, kind
}

func addResetDelta(observedAt time.Time, seconds int64) (time.Time, bool) {
	const maxDurationSeconds = int64((time.Duration(1<<63 - 1)) / time.Second)
	if seconds < 0 || seconds > maxDurationSeconds {
		return time.Time{}, false
	}
	base := observedAt.UTC()
	delta := time.Duration(seconds) * time.Second
	reset := base.Add(delta)
	if reset.Sub(base) != delta {
		return time.Time{}, false
	}
	return reset, true
}

func headerInt(header http.Header, name string) (int, bool) {
	raw := strings.TrimSpace(header.Get(name))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}

// plausibleReset bounds a parsed reset to a window a rate limit could occupy.
func plausibleReset(reset, now time.Time) bool {
	if reset.IsZero() || now.IsZero() {
		return false
	}
	delta := reset.Sub(now.UTC())
	return delta >= -resetMaxBehind && delta <= resetMaxAhead
}
