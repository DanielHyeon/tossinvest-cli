# Function Logic Map: `TestTheTriggerStepSkipsWhenTheGridHasNoRoom`

- Source: `internal/verifylive/steps_trigger_test.go:407-417`
- Qualified function: `TestTheTriggerStepSkipsWhenTheGridHasNoRoom`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/steps_trigger_test.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:411` — `if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepSkipsWhenTheGridHasNoRoom` |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:414` — `if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepSkipsWhenTheGridHasNoRoom` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `triggerHarness` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.withBook` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `runToCompletion` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `triggerOptions` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.verdict` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Fatalf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.broker.countRequests` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Error` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
