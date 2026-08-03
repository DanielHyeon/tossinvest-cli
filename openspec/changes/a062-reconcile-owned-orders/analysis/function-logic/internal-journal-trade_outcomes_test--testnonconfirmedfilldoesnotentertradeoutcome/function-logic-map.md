# Function Logic Map: `TestNonConfirmedFillDoesNotEnterTradeOutcome`

Source: `internal/journal/trade_outcomes_test.go`
Function: `TestNonConfirmedFillDoesNotEnterTradeOutcome`
Signature: `TestNonConfirmedFillDoesNotEnterTradeOutcome(params=1, results=0)`
Source SHA-256: `1723845dbc5c11be31276e125b182a5f02a9401abb17356649aca0f50858ada2`
Revision: `current`

## Inputs and invariants

- Inputs are `TestNonConfirmedFillDoesNotEnterTradeOutcome(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/trade_outcomes_test.go:113 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/trade_outcomes_test.go:117 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/trade_outcomes_test.go:123 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/trade_outcomes_test.go:130 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/trade_outcomes_test.go:137 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `outcomeFixture`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `withFailed.RecordFill`: errors and state follow mapped branches.
- `terminalFill`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `withFailed.OpenExitState`: errors and state follow mapped branches.
- `insertNonConfirmedFillAttempt`: errors and state follow mapped branches.
- `roundTrip`: errors and state follow mapped branches.
- `outcomeOf`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 13; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
