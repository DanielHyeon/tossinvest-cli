// Package candidate is TossOS's discovery record: what the market's ranking
// sources said about a symbol, when they said it, and what we computed from the
// difference between two sayings.
//
// # It is a time axis, not a query
//
// Every source this package reads already exists and is already callable — the
// official rankings, prices and candles, the WTS popularity and investor lists.
// What none of them answer is "when did this start", and that is the only
// question that separates early discovery from chasing. `TOP_GAINERS` is by
// definition a list of things that have already gone up; using it directly as a
// buy list reproduces the exact failure it looks like a solution to.
//
// So the unit of storage is an observation with an instant on it, and the unit of
// meaning is a candidate that remembers the instant it first crossed. Both are
// persisted, and the persistence is not an optimisation:
//
//	An in-memory first_seen_at is reset by every restart. After the reset every
//	symbol was "just discovered", so the chase check answers "not late" for all
//	of them, and an early-discovery system becomes a late one that reports
//	itself as early. Nothing fails while this happens.
//
// # What this package cannot do
//
// It cannot order. It cannot size. It cannot stop out. Today that is true by
// inspection — `go list -deps ./internal/candidate` reaches internal/clock and
// nothing else — and it is enforced by nothing. Task 6.1 owes the dependency test
// that makes "one line from candidate to order" fail to compile, and until it
// lands this paragraph is a description rather than a guarantee. That distinction
// matters: this is where strategy lanes will attach, and the proposal's own risk
// table offers the test as the mitigation.
//
// # The two rules that keep the record honest
//
//   - A value the source reported and a value we derived are different columns.
//     Reported holds the first; the second arrives with the metrics in section 3.
//     When a number is wrong, the record has to say whether the source contract
//     moved or our arithmetic did.
//   - Absent is not zero. Every decimal is a string and empty means "we do not
//     have this", never 0 — a zero day high would make every high-proximity
//     comparison pass, which is the quiet form of turning a veto off.
package candidate

import (
	"fmt"
	"strings"
	"time"
)

// Market identifiers. They are the same two strings the rest of TossOS uses; this
// package does not translate them.
const (
	MarketKR = "KR"
	MarketUS = "US"
)

// SourceID names one data source, as it appears in a candidate's `sources` list.
//
// The identity is the *source*, not the transport: an official ranking reached
// through the hybrid client's WTS fallback is still SourceOfficialTradingValue, with
// Observation.ViaFallback set. Collapsing those two into one id is what makes
// over-polling invisible — see D14.
type SourceID string

const (
	// The official ranking endpoint serves several ranking *types*, and each one
	// is its own source rather than three calls sharing an id.
	//
	// They were one id until the section-2 review, and the collision was not
	// cosmetic. A scan may only cool a candidate when every source that raised it
	// answered (scan.go), and that check is keyed by SourceID — so with one
	// shared id, a 429 on the gainers list was masked by the two rankings that
	// did answer, and a candidate only ever raised by the rate-limited one got
	// cooled by a scan that had never looked at it. The cooling clock then
	// expires it, which discards first_seen_at. The same collision also collapsed
	// the per-source rate budgets to whichever call returned last.
	//
	// Merging the three into a single Source would be the other wrong fix: two
	// lists carrying the same symbol would produce two observations with the same
	// (source, symbol, instant), and section 3's per-source rate series would then
	// difference one against the other.

	// SourceOfficialTradingValue ranks by trading value — where money is going
	// now, which is the question this change is actually asking.
	SourceOfficialTradingValue SourceID = "official_rankings_trading_value"
	// SourceOfficialTradingVolume ranks by share count.
	SourceOfficialTradingVolume SourceID = "official_rankings_trading_volume"
	// SourceOfficialGainers ranks by change. By definition a list of moves that
	// have already happened, which is why it is one input among several rather
	// than the list.
	//
	// It is not on any panel. The API refuses the duration the ranking adapter
	// sends for this type — 400 unsupported-ranking-duration — so it never
	// answered at all, and a source that cannot answer is a configuration defect
	// rather than a degradation (`retire-gainers-source`). The id stays because
	// an observation already written under it has to read back as something the
	// system recognises, and because putting the source back should be one slice
	// element rather than a rediscovery of why each ranking type needs its own
	// id. candidatesrc.Panel carries the conditions for reopening that decision.
	SourceOfficialGainers SourceID = "official_rankings_top_gainers"
	// SourceOfficialPrices is GET /api/v1/prices — 200 symbols per call, and no
	// day high in the response.
	SourceOfficialPrices SourceID = "official_prices"
	// SourceOfficialCandles is GET /api/v1/candles — one symbol per call, and
	// the only source of an intraday high.
	SourceOfficialCandles SourceID = "official_candles"
	// SourceWTSPopular, SourceWTSInvestor, SourceWTSTheme and SourceWTSScreener
	// are the WTS read-only lists. All four are additive: discovery degrades
	// when they are gone, it does not stop.
	SourceWTSPopular  SourceID = "wts_popular"
	SourceWTSInvestor SourceID = "wts_investor"
	SourceWTSTheme    SourceID = "wts_theme"
	SourceWTSScreener SourceID = "wts_screener"
)

