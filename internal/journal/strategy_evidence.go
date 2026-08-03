package journal

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// schemaV21 deliberately contains no evidence payload, revision, source response,
// header or credential storage. Those remain exclusively in evidence.db.
const schemaV21 = `
ALTER TABLE strategy_decision_lineage ADD COLUMN consumed_evidence_snapshot_id TEXT;
ALTER TABLE strategy_decision_lineage ADD COLUMN consumed_evidence_snapshot_digest TEXT;
CREATE TRIGGER strategy_evidence_reference_insert_guard BEFORE INSERT ON strategy_decision_lineage
WHEN (NEW.consumed_evidence_snapshot_id IS NULL)!=(NEW.consumed_evidence_snapshot_digest IS NULL)
 OR (NEW.consumed_evidence_snapshot_id IS NOT NULL
     AND (NEW.consumed_evidence_snapshot_id!='snapshot-'||NEW.consumed_evidence_snapshot_digest
          OR length(NEW.consumed_evidence_snapshot_digest)!=64
          OR NEW.consumed_evidence_snapshot_digest!=lower(NEW.consumed_evidence_snapshot_digest)
          OR NEW.consumed_evidence_snapshot_digest GLOB '*[^0-9a-f]*'))
BEGIN SELECT RAISE(ABORT,'invalid consumed evidence snapshot reference'); END;
CREATE TRIGGER strategy_evidence_reference_update_guard BEFORE UPDATE OF consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest ON strategy_decision_lineage
WHEN (NEW.consumed_evidence_snapshot_id IS NULL)!=(NEW.consumed_evidence_snapshot_digest IS NULL)
 OR (NEW.consumed_evidence_snapshot_id IS NOT NULL
     AND (NEW.consumed_evidence_snapshot_id!='snapshot-'||NEW.consumed_evidence_snapshot_digest
          OR length(NEW.consumed_evidence_snapshot_digest)!=64
          OR NEW.consumed_evidence_snapshot_digest!=lower(NEW.consumed_evidence_snapshot_digest)
          OR NEW.consumed_evidence_snapshot_digest GLOB '*[^0-9a-f]*'))
BEGIN SELECT RAISE(ABORT,'invalid consumed evidence snapshot reference'); END;
`

var strategyEvidenceReadOnlyColumns = []struct {
	table, column string
}{
	{table: "strategy_decision_lineage", column: "consumed_evidence_snapshot_id"},
	{table: "strategy_decision_lineage", column: "consumed_evidence_snapshot_digest"},
}

var ErrStrategyEvidenceSnapshotUnavailable = errors.New("journal strategy evidence: consumed snapshot unavailable")

// ConsumedStrategyEvidence is the complete trading-journal view of evidence.
// It intentionally contains no evidence payload or source/revision metadata.
type ConsumedStrategyEvidence struct {
	DecisionIdentity string
	Market           string
	SnapshotID       string
	SnapshotDigest   string
}

// StrategyEvidenceReadBoundary is a dormant SELECT-only adapter over ReadOnly.
// Constructing it neither enables a lane nor connects dispatch, broker, Guardian,
// apply hooks or operating toggles.
type StrategyEvidenceReadBoundary struct{ readonly *ReadOnly }

func NewStrategyEvidenceReadBoundary(readonly *ReadOnly) (*StrategyEvidenceReadBoundary, error) {
	if readonly == nil || readonly.db == nil || readonly.version < 21 {
		return nil, ErrStrategyEvidenceSnapshotUnavailable
	}
	return &StrategyEvidenceReadBoundary{readonly: readonly}, nil
}

func (boundary *StrategyEvidenceReadBoundary) ConsumedSnapshot(ctx context.Context, decisionIdentity string) (ConsumedStrategyEvidence, error) {
	if boundary == nil || boundary.readonly == nil || boundary.readonly.db == nil {
		return ConsumedStrategyEvidence{}, ErrStrategyEvidenceSnapshotUnavailable
	}
	identity := strings.TrimSpace(decisionIdentity)
	if identity == "" {
		return ConsumedStrategyEvidence{}, ErrStrategyTraceNotFound
	}
	var result ConsumedStrategyEvidence
	err := boundary.readonly.db.QueryRowContext(ctx, `SELECT entry_decision_identity,market,COALESCE(consumed_evidence_snapshot_id,''),COALESCE(consumed_evidence_snapshot_digest,'') FROM strategy_decision_lineage WHERE entry_decision_identity=?`, identity).
		Scan(&result.DecisionIdentity, &result.Market, &result.SnapshotID, &result.SnapshotDigest)
	if errors.Is(err, sql.ErrNoRows) {
		return ConsumedStrategyEvidence{}, ErrStrategyTraceNotFound
	}
	if err != nil {
		return ConsumedStrategyEvidence{}, fmt.Errorf("journal strategy evidence: reading consumed snapshot: %w", err)
	}
	if (result.Market != "KR" && result.Market != "US") || !validConsumedEvidenceReference(result.SnapshotID, result.SnapshotDigest) || result.SnapshotID == "" {
		return ConsumedStrategyEvidence{}, ErrStrategyEvidenceSnapshotUnavailable
	}
	return result, nil
}

func validConsumedEvidenceReference(id, digest string) bool {
	if id == "" && digest == "" {
		return true
	}
	if id != strings.TrimSpace(id) || digest != strings.TrimSpace(digest) || len(digest) != 64 || id != "snapshot-"+digest {
		return false
	}
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == 32 && strings.ToLower(digest) == digest
}
