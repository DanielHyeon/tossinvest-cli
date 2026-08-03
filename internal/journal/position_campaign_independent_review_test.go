package journal

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
)

func TestCreatePositionCampaignRequiresExactStrategyDecisionLineage(t *testing.T) {
	request := CreatePositionCampaignRequest{
		ID: "campaign", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "kr", LaneVersion: "v1", DecisionID: "decision", EvidenceDigest: "digest",
		ProspectiveToken: "token", CommandKey: "create",
	}
	t.Run("decision without lineage", func(t *testing.T) {
		j := openTestJournal(t)
		insertCampaignDecision(t, j, "decision", "acct")
		if _, err := j.CreatePositionCampaign(context.Background(), request); !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("cross market lineage", func(t *testing.T) {
		j := openTestJournal(t)
		insertCampaignDecision(t, j, "decision", "acct")
		insertCampaignStrategyLineage(t, j, "decision", "acct", "us", "005930", "kr", "v1", "digest")
		if _, err := j.CreatePositionCampaign(context.Background(), request); !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestCampaignReplayBindsImmutableIdentityAndDurableClaim(t *testing.T) {
	for _, tc := range []struct {
		name      string
		statement string
	}{
		{name: "claim deleted", statement: `DELETE FROM position_campaign_claims WHERE campaign_id='campaign'`},
		{name: "claim generation changed", statement: `UPDATE position_campaign_claims SET position_generation=1 WHERE campaign_id='campaign'`},
		{name: "immutable lane changed", statement: `UPDATE position_campaigns SET lane_version='forged' WHERE id='campaign'`},
		{name: "expected version changed", statement: `UPDATE position_campaigns SET expected_position_version=7 WHERE id='campaign'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			createCampaignFixture(t, j)
			if _, err := j.db.Exec(tc.statement); err != nil {
				t.Fatal(err)
			}
			replay, err := j.ReconstructPositionCampaign(context.Background(), "campaign")
			if err != nil {
				t.Fatal(err)
			}
			if replay.Valid || replay.Reason != positioncampaign.ReplaySnapshotDrift {
				t.Fatalf("replay=%+v", replay)
			}
		})
	}
}

func TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch(t *testing.T) {
	t.Run("caller ambiguity is durable and in command digest", func(t *testing.T) {
		j := openTestJournal(t)
		campaign := createCampaignFixture(t, j)
		leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
			CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
			Sequence: 1, PlanID: "plan", RequestedQuantity: "2",
		})
		insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "2")
		req := LinkCampaignOrderRequest{CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
			CommandKey: "link", OrderID: "order", IntentID: "intent", AttemptID: "attempt", RequestedCap: "2", LineageAmbiguous: true}
		if _, err := j.LinkCampaignOrder(context.Background(), req); !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
			t.Fatalf("err=%v", err)
		}
		latched, _ := j.PositionCampaign(context.Background(), campaign.ID)
		if latched.State != positioncampaign.CampaignReconcile || !latched.EntryBlocked {
			t.Fatalf("campaign=%+v", latched)
		}
		req.LineageAmbiguous = false
		if _, err := j.LinkCampaignOrder(context.Background(), req); !errors.Is(err, ErrCampaignCommandConflict) {
			t.Fatalf("same key with different ambiguity err=%v", err)
		}
	})

	t.Run("intent quantity mismatch", func(t *testing.T) {
		j := openTestJournal(t)
		campaign := createCampaignFixture(t, j)
		leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
			CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
			Sequence: 1, PlanID: "plan", RequestedQuantity: "2",
		})
		insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "2")
		_, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
			CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
			CommandKey: "link", OrderID: "order", IntentID: "intent", AttemptID: "attempt", RequestedCap: "1"})
		if !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("replacement edge quantity mismatch", func(t *testing.T) {
		j := openTestJournal(t)
		campaign := createCampaignFixture(t, j)
		leg, _ := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{
			CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
			Sequence: 1, PlanID: "plan", RequestedQuantity: "4",
		})
		insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-a", "attempt-a", "a", "", "2")
		first, err := j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
			CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
			CommandKey: "link-a", OrderID: "a", IntentID: "intent-a", AttemptID: "attempt-a", RequestedCap: "2"})
		if err != nil {
			t.Fatal(err)
		}
		insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-b", "attempt-b", "b", "a", "2")
		if _, err := j.db.Exec(`UPDATE scoped_lineage_edges SET requested_quantity='3' WHERE child_order_id='b'`); err != nil {
			t.Fatal(err)
		}
		_, err = j.LinkCampaignOrder(context.Background(), LinkCampaignOrderRequest{
			CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: first.CampaignVersion,
			CommandKey: "link-b", OrderID: "b", PredecessorOrderID: "a", IntentID: "intent-b", AttemptID: "attempt-b", RequestedCap: "2"})
		if !errors.Is(err, positioncampaign.ErrInvalidIdentity) {
			t.Fatalf("err=%v", err)
		}
	})
}
