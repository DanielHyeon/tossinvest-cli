# Function Logic Map: `isReconcileReason`

Source: `internal/reconcile/mismatch.go`
Function: `isReconcileReason`
Signature: `isReconcileReason(params=1, results=1)`
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`
Revision: `current`

## Inputs and invariants

- Inputs are `isReconcileReason(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/reconcile/mismatch.go:929 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- No nested call is present; behavior is source-local and source-hash bound.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
