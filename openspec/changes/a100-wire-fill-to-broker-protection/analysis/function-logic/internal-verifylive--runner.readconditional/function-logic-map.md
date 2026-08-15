# Function Logic Map: `Runner.readConditional`

- Source: `internal/verifylive/steps.go:975-1055`
- Qualified function: `Runner.readConditional`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/steps.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps.go:976` — `if r.m0ReadSource != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B2 | `if` at `internal/verifylive/steps.go:982` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B3 | `if` at `internal/verifylive/steps.go:985` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B4 | `if` at `internal/verifylive/steps.go:986` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B5 | `if` at `internal/verifylive/steps.go:991` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B6 | `if` at `internal/verifylive/steps.go:1001` — `if r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B7 | `if` at `internal/verifylive/steps.go:1003` — `if !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B8 | `if` at `internal/verifylive/steps.go:1012` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B9 | `if` at `internal/verifylive/steps.go:1013` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B10 | `if` at `internal/verifylive/steps.go:1018` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B11 | `if` at `internal/verifylive/steps.go:1019` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B12 | `if` at `internal/verifylive/steps.go:1034` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B13 | `if` at `internal/verifylive/steps.go:1035` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B14 | `if` at `internal/verifylive/steps.go:1040` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B15 | `if` at `internal/verifylive/steps.go:1041` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readRetry` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReadSource.ConditionalOrderRaw` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.recordM0Attempts` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Digest` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `parseDecimal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReadContext` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReceiptUsable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `reader.ConditionalOrderRaw` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.broker.ConditionalOrder` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
