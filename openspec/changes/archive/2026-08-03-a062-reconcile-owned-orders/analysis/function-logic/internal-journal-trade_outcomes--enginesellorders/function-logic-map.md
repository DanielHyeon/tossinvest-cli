# Function Logic Map: `engineSellOrders`

Source: `internal/journal/trade_outcomes.go`
Function: `engineSellOrders`
Signature: `engineSellOrders(params=6, results=2)`
Source SHA-256: `0bd43abf96107b9998ea7c9bc6c6655f162b1af3ae96490daadc78fb354a4958`
Revision: `current`

## Inputs and invariants

- Inputs are `engineSellOrders(params=6, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/trade_outcomes.go:592 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | for | internal/journal/trade_outcomes.go:598 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/trade_outcomes.go:600 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/trade_outcomes.go:606 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `r.Query`: errors and state follow mapped branches.
- `normaliseSymbol`: errors and state follow mapped branches.
- `normaliseMarket`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 4; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
