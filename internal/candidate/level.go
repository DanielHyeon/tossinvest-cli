package candidate

// level.go answers two questions about where a candidate's price is:
//
//	expansion        how far has it run since the first price we ever saw for it
//	                 (task 3.6, the input to the `extended` veto)
//	range position   how far below the day's high it is trading
//	                 (task 3.7, the input to the `near_high` veto)
//
// Both are inputs to a veto rather than to a score, because D3 keeps chase risk
// out of the sum: a score lets a strong acceleration offset a price sitting on
// its high, and that offset is the logic of chasing written down.
//
// # Why neither answer is a float64
//
// A percentage cannot say "we did not look". The intraday high is not in the
// ranking response and not in the prices response — only candles carry it, one
// call per symbol at five calls a second — so most candidates never get one, all
// day (D13 decision 3). Returning a number anyway means one of two things:
//
//	0 for absent      every price is at its high, so near_high fires on
//	                  everything and the veto is noise nobody reads.
//	NaN for absent    every comparison against it is false, so `< threshold`
//	                  reads "far from the high" and `>= threshold` reads "not
//	                  close enough to matter". It passes both, and it passes them
//	                  silently.
//
// The second is the one that happens, because it is what dividing by an absent
// value produces. D10 calls this the second bypass of the chase veto and the
// harder one to find, "왜냐하면 아무것도 실패하지 않기 때문이다" — nothing fails.
//
// So every answer here is three-state: measured and dangerous, measured and not,
// or unmeasured with a reason. The zero value of both structs is unmeasured, so a
// field nobody assigned cannot read as a candidate somebody checked.
//
// # And why it is not a float64 for a second reason
//
// The first version of this file derived the percentages with float64 division,
// which is exactly what metrics.go refuses by name: "a float64 division of 1.3
// worth of flow by 1.0 worth is 1.2999999999999998 and misses its own threshold".
// The sibling file that stakes the actual veto boundary was doing it anyway, and
// the boundary is pinned to the digit — tasks.md 4.3 fixes 1.99% true, 2.00%
// false, 2.01% false.
//
// A two-decimal price is exactly 2.00% below its high only when the high is a
// whole multiple of 50 cents. There are 20,000 such pairs below $10,000, and
// 9,598 of them — 48% — come out of float64 as 1.99999999999999267253 rather than
// 2. The first is a $2.50 high against a $2.45 price. Every one of them fires
// near_high where the contract says it must not. Whole-unit KRW prices are all
// clean, which is why nothing showed: the defect only appears on the market where
// this will actually run.
//
// So the arithmetic is metrics.go's: parseFigure reads the source's string into a
// big.Rat, formatDecimal renders it, and the comparisons the vetoes are written
// against are made on exact rationals rather than on the rendering.
//
// # Why an unusable input is unmeasured rather than believed
//
// D9 settled this for acceleration: a denominator of zero, a window of zero
// length and a negative delta all produce +Inf, undefined and a sign inversion,
// and all three clear every threshold they meet. Not computing a value is not a
// pass, and neither is a value built from an input that cannot bear it. The same
// rule applies here to a baseline price of zero, a day high of zero, and any
// decimal the source sent that is not a number.

import (
	"math/big"
	"strings"
	"time"
)

// LevelUnmeasured names the input that was missing or unusable.
//
// A bare "unmeasured" boolean would be enough for the veto and not enough for
// anybody deciding what to fix — which is the distinction D3 draws between
// seen_late and extended, and D9 draws between WARMING_UP and an unusable
// denominator. A screen full of NO_DAY_HIGH means the candle budget is the
// constraint; a screen full of UNREADABLE_DECIMAL means a source contract moved.
// One is the design working as intended and the other is an outage.
type LevelUnmeasured string