// Official reports that this source is an official Open API endpoint. The spec
// requires a configuration in which the official sources alone produce
// candidates, and this is what a test asks.
func (s SourceID) Official() bool { return strings.HasPrefix(string(s), "official_") }

// Reported is one source's answer, verbatim.
//
// Nothing in this struct is computed by us. Every decimal is a string, and empty
// means the source did not carry the field — see the package doc's second rule.
type Reported struct {
	// Rank is the symbol's 1-based position in the list that carried it, 0 when
	// the source is not a ranking. RankTotal is that list's length as it arrived,
	// which is what makes two lists of different sizes comparable as percentiles
	// (D8).
	Rank, RankTotal int
	// RankRequested is how many rows the source asked for, 0 when it declared no
	// size. RankTotal ÷ RankRequested is what separates a short list from a
	// truncated reading of a long one, and only the source knows the second
	// number — it used to be discarded at the adapter.
	RankRequested int
	// NewlyListed reports that this symbol was not in the previous reading of
	// this list.
	//
	// It is recorded as its own fact and is never folded into a rank velocity by
	// pretending the symbol was previously last (D8, correction 1). Synthesising
	// a previous rank makes every ordinary churn at the list boundary look like a
	// maximal jump, and the boundary churns constantly.
	//
	// Three-state, and unknown is the zero value: a source that has not read this
	// list before has no answer, and recording that as "it was not there" would
	// make the first scan of every session declare the whole panel newly listed.
	NewlyListed NewlyListed
	// TradingValue and TradingVolume are the cumulative intraday figures the
	// source reported. Rates and accelerations are derived from the difference
	// between two of these, never stored here.
	TradingValue, TradingVolume string
	// Price is the last traded price.
	Price string
	// DayHigh and DayLow are the intraday extremes. The ranking and price
	// endpoints carry neither; only candles do, at one call per symbol. Empty is
	// therefore the common case and it must stay distinguishable from a measured
	// value — the whole of D10 rests on this field being absent rather than zero.
	DayHigh, DayLow string
}

// Observation is one source's reading of one symbol at one instant.
type Observation struct {
	// Market and Symbol are the candidate identity (D1). The source is not part
	// of it: one symbol raised by two sources is one candidate that two sources
	// supported.
	Market, Symbol string
	// Source is who reported it.
	Source SourceID
	// ObservedAt is when. It comes from the injected clock, never from
	// time.Now(), because the time axis is the feature.
	ObservedAt time.Time
	// Reported is the source's own answer.
	Reported Reported
	// ViaFallback reports that this reading came from the hybrid client's
	// automatic WTS fallback rather than from the official endpoint (D14).
	//
	// The fallback means nothing fails when we poll the official API too fast —
	// candidates keep appearing and the screen looks normal while the source mix
	// silently changes underneath. A rising fallback rate is the measurement of
	// an undocumented rate limit, but only if it is written down.
	ViaFallback bool
}

// Key is the candidate identity this observation belongs to.
func (o Observation) Key() Key { return Key{Market: o.Market, Symbol: o.Symbol} }

// Key is a candidate's identity: the market and the symbol, and nothing else.
type Key struct{ Market, Symbol string }

func (k Key) String() string { return k.Market + ":" + k.Symbol }

// State is where a candidate is in its life.
//
// The three exist so that a symbol which briefly leaves the ranking list is not
// treated as a new discovery when it returns. That single rule is what keeps the
// chase veto from being bypassable — see D1 and Store.Promote.
type State string

const (
	// StateActive: currently above the threshold.
	StateActive State = "active"
	// StateCooled: dropped below, still inside the cooling TTL. A re-entry from
	// here keeps the original FirstSeenAt.
	StateCooled State = "cooled"
	// StateExpired: the cooling TTL elapsed. The next pass is a new candidate
	// with a new FirstSeenAt.
	StateExpired State = "expired"
)

// DefaultCoolingTTL is how long a candidate stays re-entrant after dropping below
// the threshold.
//
// Long enough that ordinary churn at the list boundary does not mint a new
// candidate, short enough that a symbol seen once in the morning is not still
// "recently discovered" at the close. It is a lifecycle knob and not a trading
// threshold: it decides which of two records a later observation joins, never
// whether anything is bought — but it is still the knob that decides whether an
// already-doubled symbol can present itself as newly discovered, so it is a
// recorded policy value (design.md "결정된 계약값") rather than a number chosen here.
const DefaultCoolingTTL = 30 * time.Minute

