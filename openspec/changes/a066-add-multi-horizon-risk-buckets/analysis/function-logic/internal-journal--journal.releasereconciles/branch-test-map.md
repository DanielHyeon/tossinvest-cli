# Branch Test Map: `Journal.ReleaseReconciles`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty batch | existing batch tests | existing | pending |
| B2 | duplicate same market refused; peer markets distinct | market batch test | pending | pending |
| B3 | missing/wrong/cross-market target rolls all back | atomic and cross-market tests | pending | pending |
| B4 | exact scopes release atomically | atomic release tests | existing | pending |
| B5 | invalid market | market validation tests | yes | yes |
| B6 | invalid release cause | existing validation tests | existing | yes |
| B7 | missing evidence | existing validation tests | existing | yes |
| B8 | invalid expected cause | existing validation tests | existing | yes |
| B9 | duplicate exact scope | duplicate-scope contract | existing | yes |
| B10 | transaction begin | storage contract | existing | yes |
| B11 | preflight loop | atomic release tests | yes | yes |
| B12 | exact scope absent | `TestAtomicMarketReleaseDoesNotCrossIntoPeerMarket` | yes | yes |
| B13 | scan failure | query contract | existing | yes |
| B14 | expected cause mismatch | rollback test | existing | yes |
| B15 | update loop | atomic success test | existing | yes |
| B16 | update error | transaction contract | existing | yes |
| B17 | rows-affected error/count | transaction contract | existing | yes |
| B18 | wrong affected count | transaction contract | existing | yes |
| B19 | commit error | transaction contract | existing | yes |
