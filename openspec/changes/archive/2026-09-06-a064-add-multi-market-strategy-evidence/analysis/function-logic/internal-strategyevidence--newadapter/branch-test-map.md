# Branch Test Map: `NewAdapter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | two official adapters on the same authority/contract share one call ceiling | `TestSharedRateBudgetCapsEveryAdapterOnOneContract` | no RED: this is a structural closure, not a behaviour fix. The property it pins (`NewOfficialAdapter` shares a budget) already held; what did not hold was that it was the only path. | PASS |
