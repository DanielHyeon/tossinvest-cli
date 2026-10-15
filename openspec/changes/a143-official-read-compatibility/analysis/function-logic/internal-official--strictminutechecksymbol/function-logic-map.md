# Function Logic Map: `strictMinuteCheckSymbol`

- Source: `internal/official/strict_minute_candles.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market | `KR` or `US`, checked by caller | `StrictMinuteCandles` | selects predicate |
| KR symbol | exactly six `[A-Z0-9]` ASCII characters | public REST snapshot example `0101N0` | `SYMBOL_INVALID` |
| US symbol | existing `[A-Z][A-Z0-9.]{0,9}` | current strict contract | unchanged |

No normalization, Unicode folding, whitespace trimming, response identity validation, or return-value mutation. Regex anchors validate full input.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == "KR"` | selects KR predicate | falls through to match | KR matrix |
| B2 | selected predicate does not match | none | typed `SYMBOL_INVALID` quoting original input | zero-request rejection |
| return | selected predicate matches | none | nil | KR and existing US valid cases |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `pattern.MatchString` | full lexical validation | no retry/I/O | AST B2 |
| `strictMinuteRefuse` | typed error | reason contract unchanged | AST B2 |
| `strconv.Quote` | error detail | exact unnormalized input retained | AST B2 |

## State mutations and fallbacks

- Pure lexical check; no mutation, I/O, retry, config, token, journal, or broker call.

## Safety conclusion

- Safe edit boundary: KR regular-expression literal only.
- High-risk impact: yes by caller role. Preserve refusal code and no-request property.
