package journal

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
)

func TestMigrationV20AddsCampaignSchemaWithoutBackfillingPositions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 19)
	if _, err := old.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price)
		VALUES ('legacy-position','acct','kr','005930',1,'OPEN','1','70000')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 20)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 20 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	for _, table := range []string{
		"position_campaigns", "campaign_legs", "campaign_order_watermarks",
		"campaign_commands", "campaign_events", "position_campaign_claims", "position_projection_versions",
	} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("table %s count=%d err=%v", table, count, err)
		}
	}
	var claims, campaigns int
	_ = j.db.QueryRow(`SELECT count(*) FROM position_campaign_claims`).Scan(&claims)
	_ = j.db.QueryRow(`SELECT count(*) FROM position_campaigns`).Scan(&campaigns)
	if claims != 0 || campaigns != 0 {
		t.Fatalf("migration synthesized claims=%d campaigns=%d", claims, campaigns)
	}
	var legacyVersions int
	_ = j.db.QueryRow(`SELECT count(*) FROM position_projection_versions`).Scan(&legacyVersions)
	if legacyVersions != 0 {
		t.Fatalf("migration backfilled %d legacy position versions", legacyVersions)
	}
	lineage, err := j.PositionCampaignLineage(context.Background(), "legacy-position")
	if err != nil {
		t.Fatal(err)
	}
	if lineage.Status != PositionCampaignLineageLegacyUnknown || lineage.CampaignID != "" || lineage.PositionGeneration != 1 {
		t.Fatalf("legacy lineage=%+v", lineage)
	}
}

func TestCreatePositionCampaignProspectiveCASAndRetry(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertCampaignDecision(t, j, "decision-1", "acct")
	insertCampaignStrategyLineage(t, j, "decision-1", "acct", "kr", "005930", "kr-swing", "v1", "sha256:evidence")
	req := CreatePositionCampaignRequest{
		ID: "campaign-1", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "kr-swing", LaneVersion: "v1", DecisionID: "decision-1",
		EvidenceDigest: "sha256:evidence", ExpectedPositionGeneration: 0,
		ExpectedPositionVersion: 0, ProspectiveToken: "prospective-1", CommandKey: "create-1",
	}
	first, err := j.CreatePositionCampaign(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := j.CreatePositionCampaign(ctx, req)
	if err != nil || retry != first {
		t.Fatalf("retry=%+v first=%+v err=%v", retry, first, err)
	}
	conflict := req
	conflict.ID = "campaign-2"
	conflict.ProspectiveToken = "prospective-2"
	conflict.CommandKey = "create-2"
	if _, err := j.CreatePositionCampaign(ctx, conflict); !errors.Is(err, ErrGenerationConflict) {
		t.Fatalf("second campaign err=%v, want generation conflict", err)
	}
}

func TestConcurrentProspectiveCampaignCreationHasOneWinner(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertCampaignDecision(t, j, "decision-a", "acct")
	insertCampaignDecision(t, j, "decision-b", "acct")
	insertCampaignStrategyLineage(t, j, "decision-a", "acct", "us", "AAPL", "us-swing", "v1", "digest-a")
	insertCampaignStrategyLineage(t, j, "decision-b", "acct", "us", "AAPL", "us-swing", "v1", "digest-b")
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i, id := range []string{"a", "b"} {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			<-start
			_, err := j.CreatePositionCampaign(ctx, CreatePositionCampaignRequest{
				ID: "campaign-" + id, AccountRef: "acct", Market: "us", Symbol: "AAPL",
				LaneID: "us-swing", LaneVersion: "v1", DecisionID: "decision-" + id,
				EvidenceDigest: "digest-" + id, ExpectedPositionGeneration: 0,
				ExpectedPositionVersion: 0, ProspectiveToken: "token-" + id,
				CommandKey: "create-" + id,
			})
			errs <- err
		}(i, id)
	}
	close(start)
	wg.Wait()
	close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrGenerationConflict):
			conflicts++
		default:
			t.Fatalf("unexpected create error: %v", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
}

