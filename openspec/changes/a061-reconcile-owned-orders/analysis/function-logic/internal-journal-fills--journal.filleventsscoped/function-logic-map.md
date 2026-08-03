# Function Logic Map: `Journal.FillEventsScoped`

Source: `internal/journal/fills.go`
Function: `Journal.FillEventsScoped`
Signature: `Journal.FillEventsScoped(params=2, results=2)`
Source SHA-256: `000918b94c8c3f776b611421c412e4604086fc4cbee2fd0e7c21fe0dd46454c0`
Revision: `current`

## Inputs and invariants

- Inputs are `Journal.FillEventsScoped(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:1096 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/fills.go:1118 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `canonicalFillSnapshotScope`: errors and state follow mapped branches.
- `scope.complete`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `j.db.QueryContext`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `scanFillEvents`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 2; return points: 3; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
