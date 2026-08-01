# Branch Test Map: `Journal.RecordStrategyDecisionAndReserve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed/cross-bound RiskIntent rejected | divergent binding tests | draft accepted raw JSON | pass |
| B2 | failure at final lineage insert leaves zero partial rows | `TestStrategyPlanRollbackLeavesNoPartialRows` | split transaction | pass |
| B3 | exact replay converges, divergent replay collides | plan exact/idempotent/collision tests | weak cardinality | pass |
| B4 | atomic commit survives restart | migration/lineage restart tests | missing v14 | pass |
| B5 | lineage completeness/manifest binding | production issuance binding tests | weak draft | pass |
| B6 | transaction begin failure propagates | journal failure suite | missing | pass |
| B7 | reservation freshness/version precheck | reservation/recollection suite | baseline | pass |
| B8 | decision insert fails | rollback test | baseline | pass |
| B9 | reservation rows fail | rollback test | baseline | pass |
| B10 | strategy decision exact insert/collision | divergent replay tests | missing | pass |
| B11 | strategy attempt exact insert/collision | divergent replay tests | missing | pass |
| B12 | DISPATCH_START exact insert/collision | divergent execution test | missing | pass |
| B13 | commit failure propagates | journal durability suite | baseline | pass |
| B14 | success returns reservation result | production atomic success | missing | pass |
| B15 | success returns canonical strategy receipt | production atomic success | missing | pass |
