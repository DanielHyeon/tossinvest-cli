package journal

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStrategyLineageIsImmutableAppendOnlyAndRestartDurable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	ctx := context.Background()
	decision := StrategyDecisionLineage{DecisionIdentity: "decision", CandidateLifeID: "life", ThresholdVersion: "threshold", ThresholdSetDigest: "set", EvidenceDigest: "evidence", LaneID: "lane", LaneVersion: "1", LaneSourceDigest: "source", LaneConstantsDigest: "constants", ActivationManifestDigest: "manifest"}
	if err := j.RecordStrategyDecision(ctx, decision); err != nil {
		t.Fatal(err)
	}
	attempt := StrategyAttemptLineage{AttemptID: "attempt", DecisionIdentity: "decision", RiskIntentID: "risk", GuardianDecisionID: "guardian", ActivationManifestDigest: "manifest", ClientOrderID: "client"}
	if err := j.RecordStrategyAttempt(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	for _, link := range []StrategyExecutionLink{{AttemptID: "attempt", Kind: "BROKER_ORDER", ExternalRef: "broker"}, {AttemptID: "attempt", Kind: "FILL", ExternalRef: "fill"}, {AttemptID: "attempt", Kind: "POSITION", ExternalRef: "position"}} {
		if err := j.AppendStrategyExecutionLink(ctx, link); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := j.db.Exec(`UPDATE strategy_decision_lineage SET lane_version='2' WHERE decision_identity='decision'`); err == nil {
		t.Fatal("decision update accepted")
	}
	if _, err := j.db.Exec(`DELETE FROM strategy_attempt_lineage WHERE attempt_id='attempt'`); err == nil {
		t.Fatal("attempt delete accepted")
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openTestJournalAt(t, path)
	defer reopened.Close()
	var links int
	if err := reopened.db.QueryRow(`SELECT count(*) FROM strategy_execution_lineage WHERE attempt_id='attempt'`).Scan(&links); err != nil || links != 3 {
		t.Fatalf("links=%d err=%v", links, err)
	}
}
func TestStrategyLineageRejectsMissingProvenanceAndDuplicates(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	ctx := context.Background()
	if err := j.RecordStrategyDecision(ctx, StrategyDecisionLineage{}); err == nil {
		t.Fatal("empty decision accepted")
	}
	decision := StrategyDecisionLineage{DecisionIdentity: "decision", CandidateLifeID: "life", ThresholdVersion: "threshold", ThresholdSetDigest: "set", EvidenceDigest: "evidence", LaneID: "lane", LaneVersion: "1", LaneSourceDigest: "source", LaneConstantsDigest: "constants", ActivationManifestDigest: "manifest"}
	if err := j.RecordStrategyDecision(ctx, decision); err != nil {
		t.Fatal(err)
	}
	if err := j.RecordStrategyDecision(ctx, decision); err == nil {
		t.Fatal("duplicate decision accepted")
	}
	if err := j.RecordStrategyAttempt(ctx, StrategyAttemptLineage{}); err == nil {
		t.Fatal("empty attempt accepted")
	}
}
