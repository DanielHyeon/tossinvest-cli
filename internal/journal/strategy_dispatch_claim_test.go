package journal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestClaimStrategyDispatchLeasePairedKRUSCurrentAuthority(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	owner, err := j.AcquireStrategyDispatchOwner(ctx, "paired-claim-owner")
	if err != nil {
		t.Fatal(err)
	}
	plans := []StrategyDispatchLeasePlan{
		seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-current-kr", "acct-kr", "KR", "005930"),
		seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-current-us", "acct-us", "US", "AAPL"),
	}

	results := make(chan struct {
		lease StrategyDispatchLease
		err   error
	}, len(plans))
	var wait sync.WaitGroup
	for _, plan := range plans {
		plan := plan
		wait.Add(1)
		go func() {
			defer wait.Done()
			lease, claimErr := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 1,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			})
			results <- struct {
				lease StrategyDispatchLease
				err   error
			}{lease: lease, err: claimErr}
		}()
	}
	wait.Wait()
	close(results)
	markets := map[StrategyDispatchMarket]bool{}
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.lease.State != StrategyDispatchLeaseClaimed || result.lease.Disposition != StrategyDispatchReservationReserved || result.lease.Revision != 2 {
			t.Fatalf("claimed lease=%+v", result.lease)
		}
		markets[result.lease.Market] = true
		assertStrategyDispatchRealHolds(t, j, result.lease, "HELD", 5)
	}
	if !markets[StrategyDispatchMarketKR] || !markets[StrategyDispatchMarketUS] {
		t.Fatalf("paired claim markets=%v", markets)
	}
	assertNoStrategyDispatchTransport(t, j)
}

