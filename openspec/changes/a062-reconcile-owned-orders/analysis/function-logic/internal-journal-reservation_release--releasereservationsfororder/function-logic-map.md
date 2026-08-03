# Function Logic Map: `releaseReservationsForOrder`

Source: `internal/journal/reservation_release.go`
Function: `releaseReservationsForOrder`
Signature: `releaseReservationsForOrder(params=7, results=2)`
Source SHA-256: `d61f428958e0ac5ba535af8148bdcb40d25f2c1893be6feb8f95b1e9af7b2ff2`
Revision: `current`

## Inputs and invariants

- Inputs are `releaseReservationsForOrder(params=7, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_release.go:147 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.TrimSpace`: errors and state follow mapped branches.
- `releaseWhere`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
