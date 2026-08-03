# Function Logic Map: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`  
Function: `EntryGate.CheckEntryFor`  
Signature: `EntryGate.CheckEntryFor(params=2, results=1)`  
Source SHA-256: `46d559bec6ca59e70876da33face5ae8dadce7d832c5c32e105dd2add8cc8617`

## Inputs and invariants

- Inputs are the parameters represented by `EntryGate.CheckEntryFor(params=2, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/execgw/symbolgate.go:223 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/execgw/symbolgate.go:224 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/execgw/symbolgate.go:231 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/execgw/symbolgate.go:234 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | range | internal/execgw/symbolgate.go:243 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/execgw/symbolgate.go:244 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/execgw/symbolgate.go:248 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | range | internal/execgw/symbolgate.go:252 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/execgw/symbolgate.go:253 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/execgw/symbolgate.go:256 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `g.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `g.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- `refresh`: errors and returned state remain governed by the function's explicit branches.
- `err.Error`: errors and returned state remain governed by the function's explicit branches.
- `g.checkAccountEntry`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `symbolKey`: errors and returned state remain governed by the function's explicit branches.
- `strings.EqualFold`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
