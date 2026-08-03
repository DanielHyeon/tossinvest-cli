package journal

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestStrategyDispatchOwnerEpochFencesPredecessorAndDormantAPIsMintNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first, err := j.AcquireStrategyDispatchOwner(ctx, "dispatch-a")
	if err != nil || first.Epoch != 1 || first.FencingToken == "" {
		t.Fatalf("first owner=%+v err=%v", first, err)
	}
	second, err := j.AcquireStrategyDispatchOwner(ctx, "dispatch-b")
	if err != nil || second.Epoch != 2 || second.FencingToken == first.FencingToken {
		t.Fatalf("second owner=%+v first=%+v err=%v", second, first, err)
	}
	if _, err := j.DiscoverStrategyDispatchRecovery(ctx, first); !errors.Is(err, ErrStrategyDispatchFenced) {
		t.Fatalf("stale recovery owner error=%v", err)
	}
	if items, err := j.DiscoverStrategyDispatchRecovery(ctx, second); err != nil || len(items) != 0 {
		t.Fatalf("current recovery items=%+v err=%v", items, err)
	}

	authority := StrategyDispatchMarketAuthority{AccountRef: "acct-1", Market: StrategyDispatchMarketKR, Symbol: "005930"}
	if _, err := j.CommitStrategyDispatchMarketAuthority(ctx, authority); !errors.Is(err, ErrStrategyDispatchDormant) {
		t.Fatalf("authority mint error=%v", err)
	}
	plan := StrategyDispatchLeasePlan{LeaseID: "lease-dormant", OwnerEpoch: second.Epoch, FencingToken: second.FencingToken}
	for name, err := range map[string]error{
		"issue": func() error { _, issueErr := j.IssueStrategyDispatchLease(ctx, plan); return issueErr }(),
		"claim": func() error {
			_, claimErr := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{})
			return claimErr
		}(),
		"submit": func() error {
			_, submitErr := j.BeginStrategyDispatchSubmitting(ctx, StrategyDispatchLeaseCAS{})
			return submitErr
		}(),
		"recover": func() error {
			_, recoverErr := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{})
			return recoverErr
		}(),
	} {
		if !errors.Is(err, ErrStrategyDispatchDormant) {
			t.Errorf("%s error=%v", name, err)
		}
	}
	for _, table := range []string{"strategy_dispatch_market_authorities", "strategy_dispatch_leases", "strategy_dispatch_outcomes"} {
		var count int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s rows=%d err=%v", table, count, err)
		}
	}
}

func TestStrategyDispatchLeaseSchemaRequiresExactQFinalAuthorityAndHolds(t *testing.T) {
	t.Run("valid-sealed-row", func(t *testing.T) {
		j := openTestJournal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-valid")
		plan := prepareStrategyDispatchLease(t, j, "schema-valid", owner, StrategyDispatchMarketKR, "005930")
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err != nil {
			t.Fatalf("exact q_final row: %v", err)
		}
	})

	t.Run("fabricated-guardian", func(t *testing.T) {
		j := openTestJournal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-forged")
		plan := prepareStrategyDispatchLease(t, j, "schema-forged", owner, StrategyDispatchMarketKR, "005930")
		plan.GuardianDecisionID = "fabricated-guardian-decision"
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil {
			t.Fatal("fabricated Guardian decision was accepted")
		}
	})

	t.Run("one-monetary-hold-released", func(t *testing.T) {
		j := openTestJournal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-released")
		plan := prepareStrategyDispatchLease(t, j, "schema-released", owner, StrategyDispatchMarketKR, "005930")
		if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',held_minor='0',updated_at='2026-03-30T00:30:02Z' WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
			t.Fatal(err)
		}
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil || !strings.Contains(err.Error(), "exact q_final authority") {
			t.Fatalf("released monetary hold error=%v", err)
		}
	})

	t.Run("authority-digest-substitution", func(t *testing.T) {
		j := openTestJournal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-authority")
		plan := prepareStrategyDispatchLease(t, j, "schema-authority", owner, StrategyDispatchMarketKR, "005930")
		plan.AuthorityDigest = "substituted-authority-digest"
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil || !strings.Contains(err.Error(), "exact q_final authority") {
			t.Fatalf("substituted authority error=%v", err)
		}
	})
}

