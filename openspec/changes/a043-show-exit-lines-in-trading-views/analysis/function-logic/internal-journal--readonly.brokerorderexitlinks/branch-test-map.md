# Branch Test Map: `ReadOnly.BrokerOrderExitLinks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty scopes skip SQL | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | old API queried globally | yes |
| B2 | scope count over bound | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | old API unbounded | yes |
| B3 | normalize each scope | scoped link tests | RED compile on new API | yes |
| B4 | missing id/account | scope validation tests | RED compile on new API | yes |
| B5 | invalid trading day | scope validation tests | RED compile on new API | yes |
| B6 | JSON encoding failure guard | typed values make failure unreachable | reviewed | yes |
| B7 | SQL query failure | missing-schema bounded test | existing | yes |
| B8 | initialize one result per scope | collision test | old bare-id map collided | yes |
| B9 | scan bounded result rows | all focused lineage tests | old N+1 query | yes |
| B10 | sentinel row limit | oversized/row-bound contract | old unbounded query | yes |
| B11 | scan failure | corruption/read tests | existing | yes |
| B12 | returned request identity mismatch | composite collision test | old bare-id lookup | yes |
| B13 | engine or descendant origin | direct/amend tests | old global id set | yes |
| B14 | unsafe-lineage switch | lineage tests | old no lineage | yes |
| B15 | cycle | cycle/branch test | old no lineage | yes |
| B16 | multiple parents | cycle/branch test | old no lineage | yes |
| B17 | account/day mismatch | cross-account edge test | old unscoped join | yes |
| B18 | depth limit | bounded lineage contract | old no lineage | yes |
| B19 | no event on node | unlinked engine order test | old global materialization | yes |
| B20 | invalid event hydration | corruption tests | old N+1 hydration | yes |
| B21 | rows terminal error | focused journal tests | existing | yes |
| B22 | finalize each requested scope | collision test | old result omitted misses | yes |
| B23 | unsafe or duplicate evidence | ambiguity/cycle tests | old fuzzy result | yes |
| B24 | select generic ambiguity reason | duplicate event test | old event choice | yes |
| B25 | attach the sole event/attempt | direct and amend tests | old N+1 lookup | yes |
