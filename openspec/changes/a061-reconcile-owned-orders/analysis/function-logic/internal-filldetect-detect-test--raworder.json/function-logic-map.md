# Function Logic Map: `rawOrder.json`

Source: `internal/filldetect/detect_test.go`  
Function: `rawOrder.json`  
Signature: `rawOrder.json(params=0, results=1)`  
Source SHA-256: `ebc9d90930d66137b55d15c99aee9d721ccf01eaa5d516b3b00755ae963be6fc`

## Inputs and invariants

- Inputs are the parameters represented by `rawOrder.json(params=0, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect_test.go:227 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/detect_test.go:230 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/detect_test.go:233 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/detect_test.go:236 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/detect_test.go:248 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/filldetect/detect_test.go:252 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | switch | internal/filldetect/detect_test.go:255 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | case | internal/filldetect/detect_test.go:256 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | case | internal/filldetect/detect_test.go:258 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/filldetect/detect_test.go:261 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `quote`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `orZero`: errors and returned state remain governed by the function's explicit branches.
- `strings.Join`: errors and returned state remain governed by the function's explicit branches.
- `json.RawMessage`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 12 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
