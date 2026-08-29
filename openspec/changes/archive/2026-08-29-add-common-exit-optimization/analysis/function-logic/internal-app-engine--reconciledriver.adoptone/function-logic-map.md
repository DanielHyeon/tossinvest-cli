# Function Logic Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go`
- Qualified function: `ReconcileDriver.adoptOne`
- AST evidence: `ast.json` (`2a1ce278711c340c4de8f53d21576c72e53f1d8b457f8f38adf6d3207c3e7a43`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/app/engine/adoption.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/app/engine/adoption.go:283` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B2 | `if` at `internal/app/engine/adoption.go:302` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B3 | `if` at `internal/app/engine/adoption.go:307` — if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.SyntheticStop`, `d.logDeferred`, `err.Error`, `d.opts.Journal.AdoptPosition`, `journal.RFC3339`, `d.clk.Now`, `d.opts.Journal.OpenAdoptedExitState`, `d.alert`, `string`, `delete` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 3 assignment(s) and 3 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
