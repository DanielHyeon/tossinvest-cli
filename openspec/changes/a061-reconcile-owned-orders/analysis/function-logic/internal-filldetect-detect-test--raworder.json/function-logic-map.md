# Function Logic Map: `rawOrder.json`

Source: `internal/filldetect/detect_test.go`  
Function: `rawOrder.json`  
Signature: `rawOrder.json(params=0, results=1)`  
Source SHA-256: `7fe5825a894d212e278325c39d6b369d975ef46f006b913627daa8c7264e2e26`

## Inputs and invariants

- Inputs are the parameters in `rawOrder.json(params=0, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect_test.go:255 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/detect_test.go:258 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/filldetect/detect_test.go:261 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/filldetect/detect_test.go:264 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/filldetect/detect_test.go:267 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/filldetect/detect_test.go:270 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/filldetect/detect_test.go:282 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/filldetect/detect_test.go:286 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | switch | internal/filldetect/detect_test.go:289 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | case | internal/filldetect/detect_test.go:290 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | case | internal/filldetect/detect_test.go:292 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/filldetect/detect_test.go:295 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `quote`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `orZero`: returned errors and state follow the mapped branches.
- `strings.Join`: returned errors and state follow the mapped branches.
- `json.RawMessage`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 14 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
