package positioncampaign

import "testing"

func TestComposeLongStopNeverRetreatsAndPreservesProvenance(t *testing.T) {
	saved := &EffectiveStop{Price: "100", Source: "saved", Policy: "v1", ObservedAt: "t0"}
	got, blocked, err := ComposeLongStop(saved, StopCandidate{Price: "90", Valid: true, Source: "lane", Policy: "v2", ObservedAt: "t1"})
	if err != nil || blocked || got.Price != "100" || got.SelectedFrom != StopFromSaved {
		t.Fatalf("lower candidate got=%+v blocked=%v err=%v", got, blocked, err)
	}
	got, blocked, err = ComposeLongStop(saved, StopCandidate{Price: "110", Valid: true, Source: "lane", Policy: "v2", ObservedAt: "t2"})
	if err != nil || blocked || got.Price != "110" || got.Source != "lane" || got.SelectedFrom != StopFromCandidate {
		t.Fatalf("higher candidate got=%+v blocked=%v err=%v", got, blocked, err)
	}
	got, blocked, err = ComposeLongStop(saved, StopCandidate{Price: "", Valid: false, Source: "lane", Policy: "v3", ObservedAt: "t3"})
	if err != nil || !blocked || got.Price != "100" {
		t.Fatalf("invalid candidate got=%+v blocked=%v err=%v", got, blocked, err)
	}
}

func TestReplayIsDeterministicAndReportsStableMismatch(t *testing.T) {
	events := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create-1", RequestDigest: "d1", CampaignState: CampaignPlanned, ExpectedPositionGeneration: 1},
		{Sequence: 2, CampaignVersion: 2, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan-1", RequestDigest: "d2", CampaignState: CampaignPlanned, LegSequence: 1, PlanID: "plan", LegState: LegPlanned, LegRequestedQuantity: "1", LegFilledQuantity: "0", LegResidualQuantity: "1"},
		{Sequence: 3, CampaignVersion: 3, EventKind: "ORDER_LINKED", CommandKind: "LINK_ORDER", CommandKey: "bind-1", RequestDigest: "d3", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan", LegState: LegSubmitted, LegRequestedQuantity: "1", LegFilledQuantity: "0", LegResidualQuantity: "1", OrderID: "order", RequestedCap: "1", CumulativeQuantity: "0", OrderRemainingQuantity: "1", PositionGeneration: 2},
	}
	one := Replay(events, Snapshot{CampaignState: CampaignActive, Version: 3, PositionGeneration: 2})
	two := Replay(events, Snapshot{CampaignState: CampaignActive, Version: 3, PositionGeneration: 2})
	if !one.Valid || one != two || one.LastValidSequence != 3 {
		t.Fatalf("replay one=%+v two=%+v", one, two)
	}
	bad := Replay(append(events, Event{Sequence: 5, CampaignVersion: 5, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan-2", RequestDigest: "d5"}), Snapshot{})
	if bad.Valid || bad.Reason != ReplaySequenceGap || bad.LastValidSequence != 3 {
		t.Fatalf("bad replay=%+v", bad)
	}
	dup := Replay(append(events, Event{Sequence: 4, CampaignVersion: 4, EventKind: "STOP_COMPOSED", CommandKind: "LINK_ORDER", CommandKey: "bind-1", RequestDigest: "d4", CampaignState: CampaignActive}), Snapshot{})
	if dup.Valid || dup.Reason != ReplayDuplicateCommandKey {
		t.Fatalf("duplicate replay=%+v", dup)
	}
}
