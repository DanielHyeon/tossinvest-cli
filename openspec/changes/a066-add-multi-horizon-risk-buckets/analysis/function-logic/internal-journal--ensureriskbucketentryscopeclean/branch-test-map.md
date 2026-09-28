# Branch Test Map: `ensureRiskBucketEntryScopeClean`

Post-edit (6.5 fix lot, 2026-09-28). Pre-edit AST and map: `analysis/pre-edit/6.5-fixes/internal-journal--ensureriskbucketentryscopeclean/`. Mutation column refers to the mutation-6.5 ledger `analysis/mutation-6.5/fixes-ledger.tsv`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | owner latch read error (275) | storage exit — `TestA066StorageErrorExitsFailClosed` | n/a (storage) | covered (suite rc 0 at the fix tree) |
| B2 | active owner RISK_OVERAGE (280) → cause named | `TestRiskBucket*OverageLatch*` admission tests (existing) | mutation V09 | covered (suite rc 0 at the fix tree) |
| B3 | active owner UNKNOWN_ACTUAL_RISK (283) | existing owner-latch admission tests (e.g. `TestRiskBucketFillHookLatchesBindGapWithoutReturningError` then admission) | — | covered (suite rc 0 at the fix tree) |
| B4 | scope latch read error (287) | storage exit — structural test | n/a | covered (suite rc 0 at the fix tree) |
| B5 | scope latch present (291) | existing scope-latch admission tests (`TestReleasedOwnerLateFillBlocksFirstFreshAdmissionOnlyInExactMarket`) | — | covered (suite rc 0 at the fix tree) |
| B6 | reconcile read error (295) | storage exit — structural test | n/a | covered (suite rc 0 at the fix tree) |
| B7 | active reconcile (300) → "active RECONCILE <cause>" | `TestA066SubmitRevalidationRefusesALatchTakenAfterIssuance/active_reconcile…`, `TestA066IssuedEntryIsRefusedAtSubmitWhenItsScopeLatchedAfterIssuance` | mutation V08 | covered (suite rc 0 at the fix tree) |
| B8 | any cause (303) → `ErrRiskBucketEntryBlocked: market/symbol: causes` | same + existing admission blocked tests | mutation V01, V04 (submit call site) | covered (suite rc 0 at the fix tree) |
