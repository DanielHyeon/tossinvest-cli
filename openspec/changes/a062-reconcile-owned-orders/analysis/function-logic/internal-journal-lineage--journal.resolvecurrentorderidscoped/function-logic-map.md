# Function Logic Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`
Function: `Journal.ResolveCurrentOrderIDScoped`
Signature: `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)`
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:313 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/lineage.go:317 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | for | internal/journal/lineage.go:323 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/lineage.go:325 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | switch | internal/journal/lineage.go:328 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | case | internal/journal/lineage.go:329 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | case | internal/journal/lineage.go:331 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | case | internal/journal/lineage.go:333 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/lineage.go:338 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.TrimSpace`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `scope.canonical`: errors and state follow mapped branches.
- `j.scopedLineageChildren`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `j.recordScopedLineageConflict`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 6; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
