# Function Logic Map: `BuildProgress`

- Source: `internal/verifylive/report.go:315-343`
- Qualified function: `BuildProgress`
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
| B1 | `range` at `internal/verifylive/report.go:317` — `for _, e := range entries {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B2 | `if` at `internal/verifylive/report.go:318` — `if strings.TrimSpace(p.AccountRef) == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B3 | `range` at `internal/verifylive/report.go:322` — `for _, step := range Steps() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B4 | `if` at `internal/verifylive/report.go:324` — `if !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B5 | `if` at `internal/verifylive/report.go:331` — `if !e.Verdict.Terminal() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B6 | `range` at `internal/verifylive/report.go:336` — `for _, e := range entries {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |
| B7 | `if` at `internal/verifylive/report.go:337` — `if e.Kind == KindM0Checkpoint && e.M0Checkpoint != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Steps` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `LastEntry` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `e.Verdict.Terminal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Outstanding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
