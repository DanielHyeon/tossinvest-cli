package journal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
)

func TestCreatePositionCampaignRejectsMissingDecisionAndOpenPosition(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	request := CreatePositionCampaignRequest{
		ID: "missing-decision", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "lane", LaneVersion: "v1", DecisionID: "does-not-exist", EvidenceDigest: "digest",
		ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
		ProspectiveToken: "missing-decision-token", CommandKey: "create",
	}
	if _, err := j.CreatePositionCampaign(ctx, request); !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
		t.Fatalf("missing decision err=%v, want invalid identity", err)
	}

	insertCampaignDecision(t, j, "decision-open", "acct")
	insertCampaignStrategyLineage(t, j, "decision-open", "acct", "kr", "005930", "lane", "v1", "digest")
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('manual-open','acct','kr','005930',1,'OPEN','1','70000','2026-03-30T00:20:00Z')`); err != nil {
		t.Fatal(err)
	}
	var version int64
	if err := j.db.QueryRow(`SELECT version FROM position_projection_versions WHERE position_id='manual-open'`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	request.ID, request.DecisionID, request.ProspectiveToken = "open-position", "decision-open", "open-position-token"
	request.ExpectedPositionGeneration, request.ExpectedPositionVersion = 1, version
	if _, err := j.CreatePositionCampaign(ctx, request); !errors.Is(err, ErrGenerationConflict) {
		t.Fatalf("OPEN position campaign err=%v, want generation conflict", err)
	}
}

func TestCreatePositionCampaignUsesAuthoritativeClosedPositionVersion(t *testing.T) {
	j := openTestJournal(t)
	insertCampaignDecision(t, j, "decision-closed", "acct")
	insertCampaignStrategyLineage(t, j, "decision-closed", "acct", "kr", "005930", "lane", "v1", "digest")
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at,closed_at)
		VALUES ('closed','acct','kr','005930',1,'CLOSED','0','70000',
		        '2026-03-30T00:00:00Z','2026-03-30T00:20:00Z')`); err != nil {
		t.Fatal(err)
	}
	var version int64
	if err := j.db.QueryRow(`SELECT version FROM position_projection_versions WHERE position_id='closed'`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	req := CreatePositionCampaignRequest{
		ID: "after-closed", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "lane", LaneVersion: "v1", DecisionID: "decision-closed", EvidenceDigest: "digest",
		ExpectedPositionGeneration: 1, ExpectedPositionVersion: version,
		ProspectiveToken: "after-closed-token", CommandKey: "create",
	}
	if _, err := j.CreatePositionCampaign(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}

func TestLinkCampaignOrderUsesAuthoritativeAttemptAndUniqueOrderScope(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg1, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-1",
		Sequence: 1, PlanID: "plan-1", RequestedQuantity: "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg1.CampaignVersion,
		CommandKey: "unbacked", OrderID: "shared", RequestedCap: "1",
		IntentID: "unbacked-intent", AttemptID: "unbacked-attempt",
	})
	if !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
		t.Fatalf("unbacked lineage err=%v, want invalid identity", err)
	}
	refused, _ := j.PositionCampaign(context.Background(), campaign.ID)
	replay, replayErr := j.ReconstructPositionCampaign(context.Background(), campaign.ID)
	if replayErr != nil || !replay.Valid {
		t.Fatalf("refusal replay=%+v err=%v", replay, replayErr)
	}
	_, retryErr := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg1.CampaignVersion,
		CommandKey: "unbacked", OrderID: "shared", RequestedCap: "1",
		IntentID: "unbacked-intent", AttemptID: "unbacked-attempt",
	})
	retried, _ := j.PositionCampaign(context.Background(), campaign.ID)
	if !errors.Is(retryErr, positioncampaign.ErrInvalidIdentity) || retried.Version != refused.Version {
		t.Fatalf("refusal retry err=%v version %d->%d", retryErr, refused.Version, retried.Version)
	}

	j = openTestJournal(t)
	campaign = createCampaignFixture(t, j)
	leg1, err = j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-1",
		Sequence: 1, PlanID: "plan-1", RequestedQuantity: "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-1", "attempt-1", "shared", "", "1")
	first, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg1.CampaignVersion,
		CommandKey: "link-1", OrderID: "shared", RequestedCap: "1",
		IntentID: "intent-1", AttemptID: "attempt-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	leg2, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: first.CampaignVersion, CommandKey: "plan-2",
		Sequence: 2, PlanID: "plan-2", RequestedQuantity: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-2", "attempt-2", "shared", "", "1")
	_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 2, ExpectedVersion: leg2.CampaignVersion,
		CommandKey: "link-2", OrderID: "shared", RequestedCap: "1",
		IntentID: "intent-2", AttemptID: "attempt-2",
	})
	if err == nil {
		t.Fatal("same scoped immutable broker order was linked to two legs")
	}
	got, _ := j.PositionCampaign(context.Background(), campaign.ID)
	if got.State != positioncampaign.CampaignReconcile || !got.EntryBlocked {
		t.Fatalf("duplicate order did not latch campaign: %+v", got)
	}
}

