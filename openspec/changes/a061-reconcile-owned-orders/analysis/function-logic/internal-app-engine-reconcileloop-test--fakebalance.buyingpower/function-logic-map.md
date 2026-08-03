# Function Logic Map: `fakeBalance.BuyingPower`

Source: `internal/app/engine/reconcileloop_test.go`  
Function: `fakeBalance.BuyingPower`  
Signature: `fakeBalance.BuyingPower(params=2, results=2)`  
Source SHA-256: `f7244f04d716230ddc2536f8e219958c52b86a6b899cf6f4df45fa09962f961e`

## Inputs and invariants

- Inputs are the parameters represented by `fakeBalance.BuyingPower(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:90 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- No outbound call is present; behavior is local and deterministic.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- No assignment point is present; the function is observational or delegates its effect.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
