# Branch Test Map: `Client.OrdersRaw`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈/공백 status는 요청 전에 거부되고 `ShouldFallback`이 false | `TestTheRawReadsRefuseARequestWithNoStatusGroup` (서버 요청 수 0 단언 포함) | yes (가드를 항상 거짓인 조건으로 바꾸면 테스트가 실제로 물었다 — issues.md I-7) | yes |
| B2 | 호출자가 준 그룹이 그대로 와이어에 실린다 | `TestBothOrderReadsSendTheGroupTheyWereGiven` (OPEN/CLOSED 양쪽) | yes | yes |
| B3 | `symbol` 생략 시 파라미터 부재 | `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne`(서버가 6개 키 부재 단언) | yes | yes |
| B4 | `from` 생략 시 부재 | 동상 | yes | yes |
| B5 | `to` 생략 시 부재 | 동상 | yes | yes |
| B6 | `cursor` 생략 시 부재 | 동상 | yes | yes |
| B7 | `limit<=0`이면 부재 | 동상 | yes | yes |
| B8 | 전송·인증·429·5xx가 분류된 오류로 올라온다 | `orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`(같은 `send` 경로) | — | yes |
| B9 | 페이지의 각 주문이 원문 문자열로 옮겨진다 | `TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne`, `TestTheRawOrderReadDerivesTheMarketFromTheCurrency`, `TestTheRawOrderReadReportsThatThePageWasTruncated` | yes | yes |
