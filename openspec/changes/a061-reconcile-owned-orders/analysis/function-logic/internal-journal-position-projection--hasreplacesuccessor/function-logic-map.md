# Function Logic Map: `hasReplaceSuccessor`

Source: `internal/journal/position_projection.go`  
Function: `hasReplaceSuccessor`  
Signature: `hasReplaceSuccessor(params=4, results=2)`  
Source SHA-256: `ae74d3ba1b66a05360e7b5851248fd6814577fa0b34068a89f52c58c10644c7b`

## Inputs and invariants

- Inputs are the parameters represented by `hasReplaceSuccessor(params=4, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection.go:243 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/position_projection.go:248 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `tx.Query`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `rows.Close`: errors and returned state remain governed by the function's explicit branches.
- `rows.Next`: errors and returned state remain governed by the function's explicit branches.
- `rows.Err`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 3 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
