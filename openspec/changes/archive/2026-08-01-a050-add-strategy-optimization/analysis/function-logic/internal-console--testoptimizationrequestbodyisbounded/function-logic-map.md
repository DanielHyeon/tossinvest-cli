# Function Logic Map: `TestOptimizationRequestBodyIsBounded`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| oversized form | authenticated test console | HTTP fixture | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | request construction/send/status and zero-command assertion | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| HTTP client | exercises 4096-byte middleware | immediate failure | AST |

## State mutations and fallbacks

- No command mutation permitted.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: request guard.
