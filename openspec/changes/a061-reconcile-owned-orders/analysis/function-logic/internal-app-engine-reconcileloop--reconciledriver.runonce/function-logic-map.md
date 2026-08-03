# Function Logic Map: `ReconcileDriver.RunOnce`

Source: `internal/app/engine/reconcileloop.go`  
Function: `ReconcileDriver.RunOnce`  
Signature: `ReconcileDriver.RunOnce(params=1, results=1)`  
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`

## Inputs and invariants

- Inputs are the parameters represented by `ReconcileDriver.RunOnce(params=1, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop.go:387 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/reconcileloop.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/reconcileloop.go:404 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/app/engine/reconcileloop.go:410 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/app/engine/reconcileloop.go:416 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/app/engine/reconcileloop.go:417 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/app/engine/reconcileloop.go:425 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/app/engine/reconcileloop.go:426 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `d.note`: errors and returned state remain governed by the function's explicit branches.
- `d.stabilise`: errors and returned state remain governed by the function's explicit branches.
- `reconcile.LocalStateFromJournal`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `Compare`: errors and returned state remain governed by the function's explicit branches.
- `d.ingest.IngestExternalPositions`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `d.opts.Converge.ConvergeQuantities`: errors and returned state remain governed by the function's explicit branches.
- `d.opts.Tracker.Refresh`: errors and returned state remain governed by the function's explicit branches.
- `d.opts.Tracker.Observe`: errors and returned state remain governed by the function's explicit branches.
- `d.opts.Tracker.Blocks`: errors and returned state remain governed by the function's explicit branches.
- `d.judgeHoldings`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 18 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
