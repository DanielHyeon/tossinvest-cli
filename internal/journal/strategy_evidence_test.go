package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func canonicalSnapshotReference(seed byte) (string, string) {
	digest := strings.Repeat(string(seed), 64)
	return "snapshot-" + digest, digest
}

func TestStrategyEvidenceLineagePersistsOnlyImmutableReference(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "evidence", "acct-1")
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('a')
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	boundary, err := NewStrategyEvidenceReadBoundary(openTestReadOnly(t, j.Path()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := boundary.ConsumedSnapshot(context.Background(), plan.Lineage.DecisionIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if got.DecisionIdentity != plan.Lineage.DecisionIdentity || got.Market != plan.Lineage.Market || got.SnapshotID != plan.Lineage.ConsumedEvidenceSnapshotID || got.SnapshotDigest != plan.Lineage.ConsumedEvidenceSnapshotDigest {
		t.Fatalf("snapshot-only lineage=%+v", got)
	}
	for _, table := range []string{"intents", "mutation_attempts", "risk_reservations"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s count after dormant read=%d err=%v; read boundary mutated execution state", table, count, err)
		}
	}
}

func TestStrategyEvidenceLineageReplayIsExact(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "replay-evidence", "acct-1")
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('b')
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if replay, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil || !replay.Idempotent {
		t.Fatalf("exact replay=%+v err=%v", replay, err)
	}
	changed := plan
	changed.Lineage.ConsumedEvidenceSnapshotID, changed.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('c')
	if _, err := j.planStrategyEntryForTest(context.Background(), changed); err == nil {
		t.Fatal("divergent snapshot replay accepted")
	} else {
		var collision *StrategyCollisionError
		if !errors.As(err, &collision) || collision.Stage != "decision lineage" {
			t.Fatalf("divergent replay error=%v", err)
		}
	}
}

func TestStrategyEvidenceLineageRejectsPartialOrMalformedReference(t *testing.T) {
	validID, validDigest := canonicalSnapshotReference('d')
	for _, test := range []struct {
		name, id, digest string
	}{
		{name: "id-only", id: validID},
		{name: "digest-only", digest: validDigest},
		{name: "wrong-id", id: "snapshot-other", digest: validDigest},
		{name: "wrong-digest", id: validID, digest: strings.Repeat("z", 64)},
		{name: "whitespace", id: validID + " ", digest: validDigest},
		{name: "oversized", id: "snapshot-" + strings.Repeat("a", 1024), digest: validDigest},
	} {
		t.Run(test.name, func(t *testing.T) {
			j := openTestJournal(t)
			plan := strategyPlanFixture(t, "invalid-"+test.name, "acct-1")
			plan.Lineage.ConsumedEvidenceSnapshotID = test.id
			plan.Lineage.ConsumedEvidenceSnapshotDigest = test.digest
			if _, err := j.planStrategyEntryForTest(context.Background(), plan); err == nil {
				t.Fatal("invalid consumed snapshot reference accepted")
			}
			for _, table := range []string{"decisions", "strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage"} {
				var count int
				if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s count=%d err=%v; invalid lineage partially committed", table, count, err)
				}
			}
		})
	}
}

func TestStrategyEvidenceReadBoundaryDistinguishesLegacyAndMissing(t *testing.T) {
	j := openTestJournal(t)
	legacy := strategyPlanFixture(t, "legacy-evidence", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), legacy); err != nil {
		t.Fatal(err)
	}
	boundary, err := NewStrategyEvidenceReadBoundary(openTestReadOnly(t, j.Path()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := boundary.ConsumedSnapshot(context.Background(), legacy.Lineage.DecisionIdentity); !errors.Is(err, ErrStrategyEvidenceSnapshotUnavailable) {
		t.Fatalf("legacy evidence error=%v", err)
	}
	if _, err := boundary.ConsumedSnapshot(context.Background(), "missing"); !errors.Is(err, ErrStrategyTraceNotFound) {
		t.Fatalf("missing decision error=%v", err)
	}
}

func TestStrategyEvidenceSchemaRejectsDirectPartialReference(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "direct-partial-source", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	id, _ := canonicalSnapshotReference('e')
	_, err := j.db.Exec(`INSERT INTO strategy_decision_lineage(
entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest)
SELECT ?,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,?,NULL
FROM strategy_decision_lineage WHERE entry_decision_identity=?`, "direct-partial", id, plan.Lineage.DecisionIdentity)
	if err == nil {
		t.Fatal("database accepted direct partial snapshot reference")
	}
	if _, err := j.db.Exec(`DROP TRIGGER strategy_decision_lineage_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE strategy_decision_lineage SET consumed_evidence_snapshot_id=?,consumed_evidence_snapshot_digest=NULL WHERE entry_decision_identity=?`, id, plan.Lineage.DecisionIdentity); err == nil {
		t.Fatal("database accepted direct partial snapshot reference update")
	}
}

func TestStrategyEvidenceReadBoundaryRejectsUnsupportedMarket(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "unsupported-market", "acct-1")
	plan.Lineage.Market = "cn"
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('f')
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	boundary, err := NewStrategyEvidenceReadBoundary(openTestReadOnly(t, j.Path()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := boundary.ConsumedSnapshot(context.Background(), plan.Lineage.DecisionIdentity); !errors.Is(err, ErrStrategyEvidenceSnapshotUnavailable) {
		t.Fatalf("unsupported market read error=%v", err)
	}
}
