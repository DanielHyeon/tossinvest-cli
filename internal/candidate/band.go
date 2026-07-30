package candidate

// band.go is task 4.10: the shadow record for the two vetoes that have no
// threshold.
//
// # Why two of three vetoes record instead of deciding
//
// D18 looked for `seen_late`'s and `extended`'s thresholds and found that they do
// not exist — not in design.md's 결정된 계약값, not in openspec/, not in docs/, not
// in any config. The two survive in the contract as the sentences "임계를 크게
// 초과" and "상한 초과", and D6 forbids a policy number with no source, applied
// market or verification state. Inventing one here would make it look like a number
// somebody chose, to the next person who reads it.
//
// So the rule §3.5 already applies to the acceleration is applied to these two:
// record the crossings against a band of values and judge nothing. After a month of
// real sessions a person can approve a threshold from the distribution, and that is
// what turns "we have no threshold" into "we will have one".
//
// # A band is an instrument and a veto threshold is a decision
//
// The distinction is the whole of D18 and it is what the rest of this file is built
// to keep. Bands may be chosen without a source precisely because nothing is decided
// by them, and the moment one decides something it is a policy number that was never
// approved — arriving through the back door D6 closed at the front.
//
// Four things hold the line, and none of them is a comment:
//
//	no verdict vocabulary   ShadowBand has no Dangerous, Clear, Raised or Passed.
//	                        The one-line mistake the §4 review left for section 5 —
//	                        `if !x.Dangerous() { score += 2 }` — has no spelling
//	                        here to make.
//	a different comparison  a crossing is `>=`, which is `crossings()`'s comparison
//	                        for the shadow acceleration and deliberately not the
//	                        veto's `>`. A band edge is a histogram bucket, not a
//	                        boundary, and one that used the veto's comparison would
//	                        be a substitution away from being one.
//	out of the closure      the band types are in TestTheVetoCannotSeeAScoreToBeOffsetBy's
//	                        score set, so a band cannot be reached from VetoInputs
//	                        by any path.
//	out of the producers    TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
//	                        fails if any function returning a VetoState or a Chase
//	                        so much as mentions a band, so there is no conversion
//	                        to call.
//
// # But the inputs are checked exactly as the veto checks them
//
// The band records; it does not relax. A baseline from a life that ended, a
// baseline later than the first sighting, a latest price too old to be about now —
// every one of those produces a plausible wrong number, and this is the dataset a
// threshold will be derived from. A month of wrong numbers is worse than a month of
// named gaps, which is MeasureExpansion's own rule. The two chains are written
// separately (the veto's decides, this one records) and
// TestABandNamesTheSameMissingInputTheVetoWould pins them together, the way
// fsguard_drift_test.go pins the filesystem allowlist to the ledger's.

import (
	"fmt"
	"math/big"
	"sort"
	"time"
)

// SeenLatePercentileBands are the positions `seen_late` is recorded against, in
// percentile points of the list that carried the first sighting.
//
// Five values spanning the top half, because that is where the question lives: D8's
// example contrasts a symbol entering at 12th with one entering at 148th, and every
// interesting first sighting is above the middle. 50 is "already in the better half
// when we arrived", 95 is "we were late by any reading". They are chosen for the
// shape of the distribution they will show, which is what a band is allowed to be
// chosen for.
var SeenLatePercentileBands = []string{"50", "70", "80", "90", "95"}