const (
	// LevelNotEvaluated: nobody ran a measurement at all. It is the reason the
	// zero value carries, so that an unassigned struct names itself instead of
	// rendering as "unmeasured ()" — a reason code that is sometimes empty is a
	// reason code an operator learns to ignore. Reason() supplies it; the Why
	// field of a struct that was actually measured is never this.
	LevelNotEvaluated LevelUnmeasured = "NOT_EVALUATED"
	// LevelNoObservations: nothing was recorded for this candidate at all. Not
	// the same fact as a source that answered without the field — this one says
	// we have not polled, that one says we polled and were not told.
	LevelNoObservations LevelUnmeasured = "NO_OBSERVATIONS"
	// LevelMixedCandidates: the series carries more than one (market, symbol).
	// A wiring defect rather than a market condition — see oneCandidate.
	LevelMixedCandidates LevelUnmeasured = "MIXED_CANDIDATES"
	// LevelNoBaseline: the candidate summary carries no first price, so there is
	// nothing to measure expansion against. See MeasureExpansion — this is the
	// reason that used to be answered with a number derived from whichever rows
	// the caller happened to be holding.
	LevelNoBaseline LevelUnmeasured = "NO_BASELINE"
	// LevelNoPrice: no reading in the series carried a price (for expansion), or
	// the reading that carried the day high carried no price of its own (for the
	// range position).
	LevelNoPrice LevelUnmeasured = "NO_PRICE"
	// LevelNoDayHigh: no reading carried an intraday high. This is the ordinary
	// state of most candidates rather than an error — D10's whole subject.
	LevelNoDayHigh LevelUnmeasured = "NO_DAY_HIGH"
	// LevelUnreadableDecimal: a value was present and is not a usable number.
	// A thousands separator, a currency mark, "N/A", or the "NaN"/"Inf" that
	// strconv.ParseFloat accepts without complaint.
	LevelUnreadableDecimal LevelUnmeasured = "UNREADABLE_DECIMAL"
	// LevelNonPositivePrice: a price at or below zero. It cannot be a baseline
	// and it is not a price.
	LevelNonPositivePrice LevelUnmeasured = "NON_POSITIVE_PRICE"
	// LevelNonPositiveDayHigh: an intraday high at or below zero. The package
	// doc names this one: a zero day high makes every high-proximity comparison
	// pass, which is the quiet form of turning a veto off.
	LevelNonPositiveDayHigh LevelUnmeasured = "NON_POSITIVE_DAY_HIGH"
)

// Expansion is how far a candidate has run since the first price anyone reported
// for it (task 3.6).
//
// # Why the baseline is per candidate and not per source
//
// D9 keeps the rate and acceleration series per source, because TradingValue is a
// different cumulative in every source — different unit, different aggregation
// window — so differencing one against another measures the sources rather than
// the market. A price is not a cumulative: it is the last trade in this
// candidate's market, and every source reporting it is reporting the same
// number. The candidate's identity is (market, symbol) and one symbol raised by
// two sources is one candidate (D1), so the baseline is the earliest price any of
// its supporters carried. FirstSource records which one, so a disagreement
// between two sources is answerable from the record rather than invisible in it.
//
// # What this does not say
//
// It says nothing about how far the symbol ran before we saw it. That is
// seen_late's question (task 4.1), and D3 keeps the two apart because they are
// fixed by different things: seen_late says the observation system was late,
// extended says the entry would be. A single reading therefore reports 0%
// expansion — a true statement about our record — rather than unmeasured.
type Expansion struct {
	// Measured is false for every unmeasured answer, and it is the zero value so
	// that an unassigned Expansion cannot read as a candidate that has not run.
	Measured bool
	// Why names the missing input. Empty only when Measured — or when nobody
	// measured anything at all, which is what Reason() answers for.
	Why LevelUnmeasured
	// Ratio is the latest price divided by the baseline; GainPct is the same
	// thing as a percentage change. Both are exact decimal strings, and both are
	// empty on an unmeasured answer — an unmeasured metric carries its inputs but
	// never arithmetic. Empty is the only spelling of "not computed", never "0".
	Ratio, GainPct string
	// FirstPrice and LastPrice are the sources' own strings, kept beside the
	// derived numbers so the record can say whether a wrong figure came off the
	// wire or out of this file (D6). They are populated as far as the measurement
	// got, so an unmeasured answer still shows what was read.
	FirstPrice, LastPrice string
	// FirstAt and LastAt are when those two prices were observed.
	FirstAt, LastAt time.Time
	// FirstSource and LastSource are who reported them.
	FirstSource, LastSource SourceID
	// BaselineStored reports that the baseline is the candidate summary's stored
	// first price (D17) rather than a reading out of the series.
	//
	// False on a measured answer means the series carried a priced reading older
	// than the stored baseline and that older one was used. A baseline may move
	// backwards on new evidence and never forwards: forwards is the failure this
	// whole column exists to stop, and it understates every expansion measured
	// after it.
	BaselineStored bool
	// Unreadable counts readings whose price was present and not a number.
	//
	// It is a count beside a measured answer rather than a reason instead of one.
	// Nine sources feed this series, and returning nothing because one of them
	// sent a malformed row would turn `extended` off for that candidate for as
	// long as the row is retained — and the answer would carry none of the inputs
	// that were read successfully. A rising count is a source contract moving; it
	// is not a reason to stop measuring.
	Unreadable int
}

