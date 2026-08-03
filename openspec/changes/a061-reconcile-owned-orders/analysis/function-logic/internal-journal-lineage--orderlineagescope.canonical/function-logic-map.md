# Function Logic Map: `OrderLineageScope.canonical`

Source: `internal/journal/lineage.go`  
Function: `OrderLineageScope.canonical`  
Signature: `OrderLineageScope.canonical(params=0, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters represented by `OrderLineageScope.canonical(params=0, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/journal/lineage.go:72 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | case | internal/journal/lineage.go:73 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | case | internal/journal/lineage.go:75 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | case | internal/journal/lineage.go:77 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | case | internal/journal/lineage.go:79 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | case | internal/journal/lineage.go:81 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `normaliseMarket`: errors and returned state remain governed by the function's explicit branches.
- `normaliseSymbol`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
