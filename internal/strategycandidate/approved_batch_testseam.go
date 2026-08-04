//go:build tossos_testseams

package strategycandidate

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategy"

// ApprovedBatchForTest copies explicit sealed approvals into an opaque batch.
// Invalid values remain impossible to smuggle through the seam.
func ApprovedBatchForTest(values ...strategy.ApprovedSnapshot) ApprovedBatch {
	for _, value := range values {
		if !value.Valid() {
			return ApprovedBatch{}
		}
	}
	return ApprovedBatch{values: append([]strategy.ApprovedSnapshot{}, values...)}
}
