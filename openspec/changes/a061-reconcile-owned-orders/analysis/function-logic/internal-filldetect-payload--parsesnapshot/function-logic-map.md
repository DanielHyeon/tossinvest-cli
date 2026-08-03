# Function Logic Map: `parseSnapshot`

Source: `internal/filldetect/payload.go`  
Function: `parseSnapshot`  
Signature: `parseSnapshot(params=3, results=2)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters represented by `parseSnapshot(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload.go:73 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/payload.go:78 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/payload.go:87 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/payload.go:88 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/payload.go:95 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/filldetect/payload.go:119 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/filldetect/payload.go:123 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `unwrapResult`: errors and returned state remain governed by the function's explicit branches.
- `json.Unmarshal`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `parseDecimal`: errors and returned state remain governed by the function's explicit branches.
- `trimPtr`: errors and returned state remain governed by the function's explicit branches.
- `parseTime`: errors and returned state remain governed by the function's explicit branches.
- `fillMarketFromCurrency`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `snapshotTradingDay`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 15 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
