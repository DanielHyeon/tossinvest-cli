# Branch Test Map: `Client.OrderByID`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/official/...` 통과.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `getAcct` 실패가 그대로 올라온다(빈 `domain.Order` + 오류) | `client_test.go:TestGetNon2xxReturnsClassifiedError`, `orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead` | — (동작 무변경) | yes |
| (무분기 꼬리) | 정상 응답이 `domain.Order`로 사상되고 계좌 헤더가 실려 나간다 | `TestOrderByIDIntegration` | — | yes |
