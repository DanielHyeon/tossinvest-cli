# Function Logic Map: `Client.StrictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client | non-nil | receiver | `CLIENT_MISSING`, no request |
| market/symbol | exact market and lexical predicate | market helper and `strictMinuteCheckCandleSymbol` | typed refusal before request |
| count | 1..200 | constant | `COUNT_INVALID`, no request |
| before | empty or valid RFC3339 | strict timestamp parser | `BEFORE_INVALID`, no request |
| response | latest-first raw page, close-time labels | strict decoder | contract refusal / propagated transport error |

Only B3 dependency may change. Query carries supplied symbol unchanged, returned `StrictMinutePage.Symbol` remains exact request text. No response-symbol field exists; no identity-attestation behavior is added.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | `CLIENT_MISSING` | existing nil-client test |
| B2 | invalid market | none | market refusal | existing argument table |
| B3 | lexical invalid symbol | none; must precede request | `SYMBOL_INVALID` | invalid KR boundary table |
| B4 | count outside 1..200 | none | `COUNT_INVALID` | existing argument table |
| B5 | before nonempty | parse before into comparison instant | proceeds to B6 | argument table and canonical-query fixture |
| B6 | before parse fails | none | `BEFORE_INVALID` | argument table |
| B7 | nonempty before | request contains unchanged cursor | none | canonical-query test |
| B8 | outer observer present | invokes outer observer | none | observer chaining test |
| B9 | transport error | no page returned | propagated error | transport-classification fixture |
| B10 | final-attempt extraction error | no page returned | typed error | preserved path; no isolated error-branch receipt |
| B11 | strict decoder error | no page returned | typed error | malformed-body fixtures |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strictMinuteCheckCandleSymbol` | minute-only KR lexical widening, otherwise shared preflight refusal | no request on invalid input | AST B3 |
| `c.get` | official read | existing transport/retry classification | AST B9 |
| `strictMinuteDecode` | validates raw order/cursor/data | typed strict refusal | AST B11 |
| `RateBudget` | attaches observed budget | unchanged | AST return |

## State mutations and fallbacks

- Captures response traces and returns page data; no order, token, setting, journal, or broker write.
- Implemented dependency change: this method calls the minute-only wrapper. The shared numeric-only validator remains unchanged for strict quote readers; US regex, cursor, decoder and returned-page propagation remain unchanged.

## Safety conclusion

- Safe edit boundary: this method call site plus the minute-only wrapper and its KR predicate.
- High-risk impact: yes by evidence-reader role. Preserve every non-lexical refusal.
