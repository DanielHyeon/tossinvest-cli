package journal

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestBeginStrategyDispatchSubmittingPairedKRUSCurrentAuthority(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-current-"+strings.ToLower(market.name), market.name, market.symbol)
			startedAt := time.Date(2026, 3, 30, 0, 30, 20, 0, time.UTC)
			j.clk.(*clock.Fake).Set(startedAt)

			lease, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if lease.State != StrategyDispatchLeaseSubmitting || lease.Disposition != StrategyDispatchReservationReserved ||
				lease.Revision != 3 || !lease.TransportStartedAt.Equal(startedAt) || lease.BrokerOrderID != "" {
				t.Fatalf("submitting lease=%+v", lease)
			}
			assertStrategyDispatchRealHolds(t, j, lease, "HELD", 5)
			assertStrategyDispatchTransition(t, j, lease.LeaseID, 2, 3, "CLAIMED", "SUBMITTING", "TRANSPORT_START_CURRENT")
			assertStrategyDispatchTransportCounts(t, j, lease.LeaseID, 1, 0)
			cas := StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken}
			if err := j.RequireCurrentStrategyDispatchTransportAuthority(context.Background(), cas); err != nil {
				t.Fatalf("current final transport authority error=%v", err)
			}
			if _, err := j.AcquireStrategyDispatchOwner(context.Background(), "racing-owner-"+strings.ToLower(market.name)); !errors.Is(err, ErrStrategyDispatchOwnerBusy) {
				t.Fatalf("active transport takeover error=%v", err)
			}
			if err := j.RequireCurrentStrategyDispatchTransportAuthority(context.Background(), cas); err != nil {
				t.Fatalf("failed takeover changed final authority: %v", err)
			}
			if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 3,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
				t.Fatalf("SUBMITTING replay error=%v", err)
			}
			replayed, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || replayed.State != StrategyDispatchLeaseSubmitting || replayed.Revision != 3 || !replayed.TransportStartedAt.Equal(startedAt) {
				t.Fatalf("SUBMITTING replay mutated lease=%+v err=%v", replayed, err)
			}
			j.clk.(*clock.Fake).Set(lease.ExpiresAt)
			if err := j.RequireCurrentStrategyDispatchTransportAuthority(context.Background(), cas); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("expired final transport authority error=%v", err)
			}
		})
	}
}

