# Function Logic Map: `Store.Dashboard`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query | explicit AsOf, 1..90 days, fixed market/lane/complete filters, positive sample minimum | typed `Query` | invalid period fails before SQL |
| result size | <=10,000 trades (+1 sentinel) | implementation bound | larger result fails closed instead of truncating |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | normalization/period validation fails | none | error | dashboard bounds test |
| B2 | bounded trade query fails or exceeds sentinel | none | error | row-bound test |
| B3 | iterate aggregates | none | continue | dashboard aggregate tests |
| B4 | aggregate below minimum | increments insufficient count | continue | insufficient-sample test |
| B5 | filtered state SQL fails | none | wrapped error | DB error contract |
| B6 | iterate filtered trades | none | continue | missing-state tests |
| B7 | iterate six metric statuses | none | continue | missing-state tests |
| B8 | first metric is not complete | increments not-measured trade count | break to next trade | missing-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `queryTrades` | bounded indexed read of filtered trades/latest snapshots | errors returned; no fallback scan | current HEAD + AST |
| `aggregateTrades` | derives ordered portfolio/lane summaries | operates on bounded slice | tests |
| filtered state SQL | counts complete/link-missing under identical predicates | one indexed query | tests |

## State mutations and fallbacks

- Read-only; never loads price observations or performs a quote poll.
- No silent truncation, no unbounded period, and no policy recommendation/apply action.

## Safety conclusion

- Safe edit boundary: bounded derived analytics query.
- High-risk impact: read correctness/performance only.
