-- schemaV35: a066 task 5.5 relaxation (user decision 2026-09-28, design D8).
--
-- Automatic paths only tighten. A relaxation is an OPERATOR act with an approval reference and an audit line written
-- before commit, and it is recorded here as append-only history — never an UPDATE or DELETE of a lock row.
--
-- Additive, with one trigger replacement: v33's first-cause-wins trigger (one lock per scope, forever) cannot coexist
-- with releases, so it is dropped here and replaced by "at most one OPEN lock per scope". The v33 file itself is not
-- changed: a committed migration is a record.

-- A lock activation that arrives while a lock is already open for the scope is recorded as a REAFFIRM event. A release
-- must name the last event its approver saw; a later REAFFIRM makes that release stale (a concurrent tightening wins).
CREATE TABLE risk_bucket_entry_loss_lock_events (
 event_seq INTEGER PRIMARY KEY AUTOINCREMENT,
 lock_seq INTEGER NOT NULL REFERENCES risk_bucket_entry_loss_locks(lock_seq),
 kind TEXT NOT NULL CHECK(kind IN ('REAFFIRM')),
 cause TEXT NOT NULL CHECK(cause<>''),
 recorded_at TEXT NOT NULL
) STRICT;
CREATE TRIGGER risk_bucket_entry_loss_lock_events_no_update BEFORE UPDATE ON risk_bucket_entry_loss_lock_events
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock events are immutable'); END;
CREATE TRIGGER risk_bucket_entry_loss_lock_events_no_delete BEFORE DELETE ON risk_bucket_entry_loss_lock_events
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock events cannot be deleted'); END;

-- One release per lock, by an operator, with the approval and reason that make it auditable.
CREATE TABLE risk_bucket_entry_loss_lock_releases (
 release_seq INTEGER PRIMARY KEY AUTOINCREMENT,
 lock_seq INTEGER NOT NULL UNIQUE REFERENCES risk_bucket_entry_loss_locks(lock_seq),
 actor TEXT NOT NULL CHECK(actor='OPERATOR'),
 approval TEXT NOT NULL CHECK(trim(approval)<>''),
 reason TEXT NOT NULL CHECK(trim(reason)<>''),
 expected_last_event INTEGER NOT NULL CHECK(expected_last_event>=0),
 released_at TEXT NOT NULL
) STRICT;
CREATE TRIGGER risk_bucket_entry_loss_lock_releases_no_update BEFORE UPDATE ON risk_bucket_entry_loss_lock_releases
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock releases are immutable'); END;
CREATE TRIGGER risk_bucket_entry_loss_lock_releases_no_delete BEFORE DELETE ON risk_bucket_entry_loss_lock_releases
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock releases cannot be deleted'); END;

DROP TRIGGER risk_bucket_entry_loss_lock_first_cause_wins;
CREATE TRIGGER risk_bucket_entry_loss_lock_one_open BEFORE INSERT ON risk_bucket_entry_loss_locks
WHEN EXISTS (
 SELECT 1 FROM risk_bucket_entry_loss_locks existing
 WHERE existing.account_ref=NEW.account_ref
   AND existing.market=NEW.market
   AND existing.horizon=NEW.horizon
   AND NOT EXISTS (SELECT 1 FROM risk_bucket_entry_loss_lock_releases r WHERE r.lock_seq=existing.lock_seq))
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock already open for this scope'); END;

-- RISK_OVERAGE latch releases. The latch itself lives in flags (risk_bucket_owners / risk_bucket_reservations); the
-- release clears the flags of one owner generation in the same transaction that writes this record, reseals the owner
-- state and writes the audit line. The record binds the state digest its approver read.
CREATE TABLE risk_bucket_latch_releases (
 release_seq INTEGER PRIMARY KEY AUTOINCREMENT,
 account_ref TEXT NOT NULL CHECK(account_ref<>''),
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL CHECK(symbol<>''),
 prospective_generation TEXT NOT NULL CHECK(prospective_generation<>''),
 latch TEXT NOT NULL CHECK(latch='RISK_OVERAGE'),
 expected_state_digest TEXT NOT NULL CHECK(expected_state_digest<>''),
 actor TEXT NOT NULL CHECK(actor='OPERATOR'),
 approval TEXT NOT NULL CHECK(trim(approval)<>''),
 reason TEXT NOT NULL CHECK(trim(reason)<>''),
 released_at TEXT NOT NULL
) STRICT;
CREATE TRIGGER risk_bucket_latch_releases_no_update BEFORE UPDATE ON risk_bucket_latch_releases
BEGIN SELECT RAISE(ABORT,'risk bucket latch releases are immutable'); END;
CREATE TRIGGER risk_bucket_latch_releases_no_delete BEFORE DELETE ON risk_bucket_latch_releases
BEGIN SELECT RAISE(ABORT,'risk bucket latch releases cannot be deleted'); END;
