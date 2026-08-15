# Branch Test Map: `TestATriggerWithNoObservableFillLeavesTheChildHeld`

- Source: `internal/verifylive/steps_trigger_test.go:289-350`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:300` — `if err == nil {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:303` — `if len(second.Outstanding) == 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:307` — `if got := h.verdict(StepConditionalTrigger); got != VerdictFail {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/steps_trigger_test.go:310` — `if entry, ok := LastEntry(h.entries(), StepConditionalTrigger); !ok \|\| strings.Contains(entry.Reason, "verify abort") \|\| !strings.Contains(entry.Reason, "수동") {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `range` at `internal/verifylive/steps_trigger_test.go:315` — `for _, a := range Outstanding(entries) {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/steps_trigger_test.go:316` — `if a.Kind == KindOrder {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/steps_trigger_test.go:320` — `if len(out) != 1 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/steps_trigger_test.go:324` — `if !child.Deliberate {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/steps_trigger_test.go:327` — `if child.HeldUntil != StepConditionalTrigger {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/steps_trigger_test.go:333` — `if targets := PendingCleanup(entries); len(targets) != 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `internal/verifylive/steps_trigger_test.go:336` — `if targets := AbortTargets(entries); len(targets) != 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `if` at `internal/verifylive/steps_trigger_test.go:341` — `if abortErr != nil \|\| len(result.Targets) != 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B13 | `if` at `internal/verifylive/steps_trigger_test.go:344` — `if n := h.broker.countRequests("POST /orders/" + child.ID + "/cancel"); n != 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B14 | `if` at `internal/verifylive/steps_trigger_test.go:347` — `if n := h.broker.countRequests("DELETE /conditional-orders/"); n != 0 {` | `TestATriggerWithNoObservableFillLeavesTheChildHeld` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
