# Function Logic Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`  
Function: `Journal.ReleaseReconcile`  
Signature: `Journal.ReleaseReconcile(params=2, results=3)`  
Source SHA-256: `f07e1a91c10a72e1226e5cf5328d461def19b571714145d31ccb838c2e402e19`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.ReleaseReconcile(params=2, results=3)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/journal/reconcile_states.go:251 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | case | internal/journal/reconcile_states.go:252 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | case | internal/journal/reconcile_states.go:255 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | case | internal/journal/reconcile_states.go:258 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/reconcile_states.go:267 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/reconcile_states.go:274 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/reconcile_states.go:277 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/reconcile_states.go:280 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/reconcile_states.go:284 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/reconcile_states.go:290 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `ValidReconcileReleaseCause`: errors and returned state remain governed by the function's explicit branches.
- `UTC`: errors and returned state remain governed by the function's explicit branches.
- `j.clk.Now`: errors and returned state remain governed by the function's explicit branches.
- `formatJournalTime`: errors and returned state remain governed by the function's explicit branches.
- `j.db.BeginTx`: errors and returned state remain governed by the function's explicit branches.
- `tx.Rollback`: errors and returned state remain governed by the function's explicit branches.
- `scanReconcileState`: errors and returned state remain governed by the function's explicit branches.
- `tx.QueryRowContext`: errors and returned state remain governed by the function's explicit branches.
- `activeScopeWhere`: errors and returned state remain governed by the function's explicit branches.
- `scopeArgs`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `tx.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `tx.Commit`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 13 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
