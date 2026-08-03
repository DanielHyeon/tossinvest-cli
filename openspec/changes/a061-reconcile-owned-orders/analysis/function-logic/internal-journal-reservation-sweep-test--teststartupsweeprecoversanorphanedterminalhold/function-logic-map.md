# Function Logic Map: `TestStartupSweepRecoversAnOrphanedTerminalHold`

Source: `internal/journal/reservation_sweep_test.go`  
Function: `TestStartupSweepRecoversAnOrphanedTerminalHold`  
Signature: `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)`  
Source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`

## Inputs and invariants

- Inputs are the parameters represented by `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:130 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/reservation_sweep_test.go:133 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/reservation_sweep_test.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/reservation_sweep_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/reservation_sweep_test.go:149 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/reservation_sweep_test.go:152 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/reservation_sweep_test.go:155 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openReservationJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `recordEntryDecision`: errors and returned state remain governed by the function's explicit branches.
- `j.Reserve`: errors and returned state remain governed by the function's explicit branches.
- `exposureReserve`: errors and returned state remain governed by the function's explicit branches.
- `mustVersion`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `j.Prepare`: errors and returned state remain governed by the function's explicit branches.
- `reservationPrepare`: errors and returned state remain governed by the function's explicit branches.
- `j.db.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `Held`: errors and returned state remain governed by the function's explicit branches.
- `reservationState`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `j.SweepReservations`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 7 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
