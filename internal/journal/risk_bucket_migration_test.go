package journal

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestMigrationV22AddsOnlyAuthoritativeRiskBucketJournalState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openJournalAtSchema(t, path, 22)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 22 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	for _, table := range []string{
		"risk_bucket_policies", "risk_bucket_snapshots", "risk_bucket_final_decisions",
		"risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_orders",
		"risk_bucket_fills", "risk_bucket_fill_allocations", "risk_bucket_events",
		"risk_bucket_state_snapshots", "risk_bucket_scope_latches",
	} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v22 table %s count=%d err=%v", table, count, err)
		}
	}
}

func TestMigrationV23AddsScopedFillTablesAndPreservesReleasedV22LegacyShapes(t *testing.T) {
	j := openJournalAtSchema(t, filepath.Join(t.TempDir(), "journal.db"), 23)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 23 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	for _, table := range []string{
		"risk_bucket_policies", "risk_bucket_snapshots", "risk_bucket_final_decisions",
		"risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_orders",
		"risk_bucket_order_reservations", "risk_bucket_fills", "risk_bucket_fill_allocations",
		"risk_bucket_fill_actual_evidence", "risk_bucket_events", "risk_bucket_state_snapshots",
		"risk_bucket_scope_latches", "risk_bucket_orders_v22", "risk_bucket_fills_v22",
		"risk_bucket_fill_allocations_v22",
	} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v23 table %s count=%d err=%v", table, count, err)
		}
	}
}

func TestMigrationV24AddsOfficialZeroAuthorityAndReleaseReceiptsWithoutRewritingV23(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 23)
	if _, err := old.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause)
		VALUES('legacy-v23-reconcile','acct-v23','AAPL','QUANTITY_MISMATCH','legacy freeform evidence',
		'2026-03-30T00:00:00Z','2026-03-30T00:01:00Z','RECHECK_MATCHED')`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause)
		VALUES('legacy-v23-active','acct-v23','MSFT','QUANTITY_MISMATCH','legacy active global scope',
		'2026-03-30T00:02:00Z',NULL,NULL)`); err != nil {
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
	if version, err := current.SchemaVersion(context.Background()); err != nil || version != 24 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	for _, table := range []string{"risk_bucket_broker_zero_observations", "risk_bucket_owner_release_receipts"} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v24 table %s count=%d err=%v", table, count, err)
		}
	}
	for _, column := range []string{"broker_zero_observation_id", "broker_zero_observation_digest", "scope_market"} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM pragma_table_info('reconcile_states') WHERE name=?`, column).Scan(&count); err != nil || count != 1 {
			t.Fatalf("v24 reconcile column %s count=%d err=%v", column, count, err)
		}
	}
	var linkedID, linkedDigest sql.NullString
	if err := current.db.QueryRow(`SELECT broker_zero_observation_id,broker_zero_observation_digest FROM reconcile_states WHERE id='legacy-v23-reconcile'`).Scan(&linkedID, &linkedDigest); err != nil {
		t.Fatal(err)
	}
	if linkedID.Valid || linkedDigest.Valid {
		t.Fatalf("v24 migration invented authority for v23 row id=%v digest=%v", linkedID, linkedDigest)
	}
	var scopeMarket sql.NullString
	if err := current.db.QueryRow(`SELECT scope_market FROM reconcile_states WHERE id='legacy-v23-reconcile'`).Scan(&scopeMarket); err != nil || scopeMarket.Valid {
		t.Fatalf("v24 migration invented market scope=%v err=%v", scopeMarket, err)
	}
	for name, want := range map[string]int{
		"idx_reconcile_active":               0,
		"idx_reconcile_active_account_wide":  1,
		"idx_reconcile_active_legacy_symbol": 1,
		"idx_reconcile_active_market_symbol": 1,
	} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name=?`, name).Scan(&count); err != nil || count != want {
			t.Fatalf("v24 index %s count=%d want=%d err=%v", name, count, want, err)
		}
	}
	if err := current.db.QueryRow(`SELECT scope_market FROM reconcile_states WHERE id='legacy-v23-active'`).Scan(&scopeMarket); err != nil || scopeMarket.Valid {
		t.Fatalf("v24 active legacy row lost global NULL scope=%v err=%v", scopeMarket, err)
	}
	if _, err := current.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,scope_market)
		VALUES('market-overlaps-global','acct-v23','MSFT','QUANTITY_MISMATCH','US overlap','2026-03-30T00:03:00Z','US')`); err == nil {
		t.Fatal("v24 accepted exact-market row overlapping migrated global active row")
	}
	for _, market := range []string{"KR", "US"} {
		if _, err := current.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,scope_market)
			VALUES(?,?,?,?,?,'2026-03-30T00:04:00Z',?)`, "exact-"+market, "acct-exact", "AAPL",
			ReconcileCauseQuantityMismatch, market+" exact", market); err != nil {
			t.Fatalf("v24 exact %s insert: %v", market, err)
		}
	}
	if _, err := current.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at)
		VALUES('global-overlaps-exact','acct-exact','AAPL','QUANTITY_MISMATCH','global overlap','2026-03-30T00:05:00Z')`); err == nil {
		t.Fatal("v24 accepted global row overlapping active exact-market rows")
	}
	var triggers int
	if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name IN (
		'risk_bucket_broker_zero_observations_no_update','risk_bucket_broker_zero_observations_no_delete',
		'risk_bucket_owner_release_receipts_no_update','risk_bucket_owner_release_receipts_no_delete')`).Scan(&triggers); err != nil || triggers != 4 {
		t.Fatalf("v24 immutable triggers=%d err=%v", triggers, err)
	}
	var scopeTriggers int
	if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name IN (
		'reconcile_active_scope_no_overlap_before_insert','reconcile_active_scope_no_overlap_before_update')`).Scan(&scopeTriggers); err != nil || scopeTriggers != 2 {
		t.Fatalf("v24 reconcile overlap triggers=%d err=%v", scopeTriggers, err)
	}
	if _, err := current.db.Exec(`UPDATE reconcile_states SET evidence='still legacy' WHERE id='legacy-v23-reconcile'`); err != nil {
		t.Fatalf("v24 additive columns unexpectedly rewrote legacy reconcile row: %v", err)
	}
}

func TestMigrationV24FailureRollsBackTablesColumnsAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 23)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(23), migration{Version: 24, SQL: schemaV24 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 24}})
	if err == nil {
		t.Fatal("broken v24 migration succeeded")
	}
	raw, openErr := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 23 {
		t.Fatalf("version after v24 rollback=%d err=%v", version, err)
	}
	for _, table := range []string{"risk_bucket_broker_zero_observations", "risk_bucket_owner_release_receipts"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back v24 table %s count=%d err=%v", table, count, err)
		}
	}
	for _, column := range []string{"broker_zero_observation_id", "broker_zero_observation_digest", "scope_market"} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM pragma_table_info('reconcile_states') WHERE name=?`, column).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back v24 column %s count=%d err=%v", column, count, err)
		}
	}
	for name, want := range map[string]int{
		"idx_reconcile_active":               1,
		"idx_reconcile_active_legacy_symbol": 0,
		"idx_reconcile_active_market_symbol": 0,
	} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name=?`, name).Scan(&count); err != nil || count != want {
			t.Fatalf("rolled-back v24 index %s count=%d want=%d err=%v", name, count, want, err)
		}
	}
	var scopeTriggers int
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name LIKE 'reconcile_active_scope_no_overlap_%'`).Scan(&scopeTriggers); err != nil || scopeTriggers != 0 {
		t.Fatalf("rolled-back v24 reconcile triggers=%d err=%v", scopeTriggers, err)
	}
}

