# Function Logic Map: `buildGateway`

Source: `internal/app/engine/gateway.go`
Function: `buildGateway`
Signature: `buildGateway(params=2, results=2)`
Source SHA-256: `3dead101adcc3b89767975b14f72de7246909ac0ef3f909e3928ebed2637ee8b`
Revision: `current`

## Inputs and invariants

- Inputs are `buildGateway(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/gateway.go:201 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/gateway.go:223 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/gateway.go:260 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `checkProjectionWired`: errors and state follow mapped branches.
- `execgw.NewEntryGate`: errors and state follow mapped branches.
- `tracker.Restore`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `entry.SetAuthorityRefresh`: errors and state follow mapped branches.
- `tracker.Refresh`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `execgw.New`: errors and state follow mapped branches.
- `in.official.BaseURL`: errors and state follow mapped branches.
- `newNotifier`: errors and state follow mapped branches.
- `newRetrier`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 13; return points: 5; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
