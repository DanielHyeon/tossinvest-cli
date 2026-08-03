# Function Logic Map: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`
Function: `TestStartupPrunesSpentNoncesOnce`
Signature: `TestStartupPrunesSpentNoncesOnce(params=1, results=0)`
Source SHA-256: `dec52090cdbaa17f2a868d8e204c1755d88ec127a3f5158efc858405f41b83b7`
Revision: `base`

## Inputs and invariants

- Inputs are `TestStartupPrunesSpentNoncesOnce(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:451 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/engine_test.go:456 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/engine_test.go:460 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `isolate`: errors and state follow mapped branches.
- `writeEngineConfig`: errors and state follow mapped branches.
- `writeCredentials`: errors and state follow mapped branches.
- `engineStub`: errors and state follow mapped branches.
- `time.Now`: errors and state follow mapped branches.
- `seedSpentNonce`: errors and state follow mapped branches.
- `now.Add`: errors and state follow mapped branches.
- `startEngine`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `eng.Journal.NonceSpent`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 9; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
