# Branch Test Map: `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로별 응답 분기 | 자체 실행 | — (base에서 이월된 단언) | yes |
| B2 | 토큰 발급 | 자체 실행 | — | yes |
| B3 | 빈 주문 페이지 응답 | 자체 실행 | — | yes |
| B4 | 여섯 키 순회 | 자체 실행 | — | yes |
| B5 | 클라이언트가 값을 지어내지 않는다 | 자체 실행 | — | yes |
| B6 | 예상 밖 경로는 404 | 자체 실행 | — | yes |
| B7 | `Orders`가 오류 없이 돌아온다 | 자체 실행 | — | yes |
| B8 | 결과 0건 | 자체 실행 | — | yes |
| B9 | 같은 입력을 `OrdersRaw`는 거부한다 | 자체 실행 | yes (가드 도입 전에는 요청이 나갔다) | yes |
