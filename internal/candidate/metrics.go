package candidate

// metrics.go is section 3: what the difference between two readings says, and the
// considerably longer list of things it does not.
//
// # Everything here is a quantity divided by an elapsed time
//
// Scan spacing is irregular by design — 429 backoff walks a source out to 300
// seconds, Schedule.Every doubles it again while the engine runs, and an operator
// starts and stops the watch. So a delta per observation is a delta per unknown
// amount of time, and comparing two of them compares two different questions. D9
// is the correction that follows: a rate is per second, an acceleration is a ratio
// of two rates, and the actual length of both windows is recorded beside the
// answer so a reader can see how much averaging is in it.
//
// The instants come from a caller's argument or from Store.Now(), which is the
// injected clock. Nothing in this package reads the wall clock — see
// TestNothingInThisPackageAsksTheWallClockWhatTimeItIs, which is a rule rather
// than a suggestion because a single time.Now() makes the time axis untestable at
// that point without anything failing.
//
// # A not-computed value is not a passed threshold
//
// This is the same rule as D10's tri-state veto, one layer down. Every way to
// have no usable acceleration produces +Inf, NaN or a sign flip in the obvious
// implementation — and +Inf clears every threshold it is ever compared against. So
// each cause is refused, and each is refused *under its own name*:
//
//	NOT_MEASURED               nobody measured this; it is the zero value
//	NO_OBSERVATIONS            the series holds nothing at all
//	READINGS_ALL_LATER         it holds readings, all later than the instant asked
//	MIXED_SERIES               it was refused at construction and measured anyway
//	WARMING_UP                 there is no prior window; the scan just started
//	PRIOR_RATE_ZERO            there is one, and nothing traded in it
//	ZERO_ELAPSED_SECONDS       there is one, and it occupied no time
//	WINDOW_NOT_POSITIVE        the caller asked for an interval that is not one
//	CUMULATIVE_WENT_BACKWARDS  the intraday cumulative walked back
//	FIGURE_ABSENT/UNREADABLE   the source did not carry the number
//	NOT_RANKED/NO_PRIOR_RANK   there is no position, or none to move from
//	PRIOR_RANK_TOO_OLD         there is one and it is not this window's other end
//
// The names are separate because the remedies are. Many WARMING_UPs means nothing
// needs fixing. Many PRIOR_RATE_ZEROs means this panel is pointed at illiquid
// symbols. Many CUMULATIVE_WENT_BACKWARDS means a session boundary is being read
// as intraday. Many NO_OBSERVATIONS means the panel does not cover these symbols
// at all. Collapsing them into one "not computed" throws away the only signal that
// says which. D3 splits seen_late from extended for exactly this reason.
//
// The first name on that list is the one the type itself has to carry. `Computed`
// asks for the value and not merely for the absence of a reason, because a struct
// nobody assigned has no reason: a map miss, a slot in a make([]Acceleration, n)
// that no read reached, a field left unset for a candidate outside the hot queue.
// Every one of those is the natural spelling, and each would otherwise read as "we
// measured it and it crossed nothing".
//
// # The series is one source's
//
// TradingValue is a different cumulative per source — different unit, different
// aggregation window, different fallback status — so differencing one source's
// reading against another's measures the gap between the two sources, not the
// market (D9). This is the correction that had to be written down because the
// natural implementation is the wrong one: Store.Observations interleaves every
// source in time order, and differencing that compiles and returns a plausible
// number. Hence SourceSeries, which cannot be built from two sources, and
// Store.SourceObservations, which is the read that feeds it.

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
)

// Field names one cumulative intraday figure a source reports.
//
// The two are separate series and never mixed for the same reason two sources
// are: a value in currency and a count of shares are different numbers with
// different units, and a rate computed across the pair is meaningless.
type Field string

const (
	// FieldTradingValue is money — where the flow is going, which is the question
	// this change is asking.
	FieldTradingValue Field = "trading_value"
	// FieldTradingVolume is share count, which moves differently in a low-priced
	// name and is kept as its own series rather than inferred from the value.
	FieldTradingVolume Field = "trading_volume"
)

// cumulative is the reported figure for a field, verbatim. Empty means the source
// did not carry it, and that stays distinguishable from "the source said zero"
// all the way to the caller — see NotComputedFigureAbsent.
func (r Reported) cumulative(f Field) string {
	if f == FieldTradingVolume {
		return r.TradingVolume
	}
	return r.TradingValue
}

// NotComputed is why a metric has no value.
//
// Empty means it was computed. Everything else is a distinct cause, and none of
// them is a threshold crossing — see the file header.
type NotComputed string

