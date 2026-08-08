# Branch Test Map: `TestEveryRouteGoesThroughTheSessionGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | inspect all registered routes | `TestEveryRouteGoesThroughTheSessionGate` | 22-route baseline | pass |
| B2 | any missing session wrapper fails | same | update route absent | pass |
| B3 | extractor sees at least 23 routes | same | floor omitted update | pass |
