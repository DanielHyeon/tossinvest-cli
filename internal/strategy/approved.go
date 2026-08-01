// Package strategy is the value-only handoff from candidate approval into a
// strategy lane. It deliberately owns no clock, broker, journal, callback, or
// mutable collection. The candidate package's static boundary audit enforces
// that property for every production file in this package.
package strategy

import "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"

// ApprovedCandidate is opaque outside this package. It can only wrap the exact
// immutable candidate.ApprovedCandidate value; a zero or refused candidate
// remains refused when a lane reads it.
type ApprovedCandidate struct {
	value candidate.ApprovedCandidate
}

// FromApproved is the only candidate-reading handoff. It performs no work and
// grants no execution authority.
func FromApproved(value candidate.ApprovedCandidate) ApprovedCandidate {
	return ApprovedCandidate{value: value}
}

// Value returns the immutable verdict so a downstream pure lane can read its
// scalar provenance. It does not turn the verdict into an order capability.
func Value(value ApprovedCandidate) candidate.ApprovedCandidate {
	return value.value
}
