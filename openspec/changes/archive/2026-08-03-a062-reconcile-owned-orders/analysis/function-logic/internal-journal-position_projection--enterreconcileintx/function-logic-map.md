# Function Logic Map: `enterReconcileInTx`

Source: `internal/journal/position_projection.go`
Function: `enterReconcileInTx`
Signature: `enterReconcileInTx(params=5, results=1)`
Source SHA-256: `5b62d380b860e67884ceb5f5e217a12fa0d8a190c54190f370ae711133002748`
Revision: `current`

## Inputs and invariants

- Inputs are `enterReconcileInTx(params=5, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/position_projection.go:370 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- `reconcileCauseFor`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `enterReconcileScopeInTx`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 2; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
