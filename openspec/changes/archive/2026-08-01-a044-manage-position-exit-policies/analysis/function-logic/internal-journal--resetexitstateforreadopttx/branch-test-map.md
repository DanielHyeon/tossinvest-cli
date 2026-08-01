# Branch Test Map: `resetExitStateForReadoptTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid price/stop refuses reset | `TestPositionPolicyReadoptResetsExitStateAtFreshT0` | yes | yes |
| B2 | missing position source rolls back | `position policy journal suite` | yes | yes |
| B3 | unknown policy identity refuses | `position policy validation suite` | yes | yes |
| B4 | update failure rolls back | `TestPositionPolicyLifecycleAndAuditRollbackTogetherOnInjectedCrashPoint` | yes | yes |
| B5 | row-count read failure rolls back | `position policy journal suite` | yes | yes |
| B6 | no open exit state refuses readopt | `TestPositionPolicyReadoptRejectsIneligibleHolding` | yes | yes |
