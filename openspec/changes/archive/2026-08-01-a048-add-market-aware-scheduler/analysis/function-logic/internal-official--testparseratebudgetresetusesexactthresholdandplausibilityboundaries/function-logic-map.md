# Function Logic Map: `TestParseRateBudgetResetUsesExactThresholdAndPlausibilityBoundaries`

- Source: `internal/official/ratebudget_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| boundary table | whitespace, threshold, plausibility, wrapping raw, zero observed-at | explicit fixtures | any raw/reset/kind tuple mismatch fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | iterate parser cases and compare complete result tuple | test failures only | fatal evidence | this test function |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ParseRateBudgetReset` | verify the authoritative derivation surface | pure | CodeGraph + AST |

## State mutations and fallbacks

- Test-only table covering inclusive bounds and one-second violations.

## Safety conclusion

- Safe edit boundary: official reset helper evidence.
- High-risk impact: no production side effect; high-value safety evidence.
