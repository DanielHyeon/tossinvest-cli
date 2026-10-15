# Function Logic Map: `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest`

- Source: `internal/official/strict_minute_candles_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| invalid argument table | invalid market/symbol/count/cursor | strict harness | typed refusal/no request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | subtest iteration | no mutation | test failure | named test |
| B2 | expected refusal | no mutation | test failure | named test |
| B3 | zero requests | no mutation | test failure | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrictMinuteCandles` | preflight refusal | no request | AST |

## State mutations and fallbacks

- Read-only assertions only.

## Safety conclusion

- Safe edit boundary: KR invalid-form table only.
- High-risk impact: no.
