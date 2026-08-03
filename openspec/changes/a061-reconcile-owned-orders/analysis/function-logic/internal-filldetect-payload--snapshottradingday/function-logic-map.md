# Function Logic Map: `snapshotTradingDay`

Source: `internal/filldetect/payload.go`  
Function: `snapshotTradingDay`  
Signature: `snapshotTradingDay(params=2, results=1)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters represented by `snapshotTradingDay(params=2, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload.go:146 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/payload.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | range | internal/filldetect/payload.go:153 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/payload.go:154 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/payload.go:156 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `trimPtr`: errors and returned state remain governed by the function's explicit branches.
- `clock.ParseMarket`: errors and returned state remain governed by the function's explicit branches.
- `time.Parse`: errors and returned state remain governed by the function's explicit branches.
- `market.TradingDay`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 4 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