const (
	// NotComputedUnset: nobody ever measured this. It is the reason the zero
	// value carries, and it exists because the alternative is that an unassigned
	// struct reports itself as measured — a map miss, a slot in a
	// make([]Acceleration, n) that no read filled, a field left unset for a
	// candidate outside the hot queue. All three are the natural spelling and all
	// three would otherwise read as "we measured it and it crossed nothing",
	// which is D10's failure one layer down. level.go pins the same rule with
	// TestTheZeroValueOfEveryLevelMetricIsUnmeasured.
	NotComputedUnset NotComputed = "NOT_MEASURED"
	// NotComputedWarmingUp: there is a newest reading, and not enough history
	// behind it to fill the prior window. The scan started, or this symbol has
	// only just appeared. Nothing is wrong — which is exactly why the causes
	// below are kept out of it.
	NotComputedWarmingUp NotComputed = "WARMING_UP"
	// NotComputedNoObservations: the series holds no readings at all. Not the
	// same fact as a young series: this source has never reported this symbol, or
	// the read that should have filled the series failed and its error was
	// dropped. D9's correction is explicit that the two must be distinguishable —
	// 전자가 많으면 원천 구성 문제이고 후자가 많으면 그냥 방금 시작한 것 — and level.go
	// draws the same line with LevelNoObservations.
	NotComputedNoObservations NotComputed = "NO_OBSERVATIONS"
	// NotComputedReadingsAllLater: the series holds readings and every one of
	// them is later than the instant being evaluated. A clock or an argument is
	// wrong, and folding it into WARMING_UP would make it permanent and
	// unremarkable: the metric would never compute and nothing would ever say so.
	NotComputedReadingsAllLater NotComputed = "READINGS_ALL_LATER"
	// NotComputedMixedSeries: the series was refused at construction because it
	// carried more than one source or more than one symbol, and the caller
	// measured it anyway. A wiring defect rather than a market condition — the
	// same call level.go makes with LevelMixedCandidates, and for the same
	// reason: unmeasured is the one state that is never read as safe (D10), and
	// a metric that panicked would push a wiring defect into a path whose job is
	// to keep scanning.
	NotComputedMixedSeries NotComputed = "MIXED_SERIES"
	// NotComputedPriorRateZero: there is a prior window and nothing traded in it.
	// A halted or illiquid symbol, and a division by zero that would otherwise
	// report the strongest acceleration the scale can express.
	NotComputedPriorRateZero NotComputed = "PRIOR_RATE_ZERO"
	// NotComputedZeroElapsed: the window's two readings carry the same instant,
	// so a real delta would be divided by no time at all. This one describes the
	// data — a replayed scan, a body that listed the symbol twice.
	NotComputedZeroElapsed NotComputed = "ZERO_ELAPSED_SECONDS"
	// NotComputedWindowNotPositive: the caller asked for a window of zero or
	// negative length, or handed MeasureWindow two readings in the wrong order.
	// Kept apart from ZERO_ELAPSED_SECONDS because that names a market's data and
	// this names a caller — an operator counting ZERO_ELAPSED_SECONDS is looking
	// for a duplicating source and would never find one. Same class as
	// SOURCE_MISMATCH.
	NotComputedWindowNotPositive NotComputed = "WINDOW_NOT_POSITIVE"
	// NotComputedReversed: the cumulative figure went backwards. It is an
	// intraday running total, so a session boundary or a source restart walks it
	// back — and a ratio built from two negatives comes out positive and large.
	NotComputedReversed NotComputed = "CUMULATIVE_WENT_BACKWARDS"
	// NotComputedFigureAbsent: the source did not carry this figure. Absent, not
	// zero — a zero would make the delta the whole of the other reading.
	NotComputedFigureAbsent NotComputed = "FIGURE_ABSENT"
	// NotComputedFigureUnreadable: the source carried something that is not a
	// plain decimal. Separate from absent because one is a contract the source
	// never made and the other is a contract it broke.
	NotComputedFigureUnreadable NotComputed = "FIGURE_UNREADABLE"
	// NotComputedSourceMismatch: the two readings are not from the same source or
	// not about the same symbol. A programming error rather than a market
	// condition, and the one that produces the most convincing wrong number.
	NotComputedSourceMismatch NotComputed = "SOURCE_MISMATCH"
	// NotComputedNotRanked: the newest reading carries no rank, because this
	// source is not a ranking.
	NotComputedNotRanked NotComputed = "NOT_RANKED"
	// NotComputedNoPriorRank: there is no earlier ranked reading to measure the
	// move from. Never filled in with "previously last" — see RankChange.
	NotComputedNoPriorRank NotComputed = "NO_PRIOR_RANK"
	// NotComputedPriorRankTooOld: there is an earlier ranked reading and it is
	// too old to be the other end of a `window`-long move. Its own reason rather
	// than NO_PRIOR_RANK because the remedies differ in the usual way: no prior
	// rank means the symbol was not in the list, and a prior rank we cannot use
	// means our scans are further apart than the move we are trying to measure.
	NotComputedPriorRankTooOld NotComputed = "PRIOR_RANK_TOO_OLD"
)

// DefaultAccelerationWindow is the contract's `acceleration.window_seconds: 30`.
//
// The companion `minimum_history_seconds: 60` is not a second knob: Accelerate
// needs two windows, so sixty seconds of history is what having both of them
// means. Faster polling does not produce an acceleration sooner — filling two
// thirty-second windows takes wall-clock time, not samples (D13). What it buys is
// the accuracy of the window edges and the resolution of first_seen_at.
const DefaultAccelerationWindow = 30 * time.Second

