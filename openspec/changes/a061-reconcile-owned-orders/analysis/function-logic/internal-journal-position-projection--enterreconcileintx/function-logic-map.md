# Function Logic Map: `enterReconcileInTx`

Source: `internal/journal/position_projection.go`  
Function: `enterReconcileInTx`  
Signature: `enterReconcileInTx(params=5, results=1)`  
Source SHA-256: `ae74d3ba1b66a05360e7b5851248fd6814577fa0b34068a89f52c58c10644c7b`

## Inputs and invariants

- Inputs are the parameters represented by `enterReconcileInTx(params=5, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/position_projection.go:369 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `reconcileCauseFor`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Sprintf`: errors and returned state remain governed by the function's explicit branches.
- `enterReconcileScopeInTx`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 2 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
