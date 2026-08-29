# Function Logic Map: `strategyflowContext`

- Source: `internal/reversallane/strategyflow_testseam.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market/lane/symbol/candidate/config/time | explicit paired test scope | production integration table | constructor error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| constructor errors | plan/cap/terms invalid | none | error | tagged integration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sealed constructors | build plan, cap and explicit terms | fail closed, no fallback | AST |

## State mutations and fallbacks

- Test seam only; no live source or order authority.

## Safety conclusion

- Safe edit boundary: bind explicit test prices to the exact plan.
- High-risk impact: no authority surface.
