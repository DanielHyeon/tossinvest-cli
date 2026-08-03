# Function Logic Map: `TestStartupSweepRecoversAnOrphanedTerminalHold`

Source: `internal/journal/reservation_sweep_test.go`  
Function: `TestStartupSweepRecoversAnOrphanedTerminalHold`  
Signature: `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)`  
Source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`

## Inputs and invariants

- Inputs are the parameters in `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:130 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/reservation_sweep_test.go:133 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/reservation_sweep_test.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/reservation_sweep_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/reservation_sweep_test.go:149 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/reservation_sweep_test.go:152 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/reservation_sweep_test.go:155 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openReservationJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `recordEntryDecision`: returned errors and state follow the mapped branches.
- `j.Reserve`: returned errors and state follow the mapped branches.
- `exposureReserve`: returned errors and state follow the mapped branches.
- `mustVersion`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.Prepare`: returned errors and state follow the mapped branches.
- `reservationPrepare`: returned errors and state follow the mapped branches.
- `j.db.ExecContext`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `Held`: returned errors and state follow the mapped branches.
- `reservationState`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `j.SweepReservations`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 7 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
