# Branch Test Map: `Store.AppendTrade`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid trade refused | validation suite | existing | yes |
| B2 | new row appends | collect/store tests | existing | yes |
| B3 | exact replay skips | restart/concurrent replay test | yes | yes |
| B4 | divergence fails closed | divergent replay test | yes | yes |
| B5 | crash/commit publishes none or whole | phase SIGKILL test | yes | yes |
