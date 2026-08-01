# Function Logic Map: `fixedOptimizationEvidence.ReadEvidence`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixed fixture | immutable evidence | test | returned unchanged |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| Happy | direct return | none | fixed evidence | commander wiring test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | test double | none | AST |

## State mutations and fallbacks

- Test-only immutable projection.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: no.
