# Branch Test Map: `verifyRateBudgetPath`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | profile journal directory resolution fails explicitly | existing journal path tests | existing | existing |
| tail | different record overrides retain one active-profile budget path | `TestA061RecordOverrideCannotMoveTheProfileRateBudgetLock` | yes | yes |
