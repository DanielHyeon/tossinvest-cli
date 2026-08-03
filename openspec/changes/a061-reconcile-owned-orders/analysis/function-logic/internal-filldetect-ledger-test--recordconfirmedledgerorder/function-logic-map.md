# Function Logic Map: `recordConfirmedLedgerOrder`

Source: `internal/filldetect/ledger_test.go`  
Function: `recordConfirmedLedgerOrder`  
Signature: `recordConfirmedLedgerOrder(params=3, results=0)`  
Source SHA-256: `3fa79a4b73613bae1f9ee086608012e2dca32a6f85b5b9128b916016fdfe571a`

## Inputs and invariants

- Inputs are the parameters represented by `recordConfirmedLedgerOrder(params=3, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger_test.go:64 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/ledger_test.go:67 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/ledger_test.go:70 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/ledger_test.go:73 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `t.Helper`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `j.Prepare`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `attempt.MarkDispatchStarted`: errors and returned state remain governed by the function's explicit branches.
- `attempt.MarkAcked`: errors and returned state remain governed by the function's explicit branches.
- `attempt.Settle`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
