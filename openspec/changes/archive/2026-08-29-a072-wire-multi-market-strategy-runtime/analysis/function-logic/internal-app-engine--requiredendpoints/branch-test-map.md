# Branch Test Map: RequiredEndpoints

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Shared startup catalog excludes strategy-only exchange-rate GET while retaining account and mutation dependencies | `TestStrategyOnlyFXReadIsNotAGlobalEngineStartupDependency`; soak/live coverage tests | production deployment and RED tests failed with the GET present | yes |
