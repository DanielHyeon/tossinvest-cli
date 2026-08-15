# Function Logic Map: `TestATriggerWithNoObservableFillLeavesTheChildHeld`

- Source: `internal/verifylive/steps_trigger_test.go:289-350`
- Qualified function: `TestATriggerWithNoObservableFillLeavesTheChildHeld`
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
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:300` — `if err == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:303` — `if len(second.Outstanding) == 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:307` — `if got := h.verdict(StepConditionalTrigger); got != VerdictFail {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B4 | `if` at `internal/verifylive/steps_trigger_test.go:310` — `if entry, ok := LastEntry(h.entries(), StepConditionalTrigger); !ok \|\| strings.Contains(entry.Reason, "verify abort") \|\| !strings.Contains(entry.Reason, "수동") {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B5 | `range` at `internal/verifylive/steps_trigger_test.go:315` — `for _, a := range Outstanding(entries) {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B6 | `if` at `internal/verifylive/steps_trigger_test.go:316` — `if a.Kind == KindOrder {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B7 | `if` at `internal/verifylive/steps_trigger_test.go:320` — `if len(out) != 1 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B8 | `if` at `internal/verifylive/steps_trigger_test.go:324` — `if !child.Deliberate {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B9 | `if` at `internal/verifylive/steps_trigger_test.go:327` — `if child.HeldUntil != StepConditionalTrigger {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B10 | `if` at `internal/verifylive/steps_trigger_test.go:333` — `if targets := PendingCleanup(entries); len(targets) != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B11 | `if` at `internal/verifylive/steps_trigger_test.go:336` — `if targets := AbortTargets(entries); len(targets) != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B12 | `if` at `internal/verifylive/steps_trigger_test.go:341` — `if abortErr != nil \|\| len(result.Targets) != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B13 | `if` at `internal/verifylive/steps_trigger_test.go:344` — `if n := h.broker.countRequests("POST /orders/" + child.ID + "/cancel"); n != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |
| B14 | `if` at `internal/verifylive/steps_trigger_test.go:347` — `if n := h.broker.countRequests("DELETE /conditional-orders/"); n != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestATriggerWithNoObservableFillLeavesTheChildHeld` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `triggerHarness` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.firesOnRead` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `triggerOptions` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `seedM0TriggerPrerequisites` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.run` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Error` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.verdict` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Fatalf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `LastEntry` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.entries` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.Contains` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Outstanding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `PendingCleanup` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `AbortTargets` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.runner` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `runner.Abort` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `context.Background` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
