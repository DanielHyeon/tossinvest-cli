# Function Logic Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`  
Function: `Journal.ResolveCurrentOrderIDScoped`  
Signature: `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:313 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/lineage.go:317 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | for | internal/journal/lineage.go:323 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/lineage.go:325 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | switch | internal/journal/lineage.go:328 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | case | internal/journal/lineage.go:329 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | case | internal/journal/lineage.go:331 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | case | internal/journal/lineage.go:333 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/lineage.go:338 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `scope.canonical`: errors and returned state remain governed by the function's explicit branches.
- `j.scopedLineageChildren`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `j.recordScopedLineageConflict`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 7 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
