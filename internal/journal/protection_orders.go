package journal

// schemaV13 is deliberately additive. A protection mutation is recorded before
// dispatch and every result is appended, so an interrupted create/replace/cancel
// remains attributable after restart. The application-level repository applies
// the stricter state-machine and optimistic-CAS checks.
const schemaV13 = `
CREATE TABLE protection_sagas (
  saga_id TEXT PRIMARY KEY,
  account_ref TEXT NOT NULL,
  profile TEXT NOT NULL,
  market TEXT NOT NULL CHECK (market IN ('KR','US')),
  symbol TEXT NOT NULL,
  generation INTEGER NOT NULL CHECK (generation >= 1),
  revision INTEGER NOT NULL CHECK (revision >= 1),
  state TEXT NOT NULL,
  trigger INTEGER NOT NULL CHECK (trigger >= 1),
  quantity INTEGER NOT NULL CHECK (quantity >= 1),
  pending_trigger INTEGER NOT NULL DEFAULT 0,
  pending_quantity INTEGER NOT NULL DEFAULT 0,
  client_order_id TEXT NOT NULL UNIQUE,
  attempt_id TEXT NOT NULL DEFAULT '',
  broker_id TEXT NOT NULL DEFAULT '',
  previous_broker_id TEXT NOT NULL DEFAULT '',
  reconcile_reason TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL,
  last_event_kind TEXT NOT NULL DEFAULT '',
  last_event_fingerprint TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_protection_sagas_account_symbol
  ON protection_sagas(account_ref, profile, market, symbol, state);
CREATE UNIQUE INDEX idx_protection_sagas_live_claim
  ON protection_sagas(account_ref, profile, market, symbol)
  WHERE state NOT IN ('TRIGGERED','CLOSED');

CREATE TABLE protection_mutation_attempts (
  attempt_id TEXT PRIMARY KEY,
  saga_id TEXT NOT NULL REFERENCES protection_sagas(saga_id),
  generation INTEGER NOT NULL CHECK (generation >= 1),
  kind TEXT NOT NULL CHECK (kind IN ('CREATE','REPLACE','CANCEL')),
  state TEXT NOT NULL CHECK (state IN ('PLANNED','DISPATCHED','ACKNOWLEDGED','IN_DOUBT','CLOSED')),
  serializer_version INTEGER NOT NULL,
  canonical_body TEXT NOT NULL,
  target_broker_id TEXT NOT NULL DEFAULT '',
  result_broker_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_protection_attempts_saga
  ON protection_mutation_attempts(saga_id, generation, created_at);
`
