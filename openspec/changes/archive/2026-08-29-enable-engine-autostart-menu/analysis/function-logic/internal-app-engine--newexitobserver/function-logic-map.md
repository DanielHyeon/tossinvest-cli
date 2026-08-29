# Function Logic Map: `NewExitObserver`

- Source: `internal/app/engine/exitloop.go`
- Qualified function: `NewExitObserver`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/app/engine/exitloop.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` at `internal/app/engine/exitloop.go:250` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B2 | `case` at `internal/app/engine/exitloop.go:251` — case opts.Journal == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B3 | `case` at `internal/app/engine/exitloop.go:253` — case opts.Prices == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B4 | `case` at `internal/app/engine/exitloop.go:256` — case opts.Retrier == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B5 | `case` at `internal/app/engine/exitloop.go:259` — case opts.Issuer == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B6 | `case` at `internal/app/engine/exitloop.go:262` — case opts.Submit == nil: | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B7 | `case` at `internal/app/engine/exitloop.go:264` — case strings.TrimSpace(opts.AccountRef) == "": | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B8 | `case` at `internal/app/engine/exitloop.go:266` — case !opts.Costs.Configured(): | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B9 | `if` at `internal/app/engine/exitloop.go:272` — if opts.Ratchet != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B10 | `if` at `internal/app/engine/exitloop.go:275` — if err := ratchet.Validate(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B11 | `if` at `internal/app/engine/exitloop.go:279` — if opts.Ladder != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B12 | `if` at `internal/app/engine/exitloop.go:282` — if err := ladder.Validate(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B13 | `if` at `internal/app/engine/exitloop.go:285` — if id := strings.TrimSpace(opts.CommonPolicy); id != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B14 | `if` at `internal/app/engine/exitloop.go:286` — if _, ok := exitpolicy.CommonPolicyByID(id); !ok { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B15 | `if` at `internal/app/engine/exitloop.go:293` — if clk == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |
| B16 | `if` at `internal/app/engine/exitloop.go:296` — if opts.NewID == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New`, `strings.TrimSpace`, `opts.Costs.Configured`, `exitpolicy.DefaultRatchetConfig`, `ratchet.Validate`, `fmt.Errorf`, `exitpolicy.DefaultLadderPolicy`, `ladder.Validate`, `exitpolicy.CommonPolicyByID`, `clock.System`, `clk.Now` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 12 assignment(s) and 11 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
