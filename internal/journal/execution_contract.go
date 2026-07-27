package journal

// execution_contract.go holds schema v5: the decision, reservation, nonce,
// RECONCILE and correction tables the extended execution contract records
// (change extend-execution-contract, design.md D9 — that table is normative and
// this file is its transcription).
//
// # What the tables are for
//
// P1 recorded what the engine *attempted*. v5 records what it was *entitled* to
// attempt, and what that entitlement consumed:
//
//   - decisions — the issuer's persisted decision, written in the issuer's own
//     transaction before the Gateway is called (D1). The Gateway re-derives the
//     preimage hash from this row rather than trusting a caller-supplied value,
//     so a decision that was never written cannot be executed.
//   - risk_reservations — what a decision held against the account's limits
//     while it was in flight (D5). Released only on a *derived* terminal broker
//     state, an unconsumed expiry, a trading-day boundary, or an operator with a
//     recorded reason. Never on an assumed expiry.
//   - spent_nonces — single use, across restarts (D5). The retention invariant
//     is that these rows outlive the longest decision TTL: no decision may
//     survive the record of its own consumption.
//   - reconcile_states — RECONCILE as durable state rather than an in-memory
//     latch (D6). The journal is authoritative; the in-process latches become
//     projections rebuilt from here at startup.
//   - execution_corrections — the broker's cumulative execution model has no
//     per-fill id and its average price moves with every partial fill, so a
//     correction at an unchanged cumulative quantity is a real event with no
//     quantity delta (D7). Written inside RecordFill's transaction, which is the
//     only place the previous snapshot still exists.
//
// # Conventions inherited from schema.go
//
// Decimals are TEXT (binary floats cannot hold a broker's decimals exactly),
// timestamps are RFC3339 UTC text from the injected clock, tables are STRICT,
// and primary keys are caller-supplied stable ids. Nothing here rewrites a
// historical row.
//
// # Broker-behaviour claims
//
// None. Every value in this file is TossOS-internal. The one broker fact the
// schema leans on is that OrderCreateRequest.clientOrderId is an idempotency key
// valid for ten minutes (docs/migration/openapi.latest.json), which is why
// client_order_id is claimed by at most one attempt per account.

// Safety classes (D4). PROTECTION_WEAKENING is reserved in the CHECK and is not
// issued in 2a: reserving the value now is what keeps 2c's change additive, and
// a value that no writer produces cannot widen anything today.
const (
	SafetyClassExposureRaising     = "EXPOSURE_RAISING"
	SafetyClassRiskReducing        = "RISK_REDUCING"
	SafetyClassProtectionWeakening = "PROTECTION_WEAKENING"
)

// Preimage schemas (D1). The class determines which one a decision carries:
// EXPOSURE_RAISING → RiskIntent, RISK_REDUCING → ReductionIntent. Recording the
// tag next to the preimage is what lets the Gateway re-derive the hash with the
// same schema the issuer used.
const (
	PreimageKindRiskIntent      = "RISK_INTENT"
	PreimageKindReductionIntent = "REDUCTION_INTENT"
)

// Reservation kinds (D5).
const (
	ReservationKindOpenExposure = "OPEN_EXPOSURE"
	ReservationKindDailyLoss    = "DAILY_LOSS"
	ReservationKindCash         = "CASH"
)

// Reservation states (D5). A reservation is HELD from the moment it exists.
const (
	ReservationHeld     = "HELD"
	ReservationReleased = "RELEASED"
)

// Release reasons (D5). Every exit from a held reservation names one of these,
// so an operator release is never indistinguishable from an automatic one.
const (
	// ReleaseReasonBrokerTerminal — the attempt reached a *derived* terminal
	// broker state. Never an assumed expiry: how an expired order appears is
	// [미측정 — 2b 2.1] and the derivation keeps that case UNKNOWN.
	ReleaseReasonBrokerTerminal = "BROKER_TERMINAL"
	// ReleaseReasonExpiredUnconsumed — the decision expired without its nonce
	// ever being consumed, so nothing was ever sent.
	ReleaseReasonExpiredUnconsumed = "EXPIRED_UNCONSUMED"
	// ReleaseReasonOperator — the only exit for a reservation held by an
	// UNKNOWN_BROKER_STATE or UNRESOLVED attempt. Requires a recorded reason and
	// is audited: a fail-closed ratchet nobody can see is not operable.
	ReleaseReasonOperator = "OPERATOR"
	// ReleaseReasonDayBoundary — a daily-loss reservation lapsing with the
	// market's trading day.
	ReleaseReasonDayBoundary = "DAY_BOUNDARY"
)

