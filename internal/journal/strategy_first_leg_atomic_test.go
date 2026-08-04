package journal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestFirstLegAtomicAdmissionDeliversKRAndUSTogether(t *testing.T) {
	j := openTestJournal(t)
	requests := []QFinalCampaignFirstLegRequest{
		firstLegAtomicFixture(t, j, "paired-kr", "acct-kr", "KR", "005930"),
		firstLegAtomicFixture(t, j, "paired-us", "acct-us", "US", "AAPL"),
	}
	results := make(chan struct {
		receipt QFinalCampaignFirstLegReceipt
		err     error
	}, len(requests))
	var wait sync.WaitGroup
	for _, request := range requests {
		request := request
		wait.Add(1)
		go func() {
			defer wait.Done()
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			results <- struct {
				receipt QFinalCampaignFirstLegReceipt
				err     error
			}{receipt: receipt, err: err}
		}()
	}
	wait.Wait()
	close(results)
	markets := map[string]bool{}
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.receipt.LegSequence != 1 || result.receipt.QFinal != 10 || result.receipt.Idempotent {
			t.Fatalf("receipt=%+v", result.receipt)
		}
		markets[result.receipt.Market] = true
	}
	if !markets["KR"] || !markets["US"] {
		t.Fatalf("paired markets=%v", markets)
	}
	for table, want := range map[string]int{
		"decisions": 2, "risk_reservations": 2, "strategy_decision_lineage": 2,
		"strategy_attempt_lineage": 2, "position_campaigns": 2, "position_campaign_claims": 2,
		"campaign_legs": 2, "risk_bucket_final_decisions": 2, "risk_bucket_owners": 2,
		"risk_bucket_reservations": 10, "strategy_first_leg_bindings": 2,
	} {
		if got := countRiskBucketRows(t, j, table); got != want {
			t.Fatalf("%s rows=%d want=%d", table, got, want)
		}
	}
	if got := countRiskBucketRows(t, j, "strategy_dispatch_leases"); got != 0 {
		t.Fatalf("first-leg admission created %d dispatch leases", got)
	}
}

func TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser(t *testing.T) {
	j := openTestJournal(t)
	type marketCase struct{ suffix, market, symbol string }
	cases := []marketCase{{"same-account-kr", "KR", "005930"}, {"same-account-us", "US", "AAPL"}}
	results := make(chan error, len(cases))
	var wait sync.WaitGroup
	for _, market := range cases {
		market := market
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := j.RecordQFinalCampaignFirstLegWithRecollection(context.Background(), func(context.Context, int) (QFinalCampaignFirstLegRequest, error) {
				return firstLegAtomicFixture(t, j, market.suffix, "acct-shared", market.market, market.symbol), nil
			}, RecollectPolicy{MaxAttempts: 3})
			results <- err
		}()
	}
	wait.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var paired int
	if err := j.db.QueryRow(`SELECT count(DISTINCT market) FROM strategy_first_leg_bindings WHERE account_ref='acct-shared'`).Scan(&paired); err != nil || paired != 2 {
		t.Fatalf("same-account paired markets=%d err=%v", paired, err)
	}
}

func TestFirstLegAtomicAdmissionCompetingSameScopeHasOneWinner(t *testing.T) {
	j := openTestJournal(t)
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, suffix := range []string{"scope-a", "scope-b"} {
		suffix := suffix
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := j.RecordQFinalCampaignFirstLegWithRecollection(context.Background(), func(context.Context, int) (QFinalCampaignFirstLegRequest, error) {
				return firstLegAtomicFixture(t, j, suffix, "acct", "KR", "005930"), nil
			}, RecollectPolicy{MaxAttempts: 3})
			results <- err
		}()
	}
	wait.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrRiskBucketOwnerConflict), errors.Is(err, ErrGenerationConflict):
			conflict++
		default:
			t.Fatalf("unexpected same-scope result=%v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	for table, want := range map[string]int{"risk_bucket_final_decisions": 1, "risk_bucket_owners": 1,
		"risk_bucket_reservations": 5, "position_campaigns": 1, "position_campaign_claims": 1,
		"campaign_legs": 1, "strategy_first_leg_bindings": 1} {
		if got := countRiskBucketRows(t, j, table); got != want {
			t.Fatalf("%s rows=%d want=%d", table, got, want)
		}
	}
}

