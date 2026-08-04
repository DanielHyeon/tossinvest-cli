package journal

import (
	"context"
	"strings"
	"testing"
)

func TestMigrationV28InstallsExactDispatchAuthorityUpdateGuard(t *testing.T) {
	j := openTestJournal(t)
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != SchemaVersion {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	var sqlText string
	err := j.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='trigger' AND name='strategy_dispatch_lease_authority_immutable_v28'`).Scan(&sqlText)
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"operation_id", "account_ref", "candidate_id", "campaign_id", "risk_reservation_id", "guardian_decision_id", "owner_epoch", "authority_digest", "lease_digest", "created_at"} {
		if !strings.Contains(sqlText, "NEW."+column+" IS NOT OLD."+column) {
			t.Fatalf("v28 guard does not freeze %s", column)
		}
	}
}
