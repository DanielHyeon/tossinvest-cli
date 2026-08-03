# Function Logic Map: `TestAnAmendmentKeepsOneInstance`

Source: `internal/journal/position_projection_test.go`  
Function: `TestAnAmendmentKeepsOneInstance`  
Signature: `TestAnAmendmentKeepsOneInstance(params=1, results=0)`  
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`

## Inputs and invariants

- Inputs are the parameters represented by `TestAnAmendmentKeepsOneInstance(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:334 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/position_projection_test.go:346 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/position_projection_test.go:349 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/position_projection_test.go:352 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/position_projection_test.go:355 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/position_projection_test.go:365 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/position_projection_test.go:368 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/position_projection_test.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/position_projection_test.go:379 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/position_projection_test.go:382 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/journal/position_projection_test.go:385 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/journal/position_projection_test.go:388 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `projectingJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `place`: errors and returned state remain governed by the function's explicit branches.
- `j.RecordFill`: errors and returned state remain governed by the function's explicit branches.
- `fillOf`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `j.Prepare`: errors and returned state remain governed by the function's explicit branches.
- `testIntentFor`: errors and returned state remain governed by the function's explicit branches.
- `amend.MarkDispatchStarted`: errors and returned state remain governed by the function's explicit branches.
- `amend.MarkAcked`: errors and returned state remain governed by the function's explicit branches.
- `amend.ResolveConfirmedWithLineage`: errors and returned state remain governed by the function's explicit branches.
- `terminalFill`: errors and returned state remain governed by the function's explicit branches.
- `currentPosition`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `j.Positions`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 14 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
