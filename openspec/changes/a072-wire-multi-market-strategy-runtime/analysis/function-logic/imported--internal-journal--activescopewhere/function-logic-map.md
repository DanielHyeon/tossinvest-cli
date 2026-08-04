# Function Logic Map: `activeScopeWhere`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| symbol/market | account-wide global, symbol global, or symbol exact KR/US | normalized caller inputs | returns SQL predicate for exactly one release scope |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | symbol empty | none | account-wide/global predicate | validation/query tests |
| B2 | market empty | none | symbol/global predicate | global scope tests |
| B3 | market set | none | symbol/exact-market predicate | cross-market release test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure predicate selection | caller supplies matching args | AST |

## State mutations and fallbacks

- Pure SQL fragment; no mutation. It must never widen exact release to peer/global scope.

## Safety conclusion

- Safe edit boundary: add market predicate branch only.
- High-risk impact: yes — selection controls which guard may be released.
