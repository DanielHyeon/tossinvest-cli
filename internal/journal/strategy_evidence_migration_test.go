package journal

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV21AddsOnlyNullableConsumedSnapshotReference(t *testing.T) {
	j := openJournalAtSchema(t, filepath.Join(t.TempDir(), "journal.db"), 21)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 21 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	columns := map[string]struct {
		notNull int
		kind    string
	}{}
	rows, err := j.db.Query(`PRAGMA table_info(strategy_decision_lineage)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var ordinal, notNull, primaryKey int
		var name, kind string
		var defaultValue any
		if err := rows.Scan(&ordinal, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		columns[name] = struct {
			notNull int
			kind    string
		}{notNull: notNull, kind: kind}
	}
	for _, name := range []string{"consumed_evidence_snapshot_id", "consumed_evidence_snapshot_digest"} {
		column, ok := columns[name]
		if !ok || column.notNull != 0 || column.kind != "TEXT" {
			t.Fatalf("v21 column %s=%+v present=%v; want nullable TEXT", name, column, ok)
		}
	}
	for _, name := range []string{"strategy_evidence_reference_insert_guard", "strategy_evidence_reference_update_guard"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name=?`, name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v21 trigger %s count=%d err=%v", name, count, err)
		}
	}
	lower := strings.ToLower(schemaV21)
	for _, forbidden := range []string{"payload", "revision", "credential", "secret", "source_response", "create table"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("v21 trading-journal migration contains forbidden evidence storage %q: %s", forbidden, schemaV21)
		}
	}
}

func TestMigrationV21PreservesV20RowsWithoutInventingSnapshotLineage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 20)
	plan := strategyPlanFixture(t, "v21-legacy", "acct-1")
	lineage := plan.Lineage
	if _, err := old.db.Exec(`INSERT INTO strategy_decision_lineage(entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		lineage.DecisionIdentity, lineage.CandidateLifeID, lineage.Market, lineage.Symbol, lineage.ThresholdVersion, lineage.ThresholdSetDigest, lineage.EvidenceDigest,
		lineage.LaneID, lineage.LaneVersion, lineage.LaneSourceDigest, lineage.LaneConstantsDigest, lineage.EntryPrice, lineage.StopPrice, lineage.TargetPrice,
		lineage.Quantity, lineage.PolicyVersion, lineage.SettingsDigest, lineage.DecisionPayload, lineage.DecisionPayloadDigest, lineage.ActivationManifestDigest,
		lineage.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	var id, digest sql.NullString
	if err := current.db.QueryRow(`SELECT consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest FROM strategy_decision_lineage WHERE entry_decision_identity=?`, plan.Lineage.DecisionIdentity).Scan(&id, &digest); err != nil {
		t.Fatal(err)
	}
	if id.Valid || digest.Valid {
		t.Fatalf("migration backfilled unverifiable snapshot lineage: id=%+v digest=%+v", id, digest)
	}
}

func TestMigrationV21FailureRollsBackBothColumnsAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 20)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(20), migration{Version: 21, SQL: schemaV21 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 21}})
	if err == nil {
		t.Fatal("broken v21 migration succeeded")
	}
	raw, openErr := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 20 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	for _, column := range []string{"consumed_evidence_snapshot_id", "consumed_evidence_snapshot_digest"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM pragma_table_info('strategy_decision_lineage') WHERE name=?`, column).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back column %s count=%d err=%v", column, count, err)
		}
	}
	for _, trigger := range []string{"strategy_evidence_reference_insert_guard", "strategy_evidence_reference_update_guard"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name=?`, trigger).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back trigger %s count=%d err=%v", trigger, count, err)
		}
	}
}

func TestOpenReadOnlyRejectsDamagedV21EvidenceLineage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 20)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`PRAGMA user_version=21`); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrSchemaTooOld) || !strings.Contains(err.Error(), "consumed_evidence_snapshot") {
		t.Fatalf("damaged v21 read-only schema error=%v", err)
	}
}
