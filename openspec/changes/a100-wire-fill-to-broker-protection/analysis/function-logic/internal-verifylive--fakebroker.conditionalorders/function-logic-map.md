# Function Logic Map: `fakeBroker.ConditionalOrders`

- Source: `internal/verifylive/fake_broker_test.go:437-463`
- Qualified function: `fakeBroker.ConditionalOrders`
- Revision: `base`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/fake_broker_test.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:441` — `if f.rejectBadConditionalStatus && status != "" && status != "OPEN" && status != "CLOSED" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B2 | `range` at `internal/verifylive/fake_broker_test.go:448` — `for _, c := range f.conds {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:451` — `if status == "OPEN" && c.Status != "WATCHING" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:454` — `if status != "" && status != "OPEN" && status != "CLOSED" && c.Status != status {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:457` — `if symbol != "" && c.Symbol != symbol {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `f.log` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Lock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Unlock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
