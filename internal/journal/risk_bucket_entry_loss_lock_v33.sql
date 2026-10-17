-- schemaV33 gives a066's horizon×market entry loss lock a durable home (a066 task 5.5).
--
-- The lock is entry-only: it is read by the q_final admission transaction and by
-- the Gateway's q_final revalidation, and by nothing on a risk-reducing path.
-- Stop, emergency exit, reconciliation and fill detection never touch this table.
--
-- Additive only. No existing table, index or trigger is changed.
--
-- Append-only history. The activation is the conservative direction and may be
-- recorded immediately (risk-management: 보수 방향 lock 활성화는 즉시 영속할 수
-- 있지만 완화는 사람 승인과 audit를 요구해야 한다). Relaxation is deliberately
-- absent: its approval flow is a pending user decision (a066 review.md, 결정 ⑤
-- 기록 "완화 경로는 … 사용자에게 묻는다"), and it will be a separate append-only
-- record next to this table, never an UPDATE or DELETE of a lock row. Until it
-- exists every recorded lock is in force.

CREATE TABLE risk_bucket_entry_loss_locks (
 lock_seq INTEGER PRIMARY KEY AUTOINCREMENT,
 account_ref TEXT NOT NULL CHECK(account_ref<>''),
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 horizon TEXT NOT NULL CHECK(horizon IN ('SHORT','MEDIUM')),
 cause TEXT NOT NULL CHECK(cause<>''),
 activated_at TEXT NOT NULL
) STRICT;

-- History. Nothing rewrites or removes a lock.
CREATE TRIGGER risk_bucket_entry_loss_lock_no_update BEFORE UPDATE ON risk_bucket_entry_loss_locks
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss locks are immutable'); END;
CREATE TRIGGER risk_bucket_entry_loss_lock_no_delete BEFORE DELETE ON risk_bucket_entry_loss_locks
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss locks cannot be deleted'); END;

-- One lock per account×market×horizon: the first cause is the one an operator
-- needs, so a repeated activation must not bury it. The journal API returns the
-- existing lock instead of inserting; this trigger is where the rule survives a
-- writer that does not go through that API.
CREATE TRIGGER risk_bucket_entry_loss_lock_first_cause_wins BEFORE INSERT ON risk_bucket_entry_loss_locks
WHEN EXISTS (
 SELECT 1 FROM risk_bucket_entry_loss_locks existing
 WHERE existing.account_ref=NEW.account_ref
   AND existing.market=NEW.market
   AND existing.horizon=NEW.horizon)
BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock already active for this scope'); END;
