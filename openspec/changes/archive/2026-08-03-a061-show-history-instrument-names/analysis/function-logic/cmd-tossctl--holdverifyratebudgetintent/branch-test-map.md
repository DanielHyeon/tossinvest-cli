# Branch Test Map: `holdVerifyRateBudgetIntent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid profile refuses before marker/API use | profile path tests | existing | existing |
| B2 | read-only stale marker refuses before rate-budget lease or broker | `TestA061VerificationRefusesBeforeBrokerWhenItsIntentCannotBePublished` | yes | yes |
| tail | intent appears before occupied budget wait and is removed on cancellation | `TestA061RunVerifyAbortPublishesItsMarkerBeforeWaitingForTheProfileBudget`, run/console contention tests | yes | yes |
