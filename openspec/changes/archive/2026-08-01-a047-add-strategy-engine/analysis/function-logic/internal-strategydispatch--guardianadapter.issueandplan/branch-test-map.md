# Branch Test Map: `GuardianAdapter.IssueAndPlan`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unconfigured/mismatched binding | adapter focused tests | fake split draft | pass |
| B2 | atomic issuance rollback | journal rollback/Guardian tests | split commits | pass |
| B3 | exact receipt becomes exact plan | adapter integration/static tests | post-commit lookup | pass |
| B4 | account read failure calls no Guardian | adapter focused tests | missing | pass |
| B5 | atomic receipt mismatch refuses handoff | adapter binding tests | missing | pass |
