# Weekly reservation fill lifecycle — pre-edit map

## Function Logic Map

### `Journal.RecordFill` / `applyWeeklyReservationLifecycleInTx`

1. Resolve the fill as a locally owned broker order before touching weekly state.
2. Join broker order → dispatch lease → first-leg decision → weekly reservation under exact account/market/symbol identity.
3. If cumulative filled quantity is positive and reservation is ACTIVE, atomically set CONSUMED, increment `positive_leg_count` once, and advance scope version.
4. If broker state is terminal with exact zero cumulative fill, atomically set RELEASED, retain the count, and advance scope version.
5. Append an immutable idempotency receipt keyed by the fill observation identity.
6. Replay of a consumed/released reservation is a no-op only when its durable state is already compatible; ambiguous or contradictory bindings fail closed.

## Branch Test Map

| Branch | Expected |
|---|---|
| First positive partial fill | ACTIVE→CONSUMED, count +1 exactly once |
| Duplicate/later positive fill | CONSUMED remains, no second increment |
| Terminal zero-fill | ACTIVE→RELEASED, count unchanged |
| Restart then next stable-week reservation | uses advanced scope version and next ordinal |
| Non-weekly or broker-only order | no weekly state mutation |
| Cross-market/order ambiguity | fill transaction refuses rather than guessing |
