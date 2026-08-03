# Function Logic Map: `TestOrphanSweepUsesScopedTerminalWhenLegacyMirrorBelongsToAnotherScope`

Source: `internal/journal/reservation_sweep_test.go`
Function: `TestOrphanSweepUsesScopedTerminalWhenLegacyMirrorBelongsToAnotherScope`
Signature: `TestOrphanSweepUsesScopedTerminalWhenLegacyMirrorBelongsToAnotherScope(params=1, results=0)`
Source SHA-256: `dcfcc7de8f1e832e67d19949578ce40bf3d79d549ad6f99de6f5a714cfcd011e`
Revision: `current`

## Inputs and invariants

- Inputs are `TestOrphanSweepUsesScopedTerminalWhenLegacyMirrorBelongsToAnotherScope(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:286 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/reservation_sweep_test.go:289 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `openReservationJournal`: errors and state follow mapped branches.
- `reserveConfirmedSweepOrder`: errors and state follow mapped branches.
- `insertTerminalFillSnapshot`: errors and state follow mapped branches.
- `insertScopedFillSnapshot`: errors and state follow mapped branches.
- `j.SweepReservations`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
