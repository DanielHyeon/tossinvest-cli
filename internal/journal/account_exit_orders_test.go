package journal

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestReadOnlyLinksBrokerOrdersToExitSnapshotsOnlyThroughTheAttemptIntent(t *testing.T) {
	j := exitFixture(t)
	_, state := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, state, "obs-order-view", "67900", "70000", "68000")
	if snapshot.Action != exitpolicy.ActionBaselineBreach || !snapshot.Orderable {
		t.Fatalf("fixture is not an orderable protection breach: %+v", snapshot)
	}
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.Proposal = &ExitProposal{
		Action: string(snapshot.Action), Level: snapshot.Level, IntentID: "exit-intent-view",
		Provenance: judgement.Provenance,
	}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}

	insertIntent(t, j, "exit-intent-view")
	insertAttemptWithBrokerOrder(t, j, "exit-attempt-view", "exit-intent-view", "broker-exit-view")
	insertIntent(t, j, "other-intent-view")
	insertAttemptWithBrokerOrder(t, j, "other-attempt-view", "other-intent-view", "same-symbol-same-time")

	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("links = %+v, want only the exact attempt/intent chain", links)
	}
	got := links[0]
	if got.BrokerOrderID != "broker-exit-view" || got.AttemptID != "exit-attempt-view" ||
		got.IntentID != "exit-intent-view" || got.Event.Evaluation.Effective.Snapshot == nil ||
		got.Event.Evaluation.Effective.Snapshot.Line.DecisionID != snapshot.DecisionID {
		t.Fatalf("exact exit link = %+v, want broker -> attempt -> intent -> event decision", got)
	}
	if got.Ambiguous {
		t.Fatal("one exact chain was marked ambiguous")
	}
	for _, link := range links {
		if link.BrokerOrderID == "same-symbol-same-time" {
			t.Fatal("an unrelated attempt was linked without the exit intent reference")
		}
	}
}

func TestReadOnlyMarksDuplicateExitEvidenceAmbiguous(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	j := openTestJournalAt(t, path)
	insertIntent(t, j, "exit-intent-ambiguous")
	insertAttemptWithBrokerOrder(t, j, "exit-attempt-ambiguous", "exit-intent-ambiguous", "broker-ambiguous")
	insertPosition(t, j, "ambiguous-position", nil)
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO exit_events(position_id, action, proposed_intent_id, created_at)
		VALUES ('ambiguous-position','LEGACY_A','exit-intent-ambiguous','2026-07-31T00:00:00Z'),
		       ('ambiguous-position','LEGACY_B','exit-intent-ambiguous','2026-07-31T00:00:01Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, path).BrokerOrderExitLinks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].Event.Evaluation.Effective.Snapshot != nil {
		t.Fatalf("ambiguous links = %+v, want one fail-closed marker without snapshot", links)
	}
}
