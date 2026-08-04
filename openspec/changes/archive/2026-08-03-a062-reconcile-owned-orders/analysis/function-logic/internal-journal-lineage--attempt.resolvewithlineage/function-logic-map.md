# Function Logic Map: `Attempt.resolveWithLineage`

Source: `internal/journal/lineage.go`
Function: `Attempt.resolveWithLineage`
Signature: `Attempt.resolveWithLineage(params=4, results=1)`
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`
Revision: `current`

## Inputs and invariants

- Inputs are `Attempt.resolveWithLineage(params=4, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:133 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/lineage.go:139 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/lineage.go:141 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/lineage.go:146 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/lineage.go:162 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/lineage.go:166 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/lineage.go:169 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/lineage.go:173 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/lineage.go:180 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/lineage.go:216 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/lineage.go:221 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/lineage.go:225 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/lineage.go:230 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `a.mu.Lock`: errors and state follow mapped branches.
- `a.mu.Unlock`: errors and state follow mapped branches.
- `a.j.nowString`: errors and state follow mapped branches.
- `a.j.db.BeginTx`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `tx.Rollback`: errors and state follow mapped branches.
- `Scan`: errors and state follow mapped branches.
- `tx.QueryRowContext`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `checkTransitionAllowed`: errors and state follow mapped branches.
- `tx.ExecContext`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `res.RowsAffected`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `scoped.RowsAffected`: errors and state follow mapped branches.
- `tx.Commit`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 13; return points: 14; deferred operations: 2.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
