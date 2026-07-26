package journal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// openTestJournal opens a journal in an isolated temp directory with the
// filesystem guard satisfied by an injected prober. Every test in this package
// goes through it so no test ever touches the developer's real data directory.
func openTestJournal(t *testing.T) *Journal {
	t.Helper()
	return openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
}

func openTestJournalAt(t *testing.T, path string) *Journal {
	t.Helper()
	j, err := Open(context.Background(), Options{
		Path:     path,
		Clock:    clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open(%s): %v", path, err)
	}
	t.Cleanup(func() {
		if err := j.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return j
}

// TestOpenAppliesDurabilityPragmas pins the connection settings the durability
// requirement depends on. A silently-missing pragma would make every later
// crash-safety claim false.
func TestOpenAppliesDurabilityPragmas(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	var journalMode string
	if err := j.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(journalMode, "wal") {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}

	var synchronous int
	if err := j.db.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&synchronous); err != nil {
		t.Fatal(err)
	}
	if synchronous != 2 { // 2 = FULL
		t.Errorf("synchronous = %d, want 2 (FULL)", synchronous)
	}

	var foreignKeys int
	if err := j.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}
}

// TestSchemaTablesAndColumns is the schema contract Phase 2's ledger import will
// read: stable primary keys, immutable intent fields, a preserved MutationAttempt
// history and lineage edges.
func TestSchemaTablesAndColumns(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	wantTables := []string{
		"alert_outbox",
		"attempt_transitions",
		"decisions",
		"execution_corrections",
		"exit_events",
		"exit_states",
		"fill_events",
		"fill_snapshots",
		"flatten_sagas",
		"flatten_steps",
		"intents",
		"lineage_edges",
		"mutation_attempts",
		"operating_modes",
		"position_adjustments",
		"position_adoptions",
		"positions",
		"reconcile_states",
		"risk_reservations",
		"schema_meta",
		"spent_nonces",
		"trade_outcomes",
	}
	rows, err := j.db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var gotTables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		gotTables = append(gotTables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if strings.Join(gotTables, ",") != strings.Join(wantTables, ",") {
		t.Fatalf("tables = %v, want %v", gotTables, wantTables)
	}

	wantColumns := map[string][]string{
		"intents": {
			"account_ref", "created_at", "currency", "fingerprint", "id", "market",
			"notes", "order_type", "price", "quantity", "side", "source", "symbol",
			"time_in_force", "trading_day",
		},
		// Schema v5 (extend-execution-contract task 0.1) extends the attempt with
		// its decision, the idempotency key it claimed and the exact bytes a
		// replay may resend.
		"mutation_attempts": {
			"account_ref", "attempt_no", "broker_order_id", "client_order_id",
			"decision_id", "detail", "dispatch_started_at", "fingerprint",
			"generation", "id", "intent_id", "kind", "last_replay_at", "reason_code",
			"recorded_at", "replay_count", "safety_class", "serializer_version",
			"settled_at", "state", "target_order_id", "wire_body",
		},
		"attempt_transitions": {
			"at", "attempt_id", "detail", "from_state", "id", "reason_code", "to_state",
		},
		"lineage_edges": {
			"attempt_id", "child_order_id", "created_at", "id", "intent_id",
			"parent_filled_quantity", "parent_order_id", "relation",
			"requested_quantity",
		},
		// Schema v2 (task 3.2): the cumulative-snapshot fill ledger.
		// filled_amount is v5: without a previous amount, an amount-only
		// execution correction cannot be detected.
		"fill_snapshots": {
			"average_price", "broker_visible_at", "committed_at", "detail",
			"fail_closed", "filled_amount", "filled_quantity", "market", "observed_at",
			"order_id", "quantity", "reason_code", "state", "symbol", "terminal",
		},
		"fill_events": {
			"average_price", "broker_visible_at", "committed_at", "cumulative_quantity",
			"delta_quantity", "id", "market", "order_id", "symbol",
		},
		// Schema v4 (task 4.4): the flatten saga's durable plan.
		"flatten_sagas": {
			"account_ref", "detail", "dry_run", "finished_at", "id", "operator",
			"phase", "reason", "started_at", "updated_at",
		},
		"flatten_steps": {
			"attempt_id", "created_at", "currency", "detail", "id", "intent_id",
			"kind", "market", "price", "quantity", "reason_code", "saga_id", "side",
			"state", "symbol", "target_order_id", "updated_at",
		},
		// Schema v3 (task 4.3): the durable alert outbox.
		"alert_outbox": {
			"acknowledged_at", "acknowledged_by", "attempts", "body", "created_at",
			"delivered_at", "event_key", "event_type", "id", "last_attempt_at",
			"last_error", "payload", "severity", "state", "title",
		},
		// Schema v5 (extend-execution-contract task 0.1): the decision contract,
		// what it reserved, what it consumed, and the two records that keep the
		// broker's picture and ours reconcilable. design.md D9 is the normative
		// table; execution_contract_test.go pins the constraints.
		"decisions": {
			"account_ref", "client_order_id", "expires_at", "generation", "id",
			"issued_at", "limits_json", "nonce", "preimage_kind", "risk_hash",
			"risk_preimage", "safety_class",
		},
		"risk_reservations": {
			"account_ref", "amount", "attempt_id", "currency", "decision_id", "id",
			"kind", "release_reason", "released_at", "snapshot_as_of", "state",
			"trading_day",
		},
		"spent_nonces": {
			"consumed_at", "decision_id", "nonce",
		},
		"reconcile_states": {
			"account_ref", "cause", "entered_at", "evidence", "id", "release_cause",
			"released_at", "symbol",
		},
		"execution_corrections": {
			"account_ref", "cumulative_qty", "id", "new_avg_price",
			"new_filled_amount", "observed_at", "order_id", "prev_avg_price",
			"prev_filled_amount",
		},
		// Schema v7 (adopt-external-positions design A1): the adoption record and
		// the one nullable column that points at it. `adoption_id` is listed here
		// because a golden column list is how a set-once column stops being
		// quietly widened — adding a second reference column would show up as a
		// diff in this literal.
		"position_adoptions": {
			"cost_basis", "cost_basis_src", "id", "market", "observed_at",
			"observed_price", "preimage_digest", "quantity", "symbol", "synthetic_stop",
		},
		"positions": {
			"account_ref", "adoption_id", "avg_price", "closed_at", "entry_decision_id",
			"id", "instance_seq", "market", "opened_at", "quantity", "state", "symbol",
		},
	}
	for table, want := range wantColumns {
		got := tableColumns(t, j, table)
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s columns =\n  %v\nwant\n  %v", table, got, want)
		}
	}
}

func tableColumns(t *testing.T, j *Journal, table string) []string {
	t.Helper()
	rows, err := j.db.QueryContext(context.Background(),
		"SELECT name FROM pragma_table_info(?)", table)
	if err != nil {
		t.Fatalf("pragma_table_info(%s): %v", table, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatalf("table %s has no columns (missing table?)", table)
	}
	return out
}

// TestSchemaVersionRecorded verifies the version field and its mirror row.
func TestSchemaVersionRecorded(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	got, err := j.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != SchemaVersion {
		t.Fatalf("SchemaVersion() = %d, want %d", got, SchemaVersion)
	}

	var mirrored string
	if err := j.db.QueryRowContext(ctx,
		"SELECT value FROM schema_meta WHERE key = 'schema_version'").Scan(&mirrored); err != nil {
		t.Fatalf("schema_meta mirror row: %v", err)
	}
	if want := strconv.Itoa(SchemaVersion); mirrored != want {
		t.Fatalf("schema_meta.schema_version = %q, want %q", mirrored, want)
	}

	var createdAt string
	if err := j.db.QueryRowContext(ctx,
		"SELECT value FROM schema_meta WHERE key = 'created_at'").Scan(&createdAt); err != nil {
		t.Fatalf("created_at row: %v", err)
	}
	// The injected clock, not time.Now(), stamps the record.
	if createdAt != "2026-03-30T00:30:00Z" {
		t.Fatalf("created_at = %q, want the injected clock's instant", createdAt)
	}
}

// TestOpenIsIdempotent proves reopening an existing journal neither re-runs the
// migration destructively nor loses data.
func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	first := openTestJournalAt(t, path)
	ctx := context.Background()

	insertIntent(t, first, "intent-1")
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second := openTestJournalAt(t, path)
	var count int
	if err := second.db.QueryRowContext(ctx, "SELECT count(*) FROM intents").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("intents after reopen = %d, want 1", count)
	}
	v, err := second.SchemaVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if v != SchemaVersion {
		t.Fatalf("schema version after reopen = %d, want %d", v, SchemaVersion)
	}
}

