# Branch Test Map: `Console.decoratePositionRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no lifecycle seam wired keeps rows unknown | existing legacy-console tests | n/a | pass |
| B2 | a failed `List` leaves lifecycle unverified | existing a052 tests | n/a | pass |
| B3 | the policy index is built per request | existing a052 tests | n/a | pass |
| B4 | no settings seam leaves designation flags false | existing tests | n/a | pass |
| B5 | one `Load` stamps both lists | existing include/exclude tests | n/a | pass |
| B6 | designation flags reach every row | existing tests | n/a | pass |
| B7 | the management pass runs only when a runtime was attempted | existing a052 tests | n/a | pass |
| B8 | every row gets a management projection | existing a052 tests | n/a | pass |
| B9 | journal rows require lifecycle proof | existing tests | n/a | pass |
| B10 | a missing policy row reads as unknown, not unmanaged | existing lifecycle-unknown test | n/a | pass |
| B11 | a present policy row supplies status and generation | existing tests | n/a | pass |
| B12 | a reconcile block renders on the row | existing reconcile tests | n/a | pass |
| tail call | the liveness verdict reaches both screens identically | `TestTheProtectionLineSurvivesAJudgementThatChangedNothing` (/positions and /dashboard) | yes | yes |
