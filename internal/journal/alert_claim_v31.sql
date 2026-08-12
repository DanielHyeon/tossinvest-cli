-- schemaV31 gives alert_outbox a delivery lease, so that two senders reaching
-- one PENDING row leave with exactly one right to send.
--
-- Until now the row carried no notion of ownership. ClaimAlertForDelivery is a
-- name for a thing it did not do: it read the state and answered "yes, a send is
-- owed" to every caller that asked, and the only exclusion in the system was one
-- Go mutex inside a single Notifier. Two senders — two Notifiers, or one
-- Notifier and a Flush — both got that answer and both published.
--
-- Four columns, not one. The token is what every settlement compares against,
-- because a name can be re-acquired by the same loop after a restart and then an
-- old settlement would land on a new lease (ABA). The other three exist so the
-- lease can be judged and read:
--
--   claimed_by       is for the operator and the log: which loop holds it.
--   claimed_at       is read only by the clock-went-backwards predicate. In a
--                    step-back the stored expiry is in the future too, so the
--                    issue *time* is the only thing that separates the cases.
--   claim_expires_at is read by the expiry predicate, and it is stored rather
--                    than derived because the issuer's lease is the authority.
--                    Deriving it at judgement time lets a claimant with a
--                    shorter lease reinterpret somebody else's live claim and
--                    steal a row nobody abandoned.
--
-- Additive per the four rules in schema.go: schemaV3 is untouched, the two TEXT
-- columns carry a DEFAULT and the two timestamps are nullable, nothing is
-- dropped or renamed, and no historical row is rewritten. An existing row comes
-- out of this migration with claim_token = '' and claim_expires_at = NULL, which
-- the predicate reads as "claimable" — exactly what it means today.
--
-- No index. idx_outbox_state(state, id) already selects the PENDING set and the
-- lease predicate is judged only inside it; that set is the undelivered critical
-- alerts, which is structurally small.
ALTER TABLE alert_outbox ADD COLUMN claim_token      TEXT NOT NULL DEFAULT '';
ALTER TABLE alert_outbox ADD COLUMN claimed_by       TEXT NOT NULL DEFAULT '';
ALTER TABLE alert_outbox ADD COLUMN claimed_at       TEXT;
ALTER TABLE alert_outbox ADD COLUMN claim_expires_at TEXT;
