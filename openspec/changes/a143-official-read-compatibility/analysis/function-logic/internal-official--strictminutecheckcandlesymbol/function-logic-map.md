# Function Logic Map: `strictMinuteCheckCandleSymbol`

- Source: `internal/official/strict_minute_candles.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| KR candle symbol | exactly six ASCII `[A-Z0-9]` | amended a143 minute-only contract | fallback shared refusal for invalid form |
| all non-matching inputs | shared strict lexical policy | `strictMinuteCheckSymbol` | preserves numeric KR quote and US policy |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `KR` and uppercase-alphanumeric six chars | none | nil, no normalization | TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol |
| fallback | invalid KR or every US input | delegates unchanged shared validator | existing typed refusal or US decision | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest; existing US minute fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strictMinuteCandleKRSymbol.MatchString` | full-form KR minute acceptance | pure lexical check | AST B1 |
| `strictMinuteCheckSymbol` | retain shared numeric KR and US policy on fallback | no I/O/retry | AST fallback |

## State mutations and fallbacks

- Pure lexical wrapper. No trimming, normalization, request, response identity validation, configuration, order, token, or journal mutation.

## Safety conclusion

- Safe edit boundary: minute reader wrapper only. Strict quote readers continue to call the shared numeric-only helper.
- High-risk impact: yes; explicit minute acceptance and fallback fixtures cover the boundary.
