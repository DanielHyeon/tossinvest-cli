# Function Logic and Branch Test Map — new `internal/strategyrouter` package

This change adds a new package only. No existing scheduler, budget, registry or owner function is edited, so
there is no base function body to map. The following map records the new high-risk boundaries before runtime
integration.

## `Route`

1. Validate canonical owner key and expected owner revision.
2. Verify the sealed owner snapshot, exact key/revision and trusted evaluation freshness.
3. Reject every cross-scope row, including inactive rows; count active same-key rows across all horizons.
4. Reject multiple active owners.
5. Verify the sealed exact-market record and expected market revision; require READY lifecycle.
6. Preserve one active ON owner before candidate scoring.
7. Otherwise validate every candidate scope/evidence, filter eligible ON candidates, refuse ties, and return
   one pure decision with zero mutations.

Branches: malformed/stale/drifted snapshot, cross-generation row, multiple active owner, durable OFF, market
revision drift, cross-market binding, existing owner preservation, no candidate, tie and deterministic winner.

## `CASMarketRecord` / `RollbackMarketRecord`

1. Validate the complete two-market state and sealed target record.
2. Compare only the selected market revision and lock.
3. Require exactly the next revision and replace only that market.
4. Rollback accepts sealed OFF targets only and records a new monotonic revision.

Branches: KR and US concurrent success, stale one-market conflict, peer revision preservation, old-or-new crash
state, OFF rollback success and historical ON rollback refusal.

## `MigrateLegacy`

1. Duplicate migration version returns the exact prior result.
2. Explicit disabled maps to two OFF records.
3. Verified non-combined exact single-market evidence copies only that sealed record.
4. Every unknown, combined, corrupt or unverified form maps to two OFF records plus typed refusal.

Branches: disabled, verified KR, verified US, combined, unverified and retry convergence.

## `QuotaAuthority.Acquire` / `Complete`

1. Install one sealed physical endpoint/reset-generation snapshot.
2. Compute allowance as the minimum of reported remaining minus reserve, observation-cycle cap and absolute cap.
3. Check freshness using the authority clock; caller timestamps cannot extend it.
4. Bind issued capability to endpoint/reset/market/horizon/poll/coordinator/request while incrementing one shared
   issued/outstanding set.
5. Duplicate identical requests return the same capability without increment; mismatched/replayed scope or
   reset cannot change commitment.

Branches: four market/horizon subscopes exhaust one counter, independent endpoints, concurrent final slot,
duplicate acquire, cross-scope completion, reset replay, completion replay, stale backdating and reserve survival.

## Authority and dormant invariants

Owner snapshots, ON market records and physical quota snapshots have package-private constructors and private
seals. External-package tests prove public struct literals cannot forge them. Public defaults produce only
sealed OFF/UNOBSERVED KR or US state. Static dependency tests reject broker, journal, activation, toggle,
campaign or owner mutation authority and runtime package imports.
