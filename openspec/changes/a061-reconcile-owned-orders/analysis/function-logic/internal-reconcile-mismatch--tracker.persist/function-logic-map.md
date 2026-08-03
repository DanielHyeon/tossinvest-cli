# Function Logic Map: `Tracker.persist`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.persist`  
Signature: `Tracker.persist(params=2, results=2)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters represented by `Tracker.persist(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:514 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | range | internal/reconcile/mismatch.go:524 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/reconcile/mismatch.go:532 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/reconcile/mismatch.go:536 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | range | internal/reconcile/mismatch.go:545 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/reconcile/mismatch.go:553 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/reconcile/mismatch.go:557 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `append`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `firstNonEmpty`: errors and returned state remain governed by the function's explicit branches.
- `t.Journal.EnterReconcile`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `scopeLabel`: errors and returned state remain governed by the function's explicit branches.
- `blockFromReconcileState`: errors and returned state remain governed by the function's explicit branches.
- `t.Journal.ReleaseReconcile`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 8 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
