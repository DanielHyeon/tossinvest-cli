# Function Logic Map: `sameSnapshot`

Source: `internal/journal/fills.go`  
Function: `sameSnapshot`  
Signature: `sameSnapshot(params=2, results=1)`  
Source SHA-256: `1a9973b325d8be62dd5d0cdebe10988ac90c6e2114d5f2e1f0b545482b141a65`

## Inputs and invariants

- Inputs are the parameters represented by `sameSnapshot(params=2, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/fills.go:552 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `orZero`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `strings.EqualFold`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- No assignment point is present; the function is observational or delegates its effect.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
