# Function Logic Map: `containsCause`

Source: `internal/journal/reservation_sweep_test.go`
Function: `containsCause`
Signature: `containsCause(params=2, results=1)`
Source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`
Revision: `base`

## Inputs and invariants

- Inputs are `containsCause(params=2, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/journal/reservation_sweep_test.go:111 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/reservation_sweep_test.go:112 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- No nested call is present; behavior is source-local and source-hash bound.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
