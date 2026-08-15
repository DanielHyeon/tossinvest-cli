# Function Logic Map: `Runner.createConditional`

- Source: `internal/verifylive/mutate.go:503-556`
- Qualified function: `Runner.createConditional`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/mutate.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/mutate.go:504` — `if err := r.checkConditionalCap(sr); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B2 | `if` at `internal/verifylive/mutate.go:508` — `if err := r.gate(sr, request{` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B3 | `if` at `internal/verifylive/mutate.go:515` — `if err := r.appendM0Checkpoint(M0Checkpoint{` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B4 | `if` at `internal/verifylive/mutate.go:522` — `if r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B5 | `if` at `internal/verifylive/mutate.go:525` — `if r.m0BeforeConditionalCreate != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B6 | `if` at `internal/verifylive/mutate.go:526` — `if err := r.m0BeforeConditionalCreate(); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B7 | `if` at `internal/verifylive/mutate.go:534` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B8 | `if` at `internal/verifylive/mutate.go:537` — `if strings.TrimSpace(ref.ID) == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B9 | `if` at `internal/verifylive/mutate.go:540` — `if r.m0AfterConditionalCreate != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B10 | `if` at `internal/verifylive/mutate.go:541` — `if err := r.m0AfterConditionalCreate(); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |
| B11 | `if` at `internal/verifylive/mutate.go:545` — `if err := r.appendM0Checkpoint(M0Checkpoint{` | Preserve source ordering; missing causal authority must HOLD | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.checkConditionalCap` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `conditionalDetail` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.gate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `parseDecimal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.appendM0Checkpoint` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReceiptUsable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0BeforeConditionalCreate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.broker.CreateConditionalOrder` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.logCall` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.TrimSpace` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0AfterConditionalCreate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.created` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Fprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
