# Function Logic Map: `EntryGate.SetAuthorityRefresh`

Source: `internal/execgw/retry.go`
Function: `EntryGate.SetAuthorityRefresh`
Signature: `EntryGate.SetAuthorityRefresh(params=1, results=0)`
Source SHA-256: `a549135ef2864ab05eb8168cfac899cad5052457189307c4ed3e3bee42e102d3`
Revision: `current`

## Inputs and invariants

- Inputs are `EntryGate.SetAuthorityRefresh(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/execgw/retry.go:451 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- `g.mu.Lock`: errors and state follow mapped branches.
- `g.mu.Unlock`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