func TestStrategyDispatchBrokerOrderIDCannotCrossKRUSWithinAccount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	owner, err := j.AcquireStrategyDispatchOwner(ctx, "cross-market-owner")
	if err != nil {
		t.Fatal(err)
	}
	kr := prepareStrategyDispatchLease(t, j, "broker-kr", owner, StrategyDispatchMarketKR, "005930")
	us := prepareStrategyDispatchLease(t, j, "broker-us", owner, StrategyDispatchMarketUS, "AAPL")
	const brokerOrderID = "broker-order-reused-across-markets"
	if err := insertStrategyDispatchLease(j, kr, StrategyDispatchLeaseSubmitted, brokerOrderID); err != nil {
		t.Fatalf("first broker binding: %v", err)
	}
	if err := insertStrategyDispatchLease(j, us, StrategyDispatchLeaseSubmitted, brokerOrderID); err == nil || !strings.Contains(err.Error(), "strategy_dispatch_leases.account_ref, strategy_dispatch_leases.broker_order_id") {
		t.Fatalf("cross-market broker reuse error=%v", err)
	}
}

func TestStrategyDispatchColdRestartDiscoversOldIssuedClaimedAndSubmitting(t *testing.T) {
	for _, tc := range []struct {
		state StrategyDispatchLeaseState
		want  StrategyDispatchRecoveryAction
	}{
		{StrategyDispatchLeaseIssued, StrategyDispatchRecoveryRefuseRelease},
		{StrategyDispatchLeaseClaimed, StrategyDispatchRecoveryRefuseRelease},
		{StrategyDispatchLeaseSubmitting, StrategyDispatchRecoveryAttestedOutcome},
	} {
		t.Run(string(tc.state), func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "journal.db")
			before := openStrategyDispatchTestJournal(t, path)
			oldOwner, err := before.AcquireStrategyDispatchOwner(ctx, "before-crash-"+strings.ToLower(string(tc.state)))
			if err != nil {
				t.Fatal(err)
			}
			plan := prepareStrategyDispatchLease(t, before, "restart-"+strings.ToLower(string(tc.state)), oldOwner, StrategyDispatchMarketKR, "005930")
			if err := insertStrategyDispatchLease(before, plan, tc.state, ""); err != nil {
				t.Fatal(err)
			}
			if err := before.Close(); err != nil {
				t.Fatal(err)
			}

			after := openStrategyDispatchTestJournal(t, path)
			defer after.Close()
			newOwner, err := after.AcquireStrategyDispatchOwner(ctx, "after-crash-"+strings.ToLower(string(tc.state)))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := after.DiscoverStrategyDispatchRecovery(ctx, oldOwner); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("old owner discovery error=%v", err)
			}
			items, err := after.DiscoverStrategyDispatchRecovery(ctx, newOwner)
			if err != nil || len(items) != 1 || items[0].Lease.LeaseID != plan.LeaseID || items[0].Lease.OwnerEpoch != oldOwner.Epoch || items[0].Action != tc.want {
				t.Fatalf("recovery items=%+v err=%v", items, err)
			}
			var state, disposition string
			if err := after.db.QueryRow(`SELECT state,disposition FROM strategy_dispatch_leases WHERE lease_id=?`, plan.LeaseID).Scan(&state, &disposition); err != nil {
				t.Fatal(err)
			}
			if state != string(tc.state) || disposition != string(StrategyDispatchReservationReserved) {
				t.Fatalf("read-only discovery mutated state=%s disposition=%s", state, disposition)
			}
		})
	}
}

