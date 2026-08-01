# Branch Test Map: `Store.Dashboard`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid as-of/period fails before query | dashboard query-bound tests | a049 baseline | PASS |
| B2 | query/corruption/row-limit failure returns no partial dashboard | dashboard corruption and row-limit tests | a049 baseline | PASS |
| B3 | aggregate parse error returns no partial dashboard | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` | a049 baseline | PASS |
| B4 | every returned trade participates in freshness maximum | source freshness propagation test | freshness absent at `948e721` | PASS |
| B5 | only later persisted source time advances `NewestSourceAt` | source freshness maximum test | freshness absent at `948e721` | PASS |
| B6 | every aggregate participates in insufficient count | first-class state test | a049 baseline | PASS |
| B7 | insufficient aggregate increments exactly once | first-class state test | a049 baseline | PASS |
| B8 | lineage count query error is wrapped | DB/query error coverage | a049 baseline | PASS |
| B9 | every trade is inspected for required observation state | first-class state test | a049 baseline | PASS |
| B10 | every required observation key is inspected until first incomplete key | first-class state test | a049 baseline | PASS |
| B11 | incomplete required observation increments not-measured once per trade | first-class state test | a049 baseline | PASS |