// ExtendedGainBands are the gains `extended` is recorded against, in percent since
// the first price recorded for the candidate.
//
// # What the first five were, and why five were not enough
//
// The scale began as 10·20·30·50·100: one ordinary intraday move to a doubling. 10
// is inside a normal day's range for a name that is on a trading-value ranking at
// all; 100 is the case D17 uses throughout as the one that must not read as "has not
// run". Those five are still here, with the same values in the same order, because
// they have been reported against — `crossed 10 0 · 20 0 · 30 0 · 50 0 · 100 0` has
// been read by people, and moving one would make a column name mean something new in
// their memory for a record that has no migration on disk.
//
// The measurement of 2026-07-28 is why more were needed: 131 candidates measured,
// and all 131 crossed nothing. The scale's finest edge was 10 and `bandCrossings`
// compares with `>=`, so everything below 10% — which on an ordinary day is
// everything — was one bucket. The tally's arithmetic held and the measured count
// was honest; what did not exist was a distribution anybody could pick a threshold
// out of. A record that answers no question is still reported as a record.
//
// # Why these new edges
//
//	0            The sign change, and the only value here that was not chosen.
//	             Arithmetic put it there: expansionBandValue returns gainPct, which
//	             is negative for a candidate below the price it was first seen at,
//	             and with every edge at 10 or above a candidate 20% down and a
//	             candidate 9% up left the identical record. Both are "no chase risk"
//	             and they are different facts. `>=` means a candidate exactly at its
//	             first price crosses it, so what 0 divides is "declined" from "did
//	             not decline".
//	2 4 6 8      A regular grid across 0–10, not a list of interesting values. The
//	             spacing is what was chosen and its basis is the sentence above: if
//	             10 is one ordinary day, five steps inside one ordinary day is 2
//	             percentage points. No edge here was picked for what it is.
//
// # 6 is on the grid, and a threshold candidate under discussion is also 6
//
// The two are the same number by coincidence — 6 is a multiple of 2 — and the
// coincidence is harmless only because a scale value cannot reach a verdict. That is
// not a claim this comment makes on its own behalf. It was false until this change:
// on 2026-07-28 a reviewer put `if v.ExtendedBand.Crossed("6") { v.Chase.Extended =
// RaisedVeto() }` into assessOne and every one of the four structural guards, `go
// vet` and 51 packages of tests passed. The guards read only functions returning a
// VetoState or a Chase, and assessOne returns a Verdict.
//
// What closed it is in this change and not in this sentence: the band statements
// left assessOne for shadowBandsFor, whose scope has no Chase in it, and
// TestNoFunctionThatProducesAVerdictCanSeeAShadowBand now reads Verdict results,
// pointer/slice/map results, the crossing vocabulary and any read of a verdict's own
// band fields. Putting that line back is a failing test.
//
// # The spacing is provisional
//
// 2 percentage points is a guess about where the mass is. Nobody has seen the 131
// values themselves — only that they all fell below 10 — which is why BandTally now
// carries the quantiles of Value. When a session's distribution is in, the spacing
// is decided from it rather than from this paragraph, and if the mass turns out to
// sit in 0–2 then this scale collapses the same way the last one did.
var ExtendedGainBands = []string{"0", "2", "4", "6", "8", "10", "20", "30", "50", "100"}

// BandsFor returns the bands recorded for a veto code, empty for a code that has
// none.
//
// near_high has none on purpose: D18 approved its threshold, so it is a veto and a
// shadow record beside a live veto would be an invitation to read the shadow
// instead.
func BandsFor(code VetoCode) []string {
	switch code {
	case VetoSeenLate:
		return SeenLatePercentileBands
	case VetoExtended:
		return ExtendedGainBands
	default:
		return nil
	}
}

// BandCrossing is one band and whether a measured value reached it.
//
// Crossings exist only on a measured band. An unmeasured one carries none, which is
// what stops "we could not measure this" from being counted as "this stayed below
// everything" — the counting form of D10, one level down from the veto.
type BandCrossing struct {
	Band    string
	Crossed bool
}

