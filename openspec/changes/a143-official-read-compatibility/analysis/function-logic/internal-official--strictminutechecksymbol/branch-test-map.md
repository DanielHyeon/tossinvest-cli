# Branch Test Map: `strictMinuteCheckSymbol`

| Branch | Condition / scenario | Named test or static evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | KR selects KR predicate; US retains existing predicate | TestStrictOrderbookTopAsksForExactlyOneSymbol; TestStrictMinuteCandlesAcceptsTheKoreanMarket; TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | quote 0101N0 RED exposed shared-caller regression | named tests PASS |
| B2 | selected predicate mismatch returns SYMBOL_INVALID; match returns nil | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; TestStrictMinuteCandlesAcceptsTheKoreanMarket | quote 0101N0 is rejected before any request | named tests PASS; no-request assertion on refused inputs |

Named test PASS is not an instrumented branch-hit claim. Initial RED and separate mutation runs are distinguished in analysis/verification-receipt.md. Unchanged error paths are not newly claimed as exhaustively covered.
