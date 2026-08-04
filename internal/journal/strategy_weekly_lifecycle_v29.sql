CREATE TABLE strategy_weekly_reservation_lifecycle_receipts (
 event_id TEXT PRIMARY KEY,
 reservation_id TEXT NOT NULL REFERENCES strategy_weekly_market_reservations(reservation_id),
 from_status TEXT NOT NULL CHECK(from_status='ACTIVE'),
 to_status TEXT NOT NULL CHECK(to_status IN ('CONSUMED','RELEASED')),
 scope_version INTEGER NOT NULL CHECK(scope_version>0),
 cumulative_quantity TEXT NOT NULL,
 record_digest TEXT NOT NULL UNIQUE CHECK(length(record_digest)=64 AND record_digest NOT GLOB '*[^0-9a-f]*'),
 observed_at TEXT NOT NULL
) STRICT;

CREATE TRIGGER strategy_weekly_lifecycle_receipts_no_update
BEFORE UPDATE ON strategy_weekly_reservation_lifecycle_receipts
BEGIN SELECT RAISE(ABORT,'weekly reservation lifecycle receipts are immutable'); END;

CREATE TRIGGER strategy_weekly_lifecycle_receipts_no_delete
BEFORE DELETE ON strategy_weekly_reservation_lifecycle_receipts
BEGIN SELECT RAISE(ABORT,'weekly reservation lifecycle receipts are immutable'); END;