func TestCampaignCommandsExpectedVersionAndDeterministicRetry(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-1",
		Sequence: 1, PlanID: "leg-plan-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-1",
		Sequence: 1, PlanID: "leg-plan-1", RequestedQuantity: "10",
	})
	if err != nil || retry != leg {
		t.Fatalf("retry=%+v leg=%+v err=%v", retry, leg, err)
	}
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan-stale",
		Sequence: 2, PlanID: "leg-plan-2", RequestedQuantity: "2",
	}); !errors.Is(err, ErrCampaignVersionConflict) {
		t.Fatalf("stale plan err=%v, want version conflict", err)
	}
}

func TestCampaignStopCompositionPersistsCandidateAndNeverRetreats(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	higher, err := j.UpdateCampaignStop(ctx, UpdateCampaignStopRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "stop-1",
		Candidate: positioncampaign.StopCandidate{Price: "100", Valid: true, Source: "risk", Policy: "v1", ObservedAt: "t1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	lower, err := j.UpdateCampaignStop(ctx, UpdateCampaignStopRequest{
		CampaignID: campaign.ID, ExpectedVersion: higher.Version, CommandKey: "stop-2",
		Candidate: positioncampaign.StopCandidate{Price: "90", Valid: true, Source: "lane", Policy: "v2", ObservedAt: "t2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if lower.EffectiveStop != "100" || lower.StopSource != "risk" || lower.StopCandidate != "90" ||
		!lower.StopCandidateValid || lower.StopSelectedFrom != positioncampaign.StopFromSaved {
		t.Fatalf("lower candidate retreated or lost provenance: %+v", lower)
	}
	invalid, err := j.UpdateCampaignStop(ctx, UpdateCampaignStopRequest{
		CampaignID: campaign.ID, ExpectedVersion: lower.Version, CommandKey: "stop-3",
		Candidate: positioncampaign.StopCandidate{Valid: false, Source: "lane", Policy: "v3", ObservedAt: "t3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if invalid.EffectiveStop != "100" || !invalid.EntryBlocked {
		t.Fatalf("invalid candidate cleared stop or did not block exposure: %+v", invalid)
	}
}

func TestCancelledProspectiveTokenIsNotReusedAndNextCASVersionCanCreate(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first := createCampaignFixture(t, j)
	closed, err := j.CancelProspectiveCampaign(ctx, CancelCampaignRequest{
		CampaignID: first.ID, ExpectedVersion: first.Version, CommandKey: "cancel", Detail: "no fill",
	})
	if err != nil || closed.State != positioncampaign.CampaignClosed {
		t.Fatalf("closed=%+v err=%v", closed, err)
	}
	insertCampaignDecision(t, j, "decision-2", "acct")
	insertCampaignStrategyLineage(t, j, "decision-2", "acct", "kr", "005930", "kr", "v2", "digest-2")
	second, err := j.CreatePositionCampaign(ctx, CreatePositionCampaignRequest{
		ID: "campaign-2", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "kr", LaneVersion: "v2", DecisionID: "decision-2", EvidenceDigest: "digest-2",
		ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
		ProspectiveToken: "token-2", CommandKey: "create-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ProspectiveToken == first.ProspectiveToken || second.State != positioncampaign.CampaignPlanned {
		t.Fatalf("second=%+v", second)
	}
}

func TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-old", "attempt-old", "old", "", "10")
	linked, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link-old", OrderID: "old", RequestedCap: "10", IntentID: "intent-old", AttemptID: "attempt-old",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('campaign-position','acct','kr','005930',1,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "old", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "4", CumulativeQuantity: "4", Terminal: false, CommittedAt: "2026-03-30T00:31:00Z"})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-old", "attempt-new", "new", "old", "6")
	linked, err = j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: linked.CampaignVersion + 1,
		CommandKey: "link-new", OrderID: "new", PredecessorOrderID: "old", RequestedCap: "6",
		IntentID: "intent-old", AttemptID: "attempt-new",
	})
	if err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "new", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "2", CumulativeQuantity: "2", CommittedAt: "2026-03-30T00:32:00Z"})
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "old", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "1", CumulativeQuantity: "5", Terminal: true, CommittedAt: "2026-03-30T00:33:00Z"})
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "old", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "0", CumulativeQuantity: "5", Terminal: true, CommittedAt: "2026-03-30T00:34:00Z"})

	got, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != positioncampaign.CampaignReconcile || !got.EntryBlocked {
		t.Fatalf("campaign=%+v, want reconcile latch", got)
	}
	legGot, err := j.CampaignLeg(ctx, campaign.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if legGot.FilledQuantity != "7" || legGot.ResidualQuantity != "3" {
		t.Fatalf("leg=%+v, want filled 7 residual 3", legGot)
	}
	watermark, err := j.CampaignOrderWatermark(ctx, campaign.ID, 1, "new")
	if err != nil || watermark.RemainingQuantity != "4" {
		t.Fatalf("successor=%+v err=%v, want successor-cap remaining 4", watermark, err)
	}
}

