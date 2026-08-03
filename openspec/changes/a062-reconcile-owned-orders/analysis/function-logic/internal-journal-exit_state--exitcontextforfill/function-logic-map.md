# Function Logic Map: `exitContextForFill`

Source: `internal/journal/exit_state.go`
Function: `exitContextForFill`
Signature: `exitContextForFill(params=3, results=3)`
Source SHA-256: `f3895fb41abc09f4de2aad1eceeeff1b39ab17ed658b2dc74e02bf7727b46f86`
Revision: `current`

## Inputs and invariants

- Inputs are `exitContextForFill(params=3, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/exit_state.go:875 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/exit_state.go:878 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/exit_state.go:890 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/exit_state.go:895 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/exit_state.go:898 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `resolveFillOrigin`: errors and state follow mapped branches.
- `tx.Query`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `instance.Close`: errors and state follow mapped branches.
- `instance.Next`: errors and state follow mapped branches.
- `instance.Err`: errors and state follow mapped branches.
- `instance.Scan`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 6; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
