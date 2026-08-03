# Function Logic Map: `seedExpiredUnconsumedReservation`

Source: `internal/app/engine/engine_test.go`
Function: `seedExpiredUnconsumedReservation`
Signature: `seedExpiredUnconsumedReservation(params=2, results=1)`
Source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`
Revision: `current`

## Inputs and invariants

- Inputs are `seedExpiredUnconsumedReservation(params=2, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:495 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/engine_test.go:511 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/engine_test.go:515 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/app/engine/engine_test.go:519 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `t.Helper`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `Add`: errors and state follow mapped branches.
- `UTC`: errors and state follow mapped branches.
- `time.Now`: errors and state follow mapped branches.
- `journal.Open`: errors and state follow mapped branches.
- `filepath.Join`: errors and state follow mapped branches.
- `clock.NewFake`: errors and state follow mapped branches.
- `journal.FixedFSProber`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `unknown`: errors and state follow mapped branches.
- `j.Close`: errors and state follow mapped branches.
- `j.RecordDecision`: errors and state follow mapped branches.
- `issued.Add`: errors and state follow mapped branches.
- `j.ReservationVersion`: errors and state follow mapped branches.
- `j.Reserve`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 1; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
