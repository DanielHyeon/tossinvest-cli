# Function Logic Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`  
Function: `Journal.ReleaseReconcile`  
Signature: `Journal.ReleaseReconcile(params=2, results=3)`  
Source SHA-256: `f07e1a91c10a72e1226e5cf5328d461def19b571714145d31ccb838c2e402e19`

## Inputs and invariants

- Inputs are the parameters in `Journal.ReleaseReconcile(params=2, results=3)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/journal/reconcile_states.go:251 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | case | internal/journal/reconcile_states.go:252 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | case | internal/journal/reconcile_states.go:255 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/journal/reconcile_states.go:258 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/reconcile_states.go:267 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/reconcile_states.go:274 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/reconcile_states.go:277 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/reconcile_states.go:280 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/reconcile_states.go:284 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/reconcile_states.go:290 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `ValidReconcileReleaseCause`: returned errors and state follow the mapped branches.
- `UTC`: returned errors and state follow the mapped branches.
- `j.clk.Now`: returned errors and state follow the mapped branches.
- `formatJournalTime`: returned errors and state follow the mapped branches.
- `j.db.BeginTx`: returned errors and state follow the mapped branches.
- `tx.Rollback`: returned errors and state follow the mapped branches.
- `scanReconcileState`: returned errors and state follow the mapped branches.
- `tx.QueryRowContext`: returned errors and state follow the mapped branches.
- `activeScopeWhere`: returned errors and state follow the mapped branches.
- `scopeArgs`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `tx.ExecContext`: returned errors and state follow the mapped branches.
- `tx.Commit`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 13 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
