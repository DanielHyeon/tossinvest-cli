-- schemaV34 gives each reservation the exact reservation-policy record it was sized with (a066 task 5.6.1, F2).
--
-- Until v33, risk_bucket_policies held ONE record per (dimension, value, policy_version), and that record carries the
-- reservation pricing (worst price, FX, fee, evaluation time). A bucket shared by several entries (strategy, horizon,
-- sector, market) therefore accepted only the first entry's pricing; every later entry at another price was refused as
-- an "immutable policy collision" (measured 2026-09-27: price 5 then 6 refused). That also hid F1 (a stale shared
-- snapshot admitted past the cap), because the collision refused first.
--
-- Additive only:
--   * risk_bucket_policy_records keeps every distinct record, keyed by the record digest the writers already compute
--     (riskBucketRecordDigest over {Key, PolicyEvidence, ReservePolicy}); rows are immutable.
--   * risk_bucket_reservations gains nullable policy_record_digest. Rows written before v34 keep NULL and keep reading
--     their policy from risk_bucket_policies (which was unique per key for them). Rows written from v34 must carry it,
--     it must name an existing record of the same bucket key, and it can never change.
--   * risk_bucket_policies is unchanged. It stays the FK parent of risk_bucket_snapshots; a v34 writer still inserts the
--     first record there (INSERT OR IGNORE) but no longer treats a different record under the same key as a collision.

CREATE TABLE risk_bucket_policy_records (
 bucket_dimension TEXT NOT NULL, bucket_value TEXT NOT NULL, policy_version TEXT NOT NULL, record_digest TEXT NOT NULL,
 policy_digest TEXT NOT NULL, policy_source TEXT NOT NULL, policy_observed_at TEXT NOT NULL, policy_fresh_until TEXT NOT NULL,
 account_currency TEXT NOT NULL, quote_currency TEXT NOT NULL, evaluated_at TEXT NOT NULL,
 worst_price_quote TEXT NOT NULL, price_source TEXT NOT NULL, price_version TEXT NOT NULL, price_digest TEXT NOT NULL, price_observed_at TEXT NOT NULL, price_fresh_until TEXT NOT NULL,
 fee_fixed_base_minor TEXT NOT NULL, fee_per_unit_base_minor TEXT NOT NULL, fee_minimum_base_minor TEXT NOT NULL, fee_version TEXT NOT NULL, fee_digest TEXT NOT NULL,
 fx_rate_quote_to_base TEXT NOT NULL, fx_haircut TEXT NOT NULL, fx_source TEXT NOT NULL, fx_version TEXT NOT NULL, fx_digest TEXT NOT NULL, fx_observed_at TEXT NOT NULL, fx_fresh_until TEXT NOT NULL,
 created_at TEXT NOT NULL,
 PRIMARY KEY(bucket_dimension,bucket_value,policy_version,record_digest)
) STRICT;
CREATE TRIGGER risk_bucket_policy_records_no_update BEFORE UPDATE ON risk_bucket_policy_records
BEGIN SELECT RAISE(ABORT,'risk bucket policy records are immutable'); END;
CREATE TRIGGER risk_bucket_policy_records_no_delete BEFORE DELETE ON risk_bucket_policy_records
BEGIN SELECT RAISE(ABORT,'risk bucket policy records are immutable'); END;

ALTER TABLE risk_bucket_reservations ADD COLUMN policy_record_digest TEXT;

-- A v34 reservation names its record, and the record belongs to the same bucket key.
CREATE TRIGGER risk_bucket_reservations_policy_record_required BEFORE INSERT ON risk_bucket_reservations
WHEN NEW.policy_record_digest IS NULL OR NOT EXISTS (
 SELECT 1 FROM risk_bucket_policy_records r
 WHERE r.bucket_dimension=NEW.bucket_dimension AND r.bucket_value=NEW.bucket_value
   AND r.policy_version=NEW.policy_version AND r.record_digest=NEW.policy_record_digest)
BEGIN SELECT RAISE(ABORT,'risk bucket reservation must name its exact policy record'); END;

-- The binding is set once. Fill accounting updates held/filled/state on the same row and must not move it.
CREATE TRIGGER risk_bucket_reservations_policy_record_immutable BEFORE UPDATE OF policy_record_digest ON risk_bucket_reservations
WHEN OLD.policy_record_digest IS NOT NEW.policy_record_digest
BEGIN SELECT RAISE(ABORT,'risk bucket reservation policy record is immutable'); END;
