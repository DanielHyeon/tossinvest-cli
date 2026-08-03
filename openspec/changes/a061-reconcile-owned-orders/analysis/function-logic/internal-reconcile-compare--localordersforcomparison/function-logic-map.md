# Function Logic Map: `localOrdersForComparison`

Source: `internal/reconcile/compare.go`
Function: `localOrdersForComparison`
Signature: `localOrdersForComparison(params=1, results=1)`
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`
Revision: `current`

## Inputs and invariants

- Inputs are `localOrdersForComparison(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:480 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | range | internal/reconcile/compare.go:482 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/reconcile/compare.go:497 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/reconcile/compare.go:498 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `identity.canonical`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `sort.Slice`: errors and state follow mapped branches.
- `less`: errors and state follow mapped branches.
- `orders.Identity`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 12; return points: 4; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
