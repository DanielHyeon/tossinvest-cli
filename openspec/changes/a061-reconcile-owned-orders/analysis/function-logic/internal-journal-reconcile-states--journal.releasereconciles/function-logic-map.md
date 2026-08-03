# Function Logic Map: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`  
Function: `Journal.ReleaseReconciles`  
Signature: `Journal.ReleaseReconciles(params=2, results=2)`  
Source SHA-256: `1a5e5aa3d3c37c940bb43adaebb05b8585256908cf2b28f0112da141ede1eb08`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.ReleaseReconciles(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reconcile_states.go:309 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | range | internal/journal/reconcile_states.go:321 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | switch | internal/journal/reconcile_states.go:329 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | case | internal/journal/reconcile_states.go:330 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | case | internal/journal/reconcile_states.go:332 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | case | internal/journal/reconcile_states.go:334 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | case | internal/journal/reconcile_states.go:336 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/reconcile_states.go:340 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/reconcile_states.go:348 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | range | internal/journal/reconcile_states.go:354 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/journal/reconcile_states.go:357 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/journal/reconcile_states.go:361 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | if | internal/journal/reconcile_states.go:364 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B14 | range | internal/journal/reconcile_states.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B15 | if | internal/journal/reconcile_states.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B16 | if | internal/journal/reconcile_states.go:383 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B17 | if | internal/journal/reconcile_states.go:384 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B18 | if | internal/journal/reconcile_states.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `len`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `ValidReconcileReleaseCause`: errors and returned state remain governed by the function's explicit branches.
- `ValidReconcileCause`: errors and returned state remain governed by the function's explicit branches.
- `symbolLabel`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `j.db.BeginTx`: errors and returned state remain governed by the function's explicit branches.
- `tx.Rollback`: errors and returned state remain governed by the function's explicit branches.
- `scanReconcileState`: errors and returned state remain governed by the function's explicit branches.
- `tx.QueryRowContext`: errors and returned state remain governed by the function's explicit branches.
- `activeScopeWhere`: errors and returned state remain governed by the function's explicit branches.
- `scopeArgs`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `UTC`: errors and returned state remain governed by the function's explicit branches.
- `j.clk.Now`: errors and returned state remain governed by the function's explicit branches.
- `formatJournalTime`: errors and returned state remain governed by the function's explicit branches.
- `tx.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `result.RowsAffected`: errors and returned state remain governed by the function's explicit branches.
- `tx.Commit`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 20 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
