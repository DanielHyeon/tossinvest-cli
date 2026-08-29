# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- Qualified function: `Console.routes`
- AST evidence: `ast.json` (`aec87df0eb373e7771cf18420a98de1dc480c98d32d42db6ee6ccda4b230d378`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/console/console.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path; function is branchless in the current AST | direct registration/delegation only | existing return contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux`, `mux.HandleFunc`, `c.session0`, `c.mutating`, `c.startExclusive`, `c.registerOverview`, `c.registerOrders`, `c.registerSignals` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 1 assignment(s) and 1 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: no; current AST hash and affected-package tests are required.