func TestReplacementPredecessorHasOneAuthoritativeSuccessor(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "3",
	})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt-a", "a", "", "3")
	a, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link-a", OrderID: "a", RequestedCap: "3", IntentID: "intent", AttemptID: "attempt-a",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt-b", "b", "a", "3")
	b, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: a.CampaignVersion,
		CommandKey: "link-b", OrderID: "b", PredecessorOrderID: "a", RequestedCap: "3",
		IntentID: "intent", AttemptID: "attempt-b",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt-c", "c", "a", "3")
	_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: b.CampaignVersion,
		CommandKey: "link-c", OrderID: "c", PredecessorOrderID: "a", RequestedCap: "3",
		IntentID: "intent", AttemptID: "attempt-c", LineageAmbiguous: false,
	})
	if err == nil {
		t.Fatal("one predecessor acquired two successors")
	}
	got, _ := j.PositionCampaign(context.Background(), campaign.ID)
	if got.State != positioncampaign.CampaignReconcile || !got.EntryBlocked {
		t.Fatalf("ambiguous successor did not latch reconcile: %+v", got)
	}
}

func TestAmbiguousCampaignFillPreservesAuthoritativeFillTransaction(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg1, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-1",
		Sequence: 1, PlanID: "plan-1", RequestedQuantity: "2",
	})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "2")
	linked, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg1.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "2", IntentID: "intent", AttemptID: "attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	leg2, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: linked.CampaignVersion, CommandKey: "plan-2",
		Sequence: 2, PlanID: "plan-2", RequestedQuantity: "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DROP INDEX idx_campaign_order_scope_identity`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO campaign_order_watermarks
		(campaign_id,leg_sequence,order_id,account_ref,market,trading_day,symbol,side,decision_id,
		 intent_id,attempt_id,carry_baseline,requested_cap,cumulative_filled,remaining_quantity,
		 terminal,lineage_ambiguous,created_at,updated_at)
		VALUES (?,2,'order','acct','kr','2026-03-30','005930','BUY',?,?,?,'0','2','0','2',0,0,
		        '2026-03-30T00:20:00Z','2026-03-30T00:20:00Z')`, campaign.ID, campaign.DecisionID, "intent", "attempt"); err != nil {
		t.Fatal(err)
	}
	_ = leg2
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	handle := &ApplyTx{tx: tx, now: "2026-03-30T00:30:00Z"}
	err = ApplyPositionCampaignFill(context.Background(), handle, AppliedFill{
		OrderID: "order", AccountRef: "acct", Market: "kr", TradingDay: "2026-03-30", Symbol: "005930", Side: "BUY",
		Delta: "1", CumulativeQuantity: "1", CommittedAt: "2026-03-30T00:30:00Z",
	})
	handle.invalidate()
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("campaign ambiguity rejected authoritative fill: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	got, _ := j.PositionCampaign(context.Background(), campaign.ID)
	if got.State != positioncampaign.CampaignReconcile || !got.EntryBlocked {
		t.Fatalf("ambiguous campaign was not isolated: %+v", got)
	}
	replay, err := j.ReconstructPositionCampaign(context.Background(), campaign.ID)
	if err != nil || !replay.Valid {
		t.Fatalf("ambiguous fill replay=%+v err=%v", replay, err)
	}
	var evidenceEvents int
	if err := j.db.QueryRow(`SELECT count(*) FROM campaign_events WHERE campaign_id=? AND event_kind='AMBIGUOUS_ORDER_FILL'`, campaign.ID).Scan(&evidenceEvents); err != nil || evidenceEvents != 1 {
		t.Fatalf("ambiguous evidence events=%d err=%v", evidenceEvents, err)
	}
	beforeRetry := got.Version
	tx, err = j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	handle = &ApplyTx{tx: tx, now: "2026-03-30T00:30:00Z"}
	err = ApplyPositionCampaignFill(context.Background(), handle, AppliedFill{
		OrderID: "order", AccountRef: "acct", Market: "kr", TradingDay: "2026-03-30", Symbol: "005930", Side: "BUY",
		Delta: "1", CumulativeQuantity: "1", CommittedAt: "2026-03-30T00:30:00Z",
	})
	handle.invalidate()
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	afterRetry, _ := j.PositionCampaign(context.Background(), campaign.ID)
	if afterRetry.Version != beforeRetry {
		t.Fatalf("ambiguous retry moved version %d->%d", beforeRetry, afterRetry.Version)
	}
}

func TestClosedCampaignLateFillStaysClosedAndLatchesReconcile(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "2",
	})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "late", "", "2")
	_, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "late", RequestedCap: "2", IntentID: "intent", AttemptID: "attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE position_campaigns SET state='CLOSED',entry_blocked=1 WHERE id=?`, campaign.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('late-position','acct','kr','005930',1,'OPEN','1','70000','2026-03-30T00:30:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "late", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "1", CumulativeQuantity: "1", CommittedAt: "2026-03-30T00:31:00Z"})
	got, _ := j.PositionCampaign(context.Background(), campaign.ID)
	watermark, _ := j.CampaignOrderWatermark(context.Background(), campaign.ID, 1, "late")
	if got.State != positioncampaign.CampaignClosed || !got.EntryBlocked || watermark.CumulativeFilled != "1" {
		t.Fatalf("closed late fill projection: campaign=%+v watermark=%+v", got, watermark)
	}
	var active int
	if err := j.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE account_ref='acct' AND released_at IS NULL`).Scan(&active); err != nil || active == 0 {
		t.Fatalf("late fill reconcile latch count=%d err=%v", active, err)
	}
}

func TestExposureAdmissionReadsPositionAndPendingRiskReduction(t *testing.T) {
	t.Run("position closing", func(t *testing.T) {
		j := openTestJournal(t)
		campaign := createCampaignFixture(t, j)
		if _, err := j.db.Exec(`INSERT INTO positions
			(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
			VALUES ('closing','acct','kr','005930',1,'CLOSING','1','70000','2026-03-30T00:20:00Z')`); err != nil {
			t.Fatal(err)
		}
		_, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
			CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
			Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
		})
		if !errors.Is(err, positioncampaign.ErrExposureBlocked) {
			t.Fatalf("CLOSING admission err=%v", err)
		}
	})
	t.Run("pending sell intent", func(t *testing.T) {
		j := openTestJournal(t)
		campaign := createCampaignFixture(t, j)
		if _, err := j.db.Exec(`INSERT INTO intents
			(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,
			 quantity,price,currency,source,fingerprint,notes)
			VALUES ('sell-pending','2026-03-30T00:20:00Z','kr','2026-03-30','acct','005930',
			        'SELL','MARKET','','1',NULL,'KRW','exit','sell-pending-fp','')`); err != nil {
			t.Fatal(err)
		}
		_, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
			CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
			Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
		})
		if !errors.Is(err, positioncampaign.ErrExposureBlocked) {
			t.Fatalf("pending SELL admission err=%v", err)
		}
	})
}

func TestZeroFillTerminalClosesCampaignAndReleasesClaim(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
	})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "cancelled", "", "1")
	_, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "cancelled", RequestedCap: "1", IntentID: "intent", AttemptID: "attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "cancelled", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "0", CumulativeQuantity: "0", Terminal: true, State: "CANCELLED", CommittedAt: "2026-03-30T00:30:00Z"})
	got, _ := j.PositionCampaign(context.Background(), campaign.ID)
	var claims int
	_ = j.db.QueryRow(`SELECT count(*) FROM position_campaign_claims WHERE campaign_id=?`, campaign.ID).Scan(&claims)
	if got.State != positioncampaign.CampaignClosed || claims != 0 {
		t.Fatalf("zero-fill terminal campaign=%+v claims=%d", got, claims)
	}
}

func TestCampaignReplayDetectsCompleteProjectionDrift(t *testing.T) {
	for _, tc := range []struct{ name, statement string }{
		{name: "entry blocked", statement: `UPDATE position_campaigns SET entry_blocked=1 WHERE id='campaign'`},
		{name: "leg residual", statement: `UPDATE campaign_legs SET residual_quantity='999' WHERE campaign_id='campaign' AND sequence=1`},
		{name: "command result", statement: `UPDATE campaign_commands SET result_version=999 WHERE campaign_id='campaign' AND command_kind='PLAN_LEG'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			campaign := createCampaignFixture(t, j)
			if _, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
				CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
				Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
			}); err != nil {
				t.Fatal(err)
			}
			if tc.name == "command result" {
				if _, err := j.db.Exec(`DROP TRIGGER campaign_commands_no_update`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := j.db.Exec(tc.statement); err != nil {
				t.Fatal(err)
			}
			replay, err := j.ReconstructPositionCampaign(context.Background(), campaign.ID)
			if err != nil {
				t.Fatal(err)
			}
			if replay.Valid || replay.Reason != positioncampaign.ReplaySnapshotDrift {
				t.Fatalf("replay=%+v", replay)
			}
		})
	}
}

