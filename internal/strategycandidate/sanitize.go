// Package strategycandidate is the audited value sanitizer between candidate
// assessment and strategy evaluation. It owns no journal, Gateway, broker,
// clock, callback, writer or operating-setting capability.
package strategycandidate

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
)

// ApprovedBatch keeps sanitized values opaque to authority-bearing packages.
// Append always copies the private slice before adding a value, so callers
// cannot mutate an earlier batch through aliasing.
type ApprovedBatch struct {
	values []strategy.ApprovedSnapshot
}

// Seal repeats approval from the exact measured verdict and sealed production
// threshold authority, then immediately converts it to the strategy package's
// opaque scalar snapshot. The candidate.ApprovedCandidate never crosses this
// package boundary.
func Seal(verdict candidate.Verdict, authority candidate.ProductionThresholdAuthority, at time.Time) (strategy.ApprovedSnapshot, error) {
	approved, err := candidate.AssessApprovedCandidate(candidate.VetoInputs{
		Candidate:  verdict.Summary.Candidate,
		Sighting:   verdict.Sighting,
		Expansion:  verdict.Expansion,
		Range:      verdict.Range,
		At:         at,
		Thresholds: authority.ThresholdSet().VetoThresholds(),
	}, authority.ThresholdSet())
	if err != nil {
		return strategy.ApprovedSnapshot{}, err
	}
	return strategy.SealApproved(approved), nil
}

func Append(batch ApprovedBatch, verdict candidate.Verdict, authority candidate.ProductionThresholdAuthority, at time.Time) (ApprovedBatch, error) {
	sealed, err := Seal(verdict, authority, at)
	if err != nil {
		return ApprovedBatch{}, err
	}
	values := append([]strategy.ApprovedSnapshot{}, batch.values...)
	values = append(values, sealed)
	return ApprovedBatch{values: values}, nil
}

func (batch ApprovedBatch) Len() int { return len(batch.values) }

func (batch ApprovedBatch) At(index int) (strategy.ApprovedSnapshot, bool) {
	if index < 0 || index >= len(batch.values) {
		return strategy.ApprovedSnapshot{}, false
	}
	return batch.values[index], true
}
