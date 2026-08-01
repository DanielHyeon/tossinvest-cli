# Function Logic Map: `conditionalOriginOf`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| conditional composite key and scoped engine set | own id/account/day only | bounded journal result | miss stays unknown |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal unreadable | none | unknown | conditional origin test |
| B2 | exact scoped hit | none | engine | conditional origin test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure classification | no retry | AST + conditional tests |

## State mutations and fallbacks

- Never borrows the conditional creation time for a triggered plain order.

## Safety conclusion

- Safe edit boundary: presentation classification.
- High-risk impact: fail-closed display only.
