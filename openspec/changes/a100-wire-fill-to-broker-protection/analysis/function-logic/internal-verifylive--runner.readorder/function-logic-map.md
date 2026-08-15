# Function Logic Map: `Runner.readOrder`

- Source: `internal/verifylive/steps.go:924-973`
- Qualified function: `Runner.readOrder`
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
| B1 | `if` at `internal/verifylive/steps.go:925` — `if r.m0ReadSource != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B2 | `if` at `internal/verifylive/steps.go:931` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B3 | `if` at `internal/verifylive/steps.go:934` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B4 | `if` at `internal/verifylive/steps.go:935` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B5 | `if` at `internal/verifylive/steps.go:940` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B6 | `if` at `internal/verifylive/steps.go:944` — `if err := json.Unmarshal(raw, &view); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B7 | `if` at `internal/verifylive/steps.go:953` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B8 | `if` at `internal/verifylive/steps.go:954` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B9 | `if` at `internal/verifylive/steps.go:959` — `if r.m0ReceiptErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B10 | `if` at `internal/verifylive/steps.go:960` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B11 | `if` at `internal/verifylive/steps.go:966` — `if err := json.Unmarshal(raw, &view); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |
| B12 | `if` at `internal/verifylive/steps.go:967` — `if r.m0CriticalWindow {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readRetry` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReadSource.OrderRawByID` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.recordM0Attempts` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `DigestBytes` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `json.Unmarshal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReadContext` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.broker.OrderRawByID` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
