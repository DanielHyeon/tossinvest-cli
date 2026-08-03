# Function Logic Map: `TestAdoptionIsSilentUnderReconcile`

Source: `internal/app/engine/reconcileloop_test.go`  
Function: `TestAdoptionIsSilentUnderReconcile`  
Signature: `TestAdoptionIsSilentUnderReconcile(params=1, results=0)`  
Source SHA-256: `feb0b59737a7c47e4ead572b77c9f2b591273fa6bd61744850a60c87830d6342`

## Inputs and invariants

- Inputs are the parameters represented by `TestAdoptionIsSilentUnderReconcile(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:359 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/reconcileloop_test.go:366 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/reconcileloop_test.go:370 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/app/engine/reconcileloop_test.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `newDriverHarness`: errors and returned state remain governed by the function's explicit branches.
- `h.holds`: errors and returned state remain governed by the function's explicit branches.
- `h.journal.EnterReconcile`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `h.cycle`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `h.alerts.count`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 4 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
