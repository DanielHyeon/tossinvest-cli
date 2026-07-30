# Branch Test Map: `routeFindings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 허용 집합 구성 | `TestNoRouteNamesAnAccountMutation` | no | yes |
| B2 | 계좌 어휘 순회 | 같은 테스트 | no | yes |
| B3 | 계좌 동사 경로 검출 | `TestRouteFindingsRejectsAnAccountVerb` 계열 | no | yes |
| B4 | 논증된 경로·읽기 예외 면제 | 같은 테스트 | no | yes |
| B5 | 행위 어휘 순회 — `exclude` 포함 | 같은 테스트 | no | yes |
| B6 | 미논증 행위 경로 검출 | 같은 테스트 | no | yes |
