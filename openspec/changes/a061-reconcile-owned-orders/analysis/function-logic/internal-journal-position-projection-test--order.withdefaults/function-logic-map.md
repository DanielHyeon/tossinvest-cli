# Function Logic Map: `order.withDefaults`

Source: `internal/journal/position_projection_test.go`  
Function: `order.withDefaults`  
Signature: `order.withDefaults(params=0, results=1)`  
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`

## Inputs and invariants

- Inputs are the parameters represented by `order.withDefaults(params=0, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:36 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/position_projection_test.go:39 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/position_projection_test.go:42 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/position_projection_test.go:45 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/position_projection_test.go:48 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/position_projection_test.go:51 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- No outbound call is present; behavior is local and deterministic.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 6 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