// RECONCILE causes (D6, D7). These are the values the writers use.
//
// They are deliberately *not* a CHECK constraint: D9 specifies "cause enum NOT
// NULL" without enumerating it, and a CHECK cannot be relaxed later without
// rebuilding the table — which the additive rules in schema.go forbid. The
// constraint is therefore enforced by the writer, and the column stores what the
// writer names.
const (
	// ReconcileCauseSnapshotUnavailable — holdings or sellable quantity could
	// not be read at the moment a consumer needed them.
	ReconcileCauseSnapshotUnavailable = "SNAPSHOT_UNAVAILABLE"
	// ReconcileCauseSnapshotStale — they were read, but older than the
	// staleness bound for the consumer that needs them.
	ReconcileCauseSnapshotStale = "SNAPSHOT_STALE"
	// ReconcileCauseQuantityMismatch — the locally derived quantity and the
	// broker snapshot disagree.
	ReconcileCauseQuantityMismatch = "QUANTITY_MISMATCH"
	// ReconcileCauseIdentifierConflict — the same identifier turned up in
	// conflicting contexts.
	ReconcileCauseIdentifierConflict = "IDENTIFIER_CONFLICT"
	// ReconcileCauseAttributionFailed — a broker record could not be attributed
	// to anything local. CANCEL_REJECTED/REPLACE_REJECTED are documented as
	// separate order records (openapi); their exact shape is
	// [형태 미측정 — 2b 2.1], so a failure to recognise one is RECONCILE rather
	// than an outside-order classification.
	ReconcileCauseAttributionFailed = "ATTRIBUTION_FAILED"
)

