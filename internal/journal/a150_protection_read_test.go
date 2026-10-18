//go:build tossos_testseams

package journal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA150ProtectionReadsFrozenAdmission(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
		t.Fatal(err)
	}
	got, err := j.StrategyProtection(ctx, req.Issue.Issue.Decision.ID, req.Campaign.CampaignID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Legacy || got.Policy == nil || got.Policy.Kind != ExitPolicyRatchet || got.Risk.StopPrice == "" {
		t.Fatalf("invalid frozen protection: %+v", got)
	}
	if _, err := j.StrategyProtection(ctx, req.Issue.Issue.Decision.ID, "other-campaign"); err == nil {
		t.Fatal("wrong campaign accepted")
	}
}

func TestA150CorruptMarkerOrMissingBindingCannotBecomeLegacy(t *testing.T) {
	for _, tc := range []struct{ name, corrupt string }{
		{"marker-zero-with-v4", `UPDATE decisions SET strategy_binding_version=0`},
		{"marker-four-without-binding", `DELETE FROM strategy_first_leg_bindings`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			j := openTestJournal(t)
			req := a150V4FirstLegFixture(t, j)
			if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
				t.Fatal(err)
			}
			if tc.name == "marker-zero-with-v4" {
				if _, err := j.db.Exec(`DROP TRIGGER decisions_strategy_binding_version_no_update`); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err := j.db.Exec(`DROP TRIGGER strategy_first_leg_bindings_no_delete`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := j.db.Exec(tc.corrupt); err != nil {
				t.Fatal(err)
			}
			if _, err := j.StrategyProtection(ctx, req.Issue.Issue.Decision.ID, req.Campaign.CampaignID); err == nil {
				t.Fatal("corrupt v4 admission treated as legacy")
			}
			if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err == nil {
				t.Fatal("corrupt v4 admission replay accepted")
			}
		})
	}
}

