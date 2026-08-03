# Function Logic Map: `LocalStateFromJournal`

Source: `internal/reconcile/compare.go`
Function: `LocalStateFromJournal`
Signature: `LocalStateFromJournal(params=3, results=2)`
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`
Revision: `current`

## Inputs and invariants

- Inputs are `LocalStateFromJournal(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:199 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/reconcile/compare.go:203 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/reconcile/compare.go:207 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/reconcile/compare.go:217 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/reconcile/compare.go:225 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `fmt.Errorf`: errors and state follow mapped branches.
- `HeldBySymbol`: errors and state follow mapped branches.
- `j.TrackedFillOrders`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `j.ResolveCurrentOrderIDScoped`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `strings.ToLower`: errors and state follow mapped branches.
- `order.Identity`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 5; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
