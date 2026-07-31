# Function Logic Map: `TestAnImplausibleResetIsNotPresentedAsADetermination`

- Source: `internal/official/ratebudget_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| malformed/extreme raw table | milliseconds, threshold-like nanoseconds, negative and huge integers | regression fixture | every case remains reported but unparsed with raw preserved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | iterate cases and assert kind/raw/reported invariants | test failures only | fatal evidence | this test function |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readRateBudget` | exercise the production parser entry | pure fixture response | CodeGraph + AST |

## State mutations and fallbacks

- Test-only assertions; no production mutation.

## Safety conclusion

- Safe edit boundary: official parser regression test.
- High-risk impact: no production side effect; high-value safety evidence.
