# Function Logic Map: `Runner.stepConditionalTrigger`

- Source: `internal/verifylive/steps_trigger.go:91-166`
- Qualified function: `Runner.stepConditionalTrigger`
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
| B1 | `if` at `internal/verifylive/steps_trigger.go:92` — `if r.deferredForm(sr.step) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B2 | `if` at `internal/verifylive/steps_trigger.go:98` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B3 | `if` at `internal/verifylive/steps_trigger.go:101` — `if sellable < MinQuantity {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B4 | `if` at `internal/verifylive/steps_trigger.go:111` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B5 | `if` at `internal/verifylive/steps_trigger.go:115` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |
| B6 | `if` at `internal/verifylive/steps_trigger.go:145` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.deferredForm` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.recordTriggerUnverified` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.readSellable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.skip` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Sprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `trim` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.marketTop` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `NearStopTrigger` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `MarketOf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `truncateError` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.observe` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strconv.FormatBool` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `topOfBook` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `newToken` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.expireDate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.readHolding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.createConditional` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `sr.joinChain` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.watchTrigger` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
