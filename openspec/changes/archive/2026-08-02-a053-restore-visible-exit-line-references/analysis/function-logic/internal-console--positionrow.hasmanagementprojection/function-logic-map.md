# Function Logic Map: `positionRow.HasManagementProjection`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Management status | empty or stable typed status | engine runtime projection | empty means the command/runtime seam did not produce a projection |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless status comparison | none | true only for non-empty typed projection | console management tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | leaf predicate | none | AST evidence |

## State mutations and fallbacks

- Pure boolean projection with no mutation or I/O.

## Safety conclusion

- Safe edit boundary: read-only row state.
- High-risk impact: no direct trading side effect.