func TestCampaignCommandsAndEventsAreAppendOnly(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	if _, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
	}); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`UPDATE campaign_commands SET result_version=999 WHERE campaign_id='campaign'`,
		`DELETE FROM campaign_commands WHERE campaign_id='campaign'`,
		`UPDATE campaign_events SET campaign_version=999 WHERE campaign_id='campaign'`,
		`DELETE FROM campaign_events WHERE campaign_id='campaign'`,
	} {
		if _, err := j.db.Exec(statement); err == nil {
			t.Fatalf("append-only mutation succeeded: %s", statement)
		}
	}
}

func TestOpenReadOnlyRejectsDamagedV20CampaignSchema(t *testing.T) {
	for _, tc := range []struct{ name, damage, want string }{
		{name: "missing table", damage: `DROP TABLE campaign_commands`, want: "campaign_commands"},
		{name: "missing column", damage: `ALTER TABLE campaign_events DROP COLUMN projection_digest`, want: "campaign_events.projection_digest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), DBFileName)
			j := openTestJournalAt(t, path)
			if _, err := j.db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(tc.damage); err != nil {
				t.Fatal(err)
			}
			if err := j.Close(); err != nil {
				t.Fatal(err)
			}
			_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
			if !errors.Is(err, ErrSchemaTooOld) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("damaged v20 open err=%v, want typed %s", err, tc.want)
			}
		})
	}
}

func TestCampaignCoreHasNoProductionBrokerOrToggleWiring(t *testing.T) {
	production, err := os.ReadFile(filepath.Join("..", "app", "engine", "gateway.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(production)
	if strings.Contains(text, "Campaign: ApplyPositionCampaignFill") || strings.Contains(text, "ApplyPositionCampaignFill,") {
		t.Fatal("a065 campaign hook is wired into the production gateway")
	}
	core, err := os.ReadFile("position_campaign.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"tossctl", "order place", "SetLaneEnabled", "AutomationEnabled"} {
		if strings.Contains(string(core), forbidden) {
			t.Fatalf("campaign persistence contains production mutation binding %q", forbidden)
		}
	}
}
