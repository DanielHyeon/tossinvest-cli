# Function Logic Map: `sequencedReleaseStore.ReleaseReconcile`

Source: `internal/reconcile/restore_test.go`
Function: `sequencedReleaseStore.ReleaseReconcile`
Signature: `sequencedReleaseStore.ReleaseReconcile(params=2, results=3)`
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`
Revision: `current`

## Inputs and invariants

- Inputs are `sequencedReleaseStore.ReleaseReconcile(params=2, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:67 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `errors.New`: errors and state follow mapped branches.
- `s.ReconcileStore.ReleaseReconcile`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
