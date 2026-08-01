# Branch Test Map: `Store.CollectClosedStrategyTrades`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil reader rejected | handoff validation test | yes | yes |
| B2 | reader failure wrapped before writes | handoff error test | yes | yes |
| B3 | exact replay produces snapshots without duplicates | replay/restart test | yes | yes |
| B4 | divergent replay fails at current trade | divergence test | yes | yes |
