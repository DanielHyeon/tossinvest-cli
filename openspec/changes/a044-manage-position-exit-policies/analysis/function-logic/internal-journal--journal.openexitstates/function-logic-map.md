# Function Logic Map: `Journal.OpenExitStates`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | trimmed account scope | engine context | query error, never widen scope |
| exit/lifecycle rows | only incomplete current managed generation | journal DB | omit released/stale generations |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | account query fails | none | wrapped error | journal query tests |
| B2 | row scan fails | none | error | schema/scan tests |
| B3 | iterator fails | none | wrapped error | journal query tests |
| B4 | current incomplete row | append result | states | exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `db.QueryContext` | select engine working set | no retry; fail closed | CodeGraph + AST |
| `scanExitState` | hydrate immutable state | corruption/error propagated | CodeGraph + AST |

## State mutations and fallbacks

- Read-only query. The edit may narrow the working set by exact lifecycle generation; it must never re-include RELEASED or older generations.

## Safety conclusion

- Safe edit boundary: add lifecycle status/generation predicates only; preserve account/completed predicates and ordering.
- High-risk impact: yes — this query is ExitObserver authority.
