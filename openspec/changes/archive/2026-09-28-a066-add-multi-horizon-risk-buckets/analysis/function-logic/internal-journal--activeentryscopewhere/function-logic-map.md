# Function Logic Map: `activeEntryScopeWhere`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| symbol/market | normalized global or exact scope | EnterReconcile | pure SQL predicate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | account-wide/global request | none | exact global predicate | global tests |
| B2 | exact market | none | global-or-same-market overlap predicate | market tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `activeScopeWhere` | reuse exact global predicate | pure; no error | AST |

## State mutations and fallbacks

- No mutation. Exact entry widens only to global NULL, never peer market.

## Safety conclusion

- Safe edit boundary: entry lookup only.
- High-risk impact: yes — determines overlap authority.
