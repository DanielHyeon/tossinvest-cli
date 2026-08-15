# Function Logic Map: `fakeBroker.CreateConditionalOrder`

- Source: `internal/verifylive/fake_broker_test.go:910-941`
- Qualified function: `fakeBroker.CreateConditionalOrder`
- Revision: `current`
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
| B1 | `if` at `internal/verifylive/fake_broker_test.go:913` — `if body.ClientOrderID != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:917` — `if seen && prior.body == canonical && f.honourIdempotency {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:930` — `if body.ClientOrderID != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:933` — `if strings.HasPrefix(body.ClientOrderID, "TRIGGER-") {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `f.log` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Sprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Lock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Unlock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.nextID` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `parseDecimal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.HasPrefix` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
