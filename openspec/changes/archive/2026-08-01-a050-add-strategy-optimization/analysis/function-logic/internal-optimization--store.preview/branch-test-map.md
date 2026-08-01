# Branch Test Map: `Store.preview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | current snapshot read | preview lifecycle tests | existing | existing |
| B2 | base version conflict | preview lifecycle tests | existing | existing |
| B3 | invalid category | preview validation tests | existing | existing |
| B4 | invalid source | preview validation tests | existing | existing |
| B5 | invalid reason | preview validation tests | existing | existing |
| B6 | evidence unavailable | evidence test | existing | existing |
| B7 | collect keys | preview lifecycle tests | existing | existing |
| B8 | read-only/mixed category field | `TestPreviewRejectsMixedCategoryChangesWithoutPersistingCandidate` | yes | yes |
| B9 | invalid option | preview validation tests | existing | existing |
| B10 | unchanged field | preview validation tests | existing | existing |
| B11 | risk derivation | lifecycle tests | existing | existing |
| B12 | restart derivation | lifecycle tests | existing | existing |
| B13 | empty changes | preview validation tests | existing | existing |
| B14 | capability token entropy | entropy fault path | n/a | n/a |
| B15 | candidate ID entropy | entropy fault path | n/a | n/a |
| B16 | rollback salt | rollback tests | existing | existing |
| B17 | risk wait | lifecycle tests | existing | existing |
| B18 | canonical payload MAC | payload MAC test | yes | yes |
| B19 | candidate insert failure | DB fault path | n/a | n/a |
| B20 | successful preview | preview lifecycle tests | existing | existing |
| B21 | immutable candidate insert failure returns no preview/capability | DB fault-path review and zero-row preview rejection tests | metadata hardening changed persisted payload | defensive branch reviewed |