func TestFirstLegAtomicAdmissionExactReplayUsesOriginalJournalToken(t *testing.T) {
	j := openTestJournal(t)
	request := firstLegAtomicFixture(t, j, "replay", "acct", "KR", "005930")
	first, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	var token string
	if err := j.db.QueryRow(`SELECT prospective_token FROM strategy_first_leg_bindings WHERE decision_id=?`, first.DecisionID).Scan(&token); err != nil {
		t.Fatal(err)
	}
	if len(token) != 64 {
		t.Fatalf("journal token length=%d", len(token))
	}
	replay, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil || !replay.Idempotent || replay.DecisionID != first.DecisionID || replay.CampaignID != first.CampaignID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	var replayToken string
	if err := j.db.QueryRow(`SELECT prospective_token FROM strategy_first_leg_bindings WHERE decision_id=?`, first.DecisionID).Scan(&replayToken); err != nil || replayToken != token {
		t.Fatalf("replay token changed=%q err=%v", replayToken, err)
	}
	request.Campaign.FirstLegPlanID = "divergent-plan"
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("divergent replay error=%v", err)
	}
}

func TestFirstLegAtomicAdmissionSurvivesReopenWithoutLegacyRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	before := openTestJournalAt(t, path)
	requests := []QFinalCampaignFirstLegRequest{
		firstLegAtomicFixture(t, before, "reopen-kr", "acct-kr", "KR", "005930"),
		firstLegAtomicFixture(t, before, "reopen-us", "acct-us", "US", "AAPL"),
	}
	receipts := make([]QFinalCampaignFirstLegReceipt, len(requests))
	for index, request := range requests {
		var err error
		receipts[index], err = before.RecordQFinalCampaignFirstLeg(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := before.Close(); err != nil {
		t.Fatal(err)
	}

	after := openTestJournalAt(t, path)
	for index, request := range requests {
		replayed, err := after.RecordQFinalCampaignFirstLeg(context.Background(), request)
		if err != nil || !replayed.Idempotent || replayed.DecisionID != receipts[index].DecisionID ||
			replayed.AttemptID != receipts[index].AttemptID {
			t.Fatalf("market=%s replay=%+v err=%v", request.Strategy.Lineage.Market, replayed, err)
		}
		pending, err := after.PendingStrategyPlans(context.Background(), request.Issue.Issue.Decision.AccountRef)
		if err != nil || len(pending) != 0 {
			t.Fatalf("market=%s first-leg attempt entered legacy recovery: pending=%+v err=%v",
				request.Strategy.Lineage.Market, pending, err)
		}
		var state string
		if err := after.db.QueryRow(`SELECT state FROM strategy_attempt_lineage WHERE attempt_id=?`, receipts[index].AttemptID).Scan(&state); err != nil || state != "PLANNED" {
			t.Fatalf("market=%s attempt state=%q err=%v", request.Strategy.Lineage.Market, state, err)
		}
	}
}

func TestFirstLegAtomicAdmissionDoesNotRepairPartialQFinalAuthority(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			request := firstLegAtomicFixture(t, j, "partial-"+strings.ToLower(market.name), "acct", market.name, market.symbol)
			partial := request.Issue
			partial.Admission.Owner.Key.ProspectiveGeneration = strings.Repeat(map[string]string{"KR": "a", "US": "b"}[market.name], 64)
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), partial); err != nil {
				t.Fatal(err)
			}
			if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); !errors.Is(err, ErrRiskBucketReplayMismatch) {
				t.Fatalf("partial q_final repair error=%v", err)
			}
			for _, table := range []string{"strategy_decision_lineage", "strategy_attempt_lineage", "position_campaigns", "campaign_legs", "strategy_first_leg_bindings"} {
				if got := countRiskBucketRows(t, j, table); got != 0 {
					t.Fatalf("%s repaired rows=%d", table, got)
				}
			}
		})
	}
}

func TestFirstLegAtomicAdmissionEntropyFailureWritesNothing(t *testing.T) {
	j := openTestJournal(t)
	j.firstLegEntropy = failingEntropy{}
	request := firstLegAtomicFixture(t, j, "entropy", "acct", "US", "AAPL")
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {
		t.Fatal("entropy failure admitted first leg")
	}
	for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "position_campaigns", "risk_bucket_final_decisions", "strategy_first_leg_bindings"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Fatalf("%s partial rows=%d", table, got)
		}
	}
}

func TestFirstLegAtomicAdmissionTokenCollisionRollsBackSecondMarket(t *testing.T) {
	j := openTestJournal(t)
	j.firstLegEntropy = bytes.NewReader(make([]byte, 64))
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), firstLegAtomicFixture(t, j, "token-kr", "acct-kr", "KR", "005930")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), firstLegAtomicFixture(t, j, "token-us", "acct-us", "US", "AAPL")); !errors.Is(err, ErrGenerationConflict) {
		t.Fatalf("token collision error=%v", err)
	}
	for table, want := range map[string]int{"decisions": 1, "risk_reservations": 1, "risk_bucket_final_decisions": 1,
		"strategy_decision_lineage": 1, "position_campaigns": 1, "strategy_first_leg_bindings": 1} {
		if got := countRiskBucketRows(t, j, table); got != want {
			t.Fatalf("%s rows=%d want=%d", table, got, want)
		}
	}
}

