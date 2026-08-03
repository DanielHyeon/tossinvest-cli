# Function Logic Map: `TestOrphanSweepKeepsCollidingIntentReservationsHeld`

Source: `internal/journal/reservation_sweep_test.go`  
Function: `TestOrphanSweepKeepsCollidingIntentReservationsHeld`  
Signature: `TestOrphanSweepKeepsCollidingIntentReservationsHeld(params=1, results=0)`  
Source SHA-256: `489ced650b9f96b30419ecc9c21f7911145af1a973372a7f167edb61cfcbdcab`

## Inputs and invariants

- Inputs are the parameters represented by `TestOrphanSweepKeepsCollidingIntentReservationsHeld(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:250 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/reservation_sweep_test.go:253 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | range | internal/journal/reservation_sweep_test.go:256 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/reservation_sweep_test.go:257 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/reservation_sweep_test.go:262 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/reservation_sweep_test.go:265 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openReservationJournal`: errors and returned state remain governed by the function's explicit branches.
- `reserveConfirmedSweepOrder`: errors and returned state remain governed by the function's explicit branches.
- `insertTerminalFillSnapshot`: errors and returned state remain governed by the function's explicit branches.
- `j.SweepReservations`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `Held`: errors and returned state remain governed by the function's explicit branches.
- `reservationState`: errors and returned state remain governed by the function's explicit branches.
- `j.ActiveReconcileStates`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
