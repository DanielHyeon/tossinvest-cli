# Branch Test Map: `ReadinessAdapter.Check`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact request and current snapshot | new adapter exact-scope test | currently account/market-only | pending |
| B2 | order type substitution | new adapter substitution test | currently accepted | pending |
| B3 | fractional/non-integral/overflow quantity | execgw pre-dispatch tests | currently reaches provider | pending |
| B4 | session/trigger/replace contract substitution | new adapter substitution test | currently not represented | pending |
| B5 | provider failure/snapshot drift | existing adapter and gateway drift tests | existing | pending |
