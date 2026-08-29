# Function Logic Map: `recordConfirmedFillOrder`

Source: `internal/journal/fills_test.go`
Function: `recordConfirmedFillOrder`
Signature: `recordConfirmedFillOrder(params=5, results=0)`
Source SHA-256: `5da6390852646dcea50c6546b6f28e0c82f15f832d8953cb57aad12da363499a`
Revision: `current`

## Inputs and invariants

- Inputs are `recordConfirmedFillOrder(params=5, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/fills_test.go:105 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- `recordConfirmedFillOrderScope`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
