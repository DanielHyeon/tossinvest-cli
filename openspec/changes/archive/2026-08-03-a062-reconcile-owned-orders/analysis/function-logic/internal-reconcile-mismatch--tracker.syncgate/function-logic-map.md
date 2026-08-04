# Function Logic Map: `Tracker.syncGate`

Source: `internal/reconcile/mismatch.go`
Function: `Tracker.syncGate`
Signature: `Tracker.syncGate(params=1, results=0)`
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`
Revision: `current`

## Inputs and invariants

- Inputs are `Tracker.syncGate(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:879 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | range | internal/reconcile/mismatch.go:883 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/reconcile/mismatch.go:884 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/reconcile/mismatch.go:894 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/reconcile/mismatch.go:895 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | range | internal/reconcile/mismatch.go:901 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/reconcile/mismatch.go:902 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/reconcile/mismatch.go:908 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | range | internal/reconcile/mismatch.go:913 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/reconcile/mismatch.go:917 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | else | internal/reconcile/mismatch.go:919 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `t.Gate.BlockSymbol`: errors and state follow mapped branches.
- `t.Gate.SymbolBlocks`: errors and state follow mapped branches.
- `isReconcileReason`: errors and state follow mapped branches.
- `t.Gate.ClearSymbol`: errors and state follow mapped branches.
- `t.Gate.Block`: errors and state follow mapped branches.
- `t.Gate.Clear`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 5; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