// TestOpenRefusesNewerSchema is the forward-compatibility guard: a journal written
// by a newer TossOS must not be silently used by an older binary that would
// misread it.
func TestOpenRefusesNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	if _, err := j.db.ExecContext(context.Background(), "PRAGMA user_version = 99"); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := Open(context.Background(), Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("Open on a newer schema: want ErrSchemaTooNew, got %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "99") {
		t.Errorf("refusal should name the found version: %v", err)
	}
}

// TestOpenRefusesForbiddenFilesystem proves the guard runs before any file is
// created: a rejected mount must leave nothing behind.
func TestOpenRefusesForbiddenFilesystem(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	path := filepath.Join(dir, "journal.db")

	_, err := Open(context.Background(), Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "fuse/fuseblk", Magic: MagicFUSE}),
	})
	if !errors.Is(err, ErrFilesystemNotAllowed) {
		t.Fatalf("want ErrFilesystemNotAllowed, got %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("journal file must not be created on a refused filesystem (stat: %v)", statErr)
	}
	if _, statErr := os.Stat(dir); !os.IsNotExist(statErr) {
		t.Fatalf("data directory must not be created on a refused filesystem (stat: %v)", statErr)
	}
}

// TestOpenCreatesPrivateDataDir keeps the journal out of other users' reach: it
// holds the full order history of a live account.
func TestOpenCreatesPrivateDataDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data", "tossos")
	j := openTestJournalAt(t, filepath.Join(dir, "journal.db"))
	if j.Path() != filepath.Join(dir, "journal.db") {
		t.Fatalf("Path() = %q", j.Path())
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Fatalf("data dir permissions = %o, want 700", perm)
	}
	dbInfo, err := os.Stat(j.Path())
	if err != nil {
		t.Fatal(err)
	}
	if perm := dbInfo.Mode().Perm(); perm&0o077 != 0 {
		t.Fatalf("journal file permissions = %o, want no group/other access", perm)
	}
}