// ShadowBand is one candidate's shadow record for one veto code.
//
// It is deliberately not a VetoState. The three-state type answers a question
// ("is this dangerous") that this one is not allowed to answer, and giving the
// record that shape would make the substitution a one-word edit.
type ShadowBand struct {
	// Code is which veto this records for. A tally keyed by a different code counts
	// it as unmeasured rather than reading a percentile as a gain.
	Code VetoCode
	// Measured is false for every unmeasured record, and it is the zero value so
	// that an unassigned ShadowBand — a map miss, an unfilled slot in
	// make([]ShadowBand, n) — cannot read as a candidate somebody looked at.
	Measured bool
	// Why names the missing input, under the same name the veto would use for it.
	// Reason() supplies NOT_EVALUATED when nobody measured anything at all.
	Why VetoUnmeasured
	// Value is the quantity the bands were compared against, as an exact decimal
	// string: percentile points for seen_late, percent gained for extended. Empty
	// on an unmeasured record — an unmeasured band carries no arithmetic, never a
	// "0".
	Value string
	// Crossings is the record itself, one entry per band of this code.
	Crossings []BandCrossing
}

// Reason names the missing input of an unmeasured band and never comes back empty
// for one. Empty exactly when Measured. Same guarantee as VetoState.Reason.
func (b ShadowBand) Reason() VetoUnmeasured {
	switch {
	case b.Measured:
		return ""
	case b.Why == "":
		return VetoNotEvaluated
	default:
		return b.Why
	}
}

// Crossed reports that a measured value reached the named band.
//
// It is the only predicate on this type. An unmeasured band crosses nothing,
// because it measured nothing — so "crossed nothing" and "we never looked" answer
// the same here and differ on Measured, which is what forces a caller to read
// Measured rather than inferring calm from a row of falses.
func (b ShadowBand) Crossed(band string) bool {
	if !b.Measured {
		return false
	}
	for _, c := range b.Crossings {
		if c.Band == band {
			return c.Crossed
		}
	}
	return false
}

// MeasureSeenLateBand records where a candidate stood the first time we saw it,
// against SeenLatePercentileBands.
//
// The percentile is recomputed from Rank and RankTotal on the exact rational rather
// than read off Sighting.PercentilePct, for the reason PercentileExceeds gives: a
// rendering must not be what decides which bucket a candidate lands in, even when
// nothing downstream decides anything.
func MeasureSeenLateBand(s Sighting) ShadowBand {
	out := ShadowBand{Code: VetoSeenLate}
	value, why := sightingBandValue(s)
	if why != "" {
		out.Why = why
		return out
	}
	out.Measured = true
	out.Value = formatDecimal(value)
	out.Crossings = bandCrossings(value, SeenLatePercentileBands)
	return out
}

// sightingBandValue is the first sighting's percentile, or the reason there is
// none.
//
// The three impossible readings PercentileExceeds refuses are refused here too — a
// rank of zero, a list of zero length, and a rank past the end of its own list,
// which makes the percentile negative and would file a symbol that cannot be where
// it says it is into the lowest bucket.
func sightingBandValue(s Sighting) (*big.Rat, VetoUnmeasured) {
	if !s.Measured {
		return nil, s.Reason()
	}
	if s.RankTotal <= 0 || s.Rank <= 0 || s.Rank > s.RankTotal {
		return nil, VetoNotRanked
	}
	return percentileOf(s.Rank, s.RankTotal), ""
}

// MeasureExtendedBand records how far a candidate has run since its first price,
// against ExtendedGainBands.
//
// The signature is AssessExtended's because the checks are AssessExtended's: the
// candidate summary dates the baseline, the instant dates the latest price, and the
// thresholds carry the input-age ceiling.
//
// # Exactly one field of VetoThresholds is unread
//
// The thresholds argument is not an unused parameter and describing it as one is
// wrong. th.inputAge() is read on the line below and it decides whether a latest
// price is recent enough to be about now — a check the band applies for
// expansionBandReason's stated reason, that a plausible wrong number in this dataset
// is worse than a named gap.
//
// What is unread is ExtendedGainPct, the threshold itself, and that single omission
// is the design: the value is recorded and nothing is compared against a number
// nobody approved. TestMeasureExtendedBandNeverReadsTheVetoThreshold fails if it is
// ever read, by source and by behaviour, because until that test existed reading it
// broke nothing.
func MeasureExtendedBand(e Expansion, c Candidate, at time.Time, th VetoThresholds) ShadowBand {
	out := ShadowBand{Code: VetoExtended}
	if why := expansionBandReason(e, c, at, th.inputAge()); why != "" {
		out.Why = why
		return out
	}
	value, ok := expansionBandValue(e)
	if !ok {
		// The expansion said it was measured and its own figures are not numbers,
		// which is what level.go calls UNREADABLE_DECIMAL. vetoFrom passes the same
		// fallback for the same reason: NOT_EVALUATED would say nobody looked when
		// somebody did.
		out.Why = VetoUnmeasured(LevelUnreadableDecimal)
		return out
	}
	out.Measured = true
	out.Value = formatDecimal(value)
	out.Crossings = bandCrossings(value, ExtendedGainBands)
	return out
}

