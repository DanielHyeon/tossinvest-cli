# Branch Test Map: `routeReadsTheAccountRecord`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 예외는 부여되고(`/orders` + readOnly + CSRF 밖), 세 사실 중 하나라도 깨지면 부여되지 않으며, 아래·옆 경로 다섯에도 상속되지 않는다 | `TestTheOrdersExceptionAppliesOnlyToTheExactPathThatReadsAndDoesNotAct`(3 negative), `TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt`(5 negative) | yes — 리뷰가 접두 일치·ToLower·TrimSuffix 세 변이를 시연했고 이 형태의 단언만 그것을 잡는다 | yes |
