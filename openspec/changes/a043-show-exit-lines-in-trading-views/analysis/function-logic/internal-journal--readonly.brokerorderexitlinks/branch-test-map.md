# Branch Test Map: `ReadOnly.BrokerOrderExitLinks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty scopes skip SQL | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | old API queried globally | yes |
| B2 | scope count over bound | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | old API unbounded | yes |
| B3 | validate/copy each scope | scoped link tests | RED compile on new API | yes |
| B4 | reject unsupported market | scope validation test | RED compile on new API | yes |
| B5 | reject zero-length id/account but preserve whitespace-only id | `TestBrokerOrderExitLinksTreatsBrokerOrderIDsAsOpaqueBytes` | `TrimSpace` collapsed distinct ids | yes |
| B6 | reject invalid trading day | scope validation tests | RED compile on new API | yes |
| B7 | JSON encoding failure guard | typed values make failure unreachable | reviewed | yes |
| B8 | SQL query failure | missing-schema bounded test | existing | yes |
| B9 | initialize one exact result per scope | market/day/opaque collision tests | old bare-id map collided | yes |
| B10 | scan bounded linear result rows | all focused lineage tests | old N+1 query | yes |
| B11 | sentinel row limit | oversized/row-bound contract | old unbounded query | yes |
| B12 | scan failure | corruption/read tests | existing | yes |
| B13 | returned request identity mismatch | opaque/composite collision tests | trimmed id mismatch | yes |
| B14 | exact CONFIRMED PLACE exists on current/ancestor node | direct/amend/state tests | overly broad attribution | yes |
| B15 | classify unsafe lineage/evidence | lineage tests | old no lineage | yes |
| B16 | exact JSON-array cycle | `TestBrokerOrderExitLinksFailsClosedForPureSingleParentCycle` | branch fixture stopped before cycle | yes |
| B17 | multiple valid parents | 1,000-parent branching test | old no lineage | yes |
| B18 | raw edge exists outside validated scope | cross-account/incomplete edge tests | old unscoped join | yes |
| B19 | further parent exists at depth 32 | depth-bound test | no pure depth fixture | yes |
| B20 | duplicate confirmed PLACE evidence | duplicate attempt tests | old first-row choice | yes |
| B21 | duplicate exit events | duplicate event test | old event choice | yes |
| B22 | node has no event | unlinked/ancestry traversal tests | existing | yes |
| B23 | invalid event hydration | corruption tests | old N+1 hydration | yes |
| B24 | rows terminal error | focused journal tests | existing | yes |
| B25 | finalize every requested scope | collision tests | old result omitted misses | yes |
| B26 | unsafe or duplicate evidence | confirmed-state/wide/cycle/depth tests | incomplete evidence was accepted | yes |
| B27 | generic ambiguity fallback | duplicate event/attempt tests | multiple candidates were selected | yes |
| B28 | attach sole canonical evidence | opaque path/market/direct/amend tests | delimiter path and trimmed IDs corrupted identity | yes |
