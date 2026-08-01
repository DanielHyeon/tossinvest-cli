# Function Logic Map: `TestOptimizationRejectsAClientInventedPolicy`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| option ID | exact owner-registered finite value | descriptor registry | HTTP 400 and zero command/save calls |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invented option is not fully rejected | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| lifecycle preview route | validates at owner registry boundary | invalid candidate maps to 400 | AST |

## State mutations and fallbacks

- Neither lifecycle preview nor legacy save may accept a client-invented value.

## Safety conclusion

- Safe edit boundary: finite option validation contract.
- High-risk impact: yes; arbitrary strategy values are fail-closed.