func TestFirstLegAtomicAdmissionLateStatementFailuresRollbackEveryFamily(t *testing.T) {
	failures := []struct {
		name, trigger string
	}{
		{"strategy-lineage", `CREATE TRIGGER fail_first_leg_strategy BEFORE INSERT ON strategy_decision_lineage BEGIN SELECT RAISE(ABORT,'synthetic strategy failure'); END`},
		{"campaign", `CREATE TRIGGER fail_first_leg_campaign BEFORE INSERT ON position_campaigns BEGIN SELECT RAISE(ABORT,'synthetic campaign failure'); END`},
		{"leg", `CREATE TRIGGER fail_first_leg_leg BEFORE INSERT ON campaign_legs BEGIN SELECT RAISE(ABORT,'synthetic leg failure'); END`},
		{"binding", `CREATE TRIGGER fail_first_leg_binding BEFORE INSERT ON strategy_first_leg_bindings BEGIN SELECT RAISE(ABORT,'synthetic binding failure'); END`},
	}
	for index, failure := range failures {
		t.Run(failure.name, func(t *testing.T) {
			j := openTestJournal(t)
			if _, err := j.db.Exec(failure.trigger); err != nil {
				t.Fatal(err)
			}
			market, symbol := "KR", "005930"
			if index%2 == 1 {
				market, symbol = "US", "AAPL"
			}
			request := firstLegAtomicFixture(t, j, "rollback-"+failure.name, "acct", market, symbol)
			if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {
				t.Fatal("synthetic late failure committed")
			}
			for _, table := range []string{
				"decisions", "risk_reservations", "risk_bucket_final_decisions", "risk_bucket_owners",
				"risk_bucket_reservations", "strategy_decision_lineage", "strategy_attempt_lineage",
				"position_campaigns", "position_campaign_claims", "campaign_legs", "strategy_first_leg_bindings",
			} {
				if got := countRiskBucketRows(t, j, table); got != 0 {
					t.Fatalf("%s partial rows=%d", table, got)
				}
			}
		})
	}
}

func TestFirstLegAtomicAdmissionRejectsCallerTokenAndCrossMarketSubstitution(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*QFinalCampaignFirstLegRequest)
	}{
		{"caller-token", func(request *QFinalCampaignFirstLegRequest) {
			request.Issue.Admission.Owner.Key.ProspectiveGeneration = strings.Repeat("a", 64)
		}},
		{"cross-market-owner", func(request *QFinalCampaignFirstLegRequest) {
			request.Issue.Admission.Owner.Key.Market = riskbucket.MarketUS
			request.Issue.Admission.Owner.Key.Symbol = "AAPL"
		}},
		{"missing-created-at", func(request *QFinalCampaignFirstLegRequest) {
			request.Strategy.CreatedAt = time.Time{}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			request := firstLegAtomicFixture(t, j, "refuse-"+tc.name, "acct", "KR", "005930")
			tc.mutate(&request)
			if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {
				t.Fatal("invalid first-leg request admitted")
			}
			for _, table := range []string{"decisions", "risk_reservations", "risk_bucket_final_decisions", "position_campaigns", "strategy_first_leg_bindings"} {
				if got := countRiskBucketRows(t, j, table); got != 0 {
					t.Fatalf("%s partial rows=%d", table, got)
				}
			}
		})
	}
}

