# Branch Test Map: `TestEveryRouteGoesThroughTheSessionGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록된 전 라우트 순회 | `TestEveryRouteGoesThroughTheSessionGate` | no | yes |
| B2 | 세션 게이트 누락 검출 | `TestEveryRouteGoesThroughTheSessionGate` | no | yes |
| B3 | 라우트 20개 하한 | `TestEveryRouteGoesThroughTheSessionGate` | yes | yes |
