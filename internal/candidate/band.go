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
	"math/big"
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
// Five values spanning one ordinary intraday move to a doubling. 10 is inside a
// normal day's range for a name that is on a trading-value ranking at all; 100 is
// the case D17 uses throughout as the one that must not read as "has not run".
var ExtendedGainBands = []string{"10", "20", "30", "50", "100"}

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
// thresholds carry the input-age ceiling. Only the threshold itself is unused, and
// that is the point — the value is recorded and nothing is compared against a number
// nobody approved.
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
}

// TallyBands counts a set of shadow records for one code.
//
// A record of a different code is counted as unmeasured under NOT_EVALUATED rather
// than dropped or measured: nobody measured *this* code for that candidate, and the
// alternative — letting a seen_late percentile land in an extended gain's buckets —
// is a wiring defect that would produce a completely plausible histogram.
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
	return out
}
