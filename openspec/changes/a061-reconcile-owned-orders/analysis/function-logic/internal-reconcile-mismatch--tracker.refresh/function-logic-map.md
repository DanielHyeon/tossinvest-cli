# Function Logic Map: `Tracker.Refresh`

Source: `internal/reconcile/mismatch.go`
Function: `Tracker.Refresh`
Signature: `Tracker.Refresh(params=1, results=1)`
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`
Revision: `current`

## Inputs and invariants

- Inputs are `Tracker.Refresh(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:623 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/reconcile/mismatch.go:631 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/reconcile/mismatch.go:637 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/reconcile/mismatch.go:638 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | range | internal/reconcile/mismatch.go:642 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/reconcile/mismatch.go:643 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/reconcile/mismatch.go:651 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `t.mu.Lock`: errors and state follow mapped branches.
- `t.Journal.ActiveReconcileStates`: errors and state follow mapped branches.
- `t.mu.Unlock`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `blockFromReconcileState`: errors and state follow mapped branches.
- `block.Key`: errors and state follow mapped branches.
- `hasPermanentQuantityAccountBlock`: errors and state follow mapped branches.
- `t.maxFailures`: errors and state follow mapped branches.
- `t.snapshotBlocks`: errors and state follow mapped branches.
- `t.syncGate`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 9; return points: 3; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
