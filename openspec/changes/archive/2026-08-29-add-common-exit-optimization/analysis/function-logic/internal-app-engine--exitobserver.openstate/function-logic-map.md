# Function Logic Map: `ExitObserver.openState`

- Source: `internal/app/engine/exitloop.go`
- Qualified function: `ExitObserver.openState`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/app/engine/exitloop.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/app/engine/exitloop.go:513` — if p.Adopted() { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B2 | `if` at `internal/app/engine/exitloop.go:517` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B3 | `if` at `internal/app/engine/exitloop.go:522` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B4 | `if` at `internal/app/engine/exitloop.go:527` — if !ok { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B5 | `if` at `internal/app/engine/exitloop.go:534` — if o.opts.CommonPolicy != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B6 | `switch` at `internal/app/engine/exitloop.go:544` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B7 | `case` at `internal/app/engine/exitloop.go:545` — case errors.Is(err, journal.ErrExitStateExists): | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B8 | `case` at `internal/app/engine/exitloop.go:548` — case err != nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `p.Adopted`, `o.openAdoptedState`, `o.opts.Journal.LookupDecision`, `fmt.Errorf`, `journal.ParsePreimage`, `o.opts.Journal.OpenExitState`, `errors.Is`, `o.log` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 6 assignment(s) and 7 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
