# Function Logic Map: `TestTerminalOrdersLeaveTheTrackedSet`

Source: `internal/filldetect/ledger_test.go`  
Function: `TestTerminalOrdersLeaveTheTrackedSet`  
Signature: `TestTerminalOrdersLeaveTheTrackedSet(params=1, results=0)`  
Source SHA-256: `3fa79a4b73613bae1f9ee086608012e2dca32a6f85b5b9128b916016fdfe571a`

## Inputs and invariants

- Inputs are the parameters represented by `TestTerminalOrdersLeaveTheTrackedSet(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger_test.go:285 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/ledger_test.go:289 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/ledger_test.go:292 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/ledger_test.go:306 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/ledger_test.go:309 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/filldetect/ledger_test.go:313 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/filldetect/ledger_test.go:316 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `clock.NewFake`: errors and returned state remain governed by the function's explicit branches.
- `filepath.Join`: errors and returned state remain governed by the function's explicit branches.
- `t.TempDir`: errors and returned state remain governed by the function's explicit branches.
- `newPager`: errors and returned state remain governed by the function's explicit branches.
- `page`: errors and returned state remain governed by the function's explicit branches.
- `newJournalDetector`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `recordConfirmedLedgerOrder`: errors and returned state remain governed by the function's explicit branches.
- `d.PollOnce`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `TrackedOrders`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `clk.Advance`: errors and returned state remain governed by the function's explicit branches.
- `rawOrders`: errors and returned state remain governed by the function's explicit branches.
- `pager.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `pager.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 11 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
