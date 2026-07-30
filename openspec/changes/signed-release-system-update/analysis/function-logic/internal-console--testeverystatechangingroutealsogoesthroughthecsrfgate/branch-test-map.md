# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Iterate every registered route | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | baseline | pass |
| B2 | Classify each route against the closed state-changing set | same | baseline | pass |
| B3 | A state-changing route without CSRF fails | same | signed download missing from expected set | pass |
| B4 | A read route incorrectly wrapped in CSRF fails | same | baseline | pass |
| B5 | Iterate every expected state-changing path | same | baseline | pass |
| B6 | An expected mutation path missing from registration fails | same | signed download route absent | pass |
