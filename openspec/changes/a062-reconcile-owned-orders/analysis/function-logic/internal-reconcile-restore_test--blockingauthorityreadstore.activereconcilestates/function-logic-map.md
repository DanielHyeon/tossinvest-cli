# Function Logic Map: `blockingAuthorityReadStore.ActiveReconcileStates`

Source: `internal/reconcile/restore_test.go`
Function: `blockingAuthorityReadStore.ActiveReconcileStates`
Signature: `blockingAuthorityReadStore.ActiveReconcileStates(params=1, results=2)`
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`
Revision: `current`

## Inputs and invariants

- Inputs are `blockingAuthorityReadStore.ActiveReconcileStates(params=1, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/reconcile/restore_test.go:51 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- `s.ReconcileStore.ActiveReconcileStates`: errors and state follow mapped branches.
- `close`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
