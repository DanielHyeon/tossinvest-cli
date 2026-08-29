# Function Logic Map: `validContext`

- Source: `internal/continuationlane/allocation_risk_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan | sealed valid campaign plan | `validPlan` fixture | test fails immediately if plan construction fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | straight-line fixture construction | none | accepted context | continuation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validStopCandidate` | supplies sealed stop | test helper panic on invalid construction | AST |
| `NewExecutionTermsPreimage` | binds explicit test prices to plan | test helper panic on invalid construction | AST |

## State mutations and fallbacks

- Pure fixture only; explicit values replace no production defaults.

## Safety conclusion

- Safe edit boundary: add a sealed plan-bound terms preimage to accepted fixtures.
- High-risk impact: no; evaluator has independent fail-closed validation.
