package journal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestRiskBucketAdmissionPersistsDistinctPolicyAndSnapshotObservationWindows(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-distinct-windows", "acct-1")
	plan := riskBucketAdmissionFixture(t, "distinct-windows", "acct-1", "lane-short", "campaign-1", "prospective-windows", "100", "0")
	policyObserved, policyFresh, snapshotObserved, snapshotFresh := setDistinctRiskReferenceWindows(t, &plan)
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	var storedPolicyObserved, storedPolicyFresh, storedSnapshotObserved, storedSnapshotFresh string
	if err := j.db.QueryRow(`SELECT policy_observed_at,policy_fresh_until FROM risk_bucket_policies
		WHERE bucket_dimension='horizon'`).Scan(&storedPolicyObserved, &storedPolicyFresh); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT observed_at,fresh_until FROM risk_bucket_snapshots
		WHERE bucket_dimension='horizon'`).Scan(&storedSnapshotObserved, &storedSnapshotFresh); err != nil {
		t.Fatal(err)
	}
	if storedPolicyObserved != canonicalRiskTime(policyObserved) || storedPolicyFresh != canonicalRiskTime(policyFresh) ||
		storedSnapshotObserved != canonicalRiskTime(snapshotObserved) || storedSnapshotFresh != canonicalRiskTime(snapshotFresh) {
		t.Fatalf("stored windows policy=%s..%s snapshot=%s..%s", storedPolicyObserved, storedPolicyFresh, storedSnapshotObserved, storedSnapshotFresh)
	}
}

func TestRiskBucketAdmissionRejectsCrossWiredDistinctObservationWindows(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-cross-wired-windows", "acct-1")
	plan := riskBucketAdmissionFixture(t, "cross-wired-windows", "acct-1", "lane-short", "campaign-1", "prospective-windows", "100", "0")
	setDistinctRiskReferenceWindows(t, &plan)
	plan.Snapshots[2].PolicyObservedAt = plan.Snapshots[2].SnapshotObservedAt
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("cross-wired windows error=%v", err)
	}
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Fatalf("cross-wired windows wrote %s rows=%d", table, got)
		}
	}
}

func setDistinctRiskReferenceWindows(t *testing.T, plan *RiskBucketAdmissionPlan) (time.Time, time.Time, time.Time, time.Time) {
	t.Helper()
	now := plan.Admission.Policy.EvaluatedAt
	policyObserved, policyFresh := now.Add(-2*time.Minute), now.Add(2*time.Minute)
	snapshotObserved, snapshotFresh := now.Add(-30*time.Second), now.Add(30*time.Second)
	for i := range plan.Admission.Buckets {
		bucket := plan.Admission.Buckets[i]
		bound := bucket.BoundEvidence()
		policy, err := riskbucket.NewPolicyProvenance(bucket.Key, riskbucket.Evidence{
			Source: riskbucket.RiskPolicyAuthoritySource, Version: bucket.Key.PolicyVersion,
			Digest: plan.Snapshots[i].PolicyDigest, Official: true, Frozen: true,
			ObservedAt: policyObserved, FreshUntil: policyFresh,
		})
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := riskbucket.NewSnapshotProvenance(bound.Snapshot, riskbucket.Evidence{
			Source: riskbucket.RiskSnapshotAuthoritySource, Version: bucket.SnapshotVersion,
			Digest: plan.Snapshots[i].SnapshotDigest, Official: true, Frozen: true,
			ObservedAt: snapshotObserved, FreshUntil: snapshotFresh,
		})
		if err != nil {
			t.Fatal(err)
		}
		bucket.PolicyProvenance = policy
		bucket.SnapshotProvenance = snapshot
		plan.Admission.Buckets[i] = bucket
		plan.Snapshots[i].PolicyObservedAt = policyObserved
		plan.Snapshots[i].PolicyFreshUntil = policyFresh
		plan.Snapshots[i].SnapshotObservedAt = snapshotObserved
		plan.Snapshots[i].SnapshotFreshUntil = snapshotFresh
	}
	return policyObserved, policyFresh, snapshotObserved, snapshotFresh
}
