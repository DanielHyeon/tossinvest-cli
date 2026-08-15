# Function Logic Map: `stepRun.resolve`

- Source: `internal/verifylive/runner.go:954-976`
- Qualified function: `stepRun.resolve`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/runner.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` at `internal/verifylive/runner.go:955` — `switch {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B2 | `case` at `internal/verifylive/runner.go:956` — `case err == nil:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B3 | `if` at `internal/verifylive/runner.go:957` — `if sr.verdict == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B4 | `case` at `internal/verifylive/runner.go:960` — `case errors.Is(err, ErrOutsidePlan):` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B5 | `case` at `internal/verifylive/runner.go:962` — `case errors.Is(err, ErrM0TerminalHold):` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B6 | `case` at `internal/verifylive/runner.go:965` — `case errors.Is(err, ErrRefused), errors.Is(err, ErrConfirmationExpired):` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B7 | `case` at `internal/verifylive/runner.go:967` — `case errors.Is(err, ErrNotATerminal):` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B8 | `case` at `internal/verifylive/runner.go:970` — `case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B9 | `case` at `internal/verifylive/runner.go:973` — `default:` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sr.pass` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `errors.Is` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.outsidePlan` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.fail` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `truncateError` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.refuse` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `err.Error` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
