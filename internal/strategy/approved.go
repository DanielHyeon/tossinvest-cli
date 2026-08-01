// Package strategy is the value-only handoff from candidate approval into a
// strategy lane. It deliberately owns no clock, broker, journal, callback, or
// mutable collection. The candidate package's static boundary audit enforces
// that property for every production file in this package.
package strategy

import "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"

// ApprovedCandidate is opaque outside this package. It can only wrap the exact
// immutable candidate.ApprovedCandidate value; a zero or refused candidate
// remains refused when a lane reads it.
type ApprovedSnapshot struct {
	valid                                                                               bool
	market, symbol, state, candidateLifeID, thresholdVersion, setDigest, evidenceDigest string
	firstSeenUnixNano, lastSeenUnixNano, validUntilUnixNano, approvedAtUnixNano         int64
}

// FromApproved is the only candidate-reading handoff. It performs no work and
// grants no execution authority.
func SealApproved(value candidate.ApprovedCandidate) ApprovedSnapshot {
	return ApprovedSnapshot{valid: value.Valid(), market: value.MarketString(), symbol: value.SymbolString(), state: value.StateString(), candidateLifeID: value.CandidateLifeIDString(), thresholdVersion: value.ThresholdVersion(), setDigest: value.SetDigest(), evidenceDigest: value.EvidenceDigest(), firstSeenUnixNano: value.FirstSeenUnixNano(), lastSeenUnixNano: value.LastSeenUnixNano(), validUntilUnixNano: value.ValidUntilUnixNano(), approvedAtUnixNano: value.ApprovedAtUnixNano()}
}

func (s ApprovedSnapshot) Valid() bool               { return s.valid }
func (s ApprovedSnapshot) Market() string            { return s.market }
func (s ApprovedSnapshot) Symbol() string            { return s.symbol }
func (s ApprovedSnapshot) State() string             { return s.state }
func (s ApprovedSnapshot) CandidateLifeID() string   { return s.candidateLifeID }
func (s ApprovedSnapshot) ThresholdVersion() string  { return s.thresholdVersion }
func (s ApprovedSnapshot) SetDigest() string         { return s.setDigest }
func (s ApprovedSnapshot) EvidenceDigest() string    { return s.evidenceDigest }
func (s ApprovedSnapshot) FirstSeenUnixNano() int64  { return s.firstSeenUnixNano }
func (s ApprovedSnapshot) LastSeenUnixNano() int64   { return s.lastSeenUnixNano }
func (s ApprovedSnapshot) ValidUntilUnixNano() int64 { return s.validUntilUnixNano }
func (s ApprovedSnapshot) ApprovedAtUnixNano() int64 { return s.approvedAtUnixNano }
