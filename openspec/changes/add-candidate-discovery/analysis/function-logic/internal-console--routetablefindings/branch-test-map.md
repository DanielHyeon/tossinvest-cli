# Branch Test Map: `routeTableFindings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표 전체 판정 | `TestNoRouteNamesAnAccountMutation` | — | yes |
| B2 | 상태변경 목록 순회 | 같은 위 | — | yes |
| B3 | 목록에만 있고 등록되지 않은 경로 | 등록 제거 변이 | yes | yes |