func TestClaimStrategyDispatchLeasePairedKRUSRefusesExpiredChangedCrossMarketAndStaleOwner(t *testing.T) {
	type mutationCase struct {
		name string
		code string
		do   func(*testing.T, *Journal, StrategyDispatchOwner, StrategyDispatchLeasePlan)
	}
	cases := []mutationCase{
		{name: "expired", code: "LEASE_EXPIRED", do: func(t *testing.T, j *Journal, _ StrategyDispatchOwner, plan StrategyDispatchLeasePlan) {
			j.clk.(*clock.Fake).Set(plan.ExpiresAt)
		}},
		{name: "authority-changed", code: "MARKET_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, _ StrategyDispatchOwner, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE strategy_dispatch_market_authorities SET revision=revision+1,record_digest=? WHERE account_ref=? AND market=? AND symbol=?`,
				"changed-authority-"+plan.LeaseID, plan.AccountRef, plan.Market, plan.Symbol); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "authority-aba", code: "MARKET_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, _ StrategyDispatchOwner, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE strategy_dispatch_market_authorities SET revision=revision+1,record_digest=? WHERE account_ref=? AND market=? AND symbol=?`,
				"intermediate-authority-"+plan.LeaseID, plan.AccountRef, plan.Market, plan.Symbol); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE strategy_dispatch_market_authorities SET revision=revision+1,record_digest=? WHERE account_ref=? AND market=? AND symbol=?`,
				plan.AuthorityDigest, plan.AccountRef, plan.Market, plan.Symbol); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "cross-market-lineage", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, _ StrategyDispatchOwner, plan StrategyDispatchLeasePlan) {
			peer := "US"
			if plan.Market == StrategyDispatchMarketUS {
				peer = "KR"
			}
			if _, err := j.db.Exec(`UPDATE position_campaign_claims SET market=? WHERE campaign_id=?`, peer, plan.CampaignID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "stale-bound-owner", code: "OWNER_STALE", do: func(t *testing.T, j *Journal, owner StrategyDispatchOwner, plan StrategyDispatchLeasePlan) {
			newOwner, err := j.AcquireStrategyDispatchOwner(context.Background(), "replacement-"+strings.ToLower(string(plan.Market)))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("stale caller error=%v", err)
			}
			plan.OwnerEpoch, plan.FencingToken = newOwner.Epoch, newOwner.FencingToken
		}},
	}
	markets := []struct{ market, symbol string }{{"KR", "005930"}, {"US", "AAPL"}}
	for _, tc := range cases {
		for _, market := range markets {
			t.Run(tc.name+"/"+market.market, func(t *testing.T) {
				j := openTestJournal(t)
				ctx := context.Background()
				owner, err := j.AcquireStrategyDispatchOwner(ctx, "owner-"+tc.name+"-"+strings.ToLower(market.market))
				if err != nil {
					t.Fatal(err)
				}
				plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-"+tc.name+"-"+strings.ToLower(market.market), "acct-"+strings.ToLower(market.market), market.market, market.symbol)
				tc.do(t, j, owner, plan)
				claimOwner := owner
				if tc.name == "stale-bound-owner" {
					var err error
					claimOwner, err = currentStrategyDispatchOwnerForTest(j)
					if err != nil {
						t.Fatal(err)
					}
				}
				lease, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: 1,
					OwnerEpoch: claimOwner.Epoch, FencingToken: claimOwner.FencingToken,
				})
				if err != nil {
					t.Fatal(err)
				}
				if lease.State != StrategyDispatchLeaseRefused || lease.Disposition != StrategyDispatchReservationReleased || lease.Revision != 2 || lease.RefusalCode != tc.code {
					t.Fatalf("refused lease=%+v want code=%s", lease, tc.code)
				}
				assertStrategyDispatchRealHolds(t, j, lease, "RELEASED", 0)
				assertNoStrategyDispatchTransport(t, j)
			})
		}
	}
}

func TestClaimStrategyDispatchLeaseReleaseCardinalityMismatchRollsBackKRUS(t *testing.T) {
	mutations := []struct {
		name string
		fail bool
		do   func(*testing.T, *Journal, StrategyDispatchLeasePlan)
	}{
		{"aggregate-already-released", false, func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at=?,release_reason=? WHERE id=?`,
				"2026-03-30T00:30:06Z", ReleaseReasonExpiredUnconsumed, plan.RiskReservationID); err != nil {
				t.Fatal(err)
			}
		}},
		{"one-bucket-already-released", false, func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',held_minor='0' WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{"one-bucket-cross-scope", true, func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			peer := "US"
			if plan.Market == StrategyDispatchMarketUS {
				peer = "KR"
			}
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET market=? WHERE decision_id=? AND bucket_dimension='symbol'`, peer, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		for _, mutation := range mutations {
			t.Run(market.name+"/"+mutation.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "release-cardinality-"+strings.ToLower(market.name)+"-"+mutation.name)
				if err != nil {
					t.Fatal(err)
				}
				plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "release-cardinality-"+strings.ToLower(market.name)+"-"+mutation.name,
					"acct-"+strings.ToLower(market.name), market.name, market.symbol)
				var beforeOutcomes int
				if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_outcomes WHERE lease_id=?`, plan.LeaseID).Scan(&beforeOutcomes); err != nil {
					t.Fatal(err)
				}
				mutation.do(t, j, plan)
				claimed, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				})
				var afterOutcomes int
				if mutation.fail {
					if !errors.Is(err, ErrStrategyDispatchLeaseUnavailable) {
						t.Fatalf("cardinality mismatch claim error=%v", err)
					}
					lease, loadErr := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
					if loadErr != nil || lease.State != StrategyDispatchLeaseIssued || lease.Disposition != StrategyDispatchReservationReserved || lease.Revision != 1 {
						t.Fatalf("cardinality rollback lease=%+v err=%v", lease, loadErr)
					}
					if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_outcomes WHERE lease_id=?`, plan.LeaseID).Scan(&afterOutcomes); err != nil || afterOutcomes != beforeOutcomes {
						t.Fatalf("outcomes before/after=%d/%d err=%v", beforeOutcomes, afterOutcomes, err)
					}
				} else {
					if err != nil || claimed.State != StrategyDispatchLeaseRefused || claimed.Disposition != StrategyDispatchReservationReleased {
						t.Fatalf("normalized refusal lease=%+v err=%v", claimed, err)
					}
					assertStrategyDispatchRealHolds(t, j, claimed, "RELEASED", 0)
				}
				assertNoStrategyDispatchTransport(t, j)
			})
		}
	}
}

func TestClaimStrategyDispatchLeaseRepeatsLiveAttemptStrategyAndClientJoinsKRUS(t *testing.T) {
	mutations := []struct {
		name string
		do   func(*testing.T, *Journal, StrategyDispatchLeasePlan)
	}{
		{"attempt-entry-decision", func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`INSERT INTO strategy_decision_lineage(
				entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,
				lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,
				policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,
				consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest)
				SELECT entry_decision_identity||':peer',candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,
				lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,
				policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,
				consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest
				FROM strategy_decision_lineage WHERE entry_decision_identity=(SELECT entry_decision_identity FROM strategy_first_leg_bindings WHERE decision_id=?)`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`DROP TRIGGER strategy_attempt_lineage_update_guard`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE strategy_attempt_lineage SET entry_decision_identity=entry_decision_identity||':peer' WHERE risk_intent_id=?`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{"activation-manifest", func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`DROP TRIGGER strategy_decision_lineage_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE strategy_decision_lineage SET activation_manifest_digest=activation_manifest_digest||':drift' WHERE entry_decision_identity=(SELECT entry_decision_identity FROM strategy_first_leg_bindings WHERE decision_id=?)`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{"client-order", func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE decisions SET client_order_id=client_order_id||'x' WHERE id=?`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		for _, mutation := range mutations {
			t.Run(market.name+"/"+mutation.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "live-join-"+strings.ToLower(market.name)+"-"+mutation.name)
				if err != nil {
					t.Fatal(err)
				}
				plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "live-join-"+strings.ToLower(market.name)+"-"+mutation.name,
					"acct-"+strings.ToLower(market.name), market.name, market.symbol)
				mutation.do(t, j, plan)
				lease, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				})
				if err != nil || lease.State != StrategyDispatchLeaseRefused || lease.RefusalCode != "FIRST_LEG_AUTHORITY_CHANGED" {
					t.Fatalf("live join refusal lease=%+v err=%v", lease, err)
				}
				assertStrategyDispatchRealHolds(t, j, lease, "RELEASED", 0)
				assertNoStrategyDispatchTransport(t, j)
			})
		}
	}
}

func TestClaimStrategyDispatchLeaseRejectsInvalidMissingAndReplayWithoutRevival(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	owner, err := j.AcquireStrategyDispatchOwner(ctx, "claim-replay-owner")
	if err != nil {
		t.Fatal(err)
	}
	for _, request := range []StrategyDispatchLeaseCAS{
		{},
		{LeaseID: "missing-lease", ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken},
	} {
		if _, err := j.ClaimStrategyDispatchLease(ctx, request); err == nil {
			t.Fatalf("invalid/missing claim %+v succeeded", request)
		}
	}
	stalePlan := seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-stale-revision", "acct-us", "US", "AAPL")
	stale, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
		LeaseID: stalePlan.LeaseID, ExpectedRevision: 2, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	})
	if err != nil || stale.State != StrategyDispatchLeaseRefused || stale.Disposition != StrategyDispatchReservationReleased || stale.RefusalCode != "LEASE_REVISION_STALE" {
		t.Fatalf("stale revision claim=%+v err=%v", stale, err)
	}
	assertStrategyDispatchRealHolds(t, j, stale, "RELEASED", 0)
	plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-replay", "acct", "KR", "005930")
	cas := StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken}
	first, err := j.ClaimStrategyDispatchLease(ctx, cas)
	if err != nil || first.State != StrategyDispatchLeaseClaimed {
		t.Fatalf("first claim=%+v err=%v", first, err)
	}
	if _, err := j.ClaimStrategyDispatchLease(ctx, cas); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
		t.Fatalf("replayed claim error=%v", err)
	}
	after, err := loadStrategyDispatchLease(ctx, j.db, plan.LeaseID)
	if err != nil || after.State != StrategyDispatchLeaseClaimed || after.Disposition != StrategyDispatchReservationReserved || after.Revision != 2 {
		t.Fatalf("replay revived/mutated lease=%+v err=%v", after, err)
	}
	assertStrategyDispatchRealHolds(t, j, after, "HELD", 5)
	assertNoStrategyDispatchTransport(t, j)
}

func TestClaimStrategyDispatchLeaseRollbackIsAtomic(t *testing.T) {
	for _, market := range []struct{ market, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.market, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			owner, err := j.AcquireStrategyDispatchOwner(ctx, "rollback-owner-"+strings.ToLower(market.market))
			if err != nil {
				t.Fatal(err)
			}
			plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "claim-rollback-"+strings.ToLower(market.market), "acct-"+strings.ToLower(market.market), market.market, market.symbol)
			if _, err := j.db.Exec(`CREATE TRIGGER fail_strategy_dispatch_claim_outcome BEFORE INSERT ON strategy_dispatch_outcomes BEGIN SELECT RAISE(ABORT,'synthetic claim outcome failure'); END`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}); err == nil {
				t.Fatal("late claim failure committed")
			}
			lease, err := loadStrategyDispatchLease(ctx, j.db, plan.LeaseID)
			if err != nil || lease.State != StrategyDispatchLeaseIssued || lease.Disposition != StrategyDispatchReservationReserved || lease.Revision != 1 {
				t.Fatalf("claim rollback lease=%+v err=%v", lease, err)
			}
			assertStrategyDispatchRealHolds(t, j, lease, "HELD", 5)
			assertNoStrategyDispatchTransport(t, j)
		})
	}
}

func seedBoundStrategyDispatchClaimLease(t *testing.T, j *Journal, owner StrategyDispatchOwner, suffix, account, market, symbol string) StrategyDispatchLeasePlan {
	t.Helper()
	request := firstLegAtomicFixture(t, j, suffix, account, market, symbol)
	receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	recordDigest := "claim-authority-" + suffix
	if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(
		authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
		protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
		risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
		VALUES(?,?,?,?,1,?,1,1,?,?,1,1,?,1,?,?,1,?,?)`, "claim-authority-id-"+suffix, account, market, symbol,
		"activation-"+suffix, "protection-serial-"+suffix, "protection-"+suffix, "risk-"+suffix,
		"guardian-"+suffix, "build-"+suffix, recordDigest, "2026-03-30T00:30:01Z"); err != nil {
		t.Fatal(err)
	}
	plan := StrategyDispatchLeasePlan{
		LeaseID: "claim-lease-" + suffix, OperationID: DeriveClientOrderID(request.Issue.Issue.Decision.ID, 0),
		AccountRef: account, Market: StrategyDispatchMarket(market), Symbol: symbol,
		CandidateID: request.Strategy.Lineage.CandidateLifeID, EvidenceDigest: request.Strategy.Lineage.EvidenceDigest,
		RouterID: request.RouterID, RouterVersion: request.RouterVersion, LaneID: request.Strategy.Lineage.LaneID,
		LaneVersion: request.Strategy.Lineage.LaneVersion, CampaignID: receipt.CampaignID,
		LegID: receipt.FirstLegPlanID, RiskReservationID: receipt.AggregateReservationID,
		GuardianDecisionID: receipt.DecisionID, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
		AuthorityRevision: 1, AuthorityDigest: recordDigest,
		IssuedAt: time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC),
	}
	if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err != nil {
		t.Fatal(err)
	}
	return plan
}

