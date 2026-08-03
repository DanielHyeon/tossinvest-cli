package journal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestQFinalIssuanceCommitsDecisionAggregateFiveBucketsAndOwnerAtomically(t *testing.T) {
	j := openTestJournal(t)
	request := qFinalIssueFixture(t, j, "atomic")
	result, err := j.RecordQFinalDecisionAndReserve(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Admission.QFinal != 10 || len(result.Admission.ReservationIDs) != 5 || len(result.Issue.Reservations) != 1 {
		t.Fatalf("result=%+v", result)
	}
	for table, want := range map[string]int{
		"decisions": 1, "risk_reservations": 1, "risk_bucket_final_decisions": 1,
		"risk_bucket_owners": 1, "risk_bucket_reservations": 5, "risk_bucket_events": 1,
	} {
		if got := countRiskBucketRows(t, j, table); got != want {
			t.Fatalf("%s rows=%d want=%d", table, got, want)
		}
	}
	if required, err := j.RevalidateQFinalAdmission(context.Background(), request.Issue.Decision.ID); err != nil || !required {
		t.Fatalf("revalidate required=%v err=%v", required, err)
	}
	replay, err := j.RecordQFinalDecisionAndReserve(context.Background(), request)
	if err != nil {
		t.Fatalf("identical replay: %v", err)
	}
	if !replay.Admission.Idempotent || replay.Admission.OwnerReused != result.Admission.OwnerReused || replay.Admission.QFinal != result.Admission.QFinal ||
		replay.Issue.Version != result.Issue.Version || replay.Issue.Decision.ID != result.Issue.Decision.ID || len(replay.Issue.Reservations) != 1 || replay.Issue.Reservations[0].ID != result.Issue.Reservations[0].ID {
		t.Fatalf("replay=%+v first=%+v", replay, result)
	}
	for table, want := range map[string]int{
		"decisions": 1, "risk_reservations": 1, "risk_bucket_final_decisions": 1,
		"risk_bucket_owners": 1, "risk_bucket_reservations": 5, "risk_bucket_events": 1,
	} {
		if got := countRiskBucketRows(t, j, table); got != want {
			t.Fatalf("identical replay duplicated %s: rows=%d want=%d", table, got, want)
		}
	}
}

func TestQFinalIssuanceDivergentReplayFailsClosedWithoutDuplicateAuthority(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *QFinalIssueRequest)
	}{
		{"decision-limits", func(_ *testing.T, request *QFinalIssueRequest) {
			request.Issue.Decision.LimitsJSON = `{"test":"divergent"}`
		}},
		{"aggregate-usage", func(_ *testing.T, request *QFinalIssueRequest) { request.Issue.Reserve.SnapshotUsage[0].Amount = "1" }},
		{"valid-same-q-final-authority", func(t *testing.T, request *QFinalIssueRequest) {
			key := request.Admission.Admission.Buckets[0].Key
			key.PolicyVersion = "policy-substituted-v2"
			rebindRiskBucket(t, &request.Admission, 0, key)
			decision := riskbucket.CalculateAdmission(request.Admission.Admission)
			if decision.Refusal != nil || decision.QFinal != 10 {
				t.Fatalf("substitute is not valid same-q_final: %+v", decision)
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			request := qFinalIssueFixture(t, j, "divergent-"+tc.name)
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			request.Issue.Reserve.SnapshotUsage = append([]AggregateAmount(nil), request.Issue.Reserve.SnapshotUsage...)
			request.Admission.Admission.Buckets = append([]riskbucket.BucketSnapshot(nil), request.Admission.Admission.Buckets...)
			request.Admission.Snapshots = append([]RiskBucketSnapshotReference(nil), request.Admission.Snapshots...)
			tc.mutate(t, &request)
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), request); !errors.Is(err, ErrRiskBucketReplayMismatch) {
				t.Fatalf("error=%v want replay mismatch", err)
			}
			for table, want := range map[string]int{"decisions": 1, "risk_reservations": 1, "risk_bucket_final_decisions": 1, "risk_bucket_reservations": 5} {
				if got := countRiskBucketRows(t, j, table); got != want {
					t.Fatalf("divergent replay changed %s: rows=%d want=%d", table, got, want)
				}
			}
		})
	}
}

