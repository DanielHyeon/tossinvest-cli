package journal

// core_domain.go holds schema v6: the position, adjustment, operating-mode, exit
// and outcome tables the core domain records (change add-core-domain, design.md
// D7 — that table is normative and this file is its transcription).
//
// # What the tables are for
//
// v5 recorded what the engine was *entitled* to attempt and what that
// entitlement consumed. v6 records what the attempts turned into and what is
// protecting it:
//
//   - positions — the projection of fills and adjustments (D4). It is a
//     projection and not an authority: nothing writes a quantity here that did
//     not come from a fill event or a recorded adjustment, and the row carries
//     `entry_decision_id` so "why does this position exist" is an explicit join
//     rather than a time-window guess (position-ledger: Provenance Lineage).
//   - position_adjustments — append-only compare-and-append corrections from
//     reconciliation. `expected_prev_quantity` is what makes it a compare: an
//     adjustment computed against a quantity that has since moved is discarded
//     rather than applied on top of a different world.
//   - operating_modes — append-only mode history; the current mode is the latest
//     row (D3). Append-only because a mode that can be overwritten cannot be
//     audited, and the conservative direction is automatic while relaxation
//     needs a person.
//   - exit_states — one row per position, the whole of the local protection
//     decision (D5): the t0 baseline (the entry stop, NOT NULL — a position with
//     no baseline is a position with no stop), the observed high-water probe,
//     the ratchet level or ladder rung, and the pending proposal that stops the
//     same level being proposed twice.
//   - exit_events — append-only judgement history. Every observation that moved
//     something leaves a row, which is the provenance join between a position
//     and the intents its protection proposed.
//   - trade_outcomes — frozen at CLOSED (D7, trade-analytics). Frozen, because
//     realised R computed later from a moving denominator is not the number the
//     decision was judged by.
//
// # Conventions inherited from schema.go
//
// Decimals are TEXT (binary floats cannot hold a broker's decimals exactly),
// timestamps are RFC3339 UTC text from the injected clock, tables are STRICT,
// and primary keys are caller-supplied stable ids except in the append-only
// event logs, which follow fill_events and attempt_transitions with an
// autoincrementing integer. Nothing here rewrites a historical row.
//
// # Broker-behaviour claims
//
// None. Every value in this file is TossOS-internal: the broker's contribution
// reaches these tables only as a `broker_as_of` stamp on an adjustment and as
// the fills the projection is built from.

// Position states (D7, position-ledger 포지션 상태기계). The transition table
// itself is task 6.1; what the schema pins is that no other value is a state.
const (
	PositionFlat    = "FLAT"
	PositionOpening = "OPENING"
	PositionOpen    = "OPEN"
	PositionScaling = "SCALING"
	PositionClosing = "CLOSING"
	PositionClosed  = "CLOSED"
)

// Adjustment kinds (D4). UNKNOWN is not a placeholder for "we did not look": it
// is the recorded verdict that a difference could not be attributed, which is a
// different fact from "external" and must stay distinguishable in the audit
// trail.
const (
	AdjustmentExternal = "EXTERNAL"
	AdjustmentManual   = "MANUAL"
	AdjustmentUnknown  = "UNKNOWN"
)

// Operating modes (D3). Three, not four: EXIT_ONLY was dropped because it
// behaved identically to ENTRY_BLOCKED, and a rung of the approval ladder that
// changes nothing is a rung nobody can justify climbing.
const (
	ModeNormal       = "NORMAL"
	ModeEntryBlocked = "ENTRY_BLOCKED"
	ModeHaltAll      = "HALT_ALL"
)

// Mode actors (D3). The asymmetry the column exists for: AUTO may only tighten
// (NORMAL → ENTRY_BLOCKED), and every relaxation carries OPERATOR plus a cause.
const (
	ModeActorAuto     = "AUTO"
	ModeActorOperator = "OPERATOR"
)

// Exit policy kinds (D5). One policy per position: the two are alternative
// execution paths in the original and must never contest one baseline.
const (
	ExitPolicyRatchet = "RATCHET"
	ExitPolicyLadder  = "LADDER"
)

// Ratchet levels (D5, exit-policy Baseline Ratchet). NONE is the level a
// position opens at — the baseline is already the entry stop at t0, so NONE
// means "no ratchet step has been taken", never "no protection".
const (
	RatchetNone        = "NONE"
	RatchetHalfRisk    = "HALF_RISK"
	RatchetBreakeven   = "BREAKEVEN"
	RatchetPartialLock = "PARTIAL_LOCK"
	RatchetProfitLock  = "PROFIT_LOCK"
)