// MaxRankPriorAge is how old the earlier ranked reading may be before RankChange
// refuses to call the difference a move at all.
//
// rankedAt reaches as far back as the retained series goes, so without a bound the
// same eighty-point gain comes back for a prior reading half an hour, nine hours
// or two days old, and is compared against `rank.percentile_gain_30s_pct: 20` — a
// threshold written for thirty seconds. Unlike a rate, a percentile gain cannot be
// normalised out of the problem: a symbol that climbed eighty points over a
// session has not accelerated.
//
// # Ten minutes, and why it is not a multiple of `window`
//
// Same derivation as DefaultStalenessTTL, against the same fact: the longest
// planned rate-limit backoff is 300 seconds, a retreat that long is normal
// operation rather than an anomaly, and the bar sits at twice it. A bound that
// refused mid-ladder would leave rank velocity unmeasured during exactly the
// retreats the design guarantees will happen — the state DefaultStalenessTTL was
// set to avoid for first_seen_at, one metric across.
//
// A multiple of `window` was the wrong shape for it. What makes a prior reading
// stale is the backoff ladder, not the length of the window somebody asked for,
// so a multiple moves the bound whenever the window is tuned and for no reason
// connected to why the bound exists.
//
// The two constants are the same number because they are the same derivation, and
// TestTheRankAgeBoundAndTheStalenessTTLShareOneDerivation fails if one is moved
// without the other. Separating them is allowed and has to be argued for in that
// test rather than by editing a literal.
const MaxRankPriorAge = 10 * time.Minute

// metricScale is how many fractional digits a non-terminating value is rendered
// with. Twelve, the same as internal/exitpolicy and internal/position, and for the
// same reason: it leaves ten digits of headroom below the smallest tick either
// market quotes.
const metricScale = 12

// ShadowThresholds are the five acceleration ratios this change records against.
//
// It records and decides nothing (design.md, "결정된 계약값": 이 change는 이 값들로
// 판정하지 않는다). Recording all five is what lets the lane thresholds in T3.2 be
// derived from what actually happened instead of asserted in advance, so leaving
// any of them out is a decision taken early by omission.
var ShadowThresholds = []string{"1.3", "1.5", "1.8", "2.0", "2.5"}

// ThresholdCrossing is one shadow threshold and whether an acceleration reached it.
//
// Crossings exist only on a computed acceleration. An uncomputed one carries none,
// which is what stops "we could not measure this" from being tallied as "this
// passed" — the counting form of D10.
type ThresholdCrossing struct {
	Threshold string
	Crossed   bool
}

// Window is one measured interval of one source's cumulative series.
//
// Seconds is the *actual* elapsed time between the two readings, not the
// configured window length, and it is recorded because D9 requires it: a 1.8
// measured across ninety seconds and a 1.8 measured across thirty are not the
// same fact, and a reader must not have to re-derive the difference from
// timestamps.
type Window struct {
	// From and To are the two readings this window was measured between.
	From, To time.Time
	// Seconds is To − From, exactly, as a decimal string.
	Seconds string
	// FromFigure and ToFigure are the source's own cumulative strings at the two
	// ends, verbatim, kept beside the derived Delta so the record can say whether
	// a wrong figure came off the wire or out of this file (D6) — the same reason
	// level.go carries FirstPrice and LastPrice. They are filled in before the
	// figures are parsed, so a FIGURE_UNREADABLE window still shows what the
	// source actually sent, and which end of it was the problem.
	FromFigure, ToFigure string
	// Delta is the change in the cumulative figure across the window. Empty when
	// the figure was absent or unreadable.
	Delta string
	// Rate is Delta ÷ Seconds, per second. Empty when the window was not usable —
	// and empty is the only spelling of that, never "0".
	Rate string
}

// Measured reports that this window produced a rate.
func (w Window) Measured() bool { return w.Rate != "" }

// Acceleration is the ratio of a recent window's rate to the prior window's (D9).
type Acceleration struct {
	Key    Key
	Source SourceID
	Field  Field
	// At is the instant the series was evaluated as of.
	At time.Time
	// Recent and Prior are the two windows, with their actual lengths. They are
	// populated as far as the measurement got, so a not-computed acceleration
	// still shows what was seen.
	Recent, Prior Window
	// Ratio is Recent.Rate ÷ Prior.Rate. Empty when Reason is set — an uncomputed
	// value has no spelling.
	Ratio string
	// Reason is why Ratio is empty. Empty itself means it was computed.
	Reason NotComputed
	// Crossings is the shadow record, present only on a computed acceleration.
	Crossings []ThresholdCrossing
}

// Computed reports that this acceleration has a ratio.
//
// It requires the value and not just the absence of a reason, because the zero
// value has no reason: an Acceleration nobody assigned would otherwise report
// itself as measured, with no ratio and no crossings, which reads as "we measured
// it and it crossed nothing". See NotComputedUnset.
func (a Acceleration) Computed() bool { return a.Reason == "" && a.Ratio != "" }

// Why is the cause there is no ratio, including the one an unassigned value
// carries. Empty exactly when Computed.
func (a Acceleration) Why() NotComputed {
	if a.Reason != "" {
		return a.Reason
	}
	if a.Ratio == "" {
		return NotComputedUnset
	}
	return ""
}

// Crossed reports that a computed acceleration reached the named threshold. An
// uncomputed one crosses nothing, because it measured nothing.
func (a Acceleration) Crossed(threshold string) bool {
	if !a.Computed() {
		return false
	}
	for _, c := range a.Crossings {
		if c.Threshold == threshold {
			return c.Crossed
		}
	}
	return false
}

