package candidate

// veto.go is section 4: the chase-risk veto.
//
// # A veto, not a term in a sum
//
// D3 keeps chase risk out of the score, because a sum lets one axis buy off
// another: "the trading value is accelerating so hard that sitting on the day's
// high is worth it" is not a trade-off, it is the argument for chasing written as
// arithmetic. So the answer here is a separate verdict with a reason code, and
// nothing in this file carries a number anybody could add to anything — Chase,
// VetoState and VetoInputs are pinned field by field by
// TestTheVetoCannotSeeAScoreToBeOffsetBy, which fixes that by inspecting the
// types rather than by inspecting today's call sites. The acceleration is not an
// input to this file at all: there is no offset to refuse because there is
// nothing here to offset with.
//
// # Three states, because one of the other two would be a lie
//
// The intraday high is not in the ranking response and not in the prices
// response. Only candles carry it, one call per symbol at five calls a second, so
// under a list of hundreds most candidates never get one, all day (D13 decision
// 3). That makes "we did not measure this" the ordinary case rather than the edge
// one, and it is why every veto here is three-state:
//
//	raised      measured, and dangerous
//	clear       measured, and not dangerous
//	unmeasured  we did not look, or what we looked at cannot bear the question
//
// D10 is the rule this file exists to enforce: an unmeasured veto is never a
// pass. Store `false` for the candidates nobody spent a candle on and the veto
// switches itself off for most of the list — the record says "no chase risk"
// where the truth is "never checked", and nothing fails while that happens. That
// is the second route around D3's veto and the harder one to find, because the
// first one (D1's re-entry rule) at least moves a stored timestamp.
//
// So Passed is true only when all three are measured AND all three are clear, and
// the types are built so that no accident produces a clear:
//
//	the zero VetoState is unmeasured           its fields are unexported and
//	                                           "measured" is not the zero of a bool
//	the zero Chase does not pass               Passed asks each veto for Clear,
//	                                           which requires the measurement
//	a map miss and an unfilled slice slot      are that zero value
//	the (value, measured) pairs level.go       are read in exactly one place —
//	returns                                    vetoFrom — so `near, _ :=` has no
//	                                           call site to be written at
//	an absent threshold                        is unmeasured, not unbounded
//	a threshold of "0"                         is an absent threshold in its second
//	                                           spelling (D18), not a comparison
//	                                           every candidate is on the safe side
//	                                           of — thresholdReason
//	an absent input-age limit                  is the default, not no limit
//	an unrecognised veto code                  is unmeasured, not the zero value
//	a veto code removed from VetoCodes         cannot be, because the list is an
//	                                           array and a caller gets a copy
//
// # Three inputs that are not about this candidate, or not about now
//
// Each of the three vetoes can be computed from a reading that belongs to another
// moment or another life, and every one of those errors points the same way —
// towards `false`.
//
//	a baseline that postdates first sighting   understates the expansion, so
//	                                           `extended` reads "has not run".
//	                                           D17's migration backfill and
//	                                           PruneObservations both produce it.
//	a baseline from a life that ended          overstates it, so `extended` reads
//	                                           "has not run" for a candidate that
//	                                           has doubled. Raw rows outlive a
//	                                           life by 48h and the baseline moves
//	                                           backwards onto them.
//	a ranked row from between two lives        was the first sighting until task
//	                                           4.9. A symbol promoted 4th of 150
//	                                           read as one caught at 148th, and
//	                                           `seen_late` cleared.
//	a reading from three minutes ago           the price has moved since, so
//	                                           `near_high` reads "there is room"
//	                                           and `extended` reads "it has not
//	                                           run". tasks.md 4.7.
//
// All of them are refused as unmeasured. A "not dangerous" built from a memory —
// or from a candidate that is not this one — is not a measurement, and the
// distinction disappearing is the failure above reproduced one level down, where it
// looks identical on a screen.

import (
	"math/big"
	"time"
)

// VetoCode is one of D3's three reasons, in the spelling the record and the
// screen use.
//
// They are separate codes rather than one "chase risk" flag because they are
// fixed by different things, and a screen that cannot tell them apart cannot say
// what to fix: many seen_late means the scan interval or the source panel is
// late, many extended means the threshold is wrong.
type VetoCode string

const (
	// VetoSeenLate: the first time we ever saw this symbol it was already well up
	// its list. We arrived late and the early part of this move happened without
	// us — a limit of the observation system rather than of the entry.
	VetoSeenLate VetoCode = "seen_late"
	// VetoExtended: we may have seen it early, and it has run far enough since
	// that entering now is late. A statement about the entry rather than about us.
	VetoExtended VetoCode = "extended"
	// VetoNearHigh: the distance to the day's high is below the floor, so there is
	// no room left for the pullback an entry would need.
	VetoNearHigh VetoCode = "near_high"
)

// VetoCodes is the three in D3's order, so every tally and every screen lists
// them the same way.
//
// Chase.Passed, Vetoed, HasUnmeasured and TallyVetoes all iterate this rather
// than naming the three fields, so a fourth code is refused by Chase.State —
// unmeasured, therefore never a pass — until somebody gives it a field.
//
// It is an array rather than a slice because that iteration cuts both ways and
// only one of the two directions was guarded (§4 review P1-3). A fourth code is
// refused; a code *removed* from the list stops being consulted, and every
// predicate here then answers as though its veto did not exist — `VetoCodes = nil`
// makes the zero Chase pass, and overwriting one entry makes a Chase with that
// veto RAISED pass. An exported slice is a shared backing array any caller can
// write through. An array is a value: `range` and `len` are unchanged, and a
// caller who takes it gets a copy.
var VetoCodes = [3]VetoCode{VetoSeenLate, VetoExtended, VetoNearHigh}

// VetoUnmeasured names the input a veto could not be measured from.
//
// A bare "unmeasured" is enough for the verdict and not enough for anybody
// deciding what to fix, which is the distinction level.go draws with
// LevelUnmeasured and D9 draws between WARMING_UP and an unusable denominator. A
// screen full of NO_DAY_HIGH is the candle budget working as designed; a screen
// full of INPUT_TOO_OLD says the hot queue is not turning over; a screen full of
// THRESHOLD_ABSENT says nobody configured the veto at all.
//
// level.go's reasons travel through unchanged rather than being folded into one
// "the input was missing" — NO_DAY_HIGH, NO_BASELINE and UNREADABLE_DECIMAL
// arrive at the verdict under the names they were measured with.
type VetoUnmeasured string

