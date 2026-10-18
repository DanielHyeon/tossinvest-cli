package journal

import (
	"context"
	"fmt"
)

// schemaV36 distinguishes legacy decisions from decisions whose protection
// contract must be recovered from an immutable v4 strategy binding.
const schemaV36 = `
ALTER TABLE decisions ADD COLUMN strategy_binding_version INTEGER NOT NULL DEFAULT 0
 CHECK(strategy_binding_version IN (0, 4));
CREATE TRIGGER decisions_strategy_binding_version_no_update
BEFORE UPDATE OF strategy_binding_version ON decisions
BEGIN SELECT RAISE(ABORT, 'strategy binding version is immutable'); END;
`

func insertV4StrategyDecision(ctx context.Context, ex execer, dec Decision) error {
	if _, err := ex.ExecContext(ctx, `INSERT INTO decisions
		(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,
		 client_order_id,limits_json,nonce,issued_at,expires_at,strategy_binding_version)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,4)`, dec.ID, dec.AccountRef, dec.Generation,
		dec.SafetyClass, dec.PreimageKind, dec.RiskPreimage, dec.RiskHash,
		nullableString(dec.ClientOrderID), nullableString(dec.LimitsJSON), dec.Nonce,
		formatJournalTime(dec.IssuedAt), formatJournalTime(dec.ExpiresAt)); err != nil {
		return fmt.Errorf("journal: recording v4 strategy decision %s: %w", dec.ID, err)
	}
	return nil
}
