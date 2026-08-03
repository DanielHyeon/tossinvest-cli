# Function Logic Map: `TestTrackedFillOrdersCarryLineage`

Source: `internal/journal/fills_test.go`  
Function: `TestTrackedFillOrdersCarryLineage`  
Signature: `TestTrackedFillOrdersCarryLineage(params=1, results=0)`  
Source SHA-256: `2004cd42dcb970e432f53ccef544de09a592485a848c3bf492015b5a4c67fbbb`

## Inputs and invariants

- Inputs are the parameters represented by `TestTrackedFillOrdersCarryLineage(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills_test.go:733 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/fills_test.go:744 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/fills_test.go:747 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/fills_test.go:750 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/fills_test.go:753 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/fills_test.go:761 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | range | internal/journal/fills_test.go:765 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/fills_test.go:766 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/fills_test.go:768 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/fills_test.go:774 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/journal/fills_test.go:781 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/journal/fills_test.go:785 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | range | internal/journal/fills_test.go:788 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B14 | if | internal/journal/fills_test.go:789 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openTestJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `recordConfirmedFillOrder`: errors and returned state remain governed by the function's explicit branches.
- `j.RecordFill`: errors and returned state remain governed by the function's explicit branches.
- `observation`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `j.Prepare`: errors and returned state remain governed by the function's explicit branches.
- `amend.MarkDispatchStarted`: errors and returned state remain governed by the function's explicit branches.
- `amend.MarkAcked`: errors and returned state remain governed by the function's explicit branches.
- `amend.ResolveConfirmedWithLineage`: errors and returned state remain governed by the function's explicit branches.
- `j.TrackedFillOrders`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 15 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
