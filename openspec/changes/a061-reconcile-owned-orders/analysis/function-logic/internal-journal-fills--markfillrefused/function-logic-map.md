# Function Logic Map: `markFillRefused`

Source: `internal/journal/fills.go`  
Function: `markFillRefused`  
Signature: `markFillRefused(params=7, results=1)`  
Source SHA-256: `1a9973b325d8be62dd5d0cdebe10988ac90c6e2114d5f2e1f0b545482b141a65`

## Inputs and invariants

- Inputs are the parameters represented by `markFillRefused(params=7, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:524 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/fills.go:525 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/fills.go:534 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `tx.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `orZero`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 2 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
