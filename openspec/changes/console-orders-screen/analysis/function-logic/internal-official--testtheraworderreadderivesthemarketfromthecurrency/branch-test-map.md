# Branch Test Map: `TestTheRawOrderReadDerivesTheMarketFromTheCurrency`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes (`OrdersRaw`/`marketFromCurrency` 부재로 컴파일 실패) | yes |
| B2 | KRW 행이 KR로 | 자체 실행 | yes | yes |
| B3 | USD 행이 US로 | 자체 실행 | yes | yes |
| B4 | 모르는 통화는 빈 문자열 | 자체 실행 | yes | yes |