func TestFirstLegBindingIsImmutableAndRequiredByFutureLease(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mutatePlan func(*StrategyDispatchLeasePlan)
		mutateDB   func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt)
		wantError  bool
	}{
		{"exact", func(*StrategyDispatchLeasePlan) {}, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, false},
		{"wrong-leg-plan", func(plan *StrategyDispatchLeasePlan) { plan.LegID = "other-plan" }, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, true},
		{"wrong-candidate", func(plan *StrategyDispatchLeasePlan) { plan.CandidateID = "other-candidate" }, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, true},
		{"wrong-operation", func(plan *StrategyDispatchLeasePlan) { plan.OperationID = "other-operation" }, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, true},
		{"wrong-router-id", func(plan *StrategyDispatchLeasePlan) { plan.RouterID = "other-router" }, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, true},
		{"wrong-router-version", func(plan *StrategyDispatchLeasePlan) { plan.RouterVersion = "other-router-version" }, func(*testing.T, *Journal, QFinalCampaignFirstLegReceipt) {}, true},
		{"campaign-evidence-drift", func(*StrategyDispatchLeasePlan) {}, func(t *testing.T, j *Journal, receipt QFinalCampaignFirstLegReceipt) {
			if _, err := j.db.Exec(`UPDATE position_campaigns SET evidence_digest='drifted' WHERE id=?`, receipt.CampaignID); err != nil {
				t.Fatal(err)
			}
		}, true},
		{"leg-residual-drift", func(*StrategyDispatchLeasePlan) {}, func(t *testing.T, j *Journal, receipt QFinalCampaignFirstLegReceipt) {
			if _, err := j.db.Exec(`UPDATE campaign_legs SET residual_quantity='9' WHERE campaign_id=? AND sequence=1`, receipt.CampaignID); err != nil {
				t.Fatal(err)
			}
		}, true},
		{"attempt-terminal-drift", func(*StrategyDispatchLeasePlan) {}, func(t *testing.T, j *Journal, receipt QFinalCampaignFirstLegReceipt) {
			if _, err := j.db.Exec(`UPDATE strategy_attempt_lineage SET state='REFUSED',revision=2 WHERE attempt_id=?`, receipt.AttemptID); err != nil {
				t.Fatal(err)
			}
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			request := firstLegAtomicFixture(t, j, "lease-"+tc.name, "acct", "US", "AAPL")
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE strategy_first_leg_bindings SET leg_plan_id='forged' WHERE decision_id=?`, receipt.DecisionID); err == nil {
				t.Fatal("immutable first-leg binding updated")
			}
			if _, err := j.db.Exec(`DELETE FROM strategy_first_leg_bindings WHERE decision_id=?`, receipt.DecisionID); err == nil {
				t.Fatal("immutable first-leg binding deleted")
			}
			owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "owner-"+tc.name)
			if err != nil {
				t.Fatal(err)
			}
			recordDigest := "dispatch-authority-" + tc.name
			if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(
				authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
				protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
				risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
				VALUES(?,?,?,?,1,?,1,1,?,?,1,1,?,1,?,?,1,?,?)`, "authority-"+tc.name, "acct", "US", "AAPL",
				"activation", "protection-serial", "protection", "risk", "guardian", "build", recordDigest,
				"2026-03-30T00:30:01Z"); err != nil {
				t.Fatal(err)
			}
			plan := StrategyDispatchLeasePlan{LeaseID: "lease-" + tc.name, OperationID: DeriveClientOrderID(request.Issue.Issue.Decision.ID, 0),
				AccountRef: "acct", Market: StrategyDispatchMarketUS, Symbol: "AAPL",
				CandidateID: request.Strategy.Lineage.CandidateLifeID, EvidenceDigest: request.Strategy.Lineage.EvidenceDigest,
				RouterID: request.RouterID, RouterVersion: request.RouterVersion, LaneID: request.Strategy.Lineage.LaneID,
				LaneVersion: request.Strategy.Lineage.LaneVersion, CampaignID: request.Campaign.CampaignID,
				LegID: request.Campaign.FirstLegPlanID, RiskReservationID: receipt.AggregateReservationID,
				GuardianDecisionID: receipt.DecisionID, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				AuthorityRevision: 1, AuthorityDigest: recordDigest,
				IssuedAt:  time.Date(2026, 3, 30, 0, 30, 5, 0, time.UTC),
				ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC)}
			tc.mutateDB(t, j, receipt)
			tc.mutatePlan(&plan)
			err = insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, "")
			if tc.wantError && (err == nil || !strings.Contains(err.Error(), "exact first-leg binding")) {
				t.Fatalf("mismatched lease error=%v", err)
			}
			if !tc.wantError && err != nil {
				t.Fatalf("exact bound lease: %v", err)
			}
		})
	}
}

func TestFirstLegRouterIdentityIsRequiredAndReplayBoundKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			request := firstLegAtomicFixture(t, j, "router-bound-"+strings.ToLower(market.name),
				"acct-"+strings.ToLower(market.name), market.name, market.symbol)
			for _, forged := range []struct {
				name string
				do   func(*QFinalCampaignFirstLegRequest)
			}{
				{"missing-router", func(value *QFinalCampaignFirstLegRequest) { value.RouterID = "" }},
				{"forged-router", func(value *QFinalCampaignFirstLegRequest) { value.RouterID = "different-router" }},
				{"forged-release", func(value *QFinalCampaignFirstLegRequest) { value.RouterVersion = "different-router-version" }},
			} {
				t.Run(forged.name, func(t *testing.T) {
					value := request
					forged.do(&value)
					if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), value); !errors.Is(err, ErrInvalidRequest) {
						t.Fatalf("fresh forged router error=%v", err)
					}
				})
			}
			if got := countRiskBucketRows(t, j, "strategy_first_leg_bindings"); got != 0 {
				t.Fatalf("fresh forged router created %d bindings", got)
			}
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			for _, mutate := range []struct {
				name string
				do   func(*QFinalCampaignFirstLegRequest)
			}{
				{"router-id", func(replay *QFinalCampaignFirstLegRequest) { replay.RouterID = "different-router" }},
				{"router-version", func(replay *QFinalCampaignFirstLegRequest) { replay.RouterVersion = "different-router-version" }},
			} {
				t.Run(mutate.name, func(t *testing.T) {
					replay := request
					mutate.do(&replay)
					if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), replay); !errors.Is(err, ErrInvalidRequest) {
						t.Fatalf("divergent router replay error=%v", err)
					}
				})
			}
			var routerID, routerVersion string
			if err := j.db.QueryRow(`SELECT router_id,router_version FROM strategy_first_leg_bindings WHERE decision_id=?`, receipt.DecisionID).
				Scan(&routerID, &routerVersion); err != nil || routerID != strategyrouter.RouterID || routerVersion != strategyrouter.RouterRelease {
				t.Fatalf("stored router=%q/%q err=%v", routerID, routerVersion, err)
			}
		})
	}
}

