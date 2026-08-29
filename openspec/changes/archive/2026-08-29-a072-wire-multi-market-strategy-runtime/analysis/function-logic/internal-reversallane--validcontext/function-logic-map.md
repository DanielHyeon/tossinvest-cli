# Function Logic Map: `validContext`

- Source: `internal/reversallane/allocation_risk_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan/ordinal | sealed campaign plan and valid leg | reversal test fixture | invalid fixture is rejected by evaluator |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | straight-line fixture construction | none | accepted context | reversal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewExecutionTermsPreimage` | binds explicit test prices to plan | test helper panic on invalid construction | AST |

## State mutations and fallbacks

- Pure fixture only; no execution, broker, Guardian or activation call.

## Safety conclusion

- Safe edit boundary: add sealed plan-bound explicit terms to accepted fixtures.
- High-risk impact: no; evaluator independently validates ordering and seal.
