# Function Logic Map: `Detector.collect`

Source: `internal/filldetect/detect.go`  
Function: `Detector.collect`  
Signature: `Detector.collect(params=4, results=2)`  
Source SHA-256: `b8e7aeaf3f2b1823793934701f01fe04bb47bc8c7ecdf300b5b41111ae88ebe9`

## Inputs and invariants

- Inputs are the parameters represented by `Detector.collect(params=4, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect.go:357 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/detect.go:367 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | range | internal/filldetect/detect.go:372 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/filldetect/detect.go:377 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/detect.go:379 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/filldetect/detect.go:385 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/filldetect/detect.go:386 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | range | internal/filldetect/detect.go:394 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/filldetect/detect.go:395 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/filldetect/detect.go:399 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/filldetect/detect.go:404 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/filldetect/detect.go:410 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | if | internal/filldetect/detect.go:413 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B14 | if | internal/filldetect/detect.go:416 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `execgw.ScanOrders`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `d.brokerVisibleFallback`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `d.Tracked.TrackedOrders`: errors and returned state remain governed by the function's explicit branches.
- `parseSnapshot`: errors and returned state remain governed by the function's explicit branches.
- `brokerstate.Derive`: errors and returned state remain governed by the function's explicit branches.
- `viewOf`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `d.Order.OrderRaw`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 24 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
