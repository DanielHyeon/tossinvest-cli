# Function Logic Map: `containsCause`

Source: `internal/journal/reservation_sweep_test.go`  
Function: `containsCause`  
Signature: `containsCause(params=2, results=1)`  
Source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`

## Inputs and invariants

- Inputs are the parameters represented by `containsCause(params=2, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/journal/reservation_sweep_test.go:111 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/reservation_sweep_test.go:112 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- No outbound call is present; behavior is local and deterministic.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- No assignment point is present; the function is observational or delegates its effect.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
