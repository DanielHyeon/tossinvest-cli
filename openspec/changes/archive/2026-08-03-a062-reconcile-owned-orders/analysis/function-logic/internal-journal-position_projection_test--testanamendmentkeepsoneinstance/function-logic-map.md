# Function Logic Map: `TestAnAmendmentKeepsOneInstance`

Source: `internal/journal/position_projection_test.go`
Function: `TestAnAmendmentKeepsOneInstance`
Signature: `TestAnAmendmentKeepsOneInstance(params=1, results=0)`
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`
Revision: `current`

## Inputs and invariants

- Inputs are `TestAnAmendmentKeepsOneInstance(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:334 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/position_projection_test.go:346 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/position_projection_test.go:349 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/position_projection_test.go:352 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/position_projection_test.go:355 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/position_projection_test.go:365 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/position_projection_test.go:368 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/position_projection_test.go:374 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/position_projection_test.go:379 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/position_projection_test.go:382 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/position_projection_test.go:385 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/position_projection_test.go:388 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `projectingJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `fillOf`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `j.Prepare`: errors and state follow mapped branches.
- `testIntentFor`: errors and state follow mapped branches.
- `amend.MarkDispatchStarted`: errors and state follow mapped branches.
- `amend.MarkAcked`: errors and state follow mapped branches.
- `amend.ResolveConfirmedWithLineage`: errors and state follow mapped branches.
- `terminalFill`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `j.Positions`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 14; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