// RankMove is a symbol's change of normalised list position, and the separate fact
// that it was not in the list before.
type RankMove struct {
	Key    Key
	Source SourceID
	// From and To are the two readings compared. From is zero when there was none.
	From, To time.Time
	// Seconds is To − From, exactly, as a decimal string, and it is here for the
	// reason D9 put it on Window: the contract knob is
	// rank.percentile_gain_30s_pct, so a gain measured across ninety seconds and
	// one measured across thirty are not the same fact, and a reader must not have
	// to re-derive the span from the two timestamps. Empty when there was no prior
	// reading to measure from.
	Seconds string
	// Rank and RankTotal are the newest reading's position and list length.
	Rank, RankTotal int
	// PercentileGain is the rise in percentile points, positive when the symbol
	// moved up the list. Empty when Reason is set.
	PercentileGain string
	// Reason is why PercentileGain is empty.
	Reason NotComputed
	// NewlyListed is the source's own report that this symbol was absent from its
	// previous reading. It is recorded here and never converted into a previous
	// rank — see RankChange. Three-state: a source's first reading has no answer.
	NewlyListed NewlyListed
}

// Computed reports that this move has a gain. Like Acceleration.Computed it
// requires the value, so an unassigned RankMove cannot read as a measured one.
func (m RankMove) Computed() bool { return m.Reason == "" && m.PercentileGain != "" }

// Why is the cause there is no gain, including the one an unassigned value
// carries. Empty exactly when Computed.
func (m RankMove) Why() NotComputed {
	if m.Reason != "" {
		return m.Reason
	}
	if m.PercentileGain == "" {
		return NotComputedUnset
	}
	return ""
}

// CrossingTally is the shadow record over a set of candidates: how many crossed
// each threshold, and how many could not be measured, by cause.
//
// The two counts are disjoint by construction. A screen that folded the second
// into the first would report the not-computed candidates as passes, and under the
// candle budget the not-computed ones are the majority (D10, D13 decision 3).
//
// Total and Computed are what make the disjointness checkable rather than
// asserted: Total == Computed + Σ NotComputed holds for every tally, so a
// candidate that fell out of both halves — the shape the unset bucket had before
// NotComputedUnset existed — shows up as an arithmetic failure instead of as a
// number that is quietly short.
type CrossingTally struct {
	// Total is how many accelerations were tallied, measured or not.
	Total int
	// Computed is how many of them had a ratio. It is not the sum of Crossed: a
	// computed acceleration that reached no threshold is counted here and nowhere
	// in Crossed, and one that reached three is counted once here and three times
	// there.
	Computed int
	// Crossed counts, per shadow threshold, the computed accelerations that
	// reached it. Its keys are exactly ShadowThresholds.
	Crossed map[string]int
	// NotComputed counts the rest by cause, including NotComputedUnset.
	NotComputed map[NotComputed]int
}

// --- the series ------------------------------------------------------------------

// SourceSeries is one source's readings of one symbol, oldest first.
//
// The type exists so that "difference two readings" cannot be spelled across two
// sources. D9 warns that the natural implementation is the wrong one and that it
// compiles: Store.Observations returns every source's rows in time order, and a
// difference taken over that is the gap between two sources' conventions rather
// than the market's acceleration. Refusing the mixed construction is what turns
// that from a comment into a compile-and-run failure.
// A refused series carries its refusal rather than coming back as the zero value,
// because the error is the one thing a scan loop is most likely to drop: it must
// keep going, and `series, _ := NewSourceSeries(rows)` compiles. Without the
// sticky state the measurement that follows answers WARMING_UP — the one reason
// that means nothing needs fixing — for what is a source composition defect.
type SourceSeries struct {
	Key    Key
	Source SourceID
	rows   []Observation
	// rejected is set when the series was refused at construction. It is never
	// cleared: a series that could not be built cannot be measured later.
	rejected NotComputed
}

// NewSourceSeries builds a series from readings that must all be one source's
// readings of one symbol.
//
// It refuses an empty input rather than returning an empty series, because a
// series with no rows has no identity to report — Store.SourceSeries knows the
// identity from its arguments and constructs that case itself.
func NewSourceSeries(rows []Observation) (SourceSeries, error) {
	if len(rows) == 0 {
		return SourceSeries{}, fmt.Errorf(
			"candidate: a source series needs at least one reading to know whose it is")
	}
	out := SourceSeries{Key: rows[0].Key(), Source: rows[0].Source}
	// The refusal travels in the returned value as well as in the error, so a
	// caller that drops the error measures a series that refuses to be measured
	// rather than one that looks like a scan that has just started.
	refused := SourceSeries{Key: out.Key, Source: out.Source, rejected: NotComputedMixedSeries}
	for _, o := range rows {
		if o.Source != out.Source {
			return refused, fmt.Errorf(
				"candidate: a series of %s cannot also hold %s readings of %s; "+
					"the two report different cumulatives and differencing them measures "+
					"the gap between the sources, not the market",
				out.Source, o.Source, out.Key)
		}
		if o.Key() != out.Key {
			return refused, fmt.Errorf(
				"candidate: a series of %s cannot also hold readings of %s", out.Key, o.Key())
		}
	}
	out.rows = append(out.rows, rows...)
	// Stable so two readings sharing an instant keep the order they were read in,
	// which is the order the store returns them in (observed_at, id).
	sort.SliceStable(out.rows, func(i, j int) bool {
		return out.rows[i].ObservedAt.Before(out.rows[j].ObservedAt)
	})
	return out, nil
}

