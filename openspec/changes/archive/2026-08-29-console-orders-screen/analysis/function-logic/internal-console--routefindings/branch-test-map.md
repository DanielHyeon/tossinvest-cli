# Branch Test Map: `routeFindings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 허용 목록 구성 | 표 판정 전체 | — | yes |
| B2 | 계좌 동사 순회 | 같은 위 | — | yes |
| B3 | 계좌 동사를 이름에 쓴 비예외 경로 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`, `TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt` | yes | yes |
| B4 | 허용 경로와 예외 경로는 두 번째 루프를 건너뛴다 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`의 마지막 단언(`/orders`가 발견 0) | yes — `reads`를 빼면 `/orders`가 두 번째 루프에 걸린다 | yes |
| B5 | 행위 동사 순회 | 표 판정 전체 | — | yes |
| B6 | 목록 밖의 행위 경로 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`(`/verify/reset`) | yes | yes |
