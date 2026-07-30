# Function Logic Map: `NewReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go`
- Qualified function: `NewReconcileDriver`
- AST evidence: `ast.json` (`157aa37d842a4ab0379b0364a9590d18d5b3ef27b9a655dd3e6ed803120dcc29`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/app/engine/reconcileloop.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` at `internal/app/engine/reconcileloop.go:279` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B2 | `case` at `internal/app/engine/reconcileloop.go:280` — case opts.Journal == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B3 | `case` at `internal/app/engine/reconcileloop.go:282` — case opts.Collector == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B4 | `case` at `internal/app/engine/reconcileloop.go:284` — case opts.Tracker == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B5 | `case` at `internal/app/engine/reconcileloop.go:286` — case opts.Ingest == nil \|\| opts.Converge == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B6 | `case` at `internal/app/engine/reconcileloop.go:288` — case strings.TrimSpace(opts.AccountRef) == "": | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B7 | `case` at `internal/app/engine/reconcileloop.go:291` — case opts.Adoption.Enabled && opts.Prices == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B8 | `if` at `internal/app/engine/reconcileloop.go:295` — if id := strings.TrimSpace(opts.CommonPolicy); id != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B9 | `if` at `internal/app/engine/reconcileloop.go:296` — if _, ok := exitpolicy.CommonPolicyByID(id); !ok { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B10 | `if` at `internal/app/engine/reconcileloop.go:308` — if d.clk == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Errorf`, `strings.TrimSpace`, `exitpolicy.CommonPolicyByID`, `clock.System` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 8 assignment(s) and 8 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
