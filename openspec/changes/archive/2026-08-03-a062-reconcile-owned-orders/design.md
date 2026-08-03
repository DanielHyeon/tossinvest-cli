## Context

Desired and effective adoption are both ON, but the live journal has carried an account-wide permanent `QUANTITY_MISMATCH` since 2026-07-31. The first symbol evidence names an order that has no local mutation attempt or lineage. Its only local evidence is a broker-observed `fill_snapshots` row with `terminal=0`. `TrackedFillOrders` currently returns every non-terminal snapshot, so the next reconciliation treats that external order as engine-owned and its absence from the broker open list as a blocking missing order.

The same investigation found two adjacent authority defects. The adoption driver and `/positions` runtime projection read only `Tracker.Blocks()`, while the journal and entry gate can contain other active RECONCILE causes. Also `Tracker.Observe` mutates memory and clears the gate before a durable release succeeds, and `RunOnce` continues into adoption after that error.

## Goals / Non-Goals

**Goals:**

- Keep broker-only orders external across fill polling and reconciliation cycles.
- Preserve tracking for every locally confirmed or lineage-owned order.
- Use active journal RECONCILE rows as the shared authority for adoption and operator display.
- Make release ordering fail closed and provide a local, explicit, audited recovery command.
- Clear the existing false block only after corrected code proves a fresh stable comparison has no blocking diff.

**Non-Goals:**

- No live order placement, cancellation, amendment, or protection-order change.
- No automatic release of permanent blocks merely because a later observation agrees.
- No web/HTTP reconciliation mutation route and no text-entry UI.
- No change to adoption, Guardian, lane, kill-switch, gate, or operating-mode settings.

## Decisions

### D1. Ownership is positive, canonically scoped journal evidence

A non-terminal fill snapshot enters `TrackedFillOrders` only when its order identifier is linked to exactly one confirmed local mutation attempt or recorded replacement lineage in the same account, market-local trading day, symbol, and side. The existing confirmed-attempt-without-snapshot branch remains. We will not infer ownership from symbol, time, account presence, or from the fact that the detector stored the broker observation. Ambiguous or conflicting ownership enters `IDENTIFIER_CONFLICT` in the same journal transaction and performs no projection, hook, or reservation release.

Schema v16 additively stores `account_ref`, `trading_day`, and `side` on new fill snapshots and indexes the canonical scope. The following migration makes that canonical scope part of snapshot identity instead of retaining `order_id` as a lossy single key, so two active orders that reuse one opaque identifier cannot overwrite one another. It also adds `scoped_lineage_edges`, because the legacy lineage table's global parent/child uniqueness cannot represent the same opaque pair reused in another account or trading session. New amendments dual-write legacy-compatible and scoped lineage in the confirmation transaction; the scoped resolver merges validated historical lineage with v16 evidence and refuses multiple successors. Historical fill rows are retained with empty scope. A terminal order identifier reused on a later market-local trading day starts a new cumulative sequence; US trading days are derived in New York time and KR trading days in Seoul time.

Legacy empty-scope snapshots are migration evidence, not a wildcard. They bind only when exactly one order-creating confirmed intent or validated lineage endpoint strictly predates the evidence commit and remains unique within the same account and market across trading days. The same temporal ownership rule applies to fully scoped broker observations: storing an external observation first cannot make it local when a later engine order reuses the identity. Cross-market reuse cannot suppress an otherwise authoritative legacy terminal snapshot, and cross-day reuse cannot promote one legacy non-terminal snapshot into multiple lineage scopes.

The detector and reconciler carry the same canonical identity. Open-list deduplication and lineage lookup cannot use `order_id` alone, and the corrected recovery comparison cannot consider a broker order a match unless all broker-visible canonical fields agree. Missing broker scope is not permission to declare a clean comparison when multiple local candidates reuse one identifier.

Schema v18 extends the same identity to appended fill events and execution corrections. Schema v19 leaves append-only fill events unchanged and writes a separate durable binding only during migration, only when exactly one intent's confirmed transition (`settled_at`) strictly predates the observation; two intents claiming the same canonical identity remain unbound. Equality in released second-resolution history is not treated as causal proof. New order-creating confirmations and fill commits serialize their fixed-width timestamps strictly after prior canonical evidence/ownership, so the strict comparison remains durable across restart without changing the released timestamp format. Scoped reads and all later provenance/outcome joins compose the migration companion with natively scoped events and require exact durable scope, order-creating confirmed ownership (`PLACE` or `AMEND`, never `CANCEL`), strict temporal precedence, and one unique intent owner. Compatibility reads by opaque order id fail closed when reused or when unbound legacy evidence coexists with a canonical scope. Partial new scope is rejected rather than being stored under an asymmetric key that could overwrite a compatibility row or re-emit cumulative deltas.

Alternative rejected: delete external snapshots. They remain useful observations and destructive cleanup would lose evidence without fixing the ownership boundary.

