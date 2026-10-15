# Branch Test Map: `strictMinuteCheckCandleSymbol`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | KR `0101N0` is accepted without rewriting | TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol | original numeric predicate rejected it | final minute suite PASS |
| fallback | invalid KR form delegates and refuses before request | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | final minute suite PASS |
| fallback | US `AAPL` delegates to unchanged US policy | existing US minute fixtures | not targeted | final minute suite PASS |