// TestOpenUsesDefaultPathFromEnv wires the env resolution into Open so the engine
// does not have to compute the path itself.
func TestOpenUsesDefaultPathFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvDataDir, dir)
	t.Setenv(EnvXDGDataHome, "")

	j, err := Open(context.Background(), Options{
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer j.Close()

	if want := filepath.Join(dir, DBFileName); j.Path() != want {
		t.Fatalf("Path() = %q, want %q", j.Path(), want)
	}
}

// TestForeignKeysRejectOrphanAttempt proves referential integrity is enforced, so
// an attempt can never exist without the immutable intent it came from.
func TestForeignKeysRejectOrphanAttempt(t *testing.T) {
	j := openTestJournal(t)
	_, err := j.db.ExecContext(context.Background(),
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		 VALUES ('a-1', 'missing-intent', 'PLACE', 'RECORDED', 1, 'fp', '2026-03-30T00:30:00Z')`)
	if err == nil {
		t.Fatal("insert with an unknown intent_id must violate the foreign key")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("want a foreign-key error, got %v", err)
	}
}

// TestStrictTablesRejectWrongType proves the STRICT table modifier is active.
//
// STRICT still performs lossless conversions (an INTEGER 10 into a TEXT column
// becomes '10'), so what it actually buys us is that no column is typeless and a
// value that cannot be converted is refused instead of stored as-is — here, text
// in the INTEGER attempt counter.
func TestStrictTablesRejectWrongType(t *testing.T) {
	j := openTestJournal(t)
	insertIntent(t, j, "intent-1")

	_, err := j.db.ExecContext(context.Background(),
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		 VALUES ('a-bad', 'intent-1', 'PLACE', 'RECORDED', 'many', 'fp', '2026-03-30T00:30:00Z')`)
	if err == nil {
		t.Fatal("STRICT table must reject a TEXT value in an INTEGER column")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "datatype mismatch") &&
		!strings.Contains(strings.ToLower(err.Error()), "cannot store") {
		t.Fatalf("want a STRICT datatype error, got %v", err)
	}
}

// TestLineageEdgeUniqueness keeps a replace relation recorded exactly once even if
// a resolution pass re-derives it.
func TestLineageEdgeUniqueness(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertIntent(t, j, "intent-1")
	insertAttempt(t, j, "attempt-1", "intent-1")

	insert := func() error {
		_, err := j.db.ExecContext(ctx,
			`INSERT INTO lineage_edges
			   (parent_order_id, child_order_id, relation, parent_filled_quantity,
			    requested_quantity, intent_id, attempt_id, created_at)
			 VALUES ('order-1', 'order-2', 'replaces', '3', '7', 'intent-1', 'attempt-1',
			         '2026-03-30T00:30:00Z')`)
		return err
	}
	if err := insert(); err != nil {
		t.Fatalf("first lineage edge: %v", err)
	}
	if err := insert(); err == nil {
		t.Fatal("duplicate (parent, child, relation) must violate the unique index")
	}
}

// TestSchemaIndexes verifies the lookup paths the IN_DOUBT resolution and fill
// detection loops depend on exist, so a resolution pass cannot degrade into a full
// table scan on a large journal.
func TestSchemaIndexes(t *testing.T) {
	j := openTestJournal(t)
	rows, err := j.db.QueryContext(context.Background(),
		`SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		got[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"idx_intents_fingerprint",
		"idx_intents_symbol_day",
		"idx_attempts_state",
		"idx_attempts_intent",
		"idx_attempts_broker_order",
		"idx_transitions_attempt",
		"idx_lineage_child",
		"idx_lineage_parent",
		"idx_fill_snapshots_symbol",
		"idx_fill_snapshots_terminal",
		"idx_fill_events_order",
		"idx_fill_events_symbol",
		// Schema v5: the decision contract's lookup paths — expiry sweeps,
		// per-account reservation sums, the restart sweep for orphaned holds and
		// the correction history of one order.
		"idx_decisions_account",
		"idx_decisions_expires",
		"idx_reservations_decision",
		"idx_reservations_attempt",
		"idx_reservations_held",
		"idx_spent_nonces_consumed",
		"idx_reconcile_account",
		"idx_corrections_order",
	} {
		if !got[want] {
			t.Errorf("missing index %s (have %v)", want, keysOf(got))
		}
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// insertIntent inserts a minimal valid intent row directly, so schema-level tests
// do not depend on the write API added by the durability task.
func insertIntent(t *testing.T, j *Journal, id string) {
	t.Helper()
	_, err := j.db.ExecContext(context.Background(),
		`INSERT INTO intents
		   (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
		    time_in_force, quantity, price, currency, source, fingerprint, notes)
		 VALUES (?, '2026-03-30T00:30:00Z', 'us', '2026-03-30', 'acct-1', 'AAPL', 'BUY',
		         'LIMIT', 'DAY', '10', '200.5', 'USD', 'engine', 'fp-1', '')`, id)
	if err != nil {
		t.Fatalf("insert intent %s: %v", id, err)
	}
}

func insertAttempt(t *testing.T, j *Journal, id, intentID string) {
	t.Helper()
	_, err := j.db.ExecContext(context.Background(),
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		 VALUES (?, ?, 'PLACE', 'RECORDED', 1, 'fp-1', '2026-03-30T00:30:00Z')`, id, intentID)
	if err != nil {
		t.Fatalf("insert attempt %s: %v", id, err)
	}
}
