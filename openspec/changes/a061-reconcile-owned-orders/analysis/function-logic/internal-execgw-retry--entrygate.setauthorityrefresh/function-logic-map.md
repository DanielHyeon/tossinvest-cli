# Function Logic Map: `EntryGate.SetAuthorityRefresh`

Source: `internal/execgw/retry.go`  
Function: `EntryGate.SetAuthorityRefresh`  
Signature: `EntryGate.SetAuthorityRefresh(params=1, results=0)`  
Source SHA-256: `a549135ef2864ab05eb8168cfac899cad5052457189307c4ed3e3bee42e102d3`

## Inputs and invariants

- Inputs are the parameters represented by `EntryGate.SetAuthorityRefresh(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/execgw/retry.go:451 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `g.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `g.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 1 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
