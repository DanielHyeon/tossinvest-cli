package riskbucket

// BucketEvidenceBinding is a read-only-by-value view of the authority evidence
// already sealed by NewPolicyProvenance and NewSnapshotProvenance. Returning
// value copies lets persistence verify exact bindings without exposing a way to
// construct or mutate either sealed provenance value.
type BucketEvidenceBinding struct {
	Key              BucketKey
	Snapshot         BucketSnapshotBinding
	PolicyEvidence   Evidence
	SnapshotEvidence Evidence
}

// BoundEvidence returns immutable value copies of the evidence consumed by
// CalculateAdmission for this bucket. Mutating the returned values cannot alter
// the BucketSnapshot or its private authority seals.
func (b BucketSnapshot) BoundEvidence() BucketEvidenceBinding {
	return BucketEvidenceBinding{
		Key:              b.PolicyProvenance.key,
		Snapshot:         b.SnapshotProvenance.binding,
		PolicyEvidence:   b.PolicyProvenance.evidence,
		SnapshotEvidence: b.SnapshotProvenance.evidence,
	}
}
