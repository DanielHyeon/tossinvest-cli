# Function Logic Map: `TestOptimizationRejectsUnexpectedAndDuplicateFields`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| malformed forms | unexpected or duplicate fields | test table | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | cases and subtests assert 400 before commander | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| HTTP harness | exercises strict handler validation | immediate failure | AST |

## State mutations and fallbacks

- No command mutation permitted.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: request guard.
