# Function Logic Map: `TestStartupRebuildsTheReconcileProjection`

Source: `internal/app/engine/engine_test.go`  
Function: `TestStartupRebuildsTheReconcileProjection`  
Signature: `TestStartupRebuildsTheReconcileProjection(params=1, results=0)`  
Source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

## Inputs and invariants

- Inputs are the parameters represented by `TestStartupRebuildsTheReconcileProjection(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:656 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/app/engine/engine_test.go:661 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/app/engine/engine_test.go:666 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/app/engine/engine_test.go:669 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/app/engine/engine_test.go:673 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/app/engine/engine_test.go:676 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `isolate`: errors and returned state remain governed by the function's explicit branches.
- `writeEngineConfig`: errors and returned state remain governed by the function's explicit branches.
- `writeCredentials`: errors and returned state remain governed by the function's explicit branches.
- `engineStub`: errors and returned state remain governed by the function's explicit branches.
- `seedAccountWideReconcile`: errors and returned state remain governed by the function's explicit branches.
- `startEngine`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `eng.Entry.CheckEntry`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `strings.Contains`: errors and returned state remain governed by the function's explicit branches.
- `eng.Reconcile.Blocks`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `t.Error`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
