# Function Logic Map: `Journal.scopedLineageChildren`

Source: `internal/journal/lineage.go`
Function: `Journal.scopedLineageChildren`
Signature: `Journal.scopedLineageChildren(params=3, results=2)`
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.scopedLineageChildren(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:348 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/lineage.go:352 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/journal/lineage.go:362 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/lineage.go:363 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/lineage.go:368 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `j.scopedLineageChildrenV16`: errors and state follow mapped branches.
- `j.legacyScopedLineageChildren`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 6; return points: 3; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
