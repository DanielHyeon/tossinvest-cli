# Function Logic Map: `exitContextForFill`

Source: `internal/journal/exit_state.go`  
Function: `exitContextForFill`  
Signature: `exitContextForFill(params=3, results=3)`  
Source SHA-256: `f3895fb41abc09f4de2aad1eceeeff1b39ab17ed658b2dc74e02bf7727b46f86`

## Inputs and invariants

- Inputs are the parameters represented by `exitContextForFill(params=3, results=3)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/exit_state.go:875 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/exit_state.go:878 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/exit_state.go:890 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/exit_state.go:895 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/exit_state.go:898 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `resolveFillOrigin`: errors and returned state remain governed by the function's explicit branches.
- `tx.Query`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `instance.Close`: errors and returned state remain governed by the function's explicit branches.
- `instance.Next`: errors and returned state remain governed by the function's explicit branches.
- `instance.Err`: errors and returned state remain governed by the function's explicit branches.
- `instance.Scan`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 4 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
