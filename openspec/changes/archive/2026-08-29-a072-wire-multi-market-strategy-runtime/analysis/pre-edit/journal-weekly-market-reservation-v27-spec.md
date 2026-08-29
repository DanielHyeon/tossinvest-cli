# Journal v27 weekly market reservation — pre-edit specification

## Scope and invariant

KR and US weekly-value reservations ship in the same migration and API. The durable uniqueness key is
`(campaign_id, market, stable_market_week_identity)`. Calendar generation or digest corrections cannot
change that key. At most one `ACTIVE` reservation may exist for a `(campaign_id, market)` scope, and a
committed weekly key can never be inserted again under another reservation ID.

This is strategy admission state, not broker or LIVE authority. Creating it cannot create a decision,
risk reservation, dispatch lease, order intent or broker request.

## Additive schema

Schema v27 adds three STRICT tables without modifying v26 rows:

- `strategy_weekly_reservation_scopes`: CAS version and positive-leg count per campaign/market;
- `strategy_weekly_market_reservations`: immutable calendar/key identity plus terminal status; and
- `strategy_weekly_reservation_receipts`: idempotency fingerprint and replay receipt.

Unique constraints cover reservation ID, canonical campaign/market/week, idempotency key, and the one
active reservation per campaign/market rule. Older builds refuse the v27 journal as too new.

## Reserve transaction and read projection

`ReserveWeeklyMarket` validates bounded canonical identities, market-specific official provider/timezone,
Monday session date, stable ISO-week identity, point-in-time freshness, ordinal 1..7 and expected scope
version. One `BEGIN IMMEDIATE` transaction performs idempotent receipt replay, CAS, active-conflict check,
reservation insert, scope advance and receipt insert. Mismatch or conflict rolls back every row.

The read-only journal handle exposes one exact reservation projection only at schema v27+. It cannot
reserve, release or consume. The weekly lane converts that projection into its existing sealed in-memory
`ReservationState`, revalidating every field before q_candidate evaluation.

The subsequent q_final/first-leg transaction must bind this reservation key before task 3.24 is complete;
the reservation cannot be treated as consumed merely because a proposal exists.

## RED matrix

- One migration preserves v26 row fingerprints and creates all KR/US reservation objects.
- KR and US reserve successfully in the same test wave and reconstruct exact projections after restart.
- Same key under a corrected calendar generation is rejected; idempotent same request replays.
- Concurrent winners for the same scope yield exactly one durable row.
- Cross-market reservations do not collide.
- A v26 read projection refuses missing v27 support, and a v26 build refuses a v27 journal.
