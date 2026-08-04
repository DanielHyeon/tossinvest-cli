package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV26AddsPairedFirstLegAuthorityWithoutChangingV25Rows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 25)
	owner, err := old.AcquireStrategyDispatchOwner(context.Background(), "migration-v26-owner")
	if err != nil {
		t.Fatal(err)
	}
	kr := prepareStrategyDispatchLease(t, old, "migration-v26-kr", owner, StrategyDispatchMarketKR, "005930")
	us := prepareStrategyDispatchLease(t, old, "migration-v26-us", owner, StrategyDispatchMarketUS, "AAPL")
	if err := insertStrategyDispatchLease(old, kr, StrategyDispatchLeaseClaimed, ""); err != nil {
		t.Fatal(err)
	}
	if err := insertStrategyDispatchLease(old, us, StrategyDispatchLeaseIssued, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO strategy_dispatch_outcomes(
		outcome_id,lease_id,from_state,to_state,from_disposition,to_disposition,expected_revision,next_revision,
		transition_code,operation_identity,broker_order_id,query_digest,observed_at,record_digest)
		VALUES(?,?,'ISSUED','CLAIMED','RESERVED','RESERVED',1,2,'CLAIMED',?,'','','2026-03-30T00:30:06Z',?)`,
		"outcome-migration-v26-kr", kr.LeaseID, kr.OperationID, "outcome-record-migration-v26-kr"); err != nil {
		t.Fatal(err)
	}
	beforeRows := journalV25RowFingerprints(t, old, nil)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	planV26 := migrationPlan{steps: migrationsThrough(26), target: 26}
	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &planV26})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	if version, err := current.SchemaVersion(context.Background()); err != nil || version != 26 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	var paired int
	if err := current.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_market_authorities
		WHERE (market='KR' AND symbol='005930' AND record_digest='sealed-authority-record-migration-v26-kr')
		   OR (market='US' AND symbol='AAPL' AND record_digest='sealed-authority-record-migration-v26-us')`).Scan(&paired); err != nil || paired != 2 {
		t.Fatalf("preserved paired authorities=%d err=%v", paired, err)
	}
	var leases, outcomes, qFinal, ownerEpochs, currentOwners int
	for table, target := range map[string]*int{
		"strategy_dispatch_leases": &leases, "strategy_dispatch_outcomes": &outcomes,
		"risk_bucket_final_decisions": &qFinal, "strategy_dispatch_owner_epochs": &ownerEpochs,
		"strategy_dispatch_owner_current": &currentOwners,
	} {
		if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if leases != 2 || outcomes != 1 || qFinal != 2 || ownerEpochs != 1 || currentOwners != 1 {
		t.Fatalf("preserved v25 rows leases=%d outcomes=%d q_final=%d owner_epochs=%d owner_current=%d",
			leases, outcomes, qFinal, ownerEpochs, currentOwners)
	}
	var krState string
	var krRevision int
	if err := current.db.QueryRow(`SELECT state,revision FROM strategy_dispatch_leases WHERE lease_id=?`, kr.LeaseID).Scan(&krState, &krRevision); err != nil || krState != "CLAIMED" || krRevision != 2 {
		t.Fatalf("preserved KR lease state=%s revision=%d err=%v", krState, krRevision, err)
	}
	afterRows := journalV25RowFingerprints(t, current, mapKeys(beforeRows))
	for table, want := range beforeRows {
		if got := afterRows[table]; got != want {
			t.Fatalf("v25 table %s rows changed across v26 migration: before=%s after=%s", table, want, got)
		}
	}
	for _, object := range []struct{ kind, name string }{
		{"table", "strategy_first_leg_bindings"},
		{"trigger", "strategy_first_leg_binding_insert_guard"},
		{"trigger", "strategy_first_leg_bindings_no_update"},
		{"trigger", "strategy_first_leg_bindings_no_delete"},
		{"trigger", "strategy_dispatch_lease_first_leg_binding"},
	} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type=? AND name=?`, object.kind, object.name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("%s %s count=%d err=%v", object.kind, object.name, count, err)
		}
	}
}

func TestMigrationV26FailureRollsBackBindingAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 25)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(25), migration{Version: 26, SQL: schemaV26 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 26}})
	if err == nil {
		t.Fatal("broken v26 migration succeeded")
	}
	raw, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version, bindings int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 25 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master
		WHERE name LIKE 'strategy_first_leg_%' OR name='strategy_dispatch_lease_first_leg_binding'`).Scan(&bindings); err != nil || bindings != 0 {
		t.Fatalf("rolled-back first-leg objects=%d err=%v", bindings, err)
	}
}

func TestReleasedV25BuildRefusesV26Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	current := openTestJournalAt(t, path)
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(25), target: 25}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("v25 build open error=%v", err)
	}
}

func journalV25RowFingerprints(t *testing.T, j *Journal, tables []string) map[string]string {
	t.Helper()
	if tables == nil {
		rows, err := j.db.Query(`SELECT name FROM sqlite_master
			WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT IN ('schema_meta','schema_migration_events')
			ORDER BY name`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			tables = append(tables, table)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
	}
	fingerprints := make(map[string]string, len(tables))
	for _, table := range tables {
		quoted := `"` + strings.ReplaceAll(table, `"`, `""`) + `"`
		rows, err := j.db.Query(`SELECT * FROM ` + quoted + ` ORDER BY rowid`)
		if err != nil {
			t.Fatalf("snapshot %s: %v", table, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			t.Fatal(err)
		}
		hash := sha256.New()
		writeFingerprintPart := func(kind string, raw []byte) {
			_, _ = hash.Write([]byte(kind + ":" + strconv.Itoa(len(raw)) + ":"))
			_, _ = hash.Write(raw)
		}
		for _, column := range columns {
			writeFingerprintPart("column", []byte(column))
		}
		for rows.Next() {
			values := make([]any, len(columns))
			targets := make([]any, len(columns))
			for index := range values {
				targets[index] = &values[index]
			}
			if err := rows.Scan(targets...); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			writeFingerprintPart("row", nil)
			for _, value := range values {
				switch typed := value.(type) {
				case nil:
					writeFingerprintPart("null", nil)
				case int64:
					writeFingerprintPart("int64", []byte(strconv.FormatInt(typed, 10)))
				case float64:
					writeFingerprintPart("float64", []byte(strconv.FormatFloat(typed, 'g', -1, 64)))
				case string:
					writeFingerprintPart("string", []byte(typed))
				case []byte:
					writeFingerprintPart("bytes", typed)
				default:
					rows.Close()
					t.Fatalf("snapshot %s unsupported %T (%v)", table, value, value)
				}
			}
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
		fingerprints[table] = hex.EncodeToString(hash.Sum(nil))
	}
	return fingerprints
}

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
