# Function Logic Map: `Runner.finishTrigger`

- Source: `internal/verifylive/steps_trigger.go:566-659`
- Qualified function: `Runner.finishTrigger`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/steps_trigger.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger.go:586` — `if !obs.triggeredAt.IsZero() && !obs.childSeenAt.IsZero() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B2 | `else` at `internal/verifylive/steps_trigger.go:591` — `} else {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B3 | `if` at `internal/verifylive/steps_trigger.go:596` — `if !obs.triggeredAt.IsZero() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B4 | `switch` at `internal/verifylive/steps_trigger.go:602` — `switch {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B5 | `case` at `internal/verifylive/steps_trigger.go:603` — `case !obs.childFilledAt.IsZero():` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B6 | `if` at `internal/verifylive/steps_trigger.go:604` — `if !r.m0PassReady() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B7 | `case` at `internal/verifylive/steps_trigger.go:617` — `case !obs.triggeredAt.IsZero():` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B8 | `case` at `internal/verifylive/steps_trigger.go:632` — `case obs.crossedWithoutFiring:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B9 | `case` at `internal/verifylive/steps_trigger.go:642` — `case obs.cancelled && obs.raceUnknown:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B10 | `case` at `internal/verifylive/steps_trigger.go:650` — `case obs.cancelled:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B11 | `case` at `internal/verifylive/steps_trigger.go:655` — `default:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sr.observe` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strconv.Itoa` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.observeStamp` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `orNone` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `orDash` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strconv.FormatBool` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.triggeredAt.IsZero` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.childSeenAt.IsZero` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strconv.FormatInt` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Milliseconds` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.childSeenAt.Sub` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.childInterval.String` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `triggerBasis` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `triggerBasisDetail` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `topOfBook` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.childFilledAt.IsZero` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0PassReady` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.fail` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.filled` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.pass` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
