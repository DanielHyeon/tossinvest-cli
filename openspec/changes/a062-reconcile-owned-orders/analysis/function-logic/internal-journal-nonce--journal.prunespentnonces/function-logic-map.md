# Function Logic Map: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`
Function: `Journal.PruneSpentNonces`
Signature: `Journal.PruneSpentNonces(params=3, results=2)`
Source SHA-256: `1466fddb8d43a5481cdc10f06b53c09a340862ab91907b7ccc70e40d35b7959c`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.PruneSpentNonces(params=3, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce.go:151 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/nonce.go:155 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/nonce.go:158 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/nonce.go:172 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/nonce.go:176 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `fmt.Errorf`: errors and state follow mapped branches.
- `j.MaxDecisionTTL`: errors and state follow mapped branches.
- `formatJournalTime`: errors and state follow mapped branches.
- `now.Add`: errors and state follow mapped branches.
- `j.db.ExecContext`: errors and state follow mapped branches.
- `res.RowsAffected`: errors and state follow mapped branches.
- `int`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 6; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
