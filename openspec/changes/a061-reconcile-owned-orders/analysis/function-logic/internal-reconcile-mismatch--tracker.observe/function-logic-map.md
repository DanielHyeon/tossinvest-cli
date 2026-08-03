# Function Logic Map: `Tracker.Observe`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.Observe`  
Signature: `Tracker.Observe(params=2, results=2)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters represented by `Tracker.Observe(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:366 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/reconcile/mismatch.go:373 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | else | internal/reconcile/mismatch.go:396 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/reconcile/mismatch.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/reconcile/mismatch.go:381 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/reconcile/mismatch.go:388 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | range | internal/reconcile/mismatch.go:398 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/reconcile/mismatch.go:399 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/reconcile/mismatch.go:405 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/reconcile/mismatch.go:418 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | range | internal/reconcile/mismatch.go:440 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | range | internal/reconcile/mismatch.go:443 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | if | internal/reconcile/mismatch.go:444 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B14 | range | internal/reconcile/mismatch.go:449 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B15 | range | internal/reconcile/mismatch.go:453 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B16 | range | internal/reconcile/mismatch.go:456 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B17 | if | internal/reconcile/mismatch.go:462 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B18 | else | internal/reconcile/mismatch.go:475 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B19 | range | internal/reconcile/mismatch.go:467 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B20 | if | internal/reconcile/mismatch.go:468 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `Now`: errors and returned state remain governed by the function's explicit branches.
- `t.clock`: errors and returned state remain governed by the function's explicit branches.
- `t.interval`: errors and returned state remain governed by the function's explicit branches.
- `t.maxFailures`: errors and returned state remain governed by the function's explicit branches.
- `t.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `diff.BlocksEntry`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `blocksFor`: errors and returned state remain governed by the function's explicit branches.
- `block.Key`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Sprintf`: errors and returned state remain governed by the function's explicit branches.
- `permanent.Key`: errors and returned state remain governed by the function's explicit branches.
- `sortBlocks`: errors and returned state remain governed by the function's explicit branches.
- `t.syncGate`: errors and returned state remain governed by the function's explicit branches.
- `t.snapshotBlocks`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `t.persist`: errors and returned state remain governed by the function's explicit branches.
- `delete`: errors and returned state remain governed by the function's explicit branches.
- `hasPermanentQuantityAccountBlock`: errors and returned state remain governed by the function's explicit branches.
- `now.Add`: errors and returned state remain governed by the function's explicit branches.
- `t.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 39 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