func TestFutureLeaseRejectsCancelledOrBlockedKRAndUSFirstLeg(t *testing.T) {
	markets := []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}}
	mutations := []string{"cancel", "entry-blocked", "claim-generation", "claim-deleted", "leg-cancelled", "attempt-refused"}
	for _, market := range markets {
		for _, mutation := range mutations {
			t.Run(market.name+"/"+mutation, func(t *testing.T) {
				j := openTestJournal(t)
				suffix := "lease-guard-" + strings.ToLower(market.name) + "-" + mutation
				request := firstLegAtomicFixture(t, j, suffix, "acct-"+strings.ToLower(market.name), market.name, market.symbol)
				receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
				if err != nil {
					t.Fatal(err)
				}
				plan := firstLegFutureLeasePlan(t, j, request, receipt, suffix)
				switch mutation {
				case "cancel":
					campaign, err := j.PositionCampaign(context.Background(), receipt.CampaignID)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := j.CancelProspectiveCampaign(context.Background(), CancelCampaignRequest{
						CampaignID: receipt.CampaignID, ExpectedVersion: campaign.Version,
						CommandKey: "cancel-" + suffix, Detail: "paired future-lease guard",
					}); err != nil {
						t.Fatal(err)
					}
				case "entry-blocked":
					_, err = j.db.Exec(`UPDATE position_campaigns SET entry_blocked=1 WHERE id=?`, receipt.CampaignID)
				case "claim-generation":
					_, err = j.db.Exec(`UPDATE position_campaign_claims SET position_generation=position_generation+1 WHERE campaign_id=?`, receipt.CampaignID)
				case "claim-deleted":
					_, err = j.db.Exec(`DELETE FROM position_campaign_claims WHERE campaign_id=?`, receipt.CampaignID)
				case "leg-cancelled":
					_, err = j.db.Exec(`UPDATE campaign_legs SET state='CANCELLED' WHERE campaign_id=? AND sequence=1`, receipt.CampaignID)
				case "attempt-refused":
					_, err = j.db.Exec(`UPDATE strategy_attempt_lineage SET state='REFUSED',revision=revision+1 WHERE attempt_id=?`, receipt.AttemptID)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil || !strings.Contains(err.Error(), "exact first-leg binding") {
					t.Fatalf("%s %s lease error=%v", market.name, mutation, err)
				}
				if got := countRiskBucketRows(t, j, "strategy_dispatch_leases"); got != 0 {
					t.Fatalf("%s %s created %d leases", market.name, mutation, got)
				}
			})
		}
	}
}

func TestFutureLeaseAcceptsExactKRAndUSFirstLegBinding(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			suffix := "exact-future-" + strings.ToLower(market.name)
			request := firstLegAtomicFixture(t, j, suffix, "acct-"+strings.ToLower(market.name), market.name, market.symbol)
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			plan := firstLegFutureLeasePlan(t, j, request, receipt, suffix)
			if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err != nil {
				t.Fatalf("%s exact first-leg lease: %v", market.name, err)
			}
		})
	}
}

