# Function Logic Map: `TestReusedOrderIDIsOwnedByTheObservedTradingDay`

Source: `internal/journal/position_projection_test.go`
Function: `TestReusedOrderIDIsOwnedByTheObservedTradingDay`
Signature: `TestReusedOrderIDIsOwnedByTheObservedTradingDay(params=1, results=0)`
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`
Revision: `current`

## Inputs and invariants

- Inputs are `TestReusedOrderIDIsOwnedByTheObservedTradingDay(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:217 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/position_projection_test.go:221 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/position_projection_test.go:225 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/position_projection_test.go:228 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `projectingJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `fillOf`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `j.ActiveReconcileStates`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