const (
	// VetoNotEvaluated: nobody ran this veto at all. It is the reason the zero
	// value carries, so an unassigned VetoState names itself rather than rendering
	// as "unmeasured ()" — a reason code that is sometimes empty is a reason code
	// an operator learns to skip.
	VetoNotEvaluated VetoUnmeasured = "NOT_EVALUATED"
	// VetoNoCandidate: there is no candidate summary, so there is no first_seen_at
	// to judge an input's lateness against. Both seen_late and extended need it —
	// each of them is a question about whether the reading they are measured from
	// is the one they claim to be.
	VetoNoCandidate VetoUnmeasured = "NO_CANDIDATE"
	// VetoNoObservations: nothing was recorded for this candidate at all.
	VetoNoObservations VetoUnmeasured = "NO_OBSERVATIONS"
	// VetoMixedCandidates: the readings carry more than one (market, symbol). A
	// wiring defect, refused rather than answered — see oneCandidate.
	VetoMixedCandidates VetoUnmeasured = "MIXED_CANDIDATES"
	// VetoNotRanked: the first sighting carried no list position, so there is no
	// "where in the list did it enter" to answer. The prices and candles sources
	// are not rankings and never carry one, so a candidate raised only by them has
	// no answer rather than a bottom-of-the-list one.
	VetoNotRanked VetoUnmeasured = "NOT_RANKED"
	// VetoNoFirstSighting: no reading we still hold could be this life's first
	// sighting — every ranked row falls outside the window around first_seen_at.
	// Raw rows are pruned after two trading days (D11) and callers window their
	// reads, so this is the ordinary way to lose it.
	VetoNoFirstSighting VetoUnmeasured = "NO_FIRST_SIGHTING"
	// VetoNoFirstRank: the candidate summary carries no first sighting position,
	// while the rows we hold do include one that could have been it.
	//
	// It is separate from NO_FIRST_SIGHTING because the remedies are: that one is
	// retention doing what it is supposed to, this one is a store from before
	// schema v3 or a caller that never wrote the column. Answering it from the row
	// instead would be the substitution D17 closed for the price, where the slice's
	// answer came back measured, re-based and unremarkable.
	VetoNoFirstRank VetoUnmeasured = "NO_FIRST_RANK"
	// VetoFirstRankUndated: a first rank is stored with no instant, so whether it
	// is this life's cannot be judged. BASELINE_UNDATED's twin.
	VetoFirstRankUndated VetoUnmeasured = "FIRST_RANK_UNDATED"
	// VetoFirstRankNotFirst: a first rank is stored and its instant is further than
	// the staleness TTL from first_seen_at, so it belongs to another moment — a
	// previous life's, or one long after this life began. NoteFirstRank applies the
	// same window on the way in, so this is a migration, a fixture or a hand-edited
	// file rather than a scan.
	VetoFirstRankNotFirst VetoUnmeasured = "FIRST_RANK_NOT_FIRST"
	// VetoBaselineTooLate: the stored first price postdates first_seen_at by more
	// than the staleness TTL, so it is not the first price and every expansion
	// measured from it is an understatement (D17).
	VetoBaselineTooLate VetoUnmeasured = "BASELINE_TOO_LATE"
	// VetoBaselineTooEarly: the stored first price predates first_seen_at by more
	// than the staleness TTL, so it belongs to a life that ended. Raw rows outlive
	// a candidate by two trading days (D11) and MeasureExpansion moves the baseline
	// backwards onto any older priced row, so this is how a dead life's price
	// becomes a live one's denominator — and it overstates the baseline, which
	// reads as "has not run" for a candidate that has doubled.
	VetoBaselineTooEarly VetoUnmeasured = "BASELINE_TOO_EARLY"
	// VetoBaselineUndated: there is a baseline and no instant for it, so its
	// lateness cannot be judged at all.
	VetoBaselineUndated VetoUnmeasured = "BASELINE_UNDATED"
	// VetoInputTooOld: the reading this veto would be computed from is older than
	// the input-age ceiling. See MaxVetoInputAge.
	VetoInputTooOld VetoUnmeasured = "INPUT_TOO_OLD"
	// VetoInputUndated: the reading carries no instant, so its age is unknown and
	// the ceiling cannot be applied to it.
	VetoInputUndated VetoUnmeasured = "INPUT_UNDATED"
	// VetoInputAfterInstant: the reading is later than the instant being asked
	// about. Its age is negative, and a negative age passes every ceiling there
	// is — metrics.go refuses the same shape under READINGS_ALL_LATER.
	VetoInputAfterInstant VetoUnmeasured = "INPUT_AFTER_INSTANT"
	// VetoInstantUnknown: the caller supplied no instant to measure ages against.
	VetoInstantUnknown VetoUnmeasured = "INSTANT_UNKNOWN"
	// VetoThresholdAbsent: no threshold was configured for this veto. Two of the
	// three have no contract value, so this is a state an operator will meet —
	// see VetoThresholds.
	VetoThresholdAbsent VetoUnmeasured = "THRESHOLD_ABSENT"
	// VetoThresholdUnreadable: a threshold was configured and is not a plain
	// decimal. Separate from absent for parseFigure's reason: one is a contract
	// nobody wrote and the other is a contract somebody broke.
	VetoThresholdUnreadable VetoUnmeasured = "THRESHOLD_UNREADABLE"
	// VetoThresholdNotPositive: a threshold parsed and is zero or negative, which
	// is D18's second spelling of "no threshold at all".
	//
	// It is separate from absent because it says something absent does not: the
	// value reached this package as a number, so the fault is in what rendered it
	// rather than in a caller who passed nothing. §5 renders the knob with
	// strconv.FormatFloat, where an absent YAML key produces "0" — a screen full of
	// THRESHOLD_ABSENT and a screen full of this one send an operator to two
	// different files.
	VetoThresholdNotPositive VetoUnmeasured = "THRESHOLD_NOT_POSITIVE"
)