// Reason names the missing input of an unmeasured answer, and never comes back
// empty for one.
//
// The Why field can be empty on a struct nobody ever passed to a measurement — a
// map lookup that missed, a field somebody forgot to assign — and that is the one
// unmeasured state which names nothing. §5 renders this rather than the field, so
// that "unmeasured ()" cannot appear on a screen whose whole job is to stop an
// unmeasured veto from looking like a passed one.
func (e Expansion) Reason() LevelUnmeasured {
	switch {
	case e.Measured:
		return ""
	case e.Why == "":
		return LevelNotEvaluated
	default:
		return e.Why
	}
}

// GainExceeds reports that the expansion is above `thresholdPct` percent — the
// comparison the `extended` veto (task 4.2) is written against.
//
// The second return is the D10 three-state: false means this candidate was not
// measured, or the threshold is not a plain decimal, and neither of those is
// "not extended". A caller that ignores it has turned the veto off for every
// candidate nobody has priced.
//
// The comparison is made on exact rationals recomputed from the two reported
// strings, not on GainPct. GainPct is truncated at metricScale digits when the
// value does not terminate in base ten, and a rendering must not be what decides
// a threshold.
func (e Expansion) GainExceeds(thresholdPct string) (exceeds, measured bool) {
	if !e.Measured {
		return false, false
	}
	first, _, firstOK := levelFigure(e.FirstPrice)
	last, _, lastOK := levelFigure(e.LastPrice)
	threshold, why := parseFigure(thresholdPct)
	if !firstOK || !lastOK || why != "" || first.Sign() <= 0 {
		return false, false
	}
	return gainPct(first, last).Cmp(threshold) > 0, true
}

// RangePosition is where a candidate is trading inside the day's range (task
// 3.7).
//
// # One reading, one instant
//
// The high and the price come from the same observation — the freshest one that
// carried an intraday high. The tempting alternative is the freshest price from
// anywhere against the freshest high from a candle, and it fails in the unsafe
// direction: when the price has fallen since the candle, the distance grows and
// near_high answers "not dangerous" on the strength of a high that is minutes
// old. One reading also makes task 4.7's age limit a single comparison against
// At, rather than a pair of them that could disagree.
type RangePosition struct {
	// Measured and Why carry the same three-state contract as Expansion's, and
	// Reason() is the same guarantee.
	Measured bool
	Why      LevelUnmeasured
	// DistancePct is how far below the day's high the price sits, in percent, as
	// an exact decimal string — truncated at metricScale digits only when the
	// value does not terminate in base ten. This is the quantity near_high's
	// threshold is written against (near_high_threshold_pct: 2.0 — design.md
	// 결정된 계약값), so smaller is more dangerous, and NearHigh is the comparison.
	//
	// It can be zero or negative: a candle's last trade can be its own high, and
	// a price can pass the high a previous candle recorded. That is the most
	// dangerous position there is rather than a missing measurement.
	//
	// Empty means unmeasured. Never "0" — a zero distance is a price sitting
	// exactly on the day's high, which is the opposite of not having looked.
	DistancePct string
	// Position is the literal range position: 0 at the day's low, 1 at its high,
	// as an exact decimal string. PositionMeasured is separate from Measured
	// because a low can be absent on its own — collapsing the two would either
	// suppress the distance (losing the veto's only input) or compute a position
	// against a low of zero, which puts every price at the top of its range.
	//
	// Values above 1 are possible for the same reason DistancePct can be
	// negative.
	Position         string
	PositionMeasured bool
	// High, Low and Price are the source's own strings from the reading this was
	// measured from (D6).
	High, Low, Price string
	// At and Source date the reading. Task 4.7 turns At into the difference
	// between a measurement and a memory: a "not dangerous" built from a
	// three-minute-old candle is the latter.
	At     time.Time
	Source SourceID
}

