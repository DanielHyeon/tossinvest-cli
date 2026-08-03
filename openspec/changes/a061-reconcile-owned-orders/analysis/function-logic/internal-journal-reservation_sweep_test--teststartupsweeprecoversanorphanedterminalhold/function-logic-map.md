# Function Logic Map: `TestStartupSweepRecoversAnOrphanedTerminalHold`

Source: `internal/journal/reservation_sweep_test.go`
Function: `TestStartupSweepRecoversAnOrphanedTerminalHold`
Signature: `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)`
Source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`
Revision: `base`

## Inputs and invariants

- Inputs are `TestStartupSweepRecoversAnOrphanedTerminalHold(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:130 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/reservation_sweep_test.go:133 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/reservation_sweep_test.go:139 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/reservation_sweep_test.go:144 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/reservation_sweep_test.go:149 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/reservation_sweep_test.go:152 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/reservation_sweep_test.go:155 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `openReservationJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `recordEntryDecision`: errors and state follow mapped branches.
- `j.Reserve`: errors and state follow mapped branches.
- `exposureReserve`: errors and state follow mapped branches.
- `mustVersion`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `j.Prepare`: errors and state follow mapped branches.
- `reservationPrepare`: errors and state follow mapped branches.
- `j.db.ExecContext`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `Held`: errors and state follow mapped branches.
- `reservationState`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `j.SweepReservations`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
