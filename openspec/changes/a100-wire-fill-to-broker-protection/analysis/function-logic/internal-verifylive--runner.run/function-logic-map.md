# Function Logic Map: `Runner.Run`

- Source: `internal/verifylive/runner.go:462-595`
- Qualified function: `Runner.Run`
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
| B1 | `if` at `internal/verifylive/runner.go:464` — `if r.includeTrigger {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B2 | `if` at `internal/verifylive/runner.go:466` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B3 | `if` at `internal/verifylive/runner.go:474` — `if r.includeTrigger && !r.m0ReceiptUsable() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B4 | `if` at `internal/verifylive/runner.go:479` — `if pending, ok, pendingErr := r.m0PendingCheckpoint(); pendingErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B5 | `else` at `internal/verifylive/runner.go:484` — `} else if ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B6 | `if` at `internal/verifylive/runner.go:484` — `} else if ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B7 | `if` at `internal/verifylive/runner.go:492` — `if halt, err := r.approveBatch(ctx); err != nil \|\| halt != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B8 | `if` at `internal/verifylive/runner.go:502` — `if outcome, err, stop := r.cleanup(ctx); outcome.Step != "" \|\| err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B9 | `if` at `internal/verifylive/runner.go:503` — `if outcome.Step != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B10 | `if` at `internal/verifylive/runner.go:506` — `if stop {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B11 | `if` at `internal/verifylive/runner.go:509` — `if outcome.Reason == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B12 | `range` at `internal/verifylive/runner.go:517` — `for _, step := range Steps() {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B13 | `if` at `internal/verifylive/runner.go:518` — `if err := ctx.Err(); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B14 | `if` at `internal/verifylive/runner.go:525` — `if settled, verdict := r.settled(step.ID); settled {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B15 | `if` at `internal/verifylive/runner.go:536` — `if reason, skip := r.preflight(step); skip {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B16 | `else` at `internal/verifylive/runner.go:538` — `} else {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B17 | `if` at `internal/verifylive/runner.go:544` — `if err := r.recorder.Append(entry); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B18 | `if` at `internal/verifylive/runner.go:553` — `if sr.verdict == VerdictAwaitingRestart {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B19 | `if` at `internal/verifylive/runner.go:559` — `if errors.Is(sr.abort, ErrNotATerminal) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B20 | `if` at `internal/verifylive/runner.go:565` — `if errors.Is(sr.abort, ErrOutsidePlan) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B21 | `if` at `internal/verifylive/runner.go:574` — `if errors.Is(sr.abort, ErrM0TerminalHold) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B22 | `if` at `internal/verifylive/runner.go:580` — `if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B23 | `if` at `internal/verifylive/runner.go:589` — `if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.m0Receipt.AcquireRunLease` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `truncateError` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `lease.Release` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.writeBanner` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0ReceiptUsable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `errors.New` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.outstanding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0PendingCheckpoint` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0RecoverPending` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.approveBatch` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.cleanup` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Steps` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `ctx.Err` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.settled` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Fprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `StepLabel` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.preflight` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
