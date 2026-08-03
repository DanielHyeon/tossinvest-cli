# Function Logic Map: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`  
Function: `TestStartupPrunesSpentNoncesOnce`  
Signature: `TestStartupPrunesSpentNoncesOnce(params=1, results=0)`  
Source SHA-256: `dec52090cdbaa17f2a868d8e204c1755d88ec127a3f5158efc858405f41b83b7`

## Inputs and invariants

- Inputs are the parameters represented by `TestStartupPrunesSpentNoncesOnce(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:451 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/engine_test.go:456 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/engine_test.go:460 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `isolate`: errors and returned state remain governed by the function's explicit branches.
- `writeEngineConfig`: errors and returned state remain governed by the function's explicit branches.
- `writeCredentials`: errors and returned state remain governed by the function's explicit branches.
- `engineStub`: errors and returned state remain governed by the function's explicit branches.
- `time.Now`: errors and returned state remain governed by the function's explicit branches.
- `seedSpentNonce`: errors and returned state remain governed by the function's explicit branches.
- `now.Add`: errors and returned state remain governed by the function's explicit branches.
- `startEngine`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `eng.Journal.NonceSpent`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 9 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
