# Function Logic Map: `snapshotTradingDay`

Source: `internal/filldetect/payload.go`
Function: `snapshotTradingDay`
Signature: `snapshotTradingDay(params=2, results=1)`
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`
Revision: `current`

## Inputs and invariants

- Inputs are `snapshotTradingDay(params=2, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload.go:146 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/payload.go:150 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/filldetect/payload.go:153 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/filldetect/payload.go:154 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/filldetect/payload.go:156 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `trimPtr`: errors and state follow mapped branches.
- `clock.ParseMarket`: errors and state follow mapped branches.
- `time.Parse`: errors and state follow mapped branches.
- `market.TradingDay`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 5; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