func TestRawFirstLegBindingMismatchMatrixRejectsKRAndUS(t *testing.T) {
	type mismatchCase struct {
		name            string
		attemptMutation string
		postSeed        func(*testing.T, *Journal, preparedQFinalFirstLeg, string)
		binding         func(*rawFirstLegBinding)
	}
	cases := []mismatchCase{
		{name: "decision", binding: func(row *rawFirstLegBinding) { row.decisionID = "other-decision" }},
		{name: "aggregate-reservation", binding: func(row *rawFirstLegBinding) { row.aggregateReservationID = "other-reservation" }},
		{name: "attempt", binding: func(row *rawFirstLegBinding) { row.attemptID = "other-attempt" }},
		{name: "campaign", binding: func(row *rawFirstLegBinding) { row.campaignID = "other-campaign" }},
		{name: "claim", postSeed: func(t *testing.T, j *Journal, prepared preparedQFinalFirstLeg, _ string) {
			if _, err := j.db.Exec(`DELETE FROM position_campaign_claims WHERE campaign_id=?`, prepared.campaign.CampaignID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "owner", postSeed: func(t *testing.T, j *Journal, prepared preparedQFinalFirstLeg, token string) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET lane_id='other-lane' WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
				prepared.decision.AccountRef, strings.ToUpper(prepared.strategyPlan.Lineage.Market), prepared.campaignSymbol, token); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "leg-plan", binding: func(row *rawFirstLegBinding) { row.legPlanID = "other-leg-plan" }},
		{name: "prospective-token", binding: func(row *rawFirstLegBinding) { row.prospectiveToken = strings.Repeat("f", 64) }},
		{name: "activation-manifest", attemptMutation: "manifest"},
		{name: "client-order", attemptMutation: "client-order"},
	}
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		for _, mismatch := range cases {
			t.Run(market.name+"/"+mismatch.name, func(t *testing.T) {
				j := openTestJournal(t)
				suffix := "raw-" + strings.ToLower(market.name) + "-" + mismatch.name
				request := firstLegAtomicFixture(t, j, suffix, "acct-"+strings.ToLower(market.name), market.name, market.symbol)
				prepared, token := seedFirstLegUpstreamWithoutBinding(t, j, request, mismatch.attemptMutation)
				if mismatch.postSeed != nil {
					mismatch.postSeed(t, j, prepared, token)
				}
				row := rawFirstLegBindingFrom(prepared, token)
				if mismatch.binding != nil {
					mismatch.binding(&row)
				}
				if err := insertRawFirstLegBinding(j, row); err == nil || !strings.Contains(err.Error(), "exact atomic authority") {
					t.Fatalf("%s %s raw companion error=%v", market.name, mismatch.name, err)
				}
				if got := countRiskBucketRows(t, j, "strategy_first_leg_bindings"); got != 0 {
					t.Fatalf("%s %s inserted %d raw companions", market.name, mismatch.name, got)
				}
			})
		}
	}
}

type rawFirstLegBinding struct {
	decisionID, aggregateReservationID, entryDecisionIdentity, attemptID, campaignID            string
	legPlanID, accountRef, market, symbol, candidateID, evidenceDigest                          string
	laneID, laneVersion, routerID, routerVersion, prospectiveToken, requestDigest, recordDigest string
	qFinal                                                                                      uint64
	createdAt                                                                                   string
}

func rawFirstLegBindingFrom(prepared preparedQFinalFirstLeg, token string) rawFirstLegBinding {
	lineage := prepared.strategyPlan.Lineage
	return rawFirstLegBinding{decisionID: prepared.decision.ID, aggregateReservationID: prepared.reservePlan.rows[0].ID,
		entryDecisionIdentity: lineage.DecisionIdentity, attemptID: prepared.strategyPlan.AttemptID,
		campaignID: prepared.campaign.CampaignID, legPlanID: prepared.campaign.FirstLegPlanID,
		accountRef: prepared.decision.AccountRef, market: strings.ToUpper(lineage.Market), symbol: prepared.campaignSymbol,
		candidateID: lineage.CandidateLifeID, evidenceDigest: lineage.EvidenceDigest, laneID: lineage.LaneID,
		laneVersion: lineage.LaneVersion, routerID: prepared.routerID, routerVersion: prepared.routerVersion,
		prospectiveToken: token, qFinal: prepared.riskDecision.QFinal,
		requestDigest: prepared.firstLegDigest, recordDigest: prepared.bindingRecordHash,
		createdAt: "2026-03-30T00:30:00Z"}
}

