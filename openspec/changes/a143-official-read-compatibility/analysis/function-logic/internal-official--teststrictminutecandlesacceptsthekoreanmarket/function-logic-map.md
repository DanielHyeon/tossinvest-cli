# Function Logic Map: `TestStrictMinuteCandlesAcceptsTheKoreanMarket`

- Source: `internal/official/strict_minute_candles_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| numeric KR read | valid page | strict harness | test fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | strict call error | no mutation | test failure | named test |
| B2 | page fields | no mutation | test failure | named test |
| B3 | canonical query | no mutation | test failure | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrictMinuteCandles` | strict read | existing contract | AST |

## State mutations and fallbacks

- Read-only assertions only.

## Safety conclusion

- Safe edit boundary: neighboring test expansion only.
- High-risk impact: no.
