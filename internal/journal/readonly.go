package journal

// readonly.go is the second way into this database, and the only one that is not
// allowed to change it (change add-operator-dashboard, task 1.1).
//
// # Why Open could not be reused
//
// Open is the engine's. It resolves the data directory, MkdirAll's it, opens with
// the durability pragmas and then migrates the schema forward — all of which are
// correct for the process that owns the journal and all of which are wrong for a
// console reading it over the engine's shoulder:
//
//	MkdirAll   a console started before the engine would create the data
//	           directory, and Open would then put an empty v6 database in it. The
//	           engine's first run would find a file somebody else made.
//	migrate    an older console would migrate a live account's journal to suit
//	           itself. There is no down step, so that is not recoverable.
//	writes     every one of them is a change nobody approved.
//
// So this is a separate constructor with a separate handle type. *ReadOnly has no
// method that writes and never will (readonly_test.go enumerates its method set),
// which makes "the dashboard cannot mutate the journal" a compile-time property
// rather than a review habit.
//
// # The one thing it does create
//
// SQLite cannot read a WAL database without the wal-index. Opening one read-only
// requires either an existing `-shm` or write access to the directory so SQLite
// can make it, and the `-wal` is read alongside. That pair is the documented
// exception in the operator-console spec — it is the precondition of reading a
// WAL database at all, not a change to the account's record. The database file
// itself is never created: a missing one is ErrJournalMissing and the screen says
// the engine has not run.
//
// # Both schema directions are named
//
// A journal newer than this build (ErrSchemaTooNew, shared with Open) wants a
// newer console. A journal older than the tables this build reads (ErrSchemaTooOld)
// wants the engine to start and migrate. Neither is rendered as "no data",
// because an empty screen is the console asserting something about an account
// that it has not read.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrJournalMissing means there is no database at that path. It is a state the
// dashboard reports ("the engine has not run yet"), not a fault: the console is
// frequently started first, and refusing to render would make it useless exactly
// when an operator is trying to find out whether anything is running.
var ErrJournalMissing = errors.New("journal: no database at that path")

// ErrSchemaTooOld means the database predates the tables this build reads. It is
// the mirror of ErrSchemaTooNew: the fix is to start the engine, which migrates.
var ErrSchemaTooOld = errors.New("journal: the database is older than the tables this build reads")

// readOnlyTables are the tables a read-only reader needs to exist before it can
// answer anything. They are the v6 core domain plus the v1 attempt log: a journal
// without them is a pre-v6 file, and every query below would fail one statement
// at a time instead of once, clearly, at open.
//
// mutation_attempts joined the list with BrokerOrderIDs (change
// console-orders-screen, task 2.2), and the failure it forecloses is specific:
// unregistered, OpenReadOnly succeeds, that one query fails, and a query failure
// the screen renders as zero rows says every order on the account was placed by
// hand. It costs nothing in practice — the table is schemaV1's and the four
// above are schemaV6's, so any journal that has them has it — which is exactly
// the property that makes registering it additive rather than a new refusal.
var readOnlyTables = []string{
	"positions", "exit_states", "exit_events", "trade_outcomes", "mutation_attempts",
}

var readOnlyColumns = []struct {
	table  string
	column string
}{
	{table: "exit_states", column: "policy_id"},
	{table: "exit_states", column: "snapshot_status"},
	{table: "exit_states", column: "effective_snapshot_json"},
	{table: "exit_events", column: "saved_snapshot_json"},
	{table: "exit_events", column: "recomputed_snapshot_json"},
	{table: "exit_events", column: "effective_snapshot_json"},
	{table: "exit_events", column: "effective_source"},
	{table: "exit_states", column: "policy_version"},
	{table: "trade_outcomes", column: "cost_total"},
}

var positionCampaignReadOnlyTables = []string{
	"position_campaigns", "campaign_legs", "campaign_order_watermarks", "campaign_commands",
	"campaign_events", "position_campaign_claims", "position_projection_versions",
}

