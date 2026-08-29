# Branch Test Map: `Client.ConditionalOrders`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/official/...` 통과.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `status`가 주어지면 그대로 실린다 | `TestConditionalOrdersIntegration`(서버가 `OPEN` 단언) | — (동작 무변경) | yes |
| B2 | `symbol` 생략 시 파라미터 부재 | 동 테스트(전달 안 함) | — | yes |
| B3 | `cursor` 생략 시 부재 | 동상 | — | yes |
| B4 | `limit<=0`이면 부재 | 동상(`limit=0` 전달) | — | yes |
| B5 | `getAcct` 실패가 그대로 올라온다 | `client_test.go:TestGetNon2xxReturnsClassifiedError` | — | yes |
| B6 | 응답의 조건주문이 domain으로 사상된다(OCO의 second nil 포함) | `TestConditionalOrdersIntegration` | — | yes |
