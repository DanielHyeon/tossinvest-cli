# Function Logic Map: `GatewayAdapter.PlaceStrategyEntry`

- Source: `internal/strategydispatch/adapters.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan | opaque decision, Guardian decision/generation, exact quantity, LIMIT/KRW | atomic issuer | reject before gateway |
| decimals | canonical and exactly representable by legacy float boundary | journal/decision receipt | reject lossy conversion |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 93: `if a.Gateway == nil \|\| !plan.Decision.Valid() \|\| plan.Order.OrderType != "LIMIT" \|\| plan.Order.Currency != "KRW" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 98: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 102: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 109: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `orderintent.NormalizePlace` | canonical official place intent | fail closed | AST |
| `execgw.Gateway.Place` | existing official mutation path | journaled Outcome preserved | compile assertion |

## State mutations and fallbacks

- No alternative broker or raw HTTP surface exists in this package.

## Safety conclusion

- Safe edit boundary: exact adapter to existing execgw only.
- High-risk impact: yes, this is the sole official mutation call.
