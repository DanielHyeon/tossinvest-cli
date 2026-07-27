package candidate

// source.go is the seam between discovery and the market data it reads.
//
// # Why this package declares its own row type
//
// The obvious move is to take domain.RankingItem, which the official adapter
// already produces. It cannot be used here, for a reason that is specific to
// what this package claims.
//
// internal/official's adaptRanking runs every field through parseDecimal into a
// float64, and a nullable field that arrives as "" becomes 0.0 — the adapter's
// own comment says so ("a null changeRate/rankedAt arrives as \"\" and maps to
// 0"). That is a reasonable choice for a CLI that prints a table. It is the wrong
// input for a package whose central rule is that absent and zero are different
// states, because by the time a domain.RankingItem exists the distinction is
// already gone and no adapter downstream can recover it.
//
// So Row carries decimals as strings, empty means the source did not report the
// field, and the conversion happens at the wiring boundary where the caller still
// knows which it was. Where the distinction is genuinely load-bearing — the
// intraday high behind the near_high veto — the value does not come from a
// ranking at all, it comes from a candle, and its absence is the absence of the
// whole reading rather than of a field.
//
// # Why the interface is here and the implementations are not
//
// Same reason the operator console declares HoldingsReader instead of accepting
// verifylive.Broker. A Source can be asked for rows and nothing else; it cannot
// place, amend or cancel anything, because the type it is handed has no way to
// say those words. cmd/tossctl supplies the implementations, and this package
// imports no client of any kind — which is what makes the isolation a property of
// the types rather than of what the handlers happen to call.

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrRateLimited is how a source reports a 429.
//
// It is a distinct sentinel because a rate-limited loss and an ordinary failure
// mean different things about what to do next, and — for the official ranking,
// which has no fallback — the rate at which it happens is the only measurement we
// have of an undocumented limit.
var ErrRateLimited = errors.New("candidate: source is rate limited")

// ErrNoSourceAnswered means every source in the panel failed.
//
// Not an empty result: "we looked and the market is quiet" and "we could not
// look" are different facts, and a screen that renders them the same way teaches
// an operator to read a failure as calm.
var ErrNoSourceAnswered = errors.New("candidate: no source answered")

// ErrFallbackNotPossible means a reading claimed to have come through the hybrid
// WTS fallback from a source that has no fallback.
//
// hybrid's Rankings, MarketIndicatorPrices, MarketIndicatorCandles and
// MarketInvestorTrading are official-only — "no WTS fallback (WTS popularity
// ranking is a different dataset)", in the client's own words. A ranking reading
// marked as a fallback is therefore not a degradation that occurred; it is a
// wiring mistake, and accepting it would file the mistake as a plausible fact and
// then feed it into the fallback-rate measurement.
var ErrFallbackNotPossible = errors.New("candidate: this source has no WTS fallback")

// Row is one symbol as one source reported it, at the fidelity the source had.
//
// Every decimal is a string and empty means the source did not carry the field.
// See the file header for why this is not domain.RankingItem.
type Row struct {
	// Symbol identifies the instrument. Required.
	Symbol string
	// Rank is the 1-based position in the list that carried it, 0 when this
	// source is not a ranking. RankTotal is that list's length.
	Rank, RankTotal int
	// NewlyListed reports that this symbol was absent from the previous reading
	// of this list. Recorded as its own fact, never folded into a rank velocity.
	NewlyListed bool
	// TradingValue and TradingVolume are the cumulative intraday figures.
	TradingValue, TradingVolume string
	// Price is the last traded price.
	Price string
	// DayHigh and DayLow come from candles only. Empty is the normal case.
	DayHigh, DayLow string
}

// Budget is what a response said about the caller's remaining rate allowance.
//
// Reported is the absent-vs-zero flag, and it is load-bearing: WTS sends no such
// header at all, and storing that as "0 calls remaining" would make the scheduler
// retreat from a source that never asked it to. Unmeasured is not zero here for
// the same reason it is not zero in the veto.
type Budget struct {
	// Limit and Remaining are the window's size and what is left of it.
	Limit, Remaining int
	// Reset is when the window rolls over.
	Reset time.Time
	// Reported is true only when the source actually sent the headers.
	Reported bool
}

// Tight reports that the remaining allowance has fallen to or below floor.
//
// An unreported budget is never tight. That is the conservative answer for a
// scheduler — it declines to slow down on evidence it does not have — and it is
// paired with the fact that an unreported budget also cannot say we are safe.
// What decides the interval when nothing is reported is the configured floor,
// not this.
func (b Budget) Tight(floor int) bool { return b.Reported && b.Remaining <= floor }

// Reading is one source's answer to one market.
type Reading struct {
	// Rows are what the source returned, in its own order.
	Rows []Row
	// ViaFallback reports that the hybrid client served this from WTS because the
	// official endpoint was unavailable. Only meaningful for sources that have a
	// fallback — see ErrFallbackNotPossible.
	ViaFallback bool
	// Budget is the rate allowance the response reported, if any.
	Budget Budget
}

