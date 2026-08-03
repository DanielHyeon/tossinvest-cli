# Function Logic Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`
Function: `NewReconcileDriver`
Signature: `NewReconcileDriver(params=1, results=2)`
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`
Revision: `current`

## Inputs and invariants

- Inputs are `NewReconcileDriver(params=1, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/app/engine/reconcileloop.go:280 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | case | internal/app/engine/reconcileloop.go:281 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | case | internal/app/engine/reconcileloop.go:283 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | case | internal/app/engine/reconcileloop.go:285 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | case | internal/app/engine/reconcileloop.go:287 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | case | internal/app/engine/reconcileloop.go:289 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | case | internal/app/engine/reconcileloop.go:292 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/app/engine/reconcileloop.go:296 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/app/engine/reconcileloop.go:297 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/app/engine/reconcileloop.go:309 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `fmt.Errorf`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `exitpolicy.CommonPolicyByID`: errors and state follow mapped branches.
- `clock.System`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 8; return points: 8; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
