# Function Logic Map: `fakeBalance.BuyingPower`

Source: `internal/app/engine/reconcileloop_test.go`
Function: `fakeBalance.BuyingPower`
Signature: `fakeBalance.BuyingPower(params=2, results=2)`
Source SHA-256: `f7244f04d716230ddc2536f8e219958c52b86a6b899cf6f4df45fa09962f961e`
Revision: `base`

## Inputs and invariants

- Inputs are `fakeBalance.BuyingPower(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:90 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- No nested call is present; behavior is source-local and source-hash bound.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
