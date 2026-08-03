# Function Logic Map: `applyHookFixture`

Source: `internal/journal/apply_hook_test.go`  
Function: `applyHookFixture`  
Signature: `applyHookFixture(params=1, results=1)`  
Source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`

## Inputs and invariants

- Inputs are the parameters represented by `applyHookFixture(params=1, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/journal/apply_hook_test.go:29 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `t.Helper`: errors and returned state remain governed by the function's explicit branches.
- `openTestJournal`: errors and returned state remain governed by the function's explicit branches.
- `recordConfirmedFillOrder`: errors and returned state remain governed by the function's explicit branches.
- `insertPosition`: errors and returned state remain governed by the function's explicit branches.
- `insertExitState`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 1 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
