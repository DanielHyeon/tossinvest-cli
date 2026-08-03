# Function Logic Map: `Journal.ReservationsAwaitingOperator`

Source: `internal/journal/reservation_release.go`
Function: `Journal.ReservationsAwaitingOperator`
Signature: `Journal.ReservationsAwaitingOperator(params=1, results=2)`
Source SHA-256: `d61f428958e0ac5ba535af8148bdcb40d25f2c1893be6feb8f95b1e9af7b2ff2`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.ReservationsAwaitingOperator(params=1, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_release.go:775 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | for | internal/journal/reservation_release.go:781 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/reservation_release.go:790 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/reservation_release.go:799 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | switch | internal/journal/reservation_release.go:809 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | case | internal/journal/reservation_release.go:810 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | case | internal/journal/reservation_release.go:814 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/reservation_release.go:821 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `j.db.QueryContext`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `parseJournalTime`: errors and state follow mapped branches.
- `AttemptState`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 14; return points: 5; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