// expansionBandReason is AssessExtended's input chain without the threshold.
//
// It is written out rather than shared with AssessExtended because the two have
// different jobs and one of them must never grow the ability to decide. The cost of
// the duplication is drift, and it is paid the way fsguard.go pays it:
// TestABandNamesTheSameMissingInputTheVetoWould walks the same input shapes through
// both and fails when the reasons stop agreeing.
//
// Every refusal here is a case where the arithmetic would still produce a number.
// That is why the band applies them: a baseline from a life that ended reports a
// candidate which has doubled as −33%, and a month of those in the dataset is how a
// threshold gets approved against a distribution that never happened.
func expansionBandReason(e Expansion, c Candidate, at time.Time, maxAge time.Duration) VetoUnmeasured {
	switch {
	case !e.Measured:
		return VetoUnmeasured(e.Reason())
	case c.FirstSeenAt.IsZero():
		return VetoNoCandidate
	case e.FirstAt.IsZero():
		return VetoBaselineUndated
	}
	if gap := e.FirstAt.Sub(c.FirstSeenAt); gap >= DefaultStalenessTTL {
		return VetoBaselineTooLate
	} else if -gap >= DefaultStalenessTTL {
		return VetoBaselineTooEarly
	}
	return inputAgeReason(at, e.LastAt, maxAge)
}

// expansionBandValue is the gain in percent, recomputed exactly from the two
// reported prices. GainExceeds' rule, minus the comparison.
func expansionBandValue(e Expansion) (*big.Rat, bool) {
	first, _, firstOK := levelFigure(e.FirstPrice)
	last, _, lastOK := levelFigure(e.LastPrice)
	if !firstOK || !lastOK || first.Sign() <= 0 {
		return nil, false
	}
	return gainPct(first, last), true
}

// bandCrossings records a value against every band, on the exact rational.
//
// The comparison is `>=`, which is crossings()' comparison for the shadow
// acceleration and is deliberately not the `>` the two vetoes are written with. See
// the file header: a band edge is a bucket, and one that shared the veto's
// comparison would be a substitution away from being the veto.
func bandCrossings(value *big.Rat, bands []string) []BandCrossing {
	out := make([]BandCrossing, 0, len(bands))
	for _, band := range bands {
		bar, ok := new(big.Rat).SetString(band)
		if !ok {
			continue
		}
		out = append(out, BandCrossing{Band: band, Crossed: value.Cmp(bar) >= 0})
	}
	return out
}

// --- the tally ---------------------------------------------------------------------

