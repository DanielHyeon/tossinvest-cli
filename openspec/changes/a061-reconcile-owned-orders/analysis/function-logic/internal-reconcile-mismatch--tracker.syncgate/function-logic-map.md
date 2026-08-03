# Function Logic Map: `Tracker.syncGate`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.syncGate`  
Signature: `Tracker.syncGate(params=1, results=0)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters represented by `Tracker.syncGate(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:879 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | range | internal/reconcile/mismatch.go:883 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/reconcile/mismatch.go:884 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/reconcile/mismatch.go:894 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/reconcile/mismatch.go:895 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | range | internal/reconcile/mismatch.go:901 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/reconcile/mismatch.go:902 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/reconcile/mismatch.go:908 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | range | internal/reconcile/mismatch.go:913 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/reconcile/mismatch.go:917 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | else | internal/reconcile/mismatch.go:919 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.BlockSymbol`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.SymbolBlocks`: errors and returned state remain governed by the function's explicit branches.
- `isReconcileReason`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.ClearSymbol`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.Block`: errors and returned state remain governed by the function's explicit branches.
- `t.Gate.Clear`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
