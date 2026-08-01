package journal

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const schemaV14 = `
CREATE TABLE strategy_decision_lineage (
  decision_identity TEXT PRIMARY KEY,
  candidate_life_id TEXT NOT NULL,
  threshold_version TEXT NOT NULL,
  threshold_set_digest TEXT NOT NULL,
  evidence_digest TEXT NOT NULL,
  lane_id TEXT NOT NULL,
  lane_version TEXT NOT NULL,
  lane_source_digest TEXT NOT NULL,
  lane_constants_digest TEXT NOT NULL,
  activation_manifest_digest TEXT NOT NULL,
  created_at TEXT NOT NULL
) STRICT;

CREATE TABLE strategy_attempt_lineage (
  attempt_id TEXT PRIMARY KEY,
  decision_identity TEXT NOT NULL REFERENCES strategy_decision_lineage(decision_identity),
  risk_intent_id TEXT NOT NULL,
  guardian_decision_id TEXT NOT NULL,
  activation_manifest_digest TEXT NOT NULL,
  client_order_id TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
) STRICT;
CREATE INDEX idx_strategy_attempt_risk_intent ON strategy_attempt_lineage(risk_intent_id);

CREATE TABLE strategy_execution_lineage (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  attempt_id TEXT NOT NULL REFERENCES strategy_attempt_lineage(attempt_id),
  kind TEXT NOT NULL CHECK (kind IN ('BROKER_ORDER','FILL','POSITION')),
  external_ref TEXT NOT NULL,
  recorded_at TEXT NOT NULL,
  UNIQUE(attempt_id, kind, external_ref)
) STRICT;

CREATE TRIGGER strategy_decision_lineage_no_update BEFORE UPDATE ON strategy_decision_lineage BEGIN SELECT RAISE(ABORT,'strategy decision lineage is immutable'); END;
CREATE TRIGGER strategy_decision_lineage_no_delete BEFORE DELETE ON strategy_decision_lineage BEGIN SELECT RAISE(ABORT,'strategy decision lineage is immutable'); END;
CREATE TRIGGER strategy_attempt_lineage_no_update BEFORE UPDATE ON strategy_attempt_lineage BEGIN SELECT RAISE(ABORT,'strategy attempt lineage is immutable'); END;
CREATE TRIGGER strategy_attempt_lineage_no_delete BEFORE DELETE ON strategy_attempt_lineage BEGIN SELECT RAISE(ABORT,'strategy attempt lineage is immutable'); END;
CREATE TRIGGER strategy_execution_lineage_no_update BEFORE UPDATE ON strategy_execution_lineage BEGIN SELECT RAISE(ABORT,'strategy execution lineage is append-only'); END;
CREATE TRIGGER strategy_execution_lineage_no_delete BEFORE DELETE ON strategy_execution_lineage BEGIN SELECT RAISE(ABORT,'strategy execution lineage is append-only'); END;
`

type StrategyDecisionLineage struct {
	DecisionIdentity, CandidateLifeID, ThresholdVersion, ThresholdSetDigest, EvidenceDigest string
	LaneID, LaneVersion, LaneSourceDigest, LaneConstantsDigest, ActivationManifestDigest    string
	CreatedAt                                                                               time.Time
}
type StrategyAttemptLineage struct {
	AttemptID, DecisionIdentity, RiskIntentID, GuardianDecisionID, ActivationManifestDigest, ClientOrderID string
	CreatedAt                                                                                              time.Time
}
type StrategyExecutionLink struct {
	AttemptID, Kind, ExternalRef string
	RecordedAt                   time.Time
}

