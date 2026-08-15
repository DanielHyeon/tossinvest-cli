# Function Logic Map: `fakeBroker.OrderRawByID`

- Source: `internal/verifylive/fake_broker_test.go:441-481`
- Qualified function: `fakeBroker.OrderRawByID`
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
| B1 | `if` at `internal/verifylive/fake_broker_test.go:442` — `if f.beforeOrderRawByID != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:449` — `if !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:453` — `if orderID != f.firedID {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B4 | `else` at `internal/verifylive/fake_broker_test.go:455` — `} else {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:460` — `if filled {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B6 | `if` at `internal/verifylive/fake_broker_test.go:464` — `if f.childIdentitySymbol != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B7 | `if` at `internal/verifylive/fake_broker_test.go:466` — `if err := json.Unmarshal(result, &decoded); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B8 | `if` at `internal/verifylive/fake_broker_test.go:472` — `if f.childIdentityQty != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |
| B9 | `if` at `internal/verifylive/fake_broker_test.go:474` — `if err := json.Unmarshal(result, &decoded); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `f.beforeOrderRawByID` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.log` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Lock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.mu.Unlock` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `mustFilledOrderJSONFrom` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `json.Unmarshal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `panic` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `json.Marshal` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
