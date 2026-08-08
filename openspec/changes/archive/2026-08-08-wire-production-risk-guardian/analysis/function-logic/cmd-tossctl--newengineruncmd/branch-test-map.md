# Branch Test Map: `newEngineRunCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Help states UNWIRED blocks exposure raising, not runtime startup or reduce-only exits | `TestEngineRunHelpDescribesUnwiredProtectionWithoutClaimingStartupIsImpossible` | stale impossible-start claim present | pass |
