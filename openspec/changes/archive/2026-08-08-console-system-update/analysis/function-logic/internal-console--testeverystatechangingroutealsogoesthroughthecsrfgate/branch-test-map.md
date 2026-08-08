# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | visit every registered route | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | baseline | pass |
| B2 | classify wrapper combination | same | update route absent | pass |
| B3 | state change without CSRF fails | same | baseline | pass |
| B4 | read route with CSRF fails | same | baseline | pass |
| B5 | visit every declared action | same | baseline | pass |
| B6 | missing install route fails | same | route absent | pass |
