package journal

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestRiskBucketAdmissionCommitsQFinalExistingReservationOwnerAndAllBucketsAtomically(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-1", "acct-1")
	plan := riskBucketAdmissionFixture(t, "1", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	receipt, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.QFinal != 10 || receipt.OwnerReused || receipt.Idempotent || len(receipt.ReservationIDs) != 5 {
		t.Fatalf("receipt=%+v", receipt)
	}
	uniqueReservationIDs := make(map[string]bool, len(receipt.ReservationIDs))
	for _, id := range receipt.ReservationIDs {
		uniqueReservationIDs[id] = true
	}
	if len(uniqueReservationIDs) != 5 {
		t.Fatalf("receipt reservation ids are not exact/unique: %v", receipt.ReservationIDs)
	}
	for table, want := range map[string]int{"risk_bucket_final_decisions": 1, "risk_bucket_owners": 1, "risk_bucket_reservations": 5, "risk_bucket_events": 1} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != want {
			t.Fatalf("%s count=%d want=%d err=%v", table, count, want, err)
		}
	}
	var existingState string
	if err := j.db.QueryRow(`SELECT state FROM risk_reservations WHERE id='existing-1'`).Scan(&existingState); err != nil || existingState != ReservationHeld {
		t.Fatalf("existing reservation state=%q err=%v", existingState, err)
	}
	replay, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil || !replay.Idempotent || replay.QFinal != receipt.QFinal {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
}

func TestRiskBucketAdmissionRollsBackPartialBucketWriteAndOwner(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-rollback", "acct-1")
	if _, err := j.db.Exec(`CREATE TRIGGER fail_sector_bucket BEFORE INSERT ON risk_bucket_reservations WHEN NEW.bucket_dimension='sector' BEGIN SELECT RAISE(ABORT,'synthetic sector failure'); END`); err != nil {
		t.Fatal(err)
	}
	plan := riskBucketAdmissionFixture(t, "rollback", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err == nil {
		t.Fatal("synthetic partial bucket failure committed")
	}
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_events"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s partial rows=%d err=%v", table, count, err)
		}
	}
}

func TestRiskBucketAdmissionConcurrentProspectiveOwnersHaveOneWinner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	first := openTestJournalAt(t, path)
	second := openTestJournalAt(t, path)
	seedExistingRiskReservation(t, first, "existing-a", "acct-1")
	seedExistingRiskReservation(t, first, "existing-b", "acct-1")
	plans := []RiskBucketAdmissionPlan{
		riskBucketAdmissionFixture(t, "a", "acct-1", "lane-short", "campaign-a", "prospective-a", "100", "0"),
		riskBucketAdmissionFixture(t, "b", "acct-1", "lane-medium", "campaign-b", "prospective-b", "100", "0"),
	}
	plans[0].ExistingReservationID = "existing-a"
	plans[1].ExistingReservationID = "existing-b"
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index, journal := range []*Journal{first, second} {
		wait.Add(1)
		go func(index int, journal *Journal) {
			defer wait.Done()
			_, err := journal.CommitRiskBucketAdmission(context.Background(), plans[index])
			results <- err
		}(index, journal)
	}
	wait.Wait()
	close(results)
	success, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrRiskBucketOwnerConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent error=%v", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
	for table, want := range map[string]int{"risk_bucket_final_decisions": 1, "risk_bucket_owners": 1, "risk_bucket_reservations": 5} {
		var count int
		if err := first.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != want {
			t.Fatalf("%s count=%d want=%d err=%v", table, count, want, err)
		}
	}
}

func TestRiskBucketAdmissionRejectsSnapshotDriftAndOrphanReferenceWithoutWrites(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-drift", "acct-1")
	plan := riskBucketAdmissionFixture(t, "drift", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	plan.Snapshots[2].SnapshotVersion = "wrong-version"
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("snapshot drift error=%v", err)
	}
	plan = riskBucketAdmissionFixture(t, "orphan", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	plan.ExistingReservationID = "missing"
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrReservationNotFound) {
		t.Fatalf("orphan existing reservation error=%v", err)
	}
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s rows=%d err=%v", table, count, err)
		}
	}
}