func insertRawFirstLegBinding(j *Journal, row rawFirstLegBinding) error {
	_, err := j.db.Exec(`INSERT INTO strategy_first_leg_bindings(
		decision_id,aggregate_reservation_id,entry_decision_identity,attempt_id,campaign_id,leg_sequence,
		leg_plan_id,account_ref,market,symbol,candidate_id,evidence_digest,lane_id,lane_version,
		router_id,router_version,prospective_token,q_final,request_digest,record_digest,created_at)
		VALUES(?,?,?,?,?,1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.decisionID, row.aggregateReservationID,
		row.entryDecisionIdentity, row.attemptID, row.campaignID, row.legPlanID, row.accountRef, row.market,
		row.symbol, row.candidateID, row.evidenceDigest, row.laneID, row.laneVersion, row.routerID, row.routerVersion, row.prospectiveToken,
		row.qFinal, row.requestDigest, row.recordDigest, row.createdAt)
	return err
}

func seedFirstLegUpstreamWithoutBinding(t *testing.T, j *Journal, request QFinalCampaignFirstLegRequest, attemptMutation string) (preparedQFinalFirstLeg, string) {
	t.Helper()
	sum := sha256.Sum256([]byte(request.Campaign.CampaignID))
	token := hex.EncodeToString(sum[:])
	request.Issue.Admission.Owner.Key.ProspectiveGeneration = token
	prepared, err := j.prepareQFinalCampaignFirstLeg(request)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := reservePrecheck(context.Background(), tx, prepared.reserve, prepared.reservePlan, j.clk.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := insertDecisionRow(context.Background(), tx, prepared.decision); err != nil {
		t.Fatal(err)
	}
	reserved, err := reserveRows(context.Background(), tx, prepared.reserve, prepared.reservePlan, j.clk.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := commitFreshRiskBucketAdmissionTx(context.Background(), tx, request.Issue.Admission, prepared.riskDecision,
		prepared.issuePreimage, prepared.issueDigest, reserved.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := insertExactStrategyDecision(context.Background(), tx, prepared.strategyPlan.Lineage); err != nil {
		t.Fatal(err)
	}
	attempt := prepared.strategyPlan
	switch attemptMutation {
	case "manifest":
		attempt.ActivationManifestDigest = "other-activation-manifest"
	case "client-order":
		attempt.ClientOrderID = "other-client-order"
	}
	if _, err := insertExactStrategyAttempt(context.Background(), tx, attempt, prepared.decision.ID, prepared.decision.AccountRef); err != nil {
		t.Fatal(err)
	}
	if err := insertFirstLegCampaignTx(context.Background(), tx, prepared, token, j.nowString()); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return prepared, token
}

func firstLegFutureLeasePlan(t *testing.T, j *Journal, request QFinalCampaignFirstLegRequest, receipt QFinalCampaignFirstLegReceipt, suffix string) StrategyDispatchLeasePlan {
	t.Helper()
	account := request.Issue.Issue.Decision.AccountRef
	market := StrategyDispatchMarket(strings.ToUpper(request.Strategy.Lineage.Market))
	symbol := request.Strategy.Lineage.Symbol
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "owner-"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	recordDigest := "dispatch-authority-" + suffix
	if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(
		authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
		protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
		risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
		VALUES(?,?,?,?,1,?,1,1,?,?,1,1,?,1,?,?,1,?,?)`, "authority-"+suffix, account, string(market), symbol,
		"activation", "protection-serial", "protection", "risk", "guardian", "build", recordDigest,
		"2026-03-30T00:30:01Z"); err != nil {
		t.Fatal(err)
	}
	return StrategyDispatchLeasePlan{LeaseID: "future-lease-" + suffix, OperationID: DeriveClientOrderID(request.Issue.Issue.Decision.ID, 0),
		AccountRef: account, Market: market, Symbol: symbol,
		CandidateID: request.Strategy.Lineage.CandidateLifeID, EvidenceDigest: request.Strategy.Lineage.EvidenceDigest,
		RouterID: request.RouterID, RouterVersion: request.RouterVersion, LaneID: request.Strategy.Lineage.LaneID,
		LaneVersion: request.Strategy.Lineage.LaneVersion, CampaignID: request.Campaign.CampaignID,
		LegID: request.Campaign.FirstLegPlanID, RiskReservationID: receipt.AggregateReservationID,
		GuardianDecisionID: receipt.DecisionID, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
		AuthorityRevision: 1, AuthorityDigest: recordDigest,
		IssuedAt: time.Date(2026, 3, 30, 0, 30, 5, 0, time.UTC), ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC)}
}

type failingEntropy struct{}

func (failingEntropy) Read([]byte) (int, error) { return 0, errors.New("synthetic entropy failure") }