// Reason names the missing input of an unmeasured range position. See
// Expansion.Reason.
func (r RangePosition) Reason() LevelUnmeasured {
	switch {
	case r.Measured:
		return ""
	case r.Why == "":
		return LevelNotEvaluated
	default:
		return r.Why
	}
}

// NearHigh reports that the price is within `thresholdPct` percent of the day's
// high — the `near_high` veto's comparison (task 4.3).
//
//	near_high = distance_pct < near_high_threshold_pct
//
// Strictly less than, and the boundary is pinned to the digit: at a threshold of
// 2.0, 1.99% is true and 2.00% is false. The comparison is therefore made on
// exact rationals recomputed from the reported high and price, not on the
// rendered DistancePct and never on a float64 — see the file header for the
// 9,598 US cent pairs whose exact distance is 2.00% and whose float64 is
// 1.99999999999999267253, every one of which fires this veto where the contract
// says it must not.
//
// The second return is the D10 three-state: false means unmeasured, and
// unmeasured is not "not near the high". A candidate nobody spent a candle on is
// the ordinary case, all day, for most of the list.
func (r RangePosition) NearHigh(thresholdPct string) (near, measured bool) {
	if !r.Measured {
		return false, false
	}
	high, _, highOK := levelFigure(r.High)
	price, _, priceOK := levelFigure(r.Price)
	threshold, why := parseFigure(thresholdPct)
	if !highOK || !priceOK || why != "" || high.Sign() <= 0 {
		return false, false
	}
	return distancePct(high, price).Cmp(threshold) < 0, true
}

// MeasureExpansion computes one candidate's expansion from its stored baseline
// and its observations.
//
// # The baseline is an argument because it cannot be a slice
//
// This used to take the series alone and pick the earliest priced row in it. That
// is right exactly when the slice is the candidate's whole history, and the store
// offers two documented ways for it not to be:
//
//	PruneObservations   raw rows are retained for two trading days (D11). A
//	                    candidate that outlives the window keeps first_seen_at and
//	                    loses the price it was first seen at.
//	SourceObservations  its own doc comment recommends `since` — "the last two
//	                    minutes of one source rather than two trading days of rows
//	                    to use three of them".
//
// Either one hands over a window, and a baseline picked out of a window is the
// earliest price *in the window*. D17 predicted the result would degrade to
// unmeasured. It did not: the answer came back Measured with an empty reason and
// a re-based number, so a candidate that had tripled reported 50% expansion and
// `extended` read "not extended" for it — on precisely the longest-running
// candidates, the ones most likely to be extended. A wrong number outranks a
// missing one, because D10 stops an unmeasured veto from being read as a pass and
// nothing stops a plausible one.
//
// So the baseline arrives as a Baseline, from the column D17 added, and a caller
// who has not got one says so rather than having one invented from whatever rows
// they were holding. Without it the answer is LevelNoBaseline: unmeasured, named,
// and never a pass.
//
// A caller who genuinely holds a candidate's whole history can still construct a
// Baseline from its earliest priced reading. That is the same measurement it used
// to get by default — the difference is that it is now a claim somebody made
// rather than one the slice made for them.
//
// # Order and ties
//
// The series need not be sorted: both ends are chosen by instant. Readings that
// share an instant are the ordinary case rather than an edge one — scan.go stamps
// a single instant on a whole scan, so every source's row for a symbol carries
// it — and the tie rule is explicit: the baseline keeps the first such reading in
// the caller's order and the latest price keeps the last. For the store's
// (observed_at, id) that is the oldest and the newest inserted row of the
// instant, which is what the two field names say. Choosing strictly on Before and
// After — which is what this did — makes both ends keep the earliest index, so
// LastPrice resolves to the oldest-inserted row of the newest instant.
func MeasureExpansion(baseline Baseline, observations []Observation) Expansion {
	out := Expansion{}
	if baseline.Recorded() {
		out.FirstPrice = strings.TrimSpace(baseline.Price)
		out.FirstAt, out.FirstSource = baseline.At, baseline.Source
		out.BaselineStored = true
	}

	if len(observations) == 0 {
		out.Why = LevelNoObservations
		return out
	}
	if !oneCandidate(observations) {
		out.Why = LevelMixedCandidates
		return out
	}

	first, last := -1, -1
	for i, o := range observations {
		_, present, usable := levelFigure(o.Reported.Price)
		// Absent is not zero. A ranking row that carries no price is the ordinary
		// case and is skipped; having none at all is a different answer, and it is
		// the one returned below.
		if !present {
			continue
		}
		// Present and not a number is a third state, and it is counted rather than
		// returned. One malformed row out of nine sources must not take the other
		// eight with it.
		if !usable {
			out.Unreadable++
			continue
		}
		if first < 0 || o.ObservedAt.Before(observations[first].ObservedAt) {
			first = i
		}
		if last < 0 || !o.ObservedAt.Before(observations[last].ObservedAt) {
			last = i
		}
	}
	if last < 0 {
		// Nothing usable in the series. Which of the two nothings it was decides
		// what an operator does about it: no price at all is the panel's
		// composition, prices that are not numbers is a source contract that moved.
		if out.Unreadable > 0 {
			out.Why = LevelUnreadableDecimal
		} else {
			out.Why = LevelNoPrice
		}
		return out
	}

	out.LastPrice = strings.TrimSpace(observations[last].Reported.Price)
	out.LastAt = observations[last].ObservedAt
	out.LastSource = observations[last].Source

	// A reading older than the stored baseline is evidence the stored one is not
	// the first price. The baseline moves backwards onto it — never forwards, and
	// never onto a reading the store has no record of being the first.
	if !out.BaselineStored || observations[first].ObservedAt.Before(out.FirstAt) {
		out.FirstPrice = strings.TrimSpace(observations[first].Reported.Price)
		out.FirstAt = observations[first].ObservedAt
		out.FirstSource = observations[first].Source
		out.BaselineStored = false
	}
	if !baseline.Recorded() {
		// Nothing was stored, so the fields above carry what the series said and
		// the answer says it did not measure. This is the case that used to return
		// a number.
		out.Why = LevelNoBaseline
		return out
	}

	base, _, baseUsable := levelFigure(out.FirstPrice)
	if !baseUsable {
		out.Why = LevelUnreadableDecimal
		return out
	}
	latest, _, _ := levelFigure(out.LastPrice)
	// A baseline of zero gives +Inf and a negative one inverts the sign, and both
	// clear every threshold they are compared against. D9's rule for the
	// acceleration denominator, on the same arithmetic.
	if base.Sign() <= 0 || latest.Sign() <= 0 {
		out.Why = LevelNonPositivePrice
		return out
	}
	out.Measured = true
	out.Why = ""
	out.Ratio = formatDecimal(new(big.Rat).Quo(latest, base))
	out.GainPct = formatDecimal(gainPct(base, latest))
	return out
}

