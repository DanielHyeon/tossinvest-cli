# Branch Test Map: `ExitObserver.record`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | executable proposal vs state-only | 1-share/state-only contract tests | yes | pending |
| B2 | collision clear required | existing clear-before-liquidate tests | yes | pending |
| B3 | clear error | existing error test | yes | pending |
| B4 | not cleared | delay/hold test | yes | pending |
| B5 | deterministic/fallback intent id | a041 identity tests | yes | pending |
| B6 | atomic record error | v10 fault matrix | yes | pending |
| B7 | no proposal | state-only persistence test | yes | pending |
| B8 | proposal count | existing cycle count test | yes | pending |
| B9 | submit after commit | crash ordering + emergency isolation test | yes | pending |
| B10 | quote timestamp source | persisted read-model test | yes | yes |
| B11 | cycle timestamp fallback | existing quote fallback test | yes | yes |