func (j *Journal) RecordStrategyDecision(ctx context.Context, value StrategyDecisionLineage) error {
	if j == nil || j.db == nil {
		return fmt.Errorf("journal strategy lineage: journal is required")
	}
	value = normalizeStrategyDecision(value, j.clk.Now().UTC())
	if err := validateStrategyDecision(value); err != nil {
		return err
	}
	_, err := j.db.ExecContext(ctx, `INSERT INTO strategy_decision_lineage(decision_identity,candidate_life_id,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,activation_manifest_digest,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, value.DecisionIdentity, value.CandidateLifeID, value.ThresholdVersion, value.ThresholdSetDigest, value.EvidenceDigest, value.LaneID, value.LaneVersion, value.LaneSourceDigest, value.LaneConstantsDigest, value.ActivationManifestDigest, value.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("journal strategy lineage: record decision: %w", err)
	}
	return nil
}
func (j *Journal) RecordStrategyAttempt(ctx context.Context, value StrategyAttemptLineage) error {
	if j == nil || j.db == nil {
		return fmt.Errorf("journal strategy lineage: journal is required")
	}
	if value.CreatedAt.IsZero() {
		value.CreatedAt = j.clk.Now().UTC()
	} else {
		value.CreatedAt = value.CreatedAt.UTC()
	}
	value.AttemptID = strings.TrimSpace(value.AttemptID)
	value.DecisionIdentity = strings.TrimSpace(value.DecisionIdentity)
	value.RiskIntentID = strings.TrimSpace(value.RiskIntentID)
	value.GuardianDecisionID = strings.TrimSpace(value.GuardianDecisionID)
	value.ActivationManifestDigest = strings.TrimSpace(value.ActivationManifestDigest)
	value.ClientOrderID = strings.TrimSpace(value.ClientOrderID)
	if err := validateStrategyAttempt(value); err != nil {
		return err
	}
	_, err := j.db.ExecContext(ctx, `INSERT INTO strategy_attempt_lineage(attempt_id,decision_identity,risk_intent_id,guardian_decision_id,activation_manifest_digest,client_order_id,created_at) VALUES(?,?,?,?,?,?,?)`, value.AttemptID, value.DecisionIdentity, value.RiskIntentID, value.GuardianDecisionID, value.ActivationManifestDigest, value.ClientOrderID, value.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("journal strategy lineage: record attempt: %w", err)
	}
	return nil
}
func (j *Journal) AppendStrategyExecutionLink(ctx context.Context, value StrategyExecutionLink) error {
	if j == nil || j.db == nil {
		return fmt.Errorf("journal strategy lineage: journal is required")
	}
	value.AttemptID = strings.TrimSpace(value.AttemptID)
	value.Kind = strings.TrimSpace(value.Kind)
	value.ExternalRef = strings.TrimSpace(value.ExternalRef)
	if value.RecordedAt.IsZero() {
		value.RecordedAt = j.clk.Now().UTC()
	} else {
		value.RecordedAt = value.RecordedAt.UTC()
	}
	if value.AttemptID == "" || value.ExternalRef == "" || (value.Kind != "BROKER_ORDER" && value.Kind != "FILL" && value.Kind != "POSITION") {
		return fmt.Errorf("journal strategy lineage: invalid execution link")
	}
	_, err := j.db.ExecContext(ctx, `INSERT INTO strategy_execution_lineage(attempt_id,kind,external_ref,recorded_at) VALUES(?,?,?,?)`, value.AttemptID, value.Kind, value.ExternalRef, value.RecordedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("journal strategy lineage: append execution link: %w", err)
	}
	return nil
}
func normalizeStrategyDecision(v StrategyDecisionLineage, now time.Time) StrategyDecisionLineage {
	v.DecisionIdentity = strings.TrimSpace(v.DecisionIdentity)
	v.CandidateLifeID = strings.TrimSpace(v.CandidateLifeID)
	v.ThresholdVersion = strings.TrimSpace(v.ThresholdVersion)
	v.ThresholdSetDigest = strings.TrimSpace(v.ThresholdSetDigest)
	v.EvidenceDigest = strings.TrimSpace(v.EvidenceDigest)
	v.LaneID = strings.TrimSpace(v.LaneID)
	v.LaneVersion = strings.TrimSpace(v.LaneVersion)
	v.LaneSourceDigest = strings.TrimSpace(v.LaneSourceDigest)
	v.LaneConstantsDigest = strings.TrimSpace(v.LaneConstantsDigest)
	v.ActivationManifestDigest = strings.TrimSpace(v.ActivationManifestDigest)
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	} else {
		v.CreatedAt = v.CreatedAt.UTC()
	}
	return v
}
func validateStrategyDecision(v StrategyDecisionLineage) error {
	if v.DecisionIdentity == "" || v.CandidateLifeID == "" || v.ThresholdVersion == "" || v.ThresholdSetDigest == "" || v.EvidenceDigest == "" || v.LaneID == "" || v.LaneVersion == "" || v.LaneSourceDigest == "" || v.LaneConstantsDigest == "" || v.ActivationManifestDigest == "" || v.CreatedAt.IsZero() {
		return fmt.Errorf("journal strategy lineage: complete immutable decision provenance is required")
	}
	return nil
}
func validateStrategyAttempt(v StrategyAttemptLineage) error {
	v.AttemptID = strings.TrimSpace(v.AttemptID)
	v.DecisionIdentity = strings.TrimSpace(v.DecisionIdentity)
	v.RiskIntentID = strings.TrimSpace(v.RiskIntentID)
	v.GuardianDecisionID = strings.TrimSpace(v.GuardianDecisionID)
	v.ActivationManifestDigest = strings.TrimSpace(v.ActivationManifestDigest)
	v.ClientOrderID = strings.TrimSpace(v.ClientOrderID)
	if v.AttemptID == "" || v.DecisionIdentity == "" || v.RiskIntentID == "" || v.GuardianDecisionID == "" || v.ActivationManifestDigest == "" || v.ClientOrderID == "" || v.CreatedAt.IsZero() {
		return fmt.Errorf("journal strategy lineage: complete immutable attempt provenance is required")
	}
	return nil
}
