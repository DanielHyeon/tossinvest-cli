//go:build tossos_testseams

package scheduler

import "time"

// ActivationForTest mints an opaque activation only in explicit repository
// test-seam binaries. Production code still obtains it exclusively from Restore.
func ActivationForTest(binding ActivationBinding) *Activation {
	generation := binding.DesiredRevision
	if generation == 0 {
		generation = 1
	}
	return &Activation{binding: binding, evidence: ActivationEvidence{Generation: generation, ExpiresAt: binding.ApprovedAt.Add(24 * time.Hour)}}
}
