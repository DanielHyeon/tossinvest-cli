package main

// vetothresholds.go is the one place this command decides what the chase veto is
// compared against.
//
// # Why one place
//
// There were two. `tossctl candidate scan` built a candidate.VetoThresholds literal
// and the /signals seam built another, and nothing held them together. Editing one
// is the ordinary way to change a threshold, and the result is two screens reporting
// different verdicts about the same store at the same instant — each internally
// consistent, each rendering its own numbers correctly, with no failure anywhere.
//
// This repository has already had that failure once and wrote the lesson down in
// internal/candidate/watch.go: a fourth shadow band appeared on the scan output and
// not on /signals, because the two surfaces assembled the same record separately.
// The repair then was one reducer plus a test that says nobody else may build it,
// and this is the same repair on the input side. It is cheap now — there is one
// approved threshold to move — and it stops being cheap the moment somebody has a
// value to introduce.
//
// # And why the values are what they are
//
// Only near_high has a contract value (design.md 결정된 계약값,
// `near_high_threshold_pct: 2.0`). seen_late and extended are left empty on purpose,
// so their vetoes report THRESHOLD_ABSENT — unmeasured, and never a pass — rather
// than a number this command invented. D18: 근거 없는 임계는 veto가 되지 못한다.
//
// The empty strings are also why nothing here renders a knob with
// strconv.FormatFloat. An absent YAML key formatted that way arrives as "0", and a
// threshold of zero parses — near_high's comparison is `distance < threshold`, which
// is false for every candidate below its high, so the only live veto in the system
// would report itself measured and clear for the whole list. internal/candidate
// refuses a non-positive threshold (THRESHOLD_NOT_POSITIVE) and this side does not
// manufacture one.

import "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"

// candidateVetoThresholds is the thresholds every surface of this command uses.
//
// A function rather than a package-level variable, so that no caller can mutate what
// the next one reads.
func candidateVetoThresholds() candidate.VetoThresholds {
	return candidate.VetoThresholds{
		NearHighDistancePct: candidate.DefaultNearHighThresholdPct,
	}
}
