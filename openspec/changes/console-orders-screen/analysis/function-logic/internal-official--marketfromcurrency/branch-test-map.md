# Branch Test Map: `marketFromCurrency`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 정규화 후 분기한다(공백·소문자 흡수) | `TestTheRawOrderReadDerivesTheMarketFromTheCurrency` | yes (함수 부재로 컴파일 실패) | yes |
| B2 | `KRW` → `KR` | 동 테스트의 KRW 주문 행 | yes | yes |
| B3 | `USD` → `US` | 동 테스트의 USD 주문 행 | yes | yes |
| B4 | 모르는 통화(`JPY`) → 빈 문자열, 추측하지 않는다 | 동 테스트의 직접 호출 단언 | yes | yes |
