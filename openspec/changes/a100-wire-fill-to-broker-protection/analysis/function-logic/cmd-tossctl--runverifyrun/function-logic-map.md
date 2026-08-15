# Function Logic Map: `runVerifyRun`

- Source: `cmd/tossctl/verify.go:272-400`
- Qualified function: `runVerifyRun`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `cmd/tossctl/verify.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `cmd/tossctl/verify.go:278` — `if opts.list {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B2 | `if` at `cmd/tossctl/verify.go:284` — `if ctx == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B3 | `if` at `cmd/tossctl/verify.go:293` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B4 | `if` at `cmd/tossctl/verify.go:302` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B5 | `if` at `cmd/tossctl/verify.go:306` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B6 | `if` at `cmd/tossctl/verify.go:309` — `if err := validateM0TriggerMode(opts, prior); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B7 | `if` at `cmd/tossctl/verify.go:313` — `if opts.includeTrigger {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B8 | `if` at `cmd/tossctl/verify.go:315` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B9 | `if` at `cmd/tossctl/verify.go:323` — `if steps := verifylive.StepCount(prior); steps > 0 && !opts.resume && len(opts.redo) == 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B10 | `if` at `cmd/tossctl/verify.go:331` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B11 | `if` at `cmd/tossctl/verify.go:336` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B12 | `if` at `cmd/tossctl/verify.go:342` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B13 | `if` at `cmd/tossctl/verify.go:347` — `if holding == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B14 | `if` at `cmd/tossctl/verify.go:351` — `if market == verifylive.MarketUS && !verifylive.SameMarket(verifylive.MarketOf(symbol), market) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B15 | `if` at `cmd/tossctl/verify.go:359` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B16 | `if` at `cmd/tossctl/verify.go:386` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |
| B17 | `if` at `cmd/tossctl/verify.go:395` — `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` | Preserve source ordering; missing causal authority must HOLD | `TestM0CLIForbiddenModesRefuseBeforeBrokerFactoryOrConfirmation` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.OutOrStdout` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `verifylive.WriteSteps` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `cmd.Context` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `context.Background` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `signal.NotifyContext` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `stop` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `acquireVerifyExecutionLock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `executionLock.Release` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Fprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `executionLock.Path` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `verifylive.NormalizeMarket` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `resolveVerifyRecordFor` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `verifylive.LoadEntries` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `validateM0TriggerMode` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `verifylive.OpenCausalReceipt` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `receipt.Close` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `verifylive.StepCount` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `holdVerifyRateBudgetIntent` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
