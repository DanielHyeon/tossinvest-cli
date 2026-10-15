# Branch Test Map: `strictMinuteMarketCurrency`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | dispatch by market | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | final minute suite PASS |
| B2 | KR becomes KRW | TestStrictMinuteCandlesAcceptsTheKoreanMarket | not targeted | final minute suite PASS |
| B3 | US becomes USD | existing US strict fixtures | not targeted | final minute suite PASS |
| B4 | invalid market refuses pre-request | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | final minute suite PASS |