func TestCampaignFillRollbackAndRestartAreDeterministic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campaign-restart.db")
	j := openTestJournalAt(t, path)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg", RequestedQuantity: "5",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "5")
	order, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "5", IntentID: "intent", AttemptID: "attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('p','acct','kr','005930',1,'OPEN','2','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	fill := AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr", Symbol: "005930",
		Delta: "2", CumulativeQuantity: "2", CommittedAt: "2026-03-30T00:31:00Z"}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	handle := &ApplyTx{tx: tx, now: fill.CommittedAt}
	if err := ApplyPositionCampaignFill(ctx, handle, fill); err != nil {
		t.Fatal(err)
	}
	handle.invalidate()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	rolledBack, err := j.CampaignOrderWatermark(ctx, campaign.ID, 1, "order")
	if err != nil || rolledBack.CumulativeFilled != "0" || rolledBack.CampaignVersion != order.CampaignVersion {
		t.Fatalf("rollback watermark=%+v err=%v", rolledBack, err)
	}
	applyCampaignFillForTest(t, j, fill)
	beforeRestart, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := openTestJournalAt(t, path)
	applyCampaignFillForTest(t, restarted, fill)
	afterRestart, err := restarted.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if afterRestart.Version != beforeRestart.Version || afterRestart.ActualPositionGeneration != 1 {
		t.Fatalf("restart moved duplicate: before=%+v after=%+v", beforeRestart, afterRestart)
	}
	replay, err := restarted.ReconstructPositionCampaign(ctx, campaign.ID)
	if err != nil || !replay.Valid || replay.LastValidSequence != afterRestart.Version {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
}

func TestLinkCampaignOrderRequiresImmutableIntentAttemptLineage(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "1",
	})
	if !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
		t.Fatalf("missing lineage err=%v, want invalid identity", err)
	}
}

func TestCampaignPlanAndOrderCapRejectZeroQuantity(t *testing.T) {
	j := openTestJournal(t)
	campaign := createCampaignFixture(t, j)
	_, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "zero-plan",
		Sequence: 1, PlanID: "zero", RequestedQuantity: "0",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("zero plan err=%v, want invalid request", err)
	}
	leg, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "positive", RequestedQuantity: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "zero-cap", OrderID: "order", RequestedCap: "0",
		IntentID: "intent", AttemptID: "attempt",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("zero cap err=%v, want invalid request", err)
	}
}