func TestBeginStrategyDispatchSubmittingPairedKRUSAuthorityDriftRefusesAndReleases(t *testing.T) {
	type driftCase struct {
		name string
		code string
		do   func(*testing.T, *Journal, StrategyDispatchLeasePlan)
	}
	cases := []driftCase{
		{name: "stale-revision", code: "LEASE_REVISION_STALE"},
		{name: "expired", code: "LEASE_EXPIRED", do: func(_ *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			j.clk.(*clock.Fake).Set(plan.ExpiresAt)
		}},
		{name: "market-authority-aba", code: "MARKET_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			for _, digest := range []string{"intermediate-" + plan.LeaseID, plan.AuthorityDigest} {
				if _, err := j.db.Exec(`UPDATE strategy_dispatch_market_authorities SET revision=revision+1,record_digest=? WHERE account_ref=? AND market=? AND symbol=?`,
					digest, plan.AccountRef, plan.Market, plan.Symbol); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{name: "cross-market-campaign-claim", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			peer := "US"
			if plan.Market == StrategyDispatchMarketUS {
				peer = "KR"
			}
			if _, err := j.db.Exec(`UPDATE position_campaign_claims SET market=? WHERE campaign_id=?`, peer, plan.CampaignID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "router-binding", code: "FIRST_LEG_BINDING_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`DROP TRIGGER strategy_first_leg_bindings_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE strategy_first_leg_bindings SET router_version=router_version||':drift' WHERE decision_id=?`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "qfinal", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`DROP TRIGGER risk_bucket_decisions_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE risk_bucket_final_decisions SET q_final=q_final-1 WHERE decision_id=?`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "attempt-client-order", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE decisions SET client_order_id=client_order_id||':drift' WHERE id=?`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "risk-owner", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET released_at=? WHERE account_ref=? AND market=? AND symbol=?`,
				"2026-03-30T00:30:15Z", plan.AccountRef, plan.Market, plan.Symbol); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "aggregate-hold", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at=?,release_reason=? WHERE id=?`,
				"2026-03-30T00:30:15Z", ReleaseReasonExpiredUnconsumed, plan.RiskReservationID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "five-bucket-hold", code: "FIRST_LEG_AUTHORITY_CHANGED", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',held_minor='0' WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		for _, tc := range cases {
			t.Run(market.name+"/"+tc.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-drift-"+strings.ToLower(market.name)+"-"+tc.name, market.name, market.symbol)
				if tc.do != nil {
					tc.do(t, j, plan)
				}
				expectedRevision := claimed.Revision
				if tc.name == "stale-revision" {
					expectedRevision--
				}
				lease, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: expectedRevision,
					OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				})
				if err != nil {
					t.Fatal(err)
				}
				if lease.State != StrategyDispatchLeaseRefused || lease.Disposition != StrategyDispatchReservationReleased ||
					lease.Revision != 3 || lease.RefusalCode != tc.code || !lease.TransportStartedAt.IsZero() {
					t.Fatalf("refused lease=%+v want code=%s", lease, tc.code)
				}
				assertStrategyDispatchRealHolds(t, j, lease, "RELEASED", 0)
				assertStrategyDispatchTransportCounts(t, j, lease.LeaseID, 0, 0)
				if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: 3,
					OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				}); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
					t.Fatalf("terminal replay error=%v", err)
				}
			})
		}
	}
}

func TestBeginStrategyDispatchSubmittingReleaseCardinalityMismatchRollsBackKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-release-proof-"+strings.ToLower(market.name), market.name, market.symbol)
			peer := "US"
			if plan.Market == StrategyDispatchMarketUS {
				peer = "KR"
			}
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET market=? WHERE decision_id=? AND bucket_dimension='symbol'`, peer, plan.GuardianDecisionID); err != nil {
				t.Fatal(err)
			}
			j.clk.(*clock.Fake).Set(plan.ExpiresAt)
			if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchLeaseUnavailable) {
				t.Fatalf("release proof error=%v", err)
			}
			after, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || after.State != StrategyDispatchLeaseClaimed || after.Revision != claimed.Revision || !after.TransportStartedAt.IsZero() {
				t.Fatalf("release proof rollback lease=%+v err=%v", after, err)
			}
			var aggregate string
			if err := j.db.QueryRow(`SELECT state FROM risk_reservations WHERE id=?`, plan.RiskReservationID).Scan(&aggregate); err != nil || aggregate != "HELD" {
				t.Fatalf("aggregate after rollback=%q err=%v", aggregate, err)
			}
			var outcomes int
			if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_outcomes WHERE lease_id=? AND next_revision=3`, plan.LeaseID).Scan(&outcomes); err != nil || outcomes != 0 {
				t.Fatalf("release proof outcomes=%d err=%v", outcomes, err)
			}
			assertStrategyDispatchTransportCounts(t, j, plan.LeaseID, 0, 0)
		})
	}
}

func TestBeginStrategyDispatchSubmittingPairedKRUSStaleOwnerAndRecoveryRefusal(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			oldOwner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-owner-"+strings.ToLower(market.name), market.name, market.symbol)
			newOwner, err := j.AcquireStrategyDispatchOwner(context.Background(), "replacement-submit-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: oldOwner.Epoch, FencingToken: oldOwner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("old owner error=%v", err)
			}
			unchanged, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || unchanged.State != StrategyDispatchLeaseClaimed || unchanged.Revision != claimed.Revision || !unchanged.TransportStartedAt.IsZero() {
				t.Fatalf("old owner mutated lease=%+v err=%v", unchanged, err)
			}
			assertStrategyDispatchRealHolds(t, j, unchanged, "HELD", 5)

			refused, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			})
			if err != nil || refused.State != StrategyDispatchLeaseRefused || refused.RefusalCode != "OWNER_STALE" || refused.Revision != 3 {
				t.Fatalf("recovery refusal lease=%+v err=%v", refused, err)
			}
			assertStrategyDispatchRealHolds(t, j, refused, "RELEASED", 0)
			assertStrategyDispatchTransportCounts(t, j, refused.LeaseID, 0, 0)
		})
	}
}

func TestBeginStrategyDispatchSubmittingRejectsInvalidMissingAndNonClaimedWithoutRevival(t *testing.T) {
	var nilJournal *Journal
	if _, err := nilJournal.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("nil journal error=%v", err)
	}
	j := openTestJournal(t)
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "submit-invalid-owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("invalid CAS error=%v", err)
	}
	if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
		LeaseID: "missing-submit-lease", ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	}); !errors.Is(err, ErrStrategyDispatchLeaseUnavailable) {
		t.Fatalf("missing lease error=%v", err)
	}
	plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "submit-issued", "acct-issued", "KR", "005930")
	if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
		LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	}); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
		t.Fatalf("ISSUED submit error=%v", err)
	}
	issued, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
	if err != nil || issued.State != StrategyDispatchLeaseIssued || issued.Revision != 1 || !issued.TransportStartedAt.IsZero() {
		t.Fatalf("ISSUED lease revived=%+v err=%v", issued, err)
	}
	assertStrategyDispatchRealHolds(t, j, issued, "HELD", 5)
}

func TestBeginStrategyDispatchSubmittingRollbackIsAtomicKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-rollback-"+strings.ToLower(market.name), market.name, market.symbol)
			if _, err := j.db.Exec(`CREATE TRIGGER fail_strategy_dispatch_submitting_outcome BEFORE INSERT ON strategy_dispatch_outcomes WHEN NEW.to_state='SUBMITTING' BEGIN SELECT RAISE(ABORT,'synthetic submitting outcome failure'); END`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.BeginStrategyDispatchSubmitting(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}); err == nil || !strings.Contains(err.Error(), "synthetic submitting outcome failure") {
				t.Fatalf("late outcome failure error=%v", err)
			}
			after, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || after.State != StrategyDispatchLeaseClaimed || after.Revision != claimed.Revision || !after.TransportStartedAt.IsZero() {
				t.Fatalf("rollback lease=%+v err=%v", after, err)
			}
			assertStrategyDispatchRealHolds(t, j, after, "HELD", 5)
			assertStrategyDispatchTransportCounts(t, j, after.LeaseID, 0, 0)
		})
	}
}

func TestBeginStrategyDispatchSubmittingConcurrentCASAtMostOneKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j, "submit-race-"+strings.ToLower(market.name), market.name, market.symbol)
			cas := StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken}
			type result struct {
				lease StrategyDispatchLease
				err   error
			}
			results := make(chan result, 2)
			var ready sync.WaitGroup
			ready.Add(2)
			start := make(chan struct{})
			for range 2 {
				go func() {
					ready.Done()
					<-start
					lease, err := j.BeginStrategyDispatchSubmitting(context.Background(), cas)
					results <- result{lease: lease, err: err}
				}()
			}
			ready.Wait()
			close(start)
			var successes int
			for range 2 {
				result := <-results
				if result.err == nil {
					successes++
					if result.lease.State != StrategyDispatchLeaseSubmitting {
						t.Fatalf("successful race lease=%+v", result.lease)
					}
					continue
				}
				if !errors.Is(result.err, ErrStrategyDispatchLeaseConsumed) {
					t.Fatalf("race loser error=%v", result.err)
				}
			}
			if successes != 1 {
				t.Fatalf("race successes=%d want=1", successes)
			}
			after, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || after.State != StrategyDispatchLeaseSubmitting || after.Revision != 3 {
				t.Fatalf("race terminal lease=%+v err=%v", after, err)
			}
			assertStrategyDispatchTransportCounts(t, j, after.LeaseID, 1, 0)
		})
	}
}

func seedClaimedStrategyDispatchLease(t *testing.T, j *Journal, suffix, market, symbol string) (StrategyDispatchOwner, StrategyDispatchLeasePlan, StrategyDispatchLease) {
	t.Helper()
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "owner-"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	plan := seedBoundStrategyDispatchClaimLease(t, j, owner, suffix, "acct-"+suffix, market, symbol)
	j.clk.(*clock.Fake).Set(time.Date(2026, 3, 30, 0, 30, 10, 0, time.UTC))
	claimed, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
		LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	})
	if err != nil || claimed.State != StrategyDispatchLeaseClaimed || claimed.Revision != 2 {
		t.Fatalf("claim lease=%+v err=%v", claimed, err)
	}
	return owner, plan, claimed
}

func assertStrategyDispatchTransition(t *testing.T, j *Journal, leaseID string, expectedRevision, nextRevision int, fromState, toState, code string) {
	t.Helper()
	var gotFrom, gotTo, gotCode string
	var gotExpected, gotNext int
	if err := j.db.QueryRow(`SELECT from_state,to_state,expected_revision,next_revision,transition_code FROM strategy_dispatch_outcomes WHERE lease_id=? AND next_revision=?`,
		leaseID, nextRevision).Scan(&gotFrom, &gotTo, &gotExpected, &gotNext, &gotCode); err != nil {
		t.Fatal(err)
	}
	if gotFrom != fromState || gotTo != toState || gotExpected != expectedRevision || gotNext != nextRevision || gotCode != code {
		t.Fatalf("transition=%s/%s rev=%d/%d code=%s", gotFrom, gotTo, gotExpected, gotNext, gotCode)
	}
}

func assertStrategyDispatchTransportCounts(t *testing.T, j *Journal, leaseID string, transport, broker int) {
	t.Helper()
	var gotTransport, gotBroker int
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_leases WHERE lease_id=? AND transport_started_at IS NOT NULL`, leaseID).Scan(&gotTransport); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_leases WHERE lease_id=? AND broker_order_id<>''`, leaseID).Scan(&gotBroker); err != nil {
		t.Fatal(err)
	}
	if gotTransport != transport || gotBroker != broker {
		t.Fatalf("transport/broker=%d/%d want=%d/%d", gotTransport, gotBroker, transport, broker)
	}
}
