# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- Qualified function: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`
- AST evidence: `ast.json` (`dc5442dd71900b066e809ce77368b8fb06ac0dd834eeb5ca7f676e03bc80b9df`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/console/static_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` at `internal/console/static_test.go:385` — for _, r := range registeredRoutes(t) { | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |
| B2 | `switch` at `internal/console/static_test.go:387` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |
| B3 | `case` at `internal/console/static_test.go:388` — case stateChanging[r.Path] && !r.CSRFGated: | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |
| B4 | `case` at `internal/console/static_test.go:390` — case !stateChanging[r.Path] && r.CSRFGated: | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |
| B5 | `range` at `internal/console/static_test.go:394` — for path := range stateChanging { | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |
| B6 | `if` at `internal/console/static_test.go:395` — if !seen[path] { | only mutations visible in this branch and its callees | existing return/error contract | `TestOptimizationRejectsAClientInventedPolicy`; `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`; console route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes`, `t.Errorf` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 3 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: no; current AST hash and affected-package tests are required.
