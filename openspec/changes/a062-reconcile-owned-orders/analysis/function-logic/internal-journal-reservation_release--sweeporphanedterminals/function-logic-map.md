# Function Logic Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`
Function: `sweepOrphanedTerminals`
Signature: `sweepOrphanedTerminals(params=3, results=2)`
Source SHA-256: `d61f428958e0ac5ba535af8148bdcb40d25f2c1893be6feb8f95b1e9af7b2ff2`
Revision: `current`

## Inputs and invariants

- Inputs are `sweepOrphanedTerminals(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_release.go:540 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/reservation_release.go:583 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | for | internal/journal/reservation_release.go:587 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/reservation_release.go:589 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/reservation_release.go:597 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | range | internal/journal/reservation_release.go:606 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/reservation_release.go:607 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/reservation_release.go:612 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/reservation_release.go:621 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `releaseWhere`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `tx.QueryContext`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- `applyTx.invalidate`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `enterReconcileScopeInTx`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 10; return points: 7; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
