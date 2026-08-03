# Function Logic Map: `fillMarketFromCurrency`

Source: `internal/filldetect/payload.go`  
Function: `fillMarketFromCurrency`  
Signature: `fillMarketFromCurrency(params=1, results=1)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters represented by `fillMarketFromCurrency(params=1, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/filldetect/payload.go:134 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | case | internal/filldetect/payload.go:135 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | case | internal/filldetect/payload.go:137 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | case | internal/filldetect/payload.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- No assignment point is present; the function is observational or delegates its effect.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
