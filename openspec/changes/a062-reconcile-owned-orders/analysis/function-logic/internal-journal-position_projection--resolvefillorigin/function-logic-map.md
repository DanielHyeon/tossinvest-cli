# Function Logic Map: `resolveFillOrigin`

Source: `internal/journal/position_projection.go`
Function: `resolveFillOrigin`
Signature: `resolveFillOrigin(params=3, results=3)`
Source SHA-256: `5b62d380b860e67884ceb5f5e217a12fa0d8a190c54190f370ae711133002748`
Revision: `current`

## Inputs and invariants

- Inputs are `resolveFillOrigin(params=3, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection.go:142 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | for | internal/journal/position_projection.go:154 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/position_projection.go:156 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/position_projection.go:163 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/position_projection.go:168 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | range | internal/journal/position_projection.go:182 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/position_projection.go:183 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/position_projection.go:191 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/position_projection.go:194 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | range | internal/journal/position_projection.go:196 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/position_projection.go:200 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/position_projection.go:204 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/position_projection.go:217 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/journal/position_projection.go:226 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `tx.Query`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `firstNonEmpty`: errors and state follow mapped branches.
- `normaliseMarket`: errors and state follow mapped branches.
- `normaliseSymbol`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `enterReconcileScopeInTx`: errors and state follow mapped branches.
- `position.RoleForSide`: errors and state follow mapped branches.
- `hasReplaceSuccessor`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 26; return points: 9; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
