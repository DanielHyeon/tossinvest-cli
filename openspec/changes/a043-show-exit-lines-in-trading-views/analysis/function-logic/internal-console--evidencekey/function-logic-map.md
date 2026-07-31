# Function Logic Map: `evidenceKey`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broker id/account/market/time | all nonblank and canonical | broker reading + market clock | return invalid key, never guess |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | id/account absent | none | invalid | origin and unlinked tests |
| B2 | timestamp invalid | none | invalid | unparsed timestamp tests |
| B3 | market invalid | none | invalid | market fallback tests |
| B4 | trading-day conversion fails | none | invalid | clock tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.ParseMarket`, `Market.TradingDay` | canonical market-local day | no retry; errors fail closed | AST + clock tests |

## State mutations and fallbacks

- Pure identity construction; no state mutation or fallback to host-local dates.

## Safety conclusion

- Safe edit boundary: broker evidence lookup identity only.
- High-risk impact: fail-closed read label only.