func TestCampaignAggregateFillExcessLatchesReconcileWithoutTruncation(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "plan", RequestedQuantity: "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-a", "attempt-a", "a", "", "4")
	first, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link-a", OrderID: "a", RequestedCap: "4", IntentID: "intent-a", AttemptID: "attempt-a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('p','acct','kr','005930',1,'OPEN','8','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "a", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "4", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:31:00Z"})
	current, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-a", "attempt-b", "b", "a", "4")
	_, err = j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: current.Version,
		CommandKey: "link-b", OrderID: "b", PredecessorOrderID: "a", RequestedCap: "4",
		IntentID: "intent-a", AttemptID: "attempt-b",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = first
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "b", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "4", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:32:00Z"})
	got, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	legGot, err := j.CampaignLeg(ctx, campaign.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != positioncampaign.CampaignReconcile || !got.EntryBlocked || legGot.FilledQuantity != "8" || legGot.ResidualQuantity != "0" {
		t.Fatalf("campaign=%+v leg=%+v", got, legGot)
	}
}

func TestCampaignZeroAndPartialUnchangedTerminalObservationsCancelResidualIdempotently(t *testing.T) {
	for _, tc := range []struct {
		name, first, requested string
	}{
		{name: "zero fill terminal", first: "0", requested: "5"},
		{name: "partial unchanged terminal", first: "2", requested: "5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			campaign := createCampaignFixture(t, j)
			leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
				CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
				Sequence: 1, PlanID: "plan", RequestedQuantity: tc.requested,
			})
			if err != nil {
				t.Fatal(err)
			}
			insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", tc.requested)
			order, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
				CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
				CommandKey: "link", OrderID: "order", RequestedCap: tc.requested,
				IntentID: "intent", AttemptID: "attempt",
			})
			if err != nil {
				t.Fatal(err)
			}
			if tc.first != "0" {
				if _, err := j.db.Exec(`INSERT INTO positions
					(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
					VALUES ('p','acct','kr','005930',1,'OPEN',?,'70000','2026-03-30T00:31:00Z')`, tc.first); err != nil {
					t.Fatal(err)
				}
				applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr", Symbol: "005930",
					Delta: tc.first, CumulativeQuantity: tc.first, CommittedAt: "2026-03-30T00:31:00Z"})
			}
			beforeTerminal, err := j.PositionCampaign(ctx, campaign.ID)
			if err != nil {
				t.Fatal(err)
			}
			applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr", Symbol: "005930",
				Delta: "0", CumulativeQuantity: tc.first, Terminal: true, State: "CANCELLED", CommittedAt: "2026-03-30T00:32:00Z"})
			legGot, err := j.CampaignLeg(ctx, campaign.ID, 1)
			if err != nil {
				t.Fatal(err)
			}
			watermark, err := j.CampaignOrderWatermark(ctx, campaign.ID, 1, "order")
			if err != nil {
				t.Fatal(err)
			}
			if legGot.State != positioncampaign.LegCancelled || legGot.FilledQuantity != tc.first ||
				!watermark.Terminal || watermark.CumulativeFilled != tc.first {
				t.Fatalf("leg=%+v watermark=%+v", legGot, watermark)
			}
			terminalVersion := watermark.CampaignVersion
			if terminalVersion != beforeTerminal.Version+1 {
				t.Fatalf("terminal version=%d before=%d", terminalVersion, beforeTerminal.Version)
			}
			applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr", Symbol: "005930",
				Delta: "0", CumulativeQuantity: tc.first, Terminal: true, State: "CANCELLED", CommittedAt: "2026-03-30T00:33:00Z"})
			retry, _ := j.PositionCampaign(ctx, campaign.ID)
			if retry.Version != terminalVersion {
				t.Fatalf("duplicate terminal moved version %d -> %d (initial order version %d)", terminalVersion, retry.Version, order.CampaignVersion)
			}
		})
	}
}

func TestHasCampaignSuccessorPropagatesQueryFailure(t *testing.T) {
	j := openTestJournal(t)
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	handle := &ApplyTx{tx: tx, now: "2026-03-30T00:31:00Z"}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := hasCampaignSuccessor(context.Background(), handle, "campaign", 1, "order"); err == nil {
		t.Fatal("closed transaction query failure was interpreted as no successor")
	}
}

