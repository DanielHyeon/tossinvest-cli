# Function Logic Map: `Runner.pollTrigger`

- Source: `internal/verifylive/steps_trigger.go:315-454`
- Qualified function: `Runner.pollTrigger`
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
| B1 | `if` at `internal/verifylive/steps_trigger.go:322` — `if obs.triggeredAt.IsZero() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B2 | `if` at `internal/verifylive/steps_trigger.go:325` — `if err == nil && obs.crossedAt.IsZero() && top.last > 0 && top.last <= w.trigger.Price {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B3 | `if` at `internal/verifylive/steps_trigger.go:334` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B4 | `if` at `internal/verifylive/steps_trigger.go:339` — `if r.m0Receipt != nil && !obs.cancelled {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B5 | `if` at `internal/verifylive/steps_trigger.go:344` — `if evidence, fired := firedEvidence(co); fired && obs.triggeredAt.IsZero() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B6 | `if` at `internal/verifylive/steps_trigger.go:350` — `if id := strings.TrimSpace(co.First.TriggeredOrderID); id != "" && obs.childID == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B7 | `if` at `internal/verifylive/steps_trigger.go:351` — `if r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B8 | `if` at `internal/verifylive/steps_trigger.go:352` — `if co.ID != w.conditionalID \|\| co.ClientOrderID != r.m0PendingClient \|\|` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B9 | `if` at `internal/verifylive/steps_trigger.go:360` — `if err := r.appendM0Checkpoint(M0Checkpoint{` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B10 | `if` at `internal/verifylive/steps_trigger.go:366` — `if r.m0AfterChildCheckpoint != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B11 | `if` at `internal/verifylive/steps_trigger.go:367` — `if err := r.m0AfterChildCheckpoint(); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B12 | `if` at `internal/verifylive/steps_trigger.go:371` — `if r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B13 | `if` at `internal/verifylive/steps_trigger.go:372` — `if r.m0ReceiptLease == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B14 | `if` at `internal/verifylive/steps_trigger.go:382` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B15 | `if` at `internal/verifylive/steps_trigger.go:386` — `if r.m0AfterParentCausal != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B16 | `if` at `internal/verifylive/steps_trigger.go:387` — `if err := r.m0AfterParentCausal(); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B17 | `if` at `internal/verifylive/steps_trigger.go:402` — `if obs.childID == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B18 | `if` at `internal/verifylive/steps_trigger.go:406` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B19 | `if` at `internal/verifylive/steps_trigger.go:407` — `if r.m0Receipt != nil && (r.m0ReceiptErr != nil \|\| r.m0CriticalWindow) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B20 | `if` at `internal/verifylive/steps_trigger.go:414` — `if strings.EqualFold(view.Status, "FILLED") &&` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B21 | `if` at `internal/verifylive/steps_trigger.go:419` — `if filledQuantity <= 0 && !strings.EqualFold(view.Status, "FILLED") {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B22 | `if` at `internal/verifylive/steps_trigger.go:422` — `if r.m0CriticalWindow && r.m0Gap {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B23 | `if` at `internal/verifylive/steps_trigger.go:425` — `if r.m0ReceiptUsable() &&` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B24 | `if` at `internal/verifylive/steps_trigger.go:431` — `if r.m0CriticalWindow && r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B25 | `if` at `internal/verifylive/steps_trigger.go:432` — `if r.m0ReceiptLease == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B26 | `if` at `internal/verifylive/steps_trigger.go:435` — `if !r.m0ChildAttemptAfterParent() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B27 | `if` at `internal/verifylive/steps_trigger.go:445` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `obs.triggeredAt.IsZero` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.marketTop` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `obs.crossedAt.IsZero` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Fprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `trim` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.readConditional` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `firedEvidence` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.TrimSpace` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReceiptUsable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.EqualFold` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.expireDate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.appendM0Checkpoint` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0AfterChildCheckpoint` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReceiptLease.RecordCausal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0Receipt.tag` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0AfterParentCausal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.created` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.markHeld` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
