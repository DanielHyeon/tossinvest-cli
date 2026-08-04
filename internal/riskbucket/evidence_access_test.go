package riskbucket

import (
	"testing"
	"time"
)

func TestBucketSnapshotBoundEvidenceReturnsImmutableValueCopy(t *testing.T) {
	now := time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)
	key := BucketKey{Dimension: DimensionMarket, Value: string(MarketKR), PolicyVersion: "policy-v1"}
	policy, err := NewPolicyProvenance(key, Evidence{Source: RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-digest", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	binding := BucketSnapshotBinding{Key: key, LimitMinor: "100", FilledMinor: "1", HeldMinor: "2", SnapshotVersion: "snapshot-v1"}
	snapshot, err := NewSnapshotProvenance(binding, Evidence{Source: RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion, Digest: "snapshot-digest", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	bucket := BucketSnapshot{Key: key, LimitMinor: "100", FilledMinor: "1", HeldMinor: "2", SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policy, SnapshotProvenance: snapshot}
	first := bucket.BoundEvidence()
	first.PolicyEvidence.Digest = "mutated"
	first.SnapshotEvidence.Digest = "mutated"
	first.Snapshot.LimitMinor = "999"
	second := bucket.BoundEvidence()
	if second.Key != key || second.Snapshot != binding || second.PolicyEvidence.Digest != "policy-digest" || second.SnapshotEvidence.Digest != "snapshot-digest" {
		t.Fatalf("bound evidence mutated through copy: %+v", second)
	}
}
