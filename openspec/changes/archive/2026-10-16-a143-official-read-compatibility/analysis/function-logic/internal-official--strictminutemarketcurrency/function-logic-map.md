# Function Logic Map: `strictMinuteMarketCurrency`

- Source: `internal/official/strict_minute_candles.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market | exact `KR` or `US` | shared strict reader contract | typed market refusal before request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | market switch | none | chooses currency/refusal | minute argument table |
| B2 | `KR` | none | `KRW`, nil | Korean minute fixture |
| B3 | `US` | none | `USD`, nil | US strict fixtures |
| B4 | other market | none | `MARKET_INVALID` | minute argument table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strictMinuteRefuse` | typed invalid-market error | no I/O or retry | AST B4 |
| `strconv.Quote` | preserve supplied market in detail | no I/O or retry | AST B4 |

## State mutations and fallbacks

- Pure market-to-currency selection. Function body behavior did not change; AST refresh is required because its source file changed.

## Safety conclusion

- Safe edit boundary: no semantic edit in this helper.
- High-risk impact: yes through strict preflight; exact contract retained.
