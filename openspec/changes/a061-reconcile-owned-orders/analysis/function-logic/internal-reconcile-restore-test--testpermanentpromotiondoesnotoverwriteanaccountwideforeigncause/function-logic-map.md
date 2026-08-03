# Function Logic Map: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`

Source: `internal/reconcile/restore_test.go`  
Function: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`  
Signature: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause(params=1, results=0)`  
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

## Inputs and invariants

- Inputs are the parameters represented by `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:565 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/reconcile/restore_test.go:572 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | for | internal/reconcile/restore_test.go:576 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/reconcile/restore_test.go:577 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | range | internal/reconcile/restore_test.go:583 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/reconcile/restore_test.go:584 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/reconcile/restore_test.go:589 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/reconcile/restore_test.go:592 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `clock.NewFake`: errors and returned state remain governed by the function's explicit branches.
- `openJournal`: errors and returned state remain governed by the function's explicit branches.
- `j.EnterReconcile`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `trackerOn`: errors and returned state remain governed by the function's explicit branches.
- `noStaleGate`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Restore`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Observe`: errors and returned state remain governed by the function's explicit branches.
- `mismatchDiff`: errors and returned state remain governed by the function's explicit branches.
- `clk.Advance`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Blocks`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Permanent`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 11 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