func currentStrategyDispatchOwnerForTest(j *Journal) (StrategyDispatchOwner, error) {
	var owner StrategyDispatchOwner
	err := j.db.QueryRow(`SELECT owner_instance,owner_epoch,fencing_token,revision,acquired_at FROM strategy_dispatch_owner_current WHERE owner_key='CENTRAL'`).Scan(
		&owner.OwnerInstance, &owner.Epoch, &owner.FencingToken, &owner.Revision, &owner.AcquiredAt)
	if err == nil {
		return owner, nil
	}
	// SQLite cannot scan RFC3339 text into time.Time through every driver path.
	var acquired string
	err = j.db.QueryRow(`SELECT owner_instance,owner_epoch,fencing_token,revision,acquired_at FROM strategy_dispatch_owner_current WHERE owner_key='CENTRAL'`).Scan(
		&owner.OwnerInstance, &owner.Epoch, &owner.FencingToken, &owner.Revision, &acquired)
	if err != nil {
		return StrategyDispatchOwner{}, err
	}
	owner.AcquiredAt, err = parseJournalTime(acquired)
	return owner, err
}

func assertStrategyDispatchRealHolds(t *testing.T, j *Journal, lease StrategyDispatchLease, aggregateState string, monetaryHeld int) {
	t.Helper()
	var aggregate string
	if err := j.db.QueryRow(`SELECT state FROM risk_reservations WHERE id=?`, lease.RiskReservationID).Scan(&aggregate); err != nil || aggregate != aggregateState {
		t.Fatalf("aggregate reservation state=%q want=%q err=%v", aggregate, aggregateState, err)
	}
	var held int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND existing_reservation_id=? AND state='HELD' AND held_minor=reserved_minor`,
		lease.GuardianDecisionID, lease.RiskReservationID).Scan(&held); err != nil || held != monetaryHeld {
		t.Fatalf("monetary held=%d want=%d err=%v", held, monetaryHeld, err)
	}
	if monetaryHeld == 0 {
		var released int
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND existing_reservation_id=? AND state='RELEASED' AND held_minor='0'`,
			lease.GuardianDecisionID, lease.RiskReservationID).Scan(&released); err != nil || released != 5 {
			t.Fatalf("monetary released=%d want=5 err=%v", released, err)
		}
	}
}

func assertNoStrategyDispatchTransport(t *testing.T, j *Journal) {
	t.Helper()
	var transport, broker int
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_leases WHERE transport_started_at IS NOT NULL`).Scan(&transport); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_leases WHERE broker_order_id<>''`).Scan(&broker); err != nil {
		t.Fatal(err)
	}
	if transport != 0 || broker != 0 {
		t.Fatalf("unexpected transport markers=%d broker ids=%d", transport, broker)
	}
	var invalid int
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_outcomes WHERE to_state NOT IN ('CLAIMED','REFUSED')`).Scan(&invalid); err != nil {
		t.Fatal(fmt.Errorf("outcome transport audit: %w", err))
	}
	if invalid != 0 {
		t.Fatalf("unexpected transport-like outcomes=%d", invalid)
	}
}
