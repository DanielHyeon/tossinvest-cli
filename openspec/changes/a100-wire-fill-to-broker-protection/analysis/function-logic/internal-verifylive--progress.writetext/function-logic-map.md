# Function Logic Map: `Progress.WriteText`

- Source: `internal/verifylive/report.go:346-379`
- Qualified function: `Progress.WriteText`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/report.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/report.go:348` — `if len(p.Steps) == 0 && len(p.M0Checkpoints) == 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B2 | `range` at `internal/verifylive/report.go:356` — `for _, s := range p.Steps {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B3 | `if` at `internal/verifylive/report.go:359` — `if len(p.Pending) > 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B4 | `if` at `internal/verifylive/report.go:362` — `if p.AwaitingRestart != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B5 | `if` at `internal/verifylive/report.go:366` — `if len(p.Outstanding) > 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B6 | `range` at `internal/verifylive/report.go:368` — `for _, a := range p.Outstanding {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B7 | `if` at `internal/verifylive/report.go:373` — `if len(p.M0Checkpoints) > 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B8 | `range` at `internal/verifylive/report.go:375` — `for _, checkpoint := range p.M0Checkpoints {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Fprintln` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `orNone` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `VerdictLabel` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `truncate` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `joinSteps` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