// Len is how many readings the series holds.
func (s SourceSeries) Len() int { return len(s.rows) }

// Rows is a copy of the readings, oldest first.
func (s SourceSeries) Rows() []Observation { return append([]Observation{}, s.rows...) }

// Unusable reports why this series cannot be measured at all, empty when it can.
//
// It is the two conditions that are properties of the series rather than of any
// particular instant: it was refused at construction, or it holds nothing. Both
// are kept out of WARMING_UP because that reason means the scan is young and
// nothing needs fixing, and these two mean a source is miswired and a source has
// never answered.
func (s SourceSeries) Unusable() NotComputed {
	if s.rejected != "" {
		return s.rejected
	}
	if len(s.rows) == 0 {
		return NotComputedNoObservations
	}
	return ""
}

// at returns the newest reading at or before t, resolving a tie to the oldest of
// the readings sharing that instant.
//
// "At or before" rather than "nearest" because every window here is measured
// backwards from a known instant: an anchor that could jump forward would let a
// window include a reading taken after the instant being evaluated.
//
// The tie goes to the oldest because two readings of one source sharing an instant
// are a replayed scan or a body that listed the symbol twice, and taking the later
// one silently drops the earlier from every window anchored there — the anchor
// moves to a figure published second and the acceleration changes with no reason
// code and nothing in the record saying it happened. Resolving to the oldest makes
// a duplicated row inert: the answer is the one the undoubled series gives.
func (s SourceSeries) at(t time.Time) (Observation, bool) {
	found := -1
	for i := len(s.rows) - 1; i >= 0; i-- {
		o := s.rows[i]
		if o.ObservedAt.After(t) {
			continue
		}
		// The rows are sorted, so the first earlier instant ends the tie.
		if found >= 0 && o.ObservedAt.Before(s.rows[found].ObservedAt) {
			break
		}
		found = i
	}
	if found < 0 {
		return Observation{}, false
	}
	return s.rows[found], true
}

// rankedAt returns the newest reading at or before t that carries a rank, with the
// same tie rule as at.
func (s SourceSeries) rankedAt(t time.Time) (Observation, bool) {
	found := -1
	for i := len(s.rows) - 1; i >= 0; i-- {
		o := s.rows[i]
		if o.ObservedAt.After(t) || o.Reported.Rank <= 0 || o.Reported.RankTotal <= 0 {
			continue
		}
		if found >= 0 && o.ObservedAt.Before(s.rows[found].ObservedAt) {
			break
		}
		found = i
	}
	if found < 0 {
		return Observation{}, false
	}
	return s.rows[found], true
}

// --- measurement -------------------------------------------------------------------

// MeasureWindow measures the interval between two readings of one source.
//
// It is the primitive every window goes through, so it is where the unusable
// denominators are refused. Two readings carrying the same instant are the case
// D9 names — one scan's rows all share an instant, and a replay or a duplicated
// row in a body produces the pair — and dividing a real delta by no elapsed time
// yields the largest number the scale has.
//
// It is exported because a window between two chosen readings is what section 4
// needs for the extension against first_seen_at, and because the guard is worth
// more where a caller can reach it than buried in Accelerate.
func MeasureWindow(from, to Observation, field Field) (Window, NotComputed) {
	w, _, reason := measureWindow(from, to, field)
	return w, reason
}

// measureWindow is MeasureWindow with the exact rate kept as a rational.
//
// Accelerate divides one rate by another and must do it on the exact values: the
// rendered Rate is truncated at metricScale, and a ratio built from two truncated
// strings would drift in a direction nobody chose.
func measureWindow(from, to Observation, field Field) (Window, *big.Rat, NotComputed) {
	w := Window{From: from.ObservedAt, To: to.ObservedAt}

	if from.Source != to.Source || from.Key() != to.Key() {
		// The most convincing wrong number in this package. Refused here as well as
		// in NewSourceSeries because this function takes raw readings.
		return w, nil, NotComputedSourceMismatch
	}

	seconds := elapsedRat(from.ObservedAt, to.ObservedAt)
	w.Seconds = formatDecimal(seconds)
	switch {
	case seconds.Sign() < 0:
		// The readings were handed over in the wrong order. A caller's mistake, and
		// named as one — an operator counting duplicated scans must not find this in
		// the same bucket.
		return w, nil, NotComputedWindowNotPositive
	case seconds.Sign() == 0:
		return w, nil, NotComputedZeroElapsed
	}

	// Recorded before either is parsed, so an unreadable figure leaves behind the
	// string the source actually sent and which end of the window it was on (D6).
	w.FromFigure = strings.TrimSpace(from.Reported.cumulative(field))
	w.ToFigure = strings.TrimSpace(to.Reported.cumulative(field))

	start, reason := parseFigure(from.Reported.cumulative(field))
	if reason != "" {
		return w, nil, reason
	}
	end, reason := parseFigure(to.Reported.cumulative(field))
	if reason != "" {
		return w, nil, reason
	}

	delta := new(big.Rat).Sub(end, start)
	w.Delta = formatDecimal(delta)
	if delta.Sign() < 0 {
		// Recorded, not computed. The delta stays in the window because the size of
		// the step backwards is what tells a reader whether this was a session
		// boundary or a source that restarted mid-list.
		return w, nil, NotComputedReversed
	}

	rate := new(big.Rat).Quo(delta, seconds)
	w.Rate = formatDecimal(rate)
	return w, rate, ""
}

