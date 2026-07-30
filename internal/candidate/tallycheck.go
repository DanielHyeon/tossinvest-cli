package candidate

// tallycheck.go is section 4: the arithmetic that catches an unmeasured veto being
// counted as a pass.
//
// # Why an identity and not a threshold check
//
// The alarm that suggests itself is "Passed is non-zero and seen_late has no
// approved threshold, so something is wrong". Two things are wrong with it.
//
// It needs the thresholds, and neither reporting surface has them. CycleResult
// carries no threshold field, SignalsMarket carries none, and CycleOptions.Thresholds
// is consumed inside the assessment and discarded. The implementation that compiles
// without touching either shape is a note function that reads a package constant
// directly — which makes the note a function of the constant rather than of the
// threshold that was applied, and the alarm branch then becomes unreachable while
// looking like it works.
//
// And it is narrower than the bug. THRESHOLD_ABSENT is one of many unmeasured
// reasons. The edit D10 exists to stop — counting an unmeasured veto as a pass — is
// invisible to a threshold check whenever the reason is NO_DAY_HIGH, INPUT_TOO_OLD,
// NOT_RANKED, BASELINE_TOO_LATE, NEW_ENTRANT_UNKNOWN or any of the others.
//
// What survives both objections is arithmetic on the tally itself. A candidate that
// passed contributed zero to every code's Raised and NotMeasured, so for every code
//
//	Passed + Raised[code] + NotMeasured[code] <= Total
//
// and the inequality is sharp rather than slack: under the candle budget
// NotMeasured[near_high] is most of the list (D13 decision 3), so Passed has very
// little room before the sum crosses Total. It needs no thresholds, so it works
// today, with none configured — where the threshold-shaped alarm would have guarded
// nothing until some later change gave it something to read.
//
// The second check is the direct contradiction: a candidate cannot both have been
// counted as passing and have had a veto go unmeasured for want of a threshold.
//
// # One rule, two screens
//
// The judgement is here and the wording is not. `tossctl candidate scan` and
// /signals each render their own sentence from the same anomalies, because the
// repository has already paid for the other arrangement once — a fourth shadow band
// appeared on the scan output and not on /signals, and a band missing from a page
// reads as a band nobody crossed rather than as one nobody counted.

import "fmt"

// TallyAnomalyKind names which arithmetic failed.
//
// Two kinds rather than one boolean, for the reason veto.go keeps many unmeasured
// reasons: they send a reader to different places. The first says the buckets do not
// add up for a particular code; the second says a pass was counted beside a veto
// that had no threshold to be measured against.
type TallyAnomalyKind string

const (
	// TallyOvercounted: Passed + Raised[code] + NotMeasured[code] exceeds Total.
	// A candidate has been counted in two buckets that are disjoint by
	// construction, and the only way that happens is a verdict being read as a pass
	// while one of its vetoes was not measured.
	TallyOvercounted TallyAnomalyKind = "OVERCOUNTED"
	// TallyPassedWithoutThreshold: a pass was counted while THRESHOLD_ABSENT was
	// among the unmeasured reasons. An absent threshold cannot produce a clear veto,
	// so the two counts contradict each other directly.
	TallyPassedWithoutThreshold TallyAnomalyKind = "PASSED_WITHOUT_THRESHOLD"
)

// TallyAnomaly is one contradiction found in a veto tally, with the numbers that
// produced it.
//
// The numbers travel with the finding because a screen that says only "the tally is
// inconsistent" gives nobody anything to check. It is the same rule D6 states for
// measurements: the reading and the conclusion drawn from it are different fields.
type TallyAnomaly struct {
	// Kind is which arithmetic failed.
	Kind TallyAnomalyKind
	// Code is the veto code the sum was taken over. Empty for
	// TallyPassedWithoutThreshold, which is not about one code.
	Code VetoCode
	// Total, Passed, Raised and NotMeasured are the four counts, as tallied.
	Total, Passed, Raised, NotMeasured int
	// Reason and ReasonCount carry the unmeasured reason behind
	// TallyPassedWithoutThreshold. Zero for the other kind.
	Reason      VetoUnmeasured
	ReasonCount int
}

// Arithmetic is the finding stated as the sum that failed, with no advice in it.
//
// The advice belongs to the surface, which knows who is reading. What must not
// differ between the two surfaces is this: the numbers, and which comparison they
// failed.
func (a TallyAnomaly) Arithmetic() string {
	switch a.Kind {
	case TallyPassedWithoutThreshold:
		return fmt.Sprintf("passed %d and %s %d, both non-zero",
			a.Passed, a.Reason, a.ReasonCount)
	default:
		return fmt.Sprintf("%s: passed %d + raised %d + unmeasured %d = %d > total %d",
			a.Code, a.Passed, a.Raised, a.NotMeasured,
			a.Passed+a.Raised+a.NotMeasured, a.Total)
	}
}

// Anomalies reports the tally's own arithmetic contradictions, in VetoCodes' order
// so that two screens list them identically.
//
// An empty result is the normal state and it is not a claim that the numbers are
// right — only that they are not provably wrong. That distinction is why the
// surfaces render the alarm beside the counts rather than instead of them.
func (t VetoTally) Anomalies() []TallyAnomaly {
	var out []TallyAnomaly
	for _, code := range VetoCodes {
		raised, notMeasured := t.Raised[code], t.NotMeasured[code]
		if t.Passed+raised+notMeasured > t.Total {
			out = append(out, TallyAnomaly{
				Kind: TallyOvercounted, Code: code,
				Total: t.Total, Passed: t.Passed,
				Raised: raised, NotMeasured: notMeasured,
			})
		}
	}
	if n := t.Reasons[VetoThresholdAbsent]; n > 0 && t.Passed > 0 {
		out = append(out, TallyAnomaly{
			Kind:  TallyPassedWithoutThreshold,
			Total: t.Total, Passed: t.Passed,
			Reason: VetoThresholdAbsent, ReasonCount: n,
		})
	}
	return out
}
