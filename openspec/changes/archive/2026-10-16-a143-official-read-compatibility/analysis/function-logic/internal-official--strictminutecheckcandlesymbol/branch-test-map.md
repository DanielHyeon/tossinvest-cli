# Branch Test Map: `strictMinuteCheckCandleSymbol`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | true: KR `0101N0` accepted unchanged; false: invalid KR delegates and refuses before request, US `AAPL` delegates unchanged | TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol; TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage | original numeric predicate rejected 0101N0; other paths not targeted | final minute suite PASS |

One AST branch, both outcomes. Named fixture success is not instrumented branch coverage.
