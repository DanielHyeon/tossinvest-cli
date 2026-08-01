package journal

// SchemaVersion is the schema version this build writes and understands. It is
// stored in the database's PRAGMA user_version and mirrored, as text, in
// schema_meta for human inspection.
const SchemaVersion = 14

// migration is one forward step. The additive rules are not negotiable, because a
// live account's order history is the thing being migrated:
//
//  1. Never edit a released step. Append a new one.
//  2. New columns are nullable or carry a DEFAULT, so an older row stays valid.
//  3. Never drop or rename a column another version might still read; deprecate
//     it in a comment instead.
//  4. Never rewrite historical rows in a migration. The attempt history and the
//     intent fields are the audit trail Phase 2's ledger imports.
//
// There is no down-migration and there will not be one: a step that could undo
// itself would have to drop a column, which rule 3 forbids. The rollback is "run
// the previous binary", which refuses the newer user_version (ErrSchemaTooNew)
// rather than misreading it, and the recovery from a migration that fails partway
// is the automatic pre-migration backup (backup.go).
type migration struct {
	Version int
	SQL     string
}

// migrationPlan is the set of forward steps to apply and the version this build
// claims to understand. Production always uses defaultMigrationPlan; the
// migration tests substitute their own to build an older database on disk and to
// inject a step that fails partway.
type migrationPlan struct {
	steps  []migration
	target int
}

func defaultMigrationPlan() migrationPlan {
	return migrationPlan{steps: migrations, target: SchemaVersion}
}

// migrations is the ordered list of forward steps. Index is irrelevant; Version is
// authoritative.
var migrations = []migration{
	{Version: 1, SQL: schemaV1},
	// schemaV2 lives in fills.go, next to the code that reads it (task 3.2).
	{Version: 2, SQL: schemaV2},
	// schemaV3 lives in outbox.go, next to the code that reads it (task 4.3).
	{Version: 3, SQL: schemaV3},
	// schemaV4 lives in flatten.go, next to the code that reads it (task 4.4).
	{Version: 4, SQL: schemaV4},
	// schemaV5 lives in execution_contract.go: it is the transcription of
	// design.md D9 for change extend-execution-contract (task 0.1), and the code
	// that reads its tables arrives in the later tasks of that change.
	{Version: 5, SQL: schemaV5},
	// schemaV6 lives in core_domain.go: it is the transcription of design.md D7
	// for change add-core-domain (task 0.1), and the code that reads its tables
	// arrives in the later tasks of that change.
	{Version: 6, SQL: schemaV6},
	// schemaV7 lives in adoption.go: it is the transcription of design.md A1 for
	// change adopt-external-positions, and the code that reads its table is in
	// the same file — the adoption record has exactly one writer and the layout
	// is what enforces it.
	{Version: 7, SQL: schemaV7},
	// schemaV8 lives in entryobservation.go: it is the transcription of design.md
	// D1 for change add-net-rr-measurement, and the code that reads and writes its
	// table is in the same file — the observation has one writer on the trading
	// path and one on the reconstruction path, and the layout is what keeps the
	// contract tables out of both.
	{Version: 8, SQL: schemaV8},
	{Version: 9, SQL: schemaV9},
	// schemaV10 lives in exit_snapshot.go beside the typed snapshot and
	// generation-scoped quarantine APIs that consume it.
	{Version: 10, SQL: schemaV10},
	// schemaV11 lives in account_views.go beside the bounded evidence query.
	{Version: 11, SQL: schemaV11},
	// schemaV12 lives in position_policy.go beside the generation-scoped CAS
	// repository. It only adds tables; existing exit snapshots are never rebound.
	{Version: 12, SQL: schemaV12},
	// schemaV13 lives in protection_orders.go. It adds the broker-resident
	// protection saga and its append-only mutation-attempt history. No existing
	// exit snapshot or position-policy generation is rebound by this step.
	{Version: 13, SQL: schemaV13},
	// schemaV14 lives in strategy_lineage.go. It adds immutable strategy
	// decision/attempt/execution provenance without changing RiskIntent or any
	// released canonical preimage bytes.
	{Version: 14, SQL: schemaV14},
}

