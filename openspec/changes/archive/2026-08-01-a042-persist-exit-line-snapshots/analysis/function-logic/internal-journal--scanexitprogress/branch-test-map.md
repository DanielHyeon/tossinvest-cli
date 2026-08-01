# Branch Test Map: `scanExitProgress`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no exit state | existing not-found test | yes | pending |
| B2 | SQL/scan failure | transaction fault test | yes | pending |
| B3 | NULL/valid snapshot data | legacy/partial/full v10 tests | yes | pending |
| B4 | invalid stored JSON | trailing JSON corruption test | yes | yes |
| B5 | valid stored snapshot | whole-candidate recovery test | yes | yes |
