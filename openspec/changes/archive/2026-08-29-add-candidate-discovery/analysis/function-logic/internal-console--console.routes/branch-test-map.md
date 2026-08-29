# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 19개 라우트가 전부 session0을 거치고, 상태변경 9개만 mutating을 거치며, readOnly는 `/orders` 하나뿐이다 | `TestEveryRouteGoesThroughTheSessionGate` + `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` + `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` + `TestEveryRouteRefusesARequestWithoutTheSessionToken`(런타임) | — | yes |
