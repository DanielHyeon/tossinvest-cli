package positioncampaign

import "testing"

func TestProtectionReconcileReplayBlocksEntryAndCannotReopen(t *testing.T) {
	events := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create", RequestDigest: "create", CampaignState: CampaignPlanned, ProspectiveToken: "token"},
		{Sequence: 2, CampaignVersion: 2, EventKind: "PROTECTION_RECONCILE", CommandKind: "PROTECTION_RECONCILE", CommandKey: "protect", RequestDigest: "protect", CampaignState: CampaignReconcile, EntryBlocked: true, ProspectiveToken: "token"},
	}
	if got := Replay(events, Snapshot{}); !got.Valid || got.State.CampaignState != CampaignReconcile || !got.State.EntryBlocked {
		t.Fatalf("replay=%+v", got)
	}
	if err := ValidateCommand("PROTECTION_RECONCILE", "protect"); err != nil {
		t.Fatal(err)
	}
	events[1].EntryBlocked = false
	if got := Replay(events, Snapshot{}); got.Valid {
		t.Fatalf("unblocked protection replay=%+v", got)
	}
}