// Rate is the per-second change of a cumulative figure across the newest window.
//
// The window is anchored at the newest reading at or before `at` and reaches back
// to the newest reading at or before `window` before that — so a window the
// sampling stretched is measured at the length it actually had, and Seconds says
// so. That is the spec's "관측 횟수당 차분을 변화율로 사용해서는 안 된다": a delta per
// observation is a delta per unknown amount of time, and it makes the densely
// sampled stretches look calm.
func Rate(s SourceSeries, field Field, at time.Time, window time.Duration) (Window, NotComputed) {
	if reason := s.Unusable(); reason != "" {
		return Window{}, reason
	}
	if window <= 0 {
		return Window{}, NotComputedWindowNotPositive
	}
	end, ok := s.at(at)
	if !ok {
		// There are readings and they are all later than the instant asked about.
		return Window{}, NotComputedReadingsAllLater
	}
	start, ok := s.at(end.ObservedAt.Add(-window))
	if !ok {
		return Window{}, NotComputedWarmingUp
	}
	return MeasureWindow(start, end, field)
}

// Accelerate is the ratio of the recent window's rate to the prior window's (D9).
//
// The three anchors are each at least `window` apart, walking back from the newest
// reading rather than from `at`:
//
//	end   the newest reading at or before `at`
//	mid   the newest reading at or before end − window
//	start the newest reading at or before mid − window
//
// Anchoring the prior window on `mid` rather than on `at − 2×window` is what makes
// the backoff case work. When a 429 stretches the recent window to ninety seconds,
// both `at − window` and `at − 2×window` land inside the gap and resolve to the
// same reading, which would leave no prior window at all. Anchoring on `mid` keeps
// the prior window on real readings and lets Seconds report each window's true
// length, which is the whole of D9's answer to a ratio of sums.
//
// A non-positive `window` is not silently replaced with the default — substituting
// a value the caller did not ask for would hide a configuration mistake behind a
// plausible number, which is this package's recurring failure shape — and it is
// refused as WINDOW_NOT_POSITIVE rather than as ZERO_ELAPSED_SECONDS, because the
// second names a condition in the data and this is a condition in the caller.
//
// Note that with a positive `window` neither ZERO_ELAPSED_SECONDS nor
// WINDOW_NOT_POSITIVE is reachable from here: mid is at or before end − window and
// start at or before mid − window, so both windows are at least `window` long.
// Both guards live on measureWindow, which is where MeasureWindow's callers reach
// them.
func Accelerate(s SourceSeries, field Field, at time.Time, window time.Duration) Acceleration {
	out := Acceleration{Key: s.Key, Source: s.Source, Field: field, At: at}

	if reason := s.Unusable(); reason != "" {
		out.Reason = reason
		return out
	}
	if window <= 0 {
		out.Reason = NotComputedWindowNotPositive
		return out
	}
	end, ok := s.at(at)
	if !ok {
		out.Reason = NotComputedReadingsAllLater
		return out
	}
	mid, ok := s.at(end.ObservedAt.Add(-window))
	if !ok {
		out.Reason = NotComputedWarmingUp
		return out
	}
	start, ok := s.at(mid.ObservedAt.Add(-window))
	if !ok {
		// One window's worth of history and no more. The spec: 비교할 직전 구간이
		// 없으면 가속도를 산출해서는 안 된다.
		out.Reason = NotComputedWarmingUp
		return out
	}

	recent, recentRate, reason := measureWindow(mid, end, field)
	out.Recent = recent
	if reason != "" {
		out.Reason = reason
		return out
	}
	prior, priorRate, reason := measureWindow(start, mid, field)
	out.Prior = prior
	if reason != "" {
		out.Reason = reason
		return out
	}
	if priorRate.Sign() == 0 {
		// A prior window in which nothing traded. The division would be +Inf, and
		// +Inf clears every threshold it is ever compared against.
		out.Reason = NotComputedPriorRateZero
		return out
	}

	ratio := new(big.Rat).Quo(recentRate, priorRate)
	out.Ratio = formatDecimal(ratio)
	out.Crossings = crossings(ratio)
	return out
}

// crossings records the ratio against every shadow threshold.
//
// The comparison is on the exact rational, not on Ratio's rendering. Both matter:
// a float64 division of 1.3 worth of flow by 1.0 worth is 1.2999999999999998 and
// misses its own threshold, and Ratio is truncated so it can never overstate what
// the arithmetic found. The flags are the record; the string is for a reader.
func crossings(ratio *big.Rat) []ThresholdCrossing {
	out := make([]ThresholdCrossing, 0, len(ShadowThresholds))
	for _, th := range ShadowThresholds {
		bar, ok := new(big.Rat).SetString(th)
		if !ok {
			continue
		}
		out = append(out, ThresholdCrossing{Threshold: th, Crossed: ratio.Cmp(bar) >= 0})
	}
	return out
}