func TestReleasedV23BuildRefusesV24Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	current := openTestJournalAt(t, path)
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(23), target: 23}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("v23 build open error=%v", err)
	}
}

func TestMigrationV22PreservesV21AsUnknownWithoutBackfill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 21)
	seedExistingRiskReservation(t, old, "legacy-existing", "acct-legacy")
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(22), target: 22}})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations", "risk_bucket_events"} {
		var count int
		if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("legacy migration invented %s rows=%d err=%v", table, count, err)
		}
	}
	if _, err := current.ReadRiskBucketState(context.Background(), riskBucketOwnerKey("acct-legacy", "legacy")); !errors.Is(err, ErrRiskBucketStateUnknown) {
		t.Fatalf("legacy state error=%v", err)
	}
}

func TestMigrationV22FailureRollsBackEveryTableAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 21)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(21), migration{Version: 22, SQL: schemaV22 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 22}})
	if err == nil {
		t.Fatal("broken v22 migration succeeded")
	}
	raw, openErr := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 21 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	var count int
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name LIKE 'risk_bucket_%'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back v22 artifacts=%d err=%v", count, err)
	}
}

func TestMigrationV22ToV23PreservesLegacyRowsWithoutPromotingThemToAuthority(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 22)
	seedReleasedV22RiskRows(t, old)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	current, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(23), target: 23}})
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	if version, err := current.SchemaVersion(context.Background()); err != nil || version != 23 {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	backups := backupsIn(t, filepath.Dir(path))
	if len(backups) != 1 {
		t.Fatalf("v22→v23 backups=%v", backups)
	}
	assertBackupAtVersion(t, backups[0], 22, map[string]int{
		"risk_bucket_orders": 1, "risk_bucket_fills": 1, "risk_bucket_fill_allocations": 1,
	}, "risk_bucket_order_reservations")
	for table, want := range map[string]int{
		"risk_bucket_final_decisions":      1,
		"risk_bucket_owners":               1,
		"risk_bucket_reservations":         1,
		"risk_bucket_orders_v22":           1,
		"risk_bucket_fills_v22":            1,
		"risk_bucket_fill_allocations_v22": 1,
		"risk_bucket_orders":               0,
		"risk_bucket_order_reservations":   0,
		"risk_bucket_fills":                0,
		"risk_bucket_fill_allocations":     0,
		"risk_bucket_fill_actual_evidence": 0,
	} {
		var got int
		if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&got); err != nil || got != want {
			t.Fatalf("table %s rows=%d want=%d err=%v", table, got, want, err)
		}
	}
	var orderID, predecessor, policyDigest string
	var quantity, cumulative uint64
	if err := current.db.QueryRow(`SELECT order_id,predecessor_order_id,order_quantity,cumulative_fill,reservation_policy_digest FROM risk_bucket_orders_v22`).Scan(&orderID, &predecessor, &quantity, &cumulative, &policyDigest); err != nil {
		t.Fatal(err)
	}
	if orderID != "v22-order" || predecessor != "v22-parent" || quantity != 10 || cumulative != 4 || policyDigest != "v22-policy-seal" {
		t.Fatalf("legacy order changed: %s %s %d %d %s", orderID, predecessor, quantity, cumulative, policyDigest)
	}
	if _, err := current.db.Exec(`UPDATE risk_bucket_orders_v22 SET cumulative_fill=5 WHERE order_id='v22-order'`); err == nil {
		t.Fatal("v22 legacy order became mutable")
	}
	fkRows, err := current.db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	defer fkRows.Close()
	if fkRows.Next() {
		t.Fatal("v23 migration left a foreign-key violation")
	}
}

