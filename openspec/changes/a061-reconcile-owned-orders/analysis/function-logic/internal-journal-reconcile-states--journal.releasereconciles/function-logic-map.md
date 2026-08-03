# Function Logic Map: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`  
Function: `Journal.ReleaseReconciles`  
Signature: `Journal.ReleaseReconciles(params=2, results=2)`  
Source SHA-256: `1a5e5aa3d3c37c940bb43adaebb05b8585256908cf2b28f0112da141ede1eb08`

## Inputs and invariants

- Inputs are the parameters in `Journal.ReleaseReconciles(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reconcile_states.go:309 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/journal/reconcile_states.go:321 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | switch | internal/journal/reconcile_states.go:329 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/journal/reconcile_states.go:330 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | case | internal/journal/reconcile_states.go:332 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | case | internal/journal/reconcile_states.go:334 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | case | internal/journal/reconcile_states.go:336 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/reconcile_states.go:340 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/reconcile_states.go:348 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | range | internal/journal/reconcile_states.go:354 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/reconcile_states.go:357 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/journal/reconcile_states.go:361 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | if | internal/journal/reconcile_states.go:364 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | range | internal/journal/reconcile_states.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B15 | if | internal/journal/reconcile_states.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B16 | if | internal/journal/reconcile_states.go:383 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B17 | if | internal/journal/reconcile_states.go:384 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B18 | if | internal/journal/reconcile_states.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `len`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `ValidReconcileReleaseCause`: returned errors and state follow the mapped branches.
- `ValidReconcileCause`: returned errors and state follow the mapped branches.
- `symbolLabel`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `j.db.BeginTx`: returned errors and state follow the mapped branches.
- `tx.Rollback`: returned errors and state follow the mapped branches.
- `scanReconcileState`: returned errors and state follow the mapped branches.
- `tx.QueryRowContext`: returned errors and state follow the mapped branches.
- `activeScopeWhere`: returned errors and state follow the mapped branches.
- `scopeArgs`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `UTC`: returned errors and state follow the mapped branches.
- `j.clk.Now`: returned errors and state follow the mapped branches.
- `formatJournalTime`: returned errors and state follow the mapped branches.
- `tx.ExecContext`: returned errors and state follow the mapped branches.
- `result.RowsAffected`: returned errors and state follow the mapped branches.
- `tx.Commit`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 20 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