func TestA150ProtectionLatchIsScopedIdempotentAndBlocksEntry(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
		t.Fatal(err)
	}
	risk := req.Issue.Issue.Decision.Preimage.(RiskIntent)
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,1,?,'OPEN','1','100')`, "a150-pos", risk.AccountRef, risk.Market, risk.Symbol, req.Issue.Issue.Decision.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE position_campaigns SET actual_position_generation=1 WHERE id=?`, req.Campaign.CampaignID); err != nil {
		t.Fatal(err)
	}
	pos := Position{ID: "a150-pos", AccountRef: risk.AccountRef, Market: risk.Market, Symbol: risk.Symbol, InstanceSeq: 1, EntryDecisionID: req.Issue.Issue.Decision.ID}
	wrongAccount := pos
	wrongAccount.AccountRef = "different-account"
	if err := j.ReconcileStrategyProtection(ctx, req.Campaign.CampaignID, wrongAccount); err == nil {
		t.Fatal("different account accepted")
	}
	if err := j.ReconcileStrategyProtection(ctx, req.Campaign.CampaignID, pos); err != nil {
		t.Fatal(err)
	}
	if err := j.ReconcileStrategyProtection(ctx, req.Campaign.CampaignID, pos); err != nil {
		t.Fatal(err)
	}
	campaign, err := j.PositionCampaign(ctx, req.Campaign.CampaignID)
	if err != nil || !campaign.EntryBlocked || string(campaign.State) != "RECONCILE" {
		t.Fatalf("campaign=%+v err=%v", campaign, err)
	}
	var events int
	if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM campaign_events WHERE campaign_id=? AND event_kind='PROTECTION_RECONCILE'`, req.Campaign.CampaignID).Scan(&events); err != nil || events != 1 {
		t.Fatalf("events=%d err=%v", events, err)
	}
	pos.InstanceSeq++
	if err := j.ReconcileStrategyProtection(ctx, req.Campaign.CampaignID, pos); err == nil {
		t.Fatal("different generation accepted")
	}
	pos.InstanceSeq--
	if _, err := j.db.Exec(`UPDATE position_campaigns SET state='CLOSED' WHERE id=?`, req.Campaign.CampaignID); err != nil {
		t.Fatal(err)
	}
	if err := j.ReconcileStrategyProtection(ctx, req.Campaign.CampaignID, pos); err == nil {
		t.Fatal("closed campaign reactivated")
	}
	var state string
	if err := j.db.QueryRow(`SELECT state FROM position_campaigns WHERE id=?`, req.Campaign.CampaignID).Scan(&state); err != nil || state != "CLOSED" {
		t.Fatalf("state=%q err=%v", state, err)
	}
}

func TestA150PositionProtectionReadsAdmittedIdentity(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
		t.Fatal(err)
	}
	risk := req.Issue.Issue.Decision.Preimage.(RiskIntent)
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,1,?,'OPEN','1','100')`, "a150-pos", risk.AccountRef, risk.Market, risk.Symbol, req.Issue.Issue.Decision.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE position_campaigns SET actual_position_generation=1 WHERE id=?`, req.Campaign.CampaignID); err != nil {
		t.Fatal(err)
	}
	pos := Position{ID: "a150-pos", AccountRef: risk.AccountRef, Market: risk.Market, Symbol: risk.Symbol, InstanceSeq: 1, EntryDecisionID: req.Issue.Issue.Decision.ID}
	got, campaign, err := j.StrategyProtectionForPosition(ctx, pos)
	if err != nil || campaign != req.Campaign.CampaignID || got.Legacy || got.Policy == nil {
		t.Fatalf("got=%+v campaign=%q err=%v", got, campaign, err)
	}
	pos.InstanceSeq++
	if _, _, err := j.StrategyProtectionForPosition(ctx, pos); err == nil {
		t.Fatal("wrong position generation accepted")
	}
}

func TestA150ForeignPositionGenerationCannotUseAdmittedSnapshot(t *testing.T) {
	for _, campaignGeneration := range []int{1, 2} {
		t.Run(fmt.Sprint(campaignGeneration), func(t *testing.T) {
			ctx := context.Background()
			j := openTestJournal(t)
			req := a150V4FirstLegFixture(t, j)
			if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
				t.Fatal(err)
			}
			risk := req.Issue.Issue.Decision.Preimage.(RiskIntent)
			if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,2,?,'OPEN','1','100')`, "foreign-pos", risk.AccountRef, risk.Market, risk.Symbol, req.Issue.Issue.Decision.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`UPDATE position_campaigns SET actual_position_generation=? WHERE id=?`, campaignGeneration, req.Campaign.CampaignID); err != nil {
				t.Fatal(err)
			}
			pos := Position{ID: "foreign-pos", AccountRef: risk.AccountRef, Market: risk.Market, Symbol: risk.Symbol, InstanceSeq: 2, EntryDecisionID: req.Issue.Issue.Decision.ID}
			got, _, err := j.StrategyProtectionForPosition(ctx, pos)
			if err == nil || got.Policy != nil {
				t.Fatalf("foreign generation received admitted policy: %+v err=%v", got, err)
			}
		})
	}
}

func TestA150NewAdmissionRequiresMatchingProjectionGeneration(t *testing.T) {
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	req.Campaign.ExpectedPositionGeneration = 1
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), req); err == nil || !strings.Contains(err.Error(), "admitted generation mismatch") {
		t.Fatalf("generation 1 projection for expected successor 2: %v", err)
	}
	for _, table := range []string{"decisions", "risk_reservations", "position_campaigns", "strategy_first_leg_bindings"} {
		if n := countRiskBucketRows(t, j, table); n != 0 {
			t.Fatalf("%s contains %d rows", table, n)
		}
	}
}

