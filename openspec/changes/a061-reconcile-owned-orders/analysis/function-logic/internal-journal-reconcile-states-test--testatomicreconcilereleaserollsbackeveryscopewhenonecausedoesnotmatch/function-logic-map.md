# Function Logic Map: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`  
Function: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`  
Signature: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch(params=1, results=0)`  
Source SHA-256: `d2de10c4ae8c4d15346e190fa03cbc7bd4db7648bd7b4ab102272225dd1785a6`

## Inputs and invariants

- Inputs are the parameters represented by `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/journal/reconcile_states_test.go:296 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/reconcile_states_test.go:300 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/reconcile_states_test.go:310 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/reconcile_states_test.go:314 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/reconcile_states_test.go:317 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | range | internal/journal/reconcile_states_test.go:320 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/reconcile_states_test.go:321 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openReservationJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `j.EnterReconcile`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `j.ReleaseReconciles`: errors and returned state remain governed by the function's explicit branches.
- `strings.Contains`: errors and returned state remain governed by the function's explicit branches.
- `err.Error`: errors and returned state remain governed by the function's explicit branches.
- `j.ActiveReconcileStates`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `state.ReleasedAt.IsZero`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
