# Branch Test Map: `ExitObserver.record`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | executable proposal vs state-only | 1-share/state-only contract tests | yes | yes |
| B2 | collision clear required | existing clear-before-liquidate tests | yes | yes |
| B3 | clear error | existing error test | yes | yes |
| B4 | not cleared: exact snapshot persists, arm reason audited | `TestUnclearedWorkingOrderPersistsExactSnapshotAndTypedArmSuppression` + delay/hold test | yes | yes |
| B5 | deterministic/fallback intent id | a041 identity tests | yes | yes |
| B6 | atomic record error | v10 fault matrix | yes | yes |
| B7 | no proposal | state-only persistence test | yes | yes |
| B8 | proposal count | existing cycle count test | yes | yes |
| B9 | submit after commit | crash ordering + emergency isolation test | yes | yes |
| B10 | quote timestamp source | persisted read-model test | yes | yes |
| B11 | cycle timestamp fallback | existing quote fallback test | yes | yes |
| B12 | saved-monotone supersedes an orderable recomputation | journal durable-result test plus engine full suite | yes | yes |
| B13 | durable result is not `armed` or has nil proposal | no submit/count increment | yes | yes |
