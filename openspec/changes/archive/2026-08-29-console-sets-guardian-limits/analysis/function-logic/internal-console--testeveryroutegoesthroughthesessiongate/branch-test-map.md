# Branch Test Map: `TestEveryRouteGoesThroughTheSessionGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록된 전 라우트가 순회된다 | 자기 자신 | yes (하한 실패로 관측) | yes |
| B2 | 세션 게이트 없는 라우트는 실패 | 자기 자신 | no (신규 라우트는 처음부터 게이트 뒤) | yes |
| B3 | 하한 미만이면 실패 | 자기 자신 | yes (20 유지 시 22개를 세지 못함을 확인) | yes |