// schemaV6 is the core domain. One atomic step, exactly as D7 specifies: a
// position without its exit state, or an exit state without somewhere to record
// its judgements, is not a state any code should have to handle.
//
// Additive per schema.go's rules — six new tables, no change to any released
// column and no historical row rewritten. There is no down-migration: an older
// binary refuses the newer user_version (ErrSchemaTooNew) and recovery is the
// pre-migration backup taken by backup.go.
//
// Where D7 names a column without spelling out its nullability, the
// transcription resolves it the way the v5 transcription did — an enum with a
// CHECK, or a column with a DEFAULT, is NOT NULL, because SQLite's CHECK passes
// on NULL and a nullable defaulted column can still be written NULL explicitly.
// The resolutions are enumerated in
// openspec/changes/add-core-domain/issues.md.
const schemaV6 = `
-- The projection of fills and adjustments. Not an authority: every quantity
-- here is derived, and the derivation runs inside the fill transaction's apply
-- hook (apply_hook.go) so the projection and the fill it came from commit
-- together or not at all.
CREATE TABLE positions (
	id                TEXT PRIMARY KEY,
	account_ref       TEXT NOT NULL,
	market            TEXT NOT NULL,
	symbol            TEXT NOT NULL,
	-- instance_seq separates a re-entry from the position it re-enters: CLOSED is
	-- final, and the next entry on the same symbol is a new instance with its own
	-- decision, not a resurrection of the old row.
	instance_seq      INTEGER NOT NULL,
	-- NULL is the external/manual position: it was folded in by an adjustment, no
	-- decision justifies it, and it is therefore not an exit-policy target (there
	-- is no entry stop to make a baseline out of).
	entry_decision_id TEXT REFERENCES decisions(id),
	state             TEXT NOT NULL
	                  CHECK (state IN ('FLAT','OPENING','OPEN','SCALING','CLOSING','CLOSED')),
	quantity          TEXT NOT NULL,             -- decimal string
	avg_price         TEXT NOT NULL,             -- decimal string
	-- Nullable: a row may exist before anything filled into it, and closed_at is
	-- by definition absent until the instance reaches CLOSED.
	opened_at         TEXT,
	closed_at         TEXT,
	UNIQUE(account_ref, market, symbol, instance_seq)
) STRICT;

CREATE INDEX idx_positions_symbol   ON positions(account_ref, market, symbol, state);
-- The provenance join: decision → … → position.
CREATE INDEX idx_positions_decision ON positions(entry_decision_id);

-- Append-only compare-and-append corrections. Rows are never updated: the
-- history of what reconciliation changed *is* the justification for the
-- projection disagreeing with the fills.
CREATE TABLE position_adjustments (
	id                     TEXT PRIMARY KEY,
	position_id            TEXT NOT NULL REFERENCES positions(id),
	kind                   TEXT NOT NULL CHECK (kind IN ('EXTERNAL','MANUAL','UNKNOWN')),
	-- The compare half of compare-and-append: the quantity the adjustment was
	-- computed against. Re-checked inside the commit transaction, and a mismatch
	-- discards the adjustment rather than applying it to a moved position.
	expected_prev_quantity TEXT NOT NULL,        -- decimal string
	prev_quantity          TEXT NOT NULL,        -- decimal string
	new_quantity           TEXT NOT NULL,        -- decimal string
	-- '' means "the broker reported no cost basis", the same convention
	-- fill_snapshots.average_price uses beside it.
	prev_avg_price         TEXT NOT NULL DEFAULT '',
	new_avg_price          TEXT NOT NULL DEFAULT '',
	broker_as_of           TEXT NOT NULL,        -- when the account snapshot was current
	evidence               TEXT,                 -- what was observed, for the operator
	created_at             TEXT NOT NULL
) STRICT;

CREATE INDEX idx_position_adjustments_position ON position_adjustments(position_id, created_at);

-- Append-only mode history. The current mode is the latest row for the account;
-- readers order by (created_at, rowid) because journal timestamps are
-- second-resolution and two transitions inside one second must still have an
-- order.
CREATE TABLE operating_modes (
	id          TEXT PRIMARY KEY,
	account_ref TEXT NOT NULL,
	mode        TEXT NOT NULL CHECK (mode IN ('NORMAL','ENTRY_BLOCKED','HALT_ALL')),
	-- Every transition names why. An automatic tightening names its trigger and
	-- an operator's relaxation names the approval; a mode change nobody can
	-- explain afterwards is not auditable (§0.7).
	cause       TEXT NOT NULL,
	actor       TEXT NOT NULL CHECK (actor IN ('AUTO','OPERATOR')),
	created_at  TEXT NOT NULL
) STRICT;

CREATE INDEX idx_operating_modes_account ON operating_modes(account_ref, created_at);

-- The local protection state of one position.
--
-- baseline_price is NOT NULL with no default, which is the whole of D5's first
-- correction: the baseline is the entry decision's stop from t0, so there is no
-- window in which a position is open and unprotected while it waits for the
-- first ratchet trigger.
CREATE TABLE exit_states (
	position_id       TEXT PRIMARY KEY REFERENCES positions(id),
	policy_kind       TEXT NOT NULL CHECK (policy_kind IN ('RATCHET','LADDER')),
	entry_price       TEXT NOT NULL,             -- decimal string
	initial_stop      TEXT NOT NULL,             -- decimal string; the t0 baseline's source
	-- initial_risk is entry − initial_stop, frozen at entry. It is the denominator
	-- of every R in this row and it does not move when a partial take-profit or an
	-- adjustment changes the quantity.
	initial_risk      TEXT NOT NULL,             -- decimal string
	baseline_price    TEXT NOT NULL,             -- decimal string; monotone non-decreasing
	-- The observed high-water mark, updated on every observation. It is the
	-- maximum of the *samples*, not the true high: a trigger brushed between two
	-- observations is missed, and the observation period is what defines that
	-- tolerance (exit-policy, stated as a limit rather than hidden).
	high_water        TEXT NOT NULL,             -- decimal string; monotone non-decreasing
	ratchet_level     TEXT NOT NULL
	                  CHECK (ratchet_level IN ('NONE','HALF_RISK','BREAKEVEN','PARTIAL_LOCK','PROFIT_LOCK')),
	active_rung       INTEGER,                   -- LADDER only; NULL under RATCHET
	-- Moved only inside the fill transaction's apply hook (apply_hook.go). The
	-- ratio is of the *initial* quantity; each individual proposal's ratio is of
	-- the remaining quantity (the original's denominator rule).
	taken_ratio_total TEXT NOT NULL DEFAULT '0', -- decimal string
	-- The pending proposal, or all three NULL. pending_level carries the RATCHET
	-- level or, under LADDER, the rung index: one column, because a position has
	-- one policy and the two can never be pending at once.
	pending_action    TEXT,
	pending_level     TEXT,
	-- No foreign key: the proposal is armed before the intent exists, which is
	-- exactly what makes a crash between arming and submitting recoverable.
	pending_intent_id TEXT,
	completed         INTEGER NOT NULL DEFAULT 0,
	updated_at        TEXT NOT NULL
) STRICT;

-- The crash-recovery scan: which positions have a proposal that was armed and
-- may never have been submitted.
CREATE INDEX idx_exit_states_pending ON exit_states(pending_intent_id)
	WHERE pending_action IS NOT NULL;
-- The observation loop's working set.
CREATE INDEX idx_exit_states_open ON exit_states(completed) WHERE completed = 0;

-- Append-only judgement history: what was observed, what it implied, and which
-- intent it proposed. This is the provenance join between a position and the
-- exits its protection asked for.
--
-- The primary key is an autoincrementing integer, following fill_events and
-- attempt_transitions: an observation log has no caller-minted stable id to key
-- on, and its order is the thing being recorded.
CREATE TABLE exit_events (
	id                 INTEGER PRIMARY KEY AUTOINCREMENT,
	position_id        TEXT NOT NULL REFERENCES positions(id),
	observed_price     TEXT,                     -- decimal string; NULL when the judgement had no price
	high_water         TEXT,                     -- decimal string, after this observation
	baseline_after     TEXT,                     -- decimal string, after this observation
	level_after        TEXT,                     -- ratchet level or rung index, after
	action             TEXT,                     -- what was proposed, empty for a no-op judgement
	proposed_intent_id TEXT,
	created_at         TEXT NOT NULL
) STRICT;

CREATE INDEX idx_exit_events_position ON exit_events(position_id, id);

-- Frozen at CLOSED, inside the transaction that closes the position. Frozen
-- because realised R recomputed later from a denominator that has moved is not
-- the number the decision was judged against.
CREATE TABLE trade_outcomes (
	position_id              TEXT PRIMARY KEY REFERENCES positions(id),
	realized_pnl_after_costs TEXT NOT NULL,      -- decimal string
	-- Realised R, not price R: the two are named apart on purpose (D5), because a
	-- partially taken position's price R and its realised R are different numbers.
	realized_r               TEXT NOT NULL,      -- decimal string
	initial_risk             TEXT NOT NULL,      -- decimal string; the frozen denominator
	initial_quantity         TEXT NOT NULL,      -- decimal string; the frozen ratio base
	held_seconds             INTEGER,
	exit_ratchet_level       TEXT,               -- RATCHET only
	exit_rung                INTEGER,            -- LADDER only
	closed_at                TEXT NOT NULL
) STRICT;

-- The 180-day retention sweep.
CREATE INDEX idx_trade_outcomes_closed ON trade_outcomes(closed_at);
`
