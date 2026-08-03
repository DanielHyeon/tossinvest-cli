# Function Logic Map: `Journal.scopedLineageChildrenV16`

Source: `internal/journal/lineage.go`  
Function: `Journal.scopedLineageChildrenV16`  
Signature: `Journal.scopedLineageChildrenV16(params=3, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.scopedLineageChildrenV16(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/lineage.go:375 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `j.queryScopedLineageChildren`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- No assignment point is present; the function is observational or delegates its effect.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
