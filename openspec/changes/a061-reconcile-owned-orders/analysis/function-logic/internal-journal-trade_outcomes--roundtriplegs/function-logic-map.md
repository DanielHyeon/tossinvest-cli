# Function Logic Map: `roundTripLegs`

Source: `internal/journal/trade_outcomes.go`
Function: `roundTripLegs`
Signature: `roundTripLegs(params=6, results=3)`
Source SHA-256: `0bd43abf96107b9998ea7c9bc6c6655f162b1af3ae96490daadc78fb354a4958`
Revision: `current`

## Inputs and invariants

- Inputs are `roundTripLegs(params=6, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/trade_outcomes.go:324 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/trade_outcomes.go:327 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/trade_outcomes.go:331 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/trade_outcomes.go:368 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | for | internal/journal/trade_outcomes.go:377 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/trade_outcomes.go:382 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/trade_outcomes.go:388 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/trade_outcomes.go:393 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/trade_outcomes.go:397 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/trade_outcomes.go:400 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/trade_outcomes.go:408 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/trade_outcomes.go:415 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/trade_outcomes.go:422 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/journal/trade_outcomes.go:425 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `new`: errors and state follow mapped branches.
- `r.Query`: errors and state follow mapped branches.
- `bound.Next`: errors and state follow mapped branches.
- `bound.Scan`: errors and state follow mapped branches.
- `bound.Close`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `normaliseSymbol`: errors and state follow mapped branches.
- `normaliseMarket`: errors and state follow mapped branches.
- `rows.Close`: errors and state follow mapped branches.
- `rows.Next`: errors and state follow mapped branches.
- `rows.Scan`: errors and state follow mapped branches.
- `strings.Join`: errors and state follow mapped branches.
- `strings.EqualFold`: errors and state follow mapped branches.
- `SetString`: errors and state follow mapped branches.
- `orZero`: errors and state follow mapped branches.
- `Mul`: errors and state follow mapped branches.
- `leg.notional.Add`: errors and state follow mapped branches.
- `Sub`: errors and state follow mapped branches.
- `rows.Err`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 18; return points: 9; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