func TestMigrationV23FailureRollsBackRenamesTablesAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 22)
	seedReleasedV22RiskRows(t, old)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	broken := append(migrationsThrough(22), migration{Version: 23, SQL: schemaV23 + `INSERT INTO absent_table(x) VALUES(1);`})
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: broken, target: 23}})
	if err == nil {
		t.Fatal("broken v23 migration succeeded")
	}
	backups := backupsIn(t, filepath.Dir(path))
	if len(backups) != 1 {
		t.Fatalf("failed v23 backups=%v", backups)
	}
	assertBackupAtVersion(t, backups[0], 22, map[string]int{
		"risk_bucket_orders": 1, "risk_bucket_fills": 1, "risk_bucket_fill_allocations": 1,
	}, "risk_bucket_order_reservations")
	raw, openErr := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer raw.Close()
	var version int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 22 {
		t.Fatalf("version after rollback=%d err=%v", version, err)
	}
	for table, want := range map[string]int{"risk_bucket_orders": 1, "risk_bucket_fills": 1, "risk_bucket_fill_allocations": 1, "risk_bucket_orders_v22": 0, "risk_bucket_order_reservations": 0, "risk_bucket_fill_actual_evidence": 0} {
		var count int
		if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != want {
			t.Fatalf("rollback table %s count=%d want=%d err=%v", table, count, want, err)
		}
	}
	var rows int
	if err := raw.QueryRow(`SELECT count(*) FROM risk_bucket_orders WHERE order_id='v22-order'`).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("v22 row after rollback=%d err=%v", rows, err)
	}
}

func TestReleasedV22BuildRefusesV23Journal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	current := openJournalAtSchema(t, path, 23)
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(22), target: 22}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("v22 build open error=%v", err)
	}
}

func seedReleasedV22RiskRows(t *testing.T, j *Journal) {
	t.Helper()
	seedExistingRiskReservation(t, j, "v22-existing", "acct-v22")
	statements := []string{
		`INSERT INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,policy_fresh_until,record_digest,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at) VALUES('symbol','AAPL','v22','pd','authority','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','record','KRW','USD','2026-03-30T00:30:00Z','10','price','v1','price-d','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','0','0','0','v1','fee-d','1','1','fx','v1','fx-d','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,policy_version,limit_minor,filled_minor,held_minor,snapshot_version,policy_digest,observed_at,fresh_until,created_at) VALUES('v22-snapshot','sd','authority','sr','symbol','AAPL','v22','100','0','50','sv','pd','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) VALUES('v22-decision','v22-tx','acct-v22','US','AAPL',10,10,10,'v22-existing','rd','rp','ss','v22-generation','v22-lane','v22-campaign',1,'2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES('acct-v22','US','AAPL','v22-generation','v22-lane','v22-campaign','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at) VALUES('v22-reservation','v22-decision','v22-existing','acct-v22','US','AAPL','v22-generation','symbol','AAPL','v22','v22-snapshot','50','30','20','0','HELD','2026-03-30T00:30:00Z','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_orders(order_id,decision_id,predecessor_order_id,order_quantity,cumulative_fill,quote_currency,base_currency,reservation_policy_digest,created_at,updated_at) VALUES('v22-order','v22-decision','v22-parent',10,4,'USD','KRW','v22-policy-seal','2026-03-30T00:30:00Z','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_fills(fill_id,order_id,cumulative_fill,delta_quantity,actual_known,fill_digest,observed_at) VALUES('v22-fill','v22-order',4,4,0,'v22-fill-digest','2026-03-30T00:31:00Z')`,
		`INSERT INTO risk_bucket_fill_allocations(fill_id,reservation_id,transfer_minor,filled_minor) VALUES('v22-fill','v22-reservation','20','20')`,
	}
	for _, statement := range statements {
		if _, err := j.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}
