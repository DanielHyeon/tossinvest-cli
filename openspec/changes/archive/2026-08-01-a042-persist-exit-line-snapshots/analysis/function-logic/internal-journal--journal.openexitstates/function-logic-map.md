# Function Logic Map: `Journal.OpenExitStates`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account | open exit-state rows | journal schema | SQL errors global; legacy wrapper preserves prior API |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | query, scan, rows error, success | none | error/states | existing working-set tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitState` | restore one state | wrapper retains semantic-error behavior | CodeGraph + AST |

## State mutations and fallbacks

- New engine path uses per-row typed results; this compatibility method remains for existing callers.

## Safety conclusion

- Safe edit boundary: unchanged function body.
- High-risk impact: yes.
