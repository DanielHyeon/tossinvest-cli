package journal

import (
	"context"
	"database/sql"
	"errors"
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

	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "broker-exit-view", AccountRef: "acct-1", TradingDay: "2026-03-30"},
			{BrokerOrderID: "same-symbol-same-time", AccountRef: "acct-1", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("links = %+v, want one answer per requested scope", links)
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
	if !links[1].Engine || links[1].Event.ID != 0 || links[1].UnknownReason != "exit_evidence_unlinked" {
		t.Fatalf("unrelated engine order evidence = %+v, want scoped origin plus typed unlinked result", links[1])
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
	links, err := openTestReadOnly(t, path).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "broker-ambiguous", AccountRef: "acct-1", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].Event.Evaluation.Effective.Snapshot != nil {
		t.Fatalf("ambiguous links = %+v, want one fail-closed marker without snapshot", links)
	}
}

func TestBrokerOrderExitLinksScopesCollidingIDsByAccountAndTradingDay(t *testing.T) {
	j := exitFixture(t)
	_, state := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, state, "obs-scope", "67900", "70000", "68000")
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.Proposal = &ExitProposal{Action: string(snapshot.Action), Level: snapshot.Level,
		IntentID: "exit-intent-scope", Provenance: judgement.Provenance}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}
	insertIntent(t, j, "exit-intent-scope")
	insertAttemptWithBrokerOrder(t, j, "exit-attempt-scope", "exit-intent-scope", "COLLIDE-1")
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,
		 time_in_force,quantity,price,currency,source,fingerprint,notes)
		VALUES ('other-account-intent','2026-03-30T00:30:00Z','us','2026-03-30','acct-2','AAPL',
		 'SELL','MARKET','DAY','1',NULL,'USD','exit-policy','fp-other','');
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,broker_order_id,fingerprint,recorded_at)
		VALUES ('other-account-attempt','other-account-intent','PLACE','RECORDED',1,'COLLIDE-1','fp','2026-03-30T00:30:01Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES (?, 'LEGACY_OTHER', 'other-account-intent', '2026-03-30T00:30:00Z')`, state.PositionID); err != nil {
		t.Fatal(err)
	}

	ro := openTestReadOnly(t, j.Path())
	links, err := ro.BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{
		{BrokerOrderID: "COLLIDE-1", AccountRef: "acct-1", TradingDay: "2026-03-30"},
		{BrokerOrderID: "COLLIDE-1", AccountRef: "acct-1", TradingDay: "2026-03-31"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].Ambiguous || links[0].Event.Evaluation.Effective.Snapshot == nil {
		t.Fatalf("account-scoped collision = %+v", links)
	}
	if links[1].Engine || links[1].Event.ID != 0 {
		t.Fatalf("cross-day collision linked: %+v", links[1])
	}
}

func TestBrokerOrderExitLinksFollowsValidatedAmendDescendants(t *testing.T) {
	j := exitFixture(t)
	_, state := openedPosition(t, j, "10")
	snapshot, recovery := ratchetSnapshotForState(t, state, "obs-amend", "67900", "70000", "68000")
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.Proposal = &ExitProposal{Action: string(snapshot.Action), Level: snapshot.Level,
		IntentID: "exit-intent-amend", Provenance: judgement.Provenance}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}
	insertIntent(t, j, "exit-intent-amend")
	insertAttemptWithBrokerOrder(t, j, "exit-attempt-amend", "exit-intent-amend", "O-1")
	insertIntent(t, j, "amend-intent")
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
		VALUES ('amend-attempt','amend-intent','AMEND','RECORDED',1,'O-1','O-2','fp','2026-03-30T00:31:00Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,
		 requested_quantity,intent_id,attempt_id,created_at)
		VALUES ('O-1','O-2','replaces','0','10','amend-intent','amend-attempt','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "O-2", AccountRef: "acct-1", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].Event.Evaluation.Effective.Snapshot == nil ||
		links[0].Event.Evaluation.Effective.Snapshot.Line.DecisionID != snapshot.DecisionID {
		t.Fatalf("amend descendant evidence = %+v", links)
	}
}

