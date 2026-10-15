# Branch Test Map: `Client.StrictMinuteCandles`

| Branch | Condition / scenario | Named test or static evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil receiver refusal | TestStrictMinuteCandlesRefusesANilClient | not targeted | see final strict-suite receipt |
| B2 | market/currency helper error | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | named test PASS |
| B3 | minute-only symbol wrapper error; valid exact symbol proceeds | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol; TestStrictMinuteCandlesAcceptsTheKoreanMarket | 0101N0 initial RED and old-KR-predicate mutant FAIL | named tests PASS |
| B4 | count outside 1..200 refused | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | named test PASS |
| B5 | nonempty before enters timestamp validation | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage | not targeted | see final strict-suite receipt |
| B6 | invalid before timestamp refused | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | named test PASS |
| B7 | nonempty before copied unchanged into query | TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage | not targeted | see final strict-suite receipt |
| B8 | outer attempt observer invoked when present | TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver | not targeted | see final strict-suite receipt |
| B9 | transport error propagates | TestStrictMinuteCandlesPropagatesTransportClassification | not targeted | see final strict-suite receipt |
| B10 | final attempt extraction error | Static preserved path; successful-attempt test does not independently prove this error branch | not targeted | not individually isolated |
| B11 | strict decode error propagates | TestStrictMinuteCandlesRefusesMalformedBodies | not targeted | see final strict-suite receipt |

Named test PASS is not an instrumented branch-hit claim. Initial RED and separate mutation runs are distinguished in analysis/verification-receipt.md. Unchanged error paths are not newly claimed as exhaustively covered.

The new alphanumeric fixture has one candle and proves exact request/query/page-symbol propagation only. Raw newest-first page/cursor evidence comes from the existing two-candle TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage in the final strict suite; officialbars close-to-open conversion is tested separately. The wire response does not independently attest symbol identity.
