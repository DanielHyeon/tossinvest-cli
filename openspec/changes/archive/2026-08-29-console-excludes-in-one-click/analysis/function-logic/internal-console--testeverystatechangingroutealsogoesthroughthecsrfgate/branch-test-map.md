# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전 라우트 순회 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | no | yes |
| B2 | 판정 분기 진입 | 같은 테스트 | no | yes |
| B3 | `/settings/exclude`가 CSRF 뒤에 있다 | 같은 테스트 | yes | yes |
| B4 | 읽기 라우트가 게이트 뒤에 없다 | 같은 테스트 | no | yes |
| B5 | 목록 순회 | 같은 테스트 | no | yes |
| B6 | 목록의 전 경로가 등록돼 있다 | 같은 테스트 | yes | yes |