// schemaV5 is the extended execution contract. One atomic step: every table and
// column below arrives together, because a decision without its reservation, or
// an attempt without its idempotency key, is not a state any code should have to
// handle.
//
// Additive per schema.go's rules — five new tables, nine new columns, no change
// to any released column and no historical row rewritten. There is no
// down-migration: an older binary refuses the newer user_version
// (ErrSchemaTooNew) and recovery is the pre-migration backup taken by
// backup.go.
const schemaV5 = `
-- The issuer's decision, persisted before the Gateway is called. The Gateway
-- reads the preimage back from here and recomputes the hash: a value supplied by
-- the caller is never trusted, which is what makes a forged or swapped decision
-- structurally impossible rather than merely unlikely.
CREATE TABLE decisions (
	id              TEXT PRIMARY KEY,               -- issuer-minted; intents do not exist yet at issue time
	account_ref     TEXT NOT NULL,
	generation      INTEGER NOT NULL DEFAULT 0,     -- reissue counter; always 0 until 2c/2d advance it
	safety_class    TEXT NOT NULL
	                CHECK (safety_class IN ('EXPOSURE_RAISING','RISK_REDUCING','PROTECTION_WEAKENING')),
	preimage_kind   TEXT NOT NULL
	                CHECK (preimage_kind IN ('RISK_INTENT','REDUCTION_INTENT')),
	risk_preimage   TEXT NOT NULL,                  -- canonical JSON, stored verbatim: the hash is over this text
	risk_hash       TEXT NOT NULL,
	client_order_id TEXT,                           -- place only; f(id, generation)
	limits_json     TEXT,                           -- required for EXPOSURE_RAISING, NULL for RISK_REDUCING
	nonce           TEXT NOT NULL UNIQUE,           -- single use; consumption is recorded in spent_nonces
	issued_at       TEXT NOT NULL,
	expires_at      TEXT NOT NULL
) STRICT;

CREATE INDEX idx_decisions_account ON decisions(account_ref, issued_at);
CREATE INDEX idx_decisions_expires ON decisions(expires_at);

-- What a decision holds against the account's limits while it is in flight.
CREATE TABLE risk_reservations (
	id             TEXT PRIMARY KEY,
	decision_id    TEXT NOT NULL REFERENCES decisions(id),
	-- attempt_id is backfilled at Prepare: the reservation is taken before the
	-- attempt exists, and this is the join that lets the attempt's terminal
	-- record release it in the same transaction.
	attempt_id     TEXT REFERENCES mutation_attempts(id),
	account_ref    TEXT NOT NULL,
	kind           TEXT NOT NULL CHECK (kind IN ('OPEN_EXPOSURE','DAILY_LOSS','CASH')),
	amount         TEXT NOT NULL,                   -- decimal string
	currency       TEXT NOT NULL,
	trading_day    TEXT,                            -- DAILY_LOSS only; market-local, YYYY-MM-DD
	snapshot_as_of TEXT NOT NULL,                   -- the as-of the sizing decision was made against
	state          TEXT NOT NULL DEFAULT 'HELD' CHECK (state IN ('HELD','RELEASED')),
	released_at    TEXT,
	release_reason TEXT CHECK (release_reason IS NULL OR release_reason IN
	                 ('BROKER_TERMINAL','EXPIRED_UNCONSUMED','OPERATOR','DAY_BOUNDARY'))
) STRICT;

CREATE INDEX idx_reservations_decision ON risk_reservations(decision_id);
CREATE INDEX idx_reservations_attempt  ON risk_reservations(attempt_id);
-- The restart sweep for orphaned holds, and the per-account exposure sum.
CREATE INDEX idx_reservations_held     ON risk_reservations(account_ref, kind, state);

-- Consumed nonces. The primary key is the single-use guarantee: it survives a
-- restart, unlike any in-memory set.
--
-- decision_id carries no foreign key on purpose. The retention invariant is that
-- a consumption record outlives the longest decision TTL, so pruning expired
-- decisions must not be blocked by (or cascade into) these rows.
CREATE TABLE spent_nonces (
	nonce       TEXT PRIMARY KEY,
	decision_id TEXT,
	consumed_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_spent_nonces_consumed ON spent_nonces(consumed_at);

-- RECONCILE as durable state. The journal is authoritative; the in-memory
-- latches are projections rebuilt from here at startup, so a restart cannot
-- forget that entries are blocked.
CREATE TABLE reconcile_states (
	id            TEXT PRIMARY KEY,
	account_ref   TEXT NOT NULL,
	symbol        TEXT,              -- NULL = account-wide
	cause         TEXT NOT NULL,     -- see the ReconcileCause* constants
	evidence      TEXT,              -- what was observed, for the operator
	entered_at    TEXT NOT NULL,
	released_at   TEXT,
	release_cause TEXT
) STRICT;

-- One active state per scope. Two active rows for one scope would let a release
-- close only one of them and re-open entries while the cause still holds.
CREATE UNIQUE INDEX idx_reconcile_active
	ON reconcile_states(account_ref, symbol) WHERE released_at IS NULL;
-- SQLite treats NULLs as distinct inside a UNIQUE index, so the index above does
-- not constrain the account-wide scope at all. This second one covers it; the
-- two together are "one active state per scope" as written.
CREATE UNIQUE INDEX idx_reconcile_active_account_wide
	ON reconcile_states(account_ref) WHERE released_at IS NULL AND symbol IS NULL;

CREATE INDEX idx_reconcile_account ON reconcile_states(account_ref, entered_at);

-- A correction to an already-observed execution: same cumulative quantity, a
-- different average price or amount. There is no quantity delta, so it is not a
-- fill event; it is its own record.
--
-- The new_* columns are NOT NULL DEFAULT '' rather than nullable because they
-- are part of the UNIQUE key and SQLite treats NULLs as distinct: a nullable
-- column would silently disable the double-insert protection exactly when the
-- broker reported no average price. '' means "not observed", as it does in
-- fill_snapshots.average_price beside it.
CREATE TABLE execution_corrections (
	id                 TEXT PRIMARY KEY,
	account_ref        TEXT NOT NULL,
	order_id           TEXT NOT NULL REFERENCES fill_snapshots(order_id),
	prev_avg_price     TEXT,
	new_avg_price      TEXT NOT NULL DEFAULT '',
	prev_filled_amount TEXT,
	new_filled_amount  TEXT NOT NULL DEFAULT '',
	cumulative_qty     TEXT NOT NULL,               -- decimal string; unchanged by definition
	observed_at        TEXT NOT NULL,
	-- Re-observing the same correction after a crash is one row, not two.
	UNIQUE(order_id, cumulative_qty, new_avg_price, new_filled_amount)
) STRICT;

CREATE INDEX idx_corrections_order ON execution_corrections(order_id, observed_at);

-- mutation_attempts: the attempt's link to the decision that authorised it, and
-- the exact bytes a replay may resend.
--
-- Every column is nullable with no default (or DEFAULT 0), so a v4 row stays
-- valid: an attempt recorded before this migration simply has no decision.
ALTER TABLE mutation_attempts ADD COLUMN decision_id TEXT REFERENCES decisions(id);
ALTER TABLE mutation_attempts ADD COLUMN safety_class TEXT;
ALTER TABLE mutation_attempts ADD COLUMN generation INTEGER;
ALTER TABLE mutation_attempts ADD COLUMN client_order_id TEXT;
-- wire_body is the canonical request body as it was sent, stored immutably at
-- RECORDED. A replay resends these bytes; it never rebuilds a body from
-- structured fields, because a change in the serialisation rules would then send
-- different bytes under the same idempotency key.
ALTER TABLE mutation_attempts ADD COLUMN wire_body TEXT;
ALTER TABLE mutation_attempts ADD COLUMN serializer_version TEXT;
ALTER TABLE mutation_attempts ADD COLUMN replay_count INTEGER DEFAULT 0;
ALTER TABLE mutation_attempts ADD COLUMN last_replay_at TEXT;
-- account_ref is denormalised from the attempt's intent so the partial UNIQUE
-- below can exist at all: D9 puts that constraint on mutation_attempts, and an
-- index cannot reach through the foreign key to intents.account_ref. Additive
-- and nullable; see openspec/changes/extend-execution-contract/issues.md.
ALTER TABLE mutation_attempts ADD COLUMN account_ref TEXT;

-- An idempotency key is claimed by exactly one attempt per account. Attempts
-- with no key — every v4 row, and any path that cannot carry one — are outside
-- the index.
CREATE UNIQUE INDEX idx_attempts_client_order
	ON mutation_attempts(account_ref, client_order_id) WHERE client_order_id IS NOT NULL;

-- fill_snapshots: the previous filled amount, without which an amount-only
-- correction cannot be detected at all.
ALTER TABLE fill_snapshots ADD COLUMN filled_amount TEXT NOT NULL DEFAULT '';
`
