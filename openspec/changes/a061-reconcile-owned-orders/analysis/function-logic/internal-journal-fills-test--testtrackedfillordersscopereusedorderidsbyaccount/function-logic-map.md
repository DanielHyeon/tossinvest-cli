# Function Logic Map: `TestTrackedFillOrdersScopeReusedOrderIDsByAccount`

Source: `internal/journal/fills_test.go`  
Function: `TestTrackedFillOrdersScopeReusedOrderIDsByAccount`  
Signature: `TestTrackedFillOrdersScopeReusedOrderIDsByAccount(params=1, results=0)`  
Source SHA-256: `2004cd42dcb970e432f53ccef544de09a592485a848c3bf492015b5a4c67fbbb`

## Inputs and invariants

- Inputs are the parameters represented by `TestTrackedFillOrdersScopeReusedOrderIDsByAccount(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills_test.go:466 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/fills_test.go:469 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/fills_test.go:472 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/fills_test.go:475 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | range | internal/journal/fills_test.go:479 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/fills_test.go:481 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/fills_test.go:484 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openTestJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `recordConfirmedFillOrder`: errors and returned state remain governed by the function's explicit branches.
- `j.Prepare`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `attempt.MarkDispatchStarted`: errors and returned state remain governed by the function's explicit branches.
- `attempt.MarkAcked`: errors and returned state remain governed by the function's explicit branches.
- `attempt.Settle`: errors and returned state remain governed by the function's explicit branches.
- `j.TrackedFillOrders`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 7 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
