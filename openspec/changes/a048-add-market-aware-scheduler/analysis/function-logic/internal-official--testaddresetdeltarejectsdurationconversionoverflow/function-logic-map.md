# Function Logic Map: `TestAddResetDeltaRejectsDurationConversionOverflow`

- Source: `internal/official/ratebudget_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| seconds | one above maximum representable duration seconds | explicit fixture | helper returns zero,false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | overflow candidate returns nonzero or true | test failure only | fatal evidence | this test function |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `addResetDelta` | directly exercise conversion guard | pure | CodeGraph + AST |

## State mutations and fallbacks

- Test-only assertion; no production mutation.

## Safety conclusion

- Safe edit boundary: official delta arithmetic evidence.
- High-risk impact: no production side effect; high-value safety evidence.
