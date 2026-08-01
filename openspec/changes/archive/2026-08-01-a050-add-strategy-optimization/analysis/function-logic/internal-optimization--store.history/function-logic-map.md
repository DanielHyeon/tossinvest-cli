# Function Logic Map: `Store.history`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| immutable snapshot history | newest-first, bounded at 1000, every row validated | control DB | query/validation error prevents View read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query fails | none | error | DB fault coverage |
| B2 | row scan/integrity fails | none | error | corruption test |
| B3 | valid rows | one bounded query, clone snapshots | return newest-first list | lifecycle history tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryContext`, `scanSnapshot` | bounded bulk read, avoiding N+1 queries | rows error propagates | AST |

## State mutations and fallbacks

- Read-only projection only; no state mutation or authority binding.

## Safety conclusion

- Safe edit boundary: private read query.
- High-risk impact: no.