// TallyCrossings counts a set of accelerations: crossings by threshold, and
// not-computed by cause.
//
// The two are disjoint. An acceleration that was not computed contributes to its
// cause and to no threshold, which is the counting form of the rule this whole
// file is built on — measuring nothing is not passing.
//
// Every input lands in exactly one of them, so Total == Computed + Σ NotComputed
// holds for every tally and a caller can assert it. Three ways to be short of that
// are refused here rather than trusted:
//
//	an unmeasured candidate carrying no reason  counted under NOT_MEASURED
//	a threshold outside ShadowThresholds        ignored, not made into a column
//	the same threshold recorded twice           counted once, first record wins
//
// The first is the one that mattered: before NotComputedUnset existed those
// candidates fell out of both halves, and the not-computed count — the number D9
// wants an operator to read — was short by exactly the candidates nobody measured.
func TallyCrossings(in []Acceleration) CrossingTally {
	out := CrossingTally{
		Total:       len(in),
		Crossed:     make(map[string]int, len(ShadowThresholds)),
		NotComputed: map[NotComputed]int{},
	}
	for _, th := range ShadowThresholds {
		out.Crossed[th] = 0
	}
	seen := make(map[string]bool, len(ShadowThresholds))
	for _, a := range in {
		if why := a.Why(); why != "" {
			out.NotComputed[why]++
			continue
		}
		out.Computed++
		for k := range seen {
			delete(seen, k)
		}
		for _, c := range a.Crossings {
			// Membership in Crossed is membership in ShadowThresholds: the map was
			// seeded with exactly those keys, and a threshold this record does not
			// shadow must not become one by being counted.
			if _, shadowed := out.Crossed[c.Threshold]; !shadowed || seen[c.Threshold] {
				continue
			}
			seen[c.Threshold] = true
			if c.Crossed {
				out.Crossed[c.Threshold]++
			}
		}
	}
	return out
}

// --- rank ---------------------------------------------------------------------------

// RankChange is a symbol's rise in normalised list position over `window`.
//
// # Normalised, because the lists are different lengths
//
// The two ranking lists this system reads are different lengths — the official
// ranking serves up to 100 rows in both markets and the WTS popularity list 30 (D8)
// — so "up thirty places" is two different distances. The move is therefore in
// percentile points, each reading normalised by its own RankTotal, which also
// survives a source returning a shorter list than it did last time.
//
// Corrected 2026-07-28: this said "the KR panel returns 150 rows and the US panel
// 100". No panel has ever returned 150 rows, in either market. The normalisation is
// needed and the sentence explaining why it was needed was fiction.
//
// # newly_listed is a fact, not a previous rank
//
// A symbol appearing in a list for the first time has no previous percentile, and
// the natural fix — treat it as having come from last place — converts having no
// evidence into having the strongest evidence available: the recorded gain becomes
// the symbol's own percentile, so anything entering above the bottom fifth clears
// the contract's 20%p on its first sighting. The rows around the list boundary
// churn in and out on every scan, so that path is taken constantly, and a symbol
// whose volume has just spiked does not enter at the boundary but in the middle.
//
// So the gain stays absent and NewlyListed is recorded beside it (D8 correction
// 1): 값을 지어내지 않고 없음을 없음으로 남긴다. The flag is worth having on its own —
// a symbol that first appears at 12th and one that first appears at 148th are
// different events, and a run of the former is how we learn our scan is late.
//
// The flag suppresses nothing. When we do hold an earlier ranked reading of our
// own, the move is measured against it and NewlyListed still travels with the
// answer, so a reader can see the series has a hole in it.
// # Both ends of the reach are bounded, and the far one is a backstop
//
// `window` says where to start looking — the answer comes from the newest ranked
// reading at or before `at − window`. MaxRankPriorAge caps how far back that
// search may land, because rankedAt otherwise reaches as far as the retained
// series does and the same gain comes back for a reading thirty seconds old and
// one from the previous session.
//
// The cap is a backstop and not the protection. **Seconds is the protection**: a
// consumer that compares a gain against `rank.percentile_gain_30s_pct` without
// reading Seconds is wrong wherever the bound sits, because 20 points over thirty
// seconds and 20 points over five minutes are different facts and only Seconds
// tells them apart. What the cap adds is the case Seconds cannot rescue — a gap
// beyond ten minutes is not explained by the planned 429 ladder at all, so it is a
// market close, a dead process or a session boundary, which is a different kind of
// fact rather than a coarser measurement of the same one.
//
// It is refused under its own reason: a prior rank too old to use says our scans
// are further apart than the move we are measuring, which is not what
// NO_PRIOR_RANK says. From and Seconds are recorded on the refusal anyway, because
// how old the only prior reading was is the number that says which of those three
// it was.
func RankChange(s SourceSeries, at time.Time, window time.Duration) RankMove {
	out := RankMove{Key: s.Key, Source: s.Source}

	if reason := s.Unusable(); reason != "" {
		out.Reason = reason
		return out
	}
	if window <= 0 {
		out.Reason = NotComputedWindowNotPositive
		return out
	}
	to, ok := s.at(at)
	if !ok {
		out.Reason = NotComputedReadingsAllLater
		return out
	}
	out.To = to.ObservedAt
	out.Rank, out.RankTotal = to.Reported.Rank, to.Reported.RankTotal
	out.NewlyListed = to.Reported.NewlyListed
	if to.Reported.Rank <= 0 || to.Reported.RankTotal <= 0 {
		out.Reason = NotComputedNotRanked
		return out
	}

	from, ok := s.rankedAt(to.ObservedAt.Add(-window))
	if !ok {
		out.Reason = NotComputedNoPriorRank
		return out
	}
	// Recorded before the age is judged, so a refused move still shows how old the
	// only prior reading was — which is the number that says whether the scan
	// spacing is the problem.
	out.From = from.ObservedAt
	out.Seconds = elapsedSeconds(from.ObservedAt, to.ObservedAt)
	// At or past the bound rather than only past it, so the boundary matches
	// DefaultStalenessTTL's — a candidate is cooled *at* the TTL, not a second
	// after, and two constants pinned to one number must not disagree about which
	// side of themselves they sit on.
	if to.ObservedAt.Sub(from.ObservedAt) >= MaxRankPriorAge {
		out.Reason = NotComputedPriorRankTooOld
		return out
	}

	out.PercentileGain = formatDecimal(new(big.Rat).Sub(percentile(to), percentile(from)))
	return out
}

