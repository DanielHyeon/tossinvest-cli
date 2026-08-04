//go:build tossos_testseams

package strategy

import "time"

// ApprovedSnapshotForTest creates an immutable production-shaped approval only
// in explicit repository test-seam binaries.
func ApprovedSnapshotForTest(market, symbol string, at time.Time) ApprovedSnapshot {
	return ApprovedSnapshot{
		valid:              true,
		market:             market,
		symbol:             symbol,
		state:              "APPROVED",
		candidateLifeID:    "candidate:" + market + ":" + symbol,
		thresholdVersion:   "candidate-veto-v1",
		setDigest:          "sha256:test-threshold-set-" + market,
		evidenceDigest:     "sha256:test-candidate-evidence-" + market + "-" + symbol,
		firstSeenUnixNano:  at.Add(-time.Minute).UnixNano(),
		lastSeenUnixNano:   at.UnixNano(),
		validUntilUnixNano: at.Add(time.Minute).UnixNano(),
		approvedAtUnixNano: at.UnixNano(),
	}
}