// DefaultNearHighThresholdPct is the contract's `near_high_threshold_pct: 2.0`.
//
// It is the distance to the day's high in percent and the comparison is `<`, so
// smaller is more dangerous: 1.99% raises the veto and 2.00% does not.
const DefaultNearHighThresholdPct = "2.0"

// MaxVetoInputAge is how old a reading may be before a veto computed from it is a
// memory rather than a measurement (tasks.md 4.7).
//
// # Why this bound is not derived the way the other two are
//
// DefaultStalenessTTL and MaxRankPriorAge both sit at twice the longest planned
// backoff (300s), because there *rejecting* is the expensive error: a bound that
// fired mid-ladder would expire first_seen_at, or blank rank velocity, during
// exactly the retreats the design guarantees will happen. Here the asymmetry runs
// the other way. Rejecting costs coverage, and missing coverage is already the
// ordinary state under the candle budget (D13 decision 3) — an unmeasured
// near_high is never read as a pass (D10). Accepting costs the veto itself.
// Ten minutes would therefore be the wrong shape, and tasks.md 4.7 says so
// itself: it names a three-minute-old candle as already a memory.
//
// # Two minutes, and the coverage arithmetic is not the argument
//
// A candle is one call per symbol at five a second, which is 75 symbols per
// 15-second scan tick (`scheduler_tick_seconds: 15`), so a candidate in a
// round-robin of N gets one every ⌈N/75⌉ ticks. The draft of D19 read that as "two
// minutes is eight ticks, so it covers about six hundred candidates *and* absorbs
// the first two steps of the backoff ladder", and the §4 review corrected it:
// those two do not hold at once. If 90 of the 120 seconds go to the 30s and 60s
// backoff steps, what is left is two ticks and roughly 150 symbols — four times
// smaller. And the six-hundredth candidate's freshest candle would be exactly the
// ceiling old, which the closed near end refuses anyway.
//
// The bound stands on the asymmetry above rather than on that arithmetic. Here
// accepting is the expensive error, tasks.md 4.7 already calls a three-minute-old
// candle a memory, and 600s would accept it. Two minutes is inside that and is the
// same span this package already uses as its example of a `since` window narrow
// enough to still be about now — "the last two minutes of one source rather than
// two trading days of rows to use three of them". The coverage sentence was a side
// observation, and an optimistic one, so it is not offered as a reason.
//
// It is a literal rather than a multiple of DefaultAccelerationWindow, for
// MaxRankPriorAge's reason: what makes a reading stale is how often one arrives,
// not the length of a window somebody tuned, and a multiple moves the bound for
// reasons unconnected to why the bound exists.
//
// The comparison is closed at the near end — a reading exactly this old is too
// old — matching stateAt's staleness boundary and MaxRankPriorAge's.
const MaxVetoInputAge = 2 * time.Minute

// VetoState is one veto's three-state answer.
//
// The fields are unexported and the zero value is unmeasured, so no map miss, no
// unfilled slice slot and no struct literal outside this package can produce a
// state that reads as "measured and safe". It is a struct rather than a defined
// integer or string so that it cannot be converted into, added to, or
// concatenated with anything — task 4.5's requirement stated at the level of the
// type rather than of the call site.
type VetoState struct {
	measured  bool
	dangerous bool
	why       VetoUnmeasured
}

// RaisedVeto is a veto that was measured and is dangerous.
func RaisedVeto() VetoState { return VetoState{measured: true, dangerous: true} }

// ClearedVeto is a veto that was measured and is not dangerous.
func ClearedVeto() VetoState { return VetoState{measured: true} }

// UnmeasuredVeto is a veto nobody could measure, under the name of the input that
// was missing. An empty reason becomes NOT_EVALUATED rather than staying empty.
func UnmeasuredVeto(why VetoUnmeasured) VetoState {
	if why == "" {
		why = VetoNotEvaluated
	}
	return VetoState{why: why}
}

// measuredVeto is the constructor the assessments use once they have an answer.
func measuredVeto(dangerous bool) VetoState {
	if dangerous {
		return RaisedVeto()
	}
	return ClearedVeto()
}

// Measured reports that this veto was evaluated.
func (v VetoState) Measured() bool { return v.measured }

// Dangerous reports that this veto was measured and fired.
func (v VetoState) Dangerous() bool { return v.measured && v.dangerous }

// Clear reports that this veto was measured and did not fire.
//
// It is the only affirmative predicate here and it requires the measurement. An
// unmeasured veto answers false to both Dangerous and Clear, so neither of them
// alone can be read as a pass — `!Dangerous()` is the natural spelling of "this
// one is fine" and it is the spelling D10 forbids.
func (v VetoState) Clear() bool { return v.measured && !v.dangerous }

// Reason names the missing input, and never comes back empty for an unmeasured
// veto. Empty exactly when Measured.
func (v VetoState) Reason() VetoUnmeasured {
	switch {
	case v.measured:
		return ""
	case v.why == "":
		return VetoNotEvaluated
	default:
		return v.why
	}
}

// String is what a screen prints. An unmeasured state always names its reason.
func (v VetoState) String() string {
	switch {
	case v.Dangerous():
		return "raised"
	case v.Clear():
		return "clear"
	default:
		return "unmeasured (" + string(v.Reason()) + ")"
	}
}

// Chase is the whole chase-risk verdict for one candidate.
type Chase struct {
	Key Key
	// At is the instant this was assessed as of, and what input ages were measured
	// against.
	At time.Time
	// SeenLate, Extended and NearHigh are D3's three, each three-state.
	SeenLate, Extended, NearHigh VetoState
}

// State returns one of the three by code.
//
// An unrecognised code answers unmeasured rather than falling through to the zero
// value silently. A default branch that returns something safe-looking is one of
// the ways an unmeasured becomes a false, and it is the one that survives review
// because it reads as tidiness.
func (c Chase) State(code VetoCode) VetoState {
	switch code {
	case VetoSeenLate:
		return c.SeenLate
	case VetoExtended:
		return c.Extended
	case VetoNearHigh:
		return c.NearHigh
	default:
		return UnmeasuredVeto(VetoNotEvaluated)
	}
}

