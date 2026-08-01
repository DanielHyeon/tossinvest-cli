# Branch Test Map: `Dispatch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero/forged decision | `TestDispatchRejectsOpaqueZeroDecisionBeforeAuthority` + purity guards | exported Validate draft | pass |
| Success | lane-minted decision only | strategyengine lane tests + compile wiring guards | bypass draft | pass |
