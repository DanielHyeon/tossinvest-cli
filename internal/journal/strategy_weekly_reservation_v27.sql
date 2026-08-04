CREATE TABLE strategy_weekly_reservation_scopes (
 campaign_id TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 version INTEGER NOT NULL CHECK(version>=0),
 positive_leg_count INTEGER NOT NULL CHECK(positive_leg_count BETWEEN 0 AND 7),
 updated_at TEXT NOT NULL,
 PRIMARY KEY(campaign_id,market)
) STRICT;

CREATE TABLE strategy_weekly_market_reservations (
 reservation_id TEXT PRIMARY KEY,
 campaign_id TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 stable_week TEXT NOT NULL,
 provider TEXT NOT NULL,
 time_zone TEXT NOT NULL,
 session_date TEXT NOT NULL,
 calendar_generation TEXT NOT NULL,
 calendar_digest TEXT NOT NULL,
 planned_ordinal INTEGER NOT NULL CHECK(planned_ordinal BETWEEN 1 AND 7),
 status TEXT NOT NULL CHECK(status IN ('ACTIVE','CONSUMED','RELEASED')),
 scope_version INTEGER NOT NULL CHECK(scope_version>0),
 observed_at TEXT NOT NULL,
 fresh_until TEXT NOT NULL,
 evaluated_at TEXT NOT NULL,
 request_digest TEXT NOT NULL CHECK(length(request_digest)=64 AND request_digest NOT GLOB '*[^0-9a-f]*'),
 record_digest TEXT NOT NULL UNIQUE CHECK(length(record_digest)=64 AND record_digest NOT GLOB '*[^0-9a-f]*'),
 created_at TEXT NOT NULL,
 updated_at TEXT NOT NULL,
 UNIQUE(campaign_id,market,stable_week),
 FOREIGN KEY(campaign_id,market) REFERENCES strategy_weekly_reservation_scopes(campaign_id,market)
) STRICT;

CREATE UNIQUE INDEX idx_strategy_weekly_one_active_scope
ON strategy_weekly_market_reservations(campaign_id,market)
WHERE status='ACTIVE';

CREATE INDEX idx_strategy_weekly_reservation_scope
ON strategy_weekly_market_reservations(campaign_id,market,stable_week,status);

CREATE TABLE strategy_weekly_reservation_receipts (
 campaign_id TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 idempotency_key TEXT NOT NULL,
 request_digest TEXT NOT NULL CHECK(length(request_digest)=64 AND request_digest NOT GLOB '*[^0-9a-f]*'),
 reservation_id TEXT NOT NULL REFERENCES strategy_weekly_market_reservations(reservation_id),
 created_at TEXT NOT NULL,
 PRIMARY KEY(campaign_id,market,idempotency_key)
) STRICT;

CREATE TABLE strategy_weekly_first_leg_bindings (
 decision_id TEXT PRIMARY KEY REFERENCES strategy_first_leg_bindings(decision_id),
 reservation_id TEXT NOT NULL UNIQUE REFERENCES strategy_weekly_market_reservations(reservation_id),
 campaign_id TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 stable_week TEXT NOT NULL,
 planned_ordinal INTEGER NOT NULL CHECK(planned_ordinal=1),
 scope_version INTEGER NOT NULL CHECK(scope_version>0),
 request_digest TEXT NOT NULL,
 reservation_record_digest TEXT NOT NULL,
 calendar_generation TEXT NOT NULL,
 calendar_digest TEXT NOT NULL,
 binding_record_digest TEXT NOT NULL UNIQUE CHECK(length(binding_record_digest)=64 AND binding_record_digest NOT GLOB '*[^0-9a-f]*'),
 created_at TEXT NOT NULL,
 UNIQUE(campaign_id,market,stable_week)
) STRICT;

CREATE TRIGGER strategy_weekly_reservations_no_delete
BEFORE DELETE ON strategy_weekly_market_reservations
BEGIN SELECT RAISE(ABORT,'weekly market reservations are append-preserved'); END;

CREATE TRIGGER strategy_weekly_receipts_no_update
BEFORE UPDATE ON strategy_weekly_reservation_receipts
BEGIN SELECT RAISE(ABORT,'weekly reservation receipts are immutable'); END;

CREATE TRIGGER strategy_weekly_receipts_no_delete
BEFORE DELETE ON strategy_weekly_reservation_receipts
BEGIN SELECT RAISE(ABORT,'weekly reservation receipts are immutable'); END;

CREATE TRIGGER strategy_weekly_first_leg_bindings_no_update
BEFORE UPDATE ON strategy_weekly_first_leg_bindings
BEGIN SELECT RAISE(ABORT,'weekly first-leg bindings are immutable'); END;

CREATE TRIGGER strategy_weekly_first_leg_bindings_no_delete
BEFORE DELETE ON strategy_weekly_first_leg_bindings
BEGIN SELECT RAISE(ABORT,'weekly first-leg bindings are immutable'); END;
