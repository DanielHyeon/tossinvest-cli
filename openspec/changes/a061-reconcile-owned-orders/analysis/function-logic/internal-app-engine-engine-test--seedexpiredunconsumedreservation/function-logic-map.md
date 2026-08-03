# Function Logic Map: `seedExpiredUnconsumedReservation`

Source: `internal/app/engine/engine_test.go`  
Function: `seedExpiredUnconsumedReservation`  
Signature: `seedExpiredUnconsumedReservation(params=2, results=1)`  
Source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

## Inputs and invariants

- Inputs are the parameters represented by `seedExpiredUnconsumedReservation(params=2, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:495 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/engine_test.go:511 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/engine_test.go:515 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/app/engine/engine_test.go:519 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `t.Helper`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `Add`: errors and returned state remain governed by the function's explicit branches.
- `UTC`: errors and returned state remain governed by the function's explicit branches.
- `time.Now`: errors and returned state remain governed by the function's explicit branches.
- `journal.Open`: errors and returned state remain governed by the function's explicit branches.
- `filepath.Join`: errors and returned state remain governed by the function's explicit branches.
- `clock.NewFake`: errors and returned state remain governed by the function's explicit branches.
- `journal.FixedFSProber`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `j.Close`: errors and returned state remain governed by the function's explicit branches.
- `j.RecordDecision`: errors and returned state remain governed by the function's explicit branches.
- `issued.Add`: errors and returned state remain governed by the function's explicit branches.
- `j.ReservationVersion`: errors and returned state remain governed by the function's explicit branches.
- `j.Reserve`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 7 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
