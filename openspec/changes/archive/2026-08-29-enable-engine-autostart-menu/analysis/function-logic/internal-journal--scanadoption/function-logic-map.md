# Function Logic Map: `scanAdoption`

- Source: `internal/journal/adoption.go`
- Qualified function: `scanAdoption`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/adoption.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/adoption.go:348` — if err := rows.Scan(&a.ID, &a.Symbol, &a.Market, &a.Quantity, &a.CostBasis, | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rows.Scan`, `fmt.Errorf` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 1 assignment(s) and 2 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
