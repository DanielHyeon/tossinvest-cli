# Function Logic Map: `buildGateway`

Source: `internal/app/engine/gateway.go`  
Function: `buildGateway`  
Signature: `buildGateway(params=2, results=2)`  
Source SHA-256: `3dead101adcc3b89767975b14f72de7246909ac0ef3f909e3928ebed2637ee8b`

## Inputs and invariants

- Inputs are the parameters represented by `buildGateway(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/gateway.go:201 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/gateway.go:223 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/gateway.go:260 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `checkProjectionWired`: errors and returned state remain governed by the function's explicit branches.
- `execgw.NewEntryGate`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Restore`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `entry.SetAuthorityRefresh`: errors and returned state remain governed by the function's explicit branches.
- `tracker.Refresh`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `execgw.New`: errors and returned state remain governed by the function's explicit branches.
- `in.official.BaseURL`: errors and returned state remain governed by the function's explicit branches.
- `newNotifier`: errors and returned state remain governed by the function's explicit branches.
- `newRetrier`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 13 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
