# Function Logic Map: `TestAFailingApplyHookRollsBackTheFill`

Source: `internal/journal/apply_hook_test.go`  
Function: `TestAFailingApplyHookRollsBackTheFill`  
Signature: `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)`  
Source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`

## Inputs and invariants

- Inputs are the parameters represented by `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/apply_hook_test.go:151 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/apply_hook_test.go:154 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/apply_hook_test.go:169 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/apply_hook_test.go:175 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/apply_hook_test.go:179 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/apply_hook_test.go:182 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/apply_hook_test.go:187 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/apply_hook_test.go:191 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/apply_hook_test.go:197 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/apply_hook_test.go:200 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `applyHookFixture`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `errors.New`: errors and returned state remain governed by the function's explicit branches.
- `j.SetApplyHooks`: errors and returned state remain governed by the function's explicit branches.
- `tx.Exec`: errors and returned state remain governed by the function's explicit branches.
- `t.Error`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `j.RecordFill`: errors and returned state remain governed by the function's explicit branches.
- `observation`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `j.LookupFill`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `j.FillEvents`: errors and returned state remain governed by the function's explicit branches.
- `len`: errors and returned state remain governed by the function's explicit branches.
- `Scan`: errors and returned state remain governed by the function's explicit branches.
- `j.db.QueryRowContext`: errors and returned state remain governed by the function's explicit branches.
- `j.TrackedFillOrders`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 10 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