// BandTally is the shadow record over a set of candidates for one code.
//
// Its shape is CrossingTally's, including the invariant that made that one
// trustworthy: Total == Measured + Σ NotMeasured holds for every tally, so a
// candidate that fell out of both halves shows up as arithmetic that does not add up
// rather than as a number that is quietly short. The §3 review found exactly that
// bug in TallyCrossings, and its effect was to flip a screen's default reading from
// "mostly unchecked" to "all measured and quiet".
type BandTally struct {
	// Code is the veto this tally records for.
	Code VetoCode
	// Total is how many records were tallied, measured or not.
	Total int
	// Measured is how many had a value. It is not the sum of Crossed: a measured
	// record below every band is counted here and nowhere in Crossed, and one above
	// three bands is counted once here and three times there.
	Measured int
	// Crossed counts, per band, the measured records that reached it. Its keys are
	// exactly BandsFor(Code) even at zero, so a column that is missing cannot read
	// as a column nobody needed.
	Crossed map[string]int
	// NotMeasured counts the rest by the input that was missing.
	NotMeasured map[VetoUnmeasured]int
	// Quantiles are positions in the distribution of the values themselves, in
	// BandQuantilePoints' order. Empty when nothing was measured.
	//
	// They exist because the crossings alone could not answer the question that
	// produced this change. On 2026-07-28 the extended tally read `measured 131 of
	// 156, crossed 10/20/30/50/100 all 0`, and from that nobody could say whether
	// the 131 values sat at +9 or at −1: ShadowBand.Value is computed for every
	// candidate and appeared on no surface at all. A scale chosen against crossings
	// that all agree is a scale chosen against no distribution.
	Quantiles []BandQuantile
	// QuantileBase is how many values the quantiles were computed from.
	//
	// It is reported rather than assumed equal to Measured. They are equal whenever
	// every measured value parses, which is every value formatDecimal produces — and
	// a day when they differ is a day the difference has to be visible, not a day
	// the median quietly describes a subset.
	QuantileBase int
}

// BandQuantile is one position in the distribution of a tally's measured values.
type BandQuantile struct {
	// Name is the position, as BandQuantilePoints names it.
	Name string
	// Value is the value at that position, an exact decimal string like
	// ShadowBand.Value. It is one of the measured values and never an average of
	// two, so it can be negative and it is always a number something actually was.
	Value string
}

// BandQuantilePoints are the positions reported, in the order they are reported.
//
// Five, spanning the whole range, for the same reason a band scale spans one: the
// question a reader has is where the mass is, and min/max alone answer it only when
// the distribution is uniform. The pairs are percent positions and are applied by
// nearest rank — no interpolation, so every figure printed is a value some candidate
// actually had rather than one halfway between two candidates that do not exist.
var BandQuantilePoints = []struct {
	Name string
	Pct  int
}{
	{"min", 0}, {"p25", 25}, {"median", 50}, {"p75", 75}, {"max", 100},
}

// bandQuantiles is BandQuantilePoints applied to a set of values by nearest rank.
//
// Nearest rank is `ceil(p/100 × n)`, one-based, clamped into the slice. The
// alternative — interpolating between neighbours — would print a number no candidate
// had, and this is the surface a threshold gets chosen from.
func bandQuantiles(values []*big.Rat) []BandQuantile {
	if len(values) == 0 {
		return nil
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Cmp(values[j]) < 0 })
	out := make([]BandQuantile, 0, len(BandQuantilePoints))
	for _, point := range BandQuantilePoints {
		// (p×n + 99) / 100 is ceil(p×n/100) in integer arithmetic.
		rank := (point.Pct*len(values) + 99) / 100
		if rank < 1 {
			rank = 1
		}
		if rank > len(values) {
			rank = len(values)
		}
		out = append(out, BandQuantile{Name: point.Name, Value: formatDecimal(values[rank-1])})
	}
	return out
}

// Collapsed reports that every measured record produced the same crossings, so this
// tally's scale resolved nothing.
//
// It is the arithmetic of the failure that produced this change, made checkable.
// `measured 131 of 156, crossed 10/20/30/50/100 all 0` is a row where the invariants
// hold, the measured count is honest and nothing is wrong except the only thing the
// record is for: there is no distribution in it to approve a threshold from. That
// state was found by a person reading the same line twice, and a state that depends
// on somebody noticing is a state that is normally not noticed.
//
// The test needs no idea of what the scale was meant to resolve: if a band was
// crossed by some measured records and not others, the scale separated something.
// Every count being 0 or Measured means it separated nothing.
//
// One measurement counts as collapsed, because one measurement is not a
// distribution either.
//
// A code with no scale is not collapsed. BandsFor returns nil for a veto whose
// threshold a person has already approved — near_high today, and extended on the day
// issues I3 is carried out — and TallyBands then seeds Crossed from an empty list,
// so the loop below runs zero times and "every band was crossed by all or none"
// would be vacuously true of any tally with one measured record in it. The alarm
// would light on that code and never go out, and an alarm that is always on is what
// teaches a reader to skip the block it is in. A scale that does not exist did not
// fail to resolve anything.
//
// It says nothing about any candidate. It is a statement about the instrument, which
// is why it may exist on a tally while ShadowBand carries no predicate but Crossed.
func (t BandTally) Collapsed() bool {
	if t.Measured < 1 || len(t.Crossed) == 0 {
		return false
	}
	for _, n := range t.Crossed {
		if n != 0 && n != t.Measured {
			return false
		}
	}
	return true
}