func firstLegAtomicFixture(t *testing.T, j *Journal, suffix, account, market, symbol string) QFinalCampaignFirstLegRequest {
	t.Helper()
	qFinal := qFinalIssueFixture(t, j, suffix)
	qFinal.Issue.Decision.AccountRef = account
	qFinal.Issue.Decision.ID = "qfinal-decision-" + suffix
	qFinal.Issue.Decision.Nonce = "nonce-" + suffix
	qFinal.Issue.Reserve.ObservedVersion = mustVersion(t, j, account)
	qFinal.Issue.Reserve.Reservations[0].ID = "qfinal-existing-" + suffix
	qFinal.Admission.DecisionID = qFinal.Issue.Decision.ID
	qFinal.Admission.ExistingReservationID = qFinal.Issue.Reserve.Reservations[0].ID
	qFinal.Admission.Owner.Key.AccountID = account
	qFinal.Admission.Owner.Key.ProspectiveGeneration = ""
	qFinal.Admission.Owner.LaneID = "lane-" + strings.ToLower(market)
	qFinal.Admission.Owner.CampaignID = "campaign-" + suffix
	qFinal.Admission.TransactionID = "risk-tx-" + suffix
	policyVersion, err := QFinalPolicyVersion("guardian-v1", qFinal.Admission.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	qFinal.Issue.Decision.Preimage = RiskIntent{AccountRef: account, Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: "5", StopPrice: "4", TargetPrice: "7", PolicyVersion: policyVersion}
	qFinal.Issue.Reserve.SnapshotUsage[0].Currency = "KRW"
	qFinal.Issue.Reserve.Limits[0].Currency = "KRW"
	qFinal.Issue.Reserve.Reservations[0].Currency = "KRW"
	qFinal.Admission.Admission.Policy.AccountCurrency = "KRW"
	qFinal.Admission.Admission.Policy.QuoteCurrency = map[string]string{"KR": "KRW", "US": "USD"}[market]
	if market == "US" {
		qFinal.Issue.Reserve.Limits[0].Amount = "1000000"
		qFinal.Issue.Reserve.Reservations[0].Amount = "73500"
		qFinal.Admission.Admission.Policy.FX = riskbucket.FXEvidence{RateQuoteToBase: "1400", Haircut: "1.05",
			Evidence: riskbucket.Evidence{Source: "official-fx", Version: "fx-v1", Digest: "fx-digest-US", Official: true, Frozen: true,
				ObservedAt: qFinal.Admission.CreatedAt.Add(-time.Minute), FreshUntil: qFinal.Admission.CreatedAt.Add(time.Minute)}}
	}
	qFinal.Admission.Owner.Key.Market = riskbucket.Market(market)
	qFinal.Admission.Owner.Key.Symbol = symbol
	for index, bucket := range qFinal.Admission.Admission.Buckets {
		if market == "US" {
			qFinal.Admission.Admission.Buckets[index].LimitMinor = "1000000"
			bucket = qFinal.Admission.Admission.Buckets[index]
		}
		key := bucket.Key
		key.PolicyVersion = "policy-v1-" + market
		switch key.Dimension {
		case riskbucket.DimensionMarket:
			key.Value = market
		case riskbucket.DimensionSymbol:
			key.Value = symbol
		}
		rebindRiskBucket(t, &qFinal.Admission, index, key)
	}

	record, _, _ := strategyDecisionRecordFixture(t, suffix)
	record.Market, record.Symbol = market, symbol
	record.LaneID, record.LaneVersion = qFinal.Admission.Owner.LaneID, "1"
	record.EvidenceDigest = "sha256:" + strings.Repeat(map[string]string{"KR": "c", "US": "d"}[market], 64)
	record.EntryPrice, record.StopPrice, record.TargetPrice = "5", "4", "7"
	record.Identity = ""
	identityPayload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	identityHash := sha256.Sum256(identityPayload)
	record.Identity = "strategy-decision:v1:sha256:" + hex.EncodeToString(identityHash[:])
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	payloadHash := sha256.Sum256(payload)
	lineage := StrategyDecisionLineage{
		DecisionIdentity: record.Identity, CandidateLifeID: record.CandidateLifeID, Market: market, Symbol: symbol,
		ThresholdVersion: record.ThresholdVersion, ThresholdSetDigest: record.ThresholdSetDigest,
		EvidenceDigest: record.EvidenceDigest, LaneID: record.LaneID, LaneVersion: record.LaneVersion,
		LaneSourceDigest: record.SourceDigest, LaneConstantsDigest: record.ConstantsDigest,
		EntryPrice: "5", StopPrice: "4", TargetPrice: "7", Quantity: "10", PolicyVersion: policyVersion,
		SettingsDigest: "sha256:" + strings.Repeat("9", 64), DecisionPayload: string(payload),
		DecisionPayloadDigest: "sha256:" + hex.EncodeToString(payloadHash[:]), ActivationManifestDigest: "sha256:" + strings.Repeat("f", 64),
		CreatedAt: qFinal.Admission.CreatedAt,
	}
	return QFinalCampaignFirstLegRequest{
		Issue:    qFinal,
		RouterID: strategyrouter.RouterID, RouterVersion: strategyrouter.RouterRelease,
		Strategy: StrategyPlanRequest{Lineage: lineage, AttemptID: "attempt-" + suffix,
			ActivationManifestDigest: lineage.ActivationManifestDigest, Revision: 1, CreatedAt: qFinal.Admission.CreatedAt},
		Campaign: FirstLegCampaignRequest{CampaignID: qFinal.Admission.Owner.CampaignID,
			ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
			CreateCommandKey: "create-" + suffix, FirstLegCommandKey: "leg-" + suffix,
			FirstLegPlanID: fmt.Sprintf("%s-first-leg", suffix)},
	}
}