// DefaultStalenessTTL is how long a candidate nobody has re-promoted stays active
// before it cools implicitly.
//
// It exists because explicit cooling needs a writer and the writer is the scan
// process — see Store's stateAt. Ten minutes is set against the longest planned
// rate-limit backoff (300s): a retreat that long is normal operation and must not
// expire the candidates it happens during, so the bar sits at twice it. Below the
// backoff, an ordinary 429 would start dropping first_seen_at values; far above
// it, a killed scan leaves candidates claiming to be known since a session that
// ended hours ago.
const DefaultStalenessTTL = 10 * time.Minute

// Candidate is the summary that outlives the raw rows.
//
// D11 splits retention in two: raw observations are pruned after a couple of
// trading days, this survives until the candidate expires. The split exists
// because the obvious tidy-up — delete old rows — would otherwise delete
// FirstSeenAt, and FirstSeenAt is the entire claim this package makes.
type Candidate struct {
	Key
	// State is evaluated against a supplied instant rather than stored, so a
	// candidate that cooled before a restart reads as expired afterwards without
	// anything having run in between.
	State State
	// FirstSeenAt is the first threshold crossing of this candidate's life. It is
	// never advanced while the candidate is alive.
	FirstSeenAt time.Time
	// LastSeenAt is the most recent crossing.
	LastSeenAt time.Time
	// CooledAt is when it last dropped below, zero while active.
	CooledAt time.Time
	// Sources are the sources that have supported this candidate, sorted.
	Sources []SourceID
	// SourcesAttempted and SourcesResponded are the completeness of the scan that
	// most recently touched this candidate (D4). A candidate produced from two of
	// five sources is not wrong, but a reader who does not know it was two of five
	// is judging on less than they think.
	SourcesAttempted, SourcesResponded int
	// Degraded reports that some source of that scan failed.
	Degraded bool
}

// Complete reports that every attempted source answered.
func (c Candidate) Complete() bool {
	return c.SourcesAttempted > 0 && c.SourcesResponded == c.SourcesAttempted
}

// Age is how long this candidate has been known at the given instant.
func (c Candidate) Age(at time.Time) time.Duration { return at.Sub(c.FirstSeenAt) }

// validate rejects an observation that could not be joined to anything later.
//
// It is deliberately strict about identity and time and silent about values: a
// source that omits a figure is an ordinary state this package records, but a
// source that omits the symbol or the instant has produced a row no query can
// ever use.
func (o Observation) validate() (Observation, error) {
	out := o
	out.Market = strings.ToUpper(strings.TrimSpace(o.Market))
	out.Symbol = strings.TrimSpace(o.Symbol)
	switch {
	case out.Market == "":
		return Observation{}, fmt.Errorf("candidate: observation has no market")
	case out.Symbol == "":
		return Observation{}, fmt.Errorf("candidate: observation for %s has no symbol", out.Market)
	case strings.TrimSpace(string(o.Source)) == "":
		return Observation{}, fmt.Errorf("candidate: observation of %s:%s names no source",
			out.Market, out.Symbol)
	case o.ObservedAt.IsZero():
		return Observation{}, fmt.Errorf(
			"candidate: observation of %s:%s from %s has no instant; the time axis is the feature",
			out.Market, out.Symbol, o.Source)
	case o.Reported.Rank < 0 || o.Reported.RankTotal < 0:
		return Observation{}, fmt.Errorf("candidate: observation of %s:%s has a negative rank (%d/%d)",
			out.Market, out.Symbol, o.Reported.Rank, o.Reported.RankTotal)
	case o.Reported.RankRequested < 0:
		// Absent is 0 here, so a negative is not a smaller absence — it is a number
		// nothing could have produced, and truncationOf would read it as unknown and
		// hide it. Refused at the boundary for the same reason a negative rank is.
		return Observation{}, fmt.Errorf(
			"candidate: observation of %s:%s asked for %d rows; absent is 0, not a negative",
			out.Market, out.Symbol, o.Reported.RankRequested)
	case o.Reported.Rank > 0 && o.Reported.RankTotal == 0:
		// A rank with no list length is an impossible reading, and it is the one
		// that does damage downstream rather than at the point of entry: the rank
		// percentile normalises by list length (D8), so rank/0 is +Inf and +Inf
		// clears every threshold it is compared against.
		return Observation{}, fmt.Errorf(
			"candidate: observation of %s:%s is ranked %d of an unstated list; "+
				"a rank without its list length cannot be normalised",
			out.Market, out.Symbol, o.Reported.Rank)
	case o.Reported.RankTotal > 0 && o.Reported.Rank > o.Reported.RankTotal:
		return Observation{}, fmt.Errorf("candidate: observation of %s:%s is ranked %d of %d",
			out.Market, out.Symbol, o.Reported.Rank, o.Reported.RankTotal)
	}
	out.ObservedAt = o.ObservedAt.UTC()
	return out, nil
}
