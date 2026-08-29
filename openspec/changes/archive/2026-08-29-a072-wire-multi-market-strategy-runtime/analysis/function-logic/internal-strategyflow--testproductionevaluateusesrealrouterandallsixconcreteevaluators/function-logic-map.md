# Function Logic Map: `TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators`

- Source: `internal/strategyflow/production_integration_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| six lane cases | KR/US continuation, reversal, weekly | fixed test table | subtest fails on any mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| table loop | one concrete lane per case | none | subtest result | this test |
| fixture errors | invalid explicit sealed fixture | none | `Fatal` | this test |
| accepted result checks | lineage/terms/authority exact | none | `Fatalf` | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| concrete test seams | construct explicit sealed lane inputs | return errors, no fallback | AST |
| `Evaluate` | production router plus evaluator | typed refusal | AST |

## State mutations and fallbacks

- Test-only observations; no broker, Guardian, toggle, journal or activation mutation.

## Safety conclusion

- Safe edit boundary: assert exact sealed entry/stop/target for every paired lane.
- High-risk impact: no production mutation; high safety coverage.