// MeasureRangePosition computes how far below the day's high a candidate is
// trading, from the freshest reading that carried an intraday high.
//
// Freshest, because a candle costs one call per symbol at five a second: a
// candidate that gets one at all gets it rarely, and the older readings stay in
// the table for the two days raw rows are retained (D11). Measuring from the
// oldest surviving candle would report the morning's range at the close.
//
// Readings that share an instant resolve to the last one in the caller's order,
// the same rule MeasureExpansion's latest price uses and for the same reason: for
// the store's (observed_at, id) that is the newest-inserted row of the instant.
func MeasureRangePosition(observations []Observation) RangePosition {
	if len(observations) == 0 {
		return RangePosition{Why: LevelNoObservations}
	}
	if !oneCandidate(observations) {
		return RangePosition{Why: LevelMixedCandidates}
	}

	// The reading is chosen by the presence of a high, not by its usability: a
	// candle that arrived with an unreadable high is still the freshest thing we
	// have been told, and falling back to an older one would answer with a stale
	// number instead of naming the problem.
	reading := -1
	for i, o := range observations {
		if _, present, _ := levelFigure(o.Reported.DayHigh); !present {
			continue
		}
		if reading < 0 || !o.ObservedAt.Before(observations[reading].ObservedAt) {
			reading = i
		}
	}
	if reading < 0 {
		// The normal case, not an error: the ranking and price endpoints carry no
		// intraday high, and the candle budget reaches a fraction of the list.
		return RangePosition{Why: LevelNoDayHigh}
	}

	o := observations[reading]
	out := RangePosition{
		High:   strings.TrimSpace(o.Reported.DayHigh),
		Low:    strings.TrimSpace(o.Reported.DayLow),
		Price:  strings.TrimSpace(o.Reported.Price),
		At:     o.ObservedAt,
		Source: o.Source,
	}

	high, _, usable := levelFigure(o.Reported.DayHigh)
	if !usable {
		out.Why = LevelUnreadableDecimal
		return out
	}
	if high.Sign() <= 0 {
		out.Why = LevelNonPositiveDayHigh
		return out
	}
	price, present, usable := levelFigure(o.Reported.Price)
	switch {
	case !present:
		out.Why = LevelNoPrice
		return out
	case !usable:
		out.Why = LevelUnreadableDecimal
		return out
	case price.Sign() <= 0:
		out.Why = LevelNonPositivePrice
		return out
	}

	out.Measured = true
	out.DistancePct = formatDecimal(distancePct(high, price))

	// The low is a separate measurement with a separate outcome. Absent,
	// unreadable, non-positive or not actually below the high — a halted symbol,
	// a limit-up print, the session's first candle — all leave the position
	// unmeasured while the distance from the high stands.
	if low, lowPresent, lowUsable := levelFigure(o.Reported.DayLow); lowPresent && lowUsable &&
		low.Sign() > 0 && high.Cmp(low) > 0 {
		out.Position = formatDecimal(new(big.Rat).Quo(
			new(big.Rat).Sub(price, low), new(big.Rat).Sub(high, low)))
		out.PositionMeasured = true
	}
	return out
}