// Passed reports that this candidate cleared the chase veto.
//
// All three measured and all three clear, which is the spec's 세 사유가 모두
// 측정되었고 모두 위험하지 않을 때만. Spelled as "nothing fired" it is true of a
// verdict nobody computed, and under the candle budget that is the majority of
// the list.
func (c Chase) Passed() bool {
	for _, code := range VetoCodes {
		if !c.State(code).Clear() {
			return false
		}
	}
	return true
}

// Vetoed reports that at least one veto fired.
func (c Chase) Vetoed() bool {
	for _, code := range VetoCodes {
		if c.State(code).Dangerous() {
			return true
		}
	}
	return false
}

// HasUnmeasured reports that at least one veto could not be measured. It is the
// number spec Requirement 8 wants on the scan output: the screen's default
// reading has to be "mostly unchecked", not "mostly safe".
func (c Chase) HasUnmeasured() bool {
	for _, code := range VetoCodes {
		if !c.State(code).Measured() {
			return true
		}
	}
	return false
}

// Raised is the codes that fired, in D3's order.
func (c Chase) Raised() []VetoCode {
	var out []VetoCode
	for _, code := range VetoCodes {
		if c.State(code).Dangerous() {
			out = append(out, code)
		}
	}
	return out
}

// NotMeasured is the codes nobody could measure, in D3's order.
func (c Chase) NotMeasured() []VetoCode {
	var out []VetoCode
	for _, code := range VetoCodes {
		if !c.State(code).Measured() {
			out = append(out, code)
		}
	}
	return out
}

// --- the first sighting (task 4.1) -----------------------------------------------

// Sighting is where a candidate stood in its list the first time we ever saw it.
//
// This is seen_late's input, and D8 says why it is the position rather than the
// price: 12위로 진입한 종목과 148위로 진입한 종목은 다른 사건이고, 전자가 많다면 우리
// 스캔이 늦다는 뜻이다. A symbol that was already twelfth the first time it appeared
// had been climbing for a while before our first scan caught it, and a run of
// those is how we learn the scan interval or the source panel is late. That is
// the remedy D3 attaches to seen_late and it is not the remedy it attaches to
// extended, which is why the two are separate codes.
//
// # Why it can be lost, and why it is a column
//
// The position at first sighting used to live only in the observations table,
// which is pruned after two trading days (D11) and which callers routinely window.
// A rank has no safe direction to be wrong in — the symbol may have climbed or
// slid since — so a surviving row's number is not an approximation of the first
// sighting, it is a different fact.
//
// Task 4.9 makes it a stored column (FirstRank) on first_price's lifecycle: set
// once, kept through cooling and re-entry, reset on expiry. A candidate with no
// stored rank is unmeasured, named, and never a pass (D10) — the answer is never
// taken from whatever row happens to be in the caller's slice. See
// MeasureFirstSighting for the two ways that substitution went wrong.
type Sighting struct {
	// Measured and Why carry the same three-state contract as Expansion's, and
	// Reason() is the same guarantee.
	Measured bool
	Why      VetoUnmeasured
	// PercentilePct is the position as a share of that list's own length: 0 at the
	// bottom, just under 100 at the top. Normalised for D8's reason — the KR panel
	// returns 150 rows and the US panel 100, so "twelfth" is two different facts.
	// Empty means unmeasured, never "0".
	PercentilePct string
	// Rank and RankTotal are the reading's own position and list length, kept
	// beside the derived percentile (D6). They are populated as far as the
	// measurement got, so an unmeasured answer still shows what was read.
	Rank, RankTotal int
	// At and Source date the reading, and FirstSeenAt is what it was checked
	// against. The two instants together are how a reader sees whether this really
	// was the first sighting — the same reason D17 stores first_price_at beside
	// first_price.
	At          time.Time
	Source      SourceID
	FirstSeenAt time.Time
	// NewlyListed is the source's own report that the symbol was absent from its
	// previous reading. Recorded beside the percentile and never folded into it
	// (D8 correction 1).
	//
	// It comes from the raw row the stored rank was taken from, when that row is
	// still held, so it is false after a prune. Nothing is decided by it — see
	// newlyListedAt.
	NewlyListed bool
}

// Reason names the missing input of an unmeasured sighting, and never comes back
// empty for one. See VetoState.Reason.
func (s Sighting) Reason() VetoUnmeasured {
	switch {
	case s.Measured:
		return ""
	case s.Why == "":
		return VetoNotEvaluated
	default:
		return s.Why
	}
}

// PercentileExceeds reports that the first sighting was above `thresholdPct`
// percentile points — the comparison seen_late is written against.
//
// The second return is the D10 three-state: false means this candidate was not
// measured or the threshold is not a plain decimal, and neither of those is "we
// saw it early". A caller that drops it has turned seen_late off for every
// candidate whose first sighting we no longer hold.
//
// The comparison is on the exact rational recomputed from Rank and RankTotal
// rather than on PercentilePct, because a rendering must not be what decides a
// threshold — the same rule GainExceeds and NearHigh are written under.
func (s Sighting) PercentileExceeds(thresholdPct string) (exceeds, measured bool) {
	if !s.Measured {
		return false, false
	}
	threshold, why := parseFigure(thresholdPct)
	// Three impossible readings of the same pair, and all three answer unmeasured.
	// A rank past the end of its own list is the third: the percentile is
	// (total − rank) ÷ total, so 200 of 150 is −33%, which is below every threshold
	// there is and clears seen_late for a symbol that cannot be where it says it
	// is. Observation.validate refuses it on the way into the store — for the same
	// reason, one field over — so what is left is a Sighting built outside a
	// measurement: a fixture, a §5 rendering, a future lane.
	if why != "" || s.RankTotal <= 0 || s.Rank <= 0 || s.Rank > s.RankTotal {
		return false, false
	}
	return percentileOf(s.Rank, s.RankTotal).Cmp(threshold) > 0, true
}

