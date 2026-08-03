# Function Logic Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`
Function: `Journal.ResolveCurrentOrderID`
Signature: `Journal.ResolveCurrentOrderID(params=2, results=2)`
Source SHA-256: `bf26e9cfd6030033e99ec6ee2ceb53dd5843a0c4c25111e8761844c598bb9d73`
Revision: `base`

## Inputs and invariants

- Inputs are `Journal.ResolveCurrentOrderID(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:200 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | for | internal/journal/lineage.go:204 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/lineage.go:206 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/journal/lineage.go:210 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/lineage.go:211 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/lineage.go:216 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/lineage.go:219 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.TrimSpace`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `j.LineageChildren`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 6; return points: 4; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