// distancePct is (high − price) ÷ high × 100, exactly.
//
// The caller has already established that high is positive. It is a function
// rather than three lines inlined twice because MeasureRangePosition and NearHigh
// have to agree to the last digit — the first renders the number a person reads
// and the second decides the veto, and a veto that disagrees with the number
// printed beside it is worse than either of them being wrong alone.
//
// This is the one value in the package that can be negative, and formatDecimal
// truncates towards zero rather than towards minus infinity — so a rendered
// negative distance is up to one unit in the last place *closer* to zero than the
// exact one. It cannot move the veto, because NearHigh compares the exact
// rational this returns and not the rendering, and a negative distance is a price
// past the day's high: below every threshold either way.
func distancePct(high, price *big.Rat) *big.Rat {
	return new(big.Rat).Mul(
		new(big.Rat).Quo(new(big.Rat).Sub(high, price), high),
		big.NewRat(100, 1))
}

// gainPct is (last ÷ first − 1) × 100, exactly. Same contract as distancePct,
// between MeasureExpansion and GainExceeds.
func gainPct(first, last *big.Rat) *big.Rat {
	return new(big.Rat).Mul(
		new(big.Rat).Sub(new(big.Rat).Quo(last, first), big.NewRat(1, 1)),
		big.NewRat(100, 1))
}

// oneCandidate reports that every observation belongs to the same (market,
// symbol).
//
// The functions here take one candidate's series, as Store.Observations returns
// it. A caller that hands over two symbols' rows would otherwise get a confident
// number computed from one symbol's baseline and another's current price, and
// nothing about it would look wrong. It is reported as unmeasured rather than as
// an error because unmeasured is the one state that is never read as safe (D10),
// and a metric that panicked or returned an error would push a wiring defect into
// a path whose job is to keep scanning.
func oneCandidate(observations []Observation) bool {
	key := observations[0].Key()
	for _, o := range observations[1:] {
		if o.Key() != key {
			return false
		}
	}
	return true
}

// levelFigure reads one of the source's decimal strings.
//
// Three outcomes rather than two, because the two failures have different
// remedies and the store keeps them apart all the way from the nullable column:
//
//	present=false               the source did not carry the field. Ordinary.
//	present=true, usable=false  it carried something that is not a number.
//	both true                   v is the value.
//
// It is metrics.go's parseFigure with that shape put back on it. Sharing the
// parser is the point rather than a convenience: the grammar it accepts is
// internal/riskcalc's, character for character, so a string this package will
// measure and a string the acceleration will measure are the same set. The
// previous reading here was strconv.ParseFloat, which accepts "NaN", "Inf" and
// "Infinity" without an error and returns a value that clears every threshold it
// is compared against — a maximal signal manufactured out of a broken field.
func levelFigure(s string) (v *big.Rat, present, usable bool) {
	value, why := parseFigure(s)
	switch why {
	case NotComputedFigureAbsent:
		return nil, false, false
	case "":
		return value, true, true
	default:
		return nil, true, false
	}
}