// MeasureFirstSighting reports where a candidate stood the first time we saw it.
//
// # The position is the stored column, never a row out of the slice
//
// Task 4.9. This used to pick the earliest ranked reading inside a ±TTL window of
// first_seen_at, and the window closed less than it looked like it did. D20's
// draft argued the two lives of a symbol cannot be closer together than the cooling
// TTL, so a ten-minute window either side could not reach across — true of
// *promotions*, and false of *observations*. D8 keeps the two independent on
// purpose, so ranked rows keep arriving for a symbol that is cooling, expired, or
// has never been promoted at all, and any of them can land inside the window:
//
//	a row from the gap between two lives   a symbol that drifted to 148th before
//	                                       it crossed again at 4th is measured
//	                                       from 148th, and seen_late clears
//	a row from before the promotion        a symbol at 5th all morning that slid
//	                                       to 100th ten minutes before it was
//	                                       promoted at 4th reads as 100/150
//
// Neither needs an expiry, and the first does not even need the two lives to be
// close: nine minutes is enough. Candidate carries nothing that separates the two
// cases, so this could not be fixed from the arguments — which is why 4.9 is the
// repair rather than an improvement on it.
//
// So the answer comes from the FirstRank column, the way MeasureExpansion's
// baseline comes from first_price, and a reading out of the slice is never
// substituted for it — that substitution is precisely the defect D17 closed for
// the price, where the slice's answer came back Measured with a re-based number
// and nothing failed. A candidate with no stored first rank is unmeasured, named,
// and never a pass (D10).
//
// # What the slice is still for
//
// Three things, none of them the measurement: the reasons that describe *why*
// there is no stored rank (nothing recorded, nothing ranked, nothing that could
// have been the first), the reading those reasons report so a screen can show how
// far it is from first_seen_at, and NewlyListed — the source's own report that the
// symbol was absent from its previous reading (D8 correction 1), which is a fact
// beside the percentile rather than an input to it. After a prune the matching row
// is gone and NewlyListed is false; it decides nothing, and PercentileExceeds does
// not read it.
func MeasureFirstSighting(c Candidate, first FirstRank, observations []Observation) Sighting {
	out := Sighting{FirstSeenAt: c.FirstSeenAt}
	// A slice carrying two symbols means the caller has lost track of which
	// candidate it is assessing, so the Candidate and the FirstRank beside it are
	// suspect too. Refused rather than answered, as before.
	if len(observations) > 0 && !oneCandidate(observations) {
		out.Why = VetoMixedCandidates
		return out
	}
	if c.FirstSeenAt.IsZero() {
		// No summary, so nothing to check the stored instant against. A first rank
		// might be this life's and might be a fortnight old, and the two are
		// indistinguishable from here.
		out.Why = VetoNoCandidate
		return out
	}
	if !first.Recorded() {
		return unstoredFirstSighting(out, c, observations)
	}

	out.Rank, out.RankTotal = first.Rank, first.Total
	out.At, out.Source = first.At, first.Source
	switch {
	case first.At.IsZero():
		// A position with no instant cannot be checked against first_seen_at at all,
		// which is the same refusal AssessExtended makes under BASELINE_UNDATED.
		out.Why = VetoFirstRankUndated
		return out
	case !nearFirstSighting(first.At, c.FirstSeenAt):
		// NoteFirstRank applies this window on the way in, so reaching it means the
		// row was written by something else — a migration, a fixture, a hand-edited
		// file — and the stored position belongs to another moment.
		out.Why = VetoFirstRankNotFirst
		return out
	}
	out.NewlyListed = newlyListedAt(observations, first)
	out.Measured = true
	out.PercentilePct = formatDecimal(percentileOf(out.Rank, out.RankTotal))
	return out
}

// unstoredFirstSighting names which of the four nothings a missing first rank is,
// and reports the reading that says so.
//
// They are four rather than one because the remedies differ: nothing recorded is a
// read that failed or a candidate nobody has observed, nothing ranked is the panel
// composition (prices and candles carry no position), nothing inside the window is
// retention or a windowed read, and a row that *is* inside the window with no
// column beside it is a store from before schema v3 or a caller that never wrote
// one.
func unstoredFirstSighting(out Sighting, c Candidate, observations []Observation) Sighting {
	// Two passes: the earliest ranked reading that could have been this life's
	// first sighting, and — for the refusal — the earliest ranked reading at all.
	//
	// Ties resolve to the first in the caller's order, which for the store's
	// (observed_at, id) is the oldest inserted row of the instant. Same rule
	// MeasureExpansion's baseline uses.
	candidateRow, earliest := -1, -1
	for i, o := range observations {
		if o.Reported.Rank <= 0 || o.Reported.RankTotal <= 0 {
			continue
		}
		if earliest < 0 || o.ObservedAt.Before(observations[earliest].ObservedAt) {
			earliest = i
		}
		if !nearFirstSighting(o.ObservedAt, c.FirstSeenAt) {
			continue
		}
		if candidateRow < 0 || o.ObservedAt.Before(observations[candidateRow].ObservedAt) {
			candidateRow = i
		}
	}
	switch {
	case len(observations) == 0:
		out.Why = VetoNoObservations
		return out
	case earliest < 0:
		// Not an error: prices and candles carry no position, so a candidate raised
		// only by them has no answer here. Having none is a different fact from
		// having entered at the bottom, and D8 refuses to convert one into the other.
		out.Why = VetoNotRanked
		return out
	}

	// Either way it reports what it read, because how far that reading is from
	// first_seen_at is the number that says which of the two it is.
	row := earliest
	out.Why = VetoNoFirstSighting
	if candidateRow >= 0 {
		row = candidateRow
		out.Why = VetoNoFirstRank
	}
	o := observations[row]
	out.Rank, out.RankTotal = o.Reported.Rank, o.Reported.RankTotal
	out.At, out.Source, out.NewlyListed = o.ObservedAt, o.Source, o.Reported.NewlyListed
	return out
}

// newlyListedAt reports the source's own "absent from the previous reading" flag
// for the stored first sighting, when the row it was taken from is still held.
//
// Matched on instant, source and position together rather than on the instant
// alone: a single scan stamps one instant on every source's row for a symbol, so
// the instant by itself selects several. A miss leaves it false, which is the
// ordinary state after a prune — the flag is recorded beside the percentile and is
// never folded into it (D8 correction 1), so nothing is decided by it.
func newlyListedAt(observations []Observation, first FirstRank) bool {
	for _, o := range observations {
		if o.Source == first.Source && o.ObservedAt.Equal(first.At) &&
			o.Reported.Rank == first.Rank && o.Reported.RankTotal == first.Total {
			return o.Reported.NewlyListed
		}
	}
	return false
}

