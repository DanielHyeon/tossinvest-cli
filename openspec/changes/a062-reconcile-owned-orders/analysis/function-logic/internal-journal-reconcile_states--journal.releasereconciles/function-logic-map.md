# Function Logic Map: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`
Function: `Journal.ReleaseReconciles`
Signature: `Journal.ReleaseReconciles(params=2, results=2)`
Source SHA-256: `1a5e5aa3d3c37c940bb43adaebb05b8585256908cf2b28f0112da141ede1eb08`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.ReleaseReconciles(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reconcile_states.go:309 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | range | internal/journal/reconcile_states.go:321 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | switch | internal/journal/reconcile_states.go:329 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | case | internal/journal/reconcile_states.go:330 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | case | internal/journal/reconcile_states.go:332 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | case | internal/journal/reconcile_states.go:334 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | case | internal/journal/reconcile_states.go:336 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/reconcile_states.go:340 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/reconcile_states.go:348 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | range | internal/journal/reconcile_states.go:354 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/reconcile_states.go:357 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/reconcile_states.go:361 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/reconcile_states.go:364 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | range | internal/journal/reconcile_states.go:374 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | if | internal/journal/reconcile_states.go:380 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | if | internal/journal/reconcile_states.go:383 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/journal/reconcile_states.go:384 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | if | internal/journal/reconcile_states.go:393 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `len`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `ValidReconcileReleaseCause`: errors and state follow mapped branches.
- `ValidReconcileCause`: errors and state follow mapped branches.
- `symbolLabel`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `j.db.BeginTx`: errors and state follow mapped branches.
- `tx.Rollback`: errors and state follow mapped branches.
- `scanReconcileState`: errors and state follow mapped branches.
- `tx.QueryRowContext`: errors and state follow mapped branches.
- `activeScopeWhere`: errors and state follow mapped branches.
- `scopeArgs`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `UTC`: errors and state follow mapped branches.
- `j.clk.Now`: errors and state follow mapped branches.
- `formatJournalTime`: errors and state follow mapped branches.
- `tx.ExecContext`: errors and state follow mapped branches.
- `result.RowsAffected`: errors and state follow mapped branches.
- `tx.Commit`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 20; return points: 15; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
