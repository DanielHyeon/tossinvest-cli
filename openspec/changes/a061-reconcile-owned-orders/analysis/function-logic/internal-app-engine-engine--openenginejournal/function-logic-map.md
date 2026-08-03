# Function Logic Map: `openEngineJournal`

Source: `internal/app/engine/engine.go`  
Function: `openEngineJournal`  
Signature: `openEngineJournal(params=3, results=2)`  
Source SHA-256: `401ab52518aac369f7567a60f711c4a019efad96ab0bcd7af1751155ba67e1f5`

## Inputs and invariants

- Inputs are the parameters represented by `openEngineJournal(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine.go:590 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/engine.go:598 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/engine.go:606 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/app/engine/engine.go:622 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/app/engine/engine.go:625 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `filepath.Join`: errors and returned state remain governed by the function's explicit branches.
- `journal.Open`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `j.SweepReservations`: errors and returned state remain governed by the function's explicit branches.
- `j.Close`: errors and returned state remain governed by the function's explicit branches.
- `j.MaxDecisionTTL`: errors and returned state remain governed by the function's explicit branches.
- `j.PruneSpentNonces`: errors and returned state remain governed by the function's explicit branches.
- `clk.Now`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 10 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
