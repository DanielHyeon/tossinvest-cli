# Branch Test Map: `BudgetCoordinator.observeLocked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | existing endpoint enters ordered update path | out-of-order/correction tests | existing | yes |
| B2 | timestamp relation switch selects older/equal/newer | observation ordering tests | existing | yes |
| B3 | older evidence returns unchanged | `TestOlderBudgetObservationCannotReplaceNewerEvidence` | existing | yes |
| B4 | equal timestamp enters correction path | equal-time tests | existing | yes |
| B5 | equal-time provenance mismatch marks conflict and returns | conflicting correction test | existing | yes |
| B6 | identical correction can only lower remaining | conservative correction test | existing | yes |
| B7 | equal trusted same-window cycle may reconcile | chronology tests | existing | yes |
| B8 | newer trusted evidence enters window classification | reset/parser tests | forged semantics could gain trust | yes |
| B9 | classification switch distinguishes initial/same/next/conflict | reset relation tests | existing | yes |
| B10 | missing prior trusted reset initializes anchor | invalid-then-valid tests | existing | yes |
| B11 | same window reconciles cycle-covered completions and retains earliest deadline | drift/chronology tests | existing | yes |
| B12 | next window enters causal-authority path | new-window tests | manual path could advance | yes |
| B13 | nonnil cycle reconciles watermark-covered commitments | held-response tests | prior chronology finding | yes |
| B14 | only nonnil cycle with zero remaining commitments advances generation | manual-cap-bypass/cap tests | empty map substituted for authority | yes |
| B15 | conflicting window marks endpoint fail-closed and returns | reset conflict tests | existing | yes |
| B16 | new endpoint records trusted reset only for valid evidence | initial/missing tests | existing | yes |
