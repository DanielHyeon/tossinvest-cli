# Function Logic Map: `hasReplaceSuccessor`

Source: `internal/journal/position_projection.go`
Function: `hasReplaceSuccessor`
Signature: `hasReplaceSuccessor(params=4, results=2)`
Source SHA-256: `5b62d380b860e67884ceb5f5e217a12fa0d8a190c54190f370ae711133002748`
Revision: `current`

## Inputs and invariants

- Inputs are `hasReplaceSuccessor(params=4, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection.go:244 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/position_projection.go:249 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `tx.Query`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 3; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
