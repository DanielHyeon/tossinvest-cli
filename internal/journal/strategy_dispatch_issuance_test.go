package journal

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestIssueVerifiedFirstLegStrategyDispatchLeasePairsKRUS(t *testing.T) {
	for _, tc := range []struct{ market, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(tc.market, func(t *testing.T) {
			j := openTestJournal(t)
			request := firstLegAtomicFixture(t, j, "issue-"+strings.ToLower(tc.market), "acct", tc.market, tc.symbol)
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "paired-owner")
			if err != nil {
				t.Fatal(err)
			}
			evidence := pairedStrategyDispatchEvidence(tc.market)
			lease, err := j.IssueVerifiedFirstLegStrategyDispatchLease(context.Background(), VerifiedFirstLegStrategyDispatchLeaseRequest{
				Receipt: receipt, Owner: owner, Evidence: evidence, TTL: 30 * time.Second,
			})
			if err != nil {
				t.Fatal(err)
			}
			if lease.Market != StrategyDispatchMarket(tc.market) || lease.Symbol != tc.symbol ||
				lease.State != StrategyDispatchLeaseIssued || lease.Disposition != StrategyDispatchReservationReserved ||
				lease.OwnerEpoch != owner.Epoch || lease.FencingToken != owner.FencingToken || lease.Revision != 1 {
				t.Fatalf("lease=%+v", lease)
			}
			replay, err := j.IssueVerifiedFirstLegStrategyDispatchLease(context.Background(), VerifiedFirstLegStrategyDispatchLeaseRequest{
				Receipt: receipt, Owner: owner, Evidence: evidence, TTL: 30 * time.Second,
			})
			if err != nil || replay.LeaseID != lease.LeaseID || replay.Revision != lease.Revision {
				t.Fatalf("replay=%+v err=%v", replay, err)
			}
			claimed, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: lease.LeaseID, ExpectedRevision: lease.Revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			})
			if err != nil || claimed.State != StrategyDispatchLeaseClaimed || claimed.Revision != 2 {
				t.Fatalf("claim=%+v err=%v", claimed, err)
			}
			assertNoStrategyDispatchTransport(t, j)
		})
	}
}

func TestIssueVerifiedFirstLegStrategyDispatchLeaseRefusesCrossMarketAndStaleOwner(t *testing.T) {
	j := openTestJournal(t)
	request := firstLegAtomicFixture(t, j, "issue-refusal", "acct", "KR", "005930")
	receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	first, err := j.AcquireStrategyDispatchOwner(context.Background(), "first-owner")
	if err != nil {
		t.Fatal(err)
	}
	current, err := j.AcquireStrategyDispatchOwner(context.Background(), "current-owner")
	if err != nil {
		t.Fatal(err)
	}
	for name, input := range map[string]VerifiedFirstLegStrategyDispatchLeaseRequest{
		"cross-market": {Receipt: receipt, Owner: current, Evidence: pairedStrategyDispatchEvidence("US"), TTL: 30 * time.Second},
		"stale-owner":  {Receipt: receipt, Owner: first, Evidence: pairedStrategyDispatchEvidence("KR"), TTL: 30 * time.Second},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := j.IssueVerifiedFirstLegStrategyDispatchLease(context.Background(), input); err == nil {
				t.Fatal("authority refusal expected")
			}
		})
	}
	var leases int
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_leases`).Scan(&leases); err != nil || leases != 0 {
		t.Fatalf("leases=%d err=%v", leases, err)
	}
}

func TestIssueVerifiedFirstLegStrategyDispatchLeaseRollsBackDerivedAuthorityWithLeaseInsert(t *testing.T) {
	j := openTestJournal(t)
	request := firstLegAtomicFixture(t, j, "issue-rollback", "acct", "US", "AAPL")
	receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "rollback-owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`CREATE TRIGGER fail_verified_lease_insert BEFORE INSERT ON strategy_dispatch_leases
		BEGIN SELECT RAISE(ABORT,'synthetic verified lease insert failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.IssueVerifiedFirstLegStrategyDispatchLease(context.Background(), VerifiedFirstLegStrategyDispatchLeaseRequest{
		Receipt: receipt, Owner: owner, Evidence: pairedStrategyDispatchEvidence("US"), TTL: 30 * time.Second,
	}); err == nil {
		t.Fatal("injected lease insert failure unexpectedly succeeded")
	}
	for _, table := range []string{"strategy_dispatch_market_authorities", "strategy_dispatch_leases"} {
		var rows int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&rows); err != nil || rows != 0 {
			t.Fatalf("rollback table=%s rows=%d err=%v", table, rows, err)
		}
	}
}

func pairedStrategyDispatchEvidence(market string) StrategyDispatchVerifiedEvidence {
	return StrategyDispatchVerifiedEvidence{Market: StrategyDispatchMarket(market),
		ActivationGeneration: 1, ActivationDigest: "sha256:" + strings.Repeat("f", 64),
		CalendarGeneration: 1, CalendarDigest: "sha256:" + strings.Repeat("1", 64),
		ProtectionGeneration: 1, ProtectionSerial: "1", ProtectionDigest: "sha256:" + strings.Repeat("2", 64),
		ReconciliationGeneration: 1, ReconciliationDigest: "sha256:" + strings.Repeat("3", 64),
		RiskPolicyGeneration: 1, GuardianGeneration: 1, BuildDigest: "sha256:" + strings.Repeat("4", 64)}
}
