# Function Logic Map: `LocalStateFromJournal`

Source: `internal/reconcile/compare.go`  
Function: `LocalStateFromJournal`  
Signature: `LocalStateFromJournal(params=3, results=2)`  
Source SHA-256: `3dabc46253e922cfeb82603773013e9b81cc3b5a4fa9086873a48833e58036f7`

## Inputs and invariants

- Inputs are the parameters represented by `LocalStateFromJournal(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:146 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/reconcile/compare.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/reconcile/compare.go:154 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/reconcile/compare.go:163 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/reconcile/compare.go:171 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `HeldBySymbol`: errors and returned state remain governed by the function's explicit branches.
- `j.TrackedFillOrders`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `j.ResolveCurrentOrderIDScoped`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToLower`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
