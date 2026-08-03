# Function Logic Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`  
Function: `Journal.ResolveCurrentOrderID`  
Signature: `Journal.ResolveCurrentOrderID(params=2, results=2)`  
Source SHA-256: `bf26e9cfd6030033e99ec6ee2ceb53dd5843a0c4c25111e8761844c598bb9d73`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.ResolveCurrentOrderID(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:200 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | for | internal/journal/lineage.go:204 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/lineage.go:206 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/journal/lineage.go:210 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/lineage.go:211 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/lineage.go:216 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/lineage.go:219 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `j.LineageChildren`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 6 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
