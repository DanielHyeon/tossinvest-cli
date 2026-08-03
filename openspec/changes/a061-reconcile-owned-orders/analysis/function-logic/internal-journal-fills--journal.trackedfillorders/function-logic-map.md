# Function Logic Map: `Journal.TrackedFillOrders`

Source: `internal/journal/fills.go`  
Function: `Journal.TrackedFillOrders`  
Signature: `Journal.TrackedFillOrders(params=2, results=2)`  
Source SHA-256: `1a9973b325d8be62dd5d0cdebe10988ac90c6e2114d5f2e1f0b545482b141a65`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.TrackedFillOrders(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:840 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/fills.go:843 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/fills.go:844 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/fills.go:953 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | for | internal/journal/fills.go:959 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/fills.go:961 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/fills.go:967 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | range | internal/journal/fills.go:975 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/fills.go:983 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/fills.go:986 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `len`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `j.guardTrackedFillIdentity`: errors and returned state remain governed by the function's explicit branches.
- `j.db.QueryContext`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `rows.Close`: errors and returned state remain governed by the function's explicit branches.
- `rows.Next`: errors and returned state remain governed by the function's explicit branches.
- `rows.Scan`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `rows.Err`: errors and returned state remain governed by the function's explicit branches.
- `j.ResolveCurrentOrderIDScoped`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 9 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