func prepareStrategyDispatchLease(t *testing.T, j *Journal, suffix string, owner StrategyDispatchOwner, market StrategyDispatchMarket, symbol string) StrategyDispatchLeasePlan {
	t.Helper()
	request := qFinalIssueFixture(t, j, suffix)
	if market == StrategyDispatchMarketUS {
		request.Admission.Owner.Key.Market = riskbucket.MarketUS
		request.Admission.Owner.Key.Symbol = symbol
		for i, bucket := range request.Admission.Admission.Buckets {
			switch bucket.Key.Dimension {
			case riskbucket.DimensionMarket:
				rebindRiskBucket(t, &request.Admission, i, riskbucket.BucketKey{Dimension: bucket.Key.Dimension, Value: string(riskbucket.MarketUS), PolicyVersion: bucket.Key.PolicyVersion})
			case riskbucket.DimensionSymbol:
				rebindRiskBucket(t, &request.Admission, i, riskbucket.BucketKey{Dimension: bucket.Key.Dimension, Value: symbol, PolicyVersion: bucket.Key.PolicyVersion})
			}
		}
		intent := request.Issue.Decision.Preimage.(RiskIntent)
		intent.Market = "us"
		intent.Symbol = symbol
		request.Issue.Decision.Preimage = intent
	}
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), request); err != nil {
		t.Fatalf("record q_final %s: %v", suffix, err)
	}
	recordDigest := "sealed-authority-record-" + suffix
	if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(
		authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
		protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
		risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
		VALUES(?,?,?,?,1,?,1,1,?,?,1,1,?,1,?,?,1,?,?)`,
		"sealed-authority-"+suffix, request.Admission.Owner.Key.AccountID, market, symbol,
		"sealed-activation-"+suffix, "sealed-protection-serial-"+suffix, "sealed-protection-digest-"+suffix,
		"sealed-risk-policy-"+suffix, "sealed-guardian-"+suffix, "build-"+suffix, recordDigest, "2026-03-30T00:30:01Z"); err != nil {
		t.Fatalf("seed sealed authority %s: %v", suffix, err)
	}
	return StrategyDispatchLeasePlan{
		LeaseID: "lease-" + suffix, OperationID: "operation-" + suffix,
		AccountRef: request.Admission.Owner.Key.AccountID, Market: market, Symbol: symbol,
		CandidateID: "candidate-" + suffix, EvidenceDigest: "evidence-" + suffix,
		RouterID: "router", RouterVersion: "router-v1", LaneID: request.Admission.Owner.LaneID, LaneVersion: "lane-v1",
		CampaignID: request.Admission.Owner.CampaignID, LegID: "leg-" + suffix,
		RiskReservationID: request.Admission.ExistingReservationID, GuardianDecisionID: request.Issue.Decision.ID,
		OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
		AuthorityRevision: 1, AuthorityDigest: recordDigest,
		IssuedAt: time.Date(2026, 3, 30, 0, 30, 5, 0, time.UTC), ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC),
	}
}

func insertStrategyDispatchLease(j *Journal, plan StrategyDispatchLeasePlan, state StrategyDispatchLeaseState, brokerOrderID string) error {
	disposition := StrategyDispatchReservationReserved
	revision := uint64(1)
	var transportStarted, outcomeObserved any
	outcomeCode, queryDigest := "", ""
	switch state {
	case StrategyDispatchLeaseClaimed:
		revision = 2
	case StrategyDispatchLeaseSubmitting:
		revision = 3
		transportStarted = "2026-03-30T00:30:10Z"
	case StrategyDispatchLeaseSubmitted:
		revision = 4
		disposition = StrategyDispatchReservationTransferred
		transportStarted = "2026-03-30T00:30:10Z"
		outcomeObserved = "2026-03-30T00:30:20Z"
		outcomeCode = "OFFICIAL_ACCEPTED"
		queryDigest = "official-query-" + plan.LeaseID
	}
	_, err := j.db.Exec(`INSERT INTO strategy_dispatch_leases(
		lease_id,operation_id,account_ref,market,symbol,candidate_id,evidence_digest,router_id,router_version,
		lane_id,lane_version,campaign_id,leg_id,risk_reservation_id,guardian_decision_id,owner_epoch,fencing_token,
		authority_revision,authority_digest,issued_at,expires_at,state,disposition,revision,transport_started_at,
		refusal_code,outcome_code,broker_order_id,query_digest,outcome_observed_at,lease_digest,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		plan.LeaseID, plan.OperationID, plan.AccountRef, plan.Market, plan.Symbol, plan.CandidateID, plan.EvidenceDigest,
		plan.RouterID, plan.RouterVersion, plan.LaneID, plan.LaneVersion, plan.CampaignID, plan.LegID,
		plan.RiskReservationID, plan.GuardianDecisionID, plan.OwnerEpoch, plan.FencingToken, plan.AuthorityRevision,
		plan.AuthorityDigest, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.ExpiresAt), state, disposition,
		revision, transportStarted, "", outcomeCode, brokerOrderID, queryDigest, outcomeObserved,
		"lease-digest-"+plan.LeaseID, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.IssuedAt))
	return err
}

func openStrategyDispatchTestJournal(t *testing.T, path string) *Journal {
	t.Helper()
	j, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	return j
}
