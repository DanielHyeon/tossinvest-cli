# Function Logic Map: `validateOptimizationForm`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| parsed POST form | server-rendered action-specific fields, one value each | handler contract | unknown action/field or duplicate returns error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B13 | action/allowlist/value-count validation | none | first invalid condition errors | request-bound tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure validation | no retry | AST + request tests |

## State mutations and fallbacks

- Reads parsed form only; no command call or mutation.

## Safety conclusion

- Safe edit boundary: request validation; high-risk impact: fail-closed control input.