func createCampaignFixture(t *testing.T, j *Journal) PositionCampaignRecord {
	t.Helper()
	insertCampaignDecision(t, j, "decision", "acct")
	insertCampaignStrategyLineage(t, j, "decision", "acct", "kr", "005930", "kr", "v1", "digest")
	got, err := j.CreatePositionCampaign(context.Background(), CreatePositionCampaignRequest{
		ID: "campaign", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "kr", LaneVersion: "v1", DecisionID: "decision", EvidenceDigest: "digest",
		ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
		ProspectiveToken: "token", CommandKey: "create",
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func applyCampaignFillForTest(t *testing.T, j *Journal, fill AppliedFill) {
	t.Helper()
	if fill.TradingDay == "" {
		fill.TradingDay = "2026-03-30"
	}
	if fill.Side == "" {
		fill.Side = "BUY"
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	handle := &ApplyTx{tx: tx, now: fill.CommittedAt}
	err = ApplyPositionCampaignFill(context.Background(), handle, fill)
	handle.invalidate()
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func insertCampaignDecision(t *testing.T, j *Journal, id, account string) {
	t.Helper()
	_, err := j.db.Exec(`INSERT OR IGNORE INTO decisions
		(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,
		 client_order_id,limits_json,nonce,issued_at,expires_at)
		VALUES (?, ?, 0, 'EXPOSURE_RAISING', 'RISK_INTENT', '{}', ?, ?, '{}', ?,
		        '2026-03-30T00:00:00Z', '2026-03-30T01:00:00Z')`,
		id, account, "hash-"+id, "client-"+id, "nonce-"+id)
	if err != nil {
		t.Fatalf("insert campaign decision %s: %v", id, err)
	}
}

func insertCampaignStrategyLineage(t *testing.T, j *Journal, decisionID, account, market, symbol, laneID, laneVersion, evidence string) {
	t.Helper()
	entryID := "entry-" + decisionID
	if _, err := j.db.Exec(`INSERT INTO strategy_decision_lineage
		(entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,
		 evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,
		 stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,
		 decision_payload_digest,activation_manifest_digest,created_at)
		VALUES (?,? ,?,?, 'threshold-v1','threshold-digest', ?,?,?, 'lane-source','lane-constants',
		        '100','90','120','10','policy-v1','settings','{}','payload-digest','manifest',
		        '2026-03-30T00:00:00Z')`, entryID, "candidate-"+decisionID, market, symbol,
		evidence, laneID, laneVersion); err != nil {
		t.Fatalf("insert campaign strategy decision lineage %s: %v", decisionID, err)
	}
	if _, err := j.db.Exec(`INSERT INTO strategy_attempt_lineage
		(attempt_id,account_ref,entry_decision_identity,risk_intent_id,guardian_decision_id,
		 activation_manifest_digest,client_order_id,revision,state,created_at)
		VALUES (?,?,?,?,?,'manifest',?,1,'PLANNED','2026-03-30T00:00:00Z')`,
		"strategy-attempt-"+decisionID, account, entryID, decisionID, decisionID, "strategy-client-"+decisionID); err != nil {
		t.Fatalf("insert campaign strategy attempt lineage %s: %v", decisionID, err)
	}
}

func insertCampaignExecutionLineage(t *testing.T, j *Journal, decisionID, intentID, attemptID, orderID, predecessor, quantity string) {
	t.Helper()
	kind := "PLACE"
	if predecessor != "" {
		kind = "AMEND"
	}
	if _, err := j.db.Exec(`INSERT OR IGNORE INTO intents
		(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,
		 quantity,price,currency,source,fingerprint,notes)
		VALUES (?, '2026-03-30T00:10:00Z', 'kr', '2026-03-30', 'acct', '005930',
		        'BUY', 'LIMIT', '', ?, '70000', 'KRW', 'campaign-test', ?, '')`,
		intentID, quantity, "fp-"+intentID); err != nil {
		t.Fatalf("insert intent %s: %v", intentID, err)
	}
	if _, err := j.db.Exec(`INSERT OR IGNORE INTO mutation_attempts
		(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,
		 recorded_at,dispatch_started_at,settled_at,decision_id,account_ref)
		VALUES (?, ?, ?, 'CONFIRMED', 1, ?, ?, ?, '2026-03-30T00:11:00Z',
		        '2026-03-30T00:11:00Z', '2026-03-30T00:12:00Z', ?, 'acct')`,
		attemptID, intentID, kind, predecessor, orderID, "fp-"+attemptID, decisionID); err != nil {
		t.Fatalf("insert attempt %s: %v", attemptID, err)
	}
	if predecessor != "" {
		if _, err := j.db.Exec(`INSERT OR IGNORE INTO scoped_lineage_edges
			(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,
			 account_ref,market,trading_day,symbol,side,intent_id,attempt_id,created_at)
			VALUES (?, ?, 'replaces', '0', ?, 'acct', 'kr', '2026-03-30', '005930',
			        'BUY', ?, ?, '2026-03-30T00:12:00Z')`, predecessor, orderID, quantity, intentID, attemptID); err != nil {
			t.Fatalf("insert scoped lineage %s->%s: %v", predecessor, orderID, err)
		}
	}
}
