# Function Logic Map: `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot`

- Source: `internal/protection/readiness_adapter_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test fixture | exact immutable inputs | test contract | fail assertion |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | test fixture | exact immutable inputs | test contract | fail assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| tested production function | verify fail-closed behavior | no retry | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; test-only setup and assertions.

## Safety conclusion

- Safe edit boundary: update exact-scope expectations only
- High-risk impact: no (test)
