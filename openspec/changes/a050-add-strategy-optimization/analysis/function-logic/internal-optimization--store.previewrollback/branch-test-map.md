# Branch Test Map: `Store.previewRollback`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | corrupt/missing current snapshot fails before candidate | snapshot corruption coverage | actor/digest hardening absent at baseline | PASS |
| B2 | stale base returns version conflict | rollback conflict coverage | baseline lifecycle | PASS |
| B3 | invalid target version cannot be previewed | rollback target coverage | baseline lifecycle | PASS |
| B4 | every current desired key enters union | `TestRollbackCreatesANewVersionAndNeverRewritesHistory` | baseline lifecycle | PASS |
| B5 | every target desired key enters union | same rollback test | baseline lifecycle | PASS |
| B6 | every union key is inspected | same rollback test | baseline lifecycle | PASS |
| B7 | inactive historical key fails closed | historical key regression | baseline lifecycle | PASS |
| B8 | unchanged or other-category key is excluded | rollback category/no-op coverage | baseline lifecycle | PASS |
