# Branch Test Map: `routeFindings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 논증된 목록이 예외로 구성된다 | `TestNoRouteNamesAnAccountMutation` | yes | yes |
| B2 | 계좌 어휘 전부 순회 | `TestNoRouteNamesAnAccountMutation` | yes | yes |
| B3 | 계좌 어휘를 담은 경로는 실패 | `TestNoRouteNamesAnAccountMutation` (`gate` 유지로 고정) | no | yes |
| B4 | 논증된 라우트는 두 번째 루프 면제 | `TestNoRouteNamesAnAccountMutation` | yes | yes |
| B5 | 행위 어휘 전부 순회 | `TestNoRouteNamesAnAccountMutation` | yes | yes |
| B6 | 목록 밖 행위 라우트는 실패 | `TestNoRouteNamesAnAccountMutation` | **yes — `/settings/limits/preset`이 `"reset"`으로 실패했다** | yes |
