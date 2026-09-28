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
| B7 | active reconcile rows counted (count, not cause text) → "active RECONCILE <causes>" (re-review 2) | `TestA066SubmitRevalidationRefusesALatchTakenAfterIssuance/active_reconcile…`, `TestA066IssuedEntryIsRefusedAtSubmitWhenItsScopeLatchedAfterIssuance` | mutation V08 | covered (suite rc 0 at the fix tree) |
| B8 | any cause (303) → `ErrRiskBucketEntryBlocked: market/symbol: causes` | same + existing admission blocked tests | mutation V01, V04 (submit call site) | covered (suite rc 0 at the fix tree) |

Re-review repair (2026-09-28, commit after 28629ec6). The reconcile block now depends on `count(*)`; the cause text only names it. A new B8 handles an empty cause ("(no cause recorded)"). Post-edit there are 9 branches; old B8 → B9. Tests: `TestA066EntryScopeRuleBlocksOnEveryActiveReconcileRow` (RED at 28629ec6, `pre-edit/6.5-fixes/red2-at-28629ec6.log`; mutation R01 CAUGHT) and `TestA066SubmitRevalidationRefusesOwnerAndScopeLatches` (owner and scope-latch branches at the submit call site; mutation R03 CAUGHT). Ledger: `mutation-6.5/rereview-ledger.tsv`.