### D2. Active journal states are the adoption-block authority

The tracker will project every active state for the configured account. Quantity mismatches retain their adjusted-reconcile behavior. Causes owned by other producers are projected as operator-held blocks and cannot be cleared by the quantity comparison. The engine runtime endpoint uses this same projection, so `/positions` and the adoption gate cannot disagree.

Alternative rejected: consult only `EntryGate`. The gate carries a decision but not the complete scope, evidence, age, and journal cause needed by adoption and the operator view.

### D3. Persistence precedes release visibility

`Tracker.Observe` may conservatively expose a newly found block when persistence fails, but it must not remove an existing block or clear a gate until the journal release commits. `ReconcileDriver.RunOnce` stops before adoption on any tracker persistence error. This keeps the failure direction closed without delaying exits.

### D4. Recovery is a local command under the engine lock

`tossctl engine reconcile-resolve` is available only as a local CLI operation. It requires explicit confirmation, operator identity, and a note. It acquires the same journal-directory lock as the engine, constructs the official read-only engine context, obtains three agreeing snapshots at the existing two-second interval, derives corrected local state, and refuses unless the diff has no blocking quantity or missing-order findings. It then validates that every active cause is a quantity mismatch and atomically records `OPERATOR` release evidence for all matching scopes before clearing memory.

The command has no broker mutation interface and adds no console or HTTP route. During deployment the engine is stopped, the command is run once with the authorization already given in this conversation, and the corrected engine is started again.

### D5. Runtime recovery remains explicit

The command releases current journal blocks; startup `Restore` rebuilds the in-memory tracker and entry gate from the released rows. Automatic adoption then runs on the next stable cycle. A new false block cannot recur from the orphan external snapshot because D1 removes it from local engine-owned open orders.

## Risks / Trade-offs

- [An engine order lacks a direct attempt because of amendment lineage] → accept either confirmed-attempt ownership or explicit lineage ownership only within the same canonical account/day/symbol/side scope and retain lineage regression tests.
- [A broker reuses an order identifier] → bind snapshots to market-local trading day, start a new cumulative sequence only after the prior order is terminal, and fail closed on same-scope ambiguity.
- [A broker reuses an amendment parent/child pair] → retain one scoped audit row per confirmed attempt and resolve only canonical-scope, confirmed-AMEND lineage; merge legacy and v16 evidence so migration boundaries cannot hide a conflict.
- [A broker reuses an identifier after fills or corrections exist] → keep fill events and corrections in schema-v18 scoped streams; use exact-scope provenance/outcome joins and make order-id-only compatibility reads fail closed when ambiguous.
- [External evidence predates the first engine owner] → require the actual order-creating confirmed transition (`settled_at`) to strictly precede snapshots/events/corrections even for an exact canonical match; equal legacy timestamps are ambiguous, while new writes serialize a durable strict order. Schema v19 leaves later-owner, equal-time, and multi-intent legacy evidence permanently unbound without rewriting append-only history.
- [A confirmed CANCEL echoes the target order id] → exclude CANCEL attempts from every order-owner set so a normal cancellation cannot create a second owner, poison live-order lookup, or suppress an emergency exit.
- [A startup sweep sees a terminal snapshot from another scope] → run reservation recovery before the engine can decide or nonce retention can delete evidence; require exact snapshot/intent/reservation scope and decision binding plus a single intent owner before releasing one reservation; otherwise keep every hold and record an identifier conflict.
- [Nonce retention encounters an old but still-held reservation] → retain the spent nonce while any reservation for its decision remains `HELD`, so a later restart cannot misclassify the request as never sent.
- [A release command could clear a real mismatch] → require engine exclusion, fresh stable official snapshots, a non-blocking corrected comparison, explicit confirmation, operator identity, note, and durable audit evidence.
- [A storage failure occurs while releasing] → preserve memory/gate blocks and abort before adoption.
- [A journal cause is unknown to this build] → fail closed; do not render it as harmless or release it implicitly.
- [Broader high-risk blast radius] → separate implementation ownership, Function Logic Maps for every edited existing function, race/full-suite verification, and independent security/test review.

## Migration Plan

1. Deploy the ownership and authority fixes; startup performs the existing local reservation recovery before any engine decision and before spent-nonce retention.
2. Stop the engine and acquire its journal lock through the recovery command.
3. Run fresh stable official reads and refuse if the corrected comparison blocks.
4. Record the operator-authorized releases with evidence.
5. Restart the corrected engine and wait for a stable reconciliation/adoption cycle.
6. Verify `/positions` shows managed exit states and stored lines for KR and US holdings.

Rollback: restore the prior binary while retaining journal history. Released rows remain auditable; if a real disagreement appears, the ordinary tracker immediately enters a new block. No toggle or broker state is changed by rollback.

## Open Questions

None. The live evidence, existing release contract, and user authorization determine the recovery path.
