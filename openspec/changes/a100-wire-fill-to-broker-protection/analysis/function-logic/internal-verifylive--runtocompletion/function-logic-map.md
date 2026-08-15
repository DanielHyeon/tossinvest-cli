# Function Logic Map: `runToCompletion`

- Source: `internal/verifylive/runner_test.go:28-59`
- Qualified function: `runToCompletion`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/runner_test.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/runner_test.go:30` — `if opts.IncludeTrigger && opts.Receipt != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B2 | `if` at `internal/verifylive/runner_test.go:35` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B3 | `if` at `internal/verifylive/runner_test.go:38` — `if only.Halted {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B4 | `if` at `internal/verifylive/runner_test.go:44` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B5 | `if` at `internal/verifylive/runner_test.go:47` — `if !first.Halted {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B6 | `if` at `internal/verifylive/runner_test.go:55` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `seedM0TriggerPrerequisites` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.run` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Fatalf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.entries` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.broker.restart` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
