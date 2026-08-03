# Function Logic Map: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`  
Function: `JournalTracked.TrackedOrders`  
Signature: `JournalTracked.TrackedOrders(params=1, results=2)`  
Source SHA-256: `59f7c08036ddcbab21f9b9e938856c39bec98b2f38c64b39ba9ce334986547db`

## Inputs and invariants

- Inputs are the parameters represented by `JournalTracked.TrackedOrders(params=1, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:108 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/ledger.go:111 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/ledger.go:115 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | range | internal/filldetect/ledger.go:119 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `t.Journal.TrackedFillOrders`: errors and returned state remain governed by the function's explicit branches.
- `make`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 3 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
