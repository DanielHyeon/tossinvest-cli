# Function Logic Map: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`
Function: `TestSchemaIndexes`
Signature: `TestSchemaIndexes(params=1, results=0)`
Source SHA-256: `50b6ce71056a4e67c8dbb344ab84bd429c6490b1ebc7f25fe7bc7d15e7c7c1be`
Revision: `current`

## Inputs and invariants

- Inputs are `TestSchemaIndexes(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/schema_test.go:526 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | for | internal/journal/schema_test.go:531 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/schema_test.go:533 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/schema_test.go:538 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | range | internal/journal/schema_test.go:541 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/schema_test.go:575 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `openTestJournal`: errors and state follow mapped branches.
- `j.db.QueryContext`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- `keysOf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 6; return points: 0; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
