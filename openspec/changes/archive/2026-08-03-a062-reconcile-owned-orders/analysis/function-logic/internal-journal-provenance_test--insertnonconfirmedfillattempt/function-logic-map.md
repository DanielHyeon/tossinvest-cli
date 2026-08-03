# Function Logic Map: `insertNonConfirmedFillAttempt`

Source: `internal/journal/provenance_test.go`
Function: `insertNonConfirmedFillAttempt`
Signature: `insertNonConfirmedFillAttempt(params=5, results=1)`
Source SHA-256: `42fabaf43a4709cf94fadf4c29b9daeb6e6967779c17161f242ab6184d7e5003`
Revision: `current`

## Inputs and invariants

- Inputs are `insertNonConfirmedFillAttempt(params=5, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/provenance_test.go:131 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/provenance_test.go:139 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `t.Helper`: errors and state follow mapped branches.
- `withDefaults`: errors and state follow mapped branches.
- `j.db.Exec`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