// nearFirstSighting reports that a reading could be the one a candidate was first
// seen in — within DefaultStalenessTTL of first_seen_at, on either side.
//
// # It is now a backstop rather than the mechanism
//
// Task 4.9 made the first sighting a stored column, so this window no longer picks
// which reading is the first sighting. It bounds two things instead: NoteFirstRank
// refuses to store a reading outside it, and MeasureFirstSighting refuses to
// measure from a stored rank whose instant falls outside it. That demotion is the
// point — as the mechanism this window was load-bearing and did not bear the load,
// because D20's argument for it was about promotions and the rows that reach it are
// observations. See MeasureFirstSighting for the two shapes that walk through it.
//
// # Later, up to the TTL
//
// A candidate can be promoted a little before or after the reading that shows
// where it entered: the scan stamps one instant, the promotion is written from the
// same one, and a backoff can separate them. Ten minutes of slack because that
// window is explained by the scan interval and the planned backoff ladder — the
// same derivation and the same number as the TTL itself, closed at the near end as
// stateAt's is. Past it, the position the reading carries is a later moment's, and
// unlike a price a rank has no safe direction to be wrong in: the symbol may have
// climbed or slid since.
//
// # Earlier, and why that bound exists at all
//
// A candidate that expires and crosses again is a new candidate with a new
// first_seen_at (D1), and the dead life's rows outlive it by two trading days
// (D11). Without the earlier bound a new life would take its entry from the old
// life's rows — and the failure has a direction: a symbol that left the list at the
// bottom and came back near the top reads as "we caught it early" about the entry
// that is actually late.
func nearFirstSighting(readingAt, firstSeenAt time.Time) bool {
	gap := readingAt.Sub(firstSeenAt)
	if gap < 0 {
		gap = -gap
	}
	return gap < DefaultStalenessTTL
}

// percentileOf is a position as a share of its own list, 0 at the bottom and just
// under 100 at the top.
//
// It is the same arithmetic metrics.go's percentile applies to an Observation, and
// TestTheSightingPercentileIsTheSameOneTheRankMoveUses pins the two together
// rather than trusting them to stay that way — a screen showing a rank move of 20
// points beside a first sighting on a different scale is worse than showing
// neither.
func percentileOf(rank, total int) *big.Rat {
	return new(big.Rat).Mul(
		new(big.Rat).Quo(big.NewRat(int64(total-rank), 1), big.NewRat(int64(total), 1)),
		big.NewRat(100, 1))
}

// --- thresholds ---------------------------------------------------------------------

// VetoThresholds are the numbers the three vetoes are compared against.
//
// Only one of them is fixed by the contract. design.md's 결정된 계약값 gives
// `near_high_threshold_pct: 2.0` and gives no value for the other two, so those
// arrive from the caller and an absent one makes its veto unmeasured rather than
// clear. That is the safe direction and it is the only one available: a number
// this package invented would be a policy value with no source, applied market,
// or verification state, which is exactly what D6 says must not exist. §5 records
// them the way the shadow acceleration thresholds are recorded, and T3.2 decides.
type VetoThresholds struct {
	// SeenLatePercentilePct is how far up its list a candidate may already have
	// been the first time we saw it, in percentile points, compared with `>`.
	SeenLatePercentilePct string
	// ExtendedGainPct is how far a candidate may have run since the first price we
	// recorded for it, in percent, compared with `>`.
	ExtendedGainPct string
	// NearHighDistancePct is how close to the day's high is too close, as a
	// distance in percent, compared with `<`. DefaultNearHighThresholdPct is the
	// contract's value; the comparison's boundary is pinned to the digit.
	NearHighDistancePct string
	// MaxInputAge is how old a reading may be before a veto computed from it is
	// unmeasured. Anything non-positive means MaxVetoInputAge: a bound nobody set
	// is the default, never no bound at all. A zero here meaning "unlimited" is
	// the same defect as a zero DayHigh meaning "at the high" — the field nobody
	// filled in switching the guard off.
	MaxInputAge time.Duration
}

func (t VetoThresholds) inputAge() time.Duration {
	if t.MaxInputAge <= 0 {
		return MaxVetoInputAge
	}
	return t.MaxInputAge
}

// thresholdReason reports why a configured threshold cannot be used, empty when
// it can.
//
// # A threshold of zero is the second spelling of no threshold (D18)
//
// Refusing only the empty and the unparseable leaves "0", "-0", "0.0" and "-1",
// and every one of them parses. The comparison then proceeds: near_high is
// `distance < threshold`, which is false for every candidate below its high, so
// the veto reads measured and clear for the whole list — and near_high is the only
// veto D18 approved a threshold for, so that is the entire live veto surface
// switching itself off.
//
// The input is not hypothetical. §5 renders the knob with
// strconv.FormatFloat(cfg.NearHighThresholdPct, 'f', -1, 64), and an absent YAML
// key arrives here as "0" rather than "". It is task 1.1's zero DayHigh — "a zero
// day high would make every high-proximity comparison pass" — reappearing on the
// other operand of the same comparison, which is also the defect VetoThresholds'
// MaxInputAge already guards against on the other field.
//
// Refusing is safe in all three codes and in both directions of every comparison.
// seen_late and extended are `>`, where a non-positive threshold fires *more*, so
// the refusal only turns a raised into an unmeasured there; near_high is `<`,
// where it turns a clear into an unmeasured. Neither direction manufactures a
// pass, which is the only property D10 asks of a change here.
func thresholdReason(s string) VetoUnmeasured {
	value, why := parseFigure(s)
	switch why {
	case "":
	case NotComputedFigureAbsent:
		return VetoThresholdAbsent
	default:
		return VetoThresholdUnreadable
	}
	if value.Sign() <= 0 {
		return VetoThresholdNotPositive
	}
	return ""
}