var positionCampaignReadOnlyColumns = []struct {
	table  string
	column string
}{
	{table: "position_campaigns", column: "decision_id"},
	{table: "position_campaigns", column: "entry_blocked"},
	{table: "campaign_legs", column: "residual_quantity"},
	{table: "campaign_order_watermarks", column: "account_ref"},
	{table: "campaign_order_watermarks", column: "trading_day"},
	{table: "campaign_order_watermarks", column: "side"},
	{table: "campaign_order_watermarks", column: "decision_id"},
	{table: "campaign_order_watermarks", column: "remaining_quantity"},
	{table: "campaign_commands", column: "result_version"},
	{table: "campaign_commands", column: "result_error"},
	{table: "campaign_events", column: "entry_blocked"},
	{table: "campaign_events", column: "projection_digest"},
	{table: "campaign_events", column: "leg_requested_quantity"},
	{table: "campaign_events", column: "order_remaining_quantity"},
	{table: "position_campaign_claims", column: "position_version"},
	{table: "position_projection_versions", column: "version"},
}

// ReadOnlyOptions configures OpenReadOnly.
type ReadOnlyOptions struct {
	// Path is the SQLite file. Empty resolves DefaultPath(), the same way Open
	// does — but unlike Open, nothing along that path is created.
	Path string

	// BusyTimeout bounds how long a read waits behind the writer's commit. Zero
	// uses defaultBusyTimeout. A reader in WAL mode is not normally blocked by a
	// writer at all; this covers the checkpoint.
	BusyTimeout time.Duration
}

// ReadOnly is a read-only handle on an existing journal.
//
// It is deliberately not a *Journal. A *Journal carries the write API, the apply
// hooks and the mode projector, and handing one to a dashboard would make "the
// console does not write" a property of the console's discipline rather than of
// its type.
type ReadOnly struct {
	db      *sql.DB
	path    string
	version int
}

// OpenReadOnly opens an existing journal for reading.
//
// It creates no database file and no directory, runs no migration and issues no
// write. A missing file, a schema newer than this build and a schema older than
// the tables this build reads are three different typed refusals, because they
// have three different answers.
func OpenReadOnly(ctx context.Context, opts ReadOnlyOptions) (*ReadOnly, error) {
	path := opts.Path
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = resolved
	}
	path = filepath.Clean(path)

	// Checked before the driver is asked, so that "there is no journal" is a
	// typed answer rather than a driver string, and so nothing can be created by
	// a DSN that lost its mode parameter.
	if info, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrJournalMissing, path)
		}
		return nil, fmt.Errorf("journal: looking for %s: %w", path, err)
	} else if info.IsDir() {
		return nil, fmt.Errorf("%w: %s is a directory", ErrJournalMissing, path)
	}

	// The filesystem allowlist (fsguard.go) is deliberately not applied. It exists
	// because fsync on a network mount does not mean what the durability contract
	// needs it to mean, and this connection never writes — refusing to *read* a
	// journal on an unusual mount would be a guard doing something it was not
	// built for.
	busy := opts.BusyTimeout
	if busy <= 0 {
		busy = defaultBusyTimeout
	}

	db, err := sql.Open("sqlite", readOnlyDSN(path, busy))
	if err != nil {
		return nil, fmt.Errorf("journal: opening %s read-only: %w", path, err)
	}
	// One connection, as Open uses: the console serves one operator and a pool
	// would only add connections that each need their own wal-index handshake.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("journal: connecting to %s read-only: %w", path, err)
	}

	ro := &ReadOnly{db: db, path: path}
	if err := ro.checkSchema(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return ro, nil
}

