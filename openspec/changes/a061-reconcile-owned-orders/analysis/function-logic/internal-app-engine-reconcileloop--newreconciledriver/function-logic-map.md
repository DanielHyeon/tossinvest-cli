# Function Logic Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`  
Function: `NewReconcileDriver`  
Signature: `NewReconcileDriver(params=1, results=2)`  
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`

## Inputs and invariants

- Inputs are the parameters represented by `NewReconcileDriver(params=1, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/app/engine/reconcileloop.go:280 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | case | internal/app/engine/reconcileloop.go:281 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | case | internal/app/engine/reconcileloop.go:283 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | case | internal/app/engine/reconcileloop.go:285 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | case | internal/app/engine/reconcileloop.go:287 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | case | internal/app/engine/reconcileloop.go:289 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | case | internal/app/engine/reconcileloop.go:292 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/app/engine/reconcileloop.go:296 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/app/engine/reconcileloop.go:297 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/app/engine/reconcileloop.go:309 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `exitpolicy.CommonPolicyByID`: errors and returned state remain governed by the function's explicit branches.
- `clock.System`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 8 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
