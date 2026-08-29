# Function Logic Map: `ExitObserver.ladderFor`

- Source: `internal/app/engine/exitloop.go`
- Qualified function: `ExitObserver.ladderFor`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/app/engine/exitloop.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/app/engine/exitloop.go:813` — if id == "" \|\| id == "default_v1" { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B2 | `if` at `internal/app/engine/exitloop.go:814` — if o.ladder.PolicyID != "default_v1" { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B3 | `if` at `internal/app/engine/exitloop.go:822` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`, `fmt.Errorf`, `exitpolicy.CommonLadderForPosition`, `m.position.Adopted` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 2 assignment(s) and 4 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