func TestQFinalIssuanceOwnerConflictRollsBackDecisionAndEveryReservation(t *testing.T) {
	j := openTestJournal(t)
	first := qFinalIssueFixture(t, j, "winner")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	loser := qFinalIssueFixture(t, j, "loser")
	loser.Admission.Owner.Key.ProspectiveGeneration = "prospective-loser"
	loser.Admission.Owner.LaneID = "lane-loser"
	loser.Admission.Owner.CampaignID = "campaign-loser"
	beforeDecisions := countRiskBucketRows(t, j, "decisions")
	beforeLegacy := countRiskBucketRows(t, j, "risk_reservations")
	beforeBuckets := countRiskBucketRows(t, j, "risk_bucket_reservations")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), loser); !errors.Is(err, ErrRiskBucketOwnerConflict) {
		t.Fatalf("error=%v want owner conflict", err)
	}
	if got := countRiskBucketRows(t, j, "decisions"); got != beforeDecisions {
		t.Fatalf("decision leaked: %d -> %d", beforeDecisions, got)
	}
	if got := countRiskBucketRows(t, j, "risk_reservations"); got != beforeLegacy {
		t.Fatalf("legacy reservation leaked: %d -> %d", beforeLegacy, got)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_reservations"); got != beforeBuckets {
		t.Fatalf("monetary reservation leaked: %d -> %d", beforeBuckets, got)
	}
}

func TestQFinalMarkedDecisionWithoutAdmissionFailsClosed(t *testing.T) {
	j := openTestJournal(t)
	request := qFinalIssueFixture(t, j, "missing")
	if _, err := j.RecordDecisionAndReserve(context.Background(), request.Issue); err != nil {
		t.Fatal(err)
	}
	required, err := j.RevalidateQFinalAdmission(context.Background(), request.Issue.Decision.ID)
	if !required || !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("required=%v error=%v", required, err)
	}
}

func qFinalIssueFixture(t *testing.T, j *Journal, suffix string) QFinalIssueRequest {
	t.Helper()
	plan := riskBucketAdmissionFixture(t, suffix, "acct-1", "lane-short", "campaign-"+suffix, "prospective-"+suffix, "100", "0")
	policyVersion, err := QFinalPolicyVersion("guardian-v1", plan.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	decisionID := "qfinal-decision-" + suffix
	reservationID := "qfinal-existing-" + suffix
	plan.DecisionID = decisionID
	plan.ExistingReservationID = reservationID
	now := time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)
	version, err := j.ReservationVersion(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	return QFinalIssueRequest{
		Issue: IssueRequest{
			Decision: DecisionRequest{
				ID: decisionID, AccountRef: "acct-1", SafetyClass: SafetyClassExposureRaising, Kind: KindPlace,
				Preimage:   RiskIntent{AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "10", EntryPrice: "5", StopPrice: "4", TargetPrice: "7", PolicyVersion: policyVersion},
				LimitsJSON: `{"test":"opaque-to-journal"}`, Nonce: "nonce-" + suffix, IssuedAt: now, ExpiresAt: now.Add(time.Minute),
			},
			Reserve: ReserveRequest{
				SnapshotAsOf: now, ObservedVersion: version,
				SnapshotUsage: []AggregateAmount{{Kind: ReservationKindOpenExposure, Amount: "0", Currency: "KRW"}},
				Limits:        []AggregateAmount{{Kind: ReservationKindOpenExposure, Amount: "100000", Currency: "KRW"}},
				Reservations:  []ReservationRequest{{ID: reservationID, Kind: ReservationKindOpenExposure, Amount: "50", Currency: "KRW"}},
			},
		},
		Admission: plan,
	}
}
