# Function Logic Map: `Client.StrictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client | non-nil | receiver | `CLIENT_MISSING`, no request |
| market/symbol | exact market and lexical predicate | market helper and `strictMinuteCheckSymbol` | typed refusal before request |
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
| B5/B6 | nonempty invalid before | none | `BEFORE_INVALID` | existing argument table |
| B7 | nonempty before | request contains unchanged cursor | none | canonical-query test |
| B8 | outer observer present | invokes outer observer | none | observer chaining test |
| B9/B10/B11 | transport/final-attempt/decode error | no page returned | propagated/typed error | existing transport/body tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strictMinuteCheckSymbol` | preflight lexical refusal | no request on invalid input | AST B3 |
| `c.get` | official read | existing transport/retry classification | AST B9 |
| `strictMinuteDecode` | validates raw order/cursor/data | typed strict refusal | AST B11 |
| `RateBudget` | attaches observed budget | unchanged | AST return |

## State mutations and fallbacks

- Captures response traces and returns page data; no order, token, setting, journal, or broker write.
- Planned edit: KR regex only. Do not normalize/case-fold, alter US regex, cursor, decoder, or returned-page propagation.

## Safety conclusion

- Safe edit boundary: `strictMinuteKRSymbol` declaration only.
- High-risk impact: yes by evidence-reader role. Preserve every non-lexical refusal.
