# Function Logic Map: `ReadOnly.BrokerOrderIDs`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| read-only journal | legacy bare-id origin query | v10 read model | removed because account/day scope is mandatory |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | query/scan/rows handling in removed implementation | read-only SQL | error or id list | base regression evidence |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite query | formerly materialized every broker id | no retry | base AST |

## State mutations and fallbacks

- Removed; callers use bounded composite `BrokerOrderExitLinks`.

## Safety conclusion

- Safe edit boundary: delete unsafe unscoped read API.
- High-risk impact: improves fail-closed attribution; no writes.
