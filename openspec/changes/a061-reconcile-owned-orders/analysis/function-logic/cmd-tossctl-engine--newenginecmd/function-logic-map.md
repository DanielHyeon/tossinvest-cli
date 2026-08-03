# Function Logic Map: `newEngineCmd`

Source: `cmd/tossctl/engine.go`  
Function: `newEngineCmd`  
Signature: `newEngineCmd(params=1, results=1)`  
Source SHA-256: `45414562be8a352d2183fb2dfc0985154e0eea5ce781e167eb6800841c495451`

## Inputs and invariants

- Inputs are the parameters represented by `newEngineCmd(params=1, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | cmd/tossctl/engine.go:112 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `cmd.AddCommand`: errors and returned state remain governed by the function's explicit branches.
- `newEngineRunCmd`: errors and returned state remain governed by the function's explicit branches.
- `newEngineReconcileResolveCmd`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 1 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
