# Function Logic Map: `harness.build`

- Source: `internal/verifylive/fake_broker_test.go:1207-1271`
- Qualified function: `harness.build`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/fake_broker_test.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:1210` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:1214` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:1221` — `if opts.IncludeTrigger {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:1222` — `if !h.broker.suppressAttemptTrace {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:1223` — `if h.m0Client == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B6 | `if` at `internal/verifylive/fake_broker_test.go:1233` — `if opts.AccountRef == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B7 | `if` at `internal/verifylive/fake_broker_test.go:1236` — `if opts.Symbol == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B8 | `if` at `internal/verifylive/fake_broker_test.go:1246` — `if h.broker.expireKeysOnSleep {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B9 | `if` at `internal/verifylive/fake_broker_test.go:1252` — `if h.fixedPID != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B10 | `if` at `internal/verifylive/fake_broker_test.go:1256` — `if h.fixedInstanceID != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B11 | `if` at `internal/verifylive/fake_broker_test.go:1266` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `h.t.Helper` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `LoadEntries` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.t.Fatalf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `OpenRecorder` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `rec.Close` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `newM0HTTPClient` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.op.confirmer` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.op.batchConfirmer` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.now.Add` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.broker.expireKeys` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Sprintf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `New` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `closeRecord` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
