# Function Logic Map: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`

Source: `internal/journal/nonce_test.go`  
Function: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`  
Signature: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld(params=1, results=0)`  
Source SHA-256: `84f12358991f53b64e3f9dbebef2730533e6375e8ea55289c80b9cdfeb35487f`

## Inputs and invariants

- Inputs are the parameters represented by `TestPrunePreservesSpentNonceWhileItsReservationIsHeld(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce_test.go:263 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/nonce_test.go:271 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/nonce_test.go:274 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openReservationJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `recordEntryDecision`: errors and returned state remain governed by the function's explicit branches.
- `j.Reserve`: errors and returned state remain governed by the function's explicit branches.
- `exposureReserve`: errors and returned state remain governed by the function's explicit branches.
- `mustVersion`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `spendNonce`: errors and returned state remain governed by the function's explicit branches.
- `fake.Advance`: errors and returned state remain governed by the function's explicit branches.
- `j.PruneSpentNonces`: errors and returned state remain governed by the function's explicit branches.
- `fake.Now`: errors and returned state remain governed by the function's explicit branches.
- `spentNonceCount`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 5 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
