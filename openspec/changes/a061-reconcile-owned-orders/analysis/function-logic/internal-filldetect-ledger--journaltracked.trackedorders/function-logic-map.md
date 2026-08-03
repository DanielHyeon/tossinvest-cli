# Function Logic Map: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`
Function: `JournalTracked.TrackedOrders`
Signature: `JournalTracked.TrackedOrders(params=1, results=2)`
Source SHA-256: `75966bf1507d412f48b88efddab42f045fff31f45c76bc9cf43dfc0a634242c4`
Revision: `current`

## Inputs and invariants

- Inputs are `JournalTracked.TrackedOrders(params=1, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:126 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/ledger.go:129 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/filldetect/ledger.go:133 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/filldetect/ledger.go:137 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `fmt.Errorf`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `t.Journal.TrackedFillOrders`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 4; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
