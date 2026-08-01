# Function Logic Map: `BudgetCoordinator.Observe`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| coordinator | nonnil receiver | caller | nil is a no-op |
| observation | endpoint path and official rate-budget provenance | `official.RateBudget` | delegated conservatively with no cycle authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | coordinator is nil | none | return | nil/manual observation tests |
| B2 | coordinator is nonnil | acquires mutex and delegates with nil cycle record | return after unlock | manual authority tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mutex lock/unlock | serialize manual evidence with cycle observations, grants, and completion | always released by defer | CodeGraph + AST |
| `observeLocked` | apply shared timestamp/reset logic with explicitly nil cycle authority | no error return; invalid evidence fails closed in endpoint state | AST + manual authority tests |

## State mutations and fallbacks

- This wrapper mutates endpoint evidence only through `observeLocked`.
- Passing a nil cycle is the load-bearing authority boundary: manual `Observe` cannot reconcile, advance generation, or clear issuance memory.

## Safety conclusion

- Safe edit boundary: manual observation wrapper only; detailed state logic is mapped under `BudgetCoordinator.observeLocked`.
- High-risk impact: yes, because a relaxed observation could spend safety reserve.