// schemaV1 is the initial schema.
//
// Design notes that the tests pin:
//
//   - Quantities and prices are stored as decimal *strings* because binary floats
//     cannot represent a broker's decimal quantities exactly. The write API is
//     what keeps them strings; STRICT tables add that no column is typeless and
//     that an unconvertible value (text in an INTEGER counter) is refused rather
//     than stored.
//   - Text timestamps in RFC3339 UTC ("2026-03-30T00:30:00Z"), always from the
//     injected clock. They sort lexicographically, which keeps range scans on
//     recorded_at usable.
//   - intents rows are immutable once written. Everything mutable about an
//     attempt lives in mutation_attempts, and every state change is additionally
//     appended to attempt_transitions so the history is never overwritten.
//   - Primary keys are caller-supplied stable ids (intent id, attempt id), which
//     is what the Phase 2 ledger import keys on.
const schemaV1 = `
CREATE TABLE schema_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;

-- Immutable record of what the engine decided to do. Never updated.
CREATE TABLE intents (
	id            TEXT PRIMARY KEY,          -- stable PK; Phase 2 ledger import key
	created_at    TEXT NOT NULL,             -- RFC3339 UTC, injected clock
	market        TEXT NOT NULL,             -- 'kr' | 'us' (internal/clock.Market)
	trading_day   TEXT NOT NULL,             -- market-local trading day, YYYY-MM-DD
	account_ref   TEXT NOT NULL,             -- account reference the intent targets
	symbol        TEXT NOT NULL,
	side          TEXT NOT NULL,             -- 'BUY' | 'SELL'
	order_type    TEXT NOT NULL,             -- 'LIMIT' | 'MARKET' | …
	time_in_force TEXT NOT NULL DEFAULT '',
	quantity      TEXT NOT NULL,             -- decimal string
	price         TEXT,                      -- decimal string; NULL for market orders
	currency      TEXT NOT NULL,
	source        TEXT NOT NULL,             -- what produced the intent (strategy id, operator)
	fingerprint   TEXT NOT NULL,             -- IN_DOUBT matching key
	notes         TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_intents_symbol_day   ON intents(symbol, trading_day);
CREATE INDEX idx_intents_fingerprint  ON intents(fingerprint);

-- One PLACE/CANCEL/AMEND attempt against the broker, with its own lifecycle.
CREATE TABLE mutation_attempts (
	id                  TEXT PRIMARY KEY,
	intent_id           TEXT NOT NULL REFERENCES intents(id),
	kind                TEXT NOT NULL,              -- 'PLACE' | 'CANCEL' | 'AMEND'
	state               TEXT NOT NULL,              -- see lifecycle states
	attempt_no          INTEGER NOT NULL,           -- 1-based, per intent
	target_order_id     TEXT NOT NULL DEFAULT '',   -- CANCEL/AMEND target order
	broker_order_id     TEXT NOT NULL DEFAULT '',   -- assigned once the broker acks
	fingerprint         TEXT NOT NULL,
	recorded_at         TEXT NOT NULL,
	dispatch_started_at TEXT,
	settled_at          TEXT,
	reason_code         TEXT NOT NULL DEFAULT '',
	detail              TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_attempts_state        ON mutation_attempts(state);
CREATE INDEX idx_attempts_intent       ON mutation_attempts(intent_id);
CREATE INDEX idx_attempts_broker_order ON mutation_attempts(broker_order_id);

-- Append-only transition log. Rows are never updated or deleted.
CREATE TABLE attempt_transitions (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	attempt_id  TEXT NOT NULL REFERENCES mutation_attempts(id),
	from_state  TEXT NOT NULL,
	to_state    TEXT NOT NULL,
	at          TEXT NOT NULL,
	reason_code TEXT NOT NULL DEFAULT '',
	detail      TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_transitions_attempt ON attempt_transitions(attempt_id, id);

-- Broker order lineage. The official API's cancel/modify issues a NEW order
-- number, so a replace is recorded as an edge rather than as a mutation of the
-- original order row: parent keeps its filled quantity, child carries the
-- requested remainder.
CREATE TABLE lineage_edges (
	id                     INTEGER PRIMARY KEY AUTOINCREMENT,
	parent_order_id        TEXT NOT NULL,
	child_order_id         TEXT NOT NULL,
	relation               TEXT NOT NULL,   -- 'replaces'
	parent_filled_quantity TEXT NOT NULL,   -- decimal string at replace time
	requested_quantity     TEXT NOT NULL,   -- decimal string requested on the child
	intent_id              TEXT NOT NULL REFERENCES intents(id),
	attempt_id             TEXT NOT NULL REFERENCES mutation_attempts(id),
	created_at             TEXT NOT NULL,
	UNIQUE(parent_order_id, child_order_id, relation)
) STRICT;

CREATE INDEX idx_lineage_child  ON lineage_edges(child_order_id);
CREATE INDEX idx_lineage_parent ON lineage_edges(parent_order_id);
`
