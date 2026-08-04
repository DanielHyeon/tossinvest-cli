# Branch Test Map: `NewContext`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing manifest pin | production assembly default test | paired UNWIRED | yes |
| B2 | protection DB collision | `TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly` | former supervisor dependency | yes |
| B3 | normal close/error cleanup | engine lifecycle tests | existing journal cleanup | yes |