// Source is one place discovery reads from.
//
// One method, and it takes a market and returns rows. There is deliberately no
// way to express anything else: this is the interface an order would have to
// travel through, and it cannot carry one.
type Source interface {
	// ID names this source as it appears in a candidate's sources list.
	ID() SourceID
	// Read returns this source's rows for a market. A 429 is reported as
	// ErrRateLimited so the caller can count it separately from a plain failure.
	Read(ctx context.Context, market string) (Reading, error)
}

// hasFallback reports whether the hybrid client can serve this source from WTS.
//
// Measured against internal/hybrid on 2026-07-28, not assumed: Rankings,
// MarketIndicatorPrices, MarketIndicatorCandles and MarketInvestorTrading all
// return ErrOfficialKeyRequired rather than routing, and only the Prices-class
// reads go through route(). The first version of this change's design assumed the
// opposite and built a whole risk analysis on it.
func (s SourceID) hasFallback() bool { return s == SourceOfficialPrices }

// --- per-source scheduling (D13) -------------------------------------------------

// Interval is how often one source may be read.
//
// The interval belongs to the source rather than to the market because the limits
// differ by orders of magnitude — WTS has no limit anyone has observed, the
// official RANKING group is not in the published limit table at all, and candles
// are one call per symbol at five a second. One interval per market ties every
// source to the strictest one and throws away the earliness this change buys.
type Interval struct {
	// Every is the configured period.
	Every time.Duration
	// Floor is the shortest period this source may be read at, whatever Every
	// says. It is the rate-budget promise expressed as a number rather than as an
	// intention.
	Floor time.Duration
}

// engineYieldFactor is how much discovery slows while the engine is running.
//
// Doubling rather than stopping: the priority is engine > verification >
// discovery, and discovery is the only consumer whose lateness costs no money —
// but a discovery that stops entirely loses the first_seen_at resolution that is
// this change's whole product. The engine's own delay is the thing being
// protected, because it reaches stop-loss immediacy through fill detection.
const engineYieldFactor = 2

// unconfiguredFloor is the interval a source gets when nobody configured one.
//
// Not zero. A source absent from the interval map used to answer "due" on every
// tick with no floor and no engine yield, which is uncapped polling of an account
// budget shared with the engine — produced by a config typo, or by a panel
// gaining a source before its interval was written down. Section 3 adds candles
// at five calls a second per symbol, which is both the most expensive source and
// the likeliest to arrive that way.
//
// Fifteen seconds is the slowest configured interval in the contract, so an
// unconfigured source is treated as the most cautious known one rather than as
// unlimited.
const unconfiguredFloor = 15 * time.Second

// Schedule decides which sources are due.
//
// It is not a loop. Section 5's watch command owns the loop; this owns the answer
// to "may I read this source yet", so that the answer is testable without one.
//
// It is safe for concurrent use, and that is a requirement rather than a courtesy:
// YieldToEngine exists precisely because something outside the scan loop watches
// whether the engine is running, so the setter and the readers are different
// goroutines by construction. An unsynchronised map here is not a subtle race —
// it is Go's unrecoverable "concurrent map read and map write", which in a process
// that also runs the engine would take fill detection down with it.
type Schedule struct {
	mu        sync.Mutex
	intervals map[SourceID]Interval
	lastRun   map[SourceID]time.Time
	yielding  bool
}

// NewSchedule returns a schedule over the given per-source intervals.
func NewSchedule(intervals map[SourceID]Interval) *Schedule {
	copied := make(map[SourceID]Interval, len(intervals))
	for id, iv := range intervals {
		copied[id] = iv
	}
	return &Schedule{intervals: copied, lastRun: map[SourceID]time.Time{}}
}

// YieldToEngine tells the schedule whether the engine is currently running.
//
// spec Requirement 7 puts the engine first and, until the section-1 review, gave
// that ordering no mechanism — the verification runlock covered the middle
// priority and the top one had only a sentence. This is the verb.
func (s *Schedule) YieldToEngine(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.yielding = running
}

// Every is the effective period for a source: the configured value raised to the
// floor, and raised again while the engine is running.
func (s *Schedule) Every(id SourceID) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.every(id)
}

// every is Every without the lock, for callers that already hold it.
func (s *Schedule) every(id SourceID) time.Duration {
	iv, ok := s.intervals[id]
	if !ok {
		// Unknown, not unlimited. See unconfiguredFloor.
		iv = Interval{Every: unconfiguredFloor, Floor: unconfiguredFloor}
	}
	every := iv.Every
	if every < iv.Floor {
		every = iv.Floor
	}
	if s.yielding {
		every *= engineYieldFactor
	}
	return every
}

// Due reports that this source may be read at the given instant. A source that
// has never run is always due.
func (s *Schedule) Due(id SourceID, at time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.due(id, at)
}

func (s *Schedule) due(id SourceID, at time.Time) bool {
	last, ran := s.lastRun[id]
	if !ran {
		return true
	}
	return !at.Before(last.Add(s.every(id)))
}

// Ran records that a source was read.
func (s *Schedule) Ran(id SourceID, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastRun[id] = at.UTC()
}

// DueSources returns the subset of sources that may be read now, preserving the
// caller's order.
func (s *Schedule) DueSources(sources []Source, at time.Time) []Source {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Source, 0, len(sources))
	for _, src := range sources {
		if s.due(src.ID(), at) {
			out = append(out, src)
		}
	}
	return out
}