func TestBrokerOrderExitLinksFailsClosedForCycleBranchAndCrossAccountEdge(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	j := openTestJournalAt(t, path)
	for _, id := range []string{"i-root-a", "i-root-b", "i-amend-a", "i-amend-b", "i-cycle"} {
		insertIntent(t, j, id)
	}
	insertPosition(t, j, "lineage-position", nil)
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
		VALUES ('a-root-a','i-root-a','PLACE','RECORDED',1,'','ROOT-A','fp','2026-03-30T00:00:00Z'),
		       ('a-root-b','i-root-b','PLACE','RECORDED',1,'','ROOT-B','fp','2026-03-30T00:00:00Z'),
		       ('a-amend-a','i-amend-a','AMEND','RECORDED',1,'ROOT-A','CHILD','fp','2026-03-30T00:01:00Z'),
		       ('a-amend-b','i-amend-b','AMEND','RECORDED',1,'ROOT-B','CHILD','fp','2026-03-30T00:01:00Z'),
		       ('a-cycle','i-cycle','AMEND','RECORDED',1,'CHILD','ROOT-A','fp','2026-03-30T00:02:00Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('lineage-position','LEGACY_A','i-root-a','2026-03-30T00:00:00Z'),
		       ('lineage-position','LEGACY_B','i-root-b','2026-03-30T00:00:01Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
		VALUES ('ROOT-A','CHILD','replaces','0','1','i-amend-a','a-amend-a','2026-03-30T00:01:00Z'),
		       ('ROOT-B','CHILD','replaces','0','1','i-amend-b','a-amend-b','2026-03-30T00:01:00Z'),
		       ('CHILD','ROOT-A','replaces','0','1','i-cycle','a-cycle','2026-03-30T00:02:00Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, path).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "CHILD", AccountRef: "acct-1", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].Event.ID != 0 || links[0].UnknownReason == "" {
		t.Fatalf("unsafe lineage did not fail closed: %+v", links)
	}
}

func TestBrokerOrderExitLinksFailsClosedForCrossAccountLineageEdge(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertPosition(t, j, "cross-account-position", nil)
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,
		 time_in_force,quantity,price,currency,source,fingerprint,notes)
		VALUES ('cross-root-intent','2026-03-30T00:00:00Z','us','2026-03-30','acct-2','AAPL',
		 'SELL','MARKET','DAY','1',NULL,'USD','exit-policy','fp-root',''),
		       ('cross-amend-intent','2026-03-30T00:01:00Z','us','2026-03-30','acct-2','AAPL',
		 'SELL','MARKET','DAY','1',NULL,'USD','exit-policy','fp-amend','');
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
		VALUES ('cross-root-attempt','cross-root-intent','PLACE','RECORDED',1,'','CROSS-ROOT','fp','2026-03-30T00:00:00Z'),
		       ('cross-amend-attempt','cross-amend-intent','AMEND','RECORDED',1,'CROSS-ROOT','CROSS-CHILD','fp','2026-03-30T00:01:00Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('cross-account-position','LEGACY','cross-root-intent','2026-03-30T00:00:00Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
		VALUES ('CROSS-ROOT','CROSS-CHILD','replaces','0','1','cross-amend-intent','cross-amend-attempt','2026-03-30T00:01:00Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "CROSS-CHILD", AccountRef: "acct-1", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].UnknownReason != "lineage_scope_mismatch" || links[0].Event.ID != 0 {
		t.Fatalf("cross-account lineage did not fail closed: %+v", links)
	}
}

func TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ro := &ReadOnly{db: db, path: "missing-schema"}
	if links, err := ro.BrokerOrderExitLinks(context.Background(), nil); err != nil || len(links) != 0 {
		t.Fatalf("empty scope queried missing schema: links=%v err=%v", links, err)
	}
	tooMany := make([]BrokerOrderScope, MaxBrokerOrderEvidenceScopes+1)
	for i := range tooMany {
		tooMany[i] = BrokerOrderScope{BrokerOrderID: "x", AccountRef: "acct", TradingDay: "2026-03-30"}
	}
	if _, err := ro.BrokerOrderExitLinks(context.Background(), tooMany); !errors.Is(err, ErrBrokerOrderEvidenceScope) {
		t.Fatalf("oversized scope error = %v", err)
	}
}