func TestA150FinalProtectionChecksExistingExit(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
		t.Fatal(err)
	}
	risk := req.Issue.Issue.Decision.Preimage.(RiskIntent)
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,1,?,'OPEN','1','100')`, "a150-pos", risk.AccountRef, risk.Market, risk.Symbol, req.Issue.Issue.Decision.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE position_campaigns SET actual_position_generation=1 WHERE id=?`, req.Campaign.CampaignID); err != nil {
		t.Fatal(err)
	}
	check := func() error {
		tx, err := j.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		protection, err := strategyProtectionInTx(ctx, tx, req.Issue.Issue.Decision.ID, req.Campaign.CampaignID)
		if err != nil {
			return err
		}
		return requireStrategyProtectionPositionInTx(ctx, tx, req.Campaign.CampaignID, protection)
	}
	if err := check(); err == nil {
		t.Fatal("missing exit accepted")
	}
	if _, err := j.OpenExitState(ctx, ExitStateSeed{PositionID: "a150-pos", PolicyKind: ExitPolicyRatchet, EntryPrice: risk.EntryPrice, InitialStop: risk.StopPrice}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatalf("admitted exit rejected: %v", err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE exit_states SET initial_stop='1' WHERE position_id='a150-pos'`); err != nil {
		t.Fatal(err)
	}
	if err := check(); err == nil {
		t.Fatal("changed stop accepted")
	}
}

func TestA150FinalTransportRefusesCommittedProtectionLatch(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	req := a150V4FirstLegFixture(t, j)
	receipt, err := j.RecordQFinalCampaignFirstLeg(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := j.AcquireStrategyDispatchOwner(ctx, "a150-owner")
	if err != nil {
		t.Fatal(err)
	}
	risk := req.Issue.Issue.Decision.Preimage.(RiskIntent)
	if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at) VALUES(?,?,?,?,1,'activation',1,1,'serial','protect',1,1,'risk',1,'guardian','build',1,'authority','2026-03-30T00:30:01Z')`, "a150-authority", risk.AccountRef, risk.Market, risk.Symbol); err != nil {
		t.Fatal(err)
	}
	plan := StrategyDispatchLeasePlan{LeaseID: "a150-lease", OperationID: DeriveClientOrderID(receipt.DecisionID, 0), AccountRef: risk.AccountRef, Market: StrategyDispatchMarket(risk.Market), Symbol: risk.Symbol, CandidateID: req.Strategy.Lineage.CandidateLifeID, EvidenceDigest: req.Strategy.Lineage.EvidenceDigest, RouterID: req.RouterID, RouterVersion: req.RouterVersion, LaneID: req.Strategy.Lineage.LaneID, LaneVersion: req.Strategy.Lineage.LaneVersion, CampaignID: receipt.CampaignID, LegID: receipt.FirstLegPlanID, RiskReservationID: receipt.AggregateReservationID, GuardianDecisionID: receipt.DecisionID, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken, AuthorityRevision: 1, AuthorityDigest: "authority", IssuedAt: time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC)}
	if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err != nil {
		t.Fatal(err)
	}
	j.clk.(*clock.Fake).Set(time.Date(2026, 3, 30, 0, 30, 10, 0, time.UTC))
	claimed, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken})
	if err != nil {
		t.Fatal(err)
	}
	cas := StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken}
	if _, err := j.BeginStrategyDispatchSubmitting(ctx, cas); err != nil {
		t.Fatal(err)
	}
	if err := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas); err != nil {
		t.Fatalf("authority before latch: %v", err)
	}
	// Two dispatches paused at explicit boundaries model a fake transport. The
	// first has already passed final authority; the second has not checked yet.
	beforeCheck := make(chan struct{})
	beforeCall := make(chan struct{})
	authorized := make(chan error, 1)
	preResult := make(chan error, 1)
	postResult := make(chan error, 1)
	transport := make(chan string, 2)
	go func() {
		<-beforeCheck
		err := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas)
		if err == nil {
			transport <- "pre"
		}
		preResult <- err
	}()
	go func() {
		err := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas)
		authorized <- err
		if err == nil {
			<-beforeCall
			transport <- "post"
		}
		postResult <- err
	}()
	if err := <-authorized; err != nil {
		t.Fatalf("pre-latch authority: %v", err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,1,?,'OPEN','1','100')`, "a150-pos", risk.AccountRef, risk.Market, risk.Symbol, receipt.DecisionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE position_campaigns SET actual_position_generation=1 WHERE id=?`, receipt.CampaignID); err != nil {
		t.Fatal(err)
	}
	pos := Position{ID: "a150-pos", AccountRef: risk.AccountRef, Market: risk.Market, Symbol: risk.Symbol, InstanceSeq: 1, EntryDecisionID: receipt.DecisionID}
	if _, err := j.db.Exec(`CREATE TRIGGER a150_latch_failure BEFORE UPDATE ON position_campaigns BEGIN SELECT RAISE(ABORT,'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := j.ReconcileStrategyProtection(ctx, receipt.CampaignID, pos); err == nil {
		t.Fatal("injected latch failure did not fail")
	}
	if err := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas); err == nil {
		t.Fatal("authority accepted durable position without admitted exit after latch failure")
	}
	if _, err := j.db.Exec(`DROP TRIGGER a150_latch_failure`); err != nil {
		t.Fatal(err)
	}
	other, err := Open(ctx, Options{Path: j.Path(), Clock: j.clk, FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if err := other.ReconcileStrategyProtection(ctx, receipt.CampaignID, pos); err != nil {
		t.Fatal(err)
	}
	close(beforeCheck)
	close(beforeCall)
	if err := <-preResult; err == nil {
		t.Fatal("new transport authorized after latch")
	}
	if err := <-postResult; err != nil {
		t.Fatalf("previous authorization lost: %v", err)
	}
	if got := <-transport; got != "post" {
		t.Fatalf("unexpected fake transport: %q", got)
	}
	select {
	case got := <-transport:
		t.Fatalf("extra fake transport: %q", got)
	default:
	}
	if err := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas); err == nil {
		t.Fatal("authority allowed after latch")
	}
	lease, err := loadStrategyDispatchLease(ctx, j.db, plan.LeaseID)
	if err != nil || lease.State != StrategyDispatchLeaseSubmitting || lease.Disposition != StrategyDispatchReservationReserved {
		t.Fatalf("in-flight lease=%+v err=%v", lease, err)
	}
}

func TestA150FinalProtectionRejectsPrefillGenerationDrift(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*testing.T, *Journal, QFinalCampaignFirstLegRequest)
	}{
		{"closed-successor", func(t *testing.T, j *Journal, r QFinalCampaignFirstLegRequest) {
			risk := r.Issue.Issue.Decision.Preimage.(RiskIntent)
			if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price) VALUES(?,?,?,?,1,'CLOSED','0','')`, "successor", risk.AccountRef, risk.Market, risk.Symbol); err != nil {
				t.Fatal(err)
			}
		}},
		{"expected-row-missing", func(t *testing.T, j *Journal, r QFinalCampaignFirstLegRequest) {
			if _, err := j.db.Exec(`UPDATE position_campaigns SET expected_position_generation=1,expected_position_version=1 WHERE id=?`, r.Campaign.CampaignID); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			j := openTestJournal(t)
			req := a150V4FirstLegFixture(t, j)
			if _, err := j.RecordQFinalCampaignFirstLeg(ctx, req); err != nil {
				t.Fatal(err)
			}
			tc.mutate(t, j, req)
			tx, err := j.db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			p, err := strategyProtectionInTx(ctx, tx, req.Issue.Issue.Decision.ID, req.Campaign.CampaignID)
			if err != nil {
				return // A corrupt expected generation was already rejected by admission verification.
			}
			if err := requireStrategyProtectionPositionInTx(ctx, tx, req.Campaign.CampaignID, p); err == nil {
				t.Fatal("prefill generation drift accepted")
			}
		})
	}
}
