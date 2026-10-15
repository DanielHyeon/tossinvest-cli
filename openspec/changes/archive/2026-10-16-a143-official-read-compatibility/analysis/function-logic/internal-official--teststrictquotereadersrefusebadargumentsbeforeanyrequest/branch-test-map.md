# Branch Test Map: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate each invalid input | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | `0101N0` exposed regression | final quote fixture PASS |
| B2 | orderbook must not accept | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | regression reached broker | assertion evaluated |
| B3 | orderbook refusal is typed | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | regression reached broker | assertion evaluated |
| B4 | orderbook reason is `SYMBOL_INVALID` | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | not present before test addition | assertion evaluated |
| B5 | last-price must not accept | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | regression reached broker | assertion evaluated |
| B6 | last-price refusal is typed | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | regression reached broker | assertion evaluated |
| B7 | last-price reason is `SYMBOL_INVALID` | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | not present before test addition | assertion evaluated |
| B8 | both readers make zero requests | TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest | 2 calls before correction | final quote fixture PASS |
