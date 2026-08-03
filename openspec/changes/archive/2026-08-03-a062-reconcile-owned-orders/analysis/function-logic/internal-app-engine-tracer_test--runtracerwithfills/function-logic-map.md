# Function Logic Map: `runTracerWithFills`

Source: `internal/app/engine/tracer_test.go`
Function: `runTracerWithFills`
Signature: `runTracerWithFills(params=4, results=2)`
Source SHA-256: `6d50eb9c3d64746ce4b3430c56a3b714fe00cf852325806b2bdc1cf73014e582`
Revision: `current`

## Inputs and invariants

- Inputs are `runTracerWithFills(params=4, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/tracer_test.go:184 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `t.Helper`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `unknown`: errors and state follow mapped branches.
- `tracer.Run`: errors and state follow mapped branches.
- `driveUntilOrders`: errors and state follow mapped branches.
- `waitForJournaledOrder`: errors and state follow mapped branches.
- `s.fill`: errors and state follow mapped branches.
- `s.broker.quote`: errors and state follow mapped branches.
- `drive`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 5; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