// readOnlyDSN is the connection contract, and every part of it is a refusal.
//
//	mode=ro       sqlite3_open_v2 with SQLITE_OPEN_READONLY: the file is never
//	              created and every write returns SQLITE_READONLY.
//	query_only    the same refusal one layer up, so a statement that somehow ran
//	              on this connection is rejected before it reaches the pager.
//	busy_timeout  a bounded wait behind a checkpoint instead of SQLITE_BUSY.
//
// journal_mode is deliberately absent: setting it is a write, and the mode is
// whatever the engine put in the file.
func readOnlyDSN(path string, busy time.Duration) string {
	q := url.Values{}
	q.Set("mode", "ro")
	q.Add("_pragma", "query_only(true)")
	q.Add("_pragma", "busy_timeout("+strconv.FormatInt(busy.Milliseconds(), 10)+")")
	return "file:" + path + "?" + q.Encode()
}

// checkSchema names the mismatch in whichever direction it is.
func (r *ReadOnly) checkSchema(ctx context.Context) error {
	if err := r.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&r.version); err != nil {
		return fmt.Errorf("journal: reading the schema version of %s: %w", r.path, err)
	}
	if r.version > SchemaVersion {
		return fmt.Errorf("%w: found version %d, this build understands %d — update the console",
			ErrSchemaTooNew, r.version, SchemaVersion)
	}

	var missing []string
	for _, table := range readOnlyTables {
		var name string
		err := r.db.QueryRowContext(ctx,
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			missing = append(missing, table)
		case err != nil:
			return fmt.Errorf("journal: inspecting the schema of %s: %w", r.path, err)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: version %d has no %s — start the engine once so it migrates",
			ErrSchemaTooOld, r.version, strings.Join(missing, ", "))
	}

	for _, required := range readOnlyColumns {
		var name string
		err := r.db.QueryRowContext(ctx,
			`SELECT name FROM pragma_table_info(?) WHERE name = ?`,
			required.table, required.column).Scan(&name)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			missing = append(missing, required.table+"."+required.column)
		case err != nil:
			return fmt.Errorf("journal: inspecting the schema of %s: %w", r.path, err)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: version %d has no %s — start the engine once so it migrates",
			ErrSchemaTooOld, r.version, strings.Join(missing, ", "))
	}
	// Keep released v8/v9/v14 diagnostics above authoritative. Campaign
	// prerequisites only exist at v20, so checking them earlier would mask the
	// older, more specific missing-column contract.
	if r.version >= 20 {
		for _, table := range positionCampaignReadOnlyTables {
			var name string
			err := r.db.QueryRowContext(ctx,
				`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				missing = append(missing, table)
			case err != nil:
				return fmt.Errorf("journal: inspecting the schema of %s: %w", r.path, err)
			}
		}
		for _, required := range positionCampaignReadOnlyColumns {
			var name string
			err := r.db.QueryRowContext(ctx,
				`SELECT name FROM pragma_table_info(?) WHERE name = ?`, required.table, required.column).Scan(&name)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				missing = append(missing, required.table+"."+required.column)
			case err != nil:
				return fmt.Errorf("journal: inspecting the schema of %s: %w", r.path, err)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("%w: version %d has no %s — start the engine once so it migrates",
				ErrSchemaTooOld, r.version, strings.Join(missing, ", "))
		}
	}
	if r.version >= 21 {
		for _, required := range strategyEvidenceReadOnlyColumns {
			var name string
			err := r.db.QueryRowContext(ctx,
				`SELECT name FROM pragma_table_info(?) WHERE name = ?`, required.table, required.column).Scan(&name)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				missing = append(missing, required.table+"."+required.column)
			case err != nil:
				return fmt.Errorf("journal: inspecting the schema of %s: %w", r.path, err)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("%w: version %d has no %s — start the engine once so it migrates",
				ErrSchemaTooOld, r.version, strings.Join(missing, ", "))
		}
	}
	return nil
}

// Close releases the handle.
func (r *ReadOnly) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

// Path reports which file was opened.
func (r *ReadOnly) Path() string { return r.path }

// SchemaVersion is the version read at open. It does not move, because nothing
// on this handle can move it.
func (r *ReadOnly) SchemaVersion() int { return r.version }
