# Function Logic Map: `scanExitState`

- Source: `internal/journal/apply_hook.go`
- Qualified function: `scanExitState`
- AST evidence: `ast.json` (`d3edb0b0a6bc08316d569e329e132bd5fc0458c8ffe276a877d816ca773c99ad`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/apply_hook.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/apply_hook.go:553` — if errors.Is(err, sql.ErrNoRows) { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B2 | `if` at `internal/journal/apply_hook.go:556` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B3 | `if` at `internal/journal/apply_hook.go:560` — if s.PolicyKind == ExitPolicyLadder && s.PolicyID == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |
| B4 | `if` at `internal/journal/apply_hook.go:563` — if rung.Valid { | only mutations visible in this branch and its callees | existing return/error contract | `TestLadderExitStateSnapshotsItsPolicyID`; `TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery`; journal adoption/exit-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `row.Scan`, `errors.Is`, `fmt.Errorf`, `int` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 5 assignment(s) and 3 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
