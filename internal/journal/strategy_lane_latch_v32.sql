-- schemaV32 gives a strategy lane's entry latch a life longer than the process.
--
-- Until now a lane latch was a bool in memory (`internal/strategyworker/lane.go`).
-- A restart cleared it, so a lane that latched for a reason that still holds came
-- back open, and the only way an operator could reopen a latched lane was to
-- restart the engine. That is backwards: restarting is the accident, not the
-- decision.
--
-- Two append-only tables, because the recovery is the interesting record. A
-- single mutable "latched" flag would let the fact that a lane was ever latched
-- disappear the moment it recovered, and the operator's question is usually
-- "why did this lane stop, and who reopened it".
--
-- The recovery condition is not "a cycle succeeded". `design.md` puts
-- "panic/unexpected return/repeated threshold" on the row whose effect is
-- "lane effective entry latch OFF, recovery evidence 필요", and a lane that
-- reopens because its next cycle happened to pass has produced no evidence
-- about the reason it latched. The evidence this schema demands is a strictly
-- newer signed activation generation: `scheduler.Activation.Generation()` comes
-- from an ed25519-signed manifest verified against a trusted key and a pinned
-- digest, so it cannot advance without a human replacing that file.

CREATE TABLE strategy_lane_latches (
 latch_seq INTEGER PRIMARY KEY AUTOINCREMENT,
 account_ref TEXT NOT NULL CHECK(account_ref<>''),
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 family TEXT NOT NULL CHECK(family IN ('CONTINUATION','REVERSAL','WEEKLY_VALUE','BREAKOUT_RETEST')),
 lane_id TEXT NOT NULL CHECK(lane_id<>''),
 lane_version TEXT NOT NULL CHECK(lane_version<>''),
 latch_id TEXT NOT NULL CHECK(latch_id<>''),
 latch_revision INTEGER NOT NULL CHECK(latch_revision>0),
 reason TEXT NOT NULL CHECK(reason<>''),
 abnormal INTEGER NOT NULL CHECK(abnormal IN (0,1)),
 -- The signed activation generation observed when this lane latched. Zero is
 -- allowed and means "latched while this market carried no verified activation
 -- manifest"; entry is already impossible in that state, and any real manifest
 -- then satisfies the strictly-greater recovery condition below.
 activation_generation INTEGER NOT NULL CHECK(activation_generation>=0),
 observed_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_strategy_lane_latch_lane
 ON strategy_lane_latches(account_ref,market,family,lane_id,lane_version,latch_seq);

CREATE TABLE strategy_lane_latch_recoveries (
 latch_seq INTEGER PRIMARY KEY REFERENCES strategy_lane_latches(latch_seq),
 activation_generation INTEGER NOT NULL CHECK(activation_generation>0),
 observed_at TEXT NOT NULL
) STRICT;

-- Both tables are history. Nothing rewrites or removes a latch or its recovery.
CREATE TRIGGER strategy_lane_latch_no_update BEFORE UPDATE ON strategy_lane_latches
BEGIN SELECT RAISE(ABORT,'strategy lane latches are immutable'); END;
CREATE TRIGGER strategy_lane_latch_no_delete BEFORE DELETE ON strategy_lane_latches
BEGIN SELECT RAISE(ABORT,'strategy lane latches cannot be deleted'); END;
CREATE TRIGGER strategy_lane_latch_recovery_no_update BEFORE UPDATE ON strategy_lane_latch_recoveries
BEGIN SELECT RAISE(ABORT,'strategy lane latch recoveries are immutable'); END;
CREATE TRIGGER strategy_lane_latch_recovery_no_delete BEFORE DELETE ON strategy_lane_latch_recoveries
BEGIN SELECT RAISE(ABORT,'strategy lane latch recoveries cannot be deleted'); END;

-- One open latch per lane. The lane's own state machine already refuses to mint
-- a second latch record while it holds one, so that the operator keeps the
-- FIRST cause rather than the last; this trigger is where that rule survives a
-- process boundary, which is the only place it could otherwise be lost.
CREATE TRIGGER strategy_lane_latch_first_cause_wins BEFORE INSERT ON strategy_lane_latches
WHEN EXISTS (
 SELECT 1 FROM strategy_lane_latches open_latch
 LEFT JOIN strategy_lane_latch_recoveries recovered ON recovered.latch_seq=open_latch.latch_seq
 WHERE open_latch.account_ref=NEW.account_ref
   AND open_latch.market=NEW.market
   AND open_latch.family=NEW.family
   AND open_latch.lane_id=NEW.lane_id
   AND open_latch.lane_version=NEW.lane_version
   AND recovered.latch_seq IS NULL)
BEGIN SELECT RAISE(ABORT,'strategy lane already holds an open entry latch'); END;

-- Recovery needs evidence, and the evidence is a strictly newer signed
-- activation generation. Passing back the same generation that was observed when
-- the lane latched proves nothing: it is the state the lane latched in.
CREATE TRIGGER strategy_lane_latch_recovery_needs_newer_activation BEFORE INSERT ON strategy_lane_latch_recoveries
WHEN NOT EXISTS (
 SELECT 1 FROM strategy_lane_latches latched
 WHERE latched.latch_seq=NEW.latch_seq
   AND NEW.activation_generation>latched.activation_generation)
BEGIN SELECT RAISE(ABORT,'strategy lane latch recovery needs a strictly newer signed activation generation'); END;
