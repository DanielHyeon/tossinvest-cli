## Why

`/positions` shows every selected holding as `RECONCILE_BLOCKED` even though desired and running automatic-adoption settings are both ON. A broker order with no local mutation attempt was persisted as a non-terminal fill snapshot, later re-read as an engine-owned open order, and its disappearance from the broker open list promoted a false symbol mismatch into an account-wide permanent block.

## What Changes

- Make the durable tracked-order set prove engine ownership from exactly one confirmed local mutation attempt or recorded replacement lineage in the same account, market-local trading day, symbol, and side before a non-terminal snapshot can be treated as local open exposure.
- Preserve broker-only orders as external observations without allowing them to become missing-engine-order mismatches on later cycles.
- Keep confirmed engine orders, including acknowledged-before-first-poll and amended lineage cases, tracked until a broker-derived terminal state is durably recorded.
- Resolve amendment lineage only through confirmed local evidence in the same canonical scope, including across the schema-v16 migration boundary and opaque parent/child identifier reuse.
- Prevent startup reservation recovery from releasing risk headroom because another account, market, trading day, symbol, or side reused an order identifier.
- Make every active journal RECONCILE cause authoritative for automatic-adoption gating and the `/positions` runtime projection, not only the quantity tracker's subset.
- Prevent a failed durable reconcile release from clearing the in-memory tracker or entry gate, and stop the current cycle before adoption whenever reconcile persistence fails.
- Add regression coverage from an external open order through disappearance, reconciliation comparison, and permanent-promotion prevention.
- Recover the already-persisted false block only after the corrected build obtains a fresh, stable, non-blocking authoritative comparison; record the operator-authorized release evidence and restart the engine projection.
- Do not place, amend, cancel, or preview an order and do not change adoption, Guardian, lane, kill-switch, gate, or operating-mode configuration.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `fill-detection`: Follow-up OrderByID tracking is limited to orders whose local ownership is proven; broker-only observations remain external.
- `reconciliation`: An external broker order must remain external across polling cycles and must never become a missing local order merely because its observation was stored.
- `exit-policy`: Automatic adoption is blocked by every authoritative covering journal RECONCILE state and fails closed when that projection cannot be read or updated.

## Impact

- Code: `internal/journal` tracked-order projection, `internal/reconcile` durable state ordering, `internal/app/engine` adoption/runtime projection, a local `tossctl engine reconcile-resolve` recovery command, and integration tests.
- Runtime data: schema v16 additively scopes new `fill_snapshots` rows by account, trading day, and side; existing rows are retained without destructive backfill, and ownership is derived from journal evidence.
- Operations: one audited operator release after a corrected, fresh comparison, followed by an engine restart so the in-memory tracker is rebuilt from released journal state.
- Safety: entry blocking remains fail-closed for every genuinely local order and every quantity mismatch; exit immediacy and live-order paths are unchanged.
