package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

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
		[]BrokerOrderScope{{BrokerOrderID: "broker-exit-view", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
			{BrokerOrderID: "same-symbol-same-time", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
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

func TestBrokerOrderExitLinksScopesCollidingIDsByMarket(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertPosition(t, j, "market-position", nil)
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,
		 time_in_force,quantity,price,currency,source,fingerprint,notes)
		VALUES ('market-us','2026-03-30T00:00:00Z','us','2026-03-30','acct-1','AAPL','SELL','MARKET','DAY','1',NULL,'USD','exit-policy','fp-us',''),
		       ('market-kr','2026-03-30T00:00:00Z','kr','2026-03-30','acct-1','005930','SELL','MARKET','DAY','1',NULL,'KRW','exit-policy','fp-kr','');
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,broker_order_id,fingerprint,recorded_at)
		VALUES ('market-us-a','market-us','PLACE','CONFIRMED',1,'SAME-ID','fp','2026-03-30T00:00:00Z'),
		       ('market-kr-a','market-kr','PLACE','CONFIRMED',1,'SAME-ID','fp','2026-03-30T00:00:00Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('market-position','US_EXIT','market-us','2026-03-30T00:00:00Z'),
		       ('market-position','KR_EXIT','market-kr','2026-03-30T00:00:01Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{
		{BrokerOrderID: "SAME-ID", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
		{BrokerOrderID: "SAME-ID", AccountRef: "acct-1", Market: "kr", TradingDay: "2026-03-30"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].IntentID != "market-us" || links[0].Event.Action != "US_EXIT" ||
		links[1].IntentID != "market-kr" || links[1].Event.Action != "KR_EXIT" {
		t.Fatalf("market collision linked across scope: %+v", links)
	}
}

func TestBrokerOrderExitLinksTreatsBrokerOrderIDsAsOpaqueBytes(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertPosition(t, j, "opaque-position", nil)
	insertIntent(t, j, "opaque-spaced-intent")
	insertIntent(t, j, "opaque-plain-intent")
	insertAttemptWithBrokerOrder(t, j, "opaque-spaced-attempt", "opaque-spaced-intent", " O-1 ")
	insertAttemptWithBrokerOrder(t, j, "opaque-plain-attempt", "opaque-plain-intent", "O-1")
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('opaque-position','SPACED','opaque-spaced-intent','2026-03-30T00:00:00Z'),
		       ('opaque-position','PLAIN','opaque-plain-intent','2026-03-30T00:00:01Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{
		{BrokerOrderID: " O-1 ", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
		{BrokerOrderID: "O-1", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].BrokerOrderID != " O-1 " || links[0].IntentID != "opaque-spaced-intent" ||
		links[0].Event.Action != "SPACED" || !links[0].Engine ||
		links[1].BrokerOrderID != "O-1" || links[1].IntentID != "opaque-plain-intent" ||
		links[1].Event.Action != "PLAIN" || !links[1].Engine {
		t.Fatalf("opaque broker order IDs were canonicalized or cross-linked: %+v", links)
	}

	if _, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{
		{BrokerOrderID: "   ", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
	}); err != nil {
		t.Fatalf("whitespace-only opaque broker order ID was rejected: %v", err)
	}
}

func TestBrokerOrderExitLinksAttributesOnlyConfirmedPlaceAttempts(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	for _, id := range []string{"kind-cancel", "kind-amend", "state-recorded", "state-failed"} {
		insertIntent(t, j, id)
	}
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,broker_order_id,fingerprint,recorded_at)
		VALUES ('kind-cancel-a','kind-cancel','CANCEL','CONFIRMED',1,'CANCEL-ONLY','fp','2026-03-30T00:00:00Z'),
		       ('kind-amend-a','kind-amend','AMEND','CONFIRMED',1,'AMEND-ONLY','fp','2026-03-30T00:00:00Z'),
		       ('state-recorded-a','state-recorded','PLACE','RECORDED',1,'RECORDED-ONLY','fp','2026-03-30T00:00:00Z'),
		       ('state-failed-a','state-failed','PLACE','FAILED_CONFIRMED',1,'FAILED-ONLY','fp','2026-03-30T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	var scopes []BrokerOrderScope
	for _, id := range []string{"CANCEL-ONLY", "AMEND-ONLY", "RECORDED-ONLY", "FAILED-ONLY"} {
		scopes = append(scopes, BrokerOrderScope{BrokerOrderID: id, AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"})
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(), scopes)
	if err != nil {
		t.Fatal(err)
	}
	for _, link := range links {
		if link.Engine || link.Event.ID != 0 {
			t.Errorf("non-confirmed PLACE attributed as engine: %+v", link)
		}
	}
}

func TestBrokerOrderExitLinksRefusesIncompleteAmendEvidence(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertPosition(t, j, "incomplete-position", nil)
	insertIntent(t, j, "incomplete-root")
	insertIntent(t, j, "incomplete-amend")
	insertAttemptWithBrokerOrder(t, j, "incomplete-root-a", "incomplete-root", "I-ROOT")
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
		VALUES ('incomplete-amend-a','incomplete-amend','AMEND','RECORDED',1,'I-ROOT','I-CHILD','fp','2026-03-30T00:01:00Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('incomplete-position','EXIT','incomplete-root','2026-03-30T00:00:00Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
		VALUES ('I-ROOT','I-CHILD','replaces','0','1','incomplete-amend','incomplete-amend-a','2026-03-30T00:01:00Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "I-CHILD", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].Event.ID != 0 || links[0].UnknownReason != "lineage_scope_mismatch" {
		t.Fatalf("incomplete AMEND lineage accepted: %+v", links)
	}
}

func TestBrokerOrderExitLinksDoesNotExpandWideBranchingLineage(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		intent, attempt, parent := fmt.Sprintf("wide-i-%d", i), fmt.Sprintf("wide-a-%d", i), fmt.Sprintf("WIDE-P-%d", i)
		if _, err := tx.Exec(`INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,quantity,price,currency,source,fingerprint,notes)
			VALUES (?,'2026-03-30T00:00:00Z','us','2026-03-30','acct-1','AAPL','SELL','MARKET','DAY','1',NULL,'USD','test',?,'')`, intent, intent); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(`INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
			VALUES (?,?,'AMEND','CONFIRMED',1,?,'WIDE-CHILD',?,'2026-03-30T00:00:00Z')`, attempt, intent, parent, attempt); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(`INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
			VALUES (?,'WIDE-CHILD','replaces','0','1',?,?,'2026-03-30T00:00:00Z')`, parent, intent, attempt); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(ctx,
		[]BrokerOrderScope{{BrokerOrderID: "WIDE-CHILD", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].UnknownReason != "lineage_ambiguous" {
		t.Fatalf("wide branching did not stop before recursion: %+v", links)
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
		[]BrokerOrderScope{{BrokerOrderID: "broker-ambiguous", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
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
		{BrokerOrderID: "COLLIDE-1", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
		{BrokerOrderID: "COLLIDE-1", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-31"},
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
	amendIntent := testIntent()
	amendIntent.ID, amendIntent.Fingerprint = "amend-intent", "fp-amend"
	amend, err := j.Prepare(context.Background(), PrepareRequest{Intent: amendIntent, Kind: KindAmend,
		AttemptID: "amend-attempt", TargetOrderID: "O-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkDispatchStarted(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkAcked(context.Background(), "O-2"); err != nil {
		t.Fatal(err)
	}
	if err := amend.ResolveConfirmedWithLineage(context.Background(), LineageEdge{
		ParentOrderID: "O-1", ChildOrderID: "O-2", Relation: RelationReplaces,
		ParentFilledQuantity: "0", RequestedQuantity: "10",
	}, "", "production lifecycle fixture"); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "O-2", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Engine || links[0].Event.Evaluation.Effective.Snapshot == nil ||
		links[0].Event.Evaluation.Effective.Snapshot.Line.DecisionID != snapshot.DecisionID {
		t.Fatalf("amend descendant evidence = %+v", links)
	}
}

func TestBrokerOrderExitLinksUsesCollisionFreeOpaqueLineagePaths(t *testing.T) {
	for _, tc := range []struct {
		name, parent, child string
	}{
		{name: "delimiter token is not a cycle", parent: "ROOT", child: "PREFIX|ROOT"},
		{name: "JSON escaping preserves exact bytes", parent: "ROOT|\"quoted\"\nline", child: "CHILD|\"quoted\"\tline"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
			insertPosition(t, j, "opaque-lineage-position", nil)
			insertIntent(t, j, "opaque-root-intent")
			insertIntent(t, j, "opaque-amend-intent")
			insertAttemptWithBrokerOrder(t, j, "opaque-root-attempt", "opaque-root-intent", tc.parent)
			if _, err := j.db.ExecContext(context.Background(), `
				INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
				VALUES ('opaque-amend-attempt','opaque-amend-intent','AMEND','CONFIRMED',1,?,?,'fp','2026-03-30T00:01:00Z');
				INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
				VALUES ('opaque-lineage-position','OPAQUE_EXIT','opaque-root-intent','2026-03-30T00:00:00Z');
				INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
				VALUES (?,?,'replaces','0','1','opaque-amend-intent','opaque-amend-attempt','2026-03-30T00:01:00Z')`,
				tc.parent, tc.child, tc.parent, tc.child); err != nil {
				t.Fatal(err)
			}
			links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
				[]BrokerOrderScope{{BrokerOrderID: tc.child, AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
			if err != nil {
				t.Fatal(err)
			}
			if len(links) != 1 || links[0].Ambiguous || !links[0].Engine || links[0].Event.Action != "OPAQUE_EXIT" {
				t.Fatalf("opaque lineage path was corrupted: %+v", links)
			}
		})
	}
}

func TestBrokerOrderExitLinksFailsClosedForPureSingleParentCycle(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertIntent(t, j, "cycle-a-intent")
	insertIntent(t, j, "cycle-b-intent")
	cycleA, cycleB := "CYCLE|A", "CYCLE|\"B\"\n"
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
		VALUES ('cycle-a-attempt','cycle-a-intent','AMEND','CONFIRMED',1,?,?,'fp','2026-03-30T00:01:00Z'),
		       ('cycle-b-attempt','cycle-b-intent','AMEND','CONFIRMED',1,?,?,'fp','2026-03-30T00:02:00Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
		VALUES (?,?,'replaces','0','1','cycle-a-intent','cycle-a-attempt','2026-03-30T00:01:00Z'),
		       (?,?,'replaces','0','1','cycle-b-intent','cycle-b-attempt','2026-03-30T00:02:00Z')`,
		cycleA, cycleB, cycleB, cycleA, cycleA, cycleB, cycleB, cycleA); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: cycleB, AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || !links[0].Ambiguous || links[0].UnknownReason != "lineage_cycle" || links[0].Event.ID != 0 {
		t.Fatalf("pure single-parent cycle did not fail closed: %+v", links)
	}
}

func TestBrokerOrderExitLinksFailsClosedOnlyWhenLineageExceedsDepthBound(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), DBFileName))
	insertPosition(t, j, "depth-position", nil)
	insertIntent(t, j, "depth-root-intent")
	insertAttemptWithBrokerOrder(t, j, "depth-root-attempt", "depth-root-intent", "DEPTH-0")
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('depth-position','DEPTH_EXIT','depth-root-intent','2026-03-30T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= maxBrokerOrderLineageDepth+1; i++ {
		intentID, attemptID := fmt.Sprintf("depth-intent-%d", i), fmt.Sprintf("depth-attempt-%d", i)
		parent, child := fmt.Sprintf("DEPTH-%d", i-1), fmt.Sprintf("DEPTH-%d", i)
		insertIntent(t, j, intentID)
		if _, err := j.db.ExecContext(context.Background(), `
			INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,target_order_id,broker_order_id,fingerprint,recorded_at)
			VALUES (?,?,'AMEND','CONFIRMED',1,?,?,'fp','2026-03-30T00:01:00Z')`,
			attemptID, intentID, parent, child); err != nil {
			t.Fatal(err)
		}
		if _, err := j.db.ExecContext(context.Background(), `
			INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
			VALUES (?,?,'replaces','0','1',?,?,'2026-03-30T00:01:00Z')`,
			parent, child, intentID, attemptID); err != nil {
			t.Fatal(err)
		}
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{
			{BrokerOrderID: fmt.Sprintf("DEPTH-%d", maxBrokerOrderLineageDepth), AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
			{BrokerOrderID: fmt.Sprintf("DEPTH-%d", maxBrokerOrderLineageDepth+1), AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"},
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].Ambiguous || !links[0].Engine || links[0].Event.Action != "DEPTH_EXIT" {
		t.Fatalf("lineage exactly at depth bound did not remain linked: %+v", links)
	}
	if !links[1].Ambiguous || links[1].UnknownReason != "lineage_depth_exceeded" || links[1].Event.ID != 0 {
		t.Fatalf("lineage beyond depth bound did not fail closed: %+v", links[1])
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
		VALUES ('a-root-a','i-root-a','PLACE','CONFIRMED',1,'','ROOT-A','fp','2026-03-30T00:00:00Z'),
		       ('a-root-b','i-root-b','PLACE','CONFIRMED',1,'','ROOT-B','fp','2026-03-30T00:00:00Z'),
		       ('a-amend-a','i-amend-a','AMEND','CONFIRMED',1,'ROOT-A','CHILD','fp','2026-03-30T00:01:00Z'),
		       ('a-amend-b','i-amend-b','AMEND','CONFIRMED',1,'ROOT-B','CHILD','fp','2026-03-30T00:01:00Z'),
		       ('a-cycle','i-cycle','AMEND','CONFIRMED',1,'CHILD','ROOT-A','fp','2026-03-30T00:02:00Z');
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
		[]BrokerOrderScope{{BrokerOrderID: "CHILD", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
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
		VALUES ('cross-root-attempt','cross-root-intent','PLACE','CONFIRMED',1,'','CROSS-ROOT','fp','2026-03-30T00:00:00Z'),
		       ('cross-amend-attempt','cross-amend-intent','AMEND','CONFIRMED',1,'CROSS-ROOT','CROSS-CHILD','fp','2026-03-30T00:01:00Z');
		INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
		VALUES ('cross-account-position','LEGACY','cross-root-intent','2026-03-30T00:00:00Z');
		INSERT INTO lineage_edges(parent_order_id,child_order_id,relation,parent_filled_quantity,requested_quantity,intent_id,attempt_id,created_at)
		VALUES ('CROSS-ROOT','CROSS-CHILD','replaces','0','1','cross-amend-intent','cross-amend-attempt','2026-03-30T00:01:00Z')`); err != nil {
		t.Fatal(err)
	}
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(),
		[]BrokerOrderScope{{BrokerOrderID: "CROSS-CHILD", AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}})
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
	if _, err := ro.BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{{
		BrokerOrderID: "x", AccountRef: "acct", Market: "crypto", TradingDay: "2026-03-30",
	}}); !errors.Is(err, ErrBrokerOrderEvidenceScope) {
		t.Fatalf("invalid market scope error = %v", err)
	}
	tooMany := make([]BrokerOrderScope, MaxBrokerOrderEvidenceScopes+1)
	for i := range tooMany {
		tooMany[i] = BrokerOrderScope{BrokerOrderID: "x", AccountRef: "acct", Market: "us", TradingDay: "2026-03-30"}
	}
	if _, err := ro.BrokerOrderExitLinks(context.Background(), tooMany); !errors.Is(err, ErrBrokerOrderEvidenceScope) {
		t.Fatalf("oversized scope error = %v", err)
	}
}