// inputAgeReason reports why a reading taken at `readingAt` cannot answer a
// question asked as of `at`, empty when it can.
//
// Three refusals rather than one, because they are three different problems: a
// caller with no instant, a reading with no instant, and a reading from after the
// question. The third matters most — its age is negative and a negative age
// passes every ceiling, so folding it into "fresh" is how a clock disagreement
// turns into a measurement nobody made.
func inputAgeReason(at, readingAt time.Time, max time.Duration) VetoUnmeasured {
	switch {
	case at.IsZero():
		return VetoInstantUnknown
	case readingAt.IsZero():
		return VetoInputUndated
	case readingAt.After(at):
		return VetoInputAfterInstant
	case at.Sub(readingAt) >= max:
		return VetoInputTooOld
	default:
		return ""
	}
}

// --- the three vetoes ------------------------------------------------------------

// vetoFrom is the only place a (value, measured) pair from a measurement is read.
//
// Having exactly one reader is the point. `near, _ := r.NearHigh(th)` compiles,
// and it turns every unmeasured candidate into a clear one, and it is the natural
// spelling — the §3 review found the same shape three times in metrics.go. There
// is no second call site at which to write it.
//
// A pair that comes back unmeasured from an input that said it was measured means
// the figures behind it are not numbers, which is what level.go calls
// UNREADABLE_DECIMAL. The caller passes that as the fallback rather than letting
// the reason fall to NOT_EVALUATED, which would say nobody looked when somebody
// did.
func vetoFrom(exceeds, measured bool, why VetoUnmeasured) VetoState {
	if !measured {
		return UnmeasuredVeto(why)
	}
	return measuredVeto(exceeds)
}

// AssessSeenLate is task 4.1: was this symbol already well up its list the first
// time we saw it.
//
// No input-age ceiling here, and that is not an omission. The first sighting is
// old by construction — being old is the fact it reports — so what has to be
// checked is not its age but its identity: whose first sighting is this, and is it
// this life's. MeasureFirstSighting is where both are decided, from the stored
// column rather than from a surviving row (task 4.9).
func AssessSeenLate(s Sighting, th VetoThresholds) VetoState {
	if why := thresholdReason(th.SeenLatePercentilePct); why != "" {
		return UnmeasuredVeto(why)
	}
	if !s.Measured {
		return UnmeasuredVeto(s.Reason())
	}
	exceeds, measured := s.PercentileExceeds(th.SeenLatePercentilePct)
	// The threshold is already known to parse, so the only way the pair comes back
	// unmeasured from a measured sighting is a position that is not one.
	return vetoFrom(exceeds, measured, VetoNotRanked)
}

// AssessExtended is task 4.2: has it run too far since the first price we
// recorded for it.
//
// # The baseline has to be this life's first price, and the window is symmetric
//
// Task 4.2b. D17's migration backfill fills first_price from the earliest
// surviving raw row, and if the rows before it were already pruned the stored
// value is later than the price the candidate was actually first seen at. A late
// baseline understates the expansion — 10,000 to 30,000 reported as 20,000 to
// 30,000 — and understating pushes `extended` towards `false`. That is the
// direction D10 forbids, arriving through a migration instead of through a
// missing input, so a baseline more than DefaultStalenessTTL later than
// first_seen_at is unmeasured rather than clear.
//
// The earlier end is the §4 review's P1-1 and it is the same failure reflected.
// MeasureExpansion moves the baseline *backwards* onto any older priced row in
// the slice — deliberately, since a baseline may move backwards on new evidence
// and never forwards — while raw rows outlive a candidate life by two trading days
// (D11) and Promote clears first_price on expiry. So a dead life's price becomes
// the live life's denominator, and it overstates the baseline: a candidate
// promoted at 10,000 that has since doubled to 20,000 is measured from a previous
// life's 30,000 and reports −33%, which reads as "has not run".
//
// A legitimate baseline is written at promotion, within seconds of first_seen_at.
// The only thing the earlier end refuses is a baseline the series moved backwards
// onto — which is why the comparison is on |FirstAt − FirstSeenAt| under two
// reasons rather than one, since a late baseline and a dead life's are fixed by
// different things.
//
// Ten minutes of slack because that window is explained by the scan interval and
// the planned backoff ladder; past it the gap is something else. The two instants
// are both stored precisely so this can be one comparison (D17).
//
// # And the latest price has an age ceiling for task 4.7's reason
//
// The gain is last ÷ first, so a stale `last` understates it exactly as a late
// baseline does, and in the same direction. tasks.md 4.7 names the candle because
// the candle is the scarce input, but the argument is about the arithmetic rather
// than about which endpoint supplied it: a "has not run" computed from a price
// somebody fetched four minutes ago is a memory too.
func AssessExtended(e Expansion, c Candidate, at time.Time, th VetoThresholds) VetoState {
	if why := thresholdReason(th.ExtendedGainPct); why != "" {
		return UnmeasuredVeto(why)
	}
	if !e.Measured {
		return UnmeasuredVeto(VetoUnmeasured(e.Reason()))
	}
	if c.FirstSeenAt.IsZero() {
		return UnmeasuredVeto(VetoNoCandidate)
	}
	if e.FirstAt.IsZero() {
		return UnmeasuredVeto(VetoBaselineUndated)
	}
	if gap := e.FirstAt.Sub(c.FirstSeenAt); gap >= DefaultStalenessTTL {
		return UnmeasuredVeto(VetoBaselineTooLate)
	} else if -gap >= DefaultStalenessTTL {
		return UnmeasuredVeto(VetoBaselineTooEarly)
	}
	if why := inputAgeReason(at, e.LastAt, th.inputAge()); why != "" {
		return UnmeasuredVeto(why)
	}
	exceeds, measured := e.GainExceeds(th.ExtendedGainPct)
	return vetoFrom(exceeds, measured, VetoUnmeasured(LevelUnreadableDecimal))
}

