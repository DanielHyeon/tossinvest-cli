# Branch Test Map: `Store.queryTrades`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | complete-only scoped CTE | filter parity test | yes | yes |
| B2 | market/lane predicates scoped | filter parity test | yes | yes |
| B3 | indexed latest snapshot per filtered trade | actual Dashboard EXPLAIN test | yes | yes |
| B4 | absent metric becomes not_measured | missing state test | existing | existing |
| B5 | 10,001 sentinel bounds read | row-limit test | yes | yes |
| B6 | present metrics populate fixed maps | dashboard metric tests | no — existing | yes |
| B7 | markout aggregation uses cost-adjusted column | cost-adjusted dashboard test | no — existing | yes |
| B8 | source and version remain paired | provenance test | no — existing | yes |
| B9 | iterator error propagates | DB error contract | no — defensive | yes |
| B10 | deterministic order builds bounded slice | dashboard aggregate tests | no — existing | yes |
