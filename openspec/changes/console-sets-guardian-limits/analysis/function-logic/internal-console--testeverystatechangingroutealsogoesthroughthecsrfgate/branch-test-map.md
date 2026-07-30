# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전 라우트 순회 | 자기 자신 | yes | yes |
| B2 | 라우트별 판정 | 자기 자신 | yes | yes |
| B3 | 상태변경인데 CSRF 없으면 실패 | 자기 자신 | no (신규 라우트는 처음부터 mutating) | yes |
| B4 | 목록 밖 CSRF 라우트는 실패 | 자기 자신 | **yes — RED에서 두 라우트가 정확히 이 문장으로 실패했다** | yes |
| B5 | 목록 순회 | 자기 자신 | yes | yes |
| B6 | 목록에만 있고 미등록이면 실패 | 자기 자신 | no | yes |
