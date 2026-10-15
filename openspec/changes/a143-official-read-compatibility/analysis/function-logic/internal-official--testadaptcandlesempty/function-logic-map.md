# Function Logic Map: `TestAdaptCandlesEmpty`

- Source: `internal/official/candle_reads_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| empty chart | zero candles | `apiCandlePage{}` | test fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | symbol assertion | no mutation | test failure | `TestAdaptCandlesEmpty` |
| B2 | length assertion | no mutation | test failure | `TestAdaptCandlesEmpty` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `adaptCandles` | builds chart | no I/O | AST |

## State mutations and fallbacks

- Read-only assertions only.

## Safety conclusion

- Safe edit boundary: test-only neighboring source change.
- High-risk impact: no.
