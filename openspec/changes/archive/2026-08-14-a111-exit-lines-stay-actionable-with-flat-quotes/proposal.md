## Why

A110 removed the false account-wide reconciliation block and the first clean production cycle adopted all five previously unmanaged holdings, but the same incident proved that management alone does not make the promised exit lines usable. Three newly adopted positions whose first official quote equalled the adoption quote remained `SEED/not_evaluated_yet`, while evaluated positions became `stale` after 30 seconds because unchanged observations were discarded even though the exit loop continued to receive fresh official prices.

## What Changes

- Require the first valid exit observation after an exit state is opened to persist one canonical `EVALUATED` snapshot even when price, high-water, protection, rung and action are unchanged.
- Separate “domain state changed” from “a fresh observation was evaluated”: unchanged later observations must refresh the authoritative effective snapshot without creating a proposal, order, state transition, or duplicate semantic exit event.
- Keep observation failures fail-closed. A missing, invalid, future or source-stale official quote must not refresh snapshot freshness; after the existing 30-second bound its prior line is hidden.
- Retire only a077's running-engine age bypass now that `last_observed_at` becomes a real successful-observation heartbeat: a positively stopped engine is stale immediately, and every other liveness state applies the same persisted 30-second bound.
- Make console and HTTP positions projections share that freshness matrix and the same actionable current/next line values.
- Preserve adoption's no-order boundary, monotone protection/high-water/rung state, pending proposal deduplication, rate-budget priority and all LIVE/order gates.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `exit-policy`: successful flat-price observations must create the initial evaluated snapshot and subsequently refresh observation evidence without inventing a state transition or order.
- `operator-console`: managed positions with continuously fresh official observations must retain actionable current and next exit lines; `SEED` and stale evidence remain hidden only when no valid evaluation/fresh observation exists.
- `http-api-service`: `/api/v1/positions` must expose the same fresh canonical exit-line verdict as the console for flat-price managed positions.

## Impact

- Production paths: `internal/app/engine/exitloop.go`, journal exit-snapshot persistence/event semantics, shared `internal/operatorview` freshness inputs, and the console/httpapi positions adapters.
- Tests: unchanged first quote for both ratchet and ladder/adopted paths, repeated identical observations, invalid/stale quote controls, restart persistence, 30-second boundary, console/API parity, and zero proposal/order/mutation side effects.
- Operations: no schema migration is expected; the deployed A110 image remains healthy and the released reconciliation state remains valid while A111 is developed and reviewed.
