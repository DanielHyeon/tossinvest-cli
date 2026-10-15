# Branch Test Map: `strictMinuteCheckSymbol`

| Branch | Condition / scenario | Named test or static evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | KR selects KR predicate; US retains existing predicate | TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol; TestStrictMinuteCandlesAcceptsTheKoreanMarket; TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | 0101N0 initial RED and old-predicate mutant FAIL | named tests PASS |
| B2 | selected predicate mismatch returns SYMBOL_INVALID; match returns nil | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol; TestStrictMinuteCandlesAcceptsTheKoreanMarket | 0101N0 failed with old predicate; invalid forms not RED targets | named tests PASS; no-request assertion on refused inputs |

Named test PASS is not an instrumented branch-hit claim. Initial RED and separate mutation runs are distinguished in analysis/verification-receipt.md. Unchanged error paths are not newly claimed as exhaustively covered.