// CollapsedAlarm is the sentence a collapsed tally is reported with, empty when the
// scale resolved something.
//
// It carries the numbers that produced it for the reason candidateReport.Alarms
// gives: a consumer told only that something is inconsistent has nothing to check.
func (t BandTally) CollapsedAlarm() string {
	if !t.Collapsed() {
		return ""
	}
	return fmt.Sprintf("ALARM %s: all %d measured record(s) produced the same crossings, so "+
		"this scale resolved nothing and no threshold may be approved from it. The counts "+
		"beside it are honest and answer no question", string(t.Code), t.Measured)
}

// TallyBands counts a set of shadow records for one code.
//
// A record of a different code is counted as unmeasured under NOT_EVALUATED rather
// than dropped or measured: nobody measured *this* code for that candidate, and the
// alternative — letting a seen_late percentile land in an extended gain's buckets —
// is a wiring defect that would produce a completely plausible histogram.
//
// It also carries the distribution of the values, not only of the crossings. The
// crossings answer "how many were above each edge" and the whole failure this change
// repairs is a set of edges against which every answer was the same; the quantiles
// answer "where were they" and do not depend on the edges being right.
func TallyBands(code VetoCode, in []ShadowBand) BandTally {
	bands := BandsFor(code)
	out := BandTally{
		Code:        code,
		Total:       len(in),
		Crossed:     make(map[string]int, len(bands)),
		NotMeasured: map[VetoUnmeasured]int{},
	}
	for _, band := range bands {
		out.Crossed[band] = 0
	}
	seen := make(map[string]bool, len(bands))
	values := make([]*big.Rat, 0, len(in))
	for _, b := range in {
		if !b.Measured || b.Code != code {
			why := b.Reason()
			if b.Measured {
				why = VetoNotEvaluated
			}
			out.NotMeasured[why]++
			continue
		}
		out.Measured++
		// The rendered value and not the exact rational, on purpose: the quantiles
		// describe the numbers a person reads, and formatDecimal truncates a
		// non-terminating one towards zero. The difference is one unit in the last of
		// twelve fractional digits and it cannot reach a crossing, which bandCrossings
		// decides on the exact value.
		//
		// A measured record always carries a Value and formatDecimal always produces a
		// decimal, so the refusal below is unreachable today. It is written as a
		// refusal rather than an assumption because the count it feeds is reported
		// separately: a value that stopped parsing would shorten the base rather than
		// move the median, and the shorter base is on the screen.
		if v, ok := new(big.Rat).SetString(b.Value); ok {
			values = append(values, v)
		}
		for k := range seen {
			delete(seen, k)
		}
		for _, c := range b.Crossings {
			// Membership in Crossed is membership in BandsFor(code): the map was
			// seeded with exactly those keys, and a band this record does not shadow
			// must not become one by being counted. Duplicates count once.
			if _, shadowed := out.Crossed[c.Band]; !shadowed || seen[c.Band] {
				continue
			}
			seen[c.Band] = true
			if c.Crossed {
				out.Crossed[c.Band]++
			}
		}
	}
	out.QuantileBase = len(values)
	out.Quantiles = bandQuantiles(values)
	return out
}
