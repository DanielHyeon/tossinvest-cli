# Function Logic Map: `engineFillDetector`

Source: `cmd/tossctl/engine.go`
Function: `engineFillDetector`
Signature: `engineFillDetector(params=3, results=2)`
Source SHA-256: `45414562be8a352d2183fb2dfc0985154e0eea5ce781e167eb6800841c495451`
Revision: `current`

## Inputs and invariants

- Inputs are `engineFillDetector(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | cmd/tossctl/engine.go:401 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | cmd/tossctl/engine.go:402 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `hints.Validate`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `ectx.AccountSweep`: errors and state follow mapped branches.
- `ectx.SnapshotCurrencies`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 2; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
