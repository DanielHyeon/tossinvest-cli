# Function Logic Map: `TestAnOperatorResolutionSurvivesARestart`

Source: `internal/reconcile/restore_test.go`  
Function: `TestAnOperatorResolutionSurvivesARestart`  
Signature: `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)`  
Source SHA-256: `06361705cca4cd1d8cfd0263dff7b47ea9c661cac3c4b09bd164ba91b75c67f4`

## Inputs and invariants

- Inputs are the parameters represented by `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/reconcile/restore_test.go:147 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/reconcile/restore_test.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/reconcile/restore_test.go:157 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/reconcile/restore_test.go:160 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/reconcile/restore_test.go:163 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `clock.NewFake`: errors and returned state remain governed by the function's explicit branches.
- `t.TempDir`: errors and returned state remain governed by the function's explicit branches.
- `openJournalAt`: errors and returned state remain governed by the function's explicit branches.
- `trackerOn`: errors and returned state remain governed by the function's explicit branches.
- `noStaleGate`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Observe`: errors and returned state remain governed by the function's explicit branches.
- `mismatchDiff`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Resolve`: errors and returned state remain governed by the function's explicit branches.
- `first.Close`: errors and returned state remain governed by the function's explicit branches.
- `restarted.Restore`: errors and returned state remain governed by the function's explicit branches.
- `restartedGate.CheckEntryFor`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `restarted.Blocks`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 13 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
