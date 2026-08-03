# Function Logic Map: `Tracker.Restore`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.Restore`  
Signature: `Tracker.Restore(params=1, results=1)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters represented by `Tracker.Restore(params=1, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:578 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/reconcile/mismatch.go:582 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | range | internal/reconcile/mismatch.go:589 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/reconcile/mismatch.go:590 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/reconcile/mismatch.go:607 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/reconcile/mismatch.go:610 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `t.Journal.ActiveReconcileStates`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `blockFromReconcileState`: errors and returned state remain governed by the function's explicit branches.
- `block.Key`: errors and returned state remain governed by the function's explicit branches.
- `t.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `t.maxFailures`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.RebuildReconcileProjection`: errors and returned state remain governed by the function's explicit branches.
- `t.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 12 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