func TestReadRiskBucketStateReturnsStableMismatchInsteadOfRepairingDrift(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-replay", "acct-1")
	plan := riskBucketAdmissionFixture(t, "replay", "acct-1", "lane-short", "campaign-1", "prospective-replay", "100", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	key := riskBucketOwnerKey("acct-1", "prospective-replay")
	if state, err := j.ReadRiskBucketState(context.Background(), key); err != nil || state.Digest == "" || len(state.Usage) != 5 {
		t.Fatalf("initial replay state=%+v err=%v", state, err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='49' WHERE bucket_dimension='sector'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ReadRiskBucketState(context.Background(), key); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("drift replay error=%v", err)
	}
	var count int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations`).Scan(&count); err != nil || count != 5 {
		t.Fatalf("replay repaired/deleted reservations count=%d err=%v", count, err)
	}
}

func TestRiskBucketAdmissionRejectsImmutablePolicyAndSnapshotCollision(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-collision-a", "acct-1")
	first := riskBucketAdmissionFixture(t, "collision-a", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-collision-b", "acct-1")
	second := riskBucketAdmissionFixture(t, "collision-b", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	second.Admission.Policy.Price.Digest = "different-price-digest"
	if _, err := j.CommitRiskBucketAdmission(context.Background(), second); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("policy collision error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("decisions=%d", got)
	}
	second = riskBucketAdmissionFixture(t, "collision-b", "acct-1", "lane-short", "campaign-1", "prospective-1", "100", "0")
	second.Snapshots[0].SnapshotID = first.Snapshots[0].SnapshotID
	if _, err := j.CommitRiskBucketAdmission(context.Background(), second); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("snapshot collision error=%v", err)
	}
}

func TestRiskBucketOwnerReuseRequiresExactProspectiveIdentity(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-owner-a", "acct-1")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), riskBucketAdmissionFixture(t, "owner-a", "acct-1", "lane-short", "campaign-1", "prospective-a", "100", "0")); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-owner-b", "acct-1")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), riskBucketAdmissionFixture(t, "owner-b", "acct-1", "lane-short", "campaign-1", "prospective-b", "100", "0")); !errors.Is(err, ErrRiskBucketOwnerConflict) {
		t.Fatalf("prospective conflict error=%v", err)
	}
}

func TestRiskBucketAdmissionRejectsOwnerMarketAndSymbolBucketMismatchWithoutWrites(t *testing.T) {
	for _, tc := range []struct {
		name  string
		index int
		key   riskbucket.BucketKey
	}{
		{"market", 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"}},
		{"symbol", 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			seedExistingRiskReservation(t, j, "existing-binding-"+tc.name, "acct-1")
			plan := riskBucketAdmissionFixture(t, "binding-"+tc.name, "acct-1", "lane-short", "campaign-1", "prospective-binding", "100", "0")
			rebindRiskBucket(t, &plan, tc.index, tc.key)
			if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
				t.Fatalf("binding error=%v", err)
			}
			for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations"} {
				if got := countRiskBucketRows(t, j, table); got != 0 {
					t.Fatalf("%s rows=%d", table, got)
				}
			}
		})
	}
}

func TestRiskBucketAdmissionRejectsSnapshotReferenceEvidenceTamperWithoutWrites(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*RiskBucketSnapshotReference)
	}{
		{"policy-digest", func(r *RiskBucketSnapshotReference) { r.PolicyDigest = "tampered-policy" }},
		{"snapshot-digest", func(r *RiskBucketSnapshotReference) { r.SnapshotDigest = "tampered-snapshot" }},
		{"observed-at", func(r *RiskBucketSnapshotReference) { r.ObservedAt = r.ObservedAt.Add(time.Second) }},
		{"fresh-until", func(r *RiskBucketSnapshotReference) { r.FreshUntil = r.FreshUntil.Add(time.Second) }},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			seedExistingRiskReservation(t, j, "existing-evidence-"+tc.name, "acct-1")
			plan := riskBucketAdmissionFixture(t, "evidence-"+tc.name, "acct-1", "lane-short", "campaign-1", "prospective-evidence", "100", "0")
			tc.mutate(&plan.Snapshots[2])
			if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
				t.Fatalf("tamper error=%v", err)
			}
			if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 0 {
				t.Fatalf("decisions=%d", got)
			}
		})
	}
}

func TestRiskBucketAdmissionDivergentRetryCannotHideChangedConsumedSnapshotAmounts(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-retry-binding", "acct-1")
	plan := riskBucketAdmissionFixture(t, "retry-binding", "acct-1", "lane-short", "campaign-1", "prospective-retry-binding", "100", "0")
	first, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if first.QFinal != 10 {
		t.Fatalf("first=%+v", first)
	}
	for i := range plan.Admission.Buckets {
		bucket := plan.Admission.Buckets[i]
		bound := bucket.BoundEvidence()
		binding := bound.Snapshot
		binding.LimitMinor = "110"
		binding.FilledMinor = "10"
		snapshot, err := riskbucket.NewSnapshotProvenance(binding, bound.SnapshotEvidence)
		if err != nil {
			t.Fatal(err)
		}
		bucket.LimitMinor = binding.LimitMinor
		bucket.FilledMinor = binding.FilledMinor
		bucket.SnapshotProvenance = snapshot
		plan.Admission.Buckets[i] = bucket
	}
	decision := riskbucket.CalculateAdmission(plan.Admission)
	if decision.Refusal != nil || decision.QFinal != first.QFinal {
		t.Fatalf("fixture did not preserve admission result: %+v", decision)
	}
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("divergent retry error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("decisions=%d", got)
	}
}

func TestRiskBucketSameOwnerScaleInUsesMonotonicAggregateReplay(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-scale-a", "acct-1")
	first := riskBucketAdmissionFixture(t, "scale-a", "acct-1", "lane-short", "campaign-1", "prospective-scale", "200", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-scale-b", "acct-1")
	second := riskBucketAdmissionFixture(t, "scale-b", "acct-1", "lane-short", "campaign-1", "prospective-scale", "200", "50")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	state, err := j.ReadRiskBucketState(context.Background(), riskBucketOwnerKey("acct-1", "prospective-scale"))
	if err != nil {
		t.Fatal(err)
	}
	if state.QFinal != 20 || len(state.Reservations) != 5 {
		t.Fatalf("aggregate state=%+v", state)
	}
	for key, amount := range state.Reservations {
		if amount != "100" {
			t.Fatalf("%+v amount=%s", key, amount)
		}
	}
	var sequences int
	if err := j.db.QueryRow(`SELECT count(DISTINCT event_sequence) FROM risk_bucket_state_snapshots`).Scan(&sequences); err != nil || sequences != 2 {
		t.Fatalf("sequences=%d err=%v", sequences, err)
	}
}

func TestRiskBucketSameOwnerScaleInRejectsBucketKeyOrPolicyVersionChange(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-key-a", "acct-1")
	first := riskBucketAdmissionFixture(t, "key-a", "acct-1", "lane-short", "campaign-1", "prospective-key", "200", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-key-b", "acct-1")
	second := riskBucketAdmissionFixture(t, "key-b", "acct-1", "lane-short", "campaign-1", "prospective-key", "200", "50")
	rebindRiskBucket(t, &second, 2, riskbucket.BucketKey{Dimension: riskbucket.DimensionStrategy, Value: "strategy-beta", PolicyVersion: "policy-v2"})
	if _, err := j.CommitRiskBucketAdmission(context.Background(), second); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("scale-in key drift error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("decisions=%d", got)
	}
}

func TestReadRiskBucketStateRejectsMissingReservationAndSnapshotRebinding(t *testing.T) {
	for _, mutation := range []string{
		`DELETE FROM risk_bucket_reservations WHERE bucket_dimension='sector'`,
		`UPDATE risk_bucket_reservations SET snapshot_id=(SELECT snapshot_id FROM risk_bucket_snapshots WHERE bucket_dimension='horizon') WHERE bucket_dimension='sector'`,
	} {
		t.Run(mutation[:6], func(t *testing.T) {
			j := openTestJournal(t)
			seedExistingRiskReservation(t, j, "existing-structural", "acct-1")
			plan := riskBucketAdmissionFixture(t, "structural", "acct-1", "lane-short", "campaign-1", "prospective-structural", "100", "0")
			if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(mutation); err != nil {
				t.Fatal(err)
			}
			if _, err := j.ReadRiskBucketState(context.Background(), riskBucketOwnerKey("acct-1", "prospective-structural")); !errors.Is(err, ErrRiskBucketReplayMismatch) {
				t.Fatalf("structural drift error=%v", err)
			}
		})
	}
}

func riskBucketAdmissionFixture(t *testing.T, suffix, account, lane, campaign, prospective, limit, held string) RiskBucketAdmissionPlan {
	t.Helper()
	now := time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)
	policy := riskbucket.ReservePolicy{
		AccountCurrency: "KRW", QuoteCurrency: "KRW", EvaluatedAt: now,
		Price: riskbucket.PriceEvidence{WorstExecutableQuote: "5", Evidence: riskbucket.Evidence{Source: "official-order-contract", Version: "price-v1", Digest: "price-digest", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}},
		FX:    riskbucket.FXEvidence{RateQuoteToBase: "1", Haircut: "1", Evidence: riskbucket.Evidence{Source: "same-currency", Version: "fx-v1", Digest: "fx-digest", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}},
		Fee:   riskbucket.FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "0", MinimumBaseMinor: "0", Version: "fee-v1", Digest: "fee-digest"},
	}
	values := map[riskbucket.Dimension]string{
		riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "KR",
		riskbucket.DimensionStrategy: "strategy-alpha", riskbucket.DimensionSector: "sector-tech",
		riskbucket.DimensionSymbol: "005930",
	}
	buckets := make([]riskbucket.BucketSnapshot, 0, 5)
	references := make([]RiskBucketSnapshotReference, 0, 5)
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		key := riskbucket.BucketKey{Dimension: dimension, Value: values[dimension], PolicyVersion: "policy-v1"}
		policyProvenance, err := riskbucket.NewPolicyProvenance(key, riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-" + string(dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		binding := riskbucket.BucketSnapshotBinding{Key: key, LimitMinor: limit, FilledMinor: "0", HeldMinor: held, SnapshotVersion: "snapshot-v1"}
		snapshotEvidenceDigest := "snapshot-" + suffix + "-" + string(dimension)
		snapshotProvenance, err := riskbucket.NewSnapshotProvenance(binding, riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion, Digest: snapshotEvidenceDigest, Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		buckets = append(buckets, riskbucket.BucketSnapshot{Key: key, LimitMinor: limit, FilledMinor: "0", HeldMinor: held, SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance})
		references = append(references, RiskBucketSnapshotReference{Key: key, SnapshotID: "snapshot-" + suffix + "-" + string(dimension), SnapshotDigest: snapshotEvidenceDigest, SnapshotVersion: binding.SnapshotVersion, PolicyDigest: "policy-" + string(dimension), ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
	}
	return RiskBucketAdmissionPlan{
		TransactionID: "risk-tx-" + suffix, DecisionID: "risk-final-" + suffix,
		ExistingReservationID: "existing-" + suffix,
		Admission:             riskbucket.AdmissionRequest{QCandidate: 10, QExistingGuardian: 10, Policy: policy, Buckets: buckets},
		Owner:                 riskbucket.OwnerClaim{Key: riskbucket.OwnerKey{AccountID: account, Market: riskbucket.MarketKR, Symbol: "005930", ProspectiveGeneration: prospective}, LaneID: lane, CampaignID: campaign},
		Snapshots:             references, CreatedAt: now,
	}
}

func riskBucketOwnerKey(account, prospective string) riskbucket.OwnerKey {
	return riskbucket.OwnerKey{AccountID: account, Market: riskbucket.MarketKR, Symbol: "005930", ProspectiveGeneration: prospective}
}

func seedExistingRiskReservation(t *testing.T, j *Journal, id, account string) {
	t.Helper()
	decisionID := "decision-" + id
	if _, err := j.db.Exec(`INSERT INTO decisions(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,nonce,issued_at,expires_at) VALUES(?,?,0,'EXPOSURE_RAISING','RISK_INTENT','{}',?,?,?,?)`, decisionID, account, "hash-"+id, "nonce-"+id, "2026-03-30T00:29:00Z", "2026-03-30T01:30:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO risk_reservations(id,decision_id,account_ref,kind,amount,currency,snapshot_as_of,state) VALUES(?,?,?,'OPEN_EXPOSURE','1','KRW','2026-03-30T00:29:00Z','HELD')`, id, decisionID, account); err != nil {
		t.Fatal(err)
	}
}

func countRiskBucketRows(t *testing.T, j *Journal, table string) int {
	t.Helper()
	var count int
	if err := j.db.QueryRow(fmt.Sprintf("SELECT count(*) FROM %s", table)).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func rebindRiskBucket(t *testing.T, plan *RiskBucketAdmissionPlan, index int, key riskbucket.BucketKey) {
	t.Helper()
	bucket := plan.Admission.Buckets[index]
	now := plan.Admission.Policy.EvaluatedAt
	policy, err := riskbucket.NewPolicyProvenance(key, riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-rebound", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	binding := riskbucket.BucketSnapshotBinding{Key: key, LimitMinor: bucket.LimitMinor, FilledMinor: bucket.FilledMinor, HeldMinor: bucket.HeldMinor, SnapshotVersion: bucket.SnapshotVersion}
	snapshot, err := riskbucket.NewSnapshotProvenance(binding, riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: bucket.SnapshotVersion, Digest: "snapshot-rebound", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	bucket.Key = key
	bucket.PolicyProvenance = policy
	bucket.SnapshotProvenance = snapshot
	plan.Admission.Buckets[index] = bucket
	plan.Snapshots[index].Key = key
	plan.Snapshots[index].PolicyDigest = "policy-rebound"
	plan.Snapshots[index].SnapshotDigest = "snapshot-rebound"
}
