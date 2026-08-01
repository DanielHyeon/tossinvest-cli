# Function Logic Map: `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| registered route AST | valid test/domain fixture | route table | fail test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | static extraction | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- static extraction.

## Safety conclusion

- Safe edit boundary: not-applicable test evidence.
- High-risk impact: no.
