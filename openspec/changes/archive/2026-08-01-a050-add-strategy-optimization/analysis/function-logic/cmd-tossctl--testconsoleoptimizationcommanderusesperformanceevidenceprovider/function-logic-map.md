# Function Logic Map: `TestConsoleOptimizationCommanderUsesPerformanceEvidenceProvider`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixed complete evidence | SHA-256-shaped fixture | test fixture | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | open/read/compare provider result | temporary control DB only | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| commander constructor/read | verifies provider propagation | immediate failure | AST |

## State mutations and fallbacks

- Test-only lifecycle read.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: no.
