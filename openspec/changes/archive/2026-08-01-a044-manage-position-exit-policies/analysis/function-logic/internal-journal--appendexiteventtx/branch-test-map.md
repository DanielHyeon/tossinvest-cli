# Branch Test Map: `appendExitEventTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exit row missing cannot bind generation | `TestLateOldGenerationJudgementIsQuarantined` | yes | yes |
| B2 | complete evaluated tuple is encoded | `exit snapshot persistence suite` | yes | yes |
| B3 | nil effective snapshot stays state-only | `exit state suite` | yes | yes |
| B4 | saved encode failure aborts transaction | `exit snapshot corruption suite` | yes | yes |
| B5 | recomputed encode failure aborts transaction | `exit snapshot corruption suite` | yes | yes |
| B6 | effective encode failure aborts transaction | `exit snapshot corruption suite` | yes | yes |
| B7 | insert failure rolls back state/event together | `TestPositionPolicyLifecycleAndAuditRollbackTogetherOnInjectedCrashPoint` | yes | yes |
