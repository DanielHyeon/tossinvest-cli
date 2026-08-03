# Function Logic Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`
Function: `Journal.ReleaseReconcile`
Signature: `Journal.ReleaseReconcile(params=2, results=3)`
Source SHA-256: `f07e1a91c10a72e1226e5cf5328d461def19b571714145d31ccb838c2e402e19`
Revision: `base`

## Inputs and invariants

- Inputs are `Journal.ReleaseReconcile(params=2, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/journal/reconcile_states.go:251 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | case | internal/journal/reconcile_states.go:252 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | case | internal/journal/reconcile_states.go:255 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | case | internal/journal/reconcile_states.go:258 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/reconcile_states.go:267 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/reconcile_states.go:274 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/reconcile_states.go:277 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/reconcile_states.go:280 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/reconcile_states.go:284 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/reconcile_states.go:290 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.TrimSpace`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `ValidReconcileReleaseCause`: errors and state follow mapped branches.
- `UTC`: errors and state follow mapped branches.
- `j.clk.Now`: errors and state follow mapped branches.
- `formatJournalTime`: errors and state follow mapped branches.
- `j.db.BeginTx`: errors and state follow mapped branches.
- `tx.Rollback`: errors and state follow mapped branches.
- `scanReconcileState`: errors and state follow mapped branches.
- `tx.QueryRowContext`: errors and state follow mapped branches.
- `activeScopeWhere`: errors and state follow mapped branches.
- `scopeArgs`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `tx.ExecContext`: errors and state follow mapped branches.
- `tx.Commit`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 13; return points: 10; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
