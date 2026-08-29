# Function Logic Map: `entryScopeArgs`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account/symbol/market | normalized inputs matching entry predicate | EnterReconcile | positional args returned |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all inputs | none | delegates exact positional construction | market/global tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scopeArgs` | construct query args | pure; no error | AST |

## State mutations and fallbacks

- Pure delegation; no fallback or mutation.

## Safety conclusion

- Safe edit boundary: query arguments only.
- High-risk impact: yes — positional identity must match selection.