// AssessNearHigh is task 4.3: is there room left below the day's high.
//
// # The comparison is `<`
//
// D3's table said 상한 초과 for several drafts, which reads as `>` and inverts the
// veto — it would fire on candidates far from the high and pass the ones sitting
// on it, so the judgement D3 calls a 거부권 turns exactly the wrong way round. The
// stored quantity is the *distance* to the day's high and smaller is more
// dangerous. RangePosition.NearHigh is where the comparison lives, on exact
// rationals recomputed from the reported high and price, and this function reaches
// for it rather than deriving its own: a veto that disagreed with the number
// printed beside it would be worse than either of them being wrong alone.
//
// # And the reading has an age ceiling
//
// Task 4.7. A candle costs one call per symbol at five a second, so a high that
// was fetched once gets reused for as long as the row is retained — two trading
// days. The price in that same reading is what the distance is measured from
// (RangePosition takes both from one observation on purpose), so a reading that
// has aged says nothing about where the price is now, and the error points at "not
// dangerous": the price has risen since and the recorded distance has not.
// Without the ceiling the failure task 4.4 blocks reappears one level down and
// looks identical on a screen.
func AssessNearHigh(r RangePosition, at time.Time, th VetoThresholds) VetoState {
	if why := thresholdReason(th.NearHighDistancePct); why != "" {
		return UnmeasuredVeto(why)
	}
	if !r.Measured {
		// The ordinary case, all day, for most of the list: NO_DAY_HIGH. It travels
		// under its own name so a screen full of them reads as the candle budget
		// rather than as an outage.
		return UnmeasuredVeto(VetoUnmeasured(r.Reason()))
	}
	if why := inputAgeReason(at, r.At, th.inputAge()); why != "" {
		return UnmeasuredVeto(why)
	}
	near, measured := r.NearHigh(th.NearHighDistancePct)
	return vetoFrom(near, measured, VetoUnmeasured(LevelUnreadableDecimal))
}

// --- assembly ---------------------------------------------------------------------

// VetoInputs is everything the three vetoes are computed from.
//
// There is deliberately no acceleration, no rank move and no tally in it. D3
// keeps chase risk out of the score, and the way to keep it out is for the veto
// to be unable to see one — "the acceleration is top of the list, so being on the
// high is acceptable" cannot be written here because there is nothing to write it
// against. TestTheVetoCannotSeeAScoreToBeOffsetBy asserts that field by field, so
// adding one later fails rather than reads as a reasonable diff.
type VetoInputs struct {
	// Candidate is the summary, and it is required: both seen_late and extended
	// check their input against first_seen_at.
	Candidate Candidate
	// Sighting, Expansion and Range are the three measurements, each already
	// three-state.
	Sighting  Sighting
	Expansion Expansion
	Range     RangePosition
	// At is the instant being assessed as of, supplied by the caller. Input ages
	// are measured against it and nothing here reads a clock.
	At time.Time
	// Thresholds are the numbers to compare against. An absent one leaves its veto
	// unmeasured.
	Thresholds VetoThresholds
}

// AssessChase is the whole chase verdict for one candidate.
func AssessChase(in VetoInputs) Chase {
	return Chase{
		Key:      in.Candidate.Key,
		At:       in.At,
		SeenLate: AssessSeenLate(in.Sighting, in.Thresholds),
		Extended: AssessExtended(in.Expansion, in.Candidate, in.At, in.Thresholds),
		NearHigh: AssessNearHigh(in.Range, in.At, in.Thresholds),
	}
}

// --- the tally (task 4.6) ------------------------------------------------------

// VetoTally is the veto record over a set of candidates.
//
// Vetoed candidates are counted rather than dropped, because D3 says so and says
// why: 지우면 "우리가 늦게 본 종목이 얼마나 되는가"를 나중에 셀 수 없고, 그 수가 이
// 시스템이 실제로 조기인지 아닌지의 유일한 증거다. The share of candidates we saw
// late is the only evidence this system is early, and it can only be counted if
// the late ones survive.
//
// # Three buckets, disjoint by construction
//
// Total == Passed + Vetoed + Unmeasured holds for every tally, so a candidate
// that fell out of all of them shows up as an arithmetic failure rather than as a
// number that is quietly short. That is the guard CrossingTally grew after the §3
// review found TallyCrossings dropping the unmeasured candidates out of both of
// its halves — which flipped the screen's default reading from "mostly unchecked"
// to "all measured and quiet", in exactly the direction D10 is about.
//
// Precedence is raised, then unmeasured, then clear. A candidate that is both
// vetoed and partly unmeasured is vetoed: the strongest thing known about it is
// the thing that fired.
type VetoTally struct {
	// Total is how many verdicts were tallied.
	Total int
	// Passed is how many had all three measured and all three clear.
	Passed int
	// Vetoed is how many had at least one veto fire.
	Vetoed int
	// Unmeasured is how many fired nothing and were not fully measured either.
	// Under the candle budget this is the majority (D13 decision 3), which is why
	// spec Requirement 8 puts it on the scan output.
	Unmeasured int

	// Raised counts, per code, the candidates whose veto fired. Its keys are
	// exactly VetoCodes even at zero, so a column that is missing cannot read as a
	// column nobody needed.
	Raised map[VetoCode]int
	// NotMeasured counts, per code, the candidates whose veto could not be
	// measured. Same keys, same reason.
	NotMeasured map[VetoCode]int
	// Reasons counts the missing inputs by name across every code, which is what
	// tells an operator whether they are looking at the candle budget, a stale hot
	// queue or an unconfigured threshold.
	Reasons map[VetoUnmeasured]int
}

// TallyVetoes counts a set of verdicts.
func TallyVetoes(in []Chase) VetoTally {
	out := VetoTally{
		Total:       len(in),
		Raised:      make(map[VetoCode]int, len(VetoCodes)),
		NotMeasured: make(map[VetoCode]int, len(VetoCodes)),
		Reasons:     map[VetoUnmeasured]int{},
	}
	for _, code := range VetoCodes {
		out.Raised[code] = 0
		out.NotMeasured[code] = 0
	}
	for _, c := range in {
		for _, code := range VetoCodes {
			state := c.State(code)
			switch {
			case state.Dangerous():
				out.Raised[code]++
			case !state.Measured():
				out.NotMeasured[code]++
				out.Reasons[state.Reason()]++
			}
		}
		switch {
		case c.Vetoed():
			out.Vetoed++
		case c.HasUnmeasured():
			out.Unmeasured++
		default:
			out.Passed++
		}
	}
	return out
}
