# Function Logic Map: `scopeArgs`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account/symbol/market | normalized and consistent with `activeScopeWhere` | caller validation | returns positional args matching predicate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | symbol empty | none | account only | account-wide tests |
| B2 | market empty | none | account + symbol | global symbol tests |
| B3 | market set | none | account + symbol + market | exact-market tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure argument construction | no error path | AST |

## State mutations and fallbacks

- No mutation; branch ordering mirrors `activeScopeWhere` exactly.

## Safety conclusion

- Safe edit boundary: append market only for exact-market predicate.
- High-risk impact: yes — positional mismatch could select/release wrong guard.