// percentile is a reading's position as a share of its own list, 0 at the bottom
// and just under 100 at the top. Dividing by RankTotal is what makes a 30-row WTS
// popularity list and a 100-row official ranking comparable.
func percentile(o Observation) *big.Rat {
	total := big.NewRat(int64(o.Reported.RankTotal), 1)
	behind := big.NewRat(int64(o.Reported.RankTotal-o.Reported.Rank), 1)
	return new(big.Rat).Mul(new(big.Rat).Quo(behind, total), big.NewRat(100, 1))
}

// --- elapsed time ---------------------------------------------------------------------

// elapsedRat is to − from in seconds, exactly. Negative when the two are the wrong
// way round, which every caller checks rather than assuming.
func elapsedRat(from, to time.Time) *big.Rat {
	return big.NewRat(to.Sub(from).Nanoseconds(), int64(time.Second))
}

// elapsedSeconds is elapsedRat rendered the way Window.Seconds is, so the two
// fields that report a span report it in one spelling.
func elapsedSeconds(from, to time.Time) string { return formatDecimal(elapsedRat(from, to)) }

// --- decimals ------------------------------------------------------------------------

// parseFigure reads a reported cumulative exactly, and says which kind of nothing
// an empty or malformed one is.
//
// Absent and unreadable are separate because the remedies are: a source that never
// carried the field is a panel composition question, a source that carried
// something unparseable has broken a contract it was keeping. Neither is zero —
// the package's second rule, and the one D10 rests on.
//
// The accepted grammar is internal/riskcalc's, character for character: an
// optional sign, digits, at most one point. Exponents are refused even though
// big.Rat reads them, because no source in TossOS produces one and accepting them
// would let a single value have two spellings that compare unequal as text.
func parseFigure(s string) (*big.Rat, NotComputed) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return nil, NotComputedFigureAbsent
	}
	if !isPlainDecimal(raw) {
		return nil, NotComputedFigureUnreadable
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, NotComputedFigureUnreadable
	}
	return r, ""
}

func isPlainDecimal(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	ints, frac, hasPoint := strings.Cut(s, ".")
	if hasPoint && frac == "" {
		return false
	}
	if ints == "" && frac == "" {
		return false
	}
	return onlyDigits(ints) && onlyDigits(frac)
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// formatDecimal renders an exact rational.
//
// Exactly when the value terminates in base ten — elapsed seconds always do, and
// so does any ratio of two terminating deltas over equal windows — and truncated
// towards zero at metricScale digits when it does not.
//
// Truncated rather than rounded to nearest because the string sits beside the
// crossing flags, and the flags are computed on the exact value. A rendering that
// could round up would let the two disagree in the one direction that matters: a
// printed ratio claiming a threshold the arithmetic did not reach.
func formatDecimal(r *big.Rat) string {
	if exact, ok := exactDecimal(r); ok {
		return exact
	}
	return truncateAt(r, metricScale)
}

// exactDecimal returns the value's exact decimal spelling, or false when it has
// none. A rational terminates in base ten exactly when its reduced denominator has
// no prime factor other than 2 and 5.
func exactDecimal(r *big.Rat) (string, bool) {
	den := new(big.Int).Set(r.Denom())
	for _, p := range []int64{2, 5} {
		prime := big.NewInt(p)
		q, m := new(big.Int), new(big.Int)
		for {
			q.QuoRem(den, prime, m)
			if m.Sign() != 0 {
				break
			}
			den.Set(q)
		}
	}
	if den.Cmp(big.NewInt(1)) != 0 {
		return "", false
	}
	// The digits needed are bounded by the powers of 2 and 5 that were removed;
	// FloatString at a scale that large is exact and the trim removes the padding.
	digits := len(r.Denom().String()) * 4
	if digits < 1 {
		digits = 1
	}
	return trimDecimal(r.FloatString(digits)), true
}

// truncateAt renders the value with `digits` fractional places, dropping what does
// not fit rather than rounding it.
func truncateAt(r *big.Rat, digits int) string {
	shift := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	num := new(big.Int).Mul(r.Num(), shift)
	// QuoRem truncates towards zero, which for the non-negative values this file
	// produces is the floor.
	q := new(big.Int).Quo(num, r.Denom())
	return trimDecimal(new(big.Rat).SetFrac(q, shift).FloatString(digits))
}

// trimDecimal removes FloatString's padding so a value has one spelling, the same
// one internal/riskcalc's CanonicalDecimal produces.
func trimDecimal(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "" || s == "-" || s == "-0" {
		return "0"
	}
	return s
}
