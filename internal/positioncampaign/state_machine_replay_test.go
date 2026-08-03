package positioncampaign

import "testing"

func TestReplayDerivesCampaignAndLegStateFromEvents(t *testing.T) {
	base := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create",
			RequestDigest: "create-digest", CampaignState: CampaignPlanned, ProspectiveToken: "token"},
		{Sequence: 2, CampaignVersion: 2, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan",
			RequestDigest: "plan-digest", CampaignState: CampaignPlanned, LegSequence: 1, PlanID: "plan",
			LegState: LegPlanned, LegRequestedQuantity: "5", LegFilledQuantity: "0", LegResidualQuantity: "5",
			ProspectiveToken: "token"},
		{Sequence: 3, CampaignVersion: 3, EventKind: "ORDER_LINKED", CommandKind: "LINK_ORDER", CommandKey: "link",
			RequestDigest: "link-digest", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan",
			LegState: LegSubmitted, LegRequestedQuantity: "5", LegFilledQuantity: "0", LegResidualQuantity: "5",
			OrderID: "order", RequestedCap: "5", CumulativeQuantity: "0", OrderRemainingQuantity: "5",
			ProspectiveToken: "token"},
	}

	t.Run("stored campaign state cannot invent a transition", func(t *testing.T) {
		events := append([]Event(nil), base...)
		events[1].CampaignState = CampaignActive
		got := Replay(events, Snapshot{})
		if got.Valid || got.Reason != ReplayInvalidCampaignTransition {
			t.Fatalf("replay=%+v", got)
		}
	})
	t.Run("stored leg state cannot skip transition", func(t *testing.T) {
		events := append([]Event(nil), base...)
		events[2].LegState = LegFilled
		got := Replay(events, Snapshot{})
		if got.Valid || got.Reason != ReplayInvalidLegTransition {
			t.Fatalf("replay=%+v", got)
		}
	})
	t.Run("campaign version must match sequence", func(t *testing.T) {
		events := append([]Event(nil), base...)
		events[2].CampaignVersion = 99
		got := Replay(events, Snapshot{})
		if got.Valid || got.Reason != ReplayCampaignVersionMismatch {
			t.Fatalf("replay=%+v", got)
		}
	})
}

func TestReplayRejectsDeltaAndOrderRemainingDrift(t *testing.T) {
	events := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create",
			RequestDigest: "create-digest", CampaignState: CampaignPlanned, ProspectiveToken: "token"},
		{Sequence: 2, CampaignVersion: 2, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan",
			RequestDigest: "plan-digest", CampaignState: CampaignPlanned, LegSequence: 1, PlanID: "plan",
			LegState: LegPlanned, LegRequestedQuantity: "5", LegFilledQuantity: "0", LegResidualQuantity: "5",
			ProspectiveToken: "token"},
		{Sequence: 3, CampaignVersion: 3, EventKind: "ORDER_LINKED", CommandKind: "LINK_ORDER", CommandKey: "link",
			RequestDigest: "link-digest", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan",
			LegState: LegSubmitted, LegRequestedQuantity: "5", LegFilledQuantity: "0", LegResidualQuantity: "5",
			OrderID: "order", RequestedCap: "5", CumulativeQuantity: "0", OrderRemainingQuantity: "5",
			ProspectiveToken: "token"},
		{Sequence: 4, CampaignVersion: 4, EventKind: "ORDER_WATERMARK_ADVANCED", CommandKind: "APPLY_FILL", CommandKey: "fill",
			RequestDigest: "fill-digest", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan",
			LegState: LegPartial, LegRequestedQuantity: "5", LegFilledQuantity: "1", LegResidualQuantity: "4",
			OrderID: "order", RequestedCap: "5", DeltaQuantity: "2", CumulativeQuantity: "1",
			OrderRemainingQuantity: "4", ProspectiveToken: "token"},
	}
	got := Replay(events, Snapshot{})
	if got.Valid || got.Reason != ReplayOrderDeltaMismatch {
		t.Fatalf("replay=%+v", got)
	}

	events[3].DeltaQuantity = "1"
	events[3].OrderRemainingQuantity = "3"
	got = Replay(events, Snapshot{})
	if got.Valid || got.Reason != ReplayOrderRemainingMismatch {
		t.Fatalf("replay=%+v", got)
	}
}
